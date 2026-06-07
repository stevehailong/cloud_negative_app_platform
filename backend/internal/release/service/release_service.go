package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"my-cloud/internal/release/model"
	"my-cloud/internal/release/repository"
	"my-cloud/pkg/k8s"

	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ReleaseService struct {
	releaseRepo         *repository.ReleaseRepository
	releaseApprovalRepo *repository.ReleaseApprovalRepository
	k8sClient           *k8s.Client
	iamDB               *gorm.DB
}

func NewReleaseService(
	releaseRepo *repository.ReleaseRepository,
	releaseApprovalRepo *repository.ReleaseApprovalRepository,
	k8sClient *k8s.Client,
	iamDB *gorm.DB,
) *ReleaseService {
	return &ReleaseService{
		releaseRepo:         releaseRepo,
		releaseApprovalRepo: releaseApprovalRepo,
		k8sClient:           k8sClient,
		iamDB:               iamDB,
	}
}

// CreateRelease 创建发布工单
func (s *ReleaseService) CreateRelease(release *model.Release) error {
	// 生成发布编号
	releaseNo := fmt.Sprintf("REL-%d-%d", release.AppID, time.Now().Unix())
	release.ReleaseNo = releaseNo
	release.ApprovalStatus = "pending"
	release.ReleaseStatus = "created"

	return s.releaseRepo.Create(release)
}

// GetRelease 获取发布工单详情
func (s *ReleaseService) GetRelease(id uint) (*model.Release, error) {
	return s.releaseRepo.GetByID(id)
}

