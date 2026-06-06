package service

import (
	"testing"
)

func TestBuildCanaryServiceSpec(t *testing.T) {
	svc := BuildCanaryServiceSpec(
		"app-1-canary-svc",
		"app-1-prod",
		"app-1",
		"app-1-canary",
		80, 8080,
	)

	if svc.Name != "app-1-canary-svc" {
		t.Errorf("expected name app-1-canary-svc, got %s", svc.Name)
	}
	if svc.Namespace != "app-1-prod" {
		t.Errorf("expected namespace app-1-prod, got %s", svc.Namespace)
	}

	// 关键断言：selector 必须包含 version=app-1-canary，确保仅选中 canary Pod
	if v, ok := svc.Spec.Selector["version"]; !ok {
		t.Error("canary service selector missing 'version' key")
	} else if v != "app-1-canary" {
		t.Errorf("expected version selector app-1-canary, got %s", v)
	}
	if v, ok := svc.Spec.Selector["app"]; !ok || v != "app-1" {
		t.Errorf("expected app selector app-1, got %s", v)
	}

	// 必须有 canary role 标签
	if v, ok := svc.Labels["role"]; !ok || v != "canary" {
		t.Errorf("expected role=canary label, got %s", v)
	}

	// 端口
	if len(svc.Spec.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(svc.Spec.Ports))
	}
	if svc.Spec.Ports[0].Port != 80 {
		t.Errorf("expected port 80, got %d", svc.Spec.Ports[0].Port)
	}
	if svc.Spec.Ports[0].TargetPort.IntValue() != 8080 {
		t.Errorf("expected targetPort 8080, got %d", svc.Spec.Ports[0].TargetPort.IntValue())
	}
}

func TestBuildCanaryIngressSpec_WeightMode(t *testing.T) {
	ing := BuildCanaryIngressSpec(
		"app-1-canary-ingress",
		"app-1-prod",
		"app-1.example.com",
		"/", "Prefix",
		"app-1-canary-svc", 80,
		"weight", 30,
		"", "", "", "",
		map[string]string{"app": "app-1"},
	)

	// 名称和命名空间
	if ing.Name != "app-1-canary-ingress" {
		t.Errorf("expected name app-1-canary-ingress, got %s", ing.Name)
	}
	if ing.Namespace != "app-1-prod" {
		t.Errorf("expected namespace app-1-prod, got %s", ing.Namespace)
	}

	// 注解：必须有 canary: "true" 和 canary-weight
	if v, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary"]; !ok || v != "true" {
		t.Errorf("expected canary: true annotation, got %q", v)
	}
	if v, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary-weight"]; !ok || v != "30" {
		t.Errorf("expected canary-weight: 30, got %q", v)
	}

	// weight 模式下不应有 header/cookie 注解
	for _, key := range []string{
		"nginx.ingress.kubernetes.io/canary-by-header",
		"nginx.ingress.kubernetes.io/canary-by-cookie",
	} {
		if v, ok := ing.Annotations[key]; ok {
			t.Errorf("unexpected annotation %s=%q in weight mode", key, v)
		}
	}

	// 后端服务
	if len(ing.Spec.Rules) == 0 {
		t.Fatal("expected at least 1 rule")
	}
	rule := ing.Spec.Rules[0]
	if rule.Host != "app-1.example.com" {
		t.Errorf("expected host app-1.example.com, got %s", rule.Host)
	}
	if rule.HTTP == nil || len(rule.HTTP.Paths) == 0 {
		t.Fatal("expected at least 1 path")
	}
	backend := rule.HTTP.Paths[0].Backend
	if backend.Service == nil || backend.Service.Name != "app-1-canary-svc" {
		t.Errorf("expected backend service app-1-canary-svc, got %v", backend.Service)
	}
}

func TestBuildCanaryIngressSpec_HeaderMode(t *testing.T) {
	ing := BuildCanaryIngressSpec(
		"app-1-canary-ingress",
		"app-1-prod",
		"app-1.example.com",
		"/api", "Exact",
		"app-1-canary-svc", 443,
		"header", 0,
		"x-version", "canary", "",
		"my-tls-secret",
		map[string]string{"app": "app-1"},
	)

	// 注解：header 模式
	if v, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary-by-header"]; !ok || v != "x-version" {
		t.Errorf("expected canary-by-header: x-version, got %q", v)
	}
	if v, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary-by-header-value"]; !ok || v != "canary" {
		t.Errorf("expected canary-by-header-value: canary, got %q", v)
	}

	// header 模式下不应有 canary-weight
	if v, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary-weight"]; ok {
		t.Errorf("unexpected canary-weight=%q in header mode", v)
	}

	// TLS
	if len(ing.Spec.TLS) == 0 || ing.Spec.TLS[0].SecretName != "my-tls-secret" {
		t.Error("expected TLS with my-tls-secret")
	}

	// 端口
	if ing.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number != 443 {
		t.Errorf("expected backend port 443, got %d", ing.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number)
	}
}

func TestBuildCanaryIngressSpec_CookieMode(t *testing.T) {
	ing := BuildCanaryIngressSpec(
		"app-1-canary-ingress",
		"app-1-prod",
		"app-1.example.com",
		"/", "Prefix",
		"app-1-canary-svc", 80,
		"cookie", 0,
		"", "", "canary_cookie",
		"",
		map[string]string{},
	)

	if v, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary-by-cookie"]; !ok || v != "canary_cookie" {
		t.Errorf("expected canary-by-cookie: canary_cookie, got %q", v)
	}
}

func TestBuildCanaryIngressSpec_WeightHeaderMode(t *testing.T) {
	ing := BuildCanaryIngressSpec(
		"app-1-canary-ingress",
		"app-1-prod",
		"app-1.example.com",
		"/", "Prefix",
		"app-1-canary-svc", 80,
		"weight_header", 50,
		"x-version", "canary", "",
		"",
		map[string]string{},
	)

	// 同时有 canary-weight 和 canary-by-header
	if v, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary-weight"]; !ok || v != "50" {
		t.Errorf("expected canary-weight: 50, got %q", v)
	}
	if v, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary-by-header"]; !ok || v != "x-version" {
		t.Errorf("expected canary-by-header: x-version, got %q", v)
	}
}

func TestStringPtr(t *testing.T) {
	s := stringPtr("test")
	if s == nil || *s != "test" {
		t.Errorf("expected 'test', got %v", s)
	}
}

func TestBuildCanaryServiceSpec_ServiceType(t *testing.T) {
	svc := BuildCanaryServiceSpec("test-svc", "ns", "myapp", "myapp-canary", 80, 3000)
	if svc.Spec.Type != "ClusterIP" {
		t.Errorf("expected ClusterIP service type, got %s", svc.Spec.Type)
	}
}

func TestBuildCanaryIngressSpec_EmptyTLS(t *testing.T) {
	ing := BuildCanaryIngressSpec(
		"test-ing", "ns", "host.example.com",
		"/", "Prefix",
		"svc", 80,
		"weight", 10,
		"", "", "", "",
		map[string]string{},
	)
	if len(ing.Spec.TLS) != 0 {
		t.Error("expected no TLS when tlsSecretName is empty")
	}
}
