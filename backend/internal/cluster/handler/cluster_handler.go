package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"my-cloud/internal/cluster/repository"
	"my-cloud/internal/common/model"
	"my-cloud/pkg/k8s"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterHandler struct {
	clusterRepo   *repository.ClusterRepository
	nodeRepo      *repository.NodeRepository
	namespaceRepo *repository.NamespaceRepository
	k8sClient     *k8s.Client
	db            *gorm.DB
}

func NewClusterHandler(clusterRepo *repository.ClusterRepository, nodeRepo *repository.NodeRepository, namespaceRepo *repository.NamespaceRepository, k8sClient *k8s.Client, db *gorm.DB) *ClusterHandler {
	return &ClusterHandler{
		clusterRepo:   clusterRepo,
		nodeRepo:      nodeRepo,
		namespaceRepo: namespaceRepo,
		k8sClient:     k8sClient,
		db:            db,
	}
}

// ListClusters 集群列表
func (h *ClusterHandler) ListClusters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	clusterTypeStr := c.Query("clusterType")

	offset := (page - 1) * pageSize

	var clusterType *string
	if clusterTypeStr != "" {
		clusterType = &clusterTypeStr
	}

	clusters, total, err := h.clusterRepo.List(offset, pageSize, keyword, clusterType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询集群列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     clusters,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateCluster 创建集群
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	var req model.Cluster
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 检查集群编码是否已存在
	existing, _ := h.clusterRepo.GetByCode(req.ClusterCode)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "集群编码已存在",
		})
		return
	}

	if err := h.clusterRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建集群失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetCluster 获取集群详情
func (h *ClusterHandler) GetCluster(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的集群ID",
		})
		return
	}

	cluster, err := h.clusterRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "集群不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    cluster,
	})
}

// UpdateCluster 更新集群
func (h *ClusterHandler) UpdateCluster(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的集群ID",
		})
		return
	}

	cluster, err := h.clusterRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "集群不存在",
		})
		return
	}

	var req model.Cluster
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = cluster.CreateTime

	if err := h.clusterRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新集群失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteCluster 删除集群
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的集群ID",
		})
		return
	}

	if err := h.clusterRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除集群失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ListNodes 节点列表
func (h *ClusterHandler) ListNodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	clusterIDStr := c.Query("clusterId")
	nodeRoleStr := c.Query("nodeRole")

	offset := (page - 1) * pageSize

	var clusterID *uint
	var nodeRole *string

	if clusterIDStr != "" {
		id, _ := strconv.ParseUint(clusterIDStr, 10, 32)
		cid := uint(id)
		clusterID = &cid
	}
	if nodeRoleStr != "" {
		nodeRole = &nodeRoleStr
	}

	nodes, total, err := h.nodeRepo.List(offset, pageSize, clusterID, nodeRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询节点列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     nodes,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateNode 创建节点
func (h *ClusterHandler) CreateNode(c *gin.Context) {
	var req model.ClusterNode
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	if err := h.nodeRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建节点失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetNode 获取节点详情
func (h *ClusterHandler) GetNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的节点ID",
		})
		return
	}

	node, err := h.nodeRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "节点不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    node,
	})
}

// UpdateNode 更新节点
func (h *ClusterHandler) UpdateNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的节点ID",
		})
		return
	}

	node, err := h.nodeRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "节点不存在",
		})
		return
	}

	var req model.ClusterNode
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = node.CreateTime

	if err := h.nodeRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新节点失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteNode 删除节点
func (h *ClusterHandler) DeleteNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的节点ID",
		})
		return
	}

	if err := h.nodeRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除节点失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ListNamespaces 命名空间列表
func (h *ClusterHandler) ListNamespaces(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	clusterIDStr := c.Query("clusterId")
	projectIDStr := c.Query("projectId")

	offset := (page - 1) * pageSize

	var clusterID, projectID *uint

	if clusterIDStr != "" {
		id, _ := strconv.ParseUint(clusterIDStr, 10, 32)
		cid := uint(id)
		clusterID = &cid
	}
	if projectIDStr != "" {
		id, _ := strconv.ParseUint(projectIDStr, 10, 32)
		pid := uint(id)
		projectID = &pid
	}

	namespaces, total, err := h.namespaceRepo.List(offset, pageSize, clusterID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "查询命名空间列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":     namespaces,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// CreateNamespace 创建命名空间
func (h *ClusterHandler) CreateNamespace(c *gin.Context) {
	var req model.Namespace
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	// 检查命名空间是否已存在
	existing, _ := h.namespaceRepo.GetByClusterAndName(req.ClusterID, req.NamespaceName)
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "该集群中命名空间已存在",
		})
		return
	}

	if err := h.namespaceRepo.Create(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "创建命名空间失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    req,
	})
}

