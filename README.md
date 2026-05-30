# 云原生应用研发交付平台

> 基于 Go 微服务（Gin + GORM）+ Vue.js 前端的企业级云原生应用研发交付平台

## 项目概述

一站式开发、测试、发布、运维平台，实现研发交付闭环：
**代码 → 构建 → 测试 → 制品 → 发布 → 部署 → 观测 → 回滚 → 治理**

## 技术栈

### 后端
- **语言**: Go 1.22+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **缓存**: Redis
- **消息队列**: Kafka/RabbitMQ

### 前端
- **框架**: Vue 3
- **UI库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **构建工具**: Vite
- **HTTP客户端**: Axios

### DevOps工具链
- **代码仓库**: GitLab
- **CI**: Jenkins/Tekton
- **镜像仓库**: Harbor
- **CD**: Argo CD
- **编排平台**: Kubernetes
- **配置中心**: Nacos
- **密钥管理**: Vault
- **监控**: Prometheus + Grafana
- **日志**: Loki
- **链路追踪**: Jaeger
- **服务网格**: Istio

## 项目结构

```
.
├── backend/                    # 后端服务
│   ├── cmd/                   # 各微服务入口
│   ├── internal/              # 内部代码
│   ├── api/                   # API定义
│   ├── pkg/                   # 公共包
│   └── go.mod
├── frontend/                  # 前端项目
│   ├── src/
│   ├── public/
│   └── package.json
├── deploy/                    # 部署配置
│   ├── helm/
│   ├── k8s/
│   └── docker/
├── sql/                       # 数据库脚本
└── docs/                      # 文档
```

## 微服务列表

### Phase I: 核心交付能力 ✅ (已完成)

| 服务名 | 端口 | 数据库 | 职责 | 状态 |
|--------|------|--------|------|------|
| gateway | 8080 | iam_db | API网关/BFF | ✅ |
| auth-service | 8081 | iam_db | 认证授权(用户/角色/权限) | ✅ |
| project-service | 8082 | org_db | 项目组织管理 | ✅ |
| application-service | 8083 | app_db | 应用管理 | ✅ |
| pipeline-service | 8084 | devops_db | 流水线管理 | ✅ |
| env-service | 8085 | env_db | 环境管理 | ✅ |
| release-service | 8086 | release_db | 发布管理 | ✅ |
| deploy-service | 8087 | deploy_db | 部署管理 | ✅ |
| cluster-service | 8088 | infra_db | 集群管理 | ✅ |

### Phase II: 治理增强能力 ⏳ (进行中 2/6)

| 服务名 | 端口 | 数据库 | 职责 | 状态 | 优先级 |
|--------|------|--------|------|------|--------|
| notification-service | 8095 | notification_db | 多渠道通知服务 | ✅ **已完成** | HIGH |
| audit-service | 8093 | audit_db | 审计日志服务 | ✅ **已完成** | HIGH |
| monitor-service | 8090 | monitor_db | 监控告警 | 📋 待实现 | MEDIUM |
| config-service | 8091 | config_db | 配置中心(Nacos) | 📋 待实现 | MEDIUM |
| secret-service | 8092 | secret_db | 密钥管理(Vault) | 📋 待实现 | MEDIUM |
| cost-service | 8096 | cost_db | 成本治理 | 📋 待实现 | LOW |

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- (可选) Go 1.22+ 用于本地开发
- (可选) Node.js 18+ 用于前端开发

### 一键启动（推荐）

```bash
# 克隆项目
git clone https://github.com/yourusername/my-cloud.git
cd my-cloud

# 启动所有服务（包括MySQL、Redis和所有微服务）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f notification-service
```

服务启动后访问：
- **网关**: http://localhost:8080
- **前端**: http://localhost:80
- **各微服务健康检查**: http://localhost:809X/health

### 数据库初始化

数据库会在容器启动时自动初始化。如需手动初始化：

```bash
# 初始化所有数据库
docker exec -i my-cloud-mysql mysql -uroot -proot123456 < sql/01-iam-db.sql
docker exec -i my-cloud-mysql mysql -uroot -proot123456 < sql/02-org-db.sql
# ... 其他数据库脚本
```

### 测试服务

```bash
# 测试 Notification Service
./scripts/test-notification-service.sh

# 测试 Audit Service
./scripts/test-audit-service.sh

# 手动测试健康检查
curl http://localhost:8095/health  # Notification
curl http://localhost:8093/health  # Audit
```

## 最新更新 🎉

### 2026-05-28: Notification Service 上线

新增多渠道通知服务，提供完整的通知管理能力：

**核心特性**:
- ✅ 多渠道支持: Email、SMS、DingTalk、Slack、Webhook
- ✅ 模板管理: 支持{{变量}}替换的灵活模板系统
- ✅ 异步发送: 不阻塞业务流程的异步通知
- ✅ 状态跟踪: pending/sent/failed状态管理
- ✅ 预置模板: 6个常用通知模板(发布/部署/流水线通知)
- ✅ 预置渠道: 4个预配置渠道

**使用文档**: [docs/notification-service.md](docs/notification-service.md)

**API端点**:
```bash
# 直接发送通知
POST /api/v1/notifications

# 通过模板发送
POST /api/v1/notifications/template

# 管理模板和渠道
GET/POST/PUT/DELETE /api/v1/notification-templates
GET/POST/PUT/DELETE /api/v1/notification-channels
```

**快速测试**:
```bash
./scripts/test-notification-service.sh
```

## API 文档

- **API基础路径**: `/api/v1`
- **网关地址**: http://localhost:8080
- **认证方式**: Bearer Token (JWT)

### 服务文档
- [Notification Service API文档](docs/notification-service.md)
- [项目整体进度](docs/implementation-progress.md)
- [设计文档](docs/design.md)

## 开发指南

### 后端开发

1. 每个微服务遵循相同的项目结构
2. 使用统一的响应格式
3. 实现统一的错误处理
4. 添加 OpenTelemetry 追踪
5. 编写单元测试

### 前端开发

1. 使用 Composition API
2. 遵循 Vue 3 最佳实践
3. 组件化开发
4. 统一的状态管理
5. 响应式设计

## 部署

### Docker 部署

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d
```

### Kubernetes 部署

```bash
# 使用 Helm
helm install my-cloud ./deploy/helm/my-cloud

# 或使用 kubectl
kubectl apply -f deploy/k8s/
```

## 监控与可观测

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
- **Jaeger**: http://localhost:16686
- **Loki**: 日志聚合

## 配置管理

配置文件位置：
- 后端: `backend/configs/`
- 前端: `frontend/.env.*`

环境变量：
- `DATABASE_URL`: 数据库连接
- `REDIS_URL`: Redis连接
- `JWT_SECRET`: JWT密钥
- `KEYCLOAK_URL`: Keycloak地址

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 许可证

MIT License

## 联系方式

- 项目地址: https://github.com/yourusername/my-cloud
- 问题反馈: https://github.com/yourusername/my-cloud/issues
