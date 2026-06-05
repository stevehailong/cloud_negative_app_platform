#!/bin/bash
# K8s Ingress Port-Forward 守护脚本
# 保持 book.com 的 K8s Ingress 在本地可访问

PORT=8888
NAMESPACE=ingress-nginx
SERVICE=ingress-nginx-controller
LOCKFILE=/tmp/k8s-ingress-pf.lock
LOGFILE=/tmp/k8s-ingress-pf.log

# 防止重复运行
if [ -f "$LOCKFILE" ] && kill -0 $(cat "$LOCKFILE") 2>/dev/null; then
    echo "Port-forward already running (PID $(cat $LOCKFILE))"
    exit 0
fi

echo $$ > "$LOCKFILE"

while true; do
    echo "[$(date)] Starting port-forward ${PORT}:80 -> ${NAMESPACE}/${SERVICE}"
    kubectl port-forward -n "$NAMESPACE" "svc/${SERVICE}" "${PORT}:80" --address 0.0.0.0 >> "$LOGFILE" 2>&1
    echo "[$(date)] Port-forward exited, restarting in 3s..."
    sleep 3
done
