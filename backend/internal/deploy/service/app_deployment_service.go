package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	commonModel "my-cloud/internal/common/model"
	"my-cloud/internal/deploy/model"
	"my-cloud/internal/deploy/repository"
	envRepo "my-cloud/internal/environment/repository"
	"my-cloud/pkg/helm"
	"my-cloud/pkg/k8s"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type AppDeploymentService struct {
	appDeployRepo *repository.AppDeploymentRepository
	historyRepo   *repository.DeploymentHistoryRepository
	envRepo       *envRepo.EnvironmentRepository
	templateRepo  *envRepo.EnvTemplateRepository
	bindingRepo   *envRepo.AppEnvBindingRepository
	k8sClient     *k8s.Client
	helmClient    *helm.Client
	chartPath     string
}

func NewAppDeploymentService(
	appDeployRepo *repository.AppDeploymentRepository,
	historyRepo *repository.DeploymentHistoryRepository,
	envRepo *envRepo.EnvironmentRepository,
	templateRepo *envRepo.EnvTemplateRepository,
	bindingRepo *envRepo.AppEnvBindingRepository,
	k8sClient *k8s.Client,
) *AppDeploymentService {
	// 获取 Helm Chart 路径
	chartPath := os.Getenv("HELM_CHART_PATH")
	if chartPath == "" {
		chartPath = "./helm-charts/mycloud-app"
	}

	// 获取 kubeconfig 路径
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	return &AppDeploymentService{
		appDeployRepo: appDeployRepo,
		historyRepo:   historyRepo,
		envRepo:       envRepo,
		templateRepo:  templateRepo,
		bindingRepo:   bindingRepo,
		k8sClient:     k8sClient,
		helmClient:    helm.NewClient(kubeconfig),
		chartPath:     chartPath,
	}
}

// ListAppDeployments 查询应用部署列表
func (s *AppDeploymentService) ListAppDeployments(appID, envID *int64, page, pageSize int) ([]model.AppDeployment, int64, error) {
	return s.appDeployRepo.List(appID, envID, page, pageSize)
}

// GetAppDeploymentDetail 获取应用部署详情
func (s *AppDeploymentService) GetAppDeploymentDetail(id int64) (*model.AppDeployment, error) {
	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 从K8s同步最新状态
	if s.k8sClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		k8sDeploy, err := s.k8sClient.GetDeployment(ctx, deployment.Namespace, deployment.WorkloadName)
		if err == nil && k8sDeploy != nil {
			// 更新副本数
			deployment.DesiredReplicas = int(*k8sDeploy.Spec.Replicas)
			deployment.AvailableReplicas = int(k8sDeploy.Status.AvailableReplicas)

			// 更新状态
			if k8sDeploy.Status.AvailableReplicas == *k8sDeploy.Spec.Replicas {
				deployment.DeploymentStatus = "running"
			} else if k8sDeploy.Status.AvailableReplicas == 0 {
				deployment.DeploymentStatus = "stopped"
			} else {
				deployment.DeploymentStatus = "progressing"
			}

			// 保存到数据库
			_ = s.appDeployRepo.UpdateFields(deployment.ID, map[string]interface{}{
				"desired_replicas":   deployment.DesiredReplicas,
				"available_replicas": deployment.AvailableReplicas,
				"deployment_status":  deployment.DeploymentStatus,
			})
		}
	}

	return deployment, nil
}

// GetDeploymentHistory 获取部署历史
func (s *AppDeploymentService) GetDeploymentHistory(appDeploymentID int64, page, pageSize int) ([]model.DeploymentHistory, int64, error) {
	return s.historyRepo.ListByAppDeployment(appDeploymentID, page, pageSize)
}

// RestartDeployment 重启部署
func (s *AppDeploymentService) RestartDeployment(id int64, userID int64) error {
	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 创建历史记录
	startTime := time.Now()
	history := &model.DeploymentHistory{
		AppDeploymentID: deployment.ID,
		Version:         deployment.CurrentVersion,
		ImageURL:        deployment.CurrentImage,
		Replicas:        deployment.DesiredReplicas,
		DeploymentType:  "restart",
		OperatorUserID:  &userID,
		StartTime:       &startTime,
		Status:          "progressing",
		Changes: model.JSONMap{
			"action": "restart",
		},
	}

	if err := s.historyRepo.Create(history); err != nil {
		return err
	}

	// 异步执行K8s重启
	go s.executeRestart(deployment, history)

	return nil
}

// executeRestart 执行K8s重启
func (s *AppDeploymentService) executeRestart(deployment *model.AppDeployment, history *model.DeploymentHistory) {
	endTime := time.Now()
	duration := int(endTime.Sub(*history.StartTime).Seconds())

	if s.k8sClient == nil {
		log.Printf("[AppDeploy] No K8s client available, marking restart as failed for %s", deployment.WorkloadName)
		_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       endTime,
			"duration":       duration,
			"failure_reason": "K8s client not available",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 检查 Deployment 是否存在
	_, err := s.k8sClient.GetDeployment(ctx, deployment.Namespace, deployment.WorkloadName)
	if err != nil {
		// Deployment 不存在，无法重启
		log.Printf("[AppDeploy] Deployment %s/%s not found, cannot restart", deployment.Namespace, deployment.WorkloadName)
		_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       time.Now(),
			"duration":       int(time.Since(*history.StartTime).Seconds()),
			"failure_reason": "Deployment not found, please deploy first",
		})
		return
	}

	// K8s重启：给Deployment的Pod模板添加annotation触发滚动重启
	err = s.k8sClient.RestartDeployment(ctx, deployment.Namespace, deployment.WorkloadName)

	endTime = time.Now()
	duration = int(endTime.Sub(*history.StartTime).Seconds())

	if err != nil {
		log.Printf("[AppDeploy] Restart failed for %s/%s: %v", deployment.Namespace, deployment.WorkloadName, err)
		_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       endTime,
			"duration":       duration,
			"failure_reason": err.Error(),
		})
		return
	}

	// 更新历史记录为成功
	_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
		"status":   "success",
		"end_time": endTime,
		"duration": duration,
	})

	// 更新app_deployments的last_deploy_id和last_deploy_time
	_ = s.appDeployRepo.UpdateFields(deployment.ID, map[string]interface{}{
		"last_deploy_id":      history.ID,
		"last_deploy_time":    endTime,
		"last_deploy_user_id": history.OperatorUserID,
	})

	log.Printf("[AppDeploy] Restart completed for %s/%s", deployment.Namespace, deployment.WorkloadName)
}

