package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type NamespaceRepository struct {
	db *gorm.DB
}

func NewNamespaceRepository(db *gorm.DB) *NamespaceRepository {
	return &NamespaceRepository{db: db}
}

// Create 创建命名空间
func (r *NamespaceRepository) Create(ns *model.Namespace) error {
	return r.db.Create(ns).Error
}

// GetByID 根据ID查询命名空间
func (r *NamespaceRepository) GetByID(id uint) (*model.Namespace, error) {
	var ns model.Namespace
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&ns).Error
	if err != nil {
		return nil, err
	}
	return &ns, nil
}

// GetByClusterAndName 根据集群和名称查询命名空间
func (r *NamespaceRepository) GetByClusterAndName(clusterID uint, name string) (*model.Namespace, error) {
	var ns model.Namespace
	err := r.db.Where("cluster_id = ? AND namespace_name = ? AND is_deleted = 0", clusterID, name).First(&ns).Error
	if err != nil {
		return nil, err
	}
	return &ns, nil
}

// Update 更新命名空间
func (r *NamespaceRepository) Update(ns *model.Namespace) error {
	return r.db.Save(ns).Error
}

// Delete 删除命名空间（软删除）
func (r *NamespaceRepository) Delete(id uint) error {
	return r.db.Model(&model.Namespace{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页查询命名空间列表
func (r *NamespaceRepository) List(offset, limit int, clusterID *uint, projectID *uint) ([]model.Namespace, int64, error) {
	var namespaces []model.Namespace
	var total int64

	query := r.db.Model(&model.Namespace{}).Where("is_deleted = 0")

	if clusterID != nil {
		query = query.Where("cluster_id = ?", *clusterID)
	}

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&namespaces).Error; err != nil {
		return nil, 0, err
	}

	return namespaces, total, nil
}

// GetByClusterID 根据集群ID查询命名空间列表
func (r *NamespaceRepository) GetByClusterID(clusterID uint) ([]model.Namespace, error) {
	var namespaces []model.Namespace
	err := r.db.Where("cluster_id = ? AND is_deleted = 0", clusterID).Find(&namespaces).Error
	return namespaces, err
}

// CountByCluster 统计集群下命名空间数量
func (r *NamespaceRepository) CountByCluster(clusterID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Namespace{}).Where("cluster_id = ? AND is_deleted = 0", clusterID).Count(&count).Error
	return count, err
}

// GetByProjectID 根据项目ID查询命名空间列表
func (r *NamespaceRepository) GetByProjectID(projectID uint) ([]model.Namespace, error) {
	var namespaces []model.Namespace
	err := r.db.Where("project_id = ? AND is_deleted = 0", projectID).Find(&namespaces).Error
	return namespaces, err
}
