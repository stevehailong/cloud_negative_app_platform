package model

import "time"

// Pipeline 流水线表
type Pipeline struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AppID      uint      `gorm:"not null;index" json:"appId"`
	Code       string    `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Name       string    `gorm:"size:128;not null" json:"name"`
	Type       string    `gorm:"size:32;not null" json:"type"` // ci/cd/full
	CITool     string    `gorm:"size:32;default:jenkins" json:"ciTool"`
	ConfigJSON string    `gorm:"type:json" json:"configJson"`
	Enabled    int       `gorm:"default:1" json:"enabled"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (Pipeline) TableName() string {
	return "pipelines"
}

// PipelineRun 流水线运行记录表
type PipelineRun struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	PipelineID      uint       `gorm:"not null;index" json:"pipelineId"`
	RunNo           string     `gorm:"size:64;not null;uniqueIndex" json:"runNo"`
	TriggerType     string     `gorm:"size:32" json:"triggerType"` // manual/webhook/mr/schedule
	GitCommit       string     `gorm:"size:64" json:"gitCommit"`
	GitBranch       string     `gorm:"size:64;index" json:"gitBranch"`
	Status          string     `gorm:"size:32;not null;index" json:"status"` // pending/running/success/failed/cancelled
	StartTime       *time.Time `json:"startTime,omitempty"`
	EndTime         *time.Time `json:"endTime,omitempty"`
	DurationSeconds int        `gorm:"default:0" json:"durationSeconds"`
	OperatorUserID  *uint      `json:"operatorUserId,omitempty"`
	LogURL          string     `gorm:"size:255" json:"logUrl"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (PipelineRun) TableName() string {
	return "pipeline_runs"
}

// Artifact 制品表
type Artifact struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PipelineRunID uint      `gorm:"not null;index" json:"pipelineRunId"`
	Type          string    `gorm:"size:32;not null;index" json:"type"` // image/chart/package/sbom/report
	Name          string    `gorm:"size:128;not null" json:"name"`
	Version       string    `gorm:"size:64;not null" json:"version"`
	RepoURL       string    `gorm:"size:255;not null" json:"repoUrl"`
	Digest        string    `gorm:"size:255" json:"digest"`
	MetadataJSON  string    `gorm:"type:json" json:"metadataJson"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (Artifact) TableName() string {
	return "artifacts"
}
