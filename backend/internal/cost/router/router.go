package router

import (
	"my-cloud/internal/cost/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, h *handler.CostHandler) {
	v1 := r.Group("/api/v1")
	{
		costs := v1.Group("/costs")
		{
			costs.GET("", h.ListCostRecords)
			costs.GET("/overview", h.GetCostOverview)
		costs.POST("/sync", h.SyncCostData)
			costs.GET("/projects/:projectId", h.GetCostByProject)
			costs.GET("/apps/:appId", h.GetCostByApp)
		}
	}

	internal := r.Group("/internal/v1")
	{
		internal.POST("/costs/sync", h.SyncCostData)
	}
}
