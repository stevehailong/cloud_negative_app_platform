# Helm 完整部署流程

## 概述

本文档说明如何使用 Helm 进行完整的应用部署，确保与环境定义严格一致，而不是单纯创建 Deployment。

## 架构设计

### 传统部署 vs Helm 部署

| 方面 | 传统部署 | Helm 部署 |
|------|---------|----------|
| 资源创建 | 手动创建各个资源 | 使用 Chart 模板一次性创建 |
| 配置管理 | 分散在多个地方 | 集中在 values.yaml |
| 环境隔离 | 代码实现 | Chart 模板实现 |
| 一致性 | 难以保证 | 严格保证 |
| 回滚能力 | 有限 | 完整支持 |

### 部署流程

```
用户请求部署
    ↓
获取环境定义（Environment）
    ↓
获取环境模板（EnvTemplate）
    ↓
构建 Helm Values
    ↓
执行 Helm Install/Upgrade
    ↓
创建完整资源栈
    ├── Namespace（如果不存在）
    ├── NetworkPolicy（网络隔离）
    ├── ResourceQuota（资源配额）
    ├── ServiceAccount（权限控制）
    ├── ConfigMap（配置）
    ├── Secret（敏感信息）
    ├── Deployment（工作负载）
    ├── Service（服务发现）
    ├── Ingress（公网访问）
    └── HPA（自动扩缩容）
```

## 核心组件

### 1. Helm 客户端 (`pkg/helm/client.go`)

提供 Helm 操作的封装：

```go
// 安装或升级 Release
func (c *Client) InstallOrUpgrade(ctx context.Context, releaseName, namespace, chartPath string, values map[string]interface{}) error

// 卸载 Release
func (c *Client) Uninstall(ctx context.Context, releaseName, namespace string) error

// 等待部署完成
func (c *Client) WaitForRelease(ctx context.Context, releaseName, namespace string, timeout time.Duration) error
```

### 2. Values 构建器 (`pkg/helm/values_builder.go`)

将环境定义转换为 Helm Values：

```go
// 从模板和环境配置构建 Values
func (b *ValuesBuilder) BuildFromTemplate(templateValues string, config DeploymentConfig) (map[string]interface{}, error)
```

支持的配置项：
- 基础配置（镜像、副本数）
- 资源配置（CPU、内存）
- 服务配置（端口、类型）
- Ingress 配置（域名、TLS）
- 健康检查（存活探针、就绪探针）
- 环境变量
- ConfigMap
- Secret
- HPA

### 3. 部署服务 (`internal/deploy/service/app_deployment_service.go`)

集成 Helm 部署到现有流程：

```go
// 创建 K8s 部署（优先使用 Helm）
func (s *AppDeploymentService) createK8sDeployment(ctx context.Context, deployment *model.AppDeployment, imageURL string) error {
    // 使用 Helm 进行完整部署
    if s.helmClient != nil {
        return s.deployWithHelmChart(ctx, deployment, imageURL)
    }
    
    // 降级：使用传统方式
    return s.createK8sDeploymentLegacy(ctx, deployment, imageURL)
}
```

## 环境定义与资源配置对应关系

### 环境类型自动配置

| 环境类型 | Service 类型 | Ingress | HPA | TLS |
|---------|-------------|---------|-----|-----|
| dev | NodePort | 不启用 | 不启用 | 不启用 |
| test | ClusterIP | 启用 | 不启用 | 不启用 |
| staging | ClusterIP | 启用 | 启用 | 启用 |
| prod | ClusterIP | 启用 | 启用 | 启用 |

### 从环境定义读取配置

```go
// Environment 模型
type Environment struct {
    EnvType     string  // 环境类型
    Namespace   string  // 命名空间
    TemplateID  *uint   // 环境模板 ID
    ConfigJSON  string  // 额外配置（JSON）
}

// EnvTemplate 模型
type EnvTemplate struct {
    ValuesYAML   string  // Helm Values YAML
}

// AppEnvBinding 模型（资源配置）
type AppEnvBinding struct {
    Replicas      int
    CPURequest    string
    CPULimit      string
    MemoryRequest string
    MemoryLimit   string
    ConfigJSON    string  // 额外配置
}
```

## 与 Go 微服务标准模板的对应

### 标准模板资源清单

| 资源 | Helm Chart 模板 | 说明 |
|------|----------------|------|
| Deployment | `templates/deployment.yaml` | ✅ 必需 |
| Service | `templates/service.yaml` | ✅ 必需 |
| ServiceAccount | `templates/serviceaccount.yaml` | ✅ 必需 |
| Ingress | `templates/ingress.yaml` | 可选（根据 values.ingress.enabled） |
| ConfigMap | `templates/configmap.yaml` | 可选（根据 values.configMap.enabled） |
| Secret | `templates/secret.yaml` | 可选（根据 values.secret.enabled） |
| HPA | `templates/hpa.yaml` | 可选（根据 values.autoscaling.enabled） |
| NetworkPolicy | ❌ | 后端服务自动创建 |
| ResourceQuota | ❌ | 后端服务自动创建 |

