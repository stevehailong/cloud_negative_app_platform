package main

import (
	"fmt"
	"log"
	"my-cloud/internal/common/config"
	"my-cloud/internal/project/router"
	"my-cloud/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库连接 - 连接到 org_db
	dsn := "root:root123456@tcp(mysql:3306)/org_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.InitDB(dsn, database.DefaultConnectionPoolConfig())
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 初始化Gin
	r := gin.Default()

	// 注册路由
	router.RegisterRoutes(r, db)

	// 启动服务
	port := cfg.Server.Port
	if port == 0 {
		port = 8082
	}
	log.Printf("Project Service 启动在端口 %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("启动服务失败:", err)
	}
}
