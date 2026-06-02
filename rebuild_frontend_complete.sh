#!/bin/bash
# 前端完全重建脚本 - 解决缓存问题

set -e

echo "🔧 开始前端完全重建..."
echo ""

# 1. 停止并删除容器
echo "1️⃣  停止并删除前端容器..."
docker-compose stop frontend
docker-compose rm -f frontend

# 2. 删除镜像
echo "2️⃣  删除旧的前端镜像..."
docker rmi my_cloud-frontend || echo "   (镜像不存在，跳过)"

# 3. 清理 Docker 构建缓存
echo "3️⃣  清理 Docker 构建缓存..."
docker builder prune -f

# 4. 重新构建（不使用缓存）
echo "4️⃣  重新构建前端镜像（不使用缓存）..."
docker-compose build --no-cache --pull frontend

# 5. 启动容器
echo "5️⃣  启动前端容器..."
docker-compose up -d frontend

# 6. 等待启动
echo "6️⃣  等待前端启动..."
sleep 5

# 7. 验证
echo "7️⃣  验证部署..."
echo ""

# 检查容器状态
if docker-compose ps frontend | grep -q "Up"; then
    echo "✅ 前端容器运行中"
else
    echo "❌ 前端容器未运行"
    exit 1
fi

# 检查文件时间
BUILD_TIME=$(docker exec my-cloud-frontend sh -c 'ls -l /usr/share/nginx/html/index.html' | awk '{print $6,$7,$8}')
echo "✅ 前端文件构建时间: $BUILD_TIME"

# 检查镜像创建时间
IMAGE_TIME=$(docker images my_cloud-frontend --format "{{.CreatedAt}}")
echo "✅ 镜像创建时间: $IMAGE_TIME"

# 检查容器创建时间
CONTAINER_TIME=$(docker inspect my-cloud-frontend | grep Created | head -1 | cut -d'"' -f4)
echo "✅ 容器创建时间: $CONTAINER_TIME"

echo ""
echo "🎉 前端重建完成！"
echo ""
echo "📝 下一步："
echo "   1. 清除浏览器缓存（Ctrl+Shift+Delete）"
echo "   2. 关闭并重新打开浏览器"
echo "   3. 访问 http://localhost:3000"
echo "   4. 强制刷新页面（Ctrl+Shift+R）"
echo ""
echo "💡 如果还是看不到更新，请使用无痕模式测试"
