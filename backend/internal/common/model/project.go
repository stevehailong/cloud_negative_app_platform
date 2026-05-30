package model

import (
	"time"
)

// Tenant 租户模型
type Tenant struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantCode    string    `gorm:"type:varchar(64);uniqueIndex;not null;comment:租户编码" json:"tenantCode"`
	TenantName    string    `gorm:"type:varchar(128);not null;comment:租户名称" json:"tenantName"`
	ContactEmail  string    `gorm:"type:varchar(128);comment:联系邮箱" json:"contactEmail"`
	ContactPhone  string    `gorm:"type:varchar(32);comment:联系电话" json:"contactPhone"`
	Status        int       `gorm:"type:tinyint;default:1;comment:状态:1-启用,0-禁用" json:"status"`
	CreateTime    time.Time `gorm:"column:create_time;comment:创建时间" json:"createTime"`
	UpdateTime    time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
	CreateBy      *uint     `gorm:"column:create_by;comment:创建人" json:"createBy,omitempty"`
	UpdateBy      *uint     `gorm:"column:update_by;comment:更新人" json:"updateBy,omitempty"`
	IsDeleted     int       `gorm:"type:tinyint;default:0;comment:是否删除" json:"isDeleted"`
}

func (Tenant) TableName() string {
	return "tenants"
}

// Organization 组织模型
type Organization struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    uint      `gorm:"column:tenant_id;not null;comment:租户ID" json:"tenantId"`
	ParentID    *uint     `gorm:"column:parent_id;comment:父组织ID" json:"parentId,omitempty"`
	OrgCode     string    `gorm:"type:varchar(64);not null;comment:组织编码" json:"orgCode"`
	OrgName     string    `gorm:"type:varchar(128);not null;comment:组织名称" json:"orgName"`
	OrgLevel    int       `gorm:"column:org_level;default:0;comment:组织层级" json:"orgLevel"`
	OrgPath     string    `gorm:"type:varchar(512);default:'/';comment:组织路径" json:"orgPath"`
	Description string    `gorm:"type:varchar(255);comment:描述" json:"description"`
	Status      int       `gorm:"type:tinyint;default:1;comment:状态:1-启用,0-禁用" json:"status"`
	CreateTime  time.Time `gorm:"column:create_time;comment:创建时间" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
	CreateBy    *uint     `gorm:"column:create_by;comment:创建人" json:"createBy,omitempty"`
	UpdateBy    *uint     `gorm:"column:update_by;comment:更新人" json:"updateBy,omitempty"`
	IsDeleted   int       `gorm:"type:tinyint;default:0;comment:是否删除" json:"isDeleted"`
}

func (Organization) TableName() string {
	return "organizations"
}

// Project 项目模型
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    uint      `gorm:"column:tenant_id;not null;comment:租户ID" json:"tenantId"`
	OrgID       *uint     `gorm:"column:org_id;comment:组织ID" json:"orgId,omitempty"`
	ProjectCode string    `gorm:"type:varchar(64);uniqueIndex;not null;comment:项目编码" json:"projectCode"`
	ProjectName string    `gorm:"type:varchar(128);not null;comment:项目名称" json:"projectName"`
	OwnerUserID *uint     `gorm:"column:owner_user_id;comment:负责人" json:"ownerUserId,omitempty"`
	Description string    `gorm:"type:varchar(255);comment:描述" json:"description"`
	Visibility  string    `gorm:"type:varchar(32);default:'private';comment:可见性" json:"visibility"`
	Status      int       `gorm:"type:tinyint;default:1;comment:状态:1-启用,0-禁用" json:"status"`
	CreateTime  time.Time `gorm:"column:create_time;comment:创建时间" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
	CreateBy    *uint     `gorm:"column:create_by;comment:创建人" json:"createBy,omitempty"`
	UpdateBy    *uint     `gorm:"column:update_by;comment:更新人" json:"updateBy,omitempty"`
	IsDeleted   int       `gorm:"type:tinyint;default:0;comment:是否删除" json:"isDeleted"`
}

func (Project) TableName() string {
	return "projects"
}

// ProjectMember 项目成员模型
type ProjectMember struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProjectID  uint      `gorm:"column:project_id;not null;comment:项目ID" json:"projectId"`
	UserID     uint      `gorm:"column:user_id;not null;comment:用户ID" json:"userId"`
	RoleCode   string    `gorm:"type:varchar(64);not null;comment:项目角色" json:"roleCode"`
	CreateTime time.Time `gorm:"column:create_time;comment:创建时间" json:"createTime"`
	CreateBy   *uint     `gorm:"column:create_by;comment:创建人" json:"createBy,omitempty"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}
