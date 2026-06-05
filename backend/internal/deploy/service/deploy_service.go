package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"my-cloud/internal/deploy/model"
	"my-cloud/internal/deploy/repository"
	"my-cloud/pkg/k8s"
	"strings"
	"time"
)

type DeployService struct {
	deploymentRepo *repository.DeploymentRepository
	k8sClient      *k8s.Client
}

func NewDeployService(deploymentRepo *repository.DeploymentRepository, k8sClient *k8s.Client) *DeployService {
	return &DeployService{
		deploymentRepo: deploymentRepo,
		k8sClient:      k8sClient,
	}
}

// CreateDeployment 创建部署
func (s *DeployService) CreateDeployment(deployment *model.Deployment) error {
	deployment.DeploymentStatus = "progressing"
	now := time.Now()
	deployment.StartTime = &now

	if err := s.deploymentRepo.Create(deployment); err != nil {
		return err
	}

	// 异步执行真实的K8s部署
	go s.executeK8sDeployment(deployment)

	return nil
}

// executeK8sDeployment 实际执行K8s部署
func (s *DeployService) executeK8sDeployment(deployment *model.Deployment) {
	if s.k8sClient == nil {
		log.Printf("[Deploy] No K8s client available, skipping real deployment for %s", deployment.WorkloadName)
		// 模拟部署成功
		time.Sleep(5 * time.Second)
		s.updateStatus(deployment.ID, "success", deployment.DesiredReplicas)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	namespace := deployment.Namespace
	if namespace == "" {
		log.Printf("[Deploy] deployment %d has empty namespace, aborting", deployment.ID)
		s.updateStatusWithReason(deployment.ID, "failed", 0, "namespace is required")
		return
	}

	// 确保namespace存在
	if err := s.k8sClient.EnsureNamespace(ctx, namespace); err != nil {
		log.Printf("[Deploy] Failed to ensure namespace %s: %v", namespace, err)
		s.updateStatusWithReason(deployment.ID, "failed", 0, fmt.Sprintf("namespace creation failed: %v", err))
		return
	}

	// 确保NetworkPolicy存在（网络隔离）
	if err := s.k8sClient.EnsureNetworkPolicy(ctx, namespace); err != nil {
		log.Printf("[Deploy] Warning: Failed to create NetworkPolicy for %s: %v", namespace, err)
		// NetworkPolicy失败不阻塞部署
	}

	// 确保ResourceQuota存在（资源配额）
	if err := s.k8sClient.EnsureResourceQuota(ctx, namespace); err != nil {
		log.Printf("[Deploy] Warning: Failed to create ResourceQuota for %s: %v", namespace, err)
		// ResourceQuota失败不阻塞部署
	}

	// 构建K8s Deployment spec
	// 金丝雀部署：主和canary共享同一个app label用于Service流量分配
	isCanary := strings.HasSuffix(deployment.WorkloadName, "-canary")
	appName := deployment.WorkloadName
	if isCanary {
		appName = strings.TrimSuffix(appName, "-canary")
	}
	
	// 确保ServiceAccount存在（RBAC）
	if err := s.k8sClient.EnsureServiceAccount(ctx, namespace, appName); err != nil {
		log.Printf("[Deploy] Warning: Failed to create ServiceAccount for %s/%s: %v", namespace, appName, err)
		// RBAC失败不阻塞部署
	}
	
	labels := map[string]string{
		"app":        appName, // 主和canary使用相同的app label
		"version":    deployment.WorkloadName, // 保留完整名称用于区分
		"managed-by": "my-cloud",
	}

	// 确保Service存在（在部署之前创建，使用统一的app label）
	serviceName := fmt.Sprintf("%s-service", appName)
	if _, err := s.k8sClient.EnsureService(ctx, namespace, serviceName, appName, 80, 80); err != nil {
		log.Printf("[Deploy] Failed to ensure service %s/%s: %v", namespace, serviceName, err)
		// Service创建失败不阻塞部署，但记录日志
	} else {
		log.Printf("[Deploy] Service %s/%s ensured with selector app=%s", namespace, serviceName, appName)
	}

	image := deployment.ImageVersion
	if image == "" {
		image = "nginx:latest"
	}

	replicas := int32(deployment.DesiredReplicas)
	if replicas <= 0 {
		replicas = 1
	}

	k8sDeploy := k8s.BuildDeploymentSpec(
		deployment.WorkloadName,
		namespace,
		image,
		replicas,
		labels,
	)

	// 尝试获取已有的deployment
	existing, err := s.k8sClient.GetDeployment(ctx, namespace, deployment.WorkloadName)
	if err == nil && existing != nil {
		// 更新已有的deployment
		existing.Spec.Template.Spec.Containers[0].Image = image
		existing.Spec.Replicas = &replicas
		_, err = s.k8sClient.UpdateDeployment(ctx, namespace, existing)
	} else {
		// 创建新的deployment
		_, err = s.k8sClient.CreateDeployment(ctx, namespace, k8sDeploy)
	}

	if err != nil {
		log.Printf("[Deploy] Failed to apply K8s deployment %s/%s: %v", namespace, deployment.WorkloadName, err)
		s.updateStatusWithReason(deployment.ID, "failed", 0, fmt.Sprintf("K8s deployment apply failed: %v", err))
		return
	}

	log.Printf("[Deploy] K8s deployment %s/%s applied, waiting for rollout...", namespace, deployment.WorkloadName)

	// 等待部署完成（轮询检查副本就绪状态）
	s.waitForRollout(ctx, deployment, namespace)
}

// waitForRollout 等待K8s部署完成
func (s *DeployService) waitForRollout(ctx context.Context, deployment *model.Deployment, namespace string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeout := time.After(90 * time.Second)

	for {
		select {
		case <-timeout:
			reason := s.getFailureReason(ctx, deployment, namespace)
			log.Printf("[Deploy] Deployment %s/%s rollout timed out: %s", namespace, deployment.WorkloadName, reason)
			s.updateStatusWithReason(deployment.ID, "failed", 0, reason)
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			k8sDeploy, err := s.k8sClient.GetDeployment(ctx, namespace, deployment.WorkloadName)
			if err != nil {
				continue
			}

			available := int(k8sDeploy.Status.AvailableReplicas)
			desired := int(*k8sDeploy.Spec.Replicas)

			if available >= desired && k8sDeploy.Status.ReadyReplicas >= int32(desired) {
				log.Printf("[Deploy] Deployment %s/%s rolled out successfully (%d/%d ready)",
					namespace, deployment.WorkloadName, available, desired)
				s.updateStatus(deployment.ID, "success", available)
				return
			}
		}
	}
}

// getFailureReason 获取部署失败原因（查询Pod事件和状态）
func (s *DeployService) getFailureReason(ctx context.Context, deployment *model.Deployment, namespace string) string {
	if s.k8sClient == nil {
		return "deployment rollout timed out"
	}

	// 查询Pod列表获取容器状态
	// 使用完整的workloadName作为version label查询，避免查到其他版本的pods
	labelSelector := fmt.Sprintf("version=%s", deployment.WorkloadName)
	pods, err := s.k8sClient.GetPods(ctx, namespace, labelSelector)
	if err != nil {
		return fmt.Sprintf("rollout timed out (unable to query pods: %v)", err)
	}

	for _, pod := range pods {
		for _, cs := range pod.Status.ContainerStatuses {
			// 1. 容器正在等待 (ImagePullBackOff, CrashLoopBackOff, ErrImagePull 等)
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				reason := cs.State.Waiting.Reason
				msg := cs.State.Waiting.Message
				// 如果是 CrashLoopBackOff，尝试从上次终止状态获取崩溃原因
				if reason == "CrashLoopBackOff" && cs.LastTerminationState.Terminated != nil {
					term := cs.LastTerminationState.Terminated
					return fmt.Sprintf("Pod %s: container %s crashed (exit=%d) — %s: %s. Last termination: %s",
						pod.Name, cs.Name, term.ExitCode, term.Reason, term.Message, msg)
				}
				return fmt.Sprintf("Pod %s: container %s — %s: %s", pod.Name, cs.Name, reason, msg)
			}
			// 2. 容器已终止（非零退出码）
			if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
				term := cs.State.Terminated
				return fmt.Sprintf("Pod %s: container %s exited with code %d — %s: %s",
					pod.Name, cs.Name, term.ExitCode, term.Reason, term.Message)
			}
			// 3. 从重启次数和上次终止状态获取原因
			if cs.RestartCount > 0 && cs.LastTerminationState.Terminated != nil {
				term := cs.LastTerminationState.Terminated
				return fmt.Sprintf("Pod %s: container %s restarted %d times, last exit=%d — %s: %s",
					pod.Name, cs.Name, cs.RestartCount, term.ExitCode, term.Reason, term.Message)
			}
		}
		// 4. 检查Pod条件
		for _, cond := range pod.Status.Conditions {
			if cond.Status == "False" && cond.Message != "" {
				return fmt.Sprintf("Pod %s: %s — %s", pod.Name, cond.Reason, cond.Message)
			}
		}
	}

	return "deployment rollout timed out (90s)"
}

