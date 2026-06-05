package model

import "time"

// AppConfig 应用配置
type AppConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AppID       uint      `gorm:"not null;index" json:"appId"`
	EnvID       uint      `gorm:"not null;index" json:"envId"`
	ConfigKey   string    `gorm:"size:128;not null" json:"configKey"`
	ConfigValue string    `gorm:"type:text" json:"configValue"`
	ValueType   string    `gorm:"size:32;default:string" json:"valueType"`
	Version     string    `gorm:"size:64" json:"version"`
	Description string    `gorm:"size:255" json:"description"`
	CreateTime  time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (AppConfig) TableName() string { return "app_configs" }
