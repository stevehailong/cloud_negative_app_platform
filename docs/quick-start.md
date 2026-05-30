# My-Cloud 快速开始指南

本指南将帮助你在5分钟内启动my-cloud平台的所有服务。

## 前置要求

确保你的机器上已安装：
- **Docker** 20.10+ 
- **Docker Compose** 2.0+

## 一键启动

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/my-cloud.git
cd my-cloud
```

### 2. 启动所有服务

```bash
# 启动所有服务（首次启动会自动构建镜像）
docker-compose up -d

# 查看启动进度
docker-compose ps
```

### 3. 等待服务就绪

```bash
# 查看所有服务状态（等待所有服务变为Up状态）
watch docker-compose ps

# 或查看特定服务日志
docker-compose logs -f notification-service
```

等待约2-3分钟，直到所有服务状态显示为 `Up` 和 `healthy`。

## 验证服务

### 检查所有服务健康状态

```bash
# Gateway
curl http://localhost:8080/health

# Auth Service
curl http://localhost:8081/health

# Project Service
curl http://localhost:8082/health

# Application Service
curl http://localhost:8083/health

# Pipeline Service
curl http://localhost:8084/health

# Environment Service
curl http://localhost:8085/health

# Release Service
curl http://localhost:8086/health

# Deploy Service
curl http://localhost:8087/health

# Cluster Service
curl http://localhost:8088/health

# Notification Service ✨
curl http://localhost:8095/health
```

所有服务应返回: `{"status":"ok"}`

### 测试Notification Service

```bash
# 运行自动化测试脚本
chmod +x scripts/test-notification-service.sh
./scripts/test-notification-service.sh
```

## 快速使用

### 1. 获取JWT Token

```bash
# 注册用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "email": "admin@example.com"
  }'

# 登录获取token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# 保存返回的token
export TOKEN="your-jwt-token-here"
```

### 2. 发送通知

#### 直接发送通知

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "测试通知",
    "content": "这是一条测试通知",
    "notifyType": "system",
    "channel": "dingtalk",
    "receiverType": "user",
    "receiverIds": "1,2,3"
  }'
```

#### 通过模板发送通知

```bash
curl -X POST http://localhost:8080/api/v1/notifications/template \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "templateCode": "RELEASE_SUCCESS",
    "params": {
      "projectName": "my-project",
      "version": "v1.0.0",
      "environment": "production",
      "operator": "张三",
      "releaseTime": "2026-05-28 15:00:00"
    },
    "receiverType": "user",
    "receiverIds": [1, 2, 3]
  }'
```

### 3. 查询通知

```bash
# 获取通知列表
curl -X GET "http://localhost:8080/api/v1/notifications?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"

# 获取通知详情
curl -X GET "http://localhost:8080/api/v1/notifications/1" \
  -H "Authorization: Bearer $TOKEN"
```

### 4. 管理模板

```bash
# 创建自定义模板
curl -X POST http://localhost:8080/api/v1/notification-templates \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "templateCode": "CUSTOM_NOTIFY",
    "templateName": "自定义通知",
    "notifyType": "system",
    "channel": "dingtalk",
    "title": "【系统通知】{{title}}",
    "content": "{{content}}",
    "variables": "[\"title\",\"content\"]"
  }'

# 获取所有模板
curl -X GET "http://localhost:8080/api/v1/notification-templates" \
  -H "Authorization: Bearer $TOKEN"
```

### 5. 管理渠道

```bash
# 获取所有渠道
curl -X GET "http://localhost:8080/api/v1/notification-channels" \
  -H "Authorization: Bearer $TOKEN"

# 创建新渠道
curl -X POST http://localhost:8080/api/v1/notification-channels \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "channelCode": "DINGTALK_DEV",
    "channelName": "钉钉开发环境",
    "channelType": "dingtalk",
    "config": "{\"webhook\":\"https://your-webhook-url\",\"secret\":\"your-secret\"}"
  }'
```

