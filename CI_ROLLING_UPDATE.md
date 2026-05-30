# CI流水线滚动更新说明

## 功能概述

当代码修改触发CI流水线后，系统会自动创建**滚动更新（rolling）**发布策略，实现：
- ✅ **保持现有副本数不变**（例如5个Pod保持5个）
- ✅ **仅更新镜像版本**
- ✅ **零停机更新**（Kubernetes原生滚动更新）

## 更新流程

### 1. 代码修改触发CI

```bash
# 开发人员提交代码
git add .
git commit -m "feat: 更新功能"
git push origin main

# Jenkins CI流水线自动触发
# - 执行代码构建
# - 构建Docker镜像
# - 推送到镜像仓库
# - 创建发布工单
```

### 2. 系统自动创建发布工单

CI流水线完成后会调用API创建发布：

```json
POST /internal/v1/releases
{
  "appId": 8,
  "releaseVersion": "v2.1.0-abc123",
  "releaseStrategy": "rolling",  // ← 关键：rolling策略
  "imageUrl": "registry.mycloud.io/app:v2.1.0-abc123"
}
```

### 3. 审批并执行发布

管理员在**发布管理**页面：
1. 查看发布工单详情
2. 审批通过
3. 点击"执行发布"

### 4. 自动滚动更新

系统执行以下步骤：

```
1. 检测现有部署
   kubectl get deployment app-8-canary -n app-8
   → 发现已存在，副本数=5

2. 保持副本数，更新镜像
   kubectl set image deployment/app-8-canary \
     container=new-image:v2.1.0 -n app-8
   → 副本数保持5个不变

3. Kubernetes滚动更新
   - 创建新Pod（新镜像）
   - 等待新Pod就绪
   - 删除旧Pod
   - 重复直到所有Pod更新完成
```

## 关键特性

### 副本数保持逻辑

系统会智能检测：

| 场景 | 行为 |
|------|------|
| **首次部署** | 使用默认副本数（5个） |
| **已有部署** | 查询现有副本数并保持 |
| **手动扩容后** | 保持扩容后的副本数 |

### 代码实现（release_service.go）

```go
func (s *ReleaseService) executeRollingUpdateDeployment(release *model.Release) {
    // 1. 查询现有部署的副本数
    existingReplicas := s.getExistingDeploymentReplicas(namespace, workloadName)
    
    // 2. 确定目标副本数
    if existingReplicas > 0 {
        targetReplicas = existingReplicas  // 保持不变
    } else {
        targetReplicas = 5  // 首次部署默认值
    }
    
    // 3. 执行部署（更新镜像）
    s.callDeployService(release, release.ImageURL, targetReplicas)
}
```

## 与灰度发布的区别

| 特性 | 滚动更新（rolling） | 灰度发布（canary） |
|------|---------------------|-------------------|
| **副本数** | 保持不变 | 按比例分配（4:1） |
| **流量切换** | 全量更新 | 渐进式切换 |
| **部署对象** | 更新现有Deployment | 创建新Deployment |
| **适用场景** | CI自动部署 | 重大版本发布 |
| **回滚** | kubectl rollout undo | 删除Canary |

## 验证示例

### 场景：app-8当前5副本，触发CI更新

**更新前：**
```bash
$ kubectl get deploy -n app-8
NAME           READY   REPLICAS   IMAGE
app-8-canary   5/5     5          httpd:alpine
```

**执行CI更新后：**
```bash
$ kubectl get deploy -n app-8
NAME           READY   REPLICAS   IMAGE
app-8-canary   5/5     5          nginx:1.25-alpine  ← 镜像已更新
```

**验证点：**
- ✅ 副本数保持5个
- ✅ 镜像更新为新版本
- ✅ 无停机时间
- ✅ 旧Pod被新Pod替换

## API端点

### 查询现有部署副本数
```bash
GET /internal/v1/k8s/deployments/{namespace}/{name}/replicas

# 示例
curl http://localhost:8087/internal/v1/k8s/deployments/app-8/app-8-canary/replicas

# 响应
{
  "code": 0,
  "data": {
    "replicas": 5
  }
}
```

## 故障排查

### 问题1：副本数被重置为默认值

**原因：** 查询API失败或返回0
**排查：**
```bash
# 检查deploy-service日志
docker logs my-cloud-deploy-service | grep "GetK8sDeploymentReplicas"

# 手动测试查询API
curl http://localhost:8087/internal/v1/k8s/deployments/app-8/app-8-canary/replicas
```

### 问题2：更新后Pod数量改变

**原因：** 数据库中的desiredReplicas与K8s不一致
**解决：**
```bash
# 手动触发扩缩容以同步状态
curl -X POST http://localhost:8087/internal/v1/deployments/scale \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "app-8",
    "workloadName": "app-8-canary",
    "replicas": 5
  }'
```

## 配置说明

### Pipeline Service配置

文件：`/backend/internal/pipeline/service/pipeline_service.go:711`

```go
// CI触发的发布策略
"releaseStrategy": "rolling"  // ← 使用rolling策略
```

### Release Service配置

文件：`/backend/internal/release/service/release_service.go:234-303`

```go
// 滚动更新实现
func (s *ReleaseService) executeRollingUpdateDeployment(release *model.Release)
```

## 最佳实践

1. **CI触发后自动审批**：可配置特定分支（如main）的CI发布自动审批
2. **监控副本数变化**：在Prometheus中添加副本数监控指标
3. **回滚机制**：保留最近3个版本的镜像用于快速回滚
4. **健康检查**：确保应用配置了readiness probe

## 总结

通过引入 `rolling` 发布策略，系统实现了：
- ✅ CI触发的自动化部署
- ✅ 副本数智能保持
- ✅ 零停机滚动更新
- ✅ 与灰度发布策略互补

现在您可以放心地提交代码，系统会自动完成从构建到部署的全流程，同时保证服务的稳定性和连续性。
