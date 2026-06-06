# 云原生应用研发交付平台

> 基于 Go 微服务（Gin + GORM）+ Vue 3 前端的企业级云原生应用研发交付平台
>
> 实现研发交付闭环：**代码 → 构建 → 制品 → 发布 → 部署 → 金丝雀灰度 → 观测 → 回滚 → 成本治理**

## 技术栈

| 分类 | 技术 |
|------|------|
| 后端 | Go + Gin + GORM |
| 前端 | Vue 3 + Element Plus + Pinia + Vite |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis |
| 容器编排 | Kubernetes (Docker Desktop) |
| CI | Jenkins |
| 镜像仓库 | Local Registry (Docker) / Harbor (可选) |
| 监控 | Prometheus + Grafana |
| 日志 | Loki |
| 链路追踪 | OpenTelemetry + Jaeger |
| 金丝雀/灰度 | Nginx Ingress Controller canary annotations |
| 成本治理 | Prometheus metrics + Kubecost (可插拔) |

## 项目结构

```
.
├── backend/
│   ├── cmd/                      # 各微服务入口 (17个服务)
│   ├── internal/                 # 业务逻辑
│   │   ├── common/               # 公共模型/中间件/响应
│   │   ├── application/          # 应用管理
│   │   ├── auth/                 # 认证授权
│   │   ├── audit/                # 审计
│   │   ├── cluster/              # 集群
│   │   ├── config/               # 配置
│   │   ├── cost/                 # 成本
│   │   ├── deploy/               # 部署
│   │   ├── environment/          # 环境
│   │   ├── gateway/              # 网关
│   │   ├── monitor/              # 监控
│   │   ├── notification/         # 通知
│   │   ├── pipeline/             # 流水线
│   │   ├── release/              # 发布
│   │   ├── resource/             # 资源
│   │   └── secret/               # 密钥
│   ├── pkg/                      # 公共包
│   │   ├── database/             # 数据库初始化
│   │   ├── gitlab/               # GitLab 客户端
│   │   ├── helm/                 # Helm 客户端 + Values Builder
│   │   ├── jenkins/              # Jenkins 客户端
│   │   ├── jwt/                  # JWT 工具
│   │   ├── k8s/                  # K8s 客户端
│   │   ├── kubecost/             # Kubecost 客户端
│   │   ├── metrics/              # Prometheus 指标
│   │   ├── prometheus/           # Prometheus 查询
│   │   └── security/             # 安全/登录限流
│   ├── configs/                  # 配置文件
│   ├── scripts/                  # SQL 脚本
│   └── go.mod
├── frontend/                     # Vue 3 前端
│   └── src/
│       ├── views/                # 页面 (20个模块)
│       ├── api/                  # API 封装
│       ├── router/               # 路由
│       ├── utils/                # 工具函数
│       └── layouts/              # 布局
├── helm-charts/                  # Helm Chart 模板
├── k8s-manifests/                # K8s 资源清单
├── jenkins/                      # Jenkins 配置
├── prometheus/                   # Prometheus 配置
├── docker-compose.yml            # Docker Compose 编排
├── docker-compose-harbor.yml     # Harbor (可选)
└── README.md
```

## 微服务清单

| 服务 | 端口 | 数据库 | 职责 |
|------|------|--------|------|
| gateway | 8080 | — | API网关/BFF，认证鉴权，服务代理 |
| auth-service | 8081 | iam_db | 用户、角色、权限、JWT 认证 |
| project-service | 8082 | org_db | 租户、组织、项目、成员管理 |
| application-service | 8083 | app_db | 应用、组件、依赖管理 |
| pipeline-service | 8084 | devops_db | 流水线、构建运行、制品 |
| env-service | 8085 | env_db | 环境、环境模板、应用环境绑定 |
| release-service | 8086 | release_db | 发布单、审批、金丝雀、回滚 |
| deploy-service | 8087 | deploy_db | K8s 部署、扩缩容、重启、回滚 |
| cluster-service | 8088 | infra_db | 集群、命名空间管理 |
| monitor-service | 8090 | monitor_db | 监控、日志、链路聚合 |
| config-service | 8097 | config_db | 应用配置管理 |
| secret-service | 8098 | secret_db | 密钥元数据 |
| audit-service | 8093 | audit_db | 审计日志 |
| notification-service | 8095 | notification_db | 多渠道通知 |
| resource-service | 8096 | resource_db | 资源配额 |
| cost-service | 8099 | cost_db | 成本统计（Prometheus + Kubecost） |

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- Kubernetes 集群 (Docker Desktop 内置)

