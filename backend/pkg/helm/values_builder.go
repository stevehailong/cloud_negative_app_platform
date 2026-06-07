package helm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValuesBuilder 构建Helm Values
type ValuesBuilder struct {
	values map[string]interface{}
}

// NewValuesBuilder 创建Values构建器
func NewValuesBuilder() *ValuesBuilder {
	return &ValuesBuilder{
		values: make(map[string]interface{}),
	}
}

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	// 应用基础信息
	AppName      string
	Image        string
	Replicas     int
	WorkloadName string

	// 资源配置
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string

	// 服务配置
	ServiceType   string
	ServicePort   int
	ContainerPort int

	// Ingress配置
	IngressEnabled    bool
	IngressHost       string
	IngressPath       string
	IngressTLSEnabled bool
	TLSSecretName     string

	// 健康检查
	LivenessPath  string
	ReadinessPath string

	// 环境变量
	EnvVars map[string]string

	// 链路追踪
	TracingEnabled     bool
	TracingEndpoint    string
	TracingServiceName string

	// ConfigMap
	ConfigMapEnabled bool
	ConfigMapData    map[string]string

	// Secret
	SecretEnabled bool
	SecretData    map[string]string

	// 自动扩缩容
	HPAEnabled     bool
	HPAMinReplicas int
	HPAMaxReplicas int
	HPATargetCPU   int

	// Istio 服务网格配置
	IstioEnabled             bool
	IstioVirtualServiceHosts []string
	IstioGateways            []string
	IstioStableSubset        string
	IstioCanarySubset        string
	IstioCanaryWeight        int
	IstioStableWeight        int
	IstioHeaderMatches       []IstioHeaderMatch
	IstioTimeout             string
	IstioRetryAttempts       int
	IstioRetryPerTryTimeout  string
	IstioDRHost              string
	IstioDRSubsets           []IstioSubsetConfig
	IstioTrafficPolicy       map[string]interface{}
	IstioGatewayEnabled      bool
	IstioGatewayServers      []IstioGatewayServer
	IstioPeerAuthEnabled     bool
	IstioPeerAuthMode        string
}

// IstioHeaderMatch Istio VirtualService header 匹配规则
type IstioHeaderMatch struct {
	HeaderName  string
	HeaderValue string
	Subset      string
}

// IstioSubsetConfig DestinationRule 子集配置
type IstioSubsetConfig struct {
	Name   string
	Labels map[string]string
}

// IstioGatewayServer Gateway 服务器配置
type IstioGatewayServer struct {
	PortNumber   int
	PortName     string
	PortProtocol string
	Hosts        []string
	TLSSecret    string
}

// BuildFromTemplate 从环境模板构建Values
func (b *ValuesBuilder) BuildFromTemplate(templateValues string, config DeploymentConfig) (map[string]interface{}, error) {
	// 先解析模板的默认值
	if templateValues != "" {
		if err := json.Unmarshal([]byte(templateValues), &b.values); err != nil {
			// 尝试作为YAML解析
			// 这里简化处理，实际应该使用yaml.Unmarshal
			b.values = make(map[string]interface{})
		}
	}

	// 覆盖配置
	b.SetBasicConfig(config)
	b.SetResourceConfig(config)
	b.SetServiceConfig(config)
	b.SetIngressConfig(config)
	b.SetHealthCheckConfig(config)
	b.SetEnvConfig(config)
	b.SetTracingConfig(config)
	b.SetConfigMapConfig(config)
	b.SetSecretConfig(config)
	b.SetHPAConfig(config)
	b.SetIstioConfig(config)

	return b.values, nil
}

// SetBasicConfig sets basic config, only overriding template if config has explicit values
func (b *ValuesBuilder) SetBasicConfig(config DeploymentConfig) {
	// replicaCount: only override if template didn't set it (or config has explicit >1)
	if _, exists := b.values["replicaCount"]; !exists || config.Replicas > 1 {
		b.values["replicaCount"] = config.Replicas
	}

	if config.Image != "" {
		// 解析镜像地址，兼容 registry:port/repo:tag
		repository := config.Image
		tag := "latest"
		lastSlash := strings.LastIndex(config.Image, "/")
		lastColon := strings.LastIndex(config.Image, ":")
		if lastColon > lastSlash {
			repository = config.Image[:lastColon]
			tag = config.Image[lastColon+1:]
		}

		image := make(map[string]interface{})
		image["repository"] = repository
		image["tag"] = tag
		image["pullPolicy"] = "IfNotPresent"
		b.values["image"] = image
	}
}

