package router

import (
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/monitor/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, monitorHandler *handler.MonitorHandler) {
	api := r.Group("/api/v1")
	api.Use(middleware.Auth())

	// Metrics相关路由
	metrics := api.Group("/metrics")
	{
		metrics.POST("", monitorHandler.CreateMetric)
		metrics.GET("", monitorHandler.ListMetrics)
		metrics.GET("/:id", monitorHandler.GetMetric)
		metrics.PUT("/:id", monitorHandler.UpdateMetric)
		metrics.DELETE("/:id", monitorHandler.DeleteMetric)
	}

	// AlertRules相关路由
	alertRules := api.Group("/alert-rules")
	{
		alertRules.POST("", monitorHandler.CreateAlertRule)
		alertRules.GET("", monitorHandler.ListAlertRules)
		alertRules.GET("/:id", monitorHandler.GetAlertRule)
		alertRules.PUT("/:id", monitorHandler.UpdateAlertRule)
		alertRules.DELETE("/:id", monitorHandler.DeleteAlertRule)
	}

	// Alerts相关路由
	alerts := api.Group("/alerts")
	{
		alerts.GET("", monitorHandler.ListAlerts)
		alerts.GET("/:id", monitorHandler.GetAlert)
		alerts.POST("/:id/resolve", monitorHandler.ResolveAlert)
		alerts.GET("/statistics", monitorHandler.GetAlertStatistics)
	}
}
