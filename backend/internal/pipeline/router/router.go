package router

import (
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/pipeline/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, pipelineHandler *handler.PipelineHandler, db *gorm.DB) {
	api := r.Group("/api/v1")
	api.Use(middleware.Auth())
	{
		// 流水线管理
		api.GET("/pipelines", pipelineHandler.ListPipelines)
		api.POST("/pipelines", pipelineHandler.CreatePipeline)
		api.GET("/pipelines/:id", pipelineHandler.GetPipeline)
		api.PUT("/pipelines/:id", pipelineHandler.UpdatePipeline)
		api.DELETE("/pipelines/:id", pipelineHandler.DeletePipeline)
		api.POST("/pipelines/:id/run", pipelineHandler.RunPipeline)
		api.POST("/pipelines/:id/deploy", pipelineHandler.DeployPipeline)

		// 流水线执行记录
		api.GET("/pipeline-runs", pipelineHandler.ListAllPipelineRuns)
		api.GET("/pipelines/:id/runs", pipelineHandler.ListPipelineRuns)
		api.GET("/pipeline-runs/:runId", pipelineHandler.GetPipelineRun)
		api.GET("/pipeline-runs/:runId/logs", pipelineHandler.GetPipelineRunLogs)

		// 制品管理
		api.GET("/artifacts", pipelineHandler.ListArtifacts)
		api.GET("/artifacts/:id", pipelineHandler.GetArtifact)
		api.DELETE("/artifacts/:id", pipelineHandler.DeleteArtifact)

		// GitLab集成
		api.POST("/gitlab/test", pipelineHandler.TestGitlabConnection)
		api.GET("/gitlab/projects", pipelineHandler.ListGitlabProjects)
		api.GET("/gitlab/projects/:projectId/branches", pipelineHandler.ListGitlabBranches)
		api.GET("/gitlab/projects/:projectId/commit", pipelineHandler.GetGitlabLatestCommit)
		api.POST("/gitlab/webhooks", pipelineHandler.CreateGitlabWebhook)
		api.PUT("/gitlab/client", pipelineHandler.UpdateGitlabClient)
	}

	// GitLab Webhook回调（不需要认证）
	r.POST("/hooks/gitlab", pipelineHandler.GitlabWebhookCallback)

	// Jenkins 构建回调（内部服务，不需要认证）
	internal := r.Group("/internal/v1")
	{
		internal.POST("/pipeline-runs/:runNo/artifact", pipelineHandler.JenkinsBuildCallback)
	}
}
