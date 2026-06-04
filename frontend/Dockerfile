FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm install --legacy-peer-deps

COPY . .
RUN npm run build

# 生产阶段
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

# trace-log-reader（预编译的 Go 二进制）
COPY trace-log-reader /usr/local/bin/trace-log-reader

# 启动脚本
COPY start.sh /start.sh
RUN chmod +x /start.sh /usr/local/bin/trace-log-reader

EXPOSE 80

CMD ["/start.sh"]
