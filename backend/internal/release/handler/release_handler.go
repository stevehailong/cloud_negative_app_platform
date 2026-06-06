package handler

import (
	"fmt"
	"my-cloud/internal/common/response"
	"my-cloud/internal/release/model"
	"my-cloud/internal/release/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReleaseHandler struct {
	releaseService *service.ReleaseService
}

func NewReleaseHandler(releaseService *service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{
		releaseService: releaseService,
	}
}

// CreateRelease 创建发布工单
type CreateReleaseRequest struct {
	AppID             uint   `json:"appId" binding:"required"`
	EnvID             uint   `json:"envId" binding:"required"`
	PipelineRunID     *uint  `json:"pipelineRunId"`
	ReleaseVersion    string `json:"releaseVersion" binding:"required"`
	ReleaseStrategy   string `json:"releaseStrategy" binding:"required"`
	ImageURL          string `json:"imageUrl" binding:"required"`
	ClusterID         uint   `json:"clusterId"`
	Namespace         string `json:"namespace"`
	CanaryPercent     int    `json:"canaryPercent"`
	CanaryRoutingMode string `json:"canaryRoutingMode"`
	CanaryHeaderName  string `json:"canaryHeaderName"`
	CanaryHeaderValue string `json:"canaryHeaderValue"`
	CanaryCookieName  string `json:"canaryCookieName"`
	Description       string `json:"description"`
}

func (h *ReleaseHandler) CreateRelease(c *gin.Context) {
	var req CreateReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	// 获取操作用户ID（优先从JWT context，其次从 X-User-ID header，用于服务间调用）
	userID, _ := c.Get("userId")
	operatorUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		operatorUserID = uid
	} else if xUserID := c.GetHeader("X-User-ID"); xUserID != "" {
		if uid, err := strconv.ParseUint(xUserID, 10, 32); err == nil {
			operatorUserID = uint(uid)
		}
	}

	clusterID := req.ClusterID
	if clusterID == 0 {
		clusterID = 1
	}
	canaryPercent := req.CanaryPercent
	if canaryPercent == 0 {
		canaryPercent = 20
	}

	// 保持namespace为空，让release_service根据AppID自动分配app-specific namespace
	// 如果有特殊需要可以传 namespace 参数

	release := &model.Release{
		AppID:             req.AppID,
		EnvID:             req.EnvID,
		PipelineRunID:     req.PipelineRunID,
		ReleaseVersion:    req.ReleaseVersion,
		ReleaseStrategy:   req.ReleaseStrategy,
		ImageURL:          req.ImageURL,
		ClusterID:         clusterID,
		Namespace:         "", // 保持为空，让release_service根据AppID自动分配
		CanaryPercent:     canaryPercent,
		CanaryRoutingMode: req.CanaryRoutingMode,
		CanaryHeaderName:  req.CanaryHeaderName,
		CanaryHeaderValue: req.CanaryHeaderValue,
		CanaryCookieName:  req.CanaryCookieName,
		OperatorUserID:    operatorUserID,
		Description:       req.Description,
	}

	if err := h.releaseService.CreateRelease(release); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, release)
}

// UpdateRelease 更新发布工单
type UpdateReleaseRequest struct {
	ReleaseStrategy string `json:"releaseStrategy" binding:"required"`
	CanaryPercent   int    `json:"canaryPercent"`
	Description     string `json:"description"`
}

func (h *ReleaseHandler) UpdateRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	var req UpdateReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	// 验证策略
	if req.ReleaseStrategy != "rolling" && req.ReleaseStrategy != "canary" && req.ReleaseStrategy != "bluegreen" {
		response.InvalidParams(c, "无效的发布策略")
		return
	}

	// 金丝雀策略需要验证比例
	if req.ReleaseStrategy == "canary" {
		if req.CanaryPercent < 5 || req.CanaryPercent > 50 {
			response.InvalidParams(c, "金丝雀比例应在5%-50%之间")
			return
		}
	}

	if err := h.releaseService.UpdateRelease(uint(id), req.ReleaseStrategy, req.CanaryPercent, req.Description); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// GetRelease 获取发布工单详情
