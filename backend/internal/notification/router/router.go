package router

import (
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/notification/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, notificationHandler *handler.NotificationHandler) {
	api := r.Group("/api/v1")
	api.Use(middleware.Auth())

	// 通知相关路由
	notifications := api.Group("/notifications")
	{
		notifications.POST("", notificationHandler.SendNotification)
		notifications.POST("/template", notificationHandler.SendByTemplate)
		notifications.GET("", notificationHandler.ListNotifications)
		notifications.GET("/:id", notificationHandler.GetNotification)
	}

	// 通知模板相关路由
	templates := api.Group("/notification-templates")
	{
		templates.POST("", notificationHandler.CreateTemplate)
		templates.GET("", notificationHandler.ListTemplates)
		templates.GET("/:id", notificationHandler.GetTemplate)
		templates.PUT("/:id", notificationHandler.UpdateTemplate)
		templates.DELETE("/:id", notificationHandler.DeleteTemplate)
	}

	// 通知渠道相关路由
	channels := api.Group("/notification-channels")
	{
		channels.POST("", notificationHandler.CreateChannel)
		channels.GET("", notificationHandler.ListChannels)
		channels.PUT("/:id", notificationHandler.UpdateChannel)
		channels.DELETE("/:id", notificationHandler.DeleteChannel)
	}
}
