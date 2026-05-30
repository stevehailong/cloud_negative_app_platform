package model

import "time"

// Deployment 部署记录
type Deployment struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	ReleaseID         uint       `gorm:"not null;index:idx_release_id" json:"releaseId"`
	ClusterID         uint       `gorm:"not null;index:idx_cluster_id" json:"clusterId"`
	Namespace         string     `gorm:"size:128;not null;index:idx_namespace" json:"namespace"`
	WorkloadName      string     `gorm:"size:128;not null" json:"workloadName"`
	WorkloadType      string     `gorm:"size:32;not null" json:"workloadType"` // deployment/statefulset/job
	ImageVersion      string     `gorm:"size:128;not null" json:"imageVersion"`
	DesiredReplicas   int        `gorm:"default:1" json:"desiredReplicas"`
	AvailableReplicas int        `gorm:"default:0" json:"availableReplicas"`
	DeploymentStatus  string     `gorm:"size:32;not null" json:"deploymentStatus"` // progressing/success/failed/rollback
	FailureReason     string     `gorm:"size:512" json:"failureReason"`
	StartTime         *time.Time `json:"startTime,omitempty"`
	EndTime           *time.Time `json:"endTime,omitempty"`
	CreateTime        time.Time  `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime        time.Time  `gorm:"autoUpdateTime" json:"updateTime"`
}

func (Deployment) TableName() string {
	return "deployments"
}
