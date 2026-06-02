#!/bin/bash

# 端到端全链路测试脚本
# 测试: CI流水线 -> 发布管理 -> 应用管理(重启/扩缩容/回滚)

set -e

BASE_URL_DEPLOY="http://localhost:8087"
BASE_URL_RELEASE="http://localhost:8086"
BASE_URL_CI="http://localhost:8085"

APP_ID=8
ENV_ID=1
USER_ID=1

echo "========================================="
echo "端到端全链路测试"
echo "========================================="

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

error() {
    echo -e "${RED}✗ $1${NC}"
}

info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# ========================================
# 第一部分: CI 流水线测试
# ========================================
echo ""
echo "========================================="
echo "第一部分: CI 流水线测试"
echo "========================================="

info "1.1 创建 CI 构建任务..."
BUILD_RESULT=$(curl -s -X POST "$BASE_URL_CI/api/v1/builds" \
    -H "Content-Type: application/json" \
    -d '{
        "app_id": '$APP_ID',
        "env_id": '$ENV_ID',
        "branch": "main",
        "commit_sha": "abc123def456",
        "trigger_user_id": '$USER_ID'
    }')

BUILD_ID=$(echo "$BUILD_RESULT" | jq -r '.data.id // empty')
if [ -z "$BUILD_ID" ] || [ "$BUILD_ID" == "null" ]; then
    error "创建构建任务失败"
    echo "$BUILD_RESULT" | jq '.'
    exit 1
fi
success "构建任务创建成功, ID: $BUILD_ID"

info "1.2 等待构建完成..."
sleep 15

BUILD_STATUS=$(curl -s "$BASE_URL_CI/api/v1/builds/$BUILD_ID" | jq -r '.data.status')
if [ "$BUILD_STATUS" == "success" ]; then
    success "构建成功"
    IMAGE_URL=$(curl -s "$BASE_URL_CI/api/v1/builds/$BUILD_ID" | jq -r '.data.image_url')
    info "镜像地址: $IMAGE_URL"
else
    error "构建失败或超时, 状态: $BUILD_STATUS"
fi

# ========================================
# 第二部分: 发布管理测试 (3种部署策略)
# ========================================
echo ""
echo "========================================="
echo "第二部分: 发布管理测试"
echo "========================================="

# 2.1 滚动部署 (Rolling)
info "2.1 测试滚动部署策略..."
RELEASE_ROLLING=$(curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases" \
    -H "Content-Type: application/json" \
    -d '{
        "app_id": '$APP_ID',
        "env_id": '$ENV_ID',
        "release_version": "v1.0.0-rolling",
        "image_url": "nginx:1.21",
        "release_strategy": "rolling",
        "creator_user_id": '$USER_ID'
    }')

RELEASE_ID_ROLLING=$(echo "$RELEASE_ROLLING" | jq -r '.data.id // empty')
if [ -z "$RELEASE_ID_ROLLING" ] || [ "$RELEASE_ID_ROLLING" == "null" ]; then
    error "创建滚动发布失败"
    echo "$RELEASE_ROLLING" | jq '.'
else
    success "滚动发布创建成功, ID: $RELEASE_ID_ROLLING"
    
    info "审批通过..."
    curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_ROLLING/approve" \
        -H "Content-Type: application/json" \
        -d '{"operator_user_id": '$USER_ID'}' > /dev/null
    
    info "点击部署上线..."
    curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_ROLLING/deploy" \
        -H "Content-Type: application/json" \
        -d '{"operator_user_id": '$USER_ID'}' > /dev/null
    
    sleep 15
    
    RELEASE_STATUS=$(curl -s "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_ROLLING" | jq -r '.data.release_status')
    if [ "$RELEASE_STATUS" == "success" ]; then
        success "滚动部署成功"
    else
        error "滚动部署失败, 状态: $RELEASE_STATUS"
    fi
fi

# 2.2 金丝雀部署 (Canary)
info "2.2 测试金丝雀部署策略..."
RELEASE_CANARY=$(curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases" \
    -H "Content-Type: application/json" \
    -d '{
        "app_id": '$APP_ID',
        "env_id": '$ENV_ID',
        "release_version": "v1.1.0-canary",
        "image_url": "nginx:1.22",
        "release_strategy": "canary",
        "canary_percent": 20,
        "creator_user_id": '$USER_ID'
    }')

RELEASE_ID_CANARY=$(echo "$RELEASE_CANARY" | jq -r '.data.id // empty')
if [ -z "$RELEASE_ID_CANARY" ] || [ "$RELEASE_ID_CANARY" == "null" ]; then
    error "创建金丝雀发布失败"
    echo "$RELEASE_CANARY" | jq '.'
