package middleware

import (
	"bytes"
	"fmt"
	"io"
	"my-cloud/internal/common/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditMiddleware 审计日志中间件
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 读取请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// 恢复请求体供后续使用
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 创建自定义ResponseWriter来捕获响应
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算请求耗时
		duration := time.Since(startTime)

		// 获取用户信息
		var userID uint
		var username string
		if uid, exists := c.Get("userId"); exists {
			if id, ok := uid.(uint); ok {
				userID = id
			}
		}
		if uname, exists := c.Get("username"); exists {
			if name, ok := uname.(string); ok {
				username = name
			}
		}

		// 跳过不需要审计的路径
		path := c.Request.URL.Path
		if shouldSkipAudit(path) {
			return
		}

		// 确定操作类型
		action := getActionFromMethod(c.Request.Method)

		// 获取资源类型和资源名称
		resourceType, resourceID := parseResourceFromPath(path)

		// 脱敏请求体中的敏感信息
		sanitizedBody := sanitizeRequestBody(requestBody)

		// 创建审计日志记录
		auditLog := &model.AuditLog{
			UserID:          userID,
			Username:        username,
			Action:          action,
			ResourceType:    resourceType,
			ResourceID:      resourceID,
			Method:          c.Request.Method,
			Path:            path,
			IPAddress:       c.ClientIP(),
			UserAgent:       c.Request.UserAgent(),
			RequestBody:     sanitizedBody,
			ResponseCode:    c.Writer.Status(),
			ResponseMessage: getResponseMessage(c.Writer.Status()),
			DurationMs:      int(duration.Milliseconds()),
		}

		// 异步写入数据库(避免阻塞请求)
		go func() {
			// 使用已初始化的auditDB，不要创建新连接
			db.Create(auditLog)
		}()
	}
}

// bodyLogWriter 用于捕获响应体的自定义Writer
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldSkipAudit 判断是否跳过审计
func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/api/v1/auth/login",    // 登录请求已有日志
		"/api/v1/auth/refresh",  // Token刷新
		"/api/v1/audit-logs",    // 审计日志查询本身
		"/health",               // 健康检查
		"/metrics",              // 监控指标
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// 只审计POST/PUT/DELETE操作和重要的GET操作
	return false
}

// getActionFromMethod 根据HTTP方法确定操作类型
func getActionFromMethod(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	case "GET":
		return "view"
	default:
		return "other"
	}
}

// parseResourceFromPath 从路径中解析资源类型和ID
func parseResourceFromPath(path string) (string, *uint) {
	// 移除/api/v1前缀
	path = strings.TrimPrefix(path, "/api/v1/")

	// 分割路径
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "unknown", nil
	}

	// 资源类型是第一部分
	resourceType := parts[0]

	// 如果有第二部分且是数字,则为资源ID
	var resourceID *uint
	if len(parts) > 1 {
		// 尝试解析为数字
		var id uint
		if _, err := fmt.Sscanf(parts[1], "%d", &id); err == nil {
			resourceID = &id
		}
	}

	// 标准化资源类型名称
	resourceType = strings.TrimSuffix(resourceType, "s") // 移除复数s
	resourceType = strings.ReplaceAll(resourceType, "-", "_")

	return resourceType, resourceID
}

// sanitizeRequestBody 脱敏请求体中的敏感信息
func sanitizeRequestBody(body string) string {
	if body == "" {
		return ""
	}

	// 限制请求体大小
	maxLength := 5000
	if len(body) > maxLength {
		body = body[:maxLength] + "...(truncated)"
	}

	// 脱敏密码字段
	sensitiveFields := []string{"password", "token", "secret", "apiKey", "accessToken"}
	for _, field := range sensitiveFields {
		// 简单的字符串替换(生产环境应使用JSON解析)
		if strings.Contains(strings.ToLower(body), field) {
			// 这里简化处理,生产环境应该解析JSON后替换
			body = strings.ReplaceAll(body, field, field+"\":\"***REDACTED***")
		}
	}

	return body
}

// getResponseMessage 根据状态码获取响应消息
func getResponseMessage(statusCode int) string {
	messages := map[int]string{
		200: "success",
		201: "created",
		204: "no content",
		400: "bad request",
		401: "unauthorized",
		403: "forbidden",
		404: "not found",
		500: "internal server error",
	}

	if msg, ok := messages[statusCode]; ok {
		return msg
	}
	return "unknown"
}