// GetNamespace 获取命名空间详情
func (h *ClusterHandler) GetNamespace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的命名空间ID",
		})
		return
	}

	ns, err := h.namespaceRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "命名空间不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    ns,
	})
}

// UpdateNamespace 更新命名空间
func (h *ClusterHandler) UpdateNamespace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的命名空间ID",
		})
		return
	}

	ns, err := h.namespaceRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "命名空间不存在",
		})
		return
	}

	var req model.Namespace
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "请求参数错误",
		})
		return
	}

	req.ID = uint(id)
	req.CreateTime = ns.CreateTime

	if err := h.namespaceRepo.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "更新命名空间失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    req,
	})
}

// DeleteNamespace 删除命名空间
func (h *ClusterHandler) DeleteNamespace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "无效的命名空间ID",
		})
		return
	}

	if err := h.namespaceRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "删除命名空间失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ========== K8s 集成功能 ==========

// SyncNodes 从 K8s 同步节点信息到数据库
func (h *ClusterHandler) SyncNodes(c *gin.Context) {
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": "无效的集群ID"})
		return
	}

	if h.k8sClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 50001, "message": "K8s客户端未初始化"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	nodes, err := h.k8sClient.GetClientset().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "message": "获取K8s节点失败: " + err.Error()})
		return
	}

	synced := 0
	for _, node := range nodes.Items {
		nodeIP := ""
		nodeRole := "worker"
		for _, addr := range node.Status.Addresses {
			if addr.Type == "InternalIP" {
				nodeIP = addr.Address
				break
			}
		}
		// 判断角色
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			nodeRole = "control-plane"
		} else if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			nodeRole = "master"
		}

		cpuCores := int(node.Status.Capacity.Cpu().Value())
		memGB := int(node.Status.Capacity.Memory().Value() / (1024 * 1024 * 1024))
		diskGB := int(node.Status.Capacity.StorageEphemeral().Value() / (1024 * 1024 * 1024))
		kubeletVer := node.Status.NodeInfo.KubeletVersion
		osImage := node.Status.NodeInfo.OSImage
		cr := node.Status.NodeInfo.ContainerRuntimeVersion

		// Upsert: 按 cluster_id + node_name 查找
		existing, _ := h.nodeRepo.GetByClusterAndName(uint(clusterID), node.Name)
		if existing != nil {
			existing.NodeIP = nodeIP
			existing.NodeRole = nodeRole
			existing.CPUCores = cpuCores
			existing.MemoryGB = memGB
			existing.DiskGB = diskGB
			existing.OSImage = osImage
			existing.ContainerRuntime = cr
			existing.KubeletVersion = kubeletVer
			existing.Status = 1
			existing.UpdateTime = time.Now()
			_ = h.nodeRepo.Update(existing)
		} else {
			now := time.Now()
			newNode := &model.ClusterNode{
				ClusterID:        uint(clusterID),
				NodeName:         node.Name,
				NodeIP:           nodeIP,
				NodeRole:         nodeRole,
				CPUCores:         cpuCores,
				MemoryGB:         memGB,
				DiskGB:           diskGB,
				OSImage:          osImage,
				ContainerRuntime: cr,
				KubeletVersion:   kubeletVer,
				Status:           1,
				CreateTime:       now,
				UpdateTime:       now,
			}
			_ = h.nodeRepo.Create(newNode)
		}
		synced++
	}

	// 更新集群版本号
	k8sVer, _ := h.k8sClient.GetClientset().ServerVersion()
	if k8sVer != nil {
		versionStr := k8sVer.GitVersion
		_ = h.clusterRepo.UpdateVersion(uint(clusterID), versionStr)
	}

	log.Printf("[Cluster] Synced %d nodes for cluster %d", synced, clusterID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "同步成功", "data": gin.H{"syncedNodes": synced}})
}

