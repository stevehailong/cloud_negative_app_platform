<template>
  <div class="release-container">
    <div class="page-header">
      <div class="header-left">
        <h2>发布管理</h2>
        <span class="subtitle">发布工单审批与金丝雀部署</span>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showCreateDialog">创建发布</el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <el-card class="filter-card" shadow="never">
      <el-form :inline="true">
        <el-form-item label="应用">
          <el-select v-model="queryParams.appId" placeholder="全部应用" clearable filterable style="width: 180px" @change="fetchList">
            <el-option v-for="app in appList" :key="app.id" :label="app.name" :value="app.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.releaseStatus" placeholder="全部" clearable style="width: 150px" @change="fetchList">
            <el-option label="待审批" value="submitted" />
            <el-option label="已审批" value="approved" />
            <el-option label="执行中" value="executing" />
            <el-option label="金丝雀中" value="canary" />
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="已回滚" value="rollback" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">查询</el-button>
          <el-button @click="queryParams.appId=null; queryParams.releaseStatus=''; fetchList()">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 发布工单列表 -->
    <el-card class="table-card" shadow="never">
      <el-table v-loading="loading" :data="releaseList" style="width: 100%">
        <el-table-column prop="releaseNo" label="发布编号" width="200" />
        <el-table-column label="应用" width="140">
          <template #default="{ row }">
            {{ getAppName(row.appId) }}
          </template>
        </el-table-column>
        <el-table-column prop="releaseVersion" label="版本" width="150" />
        <el-table-column prop="releaseStrategy" label="发布策略" width="100">
          <template #default="{ row }">
            <el-tag :type="row.releaseStrategy === 'canary' ? 'warning' : 'primary'" size="small">
              {{ strategyLabel(row.releaseStrategy) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="releaseStatus" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.releaseStatus)" size="small">
              {{ statusLabel(row.releaseStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金丝雀比例" width="100">
          <template #default="{ row }">
            <span v-if="row.releaseStrategy === 'canary' && row.canaryStatus === 'canary_running'">
              <el-progress :percentage="row.canaryPercent || 20" :stroke-width="16" :text-inside="true"
                :color="row.canaryPercent > 50 ? '#e6a23c' : '#409eff'" />
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="canaryStatus" label="金丝雀状态" width="150">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; gap: 4px; flex-wrap: wrap;">
              <el-tag v-if="row.canaryStatus" :type="canaryType(row.canaryStatus)" size="small">
                {{ canaryLabel(row.canaryStatus) }}
              </el-tag>
              <el-tag v-if="row.canaryRoutingMode && row.canaryStatus === 'canary_running'" type="info" size="small" effect="plain">
                {{ routingModeLabel(row.canaryRoutingMode) }}
              </el-tag>
            </div>
            <span v-if="!row.canaryStatus">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="operatorName" label="操作人" width="100">
          <template #default="{ row }">
            {{ row.operatorName || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="380" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.releaseStatus === 'created'" type="primary" link size="small" @click="handleEdit(row)">编辑策略</el-button>
            <el-button v-if="row.releaseStatus === 'created'" type="success" link size="small" @click="handleSubmit(row)">提交审批</el-button>
            <el-button v-if="row.releaseStatus === 'submitted'" type="success" link size="small" @click="handleApprove(row)">审批通过</el-button>
            <el-button v-if="row.releaseStatus === 'submitted'" type="danger" link size="small" @click="handleReject(row)">拒绝</el-button>
            <el-button v-if="row.releaseStatus === 'approved'" type="warning" link size="small" @click="handleExecute(row)">执行发布</el-button>
            <el-button v-if="row.releaseStatus === 'canary' && row.canaryStatus === 'canary_running'" type="success" link size="small" @click="handleConfirmCanary(row)">确认全量</el-button>
            <el-button v-if="row.releaseStatus === 'canary' && row.canaryStatus === 'canary_running' && row.canaryRoutingMode !== 'header' && row.canaryRoutingMode !== 'cookie'" type="warning" link size="small" @click="handleAdjustWeight(row)">调整权重</el-button>
            <el-button v-if="row.releaseStatus === 'canary' && row.canaryStatus === 'canary_running'" type="danger" link size="small" @click="handleRollbackCanary(row)">回滚金丝雀</el-button>
            <el-button v-if="row.releaseStatus === 'success'" type="danger" link size="small" @click="handleRollback(row)">回滚</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 创建发布对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建发布工单"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="120px">
        <el-form-item label="应用" prop="appId">
          <el-select 
            v-model="createForm.appId" 
            placeholder="请选择应用" 
            filterable 
            style="width: 100%"
            @change="handleAppChange"
          >
            <el-option
              v-for="app in appList"
              :key="app.id"
              :label="app.name"
              :value="app.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="环境" prop="envId">
          <el-select 
            v-model="createForm.envId" 
            placeholder="请先选择应用" 
            style="width: 100%"
            :disabled="!createForm.appId || envListLoading"
            :loading="envListLoading"
          >
            <el-option
              v-for="env in boundEnvironments"
              :key="env.envId"
              :label="`${env.envName} (${env.envType}) - ${env.namespace}`"
              :value="env.envId"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ env.envName }}</span>
                <div>
                  <el-tag size="small" style="margin-right: 5px;">{{ env.envType }}</el-tag>
                  <el-tag size="small" type="info">{{ env.clusterName }}</el-tag>
                </div>
              </div>
            </el-option>
          </el-select>
          <div v-if="createForm.appId && boundEnvironments.length === 0 && !envListLoading" style="color: #f56c6c; font-size: 12px; margin-top: 4px;">
            该应用还未绑定任何环境，请先在应用详情页绑定环境
          </div>
        </el-form-item>

        <el-form-item label="版本号" prop="releaseVersion">
          <el-input v-model="createForm.releaseVersion" placeholder="例如: v1.0.0" />
        </el-form-item>

        <el-form-item label="镜像地址" prop="imageUrl">
          <el-input v-model="createForm.imageUrl" placeholder="例如: nginx:1.26-alpine" />
        </el-form-item>

        <el-form-item label="发布策略" prop="releaseStrategy">
          <el-radio-group v-model="createForm.releaseStrategy">
            <el-radio value="rolling">滚动发布</el-radio>
            <el-radio value="canary">金丝雀发布</el-radio>
            <el-radio value="bluegreen">蓝绿发布</el-radio>
          </el-radio-group>
          <div class="strategy-desc">
            <div v-if="createForm.releaseStrategy === 'rolling'" class="desc-text">
              <el-icon><InfoFilled /></el-icon>
              滚动发布: 逐步替换旧版本Pod,平滑升级,适合大多数场景
            </div>
            <div v-else-if="createForm.releaseStrategy === 'canary'" class="desc-text">
              <el-icon><InfoFilled /></el-icon>
              金丝雀发布: 先发布小比例流量验证,确认无误后再全量发布,适合风险较高的版本
            </div>
            <div v-else-if="createForm.releaseStrategy === 'bluegreen'" class="desc-text">
              <el-icon><InfoFilled /></el-icon>
              蓝绿发布: 部署新版本后一键切换流量,支持快速回滚,适合对稳定性要求极高的场景
            </div>
          </div>
        </el-form-item>

        <el-form-item
          v-if="createForm.releaseStrategy === 'canary'"
          label="金丝雀比例"
          prop="canaryPercent"
        >
          <div>
            <div style="display: flex; align-items: center; gap: 15px;">
              <el-slider
                v-model="createForm.canaryPercent"
                :min="5"
                :max="50"
                :step="5"
                :marks="{ 5: '5%', 10: '10%', 20: '20%', 30: '30%', 50: '50%' }"
                style="flex: 1;"
              />
              <el-input-number
                v-model="createForm.canaryPercent"
                :min="5"
                :max="50"
                :step="5"
                style="width: 100px"
              />
              <span>%</span>
            </div>
            <div class="help-text">
              建议金丝雀比例: 5%-20%,先小流量验证,确认无误后再全量发布
            </div>
          </div>
        </el-form-item>

        <!-- 金丝雀路由模式（Ingress 分流） -->
        <template v-if="createForm.releaseStrategy === 'canary'">
          <el-form-item label="路由模式" prop="canaryRoutingMode">
            <el-radio-group v-model="createForm.canaryRoutingMode">
              <el-radio value="istio">Istio 服务网格</el-radio>
              <el-radio value="weight">权重分流 (Ingress)</el-radio>
              <el-radio value="header">Header 路由 (Ingress)</el-radio>
              <el-radio value="cookie">Cookie 路由 (Ingress)</el-radio>
              <el-radio value="weight_header">权重+Header (Ingress)</el-radio>
            </el-radio-group>
            <div class="help-text" style="margin-top: 4px;">
              <template v-if="createForm.canaryRoutingMode === 'istio'">
                基于 Istio VirtualService + DestinationRule 实现精确流量控制（需集群已安装 Istio）
              </template>
              <template v-else>
                基于 Nginx Ingress 注解实现精确流量控制（需 Ingress 存在）
              </template>
            </div>
          </el-form-item>

          <template v-if="createForm.canaryRoutingMode.includes('header')">
            <el-form-item label="Header 名称" prop="canaryHeaderName">
              <el-input v-model="createForm.canaryHeaderName" placeholder="例如: x-version" />
            </el-form-item>
            <el-form-item label="Header 值" prop="canaryHeaderValue">
              <el-input v-model="createForm.canaryHeaderValue" placeholder="例如: canary" />
            </el-form-item>
          </template>

          <template v-if="createForm.canaryRoutingMode.includes('cookie')">
            <el-form-item label="Cookie 名称" prop="canaryCookieName">
              <el-input v-model="createForm.canaryCookieName" placeholder="例如: canary" />
            </el-form-item>
          </template>
        </template>

        <el-form-item label="描述">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入发布说明(可选)"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑发布策略对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑发布策略"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="120px">
        <el-form-item label="发布策略" prop="releaseStrategy">
          <el-radio-group v-model="editForm.releaseStrategy">
            <el-radio value="rolling">滚动发布</el-radio>
            <el-radio value="canary">金丝雀发布</el-radio>
            <el-radio value="bluegreen">蓝绿发布</el-radio>
          </el-radio-group>
          <div class="strategy-desc">
            <div v-if="editForm.releaseStrategy === 'rolling'" class="desc-text">
              <el-icon><InfoFilled /></el-icon>
              滚动发布: 逐步替换旧版本Pod,平滑升级,适合大多数场景
            </div>
            <div v-else-if="editForm.releaseStrategy === 'canary'" class="desc-text">
              <el-icon><InfoFilled /></el-icon>
              金丝雀发布: 先发布小比例流量验证,确认无误后再全量发布,适合风险较高的版本
            </div>
            <div v-else-if="editForm.releaseStrategy === 'bluegreen'" class="desc-text">
              <el-icon><InfoFilled /></el-icon>
              蓝绿发布: 部署新版本后一键切换流量,支持快速回滚,适合对稳定性要求极高的场景
            </div>
          </div>
        </el-form-item>

        <el-form-item
          v-if="editForm.releaseStrategy === 'canary'"
          label="金丝雀比例"
          prop="canaryPercent"
        >
          <div>
            <div style="display: flex; align-items: center; gap: 15px;">
              <el-slider
                v-model="editForm.canaryPercent"
                :min="5"
                :max="50"
                :step="5"
                :marks="{ 5: '5%', 10: '10%', 20: '20%', 30: '30%', 50: '50%' }"
                style="flex: 1;"
              />
              <el-input-number
                v-model="editForm.canaryPercent"
                :min="5"
                :max="50"
                :step="5"
                style="width: 100px"
              />
              <span>%</span>
            </div>
            <div class="help-text">
              建议金丝雀比例: 5%-20%,先小流量验证,确认无误后再全量发布
            </div>
          </div>
        </el-form-item>

        <!-- 金丝雀路由模式（编辑时也展示） -->
        <el-form-item v-if="editForm.releaseStrategy === 'canary'" label="路由模式">
          <el-radio-group v-model="editForm.canaryRoutingMode">
            <el-radio value="istio">Istio 服务网格</el-radio>
            <el-radio value="weight">权重分流 (Ingress)</el-radio>
            <el-radio value="header">Header 路由 (Ingress)</el-radio>
            <el-radio value="cookie">Cookie 路由 (Ingress)</el-radio>
            <el-radio value="weight_header">权重+Header (Ingress)</el-radio>
          </el-radio-group>
          <div class="help-text" style="margin-top: 4px;">
            <span v-if="editForm.canaryRoutingMode === 'istio'">
              基于 Istio VirtualService + DestinationRule 实现精确流量控制（需集群已安装 Istio）
            </span>
            <span v-else>
              基于 Nginx Ingress 注解实现精确流量控制（需 Ingress 存在）
            </span>
          </div>
        </el-form-item>

        <el-form-item v-if="editForm.releaseStrategy === 'canary' && editForm.canaryRoutingMode.includes('header')" label="Header 名称">
          <el-input v-model="editForm.canaryHeaderName" placeholder="例如: x-version" />
        </el-form-item>
        <el-form-item v-if="editForm.releaseStrategy === 'canary' && editForm.canaryRoutingMode.includes('header')" label="Header 值">
          <el-input v-model="editForm.canaryHeaderValue" placeholder="例如: canary" />
        </el-form-item>

        <el-form-item v-if="editForm.releaseStrategy === 'canary' && editForm.canaryRoutingMode.includes('cookie')" label="Cookie 名称">
          <el-input v-model="editForm.canaryCookieName" placeholder="例如: canary" />
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="editForm.description"
            type="textarea"
            :rows="3"
            placeholder="可选择修改发布说明"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="handleUpdate">保存</el-button>
      </template>
    </el-dialog>

    <!-- 调整金丝雀权重对话框 -->
    <el-dialog
      v-model="weightDialogVisible"
      title="调整金丝雀权重"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="weightForm" label-width="100px">
        <el-form-item label="发布编号">
          <span>{{ weightForm.releaseNo }}</span>
        </el-form-item>
        <el-form-item label="当前权重">
          <el-tag type="warning">{{ weightForm.currentPercent }}%</el-tag>
        </el-form-item>
        <el-form-item label="新权重">
          <div style="display: flex; align-items: center; gap: 15px; width: 100%;">
            <el-slider
              v-model="weightForm.newPercent"
              :min="0"
              :max="100"
              :step="5"
              :marks="{ 0: '0%', 25: '25%', 50: '50%', 75: '75%', 100: '100%' }"
              style="flex: 1;"
            />
            <el-input-number
              v-model="weightForm.newPercent"
              :min="0"
              :max="100"
              :step="5"
              style="width: 100px"
            />
            <span>%</span>
          </div>
          <div class="help-text" style="margin-top: 8px;">
            0% = 暂停金丝雀（全部走 stable），100% = 全量走金丝雀
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="weightDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="weightLoading" @click="confirmWeight">确认调整</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { InfoFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const loading = ref(false)
const releaseList = ref([])
const total = ref(0)

// 创建发布相关
const createDialogVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref(null)
const appList = ref([])
const boundEnvironments = ref([])
const envListLoading = ref(false)

const createForm = reactive({
  appId: null,
  envId: null,
  releaseVersion: '',
  imageUrl: '',
  releaseStrategy: 'rolling',
  canaryPercent: 20,
  canaryRoutingMode: 'weight',
  canaryHeaderName: '',
  canaryHeaderValue: '',
  canaryCookieName: '',
  description: ''
})

const createRules = {
  appId: [{ required: true, message: '请选择应用', trigger: 'change' }],
  envId: [{ required: true, message: '请选择环境', trigger: 'change' }],
  releaseVersion: [{ required: true, message: '请输入版本号', trigger: 'blur' }],
  imageUrl: [{ required: true, message: '请输入镜像地址', trigger: 'blur' }],
  releaseStrategy: [{ required: true, message: '请选择发布策略', trigger: 'change' }],
  canaryPercent: [
    { required: true, message: '请设置金丝雀比例', trigger: 'change' },
    { type: 'number', min: 5, max: 50, message: '金丝雀比例应在5%-50%之间', trigger: 'change' }
  ]
}

// 编辑发布相关
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editFormRef = ref(null)
const currentEditId = ref(null)

const editForm = reactive({
  releaseStrategy: 'rolling',
  canaryPercent: 20,
  canaryRoutingMode: 'weight',
  canaryHeaderName: '',
  canaryHeaderValue: '',
  canaryCookieName: '',
  description: ''
})

const editRules = {
  releaseStrategy: [{ required: true, message: '请选择发布策略', trigger: 'change' }],
  canaryPercent: [
    { required: true, message: '请设置金丝雀比例', trigger: 'change' },
    { type: 'number', min: 5, max: 50, message: '金丝雀比例应在5%-50%之间', trigger: 'change' }
  ]
}

const queryParams = reactive({
  page: 1,
  pageSize: 10,
  appId: null,
  releaseStatus: ''
})

const strategyLabel = (s) => ({ rolling: '滚动', bluegreen: '蓝绿', canary: '金丝雀' }[s] || s)
const routingModeLabel = (m) => ({ istio: 'Istio网格', weight: '权重', header: 'Header', cookie: 'Cookie', weight_header: '权重+Header' }[m] || m)
const statusLabel = (s) => ({
  created: '已创建', submitted: '待审批', approved: '已审批', rejected: '已拒绝',
  executing: '执行中', canary: '金丝雀中', success: '成功', failed: '失败', rollback: '已回滚'
}[s] || s)
const statusType = (s) => ({
  created: 'info', submitted: 'warning', approved: 'primary', rejected: 'danger',
  executing: 'warning', canary: 'warning', success: 'success', failed: 'danger', rollback: 'info'
}[s] || 'info')
const canaryLabel = (s) => ({ canary_running: '灰度中', canary_confirmed: '已全量', canary_rollback: '已回滚' }[s] || s)
const canaryType = (s) => ({ canary_running: 'warning', canary_confirmed: 'success', canary_rollback: 'danger' }[s] || 'info')

// 权重调整
const weightDialogVisible = ref(false)
const weightLoading = ref(false)
const weightForm = reactive({
  releaseId: null,
  releaseNo: '',
  currentPercent: 0,
  newPercent: 0
})

const handleAdjustWeight = (row) => {
  weightForm.releaseId = row.id
  weightForm.releaseNo = row.releaseNo
  weightForm.currentPercent = row.canaryPercent || 20
  weightForm.newPercent = row.canaryPercent || 20
  weightDialogVisible.value = true
}

const confirmWeight = async () => {
  weightLoading.value = true
  try {
    await request.post(`/releases/${weightForm.releaseId}/canary/adjust-weight`, {
      canaryPercent: weightForm.newPercent
    })
    ElMessage.success(`权重已调整为 ${weightForm.newPercent}%`)
    weightDialogVisible.value = false
    fetchList()
  } catch (error) {
    ElMessage.error('调整权重失败: ' + (error.response?.data?.message || error.message))
  } finally {
    weightLoading.value = false
  }
}

const fetchList = async () => {
  loading.value = true
  try {
    const params = { page: queryParams.page, pageSize: queryParams.pageSize }
    if (queryParams.appId) params.appId = queryParams.appId
    if (queryParams.releaseStatus) params.releaseStatus = queryParams.releaseStatus
    const res = await request.get('/releases', { params })
    releaseList.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const getAppName = (appId) => {
  const app = appList.value.find(a => a.id === appId)
  return app ? app.name : `app-${appId}`
}

// 加载应用列表
const fetchAppList = async () => {
  try {
    const res = await request.get('/applications', { params: { page: 1, page_size: 100 } })
    appList.value = res.data.list || []
  } catch (e) {
    console.error('加载应用列表失败:', e)
  }
}

// 应用选择变化时，加载已绑定的环境
const handleAppChange = async (appId) => {
  // 重置环境选择
  createForm.envId = null
  boundEnvironments.value = []
  
  if (!appId) return
  
  envListLoading.value = true
  try {
    const res = await request.get('/app-env-bindings', {
      params: { applicationId: appId, page: 1, pageSize: 100 }
    })
    
    if (res.code === 0) {
      boundEnvironments.value = res.data.list || []
      
      if (boundEnvironments.value.length === 0) {
        ElMessage.warning('该应用还未绑定任何环境，请先在应用详情页绑定环境')
      }
    } else {
      ElMessage.error(res.message || '加载环境列表失败')
    }
  } catch (error) {
    console.error('加载环境列表失败:', error)
    ElMessage.error('加载环境列表失败')
  } finally {
    envListLoading.value = false
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  // 重置表单
  Object.assign(createForm, {
    appId: null,
    envId: null,
    releaseVersion: '',
    imageUrl: '',
    releaseStrategy: 'rolling',
    canaryPercent: 20,
    canaryRoutingMode: 'weight',
    canaryHeaderName: '',
    canaryHeaderValue: '',
    canaryCookieName: '',
    description: ''
  })
  boundEnvironments.value = []
  createFormRef.value?.clearValidate()
  createDialogVisible.value = true
  // 加载应用列表
  if (appList.value.length === 0) {
    fetchAppList()
  }
}

// 创建发布
const handleCreate = async () => {
  try {
    await createFormRef.value.validate()
    
    createLoading.value = true
    const payload = {
      appId: createForm.appId,
      envId: createForm.envId,
      releaseVersion: createForm.releaseVersion,
      imageUrl: createForm.imageUrl,
      releaseStrategy: createForm.releaseStrategy,
      description: createForm.description
    }
    
    // 只有金丝雀发布需要传 canaryPercent 和路由配置
    if (createForm.releaseStrategy === 'canary') {
      payload.canaryPercent = createForm.canaryPercent
      payload.canaryRoutingMode = createForm.canaryRoutingMode || 'weight'
      if (createForm.canaryRoutingMode.includes('header')) {
        payload.canaryHeaderName = createForm.canaryHeaderName
        payload.canaryHeaderValue = createForm.canaryHeaderValue
      }
      if (createForm.canaryRoutingMode.includes('cookie')) {
        payload.canaryCookieName = createForm.canaryCookieName
      }
    }
    
    await request.post('/releases', payload)
    ElMessage.success('发布工单创建成功')
    createDialogVisible.value = false
    fetchList()
  } catch (error) {
    if (error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    createLoading.value = false
  }
}

// 编辑发布策略
const handleEdit = (row) => {
  currentEditId.value = row.id
  // 填充现有数据（包含路由模式）
  Object.assign(editForm, {
    releaseStrategy: row.releaseStrategy,
    canaryPercent: row.canaryPercent || 20,
    canaryRoutingMode: row.canaryRoutingMode || 'weight',
    canaryHeaderName: row.canaryHeaderName || '',
    canaryHeaderValue: row.canaryHeaderValue || '',
    canaryCookieName: row.canaryCookieName || '',
    description: row.description || ''
  })
  editFormRef.value?.clearValidate()
  editDialogVisible.value = true
}

// 保存编辑
const handleUpdate = async () => {
  try {
    await editFormRef.value.validate()

    editLoading.value = true
    const payload = {
      releaseStrategy: editForm.releaseStrategy,
      canaryPercent: editForm.canaryPercent,
      description: editForm.description
    }
    if (editForm.releaseStrategy === 'canary') {
      payload.canaryRoutingMode = editForm.canaryRoutingMode || 'weight'
      if (editForm.canaryRoutingMode.includes('header')) {
        payload.canaryHeaderName = editForm.canaryHeaderName
        payload.canaryHeaderValue = editForm.canaryHeaderValue
      }
      if (editForm.canaryRoutingMode.includes('cookie')) {
        payload.canaryCookieName = editForm.canaryCookieName
      }
    }
    await request.put(`/releases/${currentEditId.value}`, payload)
    ElMessage.success('发布策略已更新')
    editDialogVisible.value = false
    fetchList()
  } catch (error) {
    if (error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    editLoading.value = false
  }
}

const handleSubmit = async (row) => {
  try {
    await ElMessageBox.confirm('提交审批后将通知审批人，确认提交？', '提交审批')
    await request.post(`/releases/${row.id}/submit`, { approverUserIds: [1] })
    ElMessage.success('已提交审批')
    fetchList()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

const handleApprove = async (row) => {
  try {
    await ElMessageBox.confirm('确认审批通过该发布工单？', '审批通过')
    await request.post(`/releases/${row.id}/approve`, { comment: '审批通过' })
    ElMessage.success('审批通过')
    fetchList()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

const handleReject = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝审批', { inputPlaceholder: '原因' })
    await request.post(`/releases/${row.id}/reject`, { comment: value || '拒绝' })
    ElMessage.success('已拒绝')
    fetchList()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

const handleExecute = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确认执行发布？策略：${strategyLabel(row.releaseStrategy)}`,
      '执行发布',
      { type: 'warning' }
    )
    await request.post(`/releases/${row.id}/execute`)
    ElMessage.success('发布已启动')
    fetchList()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

const handleConfirmCanary = async (row) => {
  try {
    await ElMessageBox.confirm('金丝雀验证通过？确认后将全量发布', '确认全量发布', { type: 'warning' })
    await request.post(`/releases/${row.id}/canary/confirm`)
    ElMessage.success('全量发布中')
    fetchList()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

const handleRollbackCanary = async (row) => {
  try {
    await ElMessageBox.confirm('确认回滚金丝雀部署？', '回滚金丝雀', { type: 'danger' })
    await request.post(`/releases/${row.id}/canary/rollback`)
    ElMessage.success('金丝雀已回滚')
    fetchList()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

const handleRollback = async (row) => {
  try {
    await ElMessageBox.confirm('确认回滚该发布？', '回滚', { type: 'danger' })
    await request.post(`/releases/${row.id}/rollback`)
    ElMessage.success('回滚已启动')
    fetchList()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

onMounted(() => {
  fetchList()
  fetchAppList()
})
</script>

<style scoped lang="scss">
.release-container {
  padding: 20px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  .header-left {
    h2 { margin: 0 0 8px 0; font-size: 24px; font-weight: 500; color: #303133; }
    .subtitle { font-size: 14px; color: #909399; }
  }
}
.filter-card { margin-bottom: 20px; :deep(.el-card__body) { padding: 16px; } }
.table-card { :deep(.el-card__body) { padding: 0; } }
.pagination-container { display: flex; justify-content: flex-end; padding: 20px; }

.strategy-desc {
  margin-top: 8px;
  .desc-text {
    display: flex;
    align-items: start;
    gap: 6px;
    padding: 8px 12px;
    background: #f4f4f5;
    border-radius: 4px;
    font-size: 13px;
    color: #606266;
    line-height: 1.6;
    .el-icon {
      margin-top: 2px;
      color: #409eff;
    }
  }
}

.help-text {
  margin-top: 35px;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}
</style>
