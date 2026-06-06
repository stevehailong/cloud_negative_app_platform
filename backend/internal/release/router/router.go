package router

import (
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/release/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, releaseHandler *handler.ReleaseHandler, db *gorm.DB) {
	api := r.Group("/api/v1")
	api.Use(middleware.Auth(), middleware.PermissionCheck(db))
	{
		// 发布工单管理
		api.GET("/releases", releaseHandler.ListReleases)
		api.POST("/releases", releaseHandler.CreateRelease)
		api.GET("/releases/:id", releaseHandler.GetRelease)
		api.PUT("/releases/:id", releaseHandler.UpdateRelease)
		api.POST("/releases/:id/submit", releaseHandler.SubmitRelease)
		api.POST("/releases/:id/approve", releaseHandler.ApproveRelease)
		api.POST("/releases/:id/reject", releaseHandler.RejectRelease)
		api.POST("/releases/:id/execute", releaseHandler.ExecuteRelease)
		api.POST("/releases/:id/rollback", releaseHandler.RollbackRelease)
		api.POST("/releases/:id/canary/confirm", releaseHandler.ConfirmCanary)
		api.POST("/releases/:id/canary/rollback", releaseHandler.RollbackCanary)
		api.POST("/releases/:id/canary/adjust-weight", releaseHandler.AdjustCanaryWeight)

		// 审批记录
		api.GET("/releases/:id/approvals", releaseHandler.ListReleaseApprovals)
	}

	// 内部服务间调用（无需认证，通过 X-User-ID 传递操作人）
	internal := r.Group("/internal/v1")
	{
		internal.POST("/releases", releaseHandler.CreateRelease)
		internal.POST("/releases/:id/execute", releaseHandler.ExecuteRelease)
		internal.POST("/releases/:id/canary/confirm", releaseHandler.ConfirmCanary)
		internal.POST("/releases/:id/canary/rollback", releaseHandler.RollbackCanary)
	}
}
