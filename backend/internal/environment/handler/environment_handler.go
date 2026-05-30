package handler

import (
	"net/http"
	"strconv"

	"my-cloud/internal/common/model"
	"my-cloud/internal/environment/repository"

	"github.com/gin-gonic/gin"
)

type EnvironmentHandler struct {
	envRepo     *repository.EnvironmentRepository
	templateRepo *repository.EnvTemplateRepository
	bindingRepo  *repository.AppEnvBindingRepository
}

func NewEnvironmentHandler(envRepo *repository.EnvironmentRepository, templateRepo *repository.EnvTemplateRepository, bindingRepo *repository.AppEnvBindingRepository) *EnvironmentHandler {
	return &EnvironmentHandler{
		envRepo:     envRepo,
		templateRepo: templateRepo,
		bindingRepo:  bindingRepo,
	}
}

// ListEnvironments 环境列表
func (h *EnvironmentHandler) ListEnvironments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	projectIDStr := c.Query("projectId")

	offset := (page - 1) * pageSize

	var projectID *uint
	if projectIDStr != "" {
		id, _ := strconv.ParseUint(projectIDStr, 10, 32)
		pid := uint(id)
		projectID = &pid
	}

	envs, total, err := h.envRepo.List(offset, pageSize, keyword, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询环境列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     envs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateEnvironment 创建环境
func (h *EnvironmentHandler) CreateEnvironment(c *gin.Context) {
	var req model.Environment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 检查环境编码是否已存在
	existing, _ := h.envRepo.GetByCode(req.EnvCode)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "环境编码已存在",
		})
		return
	}

	if err := h.envRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建环境失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetEnvironment 获取环境详情
func (h *EnvironmentHandler) GetEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的环境ID",
		})
		return
	}

	env, err := h.envRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "环境不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    env,
	})
}

// UpdateEnvironment 更新环境
func (h *EnvironmentHandler) UpdateEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的环境ID",
		})
		return
	}

	env, err := h.envRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "环境不存在",
		})
		return
	}

	var req model.Environment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = env.CreateTime

	if err := h.envRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新环境失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteEnvironment 删除环境
func (h *EnvironmentHandler) DeleteEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的环境ID",
		})
		return
	}

	if err := h.envRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除环境失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ListTemplates 模板列表
func (h *EnvironmentHandler) ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	templateTypeStr := c.Query("templateType")

	offset := (page - 1) * pageSize

	var templateType *string
	if templateTypeStr != "" {
		templateType = &templateTypeStr
	}

	templates, total, err := h.templateRepo.List(offset, pageSize, keyword, templateType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询模板列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     templates,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateTemplate 创建模板
func (h *EnvironmentHandler) CreateTemplate(c *gin.Context) {
	var req model.EnvTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 检查模板编码是否已存在
	existing, _ := h.templateRepo.GetByCode(req.TemplateCode)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "模板编码已存在",
		})
		return
	}

	if err := h.templateRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建模板失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetTemplate 获取模板详情
func (h *EnvironmentHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的模板ID",
		})
		return
	}

	template, err := h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "模板不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    template,
	})
}

// UpdateTemplate 更新模板
func (h *EnvironmentHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的模板ID",
		})
		return
	}

	template, err := h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "模板不存在",
		})
		return
	}

	var req model.EnvTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = template.CreateTime

	if err := h.templateRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新模板失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteTemplate 删除模板
func (h *EnvironmentHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的模板ID",
		})
		return
	}

	if err := h.templateRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除模板失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ListBindings 绑定列表
func (h *EnvironmentHandler) ListBindings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	appIDStr := c.Query("appId")
	envIDStr := c.Query("envId")

	offset := (page - 1) * pageSize

	var appID, envID *uint
	if appIDStr != "" {
		id, _ := strconv.ParseUint(appIDStr, 10, 32)
		aid := uint(id)
		appID = &aid
	}
	if envIDStr != "" {
		id, _ := strconv.ParseUint(envIDStr, 10, 32)
		eid := uint(id)
		envID = &eid
	}

	bindings, total, err := h.bindingRepo.List(offset, pageSize, appID, envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询绑定列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     bindings,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateBinding 创建绑定
func (h *EnvironmentHandler) CreateBinding(c *gin.Context) {
	var req model.AppEnvBinding
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 检查绑定是否已存在
	existing, _ := h.bindingRepo.GetByAppAndEnv(req.AppID, req.EnvID)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "该应用已绑定此环境",
		})
		return
	}

	if err := h.bindingRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建绑定失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetBinding 获取绑定详情
func (h *EnvironmentHandler) GetBinding(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的绑定ID",
		})
		return
	}

	binding, err := h.bindingRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "绑定不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    binding,
	})
}

// UpdateBinding 更新绑定
func (h *EnvironmentHandler) UpdateBinding(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的绑定ID",
		})
		return
	}

	binding, err := h.bindingRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "绑定不存在",
		})
		return
	}

	var req model.AppEnvBinding
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = binding.CreateTime

	if err := h.bindingRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新绑定失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteBinding 删除绑定
func (h *EnvironmentHandler) DeleteBinding(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的绑定ID",
		})
		return
	}

	if err := h.bindingRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除绑定失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}
