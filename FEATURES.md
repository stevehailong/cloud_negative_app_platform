# 新功能说明

## 1. 审计日志中间件

### 功能描述
自动记录所有通过Gateway的API请求，包括：
- 用户信息（用户ID、用户名）
- 操作类型（create/update/delete/view）
- 资源类型和资源ID
- 请求方法和路径
- 客户端IP地址和User-Agent
- 请求体（敏感信息已脱敏）
- 响应状态码和响应消息
- 请求耗时（毫秒）
- 创建时间

### 实现文件
- `/backend/internal/common/middleware/audit.go` - 审计中间件
- `/backend/internal/common/model/audit.go` - 审计日志数据模型
- `/backend/cmd/gateway/main.go` - Gateway主程序（集成审计中间件）

### 数据库表
- `audit_db.audit_logs` - 审计日志表

### 特性
1. **异步记录**：不阻塞API请求，日志写入在后台执行
2. **敏感信息脱敏**：自动过滤password、token、secret、apiKey等敏感字段
3. **智能路径解析**：自动识别资源类型（user、project、application等）
4. **跳过特定路径**：login、register、refresh、health等路径不记录
5. **完整索引**：支持按用户、操作、资源类型、路径、时间等维度查询

### 使用示例
审计日志自动记录，无需手动调用。查询审计日志：

```sql
-- 查看最近的审计日志
SELECT user_id, username, action, resource_type, path, response_code, duration_ms, create_time 
FROM audit_db.audit_logs 
ORDER BY id DESC 
LIMIT 10;

-- 查看特定用户的操作
SELECT action, resource_type, path, create_time 
FROM audit_db.audit_logs 
WHERE user_id = 9 
ORDER BY create_time DESC;

-- 查看失败的请求
SELECT user_id, username, path, response_code, response_message, create_time 
FROM audit_db.audit_logs 
WHERE response_code >= 400 
ORDER BY create_time DESC;
```

## 2. JWT Token刷新机制

### 功能描述
实现Token刷新机制，提升用户体验：
- 登录时返回access_token（24小时有效）和refresh_token（7天有效）
- 用户可使用refresh_token换取新的access_token和refresh_token
- 无需重新登录即可延长会话

### 实现文件
- `/backend/pkg/jwt/jwt.go` - JWT工具包（新增GenerateRefreshToken和ParseTokenWithoutValidation）
- `/backend/internal/auth/service/auth_service.go` - 认证服务（修改Login返回值，新增RefreshToken方法）
- `/backend/internal/auth/handler/auth_handler.go` - 认证处理器（修改Login响应，新增RefreshToken处理器）
- `/backend/internal/auth/router/router.go` - 路由配置（新增/auth/refresh公开路由）

### API端点

#### 登录（返回两个token）
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}

# 响应
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",  # access_token (24小时)
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",  # refresh_token (7天)
    "user": { ... }
  }
}
```

#### 刷新Token
```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

# 响应
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",  # 新的access_token
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  # 新的refresh_token
  }
}
```

### 使用流程
1. 用户登录成功后，前端保存两个token：
   - `access_token`：用于API请求的Authorization header
   - `refresh_token`：用于刷新token

2. 当access_token即将过期或已过期时：
   - 使用refresh_token调用`/api/v1/auth/refresh`接口
   - 获取新的access_token和refresh_token
   - 更新本地存储的token

3. 如果refresh_token也过期：
   - 提示用户重新登录

### 前端集成建议
```javascript
// 保存token
localStorage.setItem('access_token', response.data.token);
localStorage.setItem('refresh_token', response.data.refreshToken);

// 在axios拦截器中处理token过期
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response.status === 401) {
      // Token过期，尝试刷新
      const refreshToken = localStorage.getItem('refresh_token');
      const res = await axios.post('/api/v1/auth/refresh', { refreshToken });
      
      // 保存新token
      localStorage.setItem('access_token', res.data.data.token);
      localStorage.setItem('refresh_token', res.data.data.refreshToken);
      
      // 重试原请求
      error.config.headers['Authorization'] = 'Bearer ' + res.data.data.token;
      return axios.request(error.config);
    }
    return Promise.reject(error);
  }
);
```

## 测试验证

### 测试审计日志
```bash
# 1. 注册用户（会记录审计日志）
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123456","email":"test@example.com"}'

# 2. 登录获取token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123456"}'

# 3. 调用需要认证的API
curl -X GET "http://localhost:8080/api/v1/users/?page=1&pageSize=10" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# 4. 查看审计日志
docker exec my-cloud-mysql mysql -uroot -proot123456 \
  -e "SELECT * FROM audit_db.audit_logs ORDER BY id DESC LIMIT 5;"
```

### 测试Token刷新
```bash
# 1. 登录获取refresh_token
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123456"}')

REFRESH_TOKEN=$(echo $RESPONSE | jq -r '.data.refreshToken')

# 2. 使用refresh_token刷新
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\":\"$REFRESH_TOKEN\"}"

# 3. 会返回新的access_token和refresh_token
```

## 部署说明

两个功能已集成到以下服务：
- **Gateway**：审计中间件已集成，所有经过Gateway的请求都会被记录
- **Auth Service**：Token刷新功能已实现

重新构建并启动服务：
```bash
docker-compose build auth-service gateway
docker-compose up -d auth-service gateway
```

服务状态检查：
```bash
docker logs my-cloud-gateway --tail 20
docker logs my-cloud-auth-service --tail 20
```
