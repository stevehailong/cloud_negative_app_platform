# Monitor Service API 文档

## 服务概述

Monitor Service（监控告警服务）提供指标管理、告警规则配置和告警记录管理功能，支持与Prometheus、Grafana、Loki、Jaeger等监控系统集成。

- **服务端口**: 8090
- **数据库**: monitor_db
- **API前缀**: `/api/v1`

## 核心功能

1. **指标管理** - 定义和管理监控指标
2. **告警规则** - 配置告警条件和通知规则
3. **告警记录** - 记录和追踪告警事件
4. **日志查询** - 集成Loki日志查询
5. **链路追踪** - 集成Jaeger分布式追踪

## API端点

### 1. 指标管理

#### 1.1 创建指标

```bash
POST /api/v1/metrics
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "api_requests_total",
  "type": "counter",
  "description": "API请求总数",
  "unit": "requests",
  "labels": "{\"service\": \"user-service\"}",
  "enabled": 1
}
```

**指标类型**:
- `counter`: 计数器（只增不减）
- `gauge`: 仪表盘（可增可减）
- `histogram`: 直方图
- `summary`: 摘要

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 6,
    "name": "api_requests_total",
    "type": "counter",
    "description": "API请求总数",
    "unit": "requests",
    "labels": "{\"service\": \"user-service\"}",
    "enabled": 1,
    "createTime": "2026-05-28T10:30:00Z",
    "updateTime": "2026-05-28T10:30:00Z"
  }
}
```

#### 1.2 获取指标列表

```bash
GET /api/v1/metrics?type=counter&enabled=1&page=1&page_size=20
Authorization: Bearer <token>
```

**查询参数**:
- `type`: 指标类型筛选
- `enabled`: 启用状态 (0-禁用, 1-启用)
- `page`: 页码
- `page_size`: 每页数量

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "http_requests_total",
        "type": "counter",
        "unit": "requests",
        "enabled": 1
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

#### 1.3 获取指标详情

```bash
GET /api/v1/metrics/:id
Authorization: Bearer <token>
```

#### 1.4 更新指标

```bash
PUT /api/v1/metrics/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "description": "API请求总数 (更新)",
  "enabled": 0
}
```

#### 1.5 删除指标

```bash
DELETE /api/v1/metrics/:id
Authorization: Bearer <token>
```

### 2. 告警规则管理

#### 2.1 创建告警规则

```bash
POST /api/v1/alert-rules
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "高CPU使用率告警",
  "metric_name": "cpu_usage_percent",
  "condition": ">",
  "threshold": 80.0,
  "duration": 300,
  "severity": "critical",
  "enabled": 1,
  "notify_users": "admin,ops"
}
```

**告警条件**:
- `>`: 大于
- `<`: 小于
- `>=`: 大于等于
- `<=`: 小于等于
- `==`: 等于
- `!=`: 不等于

**严重级别**:
- `critical`: 严重
- `warning`: 警告
- `info`: 信息

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 6,
    "name": "高CPU使用率告警",
    "metric_name": "cpu_usage_percent",
    "condition": ">",
    "threshold": 80.0,
    "duration": 300,
    "severity": "critical",
    "enabled": 1,
    "notify_users": "admin,ops",
    "createTime": "2026-05-28T10:35:00Z"
  }
}
```

#### 2.2 获取告警规则列表

```bash
GET /api/v1/alert-rules?metric_name=cpu_usage_percent&severity=critical&enabled=1&page=1&page_size=20
Authorization: Bearer <token>
```

**查询参数**:
- `metric_name`: 指标名称筛选
- `severity`: 严重级别筛选
- `enabled`: 启用状态
- `page`: 页码
- `page_size`: 每页数量

#### 2.3 获取告警规则详情

```bash
GET /api/v1/alert-rules/:id
Authorization: Bearer <token>
```

#### 2.4 更新告警规则

```bash
PUT /api/v1/alert-rules/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "threshold": 90.0,
  "enabled": 1
}
```

#### 2.5 删除告警规则

```bash
DELETE /api/v1/alert-rules/:id
Authorization: Bearer <token>
```

### 3. 告警记录管理

