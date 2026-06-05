package model

import "time"

// ResourceQuota 资源配额
type ResourceQuota struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ScopeType    string    `gorm:"size:32;not null" json:"scopeType"` // tenant/project/env/namespace/app
	ScopeID      uint      `gorm:"not null" json:"scopeId"`
	CPULimit     string    `gorm:"size:32" json:"cpuLimit"`
	MemoryLimit  string    `gorm:"size:32" json:"memoryLimit"`
	StorageLimit string    `gorm:"size:32" json:"storageLimit"`
	PodLimit     int       `gorm:"default:0" json:"podLimit"`
	ServiceLimit int       `gorm:"default:0" json:"serviceLimit"`
	LBLimit      int       `gorm:"default:0" json:"lbLimit"`
	GPULimit     int       `gorm:"default:0" json:"gpuLimit"`
	Status       int       `gorm:"default:1" json:"status"`
	IsDeleted    int       `gorm:"column:is_deleted;default:0" json:"isDeleted"`
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:update_time" json:"updateTime"`
}

func (ResourceQuota) TableName() string { return "resource_quotas" }
