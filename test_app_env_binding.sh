#!/bin/bash

# 应用环境绑定功能测试脚本

echo "================================"
echo "应用环境绑定功能测试"
echo "================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8085/api/v1"

echo "1. 测试查询环境列表（应包含clusterName）"
echo "--------------------------------------"
ENVS=$(docker exec my-cloud-env-service wget -qO- 'http://localhost:8085/api/v1/environments?page=1&pageSize=10')
echo "$ENVS" | jq '.data.list[0] | {envName, envType, namespace, clusterName}'

if echo "$ENVS" | jq -e '.data.list[0].clusterName != null' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 环境列表包含集群名称${NC}"
else
    echo -e "${RED}✗ 环境列表缺少集群名称（但不影响基本功能）${NC}"
fi
echo ""

echo "2. 测试创建应用环境绑定"
echo "--------------------------------------"
# 假设应用ID=1，环境ID=1
BINDING_DATA='{
  "appId": 1,
  "envId": 1,
  "replicas": 2,
  "cpuRequest": "200m",
  "cpuLimit": "1",
  "memoryRequest": "256Mi",
  "memoryLimit": "1Gi",
  "configJson": "{}"
}'

RESULT=$(docker exec my-cloud-env-service wget -qO- --post-data="$BINDING_DATA" --header='Content-Type: application/json' 'http://localhost:8085/api/v1/app-env-bindings' 2>/dev/null)

if echo "$RESULT" | jq -e '.code == 0' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 创建绑定成功${NC}"
    echo "$RESULT" | jq '.'
elif echo "$RESULT" | jq -e '.code == 40001' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 绑定已存在（符合预期）${NC}"
    echo "$RESULT" | jq '.message'
else
    echo -e "${RED}✗ 创建绑定失败${NC}"
    echo "$RESULT" | jq '.'
fi
echo ""

echo "3. 测试查询应用的环境绑定列表"
echo "--------------------------------------"
BINDINGS=$(docker exec my-cloud-env-service wget -qO- 'http://localhost:8085/api/v1/app-env-bindings?applicationId=1&page=1&pageSize=10' 2>/dev/null)
echo "$BINDINGS" | jq '.data.list[] | {envName, envType, namespace, clusterName, replicas, cpuLimit, memoryLimit, configStatus}'

if echo "$BINDINGS" | jq -e '.data.list | length > 0' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 查询绑定列表成功${NC}"
else
    echo -e "${RED}✗ 绑定列表为空${NC}"
fi
echo ""

echo "================================"
echo "测试完成"
echo "================================"
echo ""
echo "前端测试步骤："
echo "1. 访问 http://localhost （清空浏览器缓存）"
echo "2. 登录系统"
echo "3. 进入应用管理 → 选择一个应用 → 查看详情"
echo "4. 在应用详情页找到"环境绑定"卡片"
echo "5. 点击【绑定环境】按钮"
echo "6. 选择环境、配置资源，点击【确定】"
echo "7. 验证环境是否成功显示在绑定列表中"
echo ""
