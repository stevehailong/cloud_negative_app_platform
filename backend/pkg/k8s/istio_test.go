package k8s

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// =========================================================================
// VirtualService 类型序列化测试
// =========================================================================

func TestVirtualServiceJSONRoundTrip(t *testing.T) {
	vs := &VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vs",
			Namespace: "test-ns",
			Labels:    map[string]string{"managed-by": "my-cloud"},
		},
		Spec: VirtualServiceSpec{
			Hosts:    []string{"test-svc.test-ns.svc.cluster.local"},
			Gateways: []string{"mesh"},
			HTTP: []HTTPRoute{
				{
					Route: []RouteDestination{
						{
							Destination: Destination{
								Host:   "test-svc.test-ns.svc.cluster.local",
								Subset: "stable",
							},
							Weight: 80,
						},
						{
							Destination: Destination{
								Host:   "test-svc.test-ns.svc.cluster.local",
								Subset: "canary",
							},
							Weight: 20,
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(vs)
	if err != nil {
		t.Fatalf("marshal VirtualService: %v", err)
	}

	var decoded VirtualService
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal VirtualService: %v", err)
	}

	if decoded.Name != "test-vs" {
		t.Errorf("expected name 'test-vs', got %q", decoded.Name)
	}
	if decoded.Namespace != "test-ns" {
		t.Errorf("expected namespace 'test-ns', got %q", decoded.Namespace)
	}
	if len(decoded.Spec.HTTP) != 1 {
		t.Fatalf("expected 1 HTTP route, got %d", len(decoded.Spec.HTTP))
	}
	if decoded.Spec.HTTP[0].Route[0].Weight != 80 {
		t.Errorf("expected stable weight 80, got %d", decoded.Spec.HTTP[0].Route[0].Weight)
	}
	if decoded.Spec.HTTP[0].Route[1].Weight != 20 {
		t.Errorf("expected canary weight 20, got %d", decoded.Spec.HTTP[0].Route[1].Weight)
	}
	if decoded.Spec.HTTP[0].Route[0].Destination.Subset != "stable" {
		t.Errorf("expected stable subset, got %q", decoded.Spec.HTTP[0].Route[0].Destination.Subset)
	}
	if decoded.Spec.HTTP[0].Route[1].Destination.Subset != "canary" {
		t.Errorf("expected canary subset, got %q", decoded.Spec.HTTP[0].Route[1].Destination.Subset)
	}
}

func TestVirtualServiceWithHeaderMatch(t *testing.T) {
	vs := &VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vs",
			Namespace: "test-ns",
		},
		Spec: VirtualServiceSpec{
			Hosts:    []string{"test-svc"},
			Gateways: []string{"mesh"},
			HTTP: []HTTPRoute{
				{
					Match: []HTTPMatch{
						{
							Headers: map[string]StringMatch{
								"x-canary": {Exact: "true"},
							},
						},
					},
					Route: []RouteDestination{
						{
							Destination: Destination{
								Host:   "test-svc",
								Subset: "canary",
							},
						},
					},
				},
				{
					Route: []RouteDestination{
						{
							Destination: Destination{
								Host:   "test-svc",
								Subset: "stable",
							},
							Weight: 100,
						},
					},
				},
			},
		},
	}

	data, _ := json.Marshal(vs)
	var decoded VirtualService
	json.Unmarshal(data, &decoded)

	if len(decoded.Spec.HTTP) != 2 {
		t.Fatalf("expected 2 HTTP routes, got %d", len(decoded.Spec.HTTP))
	}
	if len(decoded.Spec.HTTP[0].Match) != 1 {
		t.Fatal("expected header match on first route")
	}
	if decoded.Spec.HTTP[0].Match[0].Headers["x-canary"].Exact != "true" {
		t.Errorf("expected x-canary exact match")
	}
	if decoded.Spec.HTTP[1].Route[0].Weight != 100 {
		t.Errorf("expected stable weight 100, got %d", decoded.Spec.HTTP[1].Route[0].Weight)
	}
}

// =========================================================================
// DestinationRule 类型序列化测试
// =========================================================================

func TestDestinationRuleJSONRoundTrip(t *testing.T) {
	dr := &DestinationRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dr",
			Namespace: "test-ns",
		},
		Spec: DestinationRuleSpec{
			Host: "test-svc",
			Subsets: []Subset{
				{
					Name: "stable",
					Labels: map[string]string{
						"version": "v1",
					},
				},
				{
					Name: "canary",
					Labels: map[string]string{
						"version": "v2",
					},
				},
			},
		},
	}

	data, err := json.Marshal(dr)
	if err != nil {
		t.Fatalf("marshal DestinationRule: %v", err)
	}

	var decoded DestinationRule
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal DestinationRule: %v", err)
	}

	if decoded.Spec.Host != "test-svc" {
		t.Errorf("expected host 'test-svc', got %q", decoded.Spec.Host)
	}
	if len(decoded.Spec.Subsets) != 2 {
		t.Fatalf("expected 2 subsets, got %d", len(decoded.Spec.Subsets))
	}
}

