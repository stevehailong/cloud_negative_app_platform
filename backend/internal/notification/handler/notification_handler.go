package handler

import (
	"my-cloud/internal/common/response"
	"my-cloud/internal/notification/model"
	"my-cloud/internal/notification/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// SendNotification 发送通知
type SendNotificationRequest struct {
	Title        string `json:"title" binding:"required"`
	Content      string `json:"content" binding:"required"`
	NotifyType   string `json:"notifyType" binding:"required"`
	Channel      string `json:"channel" binding:"required"`
	ReceiverType string `json:"receiverType" binding:"required"`
	ReceiverIDs  string `json:"receiverIds" binding:"required"`
	TemplateID   *uint  `json:"templateId"`
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	notification := &model.Notification{
		Title:        req.Title,
		Content:      req.Content,
		NotifyType:   req.NotifyType,
		Channel:      req.Channel,
		ReceiverType: req.ReceiverType,
		ReceiverIDs:  req.ReceiverIDs,
		TemplateID:   req.TemplateID,
	}

	if err := h.notificationService.SendNotification(notification); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, notification)
}

// SendByTemplate 根据模板发送通知
type SendByTemplateRequest struct {
	TemplateCode string                 `json:"templateCode" binding:"required"`
	Params       map[string]interface{} `json:"params" binding:"required"`
	ReceiverType string                 `json:"receiverType" binding:"required"`
	ReceiverIDs  []uint                 `json:"receiverIds" binding:"required"`
}

func (h *NotificationHandler) SendByTemplate(c *gin.Context) {
	var req SendByTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	if err := h.notificationService.SendNotificationByTemplate(
		req.TemplateCode,
		req.Params,
		req.ReceiverType,
		req.ReceiverIDs,
	); err != nil {
		response.Error(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "通知发送成功"})
}

// GetNotification 获取通知详情
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的通知ID")
		return
	}

	notification, err := h.notificationService.GetNotification(uint(id))
	if err != nil {
		response.NotFound(c, "通知不存在")
		return
	}

	response.Success(c, notification)
}

// ListNotifications 获取通知列表
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	notifyType := c.Query("notifyType")
	status := c.Query("status")

	notifications, total, err := h.notificationService.ListNotifications(notifyType, status, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, notifications)
}

// CreateTemplate 创建模板
type CreateTemplateRequest struct {
	TemplateCode string `json:"templateCode" binding:"required"`
	TemplateName string `json:"templateName" binding:"required"`
	NotifyType   string `json:"notifyType" binding:"required"`
	Channel      string `json:"channel" binding:"required"`
	Title        string `json:"title"`
	Content      string `json:"content" binding:"required"`
	Variables    string `json:"variables"`
}

func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	template := &model.NotificationTemplate{
		TemplateCode: req.TemplateCode,
		TemplateName: req.TemplateName,
		NotifyType:   req.NotifyType,
		Channel:      req.Channel,
		Title:        req.Title,
		Content:      req.Content,
		Variables:    req.Variables,
		Enabled:      1,
	}

	if err := h.notificationService.CreateTemplate(template); err != nil {
		response.Error(c, response.CodeConflict, err.Error())
		return
	}

	response.Success(c, template)
}

// GetTemplate 获取模板详情
func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的模板ID")
		return
	}

	template, err := h.notificationService.GetTemplate(uint(id))
	if err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	response.Success(c, template)
}

// ListTemplates 获取模板列表
func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	notifyType := c.Query("notifyType")

	templates, total, err := h.notificationService.ListTemplates(notifyType, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithPage(c, total, page, pageSize, templates)
}

// UpdateTemplate 更新模板
type UpdateTemplateRequest struct {
	TemplateName string `json:"templateName"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Variables    string `json:"variables"`
	Enabled      *int   `json:"enabled"`
}

func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的模板ID")
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	template, err := h.notificationService.GetTemplate(uint(id))
	if err != nil {
		response.NotFound(c, "模板不存在")
		return
	}

	// 更新字段
	if req.TemplateName != "" {
		template.TemplateName = req.TemplateName
	}
	if req.Title != "" {
		template.Title = req.Title
	}
	if req.Content != "" {
		template.Content = req.Content
	}
	if req.Variables != "" {
		template.Variables = req.Variables
	}
	if req.Enabled != nil {
		template.Enabled = *req.Enabled
	}

	if err := h.notificationService.UpdateTemplate(template); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, template)
}

// DeleteTemplate 删除模板
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的模板ID")
		return
	}

	if err := h.notificationService.DeleteTemplate(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// CreateChannel 创建渠道
type CreateChannelRequest struct {
	ChannelCode string `json:"channelCode" binding:"required"`
	ChannelName string `json:"channelName" binding:"required"`
	ChannelType string `json:"channelType" binding:"required"`
	Config      string `json:"config" binding:"required"`
}

func (h *NotificationHandler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	channel := &model.NotificationChannel{
		ChannelCode: req.ChannelCode,
		ChannelName: req.ChannelName,
		ChannelType: req.ChannelType,
		Config:      req.Config,
		Enabled:     1,
	}

	if err := h.notificationService.CreateChannel(channel); err != nil {
		response.Error(c, response.CodeConflict, err.Error())
		return
	}

	response.Success(c, channel)
}

// ListChannels 获取渠道列表
func (h *NotificationHandler) ListChannels(c *gin.Context) {
	channels, err := h.notificationService.ListChannels()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, channels)
}

// UpdateChannel 更新渠道
type UpdateChannelRequest struct {
	ChannelName string `json:"channelName"`
	Config      string `json:"config"`
	Enabled     *int   `json:"enabled"`
}

func (h *NotificationHandler) UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的渠道ID")
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	channel, err := h.notificationService.GetChannel(uint(id))
	if err != nil {
		response.NotFound(c, "渠道不存在")
		return
	}

	// 更新字段
	if req.ChannelName != "" {
		channel.ChannelName = req.ChannelName
	}
	if req.Config != "" {
		channel.Config = req.Config
	}
	if req.Enabled != nil {
		channel.Enabled = *req.Enabled
	}

	if err := h.notificationService.UpdateChannel(channel); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, channel)
}

// DeleteChannel 删除渠道
func (h *NotificationHandler) DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidParams(c, "无效的渠道ID")
		return
	}

	if err := h.notificationService.DeleteChannel(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}
