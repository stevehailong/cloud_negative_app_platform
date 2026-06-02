# Harbor 完整集成 - 操作清单

## ✅ 已完成

- [x] 下载Harbor离线安装包 (v2.11.0)
- [x] 解压Harbor安装包
- [x] 创建Harbor配置脚本
- [x] 创建集成文档

## 🔄 待执行（按顺序）

### 阶段1：部署Harbor (30分钟)

#### 1.1 运行Harbor部署脚本
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor
sudo ./setup-harbor.sh
```

**预期结果**: Harbor所有容器启动成功

#### 1.2 验证Harbor服务
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose ps
```

**预期结果**: 所有服务状态为 `Up` 或 `Up (healthy)`

#### 1.3 访问Harbor UI
- 浏览器访问: http://localhost:8093
- 登录账号: admin / Harbor12345

**预期结果**: 能正常登录Harbor管理界面

---

### 阶段2：配置Harbor (15分钟)

#### 2.1 创建项目
1. 登录Harbor
2. 点击"新建项目"
3. 项目名称: `mycloud`
4. 访问级别: 私有
5. 点击"确定"

#### 2.2 创建机器人账号（用于CI/CD）
1. 进入 `mycloud` 项目
2. 点击"机器人账号"标签页
3. 点击"新建机器人账号"
4. 名称: `ci-robot`
5. 权限: 推送拉取镜像
6. 保存生成的token

#### 2.3 配置漏洞扫描
1. 系统管理 → 配置管理
2. 系统设置 → 漏洞扫描
3. 启用"推送时自动扫描"
4. 保存

---

### 阶段3：测试Harbor (15分钟)

#### 3.1 Docker登录测试
```bash
docker login localhost:8093 -u admin -p Harbor12345
```

**预期结果**: `Login Succeeded`

#### 3.2 推送测试镜像
```bash
# 拉取一个测试镜像
docker pull nginx:alpine

# 打标签
docker tag nginx:alpine localhost:8093/mycloud/nginx:test

# 推送
docker push localhost:8093/mycloud/nginx:test
```

**预期结果**: 推送成功，可以在Harbor UI中看到镜像

#### 3.3 拉取测试
```bash
# 删除本地镜像
docker rmi localhost:8093/mycloud/nginx:test

# 从Harbor拉取
docker pull localhost:8093/mycloud/nginx:test
```

**预期结果**: 拉取成功

---

### 阶段4：集成到项目 (1小时)

#### 4.1 配置Kubernetes Secret
```bash
kubectl create secret docker-registry harbor-secret \
  --docker-server=localhost:8093 \
  --docker-username=admin \
  --docker-password=Harbor12345 \
  -n default

# 验证
kubectl get secret harbor-secret
```

#### 4.2 更新Helm Chart模板
修改 `helm-charts/mycloud-app/values.yaml`:

**文件位置**: `/Users/hanhailong01/Downloads/my_cloud/helm-charts/mycloud-app/values.yaml`

**修改内容**:
```yaml
# 修改镜像配置
image:
  repository: localhost:8093/mycloud/{{ .Chart.Name }}  # 改为Harbor地址
  pullPolicy: IfNotPresent
  tag: ""

# 添加imagePullSecrets
imagePullSecrets:
  - name: harbor-secret
```

#### 4.3 创建Go配置工具类
**文件**: `backend/pkg/registry/config.go`

```go
package registry

import (
    "fmt"
    "os"
)

type Config struct {
    URL      string
    Project  string
    Username string
    Password string
}

func NewConfig() *Config {
    return &Config{
        URL:      getEnv("HARBOR_URL", "localhost:8093"),
        Project:  getEnv("HARBOR_PROJECT", "mycloud"),
        Username: getEnv("HARBOR_USERNAME", "admin"),
        Password: getEnv("HARBOR_PASSWORD", "Harbor12345"),
    }
}

func (c *Config) GetImageURL(appName, tag string) string {
    return fmt.Sprintf("%s/%s/%s:%s", c.URL, c.Project, appName, tag)
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

#### 4.4 更新现有代码使用Harbor
在需要构建镜像URL的地方，替换为:

```go
import "my-cloud/pkg/registry"

// 创建registry配置
registryConfig := registry.NewConfig()

// 生成镜像URL
imageURL := registryConfig.GetImageURL("user-service", "v1.0.0")
// 结果: localhost:8093/mycloud/user-service:v1.0.0
```

#### 4.5 更新环境变量
修改 `docker-compose.yml`，为所有Go服务添加Harbor配置:

```yaml
environment:
  - HARBOR_URL=harbor:8093  # 容器内使用容器名
  - HARBOR_PROJECT=mycloud
  - HARBOR_USERNAME=admin
  - HARBOR_PASSWORD=Harbor12345
```

---

### 阶段5：迁移现有镜像 (30分钟)

#### 5.1 列出现有镜像
```bash
# 查看旧registry中的镜像
curl http://localhost:5001/v2/_catalog
```

#### 5.2 迁移镜像脚本
创建 `scripts/migrate-images.sh`:

```bash
#!/bin/bash
# 镜像迁移脚本

OLD_REGISTRY="localhost:5001"
NEW_REGISTRY="localhost:8093"
PROJECT="mycloud"

