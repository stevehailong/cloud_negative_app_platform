package handler

import (
	"my-cloud/internal/common/model"
	"my-cloud/internal/common/response"
	"my-cloud/internal/config/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configRepo *repository.ConfigRepository
}

func NewConfigHandler(configRepo *repository.ConfigRepository) *ConfigHandler {
	return &ConfigHandler{
		configRepo: configRepo,
	}
}

// ListConfigs 配置列表
func (h *ConfigHandler) ListConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	var appID, envID uint
	if appIDStr := c.Query("appId"); appIDStr != "" {
		id, _ := strconv.ParseUint(appIDStr, 10, 32)
		appID = uint(id)
	}
	if envIDStr := c.Query("envId"); envIDStr != "" {
		id, _ := strconv.ParseUint(envIDStr, 10, 32)
		envID = uint(id)
	}

	configs, total, err := h.configRepo.List(appID, envID, keyword, page, pageSize)
	if err != nil {
		response.InternalError(c, "查询配置列表失败")
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, configs)
}

// CreateConfig 创建配置
func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var req model.AppConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, "请求参数错误")
		return
	}

	if err := h.configRepo.Create(&req); err != nil {
		response.InternalError(c, "创建配置失败")
		return
	}

	response.Success(c, req)
}

// GetConfig 获取配置详情
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的配置ID")
		return
	}

	config, err := h.configRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "配置不存在")
		return
	}

	response.Success(c, config)
}

// UpdateConfig 更新配置
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的配置ID")
		return
	}

	existing, err := h.configRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "配置不存在")
		return
	}

	var req model.AppConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, "请求参数错误")
		return
	}

	req.ID = uint(id)
	req.CreateTime = existing.CreateTime

	if err := h.configRepo.Update(&req); err != nil {
		response.InternalError(c, "更新配置失败")
		return
	}

	response.Success(c, req)
}

// DeleteConfig 删除配置
func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的配置ID")
		return
	}

	if err := h.configRepo.Delete(uint(id)); err != nil {
		response.InternalError(c, "删除配置失败")
		return
	}

	response.Success(c, nil)
}

// GetConfigsByApp 根据应用获取配置列表
func (h *ConfigHandler) GetConfigsByApp(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的应用ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	var envID uint
	if envIDStr := c.Query("envId"); envIDStr != "" {
		id, _ := strconv.ParseUint(envIDStr, 10, 32)
		envID = uint(id)
	}

	configs, total, err := h.configRepo.GetByAppAndEnv(uint(appID), envID, page, pageSize)
	if err != nil {
		response.InternalError(c, "查询配置列表失败")
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, configs)
}
