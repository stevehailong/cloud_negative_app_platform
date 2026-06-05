#!/bin/sh
# Trace Forwarder - 实时 tail nginx access_trace.log 并上报到 monitor-service
# 将 nginx trace_json 格式转换为 TraceSpan API 格式

LOG_FILE="/var/log/nginx/access_trace.log"
MONITOR_URL="http://monitor-service:8090/internal/v1/traces/spans"

echo "[$(date)] Trace forwarder started, watching: $LOG_FILE" >&2

while [ ! -f "$LOG_FILE" ]; do sleep 1; done

tail -n 0 -f "$LOG_FILE" | while IFS= read -r line; do
    [ -z "$line" ] && continue

    # 用 awk 提取字段并构造 TraceSpan JSON
    # nginx log: {"time":"...","host":"...","method":"...","path":"...","status":...,"duration":...,"request_id":"...","remote_addr":"...","user_agent":"..."}
    # TraceSpan:  {"traceId":"...","spanId":"...","serviceName":"...","operationName":"...","method":"...","durationMs":...,"startTime":"...","statusCode":...}

    span=$(echo "$line" | awk -F',' '
    {
        gsub(/[{}"]/, "")
        for(i=1; i<=NF; i++) {
            split($i, kv, ":")
            gsub(/^[ \t]+/, "", kv[1])
            gsub(/[ \t]+$/, "", kv[1])
            gsub(/^[ \t]+/, "", kv[2])
            gsub(/[ \t]+$/, "", kv[2])
            val = kv[2]
            for(j=3; j<=length(kv); j++) val = val ":" kv[j]
            if(kv[1] == "time")      time = val
            if(kv[1] == "host")      host = val
            if(kv[1] == "method")    method = val
            if(kv[1] == "path")      path = val
            if(kv[1] == "status")    status = val
            if(kv[1] == "duration")  duration = val
            if(kv[1] == "request_id") reqid = val
        }
        if(!time) time = "1970-01-01T00:00:00Z"
        if(!host) host = "nginx"
        if(!path) path = "/"
        if(!method) method = "GET"
        if(!status) status = 0
        if(!duration) duration = 0

        # host → serviceName 映射（中文名便于平台搜索）
        if(host == "book.com")  host = "图书管理"
        if(host == "app-1.local") host = "app-1"

        # duration 从秒转为毫秒
        dur_ms = int(duration * 1000 + 0.5)

        # 生成 trace_id 和 span_id
        cmd = "cat /proc/sys/kernel/random/uuid 2>/dev/null || echo fallback-" systime()
        cmd | getline tid
        close(cmd)
        cmd | getline sid
        close(cmd)
        if(!tid || tid == "") tid = "nginx-" systime()
        if(!sid || sid == "") sid = "nginx-span-" systime()
        gsub(/\n/, "", tid)
        gsub(/\n/, "", sid)

        printf "{\"traceId\":\"%s\",\"spanId\":\"%s\",\"serviceName\":\"%s\",\"operationName\":\"%s\",\"method\":\"%s\",\"durationMs\":%d,\"startTime\":\"%s\",\"statusCode\":%d}\n",
               tid, sid, host, path, method, dur_ms, time, status
    }')

    [ -z "$span" ] && continue

    curl -s -X POST "$MONITOR_URL" \
        -H "Content-Type: application/json" \
        -d "$span" \
        -o /dev/null 2>/dev/null || true
done
