package main

import (
	"fmt"
	"log"
	"my-cloud/internal/application/handler"
	"my-cloud/internal/application/repository"
	"my-cloud/internal/application/router"
	"my-cloud/internal/application/service"
	"my-cloud/internal/common/config"
	"my-cloud/pkg/database"
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/common/response"
	"my-cloud/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化JWT
	jwt.InitJWT(cfg.JWT.Secret)

	// 初始化数据库
	db, err := database.InitDB(cfg.Database.DSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化仓储层
	appRepo := repository.NewApplicationRepository(db)
	componentRepo := repository.NewComponentRepository(db)

	// 初始化服务层
	appService := service.NewApplicationService(appRepo, componentRepo)

	// 初始化处理层
	appHandler := handler.NewApplicationHandler(appService)

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())
	r.Use(middleware.RequestID())
	r.Use(middleware.Auth())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// 注册路由
	router.RegisterRoutes(r, appHandler)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Application service is running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
