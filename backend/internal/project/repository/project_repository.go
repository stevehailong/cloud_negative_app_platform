package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// GetAll 获取所有项目
func (r *ProjectRepository) GetAll() ([]*model.Project, error) {
	var projects []*model.Project
	err := r.db.Where("is_deleted = 0").Find(&projects).Error
	return projects, err
}

// GetByID 根据ID获取项目
func (r *ProjectRepository) GetByID(id uint) (*model.Project, error) {
	var project model.Project
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&project).Error
	return &project, err
}

// GetByCode 根据code获取项目
func (r *ProjectRepository) GetByCode(code string) (*model.Project, error) {
	var project model.Project
	err := r.db.Where("project_code = ? AND is_deleted = 0", code).First(&project).Error
	return &project, err
}

// Create 创建项目
func (r *ProjectRepository) Create(project *model.Project) error {
	return r.db.Create(project).Error
}

// Update 更新项目
func (r *ProjectRepository) Update(project *model.Project) error {
	return r.db.Model(project).Updates(project).Error
}

// UpdateFields 更新项目指定字段（支持零值更新）
func (r *ProjectRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Project{}).Where("id = ? AND is_deleted = 0", id).Updates(fields).Error
}

// Delete 删除项目（软删除）
func (r *ProjectRepository) Delete(id uint) error {
	return r.db.Model(&model.Project{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页获取项目列表
func (r *ProjectRepository) List(page, pageSize int, keyword string, tenantID uint) ([]*model.Project, int64, error) {
	var projects []*model.Project
	var total int64

	db := r.db.Model(&model.Project{}).Where("is_deleted = 0")
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	if keyword != "" {
		db = db.Where("project_name LIKE ? OR project_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&projects).Error

	return projects, total, err
}

// ListAccessible 分页获取用户可访问的项目列表（私有项目只对成员和Owner可见）
func (r *ProjectRepository) ListAccessible(page, pageSize int, keyword string, tenantID uint, userID uint) ([]*model.Project, int64, error) {
	var projects []*model.Project
	var total int64

	db := r.db.Model(&model.Project{}).Where("projects.is_deleted = 0")
	if tenantID > 0 {
		db = db.Where("projects.tenant_id = ?", tenantID)
	}
	if keyword != "" {
		db = db.Where("(projects.project_name LIKE ? OR projects.project_code LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 私有项目只对成员和Owner可见；公开项目对所有人可见
	db = db.Where(
		"(projects.visibility = 'public') OR "+
			"(projects.owner_user_id = ?) OR "+
			"(projects.id IN (SELECT project_id FROM project_members WHERE user_id = ?))",
		userID, userID,
	)

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Order("projects.create_time DESC").Find(&projects).Error

	return projects, total, err
}

// IsMember 检查用户是否是项目成员
func (r *ProjectRepository) IsMember(projectID, userID uint) bool {
	var count int64
	r.db.Model(&model.ProjectMember{}).Where("project_id = ? AND user_id = ?", projectID, userID).Count(&count)
	return count > 0
}

// IsAdmin 检查用户是否是超级管理员
func (r *ProjectRepository) IsAdmin(userID uint) bool {
	var count int64
	r.db.Table("user_roles").
		Joins("INNER JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.code = 'SUPER_ADMIN'", userID).
		Count(&count)
	return count > 0
}

// GetByTenantID 根据租户ID获取项目列表
func (r *ProjectRepository) GetByTenantID(tenantID uint) ([]*model.Project, error) {
	var projects []*model.Project
	err := r.db.Where("tenant_id = ? AND is_deleted = 0", tenantID).Find(&projects).Error
	return projects, err
}

// GetMembers 获取项目成员
func (r *ProjectRepository) GetMembers(projectID uint) ([]*model.ProjectMember, error) {
	var members []*model.ProjectMember
	err := r.db.Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

// AddMember 添加项目成员
func (r *ProjectRepository) AddMember(member *model.ProjectMember) error {
	return r.db.Create(member).Error
}

// RemoveMember 移除项目成员
func (r *ProjectRepository) RemoveMember(projectID, userID uint) error {
	return r.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&model.ProjectMember{}).Error
}

// UpdateMemberRole 更新成员角色
func (r *ProjectRepository) UpdateMemberRole(projectID, userID uint, roleCode string) error {
	return r.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Update("role_code", roleCode).Error
}
