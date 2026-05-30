package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type EnvironmentRepository struct {
	db *gorm.DB
}

func NewEnvironmentRepository(db *gorm.DB) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}

// Create 创建环境
func (r *EnvironmentRepository) Create(env *model.Environment) error {
	return r.db.Create(env).Error
}

// GetByID 根据ID查询环境
func (r *EnvironmentRepository) GetByID(id uint) (*model.Environment, error) {
	var env model.Environment
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// GetByCode 根据编码查询环境
func (r *EnvironmentRepository) GetByCode(code string) (*model.Environment, error) {
	var env model.Environment
	err := r.db.Where("env_code = ? AND is_deleted = 0", code).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// Update 更新环境
func (r *EnvironmentRepository) Update(env *model.Environment) error {
	return r.db.Save(env).Error
}

// Delete 删除环境（软删除）
func (r *EnvironmentRepository) Delete(id uint) error {
	return r.db.Model(&model.Environment{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页查询环境列表
func (r *EnvironmentRepository) List(offset, limit int, keyword string, projectID *uint) ([]model.Environment, int64, error) {
	var envs []model.Environment
	var total int64

	query := r.db.Model(&model.Environment{}).Where("is_deleted = 0")

	if keyword != "" {
		query = query.Where("env_name LIKE ? OR env_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&envs).Error; err != nil {
		return nil, 0, err
	}

	return envs, total, nil
}

// GetByProjectID 根据项目ID查询环境列表
func (r *EnvironmentRepository) GetByProjectID(projectID uint) ([]model.Environment, error) {
	var envs []model.Environment
	err := r.db.Where("project_id = ? AND is_deleted = 0", projectID).Find(&envs).Error
	return envs, err
}

// GetByClusterID 根据集群ID查询环境列表
func (r *EnvironmentRepository) GetByClusterID(clusterID uint) ([]model.Environment, error) {
	var envs []model.Environment
	err := r.db.Where("cluster_id = ? AND is_deleted = 0", clusterID).Find(&envs).Error
	return envs, err
}
