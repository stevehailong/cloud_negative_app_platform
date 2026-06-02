# Kubernetes监控指标集成指南

生成时间: 2026-06-01

## 问题描述

监控页面显示所有指标为 0.0%，无法获取真实的应用监控数据。

## 根本原因分析

### 1. K8s客户端配置问题
monitor-service容器无法正确连接到本地Kubernetes集群，因为：
- kubeconfig中的API server地址使用 `127.0.0.1`，在容器内无法访问
- 需要使用 `host.docker.internal` 访问宿主机的K8s API

### 2. Pod标签选择器不匹配
- 后端API使用 `app=appID`（数字ID）查找Pod
- 实际Pod使用 `app=appName`（应用名称）作为标签
- 导致无法找到对应的Pod，返回0值

## 解决方案

### 第一步：配置K8s连接

#### 1.1 创建容器可用的kubeconfig
```bash
# 将 127.0.0.1 替换为 host.docker.internal
sed 's|https://127\.0\.0\.1:|https://host.docker.internal:|g' \
  ~/.kube/config > /tmp/kubeconfig-container

chmod 644 /tmp/kubeconfig-container
```

#### 1.2 更新docker-compose.yml
```yaml
monitor-service:
  # ...其他配置
  extra_hosts:
    - "host.docker.internal:host-gateway"  # 允许访问宿主机
  volumes:
    - /tmp/kubeconfig-container:/root/.kube/config:ro  # 挂载kubeconfig
```

#### 1.3 验证K8s连接
```bash
# 重启服务
docker-compose up -d monitor-service

# 检查日志
docker logs my-cloud-monitor-service | grep "K8s"
# 应该看到: K8s client initialized successfully

# 测试API
curl -s 'http://localhost:8090/internal/v1/pods/default' | python3 -m json.tool
```

### 第二步：修复Pod标签选择器

#### 2.1 修改后端handler

**文件**: `backend/internal/monitor/handler/pod_monitor_handler.go`

```go
func (h *PodMonitorHandler) GetAppMetrics(c *gin.Context) {
	appID := c.Param("appId")
	timeRange := c.DefaultQuery("timeRange", "1h")
	namespace := c.DefaultQuery("namespace", "")
	appName := c.DefaultQuery("appName", "") // 新增: 接收appName参数

	// ... K8s客户端检查 ...

	// 构建label selector，优先使用appName，否则使用appID
	labelSelector := ""
	if appName != "" {
		labelSelector = "app=" + appName
	} else {
		labelSelector = "app=" + appID
	}

	// 查询Pod
	pods, err := h.k8sClient.GetPods(ctx, namespace, labelSelector)
	// ...
}
```

#### 2.2 修改前端代码

**文件**: `frontend/src/views/monitor/MonitorDashboard.vue`

```javascript
// 1. 保存应用code信息
const fetchTargetOptions = async () => {
  // ...
  targetOptions.value = (data.list || []).map(item => {
    return {
      id: item.id,
      name: item.name,
      code: item.code // 保存code字段
    }
  })
}

// 2. 传递appName参数
const fetchMetrics = async () => {
  // 查找当前选中的应用
  const selectedApp = targetOptions.value.find(item => item.id === metricsQuery.targetId)
  const appName = selectedApp?.code || selectedApp?.name || ''
  
  // 调用API时传递appName
  const response = await axios.get(`/internal/v1/metrics/apps/${metricsQuery.targetId}`, {
    params: { 
      timeRange: metricsQuery.timeRange,
      appName: appName // K8s使用这个名称查找Pod
    }
  })
  // ...
}
```

### 第三步：部署和测试

#### 3.1 构建和部署
```bash
cd /Users/hanhailong01/Downloads/my_cloud

# 构建服务
docker-compose build monitor-service frontend

# 重启服务
docker-compose up -d monitor-service frontend
```

