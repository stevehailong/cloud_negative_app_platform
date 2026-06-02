<template>
  <div class="app-deployment-detail">
    <!-- 页面头部 -->
    <div class="page-header">
      <el-button @click="goBack" link>
        <el-icon><ArrowLeft /></el-icon>
        返回列表
      </el-button>
      <h2>部署详情</h2>
    </div>

    <!-- 基本信息卡片 -->
    <el-card class="info-card" shadow="never" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>基本信息</span>
          <el-button type="primary" size="small" @click="handleRefresh">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-descriptions :column="3" border v-if="deployment">
        <el-descriptions-item label="应用ID">{{ deployment.app_id }}</el-descriptions-item>
        <el-descriptions-item label="环境ID">{{ deployment.env_id }}</el-descriptions-item>
        <el-descriptions-item label="集群ID">{{ deployment.cluster_id }}</el-descriptions-item>
        <el-descriptions-item label="命名空间">{{ deployment.namespace }}</el-descriptions-item>
        <el-descriptions-item label="工作负载">{{ deployment.workload_name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ deployment.workload_type }}</el-descriptions-item>
        <el-descriptions-item label="当前版本" :span="2">{{ deployment.current_version }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(deployment.deployment_status)" size="small">
            {{ getStatusText(deployment.deployment_status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="当前镜像" :span="3">
          <code style="font-size: 12px;">{{ deployment.current_image }}</code>
        </el-descriptions-item>
        <el-descriptions-item label="期望副本数">{{ deployment.desired_replicas }}</el-descriptions-item>
        <el-descriptions-item label="可用副本数">
          <span 
            :class="{ 
              'text-success': deployment.available_replicas === deployment.desired_replicas,
              'text-warning': deployment.available_replicas < deployment.desired_replicas && deployment.available_replicas > 0,
              'text-danger': deployment.available_replicas === 0
            }"
          >
            {{ deployment.available_replicas }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="副本比例">
          <el-progress 
            :percentage="getReplicaPercentage(deployment)" 
            :status="getReplicaStatus(deployment)"
            :stroke-width="20"
            :text-inside="true"
          />
        </el-descriptions-item>
        <el-descriptions-item label="最后部署时间" :span="2">
          {{ formatTime(deployment.last_deploy_time) }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatTime(deployment.create_time) }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- 操作按钮 -->
      <div class="action-buttons">
        <el-button type="primary" @click="handleRestart">
          <el-icon><RefreshRight /></el-icon>
          重启
        </el-button>
        <el-button type="success" @click="handleScale">
          <el-icon><Operation /></el-icon>
          扩缩容
        </el-button>
        <el-button type="warning" @click="activeTab = 'history'">
          <el-icon><Back /></el-icon>
          回滚
        </el-button>
        <el-button type="info" @click="handleDeploy">
          <el-icon><Upload /></el-icon>
          部署新版本
        </el-button>
      </div>
    </el-card>

    <!-- Tab页 -->
    <el-card class="tab-card" shadow="never">
      <el-tabs v-model="activeTab">
        <!-- 部署历史 -->
        <el-tab-pane label="部署历史" name="history">
          <el-table
            v-loading="historyLoading"
            :data="historyList"
            style="width: 100%"
          >
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="deployment_type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag :type="getDeploymentTypeTag(row.deployment_type)" size="small">
                  {{ getDeploymentTypeText(row.deployment_type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="version" label="版本" width="150" />
            <el-table-column prop="image_url" label="镜像" min-width="250" show-overflow-tooltip />
            <el-table-column prop="replicas" label="副本数" width="100" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getHistoryStatusType(row.status)" size="small">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="耗时" width="100">
              <template #default="{ row }">
                {{ row.duration ? row.duration + 's' : '-' }}
              </template>
            </el-table-column>
            <el-table-column label="开始时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.start_time) }}
              </template>
            </el-table-column>
            <el-table-column label="结束时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.end_time) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button 
                  v-if="row.status === 'success' && row.deployment_type !== 'rollback'"
                  type="primary" 
                  link 
                  size="small" 
                  @click="handleRollbackToHistory(row)"
                >
                  回滚到此版本
                </el-button>
                <el-button 
                  v-if="row.failure_reason"
                  type="danger" 
                  link 
                  size="small" 
                  @click="showFailureReason(row)"
                >
                  查看错误
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 历史记录分页 -->
          <div class="pagination-container">
            <el-pagination
              v-model:current-page="historyPage"
              v-model:page-size="historyPageSize"
              :page-sizes="[10, 20, 50]"
              :total="historyTotal"
              layout="total, sizes, prev, pager, next"
              @size-change="fetchHistory"
              @current-change="fetchHistory"
            />
          </div>
        </el-tab-pane>

        <!-- Pod列表 -->
        <el-tab-pane label="Pod列表" name="pods">
          <el-table
            v-loading="podsLoading"
            :data="podsList"
            style="width: 100%"
          >
            <el-table-column prop="name" label="Pod名称" min-width="300" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getPodStatusType(row.status)" size="small">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="ready" label="就绪" width="80">
              <template #default="{ row }">
                <el-icon v-if="row.ready" color="#67c23a"><CircleCheck /></el-icon>
                <el-icon v-else color="#f56c6c"><CircleClose /></el-icon>
              </template>
            </el-table-column>
            <el-table-column prop="restarts" label="重启次数" width="100" />
            <el-table-column prop="node" label="节点" width="150" show-overflow-tooltip />
            <el-table-column prop="pod_ip" label="Pod IP" width="140" />
            <el-table-column label="启动时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.start_time) }}
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 事件 -->
        <el-tab-pane label="事件" name="events">
          <el-table
            v-loading="eventsLoading"
            :data="eventsList"
            style="width: 100%"
          >
            <el-table-column prop="type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag :type="getEventTypeTag(row.type)" size="small">
                  {{ row.type }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" width="150" />
            <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
            <el-table-column prop="count" label="次数" width="80" />
            <el-table-column prop="source" label="来源" width="150" />
            <el-table-column label="首次时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.first_timestamp) }}
              </template>
            </el-table-column>
            <el-table-column label="最后时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.last_timestamp) }}
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 扩缩容对话框（复用列表页逻辑）-->
    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="500px">
      <el-form :model="scaleForm" label-width="100px">
        <el-form-item label="当前副本数">
          <span>{{ deployment?.desired_replicas }}</span>
        </el-form-item>
        <el-form-item label="目标副本数" required>
          <el-input-number
            v-model="scaleForm.replicas"
            :min="0"
            :max="100"
            style="width: 200px"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmScale" :loading="scaleLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 部署新版本对话框 -->
    <el-dialog v-model="deployDialogVisible" title="部署新版本" width="600px">
      <el-form :model="deployForm" label-width="100px">
        <el-form-item label="当前版本">
          <span>{{ deployment?.current_version }}</span>
        </el-form-item>
        <el-form-item label="新版本号" required>
          <el-input
            v-model="deployForm.version"
            placeholder="例如: v1.0.5"
            style="width: 300px"
          />
        </el-form-item>
        <el-form-item label="镜像地址" required>
          <el-input
            v-model="deployForm.image_url"
            placeholder="例如: nginx:1.26-alpine"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="deployDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmDeploy" :loading="deployLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ArrowLeft, 
  Refresh, 
  RefreshRight, 
  Operation, 
  Back, 
  Upload,
  CircleCheck,
  CircleClose
} from '@element-plus/icons-vue'
import {
  getAppDeploymentDetail,
  getDeploymentHistory,
  getAppDeploymentPods,
  getAppDeploymentEvents,
  restartDeployment,
  scaleDeployment,
  rollbackDeployment,
  deployNewVersion
} from '@/api/deployment'
import { formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()

const deploymentId = ref(route.params.id)
const loading = ref(false)
const deployment = ref(null)

// Tab
const activeTab = ref(route.query.tab || 'history')

// 部署历史
const historyLoading = ref(false)
const historyList = ref([])
const historyPage = ref(1)
const historyPageSize = ref(10)
const historyTotal = ref(0)

// Pod列表
const podsLoading = ref(false)
const podsList = ref([])

// 事件列表
const eventsLoading = ref(false)
const eventsList = ref([])

// 扩缩容对话框
const scaleDialogVisible = ref(false)
const scaleLoading = ref(false)
const scaleForm = reactive({
  replicas: 1
})

// 部署对话框
const deployDialogVisible = ref(false)
const deployLoading = ref(false)
const deployForm = reactive({
  version: '',
  image_url: ''
})

// 获取部署详情
const fetchDeployment = async () => {
  loading.value = true
  try {
    const response = await getAppDeploymentDetail(deploymentId.value)
    if (response.code === 200) {
      deployment.value = response.data
    } else {
      ElMessage.error(response.message || '获取部署详情失败')
    }
  } catch (error) {
    console.error('获取部署详情失败:', error)
    ElMessage.error('获取部署详情失败')
  } finally {
    loading.value = false
  }
}

// 获取部署历史
const fetchHistory = async () => {
  historyLoading.value = true
  try {
    const response = await getDeploymentHistory(deploymentId.value, {
      page: historyPage.value,
      page_size: historyPageSize.value
    })
    if (response.code === 200) {
      historyList.value = response.data.list || []
      historyTotal.value = response.data.total || 0
    } else {
      ElMessage.error(response.message || '获取部署历史失败')
    }
  } catch (error) {
    console.error('获取部署历史失败:', error)
    ElMessage.error('获取部署历史失败')
  } finally {
    historyLoading.value = false
  }
}

// 获取Pod列表
const fetchPods = async () => {
  if (!deployment.value) {
    console.log('[Pod列表] deployment为空,跳过')
    return
  }
  console.log('[Pod列表] deployment.value:', deployment.value)
  podsLoading.value = true
  try {
    console.log('[Pod列表] 准备调用getAppDeploymentPods API, deploymentId=', deploymentId.value)
    const response = await getAppDeploymentPods(deploymentId.value)
    console.log('[Pod列表] getAppDeploymentPods API返回:', response)
    console.log('[Pod列表] response.code=', response.code, ', response.data=', response.data)
    
    if (response.code === 0 || response.code === 200) {
      console.log('[Pod列表] 设置podsList, 数据:', response.data)
      podsList.value = response.data || []
      console.log('[Pod列表] podsList.value.length=', podsList.value.length)
      if (podsList.value.length > 0) {
        ElMessage.success(`成功获取到 ${podsList.value.length} 个Pod`)
      } else {
        ElMessage.info('当前没有Pod')
      }
    } else {
      console.error('[Pod列表] API返回失败, code=', response.code)
      ElMessage.error(response.message || '获取Pod列表失败')
    }
  } catch (error) {
    console.error('[Pod列表] 发生错误:', error)
    console.error('[Pod列表] 错误堆栈:', error.stack)
    ElMessage.error('获取Pod列表失败: ' + error.message)
  } finally {
    podsLoading.value = false
  }
}

// 获取事件列表
const fetchEvents = async () => {
  if (!deployment.value) {
    return
  }
  eventsLoading.value = true
  try {
    console.log('[事件] 准备调用getAppDeploymentEvents API, deploymentId=', deploymentId.value)
    const response = await getAppDeploymentEvents(deploymentId.value)
    console.log('[事件] getAppDeploymentEvents API返回:', response)
    
    if (response.code === 0 || response.code === 200) {
      eventsList.value = response.data || []
      if (eventsList.value.length > 0) {
        ElMessage.success(`成功获取到 ${eventsList.value.length} 个事件`)
      } else {
        ElMessage.info('当前没有事件')
      }
    } else {
      ElMessage.error(response.message || '获取事件列表失败')
    }
  } catch (error) {
    console.error('[事件] 发生错误:', error)
    ElMessage.error('获取事件列表失败: ' + error.message)
  } finally {
    eventsLoading.value = false
  }
}

// 刷新
const handleRefresh = () => {
  fetchDeployment()
  if (activeTab.value === 'history') {
    fetchHistory()
  } else if (activeTab.value === 'pods') {
    fetchPods()
  } else if (activeTab.value === 'events') {
    fetchEvents()
  }
}

// 返回
const goBack = () => {
  router.push({ name: 'app-deployments' })
}

// 重启
const handleRestart = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要重启部署 "${deployment.value.workload_name}" 吗？`,
      '确认重启',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await restartDeployment(deploymentId.value, {
      user_id: 1
    })

    if (response.code === 200) {
      ElMessage.success('重启任务已提交')
      setTimeout(() => {
        fetchDeployment()
        fetchHistory()
      }, 2000)
    } else {
      ElMessage.error(response.message || '重启失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重启失败:', error)
      ElMessage.error('重启失败')
    }
  }
}

// 扩缩容
const handleScale = () => {
  scaleForm.replicas = deployment.value.desired_replicas
  scaleDialogVisible.value = true
}

// 确认扩缩容
const confirmScale = async () => {
  if (scaleForm.replicas === deployment.value.desired_replicas) {
    ElMessage.warning('副本数未改变')
    return
  }

  scaleLoading.value = true
  try {
    const response = await scaleDeployment(deploymentId.value, {
      replicas: scaleForm.replicas,
      user_id: 1
    })

    if (response.code === 200) {
      ElMessage.success('扩缩容任务已提交')
      scaleDialogVisible.value = false
      setTimeout(() => {
        fetchDeployment()
        fetchHistory()
      }, 2000)
    } else {
      ElMessage.error(response.message || '扩缩容失败')
    }
  } catch (error) {
    console.error('扩缩容失败:', error)
    ElMessage.error('扩缩容失败')
  } finally {
    scaleLoading.value = false
  }
}

// 部署新版本
const handleDeploy = () => {
  deployForm.version = ''
  deployForm.image_url = ''
  deployDialogVisible.value = true
}

// 确认部署
const confirmDeploy = async () => {
  if (!deployForm.version || !deployForm.image_url) {
    ElMessage.warning('请填写版本号和镜像地址')
    return
  }

  deployLoading.value = true
  try {
    const response = await deployNewVersion(deploymentId.value, {
      version: deployForm.version,
      image_url: deployForm.image_url,
      user_id: 1
    })

    if (response.code === 200) {
      ElMessage.success('部署任务已提交')
      deployDialogVisible.value = false
      setTimeout(() => {
        fetchDeployment()
        fetchHistory()
      }, 2000)
    } else {
      ElMessage.error(response.message || '部署失败')
    }
  } catch (error) {
    console.error('部署失败:', error)
    ElMessage.error('部署失败')
  } finally {
    deployLoading.value = false
  }
}

// 回滚到历史版本
const handleRollbackToHistory = async (history) => {
  try {
    await ElMessageBox.confirm(
      `确定要回滚到版本 "${history.version}" 吗？`,
      '确认回滚',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await rollbackDeployment(deploymentId.value, {
      history_id: history.id,
      user_id: 1
    })

    if (response.code === 200) {
      ElMessage.success('回滚任务已提交')
      setTimeout(() => {
        fetchDeployment()
        fetchHistory()
      }, 2000)
    } else {
      ElMessage.error(response.message || '回滚失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('回滚失败:', error)
      ElMessage.error('回滚失败')
    }
  }
}

// 显示失败原因
const showFailureReason = (history) => {
  ElMessageBox.alert(history.failure_reason, '失败原因', {
    confirmButtonText: '确定',
    type: 'error'
  })
}

// 副本百分比
const getReplicaPercentage = (deployment) => {
  if (!deployment || deployment.desired_replicas === 0) return 0
  return Math.round((deployment.available_replicas / deployment.desired_replicas) * 100)
}

// 副本状态
const getReplicaStatus = (deployment) => {
  if (!deployment) return ''
  if (deployment.available_replicas === deployment.desired_replicas) return 'success'
  if (deployment.available_replicas === 0) return 'exception'
  return 'warning'
}

// 状态类型
const getStatusType = (status) => {
  const statusMap = {
    running: 'success',
    stopped: 'info',
    failed: 'danger',
    progressing: 'warning'
  }
  return statusMap[status] || 'info'
}

// 状态文本
const getStatusText = (status) => {
  const textMap = {
    running: '运行中',
    stopped: '已停止',
    failed: '失败',
    progressing: '进行中'
  }
  return textMap[status] || status
}

// 部署类型标签
const getDeploymentTypeTag = (type) => {
  const typeMap = {
    create: 'success',
    update: 'primary',
    rollback: 'warning',
    restart: 'info',
    scale: 'info'
  }
  return typeMap[type] || 'info'
}

// 部署类型文本
const getDeploymentTypeText = (type) => {
  const textMap = {
    create: '创建',
    update: '更新',
    rollback: '回滚',
    restart: '重启',
    scale: '扩缩容'
  }
  return textMap[type] || type
}

// 历史状态类型
const getHistoryStatusType = (status) => {
  const statusMap = {
    success: 'success',
    failed: 'danger',
    progressing: 'warning'
  }
  return statusMap[status] || 'info'
}

// Pod状态类型
const getPodStatusType = (status) => {
  const statusMap = {
    Running: 'success',
    Pending: 'warning',
    Failed: 'danger',
    Succeeded: 'info',
    Unknown: 'info'
  }
  return statusMap[status] || 'info'
}

// 事件类型标签
const getEventTypeTag = (type) => {
  return type === 'Normal' ? 'success' : 'warning'
}

// 监听tab变化
watch(activeTab, async (newTab) => {
  console.log('[Tab切换] activeTab变为:', newTab)
  
  // 确保deployment数据已加载
  if (!deployment.value && newTab !== 'history') {
    console.log('[Tab切换] deployment未加载,等待加载完成...')
    // 等待deployment加载完成(最多等待5秒)
    let retries = 50
    while (!deployment.value && retries > 0) {
      await new Promise(resolve => setTimeout(resolve, 100))
      retries--
    }
    if (!deployment.value) {
      console.error('[Tab切换] deployment加载超时')
      ElMessage.error('部署信息加载失败')
      return
    }
    console.log('[Tab切换] deployment加载完成')
  }
  
  if (newTab === 'history' && historyList.value.length === 0) {
    fetchHistory()
  } else if (newTab === 'pods' && podsList.value.length === 0) {
    console.log('[Tab切换] 准备调用fetchPods, podsList.length=', podsList.value.length)
    fetchPods()
  } else if (newTab === 'events' && eventsList.value.length === 0) {
    fetchEvents()
  } else {
    console.log('[Tab切换] 不调用fetch,原因: newTab=', newTab, ', podsList.length=', podsList.value.length, ', eventsList.length=', eventsList.value.length)
  }
})

// 页面加载
onMounted(() => {
  fetchDeployment()
  if (activeTab.value === 'history') {
    fetchHistory()
  } else if (activeTab.value === 'pods') {
    fetchPods()
  } else if (activeTab.value === 'events') {
    fetchEvents()
  }
})
</script>

<style scoped>
.app-deployment-detail {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.info-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-buttons {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}

.tab-card {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.text-success {
  color: #67c23a;
  font-weight: bold;
}

.text-warning {
  color: #e6a23c;
  font-weight: bold;
}

.text-danger {
  color: #f56c6c;
  font-weight: bold;
}
</style>
