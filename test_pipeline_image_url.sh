#!/bin/bash
# Pipeline Image URL 自动化测试脚本
# 用于验证流水线执行记录中的镜像地址是否正确返回

set -e

echo "================================================"
echo "Pipeline Image URL 集成测试"
echo "================================================"
echo ""

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_result="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "[$TOTAL_TESTS] 测试: $test_name ... "
    
    if eval "$test_command"; then
        if [ "$expected_result" = "success" ]; then
            echo -e "${GREEN}✓ 通过${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            echo -e "${RED}✗ 失败${NC} (期望失败但成功了)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    else
        if [ "$expected_result" = "fail" ]; then
            echo -e "${GREEN}✓ 通过${NC} (正确失败)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            echo -e "${RED}✗ 失败${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    fi
}

echo "=== 1. 数据库层测试 ==="
echo ""

# 测试1: artifacts 表是否存在
run_test "artifacts 表存在" \
    "docker-compose exec -T mysql mysql -uroot -proot123456 devops_db -e 'DESCRIBE artifacts;' 2>&1 | grep -q 'repo_url'" \
    "success"

# 测试2: 是否有 image 类型的 artifact 记录
run_test "存在 image 类型的 artifact" \
    "docker-compose exec -T mysql mysql -uroot -proot123456 devops_db -e \"SELECT COUNT(*) FROM artifacts WHERE artifact_type='image';\" 2>&1 | grep -v Warning | tail -1 | awk '{exit (\$1 > 0) ? 0 : 1}'" \
    "success"

# 测试3: pipeline_runs 和 artifacts 关联正确
run_test "pipeline_runs 和 artifacts 关联正确" \
    "docker-compose exec -T mysql mysql -uroot -proot123456 devops_db -e \"SELECT COUNT(*) FROM pipeline_runs pr INNER JOIN artifacts a ON pr.id = a.pipeline_run_id WHERE a.artifact_type='image';\" 2>&1 | grep -v Warning | tail -1 | awk '{exit (\$1 > 0) ? 0 : 1}'" \
    "success"

echo ""
echo "=== 2. 服务层测试 ==="
echo ""

# 测试4: pipeline-service 服务是否运行
run_test "pipeline-service 容器运行中" \
    "docker-compose ps pipeline-service | grep -q 'Up'" \
    "success"

# 测试5: pipeline-service 健康检查
run_test "pipeline-service 健康检查" \
    "docker-compose exec -T pipeline-service wget -q -O- http://localhost:8084/health 2>&1 | grep -q 'ok'" \
    "success"

echo ""
echo "=== 3. API 层测试 ==="
echo ""

# 获取一个测试用的 pipeline run ID
SAMPLE_RUN_ID=$(docker-compose exec -T mysql mysql -uroot -proot123456 devops_db -e "SELECT pr.id FROM pipeline_runs pr INNER JOIN artifacts a ON pr.id = a.pipeline_run_id WHERE a.artifact_type='image' LIMIT 1;" 2>&1 | grep -v Warning | tail -1)

if [ -n "$SAMPLE_RUN_ID" ] && [ "$SAMPLE_RUN_ID" != "id" ]; then
    echo "使用测试 run ID: $SAMPLE_RUN_ID"
    
    # 测试6: ListAllPipelineRuns API 返回 imageUrl 字段
    run_test "ListAllPipelineRuns API 返回 imageUrl" \
        "docker-compose exec -T gateway curl -s 'http://pipeline-service:8084/api/v1/pipeline-runs?page=1&pageSize=5' | python3 -c 'import sys, json; data=json.load(sys.stdin); runs=data.get(\"data\",{}).get(\"list\",[]); exit(0 if any(\"imageUrl\" in run and run[\"imageUrl\"] for run in runs) else 1)' 2>&1" \
        "success"
    
    # 测试7: imageUrl 字段格式正确（包含镜像仓库地址）
    run_test "imageUrl 格式正确" \
        "docker-compose exec -T gateway curl -s 'http://pipeline-service:8084/api/v1/pipeline-runs?page=1&pageSize=5' | python3 -c 'import sys, json; data=json.load(sys.stdin); runs=data.get(\"data\",{}).get(\"list\",[]); valid=[r for r in runs if \"imageUrl\" in r and r[\"imageUrl\" ] and (\"mycloud\" in r[\"imageUrl\"] or \"localhost\" in r[\"imageUrl\"])]; exit(0 if valid else 1)' 2>&1" \
        "success"
else
    echo -e "${YELLOW}⚠ 跳过 API 测试: 数据库中没有测试数据${NC}"
fi

echo ""
echo "=== 4. 前端集成测试 ==="
echo ""

# 测试8: 前端容器运行
run_test "frontend 容器运行中" \
    "docker-compose ps frontend | grep -q 'Up'" \
    "success"

# 测试9: 前端资源可访问
run_test "前端页面可访问" \
    "curl -s -o /dev/null -w '%{http_code}' http://localhost:3000 | grep -q '200'" \
    "success"

echo ""
echo "================================================"
echo "测试结果汇总"
echo "================================================"
echo "总测试数: $TOTAL_TESTS"
echo -e "通过: ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败: ${RED}$FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！${NC}"
    echo ""
    echo "=== 样本数据 ==="
    docker-compose exec -T mysql mysql -uroot -proot123456 devops_db -e "
    SELECT 
        pr.run_no, 
        pr.status, 
        a.repo_url as image_url 
    FROM pipeline_runs pr 
    INNER JOIN artifacts a ON pr.id = a.pipeline_run_id 
    WHERE a.artifact_type = 'image' 
    ORDER BY pr.id DESC 
    LIMIT 3;
    " 2>&1 | grep -v Warning
    exit 0
else
    echo -e "${RED}✗ 有测试失败，请检查日志${NC}"
    exit 1
fi
