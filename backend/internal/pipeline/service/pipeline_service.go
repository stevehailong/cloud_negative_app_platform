package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"my-cloud/internal/pipeline/model"
	"my-cloud/internal/pipeline/repository"
	"my-cloud/pkg/gitlab"
	"my-cloud/pkg/jenkins"
	"net/http"
	"strings"
	"time"
)

type PipelineService struct {
	pipelineRepo    *repository.PipelineRepository
	pipelineRunRepo *repository.PipelineRunRepository
	artifactRepo    *repository.ArtifactRepository
	jenkinsClient   *jenkins.Client
	gitlabClient    *gitlab.Client
}

func NewPipelineService(
	pipelineRepo *repository.PipelineRepository,
	pipelineRunRepo *repository.PipelineRunRepository,
	artifactRepo *repository.ArtifactRepository,
	jenkinsClient *jenkins.Client,
	gitlabClient *gitlab.Client,
) *PipelineService {
	return &PipelineService{
		pipelineRepo:    pipelineRepo,
		pipelineRunRepo: pipelineRunRepo,
		artifactRepo:    artifactRepo,
		jenkinsClient:   jenkinsClient,
		gitlabClient:    gitlabClient,
	}
}

// SetGitlabClient 动态设置GitLab客户端（从系统设置加载后更新）
func (s *PipelineService) SetGitlabClient(client *gitlab.Client) {
	s.gitlabClient = client
}

// GetGitlabClient 获取当前GitLab客户端
func (s *PipelineService) GetGitlabClient() *gitlab.Client {
	return s.gitlabClient
}

// CreatePipeline 创建流水线
func (s *PipelineService) CreatePipeline(pipeline *model.Pipeline) error {
	if existing, _ := s.pipelineRepo.GetByCode(pipeline.PipelineCode); existing != nil {
		return errors.New("流水线代码已存在")
	}

	if err := s.pipelineRepo.Create(pipeline); err != nil {
		return err
	}

	// 如果Jenkins可用，创建对应的Jenkins Job
	if s.jenkinsClient != nil {
		go s.ensureJenkinsJob(pipeline)
	}

	// 如果GitLab可用且配置了repoUrl，自动注册Webhook
	if s.gitlabClient != nil && pipeline.ConfigJSON != "" {
		go s.autoRegisterWebhook(pipeline)
	}

	return nil
}

// autoRegisterWebhook 自动为GitLab项目注册Webhook
func (s *PipelineService) autoRegisterWebhook(pipeline *model.Pipeline) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(pipeline.ConfigJSON), &config); err != nil {
		return
	}

	repoUrl, _ := config["repoUrl"].(string)
	if repoUrl == "" {
		return
	}

	// 获取Webhook回调地址
	webhookBaseURL := os.Getenv("WEBHOOK_BASE_URL")
	if webhookBaseURL == "" {
		log.Printf("[Pipeline] WEBHOOK_BASE_URL not configured, skipping webhook registration")
		return
	}
	hookURL := strings.TrimRight(webhookBaseURL, "/") + "/hooks/gitlab"

	// 通过repoUrl查找GitLab项目ID
	// 从URL提取项目路径（支持 https://gitlab.com/group/project.git 格式）
	projectPath := extractProjectPath(repoUrl)
	if projectPath == "" {
		log.Printf("[Pipeline] Cannot extract project path from URL: %s", repoUrl)
		return
	}

	// 用项目名称搜索（GitLab API 搜索不支持路径格式）
	projectName := projectPath
	if idx := strings.LastIndex(projectPath, "/"); idx != -1 {
		projectName = projectPath[idx+1:]
	}

	// 搜索项目获取ID
	projects, err := s.gitlabClient.ListProjects(projectName, 1, 20)
	if err != nil || len(projects) == 0 {
		log.Printf("[Pipeline] Cannot find GitLab project for path %s: %v", projectPath, err)
		return
	}

	// 找到匹配的项目
	var projectID string
	for _, p := range projects {
		if p.HTTPURLToRepo == repoUrl || p.HTTPURLToRepo == strings.TrimSuffix(repoUrl, ".git") ||
			strings.TrimSuffix(p.HTTPURLToRepo, ".git") == strings.TrimSuffix(repoUrl, ".git") {
			projectID = fmt.Sprintf("%d", p.ID)
			break
		}
	}
	if projectID == "" {
		// 尝试用第一个结果
		projectID = fmt.Sprintf("%d", projects[0].ID)
	}

	// 检查是否已有相同的webhook
	existingHooks, err := s.gitlabClient.ListWebhooks(projectID)
	if err == nil {
		for _, h := range existingHooks {
			if h.URL == hookURL {
				log.Printf("[Pipeline] Webhook already exists for project %s", projectPath)
				return
			}
		}
	}

	// 创建Webhook
	_, err = s.gitlabClient.CreateWebhook(projectID, hookURL, true, false)
	if err != nil {
		log.Printf("[Pipeline] Failed to register webhook for project %s: %v", projectPath, err)
		return
	}
	log.Printf("[Pipeline] Webhook registered for project %s -> %s", projectPath, hookURL)
}

