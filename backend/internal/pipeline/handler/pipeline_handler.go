package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"my-cloud/internal/common/response"
	"my-cloud/internal/pipeline/model"
	"my-cloud/internal/pipeline/service"
	"my-cloud/pkg/gitlab"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PipelineHandler struct {
	pipelineService *service.PipelineService
}

func NewPipelineHandler(pipelineService *service.PipelineService) *PipelineHandler {
	return &PipelineHandler{
		pipelineService: pipelineService,
	}
}

// CreatePipeline 创建流水线
type CreatePipelineRequest struct {
	PipelineCode string `json:"pipelineCode" binding:"required"`
	AppID        uint   `json:"appId" binding:"required"`
	PipelineName string `json:"pipelineName" binding:"required"`
	PipelineType string `json:"pipelineType" binding:"required"`
	CITool       string `json:"ciTool"`
	ConfigJSON   string `json:"configJson"`
}

func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	pipeline := &model.Pipeline{
		PipelineCode: req.PipelineCode,
		AppID:        req.AppID,
		PipelineName: req.PipelineName,
		PipelineType: req.PipelineType,
		CITool:       req.CITool,
		ConfigJSON:   req.ConfigJSON,
		Enabled:      1,
	}

	if pipeline.CITool == "" {
		pipeline.CITool = "jenkins"
	}

	if err := h.pipelineService.CreatePipeline(pipeline); err != nil {
		response.Error(c, response.CodeConflict, err.Error())
		return
	}

	response.Success(c, pipeline)
}

// GetPipeline 获取流水线详情
func (h *PipelineHandler) GetPipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的流水线ID")
		return
	}

	pipeline, err := h.pipelineService.GetPipeline(uint(id))
	if err != nil {
		response.NotFound(c, "流水线不存在")
		return
	}

	response.Success(c, pipeline)
}

// ListPipelines 获取流水线列表
func (h *PipelineHandler) ListPipelines(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	appID, _ := strconv.ParseUint(c.Query("appId"), 10, 32)

	pipelines, total, err := h.pipelineService.ListPipelines(uint(appID), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, pipelines)
}

// UpdatePipeline 更新流水线
type UpdatePipelineRequest struct {
	PipelineName string `json:"pipelineName"`
	PipelineType string `json:"pipelineType"`
	CITool       string `json:"ciTool"`
	ConfigJSON   string `json:"configJson"`
	Enabled      *int   `json:"enabled"`
}

func (h *PipelineHandler) UpdatePipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的流水线ID")
		return
	}

	var req UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	pipeline, err := h.pipelineService.GetPipeline(uint(id))
	if err != nil {
		response.NotFound(c, "流水线不存在")
		return
	}

	// 更新字段
	if req.PipelineName != "" {
		pipeline.PipelineName = req.PipelineName
	}
	if req.PipelineType != "" {
		pipeline.PipelineType = req.PipelineType
	}
	if req.CITool != "" {
		pipeline.CITool = req.CITool
	}
	if req.ConfigJSON != "" {
		pipeline.ConfigJSON = req.ConfigJSON
	}
	if req.Enabled != nil {
		pipeline.Enabled = *req.Enabled
	}

	if err := h.pipelineService.UpdatePipeline(pipeline); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, pipeline)
}

// DeletePipeline 删除流水线
func (h *PipelineHandler) DeletePipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的流水线ID")
		return
	}

	if err := h.pipelineService.DeletePipeline(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// RunPipeline 触发流水线执行
type RunPipelineRequest struct {
	TriggerType string `json:"triggerType" binding:"required"` // manual/webhook/mr/schedule
	GitCommit   string `json:"gitCommit"`
	GitBranch   string `json:"gitBranch"`
}

func (h *PipelineHandler) RunPipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的流水线ID")
		return
	}

	var req RunPipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	// 手动触发时，若未指定分支则默认使用 main
	if req.GitBranch == "" {
		req.GitBranch = "main"
	}

	// 获取操作用户ID
	userID, _ := c.Get("userId")
	operatorUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		operatorUserID = uid
	}

	run, err := h.pipelineService.RunPipeline(uint(id), req.TriggerType, req.GitCommit, req.GitBranch, operatorUserID)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, run)
}

