package router

import (
	"my-cloud/internal/project/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, iamDB *gorm.DB) {
	projectHandler := handler.NewProjectHandler(db, iamDB)

	v1 := r.Group("/api/v1")
	{
		// 租户管理路由（Gateway已经做了认证，这里不需要middleware）
		tenants := v1.Group("/tenants")
		{
			tenants.GET("/", projectHandler.ListTenants)
			tenants.POST("/", projectHandler.CreateTenant)
			tenants.GET("/:id/", projectHandler.GetTenant)
			tenants.PUT("/:id/", projectHandler.UpdateTenant)
			tenants.DELETE("/:id/", projectHandler.DeleteTenant)
		}

		// 项目管理路由
		projects := v1.Group("/projects")
		{
			projects.GET("/", projectHandler.ListProjects)
			projects.POST("/", projectHandler.CreateProject)
			projects.GET("/:id/", projectHandler.GetProject)
			projects.PUT("/:id/", projectHandler.UpdateProject)
			projects.DELETE("/:id/", projectHandler.DeleteProject)
		}
		
		// 项目成员管理
		projectMembers := v1.Group("/project-members")
		{
			projectMembers.GET("/:projectId/", projectHandler.GetProjectMembers)
			projectMembers.POST("/:projectId/", projectHandler.AddProjectMember)
			projectMembers.DELETE("/:projectId/:userId/", projectHandler.RemoveProjectMember)
		}
	}
}
