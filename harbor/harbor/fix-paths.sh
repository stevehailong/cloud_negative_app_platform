#!/bin/bash
# 修复 Harbor 在 macOS 上的路径挂载问题

set -e

cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor

echo "修复 docker-compose.yml 中的路径挂载..."

# 备份原文件
cp docker-compose.yml docker-compose.yml.backup

# 将所有 /tmp/harbor-data 的 bind mount 替换为相对路径
sed -i.bak '
  s|/tmp/harbor-data/|./harbor-data/|g
  s|source: /tmp/harbor-data/|source: ./harbor-data/|g
' docker-compose.yml

# 创建本地数据目录
mkdir -p ./harbor-data/{registry,database,redis,job_logs}
mkdir -p ./harbor-data/secret/{registry,keys,core}
chmod -R 777 ./harbor-data

echo "✅ 修复完成"
echo ""
echo "现在运行: docker-compose up -d"
