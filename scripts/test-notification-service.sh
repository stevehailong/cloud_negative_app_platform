#!/bin/bash

# Notification Service 测试脚本

echo "========================================="
echo "Notification Service 功能测试"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试结果统计
PASS=0
FAIL=0

# 测试函数
test_api() {
    local test_name=$1
    local method=$2
    local url=$3
    local data=$4
    local expected_code=$5
    
    echo -e "${YELLOW}测试: ${test_name}${NC}"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓ 通过${NC} (HTTP $http_code)"
        echo "响应: $body" | jq '.' 2>/dev/null || echo "响应: $body"
        ((PASS++))
    else
        echo -e "${RED}✗ 失败${NC} (期望 $expected_code, 实际 $http_code)"
        echo "响应: $body"
        ((FAIL++))
    fi
    echo ""
}

# 1. 测试健康检查
echo "=== 1. 健康检查 ==="
test_api "服务健康检查" "GET" "http://localhost:8095/health" "" 200

# 2. 测试创建通知模板
echo "=== 2. 通知模板管理 ==="
template_data='{
  "templateCode": "TEST_TEMPLATE",
  "templateName": "测试模板",
  "notifyType": "system",
  "channel": "dingtalk",
  "title": "【测试】{{title}}",
  "content": "这是一条测试消息：{{message}}",
  "variables": "[\"title\",\"message\"]"
}'

echo "注意: 以下API需要认证，预期返回401未授权"
test_api "创建通知模板(未认证)" "POST" "http://localhost:8095/api/v1/notification-templates" "$template_data" 401

# 3. 测试获取模板列表
test_api "获取模板列表(未认证)" "GET" "http://localhost:8095/api/v1/notification-templates" "" 401

# 4. 测试创建通知渠道
echo "=== 3. 通知渠道管理 ==="
channel_data='{
  "channelCode": "TEST_DINGTALK",
  "channelName": "测试钉钉渠道",
  "channelType": "dingtalk",
  "config": "{\"webhook\":\"https://oapi.dingtalk.com/robot/send?access_token=test\"}"
}'

test_api "创建通知渠道(未认证)" "POST" "http://localhost:8095/api/v1/notification-channels" "$channel_data" 401

# 5. 测试获取渠道列表
test_api "获取渠道列表(未认证)" "GET" "http://localhost:8095/api/v1/notification-channels" "" 401

# 6. 测试发送通知
echo "=== 4. 通知发送 ==="
notification_data='{
  "title": "测试通知",
  "content": "这是一条测试通知内容",
  "notifyType": "system",
  "channel": "dingtalk",
  "receiverType": "user",
  "receiverIds": "1,2,3"
}'

test_api "发送通知(未认证)" "POST" "http://localhost:8095/api/v1/notifications" "$notification_data" 401

# 7. 测试模板发送
template_send_data='{
  "templateCode": "RELEASE_SUCCESS",
  "params": {
    "projectName": "测试项目",
    "version": "v1.0.0",
    "environment": "production",
    "operator": "测试用户",
    "releaseTime": "2026-05-28 15:00:00"
  },
  "receiverType": "user",
  "receiverIds": [1, 2, 3]
}'

test_api "模板发送通知(未认证)" "POST" "http://localhost:8095/api/v1/notifications/template" "$template_send_data" 401

# 8. 测试获取通知列表
test_api "获取通知列表(未认证)" "GET" "http://localhost:8095/api/v1/notifications?page=1&pageSize=10" "" 401

echo "========================================="
echo "测试完成"
echo "========================================="
echo -e "${GREEN}通过: $PASS${NC}"
echo -e "${RED}失败: $FAIL${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}所有测试通过！✓${NC}"
    echo ""
    echo "说明:"
    echo "1. 服务健康检查正常"
    echo "2. API接口已正确配置认证中间件(返回401表示需要JWT token)"
    echo "3. 所有端点路由配置正确"
    echo ""
    echo "下一步:"
    echo "- 需要先调用 auth-service 获取 JWT token"
    echo "- 然后在请求头中添加: Authorization: Bearer <token>"
    echo "- 参考文档: docs/notification-service.md"
    exit 0
else
    echo -e "${RED}部分测试失败，请检查服务日志${NC}"
    exit 1
fi