### 一键启动

```bash
# 克隆项目
git clone git@github.com:stevehailong/cloud_negative_app_platform.git
cd cloud_negative_app_platform

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看特定服务日志
docker-compose logs -f deploy-service
```

### 数据库初始化

数据库表由 GORM AutoMigrate 自动创建，无需手动执行 SQL。

如需手动初始化：

```bash
docker exec -i my-cloud-mysql mysql -uroot -proot123456 --default-character-set=utf8mb4 < sql/init.sql
```

### 访问地址

| 组件 | 地址 |
|------|------|
| 前端 Portal | http://localhost |
| API 网关 | http://localhost:8080 |
| Prometheus | http://localhost:9091 |
| Grafana | http://localhost:3000 (admin/admin123456) |
| Jenkins | http://localhost:9090 (admin/admin123) |
| ChartMuseum | http://localhost:8092 |
| Local Registry | localhost:5001 |

## 核心功能

### 应用管理
- 应用 CRUD，组件与依赖管理
- 多环境绑定（dev/test/staging/prod）
- Helm Chart 环境模板
- 应用-环境配置覆盖（envVars、端口、健康检查、Ingress）

### 流水线 (CI)
- Jenkins Pipeline 集成
- 流水线创建/运行/日志
- 制品管理
- 支持 GitLab Webhook 触发

### 发布与部署
- **三种发布策略**：滚动发布、金丝雀发布、蓝绿发布
- **金丝雀灰度分流**：基于 Nginx Ingress canary annotations
  - 权重分流（0%-100% 滑块实时调整）
  - Header/Cookie 路由
  - **自动 Pod 扩缩容**：调整权重时自动同步副本数
  - **100% 全量部署**：拖动到 100% 弹窗确认 → 自动提升 canary 为 stable → 清理金丝雀资源
- 并发部署拦截（同一应用同一时刻只允许一个部署进行中）
- 部署历史与操作人追踪
- K8s Pod 管理（日志、事件、重启、扩缩容）

### 成本治理
- 基于 Prometheus 指标的日成本估算
- CPU / 内存 / 存储 / 网络 四维成本拆分
- 按项目、应用聚合汇总
- 应用名称/项目名称自动关联
- 支持 Kubecost 接入（客户端已就绪，安装后自动切换）

### 可观测
- Prometheus + Grafana 指标监控
- OpenTelemetry 链路追踪
- 容器日志聚合
- 审计日志全量记录

### 权限与安全
- JWT Bearer Token 认证
- RBAC 角色权限控制
- API 网关统一鉴权
- 操作审计全量记录

## API 规范

- **基础路径**: `/api/v1`
- **内部调用**: `/internal/v1`（无需认证）
- **响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "requestId": "trace-uuid"
}
```

## 配置环境变量

关键环境变量（通过 `docker-compose.yml` 或 K8s ConfigMap/Secret 注入）：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_DSN` | 数据库连接串 | `root:root123456@tcp(mysql:3306)/...` |
| `REDIS_HOST` | Redis 地址 | `redis` |
| `JWT_SECRET` | JWT 签名密钥 | — |
| `KUBECONFIG` | K8s 配置文件路径 | `~/.kube/config` |
| `IMAGE_PULL_SECRETS` | 私有镜像仓库凭证（逗号分隔） | — |
| `HELM_CHART_PATH` | Helm Chart 路径 | `./helm-charts/mycloud-app` |

## 文档

- [架构设计文档](design.md) — 完整技术方案
- [API 文档](design.md#9-全模块-api-详细定义) — 全模块接口定义
- [数据库设计](design.md#7-完整-mysql-ddl) — DDL 与 ER 图

## 许可证

MIT License

