#!/bin/bash
# Harbor 部署执行脚本 - 请在终端中执行

echo "=========================================="
echo "  Harbor 部署 - 自动执行"
echo "=========================================="
echo ""
echo "当前位置: $(pwd)"
echo ""

# 设置变量
HARBOR_DIR="/Users/hanhailong01/Downloads/my_cloud/harbor/harbor"
echo "Harbor目录: $HARBOR_DIR"
echo ""

# 检查是否在正确的目录
if [ ! -f "$HARBOR_DIR/install.sh" ]; then
    echo "❌ 错误: 找不到install.sh"
    echo "请确保在正确的目录中运行"
    exit 1
fi

cd "$HARBOR_DIR"
echo "✓ 已切换到Harbor目录"
echo ""

# 检查必要文件
echo "检查必要文件..."
if [ ! -f "harbor.yml" ]; then
    echo "❌ 错误: 找不到harbor.yml"
    exit 1
fi
echo "✓ harbor.yml 存在"

if [ ! -f "harbor.v2.11.0.tar.gz" ]; then
    echo "❌ 错误: 找不到harbor.v2.11.0.tar.gz"
    exit 1
fi
echo "✓ harbor.v2.11.0.tar.gz 存在"
echo ""

# 检查Docker是否运行
echo "检查Docker状态..."
if ! docker ps &> /dev/null; then
    echo "❌ 错误: Docker未运行"
    echo "请启动Docker Desktop"
    exit 1
fi
echo "✓ Docker正在运行"
echo ""

# 检查端口是否被占用
echo "检查端口8093..."
if lsof -i :8093 &> /dev/null; then
    echo "⚠️  警告: 端口8093已被占用"
    echo "占用进程:"
    lsof -i :8093
    echo ""
    read -p "是否继续? (y/n): " continue
    if [ "$continue" != "y" ]; then
        exit 1
    fi
else
    echo "✓ 端口8093可用"
fi
echo ""

echo "=========================================="
echo "  开始安装Harbor"
echo "=========================================="
echo ""
echo "⚠️  注意: 此过程需要sudo权限"
echo "⚠️  预计耗时: 5-10分钟"
echo ""
read -p "按回车键开始安装..."
echo ""

# 执行安装
echo "步骤1/3: 运行prepare脚本..."
if sudo ./prepare; then
    echo "✓ prepare完成"
else
    echo "❌ prepare失败"
    exit 1
fi
echo ""

echo "步骤2/3: 加载Harbor镜像..."
echo "⏰ 此步骤需要5-10分钟..."
if sudo docker load -i harbor.v2.11.0.tar.gz; then
    echo "✓ 镜像加载完成"
else
    echo "❌ 镜像加载失败"
    exit 1
fi
echo ""

echo "步骤3/3: 启动Harbor服务..."
if sudo docker-compose up -d; then
    echo "✓ Harbor服务启动完成"
else
    echo "❌ Harbor服务启动失败"
    exit 1
fi
echo ""

# 等待服务就绪
echo "等待Harbor服务就绪..."
sleep 10

# 检查服务状态
echo "检查服务状态..."
docker-compose ps
echo ""

echo "=========================================="
echo "  🎉 Harbor安装完成！"
echo "=========================================="
echo ""
echo "访问信息:"
echo "  URL: http://localhost:8093"
echo "  用户名: admin"
echo "  密码: Harbor12345"
echo ""
echo "下一步:"
echo "1. 打开浏览器访问 http://localhost:8093"
echo "2. 使用 admin / Harbor12345 登录"
echo "3. 创建项目 'mycloud'"
echo "4. 执行基础测试:"
echo "   docker login localhost:8093 -u admin -p Harbor12345"
echo "   docker pull nginx:alpine"
echo "   docker tag nginx:alpine localhost:8093/mycloud/nginx:test"
echo "   docker push localhost:8093/mycloud/nginx:test"
echo ""
echo "常用命令:"
echo "  查看状态: docker-compose ps"
echo "  查看日志: docker-compose logs -f"
echo "  重启: docker-compose restart"
echo "  停止: docker-compose stop"
echo ""