// ListReleases 获取发布工单列表（含真实 Ingress 权重 + 操作人）
func (s *ReleaseService) ListReleases(appID, envID uint, releaseStatus string, page, pageSize int) ([]*model.Release, int64, error) {
	releases, total, err := s.releaseRepo.List(appID, envID, releaseStatus, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 收集操作人 ID
	userIDs := make(map[int64]bool)
	for _, r := range releases {
		if r.OperatorUserID > 0 {
			userIDs[int64(r.OperatorUserID)] = true
		}
	}
	userNames := batchResolveUserNames(s.iamDB, userIDs)

	// 填充操作人 + 真实 Ingress 权重
	for _, release := range releases {
		release.OperatorName = userNames[int64(release.OperatorUserID)]
		if release.ReleaseStatus == "canary" && release.CanaryStatus == "canary_running" && s.k8sClient != nil {
			if realWeight := s.getRealCanaryWeight(release); realWeight >= 0 {
				release.CanaryPercent = realWeight
			}
		}
	}

	return releases, total, nil
}

// batchResolveUserNames 批量解析 user_id → 姓名
func batchResolveUserNames(db *gorm.DB, userIDs map[int64]bool) map[int64]string {
	result := make(map[int64]string)
	if db == nil || len(userIDs) == 0 {
		return result
	}
	ids := make([]int64, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}
	type UserInfo struct {
		ID       int64  `gorm:"column:id"`
		Username string `gorm:"column:username"`
		RealName string `gorm:"column:real_name"`
	}
	var users []UserInfo
	_ = db.Table("users").Where("id IN ?", ids).Find(&users).Error
	for _, u := range users {
		if u.RealName != "" {
			result[u.ID] = u.RealName
		} else {
			result[u.ID] = u.Username
		}
	}
	return result
}

// getRealCanaryWeight 获取真实金丝雀权重
// Istio 模式: 从 VirtualService 读取
// Ingress 模式: 从 Nginx Ingress 注解读取
func (s *ReleaseService) getRealCanaryWeight(release *model.Release) int {
	namespace := s.getAppNamespace(release.AppID, release.EnvID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Istio 模式: 读 VirtualService
	if release.CanaryRoutingMode == "istio" && s.k8sClient != nil {
		vsName := fmt.Sprintf("app-%d", release.AppID)
		if w, err := s.k8sClient.GetCanaryTrafficWeight(ctx, namespace, vsName); err == nil {
			return w
		}
		return -1
	}

	// Ingress 模式: 读 Nginx Ingress 注解
	ingressName := fmt.Sprintf("app-%d-canary", release.AppID)
	ing, err := s.k8sClient.GetIngress(ctx, namespace, ingressName)
	if err != nil {
		return -1
	}

	if weightStr, ok := ing.Annotations["nginx.ingress.kubernetes.io/canary-weight"]; ok {
		if w, err := strconv.Atoi(weightStr); err == nil {
			return w
		}
	}
	return -1
}

// SubmitRelease 提交发布工单审批
func (s *ReleaseService) SubmitRelease(id uint, approverUserIDs []uint) error {
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	if release.ReleaseStatus != "created" {
		return errors.New("只能提交处于created状态的工单")
	}

	// 创建审批记录
	for _, approverUserID := range approverUserIDs {
		approval := &model.ReleaseApproval{
			ReleaseID:      id,
			ApproverUserID: approverUserID,
			ApprovalStatus: "pending",
		}
		if err := s.releaseApprovalRepo.Create(approval); err != nil {
			return err
		}
	}

	// 更新工单状态
	release.ReleaseStatus = "submitted"
	release.ApprovalStatus = "pending"
	return s.releaseRepo.Update(release)
}

// ApproveRelease 审批通过发布工单
func (s *ReleaseService) ApproveRelease(releaseID uint, approverUserID uint, comment string) error {
	release, err := s.releaseRepo.GetByID(releaseID)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	if release.ReleaseStatus != "submitted" {
		return errors.New("只能审批处于submitted状态的工单")
	}

	// 查找该审批人的审批记录
	approvals, err := s.releaseApprovalRepo.ListByRelease(releaseID)
	if err != nil {
		return err
	}

	var targetApproval *model.ReleaseApproval
	for _, approval := range approvals {
		if approval.ApproverUserID == approverUserID && approval.ApprovalStatus == "pending" {
			targetApproval = approval
			break
		}
	}

	if targetApproval == nil {
		return errors.New("未找到待审批记录或您不是审批人")
	}

	// 更新审批记录
	now := time.Now()
	targetApproval.ApprovalStatus = "approved"
	targetApproval.CommentText = comment
	targetApproval.ApprovalTime = &now
	if err := s.releaseApprovalRepo.Update(targetApproval); err != nil {
		return err
	}

	// 检查是否所有审批都通过
	pendingCount, _ := s.releaseApprovalRepo.CountPendingApprovals(releaseID)
	if pendingCount == 0 {
		// 所有审批都完成，更新工单状态
		release.ApprovalStatus = "approved"
		release.ReleaseStatus = "approved"
		return s.releaseRepo.Update(release)
	}

	return nil
}

// RejectRelease 拒绝发布工单
func (s *ReleaseService) RejectRelease(releaseID uint, approverUserID uint, comment string) error {
	release, err := s.releaseRepo.GetByID(releaseID)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	if release.ReleaseStatus != "submitted" {
		return errors.New("只能审批处于submitted状态的工单")
	}

	// 查找该审批人的审批记录
	approvals, err := s.releaseApprovalRepo.ListByRelease(releaseID)
	if err != nil {
		return err
	}

	var targetApproval *model.ReleaseApproval
	for _, approval := range approvals {
		if approval.ApproverUserID == approverUserID && approval.ApprovalStatus == "pending" {
			targetApproval = approval
			break
		}
	}

	if targetApproval == nil {
		return errors.New("未找到待审批记录或您不是审批人")
	}

	// 更新审批记录
	now := time.Now()
	targetApproval.ApprovalStatus = "rejected"
	targetApproval.CommentText = comment
	targetApproval.ApprovalTime = &now
	if err := s.releaseApprovalRepo.Update(targetApproval); err != nil {
		return err
	}

	// 一旦有一个审批拒绝，整个工单状态变为rejected
	release.ApprovalStatus = "rejected"
	release.ReleaseStatus = "rejected"
	return s.releaseRepo.Update(release)
}

// ExecuteRelease 执行发布
func (s *ReleaseService) ExecuteRelease(id uint, operatorUserID uint) error {
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	if release.ReleaseStatus != "approved" {
		return errors.New("只能执行审批通过的工单")
	}

	// 更新状态为执行中
	release.ReleaseStatus = "executing"
	release.OperatorUserID = operatorUserID
	if err := s.releaseRepo.Update(release); err != nil {
		return err
	}

	// 异步执行部署
	go s.executeDeployment(release)

	return nil
}

// executeDeployment 实际执行部署逻辑
func (s *ReleaseService) executeDeployment(release *model.Release) {
	switch release.ReleaseStrategy {
	case "canary":
		s.executeCanaryDeployment(release)
	case "rolling":
		s.executeRollingUpdateDeployment(release)
	case "bluegreen":
		s.executeRollingDeployment(release) // 暂用rolling实现
	default:
		s.executeRollingDeployment(release)
	}
}

// getAppNamespace 获取应用在指定环境的正确命名空间(app-{appId}-{envNamespace})
func (s *ReleaseService) getAppNamespace(appID, envID uint) string {
	// 调用deploy-service内部API获取环境信息
	url := fmt.Sprintf("http://deploy-service:8087/internal/v1/environments/%d", envID)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[Release] Failed to get environment: %v", err)
		return fmt.Sprintf("app-%d", appID) // fallback
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			Namespace string `json:"namespace"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Code != 0 {
		log.Printf("[Release] Failed to parse environment: %v", err)
		return fmt.Sprintf("app-%d", appID) // fallback
	}

	// 生成应用专属命名空间: app-{appId}-{envNamespace}
	appNamespace := fmt.Sprintf("app-%d-%s", appID, result.Data.Namespace)
	log.Printf("[Release] Generated namespace: %s (appID=%d, envNamespace=%s)", appNamespace, appID, result.Data.Namespace)
	return appNamespace
}

// executeRollingDeployment 滚动部署（全量，固定副本数）
func (s *ReleaseService) executeRollingDeployment(release *model.Release) {
	log.Printf("[Release] Executing rolling deployment for release %s", release.ReleaseNo)

	workloadName := fmt.Sprintf("app-%d", release.AppID)
	success := s.callDeployServiceWithWorkloadName(release, release.ImageURL, 5, workloadName)

	if success {
		release.ReleaseStatus = "success"
		log.Printf("[Release] Rolling deployment succeeded: %s", release.ReleaseNo)
	} else {
		release.ReleaseStatus = "failed"
		log.Printf("[Release] Rolling deployment failed: %s", release.ReleaseNo)
	}
	s.releaseRepo.Update(release)
}

// executeRollingUpdateDeployment 滚动更新（保持现有副本数，仅更新镜像）
func (s *ReleaseService) executeRollingUpdateDeployment(release *model.Release) {
	log.Printf("[Release] Executing rolling update for release %s", release.ReleaseNo)

	namespace := s.getAppNamespace(release.AppID, release.EnvID)
	
	// 1. 智能检测现有部署：优先检查 canary，再检查 stable
	baseWorkloadName := fmt.Sprintf("app-%d", release.AppID)
	canaryWorkloadName := fmt.Sprintf("app-%d-canary", release.AppID)
	
	var targetWorkloadName string
	var existingReplicas int
	
	// 优先检查 canary deployment
	canaryReplicas := s.getExistingDeploymentReplicas(namespace, canaryWorkloadName)
	if canaryReplicas > 0 {
		targetWorkloadName = canaryWorkloadName
		existingReplicas = canaryReplicas
		log.Printf("[Release] Found existing canary deployment: %s with %d replicas", canaryWorkloadName, canaryReplicas)
	} else {
		// 检查 stable deployment
		stableReplicas := s.getExistingDeploymentReplicas(namespace, baseWorkloadName)
		if stableReplicas > 0 {
			targetWorkloadName = baseWorkloadName
			existingReplicas = stableReplicas
			log.Printf("[Release] Found existing stable deployment: %s with %d replicas", baseWorkloadName, stableReplicas)
		} else {
			// 都不存在，创建新的 stable deployment
			targetWorkloadName = baseWorkloadName
			existingReplicas = 0
			log.Printf("[Release] No existing deployment found, will create new: %s", baseWorkloadName)
		}
	}
	
	var targetReplicas int
	if existingReplicas > 0 {
		// 存在现有部署，保持副本数不变
		targetReplicas = existingReplicas
		log.Printf("[Release] Keeping existing replica count: %d", targetReplicas)
	} else {
		// 首次部署，使用默认值
		targetReplicas = 5
		log.Printf("[Release] First deployment, using default %d replicas", targetReplicas)
	}

	// 2. 执行部署（更新或创建指定的 workload）
	success := s.callDeployServiceWithWorkloadName(release, release.ImageURL, targetReplicas, targetWorkloadName)

	if success {
		release.ReleaseStatus = "success"
		log.Printf("[Release] Rolling update succeeded: %s (workload: %s, replicas: %d)", release.ReleaseNo, targetWorkloadName, targetReplicas)
	} else {
		release.ReleaseStatus = "failed"
		log.Printf("[Release] Rolling update failed: %s", release.ReleaseNo)
	}
	s.releaseRepo.Update(release)
}

// callDeployServiceWithWorkloadName 调用deploy-service创建/更新指定名称的部署
func (s *ReleaseService) callDeployServiceWithWorkloadName(release *model.Release, imageURL string, replicas int, workloadName string) bool {
	namespace := s.getAppNamespace(release.AppID, release.EnvID)

	// 1. 查找或创建 app_deployment 记录
	appDeploymentID, err := s.getOrCreateAppDeployment(release, namespace, workloadName, replicas)
	if err != nil {
		log.Printf("[Release] Failed to get/create app_deployment: %v", err)
		return false
	}

	log.Printf("[Release] Using app_deployment_id=%d for %s/%s", appDeploymentID, namespace, workloadName)

	// 2. 调用新版部署API（内部API，无需认证）
	deployURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/%d/deploy", appDeploymentID)
	
	payload := fmt.Sprintf(`{
		"version": "%s",
		"image_url": "%s",
		"replicas": %d,
		"strategy": "%s",
		"user_id": %d
	}`, release.ReleaseVersion, imageURL, replicas, release.ReleaseStrategy, release.OperatorUserID)

	req, _ := http.NewRequest("POST", deployURL, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[Release] Failed to call deploy service: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Release] Deploy API returned error: %d %s", resp.StatusCode, string(body))
		return false
	}

	// 解析响应获取 deployment_history_id
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[Release] Failed to parse deploy response: %v", err)
		return false
	}

	var historyID int64
	if data, ok := result["data"].(map[string]interface{}); ok {
		if id, ok := data["history_id"].(float64); ok {
			historyID = int64(id)
		}
	}

	if historyID == 0 {
		log.Printf("[Release] No history_id returned from deploy API")
		return false
	}

	log.Printf("[Release] Deployment %s/%s started, history_id=%d", namespace, workloadName, historyID)
	
	// 3. 轮询 deployment_history 状态
	return s.waitDeploymentHistoryComplete(historyID)
}

// getOrCreateAppDeployment 获取或创建 app_deployment 记录
func (s *ReleaseService) getOrCreateAppDeployment(release *model.Release, namespace, workloadName string, replicas int) (int64, error) {
	// 查询是否已存在
	queryURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/by-workload?namespace=%s&workload_name=%s", namespace, workloadName)
	
	resp, err := http.Get(queryURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if data, ok := result["data"].(map[string]interface{}); ok {
			if id, ok := data["id"].(float64); ok {
				return int64(id), nil
			}
		}
	}

	// 不存在,创建新记录
	// namespace和cluster_id会从环境自动获取,无需传入
	createURL := "http://deploy-service:8087/internal/v1/app-deployments"
	payload := fmt.Sprintf(`{
		"app_id": %d,
		"env_id": %d,
		"workload_name": "%s",
		"workload_type": "deployment",
		"desired_replicas": %d
	}`, release.AppID, release.EnvID, workloadName, replicas)

	req, _ := http.NewRequest("POST", createURL, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("create app_deployment failed: %d %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if data, ok := result["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(float64); ok {
			return int64(id), nil
		}
	}

	return 0, fmt.Errorf("no id returned from create API")
}

// getExistingDeploymentReplicas 获取现有部署的副本数
func (s *ReleaseService) getExistingDeploymentReplicas(namespace, workloadName string) int {
	queryURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/k8s/deployments/%s/%s/replicas", namespace, workloadName)
	
	resp, err := http.Get(queryURL)
	if err != nil {
		log.Printf("[Release] Failed to query existing deployment: %v", err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// 部署不存在
		return 0
	}

	if resp.StatusCode >= 400 {
		log.Printf("[Release] Query deployment failed: HTTP %d", resp.StatusCode)
		return 0
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[Release] Failed to decode response: %v", err)
		return 0
	}

	if data, ok := result["data"].(map[string]interface{}); ok {
		if replicas, ok := data["replicas"].(float64); ok {
			return int(replicas)
		}
	}

	return 0
}

// executeCanaryDeployment 金丝雀部署
// 支持三种分流模式:
//   - "istio": 使用 Istio VirtualService + DestinationRule 分流
//   - "ingress" (默认): 使用 Nginx Ingress canary annotations 分流
//   - 降级: Pod 比例分流（当上述模式不可用时）
func (s *ReleaseService) executeCanaryDeployment(release *model.Release) {
	log.Printf("[Release] 执行金丝雀发布: release=%s, 流量比例=%d%%, 路由模式=%s",
		release.ReleaseNo, release.CanaryPercent, release.CanaryRoutingMode)

	namespace := s.getAppNamespace(release.AppID, release.EnvID)
	stableWorkloadName := fmt.Sprintf("app-%d", release.AppID)
	canaryWorkloadName := fmt.Sprintf("app-%d-canary", release.AppID)

	// 1. 获取当前 stable Deployment 的副本数
	stableReplicas, err := s.getCurrentReplicas(namespace, stableWorkloadName)
	if err != nil {
		log.Printf("[Release] 获取 stable 副本数失败: %v", err)
		release.ReleaseStatus = "failed"
		release.Description = fmt.Sprintf("获取当前副本数失败: %v", err)
		s.releaseRepo.Update(release)
		return
	}

	if stableReplicas == 0 {
		log.Printf("[Release] 当前没有运行的 Pod，执行全量部署")
		defaultReplicas := s.getConfiguredReplicas(release.AppID, release.EnvID, 3)
			success := s.callDeployServiceWithWorkloadName(release, release.ImageURL, defaultReplicas, stableWorkloadName)
		if success {
			release.ReleaseStatus = "deployed"
		} else {
			release.ReleaseStatus = "failed"
			release.Description = "部署失败"
		}
		s.releaseRepo.Update(release)
		return
	}

	// 2. 计算 canary 副本数
	canaryPercent := int(release.CanaryPercent)
	if canaryPercent <= 0 || canaryPercent > 100 {
		canaryPercent = 5
	}
	totalReplicas := stableReplicas
	canaryReplicas := int(float64(totalReplicas) * float64(canaryPercent) / 100.0)
	if canaryReplicas < 1 {
		canaryReplicas = 1
	}

	// 3. 创建 canary Deployment
	canarySuccess := s.callDeployServiceWithWorkloadName(release, release.ImageURL, canaryReplicas, canaryWorkloadName)
	if !canarySuccess {
		log.Printf("[Release] 创建 canary deployment 失败")
		release.ReleaseStatus = "failed"
		release.Description = "创建金丝雀部署失败"
		s.releaseRepo.Update(release)
		return
	}

	// 4. 分流模式选择: Istio > Ingress > Pod比例
	modeUsed := "Pod比例分流"
	trafficManaged := false

	if s.k8sClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 4a. 尝试 Istio 分流
		if release.CanaryRoutingMode == "istio" && s.k8sClient.IsIstioInstalled(ctx) {
			log.Printf("[Release] 尝试 Istio 分流模式")
			appName := stableWorkloadName // e.g. "app-1"
			appLabel := appName

			serviceHost := appName  // K8s Service name matches workloadName
			hosts := []string{fmt.Sprintf("%s.%s.svc.cluster.local", serviceHost, namespace)}

			drConfig := k8s.CanaryDestinationRuleConfig{
				Name:         appName,
				Namespace:    namespace,
				Host:         serviceHost,
				StableSubset: "stable",
				CanarySubset: "canary",
				StableLabels: map[string]string{
					"app":     appLabel,
					"version": stableWorkloadName,
				},
				CanaryLabels: map[string]string{
					"app":     appLabel,
					"version": canaryWorkloadName,
				},
				Labels: map[string]string{
					"app":        appLabel,
					"managed-by": "my-cloud",
				},
			}

			if drErr := s.k8sClient.EnsureCanaryDestinationRule(ctx, drConfig); drErr != nil {
				log.Printf("[Release] Istio DestinationRule 失败: %v, 降级到 Ingress/Pod 模式", drErr)
			} else {
				var headerMatches []k8s.HeaderMatchRule
				routingMode := release.CanaryRoutingMode
				if routingMode == "" {
					routingMode = "weight"
				}

				if routingMode == "header" || routingMode == "weight_header" {
					if release.CanaryHeaderName != "" {
						headerMatches = append(headerMatches, k8s.HeaderMatchRule{
							HeaderName:  release.CanaryHeaderName,
							HeaderValue: release.CanaryHeaderValue,
							Exact:       true,
							Subset:      "canary",
						})
					}
				}

				gateways := []string{"mesh"}
				vsConfig := k8s.CanaryVirtualServiceConfig{
					Name:          appName,
					Namespace:     namespace,
					Hosts:         hosts,
					Gateways:      gateways,
					StableHost:    serviceHost,
					CanaryHost:    serviceHost,
					StableSubset:  "stable",
					CanarySubset:  "canary",
					CanaryWeight:  canaryPercent,
					StableWeight:  100 - canaryPercent,
					HeaderMatches: headerMatches,
					Labels: map[string]string{
						"app":        appLabel,
						"managed-by": "my-cloud",
					},
				}

				if vsErr := s.k8sClient.EnsureCanaryVirtualService(ctx, vsConfig); vsErr != nil {
					log.Printf("[Release] Istio VirtualService 失败: %v, 清理 DR 并降级", vsErr)
					_ = s.k8sClient.DeleteDestinationRule(ctx, namespace, appName)
				} else {
					trafficManaged = true
					modeUsed = "Istio分流"
					release.CanaryIngressName = appName  // 复用字段存 VS/DR 名称
					release.CanaryServiceName = ""
					log.Printf("[Release] Istio 分流已激活: VS=%s/%s weight=%d%%",
						namespace, appName, canaryPercent)
				}
			}
		}

		// 4b. 尝试 Ingress 分流（非 istio 模式 或 istio 失败后的降级）
		if !trafficManaged && release.CanaryRoutingMode != "istio" {
			stableIngressName := stableWorkloadName
			stableSvcName := fmt.Sprintf("%s-service", stableWorkloadName)
			canarySvcName := fmt.Sprintf("%s-canary-svc", canaryWorkloadName)
			canaryIngressName := fmt.Sprintf("%s-canary-ingress", canaryWorkloadName)
			appLabel := stableWorkloadName

			host, path, pathType, tlsSecret, svcPort, ingressErr := ExtractStableIngressConfig(
				ctx, s.k8sClient, namespace, stableIngressName)

			if ingressErr == nil {
				log.Printf("[Release] Ingress 分流模式: host=%s path=%s", host, path)

				if narrowErr := NarrowStableServiceSelector(ctx, s.k8sClient, namespace, stableSvcName, stableWorkloadName); narrowErr != nil {
					log.Printf("[Release] Narrow stable service failed: %v, falling back to pod-based canary", narrowErr)
				} else {
					canarySvc := BuildCanaryServiceSpec(canarySvcName, namespace, appLabel, canaryWorkloadName, svcPort, 8080)
					if _, svcErr := s.k8sClient.CreateService(ctx, namespace, canarySvc); svcErr != nil {
						log.Printf("[Release] Create canary service failed: %v, cleaning up narrowed selector", svcErr)
						_ = WidenStableServiceSelector(ctx, s.k8sClient, namespace, stableSvcName)
					} else {
						routingMode := release.CanaryRoutingMode
						if routingMode == "" || routingMode == "istio" {
							routingMode = "weight"
						}
						canaryIng := BuildCanaryIngressSpec(
							canaryIngressName, namespace, host, path, pathType,
							canarySvcName, svcPort,
							routingMode, canaryPercent,
							release.CanaryHeaderName, release.CanaryHeaderValue, release.CanaryCookieName,
							tlsSecret,
							map[string]string{
								"app":        appLabel,
								"version":    canaryWorkloadName,
								"managed-by": "my-cloud",
								"role":       "canary",
							},
						)

						if _, ingErr := s.k8sClient.CreateIngress(ctx, namespace, canaryIng); ingErr != nil {
							log.Printf("[Release] Create canary ingress failed: %v, cleaning up", ingErr)
							if _, updateErr := s.k8sClient.UpdateIngress(ctx, namespace, canaryIng); updateErr != nil {
								_ = s.k8sClient.DeleteService(ctx, namespace, canarySvcName)
								_ = WidenStableServiceSelector(ctx, s.k8sClient, namespace, stableSvcName)
							} else {
								trafficManaged = true
								modeUsed = "Ingress分流"
								log.Printf("[Release] Updated existing canary Ingress %s/%s", namespace, canaryIngressName)
							}
						} else {
							trafficManaged = true
							modeUsed = "Ingress分流"
						}

						if trafficManaged {
							release.CanaryIngressName = canaryIngressName
							release.CanaryServiceName = canarySvcName
							log.Printf("[Release] Canary ingress mode active: ing=%s svc=%s weight=%d%%",
								canaryIngressName, canarySvcName, canaryPercent)
						}
					}
				}
			} else {
				log.Printf("[Release] Stable ingress %s/%s not found: %v — using pod-based canary",
					namespace, stableIngressName, ingressErr)
			}
		}
	} else {
		log.Printf("[Release] No K8s client — using pod-based canary")
	}

	// 5. 如果不是流量管理模式，回退到 Pod 比例分流
	if !trafficManaged {
		newStableReplicas := totalReplicas - canaryReplicas
		if newStableReplicas < 0 {
			newStableReplicas = 0
		}
		log.Printf("[Release] Pod比例分流: total=%d, stable=%d→%d, canary=0→%d (≈%d%%)",
			totalReplicas, stableReplicas, newStableReplicas, canaryReplicas, canaryPercent)

		if newStableReplicas < stableReplicas {
			s.scaleDeployment(namespace, stableWorkloadName, newStableReplicas)
		}
	}

	release.ReleaseStatus = "canary"
	release.CanaryStatus = "canary_running"
	release.CanaryPercent = canaryPercent
	release.Description = fmt.Sprintf("金丝雀发布运行中 [%s]: %d%% 流量到新版本", modeUsed, canaryPercent)
	s.releaseRepo.Update(release)
	log.Printf("[Release] 金丝雀发布成功: %s (%s)", release.ReleaseNo, modeUsed)
}

// scaleDeployment 缩放 Deployment 副本数
func (s *ReleaseService) scaleDeployment(namespace, workloadName string, replicas int) bool {
	url := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/by-workload?namespace=%s&workload_name=%s", 
		namespace, workloadName)
	
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[Release] 查询部署记录失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	var getResult struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&getResult); err != nil {
		log.Printf("[Release] 解析部署记录失败: %v", err)
		return false
	}

	if getResult.Code != 0 || getResult.Data.ID == 0 {
		log.Printf("[Release] 未找到部署记录: %s/%s", namespace, workloadName)
		return false
	}

	// 调用扩缩容接口
	scaleURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/%d/scale", getResult.Data.ID)
	scaleData := map[string]interface{}{
		"replicas": replicas,
		"user_id":  1,
	}
	
	jsonData, _ := json.Marshal(scaleData)
	scaleResp, err := http.Post(scaleURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Release] 扩缩容请求失败: %v", err)
		return false
	}
	defer scaleResp.Body.Close()

	if scaleResp.StatusCode >= 400 {
		body, _ := io.ReadAll(scaleResp.Body)
		log.Printf("[Release] 扩缩容失败: %d %s", scaleResp.StatusCode, string(body))
		return false
	}

	log.Printf("[Release] 扩缩容成功: %s/%s -> %d 副本", namespace, workloadName, replicas)
	return true
}

// getCurrentReplicas 获取当前 Deployment 的副本数
func (s *ReleaseService) getCurrentReplicas(namespace, workloadName string) (int, error) {
	url := fmt.Sprintf("http://deploy-service:8087/internal/v1/k8s/deployments/%s/%s/replicas", namespace, workloadName)
	
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return 0, nil // Deployment 不存在
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("获取副本数失败: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			Replicas int `json:"replicas"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("解析响应失败: %w", err)
	}

	return result.Data.Replicas, nil
}

// executePartialRollingUpdate 执行部分滚动更新（金丝雀发布）
func (s *ReleaseService) executePartialRollingUpdate(namespace, workloadName, newImage string, totalReplicas, canaryPods int) bool {
	// 使用 deploy-service 的部署接口，它会自动处理滚动更新
	// K8s 的 RollingUpdate 策略会逐步替换 Pod
	
	// 调用部署接口更新镜像
	url := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/by-workload?namespace=%s&workload_name=%s", 
		namespace, workloadName)
	
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[Release] 查询部署记录失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	var getResult struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&getResult); err != nil {
		log.Printf("[Release] 解析部署记录失败: %v", err)
		return false
	}

	if getResult.Code != 0 || getResult.Data.ID == 0 {
		log.Printf("[Release] 未找到部署记录")
		return false
	}

	// 调用部署接口更新镜像
	deployURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/%d/deploy", getResult.Data.ID)
	
	deployData := map[string]interface{}{
		"version":   fmt.Sprintf("canary-%d%%", canaryPods*100/totalReplicas),
		"image_url": newImage,
		"user_id":   1,
	}
	
	jsonData, _ := json.Marshal(deployData)
	deployResp, err := http.Post(deployURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Release] 调用部署接口失败: %v", err)
		return false
	}
	defer deployResp.Body.Close()

	if deployResp.StatusCode >= 400 {
		body, _ := io.ReadAll(deployResp.Body)
		log.Printf("[Release] 部署失败: %d %s", deployResp.StatusCode, string(body))
		return false
	}

	log.Printf("[Release] 金丝雀部署已提交: %s/%s, 镜像=%s", namespace, workloadName, newImage)
	return true
}

