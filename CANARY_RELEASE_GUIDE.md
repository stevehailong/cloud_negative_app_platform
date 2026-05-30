# 完整的灰度发布验证方案

## 目录
1. [前置准备](#前置准备)
2. [场景设定](#场景设定)
3. [Step 1: 创建应用并部署初始版本](#step-1-创建应用并部署初始版本)
4. [Step 2: 本地访问配置](#step-2-本地访问配置)
5. [Step 3: 验证初始版本](#step-3-验证初始版本)
6. [Step 4: 构建新版本镜像](#step-4-构建新版本镜像)
7. [Step 5: 创建金丝雀发布](#step-5-创建金丝雀发布)
8. [Step 6: 验证金丝雀流量分配](#step-6-验证金丝雀流量分配)
9. [Step 7: 确认金丝雀](#step-7-确认金丝雀)
10. [Step 8: 验证全量发布](#step-8-验证全量发布)
11. [回滚方案](#回滚方案)
12. [常见问题](#常见问题)

---

## 前置准备

### 环境检查
```bash
# 1. 检查所有服务运行状态
docker-compose ps

# 2. 检查 K8s 集群
kubectl get nodes

# 3. 检查 Ingress Controller
kubectl get pods -n ingress-nginx

# 4. 检查本地 registry
curl http://172.18.0.1:5001/v2/_catalog
```

### 登录系统
```bash
# 获取访问 Token
TOKEN=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}' | \
  python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo "Token: ${TOKEN:0:50}..."
```

---

## 场景设定

**假设场景**：部署一个 Book Service 应用
- **应用 ID**: 8 (book-service)
- **初始版本**: v1.0 (nginx:alpine - 模拟旧版本)
- **新版本**: v2.0 (httpd:alpine - 模拟新版本)
- **灰度策略**: 20% 流量到新版本

---

## Step 1: 创建应用并部署初始版本

### 1.1 创建初始部署（v1.0 - Stable）

```bash
# 通过 deploy-service 内部 API 创建初始部署
curl -X POST http://localhost:8087/internal/v1/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "releaseId": 1,
    "clusterId": 1,
    "namespace": "app-8",
    "workloadName": "app-8",
    "workloadType": "deployment",
    "imageVersion": "nginx:alpine",
    "desiredReplicas": 4
  }' | python3 -m json.tool
```

**预期输出**：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 36,
        "deploymentStatus": "progressing"
    }
}
```

### 1.2 等待部署完成

```bash
# 等待 30 秒让部署完成
sleep 30

# 检查部署状态
kubectl get deployment,pods,svc -n app-8
```

**预期结果**：
```
NAME                    READY   UP-TO-DATE   AVAILABLE
deployment.apps/app-8   4/4     4            4

NAME                         READY   STATUS    RESTARTS
pod/app-8-xxx                1/1     Running   0
pod/app-8-xxx                1/1     Running   0
pod/app-8-xxx                1/1     Running   0
pod/app-8-xxx                1/1     Running   0

NAME                    TYPE       CLUSTER-IP      PORT(S)
service/app-8-service   NodePort   10.96.xxx.xxx   80:31xxx/TCP
```

### 1.3 验证隔离措施

```bash
# 检查 NetworkPolicy
kubectl get networkpolicy -n app-8

# 检查 ResourceQuota
kubectl get resourcequota -n app-8

# 检查 RBAC
kubectl get sa,role,rolebinding -n app-8
```

---

## Step 2: 本地访问配置

### 方式 1: 通过 Ingress（推荐）

#### 2.1 创建 Ingress 资源

```bash
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-8-ingress
  namespace: app-8
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
  - host: book-service.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: app-8-service
            port:
              number: 80
EOF
```

#### 2.2 配置本地 hosts

```bash
# macOS / Linux
sudo bash -c 'echo "127.0.0.1 book-service.local" >> /etc/hosts'

# 验证
cat /etc/hosts | grep book-service
```

#### 2.3 获取 Ingress NodePort

```bash
# 查看 Ingress Controller 的 NodePort
kubectl get svc ingress-nginx-controller -n ingress-nginx

# 通常是 30080 (HTTP) 和 30443 (HTTPS)
```

**访问地址**：
```
http://book-service.local:30080/
```

### 方式 2: 通过 Port-Forward（适合开发调试）

```bash
# 转发 Service 到本地端口
kubectl port-forward -n app-8 svc/app-8-service 8888:80 &

# 访问地址
http://localhost:8888/
```

### 方式 3: 通过 API Gateway（需要配置路由）

需要在 Gateway 中添加路由规则（暂不推荐，增加复杂度）。

---

## Step 3: 验证初始版本

### 3.1 通过 Ingress 访问

```bash
# 访问应用
curl http://book-service.local:30080/

# 预期：返回 nginx 欢迎页面
```

**预期响应**：
```html
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
...
```

### 3.2 验证流量分配

```bash
# 发送 20 次请求，应该全部命中 stable 版本
for i in {1..20}; do
  curl -s http://book-service.local:30080/ | grep -o "nginx" || echo "other"
done | sort | uniq -c
```

**预期输出**：
```
  20 nginx
```

### 3.3 检查 Pod 标签

```bash
kubectl get pods -n app-8 -L app,version --show-labels
```

**预期**：
```
NAME          READY   STATUS    AGE   APP     VERSION   LABELS
app-8-xxx     1/1     Running   5m    app-8   app-8     app=app-8,version=app-8,...
```

---

## Step 4: 构建新版本镜像

### 方式 1: 使用真实 CI 流水线（推荐）

```bash
# 1. 触发 CI 流水线
curl -X POST "http://localhost/api/v1/pipeline-runs" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pipelineId": 1,
    "branch": "main",
    "operator": "admin"
  }' | python3 -m json.tool

# 2. 查看构建日志
# 访问 Jenkins: http://localhost:9090/

# 3. 等待构建完成，记录生成的镜像地址
# 例如: 172.18.0.1:5001/mycloud/book-service-ci:1.0.1234-abc123
```

### 方式 2: 直接使用不同镜像模拟（快速测试）

使用 `httpd:alpine` 模拟新版本：
```bash
# 无需额外操作，直接进入 Step 5
NEW_IMAGE="httpd:alpine"
```

---

## Step 5: 创建金丝雀发布

### 5.1 创建 Canary Deployment

```bash
# 创建金丝雀部署（1 副本，新版本镜像）
curl -X POST http://localhost:8087/internal/v1/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "releaseId": 2,
    "clusterId": 1,
    "namespace": "app-8",
    "workloadName": "app-8-canary",
    "workloadType": "deployment",
    "imageVersion": "httpd:alpine",
    "desiredReplicas": 1
  }' | python3 -m json.tool
```

### 5.2 等待 Canary 部署完成

```bash
# 等待 20 秒
sleep 20

# 检查两个 deployment
kubectl get deployment -n app-8
```

**预期结果**：
```
NAME           READY   UP-TO-DATE   AVAILABLE   AGE
app-8          4/4     4            4           10m
app-8-canary   1/1     1            1           30s
```

### 5.3 验证 Service Endpoints

```bash
# 查看 Service 后端 Endpoints
kubectl get endpoints app-8-service -n app-8

# 应该看到 5 个 endpoint (4 stable + 1 canary)
```

### 5.4 验证 Pod 标签

```bash
kubectl get pods -n app-8 -L app,version
```

**预期**：
```
NAME                    READY   STATUS    APP     VERSION
app-8-xxx               1/1     Running   app-8   app-8
app-8-xxx               1/1     Running   app-8   app-8
app-8-xxx               1/1     Running   app-8   app-8
app-8-xxx               1/1     Running   app-8   app-8
app-8-canary-xxx        1/1     Running   app-8   app-8-canary
```

**关键点**：
- ✅ 5 个 Pod 的 `app` label 都是 `app-8`（Service 通过这个 label 选择）
- ✅ `version` label 用于区分 stable 和 canary

---

## Step 6: 验证金丝雀流量分配

### 6.1 测试流量分配比例

```bash
# 从集群内测试（避免本地网络问题）
kubectl run traffic-test --image=curlimages/curl:latest --rm -i --restart=Never -n app-8 -- sh -c '
nginx_count=0
httpd_count=0
for i in $(seq 1 100); do
  response=$(curl -s app-8-service/)
  if echo "$response" | grep -q "Welcome to nginx"; then
    nginx_count=$((nginx_count + 1))
  elif echo "$response" | grep -q "It works"; then
    httpd_count=$((httpd_count + 1))
  fi
done
total=$((nginx_count + httpd_count))
echo "Total requests: $total"
echo "Stable (nginx): $nginx_count ($((nginx_count * 100 / total))%)"
echo "Canary (httpd): $httpd_count ($((httpd_count * 100 / total))%)"
'
```

**预期输出**：
```
Total requests: 100
Stable (nginx): 80 (80%)
Canary (httpd): 20 (20%)
```

**流量比例**: 4:1 (stable:canary) = 80%:20%

### 6.2 从本地测试（通过 Ingress）

```bash
# 发送 50 次请求
nginx_count=0
httpd_count=0

for i in {1..50}; do
  response=$(curl -s http://book-service.local:30080/)
  if echo "$response" | grep -q "nginx"; then
    ((nginx_count++))
    echo -n "n"
  elif echo "$response" | grep -q "It works"; then
    ((httpd_count++))
    echo -n "h"
  fi
done

echo ""
echo "Stable (nginx): $nginx_count"
echo "Canary (httpd): $httpd_count"
```

### 6.3 观察 Canary 健康状态

```bash
# 查看 Canary Pod 状态
kubectl get pods -n app-8 -l version=app-8-canary -o wide

# 查看 Canary Pod 日志
CANARY_POD=$(kubectl get pods -n app-8 -l version=app-8-canary -o jsonpath='{.items[0].metadata.name}')
kubectl logs $CANARY_POD -n app-8 --tail=20

# 查看 Canary Pod 事件
kubectl describe pod $CANARY_POD -n app-8 | grep -A 10 Events:
```

### 6.4 监控指标（可选）

如果集成了 Prometheus/Grafana：
```bash
# 查询 Canary 错误率
# 查询 Canary 响应时间
# 对比 Stable vs Canary 指标
```

---

## Step 7: 确认金丝雀

### 场景 A: Canary 验证成功，全量发布

#### 7.1 更新 Stable Deployment

```bash
# 将 stable deployment 更新为新版本镜像，并扩容到 5 副本
curl -X POST http://localhost:8087/internal/v1/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "releaseId": 3,
    "clusterId": 1,
    "namespace": "app-8",
    "workloadName": "app-8",
    "workloadType": "deployment",
    "imageVersion": "httpd:alpine",
    "desiredReplicas": 5
  }' | python3 -m json.tool
```

#### 7.2 删除 Canary Deployment

```bash
# 删除 canary deployment（K8s 资源 + 数据库记录）
curl -X DELETE "http://localhost:8087/internal/v1/k8s/deployments/app-8/app-8-canary"

# 删除数据库中的 canary 记录
curl -X DELETE "http://localhost:8087/internal/v1/deployments/by-workload?namespace=app-8&workloadName=app-8-canary"
```

#### 7.3 等待滚动更新完成

```bash
# 监控滚动更新进度
kubectl rollout status deployment/app-8 -n app-8

# 查看部署状态
kubectl get deployment,pods -n app-8
```

**预期结果**：
```
NAME                    READY   UP-TO-DATE   AVAILABLE
deployment.apps/app-8   5/5     5            5

NAME                         READY   STATUS    RESTARTS
pod/app-8-xxx                1/1     Running   0
pod/app-8-xxx                1/1     Running   0
pod/app-8-xxx                1/1     Running   0
pod/app-8-xxx                1/1     Running   0
pod/app-8-xxx                1/1     Running   0

# app-8-canary 已删除
```

---

## Step 8: 验证全量发布

### 8.1 验证所有流量到新版本

```bash
# 发送 20 次请求，应该全部命中新版本
for i in {1..20}; do
  curl -s http://book-service.local:30080/ | grep -o "It works" || echo "other"
done | sort | uniq -c
```

**预期输出**：
```
  20 It works
```

### 8.2 验证 Pod 镜像

```bash
kubectl get deployment app-8 -n app-8 -o jsonpath='{.spec.template.spec.containers[0].image}'
```

**预期输出**：
```
httpd:alpine
```

### 8.3 验证 Service Endpoints

```bash
kubectl get endpoints app-8-service -n app-8 -o json | \
  python3 -c "import sys,json; print(f'Total endpoints: {len(json.load(sys.stdin)[\"subsets\"][0][\"addresses\"])}')"
```

**预期输出**：
```
Total endpoints: 5
```

### 8.4 功能测试

```bash
# 测试应用的主要功能接口
curl http://book-service.local:30080/api/books
curl http://book-service.local:30080/api/health

# 检查返回状态码
curl -o /dev/null -s -w "%{http_code}\n" http://book-service.local:30080/
```

---

## 回滚方案

### 场景 B: Canary 验证失败，需要回滚

#### B.1 删除 Canary Deployment

```bash
# 删除 canary
curl -X DELETE "http://localhost:8087/internal/v1/k8s/deployments/app-8/app-8-canary"
curl -X DELETE "http://localhost:8087/internal/v1/deployments/by-workload?namespace=app-8&workloadName=app-8-canary"
```

#### B.2 验证回滚结果

```bash
# 检查只有 stable deployment
kubectl get deployment -n app-8

# 验证流量 100% 到旧版本
for i in {1..10}; do
  curl -s http://book-service.local:30080/ | grep -o "nginx"
done | wc -l
# 预期输出: 10
```

### 场景 C: 全量发布后需要回滚

#### C.1 使用 K8s 原生回滚

```bash
# 查看 revision 历史
kubectl rollout history deployment/app-8 -n app-8

# 回滚到上一个版本
kubectl rollout undo deployment/app-8 -n app-8

# 监控回滚进度
kubectl rollout status deployment/app-8 -n app-8
```

#### C.2 验证回滚

```bash
# 检查镜像版本
kubectl get deployment app-8 -n app-8 -o jsonpath='{.spec.template.spec.containers[0].image}'
# 预期: nginx:alpine

# 测试访问
curl -s http://book-service.local:30080/ | grep "nginx"
```

---

## 常见问题

### Q1: 访问 Ingress 返回 503 Service Temporarily Unavailable

**原因**：后端 Pod 未就绪

**解决**：
```bash
# 检查 Pod 状态
kubectl get pods -n app-8

# 查看 Pod 日志
kubectl logs <pod-name> -n app-8

# 检查 Service Endpoints
kubectl get endpoints app-8-service -n app-8
```

### Q2: 流量分配不符合预期（全部到 stable 或 canary）

**原因**：Label 配置错误或 Service selector 不匹配

**检查**：
```bash
# 检查 Service selector
kubectl get svc app-8-service -n app-8 -o yaml | grep -A 3 selector

# 检查 Pod labels
kubectl get pods -n app-8 -L app,version
```

**修复**：
```bash
# Service selector 应该是: app=app-8
# 所有 Pod（stable + canary）的 app label 都应该是 app-8
```

### Q3: Canary 删除后 stable 流量中断

**原因**：Service selector 只匹配 canary

**预防**：
- 确保 Service selector 匹配公共 label (`app=app-8`)
- 不要使用 `version` label 作为 Service selector

### Q4: hosts 配置后仍然无法访问

**排查**：
```bash
# 1. 验证 DNS 解析
ping book-service.local

# 2. 检查端口是否正确（应该是 30080，不是 8080）
curl -I http://book-service.local:30080/

# 3. 检查 Ingress Controller 状态
kubectl get pods -n ingress-nginx

# 4. 查看 Ingress 配置
kubectl describe ingress app-8-ingress -n app-8
```

### Q5: Port-forward 连接断开

**解决**：
```bash
# 后台运行 port-forward
nohup kubectl port-forward -n app-8 svc/app-8-service 8888:80 > /dev/null 2>&1 &

# 查看进程
ps aux | grep port-forward

# 停止 port-forward
pkill -f "port-forward.*app-8-service"
```

---

## 完整验证清单

### ✅ 部署阶段
- [ ] 初始 stable deployment 创建成功（4 副本）
- [ ] Service 自动创建
- [ ] NetworkPolicy 创建
- [ ] ResourceQuota 创建
- [ ] ServiceAccount + RBAC 创建
- [ ] 所有 Pod 状态 Running

### ✅ 本地访问配置
- [ ] Ingress 资源创建
- [ ] hosts 文件配置
- [ ] 能通过 `http://book-service.local:30080/` 访问
- [ ] 返回 nginx 欢迎页面

### ✅ 金丝雀部署
- [ ] Canary deployment 创建成功（1 副本）
- [ ] Service 后端有 5 个 Endpoints
- [ ] 所有 Pod app label 一致
- [ ] 流量分配接近 80%:20%
- [ ] Canary Pod 健康运行

### ✅ 全量发布
- [ ] Stable deployment 更新完成（5 副本，新镜像）
- [ ] Canary deployment 删除
- [ ] 100% 流量到新版本
- [ ] 功能正常

---

## 总结

这个方案涵盖了：
1. ✅ 完整的灰度发布流程（stable → canary → full）
2. ✅ 本地访问配置（Ingress + hosts）
3. ✅ 流量验证方法（集群内 + 本地）
4. ✅ 回滚机制（canary 回滚 + 全量回滚）
5. ✅ 问题排查清单

**核心要点**：
- 通过 **Ingress NodePort 30080** 访问应用
- 使用 **统一的 app label** 实现流量分配
- 通过 **副本数比例** 控制流量（4:1 = 80%:20%）
- **渐进式发布**：stable → canary → full
- **零停机**：整个过程应用持续可用
