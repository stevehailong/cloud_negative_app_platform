package model

import "time"

// CostRecord 成本记录
type CostRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ClusterID    uint      `gorm:"not null;uniqueIndex:idx_unique_cost;index" json:"clusterId"`
	TenantID     *uint     `gorm:"index" json:"tenantId"`
	ProjectID    *uint     `gorm:"index" json:"projectId"`
	AppID        *uint     `gorm:"index" json:"appId"`
	EnvID        *uint     `gorm:"index" json:"envId"`
	Namespace    string    `gorm:"size:128;uniqueIndex:idx_unique_cost" json:"namespace"`
	CostDate     string    `gorm:"type:date;not null;uniqueIndex:idx_unique_cost" json:"costDate"`
	CPUCost      float64   `gorm:"type:decimal(18,4);default:0" json:"cpuCost"`
	MemoryCost   float64   `gorm:"type:decimal(18,4);default:0" json:"memoryCost"`
	StorageCost  float64   `gorm:"type:decimal(18,4);default:0" json:"storageCost"`
	NetworkCost  float64   `gorm:"type:decimal(18,4);default:0" json:"networkCost"`
	TotalCost    float64   `gorm:"type:decimal(18,4);default:0" json:"totalCost"`
	Source       string    `gorm:"size:32;default:kubecost;uniqueIndex:idx_unique_cost" json:"source"`
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`
	AppName      string    `gorm:"-" json:"appName"`
	ProjectName  string    `gorm:"-" json:"projectName"`
}

func (CostRecord) TableName() string { return "cost_records" }
