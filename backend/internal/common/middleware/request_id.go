package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取
		requestID := c.GetHeader("X-Request-Id")
		
		// 如果没有则生成新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到上下文和响应头
		c.Set("requestId", requestID)
		c.Writer.Header().Set("X-Request-Id", requestID)

		c.Next()
	}
}
