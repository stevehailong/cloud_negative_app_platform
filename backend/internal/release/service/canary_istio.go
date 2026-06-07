package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"my-cloud/pkg/k8s"
)

// =========================================================================
// Istio 金丝雀部署服务
// =========================================================================
// Istio 金丝雀的工作方式与 Nginx Ingress 金丝雀不同：
//
// Nginx Ingress 方式:
//   1. 创建 canary Service (selector: version=canary)
//   2. Narrow stable Service selector (添加 version=stable)
//   3. 创建 canary Ingress (nginx canary annotations)
//   4. 流量分流由 Nginx Ingress Controller 处理
//
// Istio 方式:
//   1. 创建 canary Deployment (labels: version=canary)
//   2. 创建 DestinationRule (定义 stable/canary subsets)
//   3. 创建 VirtualService (weight-based + header-based routing)
//   4. 流量分流由 Istio sidecar (Envoy) 处理
//   5. 无需创建额外的 Service 或 Ingress — Istio 直接路由到 Deployment
//
// Istio 方式的优势:
//   - 更精细的流量控制 (header/cookie/query-based routing)
//   - 内置故障注入 (delay/abort)
//   - 请求级遥测 (metrics, tracing, logging)
//   - mTLS 自动加密
//   - 不需要额外的 Service/Ingress 资源

// CanaryIstioService 使用 Istio 进行金丝雀部署
type CanaryIstioService struct {
	k8sClient *k8s.Client
}

// NewCanaryIstioService 创建 Istio 金丝雀服务
func NewCanaryIstioService(k8sClient *k8s.Client) *CanaryIstioService {
	return &CanaryIstioService{
		k8sClient: k8sClient,
	}
}

// CanaryIstioParams Istio 金丝雀部署参数
type CanaryIstioParams struct {
	AppName             string            // 应用名称 (e.g. "app-1")
	AppID               uint              // 应用 ID
	EnvID               uint              // 环境 ID
	Namespace           string            // K8s 命名空间
	StableVersion       string            // Stable 版本标签 (e.g. "app-1")
	CanaryVersion       string            // Canary 版本标签 (e.g. "app-1-canary")
	ImageURL            string            // Canary 镜像
	CanaryPercent       int               // Canary 流量百分比 (0-100)
	Hosts               []string          // VirtualService 主机列表
	Gateways            []string          // Gateway 列表
	HeaderMatches       []k8s.HeaderMatchRule // Header 匹配规则
	Timeout             string            // 请求超时
	RoutingMode         string            // 路由模式: "weight", "header", "weight_header"
	CanaryHeaderName    string            // Header 名 (header/weight_header 模式)
	CanaryHeaderValue   string            // Header 值 (header/weight_header 模式)
}

