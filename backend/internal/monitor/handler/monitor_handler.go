package handler

import (
	"net/http"
	"strconv"
	"time"

	"my-cloud/internal/common/response"
	"my-cloud/internal/monitor/model"
	"my-cloud/internal/monitor/service"

	"github.com/gin-gonic/gin"
)

type MonitorHandler struct {
	monitorService *service.MonitorService
}

func NewMonitorHandler(monitorService *service.MonitorService) *MonitorHandler {
	return &MonitorHandler{monitorService: monitorService}
}

// Metric相关接口

// CreateMetric 创建指标
func (h *MonitorHandler) CreateMetric(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Description string `json:"description"`
		Unit        string `json:"unit"`
		Labels      string `json:"labels"`
		Enabled     int    `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	metric := &model.Metric{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Unit:        req.Unit,
		Labels:      req.Labels,
		Enabled:     req.Enabled,
	}

	if err := h.monitorService.CreateMetric(metric); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建指标失败: "+err.Error())
		return
	}

	response.Success(c, metric)
}

// GetMetric 获取指标详情
func (h *MonitorHandler) GetMetric(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的指标ID")
		return
	}

	metric, err := h.monitorService.GetMetric(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "指标不存在")
		return
	}

	response.Success(c, metric)
}

// ListMetrics 获取指标列表
func (h *MonitorHandler) ListMetrics(c *gin.Context) {
	metricType := c.Query("type")
	enabledStr := c.Query("enabled")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	var enabled *int
	if enabledStr != "" {
		val, _ := strconv.Atoi(enabledStr)
		enabled = &val
	}

	metrics, total, err := h.monitorService.ListMetrics(metricType, enabled, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取指标列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      metrics,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateMetric 更新指标
func (h *MonitorHandler) UpdateMetric(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的指标ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Unit        string `json:"unit"`
		Labels      string `json:"labels"`
		Enabled     *int   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	metric, err := h.monitorService.GetMetric(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "指标不存在")
		return
	}

	if req.Name != "" {
		metric.Name = req.Name
	}
	if req.Type != "" {
		metric.Type = req.Type
	}
	if req.Description != "" {
		metric.Description = req.Description
	}
	if req.Unit != "" {
		metric.Unit = req.Unit
	}
	if req.Labels != "" {
		metric.Labels = req.Labels
	}
	if req.Enabled != nil {
		metric.Enabled = *req.Enabled
	}

	if err := h.monitorService.UpdateMetric(metric); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新指标失败: "+err.Error())
		return
	}

	response.Success(c, metric)
}

// DeleteMetric 删除指标
func (h *MonitorHandler) DeleteMetric(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的指标ID")
		return
	}

	if err := h.monitorService.DeleteMetric(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除指标失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "指标删除成功"})
}

// AlertRule相关接口

// CreateAlertRule 创建告警规则
func (h *MonitorHandler) CreateAlertRule(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		MetricName  string  `json:"metric_name" binding:"required"`
		Condition   string  `json:"condition" binding:"required"`
		Threshold   float64 `json:"threshold" binding:"required"`
		Duration    int     `json:"duration"`
		Severity    string  `json:"severity" binding:"required"`
		Enabled     int     `json:"enabled"`
		NotifyUsers string  `json:"notify_users"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	rule := &model.AlertRule{
		Name:        req.Name,
		MetricName:  req.MetricName,
		Condition:   req.Condition,
		Threshold:   req.Threshold,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Enabled:     req.Enabled,
		NotifyUsers: req.NotifyUsers,
	}

	if err := h.monitorService.CreateAlertRule(rule); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建告警规则失败: "+err.Error())
		return
	}

	response.Success(c, rule)
}

// GetAlertRule 获取告警规则详情
func (h *MonitorHandler) GetAlertRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的规则ID")
		return
	}

	rule, err := h.monitorService.GetAlertRule(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "告警规则不存在")
		return
	}

	response.Success(c, rule)
}

// ListAlertRules 获取告警规则列表
func (h *MonitorHandler) ListAlertRules(c *gin.Context) {
	metricName := c.Query("metric_name")
	severity := c.Query("severity")
	enabledStr := c.Query("enabled")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	var enabled *int
	if enabledStr != "" {
		val, _ := strconv.Atoi(enabledStr)
		enabled = &val
	}

	rules, total, err := h.monitorService.ListAlertRules(metricName, severity, enabled, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取告警规则列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      rules,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateAlertRule 更新告警规则
func (h *MonitorHandler) UpdateAlertRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的规则ID")
		return
	}

	var req struct {
		Name        string   `json:"name"`
		MetricName  string   `json:"metric_name"`
		Condition   string   `json:"condition"`
		Threshold   *float64 `json:"threshold"`
		Duration    *int     `json:"duration"`
		Severity    string   `json:"severity"`
		Enabled     *int     `json:"enabled"`
		NotifyUsers string   `json:"notify_users"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	rule, err := h.monitorService.GetAlertRule(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "告警规则不存在")
		return
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.MetricName != "" {
		rule.MetricName = req.MetricName
	}
	if req.Condition != "" {
		rule.Condition = req.Condition
	}
	if req.Threshold != nil {
		rule.Threshold = *req.Threshold
	}
	if req.Duration != nil {
		rule.Duration = *req.Duration
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.NotifyUsers != "" {
		rule.NotifyUsers = req.NotifyUsers
	}

	if err := h.monitorService.UpdateAlertRule(rule); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新告警规则失败: "+err.Error())
		return
	}

	response.Success(c, rule)
}

// DeleteAlertRule 删除告警规则
func (h *MonitorHandler) DeleteAlertRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的规则ID")
		return
	}

	if err := h.monitorService.DeleteAlertRule(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除告警规则失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "告警规则删除成功"})
}

// Alert相关接口

// ListAlerts 获取告警列表
func (h *MonitorHandler) ListAlerts(c *gin.Context) {
	status := c.Query("status")
	severity := c.Query("severity")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	var startTime, endTime *time.Time
	if startTimeStr != "" {
		t, err := time.Parse("2006-01-02", startTimeStr)
		if err == nil {
			startTime = &t
		}
	}
	if endTimeStr != "" {
		t, err := time.Parse("2006-01-02", endTimeStr)
		if err == nil {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTime = &t
		}
	}

	alerts, total, err := h.monitorService.ListAlerts(status, severity, startTime, endTime, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取告警列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      alerts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAlert 获取告警详情
func (h *MonitorHandler) GetAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的告警ID")
		return
	}

	alert, err := h.monitorService.GetAlert(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "告警不存在")
		return
	}

	response.Success(c, alert)
}

// ResolveAlert 解决告警
func (h *MonitorHandler) ResolveAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的告警ID")
		return
	}

	if err := h.monitorService.ResolveAlert(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "解决告警失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "告警已解决"})
}

// GetAlertStatistics 获取告警统计
func (h *MonitorHandler) GetAlertStatistics(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	stats, err := h.monitorService.GetAlertStatistics(startTime, endTime)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取告警统计失败: "+err.Error())
		return
	}

	response.Success(c, stats)
}