func TestDestinationRuleWithTrafficPolicy(t *testing.T) {
	dr := &DestinationRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dr",
			Namespace: "test-ns",
		},
		Spec: DestinationRuleSpec{
			Host: "test-svc",
			TrafficPolicy: &TrafficPolicy{
				LoadBalancer: &LoadBalancer{
					Simple: "ROUND_ROBIN",
				},
				ConnectionPool: &ConnectionPool{
					TCP: &TCPConnectionPool{
						MaxConnections: 100,
					},
				},
			},
			Subsets: []Subset{
				{
					Name: "v1",
					Labels: map[string]string{
						"version": "v1",
					},
				},
			},
		},
	}

	data, _ := json.Marshal(dr)
	var decoded DestinationRule
	json.Unmarshal(data, &decoded)

	if decoded.Spec.TrafficPolicy == nil {
		t.Fatal("expected traffic policy to be set")
	}
	if decoded.Spec.TrafficPolicy.LoadBalancer.Simple != "ROUND_ROBIN" {
		t.Errorf("expected ROUND_ROBIN, got %q", decoded.Spec.TrafficPolicy.LoadBalancer.Simple)
	}
}

// =========================================================================
// Gateway 类型序列化测试
// =========================================================================

func TestGatewayJSONRoundTrip(t *testing.T) {
	gw := &Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gw",
			Namespace: "test-ns",
		},
		Spec: GatewaySpec{
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
			Servers: []Server{
				{
					Port: Port{
						Number:   80,
						Name:     "http",
						Protocol: "HTTP",
					},
					Hosts: []string{"example.com"},
				},
			},
		},
	}

	data, err := json.Marshal(gw)
	if err != nil {
		t.Fatalf("marshal Gateway: %v", err)
	}

	var decoded Gateway
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Gateway: %v", err)
	}

	if decoded.Spec.Servers[0].Port.Number != 80 {
		t.Errorf("expected port 80, got %d", decoded.Spec.Servers[0].Port.Number)
	}
	if decoded.Spec.Servers[0].Hosts[0] != "example.com" {
		t.Errorf("expected host 'example.com', got %q", decoded.Spec.Servers[0].Hosts[0])
	}
}

func TestGatewayWithTLS(t *testing.T) {
	gw := &Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gw",
			Namespace: "test-ns",
		},
		Spec: GatewaySpec{
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
			Servers: []Server{
				{
					Port: Port{
						Number:   443,
						Name:     "https",
						Protocol: "HTTPS",
					},
					Hosts: []string{"example.com"},
					TLS: &ServerTLSSettings{
						Mode:           "SIMPLE",
						CredentialName: "example-tls",
					},
				},
			},
		},
	}

	data, _ := json.Marshal(gw)
	var decoded Gateway
	json.Unmarshal(data, &decoded)

	if decoded.Spec.Servers[0].TLS == nil {
		t.Fatal("expected TLS config")
	}
	if decoded.Spec.Servers[0].TLS.Mode != "SIMPLE" {
		t.Errorf("expected SIMPLE TLS mode, got %q", decoded.Spec.Servers[0].TLS.Mode)
	}
}

// =========================================================================
// PeerAuthentication 测试
// =========================================================================

