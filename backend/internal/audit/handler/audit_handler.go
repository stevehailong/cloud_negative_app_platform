package handler

import (
	"my-cloud/internal/audit/service"
	"my-cloud/internal/common/model"
	"my-cloud/internal/common/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// ListAuditLogs 获取审计日志列表
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 构建过滤条件
	filters := make(map[string]interface{})

	if userIDStr := c.Query("userId"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}

	if username := c.Query("username"); username != "" {
		filters["username"] = username
	}

	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	if resourceType := c.Query("resourceType"); resourceType != "" {
		filters["resource_type"] = resourceType
	}

	if resourceIDStr := c.Query("resourceId"); resourceIDStr != "" {
		if resourceID, err := strconv.ParseUint(resourceIDStr, 10, 32); err == nil {
			filters["resource_id"] = uint(resourceID)
		}
	}

	if method := c.Query("method"); method != "" {
		filters["method"] = method
	}

	if path := c.Query("path"); path != "" {
		filters["path"] = path
	}

	if ipAddress := c.Query("ipAddress"); ipAddress != "" {
		filters["ip_address"] = ipAddress
	}

	if startTime := c.Query("startTime"); startTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			filters["start_time"] = t
		} else if t, err := time.Parse("2006-01-02", startTime); err == nil {
			filters["start_time"] = t
		}
	}

	if endTime := c.Query("endTime"); endTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
			filters["end_time"] = t
		} else if t, err := time.Parse("2006-01-02", endTime); err == nil {
			// 设置为当天23:59:59
			filters["end_time"] = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	if responseCode := c.Query("responseCode"); responseCode != "" {
		if code, err := strconv.Atoi(responseCode); err == nil {
			filters["response_code"] = code
		}
	}

	logs, total, err := h.auditService.ListAuditLogs(filters, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, logs)
}

// GetAuditLog 获取审计日志详情
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的审计日志ID")
		return
	}

	log, err := h.auditService.GetAuditLog(uint(id))
	if err != nil {
		response.NotFound(c, "审计日志不存在")
		return
	}

	response.Success(c, log)
}

// GetAuditLogsByResource 根据资源获取审计日志
func (h *AuditHandler) GetAuditLogsByResource(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID, err := strconv.ParseUint(c.Param("resourceId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的资源ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, total, err := h.auditService.GetAuditLogsByResourceID(resourceType, uint(resourceID), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, logs)
}

// GetAuditLogsByUser 根据用户获取审计日志
func (h *AuditHandler) GetAuditLogsByUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的用户ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, total, err := h.auditService.GetAuditLogsByUserID(uint(userID), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, logs)
}

// GetStatistics 获取审计日志统计信息
func (h *AuditHandler) GetStatistics(c *gin.Context) {
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")

	stats, err := h.auditService.GetStatistics(startTime, endTime)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// ExportAuditLogs 导出审计日志
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	// 构建过滤条件(与ListAuditLogs相同)
	filters := make(map[string]interface{})

	if userIDStr := c.Query("userId"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters["user_id"] = uint(userID)
		}
	}

	if username := c.Query("username"); username != "" {
		filters["username"] = username
	}

	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	if resourceType := c.Query("resourceType"); resourceType != "" {
		filters["resource_type"] = resourceType
	}

	if startTime := c.Query("startTime"); startTime != "" {
		if t, err := time.Parse("2006-01-02", startTime); err == nil {
			filters["start_time"] = t
		}
	}

	if endTime := c.Query("endTime"); endTime != "" {
		if t, err := time.Parse("2006-01-02", endTime); err == nil {
			filters["end_time"] = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	csv, err := h.auditService.ExportAuditLogs(filters)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 设置CSV响应头
	filename := "audit_logs_" + time.Now().Format("20060102150405") + ".csv"
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.String(200, csv)
}

// CreateAuditLog 创建审计日志（内部API，供其他服务调用）
func (h *AuditHandler) CreateAuditLog(c *gin.Context) {
	var req model.AuditLog
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}
	if err := h.auditService.CreateAuditLog(&req); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": req.ID})
}

// CleanOldLogs 清理过期日志
type CleanOldLogsRequest struct {
	RetentionDays int `json:"retentionDays" binding:"required,min=1"`
}

func (h *AuditHandler) CleanOldLogs(c *gin.Context) {
	var req CleanOldLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	count, err := h.auditService.CleanOldLogs(req.RetentionDays)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message":      "清理完成",
		"deleted_count": count,
	})
}
