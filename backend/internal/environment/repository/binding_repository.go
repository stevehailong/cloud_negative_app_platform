package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type AppEnvBindingRepository struct {
	db *gorm.DB
}

func NewAppEnvBindingRepository(db *gorm.DB) *AppEnvBindingRepository {
	return &AppEnvBindingRepository{db: db}
}

// Create 创建应用环境绑定
func (r *AppEnvBindingRepository) Create(binding *model.AppEnvBinding) error {
	return r.db.Create(binding).Error
}

// GetByID 根据ID查询绑定
func (r *AppEnvBindingRepository) GetByID(id uint) (*model.AppEnvBinding, error) {
	var binding model.AppEnvBinding
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// GetByAppAndEnv 根据应用ID和环境ID查询绑定
func (r *AppEnvBindingRepository) GetByAppAndEnv(appID, envID uint) (*model.AppEnvBinding, error) {
	var binding model.AppEnvBinding
	err := r.db.Where("app_id = ? AND env_id = ? AND is_deleted = 0", appID, envID).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// Update 更新绑定
func (r *AppEnvBindingRepository) Update(binding *model.AppEnvBinding) error {
	return r.db.Save(binding).Error
}

// Delete 删除绑定（软删除）
func (r *AppEnvBindingRepository) Delete(id uint) error {
	return r.db.Model(&model.AppEnvBinding{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页查询绑定列表
func (r *AppEnvBindingRepository) List(offset, limit int, appID *uint, envID *uint) ([]model.AppEnvBinding, int64, error) {
	var bindings []model.AppEnvBinding
	var total int64

	query := r.db.Model(&model.AppEnvBinding{}).Where("is_deleted = 0")

	if appID != nil {
		query = query.Where("app_id = ?", *appID)
	}

	if envID != nil {
		query = query.Where("env_id = ?", *envID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&bindings).Error; err != nil {
		return nil, 0, err
	}

	return bindings, total, nil
}

// GetByAppID 根据应用ID查询绑定列表
func (r *AppEnvBindingRepository) GetByAppID(appID uint) ([]model.AppEnvBinding, error) {
	var bindings []model.AppEnvBinding
	err := r.db.Where("app_id = ? AND is_deleted = 0", appID).Find(&bindings).Error
	return bindings, err
}

// GetByEnvID 根据环境ID查询绑定列表
func (r *AppEnvBindingRepository) GetByEnvID(envID uint) ([]model.AppEnvBinding, error) {
	var bindings []model.AppEnvBinding
	err := r.db.Where("env_id = ? AND is_deleted = 0", envID).Find(&bindings).Error
	return bindings, err
}
