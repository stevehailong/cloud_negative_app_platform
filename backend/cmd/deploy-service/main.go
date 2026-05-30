package main

import (
	"fmt"
	"log"
	"my-cloud/internal/common/config"
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/common/response"
	"my-cloud/internal/deploy/handler"
	"my-cloud/internal/deploy/model"
	"my-cloud/internal/deploy/repository"
	"my-cloud/internal/deploy/router"
	"my-cloud/internal/deploy/service"
	"my-cloud/pkg/database"
	"my-cloud/pkg/jwt"
	"my-cloud/pkg/k8s"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化JWT
	jwt.InitJWT(cfg.JWT.Secret)

	// 初始化数据库
	dsn := "root:root123456@tcp(mysql:3306)/deploy_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&model.Deployment{},
		&model.AppDeployment{},
		&model.DeploymentHistory{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// 连接到iam_db用于权限检查
	iamDSN := "root:root123456@tcp(mysql:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
	iamDB, err := database.InitDB(iamDSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to iam_db: %v", err)
	}

	// 初始化Kubernetes客户端
	var k8sClient *k8s.Client
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = "/root/.kube/config"
	}

	if _, err := os.Stat(kubeconfigPath); err == nil {
		k8sClient, err = k8s.NewClientFromKubeconfig(kubeconfigPath)
		if err != nil {
			log.Printf("WARNING: Failed to create K8s client from kubeconfig: %v", err)
		} else {
			log.Printf("K8s client initialized from kubeconfig: %s", kubeconfigPath)
		}
	} else {
		// 尝试in-cluster config
		k8sClient, err = k8s.NewClientInCluster()
		if err != nil {
			log.Printf("WARNING: No K8s client available (no kubeconfig, no in-cluster): %v", err)
		} else {
			log.Println("K8s client initialized with in-cluster config")
		}
	}

	// 初始化仓库
	deploymentRepo := repository.NewDeploymentRepository(db)
	appDeploymentRepo := repository.NewAppDeploymentRepository(db)
	deploymentHistoryRepo := repository.NewDeploymentHistoryRepository(db)

	// 初始化服务
	deployService := service.NewDeployService(deploymentRepo, k8sClient)
	appDeploymentService := service.NewAppDeploymentService(appDeploymentRepo, deploymentHistoryRepo, k8sClient)

	// 初始化处理器
	deployHandler := handler.NewDeployHandler(deployService)
	appDeploymentHandler := handler.NewAppDeploymentHandler(appDeploymentService)

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
	router.RegisterRoutes(r, deployHandler, appDeploymentHandler, iamDB)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Deploy service is running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
