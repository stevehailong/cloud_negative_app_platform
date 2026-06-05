package model

import "time"

// AppSecret 应用密钥
type AppSecret struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AppID       uint      `gorm:"not null;index" json:"appId"`
	EnvID       uint      `gorm:"not null;index" json:"envId"`
	SecretKey   string    `gorm:"size:128;not null" json:"secretKey"`
	VaultPath   string    `gorm:"size:255;not null" json:"vaultPath"`
	Description string    `gorm:"size:255" json:"description"`
	CreateTime  time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (AppSecret) TableName() string { return "app_secrets" }
