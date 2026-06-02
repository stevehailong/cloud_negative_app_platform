# 重复部署记录问题修复指南

## 问题描述

**现象**: 应用管理中 app-8 命名空间出现了两个相同的部署记录

**根本原因**: 
1. 数据库唯一索引配置错误
2. 之前的索引只在 `workload_name` 上,没有正确约束 `(namespace, workload_name)` 组合
3. 导致可以创建重复的记录

## 修复内容

### 1. 修复数据库模型

**文件**: `backend/internal/deploy/model/app_deployment.go`

**修改前**:
```go
Namespace    string `gorm:"column:namespace;size:255;not null;index:idx_namespace_workload"`
WorkloadName string `gorm:"column:workload_name;size:255;not null;uniqueIndex:idx_namespace_workload;index:idx_app_env_workload"`
```

**修改后**:
```go
Namespace    string `gorm:"column:namespace;size:255;not null;uniqueIndex:uk_namespace_workload"`
WorkloadName string `gorm:"column:workload_name;size:255;not null;uniqueIndex:uk_namespace_workload"`
```

**说明**: 
- 使用 `uniqueIndex:uk_namespace_workload` 在两个字段上,创建组合唯一索引
- 确保 `(namespace, workload_name)` 组合唯一
- 同一 namespace 可以有多个 workload (stable + canary)

### 2. 清理工具

创建了两种清理方式:

#### 方式1: SQL 脚本
**文件**: `backend/scripts/cleanup_duplicate_deployments.sql`

```bash
# 连接数据库
mysql -h mysql -u root -proot123456 deploy_db

# 执行清理脚本
source backend/scripts/cleanup_duplicate_deployments.sql
```

#### 方式2: Go 程序
**文件**: `backend/cmd/cleanup-duplicates/main.go`

```bash
# 编译
cd backend
go build -o /tmp/cleanup-duplicates ./cmd/cleanup-duplicates/

# 运行 (需要 MySQL 连接)
/tmp/cleanup-duplicates
```

**功能**:
- 自动查找重复记录
- 保留最新的记录 (按 update_time 排序)
- 删除旧的重复记录
- 验证清理结果
- 显示最终数据

## 修复步骤

### 步骤1: 备份数据库
```bash
mysqldump -h mysql -u root -proot123456 deploy_db > deploy_db_backup.sql
```

### 步骤2: 停止服务
```bash
# 停止所有相关服务
pkill -f deploy-service
pkill -f release-service
pkill -f ci-service
```

### 步骤3: 清理重复记录

**选项A: 使用 Go 程序 (推荐)**
```bash
cd /Users/hanhailong01/Downloads/my_cloud/backend
go run ./cmd/cleanup-duplicates/main.go
```

**选项B: 使用 SQL 脚本**
```bash
mysql -h mysql -u root -proot123456 deploy_db < backend/scripts/cleanup_duplicate_deployments.sql
```

### 步骤4: 重新编译服务
```bash
cd backend
go build -o /tmp/deploy-service ./cmd/deploy-service/
```

### 步骤5: 启动服务
```bash
cd backend
go run ./cmd/deploy-service/main.go &
go run ./cmd/release-service/main.go &
go run ./cmd/ci-service/main.go &
```

### 步骤6: 验证修复

#### 6.1 查询部署记录
```bash
curl -s "http://localhost:8087/internal/v1/app-deployments/by-workload?namespace=app-8&workload_name=app-8" | jq '.'
curl -s "http://localhost:8087/internal/v1/app-deployments/by-workload?namespace=app-8&workload_name=app-8-canary" | jq '.'
```

**预期结果**: 每个查询只返回一条记录

#### 6.2 尝试创建重复记录
```bash
# 第一次创建 - 应该成功
curl -X POST "http://localhost:8087/internal/v1/app-deployments" \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 8,
    "env_id": 1,
    "cluster_id": 1,
    "namespace": "app-8",
    "workload_name": "app-8-test",
    "workload_type": "deployment",
    "desired_replicas": 1
  }' | jq '.'

# 第二次创建相同的 - 应该失败
curl -X POST "http://localhost:8087/internal/v1/app-deployments" \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 8,
    "env_id": 1,
    "cluster_id": 1,
    "namespace": "app-8",
    "workload_name": "app-8-test",
    "workload_type": "deployment",
    "desired_replicas": 1
  }' | jq '.'
```

**预期结果**: 第二次应该返回错误 "Duplicate entry"

#### 6.3 测试部署功能
```bash
# 获取部署ID
STABLE_ID=$(curl -s "http://localhost:8087/internal/v1/app-deployments/by-workload?namespace=app-8&workload_name=app-8" | jq -r '.data.id')

# 测试部署
curl -X POST "http://localhost:8087/internal/v1/app-deployments/$STABLE_ID/deploy" \
  -H "Content-Type: application/json" \
  -d '{
    "version": "v1.0.0",
    "image_url": "nginx:1.21",
    "replicas": 3
  }' | jq '.'

# 等待30秒后查询状态
sleep 30
HISTORY_ID=$(curl -s "http://localhost:8087/internal/v1/app-deployments/$STABLE_ID/deploy" | jq -r '.data.history_id')
curl -s "http://localhost:8087/internal/v1/deployment-history/$HISTORY_ID" | jq '.data | {status, failure_reason}'
```

**预期结果**: 
- 如果有 K8s 环境: `status: "success"`
- 如果没有 K8s: `status: "failed"`, `failure_reason: "K8s client not available"`

## 验证清单

- [ ] 数据库中没有重复的 `(namespace, workload_name)` 记录
- [ ] 唯一索引 `uk_namespace_workload` 已创建
- [ ] 无法创建重复的部署记录
- [ ] 部署功能正常 (创建新 Deployment 或更新现有 Deployment)
- [ ] Pod 查询使用 `version` label 正确区分 stable 和 canary

## 常见问题

### Q1: 清理后仍有重复记录?
**A**: 检查是否有多个服务实例同时运行,停止所有实例后重新清理

### Q2: 部署仍然失败?
**A**: 检查失败原因:
```bash
curl -s "http://localhost:8087/internal/v1/deployment-history/$HISTORY_ID" | jq '.data.failure_reason'
```

常见原因:
- `K8s client not available`: 没有 K8s 环境
- `failed to ensure namespace`: K8s 权限不足
- `deployment rollout timed out`: 镜像拉取失败或资源不足

### Q3: 如何回滚?
**A**: 使用备份恢复:
```bash
mysql -h mysql -u root -proot123456 deploy_db < deploy_db_backup.sql
```

## 测试环境要求

### 最小环境 (仅测试数据库逻辑)
- MySQL 数据库
- Go 1.26+

### 完整环境 (测试部署功能)
- MySQL 数据库
- Kubernetes 集群
- kubeconfig 配置
- Go 1.26+

## 总结

### 修复内容
1. ✅ 修复数据库唯一索引配置
2. ✅ 创建清理工具 (SQL + Go)
3. ✅ 修复部署创建逻辑 (创建或更新)
4. ✅ 添加部署就绪等待机制

### 预期效果
- ✅ 每个应用在每个环境只有一个 namespace
- ✅ 同一 namespace 可以有 stable 和 canary 两个 workload
- ✅ 无法创建重复的部署记录
- ✅ 部署功能正常工作 (创建新 Deployment 或更新现有)
- ✅ Pod 查询正确区分版本

### 下一步
1. 在测试环境执行清理
2. 验证唯一索引生效
3. 测试部署功能
4. 在生产环境执行 (需要维护窗口)
