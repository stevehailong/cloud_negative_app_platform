package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type EnvTemplateRepository struct {
	db *gorm.DB
}

func NewEnvTemplateRepository(db *gorm.DB) *EnvTemplateRepository {
	return &EnvTemplateRepository{db: db}
}

// Create 创建模板
func (r *EnvTemplateRepository) Create(template *model.EnvTemplate) error {
	return r.db.Create(template).Error
}

// GetByID 根据ID查询模板
func (r *EnvTemplateRepository) GetByID(id uint) (*model.EnvTemplate, error) {
	var template model.EnvTemplate
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetByCode 根据编码查询模板
func (r *EnvTemplateRepository) GetByCode(code string) (*model.EnvTemplate, error) {
	var template model.EnvTemplate
	err := r.db.Where("template_code = ? AND is_deleted = 0", code).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// Update 更新模板
func (r *EnvTemplateRepository) Update(template *model.EnvTemplate) error {
	return r.db.Save(template).Error
}

// Delete 删除模板（软删除）
func (r *EnvTemplateRepository) Delete(id uint) error {
	return r.db.Model(&model.EnvTemplate{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页查询模板列表
func (r *EnvTemplateRepository) List(offset, limit int, keyword string, templateType *string) ([]model.EnvTemplate, int64, error) {
	var templates []model.EnvTemplate
	var total int64

	query := r.db.Model(&model.EnvTemplate{}).Where("is_deleted = 0")

	if keyword != "" {
		query = query.Where("template_name LIKE ? OR template_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if templateType != nil && *templateType != "" {
		query = query.Where("template_type = ?", *templateType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}