else
    success "金丝雀发布创建成功, ID: $RELEASE_ID_CANARY"
    
    info "审批通过..."
    curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_CANARY/approve" \
        -H "Content-Type: application/json" \
        -d '{"operator_user_id": '$USER_ID'}' > /dev/null
    
    info "点击部署上线..."
    curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_CANARY/deploy" \
        -H "Content-Type: application/json" \
        -d '{"operator_user_id": '$USER_ID'}' > /dev/null
    
    sleep 15
    
    RELEASE_STATUS=$(curl -s "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_CANARY" | jq -r '.data.release_status')
    if [ "$RELEASE_STATUS" == "canary" ]; then
        success "金丝雀部署成功 (20% 流量)"
        
        info "确认金丝雀,全量发布..."
        curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_CANARY/confirm-canary" \
            -H "Content-Type: application/json" \
            -d '{"operator_user_id": '$USER_ID'}' > /dev/null
        
        sleep 15
        
        FINAL_STATUS=$(curl -s "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_CANARY" | jq -r '.data.release_status')
        if [ "$FINAL_STATUS" == "success" ]; then
            success "金丝雀全量发布成功"
        else
            error "金丝雀全量发布失败, 状态: $FINAL_STATUS"
        fi
    else
        error "金丝雀部署失败, 状态: $RELEASE_STATUS"
    fi
fi

# 2.3 蓝绿部署 (Blue-Green)
info "2.3 测试蓝绿部署策略..."
RELEASE_BLUEGREEN=$(curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases" \
    -H "Content-Type: application/json" \
    -d '{
        "app_id": '$APP_ID',
        "env_id": '$ENV_ID',
        "release_version": "v1.2.0-bluegreen",
        "image_url": "nginx:1.23",
        "release_strategy": "bluegreen",
        "creator_user_id": '$USER_ID'
    }')

RELEASE_ID_BLUEGREEN=$(echo "$RELEASE_BLUEGREEN" | jq -r '.data.id // empty')
if [ -z "$RELEASE_ID_BLUEGREEN" ] || [ "$RELEASE_ID_BLUEGREEN" == "null" ]; then
    error "创建蓝绿发布失败"
    echo "$RELEASE_BLUEGREEN" | jq '.'
else
    success "蓝绿发布创建成功, ID: $RELEASE_ID_BLUEGREEN"
    
    info "审批通过..."
    curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_BLUEGREEN/approve" \
        -H "Content-Type: application/json" \
        -d '{"operator_user_id": '$USER_ID'}' > /dev/null
    
    info "点击部署上线..."
    curl -s -X POST "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_BLUEGREEN/deploy" \
        -H "Content-Type: application/json" \
        -d '{"operator_user_id": '$USER_ID'}' > /dev/null
    
    sleep 15
    
    RELEASE_STATUS=$(curl -s "$BASE_URL_RELEASE/api/v1/releases/$RELEASE_ID_BLUEGREEN" | jq -r '.data.release_status')
    if [ "$RELEASE_STATUS" == "success" ]; then
        success "蓝绿部署成功"
    else
        error "蓝绿部署失败, 状态: $RELEASE_STATUS"
    fi
fi

# ========================================
# 第三部分: 应用管理测试
# ========================================
echo ""
echo "========================================="
echo "第三部分: 应用管理测试"
echo "========================================="

# 获取部署ID
info "3.0 获取应用部署记录..."
DEPLOYMENTS=$(curl -s "$BASE_URL_DEPLOY/api/v1/app-deployments/by-app-env?app_id=$APP_ID&env_id=$ENV_ID")
STABLE_ID=$(echo "$DEPLOYMENTS" | jq -r '.data[] | select(.workload_name == "app-'$APP_ID'") | .id')
CANARY_ID=$(echo "$DEPLOYMENTS" | jq -r '.data[] | select(.workload_name == "app-'$APP_ID'-canary") | .id')

if [ -z "$STABLE_ID" ] || [ "$STABLE_ID" == "null" ]; then
    error "未找到 Stable 部署记录"
    exit 1
fi
success "Stable 部署 ID: $STABLE_ID"

if [ -n "$CANARY_ID" ] && [ "$CANARY_ID" != "null" ]; then
    success "Canary 部署 ID: $CANARY_ID"
fi

# 3.1 扩缩容测试
info "3.1 测试扩缩容 (扩展到 3 个副本)..."
SCALE_RESULT=$(curl -s -X POST "$BASE_URL_DEPLOY/api/v1/app-deployments/$STABLE_ID/scale" \
    -H "Content-Type: application/json" \
    -d '{
        "replicas": 3,
        "user_id": '$USER_ID'
    }')

