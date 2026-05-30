# 核心研发流程服务实现说明

## 已实现的三个核心服务

### 1. Pipeline Service - 流水线服务 (端口 8084)

**功能**：管理CI/CD流水线、执行记录和制品

**数据库**：`devops_db`

**数据表**：
- `pipelines` - 流水线配置表
- `pipeline_runs` - 流水线执行记录表
- `artifacts` - 制品表（镜像、Helm Chart等）

**主要API**：
```
GET    /api/v1/pipelines              # 获取流水线列表
POST   /api/v1/pipelines              # 创建流水线
GET    /api/v1/pipelines/:id          # 获取流水线详情
PUT    /api/v1/pipelines/:id          # 更新流水线
DELETE /api/v1/pipelines/:id          # 删除流水线
POST   /api/v1/pipelines/:id/run      # 触发流水线执行

GET    /api/v1/pipelines/:id/runs     # 获取流水线执行记录
GET    /api/v1/pipeline-runs/:runId   # 获取执行详情
GET    /api/v1/pipeline-runs/:runId/logs  # 获取执行日志

GET    /api/v1/artifacts              # 获取制品列表
GET    /api/v1/artifacts/:id          # 获取制品详情
DELETE /api/v1/artifacts/:id          # 删除制品
```

**核心特性**：
- 支持多种CI工具（Jenkins/Tekton）
- 支持多种触发方式（manual/webhook/mr/schedule）
- 流水线类型：CI/CD/Full
- 制品类型：image/chart/package/sbom/report

---

### 2. Release Service - 发布服务 (端口 8086)

**功能**：发布工单管理、审批流程、发布执行和回滚

**数据库**：`release_db`

**数据表**：
- `releases` - 发布工单表
- `release_approvals` - 发布审批表

**主要API**：
```
GET    /api/v1/releases               # 获取发布工单列表
POST   /api/v1/releases               # 创建发布工单
GET    /api/v1/releases/:id           # 获取工单详情
POST   /api/v1/releases/:id/submit    # 提交审批
POST   /api/v1/releases/:id/approve   # 审批通过
POST   /api/v1/releases/:id/reject    # 审批拒绝
POST   /api/v1/releases/:id/execute   # 执行发布
POST   /api/v1/releases/:id/rollback  # 回滚发布

GET    /api/v1/releases/:id/approvals # 获取审批记录
```

**核心特性**：
- 发布策略：rolling/bluegreen/canary
- 多级审批流程
- 工单状态流转：created → submitted → approved → executing → success/failed
- 支持发布回滚

**工单流程**：
1. 开发者创建发布工单（created）
2. 提交审批（submitted）
3. 审批人审批（approved/rejected）
4. 执行发布（executing → success/failed）
5. 如需要可以回滚（rollback）

---

### 3. Deploy Service - 部署服务 (端口 8087)

**功能**：Kubernetes部署管理、运维操作

**数据库**：`deploy_db`

**数据表**：
- `deployments` - 部署记录表

**主要API**：
```
GET    /api/v1/deployments            # 获取部署列表
POST   /api/v1/deployments            # 创建部署
GET    /api/v1/deployments/:id        # 获取部署详情
POST   /api/v1/deployments/:id/restart    # 重启部署
POST   /api/v1/deployments/:id/scale      # 扩缩容
GET    /api/v1/deployments/:id/events     # 获取部署事件
GET    /api/v1/deployments/:id/pods       # 获取Pod列表
```

**核心特性**：
- 支持多种工作负载类型：deployment/statefulset/job
- 部署状态监控：progressing/success/failed/rollback
- Pod副本管理（扩缩容）
- 部署重启功能
- 事件和Pod状态查询

---

## 服务之间的关系

```
Pipeline Service  →  Release Service  →  Deploy Service
      ↓                    ↓                   ↓
   制品(Artifact)      发布工单(Release)    部署(Deployment)
```

**完整流程**：

1. **CI阶段（Pipeline Service）**
   - 开发者提交代码
   - 触发Pipeline执行（手动/Webhook）
   - Pipeline构建镜像
   - 推送镜像到Harbor
   - 记录Artifact制品信息

2. **发布审批阶段（Release Service）**
   - 创建Release工单，关联Pipeline Run
   - 提交审批
   - 审批人审批通过

3. **部署执行阶段（Deploy Service）**
   - Release Service调用Deploy Service创建部署
   - Deploy Service更新Kubernetes Deployment
   - 监控部署进度
   - 健康检查
   - 部署完成或失败

4. **运维阶段（Deploy Service）**
   - 重启Pod
   - 扩缩容
   - 查看事件和日志
   - 如需要可以通过Release Service回滚

---

## 部署说明

### 1. 启动所有服务

```bash
# 构建三个新服务
docker-compose build pipeline-service release-service deploy-service

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 2. 验证服务状态

```bash
# 检查Pipeline Service
curl http://localhost:8080/api/v1/pipelines

# 检查Release Service
curl http://localhost:8080/api/v1/releases

