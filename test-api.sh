#!/bin/bash

echo "======================================"
echo "My Cloud 系统测试"
echo "======================================"
echo ""

echo "1. 测试网关健康检查..."
curl -s http://localhost:8080/health | jq .
echo ""

echo "2. 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
echo "$LOGIN_RESPONSE" | jq .
echo ""

# 提取 token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    echo "✅ 登录成功！Token: ${TOKEN:0:20}..."
    echo ""
    
    echo "3. 测试获取用户信息..."
    curl -s -X GET http://localhost:8080/api/v1/auth/userinfo \
      -H "Authorization: Bearer $TOKEN" | jq .
    echo ""
    
    echo "4. 测试获取应用列表..."
    curl -s -X GET "http://localhost:8080/api/v1/applications?page=1&pageSize=10" \
      -H "Authorization: Bearer $TOKEN" | jq .
    echo ""
    
    echo "5. 测试创建应用..."
    curl -s -X POST http://localhost:8080/api/v1/applications \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "测试应用",
        "code": "test-app-001",
        "projectId": 1,
        "type": "web",
        "language": "go",
        "framework": "gin",
        "description": "这是一个测试应用"
      }' | jq .
    echo ""
    
else
    echo "❌ 登录失败！"
fi

echo "======================================"
echo "测试完成"
echo "======================================"
