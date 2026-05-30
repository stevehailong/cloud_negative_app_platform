# 应用隔离改进总结

## 改进内容

已实现三层隔离机制，确保不同应用之间的资源和网络隔离。

## 1. 网络隔离（NetworkPolicy）

### 策略说明
每个应用 namespace 自动创建 `default-deny-all` NetworkPolicy：

**Ingress 规则（入站）：**
- ✅ 允许来自同一 namespace 内的流量
- ✅ 允许来自 `ingress-nginx` namespace 的流量（用于外部访问）
- ❌ 拒绝来自其他 namespace 的流量

**Egress 规则（出站）：**
- ✅ 允许所有出站流量（DNS、外部 API 等）

### 验证结果
```
From app-5 (no NetworkPolicy):
  → app-5-service:  200 OK (same namespace)
  → app-6-service:  timeout (blocked by app-6's policy)
  → app-8-service:  200 OK (no NetworkPolicy on app-8)

From app-6 (with NetworkPolicy):
  → app-6-service:  200 OK (same namespace)
  → app-5-service:  timeout (blocked by default deny)
  → app-8-service:  timeout (blocked by default deny)
```

### 实现细节
- 文件：`/backend/pkg/k8s/client.go`
- 方法：`EnsureNetworkPolicy()`
- 触发：每次部署时自动创建

## 2. 资源配额（ResourceQuota）

### 默认配额
每个应用 namespace 自动创建 `default-quota` ResourceQuota：

| 资源类型 | 配额限制 | 说明 |
|---------|---------|------|
| Pods | 50 | 最多 50 个 Pod |
| Services | 10 | 最多 10 个 Service |
| CPU Requests | 10 cores | 请求总和不超过 10 核 |
| Memory Requests | 20 GiB | 请求总和不超过 20GiB |
| CPU Limits | 20 cores | 限制总和不超过 20 核 |
| Memory Limits | 40 GiB | 限制总和不超过 40GiB |
| PVCs | 10 | 最多 10 个持久卷声明 |

### 容器资源限制
每个容器默认配置：
- **Requests**: CPU 100m, Memory 128Mi
- **Limits**: CPU 1000m, Memory 512Mi

### 当前使用情况（app-6）
```
Resource                Used   Hard
pods                    2      50
services                1      10
requests.cpu            200m   10
requests.memory         256Mi  20Gi
limits.cpu              2      20
limits.memory           1Gi    40Gi
persistentvolumeclaims  0      10
```

### 效果
- ✅ 防止单个应用耗尽集群资源
- ✅ 确保资源公平分配
- ✅ 超出配额时部署会被拒绝

### 实现细节
- 文件：`/backend/pkg/k8s/client.go`
- 方法：`EnsureResourceQuota()`
- 触发：每次部署时自动创建

## 3. RBAC 细粒度权限控制

### ServiceAccount 隔离
每个应用自动创建专用 ServiceAccount：
- ServiceAccount: `{appName}-sa`
- Role: `{appName}-role`
- RoleBinding: `{appName}-rolebinding`

### 权限范围
应用的 Pod 只能：
- **pods**: get, list（查看 Pod 列表）
- **pods/log**: get, list（查看 Pod 日志）
- **services**: get（查看 Service 信息）

**不能**：
- 创建、更新、删除任何资源
- 访问 Secrets、ConfigMaps
- 跨 namespace 操作
- 访问集群级别资源

### 验证
```bash
# 查看 app-6 的 RBAC 配置
$ kubectl get sa,role,rolebinding -n app-6

serviceaccount/app-6-sa
role.rbac.authorization.k8s.io/app-6-role
rolebinding.rbac.authorization.k8s.io/app-6-rolebinding

# Pod 使用专用 ServiceAccount
$ kubectl get pods -n app-6 -o jsonpath='{.items[*].spec.serviceAccountName}'
app-6-sa app-6-sa
```

### 实现细节
- 文件：`/backend/pkg/k8s/client.go`
- 方法：`EnsureServiceAccount()`
- 触发：每次部署时自动创建
- 绑定：在 `BuildDeploymentSpec()` 中设置 `serviceAccountName`

## 隔离架构图

