package handler

import (
	"context"
	"io"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"my-cloud/internal/common/response"
	"my-cloud/internal/monitor/integration"
	"my-cloud/pkg/k8s"
	"my-cloud/pkg/prometheus"

	"github.com/gin-gonic/gin"
	promclient "github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 自定义指标变量（由 RegisterMetrics 注册到指定 registry）
var (
	HttpRequestsTotal   *promclient.CounterVec
	MemoryUsagePercent  promclient.GaugeFunc
	ActiveUsersGauge    promclient.Gauge
	RequestsPerSecond   promclient.Gauge
)

// 内部计数器用于计算 RPS
var (
	httpReqBucket  atomic.Int64
	prevReqCount   atomic.Int64
	prevSampleTime atomic.Int64
)

// RegisterMetrics 将自定义指标注册到指定的 Prometheus registry
// 必须在 /metrics 端点启动前调用
func RegisterMetrics(reg promclient.Registerer) {
	log.Printf("Registering custom metrics to registry...")
	HttpRequestsTotal = promclient.NewCounterVec(
		promclient.CounterOpts{
			Name: "mycloud_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	reg.MustRegister(HttpRequestsTotal)
	log.Printf("Registered mycloud_http_requests_total")

	MemoryUsagePercent = promclient.NewGaugeFunc(
		promclient.GaugeOpts{
			Name: "mycloud_memory_usage_percent",
			Help: "Current memory usage as percent of limit",
		},
		func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			limit := uint64(512 * 1024 * 1024)
			return float64(m.Alloc) / float64(limit) * 100
		},
	)
	reg.MustRegister(MemoryUsagePercent)
	log.Printf("Registered mycloud_memory_usage_percent")

	ActiveUsersGauge = promclient.NewGauge(
		promclient.GaugeOpts{
			Name: "mycloud_active_users",
			Help: "Number of currently active users",
		},
	)
	reg.MustRegister(ActiveUsersGauge)
	log.Printf("Registered mycloud_active_users")

	RequestsPerSecond = promclient.NewGauge(
		promclient.GaugeOpts{
			Name: "mycloud_requests_per_second",
			Help: "Estimated requests per second",
		},
	)
	reg.MustRegister(RequestsPerSecond)
	log.Printf("Registered mycloud_requests_per_second")

	// 启动后台协程，每 5 秒采样一次 RPS
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			cur := httpReqBucket.Load()
			prev := prevReqCount.Swap(cur)
			now := time.Now().Unix()
			prevT := prevSampleTime.Swap(int64(now))
			delta := cur - prev
			dt := now - int64(prevT)
			if dt > 0 && RequestsPerSecond != nil {
				RequestsPerSecond.Set(float64(delta) / float64(dt))
			}
		}
	}()
	log.Printf("Custom metrics registered successfully")
}

// RecordRequest 记录一次 HTTP 请求（由 middleware 调用）
func RecordRequest(method, path, status string) {
	if HttpRequestsTotal != nil {
		HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
	}
	httpReqBucket.Add(1)
}

type PodMonitorHandler struct {
	k8sClient         *k8s.Client
	integrationLoader *integration.Loader
}

func NewPodMonitorHandler(k8sClient *k8s.Client, integrationLoader *integration.Loader) *PodMonitorHandler {
	return &PodMonitorHandler{
		k8sClient:         k8sClient,
		integrationLoader: integrationLoader,
	}
}

func (h *PodMonitorHandler) promClient() *prometheus.Client {
	if h.integrationLoader == nil {
		return nil
	}
	return h.integrationLoader.PrometheusClient()
}

// GetPodMetrics 获取Pod指标
func (h *PodMonitorHandler) GetPodMetrics(c *gin.Context) {
	namespace := c.Param("namespace")
	podName := c.Param("podName")

	if h.k8sClient == nil {
		response.Error(c, http.StatusServiceUnavailable, "K8s客户端未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pod, err := h.k8sClient.GetClientset().CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Pod不存在: "+err.Error())
		return
	}

	var cpuUsage, memoryUsage int64
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			cpuUsage += container.Resources.Requests.Cpu().MilliValue()
			memoryUsage += container.Resources.Requests.Memory().Value()
		}
	}

	response.Success(c, gin.H{
		"pod_name":      pod.Name,
		"namespace":     pod.Namespace,
		"status":        string(pod.Status.Phase),
		"node":          pod.Spec.NodeName,
		"cpu_request":   cpuUsage,
		"mem_request":   memoryUsage / (1024 * 1024),
		"restart_count": getRestartCount(pod),
		"start_time":    pod.Status.StartTime,
		"containers":    len(pod.Spec.Containers),
	})
}

