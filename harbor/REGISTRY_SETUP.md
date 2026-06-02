# 本地Docker Registry 部署说明

## ✅ 部署成功

本地Docker Registry已成功部署在 `localhost:8093`

## 访问信息

- **Registry地址**: `localhost:8093`
- **认证**: 无需认证（开发环境）
- **存储**: Docker Volume (`harbor_registry_data`)

## 使用方式

### 1. 推送镜像到本地registry

```bash
# 拉取基础镜像
docker pull nginx:alpine

# 打标签
docker tag nginx:alpine localhost:8093/myapp:v1.0.0

# 推送到本地registry
docker push localhost:8093/myapp:v1.0.0
```

### 2. 从本地registry拉取镜像

```bash
docker pull localhost:8093/myapp:v1.0.0
```

### 3. 查看registry中的镜像

```bash
# 列出所有仓库
curl http://localhost:8093/v2/_catalog

# 列出某个仓库的所有标签
curl http://localhost:8093/v2/myapp/tags/list
```

## 管理命令

### 启动Registry

```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose -f docker-compose-simple.yml up -d
```

### 停止Registry

```bash
docker-compose -f docker-compose-simple.yml down
```

### 查看Registry日志

```bash
docker logs harbor-registry
```

### 查看Registry状态

```bash
docker-compose -f docker-compose-simple.yml ps
```

## 与my-cloud项目集成

### 修改docker-compose.yml使用本地registry

将项目中的镜像引用从远程仓库改为本地registry：

```yaml
services:
  application-service:
    image: localhost:8093/my-cloud-application:latest
    # ... 其他配置
```

### Kubernetes集成

在Kubernetes中使用本地registry：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: myapp
    image: localhost:8093/myapp:v1.0.0
```

**注意**: 由于registry运行在localhost:8093，Kubernetes Pod需要能访问宿主机的8093端口。

## 生产环境建议

当前配置为开发环境，生产环境建议：

1. **启用认证**: 添加htpasswd认证
2. **使用HTTPS**: 配置TLS证书
3. **持久化存储**: 使用外部存储（S3、NFS等）
4. **添加UI**: 部署Harbor完整版或registry-ui
5. **配置备份**: 定期备份镜像数据

## 故障排查

### 无法推送镜像

```bash
# 检查registry是否运行
docker ps | grep harbor-registry

# 检查日志
docker logs harbor-registry

# 重启registry
docker-compose -f docker-compose-simple.yml restart
```

### 无法从Kubernetes访问

确保Kubernetes节点能访问宿主机的8093端口：

```bash
# 在Kubernetes节点上测试
curl http://HOST_IP:8093/v2/_catalog
```

## 已验证功能

✅ 镜像推送 - `docker push localhost:8093/nginx:test`  
✅ 镜像拉取 - `docker pull localhost:8093/nginx:test`  
✅ API访问 - `curl http://localhost:8093/v2/_catalog`  
✅ 持久化存储 - 使用Docker Volume

## 下一步

1. 更新my-cloud项目的docker-compose.yml使用本地registry
2. 配置CI/CD pipeline推送镜像到本地registry
3. 在Kubernetes中配置使用本地registry的镜像
