package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// =========================================================================
// Istio API GroupVersionResource 常量
// =========================================================================

const (
	IstioNetworkingAPIGroup   = "networking.istio.io"
	IstioNetworkingAPIVersion = "v1beta1"
	IstioSecurityAPIGroup     = "security.istio.io"
	IstioSecurityAPIVersion   = "v1beta1"
)

var (
	// VirtualService GVR: networking.istio.io/v1beta1/virtualservices
	VirtualServiceGVR = schema.GroupVersionResource{
		Group:    IstioNetworkingAPIGroup,
		Version:  IstioNetworkingAPIVersion,
		Resource: "virtualservices",
	}

	// DestinationRule GVR: networking.istio.io/v1beta1/destinationrules
	DestinationRuleGVR = schema.GroupVersionResource{
		Group:    IstioNetworkingAPIGroup,
		Version:  IstioNetworkingAPIVersion,
		Resource: "destinationrules",
	}

	// Gateway GVR: networking.istio.io/v1beta1/gateways
	GatewayGVR = schema.GroupVersionResource{
		Group:    IstioNetworkingAPIGroup,
		Version:  IstioNetworkingAPIVersion,
		Resource: "gateways",
	}

	// Sidecar GVR: networking.istio.io/v1beta1/sidecars
	SidecarGVR = schema.GroupVersionResource{
		Group:    IstioNetworkingAPIGroup,
		Version:  IstioNetworkingAPIVersion,
		Resource: "sidecars",
	}

	// ServiceEntry GVR: networking.istio.io/v1beta1/serviceentries
	ServiceEntryGVR = schema.GroupVersionResource{
		Group:    IstioNetworkingAPIGroup,
		Version:  IstioNetworkingAPIVersion,
		Resource: "serviceentries",
	}

	// PeerAuthentication GVR: security.istio.io/v1beta1/peerauthentications
	PeerAuthenticationGVR = schema.GroupVersionResource{
		Group:    IstioSecurityAPIGroup,
		Version:  IstioSecurityAPIVersion,
		Resource: "peerauthentications",
	}

	// AuthorizationPolicy GVR: security.istio.io/v1beta1/authorizationpolicies
	AuthorizationPolicyGVR = schema.GroupVersionResource{
		Group:    IstioSecurityAPIGroup,
		Version:  IstioSecurityAPIVersion,
		Resource: "authorizationpolicies",
	}

	// RequestAuthentication GVR: security.istio.io/v1beta1/requestauthentications
	RequestAuthenticationGVR = schema.GroupVersionResource{
		Group:    IstioSecurityAPIGroup,
		Version:  IstioSecurityAPIVersion,
		Resource: "requestauthentications",
	}
)

// =========================================================================
// Istio CRD 类型定义 (强类型 Go 结构体)
// =========================================================================

// ---------- VirtualService ----------

// VirtualService 定义 HTTP/GRPC 流量路由规则
type VirtualService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VirtualServiceSpec   `json:"spec"`
	Status            VirtualServiceStatus `json:"status,omitempty"`
}

type VirtualServiceSpec struct {
	Hosts    []string    `json:"hosts,omitempty"`
	Gateways []string    `json:"gateways,omitempty"`
	HTTP     []HTTPRoute `json:"http,omitempty"`
	TLS      []TLSRoute  `json:"tls,omitempty"`
	TCP      []TCPRoute  `json:"tcp,omitempty"`
	ExportTo []string    `json:"exportTo,omitempty"`
}

type VirtualServiceStatus struct {
	Conditions []IstioCondition `json:"conditions,omitempty"`
}

// HTTPRoute HTTP 路由规则
type HTTPRoute struct {
	Name            string           `json:"name,omitempty"`
	Match           []HTTPMatch      `json:"match,omitempty"`
	Route           []RouteDestination `json:"route,omitempty"`
	Redirect        *HTTPRedirect    `json:"redirect,omitempty"`
	Rewrite         *HTTPRewrite     `json:"rewrite,omitempty"`
	Timeout         string           `json:"timeout,omitempty"`
	Retries         *HTTPRetries     `json:"retries,omitempty"`
	Fault           *HTTPFault       `json:"fault,omitempty"`
	Mirror          *Destination     `json:"mirror,omitempty"`
	MirrorPercent   *int             `json:"mirrorPercentage,omitempty"`
	CorsPolicy      *CorsPolicy      `json:"corsPolicy,omitempty"`
	Headers         *Headers         `json:"headers,omitempty"`
	Delegate        *Delegate        `json:"delegate,omitempty"`
}

// HTTPMatch HTTP 匹配条件
type HTTPMatch struct {
	Name        string              `json:"name,omitempty"`
	URI         *StringMatch        `json:"uri,omitempty"`
	Scheme      *StringMatch        `json:"scheme,omitempty"`
	Method      *StringMatch        `json:"method,omitempty"`
	Authority   *StringMatch        `json:"authority,omitempty"`
	Headers     map[string]StringMatch `json:"headers,omitempty"`
	Port        int                 `json:"port,omitempty"`
	SourceLabels map[string]string  `json:"sourceLabels,omitempty"`
	Gateways    []string            `json:"gateways,omitempty"`
	QueryParams map[string]StringMatch `json:"queryParams,omitempty"`
	IgnoreURICase bool              `json:"ignoreUriCase,omitempty"`
	WithoutHeaders map[string]StringMatch `json:"withoutHeaders,omitempty"`
	SourceNamespace string          `json:"sourceNamespace,omitempty"`
}

// StringMatch 字符串匹配（精确/前缀/正则）
type StringMatch struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regex  string `json:"regex,omitempty"`
}

// RouteDestination 路由目标（含权重）
type RouteDestination struct {
	Destination Destination `json:"destination"`
	Weight      int         `json:"weight,omitempty"`
	Headers     *Headers    `json:"headers,omitempty"`
}

// Destination 目标服务
type Destination struct {
	Host   string `json:"host"`
	Subset string `json:"subset,omitempty"`
	Port   *PortSelector `json:"port,omitempty"`
}

// PortSelector 端口选择器
type PortSelector struct {
	Number int `json:"number"`
}

// HTTPRedirect HTTP 重定向
type HTTPRedirect struct {
	URI          string `json:"uri,omitempty"`
	Authority    string `json:"authority,omitempty"`
	RedirectCode int    `json:"redirectCode,omitempty"`
}

// HTTPRewrite HTTP 重写
type HTTPRewrite struct {
	URI       string `json:"uri,omitempty"`
	Authority string `json:"authority,omitempty"`
}

// HTTPRetries HTTP 重试策略
type HTTPRetries struct {
	Attempts      int    `json:"attempts"`
	PerTryTimeout string `json:"perTryTimeout,omitempty"`
	RetryOn       string `json:"retryOn,omitempty"`
}

// HTTPFault HTTP 故障注入
type HTTPFault struct {
	Delay *Delay   `json:"delay,omitempty"`
	Abort *Abort   `json:"abort,omitempty"`
}

// Delay 延迟故障
type Delay struct {
	Percent    int    `json:"percent,omitempty"`
	FixedDelay string `json:"fixedDelay,omitempty"`
}

// Abort 中止故障
type Abort struct {
	Percent    int    `json:"percent,omitempty"`
	HTTPStatus int    `json:"httpStatus,omitempty"`
}