// GetPodLogs 获取Pod日志
func (h *PodMonitorHandler) GetPodLogs(c *gin.Context) {
	namespace := c.Param("namespace")
	podName := c.Param("podName")
	container := c.Query("container")
	tailLines := c.DefaultQuery("tail", "100")
	follow := c.Query("follow") == "true"

	if h.k8sClient == nil {
		response.Error(c, http.StatusServiceUnavailable, "K8s客户端未初始化")
		return
	}

	tail, _ := strconv.ParseInt(tailLines, 10, 64)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &corev1.PodLogOptions{Container: container, TailLines: &tail, Follow: follow}
	stream, err := h.k8sClient.GetClientset().CoreV1().Pods(namespace).GetLogs(podName, opts).Stream(ctx)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取日志失败: "+err.Error())
		return
	}
	defer stream.Close()

	logs, err := io.ReadAll(stream)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "读取日志失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"logs": string(logs)})
}

// ListNamespacePods 获取命名空间下所有Pod
func (h *PodMonitorHandler) ListNamespacePods(c *gin.Context) {
	namespace := c.Param("namespace")

	if h.k8sClient == nil {
		response.Error(c, http.StatusServiceUnavailable, "K8s客户端未初始化")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pods, err := h.k8sClient.GetClientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取Pod列表失败: "+err.Error())
		return
	}

	var podList []gin.H
	for _, pod := range pods.Items {
		podList = append(podList, gin.H{
			"name":          pod.Name,
			"namespace":     pod.Namespace,
			"status":        string(pod.Status.Phase),
			"node":          pod.Spec.NodeName,
			"restart_count": getRestartCount(&pod),
			"start_time":    pod.Status.StartTime,
			"ip":            pod.Status.PodIP,
		})
	}

	response.Success(c, gin.H{"pods": podList, "total": len(podList)})
}

func promRange(timeRange string) string {
	switch timeRange {
	case "1h":
		return "5m"
	case "6h":
		return "15m"
	case "24h":
		return "1h"
	case "7d":
		return "6h"
	default:
		return "5m"
	}
}

// GetAppMetrics 获取应用级别指标
// 优先从 Prometheus 查询 cAdvisor 容器指标（按 K8s namespace 过滤）
// 未配置 Prometheus 时回退到基于 K8s Pod requests/limits 的估算
func (h *PodMonitorHandler) GetAppMetrics(c *gin.Context) {
	appID := c.Param("appId")
	timeRange := c.DefaultQuery("timeRange", "1h")
	namespace := c.DefaultQuery("namespace", "")
	appName := c.DefaultQuery("appName", "")

	_ = appName // 保留参数兼容性

	// 如果前端没传 namespace，尝试从 deploy_db 查询
	if namespace == "" && h.integrationLoader != nil {
		namespace = h.integrationLoader.LookupNamespace(appID)
	}

	if pc := h.promClient(); pc != nil && namespace != "" {
		data, err := h.queryServiceMetrics(pc, namespace, timeRange)
		if err == nil {
			data["app_id"] = appID
			data["namespace"] = namespace
			data["time_range"] = timeRange
			data["data_source"] = "prometheus"
			response.Success(c, data)
			return
		}
	}

	labelValue := appName
	if labelValue == "" {
		labelValue = appID
	}
	h.getAppMetricsFromK8s(c, appID, namespace, labelValue, timeRange)
}