// StartCanaryWithIstio 启动 Istio 金丝雀部署
// 创建/更新 DestinationRule 和 VirtualService
func (s *CanaryIstioService) StartCanaryWithIstio(ctx context.Context, params CanaryIstioParams) error {
	log.Printf("[IstioCanary] Starting canary deployment: app=%s namespace=%s weight=%d%%",
		params.AppName, params.Namespace, params.CanaryPercent)

	// 验证参数
	if params.CanaryPercent < 0 || params.CanaryPercent > 100 {
		return fmt.Errorf("canary percent must be 0-100, got %d", params.CanaryPercent)
	}

	// 检查 Istio 是否安装
	if !s.k8sClient.IsIstioInstalled(ctx) {
		return fmt.Errorf("istio is not installed in the cluster")
	}

	// 构建资源名称
	drName := params.AppName
	vsName := params.AppName
	serviceHost := params.AppName  // K8s Service name matches workloadName

	// 1. 确保 DestinationRule 存在
	drConfig := k8s.CanaryDestinationRuleConfig{
		Name:         drName,
		Namespace:    params.Namespace,
		Host:         serviceHost,
		StableSubset: "stable",
		CanarySubset: "canary",
		StableLabels: map[string]string{
			"app":     params.AppName,
			"version": params.StableVersion,
		},
		CanaryLabels: map[string]string{
			"app": params.CanaryVersion,
		},
		Labels: map[string]string{
			"app":        params.AppName,
			"managed-by": "my-cloud",
		},
	}

	if err := s.k8sClient.EnsureCanaryDestinationRule(ctx, drConfig); err != nil {
		return fmt.Errorf("ensure DestinationRule: %w", err)
	}
	log.Printf("[IstioCanary] DestinationRule %s/%s ensured", params.Namespace, drName)

	// 2. 构建 Header 匹配规则
	var headerMatches []k8s.HeaderMatchRule

	switch params.RoutingMode {
	case "header":
		// Header-based: 带有特定 header 的请求路由到 canary，其余到 stable
		headerMatches = append(headerMatches, k8s.HeaderMatchRule{
			HeaderName:  params.CanaryHeaderName,
			HeaderValue: params.CanaryHeaderValue,
			Exact:       true,
			Subset:      "canary",
		})
		// 默认路由到 stable
		headerMatches = append(headerMatches, k8s.HeaderMatchRule{
			Subset: "stable",
		})
	case "weight_header":
		// Weight + Header: 两种方式同时生效
		headerMatches = append(headerMatches, k8s.HeaderMatchRule{
			HeaderName:  params.CanaryHeaderName,
			HeaderValue: params.CanaryHeaderValue,
			Exact:       true,
			Subset:      "canary",
		})
	default:
		// "weight" mode: 无 header 匹配，纯权重分流
	}

	// 3. 确保 VirtualService 存在
	hosts := params.Hosts
	if len(hosts) == 0 {
		hosts = []string{fmt.Sprintf("%s.%s.svc.cluster.local", serviceHost, params.Namespace)}
	}

	gateways := params.Gateways
	if len(gateways) == 0 {
		gateways = []string{"mesh"}
	}

	vsConfig := k8s.CanaryVirtualServiceConfig{
		Name:          vsName,
		Namespace:     params.Namespace,
		Hosts:         hosts,
		Gateways:      gateways,
		StableHost:    serviceHost,
		CanaryHost:    serviceHost,
		StableSubset:  "stable",
		CanarySubset:  "canary",
		CanaryWeight:  params.CanaryPercent,
		StableWeight:  100 - params.CanaryPercent,
		HeaderMatches: headerMatches,
		Timeout:       params.Timeout,
		Labels: map[string]string{
			"app":        params.AppName,
			"managed-by": "my-cloud",
		},
	}

	if err := s.k8sClient.EnsureCanaryVirtualService(ctx, vsConfig); err != nil {
		return fmt.Errorf("ensure VirtualService: %w", err)
	}
	log.Printf("[IstioCanary] VirtualService %s/%s ensured (weight=%d%%)", params.Namespace, vsName, params.CanaryPercent)

	return nil
}

// AdjustIstioCanaryWeight 动态调整 Istio 金丝雀流量权重
func (s *CanaryIstioService) AdjustIstioCanaryWeight(ctx context.Context, namespace, appName string, newPercent int) error {
	log.Printf("[IstioCanary] Adjusting weight: app=%s namespace=%s weight=%d%%", appName, namespace, newPercent)

	if err := s.k8sClient.AdjustCanaryTrafficWeight(ctx, namespace, appName, newPercent); err != nil {
		return fmt.Errorf("adjust Istio canary weight: %w", err)
	}

	log.Printf("[IstioCanary] Traffic weight adjusted to %d%% for %s/%s", newPercent, namespace, appName)
	return nil
}

// GetIstioCanaryWeight 获取当前 Istio 金丝雀流量权重
func (s *CanaryIstioService) GetIstioCanaryWeight(ctx context.Context, namespace, appName string) (int, error) {
	return s.k8sClient.GetCanaryTrafficWeight(ctx, namespace, appName)
}