func (s *DeployService) updateStatus(deploymentID uint, status string, availableReplicas int) {
	s.updateStatusWithReason(deploymentID, status, availableReplicas, "")
}

func (s *DeployService) updateStatusWithReason(deploymentID uint, status string, availableReplicas int, reason string) {
	deployment, err := s.deploymentRepo.GetByID(deploymentID)
	if err != nil {
		return
	}
	deployment.DeploymentStatus = status
	deployment.AvailableReplicas = availableReplicas
	if reason != "" {
		deployment.FailureReason = reason
	}
	if status == "success" || status == "failed" {
		now := time.Now()
		deployment.EndTime = &now
	}
	s.deploymentRepo.Update(deployment)
}

// GetDeployment 获取部署详情
func (s *DeployService) GetDeployment(id uint) (*model.Deployment, error) {
	return s.deploymentRepo.GetByID(id)
}

// GetDeploymentByRelease 根据Release获取部署
func (s *DeployService) GetDeploymentByRelease(releaseID uint) (*model.Deployment, error) {
	return s.deploymentRepo.GetByRelease(releaseID)
}

// ListDeployments 获取部署列表
func (s *DeployService) ListDeployments(clusterID uint, namespace string, startDate, sortBy, sortOrder string, page, pageSize int) ([]*model.Deployment, int64, error) {
	return s.deploymentRepo.List(clusterID, namespace, startDate, sortBy, sortOrder, page, pageSize)
}

