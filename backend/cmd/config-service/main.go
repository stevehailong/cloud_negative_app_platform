package main

import (
	"fmt"
	"log"

	"my-cloud/internal/common/config"
	"my-cloud/internal/common/model"
	"my-cloud/internal/config/handler"
	"my-cloud/internal/config/repository"
	"my-cloud/internal/config/router"
	"my-cloud/pkg/database"
	"my-cloud/pkg/metrics"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接 - 连接到 config_db
	dsn := "root:root123456@tcp(mysql:3306)/config_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&model.AppConfig{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化 Repository
	configRepo := repository.NewConfigRepository(db)

	// 初始化 Handler
	configHandler := handler.NewConfigHandler(configRepo)

	// 设置 Gin
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Prometheus /metrics endpoint
	r.GET("/metrics", metrics.Handler())

	// 设置路由
	router.SetupRouter(r, configHandler)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8097
	}

	log.Printf("Config Service starting on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