// ScaleDeployment 扩缩容
func (s *AppDeploymentService) ScaleDeployment(id int64, replicas int, userID int64) error {
	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return err
	}

	if replicas < 0 {
		return errors.New("replicas must be >= 0")
	}

	oldReplicas := deployment.DesiredReplicas

	// 创建历史记录
	startTime := time.Now()
	history := &model.DeploymentHistory{
		AppDeploymentID: deployment.ID,
		Version:         deployment.CurrentVersion,
		ImageURL:        deployment.CurrentImage,
		Replicas:        replicas,
		DeploymentType:  "scale",
		OperatorUserID:  &userID,
		StartTime:       &startTime,
		Status:          "progressing",
		Changes: model.JSONMap{
			"action":       "scale",
			"old_replicas": oldReplicas,
			"new_replicas": replicas,
		},
	}

	if err := s.historyRepo.Create(history); err != nil {
		return err
	}

	// 异步执行K8s扩缩容
	go s.executeScale(deployment, history, replicas)

	return nil
}

// executeScale 执行K8s扩缩容
func (s *AppDeploymentService) executeScale(deployment *model.AppDeployment, history *model.DeploymentHistory, replicas int) {
	endTime := time.Now()
	duration := int(endTime.Sub(*history.StartTime).Seconds())

	if s.k8sClient == nil {
		log.Printf("[AppDeploy] No K8s client available, marking scale as failed for %s", deployment.WorkloadName)
		_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       endTime,
			"duration":       duration,
			"failure_reason": "K8s client not available",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 执行K8s扩缩容
	err := s.k8sClient.ScaleDeployment(ctx, deployment.Namespace, deployment.WorkloadName, int32(replicas))

	endTime = time.Now()
	duration = int(endTime.Sub(*history.StartTime).Seconds())

	if err != nil {
		log.Printf("[AppDeploy] Scale failed for %s/%s: %v", deployment.Namespace, deployment.WorkloadName, err)
		_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       endTime,
			"duration":       duration,
			"failure_reason": err.Error(),
		})
		return
	}

	// 等待3秒后查询实际副本数
	time.Sleep(3 * time.Second)
	k8sDeploy, err := s.k8sClient.GetDeployment(ctx, deployment.Namespace, deployment.WorkloadName)
	availableReplicas := 0
	if err == nil && k8sDeploy != nil {
		availableReplicas = int(k8sDeploy.Status.AvailableReplicas)
	}

	// 更新历史记录为成功
	_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
		"status":   "success",
		"end_time": endTime,
		"duration": duration,
	})

	// 更新app_deployments
	_ = s.appDeployRepo.UpdateFields(deployment.ID, map[string]interface{}{
		"desired_replicas":    replicas,
		"available_replicas":  availableReplicas,
		"last_deploy_id":      history.ID,
		"last_deploy_time":    endTime,
		"last_deploy_user_id": history.OperatorUserID,
	})

	log.Printf("[AppDeploy] Scale completed for %s/%s: %d -> %d replicas", deployment.Namespace, deployment.WorkloadName, history.Changes["old_replicas"], replicas)
}

// RollbackDeployment 回滚到历史版本
func (s *AppDeploymentService) RollbackDeployment(id int64, historyID int64, userID int64) error {
	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 获取要回滚的历史记录
	targetHistory, err := s.historyRepo.GetByID(historyID)
	if err != nil {
		return err
	}

	if targetHistory.AppDeploymentID != deployment.ID {
		return errors.New("history record does not belong to this deployment")
	}

	// 创建新的历史记录
	startTime := time.Now()
	newHistory := &model.DeploymentHistory{
		AppDeploymentID: deployment.ID,
		Version:         targetHistory.Version,
		ImageURL:        targetHistory.ImageURL,
		Replicas:        deployment.DesiredReplicas, // 保持当前副本数
		DeploymentType:  "rollback",
		OperatorUserID:  &userID,
		StartTime:       &startTime,
		Status:          "progressing",
		Changes: model.JSONMap{
			"action":            "rollback",
			"target_history_id": historyID,
			"old_version":       deployment.CurrentVersion,
			"old_image":         deployment.CurrentImage,
			"new_version":       targetHistory.Version,
			"new_image":         targetHistory.ImageURL,
		},
	}

	if err := s.historyRepo.Create(newHistory); err != nil {
		return err
	}

	// 异步执行K8s回滚
	go s.executeRollback(deployment, newHistory, targetHistory)

	return nil
}

// executeRollback 执行K8s回滚
func (s *AppDeploymentService) executeRollback(deployment *model.AppDeployment, newHistory *model.DeploymentHistory, targetHistory *model.DeploymentHistory) {
	endTime := time.Now()
	duration := int(endTime.Sub(*newHistory.StartTime).Seconds())

	if s.k8sClient == nil {
		log.Printf("[AppDeploy] No K8s client available, marking rollback as failed for %s", deployment.WorkloadName)
		_ = s.historyRepo.UpdateFields(newHistory.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       endTime,
			"duration":       duration,
			"failure_reason": "K8s client not available",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 检查 Deployment 是否存在
	_, err := s.k8sClient.GetDeployment(ctx, deployment.Namespace, deployment.WorkloadName)
	if err != nil {
		// Deployment 不存在，无法回滚
		log.Printf("[AppDeploy] Deployment %s/%s not found, cannot rollback", deployment.Namespace, deployment.WorkloadName)
		_ = s.historyRepo.UpdateFields(newHistory.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       time.Now(),
			"duration":       int(time.Since(*newHistory.StartTime).Seconds()),
			"failure_reason": "Deployment not found, please deploy first",
		})
		return
	}

	// 执行K8s镜像更新
	err = s.k8sClient.UpdateDeploymentImage(ctx, deployment.Namespace, deployment.WorkloadName, targetHistory.ImageURL)

	endTime = time.Now()
	duration = int(endTime.Sub(*newHistory.StartTime).Seconds())

	if err != nil {
		log.Printf("[AppDeploy] Rollback failed for %s/%s: %v", deployment.Namespace, deployment.WorkloadName, err)
		_ = s.historyRepo.UpdateFields(newHistory.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       endTime,
			"duration":       duration,
			"failure_reason": err.Error(),
		})
		return
	}

	// 更新历史记录为成功
	_ = s.historyRepo.UpdateFields(newHistory.ID, map[string]interface{}{
		"status":   "success",
		"end_time": endTime,
		"duration": duration,
	})

	// 更新app_deployments
	_ = s.appDeployRepo.UpdateFields(deployment.ID, map[string]interface{}{
		"current_version":     targetHistory.Version,
		"current_image":       targetHistory.ImageURL,
		"last_deploy_id":      newHistory.ID,
		"last_deploy_time":    endTime,
		"last_deploy_user_id": newHistory.OperatorUserID,
	})

	log.Printf("[AppDeploy] Rollback completed for %s/%s: %s -> %s", deployment.Namespace, deployment.WorkloadName, deployment.CurrentVersion, targetHistory.Version)
}

