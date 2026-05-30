package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// GetAll 获取所有权限
func (r *PermissionRepository) GetAll() ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.Where("status = ? AND deleted_at IS NULL", 1).
		Order("resource_type, code").
		Find(&permissions).Error
	return permissions, err
}

// GetByID 根据ID获取权限
func (r *PermissionRepository) GetByID(id uint) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.Where("deleted_at IS NULL").First(&permission, id).Error
	return &permission, err
}

// GetByCode 根据编码获取权限
func (r *PermissionRepository) GetByCode(code string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&permission).Error
	return &permission, err
}

// GetRolePermissions 获取角色的所有权限
func (r *PermissionRepository) GetRolePermissions(roleID uint) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.status = 1 AND permissions.deleted_at IS NULL", roleID).
		Find(&permissions).Error
	return permissions, err
}

// GetUserPermissions 获取用户的所有权限（通过角色）
func (r *PermissionRepository) GetUserPermissions(userID uint) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("INNER JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.status = 1 AND permissions.deleted_at IS NULL", userID).
		Group("permissions.id"). // 去重
		Find(&permissions).Error
	return permissions, err
}

// Create 创建权限
func (r *PermissionRepository) Create(permission *model.Permission) error {
	return r.db.Create(permission).Error
}

// Update 更新权限
func (r *PermissionRepository) Update(permission *model.Permission) error {
	return r.db.Save(permission).Error
}

// Delete 删除权限（软删除）
func (r *PermissionRepository) Delete(id uint) error {
	return r.db.Model(&model.Permission{}).Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// AssignPermissionsToRole 为角色分配权限
func (r *PermissionRepository) AssignPermissionsToRole(roleID uint, permissionIDs []uint) error {
	// 开启事务
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 删除旧的权限关联
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}

		// 2. 添加新的权限关联
		for _, permID := range permissionIDs {
			rp := &model.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
			if err := tx.Create(rp).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// RemovePermissionFromRole 从角色移除权限
func (r *PermissionRepository) RemovePermissionFromRole(roleID, permissionID uint) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&model.RolePermission{}).Error
}

// List 权限列表（支持分页和筛选）
func (r *PermissionRepository) List(page, pageSize int, resourceType string) ([]*model.Permission, int64, error) {
	var permissions []*model.Permission
	var total int64

	query := r.db.Model(&model.Permission{}).Where("deleted_at IS NULL")
	
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("resource_type, code").
		Offset(offset).Limit(pageSize).
		Find(&permissions).Error

	return permissions, total, err
}