### 配置对应关系

#### values-go-microservice.yaml（标准模板）

```yaml
replicaCount: 3

image:
  repository: harbor.mycompany.com/mycloud/go-service
  tag: "v1.0.0"

service:
  type: ClusterIP
  port: 8080
  targetPort: 8080

ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: go-service.mycloud.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: go-service-tls
      hosts:
        - go-service.mycloud.com

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80

livenessProbe:
  enabled: true
  httpGet:
    path: /health
    port: 8080

readinessProbe:
  enabled: true
  httpGet:
    path: /ready
    port: 8080

env:
  - name: GIN_MODE
    value: "release"
  - name: LOG_LEVEL
    value: "info"
  - name: PORT
    value: "8080"

configMap:
  enabled: true
  data:
    app.yaml: |
      server:
        port: 8080

secret:
  enabled: true
  data:
    DB_PASSWORD: bXlwYXNzd29yZDEyMw==
```

#### 环境定义 -> Values 映射

```go
// 从 Environment 和 AppEnvBinding 生成上述 Values
config := helm.DeploymentConfig{
    AppName:         "go-service",
    Image:          "harbor.mycompany.com/mycloud/go-service:v1.0.0",
    Replicas:       3,
    ServiceType:    "ClusterIP",
    ServicePort:    8080,
    ContainerPort:  8080,
    
    CPURequest:     "500m",
    CPULimit:       "1000m",
    MemoryRequest:  "512Mi",
    MemoryLimit:    "1Gi",
    
    IngressEnabled:    true,
    IngressHost:       "go-service.mycloud.com",
    IngressTLSEnabled: true,
    
    LivenessPath:  "/health",
    ReadinessPath: "/ready",
    
    HPAEnabled:     true,
    HPAMinReplicas: 2,
    HPAMaxReplicas: 10,
    HPATargetCPU:   80,
}
```

## 使用方法

### 1. 测试部署

```bash
# 运行测试脚本
./test_helm_deployment.sh
```

### 2. 通过 API 部署

```bash
# 调用部署 API
curl -X POST http://localhost:8080/api/v1/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "env_id": 1,
    "image_url": "my-registry.com/my-app:v1.0.0",
    "replicas": 3
  }'
```

### 3. 手动 Helm 部署

```bash
# 使用自定义 values
helm install my-app ./helm-charts/mycloud-app \
  -n app-1-dev \
  --create-namespace \
  -f values-go-microservice.yaml
```

## 验证清单

部署完成后，验证以下资源是否创建：

- [ ] Deployment 运行正常（3/3 副本）
- [ ] Service 已创建（ClusterIP 或 NodePort）
- [ ] ServiceAccount 已创建
- [ ] Ingress 已创建（如果启用）
- [ ] ConfigMap 已创建（如果启用）
- [ ] Secret 已创建（如果启用）
- [ ] HPA 已创建（如果启用）
- [ ] NetworkPolicy 已创建（后端自动创建）
- [ ] ResourceQuota 已创建（后端自动创建）

## 故障排查

### Helm 安装失败

```bash
# 查看 Helm 状态
helm status <release-name> -n <namespace>

# 查看 Helm 历史
helm history <release-name> -n <namespace>

# 查看 Kubernetes 事件
kubectl get events -n <namespace> --sort-by='.lastTimestamp'
```

### 资源未创建

```bash
# 检查 Helm 渲染结果
helm template <release-name> ./helm-charts/mycloud-app -f values.yaml

# 检查 dry-run
helm install <release-name> ./helm-charts/mycloud-app -f values.yaml --dry-run
```

### 网络隔离问题

```bash
# 检查 NetworkPolicy
kubectl get networkpolicy -n <namespace>
kubectl describe networkpolicy <policy-name> -n <namespace>

# 测试网络连通性
kubectl run test --image=busybox -it --rm -- wget -qO- http://<service-name>:<port>
```

## 最佳实践

1. **环境模板管理**
   - 为不同环境类型创建标准模板
   - 模板中定义合理的默认值
   - 通过环境 ConfigJSON 覆盖特定配置

2. **资源隔离**
   - 每个应用在每个环境使用独立命名空间
   - 使用 NetworkPolicy 实现网络隔离
   - 使用 ResourceQuota 限制资源使用

3. **配置管理**
   - 使用 ConfigMap 存储非敏感配置
   - 使用 Secret 存储敏感信息
   - 避免在镜像中硬编码配置

4. **监控和日志**
   - 配置健康检查端点
   - 使用统一的日志格式
   - 集成到监控系统

## 后续改进

- [ ] 支持从远程 Chart 仓库部署
- [ ] 支持 Chart 版本管理
- [ ] 支持部署前预览（dry-run）
- [ ] 支持自定义 Chart 模板
- [ ] 支持多集群部署
- [ ] 集成 GitOps 工作流
