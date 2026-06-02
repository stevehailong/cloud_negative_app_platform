package k8s

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strconv"

	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetConfig returns the underlying REST config
func (c *Client) GetConfig() *rest.Config {
	return c.config
}

// NewClientFromKubeconfig creates a K8s client from kubeconfig file path
func NewClientFromKubeconfig(kubeconfigPath string) (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}
	return newClient(config)
}

// NewClientInCluster creates a K8s client using in-cluster config
func NewClientInCluster() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	return newClient(config)
}

// NewClientFromAPIServer creates a K8s client from API server URL and bearer token
func NewClientFromAPIServer(apiServer, token, caCertPath string) (*Client, error) {
	config := &rest.Config{
		Host:        apiServer,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: caCertPath == "",
			CAFile:   caCertPath,
		},
	}
	return newClient(config)
}

func newClient(config *rest.Config) (*Client, error) {
	config.Timeout = 30 * time.Second
	// 跳过TLS验证（用于开发环境）
	config.TLSClientConfig.Insecure = true
	config.TLSClientConfig.CAFile = ""
	config.TLSClientConfig.CAData = nil
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return &Client{clientset: clientset, config: config}, nil
}

// CreateDeployment creates a Kubernetes Deployment
func (c *Client) CreateDeployment(ctx context.Context, namespace string, deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Create(ctx, deploy, metav1.CreateOptions{})
}

// GetDeployment gets a Kubernetes Deployment
func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateDeployment updates a Kubernetes Deployment
func (c *Client) UpdateDeployment(ctx context.Context, namespace string, deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
}

// UpdateDeploymentImage updates the image of a deployment
func (c *Client) UpdateDeploymentImage(ctx context.Context, namespace, name, imageURL string) error {
	// Get the deployment
	deployment, err := c.GetDeployment(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Update the image of the first container
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	deployment.Spec.Template.Spec.Containers[0].Image = imageURL

	// Update the deployment
	_, err = c.UpdateDeployment(ctx, namespace, deployment)
	return err
}

// DeleteDeployment deletes a Kubernetes Deployment
func (c *Client) DeleteDeployment(ctx context.Context, namespace, name string) error {
	return c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// EnsureService ensures a NodePort Service exists for an app
func (c *Client) EnsureService(ctx context.Context, namespace, serviceName, appLabel string, port, targetPort int32) (*corev1.Service, error) {
	// 尝试获取已存在的Service
	svc, err := c.clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err == nil {
		// Service已存在，检查selector是否正确
		if svc.Spec.Selector["app"] != appLabel {
			// 更新selector
			svc.Spec.Selector = map[string]string{
				"app":        appLabel,
				"managed-by": "my-cloud",
			}
			return c.clientset.CoreV1().Services(namespace).Update(ctx, svc, metav1.UpdateOptions{})
		}
		return svc, nil
	}

	// Service不存在，创建新的
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        appLabel,
				"managed-by": "my-cloud",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app":        appLabel,
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
	return c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
}

// GetService gets a Kubernetes Service
func (c *Client) GetService(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ScaleDeployment scales a Deployment to the specified replicas
func (c *Client) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	scale, err := c.clientset.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get scale: %w", err)
	}
	scale.Spec.Replicas = replicas
	_, err = c.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	return err
}

// RestartDeployment performs a rolling restart by updating an annotation
func (c *Client) RestartDeployment(ctx context.Context, namespace, name string) error {
	deploy, err := c.GetDeployment(ctx, namespace, name)
	if err != nil {
		return err
	}
	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	_, err = c.UpdateDeployment(ctx, namespace, deploy)
	return err
}

// GetPods returns pods matching label selector in a namespace
func (c *Client) GetPods(ctx context.Context, namespace, labelSelector string) ([]corev1.Pod, error) {
	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// DeletePod deletes a Pod from a namespace
func (c *Client) DeletePod(ctx context.Context, namespace, podName string) error {
	return c.clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}

// RolloutUndo rolls back a Deployment to the previous revision (equivalent to kubectl rollout undo)
func (c *Client) RolloutUndo(ctx context.Context, namespace, name string) error {
	// Get the ReplicaSet history to find the previous revision
	rsList, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=" + name,
	})
	if err != nil {
		return fmt.Errorf("failed to list replicasets: %w", err)
	}

	// Find the second-latest revision (previous)
	var prevRevision int64
	var prevRS *appsv1.ReplicaSet
	for i := range rsList.Items {
		rs := &rsList.Items[i]
		revStr := rs.Annotations["deployment.kubernetes.io/revision"]
		rev, _ := strconv.ParseInt(revStr, 10, 64)
		if rev > prevRevision && rs.Name != name {
			prevRevision = rev
			prevRS = rs
		}
	}

	if prevRS == nil || len(prevRS.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("no previous revision found for %s/%s", namespace, name)
	}

	// Patch the deployment with the previous pod template
	deploy, err := c.GetDeployment(ctx, namespace, name)
	if err != nil {
		return err
	}
	deploy.Spec.Template.Spec.Containers = prevRS.Spec.Template.Spec.Containers
	_, err = c.UpdateDeployment(ctx, namespace, deploy)
	return err
}

