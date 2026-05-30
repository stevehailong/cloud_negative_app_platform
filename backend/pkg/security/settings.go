package security

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"gorm.io/gorm"
)

// Settings holds the security configuration loaded from system_settings
type Settings struct {
	SessionTimeout     int      // minutes
	PasswordMinLength  int
	PasswordComplexity []string // ["uppercase","lowercase","number","special"]
	LoginLockEnabled   bool
	LoginLockAttempts  int
	LoginLockDuration  int // minutes
	APIRateLimitEnabled bool
	APIRateLimit       int // requests per minute
	IPWhitelist        []string
}

// SettingsLoader loads security settings from DB with caching
type SettingsLoader struct {
	db       *gorm.DB
	cache    *Settings
	mu       sync.RWMutex
	lastLoad time.Time
	cacheTTL time.Duration
}

// NewSettingsLoader creates a new settings loader
func NewSettingsLoader(db *gorm.DB) *SettingsLoader {
	loader := &SettingsLoader{
		db:       db,
		cacheTTL: 30 * time.Second, // refresh every 30s
	}
	loader.Load() // initial load
	return loader
}

// Get returns cached settings, reloading if stale
func (l *SettingsLoader) Get() *Settings {
	l.mu.RLock()
	if l.cache != nil && time.Since(l.lastLoad) < l.cacheTTL {
		defer l.mu.RUnlock()
		return l.cache
	}
	l.mu.RUnlock()
	return l.Load()
}

// Load forces a reload from database
func (l *SettingsLoader) Load() *Settings {
	l.mu.Lock()
	defer l.mu.Unlock()

	var rows []struct {
		SettingKey   string `gorm:"column:setting_key"`
		SettingValue string `gorm:"column:setting_value"`
	}

	l.db.Table("system_settings").
		Where("setting_group = ?", "security").
		Select("setting_key, setting_value").
		Find(&rows)

	settings := &Settings{
		SessionTimeout:    30,
		PasswordMinLength: 8,
		LoginLockAttempts: 5,
		LoginLockDuration: 30,
		APIRateLimit:      1000,
	}

	kv := make(map[string]string)
	for _, r := range rows {
		kv[r.SettingKey] = r.SettingValue
	}

	if v, ok := kv["sessionTimeout"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			settings.SessionTimeout = n
		}
	}
	if v, ok := kv["passwordMinLength"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			settings.PasswordMinLength = n
		}
	}
	if v, ok := kv["passwordComplexity"]; ok {
		// stored as JSON array string like ["uppercase","lowercase"]
		v = strings.Trim(v, "[]\"")
		if v != "" {
			parts := strings.Split(v, ",")
			for i := range parts {
				parts[i] = strings.Trim(strings.TrimSpace(parts[i]), "\"")
			}
			settings.PasswordComplexity = parts
		}
	}
	if v, ok := kv["loginLockEnabled"]; ok {
		settings.LoginLockEnabled = v == "true"
	}
	if v, ok := kv["loginLockAttempts"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			settings.LoginLockAttempts = n
		}
	}
	if v, ok := kv["loginLockDuration"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			settings.LoginLockDuration = n
		}
	}
	if v, ok := kv["apiRateLimitEnabled"]; ok {
		settings.APIRateLimitEnabled = v == "true"
	}
	if v, ok := kv["apiRateLimit"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			settings.APIRateLimit = n
		}
	}
	if v, ok := kv["ipWhitelist"]; ok && v != "" {
		lines := strings.Split(v, "\n")
		var ips []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				ips = append(ips, line)
			}
		}
		settings.IPWhitelist = ips
	}

	l.cache = settings
	l.lastLoad = time.Now()
	return settings
}

// ValidatePassword checks password against complexity rules
func ValidatePassword(password string, settings *Settings) error {
	if len(password) < settings.PasswordMinLength {
		return fmt.Errorf("密码长度不能少于%d位", settings.PasswordMinLength)
	}

	for _, rule := range settings.PasswordComplexity {
		switch rule {
		case "uppercase":
			if !containsUpper(password) {
				return fmt.Errorf("密码必须包含大写字母")
			}
		case "lowercase":
			if !containsLower(password) {
				return fmt.Errorf("密码必须包含小写字母")
			}
		case "number":
			if !containsDigit(password) {
				return fmt.Errorf("密码必须包含数字")
			}
		case "special":
			if !containsSpecial(password) {
				return fmt.Errorf("密码必须包含特殊字符")
			}
		}
	}
	return nil
}

func containsUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func containsLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func containsSpecial(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