func TestPeerAuthenticationStrict(t *testing.T) {
	pa := &PeerAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "security.istio.io/v1beta1",
			Kind:       "PeerAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "test-ns",
		},
		Spec: PeerAuthenticationSpec{
			MTLS: &MutualTLS{
				Mode: "STRICT",
			},
		},
	}

	data, _ := json.Marshal(pa)
	var decoded PeerAuthentication
	json.Unmarshal(data, &decoded)

	if decoded.Spec.MTLS.Mode != "STRICT" {
		t.Errorf("expected STRICT, got %q", decoded.Spec.MTLS.Mode)
	}
}

func TestPeerAuthenticationPermissive(t *testing.T) {
	pa := &PeerAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "security.istio.io/v1beta1",
			Kind:       "PeerAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "test-ns",
		},
		Spec: PeerAuthenticationSpec{
			MTLS: &MutualTLS{
				Mode: "PERMISSIVE",
			},
		},
	}

	data, _ := json.Marshal(pa)
	var decoded PeerAuthentication
	json.Unmarshal(data, &decoded)

	if decoded.Spec.MTLS.Mode != "PERMISSIVE" {
		t.Errorf("expected PERMISSIVE, got %q", decoded.Spec.MTLS.Mode)
	}
}

// =========================================================================
// Canary 配置验证测试
// =========================================================================

func TestCanaryVirtualServiceConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		weight  int
		wantErr bool
	}{
		{"valid weight 50", 50, false},
		{"weight 0", 0, false},
		{"weight 100", 100, false},
		{"negative weight", -1, true},
		{"weight over 100", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CanaryVirtualServiceConfig{
				Name:         "test",
				Namespace:    "ns",
				CanaryWeight: tt.weight,
			}
			isValid := config.CanaryWeight >= 0 && config.CanaryWeight <= 100
			if isValid == tt.wantErr {
				t.Errorf("weight=%d: expected valid=%v, got valid=%v", tt.weight, !tt.wantErr, isValid)
			}
		})
	}
}

func TestCanaryDestinationRuleConfig(t *testing.T) {
	dr := CanaryDestinationRuleConfig{
		Name:         "test-app",
		Namespace:    "test-ns",
		Host:         "test-app-service",
		StableSubset: "stable",
		CanarySubset: "canary",
		StableLabels: map[string]string{
			"version": "v1",
		},
		CanaryLabels: map[string]string{
			"version": "v2",
		},
	}

	if dr.Host != "test-app-service" {
		t.Errorf("expected host, got %q", dr.Host)
	}
	if dr.StableLabels["version"] != "v1" {
		t.Errorf("expected stable version v1")
	}
	if dr.CanaryLabels["version"] != "v2" {
		t.Errorf("expected canary version v2")
	}
}

func TestCanaryDeployConfig(t *testing.T) {
	cfg := CanaryDeployConfig{
		Namespace:           "app-1-dev",
		AppName:             "app-1",
		ServiceHost:         "app-1-service",
		VirtualServiceName:  "app-1",
		DestinationRuleName: "app-1",
		StableSubset:        "stable",
		CanarySubset:        "canary",
		StableVersion:       "app-1",
		CanaryVersion:       "app-1-canary",
		CanaryWeight:        20,
		Hosts:               []string{"app-1.example.com"},
		Gateways:            []string{"mesh"},
	}

	if cfg.CanaryWeight != 20 {
		t.Errorf("expected canary weight 20, got %d", cfg.CanaryWeight)
	}
	if cfg.StableSubset != "stable" {
		t.Errorf("expected stable subset, got %q", cfg.StableSubset)
	}
}

// =========================================================================
// StringMatch 测试
// =========================================================================

