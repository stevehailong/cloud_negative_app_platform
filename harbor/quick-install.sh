#!/bin/bash
# Harbor 简化安装 - 跳过离线镜像，在线拉取

set -e

echo "=========================================="
echo "Harbor 简化安装（自动拉取ARM64镜像）"
echo "=========================================="

cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor

# 1. 清理
echo "[1/4] 清理旧容器..."
docker-compose down -v 2>/dev/null || true
rm -rf /tmp/harbor-data
mkdir -p /tmp/harbor-data
chmod 777 /tmp/harbor-data

# 2. 运行 prepare 生成配置
echo "[2/4] 生成配置文件..."
sudo ./prepare

# 3. 修改 docker-compose.yml，让Docker自动拉取镜像
echo "[3/4] 修改配置以支持ARM64..."
# Docker会自动拉取适配当前平台的镜像

# 4. 启动服务
echo "[4/4] 启动Harbor服务..."
docker-compose up -d

echo ""
echo "=========================================="
echo "正在拉取镜像并启动容器..."
echo "首次运行需要下载镜像，请等待3-5分钟"
echo "=========================================="
echo ""
echo "查看启动进度:"
echo "  docker-compose logs -f"
echo ""
echo "完成后访问:"
echo "  http://localhost:8093"
echo "  用户名: admin"
echo "  密码: Harbor12345"
echo ""