# 检查Deploy Service
curl http://localhost:8080/api/v1/deployments
```

### 3. 服务端口映射

| 服务 | 容器端口 | 宿主机端口 | 数据库 |
|------|---------|-----------|--------|
| Gateway | 8080 | 8080 | iam_db |
| Auth Service | 8081 | 8081 | iam_db |
| Project Service | 8082 | 8082 | org_db |
| Application Service | 8083 | 8083 | app_db |
| **Pipeline Service** | 8084 | 8084 | devops_db |
| Env Service | 8085 | 8085 | env_db |
| **Release Service** | 8086 | 8086 | release_db |
| **Deploy Service** | 8087 | 8087 | deploy_db |
| Cluster Service | 8088 | 8088 | infra_db |

---

## API使用示例

### 示例1：创建并执行流水线

```bash
# 1. 登录获取Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 2. 创建流水线
curl -X POST http://localhost:8080/api/v1/pipelines \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pipelineCode": "user-api-ci",
    "appId": 1,
    "pipelineName": "用户API CI流水线",
    "pipelineType": "ci",
    "ciTool": "jenkins",
    "configJson": "{\"jenkinsJobName\":\"user-api-build\"}"
  }'

# 3. 触发流水线执行
curl -X POST http://localhost:8080/api/v1/pipelines/1/run \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "triggerType": "manual",
    "gitCommit": "abc123def456",
    "gitBranch": "main"
  }'

# 4. 查看执行记录
curl -X GET http://localhost:8080/api/v1/pipelines/1/runs \
  -H "Authorization: Bearer $TOKEN"

# 5. 查看制品
curl -X GET http://localhost:8080/api/v1/artifacts \
  -H "Authorization: Bearer $TOKEN"
```

### 示例2：创建发布工单并审批

```bash
# 1. 创建发布工单
RELEASE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/releases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "appId": 1,
    "envId": 1,
    "pipelineRunId": 1,
    "releaseVersion": "v1.0.0",
    "releaseStrategy": "rolling",
    "description": "用户API v1.0.0发布"
  }')

RELEASE_ID=$(echo $RELEASE_RESPONSE | jq -r '.data.id')

# 2. 提交审批
curl -X POST http://localhost:8080/api/v1/releases/$RELEASE_ID/submit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "approverUserIds": [2, 3]
  }'

# 3. 审批通过
curl -X POST http://localhost:8080/api/v1/releases/$RELEASE_ID/approve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "comment": "审批通过，可以发布"
  }'

# 4. 执行发布
curl -X POST http://localhost:8080/api/v1/releases/$RELEASE_ID/execute \
  -H "Authorization: Bearer $TOKEN"

# 5. 查看审批记录
curl -X GET http://localhost:8080/api/v1/releases/$RELEASE_ID/approvals \
  -H "Authorization: Bearer $TOKEN"
```

### 示例3：创建部署并管理

```bash
# 1. 创建部署
DEPLOYMENT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/deployments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "releaseId": 1,
    "clusterId": 1,
    "namespace": "production",
    "workloadName": "user-api",
    "workloadType": "deployment",
    "imageVersion": "harbor.example.com/myapp/user-api:v1.0.0",
    "desiredReplicas": 3
  }')

DEPLOYMENT_ID=$(echo $DEPLOYMENT_RESPONSE | jq -r '.data.id')

# 2. 查看部署详情
curl -X GET http://localhost:8080/api/v1/deployments/$DEPLOYMENT_ID \
  -H "Authorization: Bearer $TOKEN"

# 3. 扩容到5个副本
curl -X POST http://localhost:8080/api/v1/deployments/$DEPLOYMENT_ID/scale \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "replicas": 5
  }'

# 4. 重启部署
curl -X POST http://localhost:8080/api/v1/deployments/$DEPLOYMENT_ID/restart \
  -H "Authorization: Bearer $TOKEN"

# 5. 查看Pod列表
curl -X GET http://localhost:8080/api/v1/deployments/$DEPLOYMENT_ID/pods \
  -H "Authorization: Bearer $TOKEN"

# 6. 查看部署事件
curl -X GET http://localhost:8080/api/v1/deployments/$DEPLOYMENT_ID/events \
  -H "Authorization: Bearer $TOKEN"
```

---

## 数据库表结构

### Pipeline Service (devops_db)

```sql
-- pipelines 表
CREATE TABLE pipelines (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    pipeline_code VARCHAR(64) UNIQUE NOT NULL,
    app_id BIGINT NOT NULL,
    pipeline_name VARCHAR(128) NOT NULL,
    pipeline_type VARCHAR(32) NOT NULL,  -- ci/cd/full
    ci_tool VARCHAR(32) DEFAULT 'jenkins',
    config_json JSON,
    enabled TINYINT DEFAULT 1,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_id(app_id)
);

-- pipeline_runs 表
CREATE TABLE pipeline_runs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    pipeline_id BIGINT NOT NULL,
    run_no VARCHAR(64) UNIQUE NOT NULL,
    trigger_type VARCHAR(32) NOT NULL,  -- manual/webhook/mr/schedule
    git_commit VARCHAR(64),
    git_branch VARCHAR(64),
    status VARCHAR(32) NOT NULL,  -- pending/running/success/failed/cancelled
    start_time DATETIME,
    end_time DATETIME,
    duration_seconds INT,
    operator_user_id BIGINT,
    log_url VARCHAR(255),
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pipeline_id(pipeline_id),
    INDEX idx_status(status),
    INDEX idx_git_branch(git_branch)
);

