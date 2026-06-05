package handler

import (
	"strconv"
	"time"

	"my-cloud/internal/common/model"
	"my-cloud/internal/common/response"
	"my-cloud/internal/secret/repository"

	"github.com/gin-gonic/gin"
)

type SecretHandler struct {
	secretRepo *repository.SecretRepository
}

func NewSecretHandler(secretRepo *repository.SecretRepository) *SecretHandler {
	return &SecretHandler{
		secretRepo: secretRepo,
	}
}

// ListSecrets 密钥列表
func (h *SecretHandler) ListSecrets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	appIdStr := c.Query("appId")
	envIdStr := c.Query("envId")

	offset := (page - 1) * pageSize

	var appId, envId *uint
	if appIdStr != "" {
		id, err := strconv.ParseUint(appIdStr, 10, 32)
		if err == nil {
			aid := uint(id)
			appId = &aid
		}
	}
	if envIdStr != "" {
		id, err := strconv.ParseUint(envIdStr, 10, 32)
		if err == nil {
			eid := uint(id)
			envId = &eid
		}
	}

	secrets, total, err := h.secretRepo.List(offset, pageSize, appId, envId, keyword)
	if err != nil {
		response.Error(c, response.CodeDatabaseError, "查询密钥列表失败")
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, secrets)
}

// CreateSecret 创建密钥引用
func (h *SecretHandler) CreateSecret(c *gin.Context) {
	var req model.AppSecret
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.CodeInvalidParams, "请求参数错误")
		return
	}

	now := time.Now()
	req.CreateTime = now
	req.UpdateTime = now

	if err := h.secretRepo.Create(&req); err != nil {
		response.Error(c, response.CodeDatabaseError, "创建密钥失败")
		return
	}

	response.Success(c, req)
}

// GetSecret 获取密钥详情
func (h *SecretHandler) GetSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParams, "无效的密钥ID")
		return
	}

	secret, err := h.secretRepo.GetByID(uint(id))
	if err != nil {
		response.Error(c, response.CodeNotFound, "密钥不存在")
		return
	}

	response.Success(c, secret)
}

// UpdateSecret 更新密钥引用
func (h *SecretHandler) UpdateSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParams, "无效的密钥ID")
		return
	}

	existing, err := h.secretRepo.GetByID(uint(id))
	if err != nil {
		response.Error(c, response.CodeNotFound, "密钥不存在")
		return
	}

	var req model.AppSecret
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.CodeInvalidParams, "请求参数错误")
		return
	}

	req.ID = uint(id)
	req.CreateTime = existing.CreateTime
	req.UpdateTime = time.Now()

	if err := h.secretRepo.Update(&req); err != nil {
		response.Error(c, response.CodeDatabaseError, "更新密钥失败")
		return
	}

	response.Success(c, req)
}

// DeleteSecret 删除密钥引用
func (h *SecretHandler) DeleteSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParams, "无效的密钥ID")
		return
	}

	if err := h.secretRepo.Delete(uint(id)); err != nil {
		response.Error(c, response.CodeDatabaseError, "删除密钥失败")
		return
	}

	response.Success(c, nil)
}

// GetSecretsByApp 根据应用ID获取密钥列表
func (h *SecretHandler) GetSecretsByApp(c *gin.Context) {
	appId, err := strconv.ParseUint(c.Param("appId"), 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParams, "无效的应用ID")
		return
	}

	envIdStr := c.DefaultQuery("envId", "0")
	envId, err := strconv.ParseUint(envIdStr, 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParams, "无效的环境ID")
		return
	}

	secrets, err := h.secretRepo.GetByAppAndEnv(uint(appId), uint(envId))
	if err != nil {
		response.Error(c, response.CodeDatabaseError, "查询密钥失败")
		return
	}

	response.Success(c, secrets)
}
