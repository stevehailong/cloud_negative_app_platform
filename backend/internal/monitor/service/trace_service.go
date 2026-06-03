package service

import (
	"time"

	"my-cloud/internal/monitor/model"
	"my-cloud/internal/monitor/repository"
)

// TraceService 链路追踪服务层
type TraceService struct {
	traceRepo *repository.TraceRepository
}

func NewTraceService(traceRepo *repository.TraceRepository) *TraceService {
	return &TraceService{traceRepo: traceRepo}
}

// SaveSpan 保存Span
func (s *TraceService) SaveSpan(span *model.TraceSpan) error {
	return s.traceRepo.SaveSpan(span)
}

// ListTraces 查询Trace列表
func (s *TraceService) ListTraces(serviceName, operationName string, startTime, endTime time.Time, page, pageSize int) ([]model.TraceSpan, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return s.traceRepo.ListTraces(serviceName, operationName, startTime, endTime, pageSize, (page-1)*pageSize)
}

// ListTracesWithFilters 按扩展过滤条件查询Trace列表
func (s *TraceService) ListTracesWithFilters(params repository.TraceQueryParams) ([]model.TraceSpan, int64, error) {
	if params.Limit < 1 {
		params.Limit = 20
	}
	return s.traceRepo.ListTracesWithParams(params)
}

// GetTraceByID 查询Trace详情
func (s *TraceService) GetTraceByID(traceID string) ([]model.TraceSpan, error) {
	return s.traceRepo.GetTraceByID(traceID)
}

// GetTraceList 按TraceID或服务查询
func (s *TraceService) GetTraceList(traceID, serviceName string, page, pageSize int) ([]model.TraceSpan, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return s.traceRepo.GetTraceList(traceID, serviceName, pageSize, (page-1)*pageSize)
}

// GetServiceList 获取服务列表
func (s *TraceService) GetServiceList() ([]string, error) {
	return s.traceRepo.GetServiceList()
}

// GetTraceStats 获取链路统计
func (s *TraceService) GetTraceStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	return s.traceRepo.GetTraceStats(startTime, endTime)
}

// CleanOldTraces 清理旧数据
func (s *TraceService) CleanOldTraces(retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 7
	}
	return s.traceRepo.CleanOldTraces(retentionDays)
}
