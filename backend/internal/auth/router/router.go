package router

import (
	"my-cloud/internal/auth/handler"
	"my-cloud/internal/common/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler, db *gorm.DB) {
	// 创建权限处理器
	permHandler := handler.NewPermissionHandler(db)
	// 创建设置处理器
	settingsHandler := handler.NewSettingsHandler(db)

	api := r.Group("/api/v1")
	{
		// 公开路由 - 不需要认证
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken) // Token刷新不需要认证
		}

		// 需要认证的路由 - 只需要登录，不需要特定权限
		authSecure := api.Group("/auth")
		authSecure.Use(middleware.Auth())
		{
			authSecure.GET("/userinfo", authHandler.GetUserInfo)
			authSecure.PUT("/password", authHandler.UpdatePassword)
			authSecure.PUT("/profile", authHandler.UpdateProfile)
			authSecure.POST("/logout", authHandler.Logout)
		}

		// 用户管理路由 - 需要权限控制
		users := api.Group("/users")
		users.Use(middleware.Auth(), middleware.PermissionCheck(db))
		{
			users.GET("/", authHandler.GetUserList)
			users.POST("/", authHandler.CreateUser)
			users.GET("/:id/", authHandler.GetUserByID)
			users.PUT("/:id/", authHandler.UpdateUser)
			users.DELETE("/:id/", authHandler.DeleteUser)
			users.PUT("/:id/status/", authHandler.UpdateUserStatus)
			users.POST("/assign-roles/", authHandler.AssignRoles)
			users.GET("/:id/roles/", authHandler.GetUserRoles)
			users.GET("/:id/permissions/", permHandler.GetUserPermissions)
		}

		// 角色管理路由 - 需要权限控制
		roles := api.Group("/roles")
		roles.Use(middleware.Auth(), middleware.PermissionCheck(db))
		{
			roles.GET("/", authHandler.GetRoleList)
			roles.POST("/", authHandler.CreateRole)
			roles.GET("/:roleId/permissions/", permHandler.GetRolePermissions)
			roles.POST("/:roleId/permissions/", permHandler.AssignPermissionsToRole)
			roles.DELETE("/:roleId/permissions/:permId/", permHandler.RemovePermissionFromRole)
		}

		// 权限管理路由 - 需要权限控制
		permissions := api.Group("/permissions")
		permissions.Use(middleware.Auth(), middleware.PermissionCheck(db))
		{
			permissions.GET("/", permHandler.GetPermissionList)
			permissions.GET("/:id/", permHandler.GetPermissionByID)
			permissions.POST("/", permHandler.CreatePermission)
			permissions.PUT("/:id/", permHandler.UpdatePermission)
			permissions.DELETE("/:id/", permHandler.DeletePermission)
		}

		// 系统设置路由 - 需要认证和权限
		settings := api.Group("/settings")
		settings.Use(middleware.Auth(), middleware.PermissionCheck(db))
		{
			settings.GET("/", settingsHandler.GetAllSettings)
			settings.GET("/:group/", settingsHandler.GetSettings)
			settings.PUT("/:group/", settingsHandler.UpdateSettings)
		}

		// 文件上传路由 - 需要认证
		upload := api.Group("/upload")
		upload.Use(middleware.Auth())
		{
			upload.POST("/", settingsHandler.UploadFile)
		}

		// 文件访问路由 - 公开访问（图片等静态资源）
		api.GET("/uploads/*filename", settingsHandler.ServeFile)
	}
}