// DeployPipeline 手动触发部署
func (h *PipelineHandler) DeployPipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的流水线ID")
		return
	}

	// 获取操作用户ID
	userID, _ := c.Get("userId")
	operatorUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		operatorUserID = uid
	}

	result, err := h.pipelineService.DeployPipeline(uint(id), operatorUserID)
	if err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetPipelineRun 获取流水线执行详情
func (h *PipelineHandler) GetPipelineRun(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("runId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的执行ID")
		return
	}

	run, err := h.pipelineService.GetPipelineRun(uint(id))
	if err != nil {
		response.NotFound(c, "执行记录不存在")
		return
	}

	response.Success(c, run)
}

// ListPipelineRuns 获取流水线执行记录列表
// ListAllPipelineRuns 获取所有流水线执行记录列表（不限定pipelineID）
func (h *PipelineHandler) ListAllPipelineRuns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	startDate := c.Query("startDate")
	sortBy := c.DefaultQuery("sortBy", "createTime")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	runs, total, err := h.pipelineService.ListAllPipelineRuns(page, pageSize, startDate, sortBy, sortOrder)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 为每个 run 查询关联的 image artifact
	runResponses := make([]*PipelineRunResponse, 0, len(runs))
	log.Printf("[DEBUG] ListAllPipelineRuns: 查询到 %d 条记录", len(runs))
	for _, run := range runs {
		runResp := &PipelineRunResponse{
			PipelineRun: run,
		}
		
		// 查询该 run 的 artifacts，找到 type=image 的制品
		artifacts, err := h.pipelineService.GetArtifactsByRunID(run.ID)
		log.Printf("[DEBUG] Run %s (ID=%d): 查询到 %d 个 artifacts, err=%v", run.RunNo, run.ID, len(artifacts), err)
		if err == nil {
			for _, artifact := range artifacts {
				log.Printf("[DEBUG]   - Artifact: type=%s, url=%s", artifact.ArtifactType, artifact.RepoURL)
				if artifact.ArtifactType == "image" {
					runResp.ImageURL = artifact.RepoURL
					log.Printf("[DEBUG] ✓ 设置 imageUrl = %s", artifact.RepoURL)
					break
				}
			}
		}
		
		runResponses = append(runResponses, runResp)
	}
	log.Printf("[DEBUG] ListAllPipelineRuns: 返回 %d 条响应", len(runResponses))

	response.SuccessWithPage(c, total, page, pageSize, runResponses)
}

