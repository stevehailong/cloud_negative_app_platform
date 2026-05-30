#!/bin/bash

# Monitor Service 测试脚本

BASE_URL="http://localhost:8090/api/v1"

# 生成测试Token (绕过认证,仅用于测试)
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNzQ4MzY0MDAwfQ.test"

echo "========================================="
echo "Monitor Service API 测试"
echo "========================================="
echo ""

# 1. 测试健康检查
echo "1. 测试健康检查"
curl -s http://localhost:8090/health | jq .
echo ""

# 3. 创建指标
echo "2. 创建新指标"
METRIC_RESPONSE=$(curl -s -X POST ${BASE_URL}/metrics \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api_requests_total",
    "type": "counter",
    "description": "API请求总数",
    "unit": "requests",
    "labels": "{\"service\": \"test-service\"}",
    "enabled": 1
  }')
echo "$METRIC_RESPONSE" | jq .
METRIC_ID=$(echo $METRIC_RESPONSE | jq -r '.data.id')
echo ""

# 4. 获取指标列表
echo "3. 获取指标列表"
curl -s "${BASE_URL}/metrics?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 5. 获取指标详情
echo "4. 获取指标详情 (ID: $METRIC_ID)"
curl -s "${BASE_URL}/metrics/${METRIC_ID}" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 6. 创建告警规则
echo "5. 创建告警规则"
RULE_RESPONSE=$(curl -s -X POST ${BASE_URL}/alert-rules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "高请求量告警",
    "metric_name": "api_requests_total",
    "condition": ">",
    "threshold": 1000,
    "duration": 300,
    "severity": "warning",
    "enabled": 1,
    "notify_users": "admin,ops"
  }')
echo "$RULE_RESPONSE" | jq .
RULE_ID=$(echo $RULE_RESPONSE | jq -r '.data.id')
echo ""

# 7. 获取告警规则列表
echo "6. 获取告警规则列表"
curl -s "${BASE_URL}/alert-rules?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 8. 获取告警规则详情
echo "7. 获取告警规则详情 (ID: $RULE_ID)"
curl -s "${BASE_URL}/alert-rules/${RULE_ID}" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 9. 获取告警列表
echo "8. 获取告警列表"
curl -s "${BASE_URL}/alerts?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 10. 获取告警统计
echo "9. 获取告警统计"
curl -s "${BASE_URL}/alerts/statistics" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 11. 更新指标
echo "10. 更新指标"
curl -s -X PUT "${BASE_URL}/metrics/${METRIC_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "API请求总数 (更新)",
    "enabled": 1
  }' | jq .
echo ""

# 12. 按类型筛选指标
echo "11. 按类型筛选指标 (type=counter)"
curl -s "${BASE_URL}/metrics?type=counter" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 13. 按严重级别筛选告警规则
echo "12. 按严重级别筛选告警规则 (severity=critical)"
curl -s "${BASE_URL}/alert-rules?severity=critical" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "========================================="
echo "Monitor Service API 测试完成"
echo "========================================="
