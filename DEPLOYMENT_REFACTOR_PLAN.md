# 部署管理重构实施计划

## 总览

**目标：** 将部署管理从"每次部署一条记录"重构为"以应用为维度的主记录+历史记录"

**收益：**
- ✅ 一个应用只有一条主记录，操作更直观
- ✅ 重启、扩缩容、回滚等操作目标明确
- ✅ 历史记录独立管理，易于查询和审计
- ✅ 前端展示更清晰（当前状态 + 操作历史）

## 阶段划分

### 第一阶段：数据库设计与迁移（2天）

**任务清单：**

1. **创建新表结构**
   - [ ] app_deployments 表（主记录表）
   - [ ] deployment_history 表（历史记录表）
   - [ ] 添加索引和外键约束

2. **数据迁移脚本**
   - [ ] 从 deployments 表提取最新记录作为主记录
   - [ ] 将所有 deployments 记录导入 deployment_history
   - [ ] 验证数据完整性

3. **向后兼容**
   - [ ] 保留旧表 deployments（标记为 deprecated）
   - [ ] 创建视图兼容旧 API

**SQL脚本：**
```sql
-- 见 NEW_DEPLOYMENT_SCHEMA.sql
```

### 第二阶段：后端API开发（3天）

**任务清单：**

1. **Model 层**
   ```go
   // backend/internal/deploy/model/
   - app_deployment.go      // 主记录模型
   - deployment_history.go  // 历史记录模型
   ```

2. **Repository 层**
   ```go
   // backend/internal/deploy/repository/
   - app_deployment_repository.go
   - deployment_history_repository.go
   ```

3. **Service 层**
   ```go
   // backend/internal/deploy/service/
   - app_deployment_service.go
   
   Methods:
   - GetAppDeployment(appId, envId)
   - ListAppDeployments(filters)
   - Deploy(appDeploymentId, version, image)
   - Restart(appDeploymentId)
   - Scale(appDeploymentId, replicas)
   - Rollback(appDeploymentId, targetHistoryId)
   - GetHistory(appDeploymentId, pagination)
   ```

4. **Handler 层**
   ```go
   // backend/internal/deploy/handler/
   - app_deployment_handler.go
   
   Endpoints:
   - GET    /api/v1/app-deployments
   - GET    /api/v1/app-deployments/:id
   - POST   /api/v1/app-deployments/:id/deploy
   - POST   /api/v1/app-deployments/:id/restart
   - POST   /api/v1/app-deployments/:id/scale
   - POST   /api/v1/app-deployments/:id/rollback
   - GET    /api/v1/app-deployments/:id/history
   ```

5. **Router 注册**
   ```go
   // backend/internal/deploy/router/router.go
   api.GET("/app-deployments", appDeployHandler.List)
   api.GET("/app-deployments/:id", appDeployHandler.Get)
   api.POST("/app-deployments/:id/deploy", appDeployHandler.Deploy)
   // ...
   ```

### 第三阶段：前端适配（3天）

**任务清单：**

1. **页面重构**
   - [ ] 部署列表页（显示主记录）
   - [ ] 部署详情页（当前状态 + 历史记录）
   - [ ] 操作按钮（重启、扩缩容、回滚）

2. **API 对接**
   ```typescript
   // frontend/src/api/deployment.ts
   
   export const deploymentApi = {
     // 列表
     list: (params) => request.get('/api/v1/app-deployments', params),
     
     // 详情
     get: (id) => request.get(`/api/v1/app-deployments/${id}`),
     
     // 部署历史
     history: (id, params) => request.get(`/api/v1/app-deployments/${id}/history`, params),
     
     // 操作
     restart: (id, reason) => request.post(`/api/v1/app-deployments/${id}/restart`, {reason}),
     scale: (id, replicas, reason) => request.post(`/api/v1/app-deployments/${id}/scale`, {replicas, reason}),
     rollback: (id, targetHistoryId) => request.post(`/api/v1/app-deployments/${id}/rollback`, {targetHistoryId}),
   }
   ```

3. **UI 组件**
   ```typescript
   // frontend/src/views/deployment/
   - List.vue            // 部署列表
   - Detail.vue          // 部署详情
   - HistoryTimeline.vue // 历史记录时间线
   - RestartDialog.vue   // 重启对话框
   - ScaleDialog.vue     // 扩缩容对话框
   - RollbackDialog.vue  // 回滚对话框
   ```

### 第四阶段：集成测试（2天）

**测试用例：**

1. **部署流程测试**
   ```
   场景：CI自动部署
   步骤：
   1. 模拟CI触发部署
   2. 验证主记录更新
   3. 验证历史记录创建
   4. 验证K8s资源更新
   ```

2. **操作功能测试**
   ```
   场景：重启、扩缩容、回滚
   步骤：
   1. 重启：验证Pod重启 + 历史记录
   2. 扩缩容：验证副本数变化 + 历史记录
   3. 回滚：验证回到历史版本 + 历史记录
   ```

3. **前端交互测试**
   ```
   场景：用户操作流程
   步骤：
   1. 查看部署列表
   2. 点击详情查看
   3. 执行扩缩容操作
   4. 查看历史记录
   5. 执行回滚操作
   ```

### 第五阶段：灰度上线（1天）

**灰度策略：**

1. **内部用户优先**
   - 开发团队使用新API
   - 收集反馈和问题

2. **双写模式**
   ```go
   // 同时写入旧表和新表
   func (s *DeployService) Deploy(...) {
       // 写入新表
       s.appDeploymentRepo.Upsert(...)
       s.deploymentHistoryRepo.Create(...)
       
       // 向后兼容：写入旧表
       s.legacyDeploymentRepo.Create(...)
   }
   ```

