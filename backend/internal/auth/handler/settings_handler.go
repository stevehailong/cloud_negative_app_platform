package handler

import (
	"fmt"
	"io"
	"my-cloud/internal/common/response"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	db        *gorm.DB
	uploadDir string
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	uploadDir := "/data/uploads"
	if dir := os.Getenv("UPLOAD_DIR"); dir != "" {
		uploadDir = dir
	}
	os.MkdirAll(uploadDir, 0755)
	return &SettingsHandler{db: db, uploadDir: uploadDir}
}

// GetSettings 获取指定分组的设置
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	group := c.Param("group")
	if group == "" {
		response.InvalidParams(c, "缺少分组参数")
		return
	}

	var settings []struct {
		SettingKey   string `gorm:"column:setting_key"`
		SettingValue string `gorm:"column:setting_value"`
	}

	if err := h.db.Table("system_settings").
		Where("setting_group = ?", group).
		Select("setting_key, setting_value").
		Find(&settings).Error; err != nil {
		response.InternalError(c, "获取设置失败")
		return
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.SettingKey] = s.SettingValue
	}

	response.Success(c, result)
}

// GetAllSettings 获取所有设置
func (h *SettingsHandler) GetAllSettings(c *gin.Context) {
	var settings []struct {
		SettingGroup string `gorm:"column:setting_group"`
		SettingKey   string `gorm:"column:setting_key"`
		SettingValue string `gorm:"column:setting_value"`
	}

	if err := h.db.Table("system_settings").
		Select("setting_group, setting_key, setting_value").
		Find(&settings).Error; err != nil {
		response.InternalError(c, "获取设置失败")
		return
	}

	result := make(map[string]map[string]string)
	for _, s := range settings {
		if result[s.SettingGroup] == nil {
			result[s.SettingGroup] = make(map[string]string)
		}
		result[s.SettingGroup][s.SettingKey] = s.SettingValue
	}

	response.Success(c, result)
}

// UpdateSettings 更新指定分组的设置
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	group := c.Param("group")
	if group == "" {
		response.InvalidParams(c, "缺少分组参数")
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.InvalidParams(c, err.Error())
		return
	}

	tx := h.db.Begin()
	for key, value := range data {
		valueStr := fmt.Sprintf("%v", value)
		result := tx.Exec(
			"INSERT INTO system_settings (setting_group, setting_key, setting_value) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE setting_value = ?",
			group, key, valueStr, valueStr,
		)
		if result.Error != nil {
			tx.Rollback()
			response.InternalError(c, "保存设置失败")
			return
		}
	}
	tx.Commit()

	response.Success(c, gin.H{"message": "设置保存成功"})
}

// UploadFile 文件上传
func (h *SettingsHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.InvalidParams(c, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 检查文件大小 (最大5MB)
	if header.Size > 5*1024*1024 {
		response.InvalidParams(c, "文件大小不能超过5MB")
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true}
	if !allowedExts[ext] {
		response.InvalidParams(c, "仅支持PNG/JPG/GIF/SVG格式")
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(h.uploadDir, filename)

	// 保存文件
	out, err := os.Create(savePath)
	if err != nil {
		response.InternalError(c, "保存文件失败")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		response.InternalError(c, "保存文件失败")
		return
	}

	// 返回访问URL
	fileURL := "/api/v1/uploads/" + filename
	response.Success(c, gin.H{
		"url":      fileURL,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

// ServeFile 静态文件服务
func (h *SettingsHandler) ServeFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		response.NotFound(c, "文件不存在")
		return
	}
	// 去掉前导斜杠
	filename = strings.TrimPrefix(filename, "/")

	// 安全检查：防止路径遍历
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		response.InvalidParams(c, "无效的文件名")
		return
	}

	filePath := filepath.Join(h.uploadDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.NotFound(c, "文件不存在")
		return
	}

	c.File(filePath)
}
