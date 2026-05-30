package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"my-cloud/internal/notification/model"
	"my-cloud/internal/notification/repository"
	"strings"
	"time"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	templateRepo     *repository.TemplateRepository
	channelRepo      *repository.ChannelRepository
}

func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	templateRepo *repository.TemplateRepository,
	channelRepo *repository.ChannelRepository,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		channelRepo:      channelRepo,
	}
}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(notification *model.Notification) error {
	notification.Status = "pending"
	if err := s.notificationRepo.Create(notification); err != nil {
		return err
	}

	// TODO: 异步发送通知
	// 根据channel类型调用不同的发送器
	go s.asyncSendNotification(notification)

	return nil
}

// SendNotificationByTemplate 根据模板发送通知
func (s *NotificationService) SendNotificationByTemplate(
	templateCode string,
	params map[string]interface{},
	receiverType string,
	receiverIDs []uint,
) error {
	// 获取模板
	template, err := s.templateRepo.GetByCode(templateCode)
	if err != nil {
		return errors.New("模板不存在")
	}

	if template.Enabled != 1 {
		return errors.New("模板已禁用")
	}

	// 渲染模板
	title := s.renderTemplate(template.Title, params)
	content := s.renderTemplate(template.Content, params)

	// 转换receiverIDs为字符串
	receiverIDStrs := make([]string, len(receiverIDs))
	for i, id := range receiverIDs {
		receiverIDStrs[i] = fmt.Sprintf("%d", id)
	}

	// 序列化params
	paramsJSON, _ := json.Marshal(params)

	notification := &model.Notification{
		Title:        title,
		Content:      content,
		NotifyType:   template.NotifyType,
		Channel:      template.Channel,
		ReceiverType: receiverType,
		ReceiverIDs:  strings.Join(receiverIDStrs, ","),
		TemplateID:   &template.ID,
		Params:       string(paramsJSON),
	}

	return s.SendNotification(notification)
}

// renderTemplate 简单的模板渲染（替换变量）
func (s *NotificationService) renderTemplate(template string, params map[string]interface{}) string {
	result := template
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// asyncSendNotification 异步发送通知
func (s *NotificationService) asyncSendNotification(notification *model.Notification) {
	// TODO: 根据channel类型调用不同的发送器
	// 这里模拟发送过程
	time.Sleep(1 * time.Second)

	// 模拟发送成功
	now := time.Now()
	notification.Status = "sent"
	notification.SentAt = &now
	s.notificationRepo.Update(notification)

	// 实际实现中，应该根据channel类型调用不同的发送器：
	// - email: 调用SMTP
	// - sms: 调用短信API
	// - dingtalk: 调用钉钉机器人API
	// - slack: 调用Slack API
	// - webhook: 调用自定义Webhook
}

// GetNotification 获取通知详情
func (s *NotificationService) GetNotification(id uint) (*model.Notification, error) {
	return s.notificationRepo.GetByID(id)
}

// ListNotifications 获取通知列表
func (s *NotificationService) ListNotifications(notifyType, status string, page, pageSize int) ([]*model.Notification, int64, error) {
	return s.notificationRepo.List(notifyType, status, page, pageSize)
}

// CreateTemplate 创建模板
func (s *NotificationService) CreateTemplate(template *model.NotificationTemplate) error {
	// 检查code是否已存在
	if existing, _ := s.templateRepo.GetByCode(template.TemplateCode); existing != nil {
		return errors.New("模板代码已存在")
	}

	return s.templateRepo.Create(template)
}

// GetTemplate 获取模板详情
func (s *NotificationService) GetTemplate(id uint) (*model.NotificationTemplate, error) {
	return s.templateRepo.GetByID(id)
}

// ListTemplates 获取模板列表
func (s *NotificationService) ListTemplates(notifyType string, page, pageSize int) ([]*model.NotificationTemplate, int64, error) {
	return s.templateRepo.List(notifyType, page, pageSize)
}

// UpdateTemplate 更新模板
func (s *NotificationService) UpdateTemplate(template *model.NotificationTemplate) error {
	// 检查模板是否存在
	existing, err := s.templateRepo.GetByID(template.ID)
	if err != nil {
		return errors.New("模板不存在")
	}

	// 如果修改了code，检查新code是否已被占用
	if existing.TemplateCode != template.TemplateCode {
		if dup, _ := s.templateRepo.GetByCode(template.TemplateCode); dup != nil {
			return errors.New("模板代码已存在")
		}
	}

	return s.templateRepo.Update(template)
}

// DeleteTemplate 删除模板
func (s *NotificationService) DeleteTemplate(id uint) error {
	return s.templateRepo.Delete(id)
}

// CreateChannel 创建渠道
func (s *NotificationService) CreateChannel(channel *model.NotificationChannel) error {
	// 检查code是否已存在
	if existing, _ := s.channelRepo.GetByCode(channel.ChannelCode); existing != nil {
		return errors.New("渠道代码已存在")
	}

	return s.channelRepo.Create(channel)
}

// GetChannel 获取渠道详情
func (s *NotificationService) GetChannel(id uint) (*model.NotificationChannel, error) {
	return s.channelRepo.GetByID(id)
}

// ListChannels 获取渠道列表
func (s *NotificationService) ListChannels() ([]*model.NotificationChannel, error) {
	return s.channelRepo.List()
}

// UpdateChannel 更新渠道
func (s *NotificationService) UpdateChannel(channel *model.NotificationChannel) error {
	// 检查渠道是否存在
	existing, err := s.channelRepo.GetByID(channel.ID)
	if err != nil {
		return errors.New("渠道不存在")
	}

	// 如果修改了code，检查新code是否已被占用
	if existing.ChannelCode != channel.ChannelCode {
		if dup, _ := s.channelRepo.GetByCode(channel.ChannelCode); dup != nil {
			return errors.New("渠道代码已存在")
		}
	}

	return s.channelRepo.Update(channel)
}

// DeleteChannel 删除渠道
func (s *NotificationService) DeleteChannel(id uint) error {
	return s.channelRepo.Delete(id)
}