// DeployNewVersion 部署新版本
func (s *AppDeploymentService) DeployNewVersion(id int64, version, imageURL string, userID int64) (int64, error) {
	log.Printf("[AppDeploy] DeployNewVersion called: id=%d, version=%s, image=%s", id, version, imageURL)

	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return 0, err
	}

	log.Printf("[AppDeploy] Found deployment: namespace=%s, workload=%s", deployment.Namespace, deployment.WorkloadName)

	// 创建历史记录
	startTime := time.Now()
	history := &model.DeploymentHistory{
		AppDeploymentID: deployment.ID,
		Version:         version,
		ImageURL:        imageURL,
		Replicas:        deployment.DesiredReplicas,
		DeploymentType:  "update",
		OperatorUserID:  &userID,
		StartTime:       &startTime,
		Status:          "progressing",
		Changes: model.JSONMap{
			"action":      "deploy",
			"old_version": deployment.CurrentVersion,
			"old_image":   deployment.CurrentImage,
			"new_version": version,
			"new_image":   imageURL,
		},
	}

	if err := s.historyRepo.Create(history); err != nil {
		return 0, err
	}

	// 异步执行K8s部署
	log.Printf("[AppDeploy] Starting async executeDeploy for %s/%s", deployment.Namespace, deployment.WorkloadName)
	go s.executeDeploy(deployment, history, version, imageURL)

	return history.ID, nil
}

// executeDeploy 执行K8s部署(创建或更新)
func (s *AppDeploymentService) executeDeploy(deployment *model.AppDeployment, history *model.DeploymentHistory, version, imageURL string) {
	log.Printf("[AppDeploy] executeDeploy started for %s/%s", deployment.Namespace, deployment.WorkloadName)

	endTime := time.Now()
	duration := int(endTime.Sub(*history.StartTime).Seconds())

	if s.k8sClient == nil {
		log.Printf("[AppDeploy] No K8s client available, marking deploy as failed for %s", deployment.WorkloadName)
		_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       endTime,
			"duration":       duration,
			"failure_reason": "K8s client not available",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 检查 Deployment 是否存在
	existingDeploy, err := s.k8sClient.GetDeployment(ctx, deployment.Namespace, deployment.WorkloadName)

	var deployErr error
	if err != nil {
		// Deployment 不存在,需要创建
		log.Printf("[AppDeploy] Deployment %s/%s not found, creating new deployment", deployment.Namespace, deployment.WorkloadName)
		deployErr = s.createK8sDeployment(ctx, deployment, imageURL)
	} else {
		// Deployment 存在时也通过 Helm 重新渲染模板，确保 Ingress、Service、探针等配置同步更新
		log.Printf("[AppDeploy] Deployment %s/%s exists, applying deployment template", deployment.Namespace, deployment.WorkloadName)
		if s.helmClient != nil {
			deployErr = s.deployWithHelmChart(ctx, deployment, imageURL)
		} else {
			existingDeploy.Spec.Template.Spec.Containers[0].Image = imageURL
			_, deployErr = s.k8sClient.UpdateDeployment(ctx, deployment.Namespace, existingDeploy)
		}
	}

	endTime = time.Now()
	duration = int(endTime.Sub(*history.StartTime).Seconds())

	if deployErr != nil {
		log.Printf("[AppDeploy] Deploy failed for %s/%s: %v", deployment.Namespace, deployment.WorkloadName, deployErr)
		_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       endTime,
			"duration":       duration,
			"failure_reason": deployErr.Error(),
		})
		return
	}

	// 等待部署完成
	if !s.waitForDeploymentReady(ctx, deployment.Namespace, deployment.WorkloadName, deployment.DesiredReplicas) {
		log.Printf("[AppDeploy] Deployment %s/%s rollout timed out", deployment.Namespace, deployment.WorkloadName)
		_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
			"status":         "failed",
			"end_time":       time.Now(),
			"duration":       int(time.Since(*history.StartTime).Seconds()),
			"failure_reason": "deployment rollout timed out",
		})
		// 即使超时失败，也更新 current_image/version，避免数据库和 K8s 实际状态不一致
		_ = s.appDeployRepo.UpdateFields(deployment.ID, map[string]interface{}{
			"current_version":   version,
			"current_image":     imageURL,
			"last_deploy_id":    history.ID,
			"last_deploy_time":  time.Now(),
			"deployment_status": "failed",
		})
		return
	}

	endTime = time.Now()
	duration = int(endTime.Sub(*history.StartTime).Seconds())

	// 更新历史记录为成功
	_ = s.historyRepo.UpdateFields(history.ID, map[string]interface{}{
		"status":   "success",
		"end_time": endTime,
		"duration": duration,
	})

	// 更新app_deployments
	_ = s.appDeployRepo.UpdateFields(deployment.ID, map[string]interface{}{
		"current_version":     version,
		"current_image":       imageURL,
		"last_deploy_id":      history.ID,
		"last_deploy_time":    endTime,
		"last_deploy_user_id": history.OperatorUserID,
		"deployment_status":   "running",
	})

	log.Printf("[AppDeploy] Deploy completed for %s/%s: %s", deployment.Namespace, deployment.WorkloadName, version)
}

