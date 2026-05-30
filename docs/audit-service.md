# Audit Service 使用文档

## 概述

Audit Service是my-cloud平台的审计日志服务，用于记录、查询和分析所有用户操作，提供完整的审计追踪能力。

## 功能特性

- **自动记录**: 通过中间件自动记录所有API操作
- **多维查询**: 支持按用户、资源、时间、操作类型等多维度查询
- **统计分析**: 提供操作统计、用户活跃度、资源访问热度等分析
- **日志导出**: 支持导出CSV格式的审计日志
- **数据清理**: 支持按保留天数清理过期审计日志
- **敏感信息脱敏**: 自动脱敏密码、token等敏感字段

## 审计日志记录内容

### 基本信息
- **用户信息**: 用户ID、用户名
- **操作信息**: 操作类型(create/update/delete/view)
- **资源信息**: 资源类型、资源ID、资源名称
- **请求信息**: HTTP方法、请求路径、请求体
- **响应信息**: 响应码、响应消息
- **环境信息**: IP地址、User-Agent
- **性能信息**: 请求耗时(毫秒)
- **时间信息**: 操作时间

### 操作类型
- `create` - 创建操作 (POST)
- `update` - 更新操作 (PUT/PATCH)
- `delete` - 删除操作 (DELETE)
- `view` - 查看操作 (GET)

### 资源类型
- `user` - 用户管理
- `role` - 角色管理
- `permission` - 权限管理
- `project` - 项目管理
- `application` - 应用管理
- `component` - 组件管理
- `pipeline` - 流水线
- `environment` - 环境管理
- `release` - 发布管理
- `deployment` - 部署管理
- `cluster` - 集群管理
- `notification` - 通知管理
- 等等...

## API接口

### 1. 获取审计日志列表

```
GET /api/v1/audit-logs
Authorization: Bearer {token}

Query Parameters:
- page: 页码 (默认: 1)
- pageSize: 每页数量 (默认: 20, 最大: 100)
- userId: 用户ID
- username: 用户名 (模糊匹配)
- action: 操作类型 (create/update/delete/view)
- resourceType: 资源类型
- resourceId: 资源ID
- method: HTTP方法 (GET/POST/PUT/DELETE)
- path: 请求路径 (模糊匹配)
- ipAddress: IP地址
- responseCode: 响应码
- startTime: 开始时间 (格式: 2006-01-02 或 2006-01-02 15:04:05)
- endTime: 结束时间 (格式: 2006-01-02 或 2006-01-02 15:04:05)
```

**示例**:

```bash
# 查询最近7天的所有审计日志
curl -X GET "http://localhost:8080/api/v1/audit-logs?page=1&pageSize=20" \
  -H "Authorization: Bearer $TOKEN"

# 查询特定用户的操作
curl -X GET "http://localhost:8080/api/v1/audit-logs?userId=1&page=1&pageSize=20" \
  -H "Authorization: Bearer $TOKEN"

# 查询特定时间范围的删除操作
curl -X GET "http://localhost:8080/api/v1/audit-logs?action=delete&startTime=2026-05-01&endTime=2026-05-31" \
  -H "Authorization: Bearer $TOKEN"

# 查询特定资源的所有操作
curl -X GET "http://localhost:8080/api/v1/audit-logs?resourceType=application&resourceId=123" \
  -H "Authorization: Bearer $TOKEN"
```

**响应**:

```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1001,
        "userId": 1,
        "username": "admin",
        "action": "create",
        "resourceType": "application",
        "resourceId": 123,
        "resourceName": "my-app",
        "method": "POST",
        "path": "/api/v1/applications",
        "ipAddress": "192.168.1.100",
        "userAgent": "Mozilla/5.0...",
        "requestBody": "{\"name\":\"my-app\",\"description\":\"...\",...}",
        "responseCode": 201,
        "responseMessage": "created",
        "durationMs": 156,
        "createTime": "2026-05-28 15:30:45"
      }
    ],
    "pagination": {
      "total": 1250,
      "page": 1,
      "pageSize": 20,
      "totalPages": 63
    }
  }
}
```

### 2. 获取审计日志详情

```
GET /api/v1/audit-logs/:id
Authorization: Bearer {token}
```

**示例**:

```bash
curl -X GET "http://localhost:8080/api/v1/audit-logs/1001" \
  -H "Authorization: Bearer $TOKEN"
```

### 3. 根据资源获取审计日志

```
GET /api/v1/audit-logs/resource/:resourceType/:resourceId
Authorization: Bearer {token}

Query Parameters:
- page: 页码 (默认: 1)
- pageSize: 每页数量 (默认: 20)
```

**示例**:

```bash
# 查看应用ID为123的所有操作历史
curl -X GET "http://localhost:8080/api/v1/audit-logs/resource/application/123?page=1&pageSize=20" \
  -H "Authorization: Bearer $TOKEN"
```

### 4. 根据用户获取审计日志

```
GET /api/v1/audit-logs/user/:userId
Authorization: Bearer {token}

Query Parameters:
- page: 页码 (默认: 1)
- pageSize: 每页数量 (默认: 20)
```

**示例**:

```bash
# 查看用户ID为1的所有操作历史
curl -X GET "http://localhost:8080/api/v1/audit-logs/user/1?page=1&pageSize=20" \
  -H "Authorization: Bearer $TOKEN"
```

### 5. 获取统计信息

```
GET /api/v1/audit-logs/statistics
Authorization: Bearer {token}

Query Parameters:
- startTime: 开始时间 (默认: 最近7天)
- endTime: 结束时间 (默认: 当前时间)
```

**示例**:

```bash
# 获取最近7天的统计信息
curl -X GET "http://localhost:8080/api/v1/audit-logs/statistics" \
  -H "Authorization: Bearer $TOKEN"

# 获取指定时间范围的统计
curl -X GET "http://localhost:8080/api/v1/audit-logs/statistics?startTime=2026-05-01&endTime=2026-05-31" \
  -H "Authorization: Bearer $TOKEN"
```

**响应**:

```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "start_time": "2026-05-21 15:35:00",
    "end_time": "2026-05-28 15:35:00",
    "total_count": 5280,
    "avg_duration_ms": 125.6,
    "action_stats": [
      {"action": "view", "count": 3150},
      {"action": "create", "count": 1020},
      {"action": "update", "count": 856},
      {"action": "delete": "count": 254}
    ],
    "resource_stats": [
      {"resourceType": "application", "count": 1580},
      {"resourceType": "deployment", "count": 1120},
      {"resourceType": "pipeline", "count": 980},
      {"resourceType": "release", "count": 756},
      {"resourceType": "user", "count": 544}
    ],
    "user_stats": [
      {"username": "admin", "count": 2150},
      {"username": "developer1", "count": 1280},
      {"username": "operator1", "count": 980}
    ],
    "status_stats": [
      {"responseCode": 200, "count": 3980},
      {"responseCode": 201, "count": 820},
      {"responseCode": 401, "count": 280},
      {"responseCode": 404, "count": 120}
    ]
  }
}
```

### 6. 导出审计日志

```
GET /api/v1/audit-logs/export
Authorization: Bearer {token}

Query Parameters: (与列表查询相同)
- userId, username, action, resourceType, startTime, endTime 等
```

**示例**:

```bash
# 导出最近30天的所有审计日志
curl -X GET "http://localhost:8080/api/v1/audit-logs/export?startTime=2026-04-28&endTime=2026-05-28" \
  -H "Authorization: Bearer $TOKEN" \
  -o audit_logs.csv

# 导出特定用户的操作记录
curl -X GET "http://localhost:8080/api/v1/audit-logs/export?userId=1" \
  -H "Authorization: Bearer $TOKEN" \
  -o user_1_audit_logs.csv
```

**导出格式**: CSV文件，包含以下列:
```
ID,用户ID,用户名,操作类型,资源类型,资源ID,请求方法,请求路径,IP地址,响应码,耗时(ms),创建时间
```

### 7. 清理过期日志

```
POST /api/v1/audit-logs/clean
Authorization: Bearer {token}
Content-Type: application/json

{
  "retentionDays": 90
}
```

**示例**:

```bash
# 清理90天前的审计日志
curl -X POST "http://localhost:8080/api/v1/audit-logs/clean" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "retentionDays": 90
  }'
```

**响应**:

```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "message": "清理完成",
    "deleted_count": 15680
  }
}
```

## 审计中间件

审计日志通过中间件自动记录，无需在业务代码中手动调用。

### 中间件配置

在gateway或各服务中启用审计中间件：

```go
import (
    "my-cloud/internal/common/middleware"
    "gorm.io/gorm"
)

func setupRouter(r *gin.Engine, db *gorm.DB) {
    // 添加审计中间件
    r.Use(middleware.AuditMiddleware(db))
    
    // 其他路由配置...
}
```

### 跳过审计的路径

以下路径默认跳过审计记录：
- `/api/v1/auth/login` - 登录请求
- `/api/v1/auth/refresh` - Token刷新
- `/api/v1/audit-logs*` - 审计日志查询本身
- `/health` - 健康检查
- `/metrics` - 监控指标

### 敏感信息脱敏

中间件会自动脱敏以下敏感字段：
- `password` - 密码
- `token` - 令牌
- `secret` - 密钥
- `apiKey` - API密钥
- `accessToken` - 访问令牌

脱敏后的值显示为: `***REDACTED***`