// extractProjectPath 从Git URL提取项目路径
func extractProjectPath(repoUrl string) string {
	// 处理 https://gitlab.com/group/project.git 格式
	repoUrl = strings.TrimSuffix(repoUrl, ".git")
	// 去掉协议部分
	if idx := strings.Index(repoUrl, "://"); idx != -1 {
		repoUrl = repoUrl[idx+3:]
	}
	// 去掉域名部分，取路径
	if idx := strings.Index(repoUrl, "/"); idx != -1 {
		return repoUrl[idx+1:]
	}
	return ""
}

// GetPipeline 获取流水线详情
func (s *PipelineService) GetPipeline(id uint) (*model.Pipeline, error) {
	return s.pipelineRepo.GetByID(id)
}

// ListPipelines 获取流水线列表
func (s *PipelineService) ListPipelines(appID uint, page, pageSize int) ([]*model.Pipeline, int64, error) {
	return s.pipelineRepo.List(appID, page, pageSize)
}

// UpdatePipeline 更新流水线
func (s *PipelineService) UpdatePipeline(pipeline *model.Pipeline) error {
	existing, err := s.pipelineRepo.GetByID(pipeline.ID)
	if err != nil {
		return errors.New("流水线不存在")
	}

	if existing.PipelineCode != pipeline.PipelineCode {
		if dup, _ := s.pipelineRepo.GetByCode(pipeline.PipelineCode); dup != nil {
			return errors.New("流水线代码已存在")
		}
	}

	return s.pipelineRepo.Update(pipeline)
}

// DeletePipeline 删除流水线
func (s *PipelineService) DeletePipeline(id uint) error {
	return s.pipelineRepo.Delete(id)
}

// RunPipeline 触发流水线执行
func (s *PipelineService) RunPipeline(pipelineID uint, triggerType, gitCommit, gitBranch string, operatorUserID uint) (*model.PipelineRun, error) {
	pipeline, err := s.pipelineRepo.GetByID(pipelineID)
	if err != nil {
		return nil, errors.New("流水线不存在")
	}

	if pipeline.Enabled != 1 {
		return nil, errors.New("流水线已被禁用")
	}

	runNo := fmt.Sprintf("%s-%d", pipeline.PipelineCode, time.Now().Unix())

	run := &model.PipelineRun{
		PipelineID:     pipelineID,
		RunNo:          runNo,
		TriggerType:    triggerType,
		GitCommit:      gitCommit,
		GitBranch:      gitBranch,
		Status:         "pending",
		OperatorUserID: operatorUserID,
	}

	err = s.pipelineRunRepo.Create(run)
	if err != nil {
		return nil, err
	}

	// 根据流水线类型选择执行策略
	switch pipeline.PipelineType {
	case "ci":
		// CI类型：只构建打包，生成制品，不部署
		go s.executeCIPipeline(run, pipeline)
	case "cd":
		// CD类型：跳过构建，直接拉取最新制品部署到K8s
		go s.executeCDPipeline(run, pipeline)
	default:
		// CI/CD完整类型：构建完成后自动触发部署
		go s.executeCICDPipeline(run, pipeline)
	}

	return run, nil
}

