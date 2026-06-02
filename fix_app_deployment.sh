#!/bin/bash

# 修复 app-1-dev 命名空间的部署
# 使用 Helm 重新部署，确保与环境定义严格一致

set -e

echo "========================================="
echo "修复 app-1-dev 部署配置"
echo "========================================="

NAMESPACE="app-1-dev"
RELEASE_NAME="app-1"
CHART_PATH="./helm-charts/mycloud-app"

echo ""
echo "1. 备份现有配置"
echo "-----------------------------------------"

# 获取当前镜像
CURRENT_IMAGE=$(kubectl get deployment app-1 -n $NAMESPACE -o jsonpath='{.spec.template.spec.containers[0].image}')
echo "当前镜像: $CURRENT_IMAGE"

# 获取当前副本数
CURRENT_REPLICAS=$(kubectl get deployment app-1 -n $NAMESPACE -o jsonpath='{.spec.replicas}')
echo "当前副本数: $CURRENT_REPLICAS"

# 备份现有资源
mkdir -p /tmp/k8s-backup-$NAMESPACE
kubectl get deployment app-1 -n $NAMESPACE -o yaml > /tmp/k8s-backup-$NAMESPACE/deployment.yaml
kubectl get service -n $NAMESPACE -o yaml > /tmp/k8s-backup-$NAMESPACE/service.yaml
kubectl get serviceaccount -n $NAMESPACE -o yaml > /tmp/k8s-backup-$NAMESPACE/serviceaccount.yaml

echo "✓ 配置已备份到 /tmp/k8s-backup-$NAMESPACE"

echo ""
echo "2. 删除现有部署（保留 PVC 和 ConfigMap）"
echo "-----------------------------------------"

read -p "确认删除现有部署并使用 Helm 重新部署？ (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "取消操作"
    exit 1
fi

# 删除 Deployment
kubectl delete deployment app-1 -n $NAMESPACE 2>/dev/null || true

# 删除 Service
kubectl delete service app-1-service -n $NAMESPACE 2>/dev/null || true

# 删除 ServiceAccount（如果由 Helm 管理）
# kubectl delete serviceaccount app-1-sa -n $NAMESPACE 2>/dev/null || true

echo "✓ 现有资源已删除"

echo ""
echo "3. 创建 Helm Values（基于 Go 微服务标准模板）"
echo "-----------------------------------------"

cat > /tmp/app-1-dev-values.yaml <<EOF
# Go 微服务标准配置
# 对应开发环境

# 基础配置
replicaCount: $CURRENT_REPLICAS

image:
  repository: $(echo $CURRENT_IMAGE | cut -d: -f1)
  tag: "$(echo $CURRENT_IMAGE | cut -d: -f2)"
  pullPolicy: IfNotPresent

# 服务配置（开发环境使用 NodePort）
service:
  type: NodePort
  port: 80
  targetPort: 8080
  # nodePort: 31145  # 可选：指定 NodePort

# Ingress 配置（开发环境不启用）
ingress:
  enabled: false

# 资源配置
resources:
  limits:
    cpu: 1000m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

# 自动扩缩容（开发环境不启用）
autoscaling:
  enabled: false

# 健康检查
livenessProbe:
  enabled: true
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  enabled: true
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3

# 环境变量
env:
  - name: APP_ENV
    value: "development"
  - name: LOG_LEVEL
    value: "debug"
  - name: PORT
    value: "8080"

# ServiceAccount
serviceAccount:
  create: true
  name: "app-1-sa"

# Pod 安全上下文
podSecurityContext: {}

# 容器安全上下文
securityContext: {}

# 标签
labels:
  app: "app-1"
  env: "dev"
  managed-by: "my-cloud"

# Pod 注解
podAnnotations: {}

# 节点选择器
nodeSelector: {}

# 容忍度
tolerations: []

# 亲和性
affinity: {}
EOF

echo "✓ Values 文件已创建: /tmp/app-1-dev-values.yaml"

