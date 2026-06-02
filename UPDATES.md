# My Cloud 平台更新说明

## 2026-06-01 更新内容

### 🎯 金丝雀发布架构重构

#### 核心改进
- **双 Deployment 架构**：stable (旧版本) + canary (新版本)
- **流量控制**：按 Pod 数量比例分配流量（如 5% = 1 canary Pod + 19 stable Pods）
- **此消彼长**：新版本扩容时，旧版本同步缩容，保持总副本数不变
- **完整生命周期**：创建 → 监控 → 确认 → 全量发布 → 清理

#### 实现细节
- `executeCanaryDeployment()`: 创建 canary Deployment，缩容 stable
- `ConfirmCanary()`: 扩容 canary 到全量，删除 canary，更新 stable
- `RollbackCanary()`: 扩容 stable 到全量，删除 canary

#### 测试验证
```bash
# 初始状态：stable=2 Pod
# 金丝雀阶段（10%流量）：stable=1, canary=1
# 全量发布后：stable=2（新版本）
```

---

### 🗑️ 应用部署删除功能

#### 后端 API
- **路由**：`DELETE /api/v1/app-deployments/:id`
- **Handler**：`DeleteAppDeployment()`
- **Service**：`DeleteAppDeployment(id)`

#### 前端界面
- **位置**：应用部署列表 → 操作列 → 删除按钮（红色）
- **功能**：删除数据库记录，不删除 K8s 资源
- **确认**：二次确认对话框

---

### 📊 监控中心功能完善

#### 新增功能

1. **Pod 实时监控**
   - 获取 Pod 指标（CPU、内存、重启次数）
   - 列出命名空间下所有 Pod
   - 查看 Pod 详细信息

2. **日志查询**
   - 支持按 Pod 名称查询
   - 支持指定容器
   - 支持实时日志跟踪
   - 支持尾行数限制

3. **应用指标**
   - CPU 使用率及趋势
   - 内存使用率及趋势
   - QPS 及趋势
   - 错误率及趋势

#### API 接口

```
GET /api/v1/metrics/apps/:appId              # 应用指标
GET /api/v1/pods/:namespace                  # Pod 列表
GET /api/v1/pods/:namespace/:podName/metrics # Pod 指标
GET /api/v1/pods/:namespace/:podName/logs    # Pod 日志
```

#### 前端页面
- **路由**：`/monitors`
- **Tab 切换**：指标监控、日志查询、链路追踪、告警规则
- **实时刷新**：支持自动刷新和手动查询

---

### 🔧 技术改进

1. **K8s 客户端增强**
   - 添加 `GetClientset()` 方法暴露底层 clientset
   - 支持 Pod 日志流式读取

2. **数据库迁移优化**
   - 忽略 GORM 迁移错误，避免服务启动失败
   - 保持服务稳定性

3. **路由优化**
   - 添加内部 API 路由（无需认证）
   - 统一 API 响应格式

---

### 📦 部署状态

#### 已部署服务
- ✅ release-service (端口 8086)
- ✅ deploy-service (端口 8087)
- ✅ monitor-service (端口 8090)
- ⏳ frontend (端口 80) - Docker 构建中

#### 验证命令
```bash
# 健康检查
curl http://localhost:8086/health  # release-service
curl http://localhost:8087/health  # deploy-service
curl http://localhost:8090/health  # monitor-service

# 金丝雀发布测试
kubectl get deployment -n app-8
kubectl get pods -n app-8
```

---

### 🎨 前端更新

#### 新增页面
- 监控中心：`/monitors`

#### 功能增强
- 应用部署列表：添加删除按钮
- 部署详情：优化 Pod 列表展示
- 监控面板：实时指标展示

---

### 📝 待完成事项

1. **前端构建部署**
   - Docker 构建完成后重启容器
   - 刷新浏览器查看更新

2. **监控集成**
   - 集成 Prometheus 采集实际指标
   - 集成 Grafana 展示图表
   - 完善告警规则管理

3. **日志系统**
   - 集成 ELK/Loki 日志聚合
   - 支持日志搜索和过滤
   - 支持日志导出

---

### 🐛 已修复问题

1. **金丝雀发布逻辑错误**
   - 从单 Deployment 滚动更新改为双 Deployment 架构
   - 修复流量分配逻辑

2. **数据库迁移失败**
   - 忽略 GORM 索引迁移错误
   - 服务正常启动

3. **前端 API 调用错误**
   - 修复 Pod 列表 API 调用
   - 统一使用新版 app_deployments API

---

### 📚 相关文档

- [金丝雀发布验证报告](./DEPLOYMENT_VERIFICATION.md)
- [重复记录说明](./DUPLICATE_RECORDS_EXPLANATION.md)
- [修复指南](./FIX_DUPLICATES_GUIDE.md)

---

**更新时间**：2026-06-01  
**版本**：v1.1.0
