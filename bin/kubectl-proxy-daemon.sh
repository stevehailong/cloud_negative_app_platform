#!/bin/bash
# kubectl proxy 守护脚本
# Prometheus 需要通过 kubectl proxy 访问 K8s cAdvisor/kubelet 指标

PORT=8001
LOCKFILE=/tmp/kubectl-proxy.lock
LOGFILE=/tmp/kubectl-proxy.log

if [ -f "$LOCKFILE" ] && kill -0 $(cat "$LOCKFILE") 2>/dev/null; then
    echo "kubectl proxy already running (PID $(cat $LOCKFILE))"
    exit 0
fi

echo $$ > "$LOCKFILE"

while true; do
    echo "[$(date)] Starting kubectl proxy on port $PORT"
    kubectl proxy --port=$PORT --address=0.0.0.0 --accept-hosts='.*' --disable-filter=true >> "$LOGFILE" 2>&1
    echo "[$(date)] kubectl proxy exited, restarting in 3s..."
    sleep 3
done
