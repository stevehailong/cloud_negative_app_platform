package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	commonModel "my-cloud/internal/common/model"
	deployModel "my-cloud/internal/deploy/model"
	"my-cloud/internal/deploy/repository"
	envRepo "my-cloud/internal/environment/repository"
	"my-cloud/pkg/helm"
	"my-cloud/pkg/k8s"
	"os"
	"path/filepath"
	"time"
)

// HelmDeployService 基于Helm的完整部署服务
type HelmDeployService struct {
	appDeployRepo *repository.AppDeploymentRepository
	historyRepo   *repository.DeploymentHistoryRepository
	envRepo       *envRepo.EnvironmentRepository
	templateRepo  *envRepo.EnvTemplateRepository
	k8sClient     *k8s.Client
	helmClient    *helm.Client
	chartPath     string // Helm Chart本地路径
}

// NewHelmDeployService 创建Helm部署服务
func NewHelmDeployService(
	appDeployRepo *repository.AppDeploymentRepository,
	historyRepo *repository.DeploymentHistoryRepository,
	envRepo *envRepo.EnvironmentRepository,
	templateRepo *envRepo.EnvTemplateRepository,
	k8sClient *k8s.Client,
	kubeconfig string,
) *HelmDeployService {
	// 获取Helm Chart路径（可以是本地路径或远程仓库）
	chartPath := os.Getenv("HELM_CHART_PATH")
	if chartPath == "" {
		// 默认使用项目内的Chart
		chartPath = "./helm-charts/mycloud-app"
	}

	return &HelmDeployService{
		appDeployRepo: appDeployRepo,
		historyRepo:   historyRepo,
		envRepo:       envRepo,
		templateRepo:  templateRepo,
		k8sClient:     k8sClient,
		helmClient:    helm.NewClient(kubeconfig),
		chartPath:     chartPath,
	}
}

// DeployWithHelm 使用Helm进行完整部署
func (s *HelmDeployService) DeployWithHelm(
	ctx context.Context,
	appDeployment *deployModel.AppDeployment,
	env *commonModel.Environment,
	appEnvBinding *commonModel.AppEnvBinding,
	image string,
) error {
	log.Printf("[HelmDeploy] Starting Helm deployment for app %d in env %d", appDeployment.AppID, appDeployment.EnvID)

	// 1. 获取环境模板
	var template *commonModel.EnvTemplate
	var err error
	if env.TemplateID != nil && *env.TemplateID > 0 {
		template, err = s.templateRepo.GetByID(uint(*env.TemplateID))
		if err != nil {
			log.Printf("[HelmDeploy] Warning: Failed to get template %d: %v", *env.TemplateID, err)
		}
	}

	// 2. 构建Helm Values
	values, err := s.buildHelmValues(appDeployment, env, appEnvBinding, template, image)
	if err != nil {
		return fmt.Errorf("failed to build helm values: %v", err)
	}

	// 3. 确保命名空间存在
	if s.k8sClient != nil {
		if err := s.k8sClient.EnsureNamespace(ctx, appDeployment.Namespace); err != nil {
			return fmt.Errorf("failed to ensure namespace: %v", err)
		}
	}

	// 4. 使用Helm部署
	releaseName := appDeployment.WorkloadName
	err = s.helmClient.InstallOrUpgrade(ctx, releaseName, appDeployment.Namespace, s.chartPath, values)
	if err != nil {
		return fmt.Errorf("helm install/upgrade failed: %v", err)
	}

	// 5. 等待部署完成
	err = s.helmClient.WaitForRelease(ctx, releaseName, appDeployment.Namespace, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("deployment rollout failed: %v", err)
	}

	log.Printf("[HelmDeploy] Successfully deployed %s in namespace %s", releaseName, appDeployment.Namespace)
	return nil
}