// CorsPolicy CORS 策略
type CorsPolicy struct {
	AllowOrigins  []StringMatch `json:"allowOrigins,omitempty"`
	AllowMethods  []string      `json:"allowMethods,omitempty"`
	AllowHeaders  []string      `json:"allowHeaders,omitempty"`
	ExposeHeaders []string      `json:"exposeHeaders,omitempty"`
	MaxAge        string        `json:"maxAge,omitempty"`
	AllowCredentials bool       `json:"allowCredentials,omitempty"`
}

// Headers HTTP 头操作
type Headers struct {
	Request  *HeaderOperations `json:"request,omitempty"`
	Response *HeaderOperations `json:"response,omitempty"`
}

// HeaderOperations 头操作（增删改）
type HeaderOperations struct {
	Set    map[string]string `json:"set,omitempty"`
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}

// Delegate 路由委托
type Delegate struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// TLSRoute TLS 路由
type TLSRoute struct {
	Match []TLSMatch       `json:"match"`
	Route []RouteDestination `json:"route,omitempty"`
}

// TLSMatch TLS 匹配条件
type TLSMatch struct {
	Port         int               `json:"port"`
	SNIHosts     []string          `json:"sniHosts,omitempty"`
	Gateways     []string          `json:"gateways,omitempty"`
	SourceLabels map[string]string `json:"sourceLabels,omitempty"`
}

// TCPRoute TCP 路由
type TCPRoute struct {
	Match []L4Match         `json:"match,omitempty"`
	Route []RouteDestination `json:"route,omitempty"`
}

// L4Match L4 匹配条件
type L4Match struct {
	Port         int               `json:"port"`
	SourceLabels map[string]string `json:"sourceLabels,omitempty"`
	Gateways     []string          `json:"gateways,omitempty"`
	SourceSubnet string            `json:"sourceSubnet,omitempty"`
}

// ---------- DestinationRule ----------

// DestinationRule 定义服务子集和流量策略
type DestinationRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DestinationRuleSpec   `json:"spec"`
	Status            DestinationRuleStatus `json:"status,omitempty"`
}

type DestinationRuleSpec struct {
	Host          string          `json:"host"`
	TrafficPolicy *TrafficPolicy  `json:"trafficPolicy,omitempty"`
	Subsets       []Subset        `json:"subsets,omitempty"`
	ExportTo      []string        `json:"exportTo,omitempty"`
	WorkloadSelector *WorkloadSelector `json:"workloadSelector,omitempty"`
}

type DestinationRuleStatus struct {
	Conditions []IstioCondition `json:"conditions,omitempty"`
}

// Subset 服务子集
type Subset struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	TrafficPolicy *TrafficPolicy `json:"trafficPolicy,omitempty"`
}

// TrafficPolicy 流量策略
type TrafficPolicy struct {
	LoadBalancer  *LoadBalancer  `json:"loadBalancer,omitempty"`
	ConnectionPool *ConnectionPool `json:"connectionPool,omitempty"`
	OutlierDetection *OutlierDetection `json:"outlierDetection,omitempty"`
	TLS           *ClientTLSSettings `json:"tls,omitempty"`
	PortLevelSettings []PortTrafficPolicy `json:"portLevelSettings,omitempty"`
}

// LoadBalancer 负载均衡
type LoadBalancer struct {
	Simple         string              `json:"simple,omitempty"`
	ConsistentHash *ConsistentHash     `json:"consistentHash,omitempty"`
	LocalityLB     *LocalityLoadBalancer `json:"localityLbSetting,omitempty"`
}

// ConsistentHash 一致性哈希
type ConsistentHash struct {
	HTTPHeaderName  string `json:"httpHeaderName,omitempty"`
	HTTPCookie      *HTTPCookie `json:"httpCookie,omitempty"`
	UseSourceIP     bool   `json:"useSourceIp,omitempty"`
	HTTPQueryParam  string `json:"httpQueryParameterName,omitempty"`
}

// HTTPCookie 基于 Cookie 的哈希
type HTTPCookie struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
	TTL  string `json:"ttl,omitempty"`
}

// LocalityLoadBalancer 地域负载均衡
type LocalityLoadBalancer struct {
	Distribute []LocalityDistribute `json:"distribute,omitempty"`
	Failover   []LocalityFailover   `json:"failover,omitempty"`
}

type LocalityDistribute struct {
	From string            `json:"from"`
	To   map[string]string `json:"to"`
}

type LocalityFailover struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ConnectionPool 连接池
type ConnectionPool struct {
	TCP  *TCPConnectionPool  `json:"tcp,omitempty"`
	HTTP *HTTPConnectionPool `json:"http,omitempty"`
}

type TCPConnectionPool struct {
	MaxConnections int `json:"maxConnections,omitempty"`
	ConnectTimeout string `json:"connectTimeout,omitempty"`
	TcpKeepalive  *TCPKeepalive `json:"tcpKeepalive,omitempty"`
}

type TCPKeepalive struct {
	Time    int `json:"time,omitempty"`
	Probes  int `json:"probes,omitempty"`
	Interval int `json:"interval,omitempty"`
}

type HTTPConnectionPool struct {
	HTTP1MaxPendingRequests      int `json:"http1MaxPendingRequests,omitempty"`
	HTTP2MaxRequests            int `json:"http2MaxRequests,omitempty"`
	MaxRequestsPerConnection    int `json:"maxRequestsPerConnection,omitempty"`
	MaxRetries                  int `json:"maxRetries,omitempty"`
	IdleTimeout                 string `json:"idleTimeout,omitempty"`
	H2UpgradePolicy             string `json:"h2UpgradePolicy,omitempty"`
}

// OutlierDetection 异常检测
type OutlierDetection struct {
	ConsecutiveErrors      int    `json:"consecutiveErrors,omitempty"`
	Interval               string `json:"interval,omitempty"`
	BaseEjectionTime       string `json:"baseEjectionTime,omitempty"`
	MaxEjectionPercent     int    `json:"maxEjectionPercent,omitempty"`
	MinHealthPercent       int    `json:"minHealthPercent,omitempty"`
	ConsecutiveGatewayErrors int  `json:"consecutiveGatewayErrors,omitempty"`
}

// ClientTLSSettings 客户端 TLS 设置
type ClientTLSSettings struct {
	Mode              string   `json:"mode"`
	ClientCertificate string   `json:"clientCertificate,omitempty"`
	PrivateKey        string   `json:"privateKey,omitempty"`
	CACertificates    string   `json:"caCertificates,omitempty"`
	SubjectAltNames   []string `json:"subjectAltNames,omitempty"`
	Sni               string   `json:"sni,omitempty"`
}

// PortTrafficPolicy 端口级流量策略
type PortTrafficPolicy struct {
	Port            *PortSelector  `json:"port"`
	LoadBalancer    *LoadBalancer  `json:"loadBalancer,omitempty"`
	ConnectionPool  *ConnectionPool `json:"connectionPool,omitempty"`
	OutlierDetection *OutlierDetection `json:"outlierDetection,omitempty"`
	TLS             *ClientTLSSettings `json:"tls,omitempty"`
}

// WorkloadSelector 工作负载选择器
type WorkloadSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// ---------- Gateway ----------

// Gateway 入口/出口网关
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GatewaySpec `json:"spec"`
}

