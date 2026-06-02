#!/bin/bash
# Harbor配置和部署脚本

set -e

echo "============================================"
echo "  Harbor 集成配置脚本"
echo "============================================"
echo ""

# 配置变量
HARBOR_DIR="/Users/hanhailong01/Downloads/my_cloud/harbor/harbor"
HARBOR_DATA_DIR="/Users/hanhailong01/Downloads/my_cloud/harbor-data"
HARBOR_PORT="8093"
HARBOR_HOSTNAME="localhost"
HARBOR_PASSWORD="Harbor12345"

echo "步骤1: 配置Harbor"
cd $HARBOR_DIR

# 备份原配置
if [ -f harbor.yml ]; then
    cp harbor.yml harbor.yml.bak
fi

# 创建配置文件
cat > harbor.yml <<EOF
# Harbor配置文件
hostname: $HARBOR_HOSTNAME

# HTTP配置
http:
  port: $HARBOR_PORT

# HTTPS配置（开发环境可以禁用）
# https:
#   port: 443
#   certificate: /your/certificate/path
#   private_key: /your/private/key/path

# Admin密码
harbor_admin_password: $HARBOR_PASSWORD

# 数据目录
data_volume: $HARBOR_DATA_DIR

# 数据库配置
database:
  password: root123456
  max_idle_conns: 100
  max_open_conns: 900
  conn_max_lifetime: 5m
  conn_max_idle_time: 0

# Redis配置（外部Redis）
# external_redis:
#   host: localhost:6379
#   password:
#   registry_db_index: 1
#   jobservice_db_index: 2
#   trivy_db_index: 5

# 日志配置
log:
  level: info
  local:
    rotate_count: 50
    rotate_size: 200M
    location: /var/log/harbor

# 存储后端
storage_service:
  ca_bundle:
  filesystem:
    maxthreads: 100

# Trivy漏洞扫描器
trivy:
  ignore_unfixed: false
  skip_update: false
  offline_scan: false
  insecure: false

# 作业服务
jobservice:
  max_job_workers: 10
  logger_sweeper_duration: 1

# 通知Webhook
notification:
  webhook_job_max_retry: 3
  webhook_job_http_client_timeout: 3

# Chart仓库
chart:
  absolute_url: disabled

# Docker Content Trust (Notary)
# 在开发环境可以禁用
# notary:
#   url: http://notary-server:4443

# Clair漏洞扫描（已废弃，使用Trivy）
# clair:
#   updaters_interval: 12

# 代理设置
proxy:
  http_proxy:
  https_proxy:
  no_proxy: 127.0.0.1,localhost

# 上传大小限制（默认）
upload_purging:
  enabled: true
  age: 168h
  interval: 24h
  dryrun: false

# 缓存配置
cache:
  enabled: false
  expire_hours: 24
EOF

echo "✓ Harbor配置文件已创建"
echo ""

echo "步骤2: 创建数据目录"
mkdir -p $HARBOR_DATA_DIR
echo "✓ 数据目录已创建: $HARBOR_DATA_DIR"
echo ""

echo "步骤3: 准备Harbor"
echo "运行 prepare 脚本..."
sudo ./prepare
echo "✓ Harbor准备完成"
echo ""

echo "步骤4: 部署Harbor"
echo "运行 install.sh 脚本..."
echo "这可能需要几分钟时间..."
sudo ./install.sh --with-trivy
echo "✓ Harbor部署完成"
echo ""

echo "============================================"
echo "  Harbor 部署成功！"
echo "============================================"
echo ""
echo "访问地址: http://$HARBOR_HOSTNAME:$HARBOR_PORT"
echo "管理员账号:"
echo "  用户名: admin"
echo "  密码: $HARBOR_PASSWORD"
echo ""
echo "下一步操作:"
echo "1. 访问Harbor Web UI"
echo "2. 创建项目 'mycloud'"
echo "3. 配置Kubernetes imagePullSecrets"
echo "4. 更新CI/CD Pipeline"
echo ""
echo "查看Harbor状态:"
echo "  cd $HARBOR_DIR && docker-compose ps"
echo ""
echo "查看Harbor日志:"
echo "  cd $HARBOR_DIR && docker-compose logs -f"
echo ""
EOF
chmod +x /Users/hanhailong01/Downloads/my_cloud/harbor/setup-harbor.sh
echo "Harbor配置脚本已创建: /Users/hanhailong01/Downloads/my_cloud/harbor/setup-harbor.sh"