// SetResourceConfig sets resources — only as fallback if template didn't define them
func (b *ValuesBuilder) SetResourceConfig(config DeploymentConfig) {
	if _, exists := b.values["resources"]; exists {
		return // template already defined resources, don't override
	}
	resources := make(map[string]interface{})

	limits := make(map[string]interface{})
	if config.CPULimit != "" {
		limits["cpu"] = config.CPULimit
	} else {
		limits["cpu"] = "1000m"
	}
	if config.MemoryLimit != "" {
		limits["memory"] = config.MemoryLimit
	} else {
		limits["memory"] = "1Gi"
	}
	resources["limits"] = limits

	requests := make(map[string]interface{})
	if config.CPURequest != "" {
		requests["cpu"] = config.CPURequest
	} else {
		requests["cpu"] = "500m"
	}
	if config.MemoryRequest != "" {
		requests["memory"] = config.MemoryRequest
	} else {
		requests["memory"] = "512Mi"
	}
	resources["requests"] = requests

	b.values["resources"] = resources
}

// SetServiceConfig sets service config — only override what template didn't define
func (b *ValuesBuilder) SetServiceConfig(config DeploymentConfig) {
	svc, _ := b.values["service"].(map[string]interface{})
	if svc == nil {
		svc = make(map[string]interface{})
		b.values["service"] = svc
	}

	if _, exists := svc["type"]; !exists {
		if config.ServiceType != "" {
			svc["type"] = config.ServiceType
		} else {
			svc["type"] = "ClusterIP"
		}
	}

	if _, exists := svc["port"]; !exists {
		if config.ServicePort > 0 {
			svc["port"] = config.ServicePort
		} else {
			svc["port"] = 80
		}
	}

	if _, exists := svc["targetPort"]; !exists {
		if config.ContainerPort > 0 {
			svc["targetPort"] = config.ContainerPort
		} else {
			svc["targetPort"] = 8080
		}
	}

	if _, exists := b.values["containerPort"]; !exists {
		b.values["containerPort"] = svc["targetPort"]
	}
	b.values["service"] = svc
}

// SetIngressConfig 设置Ingress配置
func (b *ValuesBuilder) SetIngressConfig(config DeploymentConfig) {
	ingress := make(map[string]interface{})
	ingress["enabled"] = config.IngressEnabled

	if config.IngressEnabled {
		ingress["className"] = "nginx"

		// 注解
		annotations := make(map[string]interface{})
		annotations["nginx.ingress.kubernetes.io/ssl-redirect"] = "false"
		if config.IngressTLSEnabled {
			annotations["cert-manager.io/cluster-issuer"] = "letsencrypt-prod"
			annotations["nginx.ingress.kubernetes.io/ssl-redirect"] = "true"
		}
		ingress["annotations"] = annotations

		// 主机配置
		host := make(map[string]interface{})
		host["host"] = config.IngressHost
		if config.IngressHost == "" {
			host["host"] = fmt.Sprintf("%s.example.com", config.AppName)
		}

		path := make(map[string]interface{})
		if config.IngressPath != "" {
			path["path"] = config.IngressPath
		} else {
			path["path"] = "/"
		}
		path["pathType"] = "Prefix"
		host["paths"] = []interface{}{path}

		ingress["hosts"] = []interface{}{host}

		// TLS配置
		if config.IngressTLSEnabled {
			tls := make(map[string]interface{})
			tls["secretName"] = config.TLSSecretName
			if config.TLSSecretName == "" {
				tls["secretName"] = fmt.Sprintf("%s-tls", config.AppName)
			}
			tls["hosts"] = []string{host["host"].(string)}
			ingress["tls"] = []interface{}{tls}
		}
	}

	b.values["ingress"] = ingress
}

