package handler

import (
	"my-cloud/internal/common/response"
	"my-cloud/internal/deploy/model"
	"my-cloud/internal/deploy/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeployHandler struct {
	deployService *service.DeployService
}

func NewDeployHandler(deployService *service.DeployService) *DeployHandler {
	return &DeployHandler{
		deployService: deployService,
	}
}

// CreateDeployment 创建部署
type CreateDeploymentRequest struct {
	ReleaseID       uint   `json:"releaseId" binding:"required"`
	ClusterID       uint   `json:"clusterId" binding:"required"`
	Namespace       string `json:"namespace" binding:"required"`
	WorkloadName    string `json:"workloadName" binding:"required"`
	WorkloadType    string `json:"workloadType" binding:"required"`
	ImageVersion    string `json:"imageVersion" binding:"required"`
	DesiredReplicas int    `json:"desiredReplicas"`
}

func (h *DeployHandler) CreateDeployment(c *gin.Context) {
	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if req.DesiredReplicas == 0 {
		req.DesiredReplicas = 1
	}

	deployment := &model.Deployment{
		ReleaseID:       req.ReleaseID,
		ClusterID:       req.ClusterID,
		Namespace:       req.Namespace,
		WorkloadName:    req.WorkloadName,
		WorkloadType:    req.WorkloadType,
		ImageVersion:    req.ImageVersion,
		DesiredReplicas: req.DesiredReplicas,
	}

	if err := h.deployService.CreateDeployment(deployment); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, deployment)
}

// GetDeployment 获取部署详情
func (h *DeployHandler) GetDeployment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的部署ID")
		return
	}

	deployment, err := h.deployService.GetDeployment(uint(id))
	if err != nil {
		response.NotFound(c, "部署记录不存在")
		return
	}

	response.Success(c, deployment)
}

// ListDeployments 获取部署列表
func (h *DeployHandler) ListDeployments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	clusterID, _ := strconv.ParseUint(c.Query("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	startDate := c.Query("startDate")
	sortBy := c.DefaultQuery("sortBy", "createTime")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	deployments, total, err := h.deployService.ListDeployments(uint(clusterID), namespace, startDate, sortBy, sortOrder, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, deployments)
}

// RestartDeployment 重启部署
func (h *DeployHandler) RestartDeployment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的部署ID")
		return
	}

	if err := h.deployService.RestartDeployment(uint(id)); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "重启成功"})
}

// ScaleDeployment 扩缩容
type ScaleDeploymentRequest struct {
	Replicas int `json:"replicas" binding:"required"`
}

type ScaleDeploymentByNameRequest struct {
	Namespace    string `json:"namespace" binding:"required"`
	WorkloadName string `json:"workloadName" binding:"required"`
	Replicas     int    `json:"replicas" binding:"required"`
}

func (h *DeployHandler) ScaleDeployment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的部署ID")
		return
	}

	var req ScaleDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if err := h.deployService.ScaleDeployment(uint(id), req.Replicas); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "扩缩容成功"})
}

func (h *DeployHandler) ScaleDeploymentByName(c *gin.Context) {
	var req ScaleDeploymentByNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if err := h.deployService.ScaleDeploymentByName(req.Namespace, req.WorkloadName, req.Replicas); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "扩缩容成功"})
}

// GetDeploymentEvents 获取部署事件
func (h *DeployHandler) GetDeploymentEvents(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的部署ID")
		return
	}

	events, err := h.deployService.GetDeploymentEvents(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, events)
}

// GetDeploymentPods 获取部署的Pod列表
func (h *DeployHandler) GetDeploymentPods(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的部署ID")
		return
	}

	pods, err := h.deployService.GetDeploymentPods(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, pods)
}

// DeletePod 删除Pod
func (h *DeployHandler) DeletePod(c *gin.Context) {
	namespace := c.Query("namespace")
	podName := c.Param("podName")
	if namespace == "" || podName == "" {
		response.InvalidParams(c, "namespace和podName不能为空")
		return
	}
	if err := h.deployService.DeletePod(namespace, podName); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "Pod已删除"})
}

// DeleteDeployment 删除部署
func (h *DeployHandler) DeleteDeployment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的部署ID")
		return
	}
	if err := h.deployService.DeleteDeployment(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "部署已删除"})
}

// DeleteK8sDeployment 直接删除K8s Deployment（内部API）
func (h *DeployHandler) DeleteK8sDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace == "" || name == "" {
		response.InvalidParams(c, "namespace和name不能为空")
		return
	}
	if err := h.deployService.DeleteK8sDeployment(namespace, name); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "K8s Deployment已删除"})
}

// DeleteDeploymentsByWorkload 删除指定workload的所有数据库记录（内部API）
func (h *DeployHandler) DeleteDeploymentsByWorkload(c *gin.Context) {
	namespace := c.Query("namespace")
	workloadName := c.Query("workloadName")
	if namespace == "" || workloadName == "" {
		response.InvalidParams(c, "namespace和workloadName不能为空")
		return
	}
	if err := h.deployService.DeleteDeploymentsByWorkload(namespace, workloadName); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "部署记录已删除"})
}

// GetK8sDeploymentReplicas 获取K8s部署的副本数（内部API）
func (h *DeployHandler) GetK8sDeploymentReplicas(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	
	replicas, err := h.deployService.GetK8sDeploymentReplicas(namespace, name)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    40401,
			"message": "部署不存在",
		})
		return
	}
	
	response.Success(c, gin.H{"replicas": replicas})
}

// RollbackDeployment 回滚部署
func (h *DeployHandler) RollbackDeployment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的部署ID")
		return
	}
	if err := h.deployService.RollbackDeployment(uint(id)); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "回滚成功"})
}