3. **逐步切换**
   - 第1天：10%流量使用新API
   - 第2天：50%流量使用新API
   - 第3天：100%流量使用新API

4. **监控指标**
   - API响应时间
   - 错误率
   - 数据一致性

### 第六阶段：清理与文档（1天）

**清理任务：**

1. **代码清理**
   - [ ] 删除旧API代码
   - [ ] 删除兼容层代码
   - [ ] 更新单元测试

2. **数据库清理**
   - [ ] 备份旧表 deployments
   - [ ] 删除旧表（可选）
   - [ ] 优化新表索引

3. **文档更新**
   - [ ] API文档
   - [ ] 数据库文档
   - [ ] 用户手册
   - [ ] 运维手册

## 技术细节

### 关键点1：主记录的创建时机

**方案：** 在首次部署时自动创建主记录

```go
func (s *AppDeploymentService) Deploy(appId, envId, version, image string, replicas int) error {
    // 1. 查找或创建主记录
    appDeploy, err := s.repo.GetByAppAndEnv(appId, envId)
    if err != nil {
        // 不存在，创建新的主记录
        appDeploy = &AppDeployment{
            AppId:   appId,
            EnvId:   envId,
            Namespace: fmt.Sprintf("app-%d", appId),
            // ...
        }
        s.repo.Create(appDeploy)
    }
    
    // 2. 创建历史记录
    history := &DeploymentHistory{
        AppDeploymentId: appDeploy.Id,
        Type: "update",
        Version: version,
        ImageUrl: image,
        // ...
    }
    s.historyRepo.Create(history)
    
    // 3. 执行K8s部署
    s.k8sClient.Deploy(...)
    
    // 4. 更新主记录
    appDeploy.CurrentVersion = version
    appDeploy.CurrentImage = image
    appDeploy.LastDeployId = history.Id
    s.repo.Update(appDeploy)
}
```

### 关键点2：历史记录的变更追踪

**方案：** 使用JSON字段存储变更内容

```go
type ChangeLog struct {
    Field    string `json:"field"`
    OldValue string `json:"old_value"`
    NewValue string `json:"new_value"`
}

// 扩缩容场景
changes := map[string]ChangeLog{
    "replicas": {
        Field: "副本数",
        OldValue: "5",
        NewValue: "10",
    },
}

// 更新镜像场景
changes := map[string]ChangeLog{
    "image": {
        Field: "镜像",
        OldValue: "nginx:1.25",
        NewValue: "nginx:1.26",
    },
}
```

### 关键点3：回滚操作

**方案：** 回滚到历史记录的版本

```go
func (s *AppDeploymentService) Rollback(appDeploymentId, targetHistoryId uint) error {
    // 1. 查询目标历史记录
    targetHistory, err := s.historyRepo.GetById(targetHistoryId)
    if err != nil {
        return err
    }
    
    // 2. 使用历史记录的配置重新部署
    return s.Deploy(
        appDeploymentId,
        targetHistory.Version,
        targetHistory.ImageUrl,
        targetHistory.Replicas,
    )
    // 注意：会创建新的历史记录，type="rollback"
}
```

## 风险与应对

### 风险1：数据迁移失败

**应对：**
- 编写可逆的迁移脚本
- 在测试环境充分验证
- 生产环境先备份再迁移
- 保留回滚方案

### 风险2：新旧API不兼容

**应对：**
- 双写模式过渡
- 保留旧API一段时间
- 提供迁移指南
- 充分的集成测试

### 风险3：性能问题

**应对：**
- 添加合适的数据库索引
- 使用缓存（Redis）
- 监控慢查询
- 准备性能调优方案

## 预期效果

### 用户体验提升

**重构前：**
```
用户：想重启app-8的部署
问题：有5条app-8的记录，重启哪个？
结果：困惑、操作错误风险高
```

**重构后：**
```
用户：想重启app-8的部署
操作：点击app-8主记录的"重启"按钮
结果：清晰明确，零困惑
```

### 数据管理优化

**重构前：**
```
deployments 表
- 47条记录（app-8相关）
- 数据冗余
- 难以查询"当前状态"
```

**重构后：**
```
app_deployments 表
- 1条主记录（app-8当前状态）

deployment_history 表
- 47条历史记录（完整变更历史）
```

### 开发效率提升

**重构前：**
```go
// 查询当前部署
deployments := repo.FindByNamespace("app-8")
// 哪个是当前的？需要业务逻辑判断
```

**重构后：**
```go
// 查询当前部署
appDeploy := repo.GetByAppAndEnv(8, 1)
// 直接获取唯一的主记录
```

## 时间估算

| 阶段 | 任务 | 工时 | 人员 |
|-----|------|------|------|
| 1 | 数据库设计与迁移 | 2天 | 后端1人 |
| 2 | 后端API开发 | 3天 | 后端2人 |
| 3 | 前端适配 | 3天 | 前端2人 |
| 4 | 集成测试 | 2天 | 全员 |
| 5 | 灰度上线 | 1天 | 全员 |
| 6 | 清理与文档 | 1天 | 全员 |

**总计：** 12天（约2.5周）

## 总结

这次重构将彻底解决当前部署管理的痛点：
- ✅ 一个应用一条主记录，重启/扩缩容操作清晰明确
- ✅ 历史记录独立管理，支持查看完整的变更历史
- ✅ 前端交互更友好，用户体验大幅提升
- ✅ 代码结构更清晰，维护成本降低

**建议：** 优先级高，建议尽快启动实施！
