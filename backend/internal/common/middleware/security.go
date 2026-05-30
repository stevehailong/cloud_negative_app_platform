package middleware

import (
	"my-cloud/internal/common/response"
	"my-cloud/pkg/security"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory sliding window rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
	}
	// Cleanup expired entries periodically
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-time.Minute)
	for key, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = valid
		}
	}
}

// Allow checks if a request from the given key is within the rate limit
func (rl *RateLimiter) Allow(key string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	// Filter expired entries
	var valid []time.Time
	for _, t := range rl.requests[key] {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limit {
		rl.requests[key] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

// APIRateLimit returns a rate limiting middleware
func APIRateLimit(settingsLoader *security.SettingsLoader) gin.HandlerFunc {
	limiter := NewRateLimiter()

	return func(c *gin.Context) {
		settings := settingsLoader.Get()
		if !settings.APIRateLimitEnabled {
			c.Next()
			return
		}

		// Rate limit by client IP
		clientIP := c.ClientIP()
		if !limiter.Allow(clientIP, settings.APIRateLimit) {
			response.Error(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPWhitelist returns an IP whitelist middleware
func IPWhitelist(settingsLoader *security.SettingsLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings := settingsLoader.Get()

		// If whitelist is empty, allow all
		if len(settings.IPWhitelist) == 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()

		for _, allowed := range settings.IPWhitelist {
			if allowed == clientIP {
				c.Next()
				return
			}
			// Support CIDR notation
			if strings.Contains(allowed, "/") {
				_, ipNet, err := net.ParseCIDR(allowed)
				if err == nil && ipNet.Contains(net.ParseIP(clientIP)) {
					c.Next()
					return
				}
			}
		}

		response.Error(c, 403, "您的IP地址不在允许访问列表中")
		c.Abort()
	}
}
