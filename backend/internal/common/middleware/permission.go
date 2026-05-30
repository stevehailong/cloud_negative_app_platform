package middleware

import (
	"my-cloud/internal/auth/repository"
	"my-cloud/internal/common/response"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionCheck 权限检查中间件
func PermissionCheck(db *gorm.DB) gin.HandlerFunc {
	permRepo := repository.NewPermissionRepository(db)
	
	return func(c *gin.Context) {
		// 1. 获取用户ID
		userID, exists := c.Get("userId")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 2. 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 3. 健康检查接口跳过权限检查
		if path == "/health" {
			c.Next()
			return
		}

		// 4. 获取用户的所有权限
		permissions, err := permRepo.GetUserPermissions(userID.(uint))
		if err != nil {
			response.InternalError(c, "获取权限失败")
			c.Abort()
			return
		}

	// 5. 检查是否有权限
	hasPermission := false
	for _, perm := range permissions {
		if matchPath(path, perm.Path) && matchMethod(method, perm.HttpMethod) {
			hasPermission = true
			// 设置匹配的权限信息到上下文
			c.Set("permission", perm.Code)
			break
		}
	}

		if !hasPermission {
			response.Forbidden(c, "无权限访问此资源")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 要求特定权限的中间件
func RequirePermission(db *gorm.DB, requiredPermissions ...string) gin.HandlerFunc {
	permRepo := repository.NewPermissionRepository(db)
	
	return func(c *gin.Context) {
		userID, exists := c.Get("userId")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 获取用户权限
		permissions, err := permRepo.GetUserPermissions(userID.(uint))
		if err != nil {
			response.InternalError(c, "获取权限失败")
			c.Abort()
			return
		}

		// 检查是否有任意一个所需权限
		hasPermission := false
		for _, userPerm := range permissions {
			for _, reqPerm := range requiredPermissions {
				if userPerm.Code == reqPerm {
					hasPermission = true
					c.Set("permission", userPerm.Code)
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			response.Forbidden(c, "缺少必要权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 要求特定角色的中间件（简化版，用于向后兼容）
func RequireRole(db *gorm.DB, allowedRoles ...string) gin.HandlerFunc {
	roleRepo := repository.NewRoleRepository(db)
	
	return func(c *gin.Context) {
		userID, exists := c.Get("userId")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 获取用户角色
		roles, err := roleRepo.GetUserRoles(userID.(uint))
		if err != nil {
			response.InternalError(c, "获取角色失败")
			c.Abort()
			return
		}

		// 检查是否有允许的角色
		hasRole := false
		for _, userRole := range roles {
			for _, allowedRole := range allowedRoles {
				if userRole.Code == allowedRole {
					hasRole = true
					c.Set("role", userRole.Code)
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			response.Forbidden(c, "角色权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}

// matchPath 路径匹配（支持通配符）
func matchPath(requestPath, permPath string) bool {
	// 空路径或 * 匹配所有
	if permPath == "" || permPath == "*" {
		return true
	}
	
	// 去掉尾部斜杠进行标准化比较
	requestPath = strings.TrimRight(requestPath, "/")
	permPath = strings.TrimRight(permPath, "/")
	
	// 完全匹配
	if requestPath == permPath {
		return true
	}
	
	// 段匹配（支持 :param 和 * 作为单段通配符）
	// 必须先于尾部通配符匹配，避免含中间 * 的路径被错误截断
	if strings.Contains(permPath, ":") || strings.Contains(permPath, "*") {
		permParts := strings.Split(permPath, "/")
		reqParts := strings.Split(requestPath, "/")

		if len(permParts) == len(reqParts) {
			match := true
			for i := range permParts {
				if permParts[i] == "*" || strings.HasPrefix(permParts[i], ":") {
					continue
				}
				if permParts[i] != reqParts[i] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}

		// 尾部通配符前缀匹配（如 /api/v1/users*）
		if strings.HasSuffix(permPath, "*") {
			prefix := strings.TrimSuffix(permPath, "*")
			prefix = strings.TrimRight(prefix, "/")
			return requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/")
		}
	}

	return false
}

// matchMethod 方法匹配
func matchMethod(requestMethod, permMethod string) bool {
	// 空方法或 * 匹配所有
	if permMethod == "" || permMethod == "*" {
		return true
	}
	
	// 支持逗号分隔的多个方法
	methods := strings.Split(permMethod, ",")
	for _, m := range methods {
		if strings.TrimSpace(m) == requestMethod {
			return true
		}
	}
	
	return false
}
