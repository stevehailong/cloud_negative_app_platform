#!/bin/bash
# =============================================================================
# Istio 金丝雀发布功能测试脚本
# =============================================================================
# 使用方法:
#   chmod +x test_istio_canary.sh
#   ./test_istio_canary.sh
#
# 前提条件:
#   1. 集群已安装 Istio (istioctl version)
#   2. 目标命名空间已启用 sidecar 注入 (kubectl label ns xxx istio-injection=enabled)
#   3. 后端服务已启动 (gateway:8080)
#   4. 至少有一个应用绑定到环境
# =============================================================================

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN=""
APP_ID="${APP_ID:-}"
ENV_ID="${ENV_ID:-}"
RELEASE_ID=""

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ---------------------------------------------------------------------------
# Step 0: 环境检查
# ---------------------------------------------------------------------------
log_info "===== Step 0: 环境检查 ====="

# 检查 Istio 是否安装
if command -v istioctl &> /dev/null; then
    log_ok "istioctl 已安装: $(istioctl version --short 2>/dev/null || echo 'unknown')"
elif kubectl get crd virtualservices.networking.istio.io &>/dev/null; then
    log_ok "Istio CRD 已存在于集群中"
else
    log_warn "未检测到 Istio，Istio 分流模式将降级为 Pod 比例分流"
fi

# 检查后端连通性
if curl -s --connect-timeout 3 "${BASE_URL}/health" > /dev/null 2>&1; then
    log_ok "后端服务连通: ${BASE_URL}"
else
    log_error "后端服务不可达: ${BASE_URL}"
    log_info "请先启动后端服务: docker-compose up -d 或 go run ..."
    exit 1
fi

# ---------------------------------------------------------------------------
# Step 1: 登录获取 Token
# ---------------------------------------------------------------------------
log_info "===== Step 1: 登录 ====="

LOGIN_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    log_error "登录失败: $LOGIN_RESP"
    exit 1
fi
log_ok "登录成功，Token: ${TOKEN:0:20}..."

AUTH_HEADER="Authorization: Bearer $TOKEN"

# ---------------------------------------------------------------------------
# Step 2: 查找可用的应用和环境
# ---------------------------------------------------------------------------
log_info "===== Step 2: 查找可用的应用和环境 ====="

if [ -z "$APP_ID" ]; then
    APP_LIST=$(curl -s -X GET "${BASE_URL}/api/v1/applications?page=1&pageSize=10" \
      -H "$AUTH_HEADER")
    APP_ID=$(echo "$APP_LIST" | jq -r '.data.list[0].id // empty')
    APP_NAME=$(echo "$APP_LIST" | jq -r '.data.list[0].name // empty')

    if [ -z "$APP_ID" ]; then
        log_error "没有可用的应用，请先在平台创建应用"
        exit 1
    fi
    log_ok "使用应用: id=${APP_ID}, name=${APP_NAME}"
fi

if [ -z "$ENV_ID" ]; then
    # 查找应用绑定的环境
    ENV_LIST=$(curl -s -X GET "${BASE_URL}/api/v1/app-env-bindings?applicationId=${APP_ID}&page=1&pageSize=10" \
      -H "$AUTH_HEADER")
    ENV_ID=$(echo "$ENV_LIST" | jq -r '.data.list[0].envId // empty')
    ENV_NAME=$(echo "$ENV_LIST" | jq -r '.data.list[0].envName // empty')

    if [ -z "$ENV_ID" ]; then
        log_error "应用 ${APP_ID} 未绑定任何环境，请先在应用详情页绑定环境"
        exit 1
    fi
    log_ok "使用环境: id=${ENV_ID}, name=${ENV_NAME}"
fi

# ---------------------------------------------------------------------------
# Step 3: 创建 Istio 金丝雀发布工单
# ---------------------------------------------------------------------------
log_info "===== Step 3: 创建 Istio 金丝雀发布工单 ====="

CREATE_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/releases" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d "{
    \"appId\": ${APP_ID},
    \"envId\": ${ENV_ID},
    \"releaseVersion\": \"v1.1.0-istio-canary\",
    \"releaseStrategy\": \"canary\",
    \"imageUrl\": \"nginx:1.27-alpine\",
    \"canaryPercent\": 20,
    \"canaryRoutingMode\": \"istio\",
    \"description\": \"Istio 金丝雀发布测试\"
  }")

RELEASE_ID=$(echo "$CREATE_RESP" | jq -r '.data.id // empty')

if [ -z "$RELEASE_ID" ] || [ "$RELEASE_ID" = "null" ]; then
    log_error "创建发布工单失败: $CREATE_RESP"
    exit 1
fi
log_ok "发布工单已创建: id=${RELEASE_ID}"

# ---------------------------------------------------------------------------
# Step 4: 提交审批 → 审批通过 → 执行发布
# ---------------------------------------------------------------------------
log_info "===== Step 4: 提交审批 ====="
curl -s -X POST "${BASE_URL}/api/v1/releases/${RELEASE_ID}/submit" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"approverUserIds":[1]}' | jq .
log_ok "已提交审批"

log_info "===== Step 5: 审批通过 ====="
curl -s -X POST "${BASE_URL}/api/v1/releases/${RELEASE_ID}/approve" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"comment":"审批通过 - Istio 金丝雀测试"}' | jq .
log_ok "审批已通过"

log_info "===== Step 6: 执行发布 (触发 Istio 金丝雀部署) ====="
EXEC_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/releases/${RELEASE_ID}/execute" \
  -H "$AUTH_HEADER")
echo "$EXEC_RESP" | jq .
log_ok "发布已执行，等待金丝雀部署完成..."

# 等待金丝雀部署
sleep 10

