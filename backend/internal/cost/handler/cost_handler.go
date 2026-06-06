package handler

import (
	"log"
	"strconv"
	"time"

	"my-cloud/internal/common/response"
	"my-cloud/internal/cost/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CostHandler struct {
	costRepo *repository.CostRepository
	db       *gorm.DB
}

func NewCostHandler(costRepo *repository.CostRepository, db *gorm.DB) *CostHandler {
	return &CostHandler{
		costRepo: costRepo,
		db:       db,
	}
}

// ListCostRecords 成本记录列表
func (h *CostHandler) ListCostRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	clusterIDStr := c.Query("clusterId")
	projectIDStr := c.Query("projectId")
	appIDStr := c.Query("appId")
	namespace := c.Query("namespace")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	offset := (page - 1) * pageSize

	var clusterID uint
	if clusterIDStr != "" {
		id, _ := strconv.ParseUint(clusterIDStr, 10, 32)
		clusterID = uint(id)
	}

	var projectID *uint
	if projectIDStr != "" {
		id, _ := strconv.ParseUint(projectIDStr, 10, 32)
		pid := uint(id)
		projectID = &pid
	}

	var appID *uint
	if appIDStr != "" {
		id, _ := strconv.ParseUint(appIDStr, 10, 32)
		aid := uint(id)
		appID = &aid
	}

	records, total, err := h.costRepo.List(offset, pageSize, clusterID, projectID, appID, namespace, startDate, endDate)
	if err != nil {
		response.InternalError(c, "查询成本记录列表失败: "+err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, records)
}

// GetCostOverview 成本概览
func (h *CostHandler) GetCostOverview(c *gin.Context) {
	clusterIDStr := c.Query("clusterId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var clusterID uint
	if clusterIDStr != "" {
		id, _ := strconv.ParseUint(clusterIDStr, 10, 32)
		clusterID = uint(id)
	}

	overview, err := h.costRepo.GetOverview(clusterID, startDate, endDate)
	if err != nil {
		response.InternalError(c, "查询成本概览失败: "+err.Error())
		return
	}

	response.Success(c, overview)
}

// GetCostByProject 查询指定项目的成本记录
func (h *CostHandler) GetCostByProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的项目ID")
		return
	}

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	records, err := h.costRepo.GetCostByProject(uint(projectID), startDate, endDate)
	if err != nil {
		response.InternalError(c, "查询项目成本失败: "+err.Error())
		return
	}

	response.Success(c, records)
}

// GetCostByApp 查询指定应用的成本记录
func (h *CostHandler) GetCostByApp(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的应用ID")
		return
	}

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	records, err := h.costRepo.GetCostByApp(uint(appID), startDate, endDate)
	if err != nil {
		response.InternalError(c, "查询应用成本失败: "+err.Error())
		return
	}

	response.Success(c, records)
}

// SyncCostRequest 同步成本请求参数
type SyncCostRequest struct {
	ClusterID   uint   `json:"clusterId" binding:"required"`
	CostDate    string `json:"costDate" binding:"required"`
	KubecostURL string `json:"kubecostUrl"`
	PromURL     string `json:"promUrl"`
}

// SyncCostData 同步成本数据（优先 Kubecost，降级 Prometheus）
func (h *CostHandler) SyncCostData(c *gin.Context) {
	var req SyncCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, "请求参数错误: "+err.Error())
		return
	}

	if _, err := time.Parse("2006-01-02", req.CostDate); err != nil {
		response.InvalidParams(c, "日期格式无效，请使用 YYYY-MM-DD")
		return
	}

	var synced int
	var err error
	source := "prometheus"

	// 优先 Kubecost
	if req.KubecostURL != "" {
		synced, err = h.costRepo.SyncFromKubecost(req.KubecostURL, req.ClusterID, req.CostDate)
		if err == nil {
			source = "kubecost"
		} else {
			log.Printf("[CostHandler] Kubecost sync failed: %v, falling back to Prometheus", err)
		}
	}

	// 降级 Prometheus
	if source == "prometheus" {
		promURL := req.PromURL
		if promURL == "" {
			promURL = "http://prometheus:9090"
		}
		synced, err = h.costRepo.SyncFromPrometheus(promURL, req.ClusterID, req.CostDate)
		if err != nil {
			log.Printf("[CostHandler] Sync error: %v", err)
			response.Error(c, response.CodeExternalAPIError, "同步成本数据失败: "+err.Error())
			return
		}
	}

	log.Printf("[CostHandler] Synced %d cost records (%s) for cluster=%d date=%s", synced, source, req.ClusterID, req.CostDate)
	response.Success(c, gin.H{
		"syncedRecords": synced,
		"clusterId":     req.ClusterID,
		"costDate":      req.CostDate,
		"source":        source,
	})
}