// GetEvents returns events for a specific resource
func (c *Client) GetEvents(ctx context.Context, namespace, name string) ([]corev1.Event, error) {
	eventList, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", name),
	})
	if err != nil {
		return nil, err
	}
	return eventList.Items, nil
}

// EnsureNamespace creates a namespace if it doesn't exist
func (c *Client) EnsureNamespace(ctx context.Context, name string) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"managed-by": "my-cloud",
			},
		},
	}
	_, err = c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	return err
}

// EnsureNetworkPolicy creates a default deny-all NetworkPolicy for a namespace
func (c *Client) EnsureNetworkPolicy(ctx context.Context, namespace string) error {
	netClient := c.clientset.NetworkingV1()
	policyName := "default-deny-all"
	
	// Check if policy exists
	_, err := netClient.NetworkPolicies(namespace).Get(ctx, policyName, metav1.GetOptions{})
	if err == nil {
		return nil // Already exists
	}
	
	// Create NetworkPolicy: deny all ingress by default, allow egress, allow from same namespace
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyName,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "my-cloud",
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{}, // Apply to all pods in namespace
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					// Allow traffic from same namespace
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
						},
					},
				},
				{
					// Allow traffic from ingress-nginx namespace (if exists)
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"name": "ingress-nginx",
								},
							},
						},
					},
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					// Allow all egress (DNS, external APIs, etc.)
					To: []networkingv1.NetworkPolicyPeer{},
				},
			},
		},
	}
	
	_, err = netClient.NetworkPolicies(namespace).Create(ctx, policy, metav1.CreateOptions{})
	return err
}

// EnsureResourceQuota creates a default ResourceQuota for a namespace
func (c *Client) EnsureResourceQuota(ctx context.Context, namespace string) error {
	quotaName := "default-quota"
	
	// Check if quota exists
	_, err := c.clientset.CoreV1().ResourceQuotas(namespace).Get(ctx, quotaName, metav1.GetOptions{})
	if err == nil {
		return nil // Already exists
	}
	
	// Create ResourceQuota with reasonable defaults
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      quotaName,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "my-cloud",
			},
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourcePods:             resource.MustParse("50"),      // Max 50 pods
				corev1.ResourceServices:         resource.MustParse("10"),      // Max 10 services
				corev1.ResourceRequestsCPU:      resource.MustParse("10"),      // Max 10 CPU cores requested
				corev1.ResourceRequestsMemory:   resource.MustParse("20Gi"),    // Max 20Gi memory requested
				corev1.ResourceLimitsCPU:        resource.MustParse("20"),      // Max 20 CPU cores limit
				corev1.ResourceLimitsMemory:     resource.MustParse("40Gi"),    // Max 40Gi memory limit
				corev1.ResourcePersistentVolumeClaims: resource.MustParse("10"), // Max 10 PVCs
			},
		},
	}
	
	_, err = c.clientset.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{})
	return err
}

// EnsureServiceAccount creates a ServiceAccount for the application
func (c *Client) EnsureServiceAccount(ctx context.Context, namespace, appName string) error {
	saName := appName + "-sa"
	
	// Check if ServiceAccount exists
	_, err := c.clientset.CoreV1().ServiceAccounts(namespace).Get(ctx, saName, metav1.GetOptions{})
	if err == nil {
		return nil // Already exists
	}
	
	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        appName,
				"managed-by": "my-cloud",
			},
		},
	}
	
	_, err = c.clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	
	// Create Role with minimal permissions
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName + "-role",
			Namespace: namespace,
			Labels: map[string]string{
				"app":        appName,
				"managed-by": "my-cloud",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get"},
			},
		},
	}
	
	_, err = c.clientset.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	
	// Create RoleBinding
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName + "-rolebinding",
			Namespace: namespace,
			Labels: map[string]string{
				"app":        appName,
				"managed-by": "my-cloud",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      saName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     appName + "-role",
		},
	}
	
	_, err = c.clientset.RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	return err
}

// GetNodes returns all cluster nodes
func (c *Client) GetNodes(ctx context.Context) ([]corev1.Node, error) {
	nodeList, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

// BuildDeploymentSpec creates a standard K8s Deployment spec
func BuildDeploymentSpec(name, namespace, image string, replicas int32, labels map[string]string) *appsv1.Deployment {
	// Extract app name for ServiceAccount
	appName := name
	if labels != nil {
		if app, ok := labels["app"]; ok {
			appName = app
		}
	}
	
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: appName + "-sa", // Use dedicated ServiceAccount
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}
