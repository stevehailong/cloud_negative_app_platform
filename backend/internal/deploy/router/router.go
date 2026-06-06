package router

import (
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/deploy/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, deployHandler *handler.DeployHandler, appDeployHandler *handler.AppDeploymentHandler, db *gorm.DB) {
	api := r.Group("/api/v1")
	api.Use(middleware.Auth(), middleware.PermissionCheck(db))
	{
		// 部署管理（旧版本，保持兼容）
		api.GET("/deployments", deployHandler.ListDeployments)
		api.POST("/deployments", deployHandler.CreateDeployment)
		api.GET("/deployments/:id", deployHandler.GetDeployment)
		api.POST("/deployments/:id/restart", deployHandler.RestartDeployment)
		api.POST("/deployments/:id/rollback", deployHandler.RollbackDeployment)
		api.POST("/deployments/:id/scale", deployHandler.ScaleDeployment)
		api.GET("/deployments/:id/events", deployHandler.GetDeploymentEvents)
		api.GET("/deployments/:id/pods", deployHandler.GetDeploymentPods)
		api.DELETE("/deployments/pods/:podName", deployHandler.DeletePod)
		api.DELETE("/deployments/:id", deployHandler.DeleteDeployment)

		// 新版部署管理（以app维度）
		api.GET("/app-deployments", appDeployHandler.ListAppDeployments)
		api.GET("/app-deployments/by-app-env", appDeployHandler.ListAppDeploymentsByAppEnv)
		api.DELETE("/app-deployments/cleanup", appDeployHandler.CleanupDuplicateDeployments)
		api.POST("/app-deployments/promote-canary", appDeployHandler.PromoteCanaryToStable)
		api.GET("/app-deployments/:id", appDeployHandler.GetAppDeploymentDetail)
		api.GET("/app-deployments/:id/history", appDeployHandler.GetDeploymentHistory)
		api.GET("/app-deployments/:id/pods", appDeployHandler.GetDeploymentPods)
		api.GET("/app-deployments/:id/events", appDeployHandler.GetDeploymentEvents)
		api.POST("/app-deployments/:id/restart", appDeployHandler.RestartDeployment)
		api.POST("/app-deployments/:id/scale", appDeployHandler.ScaleDeployment)
		api.POST("/app-deployments/:id/rollback", appDeployHandler.RollbackDeployment)
		api.POST("/app-deployments/:id/deploy", appDeployHandler.DeployNewVersion)
		api.POST("/app-deployments/:id/canary/adjust-weight", appDeployHandler.AdjustCanaryWeight)
		api.DELETE("/app-deployments/:id", appDeployHandler.DeleteAppDeployment)
	}

	// 内部服务间调用（无需认证）
	internal := r.Group("/internal/v1")
	{
		internal.POST("/deployments", deployHandler.CreateDeployment)
		internal.GET("/deployments/:id", deployHandler.GetDeployment)
		internal.POST("/deployments/:id/scale", deployHandler.ScaleDeployment)
		internal.POST("/deployments/scale", deployHandler.ScaleDeploymentByName)
		internal.GET("/k8s/deployments/:namespace/:name/replicas", deployHandler.GetK8sDeploymentReplicas)
		internal.DELETE("/k8s/deployments/:namespace/:name", deployHandler.DeleteK8sDeployment)
		internal.DELETE("/deployments/by-workload", deployHandler.DeleteDeploymentsByWorkload)
		
		// 新版app_deployments内部API
		internal.GET("/app-deployments/by-workload", appDeployHandler.GetAppDeploymentByWorkload)
		internal.GET("/app-deployments/canary-weight", appDeployHandler.GetCanaryWeight)
		internal.POST("/app-deployments", appDeployHandler.CreateAppDeploymentInternal)
		internal.POST("/app-deployments/:id/deploy", appDeployHandler.DeployNewVersion)
		internal.POST("/app-deployments/:id/scale", appDeployHandler.ScaleDeployment)
		internal.POST("/app-deployments/:id/canary/adjust-weight", appDeployHandler.AdjustCanaryWeight)
		internal.DELETE("/app-deployments/:id", appDeployHandler.DeleteAppDeployment)
		internal.GET("/deployment-history/:id", appDeployHandler.GetDeploymentHistoryByID)
		
		// 环境信息查询（供release-service使用）
		internal.GET("/environments/:id", appDeployHandler.GetEnvironmentInternal)
	}
}
