package handler

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"my-cloud/internal/common/model"
	"my-cloud/internal/environment/repository"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type EnvironmentHandler struct {
	envRepo     *repository.EnvironmentRepository
	templateRepo *repository.EnvTemplateRepository
	bindingRepo  *repository.AppEnvBindingRepository
	db          *gorm.DB  // env_db连接
	clusterDB   *gorm.DB  // infra_db连接（用于查询cluster）
}

func NewEnvironmentHandler(envRepo *repository.EnvironmentRepository, templateRepo *repository.EnvTemplateRepository, bindingRepo *repository.AppEnvBindingRepository, envDB *gorm.DB, clusterDB *gorm.DB) *EnvironmentHandler {
	return &EnvironmentHandler{
		envRepo:     envRepo,
		templateRepo: templateRepo,
		bindingRepo:  bindingRepo,
		db:          envDB,
		clusterDB:   clusterDB,
	}
}

// ListEnvironments 环境列表
func (h *EnvironmentHandler) ListEnvironments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	projectIDStr := c.Query("projectId")

	offset := (page - 1) * pageSize

	var projectID *uint
	if projectIDStr != "" {
		id, _ := strconv.ParseUint(projectIDStr, 10, 32)
		pid := uint(id)
		projectID = &pid
	}

	envs, total, err := h.envRepo.List(offset, pageSize, keyword, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询环境列表失败",
		})
		return
	}

	// 关联查询集群信息
	type EnvironmentWithCluster struct {
		model.Environment
		ClusterName string `json:"clusterName"`
	}

	var detailedEnvs []EnvironmentWithCluster
	for _, env := range envs {
		detail := EnvironmentWithCluster{
			Environment: env,
		}

		// 查询集群信息
		var cluster model.Cluster
		if err := h.clusterDB.Table("clusters").Where("id = ? AND is_deleted = 0", env.ClusterID).First(&cluster).Error; err == nil {
			detail.ClusterName = cluster.ClusterName
		}

		detailedEnvs = append(detailedEnvs, detail)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     detailedEnvs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateEnvironment 创建环境
func (h *EnvironmentHandler) CreateEnvironment(c *gin.Context) {
	var req model.Environment
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[CreateEnvironment] JSON绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	log.Printf("[CreateEnvironment] 收到请求: envCode=%s, namespace=%s, clusterId=%d, projectId=%d, templateId=%v", 
		req.EnvCode, req.Namespace, req.ClusterID, req.ProjectID, req.TemplateID)

	// 验证namespace格式（K8s命名规范）
	if err := validateNamespace(req.Namespace); err != nil {
		log.Printf("[CreateEnvironment] namespace验证失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40002,
			"message": err.Error(),
		})
		return
	}

	// 如果ConfigJSON为空，设置为NULL（通过空值指针）或空JSON对象
	if req.ConfigJSON == "" {
		req.ConfigJSON = "{}"
	}

	// 检查环境编码是否已存在
	existing, _ := h.envRepo.GetByCode(req.EnvCode)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "环境编码已存在",
		})
		return
	}

	if err := h.envRepo.Create(&req); err != nil {
		// 检查是否是唯一约束冲突（cluster_id + namespace）
		if strings.Contains(err.Error(), "uk_cluster_namespace") || strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    40003,
				"message": "该集群中命名空间已被占用",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建环境失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// validateNamespace 验证namespace是否符合K8s命名规范
// 规则：
// - 由小写字母、数字、短横线(-)组成
// - 必须以字母或数字开头和结尾
// - 不能包含连续的短横线
// - 长度: 1-63 字符
// - 不能使用保留名称（kube-*, default, kube-system等）
func validateNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("命名空间不能为空")
	}

	if len(namespace) > 63 {
		return fmt.Errorf("命名空间长度不能超过63个字符，当前为%d个字符", len(namespace))
	}

	// 检查是否为保留名称
	reservedPrefixes := []string{"kube-", "kubernetes-"}
	reservedNames := []string{"default", "kube-system", "kube-public", "kube-node-lease"}
	
	for _, reserved := range reservedNames {
		if namespace == reserved {
			return fmt.Errorf("不能使用保留的命名空间名称: %s", reserved)
		}
	}
	
	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(namespace, prefix) {
			return fmt.Errorf("不能使用以 %s 开头的命名空间名称", prefix)
		}
	}

	// 检查格式：小写字母、数字、短横线，必须以字母或数字开头和结尾
	pattern := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	matched, _ := regexp.MatchString(pattern, namespace)
	if !matched {
		return fmt.Errorf("命名空间格式不正确，只能包含小写字母、数字和短横线(-)，且必须以字母或数字开头和结尾")
	}

	// 检查是否包含连续的短横线
	if strings.Contains(namespace, "--") {
		return fmt.Errorf("命名空间不能包含连续的短横线(-)")
	}

	return nil
}