echo ""
echo "4. 使用 Helm 部署"
echo "-----------------------------------------"

# 检查 Helm 是否已安装
if ! command -v helm &> /dev/null; then
    echo "✗ Helm 未安装"
    exit 1
fi

# 检查 Chart 是否存在
if [ ! -d "$CHART_PATH" ]; then
    echo "✗ Chart 路径不存在: $CHART_PATH"
    exit 1
fi

# 使用 Helm 安装
helm install $RELEASE_NAME $CHART_PATH \
    -n $NAMESPACE \
    -f /tmp/app-1-dev-values.yaml

echo "✓ Helm 部署已启动"

echo ""
echo "5. 等待部署完成"
echo "-----------------------------------------"

kubectl rollout status deployment/$RELEASE_NAME -n $NAMESPACE --timeout=300s

echo "✓ 部署已完成"

echo ""
echo "6. 创建网络隔离资源"
echo "-----------------------------------------"

# 创建 NetworkPolicy（允许来自 ingress-nginx 的流量）
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: $NAMESPACE
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  # 允许来自同一命名空间的流量
  - from:
    - podSelector: {}
  # 允许来自 ingress-nginx 命名空间的流量（公网访问）
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
  egress:
  # 允许所有出站流量
  - {}
EOF

echo "✓ NetworkPolicy 已创建"

echo ""
echo "7. 创建资源配额"
echo "-----------------------------------------"

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ResourceQuota
metadata:
  name: default-quota
  namespace: $NAMESPACE
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    pods: "10"
    services: "5"
    secrets: "10"
    configmaps: "10"
EOF

echo "✓ ResourceQuota 已创建"

echo ""
echo "8. 验证部署结果"
echo "-----------------------------------------"

echo ""
echo "Deployment:"
kubectl get deployment -n $NAMESPACE

echo ""
echo "Pods:"
kubectl get pods -n $NAMESPACE

echo ""
echo "Services:"
kubectl get service -n $NAMESPACE

echo ""
echo "ServiceAccounts:"
kubectl get serviceaccount -n $NAMESPACE

echo ""
echo "NetworkPolicies:"
kubectl get networkpolicy -n $NAMESPACE

echo ""
echo "ResourceQuotas:"
kubectl get resourcequota -n $NAMESPACE

echo ""
echo "ConfigMaps:"
kubectl get configmap -n $NAMESPACE

echo ""
echo "Secrets:"
kubectl get secret -n $NAMESPACE

echo ""
echo "========================================="
echo "部署修复完成！"
echo "========================================="

echo ""
echo "对比结果:"
echo "-----------------------------------------"
echo "修复前:"
echo "  ✓ Deployment (3/3)"
echo "  ✓ Service (NodePort)"
echo "  ✓ ServiceAccount"
echo "  ✗ Ingress (未创建)"
echo "  ✗ NetworkPolicy (未创建)"
echo "  ✗ ResourceQuota (未创建)"
echo "  ✗ ConfigMap (未创建)"
echo "  ✗ Secret (未创建)"
echo "  ✗ HPA (未创建)"
echo ""
echo "修复后:"
echo "  ✓ Deployment (3/3)"
echo "  ✓ Service (NodePort)"
echo "  ✓ ServiceAccount"
echo "  ○ Ingress (开发环境不启用)"
echo "  ✓ NetworkPolicy (已创建)"
echo "  ✓ ResourceQuota (已创建)"
echo "  ○ ConfigMap (未配置)"
echo "  ○ Secret (未配置)"
echo "  ○ HPA (开发环境不启用)"

echo ""
echo "访问方式:"
NODE_PORT=$(kubectl get service -n $NAMESPACE -o jsonpath='{.items[0].spec.ports[0].nodePort}')
echo "  NodePort: http://<NODE_IP>:$NODE_PORT"

echo ""
echo "清理备份文件:"
echo "  rm -rf /tmp/k8s-backup-$NAMESPACE"
echo "  rm -f /tmp/app-1-dev-values.yaml"