// ConfirmCanary 确认金丝雀，全量发布
func (s *ReleaseService) ConfirmCanary(id uint, operatorUserID uint) error {
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	if release.ReleaseStatus != "canary" || release.CanaryStatus != "canary_running" {
		return errors.New("当前状态不允许确认金丝雀")
	}

	release.CanaryStatus = "canary_confirmed"
	release.OperatorUserID = operatorUserID
	s.releaseRepo.Update(release)

	// 异步执行全量部署
	go func() {
		log.Printf("[Release] 金丝雀已确认，执行全量部署: %s", release.ReleaseNo)

		namespace := s.getAppNamespace(release.AppID, release.EnvID)
		stableWorkloadName := fmt.Sprintf("app-%d", release.AppID)
		canaryWorkloadName := fmt.Sprintf("app-%d-canary", release.AppID)

		// 1. 获取 canary 和 stable 的副本数
		canaryReplicas, _ := s.getCurrentReplicas(namespace, canaryWorkloadName)
		stableReplicas, _ := s.getCurrentReplicas(namespace, stableWorkloadName)
		totalReplicas := canaryReplicas + stableReplicas
		if totalReplicas == 0 {
			totalReplicas = 2
		}

		log.Printf("[Release] 确认金丝雀: canary=%d, stable=%d, 目标总副本=%d", canaryReplicas, stableReplicas, totalReplicas)

		// 2. 清理 Ingress 金丝雀资源（如果存在）
		s.cleanupCanaryTrafficResources(namespace, release, stableWorkloadName)

		// 3. 删除 canary Deployment（K8s资源 + 数据库记录）
		s.deleteCanaryDeployment(namespace, canaryWorkloadName)
		time.Sleep(2 * time.Second)

		// 4. 扩容 stable 到全量并更新镜像
		s.scaleDeployment(namespace, stableWorkloadName, totalReplicas)
		time.Sleep(2 * time.Second)
		s.callDeployServiceWithWorkloadName(release, release.ImageURL, totalReplicas, stableWorkloadName)

		release.ReleaseStatus = "deployed"
		release.CanaryStatus = ""
		release.CanaryPercent = 100
		release.CanaryIngressName = ""
		release.CanaryServiceName = ""
		release.Description = fmt.Sprintf("金丝雀发布已完成，全量部署成功 (%d 副本)", totalReplicas)
		log.Printf("[Release] 全量部署成功: %s, 副本数=%d", release.ReleaseNo, totalReplicas)
		s.releaseRepo.Update(release)
	}()

	return nil
}