// UpdateDeploymentStatus 更新部署状态
func (s *DeployService) UpdateDeploymentStatus(id uint, status string, availableReplicas int) error {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return errors.New("部署记录不存在")
	}

	deployment.DeploymentStatus = status
	deployment.AvailableReplicas = availableReplicas

	if status == "success" || status == "failed" {
		now := time.Now()
		deployment.EndTime = &now
	}

	return s.deploymentRepo.Update(deployment)
}

// RestartDeployment 重启部署
func (s *DeployService) RestartDeployment(id uint) error {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return errors.New("部署记录不存在")
	}

	if s.k8sClient == nil {
		log.Printf("[Deploy] No K8s client, simulating restart for %s", deployment.WorkloadName)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.k8sClient.RestartDeployment(ctx, deployment.Namespace, deployment.WorkloadName); err != nil {
		return fmt.Errorf("重启失败: %v", err)
	}

	log.Printf("[Deploy] Restarted deployment %s/%s", deployment.Namespace, deployment.WorkloadName)
	return nil
}

// ScaleDeployment 扩缩容（通过部署ID）
func (s *DeployService) ScaleDeployment(id uint, replicas int) error {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return errors.New("部署记录不存在")
	}

	if replicas < 0 {
		return errors.New("副本数不能为负数")
	}

	if s.k8sClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.k8sClient.ScaleDeployment(ctx, deployment.Namespace, deployment.WorkloadName, int32(replicas)); err != nil {
			return fmt.Errorf("扩缩容失败: %v", err)
		}
		log.Printf("[Deploy] Scaled deployment %s/%s to %d replicas", deployment.Namespace, deployment.WorkloadName, replicas)
	}

	deployment.DesiredReplicas = replicas
	return s.deploymentRepo.Update(deployment)
}

