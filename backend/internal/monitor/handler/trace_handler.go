package handler

import (
	"strconv"
	"time"

	"my-cloud/internal/common/response"
	"my-cloud/internal/monitor/model"
	"my-cloud/internal/monitor/repository"
	"my-cloud/internal/monitor/service"

	"github.com/gin-gonic/gin"
)

// TraceHandler 链路追踪HTTP处理器
type TraceHandler struct {
	traceService *service.TraceService
}

func NewTraceHandler(traceService *service.TraceService) *TraceHandler {
	return &TraceHandler{traceService: traceService}
}

// CollectSpan 接收并保存Span（内部API，由gateway调用）
// POST /internal/v1/traces/spans
func (h *TraceHandler) CollectSpan(c *gin.Context) {
	var span model.TraceSpan
	if err := c.ShouldBindJSON(&span); err != nil {
		response.InvalidParams(c, "无效的Span数据: "+err.Error())
		return
	}
	if span.TraceID == "" || span.SpanID == "" {
		response.InvalidParams(c, "trace_id 和 span_id 不能为空")
		return
	}
	if span.StartTime.IsZero() {
		span.StartTime = time.Now()
	}
	if err := h.traceService.SaveSpan(&span); err != nil {
		response.Error(c, response.CodeInternalError, "保存Span失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"message": "span collected"})
}

// ListTraces 查询Trace列表（支持多种筛选条件）
// GET /api/v1/traces?serviceName=xxx&operationName=xxx&method=GET&minDuration=100&maxDuration=5000&hasError=1&startTime=xxx&endTime=xxx&page=1&pageSize=20
func (h *TraceHandler) ListTraces(c *gin.Context) {
	serviceName := c.Query("serviceName")
	operationName := c.Query("operationName")
	method := c.Query("method")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	minDuration, _ := strconv.Atoi(c.Query("minDuration"))
	maxDuration, _ := strconv.Atoi(c.Query("maxDuration"))

	var hasError *int
	if errStr := c.Query("hasError"); errStr != "" {
		if v, err := strconv.Atoi(errStr); err == nil {
			hasError = &v
		}
	}

	var startTime, endTime time.Time
	if st := c.Query("startTime"); st != "" {
		startTime, _ = time.Parse(time.RFC3339, st)
	}
	if et := c.Query("endTime"); et != "" {
		endTime, _ = time.Parse(time.RFC3339, et)
	}

	params := repository.TraceQueryParams{
		ServiceName:   serviceName,
		OperationName: operationName,
		Method:        method,
		MinDuration:   minDuration,
		MaxDuration:   maxDuration,
		HasError:      hasError,
		StartTime:     startTime,
		EndTime:       endTime,
		Limit:         pageSize,
		Offset:        (page - 1) * pageSize,
	}

	spans, total, err := h.traceService.ListTracesWithFilters(params)
	if err != nil {
		response.Error(c, response.CodeInternalError, "查询Trace失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  spans,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetTrace 查询指定TraceID的所有Span
// GET /api/v1/traces/:traceId
func (h *TraceHandler) GetTrace(c *gin.Context) {
	traceID := c.Param("traceId")
	if traceID == "" {
		response.InvalidParams(c, "traceId不能为空")
		return
	}
	spans, err := h.traceService.GetTraceByID(traceID)
	if err != nil {
		response.Error(c, response.CodeInternalError, "查询Trace详情失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"traceId": traceID,
		"spans":   spans,
		"total":   len(spans),
	})
}

// GetTracesByApp 按应用/服务查询Trace
// GET /api/v1/traces/apps/:appId?serviceName=xxx
func (h *TraceHandler) GetTracesByApp(c *gin.Context) {
	appID := c.Param("appId")
	// 优先使用 serviceName 查询参数（应用名），否则回退到 URL 中的 appId
	serviceName := c.Query("serviceName")
	if serviceName == "" {
		serviceName = appID
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	spans, total, err := h.traceService.GetTraceList("", serviceName, page, pageSize)
	if err != nil {
		response.Error(c, response.CodeInternalError, "查询Trace失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  spans,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetServices 获取所有产生过 Trace 的服务名列表
// GET /api/v1/traces/services/list
func (h *TraceHandler) GetServices(c *gin.Context) {
	services, err := h.traceService.GetServiceList()
	if err != nil {
		response.Error(c, response.CodeInternalError, "查询服务列表失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"services": services})
}

// GetTraceStats 获取链路统计信息
// GET /api/v1/traces/stats?startTime=xxx&endTime=xxx
func (h *TraceHandler) GetTraceStats(c *gin.Context) {
	startTime, _ := time.Parse(time.RFC3339, c.DefaultQuery("startTime", time.Now().AddDate(0, 0, -7).Format(time.RFC3339)))
	endTime, _ := time.Parse(time.RFC3339, c.DefaultQuery("endTime", time.Now().Format(time.RFC3339)))

	stats, err := h.traceService.GetTraceStats(startTime, endTime)
	if err != nil {
		response.Error(c, response.CodeInternalError, "查询统计信息失败: "+err.Error())
		return
	}
	response.Success(c, stats)
}