# ---------------------------------------------------------------------------
# Step 7: 查看金丝雀状态
# ---------------------------------------------------------------------------
log_info "===== Step 7: 查看金丝雀状态 ====="

RELEASE_DETAIL=$(curl -s -X GET "${BASE_URL}/api/v1/releases/${RELEASE_ID}" \
  -H "$AUTH_HEADER")
echo "$RELEASE_DETAIL" | jq '{
  id: .data.id,
  releaseNo: .data.releaseNo,
  releaseStatus: .data.releaseStatus,
  canaryStatus: .data.canaryStatus,
  canaryPercent: .data.canaryPercent,
  canaryRoutingMode: .data.canaryRoutingMode,
  description: .data.description
}'

CANARY_STATUS=$(echo "$RELEASE_DETAIL" | jq -r '.data.canaryStatus // empty')
RELEASE_STATUS=$(echo "$RELEASE_DETAIL" | jq -r '.data.releaseStatus // empty')

if [ "$RELEASE_STATUS" = "canary" ] && [ "$CANARY_STATUS" = "canary_running" ]; then
    log_ok "金丝雀运行中！状态: ${RELEASE_STATUS}/${CANARY_STATUS}"
else
    log_warn "当前状态: ${RELEASE_STATUS}/${CANARY_STATUS} (可能仍在部署中或降级为 Pod 模式)"
fi

# ---------------------------------------------------------------------------
# Step 8: 检查 Istio 资源 (如果 istioctl 可用)
# ---------------------------------------------------------------------------
log_info "===== Step 8: 检查 Istio 资源 ====="

NAMESPACE="app-${APP_ID}-dev"  # 根据实际的命名空间规则调整

if command -v kubectl &> /dev/null; then
    echo ""
    log_info "--- VirtualService ---"
    kubectl get virtualservice -n "${NAMESPACE}" 2>/dev/null || log_warn "未找到 VirtualService (命名空间可能不同或使用 Helm 部署)"

    echo ""
    log_info "--- DestinationRule ---"
    kubectl get destinationrule -n "${NAMESPACE}" 2>/dev/null || log_warn "未找到 DestinationRule"

    echo ""
    log_info "--- Pod (带 istio-proxy sidecar) ---"
    kubectl get pods -n "${NAMESPACE}" -o wide 2>/dev/null || log_warn "未找到 Pod"

    echo ""
    log_info "--- 检查 sidecar 注入 ---"
    SIDECAR_COUNT=$(kubectl get pods -n "${NAMESPACE}" -o json 2>/dev/null | jq '[.items[] | select(.spec.containers[]?.name == "istio-proxy")] | length' 2>/dev/null || echo "0")
    log_info "包含 istio-proxy sidecar 的 Pod 数量: ${SIDECAR_COUNT}"
fi

# ---------------------------------------------------------------------------
# Step 9: 动态调整权重 (Istio VirtualService patch)
# ---------------------------------------------------------------------------
log_info "===== Step 9: 调整金丝雀权重 20% → 50% ====="

ADJUST_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/releases/${RELEASE_ID}/canary/adjust-weight" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"canaryPercent": 50}')
echo "$ADJUST_RESP" | jq .
log_ok "权重已调整为 50%"

sleep 3

# 验证权重
RELEASE_DETAIL2=$(curl -s -X GET "${BASE_URL}/api/v1/releases/${RELEASE_ID}" \
  -H "$AUTH_HEADER")
NEW_WEIGHT=$(echo "$RELEASE_DETAIL2" | jq -r '.data.canaryPercent')
log_info "当前金丝雀权重: ${NEW_WEIGHT}%"

# ---------------------------------------------------------------------------
# Step 10: 查看发布列表（含真实 Ingress/Istio 权重）
# ---------------------------------------------------------------------------
log_info "===== Step 10: 发布列表 ====="

LIST_RESP=$(curl -s -X GET "${BASE_URL}/api/v1/releases?page=1&pageSize=5&releaseStatus=canary" \
  -H "$AUTH_HEADER")
echo "$LIST_RESP" | jq '.data.list[] | {
  id: .id,
  releaseNo: .releaseNo,
  releaseStatus: .releaseStatus,
  canaryStatus: .canaryStatus,
  canaryPercent: .canaryPercent,
  canaryRoutingMode: .canaryRoutingMode,
  operatorName: .operatorName
}'

# ---------------------------------------------------------------------------
# Step 11: 提示后续操作选项
# ---------------------------------------------------------------------------
echo ""
log_info "=============================================="
log_info "  金丝雀发布已运行，后续操作选项:"
log_info "=============================================="
echo ""
log_info "  ✅ 确认全量发布:"
echo "    curl -X POST ${BASE_URL}/api/v1/releases/${RELEASE_ID}/canary/confirm \\"
echo "      -H \"${AUTH_HEADER}\""
echo ""
log_info "  ❌ 回滚金丝雀:"
echo "    curl -X POST ${BASE_URL}/api/v1/releases/${RELEASE_ID}/canary/rollback \\"
echo "      -H \"${AUTH_HEADER}\""
echo ""
log_info "  📊 调整权重 (0-100):"
echo "    curl -X POST ${BASE_URL}/api/v1/releases/${RELEASE_ID}/canary/adjust-weight \\"
echo "      -H \"${AUTH_HEADER}\" \\"
echo "      -H \"Content-Type: application/json\" \\"
echo "      -d '{\"canaryPercent\": 80}'"
echo ""
log_info "  🔍 查看 Istio 资源:"
echo "    kubectl get virtualservice,destinationrule -n ${NAMESPACE}"
echo ""
log_info "  📝 RELEASE_ID=${RELEASE_ID}  APP_ID=${APP_ID}  ENV_ID=${ENV_ID}"
echo ""
