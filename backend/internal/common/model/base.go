package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createTime"`
	UpdatedAt time.Time      `json:"updateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy string         `gorm:"size:100" json:"createdBy"`
	UpdatedBy string         `gorm:"size:100" json:"updatedBy"`
	Status    int            `gorm:"default:1;comment:状态:1-正常,0-禁用" json:"status"`
}