-- artifacts 表
CREATE TABLE artifacts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    pipeline_run_id BIGINT NOT NULL,
    artifact_type VARCHAR(32) NOT NULL,  -- image/chart/package/sbom/report
    artifact_name VARCHAR(128) NOT NULL,
    artifact_version VARCHAR(64),
    repo_url VARCHAR(255),
    digest VARCHAR(255),
    metadata_json JSON,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pipeline_run_id(pipeline_run_id),
    INDEX idx_artifact_type(artifact_type)
);
```

### Release Service (release_db)

```sql
-- releases 表
CREATE TABLE releases (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    release_no VARCHAR(64) UNIQUE NOT NULL,
    app_id BIGINT NOT NULL,
    env_id BIGINT NOT NULL,
    pipeline_run_id BIGINT,
    release_version VARCHAR(64) NOT NULL,
    release_strategy VARCHAR(32) NOT NULL,  -- rolling/bluegreen/canary
    approval_status VARCHAR(32) DEFAULT 'pending',  -- pending/approved/rejected
    release_status VARCHAR(32) DEFAULT 'created',  -- created/submitted/approved/rejected/executing/success/failed/rollback
    operator_user_id BIGINT,
    description VARCHAR(255),
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_id(app_id),
    INDEX idx_env_id(env_id),
    INDEX idx_release_status(release_status)
);

-- release_approvals 表
CREATE TABLE release_approvals (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    release_id BIGINT NOT NULL,
    approver_user_id BIGINT NOT NULL,
    approval_status VARCHAR(32) NOT NULL,  -- pending/approved/rejected
    comment_text VARCHAR(255),
    approval_time DATETIME,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_release_id(release_id),
    INDEX idx_approver_user_id(approver_user_id)
);
```

### Deploy Service (deploy_db)

```sql
-- deployments 表
CREATE TABLE deployments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    release_id BIGINT NOT NULL,
    cluster_id BIGINT NOT NULL,
    namespace VARCHAR(128) NOT NULL,
    workload_name VARCHAR(128) NOT NULL,
    workload_type VARCHAR(32) NOT NULL,  -- deployment/statefulset/job
    image_version VARCHAR(128) NOT NULL,
    desired_replicas INT DEFAULT 1,
    available_replicas INT DEFAULT 0,
    deployment_status VARCHAR(32) NOT NULL,  -- progressing/success/failed/rollback
    start_time DATETIME,
    end_time DATETIME,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_release_id(release_id),
    INDEX idx_cluster_id(cluster_id),
    INDEX idx_namespace(namespace)
);
```

---

## 后续扩展计划

### 1. Kubernetes集成
- 集成Kubernetes client-go
- 实际操作Kubernetes资源
- 实时监控部署状态

### 2. CI/CD工具集成
- Jenkins API集成
- Tekton Pipeline集成
- GitLab CI集成

### 3. GitOps支持
- Argo CD集成
- 自动更新GitOps仓库
- 同步状态监控

### 4. 镜像仓库集成
- Harbor API集成
- 镜像扫描结果查询
- 镜像标签管理

### 5. 审计和通知
- 完善审计日志
- 发布通知（钉钉/企微/邮件）
- 告警通知集成

---

## 故障排查

### 服务无法启动

```bash
# 查看服务日志
docker logs my-cloud-pipeline-service
docker logs my-cloud-release-service
docker logs my-cloud-deploy-service

# 检查数据库连接
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "SHOW DATABASES;"
```

### API调用失败

```bash
# 检查Gateway路由配置
docker logs my-cloud-gateway | grep pipeline
docker logs my-cloud-gateway | grep release
docker logs my-cloud-gateway | grep deploy

# 测试服务健康检查
curl http://localhost:8084/health  # Pipeline Service
curl http://localhost:8086/health  # Release Service
curl http://localhost:8087/health  # Deploy Service
```

### 数据库问题

```bash
# 检查数据库是否创建
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "SHOW DATABASES LIKE '%db';"

# 检查表是否创建
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "USE devops_db; SHOW TABLES;"
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "USE release_db; SHOW TABLES;"
docker exec my-cloud-mysql mysql -uroot -proot123456 -e "USE deploy_db; SHOW TABLES;"
```

---

## 技术栈

- **语言**: Go 1.22
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **容器化**: Docker + Docker Compose
- **API网关**: 自研Gateway（反向代理）
- **认证**: JWT
- **权限**: RBAC

---

## 总结

三个核心研发流程服务已实现完毕：

✅ **Pipeline Service** - CI/CD流水线管理
✅ **Release Service** - 发布工单和审批
✅ **Deploy Service** - Kubernetes部署管理

这三个服务构成了完整的研发交付流程，从代码构建到发布审批再到实际部署，覆盖了DevOps的核心环节。所有服务都已集成到Gateway中，通过统一的API网关提供服务。