```
┌─────────────────────────────────────────────────────────┐
│                      K8s Cluster                        │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Namespace: app-5 (旧应用，无隔离)                  │ │
│  │                                                     │ │
│  │  Deployment: app-5 (2 replicas)                    │ │
│  │  Service: app-5-service                            │ │
│  │  ❌ No NetworkPolicy                               │ │
│  │  ❌ No ResourceQuota                               │ │
│  │  ❌ No ServiceAccount                              │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Namespace: app-6 (新应用，完整隔离)                │ │
│  │  Labels: managed-by=my-cloud                       │ │
│  │                                                     │ │
│  │  ✅ NetworkPolicy: default-deny-all                │ │
│  │     - Ingress: same-ns ✓, ingress-nginx ✓         │ │
│  │     - Egress: all ✓                                │ │
│  │                                                     │ │
│  │  ✅ ResourceQuota: default-quota                   │ │
│  │     - Pods: 2/50                                   │ │
│  │     - CPU: 200m/10                                 │ │
│  │     - Memory: 256Mi/20Gi                           │ │
│  │                                                     │ │
│  │  ✅ ServiceAccount: app-6-sa                       │ │
│  │     - Role: app-6-role (pods:get/list)            │ │
│  │     - RoleBinding: app-6-rolebinding               │ │
│  │                                                     │ │
│  │  Deployment: app-6 (2 replicas)                    │ │
│  │    └─ Pod: app-6-xxx (serviceAccount: app-6-sa)   │ │
│  │    └─ Pod: app-6-xxx (serviceAccount: app-6-sa)   │ │
│  │                                                     │ │
│  │  Service: app-6-service                            │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │ Namespace: app-8 (旧应用，无隔离)                  │ │
│  │                                                     │ │
│  │  Deployment: app-8 (5 replicas)                    │ │
│  │  Service: app-8-service                            │ │
│  │  ❌ No NetworkPolicy                               │ │
│  │  ❌ No ResourceQuota                               │ │
│  │  ❌ No ServiceAccount                              │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## 网络访问矩阵

| From \ To | app-5 | app-6 | app-8 | 外部 |
|-----------|-------|-------|-------|------|
| **app-5** | ✅ | ❌ | ✅ | ✅ |
| **app-6** | ✅ (from same ns) | ✅ | ❌ | ✅ |
| **app-8** | ✅ | ❌ | ✅ | ✅ |
| **Ingress** | ✅ | ✅ | ✅ | - |

说明：
- ✅ = 允许访问
- ❌ = 被 NetworkPolicy 拒绝
- app-6 有完整的 NetworkPolicy 保护
- app-5 和 app-8 是旧应用，未启用 NetworkPolicy

## 自动化流程

新建应用部署时自动执行：

1. ✅ **创建 Namespace**（带 `managed-by=my-cloud` label）
2. ✅ **创建 NetworkPolicy**（默认拒绝跨 namespace 访问）
3. ✅ **创建 ResourceQuota**（限制资源使用）
4. ✅ **创建 ServiceAccount + Role + RoleBinding**（最小权限原则）
5. ✅ **创建 Service**（自动流量分配）
6. ✅ **创建 Deployment**（使用专用 ServiceAccount，配置资源限制）

## 配置文件

### 修改的文件
1. `/backend/pkg/k8s/client.go`
   - `EnsureNamespace()` - 添加 label
   - `EnsureNetworkPolicy()` - 新增
   - `EnsureResourceQuota()` - 新增
   - `EnsureServiceAccount()` - 新增
   - `BuildDeploymentSpec()` - 添加 ServiceAccount 和资源限制

2. `/backend/internal/deploy/service/deploy_service.go`
   - `executeK8sDeployment()` - 调用隔离措施创建方法

## 测试验证

### 创建测试应用
```bash
curl -X POST http://localhost:8087/internal/v1/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "releaseId": 101,
    "clusterId": 1,
    "namespace": "app-6",
    "workloadName": "app-6",
    "workloadType": "deployment",
    "imageVersion": "httpd:latest",
    "desiredReplicas": 2
  }'
```

### 验证隔离措施
```bash
# 1. 检查 NetworkPolicy
kubectl get networkpolicy -n app-6
kubectl describe networkpolicy default-deny-all -n app-6

# 2. 检查 ResourceQuota
kubectl get resourcequota -n app-6
kubectl describe resourcequota default-quota -n app-6

# 3. 检查 RBAC
kubectl get sa,role,rolebinding -n app-6
kubectl describe role app-6-role -n app-6

# 4. 验证 Pod 配置
kubectl get pods -n app-6 -o yaml | grep serviceAccountName
kubectl get pods -n app-6 -o yaml | grep -A 4 resources:

# 5. 测试网络隔离
kubectl run test-from-app5 -n app-5 --rm -i --restart=Never \
  --image=curlimages/curl:latest -- \
  curl -s -m 2 app-6-service.app-6.svc.cluster.local
# 预期：timeout（被NetworkPolicy阻止）
```

## 优势

### 安全性
- ✅ **零信任架构**：默认拒绝，显式允许
- ✅ **最小权限原则**：每个应用只有必要的权限
- ✅ **多层防护**：网络层 + 资源层 + 权限层

### 可靠性
- ✅ **防止资源争抢**：ResourceQuota 确保公平分配
- ✅ **故障隔离**：单个应用问题不影响其他应用
- ✅ **可预测性**：资源使用有明确上限

### 可管理性
- ✅ **自动化**：部署时自动配置，无需手动干预
- ✅ **标准化**：所有应用统一的隔离策略
- ✅ **可追溯**：通过 label 识别管理的资源

## 后续改进建议

1. **动态配额调整**：根据应用类型（生产/测试）动态设置配额
2. **监控告警**：集成 Prometheus 监控配额使用率
3. **审计日志**：记录 RBAC 访问尝试
4. **更细粒度的 NetworkPolicy**：支持自定义白名单
5. **PodSecurityPolicy/PodSecurity**：限制容器特权模式

## 总结

通过这三层改进，实现了：

✅ **Namespace 级别隔离**：每个应用独立 namespace  
✅ **网络隔离**：NetworkPolicy 限制跨 namespace 访问  
✅ **资源隔离**：ResourceQuota 防止资源耗尽  
✅ **权限隔离**：RBAC 限制 Pod 的 API 访问权限  
✅ **自动化部署**：所有隔离措施自动创建  

系统从"软隔离"（仅靠命名）升级为"强隔离"（多层防护），符合生产环境的安全和稳定性要求。