// RollbackCanary 回滚金丝雀
func (s *ReleaseService) RollbackCanary(id uint, operatorUserID uint) error {
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	if release.ReleaseStatus != "canary" || release.CanaryStatus != "canary_running" {
		return errors.New("当前状态不允许回滚金丝雀")
	}

	release.CanaryStatus = "canary_rollback"
	release.OperatorUserID = operatorUserID
	s.releaseRepo.Update(release)

	// 异步回滚
	go func() {
		log.Printf("[Release] 回滚金丝雀部署: %s", release.ReleaseNo)

		namespace := s.getAppNamespace(release.AppID, release.EnvID)
		stableWorkloadName := fmt.Sprintf("app-%d", release.AppID)
		canaryWorkloadName := fmt.Sprintf("app-%d-canary", release.AppID)

		// 1. 获取 canary 和 stable 的副本数
		canaryReplicas, _ := s.getCurrentReplicas(namespace, canaryWorkloadName)
		stableReplicas, _ := s.getCurrentReplicas(namespace, stableWorkloadName)
		totalReplicas := canaryReplicas + stableReplicas

		// 2. 清理 Ingress 金丝雀资源（如果存在）
		s.cleanupCanaryTrafficResources(namespace, release, stableWorkloadName)

		// 3. 扩容 stable 到全量
		if totalReplicas > 0 {
			s.scaleDeployment(namespace, stableWorkloadName, totalReplicas)
		}

		// 4. 删除 canary Deployment
		s.deleteCanaryDeployment(namespace, canaryWorkloadName)

		log.Printf("[Release] 金丝雀回滚完成: %s", release.ReleaseNo)
		release.ReleaseStatus = "rollback"
		release.CanaryStatus = ""
		release.CanaryIngressName = ""
		release.CanaryServiceName = ""
		release.Description = "金丝雀发布已回滚"
		s.releaseRepo.Update(release)
	}()

	return nil
}

