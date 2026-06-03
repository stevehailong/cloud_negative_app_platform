package integration

import (
	"sync"
	"sync/atomic"
	"time"

	"my-cloud/pkg/prometheus"

	"gorm.io/gorm"
)

// Settings 包含监控相关的集成配置（从 iam_db.system_settings 加载）
type Settings struct {
	PrometheusURL string
	GrafanaURL    string
	GrafanaAPIKey string
}

// Loader 从 iam_db 读取集成配置，并维护一个最新的 Prometheus 客户端实例
// 提供 TTL 缓存以避免每次请求都访问数据库
type Loader struct {
	db       *gorm.DB
	cacheTTL time.Duration

	mu       sync.RWMutex
	settings *Settings
	lastLoad time.Time

	// promClient 作为热替换的指针存放，访问时无需加锁
	promClient atomic.Value // *prometheus.Client
}

// NewLoader 创建并立即执行一次加载
func NewLoader(db *gorm.DB) *Loader {
	l := &Loader{
		db:       db,
		cacheTTL: 30 * time.Second,
	}
	l.Load()
	return l
}

// Get 返回缓存的设置；过期则强制刷新
func (l *Loader) Get() *Settings {
	l.mu.RLock()
	if l.settings != nil && time.Since(l.lastLoad) < l.cacheTTL {
		s := l.settings
		l.mu.RUnlock()
		return s
	}
	l.mu.RUnlock()
	return l.Load()
}

// Load 强制从数据库重新加载
func (l *Loader) Load() *Settings {
	l.mu.Lock()
	defer l.mu.Unlock()

	settings := &Settings{}

	if l.db != nil {
		var rows []struct {
			SettingKey   string `gorm:"column:setting_key"`
			SettingValue string `gorm:"column:setting_value"`
		}
		l.db.Table("system_settings").
			Where("setting_group = ?", "integration").
			Select("setting_key, setting_value").
			Find(&rows)
		for _, r := range rows {
			switch r.SettingKey {
			case "prometheusUrl":
				settings.PrometheusURL = r.SettingValue
			case "grafanaUrl":
				settings.GrafanaURL = r.SettingValue
			case "grafanaApiKey":
				settings.GrafanaAPIKey = r.SettingValue
			}
		}
	}

	l.settings = settings
	l.lastLoad = time.Now()

	// 同步更新 Prometheus 客户端
	if settings.PrometheusURL != "" {
		l.promClient.Store(prometheus.NewClient(settings.PrometheusURL))
	} else {
		// 清空旧客户端
		l.promClient.Store((*prometheus.Client)(nil))
	}
	return settings
}

// PrometheusClient 返回当前可用的 Prometheus 客户端；未配置时返回 nil
// 同时会触发一次 TTL 检查
func (l *Loader) PrometheusClient() *prometheus.Client {
	l.Get() // 触发可能的刷新
	v := l.promClient.Load()
	if v == nil {
		return nil
	}
	c, _ := v.(*prometheus.Client)
	if c == nil {
		return nil
	}
	return c
}
