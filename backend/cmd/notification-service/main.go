package main

import (
	"fmt"
	"log"
	"my-cloud/internal/common/config"
	"my-cloud/pkg/database"
	"my-cloud/internal/notification/handler"
	"my-cloud/internal/notification/model"
	"my-cloud/internal/notification/repository"
	"my-cloud/internal/notification/router"
	"my-cloud/internal/notification/service"
	"my-cloud/pkg/metrics"

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

	// 自动迁移
	if err = db.AutoMigrate(
		&model.Notification{},
		&model.NotificationTemplate{},
		&model.NotificationChannel{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化Repository
	notificationRepo := repository.NewNotificationRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	channelRepo := repository.NewChannelRepository(db)

	// 初始化Service
	notificationService := service.NewNotificationService(notificationRepo, templateRepo, channelRepo)

	// 初始化Handler
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// 设置Gin
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
	r.GET("/metrics", metrics.Handler())
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 设置路由
	router.SetupRouter(r, notificationHandler)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8095
	}

	log.Printf("Notification Service starting on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