// deleteCanaryDeployment 删除金丝雀部署（K8s资源 + 数据库记录）
func (s *ReleaseService) deleteCanaryDeployment(namespace, workloadName string) {
	// 1. 删除K8s Deployment对象
	deleteURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/k8s/deployments/%s/%s", namespace, workloadName)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		log.Printf("[Release] 创建删除请求失败: %v", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[Release] 删除 canary deployment 失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Release] 删除 canary deployment 失败: %d %s", resp.StatusCode, string(body))
		return
	}

	log.Printf("[Release] 已删除 canary deployment: %s/%s", namespace, workloadName)

	// 2. 删除数据库中的 canary 部署记录
	queryURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/by-workload?namespace=%s&workload_name=%s", 
		namespace, workloadName)
	
	getResp, err := http.Get(queryURL)
	if err != nil {
		log.Printf("[Release] 查询 canary 部署记录失败: %v", err)
		return
	}
	defer getResp.Body.Close()

	var getResult struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(getResp.Body).Decode(&getResult); err != nil {
		log.Printf("[Release] 解析 canary 部署记录失败: %v", err)
		return
	}

	if getResult.Code == 0 && getResult.Data.ID > 0 {
		deleteRecordURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/%d", getResult.Data.ID)
		delReq, _ := http.NewRequest("DELETE", deleteRecordURL, nil)
		delResp, err := http.DefaultClient.Do(delReq)
		if err == nil {
			defer delResp.Body.Close()
			log.Printf("[Release] 已删除 canary 部署记录: ID=%d", getResult.Data.ID)
		}
	}
}

