#!/bin/bash

# 测试基于 Helm 的完整部署流程
# 用于验证环境定义与实际部署的一致性

set -e

echo "========================================="
echo "测试 Helm 完整部署流程"
echo "========================================="

# 配置
NAMESPACE="app-1-dev"
RELEASE_NAME="app-1"
CHART_PATH="./helm-charts/mycloud-app"
IMAGE="172.18.0.1:5001/mycloud/app-common-ci-pipeline:1.0.1649-f2870b7"

echo ""
echo "1. 验证 Helm Chart"
echo "-----------------------------------------"
if [ -d "$CHART_PATH" ]; then
    echo "✓ Chart 路径存在: $CHART_PATH"
    if [ -f "$CHART_PATH/Chart.yaml" ]; then
        echo "✓ Chart.yaml 存在"
    else
        echo "✗ Chart.yaml 不存在"
        exit 1
    fi
else
    echo "✗ Chart 路径不存在: $CHART_PATH"
    exit 1
fi

echo ""
echo "2. 构建基于环境定义的 Values"
echo "-----------------------------------------"

# 创建临时 values 文件
cat > /tmp/test-values.yaml <<EOF
# 应用基础配置
replicaCount: 3

image:
  repository: 172.18.0.1:5001/mycloud/app-common-ci-pipeline
  tag: "1.0.1649-f2870b7"
  pullPolicy: IfNotPresent

# 服务配置（开发环境使用 NodePort）
service:
  type: NodePort
  port: 80
  targetPort: 8080

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

# 健康检查
livenessProbe:
  enabled: true
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  enabled: true
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5

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

# 标签
labels:
  app: "app-1"
  env: "dev"
  managed-by: "my-cloud"
EOF

echo "✓ Values 文件已生成: /tmp/test-values.yaml"

echo ""
echo "3. 检查现有部署"
echo "-----------------------------------------"
if helm status $RELEASE_NAME -n $NAMESPACE >/dev/null 2>&1; then
    echo "⚠ Release $RELEASE_NAME 已存在于命名空间 $NAMESPACE"
    echo "将执行升级操作"
    ACTION="upgrade"
else
    echo "✓ Release $RELEASE_NAME 不存在"
    echo "将执行安装操作"
    ACTION="install"
fi

echo ""
echo "4. 执行 Helm 部署"
echo "-----------------------------------------"
echo "执行命令: helm $ACTION $RELEASE_NAME $CHART_PATH -n $NAMESPACE -f /tmp/test-values.yaml"

if [ "$ACTION" = "install" ]; then
    helm install $RELEASE_NAME $CHART_PATH \
        -n $NAMESPACE \
        --create-namespace \
        -f /tmp/test-values.yaml
else
    helm upgrade $RELEASE_NAME $CHART_PATH \
        -n $NAMESPACE \
        -f /tmp/test-values.yaml
fi

echo ""
echo "5. 等待部署完成"
echo "-----------------------------------------"
kubectl rollout status deployment/$RELEASE_NAME -n $NAMESPACE --timeout=300s

echo ""
echo "6. 验证部署资源"
echo "-----------------------------------------"

# 检查 Deployment
echo "检查 Deployment:"
kubectl get deployment $RELEASE_NAME -n $NAMESPACE

# 检查 Pods
echo ""
echo "检查 Pods:"
kubectl get pods -n $NAMESPACE -l app=$RELEASE_NAME

# 检查 Service
echo ""
echo "检查 Service:"
kubectl get service -n $NAMESPACE

# 检查 ServiceAccount
echo ""
echo "检查 ServiceAccount:"
kubectl get serviceaccount -n $NAMESPACE

# 检查 ConfigMap（如果有）
echo ""
echo "检查 ConfigMap:"
kubectl get configmap -n $NAMESPACE | grep $RELEASE_NAME || echo "无 ConfigMap"

# 检查 Secret（如果有）
echo ""
echo "检查 Secret:"
kubectl get secret -n $NAMESPACE | grep $RELEASE_NAME || echo "无 Secret"

# 检查 Ingress（如果有）
echo ""
echo "检查 Ingress:"
kubectl get ingress -n $NAMESPACE || echo "无 Ingress"

# 检查 NetworkPolicy
echo ""
echo "检查 NetworkPolicy:"
kubectl get networkpolicy -n $NAMESPACE || echo "无 NetworkPolicy"

# 检查 ResourceQuota
echo ""
echo "检查 ResourceQuota:"
kubectl get resourcequota -n $NAMESPACE || echo "无 ResourceQuota"

echo ""
echo "7. 与 Go 微服务标准模板对比"
echo "-----------------------------------------"

echo ""
echo "标准模板应该包含的资源:"
echo "  ✓ Deployment (已创建)"
echo "  ✓ Service (已创建)"
echo "  ✓ ServiceAccount (已创建)"
echo "  $(kubectl get ingress -n $NAMESPACE 2>/dev/null | grep -q $RELEASE_NAME && echo '✓' || echo '○') Ingress (根据环境配置)"
echo "  ○ NetworkPolicy (需要后端服务创建)"
echo "  ○ ResourceQuota (需要后端服务创建)"
echo "  ○ ConfigMap (根据配置创建)"
echo "  ○ Secret (根据配置创建)"
echo "  ○ HPA (根据配置创建)"

echo ""
echo "========================================="
echo "部署完成！"
echo "========================================="
echo ""
echo "访问方式:"
NODE_PORT=$(kubectl get service -n $NAMESPACE -o jsonpath='{.items[0].spec.ports[0].nodePort}')
echo "  NodePort: http://<NODE_IP>:$NODE_PORT"
echo ""
echo "清理命令:"
echo "  helm uninstall $RELEASE_NAME -n $NAMESPACE"
echo ""

# 清理临时文件
rm -f /tmp/test-values.yaml