// executeViaJenkins 通过Jenkins执行流水线
func (s *PipelineService) executeViaJenkins(run *model.PipelineRun, pipeline *model.Pipeline) {
	jobName := pipeline.PipelineCode

	// 确保Jenkins Job存在且配置最新
	s.ensureJenkinsJob(pipeline)
	time.Sleep(2 * time.Second)

	// 从 ConfigJSON 解析额外参数
	repoURL := ""
	serviceName := "gateway"
	if pipeline.ConfigJSON != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(pipeline.ConfigJSON), &config); err == nil {
			if url, ok := config["repoUrl"].(string); ok && url != "" {
				repoURL = url
			}
			if svc, ok := config["serviceName"].(string); ok && svc != "" {
				serviceName = svc
			}
		}
	}

	// 触发Jenkins构建
	version := fmt.Sprintf("1.0.%d", time.Now().Unix()%10000)
	commitShort := "a1b2c3d"
	if run.GitCommit != "" && len(run.GitCommit) >= 7 {
		commitShort = run.GitCommit[:7]
	}
	imageTag := fmt.Sprintf("%s-%s", version, commitShort)

	// 当分支名为空时，默认使用 main 分支（手动触发且未指定分支时）
	gitBranch := run.GitBranch
	if gitBranch == "" {
		gitBranch = "main"
		// 尝试从 pipeline config 中读取 defaultBranch
		if pipeline.ConfigJSON != "" {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(pipeline.ConfigJSON), &config); err == nil {
				if db, ok := config["defaultBranch"].(string); ok && db != "" {
					gitBranch = db
				}
			}
		}
	}

	params := map[string]string{
		"GIT_BRANCH":   gitBranch,
		"GIT_COMMIT":   run.GitCommit,
		"RUN_NO":       run.RunNo,
		"IMAGE_TAG":    imageTag,
		"GIT_REPO_URL": repoURL,
		"SERVICE_NAME": serviceName,
	}

	// 自动从 GitLab 客户端获取 Token 并传递给 Jenkins
	if s.gitlabClient != nil {
		token := s.gitlabClient.GetToken()
		if token != "" {
			params["GITLAB_TOKEN"] = token
			log.Printf("[Pipeline Jenkins] Auto-injecting GITLAB_TOKEN for private repo access")
		}
	}

	buildNumber, err := s.jenkinsClient.TriggerBuild(jobName, params)
	if err != nil {
		log.Printf("[Pipeline Jenkins] Failed to trigger build for %s: %v", jobName, err)
		s.failRun(run, fmt.Sprintf("Jenkins构建触发失败: %v", err))
		return
	}

	log.Printf("[Pipeline Jenkins] Build #%d triggered for job %s", buildNumber, jobName)

	// 更新状态为running
	now := time.Now()
	run.Status = "running"
	run.StartTime = &now
	run.LogURL = fmt.Sprintf("/jenkins/job/%s/%d/console", jobName, buildNumber)
	s.pipelineRunRepo.Update(run)

	// 等待构建完成（Docker 构建+推送可能需要较长时间）
	buildInfo, err := s.jenkinsClient.WaitForBuildComplete(jobName, buildNumber, 30*time.Minute)
	if err != nil {
		log.Printf("[Pipeline Jenkins] Build wait failed: %v", err)
		run.Status = "failed"
		endTime := time.Now()
		run.EndTime = &endTime
		run.DurationSeconds = int(endTime.Sub(*run.StartTime).Seconds())
		s.pipelineRunRepo.Update(run)
		return
	}

	// 更新最终状态
	endTime := time.Now()
	run.EndTime = &endTime
	run.DurationSeconds = int(buildInfo.Duration / 1000)

	if buildInfo.Result == "SUCCESS" {
		run.Status = "success"
		s.pipelineRunRepo.Update(run)
		log.Printf("[Pipeline Jenkins] Build #%d succeeded for %s", buildNumber, jobName)
		actualImage := fmt.Sprintf("172.18.0.1:5001/mycloud/%s:%s", pipeline.PipelineCode, imageTag)
		s.generateArtifactsWithImage(run, pipeline, actualImage)
	} else {
		run.Status = "failed"
		s.pipelineRunRepo.Update(run)
		log.Printf("[Pipeline Jenkins] Build #%d result: %s for %s", buildNumber, buildInfo.Result, jobName)
	}
}

// ensureJenkinsJob 确保Jenkins中有对应的Job
func (s *PipelineService) ensureJenkinsJob(pipeline *model.Pipeline) {
	jobName := pipeline.PipelineCode

	// 从 ConfigJSON 解析 GitLab 仓库地址和服务名
	repoURL := ""
	serviceName := "gateway"
	if pipeline.ConfigJSON != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(pipeline.ConfigJSON), &config); err == nil {
			if url, ok := config["repoUrl"].(string); ok && url != "" {
				repoURL = url
			}
			if svc, ok := config["serviceName"].(string); ok && svc != "" {
				serviceName = svc
			}
		}
	}

	// 生成 configXML
	configXML := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<project>
  <description>Pipeline: %s (auto-created by my-cloud)</description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <hudson.model.ParametersDefinitionProperty>
      <parameterDefinitions>
        <hudson.model.StringParameterDefinition>
          <name>GIT_BRANCH</name>
          <defaultValue>main</defaultValue>
        </hudson.model.StringParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>GIT_COMMIT</name>
          <defaultValue></defaultValue>
        </hudson.model.StringParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>RUN_NO</name>
          <defaultValue></defaultValue>
        </hudson.model.StringParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>IMAGE_TAG</name>
          <defaultValue></defaultValue>
        </hudson.model.StringParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>GIT_REPO_URL</name>
          <defaultValue>%s</defaultValue>
        </hudson.model.StringParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>SERVICE_NAME</name>
          <defaultValue>%s</defaultValue>
        </hudson.model.StringParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>GITLAB_TOKEN</name>
          <defaultValue></defaultValue>
        </hudson.model.StringParameterDefinition>
      </parameterDefinitions>
    </hudson.model.ParametersDefinitionProperty>
  </properties>
  <builders>
    <hudson.tasks.Shell>
      <command>
