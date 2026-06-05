package repository

import (
	"my-cloud/internal/common/model"

	"gorm.io/gorm"
)

type NodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// Create 创建节点
func (r *NodeRepository) Create(node *model.ClusterNode) error {
	return r.db.Create(node).Error
}

// GetByID 根据ID查询节点
func (r *NodeRepository) GetByID(id uint) (*model.ClusterNode, error) {
	var node model.ClusterNode
	err := r.db.Where("id = ? AND is_deleted = 0", id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// Update 更新节点
func (r *NodeRepository) Update(node *model.ClusterNode) error {
	return r.db.Save(node).Error
}

// Delete 删除节点（软删除）
func (r *NodeRepository) Delete(id uint) error {
	return r.db.Model(&model.ClusterNode{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// List 分页查询节点列表
func (r *NodeRepository) List(offset, limit int, clusterID *uint, nodeRole *string) ([]model.ClusterNode, int64, error) {
	var nodes []model.ClusterNode
	var total int64

	query := r.db.Model(&model.ClusterNode{}).Where("is_deleted = 0")

	if clusterID != nil {
		query = query.Where("cluster_id = ?", *clusterID)
	}

	if nodeRole != nil && *nodeRole != "" {
		query = query.Where("node_role = ?", *nodeRole)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&nodes).Error; err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

// GetByClusterAndName 根据集群ID和节点名查询
func (r *NodeRepository) GetByClusterAndName(clusterID uint, nodeName string) (*model.ClusterNode, error) {
	var node model.ClusterNode
	err := r.db.Where("cluster_id = ? AND node_name = ? AND is_deleted = 0", clusterID, nodeName).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetByClusterID 根据集群ID查询节点列表
func (r *NodeRepository) GetByClusterID(clusterID uint) ([]model.ClusterNode, error) {
	var nodes []model.ClusterNode
	err := r.db.Where("cluster_id = ? AND is_deleted = 0", clusterID).Find(&nodes).Error
	return nodes, err
}
