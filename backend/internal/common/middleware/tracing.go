package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderTraceID      = "X-Request-Id"
	HeaderParentSpanID = "X-Parent-Span-Id"
)

// getMonitorURL 获取 monitor-service 地址
func getMonitorURL() string {
	if url := os.Getenv("MONITOR_SERVICE_URL"); url != "" {
		return url
	}
	return "http://monitor-service:8090"
}

// SpanReport 上报 Span 的 JSON 结构
type SpanReport struct {
	TraceID       string `json:"traceId"`
	SpanID        string `json:"spanId"`
	ParentSpanID  string `json:"parentSpanId,omitempty"`
	ServiceName   string `json:"serviceName"`
	OperationName string `json:"operationName"`
	Method        string `json:"method"`
	DurationMs    uint32 `json:"durationMs"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	StatusCode    int    `json:"statusCode"`
	HasError      int    `json:"hasError"`
}

// Tracing 链路追踪中间件，为每个请求创建一个 Span 并异步上报
// serviceName: 当前服务名称，如 "auth-service"、"application-service"
func Tracing(serviceName string) gin.HandlerFunc {
	monitorURL := getMonitorURL()

	return func(c *gin.Context) {
		// 跳过内部 trace 上报端点，避免无限循环
		if c.Request.URL.Path == "/internal/v1/traces/spans" {
			c.Next()
			return
		}
		// 跳过 health 检查
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		startTime := time.Now()

		// 读取或生成 traceId（复用 X-Request-Id）
		traceID := c.GetHeader(HeaderTraceID)
		if traceID == "" {
			traceID = uuid.New().String()
			c.Request.Header.Set(HeaderTraceID, traceID)
		}

		// 读取 parentSpanId（由上游服务在 X-Parent-Span-Id 中传入）
		parentSpanID := c.GetHeader(HeaderParentSpanID)

		// 生成本次请求的 spanId
		spanID := uuid.New().String()

		// 存储到 Gin Context，供 handler 使用
		c.Set("traceId", traceID)
		c.Set("spanId", spanID)
		c.Set("parentSpanId", parentSpanID)

		// 设置响应头，方便客户端追踪
		c.Writer.Header().Set(HeaderTraceID, traceID)

		c.Next()

		// 异步上报 Span
		go reportSpan(monitorURL, SpanReport{
			TraceID:       traceID,
			SpanID:        spanID,
			ParentSpanID:  parentSpanID,
			ServiceName:   serviceName,
			OperationName: c.Request.URL.Path,
			Method:        c.Request.Method,
			DurationMs:    uint32(time.Since(startTime).Milliseconds()),
			StartTime:     startTime.Format(time.RFC3339Nano),
			EndTime:       time.Now().Format(time.RFC3339Nano),
			StatusCode:    c.Writer.Status(),
			HasError:      boolToInt(isServerError(c.Writer.Status())),
		})
	}
}

// reportSpan 异步上报单个 Span 到 monitor-service
func reportSpan(monitorURL string, span SpanReport) {
	body, err := json.Marshal(span)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", monitorURL+"/internal/v1/traces/spans", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isServerError 判断是否为服务端错误（5xx），排除客户端错误（4xx）
func isServerError(statusCode int) bool {
	return statusCode >= 500
}