# 登录Harbor
docker login $NEW_REGISTRY -u admin -p Harbor12345

# 获取所有镜像
IMAGES=$(curl -s http://$OLD_REGISTRY/v2/_catalog | jq -r '.repositories[]')

for IMAGE in $IMAGES; do
    echo "迁移镜像: $IMAGE"
    
    # 获取所有tag
    TAGS=$(curl -s http://$OLD_REGISTRY/v2/$IMAGE/tags/list | jq -r '.tags[]')
    
    for TAG in $TAGS; do
        echo "  迁移tag: $TAG"
        
        # 拉取
        docker pull $OLD_REGISTRY/$IMAGE:$TAG
        
        # 重新打标签
        docker tag $OLD_REGISTRY/$IMAGE:$TAG $NEW_REGISTRY/$PROJECT/$IMAGE:$TAG
        
        # 推送到Harbor
        docker push $NEW_REGISTRY/$PROJECT/$IMAGE:$TAG
        
        echo "  ✓ $IMAGE:$TAG 迁移完成"
    done
done

echo "所有镜像迁移完成！"
```

运行迁移:
```bash
chmod +x scripts/migrate-images.sh
./scripts/migrate-images.sh
```

---

### 阶段6：更新CI/CD (30分钟)

#### 6.1 更新Jenkins Pipeline
修改 `Jenkinsfile`:

```groovy
environment {
    HARBOR_URL = 'localhost:8093'
    HARBOR_PROJECT = 'mycloud'
    HARBOR_CREDENTIALS = credentials('harbor-credentials')
}

stages {
    stage('Build Image') {
        steps {
            sh """
                docker build -t ${HARBOR_URL}/${HARBOR_PROJECT}/${APP_NAME}:${BUILD_NUMBER} .
            """
        }
    }
    
    stage('Push to Harbor') {
        steps {
            sh """
                echo \${HARBOR_CREDENTIALS_PSW} | docker login ${HARBOR_URL} \\
                  -u \${HARBOR_CREDENTIALS_USR} --password-stdin
                docker push ${HARBOR_URL}/${HARBOR_PROJECT}/${APP_NAME}:${BUILD_NUMBER}
            """
        }
    }
    
    stage('Scan Image') {
        steps {
            sh """
                # Harbor会自动扫描，这里可以查询扫描结果
                curl -u admin:Harbor12345 \\
                  http://${HARBOR_URL}/api/v2.0/projects/${HARBOR_PROJECT}/repositories/${APP_NAME}/artifacts/${BUILD_NUMBER}/vulnerabilities
            """
        }
    }
}
```

#### 6.2 配置Jenkins凭证
1. Jenkins → 凭据 → 添加凭据
2. 类型: Username with password
3. ID: harbor-credentials
4. Username: admin
5. Password: Harbor12345

---

### 阶段7：停用旧Registry (5分钟)

#### 7.1 验证Harbor正常工作
确认以下操作都能成功:
- ✅ 推送镜像到Harbor
- ✅ 从Harbor拉取镜像
- ✅ Kubernetes能从Harbor拉取镜像
- ✅ CI/CD能推送到Harbor

#### 7.2 停用旧Registry
修改 `docker-compose.yml`:

```yaml
# 注释掉旧registry
# registry:
#   image: registry:2
#   container_name: my-cloud-registry
#   ...
```

重启:
```bash
docker-compose up -d
```

---

## 验证清单

完成后，请确认以下所有项都能正常工作:

- [ ] Harbor UI可以访问
- [ ] 可以登录Harbor
- [ ] 可以创建项目
- [ ] 可以推送镜像
- [ ] 可以拉取镜像
- [ ] 漏洞扫描正常
- [ ] Kubernetes能拉取镜像
- [ ] CI/CD能推送镜像
- [ ] 现有应用能正常部署

---

## 常用命令

### Harbor管理
```bash
# 查看Harbor状态
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose ps

# 重启Harbor
docker-compose restart

# 查看日志
docker-compose logs -f harbor-core

# 停止Harbor
docker-compose stop

# 启动Harbor
docker-compose start
```

### 镜像操作
```bash
# 登录
docker login localhost:8093 -u admin -p Harbor12345

# 推送镜像
docker push localhost:8093/mycloud/app-name:tag

# 拉取镜像
docker pull localhost:8093/mycloud/app-name:tag

# 列出项目中的镜像
curl -u admin:Harbor12345 http://localhost:8093/api/v2.0/projects/mycloud/repositories
```

---

## 预计时间

| 阶段 | 时间 |
|-----|------|
| 阶段1: 部署Harbor | 30分钟 |
| 阶段2: 配置Harbor | 15分钟 |
| 阶段3: 测试Harbor | 15分钟 |
| 阶段4: 集成到项目 | 60分钟 |
| 阶段5: 迁移镜像 | 30分钟 |
| 阶段6: 更新CI/CD | 30分钟 |
| 阶段7: 停用旧Registry | 5分钟 |
| **总计** | **~3小时** |

---

## 下一步

现在可以开始执行：

```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor
sudo ./setup-harbor.sh
```

然后按照本清单逐步完成集成！
