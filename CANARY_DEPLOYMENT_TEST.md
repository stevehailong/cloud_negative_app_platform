# 金丝雀部署完整测试流程

## 修复内容总结

### 1. Label 策略
- **问题**: 主和 canary deployment 使用不同的 app label，导致 Service 只能匹配其中一个
- **修复**: 统一使用 `app: app-X` label，用 `version` label 区分版本

### 2. Service 流量控制
- **问题**: Service selector 固定为 `app: app-X-canary`，流量 100% 到 canary
- **修复**: Service selector 改为 `app: app-X`，通过副本数比例自动分配流量

### 3. Canary 删除逻辑
- **问题**: 缩容到 0 副本，Deployment 对象仍存在
- **修复**: 真正调用 K8s API 删除 Deployment 对象

### 4. 副本数计算
- **问题**: Canary 固定 1 副本，无法根据百分比调整
- **修复**: 根据 canary_percent 计算副本数（默认总数 5，按比例分配）

## 测试环境准备

### 前置条件
1. Docker Desktop 运行中
2. K8s 集群正常
3. 本地 registry (172.18.0.1:5001) 可用
4. 所有服务已启动

### 清理环境
```bash
# 删除旧的部署和服务
kubectl delete deployment app-8 app-8-canary -n app-8 --ignore-not-found
kubectl delete svc app-8-service -n app-8 --ignore-not-found

# 确认清理完成
kubectl get all -n app-8
```

## 测试步骤

### Step 1: 创建初始部署（模拟已有应用）

由于金丝雀部署需要一个已存在的 stable 版本，我们先创建一个初始版本：

```bash
# 使用 nginx 作为初始镜像（模拟旧版本）
curl -X POST http://localhost/api/v1/releases \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "appId": 8,
    "envId": 1,
    "releaseStrategy": "rolling",
    "releaseVersion": "v1.0.0",
    "imageUrl": "nginx:latest",
    "description": "初始版本部署"
  }'

# 审批并执行（省略审批步骤，直接调用内部 API）
```

**预期结果：**
- namespace `app-8` 创建
- deployment `app-8` 创建（5副本，nginx 镜像）
- service `app-8-service` 创建（selector: `app=app-8`）

### Step 2: 触发 CI 流水线构建新版本

```bash
# 执行 CI 流水线（book-service-ci）
curl -X POST http://localhost/api/v1/pipeline-runs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "pipelineId": PIPELINE_ID,
    "branch": "main",
    "operator": "admin"
  }'

# 等待构建完成，记录 artifact ID
```

**预期结果：**
- Jenkins 执行 docker build
- 镜像推送到 `172.18.0.1:5001/mycloud/book-service-ci:1.0.xxxx`
- Artifact 记录更新，imageUrl 填充

### Step 3: 创建金丝雀发布

```bash
# 基于 CI 产物创建金丝雀发布
curl -X POST http://localhost/api/v1/releases \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "appId": 8,
    "envId": 1,
    "pipelineRunId": PIPELINE_RUN_ID,
    "releaseStrategy": "canary",
    "releaseVersion": "v1.1.0",
    "imageUrl": "172.18.0.1:5001/mycloud/book-service-ci:1.0.xxxx",
    "canaryPercent": 20,
    "description": "金丝雀发布测试"
  }'

# 记录 release ID
RELEASE_ID=xxx
```

### Step 4: 审批发布

```bash
# 提交审批
curl -X POST http://localhost/api/v1/releases/$RELEASE_ID/submit \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"approverUserIds": [1]}'

# 审批通过
curl -X POST http://localhost/api/v1/releases/$RELEASE_ID/approve \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"comment": "同意发布"}'
```

### Step 5: 执行金丝雀部署

```bash
# 执行发布
curl -X POST http://localhost/api/v1/releases/$RELEASE_ID/execute \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**预期结果：**
- Deployment `app-8-canary` 创建（1副本，新镜像）
- Deployment `app-8` 保持不变（4副本，旧镜像）
- Service `app-8-service` selector 为 `app=app-8`，匹配两个 deployment
- 流量分配：20% 到 canary，80% 到 stable

**验证命令：**
```bash
# 查看 deployments
kubectl get deployment -n app-8 -o wide

# 查看 Service
kubectl get svc app-8-service -n app-8 -o yaml | grep -A 3 selector

# 查看 Pods 和 labels
kubectl get pods -n app-8 -L app,version --show-labels

# 预期输出：
# NAME                           READY   STATUS    LABELS
# app-8-xxxxxx                   1/1     Running   app=app-8,version=app-8
# app-8-xxxxxx                   1/1     Running   app=app-8,version=app-8
# app-8-xxxxxx                   1/1     Running   app=app-8,version=app-8
# app-8-xxxxxx                   1/1     Running   app=app-8,version=app-8
# app-8-canary-xxxxxx            1/1     Running   app=app-8,version=app-8-canary
```

### Step 6: 流量测试

```bash
# 获取 Service NodePort
NODE_PORT=$(kubectl get svc app-8-service -n app-8 -o jsonpath='{.spec.ports[0].nodePort}')

