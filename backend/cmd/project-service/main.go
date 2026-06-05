package main

import (
	"fmt"
	"log"
	"my-cloud/internal/common/config"
	"my-cloud/internal/project/router"
	"my-cloud/pkg/database"
	"my-cloud/pkg/metrics"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接 - 连接到 org_db
	dsn := "root:root123456@tcp(mysql:3306)/org_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 连接到iam_db用于权限检查
	iamDSN := "root:root123456@tcp(mysql:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
	iamDB, err := database.InitDB(iamDSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatal("连接iam_db失败:", err)
	}

	// 初始化Gin
	r := gin.Default()
	// Health check and metrics
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/metrics", metrics.Handler())

	// 注册路由
	router.RegisterRoutes(r, db, iamDB)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8082
	}
	log.Printf("Project Service 启动在端口 %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
