package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type ResourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

// Create 创建资源配额
func (r *ResourceRepository) Create(quota *model.ResourceQuota) error {
	return r.db.Create(quota).Error
}

// GetByID 根据ID查询资源配额
func (r *ResourceRepository) GetByID(id uint) (*model.ResourceQuota, error) {
	var quota model.ResourceQuota
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&quota).Error
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

// Update 更新资源配额
func (r *ResourceRepository) Update(quota *model.ResourceQuota) error {
	return r.db.Save(quota).Error
}

// Delete 删除资源配额（软删除）
func (r *ResourceRepository) Delete(id uint) error {
	return r.db.Model(&model.ResourceQuota{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页查询资源配额列表
func (r *ResourceRepository) List(offset, limit int, scopeType string, scopeID *uint) ([]model.ResourceQuota, int64, error) {
	var quotas []model.ResourceQuota
	var total int64

	query := r.db.Model(&model.ResourceQuota{}).Where("is_deleted = 0")

	if scopeType != "" {
		query = query.Where("scope_type = ?", scopeType)
	}

	if scopeID != nil && *scopeID > 0 {
		query = query.Where("scope_id = ?", *scopeID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&quotas).Error; err != nil {
		return nil, 0, err
	}

	return quotas, total, nil
}