// PipelineRunResponse 包含 imageUrl 的响应结构
type PipelineRunResponse struct {
	*model.PipelineRun
	ImageURL     string `json:"imageUrl,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func (h *PipelineHandler) ListPipelineRuns(c *gin.Context) {
	pipelineID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的流水线ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	runs, total, err := h.pipelineService.ListPipelineRuns(uint(pipelineID), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 为每个 run 查询关联的 image artifact
	runResponses := make([]*PipelineRunResponse, 0, len(runs))
	for _, run := range runs {
		runResp := &PipelineRunResponse{
			PipelineRun: run,
		}
		
		// 查询该 run 的 artifacts，找到 type=image 的制品
		artifacts, err := h.pipelineService.GetArtifactsByRunID(run.ID)
		if err == nil {
			for _, artifact := range artifacts {
				if artifact.ArtifactType == "image" {
					runResp.ImageURL = artifact.RepoURL
					break
				}
			}
		}
		
		runResponses = append(runResponses, runResp)
	}

	response.SuccessWithPage(c, total, page, pageSize, runResponses)
}

// GetPipelineRunLogs 获取流水线执行日志（模拟）
func (h *PipelineHandler) GetPipelineRunLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("runId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的执行ID")
		return
	}

	run, err := h.pipelineService.GetPipelineRun(uint(id))
	if err != nil {
		response.NotFound(c, "执行记录不存在")
		return
	}

	// TODO: 从Jenkins/Tekton获取实际日志
	// 这里返回模拟日志
	response.Success(c, gin.H{
		"runId":  run.ID,
		"runNo":  run.RunNo,
		"logUrl": run.LogURL,
		"logs":   "Pipeline execution logs...\n[INFO] Starting pipeline...\n[INFO] Build successful",
	})
}

// ListArtifacts 获取制品列表
func (h *PipelineHandler) ListArtifacts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	artifactType := c.Query("artifactType")

	artifacts, total, err := h.pipelineService.ListArtifacts(artifactType, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, artifacts)
}

// GetArtifact 获取制品详情
func (h *PipelineHandler) GetArtifact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的制品ID")
		return
	}

	artifact, err := h.pipelineService.GetArtifact(uint(id))
	if err != nil {
		response.NotFound(c, "制品不存在")
		return
	}

	response.Success(c, artifact)
}

// DeleteArtifact 删除制品
func (h *PipelineHandler) DeleteArtifact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的制品ID")
		return
	}

	if err := h.pipelineService.DeleteArtifact(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// ============ GitLab Integration Handlers ============

// TestGitlabConnection 测试GitLab连接
func (h *PipelineHandler) TestGitlabConnection(c *gin.Context) {
	client := h.pipelineService.GetGitlabClient()
	if client == nil {
		response.Error(c, response.CodeInternalError, "GitLab未配置，请先在系统设置中配置GitLab地址和Token")
		return
	}

	if err := client.Ping(); err != nil {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("GitLab连接失败: %v", err))
		return
	}

	user, err := client.GetCurrentUser()
	if err != nil {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("获取用户信息失败: %v", err))
		return
	}

	response.Success(c, gin.H{
		"connected": true,
		"username":  user["username"],
		"name":      user["name"],
	})
}

// ListGitlabProjects 列出GitLab项目
func (h *PipelineHandler) ListGitlabProjects(c *gin.Context) {
	client := h.pipelineService.GetGitlabClient()
	if client == nil {
		response.Error(c, response.CodeInternalError, "GitLab未配置")
		return
	}

	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	projects, err := client.ListProjects(search, page, perPage)
	if err != nil {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("获取项目列表失败: %v", err))
		return
	}

	response.Success(c, projects)
}

// ListGitlabBranches 列出项目的分支
func (h *PipelineHandler) ListGitlabBranches(c *gin.Context) {
	client := h.pipelineService.GetGitlabClient()
	if client == nil {
		response.Error(c, response.CodeInternalError, "GitLab未配置")
		return
	}

	projectID := c.Param("projectId")
	if projectID == "" {
		response.InvalidParams(c, "缺少项目ID")
		return
	}

	search := c.Query("search")
	branches, err := client.ListBranches(projectID, search)
	if err != nil {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("获取分支列表失败: %v", err))
		return
	}

	response.Success(c, branches)
}

// GetGitlabLatestCommit 获取分支最新提交
func (h *PipelineHandler) GetGitlabLatestCommit(c *gin.Context) {
	client := h.pipelineService.GetGitlabClient()
	if client == nil {
		response.Error(c, response.CodeInternalError, "GitLab未配置")
		return
	}

	projectID := c.Param("projectId")
	branch := c.Query("branch")
	if projectID == "" || branch == "" {
		response.InvalidParams(c, "缺少项目ID或分支名")
		return
	}

	commit, err := client.GetLatestCommit(projectID, branch)
	if err != nil {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("获取最新提交失败: %v", err))
		return
	}

	response.Success(c, commit)
}

// CreateGitlabWebhook 为项目创建Webhook
func (h *PipelineHandler) CreateGitlabWebhook(c *gin.Context) {
	client := h.pipelineService.GetGitlabClient()
	if client == nil {
		response.Error(c, response.CodeInternalError, "GitLab未配置")
		return
	}

	var req struct {
		ProjectID string `json:"projectId" binding:"required"`
		HookURL   string `json:"hookUrl" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	hook, err := client.CreateWebhook(req.ProjectID, req.HookURL, true, true)
	if err != nil {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("创建Webhook失败: %v", err))
		return
	}

	response.Success(c, hook)
}

