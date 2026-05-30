package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"my-cloud/internal/release/model"
	"my-cloud/internal/release/repository"
	"net/http"
	"strings"
	"time"
)

type ReleaseService struct {
	releaseRepo         *repository.ReleaseRepository
	releaseApprovalRepo *repository.ReleaseApprovalRepository
}

func NewReleaseService(
	releaseRepo *repository.ReleaseRepository,
	releaseApprovalRepo *repository.ReleaseApprovalRepository,
) *ReleaseService {
	return &ReleaseService{
		releaseRepo:         releaseRepo,
		releaseApprovalRepo: releaseApprovalRepo,
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

// ListReleases 获取发布工单列表
func (s *ReleaseService) ListReleases(appID, envID uint, releaseStatus string, page, pageSize int) ([]*model.Release, int64, error) {
	return s.releaseRepo.List(appID, envID, releaseStatus, page, pageSize)
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

// executeRollingDeployment 滚动部署（全量，固定副本数）
func (s *ReleaseService) executeRollingDeployment(release *model.Release) {
	log.Printf("[Release] Executing rolling deployment for release %s", release.ReleaseNo)

	namespace := release.Namespace
	if namespace == "" {
		namespace = fmt.Sprintf("app-%d", release.AppID)
	}
	
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

	namespace := release.Namespace
	if namespace == "" {
		namespace = fmt.Sprintf("app-%d", release.AppID)
	}
	
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
	namespace := release.Namespace
	if namespace == "" {
		namespace = fmt.Sprintf("app-%d", release.AppID)
	}

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
		"replicas": %d
	}`, release.ReleaseVersion, imageURL, replicas)

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
	createURL := "http://deploy-service:8087/internal/v1/app-deployments"
	payload := fmt.Sprintf(`{
		"app_id": %d,
		"env_id": %d,
		"cluster_id": %d,
		"namespace": "%s",
		"workload_name": "%s",
		"workload_type": "deployment",
		"desired_replicas": %d
	}`, release.AppID, release.EnvID, release.ClusterID, namespace, workloadName, replicas)

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

// executeCanaryDeployment 金丝雀部署（通过副本数比例控制流量）
func (s *ReleaseService) executeCanaryDeployment(release *model.Release) {
	log.Printf("[Release] Executing canary deployment for release %s (canary%%=%d)", release.ReleaseNo, release.CanaryPercent)

	namespace := release.Namespace
	if namespace == "" {
		namespace = fmt.Sprintf("app-%d", release.AppID)
	}

	// 计算主deployment和canary的副本数
	// 假设总副本数为5，按照百分比分配
	totalReplicas := 5
	canaryPercent := int(release.CanaryPercent)
	if canaryPercent <= 0 || canaryPercent > 100 {
		canaryPercent = 20 // 默认20%
	}
	
	canaryReplicas := (totalReplicas * canaryPercent) / 100
	if canaryReplicas < 1 {
		canaryReplicas = 1 // 至少1个canary副本
	}
	stableReplicas := totalReplicas - canaryReplicas
	if stableReplicas < 1 {
		stableReplicas = 1 // 至少1个stable副本
	}

	log.Printf("[Release] Canary strategy: total=%d, stable=%d (old), canary=%d (new)", 
		totalReplicas, stableReplicas, canaryReplicas)

	// 1. 首先确保stable版本存在（如果是第一次部署，使用旧镜像或者占位镜像）
	// 注意：这里假设stable deployment已经存在，如果不存在需要先创建
	
	// 2. 部署canary版本 - 使用新版API
	canaryWorkloadName := fmt.Sprintf("app-%d-canary", release.AppID)
	success := s.callDeployServiceWithWorkloadName(release, release.ImageURL, canaryReplicas, canaryWorkloadName)

	if success {
		release.ReleaseStatus = "canary"
		release.CanaryStatus = "canary_running"
		log.Printf("[Release] Canary deployment running: %s", release.ReleaseNo)
	} else {
		release.ReleaseStatus = "failed"
		release.CanaryStatus = ""
		log.Printf("[Release] Canary deployment failed: %s", release.ReleaseNo)
	}
	s.releaseRepo.Update(release)
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
		log.Printf("[Release] Canary confirmed, promoting to full deployment: %s", release.ReleaseNo)

		namespace := release.Namespace
		if namespace == "" {
			namespace = fmt.Sprintf("app-%d", release.AppID)
		}
		stableWorkloadName := fmt.Sprintf("app-%d", release.AppID)
		canaryWorkloadName := fmt.Sprintf("app-%d-canary", release.AppID)

		// 1. 扩容 Canary 部署到全量副本数
		scaleSuccess := s.scaleDeployment(namespace, canaryWorkloadName, 5)
		if !scaleSuccess {
			log.Printf("[Release] Failed to scale canary to full replicas: %s", release.ReleaseNo)
			release.ReleaseStatus = "failed"
			s.releaseRepo.Update(release)
			return
		}
		log.Printf("[Release] Canary scaled to 5 replicas: %s/%s", namespace, canaryWorkloadName)

		// 2. 删除旧的稳定版本 K8s Deployment
		deleteURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/k8s/deployments/%s/%s", namespace, stableWorkloadName)
		req, err := http.NewRequest("DELETE", deleteURL, nil)
		if err != nil {
			log.Printf("[Release] Failed to create delete request for stable deployment: %v", err)
		} else {
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("[Release] Failed to delete stable deployment: %v", err)
			} else {
				defer resp.Body.Close()
				if resp.StatusCode < 400 {
					log.Printf("[Release] Stable deployment deleted: %s/%s", namespace, stableWorkloadName)
				} else {
					body, _ := io.ReadAll(resp.Body)
					log.Printf("[Release] Failed to delete stable deployment: %d %s", resp.StatusCode, string(body))
				}
			}
		}

		// 3. 删除旧的稳定版本数据库记录
		deleteRecordsURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/deployments/by-workload?namespace=%s&workloadName=%s", namespace, stableWorkloadName)
		deleteReq, err := http.NewRequest("DELETE", deleteRecordsURL, nil)
		if err != nil {
			log.Printf("[Release] Failed to create delete records request: %v", err)
		} else {
			deleteResp, err := http.DefaultClient.Do(deleteReq)
			if err != nil {
				log.Printf("[Release] Failed to delete stable deployment records: %v", err)
			} else {
				defer deleteResp.Body.Close()
				if deleteResp.StatusCode < 400 {
					log.Printf("[Release] Stable deployment records deleted: %s/%s", namespace, stableWorkloadName)
				} else {
					body, _ := io.ReadAll(deleteResp.Body)
					log.Printf("[Release] Failed to delete stable deployment records: %d %s", deleteResp.StatusCode, string(body))
				}
			}
		}

		// 4. 更新 Canary 部署记录的副本数（通过查询并更新数据库）
		// 这里假设 deploy-service 的 scale API 已经更新了数据库记录

		release.ReleaseStatus = "success"
		release.CanaryStatus = "canary_confirmed"
		s.releaseRepo.Update(release)
		log.Printf("[Release] Canary promotion completed: %s", release.ReleaseNo)
	}()

	return nil
}

// scaleDeployment 扩缩容部署
func (s *ReleaseService) scaleDeployment(namespace, workloadName string, replicas int) bool {
	scaleURL := "http://deploy-service:8087/internal/v1/deployments/scale"
	payload := fmt.Sprintf(`{
		"namespace": "%s",
		"workloadName": "%s",
		"replicas": %d
	}`, namespace, workloadName, replicas)

	resp, err := http.Post(scaleURL, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		log.Printf("[Release] Failed to scale deployment %s/%s: %v", namespace, workloadName, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Release] Scale API returned error: %d %s", resp.StatusCode, string(body))
		return false
	}

	log.Printf("[Release] Deployment scaled successfully: %s/%s to %d replicas", namespace, workloadName, replicas)
	return true
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

	// 异步回滚：删除canary workload
	go func() {
		log.Printf("[Release] Rolling back canary deployment: %s", release.ReleaseNo)
		s.deleteCanaryDeployment(release)
		release.ReleaseStatus = "rollback"
		s.releaseRepo.Update(release)
		log.Printf("[Release] Canary rollback completed: %s", release.ReleaseNo)
	}()

	return nil
}

// deleteCanaryDeployment 删除金丝雀部署（K8s资源 + 数据库记录）
func (s *ReleaseService) deleteCanaryDeployment(release *model.Release) {
	namespace := release.Namespace
	if namespace == "" {
		namespace = fmt.Sprintf("app-%d", release.AppID)
	}

	workloadName := fmt.Sprintf("app-%d-canary", release.AppID)
	
	// 1. 删除K8s Deployment对象
	deleteURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/k8s/deployments/%s/%s", namespace, workloadName)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		log.Printf("[Release] Failed to create delete request: %v", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[Release] Failed to delete canary deployment: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Release] Failed to delete canary deployment %s/%s: %d %s", namespace, workloadName, resp.StatusCode, string(body))
		return
	}

	log.Printf("[Release] Canary deployment K8s resource deleted: %s/%s", namespace, workloadName)
	
	// 2. 删除数据库中对应的deployment记录
	// 调用deploy-service的内部API删除所有canary相关记录
	deleteRecordsURL := fmt.Sprintf("http://deploy-service:8087/internal/v1/deployments/by-workload?namespace=%s&workloadName=%s", namespace, workloadName)
	deleteReq, err := http.NewRequest("DELETE", deleteRecordsURL, nil)
	if err != nil {
		log.Printf("[Release] Failed to create delete records request: %v", err)
		return
	}
	
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		log.Printf("[Release] Failed to delete canary deployment records: %v", err)
		return
	}
	defer deleteResp.Body.Close()
	
	if deleteResp.StatusCode >= 400 {
		body, _ := io.ReadAll(deleteResp.Body)
		log.Printf("[Release] Failed to delete canary records %s/%s: %d %s", namespace, workloadName, deleteResp.StatusCode, string(body))
	} else {
		log.Printf("[Release] Canary deployment records deleted: %s/%s", namespace, workloadName)
	}
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

// RollbackRelease 回滚发布
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
func (s *ReleaseService) UpdateRelease(id uint, releaseStrategy string, canaryPercent int, description string) error {
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
	if description != "" {
		release.Description = description
	}

	return s.releaseRepo.Update(release)
}

// ListReleaseApprovals 获取发布工单的审批记录
func (s *ReleaseService) ListReleaseApprovals(releaseID uint) ([]*model.ReleaseApproval, error) {
	return s.releaseApprovalRepo.ListByRelease(releaseID)
}
