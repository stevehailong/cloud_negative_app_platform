#!/bin/bash

# Audit Service 测试脚本

echo "========================================="
echo "Audit Service 功能测试"
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
test_api "服务健康检查" "GET" "http://localhost:8093/health" "" 200

# 2. 测试审计日志查询(未认证)
echo "=== 2. 审计日志查询 ==="
echo "注意: 以下API需要认证，预期返回401未授权"

test_api "获取审计日志列表(未认证)" "GET" "http://localhost:8093/api/v1/audit-logs" "" 401

test_api "获取审计日志详情(未认证)" "GET" "http://localhost:8093/api/v1/audit-logs/1" "" 401

test_api "根据资源获取审计日志(未认证)" "GET" "http://localhost:8093/api/v1/audit-logs/resource/application/123" "" 401

test_api "根据用户获取审计日志(未认证)" "GET" "http://localhost:8093/api/v1/audit-logs/user/1" "" 401

# 3. 测试统计功能
echo "=== 3. 统计分析 ==="
test_api "获取统计信息(未认证)" "GET" "http://localhost:8093/api/v1/audit-logs/statistics" "" 401

# 4. 测试导出功能
echo "=== 4. 日志导出 ==="
test_api "导出审计日志(未认证)" "GET" "http://localhost:8093/api/v1/audit-logs/export" "" 401

# 5. 测试清理功能
echo "=== 5. 日志清理 ==="
clean_data='{"retentionDays": 90}'
test_api "清理过期日志(未认证)" "POST" "http://localhost:8093/api/v1/audit-logs/clean" "$clean_data" 401

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
    echo "3. 所有7个端点路由配置正确"
    echo ""
    echo "下一步:"
    echo "- 需要先调用 auth-service 获取 JWT token"
    echo "- 然后在请求头中添加: Authorization: Bearer <token>"
    echo "- 参考文档: docs/audit-service.md"
    echo ""
    echo "测试审计记录功能:"
    echo "1. 执行任意API操作 (如创建应用、发布等)"
    echo "2. 使用有效token查询审计日志"
    echo "3. 验证操作已被记录"
    exit 0
else
    echo -e "${RED}部分测试失败，请检查服务日志${NC}"
    exit 1
fi
