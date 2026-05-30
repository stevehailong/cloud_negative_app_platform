package handler

import (
	"my-cloud/internal/application/service"
	"my-cloud/internal/common/model"
	"my-cloud/internal/common/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApplicationHandler struct {
	appService *service.ApplicationService
}

func NewApplicationHandler(appService *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{
		appService: appService,
	}
}

// CreateApplication 创建应用
type CreateApplicationRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	ProjectID   uint   `json:"projectId" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	RepoURL     string `json:"repoUrl"`
	RepoBranch  string `json:"repoBranch"`
	BuildTool   string `json:"buildTool"`
	BuildPath   string `json:"buildPath"`
	DockerFile  string `json:"dockerFile"`
	Owner       string `json:"owner"`
}

func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	username, _ := c.Get("username")
	app := &model.Application{
		Name:        req.Name,
		Code:        req.Code,
		ProjectID:   req.ProjectID,
		Description: req.Description,
		Type:        req.Type,
		Language:    req.Language,
		Framework:   req.Framework,
		RepoURL:     req.RepoURL,
		RepoBranch:  req.RepoBranch,
		BuildTool:   req.BuildTool,
		BuildPath:   req.BuildPath,
		DockerFile:  req.DockerFile,
		Owner:       req.Owner,
	}
	app.CreatedBy = username.(string)

	if err := h.appService.CreateApplication(app); err != nil {
		response.Error(c, response.CodeConflict, err.Error())
		return
	}

	response.Success(c, app)
}

// GetApplication 获取应用详情
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的应用ID")
		return
	}

	app, components, err := h.appService.GetApplication(uint(id))
	if err != nil {
		response.NotFound(c, "应用不存在")
		return
	}

	response.Success(c, gin.H{
		"application": app,
		"components":  components,
	})
}

// UpdateApplication 更新应用
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的应用ID")
		return
	}

	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	username, _ := c.Get("username")
	app := &model.Application{
		Name:        req.Name,
		Code:        req.Code,
		ProjectID:   req.ProjectID,
		Description: req.Description,
		Type:        req.Type,
		Language:    req.Language,
		Framework:   req.Framework,
		RepoURL:     req.RepoURL,
		RepoBranch:  req.RepoBranch,
		BuildTool:   req.BuildTool,
		BuildPath:   req.BuildPath,
		DockerFile:  req.DockerFile,
		Owner:       req.Owner,
	}
	app.ID = uint(id)
	app.UpdatedBy = username.(string)

	if err := h.appService.UpdateApplication(app); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, app)
}

// DeleteApplication 删除应用
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的应用ID")
		return
	}

	if err := h.appService.DeleteApplication(uint(id)); err != nil {
		response.Error(c, response.CodeConflict, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// ListApplications 应用列表
func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 32)
	keyword := c.Query("keyword")

	apps, total, err := h.appService.ListApplications(page, pageSize, uint(projectID), keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, apps)
}

// CreateComponent 创建组件
type CreateComponentRequest struct {
	ApplicationID uint   `json:"applicationId" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Type          string `json:"type"`
	Version       string `json:"version"`
	Image         string `json:"image"`
	Port          int    `json:"port"`
	Replicas      int    `json:"replicas"`
	CPU           string `json:"cpu"`
	Memory        string `json:"memory"`
	EnvVars       string `json:"envVars"`
	ConfigMaps    string `json:"configMaps"`
	Secrets       string `json:"secrets"`
	Volumes       string `json:"volumes"`
}

func (h *ApplicationHandler) CreateComponent(c *gin.Context) {
	var req CreateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	username, _ := c.Get("username")
	component := &model.Component{
		ApplicationID: req.ApplicationID,
		Name:          req.Name,
		Type:          req.Type,
		Version:       req.Version,
		Image:         req.Image,
		Port:          req.Port,
		Replicas:      req.Replicas,
		CPU:           req.CPU,
		Memory:        req.Memory,
		EnvVars:       req.EnvVars,
		ConfigMaps:    req.ConfigMaps,
		Secrets:       req.Secrets,
		Volumes:       req.Volumes,
	}
	component.CreatedBy = username.(string)

	if err := h.appService.CreateComponent(component); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, component)
}

// GetComponent 获取组件详情
func (h *ApplicationHandler) GetComponent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的组件ID")
		return
	}

	component, err := h.appService.GetComponent(uint(id))
	if err != nil {
		response.NotFound(c, "组件不存在")
		return
	}

	response.Success(c, component)
}

// UpdateComponent 更新组件
func (h *ApplicationHandler) UpdateComponent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的组件ID")
		return
	}

	var req CreateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	username, _ := c.Get("username")
	component := &model.Component{
		ApplicationID: req.ApplicationID,
		Name:          req.Name,
		Type:          req.Type,
		Version:       req.Version,
		Image:         req.Image,
		Port:          req.Port,
		Replicas:      req.Replicas,
		CPU:           req.CPU,
		Memory:        req.Memory,
		EnvVars:       req.EnvVars,
		ConfigMaps:    req.ConfigMaps,
		Secrets:       req.Secrets,
		Volumes:       req.Volumes,
	}
	component.ID = uint(id)
	component.UpdatedBy = username.(string)

	if err := h.appService.UpdateComponent(component); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, component)
}

// DeleteComponent 删除组件
func (h *ApplicationHandler) DeleteComponent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的组件ID")
		return
	}

	if err := h.appService.DeleteComponent(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// ListComponents 组件列表
func (h *ApplicationHandler) ListComponents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	appID, _ := strconv.ParseUint(c.Query("applicationId"), 10, 32)

	components, total, err := h.appService.ListComponents(page, pageSize, uint(appID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, components)
}