## 使用场景

### 1. 安全审计

查看所有删除操作，识别潜在的安全风险：

```bash
curl -X GET "http://localhost:8080/api/v1/audit-logs?action=delete&startTime=2026-05-01" \
  -H "Authorization: Bearer $TOKEN"
```

### 2. 故障排查

查看特定时间段的操作，定位故障原因：

```bash
curl -X GET "http://localhost:8080/api/v1/audit-logs?startTime=2026-05-28 14:00:00&endTime=2026-05-28 15:00:00&responseCode=500" \
  -H "Authorization: Bearer $TOKEN"
```

### 3. 用户行为分析

分析用户操作习惯，优化产品功能：

```bash
curl -X GET "http://localhost:8080/api/v1/audit-logs/statistics?startTime=2026-05-01&endTime=2026-05-31" \
  -H "Authorization: Bearer $TOKEN"
```

### 4. 合规审计

导出审计日志用于合规审查：

```bash
curl -X GET "http://localhost:8080/api/v1/audit-logs/export?startTime=2026-01-01&endTime=2026-12-31" \
  -H "Authorization: Bearer $TOKEN" \
  -o annual_audit_2026.csv
```

### 5. 资源追踪

追踪特定资源的完整生命周期：

```bash
# 查看应用从创建到删除的所有操作
curl -X GET "http://localhost:8080/api/v1/audit-logs/resource/application/123" \
  -H "Authorization: Bearer $TOKEN"
```

## 性能优化

### 索引优化

审计日志表建立了以下索引：
- `idx_user_id` - 用户ID索引
- `idx_username` - 用户名索引
- `idx_action` - 操作类型索引
- `idx_resource_type` - 资源类型索引
- `idx_resource_id` - 资源ID索引
- `idx_create_time` - 创建时间索引
- `idx_path` - 请求路径索引(前255字符)

### 异步写入

审计日志采用异步写入方式，不阻塞主请求处理：

```go
// 异步写入数据库
go func() {
    auditDB.Create(auditLog)
}()
```

### 定期清理

建议设置定时任务定期清理过期日志：

```bash
# Crontab示例: 每月1号凌晨2点清理90天前的日志
0 2 1 * * curl -X POST "http://localhost:8093/api/v1/audit-logs/clean" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"retentionDays": 90}'
```

## 最佳实践

### 1. 合理的保留期限

根据业务需求和合规要求设置日志保留期：
- **安全敏感**: 保留180天以上
- **一般业务**: 保留90天
- **测试环境**: 保留30天

### 2. 定期导出归档

对于重要的审计日志，建议定期导出到外部存储：

```bash
#!/bin/bash
# 每月导出上月审计日志
LAST_MONTH=$(date -d "last month" +%Y-%m)
curl -X GET "http://localhost:8080/api/v1/audit-logs/export?startTime=$LAST_MONTH-01&endTime=$LAST_MONTH-31" \
  -H "Authorization: Bearer $TOKEN" \
  -o "audit_logs_$LAST_MONTH.csv"
```

### 3. 监控审计日志

监控异常操作，及时发现安全风险：
- 监控删除操作的频率
- 监控失败请求(401/403/500)的数量
- 监控敏感资源的访问
- 监控异常IP地址的请求

### 4. 权限控制

严格控制审计日志的查看权限：
- 只有管理员可以查看所有审计日志
- 普通用户只能查看自己的操作记录
- 审计日志导出需要特殊权限

## 注意事项

1. **性能影响**: 审计日志采用异步写入，对性能影响很小(< 1ms)
2. **存储空间**: 审计日志会持续增长，需要定期清理或归档
3. **敏感信息**: 虽然已脱敏，但请求体中仍可能包含敏感信息，访问权限需严格控制
4. **时区问题**: 所有时间均为服务器时区，导出时注意时区转换
5. **查询性能**: 大范围时间查询可能较慢，建议添加其他过滤条件

## 故障排查

### 审计日志未记录

1. 检查中间件是否启用
2. 检查路径是否在跳过列表中
3. 检查数据库连接是否正常
4. 查看服务日志是否有错误

### 查询性能慢

1. 缩小时间范围
2. 添加更多过滤条件
3. 检查数据库索引
4. 考虑分表或归档历史数据

### 导出失败

1. 检查过滤条件是否合理
2. 限制导出数量(最多10000条)
3. 检查磁盘空间
4. 检查权限设置

## 后续优化

- [ ] 支持更细粒度的审计配置
- [ ] 支持自定义审计规则
- [ ] 支持审计日志分析报告
- [ ] 支持实时审计告警
- [ ] 支持审计日志可视化
- [ ] 集成到SIEM系统
- [ ] 支持审计日志加密存储
- [ ] 支持审计日志完整性校验
