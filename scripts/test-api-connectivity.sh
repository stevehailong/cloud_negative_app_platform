#!/bin/bash

echo "========================================"
echo "API接口连通性测试"
echo "========================================"
echo ""

# 测试Gateway直连
echo "1. 测试Gateway直连 (8080端口)"
GATEWAY_HEALTH=$(curl -s http://localhost:8080/health | jq -r '.data.status')
if [ "$GATEWAY_HEALTH" = "ok" ]; then
    echo "   ✅ Gateway健康检查通过"
else
    echo "   ❌ Gateway健康检查失败"
fi
echo ""

# 测试前端Nginx代理
echo "2. 测试前端Nginx代理 (80端口)"
NGINX_API=$(curl -s -o /dev/null -w "%{http_code}" http://localhost/api/v1/projects)
if [ "$NGINX_API" = "401" ] || [ "$NGINX_API" = "200" ]; then
    echo "   ✅ Nginx代理正常 (HTTP $NGINX_API)"
else
    echo "   ❌ Nginx代理异常 (HTTP $NGINX_API)"
fi
echo ""

# 测试各个微服务
echo "3. 测试各个微服务健康检查"

services=(
    "auth-service:8081"
    "project-service:8082"
    "application-service:8083"
    "pipeline-service:8084"
    "env-service:8085"
    "release-service:8086"
    "deploy-service:8087"
    "cluster-service:8088"
    "monitor-service:8090"
    "audit-service:8093"
    "notification-service:8095"
)

for service in "${services[@]}"; do
    name=$(echo $service | cut -d: -f1)
    port=$(echo $service | cut -d: -f2)
    
    health=$(curl -s http://localhost:${port}/health | jq -r '.data.status' 2>/dev/null)
    if [ "$health" = "ok" ]; then
        echo "   ✅ ${name} (${port})"
    else
        echo "   ❌ ${name} (${port})"
    fi
done
echo ""

# 测试前端页面
echo "4. 测试前端页面"
FRONTEND=$(curl -s -o /dev/null -w "%{http_code}" http://localhost/)
if [ "$FRONTEND" = "200" ]; then
    echo "   ✅ 前端页面正常 (HTTP 200)"
else
    echo "   ❌ 前端页面异常 (HTTP $FRONTEND)"
fi
echo ""

# 测试登录接口
echo "5. 测试登录接口"
LOGIN_RESULT=$(curl -s http://localhost/api/v1/auth/login \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}')

LOGIN_CODE=$(echo $LOGIN_RESULT | jq -r '.code')
if [ "$LOGIN_CODE" = "40101" ] || [ "$LOGIN_CODE" = "0" ]; then
    echo "   ✅ 登录接口响应正常 (code: $LOGIN_CODE)"
else
    echo "   ❌ 登录接口异常 (code: $LOGIN_CODE)"
fi
echo ""

# 检查容器网络
echo "6. 检查容器网络连通性"
FRONTEND_TO_GATEWAY=$(docker exec my-cloud-frontend wget -q -O- --timeout=2 http://gateway:8080/health 2>/dev/null | jq -r '.data.status')
if [ "$FRONTEND_TO_GATEWAY" = "ok" ]; then
    echo "   ✅ 前端容器 -> Gateway 连通"
else
    echo "   ❌ 前端容器 -> Gateway 不通"
fi
echo ""

# 总结
echo "========================================"
echo "测试完成"
echo "========================================"
echo ""
echo "📝 说明:"
echo "  - 如果所有测试都显示 ✅，说明服务正常"
echo "  - 如果看到 401 错误，这是正常的（需要登录）"
echo "  - 如果看到 502 错误，说明服务连接有问题"
echo ""
echo "🌐 访问地址:"
echo "  - 前端: http://localhost"
echo "  - Gateway: http://localhost:8080"
echo ""