# 发送 100 次请求，统计响应
for i in {1..100}; do
  curl -s http://localhost:$NODE_PORT/ | grep -o "nginx\|book-service" >> /tmp/traffic_test.log
done

# 统计流量分布
echo "Stable (nginx): $(grep -c nginx /tmp/traffic_test.log)"
echo "Canary (book-service): $(grep -c book-service /tmp/traffic_test.log)"
rm /tmp/traffic_test.log
```

**预期结果：**
- 约 80 次命中 stable (nginx)
- 约 20 次命中 canary (book-service)

### Step 7: 确认金丝雀

```bash
# 确认金丝雀，全量发布
curl -X POST http://localhost/api/v1/releases/$RELEASE_ID/canary/confirm \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**预期结果：**
- Deployment `app-8` 更新为新镜像（5副本）
- Deployment `app-8-canary` 被删除
- Service 继续工作，所有流量到新版本

**验证命令：**
```bash
# 等待 10 秒让异步任务执行
sleep 10

# 查看 deployments（应该只有 app-8）
kubectl get deployment -n app-8

# 查看镜像版本
kubectl get deployment app-8 -n app-8 -o jsonpath='{.spec.template.spec.containers[0].image}'

# 查看 release 状态
curl http://localhost/api/v1/releases/$RELEASE_ID \
  -H "Authorization: Bearer YOUR_TOKEN" | jq '.data.releaseStatus, .data.canaryStatus'

# 预期输出：
# "success"
# "canary_confirmed"
```

### Step 8: 回滚测试（可选）

如果在 Step 6 之后不确认，而是回滚：

```bash
# 回滚金丝雀
curl -X POST http://localhost/api/v1/releases/$RELEASE_ID/canary/rollback \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**预期结果：**
- Deployment `app-8-canary` 被删除
- Deployment `app-8` 保持原样（旧版本）
- 流量 100% 到 stable

## 验证清单

### 金丝雀阶段
- [ ] 两个 deployment 都存在（app-8 和 app-8-canary）
- [ ] 两者的 `app` label 相同
- [ ] Service selector 匹配两者
- [ ] 副本数比例正确（默认 4:1）
- [ ] 流量分配符合比例（误差 ±5%）
- [ ] Pods 都是 Running 状态

### 确认后
- [ ] 只有 app-8 deployment 存在
- [ ] app-8-canary 已删除（不是 0 副本）
- [ ] app-8 使用新镜像
- [ ] app-8 副本数为 5
- [ ] Service 继续工作
- [ ] Release 状态为 success

### 回滚后
- [ ] 只有 app-8 deployment 存在
- [ ] app-8-canary 已删除
- [ ] app-8 使用旧镜像
- [ ] Release 状态为 rollback

## 常见问题排查

### 1. Service 找不到后端
**症状**: `curl` 返回 503 或超时  
**原因**: Service selector 不匹配 Pod labels  
**排查**:
```bash
kubectl get svc app-8-service -n app-8 -o yaml | grep -A 3 selector
kubectl get pods -n app-8 -L app,version
```

### 2. 流量 100% 到一个版本
**症状**: 所有请求都返回同一个响应  
**原因**: 只有一个 deployment 的 labels 匹配 Service selector  
**排查**:
```bash
kubectl get endpoints app-8-service -n app-8
kubectl describe svc app-8-service -n app-8
```

### 3. Canary 删除失败
**症状**: 确认后 canary deployment 仍然存在  
**原因**: DELETE API 调用失败  
**排查**:
```bash
docker logs --tail 50 my-cloud-release-service | grep -i canary
docker logs --tail 50 my-cloud-deploy-service | grep -i delete
```

### 4. 副本数不符合预期
**症状**: Canary 副本数不是 1  
**原因**: 计算逻辑错误或 canary_percent 设置不当  
**排查**:
```bash
docker logs my-cloud-release-service | grep "Canary strategy"
```

## 清理测试环境

```bash
# 删除所有测试资源
kubectl delete namespace app-8

# 删除数据库中的测试数据（可选）
docker exec my-cloud-mysql mysql -uroot -proot123456 release_db \
  -e "DELETE FROM releases WHERE app_id=8;"
```

## 总结

修复后的金丝雀部署实现了：
1. ✅ 正确的流量分配（通过副本数比例）
2. ✅ Service 自动管理和正确的 selector
3. ✅ Canary 真正删除（不是缩容）
4. ✅ 基于百分比的副本数计算
5. ✅ 完整的部署→观察→确认/回滚流程
