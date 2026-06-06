package model

import "time"

// AppDeployment 应用部署主记录表
// 每个应用在每个环境只有一个namespace,但可以有多个workload(stable和canary)
// 通过workload_name区分: app-{AppID} (stable) 和 app-{AppID}-canary (canary)
// 唯一约束: (namespace, workload_name) 组合唯一
type AppDeployment struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AppID             int64      `gorm:"column:app_id;not null;index:idx_app_env" json:"app_id"`
	EnvID             int64      `gorm:"column:env_id;not null;index:idx_app_env" json:"env_id"`
	ClusterID         int64      `gorm:"column:cluster_id;not null" json:"cluster_id"`
	Namespace         string     `gorm:"column:namespace;size:255;not null;uniqueIndex:uk_namespace_workload" json:"namespace"`
	WorkloadName      string     `gorm:"column:workload_name;size:255;not null;uniqueIndex:uk_namespace_workload" json:"workload_name"`
	WorkloadType      string     `gorm:"column:workload_type;size:50;default:deployment" json:"workload_type"`
	CurrentVersion    string     `gorm:"column:current_version;size:255" json:"current_version"`
	CurrentImage      string     `gorm:"column:current_image;size:500" json:"current_image"`
	DesiredReplicas   int        `gorm:"column:desired_replicas;default:1" json:"desired_replicas"`
	AvailableReplicas int        `gorm:"column:available_replicas;default:0" json:"available_replicas"`
	DeploymentStatus  string     `gorm:"column:deployment_status;size:50" json:"deployment_status"`
	LastDeployID      *int64     `gorm:"column:last_deploy_id" json:"last_deploy_id"`
	LastDeployTime    *time.Time `gorm:"column:last_deploy_time" json:"last_deploy_time"`
	LastDeployUserID  *int64     `gorm:"column:last_deploy_user_id" json:"last_deploy_user_id"`
	CreateTime        time.Time  `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime        time.Time  `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
	AppName           string     `gorm:"-" json:"app_name"`          // 非持久化字段，关联填充
	EnvName           string     `gorm:"-" json:"env_name"`          // 非持久化字段，关联填充
	IsDeploying       bool       `gorm:"-" json:"is_deploying"`      // 是否有部署正在进行中
	CanaryWeight      int        `gorm:"-" json:"canary_weight"`     // 金丝雀权重 (0-100)，从 Ingress 注解读取
	DeployStrategy    string     `gorm:"column:deploy_strategy;size:32" json:"deploy_strategy"` // 部署策略: rolling/canary/bluegreen
	OperatorName      string     `gorm:"-" json:"operator_name"`     // 最后操作人姓名
}

func (AppDeployment) TableName() string {
	return "app_deployments"
}
