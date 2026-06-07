package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"my-cloud/internal/deploy/service"

	"github.com/gin-gonic/gin"
)

type AppDeploymentHandler struct {
	appDeployService *service.AppDeploymentService
}

func NewAppDeploymentHandler(appDeployService *service.AppDeploymentService) *AppDeploymentHandler {
	return &AppDeploymentHandler{
		appDeployService: appDeployService,
	}
}

// ListAppDeployments 查询应用部署列表
// GET /api/v1/app-deployments?app_id=8&env_id=1&page=1&page_size=20
func (h *AppDeploymentHandler) ListAppDeployments(c *gin.Context) {
	var appID, envID *int64
	
	if appIDStr := c.Query("app_id"); appIDStr != "" {
		if id, err := strconv.ParseInt(appIDStr, 10, 64); err == nil {
			appID = &id
		}
	}
	
	if envIDStr := c.Query("env_id"); envIDStr != "" {
		if id, err := strconv.ParseInt(envIDStr, 10, 64); err == nil {
			envID = &id
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	deployments, total, err := h.appDeployService.ListAppDeployments(appID, envID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":      deployments,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetAppDeploymentDetail 获取应用部署详情
// GET /api/v1/app-deployments/:id
func (h *AppDeploymentHandler) GetAppDeploymentDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	deployment, err := h.appDeployService.GetAppDeploymentDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "部署记录不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    deployment,
	})
}

// GetDeploymentHistory 获取部署历史
// GET /api/v1/app-deployments/:id/history?page=1&page_size=20
func (h *AppDeploymentHandler) GetDeploymentHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	histories, total, err := h.appDeployService.GetDeploymentHistory(id, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":      histories,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// AppRestartDeploymentRequest 重启请求
type AppRestartDeploymentRequest struct {
	UserID int64 `json:"user_id"`
}

// RestartDeployment 重启部署
// POST /api/v1/app-deployments/:id/restart
func (h *AppDeploymentHandler) RestartDeployment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	var req AppRestartDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.UserID = 1 // 默认用户ID
	}

	if err := h.appDeployService.RestartDeployment(id, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "重启失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "重启任务已提交，正在执行中",
	})
}

// AppScaleDeploymentRequest 扩缩容请求
type AppScaleDeploymentRequest struct {
	Replicas int   `json:"replicas" binding:"required,min=0"`
	UserID   int64 `json:"user_id"`
}

// ScaleDeployment 扩缩容
// POST /api/v1/app-deployments/:id/scale
func (h *AppDeploymentHandler) ScaleDeployment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	var req AppScaleDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.UserID == 0 {
		req.UserID = 1 // 默认用户ID
	}

	if err := h.appDeployService.ScaleDeployment(id, req.Replicas, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "扩缩容失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "扩缩容任务已提交，正在执行中",
	})
}

// AppRollbackDeploymentRequest 回滚请求
type AppRollbackDeploymentRequest struct {
	HistoryID int64 `json:"history_id" binding:"required"`
	UserID    int64 `json:"user_id"`
}

// RollbackDeployment 回滚到历史版本
// POST /api/v1/app-deployments/:id/rollback
func (h *AppDeploymentHandler) RollbackDeployment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	var req AppRollbackDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.UserID == 0 {
		req.UserID = 1 // 默认用户ID
	}

	if err := h.appDeployService.RollbackDeployment(id, req.HistoryID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "回滚失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "回滚任务已提交，正在执行中",
	})
}

// AppDeployNewVersionRequest 部署新版本请求
type AppDeployNewVersionRequest struct {
	Version  string `json:"version" binding:"required"`
	ImageURL string `json:"image_url" binding:"required"`
	UserID   int64  `json:"user_id"`
	Strategy string `json:"strategy"` // 部署策略: rolling/canary/bluegreen
}

