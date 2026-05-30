package main

import (
	"fmt"
	"log"
	"my-cloud/internal/common/config"
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/common/response"
	"my-cloud/internal/pipeline/handler"
	"my-cloud/internal/pipeline/model"
	"my-cloud/internal/pipeline/repository"
	"my-cloud/internal/pipeline/router"
	"my-cloud/internal/pipeline/service"
	"my-cloud/pkg/database"
	"my-cloud/pkg/gitlab"
	"my-cloud/pkg/jenkins"
	"my-cloud/pkg/jwt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化JWT
	jwt.InitJWT(cfg.JWT.Secret)

	// 初始化数据库
	dsn := "root:root123456@tcp(mysql:3306)/devops_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&model.Pipeline{},
		&model.PipelineRun{},
		&model.Artifact{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// 连接到iam_db用于权限检查和读取系统设置
	iamDSN := "root:root123456@tcp(mysql:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
	iamDB, err := database.InitDB(iamDSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to iam_db: %v", err)
	}

	// 初始化Jenkins客户端
	jenkinsURL := os.Getenv("JENKINS_URL")
	if jenkinsURL == "" {
		jenkinsURL = "http://jenkins:8080"
	}
	jenkinsUser := os.Getenv("JENKINS_USER")
	if jenkinsUser == "" {
		jenkinsUser = "admin"
	}
	jenkinsToken := os.Getenv("JENKINS_TOKEN")
	if jenkinsToken == "" {
		jenkinsToken = "admin123"
	}

	jenkinsClient := jenkins.NewClient(jenkinsURL, jenkinsUser, jenkinsToken)
	if err := jenkinsClient.Ping(); err != nil {
		log.Printf("WARNING: Jenkins not reachable at %s: %v (will use simulation mode)", jenkinsURL, err)
		jenkinsClient = nil
	} else {
		log.Printf("Jenkins client connected: %s", jenkinsURL)
	}

	// 初始化GitLab客户端（从系统设置读取）
	var gitlabClient *gitlab.Client
	var gitlabURL, gitlabToken string
	iamDB.Table("system_settings").Where("setting_group = ? AND setting_key = ?", "integration", "gitlabUrl").Pluck("setting_value", &gitlabURL)
	iamDB.Table("system_settings").Where("setting_group = ? AND setting_key = ?", "integration", "gitlabToken").Pluck("setting_value", &gitlabToken)

	if gitlabURL != "" && gitlabToken != "" {
		gitlabClient = gitlab.NewClient(gitlabURL, gitlabToken)
		if err := gitlabClient.Ping(); err != nil {
			log.Printf("WARNING: GitLab not reachable at %s: %v", gitlabURL, err)
			gitlabClient = nil
		} else {
			log.Printf("GitLab client connected: %s", gitlabURL)
		}
	} else {
		log.Printf("GitLab not configured (set in system settings)")
	}

	// 初始化仓库
	pipelineRepo := repository.NewPipelineRepository(db)
	pipelineRunRepo := repository.NewPipelineRunRepository(db)
	artifactRepo := repository.NewArtifactRepository(db)

	// 初始化服务
	pipelineService := service.NewPipelineService(pipelineRepo, pipelineRunRepo, artifactRepo, jenkinsClient, gitlabClient)

	// 初始化处理器
	pipelineHandler := handler.NewPipelineHandler(pipelineService)

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
	router.RegisterRoutes(r, pipelineHandler, iamDB)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Pipeline service is running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