func (h *ReleaseHandler) GetRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	release, err := h.releaseService.GetRelease(uint(id))
	if err != nil {
		response.NotFound(c, "发布工单不存在")
		return
	}

	response.Success(c, release)
}

// ListReleases 获取发布工单列表
func (h *ReleaseHandler) ListReleases(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	appID, _ := strconv.ParseUint(c.Query("appId"), 10, 32)
	envID, _ := strconv.ParseUint(c.Query("envId"), 10, 32)
	releaseStatus := c.Query("releaseStatus")

	releases, total, err := h.releaseService.ListReleases(uint(appID), uint(envID), releaseStatus, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, releases)
}

// SubmitRelease 提交发布工单审批
type SubmitReleaseRequest struct {
	ApproverUserIDs []uint `json:"approverUserIds" binding:"required"`
}

func (h *ReleaseHandler) SubmitRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	var req SubmitReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if err := h.releaseService.SubmitRelease(uint(id), req.ApproverUserIDs); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "提交审批成功"})
}

// ApproveRelease 审批通过
type ApprovalRequest struct {
	Comment string `json:"comment"`
}

func (h *ReleaseHandler) ApproveRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	// 获取审批人ID
	userID, _ := c.Get("userId")
	approverUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		approverUserID = uid
	}

	if err := h.releaseService.ApproveRelease(uint(id), approverUserID, req.Comment); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "审批通过"})
}

// RejectRelease 审批拒绝
func (h *ReleaseHandler) RejectRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	// 获取审批人ID
	userID, _ := c.Get("userId")
	approverUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		approverUserID = uid
	}

	if err := h.releaseService.RejectRelease(uint(id), approverUserID, req.Comment); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "审批拒绝"})
}

// ExecuteRelease 执行发布
func (h *ReleaseHandler) ExecuteRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	// 获取操作用户ID
	userID, _ := c.Get("userId")
	operatorUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		operatorUserID = uid
	}

	if err := h.releaseService.ExecuteRelease(uint(id), operatorUserID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "发布已启动"})
}

// RollbackRelease 回滚发布
func (h *ReleaseHandler) RollbackRelease(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	// 获取操作用户ID
	userID, _ := c.Get("userId")
	operatorUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		operatorUserID = uid
	}

	if err := h.releaseService.RollbackRelease(uint(id), operatorUserID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "回滚已启动"})
}

// ConfirmCanary 确认金丝雀，全量发布
func (h *ReleaseHandler) ConfirmCanary(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	userID, _ := c.Get("userId")
	operatorUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		operatorUserID = uid
	}

	if err := h.releaseService.ConfirmCanary(uint(id), operatorUserID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "金丝雀确认，全量发布中"})
}

// RollbackCanary 回滚金丝雀
func (h *ReleaseHandler) RollbackCanary(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	userID, _ := c.Get("userId")
	operatorUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		operatorUserID = uid
	}

	if err := h.releaseService.RollbackCanary(uint(id), operatorUserID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "金丝雀已回滚"})
}

// AdjustCanaryWeight 动态调整金丝雀流量权重
func (h *ReleaseHandler) AdjustCanaryWeight(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	var req struct {
		CanaryPercent int `json:"canaryPercent" binding:"required,min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userId")
	operatorUserID := uint(0)
	if uid, ok := userID.(uint); ok {
		operatorUserID = uid
	}

	if err := h.releaseService.AdjustCanaryWeight(uint(id), req.CanaryPercent, operatorUserID); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": fmt.Sprintf("权重已调整为 %d%%", req.CanaryPercent)})
}

// ListReleaseApprovals 获取审批记录
func (h *ReleaseHandler) ListReleaseApprovals(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的工单ID")
		return
	}

	approvals, err := h.releaseService.ListReleaseApprovals(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, approvals)
}