set -e
REGISTRY="host.docker.internal:5001"
IMAGE_NAME="${REGISTRY}/mycloud/%s"
IMAGE_FULL="${IMAGE_NAME}:${IMAGE_TAG}"
BUILD_DIR="/tmp/build-${RUN_NO}"

# 默认分支为 main（兼容手动触发时未传分支的情况）
GIT_BRANCH="${GIT_BRANCH:-main}"

echo "========================================="
echo "Pipeline: %s"
echo "Service: ${SERVICE_NAME}"
echo "Branch: ${GIT_BRANCH}"
echo "Commit: ${GIT_COMMIT}"
echo "Repo: ${GIT_REPO_URL}"
echo "Image: ${IMAGE_FULL}"
echo "========================================="

echo "Step 1: Cloning GitLab repository..."
if [ -n "${GITLAB_TOKEN}" ]; then
    AUTH_URL=$(echo "${GIT_REPO_URL}" | sed "s|https://|https://gitlab-ci-token:${GITLAB_TOKEN}@|")
    echo "Using GitLab token for authentication"
    git clone --branch "${GIT_BRANCH}" "${AUTH_URL}" "${BUILD_DIR}" || {
        echo "ERROR: GitLab clone failed with token, check GITLAB_TOKEN"
        exit 1
    }
else
    echo "No GITLAB_TOKEN provided, trying public clone..."
    git clone --branch "${GIT_BRANCH}" "${GIT_REPO_URL}" "${BUILD_DIR}" || {
        echo "ERROR: Failed to clone repository. If private, set GITLAB_TOKEN parameter."
        exit 1
    }
fi

cd "${BUILD_DIR}"

# 如果指定了commit，checkout到该commit
if [ -n "${GIT_COMMIT}" ]; then
    echo "Checking out commit: ${GIT_COMMIT}"
    git checkout "${GIT_COMMIT}"
fi

echo "Step 2: Building Docker Image with project Dockerfile..."
# 自动检测 Dockerfile 路径
if [ -f backend/Dockerfile ]; then
    echo "Using backend/Dockerfile"
    sed -i.bak '/^FROM golang/a ENV GOPROXY=https://goproxy.cn,direct GO111MODULE=on' backend/Dockerfile || true
    docker build \
        -f backend/Dockerfile \
        --build-arg SERVICE_NAME="${SERVICE_NAME}" \
        -t "${IMAGE_FULL}" \
        .
elif [ -f Dockerfile ]; then
    echo "Using root Dockerfile"
    sed -i.bak '/^FROM golang/a ENV GOPROXY=https://goproxy.cn,direct GO111MODULE=on' Dockerfile || true
    docker build -t "${IMAGE_FULL}" .
else
    echo "ERROR: No Dockerfile found in root or backend/"
    exit 1
fi

echo "Step 3: Pushing to Registry..."
docker push "${IMAGE_FULL}"

echo "Step 4: Notifying pipeline-service..."
EXTERNAL_IMAGE="172.18.0.1:5001/mycloud/%s:${IMAGE_TAG}"
curl -s -X POST "http://pipeline-service:8084/internal/v1/pipeline-runs/${RUN_NO}/artifact" \
  -H "Content-Type: application/json" \
  -d "{\"imageUrl\":\"${EXTERNAL_IMAGE}\"}" || true

# 清理构建目录
rm -rf "${BUILD_DIR}"
echo "========================================="
echo "BUILD SUCCESSFUL: ${IMAGE_FULL}"
echo "========================================="
      </command>
    </hudson.tasks.Shell>
  </builders>
  <publishers/>
  <buildWrappers/>