// ScaleDeploymentByName 扩缩容（通过命名空间和工作负载名称）
func (s *DeployService) ScaleDeploymentByName(namespace, workloadName string, replicas int) error {
	if replicas < 0 {
		return errors.New("副本数不能为负数")
	}

	// 1. 执行 K8s 扩缩容
	if s.k8sClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.k8sClient.ScaleDeployment(ctx, namespace, workloadName, int32(replicas)); err != nil {
			return fmt.Errorf("扩缩容失败: %v", err)
		}
		log.Printf("[Deploy] Scaled deployment %s/%s to %d replicas", namespace, workloadName, replicas)

		// 等待扩缩容生效
		time.Sleep(3 * time.Second)

		// 2. 获取最新的 K8s 状态并同步到数据库
		deployment, err := s.k8sClient.GetDeployment(ctx, namespace, workloadName)
		if err != nil {
			log.Printf("[Deploy] Warning: Failed to get deployment status from K8s: %v", err)
		} else {
			// 3. 更新数据库记录
			deployments, err := s.deploymentRepo.FindByWorkload(namespace, workloadName)
			if err != nil {
				log.Printf("[Deploy] Warning: Failed to find deployment records for %s/%s: %v", namespace, workloadName, err)
				return nil // K8s 扩缩容已成功，数据库更新失败不应该影响结果
			}

			availableReplicas := int(deployment.Status.AvailableReplicas)
			for _, deployRecord := range deployments {
				deployRecord.DesiredReplicas = replicas
				deployRecord.AvailableReplicas = availableReplicas
				if err := s.deploymentRepo.Update(deployRecord); err != nil {
					log.Printf("[Deploy] Warning: Failed to update deployment record %d: %v", deployRecord.ID, err)
				}
			}
			log.Printf("[Deploy] Updated %d deployment records for %s/%s (desired=%d, available=%d)", 
				len(deployments), namespace, workloadName, replicas, availableReplicas)
		}
	} else {
		// K8s client 为 nil 时，仅更新 desiredReplicas
		deployments, err := s.deploymentRepo.FindByWorkload(namespace, workloadName)
		if err != nil {
			log.Printf("[Deploy] Warning: Failed to find deployment records for %s/%s: %v", namespace, workloadName, err)
			return nil
		}

		for _, deployment := range deployments {
			deployment.DesiredReplicas = replicas
			if err := s.deploymentRepo.Update(deployment); err != nil {
				log.Printf("[Deploy] Warning: Failed to update deployment record %d: %v", deployment.ID, err)
			}
		}
		log.Printf("[Deploy] Updated %d deployment records for %s/%s", len(deployments), namespace, workloadName)
	}

	return nil
}

// GetDeploymentEvents 获取部署事件
func (s *DeployService) GetDeploymentEvents(id uint) ([]map[string]interface{}, error) {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("部署记录不存在")
	}

	if s.k8sClient == nil {
		return []map[string]interface{}{
			{"type": "Normal", "reason": "Scheduled", "message": "Successfully assigned pod", "time": time.Now().Add(-5 * time.Minute)},
			{"type": "Normal", "reason": "Pulled", "message": "Container image pulled", "time": time.Now().Add(-4 * time.Minute)},
			{"type": "Normal", "reason": "Started", "message": "Started container", "time": time.Now().Add(-3 * time.Minute)},
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := s.k8sClient.GetEvents(ctx, deployment.Namespace, deployment.WorkloadName)
	if err != nil {
		log.Printf("[Deploy] Failed to get events for %s/%s: %v", deployment.Namespace, deployment.WorkloadName, err)
		return []map[string]interface{}{}, nil
	}

	result := make([]map[string]interface{}, 0, len(events))
	for _, ev := range events {
		result = append(result, map[string]interface{}{
			"type":    ev.Type,
			"reason":  ev.Reason,
			"message": ev.Message,
			"time":    ev.LastTimestamp.Time,
		})
	}
	return result, nil
}

// GetDeploymentPods 获取部署的Pod列表
func (s *DeployService) GetDeploymentPods(id uint) ([]map[string]interface{}, error) {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("部署记录不存在")
	}

	if s.k8sClient == nil {
		return []map[string]interface{}{
			{"name": deployment.WorkloadName + "-abc123", "status": "Running", "ready": "1/1", "restarts": 0, "age": "10m", "node": "worker-1", "ip": "10.244.0.5"},
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 修复：使用 version 标签来匹配 Pod，因为 app 标签是共享的
	// app-8 和 app-8-canary 共享 app=app-8 标签，但有不同的 version 标签
	labelSelector := fmt.Sprintf("version=%s", deployment.WorkloadName)
	pods, err := s.k8sClient.GetPods(ctx, deployment.Namespace, labelSelector)
	if err != nil {
		log.Printf("[Deploy] Failed to get pods for %s/%s: %v", deployment.Namespace, deployment.WorkloadName, err)
		return []map[string]interface{}{}, nil
	}

	result := make([]map[string]interface{}, 0, len(pods))
	for _, pod := range pods {
		ready := 0
		total := len(pod.Status.ContainerStatuses)
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				ready++
			}
		}

		restarts := int32(0)
		if len(pod.Status.ContainerStatuses) > 0 {
			restarts = pod.Status.ContainerStatuses[0].RestartCount
		}

		age := time.Since(pod.CreationTimestamp.Time).Round(time.Second).String()

		result = append(result, map[string]interface{}{
			"name":     pod.Name,
			"status":   string(pod.Status.Phase),
			"ready":    fmt.Sprintf("%d/%d", ready, total),
			"restarts": restarts,
			"age":      age,
			"node":     pod.Spec.NodeName,
			"ip":       pod.Status.PodIP,
		})
	}
	return result, nil
}

