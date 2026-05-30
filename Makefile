.PHONY: all build-all run-all test-all clean docker-build k8s-deploy init-db

# 变量定义
SERVICES := gateway auth-service project-service application-service pipeline-service \
            env-service release-service deploy-service cluster-service resource-service \
            monitor-service config-service secret-service audit-service notification-service \
            cost-service

# 默认目标
all: build-all

# 构建所有服务
build-all:
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		cd backend/cmd/$$service && go build -o ../../../bin/$$service main.go || exit 1; \
		cd ../../..; \
	done
	@echo "All services built successfully!"

# 运行所有服务（开发模式）
run-all:
	@echo "Running all services..."
	@docker-compose up -d

# 停止所有服务
stop-all:
	@echo "Stopping all services..."
	@docker-compose down

# 测试所有服务
test-all:
	@echo "Testing all services..."
	@cd backend && go test ./... -v

# 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf backend/bin/
	@rm -rf frontend/dist/
	@echo "Clean completed!"

# 构建Docker镜像
docker-build:
	@echo "Building Docker images..."
	@docker-compose build
	@echo "Docker images built successfully!"

# 部署到Kubernetes
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deploy/k8s/namespace.yaml
	@kubectl apply -f deploy/k8s/
	@echo "Deployment completed!"

# 使用Helm部署
helm-deploy:
	@echo "Deploying with Helm..."
	@helm install my-cloud ./deploy/helm/my-cloud
	@echo "Helm deployment completed!"

# 初始化数据库
init-db:
	@echo "Initializing databases..."
	@echo "Waiting for MySQL to be ready..."
	@sleep 10
	@docker exec -i my-cloud-mysql mysql -uroot -proot123456 --default-character-set=utf8mb4 < sql/00_init.sql
	@docker exec -i my-cloud-mysql mysql -uroot -proot123456 --default-character-set=utf8mb4 < sql/01_iam_db.sql
	@docker exec -i my-cloud-mysql mysql -uroot -proot123456 --default-character-set=utf8mb4 < sql/02_org_db.sql
	@docker exec -i my-cloud-mysql mysql -uroot -proot123456 --default-character-set=utf8mb4 < sql/03_app_db.sql
	@echo "Database initialization completed!"
	@mysql -uroot -proot < sql/13_notification_db.sql
	@mysql -uroot -proot < sql/14_cost_db.sql
	@echo "Database initialization completed!"

# 前端构建
frontend-build:
	@echo "Building frontend..."
	@cd frontend && npm install && npm run build
	@echo "Frontend build completed!"

# 前端开发
frontend-dev:
	@echo "Starting frontend dev server..."
	@cd frontend && npm run dev

# 代码格式化
fmt:
	@echo "Formatting Go code..."
	@cd backend && go fmt ./...
	@echo "Formatting frontend code..."
	@cd frontend && npm run lint:fix

# 生成API文档
docs:
	@echo "Generating API documentation..."
	@cd backend && swag init -g cmd/gateway/main.go -o api/swagger
	@echo "API documentation generated!"

# 查看日志
logs:
	@docker-compose logs -f

# 健康检查
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health || echo "Gateway not responding"
	@curl -s http://localhost:8081/health || echo "Auth service not responding"
