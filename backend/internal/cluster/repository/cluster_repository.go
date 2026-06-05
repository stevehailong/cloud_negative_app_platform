package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type ClusterRepository struct {
	db *gorm.DB
}

func NewClusterRepository(db *gorm.DB) *ClusterRepository {
	return &ClusterRepository{db: db}
}

// Create 创建集群
func (r *ClusterRepository) Create(cluster *model.Cluster) error {
	return r.db.Create(cluster).Error
}

// GetByID 根据ID查询集群
func (r *ClusterRepository) GetByID(id uint) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

// GetByCode 根据编码查询集群
func (r *ClusterRepository) GetByCode(code string) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.Where("cluster_code = ? AND is_deleted = 0", code).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

// Update 更新集群
func (r *ClusterRepository) Update(cluster *model.Cluster) error {
	return r.db.Save(cluster).Error
}

// Delete 删除集群（软删除）
func (r *ClusterRepository) Delete(id uint) error {
	return r.db.Model(&model.Cluster{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页查询集群列表
func (r *ClusterRepository) List(offset, limit int, keyword string, clusterType *string) ([]model.Cluster, int64, error) {
	var clusters []model.Cluster
	var total int64

	query := r.db.Model(&model.Cluster{}).Where("is_deleted = 0")

	if keyword != "" {
		query = query.Where("cluster_name LIKE ? OR cluster_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if clusterType != nil && *clusterType != "" {
		query = query.Where("cluster_type = ?", *clusterType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&clusters).Error; err != nil {
		return nil, 0, err
	}

	return clusters, total, nil
}

// UpdateVersion 更新集群版本号
func (r *ClusterRepository) UpdateVersion(id uint, version string) error {
	return r.db.Model(&model.Cluster{}).Where("id = ?", id).Update("version", version).Error
}

// GetByRegion 根据区域查询集群列表
func (r *ClusterRepository) GetByRegion(region string) ([]model.Cluster, error) {
	var clusters []model.Cluster
	err := r.db.Where("region = ? AND is_deleted = 0", region).Find(&clusters).Error
	return clusters, err
}
