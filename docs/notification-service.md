# Notification Service 使用文档

## 概述

Notification Service是my-cloud平台的通知服务，支持多渠道通知发送（邮件、短信、钉钉、Slack、Webhook），提供模板化通知管理功能。

## 功能特性

- **多渠道支持**: Email、SMS、DingTalk、Slack、Webhook
- **模板管理**: 支持通知模板的创建、更新、删除，支持变量替换
- **渠道配置**: 灵活配置不同通知渠道的参数
- **异步发送**: 通知异步发送，不阻塞业务流程
- **状态跟踪**: 记录通知发送状态（pending/sent/failed）
- **历史记录**: 完整的通知发送历史记录

## 数据库表结构

### notifications 通知记录表
- id: 通知ID
- title: 通知标题
- content: 通知内容
- notify_type: 通知类型(release/deploy/pipeline/system)
- channel: 通知渠道(email/sms/dingtalk/slack/webhook)
- status: 发送状态(pending/sent/failed)
- receiver_type: 接收者类型(user/role/group)
- receiver_ids: 接收者ID列表
- template_id: 关联模板ID
- params: 模板参数(JSON)
- error_msg: 错误信息
- sent_at: 发送时间
- created_at/updated_at: 创建/更新时间

### notification_templates 通知模板表
- id: 模板ID
- template_code: 模板编码(唯一)
- template_name: 模板名称
- notify_type: 通知类型
- channel: 通知渠道
- title: 标题模板
- content: 内容模板
- variables: 模板变量(JSON)
- enabled: 是否启用
- created_at/updated_at: 创建/更新时间

### notification_channels 通知渠道配置表
- id: 渠道ID
- channel_code: 渠道编码(唯一)
- channel_name: 渠道名称
- channel_type: 渠道类型(email/sms/dingtalk/slack/webhook)
- config: 渠道配置(JSON)
- enabled: 是否启用
- created_at/updated_at: 创建/更新时间

## API接口

### 1. 通知管理

#### 发送通知
```
POST /api/v1/notifications
Content-Type: application/json
Authorization: Bearer {token}

{
  "title": "发布成功通知",
  "content": "项目my-project版本v1.0.0发布成功",
  "notifyType": "release",
  "channel": "dingtalk",
  "receiverType": "user",
  "receiverIds": "1,2,3"
}
```

#### 通过模板发送通知
```
POST /api/v1/notifications/template
Content-Type: application/json
Authorization: Bearer {token}

{
  "templateCode": "RELEASE_SUCCESS",
  "params": {
    "projectName": "my-project",
    "version": "v1.0.0",
    "environment": "production",
    "operator": "张三",
    "releaseTime": "2026-05-28 14:00:00"
  },
  "receiverType": "user",
  "receiverIds": [1, 2, 3]
}
```

#### 获取通知列表
```
GET /api/v1/notifications?page=1&pageSize=10&notifyType=release&status=sent
Authorization: Bearer {token}
```

#### 获取通知详情
```
GET /api/v1/notifications/{id}
Authorization: Bearer {token}
```

### 2. 模板管理

#### 创建模板
```
POST /api/v1/notification-templates
Content-Type: application/json
Authorization: Bearer {token}

{
  "templateCode": "CUSTOM_NOTIFY",
  "templateName": "自定义通知",
  "notifyType": "system",
  "channel": "email",
  "title": "【系统通知】{{title}}",
  "content": "尊敬的用户{{userName}}，\n\n{{content}}\n\n时间：{{time}}",
  "variables": "[\"title\",\"userName\",\"content\",\"time\"]"
}
```

#### 获取模板列表
```
GET /api/v1/notification-templates?page=1&pageSize=10&notifyType=release
Authorization: Bearer {token}
```

#### 获取模板详情
```
GET /api/v1/notification-templates/{id}
Authorization: Bearer {token}
```

#### 更新模板
```
PUT /api/v1/notification-templates/{id}
Content-Type: application/json
Authorization: Bearer {token}

{
  "templateName": "更新后的模板名称",
  "content": "更新后的内容模板",
  "enabled": 1
}
```

#### 删除模板
```
DELETE /api/v1/notification-templates/{id}
Authorization: Bearer {token}
```

### 3. 渠道管理

#### 创建渠道
```
POST /api/v1/notification-channels
Content-Type: application/json
Authorization: Bearer {token}

{
  "channelCode": "DINGTALK_DEV",
  "channelName": "钉钉开发环境",
  "channelType": "dingtalk",
  "config": "{\"webhook\":\"https://oapi.dingtalk.com/robot/send?access_token=xxx\",\"secret\":\"xxx\"}"
}
```

