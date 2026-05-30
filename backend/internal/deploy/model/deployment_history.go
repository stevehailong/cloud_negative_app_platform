package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// DeploymentHistory 部署历史记录表
type DeploymentHistory struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	AppDeploymentID int64          `gorm:"column:app_deployment_id;not null;index" json:"app_deployment_id"`
	ReleaseID       *int64         `gorm:"column:release_id;index" json:"release_id"`
	Version         string         `gorm:"column:version;size:255" json:"version"`
	ImageURL        string         `gorm:"column:image_url;size:500" json:"image_url"`
	Replicas        int            `gorm:"column:replicas" json:"replicas"`
	DeploymentType  string         `gorm:"column:deployment_type;size:50" json:"deployment_type"` // create, update, rollback, restart, scale
	OperatorUserID  *int64         `gorm:"column:operator_user_id;index" json:"operator_user_id"`
	StartTime       *time.Time     `gorm:"column:start_time" json:"start_time"`
	EndTime         *time.Time     `gorm:"column:end_time" json:"end_time"`
	Duration        *int           `gorm:"column:duration" json:"duration"` // 耗时(秒)
	Status          string         `gorm:"column:status;size:50" json:"status"` // success, failed, progressing
	FailureReason   string         `gorm:"column:failure_reason;type:text" json:"failure_reason"`
	Changes         JSONMap        `gorm:"column:changes;type:json" json:"changes"`
	CreateTime      time.Time      `gorm:"column:create_time;autoCreateTime;index" json:"create_time"`
}

func (DeploymentHistory) TableName() string {
	return "deployment_history"
}

// JSONMap 用于处理JSON字段
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