// SetHealthCheckConfig sets health checks as fallback only if template didn't define them
func (b *ValuesBuilder) SetHealthCheckConfig(config DeploymentConfig) {
	_, hasLiveness := b.values["livenessProbe"]
	_, hasReadiness := b.values["readinessProbe"]

	if !hasLiveness {
		liveness := make(map[string]interface{})
		liveness["enabled"] = true
		httpGet := make(map[string]interface{})
		if config.LivenessPath != "" {
			httpGet["path"] = config.LivenessPath
		} else {
			httpGet["path"] = "/health"
		}
		httpGet["port"] = config.ContainerPort
		if httpGet["port"] == 0 {
			httpGet["port"] = 8080
		}
		liveness["httpGet"] = httpGet
		liveness["initialDelaySeconds"] = 30
		liveness["periodSeconds"] = 10
		liveness["timeoutSeconds"] = 5
		liveness["failureThreshold"] = 3
		b.values["livenessProbe"] = liveness
	}

	if !hasReadiness {
		readiness := make(map[string]interface{})
		readiness["enabled"] = true
		httpGetReady := make(map[string]interface{})
		if config.ReadinessPath != "" {
			httpGetReady["path"] = config.ReadinessPath
		} else {
			httpGetReady["path"] = "/ready"
		}
		httpGetReady["port"] = config.ContainerPort
		if httpGetReady["port"] == 0 {
			httpGetReady["port"] = 8080
		}
		readiness["httpGet"] = httpGetReady
		readiness["initialDelaySeconds"] = 10
		readiness["periodSeconds"] = 5
		readiness["timeoutSeconds"] = 3
		readiness["failureThreshold"] = 3
		b.values["readinessProbe"] = readiness
	}
}

// SetEnvConfig 设置环境变量配置
func (b *ValuesBuilder) SetEnvConfig(config DeploymentConfig) {
	// 先保留模板中已有的 env（避免 BuildFromTemplate 流程中模板 env 被覆盖）
	existingEnv := make(map[string]string)
	existingIdx := make(map[string]int)
	if existingList, ok := b.values["env"].([]interface{}); ok {
		for i, item := range existingList {
			if m, ok := item.(map[string]interface{}); ok {
				if name, ok := m["name"].(string); ok {
					if val, ok := m["value"].(string); ok {
						existingEnv[name] = val
						existingIdx[name] = i
					}
				}
			}
		}
	}

	// 基础环境变量：config.EnvVars 覆盖模板值
	if len(config.EnvVars) > 0 {
		for name, value := range config.EnvVars {
			existingEnv[name] = value
		}
	}

	// 如果没有模板 env 也没有 config env，给默认值
	if len(existingEnv) == 0 {
		existingEnv["APP_ENV"] = "production"
		existingEnv["LOG_LEVEL"] = "info"
	}

	// 链路追踪环境变量（自动注入）
	if config.TracingEnabled {
		endpoint := config.TracingEndpoint
		if endpoint == "" {
			endpoint = "http://monitor-service:8090/internal/v1/traces/spans"
		}
		serviceName := config.TracingServiceName
		if serviceName == "" {
			serviceName = config.WorkloadName
		}
		existingEnv["TRACE_ENABLED"] = "true"
		existingEnv["TRACE_ENDPOINT"] = endpoint
		existingEnv["TRACE_SERVICE_NAME"] = serviceName
	}

	// 重建 env 列表（保持顺序：模板原有 + 新增）
	seen := make(map[string]bool)
	env := make([]interface{}, 0)

	// 先放模板中原有的（保留顺序）
	if existingList, ok := b.values["env"].([]interface{}); ok {
		for _, item := range existingList {
			if m, ok := item.(map[string]interface{}); ok {
				if name, ok := m["name"].(string); ok {
					if val, exists := existingEnv[name]; exists {
						env = append(env, map[string]interface{}{
							"name":  name,
							"value": val,
						})
						seen[name] = true
					}
				}
			}
		}
	}

	// 再放 config.EnvVars 中新增的（不在模板中的）
	for name, value := range existingEnv {
		if !seen[name] {
			env = append(env, map[string]interface{}{
				"name":  name,
				"value": value,
			})
		}
	}

	b.values["env"] = env
}