// DeletePod 删除指定的Pod
func (s *DeployService) DeletePod(namespace, podName string) error {
	if s.k8sClient == nil {
		return errors.New("K8s客户端未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.k8sClient.DeletePod(ctx, namespace, podName); err != nil {
		return fmt.Errorf("删除Pod失败: %v", err)
	}
	log.Printf("[Deploy] Pod deleted: %s/%s", namespace, podName)
	return nil
}


// DeleteDeployment 删除部署（K8s workload + 数据库记录）
func (s *DeployService) DeleteDeployment(id uint) error {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return errors.New("部署记录不存在")
	}
	if s.k8sClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = s.k8sClient.DeleteDeployment(ctx, deployment.Namespace, deployment.WorkloadName)
	}
	return s.deploymentRepo.Delete(id)
}

// DeleteK8sDeployment 直接删除K8s Deployment（不删除数据库记录）
func (s *DeployService) DeleteK8sDeployment(namespace, name string) error {
	if s.k8sClient == nil {
		return errors.New("K8s客户端未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.k8sClient.DeleteDeployment(ctx, namespace, name); err != nil {
		return fmt.Errorf("删除Deployment失败: %v", err)
	}
	log.Printf("[Deploy] K8s Deployment deleted: %s/%s", namespace, name)
	return nil
}

// DeleteDeploymentsByWorkload 删除指定workload的所有数据库记录
func (s *DeployService) DeleteDeploymentsByWorkload(namespace, workloadName string) error {
	// 查询所有匹配的记录
	deployments, err := s.deploymentRepo.FindByWorkload(namespace, workloadName)
	if err != nil {
		return fmt.Errorf("查询部署记录失败: %v", err)
	}
	
	// 删除所有记录
	deletedCount := 0
	for _, deployment := range deployments {
		if err := s.deploymentRepo.Delete(deployment.ID); err != nil {
			log.Printf("[Deploy] Failed to delete deployment record %d: %v", deployment.ID, err)
		} else {
			deletedCount++
		}
	}
	
	log.Printf("[Deploy] Deleted %d deployment records for %s/%s", deletedCount, namespace, workloadName)
	return nil
}

// GetK8sDeploymentReplicas 获取K8s Deployment的副本数
func (s *DeployService) GetK8sDeploymentReplicas(namespace, name string) (int, error) {
	if s.k8sClient == nil {
		return 0, errors.New("K8s client not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deployment, err := s.k8sClient.GetDeployment(ctx, namespace, name)
	if err != nil {
		return 0, fmt.Errorf("获取部署失败: %v", err)
	}

	return int(*deployment.Spec.Replicas), nil
}

// RollbackDeployment 回滚部署到上一个版本
func (s *DeployService) RollbackDeployment(id uint) error {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return errors.New("部署记录不存在")
	}
	if s.k8sClient == nil {
		return errors.New("K8s客户端未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.k8sClient.RolloutUndo(ctx, deployment.Namespace, deployment.WorkloadName); err != nil {
		return fmt.Errorf("回滚失败: %v", err)
	}
	deployment.DeploymentStatus = "rollback"
	s.deploymentRepo.Update(deployment)
	log.Printf("[Deploy] Rolled back deployment %s/%s", deployment.Namespace, deployment.WorkloadName)
	return nil
}
