package main

import (
	"fmt"
	"log"

	"my-cloud/internal/common/config"
	"my-cloud/internal/cost/handler"
	"my-cloud/internal/cost/repository"
	"my-cloud/internal/cost/router"
	"my-cloud/pkg/database"
	"my-cloud/pkg/metrics"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接 — 优先使用 DB_DSN 环境变量，否则使用配置文件中的 dsn
	dsn := cfg.Database.DSN
	if dsn == "" {
		log.Fatal("数据库 DSN 未配置: 请设置 DB_DSN 环境变量或在 configs/config.yaml 中配置 database.dsn")
	}
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 初始化仓储层
	costRepo := repository.NewCostRepository(db)

	// 初始化处理器
	costHandler := handler.NewCostHandler(costRepo, db)

	// 初始化 Gin 路由
	r := gin.Default()

	// Health check and metrics
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/metrics", metrics.Handler())

	// 添加中间件确保正确的 Content-Type
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// 注册路由
	router.SetupRouter(r, costHandler)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8099
	}
	log.Printf("Cost Service 启动在端口 %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
