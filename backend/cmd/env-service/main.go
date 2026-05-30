package main

import (
	"fmt"
	"log"
	"my-cloud/internal/common/config"
	"my-cloud/internal/environment/handler"
	"my-cloud/internal/environment/repository"
	"my-cloud/internal/environment/router"
	"my-cloud/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接 - 连接到 env_db
	dsn := "root:root123456@tcp(mysql:3306)/env_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 初始化仓储层
	envRepo := repository.NewEnvironmentRepository(db)
	templateRepo := repository.NewEnvTemplateRepository(db)
	bindingRepo := repository.NewAppEnvBindingRepository(db)
	configMapRepo := repository.NewConfigMapRepository(db)
	secretRepo := repository.NewSecretRepository(db)

	// 初始化处理器
	envHandler := handler.NewEnvironmentHandler(envRepo, templateRepo, bindingRepo)
	configHandler := handler.NewConfigHandler(configMapRepo, secretRepo)

	// 初始化 Gin 路由
	r := gin.Default()

	// 注册路由
	router.RegisterRoutes(r, envHandler, configHandler)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8085
	}
	log.Printf("Environment Service 启动在端口 %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
