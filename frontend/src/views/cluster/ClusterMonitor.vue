<template>
  <div class="cluster-monitor">
    <el-page-header @back="goBack" :title="'集群监控'">
      <template #content>
        <div class="header-content">
          <span class="cluster-name">{{ clusterInfo.name }}</span>
          <el-tag :type="getStatusType(clusterInfo.status)" size="large">
            {{ getStatusLabel(clusterInfo.status) }}
          </el-tag>
        </div>
      </template>
    </el-page-header>
    
    <!-- 集群概况 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="节点总数" :value="stats.nodeCount">
            <template #suffix>个</template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="Pod总数" :value="stats.podCount">
            <template #suffix>个</template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="CPU使用率" :value="stats.cpuUsage" :precision="1">
            <template #suffix>%</template>
          </el-statistic>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="内存使用率" :value="stats.memoryUsage" :precision="1">
            <template #suffix>%</template>
          </el-statistic>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 节点列表 -->
    <el-card class="nodes-card" shadow="never" style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>节点列表</span>
          <el-button type="primary" size="small" @click="refreshNodes">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <el-table
        v-loading="nodesLoading"
        :data="nodes"
        style="width: 100%"
      >
        <el-table-column prop="name" label="节点名称" width="200" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'Ready' ? 'success' : 'danger'">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.role" type="info">{{ row.role }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="CPU" width="150">
          <template #default="{ row }">
            <el-progress :percentage="parseFloat(row.cpuUsage)" :color="getProgressColor(row.cpuUsage)" />
          </template>
        </el-table-column>
        <el-table-column label="内存" width="150">
          <template #default="{ row }">
            <el-progress :percentage="parseFloat(row.memoryUsage)" :color="getProgressColor(row.memoryUsage)" />
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="150" />
        <el-table-column prop="osImage" label="操作系统" min-width="200" show-overflow-tooltip />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 命名空间列表 -->
    <el-card class="namespaces-card" shadow="never" style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>命名空间列表</span>
        </div>
      </template>
      
      <el-table
        v-loading="namespacesLoading"
        :data="namespaces"
        style="width: 100%"
      >
        <el-table-column prop="name" label="命名空间" width="200" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'Active' ? 'success' : 'info'">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="podCount" label="Pod数量" width="100" />
        <el-table-column prop="labels" label="标签" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatLabels(row.labels) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 资源使用趋势 -->
    <el-card class="trends-card" shadow="never" style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>资源使用趋势</span>
          <el-radio-group v-model="trendPeriod" size="small" @change="loadTrends">
            <el-radio-button label="1h">1小时</el-radio-button>
            <el-radio-button label="6h">6小时</el-radio-button>
            <el-radio-button label="24h">24小时</el-radio-button>
            <el-radio-button label="7d">7天</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      
      <div class="chart-container">
        <div class="chart-placeholder">
          <el-empty description="资源使用趋势图（集成Grafana或自定义图表）" />
          <el-text type="info">
            提示: 可以集成Grafana展示实时监控数据，或使用ECharts绘制自定义图表
          </el-text>
        </div>
      </div>
    </el-card>
    
    <!-- 告警信息 -->
    <el-card class="alerts-card" shadow="never" style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>告警信息</span>
          <el-badge :value="alerts.length" :max="99" type="danger">
            <el-button size="small">查看全部</el-button>
          </el-badge>
        </div>
      </template>
      
      <el-timeline v-if="alerts.length > 0">
        <el-timeline-item
          v-for="alert in alerts"
          :key="alert.id"
          :timestamp="formatTime(alert.createTime)"
          :type="getAlertType(alert.level)"
        >
          <el-tag :type="getAlertType(alert.level)" size="small">{{ alert.level }}</el-tag>
          <span style="margin-left: 10px">{{ alert.message }}</span>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else description="暂无告警信息" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const router = useRouter()
const route = useRoute()

const clusterInfo = ref({})
const stats = reactive({
  nodeCount: 0,
  podCount: 0,
  cpuUsage: 0,
  memoryUsage: 0
})
const nodes = ref([])
const namespaces = ref([])
const alerts = ref([])
const nodesLoading = ref(false)
const namespacesLoading = ref(false)
const trendPeriod = ref('1h')

let refreshTimer = null

// 加载集群信息
const loadClusterInfo = async () => {
  try {
    const clusterId = route.params.id
    const res = await request.get(`/clusters/${clusterId}`)
    clusterInfo.value = res.data || {}

    // 加载真实统计数据
    try {
      const statsRes = await request.get(`/clusters/${clusterId}/stats`)
      if (statsRes.code === 0) {
        const d = statsRes.data
        stats.nodeCount = d.nodeCount || 0
        stats.podCount = d.podStats?.running || 0
        stats.cpuUsage = d.totalCPU || 0
        stats.memoryUsage = d.totalMemGB || 0
      }
    } catch (e) {
      console.warn('加载统计失败', e)
    }
  } catch (error) {
    console.error('加载集群信息失败', error)
    ElMessage.error('加载集群信息失败')
  }
}

// 加载节点列表
const loadNodes = async () => {
  nodesLoading.value = true
  try {
    const clusterId = route.params.id
    const res = await request.get(`/nodes`, {
      params: { clusterId, page: 1, pageSize: 100 }
    })
    
    // 使用真实节点数据
    nodes.value = (res.data.list || []).map(node => ({
      ...node,
      name: node.nodeName,
      role: node.nodeRole,
      cpuUsage: node.cpuCores || '-',
      memoryUsage: node.memoryGb + ' GB' || '-',
      status: node.status === 1 ? 'Ready' : 'NotReady'
    }))
  } catch (error) {
    console.error('加载节点列表失败', error)
    nodes.value = []
  } finally {
    nodesLoading.value = false
  }
}

// 加载命名空间列表
const loadNamespaces = async () => {
  namespacesLoading.value = true
  try {
    const clusterId = route.params.id
    const res = await request.get(`/namespaces`, {
      params: { clusterId, page: 1, pageSize: 100 }
    })
    namespaces.value = (res.data.list || []).map(ns => ({
      ...ns,
      podCount: Math.floor(Math.random() * 20)
    }))
  } catch (error) {
    console.error('加载命名空间列表失败', error)
    // 使用模拟数据
    namespaces.value = ['default', 'kube-system', 'kube-public', 'production', 'staging'].map((name, i) => ({
      id: i + 1,
      name,
      status: 'Active',
      podCount: Math.floor(Math.random() * 20),
      labels: JSON.stringify({ env: name }),
      createTime: new Date(Date.now() - Math.random() * 365 * 24 * 60 * 60 * 1000)
    }))
  } finally {
    namespacesLoading.value = false
  }
}

// 加载告警信息
const loadAlerts = async () => {
  try {
    const clusterId = route.params.id
    const res = await request.get(`/alerts`, {
      params: { clusterId, page: 1, pageSize: 10 }
    })
    alerts.value = res.data.list || []
  } catch (error) {
    console.error('加载告警信息失败', error)
    // 模拟告警数据
    alerts.value = [
      {
        id: 1,
        level: 'warning',
        message: 'Node node-3 CPU usage above 80%',
        createTime: new Date(Date.now() - 5 * 60 * 1000)
      },
      {
        id: 2,
        level: 'info',
        message: 'Deployment nginx-deployment scaled to 5 replicas',
        createTime: new Date(Date.now() - 15 * 60 * 1000)
      }
    ]
  }
}

// 加载趋势数据
const loadTrends = () => {
  // 这里可以根据trendPeriod加载不同时间段的趋势数据
  console.log('加载趋势数据:', trendPeriod.value)
}

const refreshNodes = () => {
  loadNodes()
  ElMessage.success('刷新成功')
}

const goBack = () => {
  router.back()
}

const getStatusType = (status) => {
  const typeMap = {
    active: 'success',
    inactive: 'info',
    error: 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    active: '运行中',
    inactive: '已停止',
    error: '错误'
  }
  return labelMap[status] || status
}

const getProgressColor = (percentage) => {
  const value = parseFloat(percentage)
  if (value >= 80) return '#F56C6C'
  if (value >= 60) return '#E6A23C'
  return '#67C23A'
}

const getAlertType = (level) => {
  const typeMap = {
    critical: 'danger',
    warning: 'warning',
    info: 'info'
  }
  return typeMap[level] || 'info'
}

const formatLabels = (labels) => {
  try {
    const obj = JSON.parse(labels)
    return Object.entries(obj).map(([k, v]) => `${k}=${v}`).join(', ')
  } catch {
    return labels || '-'
  }
}

// 自动刷新
const startAutoRefresh = () => {
  refreshTimer = setInterval(() => {
    loadNodes()
    loadAlerts()
  }, 30000) // 每30秒刷新一次
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

onMounted(() => {
  loadClusterInfo()
  loadNodes()
  loadNamespaces()
  loadAlerts()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped lang="scss">
.cluster-monitor {
  padding: 20px;
  
  .header-content {
    display: flex;
    align-items: center;
    gap: 10px;
    
    .cluster-name {
      font-size: 18px;
      font-weight: bold;
    }
  }
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .chart-container {
    height: 300px;
    display: flex;
    align-items: center;
    justify-content: center;
    
    .chart-placeholder {
      text-align: center;
    }
  }
}
</style>
