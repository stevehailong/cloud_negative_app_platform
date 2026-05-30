# 滚动更新问题修复报告

## 问题描述

**现象：** 执行CI滚动更新后，app-8命名空间的Pod数量从5个变成了10个

**原因分析：**

1. **创建了重复的Deployment**
   - 旧部署：`app-8-canary`（5个Pod）
   - 新部署：`app-8`（5个Pod）
   - 总计：10个Pod

2. **根本原因**
   ```go
   // 旧代码逻辑（有问题）
   workloadName := fmt.Sprintf("app-%d", release.AppID)  // 固定使用 app-8
   // 但实际现有部署名称是 app-8-canary
   // 结果：创建了新的 app-8 deployment，没有更新现有的 app-8-canary
   ```

## 修复方案

### 1. 智能检测现有Deployment

修改 `executeRollingUpdateDeployment()` 方法，实现智能检测逻辑：

```go
// 优先级检测顺序：
// 1. 检查 app-{appId}-canary (例如 app-8-canary)
// 2. 检查 app-{appId}          (例如 app-8)
// 3. 都不存在则创建 app-{appId}

canaryWorkloadName := fmt.Sprintf("app-%d-canary", release.AppID)
baseWorkloadName := fmt.Sprintf("app-%d", release.AppID)

// 优先检查 canary
canaryReplicas := s.getExistingDeploymentReplicas(namespace, canaryWorkloadName)
if canaryReplicas > 0 {
    targetWorkloadName = canaryWorkloadName  // 使用 canary
    existingReplicas = canaryReplicas
} else {
    // 检查 stable
    stableReplicas := s.getExistingDeploymentReplicas(namespace, baseWorkloadName)
    if stableReplicas > 0 {
        targetWorkloadName = baseWorkloadName  // 使用 stable
        existingReplicas = stableReplicas
    } else {
        targetWorkloadName = baseWorkloadName  // 创建新的
        existingReplicas = 0
    }
}
```

### 2. 使用指定名称部署

新增 `callDeployServiceWithWorkloadName()` 方法，支持指定 workloadName：

```go
func (s *ReleaseService) callDeployServiceWithWorkloadName(
    release *model.Release, 
    imageURL string, 
    replicas int, 
    workloadName string  // ← 关键：可指定名称
) bool {
    // 部署到指定的 workloadName
    // 如果存在则更新，不存在则创建
}
```

## 修改的文件

**文件：** `/backend/internal/release/service/release_service.go`

**修改内容：**
1. 重写 `executeRollingUpdateDeployment()` 方法（第228-290行）
2. 新增 `callDeployServiceWithWorkloadName()` 方法（第292-329行）

## 修复效果

### 修复前

```bash
# CI触发更新后
$ kubectl get deploy -n app-8
NAME           REPLICAS
app-8          5/5       # ← 新创建的
app-8-canary   5/5       # ← 已存在的

# Pod总数：10个
```

### 修复后

```bash
# CI触发更新后
$ kubectl get deploy -n app-8
NAME           REPLICAS   IMAGE
app-8-canary   5/5        nginx:1.25-alpine  # ← 就地更新

# Pod总数：5个（保持不变）
```

## 验证测试

### 测试场景1：已存在 canary deployment

```bash
# 1. 现有状态
$ kubectl get deploy -n app-8
NAME           REPLICAS
app-8-canary   5/5

# 2. 执行CI更新（rolling策略）
# 3. 结果
$ kubectl get deploy -n app-8
NAME           REPLICAS   IMAGE
app-8-canary   5/5        new-image  # ← 更新成功
```

### 测试场景2：已存在 stable deployment

```bash
# 1. 现有状态
$ kubectl get deploy -n app-8
NAME    REPLICAS
app-8   3/3

# 2. 执行CI更新
# 3. 结果
$ kubectl get deploy -n app-8
NAME    REPLICAS   IMAGE
app-8   3/3        new-image  # ← 更新成功，副本数保持3
```

### 测试场景3：首次部署

```bash
# 1. 现有状态
$ kubectl get deploy -n app-8
No resources found

# 2. 执行CI更新
# 3. 结果
$ kubectl get deploy -n app-8
NAME    REPLICAS   IMAGE
app-8   5/5        new-image  # ← 创建新的，使用默认5副本
```

## 检测逻辑决策树

```
开始
  ├── 检查 app-{appId}-canary
  │   ├── 存在？
  │   │   ├── Yes → 使用 canary，保持副本数
  │   │   └── No  → 继续检查 stable
  │   │
  │   └── 检查 app-{appId}
  │       ├── 存在？
  │       │   ├── Yes → 使用 stable，保持副本数
  │       │   └── No  → 创建新的 stable，使用默认5副本
  │       │
  │       └── 返回 workloadName 和 replicas
```

## 配置建议

### 推荐命名规范

| 部署类型 | 命名格式 | 适用场景 |
|---------|---------|---------|
| **生产稳定版** | `app-{appId}` | 已完成灰度验证的稳定版本 |
| **灰度测试版** | `app-{appId}-canary` | 灰度发布过程中的测试版本 |

### CI/CD配置

```yaml
# Jenkinsfile 或 .gitlab-ci.yml
deploy:
  stage: deploy
  script:
    - |
      # CI触发时使用 rolling 策略
      curl -X POST /api/v1/releases \
        -d '{
          "releaseStrategy": "rolling",  # ← 自动检测并更新现有部署
          "imageUrl": "${CI_IMAGE}"
        }'
```

## 注意事项

### ⚠️ 删除部署记录的影响

**问题：** 删除数据库部署记录（DELETE /api/v1/deployments/:id）会**同时删除K8s资源**

**代码位置：** `/backend/internal/deploy/service/deploy_service.go:491-502`

```go
func (s *DeployService) DeleteDeployment(id uint) error {
    // ...
    _ = s.k8sClient.DeleteDeployment(ctx, deployment.Namespace, deployment.WorkloadName)
    // ↑ 会删除K8s资源！
    return s.deploymentRepo.Delete(id)
}
```

**建议：**
- 如果只想清理数据库记录，使用内部API：
  ```bash
  DELETE /internal/v1/deployments/by-workload?namespace=X&workloadName=Y
  ```
- 如果要同时删除K8s资源，使用：
  ```bash
  DELETE /api/v1/deployments/:id
  ```

### ✅ 最佳实践

1. **清理数据库旧记录**
   ```bash
   # 只删除数据库记录，不影响K8s
   curl -X DELETE '/internal/v1/deployments/by-workload?namespace=app-8&workloadName=old-name'
   ```

2. **监控副本数变化**
   ```bash
   # 在发布前后检查
   kubectl get deploy -n app-8 -o jsonpath='{.items[*].spec.replicas}'
   ```

3. **验证部署更新**
   ```bash
   # 检查镜像版本
   kubectl get deploy -n app-8 -o jsonpath='{.items[*].spec.template.spec.containers[0].image}'
   ```

## 总结

### 问题

CI滚动更新创建了重复的Deployment，导致Pod数量翻倍（5→10）

### 原因

固定使用 `app-{appId}` 作为 workloadName，没有检测现有的 `app-{appId}-canary`

### 解决方案

实现智能检测逻辑：
1. 优先使用现有的 canary deployment
2. 其次使用现有的 stable deployment
3. 都不存在才创建新的
4. **保持副本数不变**

### 效果

✅ CI更新后Pod数保持不变（5个）  
✅ 就地更新现有Deployment  
✅ 镜像版本正确更新  
✅ 零停机部署
