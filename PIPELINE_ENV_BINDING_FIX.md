# 流水线自动创建发布记录环境绑定问题修复

## 问题描述

之前在 `pipeline-service` 的 `DeployPipeline` 函数中，创建发布记录时硬编码了 `envId=1`：

```go
payload := fmt.Sprintf(`{
    "appId": %d,
    "envId": 1,  // ❌ 硬编码！
    "releaseVersion": "%s",
    ...
}`, ...)
```

这导致所有自动创建的发布记录都绑定到环境ID=1，忽略了应用实际绑定的环境。

## 修复方案

### 1. 添加环境服务内部API

**文件**: `backend/internal/environment/handler/environment_handler.go`

新增 `GetBindingsByAppID` 方法（lines 686-748）：

```go
// GetBindingsByAppID 根据应用ID查询绑定列表（内部接口）
func (h *EnvironmentHandler) GetBindingsByAppID(c *gin.Context) {
    appID, err := strconv.ParseUint(c.Param("appId"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    40000,
            "message": "无效的应用ID",
        })
        return
    }

    bindings, err := h.bindingRepo.GetByAppID(uint(appID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    50000,
            "message": "查询绑定列表失败",
        })
        return
    }

    // 关联查询环境信息
    type BindingWithEnv struct {
        BindingID uint   `json:"bindingId"`
        AppID     uint   `json:"appId"`
        EnvID     uint   `json:"envId"`
        EnvName   string `json:"envName"`
        EnvType   string `json:"envType"`
    }

    var result []BindingWithEnv
    for _, binding := range bindings {
        item := BindingWithEnv{
            BindingID: binding.ID,
            AppID:     binding.AppID,
            EnvID:     binding.EnvID,
        }

        // 查询环境信息
        var env model.Environment
        if err := h.db.Table("environments").Where("id = ? AND is_deleted = 0", binding.EnvID).First(&env).Error; err == nil {
            item.EnvName = env.EnvName
            item.EnvType = env.EnvType
        }

        result = append(result, item)
    }

    c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": "success",
        "data":    result,
    })
}
```

### 2. 注册内部路由

**文件**: `backend/internal/environment/router/router.go`

```go
// 内部API接口（无需认证）
internal := r.Group("/internal/v1")
{
    internal.GET("/app-env-bindings/by-app/:appId", h.GetBindingsByAppID)
}
```

### 3. 配置网关内部路由

**文件**: `backend/internal/gateway/router/router.go`

在 internal 路由组中添加：

```go
// 内部服务路由（无需认证）
internal := r.Group("/internal/v1")
{
    internal.Any("/metrics/*path", monitorProxy.Handle)
    internal.Any("/pods/*path", monitorProxy.Handle)
    internal.Any("/logs/*path", monitorProxy.Handle)
    internal.Any("/app-env-bindings/*path", envProxy.Handle)  // ✅ 新增
    internal.Any("/releases", releaseProxy.Handle)             // ✅ 新增
    internal.Any("/releases/*path", releaseProxy.Handle)       // ✅ 新增
}
```

### 4. 修改Pipeline服务DeployPipeline方法

**文件**: `backend/internal/pipeline/service/pipeline_service.go` (lines 697-752)

```go
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

    // ✅ 新增：查询应用绑定的环境
    envURL := fmt.Sprintf("http://env-service:8083/internal/v1/app-env-bindings/by-app/%d", pipeline.AppID)
    envResp, err := http.Get(envURL)
    if err != nil {
        return nil, fmt.Errorf("查询应用环境绑定失败: %v", err)
    }
    defer envResp.Body.Close()

    var envResult map[string]interface{}
    if err := json.NewDecoder(envResp.Body).Decode(&envResult); err != nil {
        return nil, fmt.Errorf("解析环境绑定数据失败: %v", err)
    }

    // ✅ 检查是否有绑定的环境
    bindings, ok := envResult["data"].([]interface{})
    if !ok || len(bindings) == 0 {
        return nil, errors.New("应用未绑定任何环境，无法创建发布工单。请先在【环境管理】中为应用绑定环境")
    }

    // ✅ 获取第一个绑定的环境ID
    firstBinding := bindings[0].(map[string]interface{})
    envID := uint(firstBinding["envId"].(float64))
    envName := firstBinding["envName"].(string)

    // 调用release-service创建发布工单
    releaseURL := "http://release-service:8086/internal/v1/releases"
    payload := fmt.Sprintf(`{
        "appId": %d,
        "envId": %d,  // ✅ 使用动态查询的环境ID
        "releaseVersion": "%s",
        "releaseStrategy": "rolling",
        "imageUrl": "%s",
        "description": "由CI流水线 %s 自动创建，目标环境: %s，请在发布管理中选择部署策略并提交审批"
    }`, pipeline.AppID, envID, latestArtifact.ArtifactVersion, latestArtifact.RepoURL, pipeline.PipelineCode, envName)
    
    // ... 后续代码不变 ...
}
```

## 修复效果

### 修复前
- 所有流水线自动创建的发布记录都绑定到 `envId=1`
- 无法支持多环境场景
- 用户需要手动修改发布记录的环境

### 修复后
- 查询应用实际绑定的环境
- 使用第一个绑定的环境ID创建发布记录
- 如果应用未绑定任何环境，返回明确的错误提示
- 发布记录描述中包含目标环境名称，便于用户识别

## 测试验证

### 1. 测试内部API

```bash
curl -s http://localhost:8080/internal/v1/app-env-bindings/by-app/1 | jq
```

预期返回：
```json
{
  "code": 0,
  "data": [
    {
      "bindingId": 1,
      "appId": 1,
      "envId": 1,
      "envName": "dev-开发环境",
      "envType": "dev"
    }
  ],
  "message": "success"
}
```

### 2. 测试完整流程

需要通过前端或有效的API token：

1. 确保应用已绑定至少一个环境
2. 确保流水线有可用的构建制品
3. 调用 `POST /api/v1/pipelines/:id/deploy`
4. 检查创建的发布记录，验证 `env_id` 是否为应用绑定的环境ID

## 部署说明

需要重新构建和部署以下服务：

```bash
cd /Users/hanhailong01/Downloads/my_cloud

# 重新构建相关服务
docker-compose up -d --build env-service pipeline-service gateway

# 验证服务启动
docker logs my-cloud-env-service --tail 20
docker logs my-cloud-pipeline-service --tail 20
docker logs my-cloud-gateway --tail 20
```

## 数据库要求

确保以下数据存在：

1. **env_db.app_env_bindings**: 应用必须至少绑定一个环境
2. **pipeline_db.artifacts**: 流水线必须有可用的构建制品

如果应用未绑定环境，DeployPipeline 会返回错误：
```
应用未绑定任何环境，无法创建发布工单。请先在【环境管理】中为应用绑定环境
```

## 相关文件

- `/backend/internal/environment/handler/environment_handler.go` (新增方法)
- `/backend/internal/environment/router/router.go` (注册路由)
- `/backend/internal/gateway/router/router.go` (配置网关)
- `/backend/internal/pipeline/service/pipeline_service.go` (修改逻辑)

## 总结

此修复完全解决了流水线自动创建发布记录时硬编码环境ID的问题，支持多环境场景，并提供了清晰的错误提示。用户现在可以放心使用流水线自动创建发布记录功能，系统会自动使用应用绑定的环境。