type GatewaySpec struct {
	Servers  []Server         `json:"servers,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
}

// Server 网关服务器
type Server struct {
	Port  Port           `json:"port"`
	Hosts []string       `json:"hosts"`
	TLS   *ServerTLSSettings `json:"tls,omitempty"`
}

// Port 端口定义
type Port struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
	Name     string `json:"name"`
}

// ServerTLSSettings 服务器端 TLS
type ServerTLSSettings struct {
	HTTPSRedirect  bool     `json:"httpsRedirect,omitempty"`
	Mode           string   `json:"mode,omitempty"`
	ServerCertificate string `json:"serverCertificate,omitempty"`
	PrivateKey        string `json:"privateKey,omitempty"`
	CACertificates    string `json:"caCertificates,omitempty"`
	SubjectAltNames   []string `json:"subjectAltNames,omitempty"`
	CredentialName   string   `json:"credentialName,omitempty"`
	MinProtocolVersion string `json:"minProtocolVersion,omitempty"`
	MaxProtocolVersion string `json:"maxProtocolVersion,omitempty"`
}

// ---------- PeerAuthentication ----------

// PeerAuthentication 对等认证（mTLS）
type PeerAuthentication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PeerAuthenticationSpec `json:"spec"`
}

type PeerAuthenticationSpec struct {
	Selector *WorkloadSelector `json:"selector,omitempty"`
	MTLS     *MutualTLS        `json:"mtls,omitempty"`
	PortLevelMTLS []PortMutualTLS `json:"portLevelMtls,omitempty"`
}

// MutualTLS mTLS 模式
type MutualTLS struct {
	Mode string `json:"mode"` // UNSET, DISABLE, PERMISSIVE, STRICT
}

type PortMutualTLS struct {
	Port *PortSelector `json:"port"`
	MTLS *MutualTLS     `json:"mtls,omitempty"`
}

// ---------- AuthorizationPolicy ----------

// AuthorizationPolicy 授权策略
type AuthorizationPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AuthorizationPolicySpec `json:"spec"`
}

type AuthorizationPolicySpec struct {
	Selector *WorkloadSelector `json:"selector,omitempty"`
	Rules    []Rule            `json:"rules,omitempty"`
	Action   string            `json:"action,omitempty"` // ALLOW or DENY
}

// Rule 授权规则
type Rule struct {
	From []RuleFrom `json:"from,omitempty"`
	To   []RuleTo   `json:"to,omitempty"`
	When []Condition `json:"when,omitempty"`
}

type RuleFrom struct {
	Source *Source `json:"source,omitempty"`
}

type RuleTo struct {
	Operation *Operation `json:"operation,omitempty"`
}

type Source struct {
	Principals       []string          `json:"principals,omitempty"`
	NotPrincipals    []string          `json:"notPrincipals,omitempty"`
	RequestPrincipals []string         `json:"requestPrincipals,omitempty"`
	Namespaces       []string          `json:"namespaces,omitempty"`
	NotNamespaces    []string          `json:"notNamespaces,omitempty"`
	IPBlocks         []string          `json:"ipBlocks,omitempty"`
	NotIPBlocks      []string          `json:"notIpBlocks,omitempty"`
	RemoteIPBlocks   []string          `json:"remoteIpBlocks,omitempty"`
	NotRemoteIPBlocks []string         `json:"notRemoteIpBlocks,omitempty"`
}

type Operation struct {
	Hosts   []string `json:"hosts,omitempty"`
	NotHosts []string `json:"notHosts,omitempty"`
	Ports   []string `json:"ports,omitempty"`
	NotPorts []string `json:"notPorts,omitempty"`
	Methods []string `json:"methods,omitempty"`
	NotMethods []string `json:"notMethods,omitempty"`
	Paths   []string `json:"paths,omitempty"`
	NotPaths []string `json:"notPaths,omitempty"`
}

type Condition struct {
	Key    string   `json:"key"`
	Values []string `json:"values,omitempty"`
	NotValues []string `json:"notValues,omitempty"`
}

// ---------- Sidecar ----------

// Sidecar Sidecar 配置
type Sidecar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SidecarSpec `json:"spec"`
}

type SidecarSpec struct {
	WorkloadSelector *WorkloadSelector `json:"workloadSelector,omitempty"`
	Ingress          []IstioIngress   `json:"ingress,omitempty"`
	Egress           []IstioEgress    `json:"egress,omitempty"`
	OutboundTrafficPolicy *OutboundTrafficPolicy `json:"outboundTrafficPolicy,omitempty"`
}

type IstioIngress struct {
	Port               *PortSelector `json:"port,omitempty"`
	Bind               string        `json:"bind,omitempty"`
	CaptureMode        string        `json:"captureMode,omitempty"`
	DefaultEndpoint    string        `json:"defaultEndpoint,omitempty"`
}

type IstioEgress struct {
	Port  *PortSelector `json:"port,omitempty"`
	Bind  string        `json:"bind,omitempty"`
	Hosts []string      `json:"hosts,omitempty"`
}

type OutboundTrafficPolicy struct {
	Mode string `json:"mode,omitempty"` // REGISTRY_ONLY or ALLOW_ANY
}

// ---------- ServiceEntry ----------

// ServiceEntry 外部服务入口
type ServiceEntry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServiceEntrySpec `json:"spec"`
}

type ServiceEntrySpec struct {
	Hosts      []string          `json:"hosts"`
	Addresses  []string          `json:"addresses,omitempty"`
	Ports      []ServicePort     `json:"ports"`
	Location   string            `json:"location,omitempty"` // MESH_EXTERNAL or MESH_INTERNAL
	Resolution string            `json:"resolution,omitempty"` // NONE, STATIC, DNS, DNS_ROUND_ROBIN
	Endpoints  []WorkloadEntry   `json:"endpoints,omitempty"`
	ExportTo   []string          `json:"exportTo,omitempty"`
	SubjectAltNames []string     `json:"subjectAltNames,omitempty"`
	WorkloadSelector *WorkloadSelector `json:"workloadSelector,omitempty"`
}

type ServicePort struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
	Name     string `json:"name"`
}

type WorkloadEntry struct {
	Address string            `json:"address"`
	Ports   map[string]int    `json:"ports,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
	Weight  int               `json:"weight,omitempty"`
}

// ---------- 通用类型 ----------

