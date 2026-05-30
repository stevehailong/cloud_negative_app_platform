package model

import "time"

// Release 发布工单
type Release struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	ReleaseNo       string     `gorm:"size:64;uniqueIndex;not null" json:"releaseNo"`
	AppID           uint       `gorm:"not null;index:idx_app_id" json:"appId"`
	EnvID           uint       `gorm:"not null;index:idx_env_id" json:"envId"`
	PipelineRunID   *uint      `json:"pipelineRunId,omitempty"`
	ReleaseVersion  string     `gorm:"size:64;not null" json:"releaseVersion"`
	ReleaseStrategy string     `gorm:"size:32;not null" json:"releaseStrategy"` // rolling/bluegreen/canary
	ImageURL        string     `gorm:"size:256" json:"imageUrl"`                // 要发布的镜像地址
	ClusterID       uint       `gorm:"default:1" json:"clusterId"`
	Namespace       string     `gorm:"size:128" json:"namespace"`
	CanaryPercent   int        `gorm:"default:20" json:"canaryPercent"`  // 金丝雀流量百分比
	CanaryStatus    string     `gorm:"size:32" json:"canaryStatus"`     // canary_running/canary_confirmed/canary_rollback
	ApprovalStatus  string     `gorm:"size:32;not null;default:'pending'" json:"approvalStatus"` // pending/approved/rejected
	ReleaseStatus   string     `gorm:"size:32;not null;default:'created';index:idx_release_status" json:"releaseStatus"` // created/submitted/approved/rejected/executing/canary/success/failed/rollback
	OperatorUserID  uint       `json:"operatorUserId"`
	Description     string     `gorm:"size:255" json:"description"`
	CreateTime      time.Time  `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime      time.Time  `gorm:"autoUpdateTime" json:"updateTime"`
}

func (Release) TableName() string {
	return "releases"
}

// ReleaseApproval 发布审批
type ReleaseApproval struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ReleaseID      uint       `gorm:"not null;index:idx_release_id" json:"releaseId"`
	ApproverUserID uint       `gorm:"not null;index:idx_approver_user_id" json:"approverUserId"`
	ApprovalStatus string     `gorm:"size:32;not null" json:"approvalStatus"` // pending/approved/rejected
	CommentText    string     `gorm:"size:255" json:"commentText"`
	ApprovalTime   *time.Time `json:"approvalTime,omitempty"`
	CreateTime     time.Time  `gorm:"autoCreateTime" json:"createTime"`
}

func (ReleaseApproval) TableName() string {
	return "release_approvals"
}
