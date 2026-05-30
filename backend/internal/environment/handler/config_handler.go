package handler

import (
	"net/http"
	"strconv"

	"my-cloud/internal/common/model"
	"my-cloud/internal/environment/repository"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configMapRepo *repository.ConfigMapRepository
	secretRepo    *repository.SecretRepository
}

func NewConfigHandler(configMapRepo *repository.ConfigMapRepository, secretRepo *repository.SecretRepository) *ConfigHandler {
	return &ConfigHandler{
		configMapRepo: configMapRepo,
		secretRepo:    secretRepo,
	}
}

// ListConfigMaps ConfigMap列表
func (h *ConfigHandler) ListConfigMaps(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	envIDStr := c.Query("envId")
	namespace := c.Query("namespace")

	offset := (page - 1) * pageSize

	var envID *uint
	if envIDStr != "" {
		id, _ := strconv.ParseUint(envIDStr, 10, 32)
		eid := uint(id)
		envID = &eid
	}

	configMaps, total, err := h.configMapRepo.List(offset, pageSize, envID, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询ConfigMap列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     configMaps,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateConfigMap 创建ConfigMap
func (h *ConfigHandler) CreateConfigMap(c *gin.Context) {
	var req model.ConfigMap
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 检查是否已存在
	existing, _ := h.configMapRepo.GetByEnvAndName(req.EnvID, req.Name)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "该环境下已存在同名ConfigMap",
		})
		return
	}

	if err := h.configMapRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建ConfigMap失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetConfigMap 获取ConfigMap详情
func (h *ConfigHandler) GetConfigMap(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的ID",
		})
		return
	}

	configMap, err := h.configMapRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "ConfigMap不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    configMap,
	})
}

// UpdateConfigMap 更新ConfigMap
func (h *ConfigHandler) UpdateConfigMap(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的ID",
		})
		return
	}

	configMap, err := h.configMapRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "ConfigMap不存在",
		})
		return
	}

	var req model.ConfigMap
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = configMap.CreateTime

	if err := h.configMapRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新ConfigMap失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteConfigMap 删除ConfigMap
func (h *ConfigHandler) DeleteConfigMap(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的ID",
		})
		return
	}

	if err := h.configMapRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除ConfigMap失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ListSecrets Secret列表
func (h *ConfigHandler) ListSecrets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	envIDStr := c.Query("envId")
	namespace := c.Query("namespace")

	offset := (page - 1) * pageSize

	var envID *uint
	if envIDStr != "" {
		id, _ := strconv.ParseUint(envIDStr, 10, 32)
		eid := uint(id)
		envID = &eid
	}

	secrets, total, err := h.secretRepo.List(offset, pageSize, envID, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询Secret列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     secrets,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateSecret 创建Secret
func (h *ConfigHandler) CreateSecret(c *gin.Context) {
	var req model.Secret
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 检查是否已存在
	existing, _ := h.secretRepo.GetByEnvAndName(req.EnvID, req.Name)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "该环境下已存在同名Secret",
		})
		return
	}

	if err := h.secretRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建Secret失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetSecret 获取Secret详情
func (h *ConfigHandler) GetSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的ID",
		})
		return
	}

	secret, err := h.secretRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "Secret不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    secret,
	})
}

// UpdateSecret 更新Secret
func (h *ConfigHandler) UpdateSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的ID",
		})
		return
	}

	secret, err := h.secretRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "Secret不存在",
		})
		return
	}

	var req model.Secret
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = secret.CreateTime

	if err := h.secretRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新Secret失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteSecret 删除Secret
func (h *ConfigHandler) DeleteSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的ID",
		})
		return
	}

	if err := h.secretRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除Secret失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}