// IstioCondition Istio 资源状态条件
type IstioCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastProbeTime      string `json:"lastProbeTime,omitempty"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
}

// =========================================================================
// Dynamic Client 初始化
// =========================================================================

// GetDynamicClient 返回 dynamic.Interface 用于操作 Istio CRD
func (c *Client) GetDynamicClient() (dynamic.Interface, error) {
	if c.dynamicClient != nil {
		return c.dynamicClient, nil
	}
	dc, err := dynamic.NewForConfig(c.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	c.dynamicClient = dc
	return dc, nil
}

// =========================================================================
// 通用 CRD 操作方法
// =========================================================================

// createIstioResource 创建 Istio CRD 资源（通用方法）
func (c *Client) createIstioResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	dc, err := c.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		return dc.Resource(gvr).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
	}
	return dc.Resource(gvr).Create(ctx, obj, metav1.CreateOptions{})
}

// getIstioResource 获取 Istio CRD 资源（通用方法）
func (c *Client) getIstioResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	dc, err := c.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		return dc.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	}
	return dc.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
}

// updateIstioResource 更新 Istio CRD 资源（通用方法）
func (c *Client) updateIstioResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	dc, err := c.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		return dc.Resource(gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	}
	return dc.Resource(gvr).Update(ctx, obj, metav1.UpdateOptions{})
}

// deleteIstioResource 删除 Istio CRD 资源（通用方法）
func (c *Client) deleteIstioResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
	dc, err := c.GetDynamicClient()
	if err != nil {
		return err
	}
	if namespace != "" {
		return dc.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	}
	return dc.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
}

// listIstioResources 列出 Istio CRD 资源（通用方法）
func (c *Client) listIstioResources(ctx context.Context, gvr schema.GroupVersionResource, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	dc, err := c.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		return dc.Resource(gvr).Namespace(namespace).List(ctx, opts)
	}
	return dc.Resource(gvr).List(ctx, opts)
}

// patchIstioResource patches an Istio CRD resource (generic)
func (c *Client) patchIstioResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, patchData []byte, patchType types.PatchType) (*unstructured.Unstructured, error) {
	dc, err := c.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		return dc.Resource(gvr).Namespace(namespace).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	}
	return dc.Resource(gvr).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
}

// =========================================================================
// VirtualService 操作
// =========================================================================

// CreateVirtualService 创建 VirtualService
func (c *Client) CreateVirtualService(ctx context.Context, namespace string, vs *VirtualService) error {
	obj, err := toUnstructured(vs)
	if err != nil {
		return fmt.Errorf("convert VirtualService to unstructured: %w", err)
	}
	_, err = c.createIstioResource(ctx, VirtualServiceGVR, namespace, obj)
	if err != nil {
		return fmt.Errorf("create VirtualService %s/%s: %w", namespace, vs.Name, err)
	}
	log.Printf("[Istio] Created VirtualService %s/%s", namespace, vs.Name)
	return nil
}

// GetVirtualService 获取 VirtualService
func (c *Client) GetVirtualService(ctx context.Context, namespace, name string) (*VirtualService, error) {
	obj, err := c.getIstioResource(ctx, VirtualServiceGVR, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("get VirtualService %s/%s: %w", namespace, name, err)
	}
	var vs VirtualService
	if err := fromUnstructured(obj, &vs); err != nil {
		return nil, fmt.Errorf("convert from unstructured: %w", err)
	}
	return &vs, nil
}

// UpdateVirtualService 更新 VirtualService
func (c *Client) UpdateVirtualService(ctx context.Context, namespace string, vs *VirtualService) error {
	obj, err := toUnstructured(vs)
	if err != nil {
		return fmt.Errorf("convert VirtualService to unstructured: %w", err)
	}
	_, err = c.updateIstioResource(ctx, VirtualServiceGVR, namespace, obj)
	if err != nil {
		return fmt.Errorf("update VirtualService %s/%s: %w", namespace, vs.Name, err)
	}
	log.Printf("[Istio] Updated VirtualService %s/%s", namespace, vs.Name)
	return nil
}

// DeleteVirtualService 删除 VirtualService
func (c *Client) DeleteVirtualService(ctx context.Context, namespace, name string) error {
	if err := c.deleteIstioResource(ctx, VirtualServiceGVR, namespace, name); err != nil {
		return fmt.Errorf("delete VirtualService %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Deleted VirtualService %s/%s", namespace, name)
	return nil
}

// ListVirtualServices 列出命名空间下的 VirtualService
func (c *Client) ListVirtualServices(ctx context.Context, namespace string) ([]VirtualService, error) {
	objList, err := c.listIstioResources(ctx, VirtualServiceGVR, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list VirtualServices: %w", err)
	}
	var result []VirtualService
	for _, obj := range objList.Items {
		var vs VirtualService
		if err := fromUnstructured(&obj, &vs); err != nil {
			log.Printf("[Istio] Warning: failed to parse VirtualService: %v", err)
			continue
		}
		result = append(result, vs)
	}
	return result, nil
}

// PatchVirtualService patches a VirtualService using JSON merge patch
func (c *Client) PatchVirtualService(ctx context.Context, namespace, name string, patchData []byte) error {
	_, err := c.patchIstioResource(ctx, VirtualServiceGVR, namespace, name, patchData, types.MergePatchType)
	if err != nil {
		return fmt.Errorf("patch VirtualService %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Patched VirtualService %s/%s", namespace, name)
	return nil
}

// PatchVirtualServiceStrategic patches a VirtualService using strategic merge patch
func (c *Client) PatchVirtualServiceStrategic(ctx context.Context, namespace, name string, patchData []byte) error {
	_, err := c.patchIstioResource(ctx, VirtualServiceGVR, namespace, name, patchData, types.StrategicMergePatchType)
	if err != nil {
		return fmt.Errorf("strategic patch VirtualService %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Strategic-patched VirtualService %s/%s", namespace, name)
	return nil
}

// =========================================================================
// DestinationRule 操作
// =========================================================================

// CreateDestinationRule 创建 DestinationRule
func (c *Client) CreateDestinationRule(ctx context.Context, namespace string, dr *DestinationRule) error {
	obj, err := toUnstructured(dr)
	if err != nil {
		return fmt.Errorf("convert DestinationRule to unstructured: %w", err)
	}
	_, err = c.createIstioResource(ctx, DestinationRuleGVR, namespace, obj)
	if err != nil {
		return fmt.Errorf("create DestinationRule %s/%s: %w", namespace, dr.Name, err)
	}
	log.Printf("[Istio] Created DestinationRule %s/%s", namespace, dr.Name)
	return nil
}

// GetDestinationRule 获取 DestinationRule
func (c *Client) GetDestinationRule(ctx context.Context, namespace, name string) (*DestinationRule, error) {
	obj, err := c.getIstioResource(ctx, DestinationRuleGVR, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("get DestinationRule %s/%s: %w", namespace, name, err)
	}
	var dr DestinationRule
	if err := fromUnstructured(obj, &dr); err != nil {
		return nil, fmt.Errorf("convert from unstructured: %w", err)
	}
	return &dr, nil
}

// UpdateDestinationRule 更新 DestinationRule
func (c *Client) UpdateDestinationRule(ctx context.Context, namespace string, dr *DestinationRule) error {
	obj, err := toUnstructured(dr)
	if err != nil {
		return fmt.Errorf("convert DestinationRule to unstructured: %w", err)
	}
	_, err = c.updateIstioResource(ctx, DestinationRuleGVR, namespace, obj)
	if err != nil {
		return fmt.Errorf("update DestinationRule %s/%s: %w", namespace, dr.Name, err)
	}
	log.Printf("[Istio] Updated DestinationRule %s/%s", namespace, dr.Name)
	return nil
}

// DeleteDestinationRule 删除 DestinationRule
func (c *Client) DeleteDestinationRule(ctx context.Context, namespace, name string) error {
	if err := c.deleteIstioResource(ctx, DestinationRuleGVR, namespace, name); err != nil {
		return fmt.Errorf("delete DestinationRule %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Deleted DestinationRule %s/%s", namespace, name)
	return nil
}

// PatchDestinationRule patches a DestinationRule
func (c *Client) PatchDestinationRule(ctx context.Context, namespace, name string, patchData []byte) error {
	_, err := c.patchIstioResource(ctx, DestinationRuleGVR, namespace, name, patchData, types.MergePatchType)
	if err != nil {
		return fmt.Errorf("patch DestinationRule %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Patched DestinationRule %s/%s", namespace, name)
	return nil
}

// =========================================================================
// Gateway 操作
// =========================================================================

// CreateGateway 创建 Gateway
func (c *Client) CreateGateway(ctx context.Context, namespace string, gw *Gateway) error {
	obj, err := toUnstructured(gw)
	if err != nil {
		return fmt.Errorf("convert Gateway to unstructured: %w", err)
	}
	_, err = c.createIstioResource(ctx, GatewayGVR, namespace, obj)
	if err != nil {
		return fmt.Errorf("create Gateway %s/%s: %w", namespace, gw.Name, err)
	}
	log.Printf("[Istio] Created Gateway %s/%s", namespace, gw.Name)
	return nil
}

// GetGateway 获取 Gateway
func (c *Client) GetGateway(ctx context.Context, namespace, name string) (*Gateway, error) {
	obj, err := c.getIstioResource(ctx, GatewayGVR, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("get Gateway %s/%s: %w", namespace, name, err)
	}
	var gw Gateway
	if err := fromUnstructured(obj, &gw); err != nil {
		return nil, fmt.Errorf("convert from unstructured: %w", err)
	}
	return &gw, nil
}

// UpdateGateway 更新 Gateway
func (c *Client) UpdateGateway(ctx context.Context, namespace string, gw *Gateway) error {
	obj, err := toUnstructured(gw)
	if err != nil {
		return fmt.Errorf("convert Gateway to unstructured: %w", err)
	}
	_, err = c.updateIstioResource(ctx, GatewayGVR, namespace, obj)
	if err != nil {
		return fmt.Errorf("update Gateway %s/%s: %w", namespace, gw.Name, err)
	}
	log.Printf("[Istio] Updated Gateway %s/%s", namespace, gw.Name)
	return nil
}

// DeleteGateway 删除 Gateway
func (c *Client) DeleteGateway(ctx context.Context, namespace, name string) error {
	if err := c.deleteIstioResource(ctx, GatewayGVR, namespace, name); err != nil {
		return fmt.Errorf("delete Gateway %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Deleted Gateway %s/%s", namespace, name)
	return nil
}

// =========================================================================
// PeerAuthentication 操作
// =========================================================================

// EnsurePeerAuthentication 确保命名空间级别的 mTLS 策略存在
// mode: "UNSET", "DISABLE", "PERMISSIVE", "STRICT"
func (c *Client) EnsurePeerAuthentication(ctx context.Context, namespace string, mode string) error {
	name := "default"
	if mode == "" {
		mode = "PERMISSIVE"
	}

	existing, err := c.getIstioResource(ctx, PeerAuthenticationGVR, namespace, name)
	if err == nil {
		// 已存在，检查是否需要更新
		var pa PeerAuthentication
		_ = fromUnstructured(existing, &pa)
		if pa.Spec.MTLS != nil && strings.EqualFold(pa.Spec.MTLS.Mode, mode) {
			return nil // 模式已匹配
		}
	}

	pa := &PeerAuthentication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: IstioSecurityAPIGroup + "/" + IstioSecurityAPIVersion,
			Kind:       "PeerAuthentication",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "my-cloud",
			},
		},
		Spec: PeerAuthenticationSpec{
			MTLS: &MutualTLS{Mode: mode},
		},
	}

	obj, err := toUnstructured(pa)
	if err != nil {
		return fmt.Errorf("convert PeerAuthentication: %w", err)
	}

	if existing != nil {
		obj.SetResourceVersion(existing.GetResourceVersion())
		_, err = c.updateIstioResource(ctx, PeerAuthenticationGVR, namespace, obj)
		if err != nil {
			return fmt.Errorf("update PeerAuthentication %s/%s: %w", namespace, name, err)
		}
		log.Printf("[Istio] Updated PeerAuthentication %s/%s mode=%s", namespace, name, mode)
	} else {
		_, err = c.createIstioResource(ctx, PeerAuthenticationGVR, namespace, obj)
		if err != nil {
			return fmt.Errorf("create PeerAuthentication %s/%s: %w", namespace, name, err)
		}
		log.Printf("[Istio] Created PeerAuthentication %s/%s mode=%s", namespace, name, mode)
	}

	return nil
}

// GetPeerAuthentication 获取 PeerAuthentication
func (c *Client) GetPeerAuthentication(ctx context.Context, namespace, name string) (*PeerAuthentication, error) {
	obj, err := c.getIstioResource(ctx, PeerAuthenticationGVR, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("get PeerAuthentication %s/%s: %w", namespace, name, err)
	}
	var pa PeerAuthentication
	if err := fromUnstructured(obj, &pa); err != nil {
		return nil, fmt.Errorf("convert from unstructured: %w", err)
	}
	return &pa, nil
}

// DeletePeerAuthentication 删除 PeerAuthentication
func (c *Client) DeletePeerAuthentication(ctx context.Context, namespace, name string) error {
	if err := c.deleteIstioResource(ctx, PeerAuthenticationGVR, namespace, name); err != nil {
		return fmt.Errorf("delete PeerAuthentication %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Deleted PeerAuthentication %s/%s", namespace, name)
	return nil
}

// =========================================================================
// AuthorizationPolicy 操作
// =========================================================================

// EnsureAuthorizationPolicy 确保命名空间级别的授权策略存在
func (c *Client) EnsureAuthorizationPolicy(ctx context.Context, namespace string, rules []Rule, action string) error {
	name := "default"
	if action == "" {
		action = "ALLOW"
	}

	ap := &AuthorizationPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: IstioSecurityAPIGroup + "/" + IstioSecurityAPIVersion,
			Kind:       "AuthorizationPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "my-cloud",
			},
		},
		Spec: AuthorizationPolicySpec{
			Action: action,
			Rules:  rules,
		},
	}

	obj, err := toUnstructured(ap)
	if err != nil {
		return fmt.Errorf("convert AuthorizationPolicy: %w", err)
	}

	existing, getErr := c.getIstioResource(ctx, AuthorizationPolicyGVR, namespace, name)
	if getErr == nil {
		obj.SetResourceVersion(existing.GetResourceVersion())
		_, err = c.updateIstioResource(ctx, AuthorizationPolicyGVR, namespace, obj)
	} else {
		_, err = c.createIstioResource(ctx, AuthorizationPolicyGVR, namespace, obj)
	}

	if err != nil {
		return fmt.Errorf("ensure AuthorizationPolicy %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Ensured AuthorizationPolicy %s/%s action=%s", namespace, name, action)
	return nil
}

// =========================================================================
// Sidecar 操作
// =========================================================================

// EnsureDefaultSidecar 为命名空间创建默认 Sidecar 配置
func (c *Client) EnsureDefaultSidecar(ctx context.Context, namespace string) error {
	name := "default"

	sidecar := &Sidecar{
		TypeMeta: metav1.TypeMeta{
			APIVersion: IstioNetworkingAPIGroup + "/" + IstioNetworkingAPIVersion,
			Kind:       "Sidecar",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "my-cloud",
			},
		},
		Spec: SidecarSpec{
			OutboundTrafficPolicy: &OutboundTrafficPolicy{
				Mode: "ALLOW_ANY",
			},
		},
	}

	obj, err := toUnstructured(sidecar)
	if err != nil {
		return fmt.Errorf("convert Sidecar: %w", err)
	}

	existing, getErr := c.getIstioResource(ctx, SidecarGVR, namespace, name)
	if getErr == nil {
		obj.SetResourceVersion(existing.GetResourceVersion())
		_, err = c.updateIstioResource(ctx, SidecarGVR, namespace, obj)
	} else {
		_, err = c.createIstioResource(ctx, SidecarGVR, namespace, obj)
	}

	if err != nil {
		return fmt.Errorf("ensure Sidecar %s/%s: %w", namespace, name, err)
	}
	log.Printf("[Istio] Ensured Sidecar %s/%s", namespace, name)
	return nil
}

// =========================================================================
// 金丝雀部署辅助方法
// =========================================================================

// CanaryVirtualServiceConfig 金丝雀 VirtualService 配置
type CanaryVirtualServiceConfig struct {
	Name           string            // VirtualService 名称
	Namespace      string            // 命名空间
	Hosts          []string          // 目标主机列表
	Gateways       []string          // Gateway 列表（如 "mesh", "istio-system/istio-ingressgateway"）
	StableHost     string            // stable 子集对应的 K8s Service host
	CanaryHost     string            // canary 子集对应的 K8s Service host
	StableSubset   string            // stable 子集名（默认 "stable"）
	CanarySubset   string            // canary 子集名（默认 "canary"）
	CanaryWeight   int               // canary 流量权重 (0-100)
	StableWeight   int               // stable 流量权重 (0-100) 自动计算: 100 - canaryWeight
	HeaderMatches  []HeaderMatchRule // 基于 Header 的匹配规则（优先于权重）
	UriMatch       *StringMatch      // URI 匹配
	Timeout        string            // 请求超时
	Retries        *HTTPRetries      // 重试策略
	Labels         map[string]string // 标签
}

// HeaderMatchRule Header 匹配规则
type HeaderMatchRule struct {
	HeaderName  string // e.g. "x-canary-version"
	HeaderValue string // e.g. "v2"
	Exact       bool   // true: exact match, false: regex match
	Subset      string // 路由到哪个子集
}

// CanaryDestinationRuleConfig 金丝雀 DestinationRule 配置
type CanaryDestinationRuleConfig struct {
	Name            string            // DestinationRule 名称
	Namespace       string            // 命名空间
	Host            string            // K8s Service host
	StableSubset    string            // stable 子集名
	CanarySubset    string            // canary 子集名
	StableLabels    map[string]string // stable 子集 Pod 标签选择器
	CanaryLabels    map[string]string // canary 子集 Pod 标签选择器
	TrafficPolicy   *TrafficPolicy    // 全局流量策略
	Labels          map[string]string // 标签
}

// EnsureCanaryVirtualService 创建或更新金丝雀 VirtualService
// 自动构建 header match + weight 路由规则
func (c *Client) EnsureCanaryVirtualService(ctx context.Context, config CanaryVirtualServiceConfig) error {
	// 默认值
	if config.StableSubset == "" {
		config.StableSubset = "stable"
	}
	if config.CanarySubset == "" {
		config.CanarySubset = "canary"
	}
	if config.CanaryWeight < 0 || config.CanaryWeight > 100 {
		return fmt.Errorf("canary weight must be 0-100, got %d", config.CanaryWeight)
	}
	if config.StableHost == "" {
		config.StableHost = config.Name
	}
	if config.CanaryHost == "" {
		config.CanaryHost = config.StableHost
	}

	// 构建路由规则
	var httpRoutes []HTTPRoute

	// 1. Header-based 路由（优先级最高）
	for _, hm := range config.HeaderMatches {
		matchType := StringMatch{Regex: ".*" + hm.HeaderValue + ".*"}
		if hm.Exact {
			matchType = StringMatch{Exact: hm.HeaderValue}
		}

		route := HTTPRoute{
			Match: []HTTPMatch{
				{
					Headers: map[string]StringMatch{
						hm.HeaderName: matchType,
					},
				},
			},
			Route: []RouteDestination{
				{
					Destination: Destination{
						Host:   hm.Subset,
						Subset: config.CanarySubset,
					},
				},
			},
		}

		if config.UriMatch != nil {
			route.Match[0].URI = config.UriMatch
		}

		if hm.Subset == config.StableSubset {
			route.Route[0].Destination.Subset = config.StableSubset
            if config.StableHost != config.CanaryHost {
                route.Route[0].Destination.Host = config.StableHost
            }
		} else {
            if config.CanaryHost != config.StableHost && hm.Subset == config.CanarySubset {
                route.Route[0].Destination.Host = config.CanaryHost
            }
        }

		httpRoutes = append(httpRoutes, route)
	}

	// 2. Weight-based 路由（默认规则）
	weightRoute := HTTPRoute{
		Route: []RouteDestination{
			{
				Destination: Destination{
					Host:   config.StableHost,
					Subset: config.StableSubset,
				},
				Weight: config.StableWeight,
			},
			{
				Destination: Destination{
					Host:   config.CanaryHost,
					Subset: config.CanarySubset,
				},
				Weight: config.CanaryWeight,
			},
		},
	}

	if config.Timeout != "" {
		weightRoute.Timeout = config.Timeout
	}
	if config.Retries != nil {
		weightRoute.Retries = config.Retries
	}

	httpRoutes = append(httpRoutes, weightRoute)

	// 构建 VirtualService
	gateways := config.Gateways
	if len(gateways) == 0 {
		gateways = []string{"mesh"}
	}

	labels := config.Labels
	if labels == nil {
		labels = map[string]string{
			"app":        config.Name,
			"managed-by": "my-cloud",
		}
	}

	vs := &VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: IstioNetworkingAPIGroup + "/" + IstioNetworkingAPIVersion,
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
			Labels:    labels,
		},
		Spec: VirtualServiceSpec{
			Hosts:    config.Hosts,
			Gateways: gateways,
			HTTP:     httpRoutes,
		},
	}

	// 检查是否已存在
	existing, err := c.GetVirtualService(ctx, config.Namespace, config.Name)
	if err == nil {
		// 更新
		vs.ResourceVersion = existing.ResourceVersion
		return c.UpdateVirtualService(ctx, config.Namespace, vs)
	}

	// 创建
	return c.CreateVirtualService(ctx, config.Namespace, vs)
}

// EnsureCanaryDestinationRule 创建或更新金丝雀 DestinationRule
func (c *Client) EnsureCanaryDestinationRule(ctx context.Context, config CanaryDestinationRuleConfig) error {
	// 默认值
	if config.StableSubset == "" {
		config.StableSubset = "stable"
	}
	if config.CanarySubset == "" {
		config.CanarySubset = "canary"
	}
	if config.Host == "" {
		config.Host = config.Name
	}

	labels := config.Labels
	if labels == nil {
		labels = map[string]string{
			"app":        config.Name,
			"managed-by": "my-cloud",
		}
	}

	dr := &DestinationRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: IstioNetworkingAPIGroup + "/" + IstioNetworkingAPIVersion,
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
			Labels:    labels,
		},
		Spec: DestinationRuleSpec{
			Host: config.Host,
			Subsets: []Subset{
				{
					Name:   config.StableSubset,
					Labels: config.StableLabels,
				},
				{
					Name:   config.CanarySubset,
					Labels: config.CanaryLabels,
				},
			},
			TrafficPolicy: config.TrafficPolicy,
		},
	}

	// 检查是否已存在
	existing, err := c.GetDestinationRule(ctx, config.Namespace, config.Name)
	if err == nil {
		dr.ResourceVersion = existing.ResourceVersion
		return c.UpdateDestinationRule(ctx, config.Namespace, dr)
	}

	return c.CreateDestinationRule(ctx, config.Namespace, dr)
}

// AdjustCanaryTrafficWeight 动态调整金丝雀流量权重（patch VirtualService）
func (c *Client) AdjustCanaryTrafficWeight(ctx context.Context, namespace, virtualServiceName string, canaryWeight int) error {
	if canaryWeight < 0 || canaryWeight > 100 {
		return fmt.Errorf("canary weight must be 0-100, got %d", canaryWeight)
	}

	// GET → 修改 → UPDATE，避免 JSON merge patch 覆盖 destination 字段
	vs, err := c.GetVirtualService(ctx, namespace, virtualServiceName)
	if err != nil {
		return fmt.Errorf("get VirtualService: %w", err)
	}

	stableWeight := 100 - canaryWeight
	found := false

	for i := range vs.Spec.HTTP {
		routes := vs.Spec.HTTP[i].Route
		// 找到同时包含 stable 和 canary subset 的权重路由
		hasStable := false
		hasCanary := false
		for _, r := range routes {
			if r.Destination.Subset == "stable" {
				hasStable = true
			}
			if r.Destination.Subset == "canary" {
				hasCanary = true
			}
		}
		if hasStable && hasCanary {
			for j := range routes {
				if routes[j].Destination.Subset == "stable" {
					vs.Spec.HTTP[i].Route[j].Weight = stableWeight
				} else if routes[j].Destination.Subset == "canary" {
					vs.Spec.HTTP[i].Route[j].Weight = canaryWeight
				}
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no canary weight route found in VirtualService %s/%s", namespace, virtualServiceName)
	}

	return c.UpdateVirtualService(ctx, namespace, vs)
}

// GetCanaryTrafficWeight 读取金丝雀流量权重（从 VirtualService 的最后一条 HTTP 路由）
func (c *Client) GetCanaryTrafficWeight(ctx context.Context, namespace, virtualServiceName string) (int, error) {
	vs, err := c.GetVirtualService(ctx, namespace, virtualServiceName)
	if err != nil {
		return 0, err
	}

	// 找到包含 canary subset 的路由，返回其权重
	for _, http := range vs.Spec.HTTP {
		for _, route := range http.Route {
			if route.Destination.Subset == "canary" {
				return route.Weight, nil
			}
		}
	}

	return 0, fmt.Errorf("canary subset not found in VirtualService %s/%s", namespace, virtualServiceName)
}

// RemoveCanaryFromVirtualService 从 VirtualService 中移除金丝雀路由
// 将 VirtualService 恢复为仅路由到 stable 的全量状态
func (c *Client) RemoveCanaryFromVirtualService(ctx context.Context, namespace, virtualServiceName, stableHost, stableSubset string) error {
	if stableSubset == "" {
		stableSubset = "stable"
	}

	vs, err := c.GetVirtualService(ctx, namespace, virtualServiceName)
	if err != nil {
		return fmt.Errorf("get VirtualService: %w", err)
	}

	// 移除所有 header match 规则，只保留全量路由到 stable
	vs.Spec.HTTP = []HTTPRoute{
		{
			Route: []RouteDestination{
				{
					Destination: Destination{
						Host:   stableHost,
						Subset: stableSubset,
					},
					Weight: 100,
				},
			},
		},
	}

	return c.UpdateVirtualService(ctx, namespace, vs)
}

// =========================================================================
// 完整的金丝雀发布生命周期方法
// =========================================================================

// DeployCanaryResources 部署金丝雀所需的完整 Istio 资源
// 包括: DestinationRule (子集定义) + VirtualService (流量路由)
func (c *Client) DeployCanaryResources(ctx context.Context, cfg CanaryDeployConfig) error {
	log.Printf("[Istio] Deploying canary resources: VS=%s/%s DR=%s/%s weight=%d%%",
		cfg.Namespace, cfg.VirtualServiceName, cfg.Namespace, cfg.DestinationRuleName, cfg.CanaryWeight)

	// 1. 确保 DestinationRule 存在
	drConfig := CanaryDestinationRuleConfig{
		Name:         cfg.DestinationRuleName,
		Namespace:    cfg.Namespace,
		Host:         cfg.ServiceHost,
		StableSubset: cfg.StableSubset,
		CanarySubset: cfg.CanarySubset,
		StableLabels: map[string]string{
			"app":     cfg.AppName,
			"version": cfg.StableVersion,
		},
		CanaryLabels: map[string]string{
			"app":     cfg.AppName,
			"version": cfg.CanaryVersion,
		},
		Labels: map[string]string{
			"app":        cfg.AppName,
			"managed-by": "my-cloud",
		},
	}

	if err := c.EnsureCanaryDestinationRule(ctx, drConfig); err != nil {
		return fmt.Errorf("ensure DestinationRule: %w", err)
	}

	// 2. 确保 VirtualService 存在
	vsConfig := CanaryVirtualServiceConfig{
		Name:         cfg.VirtualServiceName,
		Namespace:    cfg.Namespace,
		Hosts:        cfg.Hosts,
		Gateways:     cfg.Gateways,
		StableHost:   cfg.ServiceHost,
		CanaryHost:   cfg.ServiceHost,
		StableSubset: cfg.StableSubset,
		CanarySubset: cfg.CanarySubset,
		CanaryWeight: cfg.CanaryWeight,
		StableWeight: 100 - cfg.CanaryWeight,
		Timeout:      cfg.Timeout,
		Labels: map[string]string{
			"app":        cfg.AppName,
			"managed-by": "my-cloud",
		},
	}

	if err := c.EnsureCanaryVirtualService(ctx, vsConfig); err != nil {
		return fmt.Errorf("ensure VirtualService: %w", err)
	}

	log.Printf("[Istio] Canary resources deployed: %d%% traffic → canary, %d%% → stable",
		cfg.CanaryWeight, 100-cfg.CanaryWeight)
	return nil
}

// PromoteCanaryToStable 将金丝雀版本提升为稳定版本
// 将 VirtualService 恢复为全量路由到 stable，删除 canary subset
func (c *Client) PromoteCanaryToStable(ctx context.Context, cfg CanaryDeployConfig) error {
	log.Printf("[Istio] Promoting canary to stable: VS=%s/%s", cfg.Namespace, cfg.VirtualServiceName)

	// 1. 更新 VirtualService 为全量路由到 stable
	vs, err := c.GetVirtualService(ctx, cfg.Namespace, cfg.VirtualServiceName)
	if err != nil {
		return fmt.Errorf("get VirtualService: %w", err)
	}

	// 保留原有的 hosts 和 gateways，修改路由为全量 stable
	vs.Spec.HTTP = []HTTPRoute{
		{
			Route: []RouteDestination{
				{
					Destination: Destination{
						Host:   cfg.ServiceHost,
						Subset: cfg.StableSubset,
					},
					Weight: 100,
				},
			},
		},
	}

	if err := c.UpdateVirtualService(ctx, cfg.Namespace, vs); err != nil {
		return fmt.Errorf("update VirtualService: %w", err)
	}

	// 2. 更新 DestinationRule 的 stable 子集标签指向新版本
	dr, err := c.GetDestinationRule(ctx, cfg.Namespace, cfg.DestinationRuleName)
	if err != nil {
		return fmt.Errorf("get DestinationRule: %w", err)
	}

	// 更新 stable subset 的 labels 指向 canary version（新稳定版）
	for i, subset := range dr.Spec.Subsets {
		if subset.Name == cfg.StableSubset {
			dr.Spec.Subsets[i].Labels = map[string]string{
				"app":     cfg.AppName,
				"version": cfg.CanaryVersion, // canary becomes new stable
			}
		}
	}

	if err := c.UpdateDestinationRule(ctx, cfg.Namespace, dr); err != nil {
		return fmt.Errorf("update DestinationRule: %w", err)
	}

	log.Printf("[Istio] Canary promoted to stable: version=%s at 100%% traffic", cfg.CanaryVersion)
	return nil
}

// RollbackCanaryResources 回滚金丝雀资源
// 删除 VirtualService 中的 canary 路由，恢复到 100% stable
func (c *Client) RollbackCanaryResources(ctx context.Context, cfg CanaryDeployConfig) error {
	log.Printf("[Istio] Rolling back canary: VS=%s/%s", cfg.Namespace, cfg.VirtualServiceName)

	// 将 VirtualService 恢复为 100% stable
	if err := c.RemoveCanaryFromVirtualService(ctx, cfg.Namespace, cfg.VirtualServiceName, cfg.ServiceHost, cfg.StableSubset); err != nil {
		return fmt.Errorf("remove canary from VirtualService: %w", err)
	}

	log.Printf("[Istio] Canary rolled back: 100%% traffic restored to stable")
	return nil
}

// CleanupCanaryResources 彻底清理金丝雀资源
func (c *Client) CleanupCanaryResources(ctx context.Context, namespace string, resourceNames CanaryResourceNames) []error {
	var errs []error

	if resourceNames.VirtualServiceName != "" {
		if err := c.DeleteVirtualService(ctx, namespace, resourceNames.VirtualServiceName); err != nil {
			errs = append(errs, fmt.Errorf("delete VirtualService: %w", err))
		}
	}

	if resourceNames.DestinationRuleName != "" {
		if err := c.DeleteDestinationRule(ctx, namespace, resourceNames.DestinationRuleName); err != nil {
			errs = append(errs, fmt.Errorf("delete DestinationRule: %w", err))
		}
	}

	if resourceNames.GatewayName != "" {
		if err := c.DeleteGateway(ctx, namespace, resourceNames.GatewayName); err != nil {
			errs = append(errs, fmt.Errorf("delete Gateway: %w", err))
		}
	}

	if len(errs) > 0 {
		log.Printf("[Istio] Cleanup completed with %d errors for %s", len(errs), namespace)
	} else {
		log.Printf("[Istio] Cleanup completed successfully for %s", namespace)
	}

	return errs
}

// CanaryDeployConfig 金丝雀部署配置
type CanaryDeployConfig struct {
	Namespace           string   // 命名空间
	AppName             string   // 应用名称
	ServiceHost         string   // K8s Service host
	VirtualServiceName  string   // VirtualService 资源名
	DestinationRuleName string   // DestinationRule 资源名
	GatewayName         string   // Gateway 资源名（可选）
	Hosts               []string // VirtualService hosts
	Gateways            []string // VirtualService gateways
	StableSubset        string   // stable 子集名
	CanarySubset        string   // canary 子集名
	StableVersion       string   // stable 版本标签值
	CanaryVersion       string   // canary 版本标签值
	CanaryWeight        int      // canary 流量权重 (0-100)
	Timeout             string   // 请求超时
}

// CanaryResourceNames 金丝雀资源名称
type CanaryResourceNames struct {
	VirtualServiceName  string
	DestinationRuleName string
	GatewayName         string
}

// =========================================================================
// 向后兼容：保留旧版 API（使用 Generic REST Client）
// =========================================================================

// SetCanaryWeight patches an Istio VirtualService to adjust canary traffic weight
// Deprecated: 使用 AdjustCanaryTrafficWeight 代替
func (c *Client) SetCanaryWeight(ctx context.Context, namespace, virtualServiceName string, weight int) error {
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"http": []map[string]interface{}{
				{
					"route": []map[string]interface{}{
						{
							"destination": map[string]interface{}{
								"host": virtualServiceName, "subset": "stable",
							},
							"weight": 100 - weight,
						},
						{
							"destination": map[string]interface{}{
								"host": virtualServiceName, "subset": "canary",
							},
							"weight": weight,
						},
					},
				},
			},
		},
	}
	patchBytes, _ := json.Marshal(patch)

	_, err := c.clientset.CoreV1().RESTClient().
		Patch(types.MergePatchType).
		Namespace(namespace).
		Resource("virtualservices").
		Name(virtualServiceName).
		SetHeader("Content-Type", "application/merge-patch+json").
		Body(patchBytes).
		Do(ctx).
		Raw()
	return err
}

// GetCanaryWeight reads the current canary weight from an Istio VirtualService
// Deprecated: 使用 GetCanaryTrafficWeight 代替
func (c *Client) GetCanaryWeight(ctx context.Context, namespace, virtualServiceName string) (int, error) {
	data, err := c.clientset.CoreV1().RESTClient().
		Get().
		Namespace(namespace).
		Resource("virtualservices").
		Name(virtualServiceName).
		Do(ctx).
		Raw()
	if err != nil {
		return 0, fmt.Errorf("get VirtualService %s/%s: %w", namespace, virtualServiceName, err)
	}

	var vs VirtualService
	if err := json.Unmarshal(data, &vs); err != nil {
		return 0, fmt.Errorf("unmarshal VirtualService: %w", err)
	}

	for _, http := range vs.Spec.HTTP {
		for _, route := range http.Route {
			if route.Destination.Subset == "canary" {
				return route.Weight, nil
			}
		}
	}

	return 0, fmt.Errorf("canary subset not found in VirtualService %s/%s", namespace, virtualServiceName)
}

// =========================================================================
// 工具函数
// =========================================================================

// toUnstructured 将任意结构体转换为 unstructured.Unstructured
func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var unstruct map[string]interface{}
	if err := json.Unmarshal(data, &unstruct); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: unstruct}, nil
}

// fromUnstructured 将 unstructured.Unstructured 转换为指定的结构体
func fromUnstructured(obj *unstructured.Unstructured, target interface{}) error {
	data, err := json.Marshal(obj.Object)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// =========================================================================
// 健康检查辅助方法
// =========================================================================

// IsIstioInstalled 检查集群中是否安装了 Istio
// 通过检查是否可以列出 VirtualService CRD 来判断
func (c *Client) IsIstioInstalled(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.listIstioResources(ctx, VirtualServiceGVR, "", metav1.ListOptions{Limit: 1})
	return err == nil
}

// GetIstioVersion 获取 Istio 版本
func (c *Client) GetIstioVersion(ctx context.Context) (string, error) {
	dc, err := c.GetDynamicClient()
	if err != nil {
		return "", err
	}

	// 查询 istio-system namespace 或使用 discovery API
	data, err := dc.Resource(schema.GroupVersionResource{
		Group:   "apiextensions.k8s.io",
		Version: "v1",
		Resource: "customresourcedefinitions",
	}).Get(ctx, "virtualservices.networking.istio.io", metav1.GetOptions{})

	if err != nil {
		return "", fmt.Errorf("Istio CRD not found: %w", err)
	}

	labels := data.GetLabels()
	if version, ok := labels["version"]; ok {
		return version, nil
	}
	if release, ok := labels["release"]; ok {
		return release, nil
	}

	return "unknown", nil
}