// createK8sDeployment 创建新的K8s Deployment（使用Helm完整部署）
func (s *AppDeploymentService) createK8sDeployment(ctx context.Context, deployment *model.AppDeployment, imageURL string) error {
	// 使用 Helm 进行完整部署
	if s.helmClient != nil {
		return s.deployWithHelmChart(ctx, deployment, imageURL)
	}

	// 降级：如果 Helm 客户端不可用，使用传统方式
	log.Printf("[AppDeploy] Helm client not available, using legacy deployment method")
	return s.createK8sDeploymentLegacy(ctx, deployment, imageURL)
}

// deployWithHelmChart 使用Helm Chart进行完整部署
func (s *AppDeploymentService) deployWithHelmChart(ctx context.Context, deployment *model.AppDeployment, imageURL string) error {
	log.Printf("[AppDeploy] Starting Helm deployment for %s/%s", deployment.Namespace, deployment.WorkloadName)

	// 1. 获取环境信息
	env, err := s.envRepo.GetByID(uint(deployment.EnvID))
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	// 2. 获取环境模板
	var template *commonModel.EnvTemplate
	if env.TemplateID != nil && *env.TemplateID > 0 {
		template, _ = s.templateRepo.GetByID(*env.TemplateID)
	}

	// 3. 获取应用环境绑定配置（资源配置）
	// TODO: 从 app_env_bindings 表查询资源配置

	// 4. 构建 Helm Values
	values, err := s.buildHelmValuesFromEnv(deployment, env, template, imageURL)
	if err != nil {
		return fmt.Errorf("failed to build helm values: %w", err)
	}

	// 5. 确保命名空间存在
	if s.k8sClient != nil {
		if err := s.k8sClient.EnsureNamespace(ctx, deployment.Namespace); err != nil {
			return fmt.Errorf("failed to ensure namespace: %w", err)
		}
	}

	// 6. 使用 Helm 部署
	releaseName := deployment.WorkloadName
	err = s.helmClient.InstallOrUpgrade(ctx, releaseName, deployment.Namespace, s.chartPath, values)
	if err != nil {
		return fmt.Errorf("helm install/upgrade failed: %w", err)
	}

	// 7. 等待部署完成
	err = s.helmClient.WaitForRelease(ctx, releaseName, deployment.Namespace, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("deployment rollout failed: %w", err)
	}

	log.Printf("[AppDeploy] Helm deployment completed successfully for %s/%s", deployment.Namespace, deployment.WorkloadName)
	return nil
}

// buildHelmValuesFromEnv 从环境定义构建Helm Values
func (s *AppDeploymentService) buildHelmValuesFromEnv(
	deployment *model.AppDeployment,
	env *commonModel.Environment,
	template *commonModel.EnvTemplate,
	imageURL string,
) (map[string]interface{}, error) {
	builder := helm.NewValuesBuilder()

	// 构建部署配置
	config := helm.DeploymentConfig{
		AppName:      deployment.WorkloadName,
		Image:        imageURL,
		Replicas:     deployment.DesiredReplicas,
		WorkloadName: deployment.WorkloadName,

		// 链路追踪：自动为部署应用注入 trace 环境变量
		TracingEnabled:     true,
		TracingEndpoint:    "http://monitor-service:8090/internal/v1/traces/spans",
		TracingServiceName: deployment.WorkloadName,
	}

	// 从环境类型推断服务配置
	switch env.EnvType {
	case "dev", "development":
		config.ServiceType = "NodePort"
		config.IngressEnabled = true
		config.IngressHost = fmt.Sprintf("%s.local", deployment.WorkloadName)
		config.HPAEnabled = false
	case "test", "testing":
		config.ServiceType = "ClusterIP"
		config.IngressEnabled = true
		config.IngressHost = fmt.Sprintf("%s-test.example.com", deployment.WorkloadName)
		config.HPAEnabled = false
	case "staging":
		config.ServiceType = "ClusterIP"
		config.IngressEnabled = true
		config.IngressHost = fmt.Sprintf("%s-staging.example.com", deployment.WorkloadName)
		config.IngressTLSEnabled = true
		config.HPAEnabled = true
		config.HPAMinReplicas = 2
		config.HPAMaxReplicas = 5
	case "prod", "production":
		config.ServiceType = "ClusterIP"
		config.IngressEnabled = true
		config.IngressHost = fmt.Sprintf("%s.example.com", deployment.WorkloadName)
		config.IngressTLSEnabled = true
		config.HPAEnabled = true
		config.HPAMinReplicas = 3
		config.HPAMaxReplicas = 10
		config.HPATargetCPU = 70
	default:
		config.ServiceType = "ClusterIP"
		config.IngressEnabled = true
	}

	// 健康检查配置
	config.LivenessPath = "/health"
	config.ReadinessPath = "/ready"
	config.ContainerPort = 8080
	config.ServicePort = 80

	// 从应用环境绑定配置获取端口、探针等覆盖项
	if s.bindingRepo != nil {
		if binding, err := s.bindingRepo.GetByAppAndEnv(uint(deployment.AppID), uint(deployment.EnvID)); err == nil && binding != nil {
			applyAppRuntimeConfig(&config, binding.ConfigJSON)
		}
	}

	// 从环境的 ConfigJSON 获取额外配置
	if env.ConfigJSON != "" {
		var envConfig map[string]interface{}
		if err := json.Unmarshal([]byte(env.ConfigJSON), &envConfig); err == nil {
			if v, ok := envConfig["ingressHost"].(string); ok {
				config.IngressHost = v
			}
			if v, ok := envConfig["ingressEnabled"].(bool); ok {
				config.IngressEnabled = v
			}
			if v, ok := envConfig["serviceType"].(string); ok {
				config.ServiceType = v
			}
			applyAppRuntimeConfigFromMap(&config, envConfig)
		}
	}

	ensurePortEnv(&config)

	// 从模板获取 ValuesYAML
	templateValues := ""
	if template != nil && template.ValuesYAML != "" {
		templateValues = template.ValuesYAML
	}

	// 构建最终的 Values
	values, err := builder.BuildFromTemplate(templateValues, config)
	if err != nil {
		return nil, err
	}

	// 让 Helm 资源名称与平台 workload_name 保持一致，避免生成 app-1-mycloud-app 这类派生名称
	values["nameOverride"] = deployment.WorkloadName
	values["fullnameOverride"] = deployment.WorkloadName

	// 设置 ServiceAccount
	builder.SetServiceAccount(true, fmt.Sprintf("%s-sa", deployment.WorkloadName))

	// 添加额外的标签
	labels := map[string]interface{}{
		"app":        deployment.WorkloadName,
		"env":        env.EnvType,
		"envCode":    env.EnvCode,
		"managed-by": "my-cloud",
	}
	values["labels"] = labels

	return values, nil
}

