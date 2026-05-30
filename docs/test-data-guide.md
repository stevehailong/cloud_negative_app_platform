# 测试数据导入说明

## 概述

已为流水线管理和部署管理模块添加测试数据，方便前端页面展示和功能测试。

## 数据统计

### 流水线数据 (devops_db)

| 数据类型 | 数量 | 说明 |
|---------|------|------|
| 流水线 (pipelines) | 5条 | 包含不同类型的CI/CD流水线 |
| 流水线执行记录 (pipeline_runs) | 10条 | 包含成功、失败、运行中的执行记录 |
| 构建产物 (artifacts) | 8条 | Docker镜像、压缩包、报告等 |

### 部署数据 (deploy_db)

| 数据类型 | 数量 | 说明 |
|---------|------|------|
| 部署记录 (deployments) | 13条 | 跨dev/test/prod环境的部署 |
| - 运行中 (running) | 11条 | 正常运行的部署 |
| - 部署中 (deploying) | 1条 | 正在部署 |
| - 失败 (failed) | 1条 | 部署失败 |

## 流水线测试数据详情

### 1. 流水线列表

| 流水线名称 | 类型 | CI工具 | 状态 |
|-----------|------|--------|------|
| 前端构建流水线 | build | jenkins | 启用 |
| 后端服务流水线 | ci-cd | jenkins | 启用 |
| 全栈应用流水线 | ci-cd | jenkins | 启用 |
| 数据处理流水线 | build | gitlab-ci | 启用 |
| 移动端构建流水线 | build | jenkins | 禁用 |

### 2. 流水线执行记录示例

**前端构建流水线执行**:
- ✅ PIPE-FRONTEND-001-20260528-001 (成功, 5分钟)
- ✅ PIPE-FRONTEND-001-20260528-002 (成功, 4分钟)
- ❌ PIPE-FRONTEND-001-20260527-001 (失败, 2分钟)

**后端服务流水线执行**:
- ✅ PIPE-BACKEND-001-20260528-001 (成功, 8分钟)
- 🔄 PIPE-BACKEND-001-20260528-002 (运行中)
- ✅ PIPE-BACKEND-001-20260528-003 (成功, 7分钟)

### 3. 构建产物示例

| 产物名称 | 类型 | 版本 | 大小 |
|---------|------|------|------|
| frontend-app-v1.2.3.tar.gz | package | v1.2.3 | 15MB |
| backend-service:v2.1.0 | docker | v2.1.0 | 50MB |
| fullstack-app-v3.0.0.tar.gz | package | v3.0.0 | 100MB |
| data-processing-results.csv | report | v1.0.0 | 1MB |

## 部署测试数据详情

### 1. 开发环境 (dev)

| 应用 | 版本 | 副本数 | 状态 |
|------|------|--------|------|
| frontend-app | v1.2.3 | 2/2 | running |
| backend-service | v2.1.0 | 3/3 | running |
| fullstack-app | v3.0.0 | 2/2 | running |
| data-service | v1.5.0 | 1/2 | deploying |

### 2. 测试环境 (test)

| 应用 | 版本 | 副本数 | 状态 |
|------|------|--------|------|
| frontend-app | v1.2.3 | 2/2 | running |
| backend-service | v2.1.0 | 3/3 | running |
| fullstack-app | v3.0.0 | 2/2 | running |
| mobile-service | v0.9.0 | 0/2 | failed |
| database-service (StatefulSet) | v3.2.1 | 3/3 | running |

### 3. 生产环境 (prod)

| 应用 | 版本 | 副本数 | 状态 |
|------|------|--------|------|
| frontend-app | v1.2.2 | 4/4 | running |
| backend-service | v2.0.9 | 6/6 | running |
| fullstack-app | v2.9.8 | 4/4 | running |

## 数据特点

### 流水线数据特点

1. **多种状态**: 成功、失败、运行中
2. **时间分布**: 最近2小时到2天前
3. **不同触发方式**: manual、webhook、scheduled
4. **真实日志URL**: Jenkins和GitLab CI的控制台链接
5. **Git信息**: 包含commit hash和分支名