// buildHelmValues 构建Helm Values配置
func (s *HelmDeployService) buildHelmValues(
	appDeployment *deployModel.AppDeployment,
	env *commonModel.Environment,
	appEnvBinding *commonModel.AppEnvBinding,
	template *commonModel.EnvTemplate,
	image string,
) (map[string]interface{}, error) {
	// 使用ValuesBuilder构建配置
	builder := helm.NewValuesBuilder()

	// 构建部署配置
	config := helm.DeploymentConfig{
		AppName:      appDeployment.WorkloadName,
		Image:        image,
		Replicas:     appDeployment.DesiredReplicas,
		WorkloadName: appDeployment.WorkloadName,
	}

	// 从AppEnvBinding获取资源配置
	if appEnvBinding != nil {
		config.CPURequest = appEnvBinding.CPURequest
		config.CPULimit = appEnvBinding.CPULimit
		config.MemoryRequest = appEnvBinding.MemoryRequest
		config.MemoryLimit = appEnvBinding.MemoryLimit
	}

	// 从环境类型推断服务配置
	switch env.EnvType {
	case "dev", "development":
		config.ServiceType = "NodePort" // 开发环境使用NodePort方便调试
		config.IngressEnabled = false   // 开发环境可以不启用Ingress
		config.HPAEnabled = false
	case "test", "testing":
		config.ServiceType = "ClusterIP"
		config.IngressEnabled = true
		config.IngressHost = fmt.Sprintf("%s-test.example.com", appDeployment.WorkloadName)
		config.HPAEnabled = false
	case "staging":
		config.ServiceType = "ClusterIP"
		config.IngressEnabled = true
		config.IngressHost = fmt.Sprintf("%s-staging.example.com", appDeployment.WorkloadName)
		config.IngressTLSEnabled = true
		config.HPAEnabled = true
		config.HPAMinReplicas = 2
		config.HPAMaxReplicas = 5
	case "prod", "production":
		config.ServiceType = "ClusterIP"
		config.IngressEnabled = true
		config.IngressHost = fmt.Sprintf("%s.example.com", appDeployment.WorkloadName)
		config.IngressTLSEnabled = true
		config.HPAEnabled = true
		config.HPAMinReplicas = 3
		config.HPAMaxReplicas = 10
		config.HPATargetCPU = 70
	default:
		// 默认配置
		config.ServiceType = "ClusterIP"
		config.IngressEnabled = true
	}

	// 健康检查配置
	config.LivenessPath = "/health"
	config.ReadinessPath = "/ready"
	config.ContainerPort = 8080

	// 从AppEnvBinding的ConfigJSON获取额外配置
	if appEnvBinding != nil && appEnvBinding.ConfigJSON != "" {
		var extraConfig map[string]interface{}
		if err := json.Unmarshal([]byte(appEnvBinding.ConfigJSON), &extraConfig); err == nil {
			// 覆盖默认配置
			if v, ok := extraConfig["serviceType"].(string); ok {
				config.ServiceType = v
			}
			if v, ok := extraConfig["ingressEnabled"].(bool); ok {
				config.IngressEnabled = v
			}
			if v, ok := extraConfig["ingressHost"].(string); ok {
				config.IngressHost = v
			}
			if v, ok := extraConfig["containerPort"].(float64); ok {
				config.ContainerPort = int(v)
			}
			if v, ok := extraConfig["envVars"].(map[string]interface{}); ok {
				envVars := make(map[string]string)
				for k, val := range v {
					if str, ok := val.(string); ok {
						envVars[k] = str
					}
				}
				config.EnvVars = envVars
			}
		}
	}

	// 获取模板的ValuesYAML
	templateValues := ""
	if template != nil && template.ValuesYAML != "" {
		templateValues = template.ValuesYAML
	}

	// 构建最终的Values
	values, err := builder.BuildFromTemplate(templateValues, config)
	if err != nil {
		return nil, err
	}

	// 设置ServiceAccount
	builder.SetServiceAccount(true, fmt.Sprintf("%s-sa", appDeployment.WorkloadName))

	// 添加额外的标签
	labels := map[string]interface{}{
		"app":        appDeployment.WorkloadName,
		"env":        env.EnvType,
		"managed-by": "my-cloud",
	}
	values["labels"] = labels

	return values, nil
}

// UninstallWithHelm 使用Helm卸载应用
func (s *HelmDeployService) UninstallWithHelm(ctx context.Context, appDeployment *deployModel.AppDeployment) error {
	releaseName := appDeployment.WorkloadName

	err := s.helmClient.Uninstall(ctx, releaseName, appDeployment.Namespace)
	if err != nil {
		return fmt.Errorf("helm uninstall failed: %v", err)
	}

	log.Printf("[HelmDeploy] Successfully uninstalled %s from namespace %s", releaseName, appDeployment.Namespace)
	return nil
}

// GetDeploymentStatus 获取部署状态
func (s *HelmDeployService) GetDeploymentStatus(ctx context.Context, appDeployment *deployModel.AppDeployment) (*helm.ReleaseInfo, error) {
	releaseName := appDeployment.WorkloadName

	info, err := s.helmClient.GetRelease(ctx, releaseName, appDeployment.Namespace)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// ValidateChart 验证Helm Chart是否存在且有效
func (s *HelmDeployService) ValidateChart() error {
	// 检查Chart路径是否存在
	absPath, err := filepath.Abs(s.chartPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for chart: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("chart path does not exist: %s", absPath)
	}

	// 检查Chart.yaml是否存在
	chartYAML := filepath.Join(absPath, "Chart.yaml")
	if _, err := os.Stat(chartYAML); os.IsNotExist(err) {
		return fmt.Errorf("Chart.yaml not found in %s", absPath)
	}

	log.Printf("[HelmDeploy] Chart validated at %s", absPath)
	return nil
}