// GetClusterStats 获取集群资源统计
func (h *ClusterHandler) GetClusterStats(c *gin.Context) {
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": "无效的集群ID"})
		return
	}

	var totalCPU, totalMem, totalDisk int64
	var nodeCount int

	nodes, _, _ := h.nodeRepo.List(0, 1000, (*uint)(&[]uint{uint(clusterID)}[0]), nil)
	for _, n := range nodes {
		totalCPU += int64(n.CPUCores)
		totalMem += int64(n.MemoryGB)
		totalDisk += int64(n.DiskGB)
		nodeCount++
	}

	// 从 K8s 获取 Pod 统计
	podStats := gin.H{"total": 0, "running": 0, "pending": 0, "failed": 0}
	if h.k8sClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pods, err := h.k8sClient.GetClientset().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err == nil {
			podStats["total"] = len(pods.Items)
			for _, p := range pods.Items {
				switch p.Status.Phase {
				case "Running":
					podStats["running"] = podStats["running"].(int) + 1
				case "Pending":
					podStats["pending"] = podStats["pending"].(int) + 1
				case "Failed":
					podStats["failed"] = podStats["failed"].(int) + 1
				}
			}
		}
	}

	// 命名空间统计
	nsCount, _ := h.namespaceRepo.CountByCluster(uint(clusterID))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"nodeCount":   nodeCount,
			"totalCPU":    totalCPU,
			"totalMemGB":  totalMem,
			"totalDiskGB": totalDisk,
			"podStats":    podStats,
			"nsCount":     nsCount,
		},
	})
}

// SyncNamespaces 从 K8s 同步命名空间
func (h *ClusterHandler) SyncNamespaces(c *gin.Context) {
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": "无效的集群ID"})
		return
	}

	if h.k8sClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 50001, "message": "K8s客户端未初始化"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	nsList, err := h.k8sClient.GetClientset().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "message": "获取命名空间失败: " + err.Error()})
		return
	}

	synced := 0
	for _, ns := range nsList.Items {
		// 查找 namespace 关联的 app 和 project
		projectID, appName := h.lookupNamespaceApp(ns.Name)

		existing, _ := h.namespaceRepo.GetByClusterAndName(uint(clusterID), ns.Name)
		desc := "K8s namespace: " + ns.Name
		if appName != "" {
			desc = "应用: " + appName + " | " + desc
		}

		if existing != nil {
			if projectID > 0 {
				existing.ProjectID = projectID
			}
			existing.Description = desc
			existing.Status = 1
			existing.UpdateTime = time.Now()
			_ = h.namespaceRepo.Update(existing)
		} else {
			now := time.Now()
			newNS := &model.Namespace{
				ClusterID:         uint(clusterID),
				NamespaceName:     ns.Name,
				ProjectID:         projectID,
				ResourceQuotaJSON: "{}",
				LimitRangeJSON:    "{}",
				Description:       desc,
				Status:            1,
				CreateTime:        now,
				UpdateTime:        now,
			}
			_ = h.namespaceRepo.Create(newNS)
		}
		synced++
	}

	log.Printf("[Cluster] Synced %d namespaces for cluster %d", synced, clusterID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "同步成功", "data": gin.H{"syncedNamespaces": synced}})
}

// HealthCheck 检测集群连通性
func (h *ClusterHandler) HealthCheck(c *gin.Context) {
	if h.k8sClient == nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"healthy": false, "reason": "K8s客户端未初始化"}})
		return
	}

	// 尝试获取 API Server 版本
	ver, err := h.k8sClient.GetClientset().ServerVersion()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"healthy": false, "reason": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"healthy":     true,
			"version":     ver.GitVersion,
			"platform":    ver.Platform,
			"checkedAt":   time.Now(),
		},
	})
}

// lookupNamespaceApp 通过 namespace 查找对应的 project_id 和 app 名称
func (h *ClusterHandler) lookupNamespaceApp(namespace string) (uint, string) {
	if h.db == nil {
		return 0, ""
	}
	var appID uint
	err := h.db.Raw("SELECT app_id FROM deploy_db.app_deployments WHERE namespace = ? LIMIT 1", namespace).Scan(&appID).Error
	if err != nil || appID == 0 {
		return 0, ""
	}
	var result struct {
		ProjectID uint
		AppName   string
	}
	err = h.db.Raw("SELECT project_id, name AS app_name FROM app_db.applications WHERE id = ? LIMIT 1", appID).Scan(&result).Error
	if err != nil {
		return 0, ""
	}
	return result.ProjectID, result.AppName
}