#### 获取渠道列表
```
GET /api/v1/notification-channels
Authorization: Bearer {token}
```

#### 更新渠道
```
PUT /api/v1/notification-channels/{id}
Content-Type: application/json
Authorization: Bearer {token}

{
  "channelName": "更新后的渠道名称",
  "config": "{\"webhook\":\"https://new-webhook-url\"}",
  "enabled": 1
}
```

#### 删除渠道
```
DELETE /api/v1/notification-channels/{id}
Authorization: Bearer {token}
```

## 模板变量语法

模板内容支持变量替换，使用 `{{变量名}}` 语法：

```
标题: 【发布通知】{{projectName}}
内容: 
项目: {{projectName}}
版本: {{version}}
环境: {{environment}}
操作人: {{operator}}
时间: {{releaseTime}}
```

发送时传入参数：
```json
{
  "projectName": "my-project",
  "version": "v1.0.0",
  "environment": "production",
  "operator": "张三",
  "releaseTime": "2026-05-28 14:00:00"
}
```

## 预置模板

系统已预置以下通知模板：

1. **RELEASE_SUCCESS** - 发布成功通知
2. **RELEASE_FAILED** - 发布失败通知
3. **PIPELINE_SUCCESS** - 流水线成功通知
4. **PIPELINE_FAILED** - 流水线失败通知
5. **DEPLOY_SUCCESS** - 部署成功通知
6. **DEPLOY_FAILED** - 部署失败通知

## 预置渠道

系统已预置以下通知渠道配置（需要更新配置后启用）：

1. **DINGTALK_DEFAULT** - 钉钉默认渠道
2. **EMAIL_DEFAULT** - 邮件默认渠道
3. **SLACK_DEFAULT** - Slack默认渠道（默认禁用）
4. **WEBHOOK_DEFAULT** - Webhook默认渠道（默认禁用）

## 渠道配置示例

### 钉钉 (DingTalk)
```json
{
  "webhook": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN",
  "secret": "YOUR_SECRET"
}
```

### 邮件 (Email)
```json
{
  "smtp_host": "smtp.example.com",
  "smtp_port": "465",
  "smtp_user": "noreply@example.com",
  "smtp_pass": "YOUR_PASSWORD",
  "from": "noreply@example.com"
}
```

### Slack
```json
{
  "webhook": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
}
```

### Webhook
```json
{
  "url": "https://your-webhook-endpoint.com/notifications",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer YOUR_TOKEN"
  }
}
```

## 使用示例

### 在Release Service中集成通知

```go
// 发布成功后发送通知
func (s *ReleaseService) notifyReleaseSuccess(release *model.Release) {
    notificationClient := &http.Client{}
    
    payload := map[string]interface{}{
        "templateCode": "RELEASE_SUCCESS",
        "params": map[string]interface{}{
            "projectName": release.ProjectName,
            "version":     release.Version,
            "environment": release.Environment,
            "operator":    release.CreatedBy,
            "releaseTime": release.CreatedAt.Format("2006-01-02 15:04:05"),
        },
        "receiverType": "user",
        "receiverIds":  []uint{release.CreatedBy},
    }
    
    jsonData, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", "http://notification-service:8095/api/v1/notifications/template", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    
    notificationClient.Do(req)
}
```

## 部署说明

### Docker部署

服务已在docker-compose.yml中配置：

```yaml
notification-service:
  ports:
    - "8095:8095"
  environment:
    - SERVER_PORT=8095
    - DB_DSN=root:root123456@tcp(mysql:3306)/notification_db?charset=utf8mb4&parseTime=True&loc=Local
```

### 数据库初始化

服务启动时会自动创建数据表，也可以手动执行SQL脚本：

```bash
mysql -u root -p < sql/10-notification-db.sql
```

## 监控指标

- 通知发送总数
- 通知发送成功率
- 各渠道发送量统计
- 发送延迟统计
- 失败重试次数

## 注意事项

1. 通知采用异步发送，不会阻塞主业务流程
2. 发送失败会记录错误信息和重试次数
3. 模板变量必须在params中提供，否则会保留原始占位符
4. 渠道配置中的敏感信息（如token、密码）应使用secret管理
5. 建议为不同环境配置独立的通知渠道

## 后续优化

- [ ] 实现真实的渠道发送逻辑（目前为模拟发送）
- [ ] 添加发送失败自动重试机制
- [ ] 支持发送频率限制
- [ ] 支持通知优先级
- [ ] 添加通知订阅功能
- [ ] 集成更多通知渠道（企业微信、飞书等）
- [ ] 支持富文本和markdown格式
- [ ] 添加通知统计分析功能
