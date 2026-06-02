# My-Cloud 本地Registry使用指南

## 📦 Registry信息

- **地址**: `localhost:5001` (宿主机访问) 或 `registry:5000` (容器内访问)
- **容器名**: `my-cloud-registry`
- **镜像**: `registry:2`
- **状态**: ✅ 运行中
- **认证**: 无需认证（内网环境）
- **存储**: Docker Volume `registry_data`

## 🎯 已有镜像

当前Registry中已有以下镜像：

```bash
mycloud/book-service-ci:1.0.882-a1b2c3d
mycloud/book-service-ci:1.0.1626-e062150
mycloud/test-app:v1.0.0
test-host:v1
test-alpine:v1  # 测试镜像
```

## 📝 使用方法

### 1. 推送镜像到本地Registry

```bash
# 构建镜像
docker build -t my-app:v1.0.0 .

# 打标签（指向本地registry）
docker tag my-app:v1.0.0 localhost:5001/mycloud/my-app:v1.0.0

# 推送到本地registry
docker push localhost:5001/mycloud/my-app:v1.0.0
```

### 2. 从本地Registry拉取镜像

```bash
# 从宿主机拉取
docker pull localhost:5001/mycloud/my-app:v1.0.0

# 在docker-compose中使用
services:
  my-service:
    image: localhost:5001/mycloud/my-app:v1.0.0
```

### 3. 在容器内访问Registry

如果你的服务在`my-cloud-network`网络中，可以直接使用`registry:5000`：

```yaml
services:
  my-service:
    image: registry:5000/mycloud/my-app:v1.0.0
    networks:
      - my-cloud-network
```

## 🔍 查询Registry内容

### 查看所有镜像

```bash
curl http://localhost:5001/v2/_catalog
```

输出示例：
```json
{
  "repositories": [
    "mycloud/book-service-ci",
    "mycloud/test-app",
    "test-host"
  ]
}
```

### 查看镜像的所有标签

```bash
curl http://localhost:5001/v2/mycloud/book-service-ci/tags/list
```

输出示例：
```json
{
  "name": "mycloud/book-service-ci",
  "tags": ["1.0.882-a1b2c3d", "1.0.1626-e062150"]
}
```

### 查看镜像清单(Manifest)

```bash
curl http://localhost:5001/v2/mycloud/book-service-ci/manifests/1.0.882-a1b2c3d
```

## 🛠️ 管理命令

### 启动Registry

```bash
cd /Users/hanhailong01/Downloads/my_cloud
docker-compose up -d registry
```

### 停止Registry

```bash
docker-compose stop registry
```

### 重启Registry

```bash
docker-compose restart registry
```

### 查看Registry日志

```bash
docker logs my-cloud-registry

# 实时查看日志
docker logs -f my-cloud-registry
```

### 查看Registry状态

```bash
docker ps | grep my-cloud-registry
```

## 🔧 在CI/CD中使用

### Jenkins Pipeline示例

```groovy
pipeline {
    agent any
    
    environment {
        REGISTRY = 'localhost:5001'
        IMAGE_NAME = 'mycloud/my-app'
        IMAGE_TAG = "${env.BUILD_NUMBER}"
    }
    
    stages {
        stage('Build') {
            steps {
                sh 'docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .'
            }
        }
        
        stage('Push to Registry') {
            steps {
                sh '''
                    docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
                    docker push ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
                '''
            }
        }
    }
}
```

### GitLab CI示例

```yaml
build_and_push:
  stage: build
  script:
    - docker build -t $CI_PROJECT_NAME:$CI_COMMIT_SHORT_SHA .
    - docker tag $CI_PROJECT_NAME:$CI_COMMIT_SHORT_SHA localhost:5001/mycloud/$CI_PROJECT_NAME:$CI_COMMIT_SHORT_SHA
    - docker push localhost:5001/mycloud/$CI_PROJECT_NAME:$CI_COMMIT_SHORT_SHA
```

## 📊 在Kubernetes中使用

### 直接使用本地Registry镜像

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: my-app
        image: localhost:5001/mycloud/my-app:v1.0.0
        # 或使用registry服务名（如果K8s在同一网络）
        # image: registry:5000/mycloud/my-app:v1.0.0
```

### 如果Kubernetes集群无法访问localhost:5001

需要配置Docker Daemon或使用NodePort/LoadBalancer暴露Registry服务。

## 🗑️ 清理镜像

### 删除本地标签

```bash
docker rmi localhost:5001/mycloud/my-app:v1.0.0
```

### 从Registry删除镜像（需要启用delete）

当前Registry未启用删除功能。如需启用，修改`docker-compose.yml`：

```yaml
registry:
  environment:
    - REGISTRY_STORAGE_DELETE_ENABLED=true
```

然后重启Registry并使用API删除：

```bash
# 获取镜像digest
DIGEST=$(curl -I -s -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
  http://localhost:5001/v2/mycloud/my-app/manifests/v1.0.0 \
  | grep Docker-Content-Digest | awk '{print $2}' | tr -d '\r')

# 删除镜像
curl -X DELETE http://localhost:5001/v2/mycloud/my-app/manifests/${DIGEST}

# 运行垃圾回收
docker exec my-cloud-registry bin/registry garbage-collect /etc/docker/registry/config.yml
```

## 🔒 安全建议

当前Registry配置为**无认证模式**，适合内网开发环境。

### 如需启用认证

1. 创建htpasswd文件：
```bash
mkdir -p auth
docker run --rm --entrypoint htpasswd httpd:2 -Bbn admin password123 > auth/htpasswd
```

2. 修改docker-compose.yml：
```yaml
registry:
  environment:
    - REGISTRY_AUTH=htpasswd
    - REGISTRY_AUTH_HTPASSWD_REALM=Registry
    - REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd
  volumes:
    - ./auth:/auth
```

3. 使用时需要登录：
```bash
docker login localhost:5001 -u admin -p password123
```

## 📈 监控和维护

### 检查存储空间

```bash
docker volume inspect registry_data
docker system df -v | grep registry_data
```

### 查看Registry版本

```bash
docker exec my-cloud-registry registry --version
```

### 备份Registry数据

```bash
# 备份volume
docker run --rm -v registry_data:/data -v $(pwd):/backup alpine tar czf /backup/registry-backup.tar.gz /data

# 恢复
docker run --rm -v registry_data:/data -v $(pwd):/backup alpine tar xzf /backup/registry-backup.tar.gz -C /
```

## ✅ 测试验证

以下功能已验证通过：

- ✅ 推送镜像：`docker push localhost:5001/test-alpine:v1`
- ✅ 拉取镜像：`docker pull localhost:5001/test-alpine:v1`
- ✅ 查询catalog：`curl http://localhost:5001/v2/_catalog`
- ✅ 查询tags：`curl http://localhost:5001/v2/test-alpine/tags/list`
- ✅ 持久化存储：重启容器后数据保留

## 🆘 故障排查

### 推送失败：connection refused

检查Registry是否运行：
```bash
docker ps | grep my-cloud-registry
docker logs my-cloud-registry
```

### 推送失败：unauthorized

如果启用了认证，先登录：
```bash
docker login localhost:5001
```

### 存储空间不足

清理未使用的镜像和volume：
```bash
docker system prune -a
```

## 📚 相关资源

- [Docker Registry官方文档](https://docs.docker.com/registry/)
- [Registry HTTP API V2](https://docs.docker.com/registry/spec/api/)
- My-Cloud项目文档：`/Users/hanhailong01/Downloads/my_cloud/README.md`
