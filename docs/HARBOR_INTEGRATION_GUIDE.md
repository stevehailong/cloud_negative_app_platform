# Harbor 完整集成方案

## 方案概述

将现有的简单Registry替换为Harbor企业级镜像仓库，提供Web UI、权限管理、漏洞扫描等企业级功能。

## 集成步骤

### 步骤1：添加Harbor到docker-compose.yml

由于Harbor是多个服务组成的复杂系统，我们采用**Harbor Standalone模式**，使用官方的docker-compose配置。

### 步骤2：配置Harbor

#### 2.1 创建Harbor配置文件

创建 `harbor/harbor.yml`:

```yaml
# Harbor配置文件
hostname: localhost

# HTTP配置
http:
  port: 8093

# Admin密码
harbor_admin_password: Harbor12345

# 数据目录
data_volume: /Users/hanhailong01/Downloads/my_cloud/harbor-data

# 数据库配置（使用现有MySQL）
database:
  password: root123456
  max_idle_conns: 100
  max_open_conns: 900

# Redis配置（使用现有Redis）
external_redis:
  host: my-cloud-redis:6379

# 日志
log:
  level: info
  local:
    location: /var/log/harbor
```

### 步骤3：使用docker-compose部署Harbor

```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
cp harbor.yml.tmpl harbor.yml
# 编辑harbor.yml，修改hostname为localhost
sudo ./install.sh
```

### 步骤4：验证Harbor部署

访问 http://localhost:8093
- 用户名: admin
- 密码: Harbor12345

### 步骤5：创建项目和配置

1. 登录Harbor Web UI
2. 创建项目 `mycloud`（设置为私有）
3. 添加机器人账号用于CI/CD
4. 配置漏洞扫描策略

### 步骤6：配置Kubernetes访问Harbor

```bash
# 创建Harbor访问凭证
kubectl create secret docker-registry harbor-secret \
  --docker-server=localhost:8093 \
  --docker-username=admin \
  --docker-password=Harbor12345 \
  -n default
```

### 步骤7：更新Helm Chart模板

修改 `helm-charts/mycloud-app/values.yaml`:

```yaml
image:
  repository: localhost:8093/mycloud/myapp
  pullPolicy: IfNotPresent
  tag: ""

imagePullSecrets:
  - name: harbor-secret
```

### 步骤8：更新CI/CD Pipeline

#### Jenkins Pipeline

```groovy
environment {
    HARBOR_URL = 'localhost:8093'
    HARBOR_PROJECT = 'mycloud'
    HARBOR_CREDENTIALS = credentials('harbor-credentials')
}

stage('Build & Push') {
    sh """
        docker build -t ${HARBOR_URL}/${HARBOR_PROJECT}/${APP_NAME}:${BUILD_NUMBER} .
        echo \${HARBOR_CREDENTIALS_PSW} | docker login ${HARBOR_URL} -u \${HARBOR_CREDENTIALS_USR} --password-stdin
        docker push ${HARBOR_URL}/${HARBOR_PROJECT}/${APP_NAME}:${BUILD_NUMBER}
    """
}
```

### 步骤9：更新Go代码中的镜像配置

修改所有服务中的镜像URL：

**之前**:
```go
imageURL := "localhost:5001/mycloud/user-service:v1.0.0"
```

**之后**:
```go
imageURL := "localhost:8093/mycloud/user-service:v1.0.0"
```

或使用配置文件：
```yaml
# config.yaml
registry:
  url: "localhost:8093"
  project: "mycloud"
  username: "admin"
  password: "Harbor12345"
```

```go
type RegistryConfig struct {
    URL      string `yaml:"url"`
    Project  string `yaml:"project"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

func GetImageURL(appName, tag string) string {
    return fmt.Sprintf("%s/%s/%s:%s", 
        config.Registry.URL,
        config.Registry.Project,
        appName,
        tag,
    )
}
```

### 步骤10：迁移现有镜像

```bash
# 从旧registry拉取镜像
docker pull localhost:5001/mycloud/user-service:v1.0.0