// GitlabWebhookCallback 处理GitLab Webhook回调
func (h *PipelineHandler) GitlabWebhookCallback(c *gin.Context) {
	// 验证webhook token
	token := c.GetHeader("X-Gitlab-Token")
	if token != "my-cloud-webhook-secret" {
		response.Error(c, response.CodeForbidden, "无效的Webhook Token")
		return
	}

	eventType := c.GetHeader("X-Gitlab-Event")
	if eventType != "Push Hook" {
		// 仅处理Push事件
		response.Success(c, gin.H{"message": "event ignored", "event": eventType})
		return
	}

	var pushEvent gitlab.PushEvent
	if err := c.ShouldBindJSON(&pushEvent); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	// 从ref中提取分支名 (refs/heads/main -> main)
	branch := strings.TrimPrefix(pushEvent.Ref, "refs/heads/")

	// 获取最新的commit
	gitCommit := pushEvent.After
	if len(pushEvent.Commits) > 0 {
		gitCommit = pushEvent.Commits[0].ID
	}

	log.Printf("[GitLab Webhook] Push event: project=%s branch=%s commit=%s",
		pushEvent.Project.PathWithNamespace, branch, gitCommit)

	// 查找关联的流水线并触发
	triggered := h.triggerPipelinesForRepo(pushEvent.Project.WebURL, branch, gitCommit)

	response.Success(c, gin.H{
		"message":   "webhook processed",
		"branch":    branch,
		"commit":    gitCommit,
		"triggered": triggered,
	})
}

// triggerPipelinesForRepo 根据仓库URL触发关联的流水线
func (h *PipelineHandler) triggerPipelinesForRepo(repoURL, branch, commit string) int {
	// 获取所有流水线，检查configJson中的repoUrl是否匹配
	pipelines, _, err := h.pipelineService.ListPipelines(0, 1, 100)
	if err != nil {
		log.Printf("[GitLab Webhook] Failed to list pipelines: %v", err)
		return 0
	}

	triggered := 0
	for _, pipeline := range pipelines {
		if pipeline.ConfigJSON == "" || pipeline.Enabled != 1 {
			continue
		}

		var config map[string]interface{}
		if err := json.Unmarshal([]byte(pipeline.ConfigJSON), &config); err != nil {
			continue
		}

		configRepoURL, _ := config["repoUrl"].(string)
		if configRepoURL == "" {
			continue
		}

		// 匹配仓库URL
		if !strings.Contains(repoURL, strings.TrimSuffix(configRepoURL, ".git")) &&
			!strings.Contains(configRepoURL, repoURL) {
			continue
		}

		// 检查分支过滤（如果配置了defaultBranch，只在该分支触发）
		defaultBranch, _ := config["defaultBranch"].(string)
		if defaultBranch != "" && defaultBranch != branch {
			continue
		}

		// 触发流水线
		_, err := h.pipelineService.RunPipeline(pipeline.ID, "webhook", commit, branch, 0)
		if err != nil {
			log.Printf("[GitLab Webhook] Failed to trigger pipeline %s: %v", pipeline.PipelineCode, err)
			continue
		}
		log.Printf("[GitLab Webhook] Triggered pipeline %s for branch %s", pipeline.PipelineCode, branch)
		triggered++
	}

	return triggered
}

// UpdateGitlabClient 动态更新GitLab客户端配置
func (h *PipelineHandler) UpdateGitlabClient(c *gin.Context) {
	var req struct {
		GitlabURL   string `json:"gitlabUrl" binding:"required"`
		GitlabToken string `json:"gitlabToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	client := gitlab.NewClient(req.GitlabURL, req.GitlabToken)
	if err := client.Ping(); err != nil {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("GitLab连接失败: %v", err))
		return
	}

	h.pipelineService.SetGitlabClient(client)
	response.Success(c, gin.H{"message": "GitLab配置已更新"})
}

// JenkinsBuildCallback Jenkins构建完成回调，更新制品的实际镜像地址
func (h *PipelineHandler) JenkinsBuildCallback(c *gin.Context) {
	runNo := c.Param("runNo")
	var req struct {
		ImageURL string `json:"imageUrl" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}
	if err := h.pipelineService.UpdateLatestArtifactImage(runNo, req.ImageURL); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "artifact updated", "imageUrl": req.ImageURL})
}
