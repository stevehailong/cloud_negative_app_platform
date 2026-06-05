package model

import "time"

// Metric 指标模型
type Metric struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Type        string    `gorm:"size:20;not null" json:"type"` // counter/gauge/histogram/summary
	Description string    `gorm:"size:500" json:"description"`
	Unit        string    `gorm:"size:20" json:"unit"`
	Labels      string    `gorm:"type:json" json:"labels"` // JSON格式标签
	Enabled     int       `gorm:"default:1" json:"enabled"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (Metric) TableName() string {
	return "metrics"
}

// AlertRule 告警规则模型
type AlertRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;index" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	MetricName  string    `gorm:"size:100;not null;index" json:"metricName"`
	Condition   string    `gorm:"size:50;not null" json:"condition"` // >/</==/>=/<=/!=
	Threshold   float64   `gorm:"not null" json:"threshold"`
	Duration    int       `gorm:"not null" json:"duration"`          // 持续时间(秒)
	Severity    string    `gorm:"size:20;not null" json:"severity"`  // critical/warning/info
	Enabled     int       `gorm:"default:1" json:"enabled"`
	NotifyUsers string    `gorm:"type:text" json:"notifyUsers"` // 通知用户列表(逗号分隔)
	CreateTime  time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (AlertRule) TableName() string {
	return "alert_rules"
}

// Alert 告警记录模型
type Alert struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	RuleID      uint       `gorm:"not null;index" json:"ruleId"`
	RuleName    string     `gorm:"size:100;not null" json:"ruleName"`
	MetricName  string     `gorm:"size:100;not null;index" json:"metricName"`
	CurrentValue float64   `gorm:"not null" json:"currentValue"`
	Threshold   float64    `gorm:"not null" json:"threshold"`
	Severity    string     `gorm:"size:20;not null;index" json:"severity"`
	Status      string     `gorm:"size:20;not null;index" json:"status"` // firing/resolved
	Message     string     `gorm:"type:text" json:"message"`
	FiredAt     time.Time  `json:"firedAt"`
	ResolvedAt  *time.Time `json:"resolvedAt,omitempty"`
	CreateTime  time.Time  `gorm:"autoCreateTime;index" json:"createTime"`
}

func (Alert) TableName() string {
	return "alerts"
}

// LogQuery 日志查询模型
type LogQuery struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Query       string    `gorm:"type:text;not null" json:"query"` // Loki LogQL查询语句
	Labels      string    `gorm:"type:json" json:"labels"`
	UserID      uint      `gorm:"not null;index" json:"userId"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (LogQuery) TableName() string {
	return "log_queries"
}

// TraceQuery 链路追踪查询模型
type TraceQuery struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	ServiceName string    `gorm:"size:100;not null;index" json:"serviceName"`
	Operation   string    `gorm:"size:100" json:"operation"`
	MinDuration int       `json:"minDuration"` // 最小持续时间(ms)
	MaxDuration int       `json:"maxDuration"` // 最大持续时间(ms)
	UserID      uint      `gorm:"not null;index" json:"userId"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (TraceQuery) TableName() string {
	return "trace_queries"
}

// TraceSpan 链路追踪Span模型
type TraceSpan struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	TraceID      string     `gorm:"size:64;not null;index" json:"traceId"`
	SpanID       string     `gorm:"size:64;not null" json:"spanId"`
	ParentSpanID string     `gorm:"size:64" json:"parentSpanId,omitempty"`
	ServiceName  string     `gorm:"size:100;not null;index" json:"serviceName"`
	OperationName string    `gorm:"size:255;not null" json:"operationName"`
	Method       string     `gorm:"size:10" json:"method"`
	DurationMs   uint32     `json:"durationMs"`
	StartTime    time.Time  `gorm:"not null;index" json:"startTime"`
	EndTime      *time.Time `json:"endTime,omitempty"`
	StatusCode   int        `json:"statusCode"`
	Tags         *string    `gorm:"type:json" json:"tags,omitempty"`
	HasError     int        `gorm:"default:0" json:"hasError"`
}

func (TraceSpan) TableName() string {
	return "trace_spans"
}