# 重新打标签
docker tag localhost:5001/mycloud/user-service:v1.0.0 \
           localhost:8093/mycloud/user-service:v1.0.0

# 推送到Harbor
docker login localhost:8093 -u admin -p Harbor12345
docker push localhost:8093/mycloud/user-service:v1.0.0
```

### 步骤11：停用旧Registry

在确认Harbor工作正常后，注释掉docker-compose.yml中的registry服务：

```yaml
# registry:
#   image: registry:2
#   container_name: my-cloud-registry
#   ports:
#     - "5001:5000"
#   ...
```

## 集成后的优势

### 功能对比

| 功能 | 旧Registry | Harbor |
|-----|-----------|---------|
| Web UI | ❌ 无 | ✅ 功能强大 |
| 权限管理 | ❌ 无 | ✅ RBAC |
| 漏洞扫描 | ❌ 无 | ✅ Trivy集成 |
| 镜像复制 | ❌ 无 | ✅ 多Harbor复制 |
| Webhook | ❌ 无 | ✅ 支持 |
| 配额管理 | ❌ 无 | ✅ 支持 |

### 实际使用示例

#### 1. 查看镜像漏洞

登录Harbor → 进入mycloud项目 → 点击镜像 → 查看漏洞扫描结果

#### 2. 配置自动扫描

项目配置 → 自动扫描 → 开启推送时自动扫描

#### 3. 配置Webhook

项目配置 → Webhooks → 添加 → 配置通知URL

当有新镜像推送时，自动通知CI/CD系统。

## 故障排查

### 问题1：无法访问Harbor

```bash
# 检查Harbor容器状态
cd harbor/harbor
docker-compose ps

# 查看日志
docker-compose logs -f harbor-core
```

### 问题2：推送镜像失败

```bash
# 检查是否登录
docker login localhost:8093

# 检查项目是否存在
# 访问 http://localhost:8093 → 项目

# 检查磁盘空间
df -h
```

### 问题3：Kubernetes无法拉取镜像

```bash
# 检查Secret是否创建
kubectl get secret harbor-secret -n default

# 检查Pod是否使用了imagePullSecrets
kubectl get pod <pod-name> -o yaml | grep imagePullSecrets -A 5
```

## 性能优化

### 1. 配置镜像缓存

```yaml
# harbor.yml
proxy:
  http_proxy:
  https_proxy:
  no_proxy: localhost,127.0.0.1
```

### 2. 配置存储清理策略

登录Harbor → 系统管理 → 垃圾清理 → 配置定时任务

### 3. 配置日志轮转

```yaml
# harbor.yml
log:
  rotate_count: 50
  rotate_size: 200M
```

## 安全配置

### 1. 配置HTTPS（生产环境必须）

```bash
# 生成自签名证书
openssl req -newkey rsa:4096 -nodes -sha256 \
  -keyout harbor.key -x509 -days 365 \
  -out harbor.crt

# 更新harbor.yml
https:
  port: 443
  certificate: /path/to/harbor.crt
  private_key: /path/to/harbor.key
```

### 2. 配置LDAP集成

```yaml
# harbor.yml
auth_mode: ldap_auth
ldap:
  url: ldaps://ldap.mycompany.com
  search_dn: cn=admin,dc=example,dc=com
  search_password: password
  base_dn: dc=example,dc=com
```

### 3. 配置审计日志

Harbor → 系统管理 → 日志 → 查看所有操作审计

## 监控和告警

### 1. Prometheus监控

Harbor自带Prometheus metrics:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'harbor'
    static_configs:
      - targets: ['localhost:8093']
```

### 2. 配置告警

- 镜像存储空间告警
- 漏洞扫描告警
- 推送失败告警

## 总结

Harbor集成后，团队可以享受：
- ✅ 企业级镜像管理
- ✅ 自动化漏洞扫描
- ✅ 完善的权限控制
- ✅ 直观的Web界面
- ✅ 完整的审计日志

**预计集成时间**: 2-4小时
**维护成本**: 极低（每月<1小时）
**收益**: 显著提升镜像管理效率和安全性
