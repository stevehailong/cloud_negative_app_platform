package repository

import (
	"my-cloud/internal/common/model"
	"time"

	"gorm.io/gorm"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Create 创建配置
func (r *ConfigRepository) Create(config *model.AppConfig) error {
	config.CreateTime = time.Now()
	config.UpdateTime = time.Now()
	return r.db.Create(config).Error
}

// GetByID 根据ID查询配置
func (r *ConfigRepository) GetByID(id uint) (*model.AppConfig, error) {
	var config model.AppConfig
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Update 更新配置
func (r *ConfigRepository) Update(config *model.AppConfig) error {
	config.UpdateTime = time.Now()
	return r.db.Save(config).Error
}

// Delete 删除配置（软删除）
func (r *ConfigRepository) Delete(id uint) error {
	return r.db.Model(&model.AppConfig{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_deleted":  1,
		"update_time": time.Now(),
	}).Error
}

// List 分页查询配置列表
func (r *ConfigRepository) List(appID, envID uint, keyword string, page, pageSize int) ([]model.AppConfig, int64, error) {
	var configs []model.AppConfig
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.Model(&model.AppConfig{}).Where("is_deleted = 0")

	if appID > 0 {
		query = query.Where("app_id = ?", appID)
	}
	if envID > 0 {
		query = query.Where("env_id = ?", envID)
	}
	if keyword != "" {
		query = query.Where("config_key LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

// GetByAppAndEnv 根据应用ID和环境ID获取配置列表
func (r *ConfigRepository) GetByAppAndEnv(appID, envID uint, page, pageSize int) ([]model.AppConfig, int64, error) {
	var configs []model.AppConfig
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.Model(&model.AppConfig{}).Where("app_id = ? AND env_id = ? AND is_deleted = 0", appID, envID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}
