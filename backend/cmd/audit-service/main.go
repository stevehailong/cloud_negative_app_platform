package main

import (
	"fmt"
	"log"
	"my-cloud/internal/audit/handler"
	"my-cloud/internal/audit/repository"
	"my-cloud/internal/audit/router"
	"my-cloud/internal/audit/service"
	"my-cloud/internal/common/config"
	"my-cloud/pkg/database"
	"my-cloud/internal/common/model"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接
	db, err := database.InitDB(cfg.Database.DSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 切换到audit_db数据库
	db.Exec("USE audit_db")

	// 自动迁移
	err = db.AutoMigrate(&model.AuditLog{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化Repository
	auditRepo := repository.NewAuditRepository(db)

	// 初始化Service
	auditService := service.NewAuditService(auditRepo)

	// 初始化Handler
	auditHandler := handler.NewAuditHandler(auditService)

	// 设置Gin
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 设置路由
	router.SetupRouter(r, auditHandler)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8093
	}

	log.Printf("Audit Service starting on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
