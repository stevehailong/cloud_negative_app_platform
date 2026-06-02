# Harbor 分阶段迁移计划

## 迁移策略：稳妥渐进式

### 原则
- ✅ 不影响现有服务
- ✅ 保留旧registry作为备份
- ✅ 逐个应用迁移
- ✅ 充分测试后再进行下一步

---

## 阶段0：准备期（已完成）

- [x] Harbor部署完成
- [x] Harbor基础测试通过
- [x] 旧registry (localhost:5001) 保持运行
- [x] Harbor (localhost:8093) 并行运行

---

## 阶段1：试点应用（第1周）

### 目标
选择1-2个非关键应用进行试点迁移

### 推荐试点应用
```
✓ 推荐：notification-service（通知服务，非核心）
✓ 推荐：monitor-service（监控服务，可以中断）
✗ 不推荐：gateway（核心服务）
✗ 不推荐：auth-service（认证服务）
```

### 操作步骤

#### 1.1 迁移镜像
```bash
# 假设迁移notification-service
APP_NAME="notification-service"
OLD_TAG="v1.0.0"

# 从旧registry拉取
docker pull localhost:5001/${APP_NAME}:${OLD_TAG}

# 重新打标签
docker tag localhost:5001/${APP_NAME}:${OLD_TAG} \
           localhost:8093/mycloud/${APP_NAME}:${OLD_TAG}

# 推送到Harbor
docker push localhost:8093/mycloud/${APP_NAME}:${OLD_TAG}
```

#### 1.2 更新部署配置
修改对应服务的Helm values或deployment配置：

**之前**:
```yaml
image: localhost:5001/notification-service:v1.0.0
```

**之后**:
```yaml
image: localhost:8093/mycloud/notification-service:v1.0.0
imagePullSecrets:
  - name: harbor-secret
```

#### 1.3 验证
- [ ] 应用能正常启动
- [ ] 应用功能正常
- [ ] 日志无异常
- [ ] 观察1-2天确保稳定

---

## 阶段2：非核心服务（第2周）

### 迁移应用列表
```
- [ ] monitor-service
- [ ] notification-service
- [ ] audit-service
- [ ] release-service
```

### 批量迁移脚本
创建 `scripts/migrate-to-harbor.sh`:

```bash
#!/bin/bash
# 批量迁移脚本

APPS=(
  "monitor-service:v1.0.0"
  "notification-service:v1.0.0"
  "audit-service:v1.0.0"
  "release-service:v1.0.0"
)

OLD_REGISTRY="localhost:5001"
NEW_REGISTRY="localhost:8093"
PROJECT="mycloud"

# 登录Harbor
docker login $NEW_REGISTRY -u admin -p Harbor12345

for APP_TAG in "${APPS[@]}"; do
  APP=$(echo $APP_TAG | cut -d: -f1)
  TAG=$(echo $APP_TAG | cut -d: -f2)
  
  echo "迁移: $APP:$TAG"
  
  # 拉取
  docker pull $OLD_REGISTRY/$APP:$TAG || echo "警告: 拉取失败，跳过"
  
  # 重新打标签
  docker tag $OLD_REGISTRY/$APP:$TAG $NEW_REGISTRY/$PROJECT/$APP:$TAG
  
  # 推送
  docker push $NEW_REGISTRY/$PROJECT/$APP:$TAG
  
  echo "✓ $APP:$TAG 迁移完成"
  echo ""
done

echo "所有应用迁移完成！"
```

---

## 阶段3：核心服务（第3周）

### 迁移应用列表
```
- [ ] gateway
- [ ] auth-service
- [ ] env-service
- [ ] cluster-service
- [ ] deploy-service
- [ ] application-service
- [ ] project-service
- [ ] pipeline-service
```

### 迁移策略
每个服务迁移后：
1. 先在测试环境验证
2. 灰度发布（部分Pod使用新镜像）
3. 全量切换
4. 观察24小时

### 灰度发布示例
```yaml
# deployment.yaml
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: gateway
        # 第1天：1个Pod用Harbor镜像，2个用旧镜像
        image: localhost:8093/mycloud/gateway:v1.0.0
        # 第2天：2个Pod用Harbor镜像，1个用旧镜像
        # 第3天：全部切换到Harbor
```

---

## 阶段4：CI/CD更新（第4周）

### 更新Jenkins Pipeline

#### 方式1：渐进式（推荐）
保留对旧registry的支持，添加Harbor支持：

```groovy
pipeline {
    environment {
        // 新增Harbor配置
        HARBOR_URL = 'localhost:8093'
        HARBOR_PROJECT = 'mycloud'
        
        // 保留旧配置
        OLD_REGISTRY = 'localhost:5001'
        
        // 使用哪个registry由参数控制
        USE_HARBOR = true
    }
    
    stages {
        stage('Build') {
            steps {
                script {
                    if (env.USE_HARBOR == 'true') {
                        sh "docker build -t ${HARBOR_URL}/${HARBOR_PROJECT}/${APP_NAME}:${BUILD_NUMBER} ."
                        sh "docker push ${HARBOR_URL}/${HARBOR_PROJECT}/${APP_NAME}:${BUILD_NUMBER}"
                    } else {
                        sh "docker build -t ${OLD_REGISTRY}/${APP_NAME}:${BUILD_NUMBER} ."
                        sh "docker push ${OLD_REGISTRY}/${APP_NAME}:${BUILD_NUMBER}"
                    }
                }
            }
        }
    }
}
```

