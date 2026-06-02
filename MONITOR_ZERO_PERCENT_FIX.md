# 监控页面显示0%问题修复

## 问题描述
监控页面在查询book-service应用时，所有指标显示为0.0%。

## 根本原因
**应用名称字段不匹配**：
- 数据库中应用有两个字段：
  - `name`: "book-service" 
  - `code`: "hanhailong-book-001"
- K8s中Pod的标签使用的是应用名称：`app=book-service`
- 前端代码错误地传递了 `code` 字段（`hanhailong-book-001`），导致无法找到对应的Pod

## 数据流分析

### 错误的流程
```
前端 → appName=hanhailong-book-001 (code字段)
  ↓
后端 → 查询 K8s: app=hanhailong-book-001
  ↓
结果 → 找不到Pod → 返回0值
```

### 正确的流程
```
前端 → appName=book-service (name字段)
  ↓
后端 → 查询 K8s: app=book-service
  ↓
结果 → 找到4个Pod → 返回真实数据
```

## 解决方案

### 修改前端代码

**文件**: `frontend/src/views/monitor/MonitorDashboard.vue`

```javascript
// 错误的代码（已修复）
const appName = selectedApp?.code || selectedApp?.name || ''

// 正确的代码
const appName = selectedApp?.name || ''
```

**修改位置**: fetchMetrics 函数第7行

### 验证修复

#### 1. 测试API（使用code - 错误）
```bash
curl 'http://localhost/internal/v1/metrics/apps/8?appName=hanhailong-book-001'
# 返回: cpu: 0, memory: 0, qps: 0
```

#### 2. 测试API（使用name - 正确）
```bash
curl 'http://localhost/internal/v1/metrics/apps/8?appName=book-service'
# 返回: cpu: 14, memory: 30, qps: 60, pod_count: 4
```

## 部署步骤

```bash
cd /Users/hanhailong01/Downloads/my_cloud

# 重新构建前端
docker-compose build frontend

# 重启前端服务
docker-compose up -d frontend

# 等待服务启动（约5秒）
sleep 5

# 清除浏览器缓存
# Cmd + Shift + R (Mac)

# 访问页面测试
open http://localhost/monitors
```

## 预期结果

访问 http://localhost/monitors，选择book-service应用：

| 指标 | 预期值 |
|-----|-------|
| CPU使用率 | 14.0% |
| 内存使用率 | 30.0% |
| 请求QPS | 60 |
| 错误率 | 0.1% |
| Pod数量 | 4 |

## 日志验证

### 查看nginx访问日志
```bash
docker logs my-cloud-frontend --tail 5
```

应该看到：
```
GET /internal/v1/metrics/apps/8?timeRange=1h&appName=book-service HTTP/1.1" 200
```

### 查看API返回
打开浏览器开发者工具 → Network → 找到metrics请求 → Response：
```json
{
  "code": 0,
  "data": {
    "cpu": 14,
    "memory": 30,
    "qps": 60,
    "pod_count": 4
  }
}
```

## 经验教训

### 1. 应用标识规范
在微服务架构中，应用的标识需要统一：
- 数据库中的应用名称 = K8s中的标签值
- 避免使用code等其他字段作为K8s标识

### 2. 建议的应用字段规范
```sql
CREATE TABLE applications (
  id BIGINT PRIMARY KEY,
  name VARCHAR(100),      -- 应用名称，用于K8s标签
  display_name VARCHAR(100), -- 显示名称，用于UI展示
  code VARCHAR(100),      -- 应用编码，用于内部标识
  ...
);
```

**使用规则**：
- K8s Pod标签：使用 `name` 字段
- 前端显示：使用 `display_name` 或 `name`
- 监控查询：使用 `name` 作为K8s标签选择器

### 3. K8s部署规范
部署应用到K8s时，确保标签与数据库一致：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: book-service
spec:
  selector:
    matchLabels:
      app: book-service  # 必须与数据库的name字段一致
  template:
    metadata:
      labels:
        app: book-service  # 保持一致
```

## 故障排查清单

如果监控数据仍为0，请按顺序检查：

- [ ] 1. K8s集群是否连接成功
  ```bash
  docker logs my-cloud-monitor-service | grep "K8s client initialized"
  ```

- [ ] 2. Pod标签是否正确
  ```bash
  kubectl get pods -n default -l app=book-service
  ```

- [ ] 3. 前端传递的appName参数
  ```bash
  docker logs my-cloud-frontend | grep "appName"
  ```

- [ ] 4. API返回的数据
  ```bash
  curl 'http://localhost/internal/v1/metrics/apps/8?appName=book-service'
  ```

- [ ] 5. 浏览器缓存是否清除
  - 使用 Cmd + Shift + R 硬刷新
  - 或使用隐私模式

## 总结

**问题**: 前端传递了错误的字段（code而非name）作为K8s标签选择器

**修复**: 使用应用的name字段匹配K8s Pod标签

**验证**: API返回真实的监控数据（CPU 14%, Memory 30%, QPS 60）

**状态**: ✅ 已修复，等待前端构建和部署
