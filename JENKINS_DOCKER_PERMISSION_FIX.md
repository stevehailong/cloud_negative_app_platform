# Jenkins Docker权限问题解决方案

## 🔍 问题描述

### 错误信息
```
ERROR: permission denied while trying to connect to the Docker daemon socket 
at unix:///var/run/docker.sock
```

### 问题原因
Jenkins容器在执行`docker build`命令时无法访问宿主机的Docker daemon，因为：
1. Jenkins容器默认以`jenkins`用户运行
2. `/var/run/docker.sock`文件需要root权限
3. Jenkins镜像默认不包含Docker CLI

## ✅ 解决方案

### 方案1：使用Docker-in-Docker（已实施）

#### 步骤1：创建自定义Jenkins镜像
创建 `jenkins/Dockerfile`：
```dockerfile
FROM jenkins/jenkins:lts-jdk17

USER root

# 安装Docker CLI
RUN apt-get update && \
    apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \
    apt-get update && \
    apt-get install -y docker-ce-cli && \
    rm -rf /var/lib/apt/lists/*

# 给Jenkins用户添加到docker组
RUN groupadd -f docker && usermod -aG docker jenkins

USER jenkins
```

#### 步骤2：修改docker-compose.yml
```yaml
jenkins:
  build:
    context: ./jenkins
    dockerfile: Dockerfile
  container_name: my-cloud-jenkins
  user: root  # 使用root用户运行
  environment:
    - DOCKER_HOST=unix:///var/run/docker.sock
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock  # 挂载Docker socket
```

#### 步骤3：重新构建并启动
```bash
# 构建新的Jenkins镜像
docker-compose build jenkins

# 启动Jenkins
docker-compose up -d jenkins

# 验证Docker CLI
docker exec my-cloud-jenkins docker --version
```

### 方案2：使用Docker组权限（备选）

如果不想以root运行，可以：

```yaml
jenkins:
  image: jenkins/jenkins:lts-jdk17
  user: "1000:999"  # jenkins用户ID:docker组ID
  group_add:
    - "999"  # docker组ID
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock
```

**注意**：需要先查询宿主机的docker组ID：
```bash
getent group docker | cut -d: -f3
```

## 🧪 验证步骤

### 1. 验证Docker CLI安装
```bash
docker exec my-cloud-jenkins docker --version
```
预期输出：`Docker version xx.xx.x`

### 2. 验证Docker Socket访问
```bash
docker exec my-cloud-jenkins docker ps
```
预期输出：显示正在运行的容器列表

### 3. 验证镜像构建
```bash
docker exec my-cloud-jenkins docker build -t test:latest -<<EOF
FROM alpine:latest
CMD echo "Hello from Jenkins"
EOF
```
预期输出：成功构建镜像

### 4. 重新触发流水线
1. 访问流水线页面
2. 手动触发构建
3. 查看构建日志
4. 验证`docker build`命令成功执行

## 📋 完整的CI/CD流程

### Jenkins Pipeline构建脚本
```bash
#!/bin/bash
set -e

# 配置
REGISTRY="host.docker.internal:5001"
IMAGE_NAME="${REGISTRY}/mycloud/book-service-ci"
IMAGE_TAG="1.0.${BUILD_NUMBER}-${GIT_COMMIT:0:7}"
IMAGE_FULL="${IMAGE_NAME}:${IMAGE_TAG}"

echo "========================================="
echo "Pipeline: ${JOB_NAME}"
echo "Branch: ${GIT_BRANCH}"
echo "Commit: ${GIT_COMMIT}"
echo "Image: ${IMAGE_FULL}"
echo "========================================="

# 步骤1：构建Docker镜像
echo "Step 1: Building Docker Image..."
docker build -t ${IMAGE_FULL} .

# 步骤2：推送到镜像仓库
echo "Step 2: Pushing to Registry..."
docker push ${IMAGE_FULL}

# 步骤3：清理本地镜像
echo "Step 3: Cleaning up..."
docker rmi ${IMAGE_FULL} || true

echo "Build completed successfully!"
echo "Image: ${IMAGE_FULL}"
```

## 🔧 故障排查

### 问题1：构建后仍然权限错误
**检查**：
```bash
# 检查docker.sock权限
docker exec my-cloud-jenkins ls -l /var/run/docker.sock

# 检查当前用户
docker exec my-cloud-jenkins whoami
```

**解决**：
- 确认使用了`user: root`
- 检查docker.sock是否正确挂载

### 问题2：Docker CLI未安装
**检查**：
```bash
docker exec my-cloud-jenkins which docker
```

**解决**：
- 确认使用了自定义Dockerfile
- 重新构建镜像：`docker-compose build --no-cache jenkins`

### 问题3：无法连接Registry
**错误**：`denied: requested access to the resource is denied`

**解决**：
```bash
# 检查Registry是否运行
docker ps | grep registry

# 测试Registry连接
docker exec my-cloud-jenkins curl http://host.docker.internal:5001/v2/
```

### 问题4：镜像仓库不可信
**错误**：`http: server gave HTTP response to HTTPS client`

**解决**：
在Jenkins容器中配置Docker daemon：
```bash
# 创建daemon.json
docker exec my-cloud-jenkins bash -c 'mkdir -p /etc/docker && echo "{\"insecure-registries\": [\"host.docker.internal:5001\"]}" > /etc/docker/daemon.json'

# 注意：由于是共享宿主机Docker，需要在宿主机配置
# Mac Docker Desktop: Settings → Docker Engine → 添加insecure-registries
```

## 📚 相关资源

### Docker-in-Docker最佳实践
- [Docker官方文档](https://docs.docker.com/engine/security/userns-remap/)
- [Jenkins Docker插件](https://plugins.jenkins.io/docker-plugin/)

### 安全建议
1. **生产环境**：不建议使用`user: root`
2. **权限最小化**：使用docker组替代root
3. **镜像安全**：定期扫描Jenkins镜像漏洞
4. **网络隔离**：限制Jenkins访问内部网络

## 🎯 验证清单

完成以下检查确认问题已解决：

- [ ] Jenkins镜像已重新构建（包含Docker CLI）
- [ ] docker-compose.yml已更新（user: root + 环境变量）
- [ ] Jenkins容器已重启
- [ ] `docker --version`命令成功执行
- [ ] `docker ps`命令返回容器列表
- [ ] 手动触发流水线
- [ ] docker build命令成功执行
- [ ] 镜像成功推送到Registry
- [ ] 执行记录显示成功状态
- [ ] 镜像地址正确显示在界面

## ✨ 预期结果

修复后，Jenkins构建日志应该显示：

```
Step 1: Building Docker Image...
Sending build context to Docker daemon  2.048kB
Step 1/3 : FROM alpine:latest
 ---> xxxxx
Step 2/3 : WORKDIR /app
 ---> Running in xxxxx
 ---> xxxxx
Step 3/3 : CMD ["sh"]
 ---> Running in xxxxx
 ---> xxxxx
Successfully built xxxxx
Successfully tagged host.docker.internal:5001/mycloud/book-service-ci:1.0.9576-a1b2c3d

Step 2: Pushing to Registry...
The push refers to repository [host.docker.internal:5001/mycloud/book-service-ci]
xxxxx: Pushed
1.0.9576-a1b2c3d: digest: sha256:xxxxx size: 527

Build completed successfully!
```

---

**问题状态**: 🔄 修复中（Jenkins镜像构建中）
**预计时间**: 2-3分钟（取决于网络速度）
**下一步**: 构建完成后重启Jenkins并重新触发流水线
