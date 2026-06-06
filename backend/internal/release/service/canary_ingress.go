package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"my-cloud/pkg/k8s"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ---------- Canary Ingress Builder ----------

// BuildCanaryIngressSpec 构建带 Nginx Canary 注解的 Ingress 规范
// routingMode: "weight", "header", "cookie", "weight_header"
func BuildCanaryIngressSpec(
	name, namespace, host, path, pathType string,
	canarySvcName string, servicePort int32,
	routingMode string, weight int,
	headerName, headerValue, cookieName, tlsSecretName string,
	labels map[string]string,
) *networkingv1.Ingress {

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/canary": "true",
	}

	switch routingMode {
	case "weight":
		annotations["nginx.ingress.kubernetes.io/canary-weight"] = fmt.Sprintf("%d", weight)
	case "header":
		annotations["nginx.ingress.kubernetes.io/canary-by-header"] = headerName
		if headerValue != "" {
			annotations["nginx.ingress.kubernetes.io/canary-by-header-value"] = headerValue
		}
	case "cookie":
		annotations["nginx.ingress.kubernetes.io/canary-by-cookie"] = cookieName
	case "weight_header":
		annotations["nginx.ingress.kubernetes.io/canary-weight"] = fmt.Sprintf("%d", weight)
		annotations["nginx.ingress.kubernetes.io/canary-by-header"] = headerName
		if headerValue != "" {
			annotations["nginx.ingress.kubernetes.io/canary-by-header-value"] = headerValue
		}
	}

	ingressPathType := networkingv1.PathType(pathType)
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: stringPtr("nginx"),
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     path,
									PathType: &ingressPathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: canarySvcName,
											Port: networkingv1.ServiceBackendPort{
												Number: servicePort,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// TLS
	if tlsSecretName != "" {
		ing.Spec.TLS = []networkingv1.IngressTLS{
			{
				SecretName: tlsSecretName,
				Hosts:      []string{host},
			},
		}
	}

	return ing
}

// ---------- Canary Service Builder ----------

// BuildCanaryServiceSpec 构建仅选中 canary Pod 的 Service
// selector 包含 version=canaryWorkloadName，确保流量只到 canary Pod
func BuildCanaryServiceSpec(
	name, namespace, appName, canaryWorkloadName string,
	port, targetPort int32,
) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        appName,
				"version":    canaryWorkloadName,
				"managed-by": "my-cloud",
				"role":       "canary",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app":        appName,
				"version":    canaryWorkloadName, // 关键：只匹配 canary Pod
				"managed-by": "my-cloud",
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       port,
					TargetPort: intstr.FromInt(int(targetPort)),
				},
			},
		},
	}
}

// ---------- Stable Ingress Config Extractor ----------

// ExtractStableIngressConfig 从已有 stable Ingress 中提取 host/path/pathType/TLS 配置
func ExtractStableIngressConfig(
	ctx context.Context,
	k8sClient *k8s.Client,
	namespace, stableIngressName string,
) (host, path, pathType, tlsSecretName string, servicePort int32, err error) {
	ing, err := k8sClient.GetIngress(ctx, namespace, stableIngressName)
	if err != nil {
		return "", "", "", "", 0, fmt.Errorf("failed to get stable ingress %s/%s: %w", namespace, stableIngressName, err)
	}

	if len(ing.Spec.Rules) == 0 {
		return "", "", "", "", 0, fmt.Errorf("stable ingress %s/%s has no rules", namespace, stableIngressName)
	}

	rule := ing.Spec.Rules[0]
	host = rule.Host

	if rule.HTTP != nil && len(rule.HTTP.Paths) > 0 {
		path = rule.HTTP.Paths[0].Path
		if rule.HTTP.Paths[0].PathType != nil {
			pathType = string(*rule.HTTP.Paths[0].PathType)
		} else {
			pathType = "Prefix"
		}
		if rule.HTTP.Paths[0].Backend.Service != nil {
			servicePort = rule.HTTP.Paths[0].Backend.Service.Port.Number
		}
	}

	if path == "" {
		path = "/"
	}
	if pathType == "" {
		pathType = "Prefix"
	}
	if servicePort == 0 {
		servicePort = 80
	}

	if len(ing.Spec.TLS) > 0 && ing.Spec.TLS[0].SecretName != "" {
		tlsSecretName = ing.Spec.TLS[0].SecretName
	}

	log.Printf("[CanaryIngress] Extracted from stable ingress %s/%s: host=%s path=%s pathType=%s port=%d tls=%s",
		namespace, stableIngressName, host, path, pathType, servicePort, tlsSecretName)

	return host, path, pathType, tlsSecretName, servicePort, nil
}

