package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type ConfigMapRepository struct {
	db *gorm.DB
}

func NewConfigMapRepository(db *gorm.DB) *ConfigMapRepository {
	return &ConfigMapRepository{db: db}
}

func (r *ConfigMapRepository) Create(cm *model.ConfigMap) error {
	return r.db.Create(cm).Error
}

func (r *ConfigMapRepository) GetByID(id uint) (*model.ConfigMap, error) {
	var cm model.ConfigMap
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&cm).Error; err != nil {
		return nil, err
	}
	return &cm, nil
}

func (r *ConfigMapRepository) GetByEnvAndName(envID uint, name string) (*model.ConfigMap, error) {
	var cm model.ConfigMap
	if err := r.db.Where("env_id = ? AND name = ? AND is_deleted = 0", envID, name).First(&cm).Error; err != nil {
		return nil, err
	}
	return &cm, nil
}

func (r *ConfigMapRepository) Update(cm *model.ConfigMap) error {
	return r.db.Save(cm).Error
}

func (r *ConfigMapRepository) Delete(id uint) error {
	return r.db.Model(&model.ConfigMap{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

func (r *ConfigMapRepository) List(offset, pageSize int, envID *uint, namespace string) ([]*model.ConfigMap, int64, error) {
	var configMaps []*model.ConfigMap
	var total int64

	query := r.db.Model(&model.ConfigMap{}).Where("is_deleted = 0")

	if envID != nil {
		query = query.Where("env_id = ?", *envID)
	}
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&configMaps).Error; err != nil {
		return nil, 0, err
	}

	return configMaps, total, nil
}

type SecretRepository struct {
	db *gorm.DB
}

func NewSecretRepository(db *gorm.DB) *SecretRepository {
	return &SecretRepository{db: db}
}

func (r *SecretRepository) Create(secret *model.Secret) error {
	return r.db.Create(secret).Error
}

func (r *SecretRepository) GetByID(id uint) (*model.Secret, error) {
	var secret model.Secret
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&secret).Error; err != nil {
		return nil, err
	}
	return &secret, nil
}

func (r *SecretRepository) GetByEnvAndName(envID uint, name string) (*model.Secret, error) {
	var secret model.Secret
	if err := r.db.Where("env_id = ? AND name = ? AND is_deleted = 0", envID, name).First(&secret).Error; err != nil {
		return nil, err
	}
	return &secret, nil
}

func (r *SecretRepository) Update(secret *model.Secret) error {
	return r.db.Save(secret).Error
}

func (r *SecretRepository) Delete(id uint) error {
	return r.db.Model(&model.Secret{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

func (r *SecretRepository) List(offset, pageSize int, envID *uint, namespace string) ([]*model.Secret, int64, error) {
	var secrets []*model.Secret
	var total int64

	query := r.db.Model(&model.Secret{}).Where("is_deleted = 0")

	if envID != nil {
		query = query.Where("env_id = ?", *envID)
	}
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&secrets).Error; err != nil {
		return nil, 0, err
	}

	return secrets, total, nil
}
