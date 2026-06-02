# 命名空间隔离设计方案

## 🎯 设计目标

采用**企业级Kubernetes命名空间隔离方案**，这是最流行的多租户隔离模式。

## 📐 架构设计

### 隔离层次

```
项目 (Project)
  └── 环境 (Environment)  ← 核心隔离单元
       ├── 命名空间 (Namespace) 1:1映射
       ├── 集群 (Cluster)
       └── 应用部署 (AppDeployment)
            └── Workloads (Deployment/StatefulSet/DaemonSet)
```

### 核心原则

1. **一个环境 = 一个命名空间**（在特定集群中）
2. **应用部署的namespace必须从环境获取**，不允许手动指定
3. **同一集群中，namespace全局唯一**（通过数据库约束保证）
4. **支持K8s原生隔离机制**：
   - RBAC（角色访问控制）
   - ResourceQuota（资源配额）
   - LimitRange（资源限制范围）
   - NetworkPolicy（网络策略）
   - Pod Security（Pod安全策略）
   - Ingress/Gateway（访问隔离）

## 🔧 实现细节

### 1. 数据库约束

```sql
-- 环境表添加组合唯一索引
ALTER TABLE environments 
ADD UNIQUE KEY uk_cluster_namespace (cluster_id, namespace);
```

**说明**：
- 确保在同一个集群中，每个namespace只能被一个环境使用
- 不同集群可以有相同名称的namespace（但不推荐）

### 2. 环境创建流程

```
用户创建环境
  ├── 选择项目
  ├── 选择集群
  ├── 指定namespace名称
  ├── 检查: cluster_id + namespace 是否已存在
  ├── 创建环境记录
  └── (可选) 在K8s中创建namespace及相关资源
       ├── Namespace
       ├── ResourceQuota
       ├── LimitRange
       └── NetworkPolicy
```

### 3. 应用部署流程

```
应用部署到环境
  ├── 指定 app_id 和 env_id
  ├── 查询环境信息，获取 namespace 和 cluster_id
  ├── 构建 workload_name (app-{appID} 或 app-{appID}-canary)
  ├── 创建 AppDeployment 记录
  │    ├── namespace: 从环境获取
  │    ├── cluster_id: 从环境获取
  │    └── workload_name: 自动生成
  └── 在K8s中创建/更新 Deployment
```

## 📊 数据模型

### Environment (环境表)

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| id | BIGINT | 主键 | PK |
| env_code | VARCHAR(64) | 环境编码 | UNIQUE |
| env_name | VARCHAR(128) | 环境名称 | |
| env_type | VARCHAR(32) | 环境类型 | dev/test/staging/prod/preview |
| cluster_id | BIGINT | 集群ID | NOT NULL |
| **namespace** | VARCHAR(128) | **命名空间** | **NOT NULL** |
| project_id | BIGINT | 项目ID | NOT NULL |
| description | VARCHAR(255) | 描述 | |
| config_json | JSON | 环境配置 | ResourceQuota/LimitRange等 |
| status | TINYINT | 状态 | 1-启用 0-禁用 |

**索引**：
- `uk_cluster_namespace (cluster_id, namespace)` - 组合唯一索引

### AppDeployment (应用部署表)

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| id | BIGINT | 主键 | PK |
| app_id | BIGINT | 应用ID | NOT NULL |
| env_id | BIGINT | 环境ID | NOT NULL |
| cluster_id | BIGINT | 集群ID | 从环境继承 |
| **namespace** | VARCHAR(255) | **命名空间** | **从环境继承** |
| workload_name | VARCHAR(255) | 工作负载名称 | 自动生成 |
| workload_type | VARCHAR(50) | 类型 | deployment/statefulset |
| ... | ... | ... | ... |

**索引**：
- `uk_namespace_workload (namespace, workload_name)` - 组合唯一索引
- `idx_app_env (app_id, env_id)` - 复合索引

**约束逻辑**：
- 一个应用在一个环境最多2条记录：stable + canary
- workload_name 格式：`app-{appID}` 或 `app-{appID}-canary`

## 🚀 使用场景

