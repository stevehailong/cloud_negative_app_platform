package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"my-cloud/internal/common/response"
	"net/http"
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

// integrationSetting 从 system_settings.integration 读取单个值
func (h *SettingsHandler) integrationSetting(key string) string {
	var v string
	h.db.Table("system_settings").
		Where("setting_group = ? AND setting_key = ?", "integration", key).
		Pluck("setting_value", &v)
	return v
}

// testConnectionRequest 是测试连接的可选请求体
// 前端未填表单时可不传，由后端从已保存的设置读取
type testConnectionRequest struct {
	URL    string `json:"url"`
	APIKey string `json:"apiKey"`
}

// TestPrometheusConnection 测试 Prometheus 连通性
// POST /api/v1/settings/integration/test-prometheus
func (h *SettingsHandler) TestPrometheusConnection(c *gin.Context) {
	var req testConnectionRequest
	_ = c.ShouldBindJSON(&req)

	url := strings.TrimSpace(req.URL)
	if url == "" {
		url = h.integrationSetting("prometheusUrl")
	}
	if url == "" {
		response.InvalidParams(c, "请先填写 Prometheus 地址")
		return
	}

	client := &http.Client{Timeout: 8 * time.Second}
	endpoint := strings.TrimRight(url, "/") + "/api/v1/query?query=up"
	resp, err := client.Get(endpoint)
	if err != nil {
		response.Error(c, response.CodeInternalError, "连接失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("连接失败: HTTP %d", resp.StatusCode))
		return
	}

	// 解析返回的样本数（不一定有，up{job} 通常会有）
	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			Result []interface{} `json:"result"`
		} `json:"data"`
	}
	_ = json.Unmarshal(body, &parsed)
	if parsed.Status != "success" {
		response.Error(c, response.CodeInternalError, "Prometheus 响应异常")
		return
	}
	response.Success(c, gin.H{
		"message": fmt.Sprintf("连接成功，up 指标返回 %d 个样本", len(parsed.Data.Result)),
		"url":     url,
	})
}

// TestGrafanaConnection 测试 Grafana 连通性
// POST /api/v1/settings/integration/test-grafana
func (h *SettingsHandler) TestGrafanaConnection(c *gin.Context) {
	var req testConnectionRequest
	_ = c.ShouldBindJSON(&req)

	url := strings.TrimSpace(req.URL)
	apiKey := strings.TrimSpace(req.APIKey)
	if url == "" {
		url = h.integrationSetting("grafanaUrl")
	}
	if apiKey == "" {
		apiKey = h.integrationSetting("grafanaApiKey")
	}
	if url == "" {
		response.InvalidParams(c, "请先填写 Grafana 地址")
		return
	}

	client := &http.Client{Timeout: 8 * time.Second}
	endpoint := strings.TrimRight(url, "/") + "/api/health"
	httpReq, _ := http.NewRequest("GET", endpoint, nil)
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		response.Error(c, response.CodeInternalError, "连接失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		response.Error(c, response.CodeInternalError, fmt.Sprintf("连接失败: HTTP %d - %s", resp.StatusCode, string(body)))
		return
	}

	var health struct {
		Database string `json:"database"`
		Version  string `json:"version"`
		Commit   string `json:"commit"`
	}
	_ = json.Unmarshal(body, &health)

	msg := "连接成功"
	if health.Version != "" {
		msg = fmt.Sprintf("连接成功，Grafana 版本 %s", health.Version)
	}
	response.Success(c, gin.H{
		"message": msg,
		"version": health.Version,
		"url":     url,
	})
}

// GetGrafanaConfig 返回前端嵌入 Grafana 所需的配置（不含 API Key）
//
// 优先返回 grafanaPublicUrl（浏览器可访问）；未配置时把 grafanaUrl 里的 docker
// 内部主机名 'grafana' 自动翻译成 'localhost'，方便本地开发直接嵌 iframe。
// GET /api/v1/settings/grafana-config
func (h *SettingsHandler) GetGrafanaConfig(c *gin.Context) {
	publicURL := h.integrationSetting("grafanaPublicUrl")
	internalURL := h.integrationSetting("grafanaUrl")

	url := publicURL
	if url == "" {
		url = browserAccessibleURL(internalURL)
	}
	response.Success(c, gin.H{
		"grafanaUrl": url,
		"enabled":    url != "",
	})
}

// browserAccessibleURL 把 docker 内部主机名翻译成浏览器可达地址
// 当前只针对 'grafana' 这个固定容器名做替换；其它地址原样返回
func browserAccessibleURL(internalURL string) string {
	if internalURL == "" {
		return ""
	}
	// http://grafana:3000 → http://localhost:3000
	// http://grafana:3000/path → http://localhost:3000/path
	return strings.Replace(internalURL, "://grafana:", "://localhost:", 1)
}
