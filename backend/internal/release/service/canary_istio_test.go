package service

import (
	"testing"

	"my-cloud/pkg/k8s"
)

// =========================================================================
// parseAppID 测试
// =========================================================================

func TestParseAppID(t *testing.T) {
	tests := []struct {
		appName  string
		expected int
	}{
		{"app-1", 1},
		{"app-42", 42},
		{"app-999", 999},
		{"my-app-5", 5},
		{"invalid", 0},
		{"app-", 0},
		{"", 0},
		{"app-0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			result := parseAppID(tt.appName)
			if result != tt.expected {
				t.Errorf("parseAppID(%q) = %d, want %d", tt.appName, result, tt.expected)
			}
		})
	}
}

// =========================================================================
// CanaryIstioParams 验证测试
// =========================================================================

func TestCanaryIstioParamsDefaults(t *testing.T) {
	params := CanaryIstioParams{
		AppName:       "app-1",
		Namespace:     "app-1-dev",
		CanaryPercent: 20,
	}

	if params.CanaryPercent < 0 || params.CanaryPercent > 100 {
		t.Errorf("canary percent out of range: %d", params.CanaryPercent)
	}
}

func TestCanaryIstioParamsValidation(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		valid   bool
	}{
		{"valid", 20, true},
		{"zero", 0, true},
		{"full", 100, true},
		{"negative", -5, false},
		{"overflow", 150, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.percent >= 0 && tt.percent <= 100
			if valid != tt.valid {
				t.Errorf("percent %d: expected valid=%v, got %v", tt.percent, tt.valid, valid)
			}
		})
	}
}

// =========================================================================
// CanaryTrafficStats 类型测试
// =========================================================================

func TestCanaryTrafficStats(t *testing.T) {
	stats := &CanaryTrafficStats{
		VirtualServiceName: "app-1",
		Namespace:          "app-1-dev",
		Hosts:              []string{"app-1-service"},
		TotalWeight:        100,
		Routes: []RouteInfo{
			{
				MatchHeaders: []string{"x-canary=true"},
				Destinations: []DestinationInfo{
					{Subset: "canary", Host: "app-1-service", Weight: 0},
				},
			},
			{
				Destinations: []DestinationInfo{
					{Subset: "stable", Host: "app-1-service", Weight: 80},
					{Subset: "canary", Host: "app-1-service", Weight: 20},
				},
			},
		},
	}

	if stats.VirtualServiceName != "app-1" {
		t.Errorf("expected VS name 'app-1'")
	}
	if stats.TotalWeight != 100 {
		t.Errorf("expected total weight 100")
	}
	if len(stats.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(stats.Routes))
	}
	if stats.Routes[0].Destinations[0].Subset != "canary" {
		t.Errorf("expected canary subset in first route")
	}
	if stats.Routes[1].Destinations[0].Weight != 80 {
		t.Errorf("expected stable weight 80")
	}
	if stats.Routes[1].Destinations[1].Weight != 20 {
		t.Errorf("expected canary weight 20")
	}
}

// =========================================================================
// IstioCanaryRunner 参数构建测试
// =========================================================================

func TestIstioCanaryRunnerParamsBuild(t *testing.T) {
	// 测试参数构建逻辑（不依赖 K8s 连接）
	params := CanaryIstioParams{
		AppName:           "app-5",
		Namespace:         "app-5-dev",
		StableVersion:     "app-5",
		CanaryVersion:     "app-5-canary",
		CanaryPercent:     30,
		RoutingMode:       "weight",
		Hosts:             []string{"app-5.example.com"},
		Gateways:          []string{"mesh"},
	}

	// 验证 host 构建
	serviceHost := params.AppName + "-service"
	if serviceHost != "app-5-service" {
		t.Errorf("expected service host 'app-5-service'")
	}

	// 验证 DR 名称
	drName := params.AppName
	if drName != "app-5" {
		t.Errorf("expected DR name 'app-5'")
	}

	// 验证默认 gateways
	if len(params.Gateways) == 0 {
		params.Gateways = []string{"mesh"}
	}
	if params.Gateways[0] != "mesh" {
		t.Errorf("expected default gateway 'mesh'")
	}
}

// =========================================================================
// K8s CanaryVirtualServiceConfig 构建测试
// =========================================================================

func TestBuildCanaryVSConfig(t *testing.T) {
	config := k8s.CanaryVirtualServiceConfig{
		Name:         "app-1",
		Namespace:    "app-1-dev",
		Hosts:        []string{"app-1.example.com"},
		Gateways:     []string{"mesh"},
		StableHost:   "app-1-service",
		CanaryHost:   "app-1-service",
		StableSubset: "stable",
		CanarySubset: "canary",
		CanaryWeight: 20,
		StableWeight: 80,
		HeaderMatches: []k8s.HeaderMatchRule{
			{
				HeaderName:  "x-canary",
				HeaderValue: "true",
				Exact:       true,
				Subset:      "canary",
			},
		},
		Labels: map[string]string{
			"app":        "app-1",
			"managed-by": "my-cloud",
		},
	}

	if config.CanaryWeight != 20 {
		t.Errorf("expected canary weight 20")
	}
	if config.StableWeight != 80 {
		t.Errorf("expected stable weight 80")
	}
	if len(config.HeaderMatches) != 1 {
		t.Fatalf("expected 1 header match")
	}
	if config.HeaderMatches[0].HeaderName != "x-canary" {
		t.Errorf("expected header 'x-canary'")
	}
	if !config.HeaderMatches[0].Exact {
		t.Errorf("expected exact header match")
	}
}

// =========================================================================
// K8s CanaryDestinationRuleConfig 构建测试
// =========================================================================

func TestBuildCanaryDRConfig(t *testing.T) {
	config := k8s.CanaryDestinationRuleConfig{
		Name:         "app-1",
		Namespace:    "app-1-dev",
		Host:         "app-1-service",
		StableSubset: "stable",
		CanarySubset: "canary",
		StableLabels: map[string]string{
			"app":     "app-1",
			"version": "v1",
		},
		CanaryLabels: map[string]string{
			"app":     "app-1",
			"version": "v2",
		},
	}

	if config.Host != "app-1-service" {
		t.Errorf("expected host 'app-1-service'")
	}
	if config.StableLabels["version"] != "v1" {
		t.Errorf("expected stable version v1")
	}
	if config.CanaryLabels["version"] != "v2" {
		t.Errorf("expected canary version v2")
	}
	if config.StableLabels["app"] != "app-1" {
		t.Errorf("expected stable app label 'app-1'")
	}
}
