package repository

import (
	"time"

	"my-cloud/internal/monitor/model"

	"gorm.io/gorm"
)

// TraceRepository 链路追踪数据访问层
type TraceRepository struct {
	db *gorm.DB
}

func NewTraceRepository(db *gorm.DB) *TraceRepository {
	return &TraceRepository{db: db}
}

// TraceQueryParams 链路查询参数
type TraceQueryParams struct {
	ServiceName   string
	OperationName string
	Method        string
	MinDuration   int // ms
	MaxDuration   int // ms
	HasError      *int // nil=不限, 0=正常, 1=错误
	StartTime     time.Time
	EndTime       time.Time
	Limit         int
	Offset        int
}

// SaveSpan 保存单个Span
func (r *TraceRepository) SaveSpan(span *model.TraceSpan) error {
	return r.db.Create(span).Error
}

// ListTraces 按条件查询Trace列表（按 trace_id 去重，取每个 trace 的根span）
func (r *TraceRepository) ListTraces(serviceName, operationName string, startTime, endTime time.Time, limit, offset int) ([]model.TraceSpan, int64, error) {
	return r.ListTracesWithParams(TraceQueryParams{
		ServiceName:   serviceName,
		OperationName: operationName,
		StartTime:     startTime,
		EndTime:       endTime,
		Limit:         limit,
		Offset:        offset,
	})
}

// ListTracesWithParams 按扩展参数查询Trace列表
func (r *TraceRepository) ListTracesWithParams(params TraceQueryParams) ([]model.TraceSpan, int64, error) {
	var spans []model.TraceSpan
	var total int64

	// 子查询：取每个 trace_id 最新的 span 作为代表
	subQuery := r.db.Model(&model.TraceSpan{}).
		Select("trace_id, MAX(start_time) as max_time").
		Group("trace_id")

	query := r.db.Table("trace_spans as t").
		Joins("JOIN (?) as latest ON t.trace_id = latest.trace_id AND t.start_time = latest.max_time", subQuery)

	query = applyTraceFilters(query, params)

	query.Count(&total)
	err := query.Order("t.start_time DESC").Limit(params.Limit).Offset(params.Offset).Find(&spans).Error
	return spans, total, err
}

// GetTraceByID 查询指定TraceID的所有Span
func (r *TraceRepository) GetTraceByID(traceID string) ([]model.TraceSpan, error) {
	var spans []model.TraceSpan
	err := r.db.Where("trace_id = ?", traceID).Order("start_time ASC").Find(&spans).Error
	return spans, err
}

// GetTraceList 按TraceID模糊查询或按应用查询
func (r *TraceRepository) GetTraceList(traceID, serviceName string, limit, offset int) ([]model.TraceSpan, int64, error) {
	var spans []model.TraceSpan
	var total int64

	query := r.db.Model(&model.TraceSpan{})
	if traceID != "" {
		query = query.Where("trace_id = ?", traceID)
	}
	if serviceName != "" {
		query = query.Where("service_name = ?", serviceName)
	}

	query.Count(&total)
	err := query.Order("start_time DESC").Limit(limit).Offset(offset).Find(&spans).Error
	return spans, total, err
}

// GetServiceList 获取所有产生过 trace 的服务名列表
func (r *TraceRepository) GetServiceList() ([]string, error) {
	var services []string
	err := r.db.Model(&model.TraceSpan{}).
		Distinct("service_name").
		Order("service_name").
		Pluck("service_name", &services).Error
	return services, err
}

// GetTraceStats 获取链路统计信息
func (r *TraceRepository) GetTraceStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总 trace 数
	var totalTraces int64
	r.db.Model(&model.TraceSpan{}).
		Select("COUNT(DISTINCT trace_id)").
		Where("start_time >= ? AND start_time <= ?", startTime, endTime).
		Scan(&totalTraces)
	stats["totalTraces"] = totalTraces

	// 平均耗时
	var avgDuration float64
	r.db.Model(&model.TraceSpan{}).
		Select("AVG(duration_ms)").
		Where("start_time >= ? AND start_time <= ?", startTime, endTime).
		Scan(&avgDuration)
	stats["avgDurationMs"] = avgDuration

	// 错误率
	var errorCount int64
	var totalSpans int64
	r.db.Model(&model.TraceSpan{}).
		Where("start_time >= ? AND start_time <= ?", startTime, endTime).
		Count(&totalSpans)
	r.db.Model(&model.TraceSpan{}).
		Where("start_time >= ? AND start_time <= ? AND has_error = 1", startTime, endTime).
		Count(&errorCount)
	if totalSpans > 0 {
		stats["errorRate"] = float64(errorCount) / float64(totalSpans) * 100
	} else {
		stats["errorRate"] = float64(0)
	}

	// 各服务 span 数量
	type ServiceCount struct {
		ServiceName string `json:"serviceName"`
		Count       int64  `json:"count"`
	}
	var serviceCounts []ServiceCount
	r.db.Model(&model.TraceSpan{}).
		Select("service_name, COUNT(*) as count").
		Where("start_time >= ? AND start_time <= ?", startTime, endTime).
		Group("service_name").
		Order("count DESC").
		Limit(10).
		Scan(&serviceCounts)
	stats["topServices"] = serviceCounts

	return stats, nil
}

// CleanOldTraces 清理超过 retentionDays 天的旧 trace 数据
func (r *TraceRepository) CleanOldTraces(retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := r.db.Where("start_time < ?", cutoff).Delete(&model.TraceSpan{})
	return result.RowsAffected, result.Error
}

// applyTraceFilters 应用通用过滤条件
func applyTraceFilters(query *gorm.DB, params TraceQueryParams) *gorm.DB {
	if params.ServiceName != "" {
		query = query.Where("t.service_name = ?", params.ServiceName)
	}
	if params.OperationName != "" {
		query = query.Where("t.operation_name LIKE ?", "%"+params.OperationName+"%")
	}
	if params.Method != "" {
		query = query.Where("t.method = ?", params.Method)
	}
	if params.MinDuration > 0 {
		query = query.Where("t.duration_ms >= ?", params.MinDuration)
	}
	if params.MaxDuration > 0 {
		query = query.Where("t.duration_ms <= ?", params.MaxDuration)
	}
	if params.HasError != nil {
		query = query.Where("t.has_error = ?", *params.HasError)
	}
	if !params.StartTime.IsZero() {
		query = query.Where("t.start_time >= ?", params.StartTime)
	}
	if !params.EndTime.IsZero() {
		query = query.Where("t.start_time <= ?", params.EndTime)
	}
	return query
}