### 场景1：多团队共享集群

```
集群: production-k8s
  ├── namespace: team-frontend  (前端团队环境)
  ├── namespace: team-backend   (后端团队环境)
  └── namespace: team-data      (数据团队环境)
```

### 场景2：同一应用的多环境隔离

```
项目: 电商平台
  ├── 环境: dev-001 → namespace: ecom-dev
  ├── 环境: test-001 → namespace: ecom-test
  ├── 环境: staging-001 → namespace: ecom-staging
  └── 环境: prod-001 → namespace: ecom-prod
```

### 场景3：多业务线隔离

```
集群: shared-cluster
  ├── namespace: bizline-retail   (零售业务线)
  ├── namespace: bizline-finance  (金融业务线)
  └── namespace: bizline-logistics (物流业务线)
```

## ⚙️ K8s资源配额示例

### ResourceQuota (资源配额)

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: env-quota
  namespace: ecom-dev
spec:
  hard:
    requests.cpu: "10"
    requests.memory: 20Gi
    limits.cpu: "20"
    limits.memory: 40Gi
    persistentvolumeclaims: "10"
    pods: "50"
```

### LimitRange (默认资源限制)

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: env-limits
  namespace: ecom-dev
spec:
  limits:
  - default:
      cpu: 500m
      memory: 512Mi
    defaultRequest:
      cpu: 100m
      memory: 128Mi
    type: Container
```

### NetworkPolicy (网络隔离)

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-from-other-namespaces
  namespace: ecom-dev
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector: {}  # 只允许同namespace内的Pod通信
```

## 🔒 安全增强

### 1. RBAC示例

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: env-developer
  namespace: ecom-dev
rules:
- apiGroups: ["", "apps", "batch"]
  resources: ["pods", "deployments", "jobs", "services"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
```

### 2. Pod Security Standard

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ecom-dev
  labels:
    pod-security.kubernetes.io/enforce: baseline
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

## 📝 前端界面建议

### 环境创建表单

```
┌─────────────────────────────────────┐
│  新建环境                           │
├─────────────────────────────────────┤
│  环境编码: [dev-001             ]  │
│  环境名称: [开发环境01          ]  │
│  环境类型: [开发环境 ▼]           │
│  所属项目: [电商平台 ▼]           │
│  所属集群: [本地K8s集群 ▼]        │
│  命名空间: [ecom-dev            ]  │
│                                      │
│  ⚠️ 提示: 命名空间在集群中必须唯一 │
│  命名规范: 小写字母、数字、短横线  │
│                                      │
│  [ 取消 ]           [ 创建 ]        │
└─────────────────────────────────────┘
```

### 应用部署表单

```
┌─────────────────────────────────────┐
│  部署应用: 用户服务                │
├─────────────────────────────────────┤
│  目标环境: [开发环境01 ▼]         │
│                                      │
│  📍 部署信息                        │
│  集群: 本地K8s集群 (自动)          │
│  命名空间: ecom-dev (自动)         │
│  工作负载: app-8 (自动)            │
│                                      │
│  镜像版本: [v1.2.3              ]  │
│  副本数量: [3                   ]  │
│                                      │
│  [ 取消 ]           [ 部署 ]        │
└─────────────────────────────────────┘
```

## ✅ 优势

1. **隔离性强**：利用K8s原生namespace机制
2. **易于管理**：一个环境对应一个namespace，概念清晰
3. **资源可控**：可配置ResourceQuota和LimitRange
4. **网络隔离**：可配置NetworkPolicy限制跨namespace通信
5. **权限明确**：基于RBAC的细粒度权限控制
6. **成本追踪**：可按namespace统计资源使用和成本
7. **企业标准**：符合大多数企业的Kubernetes多租户实践

## 🔄 迁移建议

对于已有的不规范数据：

1. **清理重复记录**：使用 `CleanupDuplicateDeployments` 函数
2. **验证namespace一致性**：确保app_deployment的namespace与环境定义一致
3. **添加约束**：在确保数据一致后，添加数据库唯一约束

---

**实施日期**: 2026-06-01  
**版本**: v1.0