// SetConfigMapConfig 设置ConfigMap配置
func (b *ValuesBuilder) SetConfigMapConfig(config DeploymentConfig) {
	configMap := make(map[string]interface{})
	configMap["enabled"] = config.ConfigMapEnabled

	if config.ConfigMapEnabled && len(config.ConfigMapData) > 0 {
		configMap["data"] = config.ConfigMapData
	}

	b.values["configMap"] = configMap
}

// SetSecretConfig 设置Secret配置
func (b *ValuesBuilder) SetSecretConfig(config DeploymentConfig) {
	secret := make(map[string]interface{})
	secret["enabled"] = config.SecretEnabled

	if config.SecretEnabled && len(config.SecretData) > 0 {
		secret["data"] = config.SecretData
	}

	b.values["secret"] = secret
}

// SetTracingConfig 设置链路追踪配置
func (b *ValuesBuilder) SetTracingConfig(config DeploymentConfig) {
	tracing := make(map[string]interface{})
	tracing["enabled"] = config.TracingEnabled

	if config.TracingEnabled {
		endpoint := config.TracingEndpoint
		if endpoint == "" {
			endpoint = "http://monitor-service:8090/internal/v1/traces/spans"
		}
		serviceName := config.TracingServiceName
		if serviceName == "" {
			serviceName = config.WorkloadName
		}
		tracing["endpoint"] = endpoint
		tracing["serviceName"] = serviceName
	}

	b.values["tracing"] = tracing
}

// SetHPAConfig sets HPA config, preserving template values
func (b *ValuesBuilder) SetHPAConfig(config DeploymentConfig) {
	if _, exists := b.values["autoscaling"]; exists {
		return // template already has HPA config
	}
	autoscaling := make(map[string]interface{})
	autoscaling["enabled"] = config.HPAEnabled

	if config.HPAEnabled {
		if config.HPAMinReplicas > 0 {
			autoscaling["minReplicas"] = config.HPAMinReplicas
		} else {
			autoscaling["minReplicas"] = 2
		}

		if config.HPAMaxReplicas > 0 {
			autoscaling["maxReplicas"] = config.HPAMaxReplicas
		} else {
			autoscaling["maxReplicas"] = 10
		}

		if config.HPATargetCPU > 0 {
			autoscaling["targetCPUUtilizationPercentage"] = config.HPATargetCPU
		} else {
			autoscaling["targetCPUUtilizationPercentage"] = 80
		}
	}

	b.values["autoscaling"] = autoscaling
}

// SetServiceAccount 设置ServiceAccount
func (b *ValuesBuilder) SetServiceAccount(create bool, name string) {
	sa := make(map[string]interface{})
	sa["create"] = create
	if name != "" {
		sa["name"] = name
	}
	b.values["serviceAccount"] = sa
}

// Build 构建最终的Values
func (b *ValuesBuilder) Build() map[string]interface{} {
	return b.values
}