// GetEnvironment 获取环境详情
func (h *EnvironmentHandler) GetEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的环境ID",
		})
		return
	}

	env, err := h.envRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "环境不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    env,
	})
}

// UpdateEnvironment 更新环境
func (h *EnvironmentHandler) UpdateEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的环境ID",
		})
		return
	}

	env, err := h.envRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "环境不存在",
		})
		return
	}

	var req model.Environment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = env.CreateTime

	if err := h.envRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新环境失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteEnvironment 删除环境
func (h *EnvironmentHandler) DeleteEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的环境ID",
		})
		return
	}

	if err := h.envRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除环境失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ListTemplates 模板列表
func (h *EnvironmentHandler) ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	templateTypeStr := c.Query("templateType")

	offset := (page - 1) * pageSize

	var templateType *string
	if templateTypeStr != "" {
		templateType = &templateTypeStr
	}

	templates, total, err := h.templateRepo.List(offset, pageSize, keyword, templateType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询模板列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     templates,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateTemplate 创建模板
func (h *EnvironmentHandler) CreateTemplate(c *gin.Context) {
	var req model.EnvTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("绑定JSON失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	log.Printf("创建模板请求: TemplateCode=%s, ChartName=%s, RepoURL=%s", req.TemplateCode, req.ChartName, req.RepoURL)

	// 检查模板编码是否已存在
	existing, _ := h.templateRepo.GetByCode(req.TemplateCode)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "模板编码已存在",
		})
		return
	}

	// 设置时间戳
	now := time.Now()
	req.CreateTime = now
	req.UpdateTime = now

	if err := h.templateRepo.Create(&req); err != nil {
		log.Printf("创建模板数据库失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建模板失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetTemplate 获取模板详情
func (h *EnvironmentHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的模板ID",
		})
		return
	}

	template, err := h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "模板不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    template,
	})
}

// UpdateTemplate 更新模板
func (h *EnvironmentHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的模板ID",
		})
		return
	}

	template, err := h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "模板不存在",
		})
		return
	}

	var req model.EnvTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = template.CreateTime

	if err := h.templateRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新模板失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteTemplate 删除模板
func (h *EnvironmentHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的模板ID",
		})
		return
	}

	if err := h.templateRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除模板失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ListBindings 绑定列表
