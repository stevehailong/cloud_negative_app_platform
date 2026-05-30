package model

import "time"

// Pipeline 流水线
type Pipeline struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	PipelineCode string    `gorm:"size:64;uniqueIndex;not null" json:"pipelineCode"`
	AppID        uint      `gorm:"not null;index:idx_app_id" json:"appId"`
	PipelineName string    `gorm:"size:128;not null" json:"pipelineName"`
	PipelineType string    `gorm:"size:32;not null" json:"pipelineType"` // ci/cd/full
	CITool       string    `gorm:"size:32;not null;default:'jenkins'" json:"ciTool"`
	ConfigJSON   string    `gorm:"type:json" json:"configJson"`
	Enabled      int       `gorm:"default:1" json:"enabled"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime   time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (Pipeline) TableName() string {
	return "pipelines"
}

// PipelineRun 流水线执行记录
type PipelineRun struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	PipelineID      uint       `gorm:"not null;index:idx_pipeline_id" json:"pipelineId"`
	RunNo           string     `gorm:"size:64;uniqueIndex;not null" json:"runNo"`
	TriggerType     string     `gorm:"size:32;not null" json:"triggerType"` // manual/webhook/mr/schedule
	GitCommit       string     `gorm:"size:64" json:"gitCommit"`
	GitBranch       string     `gorm:"size:64;index:idx_git_branch" json:"gitBranch"`
	Status          string     `gorm:"size:32;not null;index:idx_status" json:"status"` // pending/running/success/failed/cancelled
	StartTime       *time.Time `json:"startTime,omitempty"`
	EndTime         *time.Time `json:"endTime,omitempty"`
	DurationSeconds int        `json:"durationSeconds"`
	OperatorUserID  uint       `json:"operatorUserId"`
	LogURL          string     `gorm:"size:255" json:"logUrl"`
	CreateTime      time.Time  `gorm:"autoCreateTime" json:"createTime"`
}

func (PipelineRun) TableName() string {
	return "pipeline_runs"
}

// Artifact 制品
type Artifact struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PipelineRunID uint      `gorm:"not null;index:idx_pipeline_run_id" json:"pipelineRunId"`
	ArtifactType  string    `gorm:"size:32;not null;index:idx_artifact_type" json:"artifactType"` // image/chart/package/sbom/report
	ArtifactName  string    `gorm:"size:128;not null" json:"artifactName"`
	ArtifactVersion string  `gorm:"size:64" json:"artifactVersion"`
	RepoURL       string    `gorm:"size:255" json:"repoUrl"`
	Digest        string    `gorm:"size:255" json:"digest"`
	MetadataJSON  string    `gorm:"type:json" json:"metadataJson"`
	CreateTime    time.Time `gorm:"autoCreateTime" json:"createTime"`
}

func (Artifact) TableName() string {
	return "artifacts"
}
