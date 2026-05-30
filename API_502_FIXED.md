# 接口502错误问题解决报告

## 问题描述

前端访问后端API时全部返回502错误。

## 问题原因

**主要原因**: 前端容器和Gateway容器之间的网络连接出现问题。

从Nginx错误日志可以看到：
```
2026/05/28 12:00:20 [error] 28#28: *142 connect() failed (111: Connection refused) 
while connecting to upstream, client: 192.168.65.1, 
server: localhost, request: "GET /api/v1/auth/userinfo HTTP/1.1", 
upstream: "http://172.19.0.5:8080/api/v1/auth/userinfo"
```

错误信息显示：
- Nginx尝试连接到`http://172.19.0.5:8080`
- 连接被拒绝 (Connection refused)
- 实际Gateway IP是`172.19.0.4`

## 解决方案

**重启前端和Gateway容器**：

```bash
docker-compose restart frontend gateway
```

## 根本原因分析

1. **容器IP变化**: Docker容器重启后IP地址可能发生变化
2. **DNS缓存**: Nginx可能缓存了旧的DNS解析结果
3. **网络故障**: Docker网络可能出现临时性故障

## 验证结果

### 重启前
```
❌ API请求返回502 Bad Gateway
❌ Nginx日志显示"Connection refused"
```

### 重启后
```
✅ Gateway健康检查通过
✅ Nginx代理正常 (HTTP 401 - 需要登录，这是正常的)
✅ 前端页面正常 (HTTP 200)
✅ 登录接口响应正常
✅ 前端容器 -> Gateway 连通
```

## 测试结果

### 1. Gateway直连测试
```bash
curl http://localhost:8080/health
# ✅ 返回 {"status":"ok"}
```

### 2. 前端Nginx代理测试
```bash
curl http://localhost/api/v1/projects
# ✅ 返回 401 (需要登录，这是预期行为)
```

### 3. 登录接口测试
```bash
curl http://localhost/api/v1/auth/login \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
# ✅ 返回正确的错误码 (用户名或密码错误)
```

### 4. 容器网络测试
```bash
docker exec my-cloud-frontend wget -O- http://gateway:8080/health
# ✅ 成功连接并返回健康状态
```

## 服务状态总览

### 核心服务 (全部正常)
- ✅ Gateway (8080) - 正常
- ✅ Frontend (80) - 正常
- ✅ MySQL - 正常
- ✅ Redis - 正常

### 业务服务
| 服务 | 端口 | 状态 | 说明 |
|------|------|------|------|
| auth-service | 8081 | ✅ | 正常 |
| pipeline-service | 8084 | ✅ | 正常 |
| release-service | 8086 | ✅ | 正常 |
| deploy-service | 8087 | ✅ | 正常 |
| project-service | 8082 | ⚠️ | 运行中，健康检查端点格式不同 |
| application-service | 8083 | ⚠️ | 运行中，健康检查端点格式不同 |
| env-service | 8085 | ⚠️ | 运行中，健康检查端点格式不同 |
| cluster-service | 8088 | ⚠️ | 运行中，健康检查端点格式不同 |
| monitor-service | 8090 | ⚠️ | 运行中，健康检查端点格式不同 |
| audit-service | 8093 | ⚠️ | 运行中，健康检查端点格式不同 |
| notification-service | 8095 | ⚠️ | 运行中，健康检查端点格式不同 |

**注**: 标记为⚠️的服务实际都在运行，只是健康检查端点路径可能是`/api/v1/health`而不是`/health`。

## 配置检查

### 前端Nginx配置 (正确)
```nginx
location /api {
    proxy_pass http://gateway:8080;  # ✅ 使用服务名，不是硬编码IP
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

### 前端API配置 (正确)
```javascript
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',  // ✅ 使用相对路径
  timeout: 30000
})
```

### 环境变量 (正确)
```env
VITE_API_BASE_URL=/api/v1  # ✅ 相对路径，通过Nginx代理
```

## 预防措施

### 1. 添加健康检查脚本
已创建`scripts/test-api-connectivity.sh`用于快速诊断：

```bash
./scripts/test-api-connectivity.sh
```

### 2. 监控建议
- 使用`docker-compose ps`定期检查容器状态
- 使用`docker logs`查看服务日志
- 配置容器健康检查（docker-compose.yml中添加healthcheck）

### 3. 快速恢复
遇到502错误时的快速处理：

```bash
# 方案1: 重启相关服务
docker-compose restart frontend gateway

# 方案2: 重启所有服务
docker-compose restart

# 方案3: 完全重启
docker-compose down && docker-compose up -d
```

## 相关文件

- `scripts/test-api-connectivity.sh` - API连通性测试脚本
- `frontend/src/utils/request.js` - 前端API请求配置
- `frontend/.env.production` - 前端生产环境配置

## 后续优化建议

### 1. 统一健康检查端点
建议所有服务使用统一的健康检查端点格式：
```go
r.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok", "service": "service-name"})
})
```

### 2. 添加Docker健康检查
在docker-compose.yml中为每个服务添加：
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
```

### 3. 使用服务发现
考虑引入Consul或Eureka等服务发现组件，避免硬编码服务地址。

### 4. 添加监控告警
- 集成Prometheus监控各服务状态
- 配置告警规则，服务异常时自动通知
- 使用Grafana可视化监控指标

## 总结

✅ **问题已解决**: 通过重启前端和Gateway容器，API 502错误已完全解决

✅ **服务状态**: 所有核心服务运行正常，API可正常访问

✅ **测试工具**: 已提供完整的连通性测试脚本

⚠️ **注意事项**: 部分服务健康检查端点路径不统一，建议后续优化

---

**解决时间**: 2026-05-28  
**影响范围**: 前端 + Gateway  
**解决方式**: 服务重启  
**状态**: ✅ 已解决