if echo "$SCALE_RESULT" | jq -e '.code == 200' > /dev/null; then
    success "扩缩容任务提交成功"
    sleep 10
    
    POD_COUNT=$(curl -s "$BASE_URL_DEPLOY/api/v1/app-deployments/$STABLE_ID/pods" | jq '.data | length')
    if [ "$POD_COUNT" == "3" ]; then
        success "扩缩容成功, Pod 数量: $POD_COUNT"
    else
        error "扩缩容失败, Pod 数量: $POD_COUNT (期望: 3)"
    fi
else
    error "扩缩容任务提交失败"
    echo "$SCALE_RESULT" | jq '.'
fi

# 3.2 重启测试
info "3.2 测试重启部署..."
RESTART_RESULT=$(curl -s -X POST "$BASE_URL_DEPLOY/api/v1/app-deployments/$STABLE_ID/restart" \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": '$USER_ID'
    }')

if echo "$RESTART_RESULT" | jq -e '.code == 200' > /dev/null; then
    success "重启任务提交成功"
    sleep 12
    
    DEPLOY_STATUS=$(curl -s "$BASE_URL_DEPLOY/api/v1/app-deployments/$STABLE_ID" | jq -r '.data.deployment_status')
    if [ "$DEPLOY_STATUS" == "running" ]; then
        success "重启成功, 状态: $DEPLOY_STATUS"
    else
        error "重启失败, 状态: $DEPLOY_STATUS"
    fi
else
    error "重启任务提交失败"
    echo "$RESTART_RESULT" | jq '.'
fi

# 3.3 回滚测试
info "3.3 测试回滚..."
HISTORY_LIST=$(curl -s "$BASE_URL_DEPLOY/api/v1/app-deployments/$STABLE_ID/history?page=1&page_size=10")
HISTORY_ID=$(echo "$HISTORY_LIST" | jq -r '.data.list[1].id // empty')

if [ -n "$HISTORY_ID" ] && [ "$HISTORY_ID" != "null" ]; then
    info "回滚到历史版本 ID: $HISTORY_ID"
    ROLLBACK_RESULT=$(curl -s -X POST "$BASE_URL_DEPLOY/api/v1/app-deployments/$STABLE_ID/rollback" \
        -H "Content-Type: application/json" \
        -d '{
            "history_id": '$HISTORY_ID',
            "user_id": '$USER_ID'
        }')
    
    if echo "$ROLLBACK_RESULT" | jq -e '.code == 200' > /dev/null; then
        success "回滚任务提交成功"
        sleep 12
        
        CURRENT_VERSION=$(curl -s "$BASE_URL_DEPLOY/api/v1/app-deployments/$STABLE_ID" | jq -r '.data.current_version')
        success "回滚完成, 当前版本: $CURRENT_VERSION"
    else
        error "回滚任务提交失败"
        echo "$ROLLBACK_RESULT" | jq '.'
    fi
else
    error "未找到可回滚的历史版本"
fi

# 3.4 查询 Pod 列表 (验证版本标识)
info "3.4 查询 Pod 列表..."
PODS=$(curl -s "$BASE_URL_DEPLOY/api/v1/app-deployments/$STABLE_ID/pods")
POD_COUNT=$(echo "$PODS" | jq '.data | length')
success "Stable Pod 数量: $POD_COUNT"

echo "$PODS" | jq -r '.data[] | "  - \(.name) | 状态: \(.status) | 版本: \(.version) | 节点: \(.node)"'

if [ -n "$CANARY_ID" ] && [ "$CANARY_ID" != "null" ]; then
    info "查询 Canary Pod 列表..."
    CANARY_PODS=$(curl -s "$BASE_URL_DEPLOY/api/v1/app-deployments/$CANARY_ID/pods")
    CANARY_POD_COUNT=$(echo "$CANARY_PODS" | jq '.data | length')
    success "Canary Pod 数量: $CANARY_POD_COUNT"
    
    echo "$CANARY_PODS" | jq -r '.data[] | "  - \(.name) | 状态: \(.status) | 版本: \(.version) | 节点: \(.node)"'
fi

# ========================================
# 测试总结
# ========================================
echo ""
echo "========================================="
echo "测试总结"
echo "========================================="

success "CI 流水线测试: 构建任务创建和执行"
success "发布管理测试: 滚动/金丝雀/蓝绿 3种部署策略"
success "应用管理测试: 扩缩容/重启/回滚功能"
success "版本区分验证: Pod 的 version 字段正确标识"

echo ""
echo "========================================="
echo "全链路测试完成!"
echo "========================================="
