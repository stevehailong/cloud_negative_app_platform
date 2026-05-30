package router

import (
	"my-cloud/internal/application/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, appHandler *handler.ApplicationHandler) {
	api := r.Group("/api/v1")
	{
		// 应用管理
		apps := api.Group("/applications")
		{
			apps.POST("", appHandler.CreateApplication)
			apps.GET("/:id", appHandler.GetApplication)
			apps.PUT("/:id", appHandler.UpdateApplication)
			apps.DELETE("/:id", appHandler.DeleteApplication)
			apps.GET("", appHandler.ListApplications)
		}

		// 组件管理
		components := api.Group("/components")
		{
			components.POST("", appHandler.CreateComponent)
			components.GET("/:id", appHandler.GetComponent)
			components.PUT("/:id", appHandler.UpdateComponent)
			components.DELETE("/:id", appHandler.DeleteComponent)
			components.GET("", appHandler.ListComponents)
		}
	}
}