func applyAppRuntimeConfig(config *helm.DeploymentConfig, configJSON string) {
	if configJSON == "" {
		return
	}
	var extraConfig map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &extraConfig); err != nil {
		return
	}
	applyAppRuntimeConfigFromMap(config, extraConfig)
}

func applyAppRuntimeConfigFromMap(config *helm.DeploymentConfig, extraConfig map[string]interface{}) {
	if v, ok := getIntConfig(extraConfig, "containerPort", "targetPort", "appPort", "port"); ok && v > 0 {
		config.ContainerPort = v
	}
	if v, ok := getIntConfig(extraConfig, "servicePort"); ok && v > 0 {
		config.ServicePort = v
	}
	if healthCheck, ok := extraConfig["healthCheck"].(map[string]interface{}); ok {
		if v, ok := getIntConfig(healthCheck, "port", "containerPort", "targetPort"); ok && v > 0 {
			config.ContainerPort = v
		}
		if v, ok := healthCheck["path"].(string); ok && v != "" {
			config.LivenessPath = v
			if config.ReadinessPath == "" || config.ReadinessPath == "/ready" {
				config.ReadinessPath = v
			}
		}
		if v, ok := healthCheck["livenessPath"].(string); ok && v != "" {
			config.LivenessPath = v
		}
		if v, ok := healthCheck["readinessPath"].(string); ok && v != "" {
			config.ReadinessPath = v
		}
	}
	if v, ok := extraConfig["ingressEnabled"].(bool); ok {
		config.IngressEnabled = v
	}
	if v, ok := extraConfig["ingressHost"].(string); ok && v != "" {
		config.IngressHost = v
	}
	if v, ok := extraConfig["ingressPath"].(string); ok && v != "" {
		config.IngressPath = v
	}
	if v, ok := extraConfig["livenessPath"].(string); ok && v != "" {
		config.LivenessPath = v
	}
	if v, ok := extraConfig["readinessPath"].(string); ok && v != "" {
		config.ReadinessPath = v
	}
	if v, ok := extraConfig["healthPath"].(string); ok && v != "" {
		config.LivenessPath = v
		if config.ReadinessPath == "" || config.ReadinessPath == "/ready" {
			config.ReadinessPath = v
		}
	}
	if envVars := parseEnvVars(extraConfig["envVars"]); len(envVars) > 0 {
		config.EnvVars = envVars
	}
}

func parseEnvVars(value interface{}) map[string]string {
	envVars := make(map[string]string)
	switch vars := value.(type) {
	case map[string]interface{}:
		for key, val := range vars {
			if str, ok := val.(string); ok {
				envVars[key] = str
			}
		}
	case []interface{}:
		for _, item := range vars {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			key, _ := itemMap["key"].(string)
			if key == "" {
				key, _ = itemMap["name"].(string)
			}
			val, _ := itemMap["value"].(string)
			if key != "" {
				envVars[key] = val
			}
		}
	}
	return envVars
}

