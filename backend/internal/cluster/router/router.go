package router

import (
	"my-cloud/internal/cluster/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handler.ClusterHandler) {
	v1 := r.Group("/api/v1")
	{
		// 集群管理
		clusterGroup := v1.Group("/clusters")
		{
			clusterGroup.GET("", h.ListClusters)
			clusterGroup.POST("", h.CreateCluster)
			clusterGroup.GET("/:id", h.GetCluster)
			clusterGroup.PUT("/:id", h.UpdateCluster)
			clusterGroup.DELETE("/:id", h.DeleteCluster)
		}

		// 节点管理
		nodeGroup := v1.Group("/nodes")
		{
			nodeGroup.GET("", h.ListNodes)
			nodeGroup.POST("", h.CreateNode)
			nodeGroup.GET("/:id", h.GetNode)
			nodeGroup.PUT("/:id", h.UpdateNode)
			nodeGroup.DELETE("/:id", h.DeleteNode)
		}

		// 命名空间管理
		namespaceGroup := v1.Group("/namespaces")
		{
			namespaceGroup.GET("", h.ListNamespaces)
			namespaceGroup.POST("", h.CreateNamespace)
			namespaceGroup.GET("/:id", h.GetNamespace)
			namespaceGroup.PUT("/:id", h.UpdateNamespace)
			namespaceGroup.DELETE("/:id", h.DeleteNamespace)
		}
	}
}
