#!/bin/bash

echo "========================================="
echo "启动Serveo隧道转发"
echo "========================================="
echo ""

# 停止旧的隧道
pkill -f "ssh.*serveo" 2>/dev/null

# 启动新隧道
echo "正在连接serveo.net..."
ssh -o StrictHostKeyChecking=no -R 80:localhost:8080 serveo.net 2>&1 | tee /tmp/serveo_output.log &

TUNNEL_PID=$!
echo "隧道进程PID: $TUNNEL_PID"
echo ""
echo "等待获取公网URL..."
sleep 5

# 从输出中提取URL
WEBHOOK_URL=$(grep -o "https://[a-z0-9\-]*\.serveousercontent\.com" /tmp/serveo_output.log | head -1)

if [ -n "$WEBHOOK_URL" ]; then
    echo ""
    echo "✅ 隧道已建立！"
    echo "公网URL: $WEBHOOK_URL"
    echo "Webhook URL: $WEBHOOK_URL/hooks/gitlab"
    echo ""
    echo "下一步操作："
    echo "1. 更新docker-compose.yml中的WEBHOOK_BASE_URL=$WEBHOOK_URL"
    echo "2. 执行: docker-compose restart pipeline-service"
    echo "3. 在GitLab webhook设置中更新URL为: $WEBHOOK_URL/hooks/gitlab"
else
    echo "❌ 未能获取隧道URL，请查看日志: /tmp/serveo_output.log"
    echo "或手动运行: ssh -R 80:localhost:8080 serveo.net"
fi

echo ""
echo "隧道将保持运行，按Ctrl+C停止"
wait $TUNNEL_PID
