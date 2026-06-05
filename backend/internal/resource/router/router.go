package router

import (
	"my-cloud/internal/resource/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, h *handler.ResourceHandler) {
	v1 := r.Group("/api/v1")
	{
		resourceGroup := v1.Group("/resource-quotas")
		{
			resourceGroup.GET("", h.ListResourceQuotas)
			resourceGroup.POST("", h.CreateResourceQuota)
			resourceGroup.GET("/:id", h.GetResourceQuota)
			resourceGroup.PUT("/:id", h.UpdateResourceQuota)
			resourceGroup.DELETE("/:id", h.DeleteResourceQuota)
		}
	}

	internal := r.Group("/internal/v1")
	{
		internal.POST("/resource-quotas/sync", h.SyncFromK8s)
	}
}
