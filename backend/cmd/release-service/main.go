package main

import (
	"fmt"
	"log"
	"my-cloud/internal/common/config"
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/common/response"
	"my-cloud/internal/release/handler"
	"my-cloud/internal/release/model"
	"my-cloud/internal/release/repository"
	"my-cloud/internal/release/router"
	"my-cloud/internal/release/service"
	"my-cloud/pkg/database"
	"my-cloud/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化JWT
	jwt.InitJWT(cfg.JWT.Secret)

	// 初始化数据库
	dsn := "root:root123456@tcp(mysql:3306)/release_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&model.Release{},
		&model.ReleaseApproval{},
	)
	if err != nil {
		// 忽略 GORM 的索引迁移错误
		log.Printf("Warning: Database migration error (ignored): %v", err)
	} else {
		log.Println("Database migration completed")
	}

	// 连接到iam_db用于权限检查
	iamDSN := "root:root123456@tcp(mysql:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
	iamDB, err := database.InitDB(iamDSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to iam_db: %v", err)
	}

	// 初始化仓库
	releaseRepo := repository.NewReleaseRepository(db)
	releaseApprovalRepo := repository.NewReleaseApprovalRepository(db)

	// 初始化服务
	releaseService := service.NewReleaseService(releaseRepo, releaseApprovalRepo)

	// 初始化处理器
	releaseHandler := handler.NewReleaseHandler(releaseService)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()

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
	router.RegisterRoutes(r, releaseHandler, iamDB)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Release service is running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