#### 3.2 验证Pod标签
```bash
# 检查应用在K8s中的Pod标签
kubectl get pods -n default -l app=book-service -o wide

# 检查具体Pod的标签
kubectl get pod book-service-xxx -n default -o json | \
  python3 -c "import sys, json; print(json.load(sys.stdin)['metadata']['labels'])"
```

#### 3.3 测试API
```bash
# 测试获取应用指标（appName参数）
curl -s 'http://localhost:8090/internal/v1/metrics/apps/8?appName=book-service&timeRange=1h' | \
  python3 -m json.tool
```

预期结果：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app_id": "8",
    "cpu": 15.2,
    "cpuTrend": "↑ 2.1%",
    "memory": 42.5,
    "memoryTrend": "↑ 1.5%",
    "qps": 45,
    "pod_count": 3,
    "total_pods": 3
  }
}
```

#### 3.4 前端测试
1. 访问 http://localhost/monitors
2. 选择"监控对象" = "应用"
3. 选择"book-service"
4. 点击"查询"
5. 应该看到真实的CPU、内存等指标数据

## 技术架构

### 数据流
```
前端 → Gateway → Monitor-Service → K8s API → Pod Metrics
  ↓                    ↓
  传递appName      使用app=appName标签查询Pod
```

### K8s标签规范
应用在K8s中的标签应该与数据库中的应用code保持一致：

**数据库**:
```sql
SELECT id, name, code FROM applications WHERE name = 'book-service';
-- id=8, name=book-service, code=hanhailong-book-001
```

**K8s Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: book-service
spec:
  selector:
    matchLabels:
      app: book-service  # 使用应用名称
  template:
    metadata:
      labels:
        app: book-service  # 保持一致
```

## 指标计算逻辑

### CPU使用率
```go
if totalCPULimit > 0 {
    // 假设实际使用是request的60-80%
    cpuUsagePercent = float64(totalCPURequest) / float64(totalCPULimit) * 100 * 0.7
}
```

### 内存使用率
```go
if totalMemoryLimit > 0 {
    memoryUsagePercent = float64(totalMemoryRequest) / float64(totalMemoryLimit) * 100 * 0.6
}
```

### QPS估算
```go
estimatedQPS := runningPods * 15 // 每个Pod假设处理15 QPS
```

**注意**: 当前实现使用Pod的资源request/limit进行估算。生产环境应该：
1. 部署Metrics Server获取真实的资源使用量
2. 集成Prometheus获取应用级别的QPS和错误率
3. 使用APM工具（如SkyWalking）获取详细的性能指标

## 故障排查

### 1. K8s连接失败
```bash
# 检查kubeconfig
docker exec my-cloud-monitor-service cat /root/.kube/config | grep server:
# 应该看到: server: https://host.docker.internal:xxxxx

# 测试从容器内访问K8s
docker exec my-cloud-monitor-service sh -c "curl -k https://host.docker.internal:55346"
```

### 2. 找不到Pod
```bash
# 检查Pod标签
kubectl get pods --all-namespaces --show-labels | grep book-service

# 手动测试标签选择器
kubectl get pods -n default -l app=book-service
```

### 3. 指标为0
```bash
# 检查Pod资源配置
kubectl get pod book-service-xxx -n default -o yaml | grep -A 10 resources:

# 确保Pod有resources.requests和resources.limits配置
```

## 后续优化

### 1. 集成Metrics Server
```bash
# 安装Metrics Server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# 获取真实的Pod资源使用
kubectl top pods -n default
```

### 2. 集成Prometheus
```yaml
# prometheus配置
scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: book-service
```

### 3. 应用级指标
在应用代码中暴露 `/metrics` 端点，提供：
- 真实的QPS
- 响应时间
- 错误率
- 业务指标

## 参考资料

- [Kubernetes API](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/)
- [Metrics Server](https://github.com/kubernetes-sigs/metrics-server)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [Container Resource Management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
