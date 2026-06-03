package router

import (
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/monitor/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, monitorHandler *handler.MonitorHandler, podMonitorHandler *handler.PodMonitorHandler, traceHandler *handler.TraceHandler) {
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

		// 应用指标
		metrics.GET("/apps/:appId", podMonitorHandler.GetAppMetrics)
	}

	// Pod监控路由
	pods := api.Group("/pods")
	{
		pods.GET("/:namespace", podMonitorHandler.ListNamespacePods)
		pods.GET("/:namespace/:podName/metrics", podMonitorHandler.GetPodMetrics)
		pods.GET("/:namespace/:podName/logs", podMonitorHandler.GetPodLogs)
	}

	// 日志查询路由
	logs := api.Group("/logs")
	{
		logs.GET("/pods/:podName", podMonitorHandler.GetPodLogs)
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

	// 链路追踪路由
	traces := api.Group("/traces")
	{
		traces.GET("", traceHandler.ListTraces)
		// 具体路径放在通配符 :traceId 之前，避免路由冲突
		traces.GET("/services/list", traceHandler.GetServices)
		traces.GET("/stats", traceHandler.GetTraceStats)
		traces.GET("/apps/:appId", traceHandler.GetTracesByApp)
		traces.GET("/:traceId", traceHandler.GetTrace)
	}

	// 内部 API（无需认证）
	internal := r.Group("/internal/v1")
	{
		// 应用指标
		internal.GET("/metrics/apps/:appId", podMonitorHandler.GetAppMetrics)

		// Pod 监控
		internal.GET("/pods/:namespace", podMonitorHandler.ListNamespacePods)
		internal.GET("/pods/:namespace/:podName/metrics", podMonitorHandler.GetPodMetrics)
		internal.GET("/pods/:namespace/:podName/logs", podMonitorHandler.GetPodLogs)

		// 日志查询
		internal.GET("/logs/pods/:podName", podMonitorHandler.GetPodLogs)
		internal.GET("/logs/apps/:appId", podMonitorHandler.GetAppMetrics)

		// 链路追踪 Span 采集（由 gateway 调用）
		internal.POST("/traces/spans", traceHandler.CollectSpan)
	}
}
