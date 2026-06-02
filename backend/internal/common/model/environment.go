package model

import "time"

// Environment 环境表
type Environment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EnvCode     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"envCode"`
	EnvName     string    `gorm:"type:varchar(128);not null" json:"envName"`
	EnvType     string    `gorm:"type:varchar(32);not null" json:"envType"` // dev/test/staging/prod/preview
	ClusterID   uint      `gorm:"column:cluster_id;not null" json:"clusterId"`
	Namespace   string    `gorm:"type:varchar(128);not null" json:"namespace"`
	ProjectID   uint      `gorm:"column:project_id;not null" json:"projectId"`
	TemplateID  *uint     `gorm:"column:template_id" json:"templateId"` // 关联的环境模板ID
	Description string    `gorm:"type:varchar(255)" json:"description"`
	ConfigJSON  string    `gorm:"type:json" json:"configJson"`
	Status      int       `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	CreateBy    *uint     `gorm:"column:create_by" json:"createBy"`
	UpdateBy    *uint     `gorm:"column:update_by" json:"updateBy"`
	IsDeleted   int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}

func (Environment) TableName() string {
	return "environments"
}

// EnvTemplate 环境模板表
type EnvTemplate struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TemplateCode string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"templateCode"`
	TemplateName string    `gorm:"type:varchar(128);not null" json:"templateName"`
	TemplateType string    `gorm:"type:varchar(32);not null" json:"templateType"` // helm/kustomize/yaml
	RepoURL      string    `gorm:"type:varchar(255)" json:"repoUrl"`
	ChartName    string    `gorm:"type:varchar(128)" json:"chartName"`
	ChartVersion string    `gorm:"type:varchar(64)" json:"chartVersion"`
	ValuesYAML   string    `gorm:"type:text" json:"valuesYaml"`
	Description  string    `gorm:"type:varchar(255)" json:"description"`
	Status       int       `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime   time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:update_time" json:"updateTime"`
	CreateBy     *uint     `gorm:"column:create_by" json:"createBy"`
	UpdateBy     *uint     `gorm:"column:update_by" json:"updateBy"`
	IsDeleted    int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}

func (EnvTemplate) TableName() string {
	return "env_templates"
}

// AppEnvBinding 应用环境绑定表
type AppEnvBinding struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AppID         uint      `gorm:"column:app_id;not null" json:"appId"`
	EnvID         uint      `gorm:"column:env_id;not null" json:"envId"`
	TemplateID    *uint     `gorm:"column:template_id" json:"templateId"`
	Replicas      int       `gorm:"default:1" json:"replicas"`
	CPURequest    string    `gorm:"type:varchar(32);default:'100m'" json:"cpuRequest"`
	CPULimit      string    `gorm:"type:varchar(32);default:'500m'" json:"cpuLimit"`
	MemoryRequest string    `gorm:"type:varchar(32);default:'128Mi'" json:"memoryRequest"`
	MemoryLimit   string    `gorm:"type:varchar(32);default:'512Mi'" json:"memoryLimit"`
	ConfigJSON    string    `gorm:"type:json" json:"configJson"`
	Status        int       `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime    time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime    time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	CreateBy      *uint     `gorm:"column:create_by" json:"createBy"`
	UpdateBy      *uint     `gorm:"column:update_by" json:"updateBy"`
	IsDeleted     int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}

func (AppEnvBinding) TableName() string {
	return "app_env_bindings"
}