// cleanupCanaryTrafficResources 清理金丝雀分流资源（支持 Istio 和 Nginx Ingress 两种模式）
// 根据 release.CanaryRoutingMode 自动判断清理策略:
//   - "istio" 模式: 清理 Istio VirtualService + DestinationRule
//   - Ingress 模式: 清理 canary Ingress + canary Service，恢复 stable Service selector
func (s *ReleaseService) cleanupCanaryTrafficResources(namespace string, release *model.Release, stableWorkloadName string) {
	if s.k8sClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Istio 模式：清理 VirtualService + DestinationRule
	if release.CanaryRoutingMode == "istio" && release.CanaryIngressName != "" {
		appName := release.CanaryIngressName
		// 将 VirtualService 恢复为 100% stable（幂等操作）
		serviceHost := fmt.Sprintf("%s-service", stableWorkloadName)
		if vs, vsErr := s.k8sClient.GetVirtualService(ctx, namespace, appName); vsErr == nil {
			vs.Spec.HTTP = []k8s.HTTPRoute{
				{
					Route: []k8s.RouteDestination{
						{
							Destination: k8s.Destination{
								Host:   serviceHost,
								Subset: "stable",
							},
							Weight: 100,
						},
					},
				},
			}
			if err := s.k8sClient.UpdateVirtualService(ctx, namespace, vs); err != nil {
				log.Printf("[Release] Warning: Failed to update VirtualService for cleanup: %v", err)
			} else {
				log.Printf("[Release] Restored VirtualService %s/%s to 100%% stable", namespace, appName)
			}
		}
		// 注意: 不删除 DestinationRule，子集定义在一次发布后仍有价值
		log.Printf("[Release] Istio canary resources cleaned up for %s/%s", namespace, appName)
		return
	}

	// Nginx Ingress 模式：恢复 stable Service selector + 删除 canary Ingress/Service
	stableSvcName := fmt.Sprintf("%s-service", stableWorkloadName)
	if err := WidenStableServiceSelector(ctx, s.k8sClient, namespace, stableSvcName); err != nil {
		log.Printf("[Release] Warning: Failed to widen stable service selector: %v", err)
	}

	if release.CanaryIngressName != "" {
		if err := s.k8sClient.DeleteIngress(ctx, namespace, release.CanaryIngressName); err != nil {
			log.Printf("[Release] Warning: Failed to delete canary ingress %s/%s: %v", namespace, release.CanaryIngressName, err)
		} else {
			log.Printf("[Release] Deleted canary ingress %s/%s", namespace, release.CanaryIngressName)
		}
	}

	if release.CanaryServiceName != "" {
		if err := s.k8sClient.DeleteService(ctx, namespace, release.CanaryServiceName); err != nil {
			log.Printf("[Release] Warning: Failed to delete canary service %s/%s: %v", namespace, release.CanaryServiceName, err)
		} else {
			log.Printf("[Release] Deleted canary service %s/%s", namespace, release.CanaryServiceName)
		}
	}
}