func getIntConfig(config map[string]interface{}, keys ...string) (int, bool) {
	for _, key := range keys {
		value, exists := config[key]
		if !exists {
			continue
		}
		switch v := value.(type) {
		case float64:
			return int(v), true
		case int:
			return v, true
		case string:
			var parsed int
			if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func ensurePortEnv(config *helm.DeploymentConfig) {
	if config.ContainerPort <= 0 {
		config.ContainerPort = 8080
	}
	if config.EnvVars == nil {
		config.EnvVars = make(map[string]string)
	}
	if _, exists := config.EnvVars["PORT"]; !exists {
		config.EnvVars["PORT"] = fmt.Sprintf("%d", config.ContainerPort)
	}
}

// createK8sDeploymentLegacy 传统部署方式（降级方案）
func (s *AppDeploymentService) createK8sDeploymentLegacy(ctx context.Context, deployment *model.AppDeployment, imageURL string) error {
	// 确保namespace存在
	if err := s.k8sClient.EnsureNamespace(ctx, deployment.Namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// 获取模版配置
	var templateValues map[string]interface{}
	if deployment.EnvID > 0 {
		env, err := s.envRepo.GetByID(uint(deployment.EnvID))
		if err == nil && env != nil && env.TemplateID != nil {
			templateValues, _ = s.getTemplateValues(*env.TemplateID)
			if templateValues != nil {
				log.Printf("[AppDeploy] Using template %d for deployment configuration", *env.TemplateID)
			}
		}
	}

	// 构建labels
	isCanary := strings.HasSuffix(deployment.WorkloadName, "-canary")
	appName := deployment.WorkloadName
	if isCanary {
		appName = strings.TrimSuffix(appName, "-canary")
	}

	labels := map[string]string{
		"app":        appName,
		"version":    deployment.WorkloadName,
		"managed-by": "my-cloud",
	}

	// 确保Service存在
	serviceName := fmt.Sprintf("%s-service", appName)
	targetPort := 8080
	if templateValues != nil {
		if service, ok := templateValues["service"].(map[string]interface{}); ok {
			if port, ok := service["targetPort"].(int); ok {
				targetPort = port
			}
		}
	}
	if _, err := s.k8sClient.EnsureService(ctx, deployment.Namespace, serviceName, appName, 80, int32(targetPort)); err != nil {
		log.Printf("[AppDeploy] Warning: Failed to ensure service: %v", err)
	}

	// 确保ServiceAccount存在
	if err := s.k8sClient.EnsureServiceAccount(ctx, deployment.Namespace, appName); err != nil {
		log.Printf("[AppDeploy] Warning: Failed to ensure service account: %v", err)
	}

	// 构建Deployment spec
	replicas := int32(deployment.DesiredReplicas)
	if replicas <= 0 {
		// 尝试从模版获取副本数
		if templateValues != nil {
			if rc, ok := templateValues["replicaCount"].(int); ok && rc > 0 {
				replicas = int32(rc)
			} else {
				replicas = 1
			}
		} else {
			replicas = 1
		}
	}

	k8sDeploy := k8s.BuildDeploymentSpec(
		deployment.WorkloadName,
		deployment.Namespace,
		imageURL,
		replicas,
		labels,
	)

	// 应用模版配置
	if templateValues != nil {
		// 应用资源配置
		if resources, ok := templateValues["resources"].(map[string]interface{}); ok {
			if limits, ok := resources["limits"].(map[string]interface{}); ok {
				if cpu, ok := limits["cpu"].(string); ok {
					k8sDeploy.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"] = resource.MustParse(cpu)
				}
				if mem, ok := limits["memory"].(string); ok {
					k8sDeploy.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] = resource.MustParse(mem)
				}
			}
			if requests, ok := resources["requests"].(map[string]interface{}); ok {
				if cpu, ok := requests["cpu"].(string); ok {
					k8sDeploy.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] = resource.MustParse(cpu)
				}
				if mem, ok := requests["memory"].(string); ok {
					k8sDeploy.Spec.Template.Spec.Containers[0].Resources.Requests["memory"] = resource.MustParse(mem)
				}
			}
		}

		// 应用环境变量
		if envList, ok := templateValues["env"].([]interface{}); ok {
			for _, envItem := range envList {
				if envMap, ok := envItem.(map[string]interface{}); ok {
					name, _ := envMap["name"].(string)
					value, _ := envMap["value"].(string)
					if name != "" {
						k8sDeploy.Spec.Template.Spec.Containers[0].Env = append(
							k8sDeploy.Spec.Template.Spec.Containers[0].Env,
							corev1.EnvVar{Name: name, Value: value},
						)
					}
				}
			}
		}

		// 应用健康检查配置
		if liveness, ok := templateValues["livenessProbe"].(map[string]interface{}); ok {
			if enabled, ok := liveness["enabled"].(bool); ok && enabled {
				if httpGet, ok := liveness["httpGet"].(map[string]interface{}); ok {
					path, _ := httpGet["path"].(string)
					port, _ := httpGet["port"].(int)
					initialDelay, _ := liveness["initialDelaySeconds"].(int)
					period, _ := liveness["periodSeconds"].(int)

					k8sDeploy.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: path,
								Port: intstr.FromInt(port),
							},
						},
						InitialDelaySeconds: int32(initialDelay),
						PeriodSeconds:       int32(period),
					}
				}
			}
		}

		if readiness, ok := templateValues["readinessProbe"].(map[string]interface{}); ok {
			if enabled, ok := readiness["enabled"].(bool); ok && enabled {
				if httpGet, ok := readiness["httpGet"].(map[string]interface{}); ok {
					path, _ := httpGet["path"].(string)
					port, _ := httpGet["port"].(int)
					initialDelay, _ := readiness["initialDelaySeconds"].(int)
					period, _ := readiness["periodSeconds"].(int)

					k8sDeploy.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: path,
								Port: intstr.FromInt(port),
							},
						},
						InitialDelaySeconds: int32(initialDelay),
						PeriodSeconds:       int32(period),
					}
				}
			}
		}

		log.Printf("[AppDeploy] Applied template configuration: replicas=%d, envs=%d", replicas, len(k8sDeploy.Spec.Template.Spec.Containers[0].Env))
	}

	// 创建Deployment
	_, err := s.k8sClient.CreateDeployment(ctx, deployment.Namespace, k8sDeploy)
	return err
}

// waitForDeploymentReady 等待Deployment就绪
func (s *AppDeploymentService) waitForDeploymentReady(ctx context.Context, namespace, name string, desiredReplicas int) bool {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeout := time.After(90 * time.Second)

	for {
		select {
		case <-timeout:
			return false
		case <-ctx.Done():
			return false
		case <-ticker.C:
			k8sDeploy, err := s.k8sClient.GetDeployment(ctx, namespace, name)
			if err != nil {
				continue
			}

			available := int(k8sDeploy.Status.AvailableReplicas)
			desired := int(*k8sDeploy.Spec.Replicas)

			if available >= desired && k8sDeploy.Status.ReadyReplicas >= int32(desired) {
				log.Printf("[AppDeploy] Deployment %s/%s ready (%d/%d)", namespace, name, available, desired)
				return true
			}
		}
	}
}

