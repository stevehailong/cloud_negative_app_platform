# My Cloud 平台快速访问指南

## 🌐 访问地址

### 前端页面
- **主页**：http://localhost/
- **应用管理**：http://localhost/apps
- **应用部署**：http://localhost/deployments
- **发布管理**：http://localhost/releases
- **监控中心**：http://localhost/monitors ⭐ 新增
- **集群管理**：http://localhost/clusters

### 后端服务
- **Gateway**：http://localhost:8080
- **Release Service**：http://localhost:8086
- **Deploy Service**：http://localhost:8087
- **Monitor Service**：http://localhost:8090 ⭐ 新增

---

## 🚀 核心功能

### 1. 金丝雀发布（已重构）

**操作流程**：
1. 进入发布管理页面
2. 创建新发布，选择"金丝雀发布"策略
3. 设置流量比例（如 5%、10%、20%）
4. 提交发布
5. 监控金丝雀 Pod 运行状态
6. 确认无误后，点击"确认金丝雀"完成全量发布

**验证命令**：
```bash
# 查看 Deployment
kubectl get deployment -n app-8

# 查看 Pod 分布
kubectl get pods -n app-8 -o wide

# 查看发布状态
curl http://localhost:8086/internal/v1/releases
```

---

### 2. 应用部署管理

**新增功能**：
- ✅ 删除部署记录（红色删除按钮）
- ✅ 查看 Pod 列表
- ✅ 查看部署历史
- ✅ 扩缩容操作
- ✅ 重启/回滚操作

**操作位置**：
- 应用部署列表：http://localhost/deployments
- 操作列：详情、重启、扩缩容、回滚、部署、**删除**

---

### 3. 监控中心（新增）

**访问地址**：http://localhost/monitors

**功能模块**：

#### 📊 指标监控
- CPU 使用率及趋势
- 内存使用率及趋势
- QPS（每秒请求数）
- 错误率统计

#### 📝 日志查询
- 应用日志查询
- Pod 日志查询
- 支持日志级别过滤（ERROR、WARN、INFO、DEBUG）
- 支持关键词搜索

#### 🔍 链路追踪
- TraceID 查询
- 应用链路查询
- Span 详情查看

#### 🚨 告警规则
- 告警规则管理（开发中）
- 告警历史查看
- 告警统计

---

## 🔧 API 接口

### 发布管理 API
```bash
# 创建金丝雀发布
POST http://localhost:8086/internal/v1/releases
{
  "appId": 8,
  "envId": 1,
  "releaseVersion": "v1.0.5",
  "imageUrl": "registry/image:tag",
  "releaseStrategy": "canary",
  "canaryPercent": 10
}

# 执行发布
POST http://localhost:8086/internal/v1/releases/{id}/execute

# 确认金丝雀
POST http://localhost:8086/internal/v1/releases/{id}/canary/confirm

# 回滚金丝雀
POST http://localhost:8086/internal/v1/releases/{id}/canary/rollback
```

### 监控 API
```bash
# 获取应用指标
GET http://localhost:8090/api/v1/metrics/apps/{appId}?timeRange=1h

# 获取 Pod 列表
GET http://localhost:8090/api/v1/pods/{namespace}

# 获取 Pod 指标
GET http://localhost:8090/api/v1/pods/{namespace}/{podName}/metrics

# 获取 Pod 日志
GET http://localhost:8090/api/v1/pods/{namespace}/{podName}/logs?container=xxx&tail=100
```

### 部署管理 API
```bash
# 删除应用部署
DELETE http://localhost:8087/api/v1/app-deployments/{id}

# 扩缩容
POST http://localhost:8087/api/v1/app-deployments/{id}/scale
{
  "replicas": 3,
  "user_id": 1
}
```

---

## 🐳 Docker 容器管理

### 查看容器状态
```bash
docker ps | grep my-cloud
```

### 重启服务
```bash
# 重启所有服务
docker-compose restart

# 重启单个服务
docker restart my-cloud-frontend
docker restart my-cloud-release-service
docker restart my-cloud-deploy-service
docker restart my-cloud-monitor-service
```

### 查看日志
```bash
docker logs my-cloud-frontend --tail 50
docker logs my-cloud-release-service --tail 50
docker logs my-cloud-deploy-service --tail 50
docker logs my-cloud-monitor-service --tail 50
```

---

## ☸️ Kubernetes 操作

### 查看资源
```bash
# 查看所有 Deployment
kubectl get deployment -A

# 查看 app-8 的资源
kubectl get all -n app-8

# 查看 Pod 详情
kubectl describe pod <pod-name> -n app-8

# 查看 Pod 日志
kubectl logs <pod-name> -n app-8 -f
```

### 金丝雀发布验证
```bash
# 查看 Deployment
kubectl get deployment -n app-8
# 应该看到：app-8 (stable) 和 app-8-canary

# 查看 Pod 分布
kubectl get pods -n app-8 -o wide
# 验证 Pod 数量比例

# 查看 Service
kubectl get svc -n app-8
# 验证流量分配
```

---

## 🎯 测试场景

### 场景1：金丝雀发布完整流程
1. 创建发布（10% 流量）
2. 验证：stable=1 Pod, canary=1 Pod
3. 确认金丝雀
4. 验证：stable=2 Pod（新版本），canary 已删除

### 场景2：监控中心使用
1. 访问 http://localhost/monitors
2. 选择"指标监控" Tab
3. 选择应用，查看 CPU/内存/QPS
4. 切换到"日志查询" Tab
5. 输入 Pod 名称，查询日志

### 场景3：删除部署记录
1. 访问 http://localhost/deployments
2. 找到要删除的记录
3. 点击"删除"按钮
4. 确认删除
5. 验证：数据库记录已删除，K8s 资源保留

---

## 📞 故障排查

### 前端无法访问
```bash
# 检查容器状态
docker ps | grep frontend

# 重启前端
docker restart my-cloud-frontend

# 查看日志
docker logs my-cloud-frontend
```

### 后端服务异常
```bash
# 检查服务健康
curl http://localhost:8086/health
curl http://localhost:8087/health
curl http://localhost:8090/health

# 查看服务日志
docker logs my-cloud-release-service --tail 100
```

### K8s 连接失败
```bash
# 检查 kubeconfig
kubectl cluster-info

# 检查服务账号权限
kubectl auth can-i get pods --all-namespaces
```

---

**更新时间**：2026-06-01  
**文档版本**：v1.0