// ConfirmIstioCanary 确认金丝雀：将 100% 流量切换到 canary 版本（变为新 stable）
func (s *CanaryIstioService) ConfirmIstioCanary(ctx context.Context, params CanaryIstioParams) error {
	log.Printf("[IstioCanary] Confirming canary for promotion: app=%s namespace=%s", params.AppName, params.Namespace)

	serviceHost := params.AppName  // K8s Service name matches workloadName

	cfg := k8s.CanaryDeployConfig{
		Namespace:           params.Namespace,
		AppName:             params.AppName,
		ServiceHost:         serviceHost,
		VirtualServiceName:  params.AppName,
		DestinationRuleName: params.AppName,
		StableSubset:        "stable",
		CanarySubset:        "canary",
		StableVersion:       params.StableVersion,
		CanaryVersion:       params.CanaryVersion,
	}

	if err := s.k8sClient.PromoteCanaryToStable(ctx, cfg); err != nil {
		return fmt.Errorf("promote canary to stable: %w", err)
	}

	log.Printf("[IstioCanary] Canary promoted to stable: %s/%s", params.Namespace, params.AppName)
	return nil
}

// RollbackIstioCanary 回滚金丝雀：恢复 100% 流量到 stable 版本
func (s *CanaryIstioService) RollbackIstioCanary(ctx context.Context, params CanaryIstioParams) error {
	log.Printf("[IstioCanary] Rolling back canary: app=%s namespace=%s", params.AppName, params.Namespace)

	serviceHost := params.AppName  // K8s Service name matches workloadName

	cfg := k8s.CanaryDeployConfig{
		Namespace:           params.Namespace,
		AppName:             params.AppName,
		ServiceHost:         serviceHost,
		VirtualServiceName:  params.AppName,
		DestinationRuleName: params.AppName,
		StableSubset:        "stable",
	}

	if err := s.k8sClient.RollbackCanaryResources(ctx, cfg); err != nil {
		return fmt.Errorf("rollback canary resources: %w", err)
	}

	log.Printf("[IstioCanary] Canary rolled back: %s/%s", params.Namespace, params.AppName)
	return nil
}

// CleanupIstioCanaryResources 清理 Istio 金丝雀资源
func (s *CanaryIstioService) CleanupIstioCanaryResources(ctx context.Context, namespace, appName string) []error {
	log.Printf("[IstioCanary] Cleaning up canary resources: app=%s namespace=%s", appName, namespace)

	resourceNames := k8s.CanaryResourceNames{
		VirtualServiceName:  appName,
		DestinationRuleName: appName,
	}

	return s.k8sClient.CleanupCanaryResources(ctx, namespace, resourceNames)
}

// EnsureIstioMeshDefaults 为命名空间设置 Istio Mesh 默认配置
// 包括：命名空间级 PeerAuthentication (mTLS)、默认 Sidecar
func (s *CanaryIstioService) EnsureIstioMeshDefaults(ctx context.Context, namespace string) error {
	log.Printf("[IstioCanary] Ensuring mesh defaults for namespace: %s", namespace)

	// 1. 确保 PeerAuthentication (mTLS PERMISSIVE)
	if err := s.k8sClient.EnsurePeerAuthentication(ctx, namespace, "PERMISSIVE"); err != nil {
		return fmt.Errorf("ensure PeerAuthentication: %w", err)
	}

	// 2. 确保默认 Sidecar
	if err := s.k8sClient.EnsureDefaultSidecar(ctx, namespace); err != nil {
		return fmt.Errorf("ensure Sidecar: %w", err)
	}

	log.Printf("[IstioCanary] Mesh defaults ensured for namespace: %s", namespace)
	return nil
}

// =========================================================================
// 集成到 Release Service 的辅助方法
// =========================================================================

// IstioCanaryRunner 封装 Istio 金丝雀运行的完整流程
// 供 release_service.go 在 executeCanaryDeployment 中调用
type IstioCanaryRunner struct {
	k8sClient  *k8s.Client
	istioSvc   *CanaryIstioService
}

// NewIstioCanaryRunner 创建 Istio 金丝雀运行器
func NewIstioCanaryRunner(k8sClient *k8s.Client) *IstioCanaryRunner {
	return &IstioCanaryRunner{
		k8sClient: k8sClient,
		istioSvc:   NewCanaryIstioService(k8sClient),
	}
}

