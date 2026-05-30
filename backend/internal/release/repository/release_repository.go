package repository

import (
	"my-cloud/internal/release/model"

	"gorm.io/gorm"
)

type ReleaseRepository struct {
	db *gorm.DB
}

func NewReleaseRepository(db *gorm.DB) *ReleaseRepository {
	return &ReleaseRepository{db: db}
}

// Create 创建发布工单
func (r *ReleaseRepository) Create(release *model.Release) error {
	return r.db.Create(release).Error
}

// GetByID 根据ID获取发布工单
func (r *ReleaseRepository) GetByID(id uint) (*model.Release, error) {
	var release model.Release
	err := r.db.First(&release, id).Error
	return &release, err
}

// GetByReleaseNo 根据发布编号获取工单
func (r *ReleaseRepository) GetByReleaseNo(releaseNo string) (*model.Release, error) {
	var release model.Release
	err := r.db.Where("release_no = ?", releaseNo).First(&release).Error
	return &release, err
}

// List 获取发布工单列表
func (r *ReleaseRepository) List(appID, envID uint, releaseStatus string, page, pageSize int) ([]*model.Release, int64, error) {
	var releases []*model.Release
	var total int64

	query := r.db.Model(&model.Release{})
	if appID > 0 {
		query = query.Where("app_id = ?", appID)
	}
	if envID > 0 {
		query = query.Where("env_id = ?", envID)
	}
	if releaseStatus != "" {
		query = query.Where("release_status = ?", releaseStatus)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&releases).Error
	return releases, total, err
}

// Update 更新发布工单
func (r *ReleaseRepository) Update(release *model.Release) error {
	return r.db.Save(release).Error
}

// ReleaseApprovalRepository 发布审批仓库
type ReleaseApprovalRepository struct {
	db *gorm.DB
}

func NewReleaseApprovalRepository(db *gorm.DB) *ReleaseApprovalRepository {
	return &ReleaseApprovalRepository{db: db}
}

// Create 创建审批记录
func (r *ReleaseApprovalRepository) Create(approval *model.ReleaseApproval) error {
	return r.db.Create(approval).Error
}

// GetByID 根据ID获取审批记录
func (r *ReleaseApprovalRepository) GetByID(id uint) (*model.ReleaseApproval, error) {
	var approval model.ReleaseApproval
	err := r.db.First(&approval, id).Error
	return &approval, err
}

// ListByRelease 获取发布工单的审批记录列表
func (r *ReleaseApprovalRepository) ListByRelease(releaseID uint) ([]*model.ReleaseApproval, error) {
	var approvals []*model.ReleaseApproval
	err := r.db.Where("release_id = ?", releaseID).Order("id DESC").Find(&approvals).Error
	return approvals, err
}

// Update 更新审批记录
func (r *ReleaseApprovalRepository) Update(approval *model.ReleaseApproval) error {
	return r.db.Save(approval).Error
}

// CountPendingApprovals 统计待审批数量
func (r *ReleaseApprovalRepository) CountPendingApprovals(releaseID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ReleaseApproval{}).
		Where("release_id = ? AND approval_status = ?", releaseID, "pending").
		Count(&count).Error
	return count, err
}

// CountApprovedApprovals 统计已审批通过数量
func (r *ReleaseApprovalRepository) CountApprovedApprovals(releaseID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ReleaseApproval{}).
		Where("release_id = ? AND approval_status = ?", releaseID, "approved").
		Count(&count).Error
	return count, err
}
