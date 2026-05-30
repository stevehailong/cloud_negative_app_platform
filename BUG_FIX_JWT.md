# Bug 修复记录

## 问题描述

用户使用 admin 账号登录后，访问应用管理页面时返回错误：
```json
{
    "code": 40101,
    "message": "invalid token: token signature is invalid: signature is invalid",
    "requestId": "32d511c2-5378-44e9-8632-e2f2bce27ec9"
}
```

## 根本原因

**JWT Secret 未初始化**

application-service 在启动时没有调用 `jwt.InitJWT(cfg.JWT.Secret)` 来初始化 JWT 密钥。当请求通过认证中间件时，JWT 包使用的是未初始化的空密钥来验证 token，导致签名验证失败。

## 修复方案

### 1. 后端修复

在所有需要验证 JWT token 的服务启动代码中添加 JWT 初始化：

**修改文件**: `backend/cmd/application-service/main.go`
```go
import (
    // ... 其他导入
    "my-cloud/pkg/jwt"
)

func main() {
    cfg := config.LoadConfig()
    
    // 添加这一行 - 初始化JWT
    jwt.InitJWT(cfg.JWT.Secret)
    
    // ... 其余代码
}
```

**修改文件**: `backend/cmd/gateway/main.go`
```go
import (
    // ... 其他导入
    "my-cloud/pkg/jwt"
)

func main() {
    cfg := config.LoadConfig()
    
    // 添加这一行 - 初始化JWT
    jwt.InitJWT(cfg.JWT.Secret)
    
    // ... 其余代码
}
```

### 2. 前端修复 - API 路径规范化

由于 Gin 框架的路由规则，需要在 API 路径末尾添加斜杠以避免 301 重定向。

**修改文件**: `frontend/src/api/application.js`

将所有 API 路径从：
```javascript
url: '/applications'          // ❌ 会导致 301 重定向
url: '/applications/:id'
```

改为：
```javascript
url: '/applications/'         // ✅ 直接匹配路由
url: '/applications/:id/'
```

## 验证测试

### 测试1: 登录并获取 Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

预期结果：返回包含有效 token 的响应

### 测试2: 使用 Token 访问应用列表
```bash
TOKEN="<从登录获取的token>"
curl -X GET "http://localhost:8080/api/v1/applications/?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"
```

预期结果：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 0,
    "page": 1,
    "pageSize": 10,
    "items": []
  }
}
```

### 测试3: 创建应用
```bash
curl -X POST "http://localhost:8080/api/v1/applications/" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试应用",
    "code": "test-app-001",
    "projectId": 1,
    "type": "web",
    "language": "go"
  }'
```

预期结果：成功创建应用并返回应用详情

## 部署修复

### 重新构建服务
```bash
cd /Users/hanhailong01/Downloads/my_cloud

# 重新构建后端服务
docker-compose up -d --build gateway application-service

# 重新构建前端
docker-compose up -d --build frontend
```

### 验证服务状态
```bash
# 检查服务是否正常运行
docker-compose ps

# 查看日志
docker-compose logs -f gateway
docker-compose logs -f application-service
```

## 影响范围

- ✅ 修复了 JWT token 验证失败的问题
- ✅ 修复了 API 路径重定向问题
- ✅ 用户登录后可以正常访问所有需要认证的接口
- ✅ 前端应用管理页面可以正常使用

## 相关文件

**后端**:
- `backend/cmd/gateway/main.go`
- `backend/cmd/application-service/main.go`
- `backend/pkg/jwt/jwt.go`
- `backend/internal/common/middleware/auth.go`

**前端**:
- `frontend/src/api/application.js`

## 预防措施

为了避免类似问题，建议：

1. **服务启动检查清单**
   - [ ] 配置文件加载
   - [ ] 数据库连接初始化
   - [ ] JWT 密钥初始化
   - [ ] Redis 连接初始化
   - [ ] 中间件注册

2. **添加启动日志**
   在每个服务的 main 函数中添加关键初始化步骤的日志输出

3. **集成测试**
   添加端到端测试来验证完整的认证流程

4. **文档更新**
   在服务开发指南中明确说明 JWT 初始化的必要性

## 修复时间

- 问题发现: 2026-05-28 05:42
- 问题修复: 2026-05-28 05:53
- 验证完成: 2026-05-28 05:54
- 总耗时: 约 12 分钟

## 状态

✅ **已修复并验证通过**

所有功能恢复正常，用户可以顺利登录并使用应用管理功能。