#### 方式2：完全切换
所有新构建都使用Harbor：

```groovy
pipeline {
    environment {
        HARBOR_URL = 'localhost:8093'
        HARBOR_PROJECT = 'mycloud'
        HARBOR_CREDENTIALS = credentials('harbor-credentials')
    }
    
    stages {
        stage('Build & Push') {
            steps {
                sh """
                    docker build -t ${HARBOR_URL}/${HARBOR_PROJECT}/${APP_NAME}:${BUILD_NUMBER} .
                    echo \${HARBOR_CREDENTIALS_PSW} | docker login ${HARBOR_URL} -u \${HARBOR_CREDENTIALS_USR} --password-stdin
                    docker push ${HARBOR_URL}/${HARBOR_PROJECT}/${APP_NAME}:${BUILD_NUMBER}
                """
            }
        }
    }
}
```

---

## 阶段5：验证和清理（第5周）

### 验证清单
- [ ] 所有应用都使用Harbor镜像
- [ ] CI/CD构建推送到Harbor
- [ ] 没有新镜像推送到旧registry
- [ ] 观察1周，确保稳定

### 查询使用情况
```bash
# 检查哪些应用还在使用旧registry
kubectl get pods -A -o json | jq -r '.items[].spec.containers[].image' | grep "localhost:5001" | sort | uniq

# 检查哪些应用使用Harbor
kubectl get pods -A -o json | jq -r '.items[].spec.containers[].image' | grep "localhost:8093" | sort | uniq
```

### 停用旧Registry
确认所有应用迁移后：

```bash
cd /Users/hanhailong01/Downloads/my_cloud

# 备份docker-compose.yml
cp docker-compose.yml docker-compose.yml.bak

# 编辑docker-compose.yml，注释registry
# registry:
#   image: registry:2
#   ...
```

重启服务：
```bash
docker-compose up -d
```

---

## 回滚计划

### 如果Harbor出现问题

#### 临时回滚
1. 重新启动旧registry
   ```bash
   docker-compose up -d registry
   ```

2. 更新Pod镜像地址
   ```bash
   kubectl set image deployment/app-name \
     container-name=localhost:5001/app-name:v1.0.0
   ```

#### 完全回滚
1. 恢复docker-compose.yml备份
   ```bash
   cp docker-compose.yml.bak docker-compose.yml
   docker-compose up -d
   ```

2. 更新所有部署配置使用旧registry

---

## 监控指标

### 每周检查

#### Harbor健康状态
```bash
cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
docker-compose ps
```

#### 存储使用
```bash
du -sh /Users/hanhailong01/Downloads/my_cloud/harbor-data
```

#### 镜像统计
访问 Harbor UI → 项目 → mycloud
- 查看镜像数量
- 查看存储使用
- 查看下载次数

---

## 迁移时间表

| 周次 | 阶段 | 任务 | 预期结果 |
|-----|------|------|---------|
| 第1周 | 试点 | 迁移1-2个非核心应用 | 验证Harbor可用性 |
| 第2周 | 扩展 | 迁移4-6个非核心服务 | 建立迁移流程 |
| 第3周 | 核心 | 迁移核心服务 | 主要服务切换完成 |
| 第4周 | CI/CD | 更新构建流程 | 新构建使用Harbor |
| 第5周 | 收尾 | 验证并停用旧registry | 完全切换到Harbor |

---

## 成功标志

### 阶段目标
- ✅ 所有应用镜像都在Harbor中
- ✅ CI/CD推送到Harbor
- ✅ 旧registry没有新流量
- ✅ 运行稳定1周+

### 验证命令
```bash
# 检查Harbor中的镜像
curl -u admin:Harbor12345 http://localhost:8093/api/v2.0/projects/mycloud/repositories | jq

# 检查所有Pod的镜像来源
kubectl get pods -A -o json | jq -r '.items[].spec.containers[].image' | grep -c "localhost:8093"

# 应该看到所有镜像都来自localhost:8093
```

---

## 联系支持

遇到问题？

1. **查看日志**
   ```bash
   cd /Users/hanhailong01/Downloads/my_cloud/harbor/harbor
   docker-compose logs -f
   ```

2. **检查文档**
   - `/Users/hanhailong01/Downloads/my_cloud/docs/HARBOR_DEPLOYMENT_TEST.md`
   - `/Users/hanhailong01/Downloads/my_cloud/docs/HARBOR_INTEGRATION_GUIDE.md`

3. **社区支持**
   - Harbor GitHub: https://github.com/goharbor/harbor/issues
   - Harbor文档: https://goharbor.io/docs/

---

开始你的迁移之旅！记住：稳妥第一，逐步推进。🚀
