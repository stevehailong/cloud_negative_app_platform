package main

import (
	"fmt"
	"log"
	"os"

	"my-cloud/internal/common/config"
	"my-cloud/internal/common/middleware"
	"my-cloud/internal/monitor/handler"
	"my-cloud/internal/monitor/integration"
	"my-cloud/internal/monitor/model"
	"my-cloud/internal/monitor/repository"
	"my-cloud/internal/monitor/router"
	"my-cloud/internal/monitor/service"
	"my-cloud/pkg/database"
	"my-cloud/pkg/jwt"
	"my-cloud/pkg/k8s"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化JWT（用于 auth 中间件验证 token）
	jwt.InitJWT(cfg.JWT.Secret)

	// 初始化数据库
	db, err := database.InitDB(cfg.Database.DSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库表（分开迁移，避免一个失败阻止其他表创建）
	err = db.AutoMigrate(&model.TraceSpan{}, &model.TraceQuery{})
	if err != nil {
		log.Printf("Warning: Trace table migration error (ignored): %v", err)
	}
	err = db.AutoMigrate(
		&model.Metric{},
		&model.AlertRule{},
		&model.Alert{},
		&model.LogQuery{},
	)
	if err != nil {
		log.Printf("Warning: Database migration error (ignored): %v", err)
	}

	// 初始化K8s客户端
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}
	k8sClient, err := k8s.NewClientFromKubeconfig(kubeconfigPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize K8s client: %v", err)
	} else {
		log.Printf("K8s client initialized successfully with kubeconfig: %s", kubeconfigPath)
	}

	// 连接 iam_db 以读取系统集成配置（Prometheus/Grafana 等）
	iamDSN := os.Getenv("IAM_DB_DSN")
	if iamDSN == "" {
		iamDSN = "root:root123456@tcp(mysql:3306)/iam_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	iamDB, err := database.InitDB(iamDSN, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Printf("Warning: Failed to connect to iam_db (integration settings disabled): %v", err)
	}

	// 初始化集成配置加载器（含 Prometheus 客户端）
	integrationLoader := integration.NewLoader(iamDB)
	if c := integrationLoader.PrometheusClient(); c != nil {
		log.Printf("Prometheus client initialized: %s", c.BaseURL())
	} else {
		log.Printf("Prometheus not configured (set in system settings)")
	}

	// 初始化Repository
	monitorRepo := repository.NewMonitorRepository(db)
	traceRepo := repository.NewTraceRepository(db)

	// 初始化Service
	monitorService := service.NewMonitorService(monitorRepo)
	traceService := service.NewTraceService(traceRepo)

	// 初始化Handler
	monitorHandler := handler.NewMonitorHandler(monitorService)
	podMonitorHandler := handler.NewPodMonitorHandler(k8sClient, integrationLoader)
	traceHandler := handler.NewTraceHandler(traceService)

	// 初始化Gin路由
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())
	r.Use(middleware.Tracing("monitor-service"))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "monitor-service"})
	})

	// Prometheus metrics 端点（注册自定义指标 + Go runtime + process 指标到同一 registry）
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)
	handler.RegisterMetrics(registry)
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	// 记录每个 HTTP 请求到 mycloud_http_requests_total 计数器（用于 QPS / 错误率）
	r.Use(func(c *gin.Context) {
		c.Next()
		handler.RecordRequest(c.Request.Method, c.FullPath(), fmt.Sprintf("%d", c.Writer.Status()))
	})

	// 设置路由
	router.SetupRouter(r, monitorHandler, podMonitorHandler, traceHandler)

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