</project>`, pipeline.PipelineName, repoURL, serviceName, pipeline.PipelineCode, pipeline.PipelineName, pipeline.PipelineCode)

	if s.jenkinsClient.JobExists(jobName) {
		if err := s.jenkinsClient.UpdateJob(jobName, configXML); err != nil {
			log.Printf("[Pipeline Jenkins] Failed to update job %s: %v", jobName, err)
		} else {
			log.Printf("[Pipeline Jenkins] Updated job: %s", jobName)
		}
		return
	}

	if err := s.jenkinsClient.CreateJob(jobName, configXML); err != nil {
		log.Printf("[Pipeline Jenkins] Failed to create job %s: %v", jobName, err)
	} else {
		log.Printf("[Pipeline Jenkins] Created job: %s", jobName)
	}
}



// generateArtifacts 生成流水线制品
func (s *PipelineService) generateArtifacts(run *model.PipelineRun, pipeline *model.Pipeline) {
	version := fmt.Sprintf("1.0.%d", time.Now().Unix()%10000)
	commitShort := "a1b2c3d"
	if run.GitCommit != "" && len(run.GitCommit) >= 7 {
		commitShort = run.GitCommit[:7]
	}
	artifactName := fmt.Sprintf("mycloud/%s", pipeline.PipelineCode)
	artifactVersion := fmt.Sprintf("%s-%s", version, commitShort)
	s.generateArtifactsWithImage(run, pipeline, fmt.Sprintf("%s:%s", artifactName, artifactVersion))
}

// generateArtifactsWithImage 用指定镜像地址生成制品记录
func (s *PipelineService) generateArtifactsWithImage(run *model.PipelineRun, pipeline *model.Pipeline, imageURL string) {
	version := fmt.Sprintf("1.0.%d", time.Now().Unix()%10000)
	commitShort := "a1b2c3d"
	if run.GitCommit != "" && len(run.GitCommit) >= 7 {
		commitShort = run.GitCommit[:7]
	}
	artifactName := fmt.Sprintf("mycloud/%s", pipeline.PipelineCode)
	artifactVersion := fmt.Sprintf("%s-%s", version, commitShort)
	imageArtifact := &model.Artifact{
		PipelineRunID:   run.ID,
		ArtifactType:    "image",
		ArtifactName:    artifactName,
		ArtifactVersion: artifactVersion,
		RepoURL:         imageURL,
		Digest:          fmt.Sprintf("sha256:%x", time.Now().UnixNano()),
		MetadataJSON:    fmt.Sprintf(`{"branch":"%s","buildTool":"%s","pipelineType":"%s"}`, run.GitBranch, pipeline.CITool, pipeline.PipelineType),
	}
	if err := s.artifactRepo.Create(imageArtifact); err != nil {
		log.Printf("[Pipeline] Failed to create artifact: %v", err)
	} else {
		log.Printf("[Pipeline] Artifact created: %s (image: %s)", imageArtifact.ArtifactName, imageURL)
	}
}

// ============ 按流水线类型分支执行 ============

// failRun 标记流水线执行失败
func (s *PipelineService) failRun(run *model.PipelineRun, reason string) {
	now := time.Now()
	if run.StartTime == nil {
		run.StartTime = &now
	}
	run.Status = "failed"
	run.EndTime = &now
	run.DurationSeconds = int(now.Sub(*run.StartTime).Seconds())
	run.LogURL = fmt.Sprintf("/logs/pipeline-runs/%d/output.log", run.ID)
	s.pipelineRunRepo.Update(run)
	log.Printf("[Pipeline] Run %s failed: %s", run.RunNo, reason)
}

// executeCIPipeline CI类型：只执行构建，生成制品，不触发部署
func (s *PipelineService) executeCIPipeline(run *model.PipelineRun, pipeline *model.Pipeline) {
	log.Printf("[Pipeline CI] Starting CI pipeline: %s", run.RunNo)

	if s.jenkinsClient == nil {
		s.failRun(run, "Jenkins未连接，无法执行构建")
		return
	}
	s.executeViaJenkins(run, pipeline)
	// CI类型到此结束，制品已在构建成功后通过 generateArtifacts 生成
	log.Printf("[Pipeline CI] CI pipeline completed: %s", run.RunNo)
}

// executeCDPipeline CD类型：跳过构建，直接使用已有制品部署
func (s *PipelineService) executeCDPipeline(run *model.PipelineRun, pipeline *model.Pipeline) {
	log.Printf("[Pipeline CD] Starting CD pipeline: %s", run.RunNo)

	now := time.Now()
	run.Status = "running"
	run.StartTime = &now
	s.pipelineRunRepo.Update(run)

	// 获取该流水线最新的制品
	latestArtifact := s.getLatestArtifact(pipeline)
	if latestArtifact == nil {
		log.Printf("[Pipeline CD] No artifact found for pipeline %s, cannot deploy", pipeline.PipelineCode)
		run.Status = "failed"
		endTime := time.Now()
		run.EndTime = &endTime
		run.DurationSeconds = int(endTime.Sub(*run.StartTime).Seconds())
		run.LogURL = fmt.Sprintf("/logs/pipeline-runs/%d/output.log", run.ID)
		s.pipelineRunRepo.Update(run)
		return
	}

	log.Printf("[Pipeline CD] Using artifact: %s:%s", latestArtifact.ArtifactName, latestArtifact.ArtifactVersion)

	// 执行部署（调用deploy-service API）
	deploySuccess := s.triggerDeployment(pipeline, latestArtifact)

	endTime := time.Now()
	run.EndTime = &endTime
	run.DurationSeconds = int(endTime.Sub(*run.StartTime).Seconds())

	if deploySuccess {
		run.Status = "success"
		run.LogURL = fmt.Sprintf("/logs/pipeline-runs/%d/output.log", run.ID)
		log.Printf("[Pipeline CD] Deployment successful: %s", run.RunNo)
	} else {
		run.Status = "failed"
		run.LogURL = fmt.Sprintf("/logs/pipeline-runs/%d/output.log", run.ID)
		log.Printf("[Pipeline CD] Deployment failed: %s", run.RunNo)
	}
	s.pipelineRunRepo.Update(run)
}

// executeCICDPipeline CI/CD类型：先构建，构建成功后自动触发部署
func (s *PipelineService) executeCICDPipeline(run *model.PipelineRun, pipeline *model.Pipeline) {
	log.Printf("[Pipeline CI/CD] Starting full pipeline: %s", run.RunNo)

	// 第一阶段：CI构建
	if s.jenkinsClient == nil {
		s.failRun(run, "Jenkins未连接，无法执行构建")
		return
	}
	s.executeViaJenkins(run, pipeline)

	// 检查构建是否成功
	updatedRun, err := s.pipelineRunRepo.GetByID(run.ID)
	if err != nil || updatedRun.Status != "success" {
		log.Printf("[Pipeline CI/CD] Build phase failed, skipping deploy: %s", run.RunNo)
		return
	}

	// 第二阶段：CD部署
	log.Printf("[Pipeline CI/CD] Build succeeded, starting deploy phase: %s", run.RunNo)

	// 获取刚刚构建生成的制品
	latestArtifact := s.getLatestArtifact(pipeline)
	if latestArtifact == nil {
		log.Printf("[Pipeline CI/CD] No artifact found after build, skipping deploy: %s", run.RunNo)
		return
	}

	deploySuccess := s.triggerDeployment(pipeline, latestArtifact)
	if deploySuccess {
		log.Printf("[Pipeline CI/CD] Deploy phase succeeded: %s", run.RunNo)
	} else {
		// 部署失败，更新状态（CI阶段已经标记success，现在改为partial）
		updatedRun.Status = "partial"
		s.pipelineRunRepo.Update(updatedRun)
		log.Printf("[Pipeline CI/CD] Deploy phase failed: %s", run.RunNo)
	}
}

// getLatestArtifact 获取流水线最新的image类型制品
func (s *PipelineService) getLatestArtifact(pipeline *model.Pipeline) *model.Artifact {
	// 查找该流水线最近一次成功构建的制品
	artifacts, _, err := s.artifactRepo.List("image", 1, 10)
	if err != nil || len(artifacts) == 0 {
		return nil
	}

	// 匹配该流水线的制品（通过 artifactName 包含 pipelineCode 判断）
	expectedName := fmt.Sprintf("mycloud/%s", pipeline.PipelineCode)
	for _, a := range artifacts {
		if a.ArtifactName == expectedName {
			return a
		}
	}
	return nil
}

// triggerDeployment 触发部署（调用deploy-service内部API）
func (s *PipelineService) triggerDeployment(pipeline *model.Pipeline, artifact *model.Artifact) bool {
	// 解析pipeline配置获取部署参数
	deployNamespace := "default"
	deployClusterID := uint(1)

	if pipeline.ConfigJSON != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(pipeline.ConfigJSON), &config); err == nil {
			if ns, ok := config["namespace"].(string); ok && ns != "" {
				deployNamespace = ns
			}
			if cid, ok := config["clusterId"].(float64); ok && cid > 0 {
				deployClusterID = uint(cid)
			}
		}
	}

	imageVersion := artifact.RepoURL
	workloadName := pipeline.PipelineCode

	log.Printf("[Pipeline Deploy] Deploying %s to namespace=%s cluster=%d image=%s",
		workloadName, deployNamespace, deployClusterID, imageVersion)

	// 通过HTTP调用deploy-service
	deployURL := "http://deploy-service:8087/api/v1/deployments"
	payload := fmt.Sprintf(`{
		"releaseId": 0,
		"clusterId": %d,
		"namespace": "%s",
		"workloadName": "%s",
		"workloadType": "deployment",
		"imageVersion": "%s",
		"desiredReplicas": 1
	}`, deployClusterID, deployNamespace, workloadName, imageVersion)

	resp, err := http.Post(deployURL, "application/json", strings.NewReader(payload))
	if err != nil {
		log.Printf("[Pipeline Deploy] Failed to call deploy-service: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Pipeline Deploy] Deploy-service returned %d: %s", resp.StatusCode, string(body))
		return false
	}

	log.Printf("[Pipeline Deploy] Deployment created successfully for %s", workloadName)
	return true
}

// GetPipelineRun 获取流水线执行详情
func (s *PipelineService) GetPipelineRun(id uint) (*model.PipelineRun, error) {
	return s.pipelineRunRepo.GetByID(id)
}

// ListPipelineRuns 获取流水线执行记录列表
func (s *PipelineService) ListPipelineRuns(pipelineID uint, page, pageSize int) ([]*model.PipelineRun, int64, error) {
	return s.pipelineRunRepo.ListByPipeline(pipelineID, page, pageSize)
}

// ListAllPipelineRuns 获取所有流水线执行记录列表
func (s *PipelineService) ListAllPipelineRuns(page, pageSize int, startDate, sortBy, sortOrder string) ([]*model.PipelineRun, int64, error) {
	return s.pipelineRunRepo.ListAll(page, pageSize, startDate, sortBy, sortOrder)
}

// UpdatePipelineRunStatus 更新流水线执行状态
func (s *PipelineService) UpdatePipelineRunStatus(id uint, status string) error {
	run, err := s.pipelineRunRepo.GetByID(id)
	if err != nil {
		return errors.New("执行记录不存在")
	}

	run.Status = status
	now := time.Now()
	if status == "running" && run.StartTime == nil {
		run.StartTime = &now
	}
	if status == "success" || status == "failed" || status == "cancelled" {
		if run.EndTime == nil {
			run.EndTime = &now
		}
		if run.StartTime != nil {
			run.DurationSeconds = int(run.EndTime.Sub(*run.StartTime).Seconds())
		}
	}

	return s.pipelineRunRepo.Update(run)
}

// CreateArtifact 创建制品
func (s *PipelineService) CreateArtifact(artifact *model.Artifact) error {
	_, err := s.pipelineRunRepo.GetByID(artifact.PipelineRunID)
	if err != nil {
		return errors.New("流水线执行记录不存在")
	}
	return s.artifactRepo.Create(artifact)
}

// GetArtifact 获取制品详情
func (s *PipelineService) GetArtifact(id uint) (*model.Artifact, error) {
	return s.artifactRepo.GetByID(id)
}

// ListArtifacts 获取制品列表
func (s *PipelineService) ListArtifacts(artifactType string, page, pageSize int) ([]*model.Artifact, int64, error) {
	return s.artifactRepo.List(artifactType, page, pageSize)
}

// ListArtifactsByPipelineRun 获取流水线执行的制品列表
func (s *PipelineService) ListArtifactsByPipelineRun(pipelineRunID uint) ([]*model.Artifact, error) {
	return s.artifactRepo.ListByPipelineRun(pipelineRunID)
}

// GetArtifactsByRunID 获取流水线执行的制品列表（别名）
func (s *PipelineService) GetArtifactsByRunID(pipelineRunID uint) ([]*model.Artifact, error) {
	return s.artifactRepo.ListByPipelineRun(pipelineRunID)
}

// DeleteArtifact 删除制品
func (s *PipelineService) DeleteArtifact(id uint) error {
	return s.artifactRepo.Delete(id)
}

// DeployPipeline 手动触发部署（创建发布工单）
func (s *PipelineService) DeployPipeline(pipelineID uint, operatorUserID uint) (map[string]interface{}, error) {
	pipeline, err := s.pipelineRepo.GetByID(pipelineID)
	if err != nil {
		return nil, errors.New("流水线不存在")
	}

	// 查找该流水线最新的构建制品
	latestArtifact := s.getLatestArtifact(pipeline)
	if latestArtifact == nil {
		return nil, errors.New("没有可用的构建制品，请先执行CI构建")
	}

	// 查询应用绑定的环境
	envURL := fmt.Sprintf("http://env-service:8085/internal/v1/app-env-bindings/by-app/%d", pipeline.AppID)
	envResp, err := http.Get(envURL)
	if err != nil {
		return nil, fmt.Errorf("查询应用环境绑定失败: %v", err)
	}
	defer envResp.Body.Close()

	var envResult map[string]interface{}
	if err := json.NewDecoder(envResp.Body).Decode(&envResult); err != nil {
		return nil, fmt.Errorf("解析环境绑定数据失败: %v", err)
	}

	// 检查是否有绑定的环境
	bindings, ok := envResult["data"].([]interface{})
	if !ok || len(bindings) == 0 {
		return nil, errors.New("应用未绑定任何环境，无法创建发布工单。请先在【环境管理】中为应用绑定环境")
	}

	// 获取第一个绑定的环境ID
	firstBinding := bindings[0].(map[string]interface{})
	envID := uint(firstBinding["envId"].(float64))
	envName := firstBinding["envName"].(string)

	// 调用release-service创建发布工单（使用内部接口，无需认证）
	// 注意:创建的工单状态为'created',用户可以在发布管理中修改策略后再提交审批
	releaseURL := "http://release-service:8086/internal/v1/releases"
	payload := fmt.Sprintf(`{
		"appId": %d,
		"envId": %d,
		"releaseVersion": "%s",
		"releaseStrategy": "rolling",
		"imageUrl": "%s",
		"description": "由CI流水线 %s 自动创建，目标环境: %s，请在发布管理中选择部署策略并提交审批"
	}`, pipeline.AppID, envID, latestArtifact.ArtifactVersion, latestArtifact.RepoURL, pipeline.PipelineCode, envName)

	req, _ := http.NewRequest("POST", releaseURL, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", operatorUserID))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用发布服务失败: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode >= 400 {
		msg, _ := result["message"].(string)
		return nil, fmt.Errorf("创建发布工单失败: %s", msg)
	}

	log.Printf("[Pipeline] Release created for pipeline %s, artifact: %s", pipeline.PipelineCode, latestArtifact.RepoURL)

	return map[string]interface{}{
		"message":         "发布工单已创建，请在发布管理中审批后执行",
		"artifactVersion": latestArtifact.ArtifactVersion,
		"imageUrl":        latestArtifact.RepoURL,
		"release":         result["data"],
	}, nil
}

// UpdateLatestArtifactImage Jenkins回调：更新最新制品的实际镜像地址
// 同时处理超时恢复场景：如果Jenkins构建成功但pipeline-service等待超时已标记失败，则恢复为成功
func (s *PipelineService) UpdateLatestArtifactImage(runNo, imageURL string) error {
	run, err := s.pipelineRunRepo.GetByRunNo(runNo)
	if err != nil {
		return fmt.Errorf("pipeline run not found: %s", runNo)
	}
	artifacts, err := s.artifactRepo.ListByPipelineRun(run.ID)
	if err != nil || len(artifacts) == 0 {
		// 超时恢复场景：构建实际成功但等待超时，制品尚未创建
		// 使用 Jenkins 回调的镜像地址创建制品并恢复运行状态
		log.Printf("[Pipeline] No artifact found, creating from callback for run %s", runNo)
		artifact := &model.Artifact{
			PipelineRunID:   run.ID,
			ArtifactType:    "image",
			ArtifactName:    runNo,
			ArtifactVersion: "latest",
			RepoURL:         imageURL,
		}
		if createErr := s.artifactRepo.Create(artifact); createErr != nil {
			return fmt.Errorf("failed to create artifact for run %s: %v", runNo, createErr)
		}
		// 恢复运行状态：Jenkins成功了就标记为success
		if run.Status == "failed" {
			now := time.Now()
			run.Status = "success"
			run.EndTime = &now
			s.pipelineRunRepo.Update(run)
			log.Printf("[Pipeline] Run %s recovered to success via Jenkins callback", runNo)
		}
		return nil
	}
	artifact := artifacts[len(artifacts)-1]
	artifact.RepoURL = imageURL
	if err := s.artifactRepo.Update(artifact); err != nil {
		return fmt.Errorf("failed to update artifact: %v", err)
	}
	// 即使run已经因超时标记失败，Jenkins回调成功也应该恢复状态
	if run.Status == "failed" {
		now := time.Now()
		run.Status = "success"
		run.EndTime = &now
		s.pipelineRunRepo.Update(run)
		log.Printf("[Pipeline] Run %s recovered to success via Jenkins callback", runNo)
	}
	log.Printf("[Pipeline] Artifact %d updated with image: %s", artifact.ID, imageURL)
	return nil
}
