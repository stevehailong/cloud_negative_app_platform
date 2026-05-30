package handler

import (
	"net/http"
	"strconv"

	"my-cloud/internal/cluster/repository"
	"my-cloud/internal/common/model"

	"github.com/gin-gonic/gin"
)

type ClusterHandler struct {
	clusterRepo   *repository.ClusterRepository
	nodeRepo      *repository.NodeRepository
	namespaceRepo *repository.NamespaceRepository
}

func NewClusterHandler(clusterRepo *repository.ClusterRepository, nodeRepo *repository.NodeRepository, namespaceRepo *repository.NamespaceRepository) *ClusterHandler {
	return &ClusterHandler{
		clusterRepo:   clusterRepo,
		nodeRepo:      nodeRepo,
		namespaceRepo: namespaceRepo,
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
