# 构建阶段
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 设置代理加速
ENV GOPROXY=https://goproxy.cn,direct

# 复制go.mod和go.sum
COPY go.mod go.sum ./
RUN go mod download

# Build cache-bust timestamp: 20260604T1855Z
ARG CACHE_BUST=1

# 复制源代码 (cache-bust: 2026-06-04-system-health)
COPY . .

# 构建指定服务
ARG SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/${SERVICE_NAME}

# 运行阶段
FROM alpine:latest

# Stage cache-bust: 20260604T1855Z
RUN apk --no-cache add ca-certificates tzdata helm

WORKDIR /root/

# 从构建阶段复制二进制文件 (cache-bust: 2026-06-04T18:55)
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 设置时区
ENV TZ=Asia/Shanghai

ENV BUILD_ID=20260604T190000

EXPOSE 8080

CMD ["./main"]
