package model

import "time"

// Notification 通知记录
type Notification struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Title        string    `gorm:"size:128;not null" json:"title"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	NotifyType   string    `gorm:"size:32;not null;index:idx_notify_type" json:"notifyType"` // system/release/deploy/alert
	Channel      string    `gorm:"size:32;not null" json:"channel"`                           // email/sms/dingtalk/slack/webhook
	Status       string    `gorm:"size:32;not null;default:'pending'" json:"status"`          // pending/sent/failed
	ReceiverType string    `gorm:"size:32;not null" json:"receiverType"`                      // user/role/group
	ReceiverIDs  string    `gorm:"type:text" json:"receiverIds"`                              // 逗号分隔的ID列表
	TemplateID   *uint     `json:"templateId,omitempty"`
	Params       string    `gorm:"type:json" json:"params"`           // 模板参数
	ErrorMsg     string    `gorm:"size:500" json:"errorMsg"`
	SentAt       *time.Time `json:"sentAt,omitempty"`
	CreateTime   time.Time `gorm:"autoCreateTime;index:idx_create_time" json:"createTime"`
}

func (Notification) TableName() string {
	return "notifications"
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TemplateCode string   `gorm:"size:64;uniqueIndex;not null" json:"templateCode"`
	TemplateName string   `gorm:"size:128;not null" json:"templateName"`
	NotifyType   string   `gorm:"size:32;not null" json:"notifyType"`
	Channel      string   `gorm:"size:32;not null" json:"channel"`
	Title        string   `gorm:"size:255" json:"title"`           // 标题模板（支持变量）
	Content      string   `gorm:"type:text;not null" json:"content"` // 内容模板（支持变量）
	Variables    string   `gorm:"type:json" json:"variables"`      // 可用变量说明
	Enabled      int      `gorm:"default:1" json:"enabled"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime   time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// NotificationChannel 通知渠道配置
type NotificationChannel struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ChannelCode string    `gorm:"size:64;uniqueIndex;not null" json:"channelCode"`
	ChannelName string    `gorm:"size:128;not null" json:"channelName"`
	ChannelType string    `gorm:"size:32;not null" json:"channelType"` // email/sms/dingtalk/slack/webhook
	Config      string    `gorm:"type:json;not null" json:"config"`    // 渠道配置（SMTP、API Key等）
	Enabled     int       `gorm:"default:1" json:"enabled"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (NotificationChannel) TableName() string {
	return "notification_channels"
}