### 部署数据特点

1. **多环境**: dev、test、prod三个环境
2. **多工作负载类型**: Deployment、StatefulSet、DaemonSet
3. **不同副本数**: 2-6个副本
4. **版本迭代**: 体现版本升级关系
5. **异常场景**: 包含失败和部署中的状态

## 前端展示场景

### 流水线管理页面可展示

1. **流水线列表**: 5条流水线，含禁用状态
2. **执行历史**: 10条执行记录，各种状态
3. **成功率统计**: 可计算成功率 70% (7/10)
4. **平均耗时**: 可计算平均构建时间
5. **构建产物**: 8个不同类型的产物

### 部署管理页面可展示

1. **环境总览**: dev/test/prod三环境
2. **应用状态**: 11个运行中，1个部署中，1个失败
3. **版本信息**: 各环境不同版本号
4. **副本状态**: desired vs available
5. **工作负载类型**: Deployment/StatefulSet/DaemonSet

## 数据导入方法

### 方式1: 使用SQL脚本

```bash
# 导入流水线数据
docker exec -i my-cloud-mysql mysql -uroot -proot123456 < sql/test-data-pipeline.sql

# 导入部署数据
docker exec -i my-cloud-mysql mysql -uroot -proot123456 < sql/test-data-deployment.sql
```

### 方式2: 重新初始化（包含测试数据）

```bash
# 清空现有数据
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "USE devops_db; DELETE FROM artifacts; DELETE FROM pipeline_runs; DELETE FROM pipelines;"

# 导入新数据
docker exec -i my-cloud-mysql mysql -uroot -proot123456 < sql/test-data-pipeline.sql
```

## API测试建议

### 流水线API测试

```bash
# 获取流水线列表
curl http://localhost:8084/api/v1/pipelines \
  -H "Authorization: Bearer $TOKEN"

# 获取流水线执行记录
curl http://localhost:8084/api/v1/pipeline-runs?pipeline_id=1 \
  -H "Authorization: Bearer $TOKEN"

# 获取构建产物
curl http://localhost:8084/api/v1/artifacts?pipeline_run_id=1 \
  -H "Authorization: Bearer $TOKEN"
```

### 部署API测试

```bash
# 获取部署列表
curl http://localhost:8087/api/v1/deployments \
  -H "Authorization: Bearer $TOKEN"

# 按环境筛选
curl "http://localhost:8087/api/v1/deployments?namespace=prod" \
  -H "Authorization: Bearer $TOKEN"

# 获取部署详情
curl http://localhost:8087/api/v1/deployments/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 注意事项

1. **时间戳动态**: 所有时间都是相对NOW()计算的，数据始终保持"新鲜"
2. **关联关系**: pipeline_id和release_id需要对应实际数据
3. **唯一约束**: pipeline_code和run_no必须唯一
4. **外键依赖**: 确保application和cluster数据存在

## 数据维护

### 定期清理

```sql
-- 清理30天前的流水线执行记录
DELETE FROM pipeline_runs WHERE create_time < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 清理失败的部署记录
DELETE FROM deployments WHERE deployment_status = 'failed' AND create_time < DATE_SUB(NOW(), INTERVAL 7 DAY);
```

### 数据备份

```bash
# 备份devops_db
docker exec my-cloud-mysql mysqldump -uroot -proot123456 devops_db > backup_devops_$(date +%Y%m%d).sql

# 备份deploy_db
docker exec my-cloud-mysql mysqldump -uroot -proot123456 deploy_db > backup_deploy_$(date +%Y%m%d).sql
```

## 后续扩展

可以根据实际需要添加更多测试数据：

1. **更多流水线类型**: 测试、安全扫描、性能测试流水线
2. **更复杂的配置**: 多阶段、并行任务、条件执行
3. **更多部署策略**: 金丝雀、A/B测试
4. **回滚记录**: 版本回滚历史
5. **审批流程**: 部署审批记录

---

**创建时间**: 2026-05-28  
**数据有效期**: 永久（相对时间自动更新）  
**维护者**: DevOps团队
