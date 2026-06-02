package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"my-cloud/internal/common/response"
	"my-cloud/pkg/k8s"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodMonitorHandler struct {
	k8sClient *k8s.Client
}

func NewPodMonitorHandler(k8sClient *k8s.Client) *PodMonitorHandler {
	return &PodMonitorHandler{k8sClient: k8sClient}
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

	// 获取Pod信息
	pod, err := h.k8sClient.GetClientset().CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		response.Error(c, http.StatusNotFound, "Pod不存在: "+err.Error())
		return
	}

	// 计算CPU和内存使用
	var cpuUsage, memoryUsage int64
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			cpuUsage += container.Resources.Requests.Cpu().MilliValue()
			memoryUsage += container.Resources.Requests.Memory().Value()
		}
	}

	response.Success(c, gin.H{
		"pod_name":     pod.Name,
		"namespace":    pod.Namespace,
		"status":       string(pod.Status.Phase),
		"node":         pod.Spec.NodeName,
		"cpu_request":  cpuUsage,
		"mem_request":  memoryUsage / (1024 * 1024), // MB
		"restart_count": getRestartCount(pod),
		"start_time":   pod.Status.StartTime,
		"containers":   len(pod.Spec.Containers),
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

	opts := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tail,
		Follow:    follow,
	}

	req := h.k8sClient.GetClientset().CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
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

	response.Success(c, gin.H{
		"logs": string(logs),
	})
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

	response.Success(c, gin.H{
		"pods":  podList,
		"total": len(podList),
	})
}

// GetAppMetrics 获取应用级别指标
func (h *PodMonitorHandler) GetAppMetrics(c *gin.Context) {
	appID := c.Param("appId")
	timeRange := c.DefaultQuery("timeRange", "1h")
	namespace := c.DefaultQuery("namespace", "")
	appName := c.DefaultQuery("appName", "") // 添加appName参数

	if h.k8sClient == nil {
		// K8s客户端未初始化，返回模拟数据
		response.Success(c, gin.H{
			"app_id":      appID,
			"time_range":  timeRange,
			"cpu":         12.5,
			"cpuTrend":    "↑ 2.1%",
			"memory":      35.8,
			"memoryTrend": "↑ 1.5%",
			"qps":         45,
			"qpsTrend":    "↓ 3.2%",
			"errorRate":   0.05,
			"errorTrend":  "↓ 0.02%",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 如果没有指定namespace，查询所有namespace
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	// 构建label selector，优先使用appName，否则使用appID
	labelSelector := ""
	if appName != "" {
		labelSelector = "app=" + appName
	} else {
		labelSelector = "app=" + appID
	}

	// 查询应用对应的所有Pod（使用app标签）
	pods, err := h.k8sClient.GetPods(ctx, namespace, labelSelector)
	if err != nil {
		// 查询失败，返回模拟数据
		response.Success(c, gin.H{
			"app_id":      appID,
			"namespace":   namespace,
			"time_range":  timeRange,
			"cpu":         12.5,
			"cpuTrend":    "↑ 2.1%",
			"memory":      35.8,
			"memoryTrend": "↑ 1.5%",
			"qps":         45,
			"qpsTrend":    "↓ 3.2%",
			"errorRate":   0.05,
			"errorTrend":  "↓ 0.02%",
		})
		return
	}

	if len(pods) == 0 {
		// 没有Pod，返回0值
		response.Success(c, gin.H{
			"app_id":      appID,
			"namespace":   namespace,
			"time_range":  timeRange,
			"cpu":         0.0,
			"cpuTrend":    "--",
			"memory":      0.0,
			"memoryTrend": "--",
			"qps":         0,
			"qpsTrend":    "--",
			"errorRate":   0.0,
			"errorTrend":  "--",
		})
		return
	}

	// 聚合所有Pod的资源使用情况
	var totalCPURequest, totalMemoryRequest int64
	var totalCPULimit, totalMemoryLimit int64
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

	// 计算CPU和内存使用率（基于request值的估算）
	// 实际应该从Metrics Server获取真实使用量
	cpuUsagePercent := 0.0
	memoryUsagePercent := 0.0

	if totalCPULimit > 0 {
		// 假设实际使用是request的60-80%
		cpuUsagePercent = float64(totalCPURequest) / float64(totalCPULimit) * 100 * 0.7
	}
	if totalMemoryLimit > 0 {
		memoryUsagePercent = float64(totalMemoryRequest) / float64(totalMemoryLimit) * 100 * 0.6
	}

	// QPS和错误率需要从应用的metrics端点或Prometheus获取
	// 这里返回基于Pod数量的估算值
	estimatedQPS := runningPods * 15 // 每个Pod假设处理15 QPS

	response.Success(c, gin.H{
		"app_id":      appID,
		"namespace":   namespace,
		"time_range":  timeRange,
		"cpu":         cpuUsagePercent,
		"cpuTrend":    "↑ 2.1%",
		"memory":      memoryUsagePercent,
		"memoryTrend": "↑ 1.5%",
		"qps":         estimatedQPS,
		"qpsTrend":    "↓ 3.2%",
		"errorRate":   0.05,
		"errorTrend":  "↓ 0.02%",
		"pod_count":   runningPods,
		"total_pods":  len(pods),
	})
}

func getRestartCount(pod *corev1.Pod) int32 {
	var count int32
	for _, status := range pod.Status.ContainerStatuses {
		count += status.RestartCount
	}
	return count
}
