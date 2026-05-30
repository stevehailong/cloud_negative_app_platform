package model

import "time"

// Cluster 集群表
type Cluster struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ClusterCode    string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"clusterCode"`
	ClusterName    string    `gorm:"type:varchar(128);not null" json:"clusterName"`
	ClusterType    string    `gorm:"type:varchar(32);not null" json:"clusterType"` // kubernetes/docker-swarm
	APIServer      string    `gorm:"type:varchar(255);not null" json:"apiServer"`
	Kubeconfig     string    `gorm:"type:text" json:"kubeconfig"`
	Version        string    `gorm:"type:varchar(64)" json:"version"`
	Region         string    `gorm:"type:varchar(64)" json:"region"`
	Zone           string    `gorm:"type:varchar(64)" json:"zone"`
	Description    string    `gorm:"type:varchar(255)" json:"description"`
	Status         int       `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime     time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime     time.Time `gorm:"column:update_time" json:"updateTime"`
	CreateBy       *uint     `gorm:"column:create_by" json:"createBy"`
	UpdateBy       *uint     `gorm:"column:update_by" json:"updateBy"`
	IsDeleted      int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}

func (Cluster) TableName() string {
	return "clusters"
}

// ClusterNode 集群节点表
type ClusterNode struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ClusterID        uint      `gorm:"column:cluster_id;not null" json:"clusterId"`
	NodeName         string    `gorm:"type:varchar(128);not null" json:"nodeName"`
	NodeIP           string    `gorm:"type:varchar(64);not null" json:"nodeIp"`
	NodeRole         string    `gorm:"type:varchar(32);not null" json:"nodeRole"` // master/worker
	CPUCores         int       `gorm:"default:0" json:"cpuCores"`
	MemoryGB         int       `gorm:"column:memory_gb;default:0" json:"memoryGb"`
	DiskGB           int       `gorm:"column:disk_gb;default:0" json:"diskGb"`
	OSImage          string    `gorm:"type:varchar(128)" json:"osImage"`
	ContainerRuntime string    `gorm:"type:varchar(64)" json:"containerRuntime"`
	KubeletVersion   string    `gorm:"type:varchar(64)" json:"kubeletVersion"`
	Status           int       `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime       time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime       time.Time `gorm:"column:update_time" json:"updateTime"`
	IsDeleted        int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}

func (ClusterNode) TableName() string {
	return "cluster_nodes"
}

// Namespace 命名空间表
type Namespace struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ClusterID         uint      `gorm:"column:cluster_id;not null" json:"clusterId"`
	NamespaceName     string    `gorm:"type:varchar(128);not null" json:"namespaceName"`
	ProjectID         uint      `gorm:"column:project_id;not null" json:"projectId"`
	ResourceQuotaJSON string    `gorm:"type:json" json:"resourceQuotaJson"`
	LimitRangeJSON    string    `gorm:"type:json" json:"limitRangeJson"`
	Description       string    `gorm:"type:varchar(255)" json:"description"`
	Status            int       `gorm:"type:tinyint;default:1" json:"status"`
	CreateTime        time.Time `gorm:"column:create_time" json:"createTime"`
	UpdateTime        time.Time `gorm:"column:update_time" json:"updateTime"`
	CreateBy          *uint     `gorm:"column:create_by" json:"createBy"`
	UpdateBy          *uint     `gorm:"column:update_by" json:"updateBy"`
	IsDeleted         int       `gorm:"type:tinyint;default:0" json:"isDeleted"`
}

func (Namespace) TableName() string {
	return "namespaces"
}
