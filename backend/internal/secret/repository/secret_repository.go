package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type SecretRepository struct {
	db *gorm.DB
}

func NewSecretRepository(db *gorm.DB) *SecretRepository {
	return &SecretRepository{db: db}
}

// Create 创建密钥引用
func (r *SecretRepository) Create(secret *model.AppSecret) error {
	return r.db.Create(secret).Error
}

// GetByID 根据ID查询密钥
func (r *SecretRepository) GetByID(id uint) (*model.AppSecret, error) {
	var secret model.AppSecret
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&secret).Error
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

// Update 更新密钥引用
func (r *SecretRepository) Update(secret *model.AppSecret) error {
	return r.db.Save(secret).Error
}

// Delete 删除密钥引用（软删除）
func (r *SecretRepository) Delete(id uint) error {
	return r.db.Model(&model.AppSecret{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页查询密钥列表
func (r *SecretRepository) List(offset, limit int, appId *uint, envId *uint, keyword string) ([]model.AppSecret, int64, error) {
	var secrets []model.AppSecret
	var total int64

	query := r.db.Model(&model.AppSecret{}).Where("is_deleted = 0")

	if appId != nil && *appId > 0 {
		query = query.Where("app_id = ?", *appId)
	}
	if envId != nil && *envId > 0 {
		query = query.Where("env_id = ?", *envId)
	}
	if keyword != "" {
		query = query.Where("secret_key LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&secrets).Error; err != nil {
		return nil, 0, err
	}

	return secrets, total, nil
}

// GetByAppAndEnv 根据应用ID和环境ID查询密钥列表
func (r *SecretRepository) GetByAppAndEnv(appId, envId uint) ([]model.AppSecret, error) {
	var secrets []model.AppSecret
	err := r.db.Where("app_id = ? AND env_id = ? AND is_deleted = 0", appId, envId).Find(&secrets).Error
	return secrets, err
}
