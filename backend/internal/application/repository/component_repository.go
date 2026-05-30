package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type ComponentRepository struct {
	db *gorm.DB
}

func NewComponentRepository(db *gorm.DB) *ComponentRepository {
	return &ComponentRepository{db: db}
}

// Create 创建组件
func (r *ComponentRepository) Create(component *model.Component) error {
	return r.db.Create(component).Error
}

// GetByID 根据ID获取组件
func (r *ComponentRepository) GetByID(id uint) (*model.Component, error) {
	var component model.Component
	err := r.db.First(&component, id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// Update 更新组件
func (r *ComponentRepository) Update(component *model.Component) error {
	return r.db.Save(component).Error
}

// Delete 删除组件
func (r *ComponentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Component{}, id).Error
}

// GetByApplicationID 根据应用ID获取组件列表
func (r *ComponentRepository) GetByApplicationID(appID uint) ([]*model.Component, error) {
	var components []*model.Component
	err := r.db.Where("application_id = ?", appID).Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}

// List 组件列表
func (r *ComponentRepository) List(page, pageSize int, appID uint) ([]*model.Component, int64, error) {
	var components []*model.Component
	var total int64

	query := r.db.Model(&model.Component{})
	
	if appID > 0 {
		query = query.Where("application_id = ?", appID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}
