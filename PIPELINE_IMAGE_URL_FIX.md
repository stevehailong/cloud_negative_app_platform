# 流水线镜像地址显示修复

## 问题描述
流水线执行记录列表中没有显示"构建镜像"列和镜像地址。

## 根本原因
1. **数据库表缺失**: `devops_db` 数据库中的 `artifacts` 表在之前的部署中已自动创建（通过 GORM AutoMigrate）
2. **后端响应缺少字段**: `PipelineRun` 模型中没有 `imageUrl` 字段，需要从关联的 `Artifact` 表查询并填充

## 修复内容

### 1. 后端修改

#### a. 修改 `/backend/internal/pipeline/handler/pipeline_handler.go`

```go
// 添加响应 DTO 结构体
type PipelineRunResponse struct {
	*model.PipelineRun
	ImageURL     string `json:"imageUrl,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// 修改 ListPipelineRuns 方法
func (h *PipelineHandler) ListPipelineRuns(c *gin.Context) {
	// ... 查询 runs ...
	
	// 为每个 run 查询关联的 image artifact
	runResponses := make([]*PipelineRunResponse, 0, len(runs))
	for _, run := range runs {
		runResp := &PipelineRunResponse{
			PipelineRun: run,
		}
		
		// 查询该 run 的 artifacts，找到 type=image 的制品
		artifacts, err := h.pipelineService.GetArtifactsByRunID(run.ID)
		if err == nil {
			for _, artifact := range artifacts {
				if artifact.ArtifactType == "image" {
					runResp.ImageURL = artifact.RepoURL
					break
				}
			}
		}
		
		runResponses = append(runResponses, runResp)
	}

	response.SuccessWithPage(c, total, page, pageSize, runResponses)
}
```

#### b. 添加服务方法 `/backend/internal/pipeline/service/pipeline_service.go`

```go
// GetArtifactsByRunID 获取流水线执行的制品列表（别名）
func (s *PipelineService) GetArtifactsByRunID(pipelineRunID uint) ([]*model.Artifact, error) {
	return s.artifactRepo.ListByPipelineRun(pipelineRunID)
}
```

### 2. 数据库表结构

创建了 `/sql/20_pipeline_db_tables.sql` 文件（备用，实际 devops_db 中的表已通过 GORM AutoMigrate 自动创建）：

```sql
-- 制品表
CREATE TABLE IF NOT EXISTS artifacts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    pipeline_run_id INT UNSIGNED NOT NULL,
    artifact_type VARCHAR(32) NOT NULL COMMENT 'image/chart/package/sbom/report',
    artifact_name VARCHAR(128) NOT NULL,
    artifact_version VARCHAR(64),
    repo_url VARCHAR(255),
    digest VARCHAR(255),
    metadata_json JSON,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_pipeline_run_id (pipeline_run_id),
    INDEX idx_artifact_type (artifact_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='制品表';
```

### 3. 前端代码

前端代码 `/frontend/src/views/pipeline/components/PipelineRuns.vue` 已经正确实现：
- 第79-96行：显示"构建镜像"列
- 第81-94行：根据 `row.imageUrl` 显示镜像地址和复制按钮
- 第340-360行：实现了复制到剪贴板功能

## 数据流程

1. **Jenkins 构建完成后**:
   ```bash
   # Jenkins 脚本会调用回调接口更新 artifact
   curl -X POST "http://pipeline-service:8084/internal/v1/pipeline-runs/${RUN_NO}/artifact" \
     -H "Content-Type: application/json" \
     -d "{\"imageUrl\":\"172.18.0.1:5001/mycloud/xxx:tag\"}"
   ```

2. **后端处理**:
   - `JenkinsBuildCallback` 接收镜像地址
   - `UpdateLatestArtifactImage` 更新 artifact 表的 `repo_url` 字段

3. **前端查询**:
   - 调用 `/api/v1/pipelines/:id/runs` 接口
   - 后端从 `artifacts` 表查询 `artifact_type='image'` 的记录
   - 将 `repo_url` 作为 `imageUrl` 返回给前端

4. **前端展示**:
   - 在"构建镜像"列显示完整镜像地址
   - 提供复制按钮，点击可复制到剪贴板

## 验证方法

### 1. 数据库验证
```sql
-- 检查 artifacts 表中的镜像记录
SELECT 
    pr.id, 
    pr.run_no, 
    pr.status, 
    a.artifact_type, 
    a.repo_url as image_url 
FROM pipeline_runs pr 
LEFT JOIN artifacts a ON pr.id = a.pipeline_run_id 
WHERE a.artifact_type = 'image' 
ORDER BY pr.id DESC 
LIMIT 5;
```

预期结果：能看到镜像地址如 `172.18.0.1:5001/mycloud/book-service-ci:1.0.882-a1b2c3d`

### 2. 前端验证
1. 打开浏览器访问 `http://localhost:3000`
2. 进入"流水线管理"页面
3. 点击某个流水线查看执行记录
4. 检查"构建镜像"列是否显示镜像地址
5. 点击复制按钮，验证是否成功复制到剪贴板

### 3. API 验证
```bash
# 通过网关访问（需要登录 token）
curl 'http://localhost:8080/api/v1/pipelines/1/runs?page=1&pageSize=5' \
  -H "Authorization: Bearer YOUR_TOKEN"
```

预期响应包含 `imageUrl` 字段：
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 52,
        "runNo": "book-service-ci-1780320879",
        "status": "success",
        "imageUrl": "172.18.0.1:5001/mycloud/book-service-ci:1.0.882-a1b2c3d",
        ...
      }
    ],
    "total": 10,
    "page": 1,
    "pageSize": 5
  }
}
```

## 部署步骤

1. 重新构建并重启 pipeline-service:
   ```bash
   cd /Users/hanhailong01/Downloads/my_cloud
   docker-compose build pipeline-service
   docker-compose restart pipeline-service
   ```

2. 验证服务启动:
   ```bash
   docker-compose logs --tail=20 pipeline-service
   ```

3. 刷新浏览器页面，查看流水线执行记录

## 测试数据
数据库中已有的测试记录：
- run_no: `book-service-ci-1780320879`, imageUrl: `172.18.0.1:5001/mycloud/book-service-ci:1.0.882-a1b2c3d`
- run_no: `book-service-ci-1780101624`, imageUrl: `172.18.0.1:5001/mycloud/book-service-ci:1.0.1626-e062150`

## 注意事项
1. 只有状态为 `success` 的构建才会有镜像地址
2. 失败的构建（status='failed'）不会生成 artifact 记录
3. 运行中的构建（status='running'）显示"构建中..."
4. 镜像地址格式: `172.18.0.1:5001/mycloud/<pipeline-code>:<version>-<commit>`
