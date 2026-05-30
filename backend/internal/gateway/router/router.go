package router

import (
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/gateway/proxy"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api/v1")
	
	// 初始化代理
	authProxy := proxy.NewServiceProxy("http://auth-service:8081")
	projectProxy := proxy.NewServiceProxy("http://project-service:8082")
	appProxy := proxy.NewServiceProxy("http://application-service:8083")
	componentProxy := proxy.NewServiceProxy("http://application-service:8083")
	pipelineProxy := proxy.NewServiceProxy("http://pipeline-service:8084")
	envProxy := proxy.NewServiceProxy("http://env-service:8085")
	releaseProxy := proxy.NewServiceProxy("http://release-service:8086")
	deployProxy := proxy.NewServiceProxy("http://deploy-service:8087")
	clusterProxy := proxy.NewServiceProxy("http://cluster-service:8088")
	resourceProxy := proxy.NewServiceProxy("http://resource-service:8089")
	monitorProxy := proxy.NewServiceProxy("http://monitor-service:8090")
	alertProxy := proxy.NewServiceProxy("http://monitor-service:8090")
	auditProxy := proxy.NewServiceProxy("http://audit-service:8093")
	notificationProxy := proxy.NewServiceProxy("http://notification-service:8095")
	costProxy := proxy.NewServiceProxy("http://cost-service:8096")

	// 认证服务路由（公开）
	api.Any("/auth/*path", authProxy.Handle)

	// 需要认证和权限检查的路由
	authenticated := api.Group("")
	authenticated.Use(middleware.Auth(), middleware.PermissionCheck(db))
	{
		// 用户、角色、权限管理
		authenticated.Any("/users", authProxy.Handle)
		authenticated.Any("/users/*path", authProxy.Handle)
		authenticated.Any("/roles", authProxy.Handle)
		authenticated.Any("/roles/*path", authProxy.Handle)
		authenticated.Any("/permissions", authProxy.Handle)
		authenticated.Any("/permissions/*path", authProxy.Handle)
		
		// 项目组织管理
		authenticated.Any("/projects", projectProxy.Handle)
		authenticated.Any("/projects/*path", projectProxy.Handle)
		authenticated.Any("/project-members", projectProxy.Handle)
		authenticated.Any("/project-members/*path", projectProxy.Handle)
		authenticated.Any("/tenants", projectProxy.Handle)
		authenticated.Any("/tenants/*path", projectProxy.Handle)
		authenticated.Any("/organizations", projectProxy.Handle)
		authenticated.Any("/organizations/*path", projectProxy.Handle)
		
		// 应用管理
		authenticated.Any("/applications", appProxy.Handle)
		authenticated.Any("/applications/*path", appProxy.Handle)
		authenticated.Any("/components", componentProxy.Handle)
		authenticated.Any("/components/*path", componentProxy.Handle)
		
		// 流水线
		authenticated.Any("/pipelines", pipelineProxy.Handle)
		authenticated.Any("/pipelines/*path", pipelineProxy.Handle)
		authenticated.Any("/pipeline-runs", pipelineProxy.Handle)
		authenticated.Any("/pipeline-runs/*path", pipelineProxy.Handle)
		authenticated.Any("/artifacts", pipelineProxy.Handle)
		authenticated.Any("/artifacts/*path", pipelineProxy.Handle)

		// GitLab集成（通过pipeline-service）
		authenticated.Any("/gitlab", pipelineProxy.Handle)
		authenticated.Any("/gitlab/*path", pipelineProxy.Handle)
		
		// 环境
		authenticated.Any("/environments", envProxy.Handle)
		authenticated.Any("/environments/*path", envProxy.Handle)
		authenticated.Any("/env-templates", envProxy.Handle)
		authenticated.Any("/env-templates/*path", envProxy.Handle)
		authenticated.Any("/app-env-bindings", envProxy.Handle)
		authenticated.Any("/app-env-bindings/*path", envProxy.Handle)
		authenticated.Any("/config-maps", envProxy.Handle)
		authenticated.Any("/config-maps/*path", envProxy.Handle)
		authenticated.Any("/secrets", envProxy.Handle)
		authenticated.Any("/secrets/*path", envProxy.Handle)
		
		// 发布
		authenticated.Any("/releases", releaseProxy.Handle)
		authenticated.Any("/releases/*path", releaseProxy.Handle)
		
		// 部署（旧版）
		authenticated.Any("/deployments", deployProxy.Handle)
		authenticated.Any("/deployments/*path", deployProxy.Handle)
		
		// 应用部署（新版）
		authenticated.Any("/app-deployments", deployProxy.Handle)
		authenticated.Any("/app-deployments/*path", deployProxy.Handle)
		
		// 集群
		authenticated.Any("/clusters", clusterProxy.Handle)
		authenticated.Any("/clusters/*path", clusterProxy.Handle)
		authenticated.Any("/nodes", clusterProxy.Handle)
		authenticated.Any("/nodes/*path", clusterProxy.Handle)
		authenticated.Any("/namespaces", clusterProxy.Handle)
		authenticated.Any("/namespaces/*path", clusterProxy.Handle)
		
		// 资源
		authenticated.Any("/resources", resourceProxy.Handle)
		authenticated.Any("/resources/*path", resourceProxy.Handle)
		
		// 监控
		authenticated.Any("/monitors", monitorProxy.Handle)
		authenticated.Any("/monitors/*path", monitorProxy.Handle)
		authenticated.Any("/metrics", monitorProxy.Handle)
		authenticated.Any("/metrics/*path", monitorProxy.Handle)
		authenticated.Any("/alert-rules", alertProxy.Handle)
		authenticated.Any("/alert-rules/*path", alertProxy.Handle)
		authenticated.Any("/alerts", alertProxy.Handle)
		authenticated.Any("/alerts/*path", alertProxy.Handle)
		
		// 审计
		authenticated.Any("/audit-logs", auditProxy.Handle)
		authenticated.Any("/audit-logs/*path", auditProxy.Handle)
		
		// 通知
		authenticated.Any("/notifications", notificationProxy.Handle)
		authenticated.Any("/notifications/*path", notificationProxy.Handle)
		authenticated.Any("/notification-templates", notificationProxy.Handle)
		authenticated.Any("/notification-templates/*path", notificationProxy.Handle)
		authenticated.Any("/notification-channels", notificationProxy.Handle)
		authenticated.Any("/notification-channels/*path", notificationProxy.Handle)
		
		// 成本
		authenticated.Any("/costs", costProxy.Handle)
		authenticated.Any("/costs/*path", costProxy.Handle)

		// 系统设置
		authenticated.Any("/settings", authProxy.Handle)
		authenticated.Any("/settings/*path", authProxy.Handle)

		// 文件上传
		authenticated.Any("/upload", authProxy.Handle)
		authenticated.Any("/upload/*path", authProxy.Handle)
	}

	// 公开文件访问（无需认证）
	api.GET("/uploads/*path", authProxy.Handle)

	// GitLab Webhook回调（无需认证，由webhook secret验证）
	// 使用独立前缀避免与 /gitlab/*path 通配符冲突
	r.POST("/hooks/gitlab", pipelineProxy.Handle)
}