// SetIstioConfig 设置 Istio 服务网格配置
func (b *ValuesBuilder) SetIstioConfig(config DeploymentConfig) {
	istio := make(map[string]interface{})
	istio["enabled"] = config.IstioEnabled

	if !config.IstioEnabled {
		b.values["istio"] = istio
		return
	}

	// VirtualService 配置
	vs := make(map[string]interface{})

	if len(config.IstioVirtualServiceHosts) > 0 {
		vs["hosts"] = config.IstioVirtualServiceHosts
	}

	if len(config.IstioGateways) > 0 {
		vs["gateways"] = config.IstioGateways
	} else {
		vs["gateways"] = []string{"mesh"}
	}

	if config.IstioStableSubset != "" {
		vs["stableSubset"] = config.IstioStableSubset
	} else {
		vs["stableSubset"] = "stable"
	}

	if config.IstioCanarySubset != "" {
		vs["canarySubset"] = config.IstioCanarySubset
	} else {
		vs["canarySubset"] = "canary"
	}

	if config.IstioCanaryWeight > 0 {
		vs["canaryWeight"] = config.IstioCanaryWeight
		vs["stableWeight"] = 100 - config.IstioCanaryWeight
	} else {
		vs["canaryWeight"] = 0
		vs["stableWeight"] = 100
	}

	if config.IstioStableWeight > 0 {
		vs["stableWeight"] = config.IstioStableWeight
	}

	if config.IstioTimeout != "" {
		vs["timeout"] = config.IstioTimeout
	}

	if config.IstioRetryAttempts > 0 {
		retries := make(map[string]interface{})
		retries["attempts"] = config.IstioRetryAttempts
		if config.IstioRetryPerTryTimeout != "" {
			retries["perTryTimeout"] = config.IstioRetryPerTryTimeout
		} else {
			retries["perTryTimeout"] = "2s"
		}
		retries["retryOn"] = "connect-failure,refused-stream"
		vs["retries"] = retries
	}

	// Header 匹配规则
	if len(config.IstioHeaderMatches) > 0 {
		matches := make([]interface{}, 0, len(config.IstioHeaderMatches))
		for _, hm := range config.IstioHeaderMatches {
			match := map[string]interface{}{
				"headerName":  hm.HeaderName,
				"headerValue": hm.HeaderValue,
				"subset":      hm.Subset,
			}
			matches = append(matches, match)
		}
		vs["headerMatches"] = matches
	}

	vs["labels"] = map[string]string{
		"managed-by": "my-cloud",
	}

	istio["virtualService"] = vs

	// DestinationRule 配置
	dr := make(map[string]interface{})

	if config.IstioDRHost != "" {
		dr["host"] = config.IstioDRHost
	}

	if len(config.IstioDRSubsets) > 0 {
		subsets := make([]interface{}, 0, len(config.IstioDRSubsets))
		for _, s := range config.IstioDRSubsets {
			subset := map[string]interface{}{
				"name":   s.Name,
				"labels": s.Labels,
			}
			subsets = append(subsets, subset)
		}
		dr["subsets"] = subsets
	} else {
		dr["subsets"] = []interface{}{
			map[string]interface{}{
				"name": "stable",
				"labels": map[string]string{
					"version": "stable",
				},
			},
			map[string]interface{}{
				"name": "canary",
				"labels": map[string]string{
					"version": "canary",
				},
			},
		}
	}

	if config.IstioTrafficPolicy != nil {
		dr["trafficPolicy"] = config.IstioTrafficPolicy
	}

	dr["exportTo"] = []string{"."}
	dr["labels"] = map[string]string{
		"managed-by": "my-cloud",
	}

	istio["destinationRule"] = dr

	// Gateway 配置
	gw := make(map[string]interface{})
	gw["enabled"] = config.IstioGatewayEnabled

	if config.IstioGatewayEnabled {
		gw["selector"] = map[string]string{
			"istio": "ingressgateway",
		}

		if len(config.IstioGatewayServers) > 0 {
			servers := make([]interface{}, 0, len(config.IstioGatewayServers))
			for _, s := range config.IstioGatewayServers {
				server := map[string]interface{}{
					"port": map[string]interface{}{
						"number":   s.PortNumber,
						"name":     s.PortName,
						"protocol": s.PortProtocol,
					},
					"hosts": s.Hosts,
				}
				if s.TLSSecret != "" {
					server["tls"] = map[string]interface{}{
						"mode":           "SIMPLE",
						"credentialName": s.TLSSecret,
					}
				}
				servers = append(servers, server)
			}
			gw["servers"] = servers
		}

		gw["labels"] = map[string]string{
			"managed-by": "my-cloud",
		}
	}

	istio["gateway"] = gw

	// PeerAuthentication 配置
	pa := make(map[string]interface{})
	pa["enabled"] = config.IstioPeerAuthEnabled

	if config.IstioPeerAuthEnabled {
		pa["name"] = "default"
		if config.IstioPeerAuthMode != "" {
			pa["mode"] = config.IstioPeerAuthMode
		} else {
			pa["mode"] = "PERMISSIVE"
		}
	}

	istio["peerAuthentication"] = pa

	b.values["istio"] = istio
}
