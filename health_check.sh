#!/bin/bash
# 部署后健康检查脚本
# 在每次服务部署/重启后自动运行，确保关键功能正常

echo "🔍 开始健康检查..."
echo ""

# 1. 检查所有服务是否运行
echo "1️⃣  检查服务状态..."
SERVICES=(
    "gateway"
    "pipeline-service"
    "frontend"
    "mysql"
    "jenkins"
)

ALL_UP=true
for service in "${SERVICES[@]}"; do
    if docker-compose ps "$service" | grep -q "Up"; then
        echo "  ✓ $service: 运行中"
    else
        echo "  ✗ $service: 未运行"
        ALL_UP=false
    fi
done

if [ "$ALL_UP" = false ]; then
    echo ""
    echo "❌ 部分服务未运行，健康检查失败"
    exit 1
fi

echo ""
echo "2️⃣  检查关键 API..."

# 2. 检查 pipeline API 返回 imageUrl
echo -n "  检查 pipeline-runs API 是否返回 imageUrl... "
RESPONSE=$(docker-compose exec -T gateway curl -s 'http://pipeline-service:8084/api/v1/pipeline-runs?page=1&pageSize=1' 2>&1)

if echo "$RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); runs=data.get('data',{}).get('list',[]); exit(0 if runs and 'imageUrl' in runs[0] else 1)" 2>/dev/null; then
    echo "✓"
else
    echo "✗"
    echo ""
    echo "⚠️  警告: API 未返回 imageUrl 字段"
    echo "   响应示例: $RESPONSE" | head -c 200
    echo ""
fi

echo ""
echo "3️⃣  检查数据库连接..."

# 3. 检查数据库表
echo -n "  检查 artifacts 表... "
if docker-compose exec -T mysql mysql -uroot -proot123456 devops_db -e "DESCRIBE artifacts;" 2>&1 | grep -q "repo_url"; then
    echo "✓"
else
    echo "✗"
fi

echo ""
echo "✅ 健康检查完成"
echo ""
echo "💡 提示: 如果发现问题，请运行完整测试: ./test_pipeline_image_url.sh"
