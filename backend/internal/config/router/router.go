package router

import (
	"my-cloud/internal/config/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, h *handler.ConfigHandler) {
	api := r.Group("/api/v1")

	// 配置管理 CRUD
	configs := api.Group("/app-configs")
	{
		configs.GET("", h.ListConfigs)
		configs.POST("", h.CreateConfig)
		configs.GET("/:id", h.GetConfig)
		configs.PUT("/:id", h.UpdateConfig)
		configs.DELETE("/:id", h.DeleteConfig)
	}

	// 按应用查询配置
	api.GET("/app-configs/apps/:appId", h.GetConfigsByApp)
}
