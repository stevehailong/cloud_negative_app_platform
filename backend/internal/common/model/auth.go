package model

import (
	"time"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	BaseModel
	Username    string `gorm:"size:100;uniqueIndex;not null;comment:用户名" json:"username"`
	Password    string `gorm:"size:255;not null;comment:密码" json:"-"`
	Email       string `gorm:"size:100;uniqueIndex;comment:邮箱" json:"email"`
	Phone       string `gorm:"size:20;comment:手机号" json:"phone"`
	RealName    string `gorm:"size:100;comment:真实姓名" json:"realName"`
	Avatar      string `gorm:"size:500;comment:头像" json:"avatar"`
	Department  string `gorm:"size:100;comment:部门" json:"department"`
	Position    string `gorm:"size:100;comment:职位" json:"position"`
	LastLoginAt *int64 `gorm:"comment:最后登录时间" json:"lastLoginAt"`
	LastLoginIP string `gorm:"size:50;comment:最后登录IP" json:"lastLoginIp"`
}

// Role 角色模型
type Role struct {
	BaseModel
	Name        string `gorm:"size:100;uniqueIndex;not null;comment:角色名称" json:"name"`
	Code        string `gorm:"size:100;uniqueIndex;not null;comment:角色编码" json:"code"`
	Description string `gorm:"size:500;comment:角色描述" json:"description"`
	Sort        int    `gorm:"default:0;comment:排序" json:"sort"`
}

// Permission 权限模型
type Permission struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Code         string         `gorm:"type:varchar(128);uniqueIndex;not null;comment:权限编码" json:"code"`
	Name         string         `gorm:"type:varchar(128);not null;comment:权限名称" json:"name"`
	ResourceType string         `gorm:"type:varchar(32);column:resource_type;comment:资源类型" json:"resourceType"`
	HttpMethod   string         `gorm:"type:varchar(16);column:http_method;comment:HTTP方法" json:"httpMethod"`
	Path         string         `gorm:"type:varchar(255);comment:API路径" json:"path"`
	Description  string         `gorm:"type:varchar(255);comment:描述" json:"description"`
	Status       int            `gorm:"type:tinyint;default:1;comment:状态:1-正常,0-禁用" json:"status"`
	CreatedAt    time.Time      `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"comment:更新时间" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间" json:"deletedAt,omitempty"`
	CreatedBy    string         `gorm:"type:varchar(100);comment:创建人" json:"createdBy"`
	UpdatedBy    string         `gorm:"type:varchar(100);comment:更新人" json:"updatedBy"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID     uint `gorm:"primarykey" json:"id"`
	UserID uint `gorm:"not null;index;comment:用户ID" json:"userId"`
	RoleID uint `gorm:"not null;index;comment:角色ID" json:"roleId"`
}

// RolePermission 角色权限关联
type RolePermission struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	RoleID       uint      `gorm:"not null;index;comment:角色ID" json:"roleId"`
	PermissionID uint      `gorm:"not null;index;comment:权限ID;column:permission_id" json:"permissionId"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"createdAt"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
