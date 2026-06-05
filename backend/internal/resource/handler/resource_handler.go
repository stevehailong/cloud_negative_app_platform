package handler

import (
	"log"
	"strconv"
	"time"

	"my-cloud/internal/common/model"
	"my-cloud/internal/common/response"
	"my-cloud/internal/resource/repository"

	"github.com/gin-gonic/gin"
)

type ResourceHandler struct {
	resourceRepo *repository.ResourceRepository
}

func NewResourceHandler(resourceRepo *repository.ResourceRepository) *ResourceHandler {
	return &ResourceHandler{
		resourceRepo: resourceRepo,
	}
}

// ListResourceQuotas 资源配额列表
func (h *ResourceHandler) ListResourceQuotas(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	scopeType := c.Query("scopeType")
	scopeIDStr := c.Query("scopeId")

	offset := (page - 1) * pageSize

	var scopeID *uint
	if scopeIDStr != "" {
		id, _ := strconv.ParseUint(scopeIDStr, 10, 32)
		sid := uint(id)
		scopeID = &sid
	}

	quotas, total, err := h.resourceRepo.List(offset, pageSize, scopeType, scopeID)
	if err != nil {
		response.DatabaseError(c, "查询资源配额列表失败")
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, quotas)
}

// CreateResourceQuota 创建资源配额
func (h *ResourceHandler) CreateResourceQuota(c *gin.Context) {
	var req model.ResourceQuota
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, "请求参数错误")
		return
	}

	now := time.Now()
	req.CreateTime = now
	req.UpdateTime = now
	req.IsDeleted = 0
	if req.Status == 0 {
		req.Status = 1
	}

	if err := h.resourceRepo.Create(&req); err != nil {
		response.DatabaseError(c, "创建资源配额失败")
		return
	}

	response.Success(c, req)
}

// GetResourceQuota 获取资源配额详情
func (h *ResourceHandler) GetResourceQuota(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的资源配额ID")
		return
	}

	quota, err := h.resourceRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "资源配额不存在")
		return
	}

	response.Success(c, quota)
}

// UpdateResourceQuota 更新资源配额
func (h *ResourceHandler) UpdateResourceQuota(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的资源配额ID")
		return
	}

	quota, err := h.resourceRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "资源配额不存在")
		return
	}

	var req model.ResourceQuota
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, "请求参数错误")
		return
	}

	req.ID = uint(id)
	req.CreateTime = quota.CreateTime
	req.UpdateTime = time.Now()

	if err := h.resourceRepo.Update(&req); err != nil {
		response.DatabaseError(c, "更新资源配额失败")
		return
	}

	response.Success(c, req)
}

// DeleteResourceQuota 删除资源配额（软删除）
func (h *ResourceHandler) DeleteResourceQuota(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的资源配额ID")
		return
	}

	if err := h.resourceRepo.Delete(uint(id)); err != nil {
		response.DatabaseError(c, "删除资源配额失败")
		return
	}

	response.Success(c, nil)
}

// ========== 内部 API ==========

// SyncFromK8s 从 K8s/Prometheus 同步资源使用情况并计算配额建议
func (h *ResourceHandler) SyncFromK8s(c *gin.Context) {
	var req struct {
		ScopeType string `json:"scopeType" binding:"required"` // tenant/project/env/namespace/app
		ScopeID   uint   `json:"scopeId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, "请求参数错误: scopeType和scopeId为必填项")
		return
	}

	// 检查是否已存在配额记录
	quotas, total, err := h.resourceRepo.List(0, 1, req.ScopeType, &req.ScopeID)
	if err != nil {
		response.DatabaseError(c, "查询已有配额失败")
		return
	}

	// 计算配额建议值
	recommendation := h.computeQuotaRecommendation(req.ScopeType, req.ScopeID)

	now := time.Now()

	if total > 0 {
		// 更新已有记录
		existing := &quotas[0]
		existing.CPULimit = recommendation.CPULimit
		existing.MemoryLimit = recommendation.MemoryLimit
		existing.StorageLimit = recommendation.StorageLimit
		existing.PodLimit = recommendation.PodLimit
		existing.ServiceLimit = recommendation.ServiceLimit
		existing.LBLimit = recommendation.LBLimit
		existing.GPULimit = recommendation.GPULimit
		existing.UpdateTime = now
		if err := h.resourceRepo.Update(existing); err != nil {
			response.DatabaseError(c, "更新配额建议失败")
			return
		}
		response.Success(c, existing)
	} else {
		// 创建新记录
		newQuota := &model.ResourceQuota{
			ScopeType:    req.ScopeType,
			ScopeID:      req.ScopeID,
			CPULimit:     recommendation.CPULimit,
			MemoryLimit:  recommendation.MemoryLimit,
			StorageLimit: recommendation.StorageLimit,
			PodLimit:     recommendation.PodLimit,
			ServiceLimit: recommendation.ServiceLimit,
			LBLimit:      recommendation.LBLimit,
			GPULimit:     recommendation.GPULimit,
			Status:       1,
			IsDeleted:    0,
			CreateTime:   now,
			UpdateTime:   now,
		}
		if err := h.resourceRepo.Create(newQuota); err != nil {
			response.DatabaseError(c, "创建配额建议失败")
			return
		}
		response.Success(c, newQuota)
	}

	log.Printf("[Resource] Synced quota recommendation for %s:%d", req.ScopeType, req.ScopeID)
}

// QuotaRecommendation 配额建议值
type QuotaRecommendation struct {
	CPULimit     string `json:"cpuLimit"`
	MemoryLimit  string `json:"memoryLimit"`
	StorageLimit string `json:"storageLimit"`
	PodLimit     int    `json:"podLimit"`
	ServiceLimit int    `json:"serviceLimit"`
	LBLimit      int    `json:"lbLimit"`
	GPULimit     int    `json:"gpuLimit"`
}

// computeQuotaRecommendation 根据 scope 级别计算配额推荐值
// 后续可集成 Prometheus 查询实际使用量来动态计算
func (h *ResourceHandler) computeQuotaRecommendation(scopeType string, scopeID uint) QuotaRecommendation {
	// 基于 scope 类型的默认推荐配额
	defaults := map[string]QuotaRecommendation{
		"tenant": {
			CPULimit: "100", MemoryLimit: "256Gi", StorageLimit: "1000Gi",
			PodLimit: 200, ServiceLimit: 50, LBLimit: 20, GPULimit: 8,
		},
		"project": {
			CPULimit: "50", MemoryLimit: "128Gi", StorageLimit: "500Gi",
			PodLimit: 100, ServiceLimit: 30, LBLimit: 10, GPULimit: 4,
		},
		"env": {
			CPULimit: "20", MemoryLimit: "64Gi", StorageLimit: "200Gi",
			PodLimit: 50, ServiceLimit: 15, LBLimit: 5, GPULimit: 2,
		},
		"namespace": {
			CPULimit: "10", MemoryLimit: "32Gi", StorageLimit: "100Gi",
			PodLimit: 30, ServiceLimit: 10, LBLimit: 3, GPULimit: 1,
		},
		"app": {
			CPULimit: "4", MemoryLimit: "8Gi", StorageLimit: "50Gi",
			PodLimit: 10, ServiceLimit: 5, LBLimit: 2, GPULimit: 1,
		},
	}

	if rec, ok := defaults[scopeType]; ok {
		return rec
	}

	// 默认返回 app 级别的限制
	return defaults["app"]
}
