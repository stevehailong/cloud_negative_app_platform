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

	// 初始化数据库连接 - 连接到 cost_db
	dsn := "root:root123456@tcp(mysql:3306)/cost_db?charset=utf8mb4&parseTime=True&loc=Local"
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