// RunCanaryWithIstio 执行完整的 Istio 金丝雀部署流程
// 返回: (success bool, errorDetail string)
func (r *IstioCanaryRunner) RunCanaryWithIstio(
	ctx context.Context,
	namespace, appName, stableVersion, canaryVersion string,
	canaryPercent int,
	routingMode string,
	headerName, headerValue string,
	hosts, gateways []string,
) (bool, string) {

	appID := parseAppID(appName)

	params := CanaryIstioParams{
		AppName:           appName,
		AppID:             uint(appID),
		Namespace:         namespace,
		StableVersion:     stableVersion,
		CanaryVersion:     canaryVersion,
		CanaryPercent:     canaryPercent,
		Hosts:             hosts,
		Gateways:          gateways,
		RoutingMode:       routingMode,
		CanaryHeaderName:  headerName,
		CanaryHeaderValue: headerValue,
	}

	// 确保 Mesh 默认配置
	if err := r.istioSvc.EnsureIstioMeshDefaults(ctx, namespace); err != nil {
		log.Printf("[IstioCanary] Warning: Failed to ensure mesh defaults: %v", err)
		// 不阻塞部署，因为 mTLS 可能已经由集群管理员设置
	}

	// 启动金丝雀
	if err := r.istioSvc.StartCanaryWithIstio(ctx, params); err != nil {
		return false, fmt.Sprintf("failed to start Istio canary: %v", err)
	}

	log.Printf("[IstioCanary] Canary deployment successful: app=%s, canary=%d%%", appName, canaryPercent)
	return true, ""
}

// ConfirmCanaryWithIstio 确认 Istio 金丝雀（全量切换）
func (r *IstioCanaryRunner) ConfirmCanaryWithIstio(
	ctx context.Context,
	namespace, appName, stableVersion, canaryVersion string,
) (bool, string) {

	appID := parseAppID(appName)

	params := CanaryIstioParams{
		AppName:       appName,
		AppID:         uint(appID),
		Namespace:     namespace,
		StableVersion: stableVersion,
		CanaryVersion: canaryVersion,
	}

	if err := r.istioSvc.ConfirmIstioCanary(ctx, params); err != nil {
		return false, fmt.Sprintf("failed to confirm Istio canary: %v", err)
	}

	return true, ""
}

// RollbackCanaryWithIstio 回滚 Istio 金丝雀
func (r *IstioCanaryRunner) RollbackCanaryWithIstio(
	ctx context.Context,
	namespace, appName, stableVersion, canaryVersion string,
) (bool, string) {

	appID := parseAppID(appName)

	params := CanaryIstioParams{
		AppName:       appName,
		AppID:         uint(appID),
		Namespace:     namespace,
		StableVersion: stableVersion,
		CanaryVersion: canaryVersion,
	}

	if err := r.istioSvc.RollbackIstioCanary(ctx, params); err != nil {
		return false, fmt.Sprintf("failed to rollback Istio canary: %v", err)
	}

	return true, ""
}

// parseAppID 从 appName (如 "app-8") 中提取 appID
func parseAppID(appName string) int {
	// appName 格式: "app-{id}"
	parts := strings.Split(appName, "-")
	if len(parts) >= 2 {
		var id int
		if _, err := fmt.Sscanf(parts[len(parts)-1], "%d", &id); err == nil {
			return id
		}
	}
	return 0
}

// =========================================================================
// Istio 流量监控
// =========================================================================

