#!/bin/bash
# Harbor ARM64 在线安装脚本

set -e

echo "=========================================="
echo "Harbor ARM64 在线安装"
echo "=========================================="

# 1. 清理旧安装
echo "清理旧容器..."
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose down -v 2>/dev/null || true

# 2. 下载 Harbor 在线安装器（支持 ARM64）
echo "下载 Harbor 在线安装器..."
cd /Users/hanhailong01/Downloads/my_cloud/harbor
if [ ! -f "harbor-online-installer-v2.11.0.tgz" ]; then
    curl -L https://github.com/goharbor/harbor/releases/download/v2.11.0/harbor-online-installer-v2.11.0.tgz -o harbor-online-installer-v2.11.0.tgz
fi

# 3. 解压
echo "解压安装包..."
tar xzvf harbor-online-installer-v2.11.0.tgz -C /tmp/

# 4. 复制配置文件
echo "配置 Harbor..."
cp /Users/hanhailong01/Downloads/my_cloud/harbor/harbor/harbor.yml /tmp/harbor/harbor.yml

# 5. 执行安装
echo "开始安装 Harbor..."
cd /tmp/harbor
sudo ./install.sh

echo ""
echo "=========================================="
echo "✅ Harbor 安装完成！"
echo "=========================================="
echo ""
echo "访问地址: http://localhost:8093"
echo "用户名: admin"
echo "密码: Harbor12345"
echo ""
