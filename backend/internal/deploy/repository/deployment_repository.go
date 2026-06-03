package repository

import (
	"my-cloud/internal/deploy/model"

	"gorm.io/gorm"
)

type DeploymentRepository struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

// Create 创建部署记录
func (r *DeploymentRepository) Create(deployment *model.Deployment) error {
	return r.db.Create(deployment).Error
}

// GetByID 根据ID获取部署记录
func (r *DeploymentRepository) GetByID(id uint) (*model.Deployment, error) {
	var deployment model.Deployment
	err := r.db.First(&deployment, id).Error
	return &deployment, err
}

// GetByRelease 根据ReleaseID获取部署记录
func (r *DeploymentRepository) GetByRelease(releaseID uint) (*model.Deployment, error) {
	var deployment model.Deployment
	err := r.db.Where("release_id = ?", releaseID).First(&deployment).Error
	return &deployment, err
}

// List 获取部署记录列表
func (r *DeploymentRepository) List(clusterID uint, namespace, startDate, sortBy, sortOrder string, page, pageSize int) ([]*model.Deployment, int64, error) {
	var deployments []*model.Deployment
	var total int64

	query := r.db.Model(&model.Deployment{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if startDate != "" {
		query = query.Where("DATE(create_time) = ?", startDate)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 排序映射
	orderClause := "id DESC"
	if sortBy == "createTime" {
		orderClause = "create_time"
		if sortOrder == "asc" {
			orderClause += " ASC"
		} else {
			orderClause += " DESC"
		}
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order(orderClause).Find(&deployments).Error
	return deployments, total, err
}

// Update 更新部署记录
func (r *DeploymentRepository) Update(deployment *model.Deployment) error {
	return r.db.Save(deployment).Error
}

// Delete 删除部署记录
func (r *DeploymentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Deployment{}, id).Error
}

// ListByClusterAndNamespace 根据集群和命名空间查询部署
func (r *DeploymentRepository) ListByClusterAndNamespace(clusterID uint, namespace string) ([]*model.Deployment, error) {
	var deployments []*model.Deployment
	err := r.db.Where("cluster_id = ? AND namespace = ?", clusterID, namespace).
		Order("id DESC").
		Find(&deployments).Error
	return deployments, err
}

// FindByWorkload 根据namespace和workloadName查询所有部署记录
func (r *DeploymentRepository) FindByWorkload(namespace, workloadName string) ([]*model.Deployment, error) {
	var deployments []*model.Deployment
	err := r.db.Where("namespace = ? AND workload_name = ?", namespace, workloadName).
		Find(&deployments).Error
	return deployments, err
}
