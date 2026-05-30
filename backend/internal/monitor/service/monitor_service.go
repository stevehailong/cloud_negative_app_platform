package service

import (
	"errors"
	"my-cloud/internal/monitor/model"
	"my-cloud/internal/monitor/repository"
	"time"
)

type MonitorService struct {
	monitorRepo *repository.MonitorRepository
}

func NewMonitorService(monitorRepo *repository.MonitorRepository) *MonitorService {
	return &MonitorService{monitorRepo: monitorRepo}
}

// Metric管理
func (s *MonitorService) CreateMetric(metric *model.Metric) error {
	if metric.Name == "" || metric.Type == "" {
		return errors.New("指标名称和类型不能为空")
	}
	return s.monitorRepo.CreateMetric(metric)
}

func (s *MonitorService) GetMetric(id uint) (*model.Metric, error) {
	return s.monitorRepo.GetMetric(id)
}

func (s *MonitorService) ListMetrics(metricType string, enabled *int, page, pageSize int) ([]*model.Metric, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.monitorRepo.ListMetrics(metricType, enabled, page, pageSize)
}

func (s *MonitorService) UpdateMetric(metric *model.Metric) error {
	return s.monitorRepo.UpdateMetric(metric)
}

func (s *MonitorService) DeleteMetric(id uint) error {
	return s.monitorRepo.DeleteMetric(id)
}

// AlertRule管理
func (s *MonitorService) CreateAlertRule(rule *model.AlertRule) error {
	if rule.Name == "" || rule.MetricName == "" {
		return errors.New("规则名称和指标名称不能为空")
	}
	if rule.Condition == "" || rule.Threshold == 0 {
		return errors.New("告警条件和阈值不能为空")
	}
	return s.monitorRepo.CreateAlertRule(rule)
}

func (s *MonitorService) GetAlertRule(id uint) (*model.AlertRule, error) {
	return s.monitorRepo.GetAlertRule(id)
}

func (s *MonitorService) ListAlertRules(metricName, severity string, enabled *int, page, pageSize int) ([]*model.AlertRule, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.monitorRepo.ListAlertRules(metricName, severity, enabled, page, pageSize)
}

func (s *MonitorService) UpdateAlertRule(rule *model.AlertRule) error {
	return s.monitorRepo.UpdateAlertRule(rule)
}

func (s *MonitorService) DeleteAlertRule(id uint) error {
	return s.monitorRepo.DeleteAlertRule(id)
}

// Alert管理
func (s *MonitorService) CreateAlert(alert *model.Alert) error {
	return s.monitorRepo.CreateAlert(alert)
}

func (s *MonitorService) GetAlert(id uint) (*model.Alert, error) {
	return s.monitorRepo.GetAlert(id)
}

func (s *MonitorService) ListAlerts(status, severity string, startTime, endTime *time.Time, page, pageSize int) ([]*model.Alert, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.monitorRepo.ListAlerts(status, severity, startTime, endTime, page, pageSize)
}

func (s *MonitorService) ResolveAlert(id uint) error {
	return s.monitorRepo.ResolveAlert(id)
}

func (s *MonitorService) GetAlertStatistics(startTimeStr, endTimeStr string) (map[string]interface{}, error) {
	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse("2006-01-02", startTimeStr)
		if err != nil {
			return nil, errors.New("开始时间格式错误")
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse("2006-01-02", endTimeStr)
		if err != nil {
			return nil, errors.New("结束时间格式错误")
		}
		endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		endTime = time.Now()
	}

	stats, err := s.monitorRepo.GetAlertStatistics(startTime, endTime)
	if err != nil {
		return nil, err
	}

	stats["start_time"] = startTime.Format("2006-01-02 15:04:05")
	stats["end_time"] = endTime.Format("2006-01-02 15:04:05")

	return stats, nil
}
