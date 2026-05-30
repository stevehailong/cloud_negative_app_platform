package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create 创建角色
func (r *RoleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// GetByID 根据ID获取角色
func (r *RoleRepository) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Update 更新角色
func (r *RoleRepository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

// Delete 删除角色
func (r *RoleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Role{}, id).Error
}

// List 角色列表
func (r *RoleRepository) List(page, pageSize int) ([]*model.Role, int64, error) {
	var roles []*model.Role
	var total int64

	query := r.db.Model(&model.Role{})
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("sort ASC").Offset(offset).Limit(pageSize).Find(&roles).Error
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// GetUserRoles 获取用户角色
func (r *RoleRepository) GetUserRoles(userID uint) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.Table("roles").
		Joins("INNER JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// AssignRolesToUser 为用户分配角色
func (r *RoleRepository) AssignRolesToUser(userID uint, roleIDs []uint) error {
	// 先删除旧的关联
	if err := r.db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return err
	}

	// 添加新的关联
	for _, roleID := range roleIDs {
		userRole := &model.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := r.db.Create(userRole).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteUserRoles 删除用户的所有角色
func (r *RoleRepository) DeleteUserRoles(userID uint) error {
	return r.db.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error
}

// AssignRole 为用户分配单个角色
func (r *RoleRepository) AssignRole(userID, roleID uint) error {
	return r.db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

// GetAllRoles 获取所有角色（不分页）
func (r *RoleRepository) GetAllRoles() ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.Where("status = ?", 1).Order("sort ASC").Find(&roles).Error
	return roles, err
}

// GetByCode 根据角色编码获取角色
func (r *RoleRepository) GetByCode(code string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("code = ? AND status = ?", code, 1).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// AssignRoleToUser 为用户分配单个角色（不删除其他角色）
func (r *RoleRepository) AssignRoleToUser(userID, roleID uint) error {
	userRole := &model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.Create(userRole).Error
}