// ---------- Stable Service Selector Narrow/Widen ----------

// NarrowStableServiceSelector 缩小 stable Service 的 selector，加上 version 限定
// 这样 stable Service 只会路由到 stable Pod，不会误路由到 canary Pod
func NarrowStableServiceSelector(
	ctx context.Context,
	k8sClient *k8s.Client,
	namespace, stableSvcName, stableWorkloadName string,
) error {
	// 读取当前 stable Service
	svc, err := k8sClient.GetService(ctx, namespace, stableSvcName)
	if err != nil {
		return fmt.Errorf("failed to get stable service %s/%s: %w", namespace, stableSvcName, err)
	}

	// 检查是否已经被 narrow 过（幂等）
	if existingVersion, ok := svc.Spec.Selector["version"]; ok && existingVersion != "" {
		log.Printf("[CanaryIngress] Stable service %s/%s already narrowed (version=%s), skipping",
			namespace, stableSvcName, existingVersion)
		return nil
	}

	// 保存原始 selector（用于后续恢复）
	// 使用 strategic merge patch 添加 version selector
	patchData := []byte(fmt.Sprintf(`{"spec":{"selector":{"version":"%s"}}}`, stableWorkloadName))

	_, err = k8sClient.PatchService(ctx, namespace, stableSvcName, patchData)
	if err != nil {
		return fmt.Errorf("failed to narrow stable service selector: %w", err)
	}

	log.Printf("[CanaryIngress] Narrowed stable service %s/%s selector with version=%s",
		namespace, stableSvcName, stableWorkloadName)
	return nil
}

// WidenStableServiceSelector 恢复 stable Service 的 selector，移除 version 限定
// 在 ConfirmCanary 或 RollbackCanary 后调用
func WidenStableServiceSelector(
	ctx context.Context,
	k8sClient *k8s.Client,
	namespace, stableSvcName string,
) error {
	svc, err := k8sClient.GetService(ctx, namespace, stableSvcName)
	if err != nil {
		return fmt.Errorf("failed to get stable service %s/%s: %w", namespace, stableSvcName, err)
	}

	// 检查是否已经是 widened 状态
	if _, ok := svc.Spec.Selector["version"]; !ok {
		log.Printf("[CanaryIngress] Stable service %s/%s already widened (no version selector), skipping",
			namespace, stableSvcName)
		return nil
	}

	// 移除 version key：设为 null
	patchData := []byte(`{"spec":{"selector":{"version":null}}}`)

	_, err = k8sClient.PatchService(ctx, namespace, stableSvcName, patchData)
	if err != nil {
		// 如果 selector 中没有 version key 需要删除，忽略错误
		if strings.Contains(err.Error(), "version") {
			log.Printf("[CanaryIngress] Version key already absent in %s/%s, ignoring", namespace, stableSvcName)
			return nil
		}
		return fmt.Errorf("failed to widen stable service selector: %w", err)
	}

	log.Printf("[CanaryIngress] Widened stable service %s/%s selector (removed version constraint)",
		namespace, stableSvcName)
	return nil
}

func stringPtr(s string) *string {
	return &s
}
