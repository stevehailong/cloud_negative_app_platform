# MyCloud Application Helm Chart

这是为 My Cloud 平台设计的通用 Helm Chart，支持部署 Go、Python 等微服务应用。

## Chart 特性

- ✅ 支持多种应用类型（Go/Gin、Python/FastAPI）
- ✅ 自动扩缩容（HPA）
- ✅ 健康检查（Liveness & Readiness Probes）
- ✅ Ingress 路由配置
- ✅ ConfigMap 和 Secret 管理
- ✅ 持久化存储支持
- ✅ 资源限制和请求
- ✅ 服务账户和 RBAC

## 快速开始

### 1. 安装 Chart

```bash
# 基础安装
helm install my-app ./mycloud-app

# 使用自定义 values
helm install my-app ./mycloud-app -f values-go-microservice.yaml

# 指定命名空间
helm install my-app ./mycloud-app -n production --create-namespace
```

### 2. 使用特定值覆盖

```bash
helm install my-app ./mycloud-app \
  --set image.repository=myregistry.com/myapp \
  --set image.tag=v2.0.0 \
  --set replicaCount=5
```

### 3. 升级应用

```bash
# 升级到新版本
helm upgrade my-app ./mycloud-app -f values-go-microservice.yaml

# 升级并等待就绪
helm upgrade my-app ./mycloud-app --wait --timeout 5m
```

### 4. 回滚

```bash
# 查看历史
helm history my-app

# 回滚到上一个版本
helm rollback my-app

# 回滚到指定版本
helm rollback my-app 2
```

## 配置说明

### 镜像配置

```yaml
image:
  repository: myregistry.com/myapp  # 镜像仓库
  pullPolicy: IfNotPresent          # 拉取策略
  tag: "v1.0.0"                     # 镜像标签
```

### 资源配置

```yaml
resources:
  limits:
    cpu: 1000m      # CPU 限制
    memory: 1Gi     # 内存限制
  requests:
    cpu: 500m       # CPU 请求
    memory: 512Mi   # 内存请求
```

### 自动扩缩容

```yaml
autoscaling:
  enabled: true                       # 启用 HPA
  minReplicas: 2                      # 最小副本数
  maxReplicas: 10                     # 最大副本数
  targetCPUUtilizationPercentage: 80  # CPU 目标利用率
```

### Ingress 配置

```yaml
ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: myapp.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: myapp-tls
      hosts:
        - myapp.example.com
```

### 环境变量

```yaml
env:
  - name: APP_ENV
    value: "production"
  - name: LOG_LEVEL
    value: "info"
  - name: DB_HOST
    value: "mysql-service"
```

### ConfigMap 和 Secret

```yaml
configMap:
  enabled: true
  data:
    app.yaml: |
      key: value

secret:
  enabled: true
  data:
    DB_PASSWORD: cGFzc3dvcmQ=  # base64 编码
```

## 应用示例

### Go/Gin 微服务

```bash
helm install go-service ./mycloud-app -f values-go-microservice.yaml
```

### Python/FastAPI 微服务

```bash
helm install python-service ./mycloud-app -f values-python-microservice.yaml
```

## 验证部署

```bash
# 检查 Chart 语法
helm lint ./mycloud-app

# 模拟安装（不实际部署）
helm install my-app ./mycloud-app --dry-run --debug

# 渲染模板查看生成的 YAML
helm template my-app ./mycloud-app -f values-go-microservice.yaml

# 查看部署状态
helm status my-app

# 查看 Pod 状态
kubectl get pods -l app.kubernetes.io/instance=my-app

# 查看服务
kubectl get svc -l app.kubernetes.io/instance=my-app
```

## 卸载

```bash
helm uninstall my-app
```

## 与 My Cloud 平台集成

在 My Cloud 平台的"环境模板管理"中配置：

1. **模板类型**：选择 `Helm`
2. **仓库地址**：Chart 仓库地址（如 Harbor）
3. **Chart名称**：`mycloud-app`
4. **版本**：`1.0.0`
5. **自定义Values**：根据应用类型选择对应的 values 文件

## 高级配置

### 持久化存储

```yaml
persistence:
  enabled: true
  storageClass: "nfs-client"
  size: 20Gi
  mountPath: /data
```

### 节点选择

```yaml
nodeSelector:
  disktype: ssd
  
tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "app"
    effect: "NoSchedule"
```

### Pod 中断预算

```yaml
podDisruptionBudget:
  enabled: true
  minAvailable: 1
```

## 最佳实践

1. **使用特定标签**：为不同环境使用不同的 tag
2. **设置资源限制**：避免资源耗尽
3. **配置健康检查**：确保应用可用性
4. **启用 HPA**：应对流量波动
5. **使用 Secret 管理敏感信息**：不要硬编码密码
6. **配置 Ingress TLS**：启用 HTTPS

## 故障排查

```bash
# 查看 Pod 日志
kubectl logs -l app.kubernetes.io/instance=my-app

# 查看事件
kubectl get events --sort-by='.lastTimestamp'

# 描述 Pod
kubectl describe pod <pod-name>

# 进入容器
kubectl exec -it <pod-name> -- /bin/sh
```

## 版本历史

- v1.0.0: 初始版本，支持 Go 和 Python 微服务

## 维护者

My Cloud Team - team@mycloud.com
