#!/bin/bash
# 直接修复 nginx DNS 问题的脚本

cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor

# 停止所有服务
docker-compose down

# 修改 nginx 配置，使用 IP 而不是主机名
CORE_IP=$(docker inspect harbor-core 2>/dev/null | grep '"IPAddress"' | head -1 | awk -F'"' '{print $4}')

if [ -z "$CORE_IP" ]; then
    echo "Harbor core 未运行，先启动依赖服务..."
    docker-compose up -d log postgresql redis registry registryctl portal core jobservice
    sleep 10
    CORE_IP=$(docker inspect harbor-core | grep '"IPAddress"' | head -1 | awk -F'"' '{print $4}')
fi

echo "Harbor core IP: $CORE_IP"

# 备份原始 nginx 配置
cp ./common/config/nginx/nginx.conf ./common/config/nginx/nginx.conf.backup 2>/dev/null ||  true

# 将 core:8080 替换为 IP
sed -i.bak "s/core:8080/${CORE_IP}:8080/g" ./common/config/nginx/nginx.conf

echo "已修复 nginx 配置"
echo "现在启动 proxy..."

docker-compose up -d proxy

echo ""
echo "等待 nginx 启动..."
sleep 5

docker-compose ps

echo ""
echo "如果 nginx 状态为 Up，请访问: http://localhost:8093"