// GetOrCreateAppDeployment 获取或创建应用部署记录（用于CI集成）
// namespace和clusterID会从环境自动获取，确保命名空间隔离
func (s *AppDeploymentService) GetOrCreateAppDeployment(appID, envID int64, workloadName string) (*model.AppDeployment, error) {
	// 先尝试获取
	deployment, err := s.appDeployRepo.GetByAppAndEnv(appID, envID)
	if err == nil {
		return deployment, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 从环境获取namespace和cluster_id
	env, err := s.envRepo.GetByID(uint(envID))
	if err != nil {
		return nil, fmt.Errorf("环境不存在: %w", err)
	}

	// 生成应用专属命名空间: app-{appId}-{envNamespace}
	appNamespace := fmt.Sprintf("app-%d-%s", appID, env.Namespace)

	// 不存在则创建
	deployment = &model.AppDeployment{
		AppID:             appID,
		EnvID:             envID,
		ClusterID:         int64(env.ClusterID),
		Namespace:         appNamespace,
		WorkloadName:      workloadName,
		WorkloadType:      "deployment",
		DeploymentStatus:  "created",
		DesiredReplicas:   1,
		AvailableReplicas: 0,
	}

	if err := s.appDeployRepo.Create(deployment); err != nil {
		return nil, err
	}

	log.Printf("[AppDeploy] Created deployment: app=%d, env=%d, namespace=%s", appID, envID, appNamespace)
	return deployment, nil
}

// SyncFromK8s 从K8s同步状态到数据库
func (s *AppDeploymentService) SyncFromK8s(id int64) error {
	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return err
	}

	if s.k8sClient == nil {
		return errors.New("k8s client not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	k8sDeploy, err := s.k8sClient.GetDeployment(ctx, deployment.Namespace, deployment.WorkloadName)
	if err != nil {
		return fmt.Errorf("failed to get k8s deployment: %w", err)
	}

	// 提取镜像信息
	var imageURL string
	if len(k8sDeploy.Spec.Template.Spec.Containers) > 0 {
		imageURL = k8sDeploy.Spec.Template.Spec.Containers[0].Image
	}

	// 更新数据库
	updates := map[string]interface{}{
		"current_image":      imageURL,
		"desired_replicas":   int(*k8sDeploy.Spec.Replicas),
		"available_replicas": int(k8sDeploy.Status.AvailableReplicas),
	}

	if k8sDeploy.Status.AvailableReplicas == *k8sDeploy.Spec.Replicas {
		updates["deployment_status"] = "running"
	} else if k8sDeploy.Status.AvailableReplicas == 0 {
		updates["deployment_status"] = "stopped"
	} else {
		updates["deployment_status"] = "progressing"
	}

	return s.appDeployRepo.UpdateFields(deployment.ID, updates)
}

// RecordDeploymentFromRelease 从Release记录部署历史（用于CI集成）
func (s *AppDeploymentService) RecordDeploymentFromRelease(appDeploymentID, releaseID int64, version, imageURL string, replicas int, userID int64) (*model.DeploymentHistory, error) {
	startTime := time.Now()

	changesData := map[string]interface{}{
		"source":      "ci",
		"release_id":  releaseID,
		"auto_deploy": true,
	}
	changesJSON, _ := json.Marshal(changesData)

	var changes model.JSONMap
	_ = json.Unmarshal(changesJSON, &changes)

	history := &model.DeploymentHistory{
		AppDeploymentID: appDeploymentID,
		ReleaseID:       &releaseID,
		Version:         version,
		ImageURL:        imageURL,
		Replicas:        replicas,
		DeploymentType:  "update",
		OperatorUserID:  &userID,
		StartTime:       &startTime,
		Status:          "progressing",
		Changes:         changes,
	}

	if err := s.historyRepo.Create(history); err != nil {
		return nil, err
	}

	return history, nil
}

// UpdateDeploymentHistoryStatus 更新部署历史状态
func (s *AppDeploymentService) UpdateDeploymentHistoryStatus(historyID int64, status string, failureReason string) error {
	endTime := time.Now()

	history, err := s.historyRepo.GetByID(historyID)
	if err != nil {
		return err
	}

	duration := 0
	if history.StartTime != nil {
		duration = int(endTime.Sub(*history.StartTime).Seconds())
	}

	updates := map[string]interface{}{
		"status":   status,
		"end_time": endTime,
		"duration": duration,
	}

	if failureReason != "" {
		updates["failure_reason"] = failureReason
	}

	return s.historyRepo.UpdateFields(historyID, updates)
}

// GetDeploymentPods 获取部署的Pod列表
func (s *AppDeploymentService) GetDeploymentPods(id int64) ([]map[string]interface{}, error) {
	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if s.k8sClient == nil {
		return nil, errors.New("k8s client not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 优先使用 Helm 标准标签查询（app.kubernetes.io/instance=<workload_name>）
	// Helm chart 通过 _helpers.tpl 设置了这些标签
	helmSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", deployment.WorkloadName)
	pods, err := s.k8sClient.GetPods(ctx, deployment.Namespace, helmSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	// 如果 Helm 标签没找到 Pod，回退到旧版标签（version=<workload_name>,managed-by=my-cloud）
	if len(pods) == 0 {
		legacySelector := fmt.Sprintf("version=%s,managed-by=my-cloud", deployment.WorkloadName)
		pods, err = s.k8sClient.GetPods(ctx, deployment.Namespace, legacySelector)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods: %w", err)
		}
	}

	// 转换为简化格式
	result := make([]map[string]interface{}, 0, len(pods))
	for _, pod := range pods {
		podInfo := map[string]interface{}{
			"name":       pod.Name,
			"namespace":  pod.Namespace,
			"status":     string(pod.Status.Phase),
			"node":       pod.Spec.NodeName,
			"pod_ip":     pod.Status.PodIP,
			"host_ip":    pod.Status.HostIP,
			"start_time": pod.Status.StartTime,
			"restarts":   0,
			"version":    deployment.WorkloadName, // 标识是stable还是canary
		}

		// 计算重启次数
		for _, containerStatus := range pod.Status.ContainerStatuses {
			podInfo["restarts"] = int(containerStatus.RestartCount)
			podInfo["ready"] = containerStatus.Ready
			break // 只取第一个容器
		}

		result = append(result, podInfo)
	}

	return result, nil
}

// GetDeploymentEvents 获取部署的事件列表
func (s *AppDeploymentService) GetDeploymentEvents(id int64) ([]map[string]interface{}, error) {
	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if s.k8sClient == nil {
		return nil, errors.New("k8s client not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取Deployment的事件
	events, err := s.k8sClient.GetEvents(ctx, deployment.Namespace, deployment.WorkloadName)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// 转换为简化格式
	result := make([]map[string]interface{}, 0, len(events))
	for _, event := range events {
		eventInfo := map[string]interface{}{
			"type":            event.Type,
			"reason":          event.Reason,
			"message":         event.Message,
			"count":           event.Count,
			"first_timestamp": event.FirstTimestamp.Time,
			"last_timestamp":  event.LastTimestamp.Time,
			"source":          event.Source.Component,
		}
		result = append(result, eventInfo)
	}

	return result, nil
}

// GetByWorkloadName 根据workload_name查询app_deployment
func (s *AppDeploymentService) GetByWorkloadName(namespace, workloadName string) (*model.AppDeployment, error) {
	return s.appDeployRepo.GetByWorkloadName(namespace, workloadName)
}

// CreateAppDeployment 创建app_deployment记录
// namespace和clusterID会从环境自动获取，确保命名空间隔离
func (s *AppDeploymentService) CreateAppDeployment(appID, envID int64, workloadName, workloadType string, desiredReplicas int) (*model.AppDeployment, error) {
	// 从环境获取namespace和cluster_id
	env, err := s.envRepo.GetByID(uint(envID))
	if err != nil {
		return nil, fmt.Errorf("环境不存在: %w", err)
	}

	// 生成应用专属命名空间: app-{appId}-{envNamespace}
	appNamespace := fmt.Sprintf("app-%d-%s", appID, env.Namespace)

	if workloadType == "" {
		workloadType = "deployment"
	}
	if desiredReplicas <= 0 {
		desiredReplicas = 1
	}

	deployment := &model.AppDeployment{
		AppID:            appID,
		EnvID:            envID,
		ClusterID:        int64(env.ClusterID),
		Namespace:        appNamespace,
		WorkloadName:     workloadName,
		WorkloadType:     workloadType,
		DesiredReplicas:  desiredReplicas,
		DeploymentStatus: "created",
	}

	if err := s.appDeployRepo.Create(deployment); err != nil {
		return nil, err
	}

	log.Printf("[AppDeploy] Created deployment: app=%d, env=%d, namespace=%s", appID, envID, appNamespace)
	return deployment, nil
}

// GetDeploymentHistoryByID 根据ID查询单条deployment_history记录
func (s *AppDeploymentService) GetDeploymentHistoryByID(id int64) (*model.DeploymentHistory, error) {
	return s.historyRepo.GetByID(id)
}

// ListByAppAndEnv 查询应用在指定环境的所有部署(包括stable和canary)
func (s *AppDeploymentService) ListByAppAndEnv(appID, envID int64) ([]model.AppDeployment, error) {
	return s.appDeployRepo.ListByAppAndEnv(appID, envID)
}

// CleanupDuplicateDeployments 清理不合理的重复部署记录
// 保留规则: 每个app+env只保留stable(app-{id})和canary(app-{id}-canary)各一条
func (s *AppDeploymentService) CleanupDuplicateDeployments(appID, envID int64) (int, error) {
	deployments, err := s.appDeployRepo.ListByAppAndEnv(appID, envID)
	if err != nil {
		return 0, err
	}

	if len(deployments) <= 2 {
		return 0, nil // 最多2条记录(stable+canary),无需清理
	}

	expectedStable := fmt.Sprintf("app-%d", appID)
	expectedCanary := fmt.Sprintf("app-%d-canary", appID)

	var stableDeployment, canaryDeployment *model.AppDeployment
	var toDelete []int64

	for i := range deployments {
		d := &deployments[i]
		if d.WorkloadName == expectedStable {
			if stableDeployment == nil {
				stableDeployment = d
			} else {
				// 重复的stable记录,保留最新的
				if d.UpdateTime.After(stableDeployment.UpdateTime) {
					toDelete = append(toDelete, stableDeployment.ID)
					stableDeployment = d
				} else {
					toDelete = append(toDelete, d.ID)
				}
			}
		} else if d.WorkloadName == expectedCanary {
			if canaryDeployment == nil {
				canaryDeployment = d
			} else {
				// 重复的canary记录,保留最新的
				if d.UpdateTime.After(canaryDeployment.UpdateTime) {
					toDelete = append(toDelete, canaryDeployment.ID)
					canaryDeployment = d
				} else {
					toDelete = append(toDelete, d.ID)
				}
			}
		} else {
			// 不符合命名规范的记录,标记删除
			toDelete = append(toDelete, d.ID)
		}
	}

	// 执行删除
	deletedCount := 0
	for _, id := range toDelete {
		if err := s.appDeployRepo.Delete(id); err != nil {
			log.Printf("[AppDeploy] Failed to delete deployment %d: %v", id, err)
		} else {
			deletedCount++
			log.Printf("[AppDeploy] Deleted duplicate/invalid deployment record: id=%d", id)
		}
	}

	return deletedCount, nil
}

// PromoteCanaryToStable 将金丝雀版本提升为稳定版本并删除canary记录
func (s *AppDeploymentService) PromoteCanaryToStable(appID, envID int64, userID int64) error {
	deployments, err := s.appDeployRepo.ListByAppAndEnv(appID, envID)
	if err != nil {
		return err
	}

	var stableDeployment, canaryDeployment *model.AppDeployment
	for i := range deployments {
		d := &deployments[i]
		if strings.HasSuffix(d.WorkloadName, "-canary") {
			canaryDeployment = d
		} else {
			stableDeployment = d
		}
	}

	if canaryDeployment == nil {
		return errors.New("canary deployment not found")
	}
	if stableDeployment == nil {
		return errors.New("stable deployment not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 1. 更新stable版本的镜像为canary的镜像
	if s.k8sClient != nil {
		err = s.k8sClient.UpdateDeploymentImage(ctx, stableDeployment.Namespace, stableDeployment.WorkloadName, canaryDeployment.CurrentImage)
		if err != nil {
			return fmt.Errorf("failed to update stable deployment image: %w", err)
		}
		log.Printf("[AppDeploy] Updated stable deployment %s to image %s", stableDeployment.WorkloadName, canaryDeployment.CurrentImage)
	}

	// 2. 更新stable记录的版本信息
	_ = s.appDeployRepo.UpdateFields(stableDeployment.ID, map[string]interface{}{
		"current_version": canaryDeployment.CurrentVersion,
		"current_image":   canaryDeployment.CurrentImage,
	})

	// 3. 删除K8s中的canary Deployment
	if s.k8sClient != nil {
		err = s.k8sClient.DeleteDeployment(ctx, canaryDeployment.Namespace, canaryDeployment.WorkloadName)
		if err != nil {
			log.Printf("[AppDeploy] Warning: failed to delete canary deployment from K8s: %v", err)
		} else {
			log.Printf("[AppDeploy] Deleted canary deployment %s from K8s", canaryDeployment.WorkloadName)
		}
	}

	// 4. 删除数据库中的canary记录
	if err := s.appDeployRepo.Delete(canaryDeployment.ID); err != nil {
		return fmt.Errorf("failed to delete canary deployment record: %w", err)
	}

	log.Printf("[AppDeploy] Promoted canary to stable for app %d env %d, deleted canary record", appID, envID)
	return nil
}

// DeleteAppDeployment 删除应用部署记录
func (s *AppDeploymentService) DeleteAppDeployment(id int64) error {
	return s.appDeployRepo.Delete(id)
}

// GetEnvironmentByID 获取环境信息（供release-service调用）
func (s *AppDeploymentService) GetEnvironmentByID(envID uint) (*commonModel.Environment, error) {
	return s.envRepo.GetByID(envID)
}

// getTemplateValues 获取模版配置
func (s *AppDeploymentService) getTemplateValues(templateID uint) (map[string]interface{}, error) {
	template, err := s.templateRepo.GetByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("获取模版失败: %w", err)
	}

	var values map[string]interface{}
	if err := yaml.Unmarshal([]byte(template.ValuesYAML), &values); err != nil {
		return nil, fmt.Errorf("解析模版配置失败: %w", err)
	}

	return values, nil
}
