#!/bin/bash
# API 响应诊断脚本

echo "=== Pipeline Runs API 响应诊断 ==="
echo ""

# 获取 run_no 为 book-service-ci-1780320879 的记录
RUN_NO="book-service-ci-1780320879"

echo "1️⃣  数据库中的数据："
docker-compose exec -T mysql mysql -uroot -proot123456 devops_db -e "
SELECT 
    pr.id,
    pr.run_no,
    pr.status,
    a.artifact_type,
    a.repo_url as imageUrl
FROM pipeline_runs pr 
LEFT JOIN artifacts a ON pr.id = a.pipeline_run_id 
WHERE pr.run_no = '$RUN_NO';
" 2>&1 | grep -v Warning

echo ""
echo "2️⃣  测试 API 响应（通过容器内部）："

# 使用 wget 从 pipeline-service 容器内部测试
docker-compose exec -T pipeline-service wget -q -O- 'http://localhost:8084/api/v1/pipeline-runs?page=1&pageSize=10&pipeline_id=22' 2>&1 > /tmp/api_response.json

echo "API 响应已保存到 /tmp/api_response.json"
echo ""

echo "3️⃣  检查响应中的 imageUrl 字段："
if command -v jq &> /dev/null; then
    # 使用 jq 格式化并检查
    cat /tmp/api_response.json | jq '.data.list[0] | {id, runNo, status, imageUrl}' 2>&1 | head -20
else
    # 没有 jq，直接搜索 imageUrl
    echo "搜索 imageUrl 字段："
    grep -o '"imageUrl":"[^"]*"' /tmp/api_response.json | head -5 || echo "未找到 imageUrl 字段"
fi

echo ""
echo "4️⃣  完整的第一条记录："
if command -v jq &> /dev/null; then
    cat /tmp/api_response.json | jq '.data.list[0]' 2>&1 | head -30
else
    head -50 /tmp/api_response.json
fi

echo ""
echo "=== 诊断完成 ==="