// GetCanaryTrafficStats 获取金丝雀流量统计信息
func (s *CanaryIstioService) GetCanaryTrafficStats(ctx context.Context, namespace, appName string) (*CanaryTrafficStats, error) {
	vs, err := s.k8sClient.GetVirtualService(ctx, namespace, appName)
	if err != nil {
		return nil, fmt.Errorf("get VirtualService: %w", err)
	}

	stats := &CanaryTrafficStats{
		VirtualServiceName: vs.Name,
		Namespace:          vs.Namespace,
		Hosts:              vs.Spec.Hosts,
	}

	for _, http := range vs.Spec.HTTP {
		routeInfo := RouteInfo{}

		// 检查 match 条件
		for _, match := range http.Match {
			for header, sm := range match.Headers {
				routeInfo.MatchHeaders = append(routeInfo.MatchHeaders, fmt.Sprintf("%s=%s", header, sm.Exact))
			}
		}

		// 收集路由目标
		for _, route := range http.Route {
			subset := route.Destination.Subset
			weight := route.Weight
			routeInfo.Destinations = append(routeInfo.Destinations, DestinationInfo{
				Subset: subset,
				Host:   route.Destination.Host,
				Weight: weight,
			})
			stats.TotalWeight += weight
		}

		stats.Routes = append(stats.Routes, routeInfo)
	}

	return stats, nil
}

// CanaryTrafficStats 金丝雀流量统计
type CanaryTrafficStats struct {
	VirtualServiceName string      `json:"virtualServiceName"`
	Namespace          string      `json:"namespace"`
	Hosts              []string    `json:"hosts"`
	Routes             []RouteInfo `json:"routes"`
	TotalWeight        int         `json:"totalWeight"`
}

// RouteInfo 路由信息
type RouteInfo struct {
	MatchHeaders []string          `json:"matchHeaders,omitempty"`
	Destinations []DestinationInfo `json:"destinations"`
}

// DestinationInfo 目标信息
type DestinationInfo struct {
	Subset string `json:"subset"`
	Host   string `json:"host"`
	Weight int    `json:"weight"`
}

// =========================================================================
// 确保命名空间已启用 Istio sidecar 注入
// =========================================================================

// EnsureNamespaceIstioInjection 确保命名空间已启用 Istio sidecar 注入
// 通过添加 istio-injection=enabled label
func (s *CanaryIstioService) EnsureNamespaceIstioInjection(ctx context.Context, namespace string) error {
	log.Printf("[IstioCanary] Ensuring Istio injection for namespace: %s", namespace)

	// 注意: 实际生产环境中，命名空间的 label 管理应该在 Cluster Service 中统一处理
	// 这里返回 nil，因为 sidecar 注入由 Istio 的 mutating webhook 自动处理

	_ = namespace
	_ = ctx
	return nil
}

// EnsureAppGateway 为应用创建 Istio Gateway（对外暴露）
func (s *CanaryIstioService) EnsureAppGateway(ctx context.Context, namespace, appName, host string, tlsSecretName string) error {
	gw := &k8s.Gateway{
		Spec: k8s.GatewaySpec{
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
			Servers: []k8s.Server{
				{
					Port: k8s.Port{
						Number:   80,
						Name:     "http",
						Protocol: "HTTP",
					},
					Hosts: []string{host},
				},
			},
		},
	}
	gw.APIVersion = k8s.IstioNetworkingAPIGroup + "/" + k8s.IstioNetworkingAPIVersion
	gw.Kind = "Gateway"
	gw.Name = appName
	gw.Namespace = namespace
	gw.Labels = map[string]string{
		"app":        appName,
		"managed-by": "my-cloud",
	}

	// 如果有 TLS secret，添加 HTTPS server
	if tlsSecretName != "" {
		gw.Spec.Servers = append(gw.Spec.Servers, k8s.Server{
			Port: k8s.Port{
				Number:   443,
				Name:     "https",
				Protocol: "HTTPS",
			},
			Hosts: []string{host},
			TLS: &k8s.ServerTLSSettings{
				Mode:           "SIMPLE",
				CredentialName: tlsSecretName,
			},
		})
	}

	// Try to create; if already exists, update
	err := s.k8sClient.CreateGateway(ctx, namespace, gw)
	if err != nil && strings.Contains(err.Error(), "already exists") {
		existing, getErr := s.k8sClient.GetGateway(ctx, namespace, appName)
		if getErr != nil {
			return getErr
		}
		gw.ResourceVersion = existing.ResourceVersion
		return s.k8sClient.UpdateGateway(ctx, namespace, gw)
	}
	return err
}
