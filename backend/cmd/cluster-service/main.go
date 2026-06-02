package main

import (
	"fmt"
	"log"
	"my-cloud/internal/cluster/handler"
	"my-cloud/internal/cluster/repository"
	"my-cloud/internal/cluster/router"
	"my-cloud/internal/common/config"
	"my-cloud/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接 - 连接到 infra_db
	dsn := "root:root123456@tcp(mysql:3306)/infra_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 初始化仓储层
	clusterRepo := repository.NewClusterRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	namespaceRepo := repository.NewNamespaceRepository(db)

	// 初始化处理器
	clusterHandler := handler.NewClusterHandler(clusterRepo, nodeRepo, namespaceRepo)

	// 初始化 Gin 路由
	r := gin.Default()

	// 添加中间件确保正确的Content-Type
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// 注册路由
	router.RegisterRoutes(r, clusterHandler)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8088
	}
	log.Printf("Cluster Service 启动在端口 %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