func (h *EnvironmentHandler) ListBindings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	appIDStr := c.Query("applicationId")
	if appIDStr == "" {
		appIDStr = c.Query("appId")
	}
	envIDStr := c.Query("envId")

	offset := (page - 1) * pageSize

	var appID, envID *uint
	if appIDStr != "" {
		id, _ := strconv.ParseUint(appIDStr, 10, 32)
		aid := uint(id)
		appID = &aid
	}
	if envIDStr != "" {
		id, _ := strconv.ParseUint(envIDStr, 10, 32)
		eid := uint(id)
		envID = &eid
	}

	bindings, total, err := h.bindingRepo.List(offset, pageSize, appID, envID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询绑定列表失败",
		})
		return
	}

	// 关联查询环境和集群信息
	type BindingWithDetails struct {
		model.AppEnvBinding
		EnvName     string `json:"envName"`
		EnvType     string `json:"envType"`
		Namespace   string `json:"namespace"`
		ClusterName string `json:"clusterName"`
		ConfigStatus string `json:"configStatus"`
	}

	var detailedBindings []BindingWithDetails
	for _, binding := range bindings {
		detail := BindingWithDetails{
			AppEnvBinding: binding,
			ConfigStatus: "pending",
		}

		// 查询环境信息
		var env model.Environment
		if err := h.db.Table("environments").Where("id = ? AND is_deleted = 0", binding.EnvID).First(&env).Error; err == nil {
			detail.EnvName = env.EnvName
			detail.EnvType = env.EnvType
			detail.Namespace = env.Namespace

			// 查询集群信息
			var cluster model.Cluster
			if err := h.clusterDB.Table("clusters").Where("id = ? AND is_deleted = 0", env.ClusterID).First(&cluster).Error; err == nil {
				detail.ClusterName = cluster.ClusterName
			}
		}

		// 判断配置状态
		if binding.ConfigJSON != "" && binding.ConfigJSON != "{}" {
			detail.ConfigStatus = "ready"
		}

		detailedBindings = append(detailedBindings, detail)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     detailedBindings,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateBinding 创建绑定
func (h *EnvironmentHandler) CreateBinding(c *gin.Context) {
	var req model.AppEnvBinding
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 检查绑定是否已存在
	existing, _ := h.bindingRepo.GetByAppAndEnv(req.AppID, req.EnvID)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "该应用已绑定此环境",
		})
		return
	}

	// 自动从环境继承模版ID和默认配置
	env, err := h.envRepo.GetByID(req.EnvID)
	if err == nil && env != nil {
		// 继承环境的模版ID
		if req.TemplateID == nil && env.TemplateID != nil {
			req.TemplateID = env.TemplateID
		}

		// 如果模版存在，使用模版的默认配置
		if req.TemplateID != nil {
			template, err := h.templateRepo.GetByID(*req.TemplateID)
			if err == nil && template != nil && template.ValuesYAML != "" {
				// 解析模版配置中的资源限制
				var values map[string]interface{}
				if err := yaml.Unmarshal([]byte(template.ValuesYAML), &values); err == nil {
					if resources, ok := values["resources"].(map[string]interface{}); ok {
						// 设置资源限制
						if requests, ok := resources["requests"].(map[string]interface{}); ok {
							if req.CPURequest == "" {
								if cpu, ok := requests["cpu"].(string); ok {
									req.CPURequest = cpu
								}
							}
							if req.MemoryRequest == "" {
								if mem, ok := requests["memory"].(string); ok {
									req.MemoryRequest = mem
								}
							}
						}
						if limits, ok := resources["limits"].(map[string]interface{}); ok {
							if req.CPULimit == "" {
								if cpu, ok := limits["cpu"].(string); ok {
									req.CPULimit = cpu
								}
							}
							if req.MemoryLimit == "" {
								if mem, ok := limits["memory"].(string); ok {
									req.MemoryLimit = mem
								}
							}
						}
					}
					// 设置副本数
					if req.Replicas == 0 {
						if replicaCount, ok := values["replicaCount"].(int); ok {
							req.Replicas = replicaCount
						}
					}
				}
			}
		}
	}

	if err := h.bindingRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建绑定失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetBinding 获取绑定详情
func (h *EnvironmentHandler) GetBinding(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的绑定ID",
		})
		return
	}

	binding, err := h.bindingRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "绑定不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    binding,
	})
}

// UpdateBinding 更新绑定
func (h *EnvironmentHandler) UpdateBinding(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的绑定ID",
		})
		return
	}

	binding, err := h.bindingRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "绑定不存在",
		})
		return
	}

	// 只接收允许更新的字段
	var req struct {
		Replicas      int    `json:"replicas"`
		CPURequest    string `json:"cpuRequest"`
		CPULimit      string `json:"cpuLimit"`
		MemoryRequest string `json:"memoryRequest"`
		MemoryLimit   string `json:"memoryLimit"`
		ConfigJSON    string `json:"configJson"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 只更新允许修改的字段
	binding.Replicas = req.Replicas
	binding.CPURequest = req.CPURequest
	binding.CPULimit = req.CPULimit
	binding.MemoryRequest = req.MemoryRequest
	binding.MemoryLimit = req.MemoryLimit
	binding.ConfigJSON = req.ConfigJSON

	if err := h.bindingRepo.Update(binding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新绑定失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    binding,
	})
}

// DeleteBinding 删除绑定
func (h *EnvironmentHandler) DeleteBinding(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的绑定ID",
		})
		return
	}

	if err := h.bindingRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除绑定失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

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

