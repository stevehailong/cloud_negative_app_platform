package repository

import (
	"my-cloud/internal/monitor/model"
	"time"

	"gorm.io/gorm"
)

type MonitorRepository struct {
	db *gorm.DB
}

func NewMonitorRepository(db *gorm.DB) *MonitorRepository {
	return &MonitorRepository{db: db}
}

// Metric相关方法
func (r *MonitorRepository) CreateMetric(metric *model.Metric) error {
	return r.db.Create(metric).Error
}

func (r *MonitorRepository) GetMetric(id uint) (*model.Metric, error) {
	var metric model.Metric
	err := r.db.First(&metric, id).Error
	return &metric, err
}

func (r *MonitorRepository) ListMetrics(metricType string, enabled *int, page, pageSize int) ([]*model.Metric, int64, error) {
	var metrics []*model.Metric
	var total int64

	query := r.db.Model(&model.Metric{})

	if metricType != "" {
		query = query.Where("type = ?", metricType)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&metrics).Error
	return metrics, total, err
}

func (r *MonitorRepository) UpdateMetric(metric *model.Metric) error {
	return r.db.Save(metric).Error
}

func (r *MonitorRepository) DeleteMetric(id uint) error {
	return r.db.Delete(&model.Metric{}, id).Error
}

// AlertRule相关方法
func (r *MonitorRepository) CreateAlertRule(rule *model.AlertRule) error {
	return r.db.Create(rule).Error
}

func (r *MonitorRepository) GetAlertRule(id uint) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.First(&rule, id).Error
	return &rule, err
}

func (r *MonitorRepository) ListAlertRules(metricName, severity string, enabled *int, page, pageSize int) ([]*model.AlertRule, int64, error) {
	var rules []*model.AlertRule
	var total int64

	query := r.db.Model(&model.AlertRule{})

	if metricName != "" {
		query = query.Where("metric_name = ?", metricName)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&rules).Error
	return rules, total, err
}

func (r *MonitorRepository) UpdateAlertRule(rule *model.AlertRule) error {
	return r.db.Save(rule).Error
}

func (r *MonitorRepository) DeleteAlertRule(id uint) error {
	return r.db.Delete(&model.AlertRule{}, id).Error
}

// Alert相关方法
func (r *MonitorRepository) CreateAlert(alert *model.Alert) error {
	return r.db.Create(alert).Error
}

func (r *MonitorRepository) GetAlert(id uint) (*model.Alert, error) {
	var alert model.Alert
	err := r.db.First(&alert, id).Error
	return &alert, err
}

func (r *MonitorRepository) ListAlerts(status, severity string, startTime, endTime *time.Time, page, pageSize int) ([]*model.Alert, int64, error) {
	var alerts []*model.Alert
	var total int64

	query := r.db.Model(&model.Alert{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if startTime != nil {
		query = query.Where("create_time >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("create_time <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&alerts).Error
	return alerts, total, err
}

func (r *MonitorRepository) ResolveAlert(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "resolved",
		"resolved_at": now,
	}).Error
}

func (r *MonitorRepository) GetAlertStatistics(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总告警数
	var totalCount int64
	if err := r.db.Model(&model.Alert{}).
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// 按严重程度统计
	var severityStats []struct {
		Severity string
		Count    int64
	}
	if err := r.db.Model(&model.Alert{}).
		Select("severity, COUNT(*) as count").
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Group("severity").
		Scan(&severityStats).Error; err != nil {
		return nil, err
	}
	stats["severity_stats"] = severityStats

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&model.Alert{}).
		Select("status, COUNT(*) as count").
		Where("create_time BETWEEN ? AND ?", startTime, endTime).
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, err
	}
	stats["status_stats"] = statusStats

	return stats, nil
}
