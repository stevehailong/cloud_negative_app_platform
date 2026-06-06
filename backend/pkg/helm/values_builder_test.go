package helm

import (
	"testing"
)

func TestSetEnvConfig_MergesWithTemplate(t *testing.T) {
	b := NewValuesBuilder()

	// 模拟模板已有的 env（GIN_MODE, DB_HOST 等）
	b.values["env"] = []interface{}{
		map[string]interface{}{"name": "GIN_MODE", "value": "release"},
		map[string]interface{}{"name": "DB_HOST", "value": "mysql-service"},
		map[string]interface{}{"name": "DB_PORT", "value": "3306"},
	}

	config := DeploymentConfig{
		WorkloadName:    "app-1-canary",
		EnvVars:         map[string]string{"PORT": "9891"},
		TracingEnabled:  true,
		TracingEndpoint: "http://monitor-service:8090/internal/v1/traces/spans",
	}

	b.SetEnvConfig(config)

	envList, ok := b.values["env"].([]interface{})
	if !ok {
		t.Fatal("env is not a list")
	}

	// 收集 env names
	envMap := make(map[string]string)
	for _, item := range envList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		value, _ := m["value"].(string)
		envMap[name] = value
	}

	// 模板原有的必须保留
	if v, ok := envMap["GIN_MODE"]; !ok || v != "release" {
		t.Errorf("GIN_MODE lost or wrong: %q", v)
	}
	if v, ok := envMap["DB_HOST"]; !ok || v != "mysql-service" {
		t.Errorf("DB_HOST lost or wrong: %q", v)
	}
	if v, ok := envMap["DB_PORT"]; !ok || v != "3306" {
		t.Errorf("DB_PORT lost or wrong: %q", v)
	}

	// Config 新增的 PORT 必须有
	if v, ok := envMap["PORT"]; !ok || v != "9891" {
		t.Errorf("PORT missing or wrong: %q", v)
	}

	// Tracing 自动注入的必须有
	if v, ok := envMap["TRACE_ENABLED"]; !ok || v != "true" {
		t.Errorf("TRACE_ENABLED missing or wrong: %q", v)
	}
	if v, ok := envMap["TRACE_SERVICE_NAME"]; !ok || v != "app-1-canary" {
		t.Errorf("TRACE_SERVICE_NAME missing or wrong: %q", v)
	}
}

func TestSetEnvConfig_OverridesTemplate(t *testing.T) {
	b := NewValuesBuilder()

	b.values["env"] = []interface{}{
		map[string]interface{}{"name": "PORT", "value": "8080"},
		map[string]interface{}{"name": "GIN_MODE", "value": "debug"},
	}

	config := DeploymentConfig{
		WorkloadName: "app-1",
		EnvVars:      map[string]string{"PORT": "9891"},
	}

	b.SetEnvConfig(config)

	envList := b.values["env"].([]interface{})
	envMap := make(map[string]string)
	for _, item := range envList {
		m := item.(map[string]interface{})
		envMap[m["name"].(string)] = m["value"].(string)
	}

	// PORT 应该被 config 覆盖
	if v := envMap["PORT"]; v != "9891" {
		t.Errorf("PORT should be overridden to 9891, got %q", v)
	}
	// GIN_MODE 应该保留模板值
	if v := envMap["GIN_MODE"]; v != "debug" {
		t.Errorf("GIN_MODE should be debug, got %q", v)
	}
}

func TestSetEnvConfig_NoTemplate(t *testing.T) {
	b := NewValuesBuilder()

	config := DeploymentConfig{
		WorkloadName: "app-1",
		EnvVars:      map[string]string{"PORT": "8080"},
	}

	b.SetEnvConfig(config)

	envList := b.values["env"].([]interface{})
	envMap := make(map[string]string)
	for _, item := range envList {
		m := item.(map[string]interface{})
		envMap[m["name"].(string)] = m["value"].(string)
	}

	if v := envMap["PORT"]; v != "8080" {
		t.Errorf("PORT missing: %q", v)
	}
}
