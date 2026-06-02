#!/bin/bash
# 使用 Docker named volumes 替代 bind mounts 解决 macOS 路径问题

cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor

echo "停止 Harbor..."
docker-compose down -v

echo "修改 docker-compose.yml 使用 named volumes..."

# 备份
cp docker-compose.yml docker-compose.yml.volumes-backup

# 创建使用 volumes 的配置
cat > docker-compose.yml.new <<'EOF'
version: '2.3'
services:
  log:
    image: goharbor/harbor-log:v2.11.0
    container_name: harbor-log
    restart: always
    volumes:
      - harbor_log:/var/log/docker/:z
      - ./common/config/log/logrotate.conf:/etc/logrotate.d/logrotate.conf:z
      - ./common/config/log/rsyslog_docker.conf:/etc/rsyslog.d/rsyslog_docker.conf:z
    ports:
      - 127.0.0.1:1514:10514
    networks:
      - harbor
  registry:
    image: goharbor/registry-photon:v2.11.0
    container_name: registry
    restart: always
    volumes:
      - harbor_registry:/storage:z
      - ./common/config/registry/:/etc/registry/:z
    networks:
      - harbor
    depends_on:
      - log
  registryctl:
    image: goharbor/harbor-registryctl:v2.11.0
    container_name: registryctl
    env_file:
      - ./common/config/registryctl/env
    restart: always
    volumes:
      - harbor_registry:/storage:z
      - ./common/config/registry/:/etc/registry/:z
      - ./common/config/registryctl/config.yml:/etc/registryctl/config.yml:z
    networks:
      - harbor
    depends_on:
      - log
  postgresql:
    image: goharbor/harbor-db:v2.11.0
    container_name: harbor-db
    restart: always
    volumes:
      - harbor_database:/var/lib/postgresql/data:z
    networks:
      harbor:
    env_file:
      - ./common/config/db/env
  core:
    image: goharbor/harbor-core:v2.11.0
    container_name: harbor-core
    env_file:
      - ./common/config/core/env
    restart: always
    volumes:
      - harbor_ca:/etc/core/ca/:z
      - harbor_data:/data/:z
      - ./common/config/core/certificates/:/etc/core/certificates/:z
      - ./common/config/core/app.conf:/etc/core/app.conf:z
      - ./harbor-data/secret/core/private_key.pem:/etc/core/private_key.pem:z
      - ./harbor-data/secret/keys/secretkey:/etc/core/key:z
    networks:
      harbor:
    depends_on:
      - log
      - registry
      - redis
      - postgresql
  portal:
    image: goharbor/harbor-portal:v2.11.0
    container_name: harbor-portal
    restart: always
    volumes:
      - ./common/config/portal/nginx.conf:/etc/nginx/nginx.conf:z
    networks:
      - harbor
    depends_on:
      - log
  jobservice:
    image: goharbor/harbor-jobservice:v2.11.0
    container_name: harbor-jobservice
    env_file:
      - ./common/config/jobservice/env
    restart: always
    volumes:
      - harbor_job_logs:/var/log/jobs:z
      - ./common/config/jobservice/config.yml:/etc/jobservice/config.yml:z
    networks:
      - harbor
    depends_on:
      - core
  redis:
    image: goharbor/redis-photon:v2.11.0
    container_name: redis
    restart: always
    volumes:
      - harbor_redis:/var/lib/redis
    networks:
      - harbor
    depends_on:
      - log
  proxy:
    image: goharbor/nginx-photon:v2.11.0
    container_name: nginx
    restart: always
    volumes:
      - ./common/config/nginx:/etc/nginx:z
    networks:
      - harbor
    ports:
      - 8093:8080
    depends_on:
      registry:
        condition: service_healthy
      core:
        condition: service_healthy
      portal:
        condition: service_healthy
      log:
        condition: service_healthy
networks:
  harbor:
    external: false
volumes:
  harbor_log:
  harbor_registry:
  harbor_database:
  harbor_ca:
  harbor_data:
  harbor_job_logs:
  harbor_redis:
EOF

mv docker-compose.yml.new docker-compose.yml

echo ""
echo "✅ 已切换到 Docker named volumes"
echo "启动 Harbor..."

docker-compose up -d

echo ""
echo "等待服务启动..."
sleep 20

docker-compose ps

echo ""
echo "修复 nginx DNS..."
CORE_IP=$(docker inspect harbor-core | grep '"IPAddress"' | head -1 | awk -F'"' '{print $4}')
sed -i.bak "s/core:8080/${CORE_IP}:8080/g" ./common/config/nginx/nginx.conf
docker exec nginx nginx -s reload

echo ""
echo "完成！访问 http://localhost:8093"
