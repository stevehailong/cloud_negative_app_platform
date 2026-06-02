package router

import (
	"my-cloud/internal/audit/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, auditHandler *handler.AuditHandler) {
	api := r.Group("/api/v1")
	// 审计服务不需要Auth中间件，因为gateway已经做了认证
	// api.Use(middleware.Auth())

	// 审计日志相关路由
	auditLogs := api.Group("/audit-logs")
	{
		auditLogs.GET("", auditHandler.ListAuditLogs)                                    // 获取审计日志列表
		auditLogs.GET("/:id", auditHandler.GetAuditLog)                                  // 获取审计日志详情
		auditLogs.GET("/resource/:resourceType/:resourceId", auditHandler.GetAuditLogsByResource) // 根据资源获取审计日志
		auditLogs.GET("/user/:userId", auditHandler.GetAuditLogsByUser)                 // 根据用户获取审计日志
		auditLogs.GET("/statistics", auditHandler.GetStatistics)                         // 获取统计信息
		auditLogs.GET("/export", auditHandler.ExportAuditLogs)                          // 导出审计日志
		auditLogs.POST("/clean", auditHandler.CleanOldLogs)                             // 清理过期日志
	}
}
