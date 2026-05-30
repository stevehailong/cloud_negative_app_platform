package model

import "time"

// ConfigMap ConfigMap配置模型
type ConfigMap struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(128);not null" json:"name"`
	EnvID        uint      `gorm:"column:env_id;not null;index" json:"envId"`
	Namespace    string    `gorm:"type:varchar(128);not null;index" json:"namespace"`
	Data         string    `gorm:"type:json;not null" json:"data"`
	Description  string    `gorm:"type:varchar(255)" json:"description"`
	SyncStatus   string    `gorm:"type:varchar(32);default:'pending';index" json:"syncStatus"`
	SyncMessage  string    `gorm:"type:text" json:"syncMessage"`
	LastSyncTime *time.Time `gorm:"column:last_sync_time" json:"lastSyncTime"`
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:update_time" json:"updateTime"`
	CreateBy     *uint     `gorm:"column:create_by" json:"createBy"`
	UpdateBy     *uint     `gorm:"column:update_by" json:"updateBy"`
	IsDeleted    int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}

func (ConfigMap) TableName() string {
	return "config_maps"
}

// Secret Secret密钥模型
type Secret struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(128);not null" json:"name"`
	EnvID        uint      `gorm:"column:env_id;not null;index" json:"envId"`
	Namespace    string    `gorm:"type:varchar(128);not null;index" json:"namespace"`
	SecretType   string    `gorm:"type:varchar(32);default:'Opaque';index" json:"secretType"`
	Data         string    `gorm:"type:json;not null" json:"data"`
	Description  string    `gorm:"type:varchar(255)" json:"description"`
	SyncStatus   string    `gorm:"type:varchar(32);default:'pending';index" json:"syncStatus"`
	SyncMessage  string    `gorm:"type:text" json:"syncMessage"`
	LastSyncTime *time.Time `gorm:"column:last_sync_time" json:"lastSyncTime"`
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:update_time" json:"updateTime"`
	CreateBy     *uint     `gorm:"column:create_by" json:"createBy"`
	UpdateBy     *uint     `gorm:"column:update_by" json:"updateBy"`
	IsDeleted    int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}

func (Secret) TableName() string {
	return "secrets"
}
