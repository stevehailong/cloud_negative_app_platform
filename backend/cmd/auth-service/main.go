package main

import (
	"fmt"
	"log"
	"my-cloud/internal/auth/handler"
	"my-cloud/internal/auth/repository"
	"my-cloud/internal/auth/router"
	"my-cloud/internal/auth/service"
	"my-cloud/internal/common/config"
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/common/response"
	"my-cloud/pkg/database"
	"my-cloud/pkg/security"

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

	// 初始化安全设置加载器
	settingsLoader := security.NewSettingsLoader(db)
	log.Println("Security settings loaded from database")

	// 初始化仓储层
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	// 初始化服务层
	authService := service.NewAuthService(userRepo, roleRepo, cfg.JWT.Secret, settingsLoader)

	// 初始化处理层
	authHandler := handler.NewAuthHandler(authService)

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()

	// 禁用 JSON 转义，解决中文乱码问题
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())
	r.Use(middleware.RequestID())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// 注册路由
	router.RegisterRoutes(r, authHandler, db)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Auth service is running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