#### 3.1 获取告警列表

```bash
GET /api/v1/alerts?status=firing&severity=critical&start_time=2026-05-01&end_time=2026-05-28&page=1&page_size=20
Authorization: Bearer <token>
```

**查询参数**:
- `status`: 告警状态 (firing-触发中, resolved-已解决)
- `severity`: 严重级别
- `start_time`: 开始时间 (格式: YYYY-MM-DD)
- `end_time`: 结束时间 (格式: YYYY-MM-DD)
- `page`: 页码
- `page_size`: 每页数量

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "rule_id": 1,
        "rule_name": "高CPU使用率告警",
        "metric_name": "cpu_usage_percent",
        "current_value": 85.6,
        "threshold": 80.0,
        "severity": "critical",
        "status": "firing",
        "message": "CPU使用率超过阈值",
        "fired_at": "2026-05-28T10:00:00Z",
        "resolved_at": null
      }
    ],
    "total": 15,
    "page": 1,
    "page_size": 20
  }
}
```

#### 3.2 获取告警详情

```bash
GET /api/v1/alerts/:id
Authorization: Bearer <token>
```

#### 3.3 解决告警

```bash
POST /api/v1/alerts/:id/resolve
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "告警已解决"
  }
}
```

#### 3.4 获取告警统计

```bash
GET /api/v1/alerts/statistics?start_time=2026-05-01&end_time=2026-05-28
Authorization: Bearer <token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_alerts": 156,
    "firing_alerts": 8,
    "resolved_alerts": 148,
    "by_severity": {
      "critical": 45,
      "warning": 89,
      "info": 22
    },
    "by_status": {
      "firing": 8,
      "resolved": 148
    },
    "start_time": "2026-05-01 00:00:00",
    "end_time": "2026-05-28 23:59:59"
  }
}
```

## 数据模型

### Metric 指标模型

```go
type Metric struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`         // 指标名称
    Type        string    `json:"type"`         // counter/gauge/histogram/summary
    Description string    `json:"description"`  // 指标描述
    Unit        string    `json:"unit"`         // 单位
    Labels      string    `json:"labels"`       // JSON格式标签
    Enabled     int       `json:"enabled"`      // 是否启用
    CreateTime  time.Time `json:"createTime"`
    UpdateTime  time.Time `json:"updateTime"`
}
```

### AlertRule 告警规则模型

```go
type AlertRule struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`         // 规则名称
    MetricName  string    `json:"metricName"`   // 指标名称
    Condition   string    `json:"condition"`    // 告警条件
    Threshold   float64   `json:"threshold"`    // 阈值
    Duration    int       `json:"duration"`     // 持续时间(秒)
    Severity    string    `json:"severity"`     // 严重级别
    Enabled     int       `json:"enabled"`      // 是否启用
    NotifyUsers string    `json:"notifyUsers"`  // 通知用户列表
    CreateTime  time.Time `json:"createTime"`
    UpdateTime  time.Time `json:"updateTime"`
}
```

### Alert 告警记录模型

```go
type Alert struct {
    ID           uint       `json:"id"`
    RuleID       uint       `json:"ruleId"`
    RuleName     string     `json:"ruleName"`
    MetricName   string     `json:"metricName"`
    CurrentValue float64    `json:"currentValue"`
    Threshold    float64    `json:"threshold"`
    Severity     string     `json:"severity"`
    Status       string     `json:"status"`      // firing/resolved
    Message      string     `json:"message"`
    FiredAt      time.Time  `json:"firedAt"`
    ResolvedAt   *time.Time `json:"resolvedAt"`
    CreateTime   time.Time  `json:"createTime"`
    UpdateTime   time.Time  `json:"updateTime"`
}
```

## 预加载数据

### 预置指标

1. **http_requests_total** - HTTP请求总数 (counter)
2. **cpu_usage_percent** - CPU使用率 (gauge)
3. **memory_usage_bytes** - 内存使用量 (gauge)
4. **response_time_seconds** - 响应时间 (histogram)
5. **error_rate** - 错误率 (gauge)

### 预置告警规则

1. **高CPU使用率告警** - CPU > 80%, critical
2. **内存使用率告警** - Memory > 8GB, warning
3. **高错误率告警** - Error Rate > 5%, critical
4. **慢响应告警** - Response Time > 2s, warning
5. **请求量异常告警** - Requests < 10, info

## 集成说明

### Prometheus集成

Monitor Service可与Prometheus集成，使用Prometheus作为指标数据源：

```go
// 示例：查询Prometheus指标
prometheusURL := "http://prometheus:9090"
query := fmt.Sprintf("%s{%s}", metricName, labels)
result := queryPrometheus(prometheusURL, query)
```

### Grafana集成

可在Grafana中创建Dashboard，使用Monitor Service的指标数据：

1. 添加Prometheus数据源
2. 创建Panel使用指标名称
3. 配置告警通知渠道

### Loki日志查询

```go
type LogQuery struct {
    Name        string    `json:"name"`
    Query       string    `json:"query"`      // LogQL语句
    Description string    `json:"description"`
    Labels      string    `json:"labels"`     // JSON格式标签
    UserID      uint      `json:"userId"`
}
```

### Jaeger链路追踪

```go
type TraceQuery struct {
    Name        string    `json:"name"`
    ServiceName string    `json:"serviceName"`
    Operation   string    `json:"operation"`
    MinDuration int       `json:"minDuration"` // 微秒
    MaxDuration int       `json:"maxDuration"`
    UserID      uint      `json:"userId"`
}
```

## 使用场景

### 场景1：CPU使用率监控

```bash
# 1. 创建CPU指标
POST /api/v1/metrics
{
  "name": "cpu_usage_percent",
  "type": "gauge",
  "unit": "percent"
}

