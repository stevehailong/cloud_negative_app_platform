package repository

import (
	"my-cloud/internal/common/model"
	"time"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create 创建审计日志
func (r *AuditRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// List 获取审计日志列表
func (r *AuditRepository) List(filters map[string]interface{}, page, pageSize int) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{})

	// 应用过滤条件
	if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if username, ok := filters["username"].(string); ok && username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}

	if resourceType, ok := filters["resource_type"].(string); ok && resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if resourceID, ok := filters["resource_id"].(uint); ok && resourceID > 0 {
		query = query.Where("resource_id = ?", resourceID)
	}

	if method, ok := filters["method"].(string); ok && method != "" {
		query = query.Where("method = ?", method)
	}

	if path, ok := filters["path"].(string); ok && path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}

	if ipAddress, ok := filters["ip_address"].(string); ok && ipAddress != "" {
		query = query.Where("ip_address = ?", ipAddress)
	}

	// 时间范围过滤
	if startTime, ok := filters["start_time"].(time.Time); ok {
		query = query.Where("create_time >= ?", startTime)
	}

	if endTime, ok := filters["end_time"].(time.Time); ok {
		query = query.Where("create_time <= ?", endTime)
	}

	// 响应码过滤
	if responseCode, ok := filters["response_code"].(int); ok && responseCode > 0 {
		query = query.Where("response_code = ?", responseCode)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID 根据ID获取审计日志
func (r *AuditRepository) GetByID(id uint) (*model.AuditLog, error) {
	var log model.AuditLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByResourceID 根据资源ID获取相关审计日志
func (r *AuditRepository) GetByResourceID(resourceType string, resourceID uint, page, pageSize int) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByUserID 根据用户ID获取审计日志
func (r *AuditRepository) GetByUserID(userID uint, page, pageSize int) ([]*model.AuditLog, int64, error) {
	var logs []*model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetStatistics 获取审计日志统计信息
func (r *AuditRepository) GetStatistics(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总操作次数
	var totalCount int64
	if err := r.db.Model(&model.AuditLog{}).
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// 按操作类型统计
	var actionStats []struct {
		Action string
		Count  int64
	}
	if err := r.db.Model(&model.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Group("action").
		Scan(&actionStats).Error; err != nil {
		return nil, err
	}
	stats["action_stats"] = actionStats

	// 按资源类型统计
	var resourceStats []struct {
		ResourceType string
		Count        int64
	}
	if err := r.db.Model(&model.AuditLog{}).
		Select("resource_type, COUNT(*) as count").
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Group("resource_type").
		Order("count DESC").
		Limit(10).
		Scan(&resourceStats).Error; err != nil {
		return nil, err
	}
	stats["resource_stats"] = resourceStats

	// 按用户统计(Top 10)
	var userStats []struct {
		Username string
		Count    int64
	}
	if err := r.db.Model(&model.AuditLog{}).
		Select("username, COUNT(*) as count").
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Group("username").
		Order("count DESC").
		Limit(10).
		Scan(&userStats).Error; err != nil {
		return nil, err
	}
	stats["user_stats"] = userStats

	// 按响应码统计
	var statusStats []struct {
		ResponseCode int
		Count        int64
	}
	if err := r.db.Model(&model.AuditLog{}).
		Select("response_code, COUNT(*) as count").
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Group("response_code").
		Order("count DESC").
		Scan(&statusStats).Error; err != nil {
		return nil, err
	}
	stats["status_stats"] = statusStats

	// 平均响应时间
	var avgDuration float64
	if err := r.db.Model(&model.AuditLog{}).
		Select("AVG(duration_ms) as avg_duration").
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Scan(&avgDuration).Error; err != nil {
		return nil, err
	}
	stats["avg_duration_ms"] = avgDuration

	return stats, nil
}

// Delete 删除过期审计日志
func (r *AuditRepository) DeleteOldLogs(beforeTime time.Time) (int64, error) {
	result := r.db.Where("create_time < ?", beforeTime).Delete(&model.AuditLog{})
	return result.RowsAffected, result.Error
}