// DeployNewVersion 部署新版本
// POST /api/v1/app-deployments/:id/deploy
func (h *AppDeploymentHandler) DeployNewVersion(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	var req AppDeployNewVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.UserID == 0 {
		req.UserID = 1 // 默认用户ID
	}

	historyID, err := h.appDeployService.DeployNewVersion(id, req.Version, req.ImageURL, req.UserID, req.Strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "部署失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "部署任务已提交，正在执行中",
		"data": gin.H{
			"history_id": historyID,
		},
	})
}

// GetDeploymentPods 获取部署的Pod列表
// GET /api/v1/app-deployments/:id/pods
func (h *AppDeploymentHandler) GetDeploymentPods(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	pods, err := h.appDeployService.GetDeploymentPods(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取Pod列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    pods,
	})
}

// GetDeploymentEvents 获取部署的事件列表
// GET /api/v1/app-deployments/:id/events
func (h *AppDeploymentHandler) GetDeploymentEvents(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	events, err := h.appDeployService.GetDeploymentEvents(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取事件列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    events,
	})
}

// GetAppDeploymentByWorkload 根据workload_name查询app_deployment (内部API)
// GET /internal/v1/app-deployments/by-workload?namespace=app-8&workload_name=app-8-canary
func (h *AppDeploymentHandler) GetAppDeploymentByWorkload(c *gin.Context) {
	namespace := c.Query("namespace")
	workloadName := c.Query("workload_name")

	if namespace == "" || workloadName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "namespace和workload_name不能为空",
		})
		return
	}

	deployment, err := h.appDeployService.GetByWorkloadName(namespace, workloadName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "未找到对应的部署记录",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": deployment,
	})
}

