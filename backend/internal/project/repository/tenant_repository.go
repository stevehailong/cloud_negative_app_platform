package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// GetAll 获取所有租户
func (r *TenantRepository) GetAll() ([]*model.Tenant, error) {
	var tenants []*model.Tenant
	err := r.db.Where("is_deleted = 0").Find(&tenants).Error
	return tenants, err
}

// GetByID 根据ID获取租户
func (r *TenantRepository) GetByID(id uint) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&tenant).Error
	return &tenant, err
}

// GetByCode 根据code获取租户
func (r *TenantRepository) GetByCode(code string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.Where("tenant_code = ? AND is_deleted = 0", code).First(&tenant).Error
	return &tenant, err
}

// Create 创建租户
func (r *TenantRepository) Create(tenant *model.Tenant) error {
	return r.db.Create(tenant).Error
}

// Update 更新租户
func (r *TenantRepository) Update(tenant *model.Tenant) error {
	return r.db.Model(tenant).Updates(tenant).Error
}

// UpdateFields 更新租户指定字段（支持零值更新）
func (r *TenantRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Tenant{}).Where("id = ? AND is_deleted = 0", id).Updates(fields).Error
}

// Delete 删除租户（软删除）
func (r *TenantRepository) Delete(id uint) error {
	return r.db.Model(&model.Tenant{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页获取租户列表
func (r *TenantRepository) List(page, pageSize int, keyword string) ([]*model.Tenant, int64, error) {
	var tenants []*model.Tenant
	var total int64

	db := r.db.Model(&model.Tenant{}).Where("is_deleted = 0")
	if keyword != "" {
		db = db.Where("tenant_name LIKE ? OR tenant_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Find(&tenants).Error

	return tenants, total, err
}
