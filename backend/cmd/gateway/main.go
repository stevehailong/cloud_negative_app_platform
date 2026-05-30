package main

import (
	"fmt"
	"log"
	"my-cloud/internal/common/config"
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/common/response"
	"my-cloud/internal/gateway/router"
	"my-cloud/pkg/database"
	"my-cloud/pkg/jwt"
	"my-cloud/pkg/security"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化JWT
	jwt.InitJWT(cfg.JWT.Secret)

	// 初始化数据库连接 - Gateway需要连接iam_db进行权限验证
	iamDSN := "root:root123456@tcp(mysql:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
	iamDB, err := database.InitDB(iamDSN, &database.ConnectionPoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    50,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to connect to iam_db: %v", err)
	}
	log.Println("Gateway connected to iam_db for permission check")

	// 初始化审计数据库连接
	auditDSN := "root:root123456@tcp(mysql:3306)/audit_db?charset=utf8mb4&parseTime=True&loc=Local"
	auditDB, err := database.InitDB(auditDSN, &database.ConnectionPoolConfig{
		MaxIdleConns:    5,
		MaxOpenConns:    20,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to connect to audit_db: %v", err)
	}
	log.Println("Gateway connected to audit_db for audit logging")

	// 初始化安全设置加载器
	settingsLoader := security.NewSettingsLoader(iamDB)
	log.Println("Security settings loaded for rate limiting and IP whitelist")

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()

	// 配置路由 - 自动重定向尾部斜杠
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())
	r.Use(middleware.RequestID())
	r.Use(middleware.IPWhitelist(settingsLoader))
	r.Use(middleware.APIRateLimit(settingsLoader))
	r.Use(middleware.AuditMiddleware(auditDB))

	// 设置JSON渲染器以正确处理中文
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// 注册路由（传入iam数据库连接）
	router.RegisterRoutes(r, iamDB)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Gateway server is running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
