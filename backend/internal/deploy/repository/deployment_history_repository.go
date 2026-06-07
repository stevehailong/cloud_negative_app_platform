package repository

import (
	"my-cloud/internal/deploy/model"

	"gorm.io/gorm"
)

type DeploymentHistoryRepository struct {
	db *gorm.DB
}

func NewDeploymentHistoryRepository(db *gorm.DB) *DeploymentHistoryRepository {
	return &DeploymentHistoryRepository{db: db}
}

// Create 创建历史记录
func (r *DeploymentHistoryRepository) Create(history *model.DeploymentHistory) error {
	return r.db.Create(history).Error
}

// GetByID 根据ID获取
func (r *DeploymentHistoryRepository) GetByID(id int64) (*model.DeploymentHistory, error) {
	var history model.DeploymentHistory
	err := r.db.Where("id = ?", id).First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ListByAppDeployment 根据app_deployment_id查询历史记录
func (r *DeploymentHistoryRepository) ListByAppDeployment(appDeploymentID int64, page, pageSize int) ([]model.DeploymentHistory, int64, error) {
	var histories []model.DeploymentHistory
	var total int64

	query := r.db.Model(&model.DeploymentHistory{}).Where("app_deployment_id = ?", appDeploymentID)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// Update 更新
func (r *DeploymentHistoryRepository) Update(history *model.DeploymentHistory) error {
	return r.db.Save(history).Error
}

// UpdateFields 更新指定字段
func (r *DeploymentHistoryRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.DeploymentHistory{}).Where("id = ?", id).Updates(fields).Error
}

// RecoverStuckProgressing 将因服务重启而卡在 progressing 状态的记录标记为 failed
func (r *DeploymentHistoryRepository) RecoverStuckProgressing(reason string) (int64, error) {
	result := r.db.Model(&model.DeploymentHistory{}).
		Where("status = ?", "progressing").
		Updates(map[string]interface{}{
			"status":         "failed",
			"failure_reason": reason,
		})
	return result.RowsAffected, result.Error
}
