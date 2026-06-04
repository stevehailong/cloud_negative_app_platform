#!/bin/sh

# 确保 trace 日志文件存在且可写
mkdir -p /var/log/nginx
touch /var/log/nginx/access_trace.log
chmod 666 /var/log/nginx/access_trace.log

# 启动 trace-log-reader（后台运行）
/usr/local/bin/trace-log-reader &

# 启动 nginx（前台运行）
exec nginx -g "daemon off;"