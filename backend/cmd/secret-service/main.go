package main

import (
	"fmt"
	"log"

	"my-cloud/internal/secret/handler"
	"my-cloud/internal/secret/repository"
	"my-cloud/internal/secret/router"
	"my-cloud/pkg/database"
	"my-cloud/pkg/metrics"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库连接 - 连接到 secret_db
	dsn := "root:root123456@tcp(mysql:3306)/secret_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 初始化仓储层
	secretRepo := repository.NewSecretRepository(db)

	// 初始化处理器
	secretHandler := handler.NewSecretHandler(secretRepo)

	// 初始化 Gin 路由
	r := gin.Default()

	// Health check and metrics
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/metrics", metrics.Handler())

	// 添加中间件确保正确的Content-Type
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// 注册路由
	router.SetupRouter(r, secretHandler)

	// 启动服务
	port := 8098
	log.Printf("Secret Service 启动在端口 %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
