package middleware

import (
	"my-cloud/internal/common/response"
	"my-cloud/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth JWT认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未授权，请先登录")
			c.Abort()
			return
		}

		// 验证Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "认证格式无效")
			c.Abort()
			return
		}

		// 解析token
		token := parts[1]
		claims, err := jwt.ParseToken(token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				response.Unauthorized(c, "登录已过期，请重新登录")
			} else {
				response.Unauthorized(c, "认证失败，Token无效")
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（允许未登录访问，但如果有token则解析）
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				if claims, err := jwt.ParseToken(token); err == nil {
					c.Set("userId", claims.UserID)
					c.Set("username", claims.Username)
				}
			}
		}
		c.Next()
	}
}