// AdjustCanaryWeight 动态调整金丝雀流量权重
// 支持 Istio VirtualService 和 Nginx Ingress 两种分流方式
// AdjustCanaryWeight 统一入口：委托 deploy-service 处理权重调整
// deploy-service 负责 Istio VS / Nginx Ingress 的权重变更 + 100%/0% 生命周期
func (s *ReleaseService) AdjustCanaryWeight(id uint, newPercent int, operatorUserID uint) error {
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	if release.ReleaseStatus != "canary" || release.CanaryStatus != "canary_running" {
		return errors.New("当前状态不允许调整金丝雀权重")
	}

	if newPercent < 0 || newPercent > 100 {
		return errors.New("权重百分比必须在 0-100 之间")
	}

	namespace := s.getAppNamespace(release.AppID, release.EnvID)
	workloadName := fmt.Sprintf("app-%d-canary", release.AppID)

	// 查找 canary 部署 ID
	getURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/by-workload?namespace=%s&workload_name=%s",
		namespace, workloadName)
	resp, err := http.Get(getURL)
	if err != nil {
		return fmt.Errorf("查询部署记录失败: %w", err)
	}
	var getResult struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&getResult)
	resp.Body.Close()

	if getResult.Data.ID == 0 {
		return errors.New("未找到金丝雀部署记录")
	}

	// 统一委托 deploy-service 处理（含 Istio VS patch + 100%/0% 生命周期）
	payload := fmt.Sprintf(`{"weight":%d}`, newPercent)
	adjustURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-deployments/%d/canary/adjust-weight", getResult.Data.ID)

	req, _ := http.NewRequest("POST", adjustURL, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", operatorUserID))

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("调用部署服务失败: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != 200 {
		return fmt.Errorf("部署服务返回错误: %d", resp2.StatusCode)
	}

	// 同步 release 权重显示
	release.CanaryPercent = newPercent
	release.OperatorUserID = operatorUserID
	release.Description = fmt.Sprintf("金丝雀权重已调整为 %d%%", newPercent)
	s.releaseRepo.Update(release)

	log.Printf("[Release] Canary weight adjusted to %d%% via deploy-service for release %s", newPercent, release.ReleaseNo)
	return nil
}

