package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"my-cloud/internal/deploy/model"
	"my-cloud/internal/deploy/repository"
	"my-cloud/pkg/k8s"
	"time"

	"gorm.io/gorm"
)

type AppDeploymentService struct {
	appDeployRepo *repository.AppDeploymentRepository
	historyRepo   *repository.DeploymentHistoryRepository
	k8sClient     *k8s.Client
}

func NewAppDeploymentService(
	appDeployRepo *repository.AppDeploymentRepository,
	historyRepo *repository.DeploymentHistoryRepository,
	k8sClient *k8s.Client,
) *AppDeploymentService {
	return &AppDeploymentService{
		appDeployRepo: appDeployRepo,
		historyRepo:   historyRepo,
		k8sClient:     k8sClient,
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

	// K8s重启：给Deployment的Pod模板添加annotation触发滚动重启
	err := s.k8sClient.RestartDeployment(ctx, deployment.Namespace, deployment.WorkloadName)
	
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

	// 执行K8s镜像更新
	err := s.k8sClient.UpdateDeploymentImage(ctx, deployment.Namespace, deployment.WorkloadName, targetHistory.ImageURL)
	
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
	deployment, err := s.appDeployRepo.GetByID(id)
	if err != nil {
		return 0, err
	}

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
	go s.executeDeploy(deployment, history, version, imageURL)

	return history.ID, nil
}

// executeDeploy 执行K8s部署
func (s *AppDeploymentService) executeDeploy(deployment *model.AppDeployment, history *model.DeploymentHistory, version, imageURL string) {
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

	// 执行K8s镜像更新
	err := s.k8sClient.UpdateDeploymentImage(ctx, deployment.Namespace, deployment.WorkloadName, imageURL)
	
	endTime = time.Now()
	duration = int(endTime.Sub(*history.StartTime).Seconds())

	if err != nil {
		log.Printf("[AppDeploy] Deploy failed for %s/%s: %v", deployment.Namespace, deployment.WorkloadName, err)
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

	// 更新app_deployments
	_ = s.appDeployRepo.UpdateFields(deployment.ID, map[string]interface{}{
		"current_version":     version,
		"current_image":       imageURL,
		"last_deploy_id":      history.ID,
		"last_deploy_time":    endTime,
		"last_deploy_user_id": history.OperatorUserID,
	})

	log.Printf("[AppDeploy] Deploy completed for %s/%s: %s", deployment.Namespace, deployment.WorkloadName, version)
}

// GetOrCreateAppDeployment 获取或创建应用部署记录（用于CI集成）
func (s *AppDeploymentService) GetOrCreateAppDeployment(appID, envID, clusterID int64, namespace, workloadName string) (*model.AppDeployment, error) {
	// 先尝试获取
	deployment, err := s.appDeployRepo.GetByAppAndEnv(appID, envID)
	if err == nil {
		return deployment, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 不存在则创建
	deployment = &model.AppDeployment{
		AppID:             appID,
		EnvID:             envID,
		ClusterID:         clusterID,
		Namespace:         namespace,
		WorkloadName:      workloadName,
		WorkloadType:      "deployment",
		DeploymentStatus:  "created",
		DesiredReplicas:   1,
		AvailableReplicas: 0,
	}

	if err := s.appDeployRepo.Create(deployment); err != nil {
		return nil, err
	}

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

	// 使用label selector查询Pod
	labelSelector := fmt.Sprintf("app=%s,managed-by=my-cloud", deployment.WorkloadName)
	pods, err := s.k8sClient.GetPods(ctx, deployment.Namespace, labelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
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
			"type":    event.Type,
			"reason":  event.Reason,
			"message": event.Message,
			"count":   event.Count,
			"first_timestamp": event.FirstTimestamp.Time,
			"last_timestamp":  event.LastTimestamp.Time,
			"source":  event.Source.Component,
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
func (s *AppDeploymentService) CreateAppDeployment(appID, envID, clusterID int64, namespace, workloadName, workloadType string, desiredReplicas int) (*model.AppDeployment, error) {
	deployment := &model.AppDeployment{
		AppID:           appID,
		EnvID:           envID,
		ClusterID:       clusterID,
		Namespace:       namespace,
		WorkloadName:    workloadName,
		WorkloadType:    workloadType,
		DesiredReplicas: desiredReplicas,
		DeploymentStatus: "created",
	}

	if err := s.appDeployRepo.Create(deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

// GetDeploymentHistoryByID 根据ID查询单条deployment_history记录
func (s *AppDeploymentService) GetDeploymentHistoryByID(id int64) (*model.DeploymentHistory, error) {
	return s.historyRepo.GetByID(id)
}

