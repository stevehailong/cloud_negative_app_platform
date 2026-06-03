package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"my-cloud/internal/common/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ServiceProxy struct {
	targetURL string
	client    *http.Client
	// monitorURL 用于上报 trace span，如 http://monitor-service:8090
	monitorURL string
}

func NewServiceProxy(targetURL string) *ServiceProxy {
	return &ServiceProxy{
		targetURL:  targetURL,
		client:     &http.Client{},
		monitorURL: "http://monitor-service:8090",
	}
}

func (p *ServiceProxy) Handle(c *gin.Context) {
	startTime := time.Now()

	// 生成或复用 trace_id（复用 X-Request-Id）
	traceID := c.GetHeader("X-Request-Id")
	if traceID == "" {
		traceID = uuid.New().String()
		c.Request.Header.Set("X-Request-Id", traceID)
	}
	spanID := uuid.New().String()

	// 设置 parentSpanId 到请求头，以便下游服务创建子 Span
	c.Request.Header.Set("X-Parent-Span-Id", spanID)

	// 获取服务名（从 targetURL 的 host 推断，如 auth-service:8081 -> auth-service）
	serviceName := extractServiceName(p.targetURL)
	operationName := c.Request.URL.Path
	method := c.Request.Method

	// 获取完整路径并构建目标URL
	fullPath := c.Request.URL.Path
	targetURL := p.targetURL + fullPath

	// 添加查询参数
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// 读取请求体
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 创建新请求
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		response.InternalError(c, "failed to create request")
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		response.InternalError(c, "failed to proxy request: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		response.InternalError(c, "failed to read response")
		return
	}

	// 返回响应
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

	// 异步记录 trace span
	go p.recordSpan(traceID, spanID, "", serviceName, operationName, method, startTime, time.Now(), resp.StatusCode)
}

// recordSpan 异步上报 span 到 monitor-service
func (p *ServiceProxy) recordSpan(traceID, spanID, parentSpanID, serviceName, operationName, method string, startTime, endTime time.Time, statusCode int) {
	if p.monitorURL == "" {
		return
	}

	durationMs := uint32(endTime.Sub(startTime).Milliseconds())
	hasError := 0
	if statusCode >= 400 {
		hasError = 1
	}

	span := map[string]interface{}{
		"traceId":       traceID,
		"spanId":        spanID,
		"parentSpanId":  parentSpanID,
		"serviceName":   serviceName,
		"operationName": operationName,
		"method":        method,
		"durationMs":    durationMs,
		"startTime":     startTime.Format(time.RFC3339Nano),
		"endTime":       endTime.Format(time.RFC3339Nano),
		"statusCode":    statusCode,
		"hasError":      hasError,
	}

	body, _ := json.Marshal(span)
	req, err := http.NewRequest("POST", p.monitorURL+"/internal/v1/traces/spans", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// 使用短超时，避免阻塞
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

// extractServiceName 从 targetURL 提取服务名，如 http://auth-service:8081 -> auth-service
func extractServiceName(targetURL string) string {
	// 去掉协议前缀
	s := strings.TrimPrefix(targetURL, "http://")
	s = strings.TrimPrefix(s, "https://")
	// 取 host:port 前的部分
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}
	// 去掉端口
	if idx := strings.Index(s, ":"); idx != -1 {
		s = s[:idx]
	}
	return s
}

// ProxyWithPath 带路径前缀的代理
func (p *ServiceProxy) ProxyWithPath(pathPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		originalPath := c.Request.URL.Path
		c.Request.URL.Path = strings.TrimPrefix(originalPath, pathPrefix)
		p.Handle(c)
	}
}
