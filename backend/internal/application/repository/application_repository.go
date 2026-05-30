package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type ApplicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

// Create 创建应用
func (r *ApplicationRepository) Create(app *model.Application) error {
	return r.db.Create(app).Error
}

// GetByID 根据ID获取应用
func (r *ApplicationRepository) GetByID(id uint) (*model.Application, error) {
	var app model.Application
	err := r.db.First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetByCode 根据编码获取应用
func (r *ApplicationRepository) GetByCode(code string) (*model.Application, error) {
	var app model.Application
	err := r.db.Where("code = ?", code).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// Update 更新应用
func (r *ApplicationRepository) Update(app *model.Application) error {
	return r.db.Save(app).Error
}

// Delete 删除应用
func (r *ApplicationRepository) Delete(id uint) error {
	return r.db.Delete(&model.Application{}, id).Error
}

// List 应用列表
func (r *ApplicationRepository) List(page, pageSize int, projectID uint, keyword string) ([]*model.Application, int64, error) {
	var apps []*model.Application
	var total int64

	query := r.db.Model(&model.Application{})
	
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&apps).Error
	if err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

// GetByProjectID 根据项目ID获取应用列表
func (r *ApplicationRepository) GetByProjectID(projectID uint) ([]*model.Application, error) {
	var apps []*model.Application
	err := r.db.Where("project_id = ?", projectID).Find(&apps).Error
	if err != nil {
		return nil, err
	}
	return apps, nil
}