# 2. 创建告警规则
POST /api/v1/alert-rules
{
  "name": "CPU过高告警",
  "metric_name": "cpu_usage_percent",
  "condition": ">",
  "threshold": 80,
  "duration": 300,
  "severity": "critical",
  "notify_users": "ops-team"
}

# 3. 查看触发的告警
GET /api/v1/alerts?status=firing&severity=critical
```

### 场景2：API响应时间监控

```bash
# 1. 创建响应时间指标
POST /api/v1/metrics
{
  "name": "api_response_time",
  "type": "histogram",
  "unit": "seconds"
}

# 2. 设置慢请求告警
POST /api/v1/alert-rules
{
  "name": "API响应慢告警",
  "metric_name": "api_response_time",
  "condition": ">",
  "threshold": 2.0,
  "severity": "warning"
}
```

### 场景3：错误率监控

```bash
# 1. 创建错误率指标
POST /api/v1/metrics
{
  "name": "error_rate",
  "type": "gauge",
  "unit": "percent"
}

# 2. 设置高错误率告警
POST /api/v1/alert-rules
{
  "name": "错误率过高",
  "metric_name": "error_rate",
  "condition": ">",
  "threshold": 5.0,
  "severity": "critical"
}

# 3. 查看告警统计
GET /api/v1/alerts/statistics
```

## 最佳实践

### 1. 指标命名规范

- 使用小写字母和下划线
- 包含单位后缀（如_seconds, _bytes, _total）
- 使用标签区分不同维度

### 2. 告警规则配置

- 设置合理的阈值和持续时间
- 按严重级别分类告警
- 配置准确的通知对象
- 定期review和调整规则

### 3. 告警处理流程

1. 告警触发 → 通知相关人员
2. 查看告警详情 → 分析根本原因
3. 解决问题 → 标记告警为已解决
4. 复盘总结 → 优化告警规则

### 4. 性能优化

- 为常用查询字段创建索引
- 定期清理历史告警记录
- 使用分页查询大数据集
- 启用缓存减少数据库查询

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 服务依赖

- **MySQL** - 数据存储
- **Redis** - 缓存和会话
- **Auth Service** - 身份认证
- **Notification Service** - 告警通知

## 端口信息

- **HTTP端口**: 8090
- **数据库**: monitor_db (MySQL)
- **Redis**: 6379

## 健康检查

```bash
GET /health

Response:
{
  "status": "ok",
  "service": "monitor-service"
}
```