// queryServiceMetrics 通过 Prometheus 查询指定应用的指标
// 使用 K8s namespace 标签过滤 cAdvisor 容器指标
func (h *PodMonitorHandler) queryServiceMetrics(pc *prometheus.Client, namespace, timeRange string) (gin.H, error) {
	rate := promRange(timeRange)

	// 过滤条件：排除 pause 容器
	containerFilter := `container!="",container!~"POD|pause.*"`

	// CPU：cAdvisor 容器 CPU 使用率（按 namespace 过滤）
	cpuQ := `sum(rate(container_cpu_usage_seconds_total{namespace="` + namespace + `",` + containerFilter + `}[` + rate + `])) * 100`

	// 内存：容器内存使用量 MB（按 namespace 过滤）
	memQ := `sum(container_memory_working_set_bytes{namespace="` + namespace + `",` + containerFilter + `}) / 1024 / 1024`

	// QPS：容器网络收包速率（近似 QPS，按 namespace 过滤）
	qpsQ := `sum(rate(container_network_receive_packets_total{namespace="` + namespace + `"}[` + rate + `]))`

	// 错误率：从应用的 /metrics 端点获取 mycloud_http_requests_total
	// 如果应用未暴露 /metrics，则为 0
	errRateQ := `0`

	// 执行查询
	cpu, _, cpuErr := pc.QueryScalar(cpuQ)
	mem, _, memErr := pc.QueryScalar(memQ)
	qps, _, _ := pc.QueryScalar(qpsQ)
	errRate, _, _ := pc.QueryScalar(errRateQ)

	// CPU 回退：cAdvisor 按 job 过滤（兼容旧版）
	if cpuErr != nil || cpu == 0 {
		cpuQ2 := `sum(rate(container_cpu_usage_seconds_total{` + containerFilter + `}[` + rate + `])) * 100`
		cpu, _, _ = pc.QueryScalar(cpuQ2)
	}

	// 内存回退
	if memErr != nil || mem == 0 {
		memQ2 := `sum(container_memory_working_set_bytes{` + containerFilter + `}) / 1024 / 1024`
		mem, _, _ = pc.QueryScalar(memQ2)
	}

	return gin.H{
		"cpu":         roundTo(cpu, 2),
		"cpuTrend":    "",
		"memory":      roundTo(mem, 2),
		"memoryTrend": "",
		"qps":         roundTo(qps, 2),
		"qpsTrend":    "",
		"errorRate":   roundTo(errRate, 2),
		"errorTrend":  "",
	}, nil
}

// getAppNamespace 已删除 — namespace 由前端通过 query param 传入

func roundTo(v float64, n int) float64 {
	if v != v { // NaN
		return 0
	}
	mult := 1.0
	for i := 0; i < n; i++ {
		mult *= 10
	}
	return float64(int64(v*mult+0.5)) / mult
}

func (h *PodMonitorHandler) getAppMetricsFromK8s(c *gin.Context, appID, namespace, labelValue, timeRange string) {
	if h.k8sClient == nil {
		response.Success(c, gin.H{
			"app_id": appID, "time_range": timeRange,
			"cpu": 0, "memory": 0, "qps": 0, "errorRate": 0,
			"data_source": "none",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	pods, err := h.k8sClient.GetPods(ctx, namespace, "app="+labelValue)
	if err != nil || len(pods) == 0 {
		response.Success(c, gin.H{
			"app_id": appID, "namespace": namespace, "time_range": timeRange,
			"cpu": 0, "memory": 0, "qps": 0, "errorRate": 0,
			"data_source": "k8s",
		})
		return
	}

	var totalCPURequest, totalMemoryRequest, totalCPULimit, totalMemoryLimit int64
	runningPods := 0

	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning {
			runningPods++
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests != nil {
					totalCPURequest += container.Resources.Requests.Cpu().MilliValue()
					totalMemoryRequest += container.Resources.Requests.Memory().Value()
				}
				if container.Resources.Limits != nil {
					totalCPULimit += container.Resources.Limits.Cpu().MilliValue()
					totalMemoryLimit += container.Resources.Limits.Memory().Value()
				}
			}
		}
	}

	cpuUsagePercent := 0.0
	memoryUsagePercent := 0.0
	if totalCPULimit > 0 {
		cpuUsagePercent = float64(totalCPURequest) / float64(totalCPULimit) * 100 * 0.7
	}
	if totalMemoryLimit > 0 {
		memoryUsagePercent = float64(totalMemoryRequest) / float64(totalMemoryLimit) * 100 * 0.6
	}

	response.Success(c, gin.H{
		"app_id":      appID,
		"namespace":   namespace,
		"time_range":  timeRange,
		"cpu":         cpuUsagePercent,
		"memory":      memoryUsagePercent,
		"qps":         runningPods * 15,
		"errorRate":   0,
		"pod_count":   runningPods,
		"total_pods":  len(pods),
		"data_source": "k8s",
	})
}

func getRestartCount(pod *corev1.Pod) int32 {
	var count int32
	for _, status := range pod.Status.ContainerStatuses {
		count += status.RestartCount
	}
	return count
}