// CreateAppDeploymentInternal 创建app_deployment记录 (内部API)
// POST /internal/v1/app-deployments
// namespace和cluster_id会从环境自动获取，无需传入
func (h *AppDeploymentHandler) CreateAppDeploymentInternal(c *gin.Context) {
	var req struct {
		AppID           int64  `json:"app_id" binding:"required"`
		EnvID           int64  `json:"env_id" binding:"required"`
		WorkloadName    string `json:"workload_name" binding:"required"`
		WorkloadType    string `json:"workload_type"`
		DesiredReplicas int    `json:"desired_replicas"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	deployment, err := h.appDeployService.CreateAppDeployment(
		req.AppID, req.EnvID,
		req.WorkloadName, req.WorkloadType,
		req.DesiredReplicas,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    deployment,
	})
}

// GetDeploymentHistoryByID 查询单条deployment_history记录 (内部API)
// GET /internal/v1/deployment-history/:id
func (h *AppDeploymentHandler) GetDeploymentHistoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	history, err := h.appDeployService.GetDeploymentHistoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "记录不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": history,
	})
}

// ListAppDeploymentsByAppEnv 查询应用在指定环境的所有部署(包括stable和canary)
// GET /api/v1/app-deployments/by-app-env?app_id=8&env_id=1
func (h *AppDeploymentHandler) ListAppDeploymentsByAppEnv(c *gin.Context) {
	appIDStr := c.Query("app_id")
	envIDStr := c.Query("env_id")

	if appIDStr == "" || envIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id和env_id不能为空",
		})
		return
	}

	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的app_id",
		})
		return
	}

	envID, err := strconv.ParseInt(envIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的env_id",
		})
		return
	}

	deployments, err := h.appDeployService.ListByAppAndEnv(appID, envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    deployments,
	})
}

// CleanupDuplicateDeployments 清理不合理的重复部署记录
// DELETE /api/v1/app-deployments/cleanup?app_id=8&env_id=1
func (h *AppDeploymentHandler) CleanupDuplicateDeployments(c *gin.Context) {
	appIDStr := c.Query("app_id")
	envIDStr := c.Query("env_id")

	if appIDStr == "" || envIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id和env_id不能为空",
		})
		return
	}

	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的app_id",
		})
		return
	}

	envID, err := strconv.ParseInt(envIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的env_id",
		})
		return
	}

	deletedCount, err := h.appDeployService.CleanupDuplicateDeployments(appID, envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清理失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "清理完成",
		"data": gin.H{
			"deleted_count": deletedCount,
		},
	})
}

// PromoteCanaryToStableRequest 提升金丝雀版本请求
type PromoteCanaryToStableRequest struct {
	UserID int64 `json:"user_id"`
}

// PromoteCanaryToStable 将金丝雀版本提升为稳定版本并删除canary记录
// POST /api/v1/app-deployments/promote-canary?app_id=8&env_id=1
func (h *AppDeploymentHandler) PromoteCanaryToStable(c *gin.Context) {
	appIDStr := c.Query("app_id")
	envIDStr := c.Query("env_id")

	if appIDStr == "" || envIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id和env_id不能为空",
		})
		return
	}

	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的app_id",
		})
		return
	}

	envID, err := strconv.ParseInt(envIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的env_id",
		})
		return
	}

	var req PromoteCanaryToStableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.UserID = 1 // 默认用户ID
	}

	if err := h.appDeployService.PromoteCanaryToStable(appID, envID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提升失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "金丝雀版本已提升为稳定版本，canary记录已删除",
	})
}

// DeleteAppDeployment 删除应用部署记录
// DELETE /internal/v1/app-deployments/:id
func (h *AppDeploymentHandler) DeleteAppDeployment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	if err := h.appDeployService.DeleteAppDeployment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetEnvironmentInternal 内部API：获取环境信息（供release-service使用）
// GET /internal/v1/environments/:id
func (h *AppDeploymentHandler) GetEnvironmentInternal(c *gin.Context) {
	envID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的环境ID",
		})
		return
	}

	env, err := h.appDeployService.GetEnvironmentByID(uint(envID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "环境不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": env,
	})
}

// GetCanaryWeight 查询金丝雀当前 Ingress 权重（内部 API）
// GET /internal/v1/app-deployments/canary-weight?namespace=xxx&workload_name=xxx
func (h *AppDeploymentHandler) GetCanaryWeight(c *gin.Context) {
	namespace := c.Query("namespace")
	workloadName := c.Query("workload_name")
	if namespace == "" || workloadName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "namespace和workload_name不能为空"})
		return
	}
	deployment, err := h.appDeployService.GetByWorkloadName(namespace, workloadName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "未找到部署记录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"canary_weight": deployment.CanaryWeight,
		"workload_name": deployment.WorkloadName,
	}})
}

// AdjustCanaryWeight 调整金丝雀流量权重
// POST /api/v1/app-deployments/:id/canary/adjust-weight
func (h *AppDeploymentHandler) AdjustCanaryWeight(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var req struct {
		Weight int `json:"weight" binding:"required,min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := h.appDeployService.AdjustCanaryWeight(id, req.Weight); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": fmt.Sprintf("权重已调整为 %d%%", req.Weight)})
}


// GetAppEnvBinding 查询 app_env_binding 配置（内部 API）
func (h *AppDeploymentHandler) GetAppEnvBinding(c *gin.Context) {
	appIDStr := c.Query("app_id")
	envIDStr := c.Query("env_id")
	appID, _ := strconv.ParseInt(appIDStr, 10, 64)
	envID, _ := strconv.ParseInt(envIDStr, 10, 64)
	if appID == 0 || envID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "app_id and env_id required"})
		return
	}
	binding, err := h.appDeployService.GetAppEnvBinding(appID, envID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "binding not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": binding})
}

// RestartByAppEnv 通过 app_id+env_id 重启所有关联部署
func (h *AppDeploymentHandler) RestartByAppEnv(c *gin.Context) {
	var req struct {
		AppID int64 `json:"app_id" binding:"required"`
		EnvID int64 `json:"env_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "app_id and env_id required"})
		return
	}

	restarted, err := h.appDeployService.RestartByAppEnv(req.AppID, req.EnvID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": fmt.Sprintf("restarted %d deployment(s)", restarted)})
}
