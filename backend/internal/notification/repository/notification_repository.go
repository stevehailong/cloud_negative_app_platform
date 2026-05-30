package repository

import (
	"my-cloud/internal/notification/model"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 创建通知
func (r *NotificationRepository) Create(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

// GetByID 根据ID获取通知
func (r *NotificationRepository) GetByID(id uint) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.First(&notification, id).Error
	return &notification, err
}

// List 获取通知列表
func (r *NotificationRepository) List(notifyType, status string, page, pageSize int) ([]*model.Notification, int64, error) {
	var notifications []*model.Notification
	var total int64

	query := r.db.Model(&model.Notification{})
	if notifyType != "" {
		query = query.Where("notify_type = ?", notifyType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&notifications).Error
	return notifications, total, err
}

// Update 更新通知
func (r *NotificationRepository) Update(notification *model.Notification) error {
	return r.db.Save(notification).Error
}

// TemplateRepository 通知模板仓库
type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create 创建模板
func (r *TemplateRepository) Create(template *model.NotificationTemplate) error {
	return r.db.Create(template).Error
}

// GetByID 根据ID获取模板
func (r *TemplateRepository) GetByID(id uint) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	err := r.db.First(&template, id).Error
	return &template, err
}

// GetByCode 根据Code获取模板
func (r *TemplateRepository) GetByCode(code string) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	err := r.db.Where("template_code = ?", code).First(&template).Error
	return &template, err
}

// List 获取模板列表
func (r *TemplateRepository) List(notifyType string, page, pageSize int) ([]*model.NotificationTemplate, int64, error) {
	var templates []*model.NotificationTemplate
	var total int64

	query := r.db.Model(&model.NotificationTemplate{})
	if notifyType != "" {
		query = query.Where("notify_type = ?", notifyType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&templates).Error
	return templates, total, err
}

// Update 更新模板
func (r *TemplateRepository) Update(template *model.NotificationTemplate) error {
	return r.db.Save(template).Error
}

// Delete 删除模板
func (r *TemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.NotificationTemplate{}, id).Error
}

// ChannelRepository 通知渠道仓库
type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

// Create 创建渠道
func (r *ChannelRepository) Create(channel *model.NotificationChannel) error {
	return r.db.Create(channel).Error
}

// GetByID 根据ID获取渠道
func (r *ChannelRepository) GetByID(id uint) (*model.NotificationChannel, error) {
	var channel model.NotificationChannel
	err := r.db.First(&channel, id).Error
	return &channel, err
}

// GetByCode 根据Code获取渠道
func (r *ChannelRepository) GetByCode(code string) (*model.NotificationChannel, error) {
	var channel model.NotificationChannel
	err := r.db.Where("channel_code = ?", code).First(&channel).Error
	return &channel, err
}

// List 获取渠道列表
func (r *ChannelRepository) List() ([]*model.NotificationChannel, error) {
	var channels []*model.NotificationChannel
	err := r.db.Where("enabled = ?", 1).Find(&channels).Error
	return channels, err
}

// Update 更新渠道
func (r *ChannelRepository) Update(channel *model.NotificationChannel) error {
	return r.db.Save(channel).Error
}

// Delete 删除渠道
func (r *ChannelRepository) Delete(id uint) error {
	return r.db.Delete(&model.NotificationChannel{}, id).Error
}
