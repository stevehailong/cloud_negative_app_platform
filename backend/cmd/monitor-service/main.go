package main

import (
	"fmt"
	"log"

	"my-cloud/internal/common/config"
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/monitor/handler"
	"my-cloud/internal/monitor/model"
	"my-cloud/internal/monitor/repository"
	"my-cloud/internal/monitor/router"
	"my-cloud/internal/monitor/service"
	"my-cloud/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	db, err := database.InitDB(cfg.Database.DSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&model.Metric{},
		&model.AlertRule{},
		&model.Alert{},
		&model.LogQuery{},
		&model.TraceQuery{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化Repository
	monitorRepo := repository.NewMonitorRepository(db)

	// 初始化Service
	monitorService := service.NewMonitorService(monitorRepo)

	// 初始化Handler
	monitorHandler := handler.NewMonitorHandler(monitorService)

	// 初始化Gin路由
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "monitor-service"})
	})

	// 设置路由
	router.SetupRouter(r, monitorHandler)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8090
	}
	log.Printf("Monitor Service starting on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
