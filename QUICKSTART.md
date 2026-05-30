# My Cloud - 快速启动指南

## 当前实现状态

✅ **已实现的服务：**
- Gateway (API网关) - 端口 8080
- Auth Service (认证服务) - 端口 8081  
- Application Service (应用服务) - 端口 8083
- Frontend (前端) - 端口 80

✅ **基础设施：**
- MySQL 8.0 - 端口 3306
- Redis 7 - 端口 6379

## 快速启动步骤

### 1. 启动所有服务

```bash
cd /Users/hanhailong01/Downloads/my_cloud

# 启动服务（会自动构建镜像）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 2. 等待服务就绪

首次启动需要等待：
- MySQL 初始化（约30秒）
- Docker 镜像构建（约3-5分钟）
- 服务启动（约10秒）

检查服务是否就绪：
```bash
# 检查 MySQL
docker exec -it my-cloud-mysql mysqladmin ping -uroot -proot123456

# 检查 Gateway
curl http://localhost:8080/health

# 检查 Auth Service  
curl http://localhost:8081/health

# 检查 Application Service
curl http://localhost:8083/health
```

### 3. 初始化数据库（可选）

SQL 脚本会在 MySQL 容器启动时自动执行，如果需要手动执行：

```bash
make init-db
```

### 4. 访问应用

- **前端地址**: http://localhost
- **API网关**: http://localhost:8080
- **默认账号**: admin
- **默认密码**: admin123

## 常用命令

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f gateway
docker-compose logs -f auth-service

# 重启服务
docker-compose restart gateway

# 重新构建并启动
docker-compose up -d --build

# 清理所有容器和数据
docker-compose down -v
```

## 本地开发（不使用Docker）

### 后端开发

1. **启动基础设施**
```bash
docker-compose up -d mysql redis
```

2. **启动 Gateway**
```bash
cd backend
export DB_DSN="root:root123456@tcp(localhost:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_HOST="localhost"
export JWT_SECRET="my-cloud-secret-key"
go run cmd/gateway/main.go
```

3. **启动 Auth Service**
```bash
cd backend
export DB_DSN="root:root123456@tcp(localhost:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_HOST="localhost"  
export JWT_SECRET="my-cloud-secret-key"
go run cmd/auth-service/main.go
```

4. **启动 Application Service**
```bash
cd backend
export DB_DSN="root:root123456@tcp(localhost:3306)/app_db?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_HOST="localhost"
export JWT_SECRET="my-cloud-secret-key"
go run cmd/application-service/main.go
```

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

访问: http://localhost:3000

## 故障排查

### 1. MySQL 连接失败

```bash
# 检查 MySQL 是否启动
docker ps | grep mysql

# 查看 MySQL 日志
docker logs my-cloud-mysql

# 进入 MySQL 容器
docker exec -it my-cloud-mysql mysql -uroot -proot123456

# 查看数据库
SHOW DATABASES;
```

### 2. 服务无法启动

```bash
# 查看具体服务日志
docker logs my-cloud-gateway
docker logs my-cloud-auth-service
docker logs my-cloud-application-service

# 重新构建镜像
docker-compose build --no-cache gateway
docker-compose up -d gateway
```

### 3. 端口冲突

修改 docker-compose.yml 中的端口映射：
```yaml
ports:
  - "8888:8080"  # 将8080改为8888
```

### 4. 前端无法访问后端

检查网关是否正常：
```bash
curl -v http://localhost:8080/health
```

检查前端 nginx 配置：
```bash
docker exec -it my-cloud-frontend cat /etc/nginx/conf.d/default.conf
```

## API 测试

### 注册用户
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test",
    "password": "test123",
    "email": "test@example.com",
    "realName": "测试用户"
  }'
```

### 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

### 获取用户信息
```bash
# 先登录获取 token，然后：
curl -X GET http://localhost:8080/api/v1/auth/userinfo \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 创建应用
```bash
curl -X POST http://localhost:8080/api/v1/applications \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试应用",
    "code": "test-app",
    "projectId": 1,
    "type": "web",
    "language": "go",
    "description": "这是一个测试应用"
  }'
```

### 查询应用列表
```bash
curl -X GET "http://localhost:8080/api/v1/applications?page=1&pageSize=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 数据库说明

当前已创建的数据库：
- `iam_db`: 用户认证和权限管理
- `org_db`: 组织和项目管理
- `app_db`: 应用和组件管理

其他数据库会在后续服务实现时创建。

## 技术栈

**后端**
- Go 1.22
- Gin Web 框架
- GORM ORM
- MySQL 8.0
- Redis 7
- JWT 认证

**前端**
- Vue 3
- Element Plus
- Pinia
- Vue Router
- Vite
- Axios

## 下一步开发

当前项目已实现基础框架和核心功能，后续可以：

1. 实现其他微服务（pipeline, deployment, cluster等）
2. 完善前端页面功能
3. 添加更多的业务逻辑
4. 集成 CI/CD
5. 添加监控和日志收集
6. 部署到 Kubernetes

## 问题反馈

如遇到问题，请检查：
1. Docker 和 Docker Compose 版本
2. 端口是否被占用
3. 服务日志输出
4. 数据库连接状态