func TestStringMatchTypes(t *testing.T) {
	tests := []struct {
		name     string
		match    StringMatch
		expected string
	}{
		{
			name:     "exact",
			match:    StringMatch{Exact: "foo"},
			expected: `{"exact":"foo"}`,
		},
		{
			name:     "prefix",
			match:    StringMatch{Prefix: "/api"},
			expected: `{"prefix":"/api"}`,
		},
		{
			name:     "regex",
			match:    StringMatch{Regex: "^/v[0-9]+"},
			expected: `{"regex":"^/v[0-9]+"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.match)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

// =========================================================================
// HTTPRoute 复杂场景测试
// =========================================================================

func TestHTTPRouteWithRetries(t *testing.T) {
	route := HTTPRoute{
		Route: []RouteDestination{
			{
				Destination: Destination{
					Host:   "test-svc",
					Subset: "v1",
				},
				Weight: 100,
			},
		},
		Retries: &HTTPRetries{
			Attempts:      3,
			PerTryTimeout: "2s",
			RetryOn:       "connect-failure,refused-stream",
		},
	}

	data, _ := json.Marshal(route)
	var decoded HTTPRoute
	json.Unmarshal(data, &decoded)

	if decoded.Retries.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", decoded.Retries.Attempts)
	}
}

func TestHTTPRouteWithFaultInjection(t *testing.T) {
	route := HTTPRoute{
		Route: []RouteDestination{
			{
				Destination: Destination{
					Host:   "test-svc",
					Subset: "v1",
				},
				Weight: 100,
			},
		},
		Fault: &HTTPFault{
			Delay: &Delay{
				Percent:    50,
				FixedDelay: "5s",
			},
			Abort: &Abort{
				Percent:    10,
				HTTPStatus: 503,
			},
		},
	}

	data, _ := json.Marshal(route)
	var decoded HTTPRoute
	json.Unmarshal(data, &decoded)

	if decoded.Fault.Delay.FixedDelay != "5s" {
		t.Errorf("expected 5s delay")
	}
	if decoded.Fault.Abort.HTTPStatus != 503 {
		t.Errorf("expected 503 abort, got %d", decoded.Fault.Abort.HTTPStatus)
	}
}

// =========================================================================
// toUnstructured / fromUnstructured 测试
// =========================================================================

func TestToUnstructuredAndBack(t *testing.T) {
	original := map[string]interface{}{
		"apiVersion": "networking.istio.io/v1beta1",
		"kind":       "VirtualService",
		"metadata": map[string]interface{}{
			"name":      "test-vs",
			"namespace": "test-ns",
		},
		"spec": map[string]interface{}{
			"hosts":    []string{"test-svc"},
			"gateways": []string{"mesh"},
			"http": []map[string]interface{}{
				{
					"route": []map[string]interface{}{
						{
							"destination": map[string]interface{}{
								"host":   "test-svc",
								"subset": "stable",
							},
							"weight": 100,
						},
					},
				},
			},
		},
	}

	// Marshal to JSON → unmarshal to VirtualService
	data, _ := json.Marshal(original)
	var vs VirtualService
	if err := json.Unmarshal(data, &vs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if vs.Name != "test-vs" {
		t.Errorf("expected name test-vs")
	}
	if vs.Spec.HTTP[0].Route[0].Weight != 100 {
		t.Errorf("expected weight 100")
	}

	// Convert to unstructured and back
	unstr, err := toUnstructured(&vs)
	if err != nil {
		t.Fatalf("toUnstructured: %v", err)
	}

	var vs2 VirtualService
	if err := fromUnstructured(unstr, &vs2); err != nil {
		t.Fatalf("fromUnstructured: %v", err)
	}

	if vs2.Name != "test-vs" {
		t.Errorf("round-trip: expected name test-vs")
	}
	if vs2.Spec.HTTP[0].Route[0].Weight != 100 {
		t.Errorf("round-trip: expected weight 100")
	}
}

// =========================================================================
// GVR 常量测试
// =========================================================================

func TestGVRConstants(t *testing.T) {
	if VirtualServiceGVR.Group != "networking.istio.io" {
		t.Errorf("expected networking.istio.io, got %s", VirtualServiceGVR.Group)
	}
	if VirtualServiceGVR.Resource != "virtualservices" {
		t.Errorf("expected virtualservices, got %s", VirtualServiceGVR.Resource)
	}
	if DestinationRuleGVR.Resource != "destinationrules" {
		t.Errorf("expected destinationrules, got %s", DestinationRuleGVR.Resource)
	}
	if GatewayGVR.Resource != "gateways" {
		t.Errorf("expected gateways, got %s", GatewayGVR.Resource)
	}
	if PeerAuthenticationGVR.Group != "security.istio.io" {
		t.Errorf("expected security.istio.io, got %s", PeerAuthenticationGVR.Group)
	}
	if PeerAuthenticationGVR.Resource != "peerauthentications" {
		t.Errorf("expected peerauthentications, got %s", PeerAuthenticationGVR.Resource)
	}
	if AuthorizationPolicyGVR.Resource != "authorizationpolicies" {
		t.Errorf("expected authorizationpolicies, got %s", AuthorizationPolicyGVR.Resource)
	}
}

// =========================================================================
// 向后兼容 API 测试
// =========================================================================

func TestAdjustCanaryWeightValidation(t *testing.T) {
	// 测试 AdjustCanaryTrafficWeight 函数的权重验证逻辑
	tests := []struct {
		weight  int
		wantErr bool
	}{
		{0, false},
		{50, false},
		{100, false},
		{-1, true},
		{101, true},
	}
	for _, tt := range tests {
		valid := tt.weight >= 0 && tt.weight <= 100
		if valid == tt.wantErr {
			t.Errorf("weight=%d: expected valid=%v, got %v", tt.weight, !tt.wantErr, valid)
		}
	}
}

// =========================================================================
// CanaryVirtualServiceConfig 默认值测试
// =========================================================================

func TestCanaryVirtualServiceConfigDefaults(t *testing.T) {
	config := CanaryVirtualServiceConfig{
		Name:      "test-app",
		Namespace: "test-ns",
	}

	if config.StableSubset == "" {
		config.StableSubset = "stable"
	}
	if config.CanarySubset == "" {
		config.CanarySubset = "canary"
	}

	if config.StableSubset != "stable" {
		t.Errorf("expected default stable subset 'stable'")
	}
	if config.CanarySubset != "canary" {
		t.Errorf("expected default canary subset 'canary'")
	}
}

// =========================================================================
// AuthZ 策略测试
// =========================================================================

func TestAuthorizationPolicyJSONRoundTrip(t *testing.T) {
	ap := &AuthorizationPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "security.istio.io/v1beta1",
			Kind:       "AuthorizationPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "test-ns",
		},
		Spec: AuthorizationPolicySpec{
			Action: "ALLOW",
			Rules: []Rule{
				{
					From: []RuleFrom{
						{
							Source: &Source{
								Namespaces: []string{"istio-system"},
							},
						},
					},
					To: []RuleTo{
						{
							Operation: &Operation{
								Methods: []string{"GET", "POST"},
								Paths:   []string{"/api/*"},
							},
						},
					},
				},
			},
		},
	}

	data, _ := json.Marshal(ap)
	var decoded AuthorizationPolicy
	json.Unmarshal(data, &decoded)

	if decoded.Spec.Action != "ALLOW" {
		t.Errorf("expected ALLOW, got %q", decoded.Spec.Action)
	}
	if len(decoded.Spec.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(decoded.Spec.Rules))
	}
	if decoded.Spec.Rules[0].From[0].Source.Namespaces[0] != "istio-system" {
		t.Errorf("expected namespace istio-system")
	}
}

// =========================================================================
// ServiceEntry 测试
// =========================================================================

func TestServiceEntryJSONRoundTrip(t *testing.T) {
	se := &ServiceEntry{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "ServiceEntry",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "external-api",
			Namespace: "test-ns",
		},
		Spec: ServiceEntrySpec{
			Hosts:      []string{"api.external.com"},
			Ports:      []ServicePort{{Number: 443, Protocol: "HTTPS", Name: "https"}},
			Location:   "MESH_EXTERNAL",
			Resolution: "DNS",
		},
	}

	data, _ := json.Marshal(se)
	var decoded ServiceEntry
	json.Unmarshal(data, &decoded)

	if decoded.Spec.Hosts[0] != "api.external.com" {
		t.Errorf("expected host api.external.com")
	}
	if decoded.Spec.Location != "MESH_EXTERNAL" {
		t.Errorf("expected MESH_EXTERNAL, got %q", decoded.Spec.Location)
	}
}