// waitDeploymentHistoryComplete 等待deployment_history完成（轮询）
func (s *ReleaseService) waitDeploymentHistoryComplete(historyID int64) bool {
	historyURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/deployment-history/%d", historyID)

	for i := 0; i < 40; i++ {
		time.Sleep(3 * time.Second)

		resp, err := http.Get(historyURL)
		if err != nil {
			log.Printf("[Release] Failed to query history %d: %v", historyID, err)
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if data, ok := result["data"].(map[string]interface{}); ok {
			status, _ := data["status"].(string)
			log.Printf("[Release] History %d status: %s", historyID, status)
			
			if status == "success" {
				return true
			}
			if status == "failed" {
				reason, _ := data["failure_reason"].(string)
				if reason != "" {
					log.Printf("[Release] Deployment history %d failed: %s", historyID, reason)
				}
				return false
			}
		}
	}
	log.Printf("[Release] Deployment history %d polling timed out after 120s", historyID)
	return false
}


// autoHealIfCanaryGone checks if the canary Deployment still exists in K8s.
// If gone, auto-completes the release and returns true.
func (s *ReleaseService) autoHealIfCanaryGone(release *model.Release) bool {
	namespace := s.getAppNamespace(release.AppID, release.EnvID)
	canaryName := fmt.Sprintf("app-%d-canary", release.AppID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.k8sClient.GetDeployment(ctx, namespace, canaryName)
	if err == nil {
		return false // canary still exists, no heal needed
	}

	// canary gone → auto-complete the release
	release.ReleaseStatus = "deployed"
	release.CanaryStatus = ""
	release.CanaryPercent = 100
	release.Description = "canary auto-completed (Deployment already gone)"
	log.Printf("[Release] Auto-healed release %s: canary Deployment %s/%s not found, marked as deployed",
		release.ReleaseNo, namespace, canaryName)
	s.releaseRepo.Update(release)
	return true
}

// confirmCanaryFullFlow triggers the full canary confirmation flow (async-safe)
func (s *ReleaseService) confirmCanaryFullFlow(release *model.Release, operatorUserID uint) {
	if err := s.ConfirmCanary(release.ID, operatorUserID); err != nil {
		log.Printf("[Release] confirmCanaryFullFlow failed for %s: %v", release.ReleaseNo, err)
	}
}

// rollbackCanaryFullFlow triggers the full canary rollback flow (async-safe)
func (s *ReleaseService) rollbackCanaryFullFlow(release *model.Release, operatorUserID uint) {
	if err := s.RollbackCanary(release.ID, operatorUserID); err != nil {
		log.Printf("[Release] rollbackCanaryFullFlow failed for %s: %v", release.ReleaseNo, err)
	}
}


// getConfiguredReplicas returns the configured replicas from app_env_bindings, or falls back to default
func (s *ReleaseService) getConfiguredReplicas(appID, envID uint, defaultReplicas int) int {
	url := fmt.Sprintf("http://deploy-service:8087/internal/v1/app-env-binding?app_id=%d&env_id=%d", appID, envID)
	resp, err := http.Get(url)
	if err != nil {
		return defaultReplicas
	}
	defer resp.Body.Close()
	var result struct {
		Code int `json:"code"`
		Data struct {
			Replicas int `json:"replicas"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Data.Replicas > 0 {
		return result.Data.Replicas
	}
	return defaultReplicas
}

// RollbackRelease
func (s *ReleaseService) RollbackRelease(id uint, operatorUserID uint) error {
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	if release.ReleaseStatus != "success" && release.ReleaseStatus != "failed" {
		return errors.New("只能回滚已完成的发布")
	}

	// 更新状态为回滚
	release.ReleaseStatus = "rollback"
	release.OperatorUserID = operatorUserID
	return s.releaseRepo.Update(release)
}

// UpdateReleaseStatus 更新发布状态
func (s *ReleaseService) UpdateReleaseStatus(id uint, status string) error {
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	release.ReleaseStatus = status
	return s.releaseRepo.Update(release)
}

// UpdateRelease 更新发布工单(仅限created状态)
func (s *ReleaseService) UpdateRelease(id uint, releaseStrategy string, canaryPercent int, canaryRoutingMode, canaryHeaderName, canaryHeaderValue, canaryCookieName, description string) error {
	release, err := s.releaseRepo.GetByID(id)
	if err != nil {
		return errors.New("发布工单不存在")
	}

	// 只有created状态的工单才能修改
	if release.ReleaseStatus != "created" {
		return errors.New("只能编辑未提交的发布工单")
	}

	release.ReleaseStrategy = releaseStrategy
	release.CanaryPercent = canaryPercent

	// 金丝雀路由模式字段
	if releaseStrategy == "canary" {
		if canaryRoutingMode != "" {
			release.CanaryRoutingMode = canaryRoutingMode
		}
		if canaryHeaderName != "" {
			release.CanaryHeaderName = canaryHeaderName
		}
		if canaryHeaderValue != "" {
			release.CanaryHeaderValue = canaryHeaderValue
		}
		if canaryCookieName != "" {
			release.CanaryCookieName = canaryCookieName
		}
	}

	if description != "" {
		release.Description = description
	}

	return s.releaseRepo.Update(release)
}

// ListReleaseApprovals 获取发布工单的审批记录
func (s *ReleaseService) ListReleaseApprovals(releaseID uint) ([]*model.ReleaseApproval, error) {
	return s.releaseApprovalRepo.ListByRelease(releaseID)
}

// SyncCanaryConfirmed 同步金丝雀确认状态（由 deploy-service 权重100%触发）
func (s *ReleaseService) SyncCanaryConfirmed(appID, envID uint) error {
	releases, _, err := s.releaseRepo.List(appID, envID, "canary", 1, 1)
	if err != nil || len(releases) == 0 {
		return fmt.Errorf("no active canary release for app=%d env=%d", appID, envID)
	}
	release := releases[0]
	if release.CanaryStatus != "canary_running" {
		return fmt.Errorf("release %s not in canary_running state", release.ReleaseNo)
	}
	release.ReleaseStatus = "deployed"
	release.CanaryStatus = ""
	release.CanaryPercent = 100
	release.Description = "金丝雀已全量发布 (自动确认)"
	log.Printf("[Release] Canary confirmed by deploy-service: %s", release.ReleaseNo)
	return s.releaseRepo.Update(release)
}

// SyncCanaryRolledBack 同步金丝雀回滚状态（由 deploy-service 权重0%触发）
func (s *ReleaseService) SyncCanaryRolledBack(appID, envID uint) error {
	releases, _, err := s.releaseRepo.List(appID, envID, "canary", 1, 1)
	if err != nil || len(releases) == 0 {
		return fmt.Errorf("no active canary release for app=%d env=%d", appID, envID)
	}
	release := releases[0]
	if release.CanaryStatus != "canary_running" {
		return fmt.Errorf("release %s not in canary_running state", release.ReleaseNo)
	}
	release.ReleaseStatus = "rollback"
	release.CanaryStatus = ""
	release.Description = "金丝雀已回滚 (自动回滚)"
	log.Printf("[Release] Canary rolled back by deploy-service: %s", release.ReleaseNo)
	return s.releaseRepo.Update(release)
}
