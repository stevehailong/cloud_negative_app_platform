<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card" v-loading="loading">
          <div class="stat-item">
            <el-icon class="stat-icon" color="#409EFF"><FolderOpened /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ stats.projectCount }}</div>
              <div class="stat-label">项目总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card" v-loading="loading">
          <div class="stat-item">
            <el-icon class="stat-icon" color="#67C23A"><Grid /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ stats.appCount }}</div>
              <div class="stat-label">应用总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card" v-loading="loading">
          <div class="stat-item">
            <el-icon class="stat-icon" color="#E6A23C"><Connection /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ stats.todayBuilds }}</div>
              <div class="stat-label">今日构建</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card" v-loading="loading">
          <div class="stat-item">
            <el-icon class="stat-icon" color="#F56C6C"><Box /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ stats.todayDeploys }}</div>
              <div class="stat-label">今日部署</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近构建</span>
              <el-button type="primary" link size="small" @click="goToPipelines">更多</el-button>
            </div>
          </template>
          <el-table v-loading="buildLoading" :data="recentBuilds" style="width: 100%">
            <el-table-column prop="pipelineName" label="流水线" width="180" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createTime" label="时间" />
          </el-table>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近部署</span>
              <el-button type="primary" link size="small" @click="goToDeployments">更多</el-button>
            </div>
          </template>
          <el-table v-loading="deployLoading" :data="recentDeploys" style="width: 100%">
            <el-table-column prop="workloadName" label="工作负载" width="180" />
            <el-table-column prop="deploymentStatus" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getDeployStatusType(row.deploymentStatus)" size="small">
                  {{ getDeployStatusText(row.deploymentStatus) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createTime" label="时间" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 环境健康状态 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>环境健康状态</span>
            </div>
          </template>
          <el-row :gutter="16" v-loading="envLoading">
            <el-col :span="8" v-for="env in environments" :key="env.id">
              <div class="env-card">
                <div class="env-name">{{ env.envName }}</div>
                <div class="env-status">
                  <el-tag :type="env.status === 'healthy' ? 'success' : 'danger'" size="large">
                    {{ env.status === 'healthy' ? '健康' : '异常' }}
                  </el-tag>
                </div>
                <div class="env-stats">
                  <span>应用数: {{ env.appCount || 0 }}</span>
                  <span>Pod数: {{ env.podCount || 0 }}</span>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { FolderOpened, Grid, Connection, Box } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const router = useRouter()

// 统计数据
const loading = ref(false)
const stats = ref({
  projectCount: 0,
  appCount: 0,
  todayBuilds: 0,
  todayDeploys: 0
})

// 最近构建
const buildLoading = ref(false)
const recentBuilds = ref([])

// 最近部署
const deployLoading = ref(false)
const recentDeploys = ref([])

// 环境状态
const envLoading = ref(false)
const environments = ref([])

// 获取统计数据
const fetchStats = async () => {
  loading.value = true
  try {
    // 获取项目总数
    const projectRes = await request({
      url: '/projects?page=1&pageSize=1',
      method: 'get'
    })
    stats.value.projectCount = projectRes.data?.total || 0

    // 获取应用总数
    const appRes = await request({
      url: '/applications?page=1&pageSize=1',
      method: 'get'
    })
    stats.value.appCount = appRes.data?.total || 0

    // 获取今日构建数（查询今天的pipeline_runs）
    const today = new Date().toISOString().split('T')[0]
    try {
      const buildRes = await request({
        url: '/pipeline-runs',
        method: 'get',
        params: {
          startDate: today,
          page: 1,
          pageSize: 1
        }
      })
      stats.value.todayBuilds = buildRes.data?.total || 0
    } catch (e) {
      // 如果接口不存在，使用默认值
      stats.value.todayBuilds = 0
    }

    // 获取今日部署数（使用新版 app-deployments 接口）
    try {
      const deployRes = await request({
        url: '/app-deployments',
        method: 'get',
        params: {
          page: 1,
          page_size: 1
        }
      })
      stats.value.todayDeploys = deployRes.data?.total || 0
    } catch (e) {
      stats.value.todayDeploys = 0
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 获取最近构建
const fetchRecentBuilds = async () => {
  buildLoading.value = true
  try {
    const { data } = await request({
      url: '/pipeline-runs',
      method: 'get',
      params: {
        page: 1,
        pageSize: 5,
        sortBy: 'createTime',
        sortOrder: 'desc'
      }
    })
    recentBuilds.value = (data.list || []).map(item => ({
      ...item,
      pipelineName: item.pipelineName || `Pipeline #${item.pipelineId}`,
      createTime: formatTime(item.createTime)
    }))
  } catch (error) {
    console.error('获取最近构建失败:', error)
    recentBuilds.value = []
  } finally {
    buildLoading.value = false
  }
}

// 获取最近部署
const fetchRecentDeploys = async () => {
  deployLoading.value = true
  try {
    const { data } = await request({
      url: '/app-deployments',
      method: 'get',
      params: {
        page: 1,
        page_size: 5
      }
    })
    recentDeploys.value = (data.list || []).map(item => ({
      ...item,
      workloadName: item.workload_name || item.workloadName || '-',
      deploymentStatus: item.deployment_status || item.deploymentStatus || '-',
      createTime: formatTime(item.create_time || item.createTime)
    }))
  } catch (error) {
    console.error('获取最近部署失败:', error)
    recentDeploys.value = []
  } finally {
    deployLoading.value = false
  }
}

// 获取环境状态
const fetchEnvironments = async () => {
  envLoading.value = true
  try {
    const { data } = await request({
      url: '/environments?page=1&pageSize=10',
      method: 'get'
    })
    const envList = data.list || []

    // 获取应用绑定计数（每个环境绑定了多少应用）
    let appCountByEnv = {}
    let podCountByEnv = {}
    try {
      const bindRes = await request({
        url: '/app-env-bindings?page=1&pageSize=1000',
        method: 'get'
      })
      const bindings = bindRes.data?.list || []
      bindings.forEach(b => {
        const envId = b.envId || b.env_id
        appCountByEnv[envId] = (appCountByEnv[envId] || 0) + 1
      })
    } catch (e) {
      console.warn('获取绑定数据失败:', e)
    }

    // 获取部署 pod 计数
    try {
      const deployRes = await request({
        url: '/app-deployments?page=1&pageSize=1000',
        method: 'get'
      })
      const deployments = deployRes.data?.list || []
      deployments.forEach(d => {
        const envId = d.env_id || d.envId
        const replicas = d.available_replicas || d.availableReplicas || 0
        podCountByEnv[envId] = (podCountByEnv[envId] || 0) + replicas
      })
    } catch (e) {
      console.warn('获取部署数据失败:', e)
    }

    environments.value = envList.map(env => ({
      ...env,
      status: 'healthy',
      appCount: appCountByEnv[env.id] || 0,
      podCount: podCountByEnv[env.id] || 0
    }))
  } catch (error) {
    console.error('获取环境状态失败:', error)
    environments.value = []
  } finally {
    envLoading.value = false
  }
}

// 构建状态
const getStatusType = (status) => {
  const map = {
    success: 'success',
    running: 'warning',
    failed: 'danger',
    pending: 'info'
  }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = {
    success: '成功',
    running: '运行中',
    failed: '失败',
    pending: '等待中'
  }
  return map[status] || status
}

// 部署状态
const getDeployStatusType = (status) => {
  const map = {
    running: 'success',
    deploying: 'warning',
    failed: 'danger',
    stopped: 'info'
  }
  return map[status] || 'info'
}

const getDeployStatusText = (status) => {
  const map = {
    running: '运行中',
    deploying: '部署中',
    failed: '失败',
    stopped: '已停止'
  }
  return map[status] || status
}

// 跳转
const goToPipelines = () => {
  router.push('/pipelines')
}

const goToDeployments = () => {
  router.push('/deployments')
}

// 初始化
onMounted(() => {
  fetchStats()
  fetchRecentBuilds()
  fetchRecentDeploys()
  fetchEnvironments()
})
</script>

<style scoped lang="scss">
.dashboard {
  padding: 20px;
  
  .stat-card {
    .stat-item {
      display: flex;
      align-items: center;
      
      .stat-icon {
        font-size: 48px;
        margin-right: 20px;
      }
      
      .stat-content {
        .stat-value {
          font-size: 28px;
          font-weight: bold;
          color: #303133;
        }
        
        .stat-label {
          font-size: 14px;
          color: #909399;
          margin-top: 4px;
        }
      }
    }
  }
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: bold;
  }

  .env-card {
    padding: 20px;
    border: 1px solid #ebeef5;
    border-radius: 4px;
    text-align: center;
    margin-bottom: 16px;
    
    .env-name {
      font-size: 16px;
      font-weight: bold;
      margin-bottom: 12px;
    }
    
    .env-status {
      margin-bottom: 12px;
    }
    
    .env-stats {
      display: flex;
      justify-content: space-around;
      font-size: 14px;
      color: #909399;
      
      span {
        padding: 0 10px;
      }
    }
  }
}
</style>
