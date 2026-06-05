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
	envRepo "my-cloud/internal/environment/repository"
	"my-cloud/pkg/database"
	"my-cloud/pkg/jwt"
	"my-cloud/pkg/k8s"
	"my-cloud/pkg/metrics"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化JWT
	jwt.InitJWT(cfg.JWT.Secret)

	// 根据环境选择数据库主机
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "mysql" // Docker 环境默认使用 mysql
	}

	// 初始化数据库
	dsn := fmt.Sprintf("root:root123456@tcp(%s:3306)/deploy_db?charset=utf8mb4&parseTime=True&loc=Local", dbHost)
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
	iamDSN := fmt.Sprintf("root:root123456@tcp(%s:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local", dbHost)
	iamDB, err := database.InitDB(iamDSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to iam_db: %v", err)
	}

	// 连接到env_db用于环境信息查询
	envDSN := fmt.Sprintf("root:root123456@tcp(%s:3306)/env_db?charset=utf8mb4&parseTime=True&loc=Local", dbHost)
	envDB, err := database.InitDB(envDSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to env_db: %v", err)
	}

	// 初始化Kubernetes客户端
	var k8sClient *k8s.Client
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		homeDir, _ := os.UserHomeDir()
		kubeconfigPath = homeDir + "/.kube/config"
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
	environmentRepo := envRepo.NewEnvironmentRepository(envDB)
	templateRepo := envRepo.NewEnvTemplateRepository(envDB)
	bindingRepo := envRepo.NewAppEnvBindingRepository(envDB)

	// 初始化服务
	deployService := service.NewDeployService(deploymentRepo, k8sClient)
	appDeploymentService := service.NewAppDeploymentService(appDeploymentRepo, deploymentHistoryRepo, environmentRepo, templateRepo, bindingRepo, k8sClient)

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
	r.Use(middleware.Tracing("deploy-service"))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// Prometheus /metrics endpoint
	r.GET("/metrics", metrics.Handler())

	// 注册路由
	router.RegisterRoutes(r, deployHandler, appDeploymentHandler, iamDB)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Deploy service is running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