## 常见操作

### 查看服务日志

```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs notification-service

# 实时查看日志
docker-compose logs -f notification-service

# 查看最近100行日志
docker-compose logs --tail=100 notification-service
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart notification-service
```

### 停止服务

```bash
# 停止所有服务
docker-compose stop

# 停止并删除容器
docker-compose down

# 停止并删除容器、网络和卷
docker-compose down -v
```

### 重新构建服务

```bash
# 重新构建所有服务
docker-compose build

# 重新构建特定服务
docker-compose build notification-service

# 重新构建并启动
docker-compose up -d --build
```

## 数据库管理

### 连接数据库

```bash
# 通过Docker容器连接MySQL
docker exec -it my-cloud-mysql mysql -uroot -proot123456

# 查看所有数据库
SHOW DATABASES;

# 使用notification_db
USE notification_db;

# 查看表
SHOW TABLES;

# 查询数据
SELECT * FROM notification_templates;
```

### 备份数据库

```bash
# 备份单个数据库
docker exec my-cloud-mysql mysqldump -uroot -proot123456 notification_db > backup_notification.sql

# 备份所有数据库
docker exec my-cloud-mysql mysqldump -uroot -proot123456 --all-databases > backup_all.sql
```

### 恢复数据库

```bash
# 恢复数据库
docker exec -i my-cloud-mysql mysql -uroot -proot123456 notification_db < backup_notification.sql
```

## 故障排查

### 服务启动失败

```bash
# 1. 查看容器状态
docker-compose ps

# 2. 查看失败服务的日志
docker-compose logs notification-service

# 3. 检查端口占用
lsof -i :8095

# 4. 重新构建并启动
docker-compose up -d --build notification-service
```

### 数据库连接失败

```bash
# 1. 检查MySQL容器状态
docker-compose ps mysql

# 2. 检查MySQL健康状态
docker exec my-cloud-mysql mysqladmin ping -h localhost -uroot -proot123456

# 3. 查看MySQL日志
docker-compose logs mysql
```

### 服务无响应

```bash
# 1. 检查服务健康状态
curl http://localhost:8095/health

# 2. 检查容器资源使用
docker stats my-cloud-notification-service

# 3. 重启服务
docker-compose restart notification-service
```

## 性能监控

### 查看容器资源使用

```bash
# 实时查看所有容器资源使用
docker stats

# 查看特定容器
docker stats my-cloud-notification-service
```

### 查看容器详情

```bash
# 查看容器详细信息
docker inspect my-cloud-notification-service

# 查看容器日志大小
docker ps -s
```

## 开发模式

如果你想在本地开发模式运行：

### 1. 安装Go环境

```bash
# 确保Go 1.22+已安装
go version
```

### 2. 启动依赖服务

```bash
# 只启动MySQL和Redis
docker-compose up -d mysql redis
```

### 3. 本地运行服务

```bash
cd backend

# 安装依赖
go mod download

# 运行notification-service
cd cmd/notification-service
go run main.go

# 服务将在 http://localhost:8095 启动
```

### 4. 热重载开发

```bash
# 安装air (Go热重载工具)
go install github.com/cosmtrek/air@latest

# 在服务目录运行
cd backend/cmd/notification-service
air
```

## 下一步

- 📖 阅读[完整API文档](docs/notification-service.md)
- 🔍 查看[项目实现进度](docs/implementation-progress.md)
- 📊 查看[实现成果报告](docs/notification-service-report.md)
- 🎨 查看[设计文档](docs/design.md)

## 获取帮助

如遇到问题，请：
1. 查看[故障排查](#故障排查)章节
2. 查看服务日志: `docker-compose logs <service-name>`
3. 提交Issue: [GitHub Issues](https://github.com/yourusername/my-cloud/issues)

---

**恭喜！🎉** 你已成功启动my-cloud平台！
