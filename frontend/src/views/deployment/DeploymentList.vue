<template>
  <div class="deployment-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2>部署管理</h2>
        <span class="subtitle">应用部署记录与状态监控</span>
      </div>
    </div>

    <!-- 筛选栏 -->
    <el-card class="filter-card" shadow="never">
      <el-form :inline="true" :model="queryParams">
        <el-form-item label="工作负载名称">
          <el-input
            v-model="queryParams.workloadName"
            placeholder="请输入工作负载名称"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item label="命名空间">
          <el-input
            v-model="queryParams.namespace"
            placeholder="请输入命名空间"
            clearable
            style="width: 150px"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="queryParams.status"
            placeholder="请选择状态"
            clearable
            style="width: 150px"
          >
            <el-option label="运行中" value="running" />
            <el-option label="部署中" value="deploying" />
            <el-option label="失败" value="failed" />
            <el-option label="已停止" value="stopped" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 部署列表 -->
    <el-card class="table-card" shadow="never">
      <el-table
        v-loading="loading"
        :data="deploymentList"
        style="width: 100%"
      >
        <el-table-column prop="workloadName" label="工作负载" min-width="180" />
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="workloadType" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.workloadType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="imageVersion" label="镜像版本" width="150" />
        <el-table-column label="副本数" width="100">
          <template #default="{ row }">
            <span>{{ row.availableReplicas }}/{{ row.desiredReplicas }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="deploymentStatus" label="状态" width="100">
          <template #default="{ row }">
            <el-tag
              :type="getStatusType(row.deploymentStatus)"
              size="small"
            >
              {{ getStatusText(row.deploymentStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleRestart(row)">重启</el-button>
            <el-button type="primary" link size="small" @click="handleScale(row)">扩缩容</el-button>
            <el-button type="primary" link size="small" @click="handleViewPods(row)">Pod</el-button>
            <el-button type="primary" link size="small" @click="handleViewEvents(row)">事件</el-button>
            <el-button type="danger" link size="small" @click="handleRollback(row)">回滚</el-button>
            <el-button v-if="canDelete" type="danger" link size="small" @click="handleDeleteDeployment(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleQuery"
          @current-change="handleQuery"
        />
      </div>
    </el-card>

    <!-- 扩缩容对话框 -->
    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="400px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="当前副本">
          <span>{{ scaleTarget.availableReplicas }} / {{ scaleTarget.desiredReplicas }}</span>
        </el-form-item>
        <el-form-item label="目标副本">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="actionLoading" @click="confirmScale">确认</el-button>
      </template>
    </el-dialog>

    <!-- Pod 列表对话框 -->
    <el-dialog v-model="podsDialogVisible" title="Pod 列表" width="700px" destroy-on-close>
      <el-table :data="podList" v-loading="podsLoading">
        <el-table-column prop="name" label="Pod名称" min-width="200" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'Running' ? 'success' : 'warning'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="nodeName" label="节点" width="150" />
        <el-table-column prop="restarts" label="重启次数" width="80" />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" link size="small" @click="handleDeletePod(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 事件对话框 -->
    <el-dialog v-model="eventsDialogVisible" title="部署事件" width="700px" destroy-on-close>
      <el-table :data="eventList" v-loading="eventsLoading">
        <el-table-column prop="type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">
              {{ row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" width="120" />
        <el-table-column prop="message" label="信息" min-width="200" />
        <el-table-column label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.time) }}
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const userStore = useUserStore()
const canDelete = computed(() => {
  const roles = userStore.userInfo?.roles || []
  return roles.some(r => r.code === 'SUPER_ADMIN' || r.code === 'PROJECT_ADMIN')
})

// 查询参数
const queryParams = reactive({
  page: 1,
  pageSize: 10,
  workloadName: '',
  namespace: '',
  status: ''
})

// 数据
const loading = ref(false)
const deploymentList = ref([])
const total = ref(0)
const actionLoading = ref(false)

// 扩缩容
const scaleDialogVisible = ref(false)
const scaleTarget = ref({})
const scaleReplicas = ref(1)

// Pod列表
const podsDialogVisible = ref(false)
const podsLoading = ref(false)
const podList = ref([])
const currentPodNamespace = ref('')

// 事件
const eventsDialogVisible = ref(false)
const eventsLoading = ref(false)
const eventList = ref([])

// 获取部署列表
const fetchDeploymentList = async () => {
  loading.value = true
  try {
    const { data } = await request({
      url: '/deployments',
      method: 'get',
      params: queryParams
    })
    deploymentList.value = data.list || []
    total.value = data.total || 0
  } catch (error) {
    console.error('获取部署列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 状态类型映射
const getStatusType = (status) => {
  const typeMap = {
    running: 'success',
    success: 'success',
    deploying: 'warning',
    progressing: 'warning',
    failed: 'danger',
    stopped: 'info',
    rollback: 'warning'
  }
  return typeMap[status] || 'info'
}

// 状态文本映射
const getStatusText = (status) => {
  const textMap = {
    running: '运行中',
    success: '成功',
    deploying: '部署中',
    progressing: '部署中',
    failed: '失败',
    stopped: '已停止',
    rollback: '已回滚'
  }
  return textMap[status] || status
}

// 重启
const handleRestart = (row) => {
  ElMessageBox.confirm(
    '确定要重启部署「' + row.workloadName + '」吗？',
    '确认重启',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  ).then(async () => {
    try {
      await request({ url: '/deployments/' + row.id + '/restart', method: 'post' })
      ElMessage.success('重启指令已发送')
      fetchDeploymentList()
    } catch (error) {
      console.error('重启失败:', error)
    }
  }).catch(() => {})
}

// 扩缩容
const handleScale = (row) => {
  scaleTarget.value = row
  scaleReplicas.value = row.desiredReplicas || 1
  scaleDialogVisible.value = true
}

const confirmScale = async () => {
  actionLoading.value = true
  try {
    await request({
      url: '/deployments/' + scaleTarget.value.id + '/scale',
      method: 'post',
      data: { replicas: scaleReplicas.value }
    })
    ElMessage.success('扩缩容指令已发送')
    scaleDialogVisible.value = false
    fetchDeploymentList()
  } catch (error) {
    console.error('扩缩容失败:', error)
  } finally {
    actionLoading.value = false
  }
}

// 查看Pod
const handleViewPods = async (row) => {
  currentPodNamespace.value = row.namespace
  podsDialogVisible.value = true
  podsLoading.value = true
  try {
    const { data } = await request({
      url: '/deployments/' + row.id + '/pods',
      method: 'get'
    })
    podList.value = data.list || data || []
  } catch (error) {
    console.error('获取Pod列表失败:', error)
    podList.value = []
  } finally {
    podsLoading.value = false
  }
}

// 删除部署
const handleDeleteDeployment = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除部署「${row.workloadName}」吗？此操作将同时删除K8s Workload，不可恢复！`, '确认删除', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    await request({ url: `/deployments/${row.id}`, method: 'delete' })
    ElMessage.success('部署已删除')
    fetchDeployments()
  } catch (e) {
    if (e !== 'cancel') console.error('删除部署失败:', e)
  }
}

// 删除Pod
const handleDeletePod = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除Pod「${row.name}」吗？`, '确认删除', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    await request({ url: `/deployments/pods/${row.name}?namespace=${currentPodNamespace.value}`, method: 'delete' })
    ElMessage.success('Pod已删除')
    podList.value = podList.value.filter(p => p.name !== row.name)
  } catch (e) {
    if (e !== 'cancel') console.error('删除Pod失败:', e)
  }
}

// 查看事件
const handleViewEvents = async (row) => {
  eventsDialogVisible.value = true
  eventsLoading.value = true
  try {
    const { data } = await request({
      url: '/deployments/' + row.id + '/events',
      method: 'get'
    })
    eventList.value = data.list || data || []
  } catch (error) {
    console.error('获取事件失败:', error)
    eventList.value = []
  } finally {
    eventsLoading.value = false
  }
}

// 回滚
const handleRollback = (row) => {
  ElMessageBox.confirm(
    '确定要回滚部署「' + row.workloadName + '」到上一个版本吗？',
    '确认回滚',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  ).then(async () => {
    try {
      await request({ url: '/deployments/' + row.id + '/rollback', method: 'post' })
      ElMessage.success('回滚指令已发送')
      fetchDeploymentList()
    } catch (error) {
      console.error('回滚失败:', error)
    }
  }).catch(() => {})
}

// 查询
const handleQuery = () => {
  queryParams.page = 1
  fetchDeploymentList()
}

// 重置
const handleReset = () => {
  queryParams.workloadName = ''
  queryParams.namespace = ''
  queryParams.status = ''
  queryParams.page = 1
  fetchDeploymentList()
}

// 初始化
onMounted(() => {
  fetchDeploymentList()
})
</script>

<style scoped lang="scss">
.deployment-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;

  .header-left {
    h2 {
      margin: 0 0 8px 0;
      font-size: 24px;
      font-weight: 500;
      color: #303133;
    }

    .subtitle {
      font-size: 14px;
      color: #909399;
    }
  }
}

.filter-card {
  margin-bottom: 20px;

  :deep(.el-card__body) {
    padding: 16px;
  }
}

.table-card {
  :deep(.el-card__body) {
    padding: 0;
  }
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  padding: 20px;
}
</style>
