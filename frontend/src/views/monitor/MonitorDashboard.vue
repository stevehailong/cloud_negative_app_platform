<template>
  <div class="monitor-dashboard">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2>监控中心</h2>
        <span class="subtitle">实时监控、日志查询、链路追踪</span>
      </div>
    </div>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" class="monitor-tabs">
      <!-- 指标监控 -->
      <el-tab-pane label="指标监控" name="metrics">
        <el-card shadow="never">
          <el-form :inline="true" :model="metricsQuery">
            <el-form-item label="监控对象">
              <el-select v-model="metricsQuery.type" style="width: 150px">
                <el-option label="应用" value="app" />
                <el-option label="环境" value="environment" />
                <el-option label="集群" value="cluster" />
              </el-select>
            </el-form-item>
            <el-form-item :label="getMetricsLabel()">
              <el-select
                v-model="metricsQuery.targetId"
                filterable
                placeholder="请选择"
                style="width: 250px"
                @change="fetchMetrics"
              >
                <el-option
                  v-for="item in targetOptions"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="时间范围">
              <el-select v-model="metricsQuery.timeRange" @change="fetchMetrics" style="width: 150px">
                <el-option label="最近1小时" value="1h" />
                <el-option label="最近6小时" value="6h" />
                <el-option label="最近24小时" value="24h" />
                <el-option label="最近7天" value="7d" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchMetrics">查询</el-button>
            </el-form-item>
          </el-form>

          <el-divider />

          <!-- 指标卡片 -->
          <el-row :gutter="16" class="metrics-cards">
            <el-col :span="6">
              <el-card class="metric-card">
                <div class="metric-title">CPU使用率</div>
                <div class="metric-value">{{ metrics.cpu }}%</div>
                <div class="metric-trend" :class="getTrendClass(metrics.cpuTrend)">
                  {{ metrics.cpuTrend }}
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="metric-card">
                <div class="metric-title">内存使用率</div>
                <div class="metric-value">{{ metrics.memory }}%</div>
                <div class="metric-trend" :class="getTrendClass(metrics.memoryTrend)">
                  {{ metrics.memoryTrend }}
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="metric-card">
                <div class="metric-title">请求QPS</div>
                <div class="metric-value">{{ metrics.qps }}</div>
                <div class="metric-trend" :class="getTrendClass(metrics.qpsTrend)">
                  {{ metrics.qpsTrend }}
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="metric-card">
                <div class="metric-title">错误率</div>
                <div class="metric-value">{{ metrics.errorRate }}%</div>
                <div class="metric-trend" :class="getTrendClass(metrics.errorTrend)">
                  {{ metrics.errorTrend }}
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 图表占位 -->
          <div class="chart-placeholder">
            <el-empty description="指标图表展示区域（可集成 ECharts 或 Grafana）" />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 日志查询 -->
      <el-tab-pane label="日志查询" name="logs">
        <el-card shadow="never">
          <el-form :inline="true" :model="logsQuery">
            <el-form-item label="查询类型">
              <el-select v-model="logsQuery.type" style="width: 150px">
                <el-option label="应用日志" value="app" />
                <el-option label="Pod日志" value="pod" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="logsQuery.type === 'app'" label="应用">
              <el-select
                v-model="logsQuery.appId"
                filterable
                placeholder="请选择应用"
                style="width: 250px"
              >
                <el-option
                  v-for="app in applications"
                  :key="app.id"
                  :label="app.name"
                  :value="app.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item v-if="logsQuery.type === 'pod'" label="Pod名称">
              <el-input
                v-model="logsQuery.podName"
                placeholder="请输入Pod名称"
                clearable
                style="width: 250px"
              />
            </el-form-item>
            <el-form-item label="日志级别">
              <el-select v-model="logsQuery.level" clearable style="width: 120px">
                <el-option label="ERROR" value="error" />
                <el-option label="WARN" value="warn" />
                <el-option label="INFO" value="info" />
                <el-option label="DEBUG" value="debug" />
              </el-select>
            </el-form-item>
            <el-form-item label="关键词">
              <el-input
                v-model="logsQuery.keyword"
                placeholder="搜索日志内容"
                clearable
                style="width: 200px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="logsLoading" @click="fetchLogs">查询</el-button>
              <el-button @click="clearLogs">清空</el-button>
            </el-form-item>
          </el-form>

          <el-divider />

          <!-- 日志列表 -->
          <div v-loading="logsLoading" class="logs-container">
            <div v-if="logs.length === 0" class="logs-empty">
              <el-empty description="暂无日志数据" />
            </div>
            <div v-else class="logs-list">
              <div
                v-for="(log, index) in logs"
                :key="index"
                class="log-item"
                :class="'log-' + log.level"
              >
                <span class="log-time">{{ log.timestamp }}</span>
                <el-tag :type="getLevelType(log.level)" size="small" class="log-level">
                  {{ log.level }}
                </el-tag>
                <span class="log-content">{{ log.message }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 链路追踪 -->
      <el-tab-pane label="链路追踪" name="traces">
        <el-card shadow="never">
          <el-form :inline="true" :model="tracesQuery">
            <el-form-item label="查询方式">
              <el-radio-group v-model="tracesQuery.type">
                <el-radio label="traceId">TraceID</el-radio>
                <el-radio label="app">应用</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="tracesQuery.type === 'traceId'" label="TraceID">
              <el-input
                v-model="tracesQuery.traceId"
                placeholder="请输入TraceID"
                clearable
                style="width: 350px"
              />
            </el-form-item>
            <el-form-item v-if="tracesQuery.type === 'app'" label="应用">
              <el-select
                v-model="tracesQuery.appId"
                filterable
                placeholder="请选择应用"
                style="width: 250px"
              >
                <el-option
                  v-for="app in applications"
                  :key="app.id"
                  :label="app.name"
                  :value="app.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="tracesLoading" @click="fetchTraces">查询</el-button>
            </el-form-item>
          </el-form>

          <el-divider />

          <!-- 链路列表 -->
          <div v-loading="tracesLoading">
            <el-table :data="traces" style="width: 100%">
              <el-table-column prop="traceId" label="TraceID" width="280" />
              <el-table-column prop="serviceName" label="服务名称" width="200" />
              <el-table-column prop="operationName" label="操作" width="200" />
              <el-table-column prop="duration" label="耗时(ms)" width="120" />
              <el-table-column prop="spanCount" label="Span数量" width="100" />
              <el-table-column prop="startTime" label="开始时间" width="180" />
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.hasError ? 'danger' : 'success'" size="small">
                    {{ row.hasError ? '错误' : '正常' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="100">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="viewTraceDetail(row)">
                    详情
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 告警规则 -->
      <el-tab-pane label="告警规则" name="alerts">
        <el-card shadow="never">
          <el-empty description="告警规则管理功能开发中" />
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import request from '@/utils/request'

// 当前Tab
const activeTab = ref('metrics')

// 应用列表
const applications = ref([])

// ========== 指标监控 ==========
const metricsQuery = reactive({
  type: 'app',
  targetId: '',
  timeRange: '1h'
})

const targetOptions = ref([])

const metrics = ref({
  cpu: '--',
  cpuTrend: '--',
  memory: '--',
  memoryTrend: '--',
  qps: '--',
  qpsTrend: '--',
  errorRate: '--',
  errorTrend: '--'
})

// 获取目标选项
const fetchTargetOptions = async () => {
  try {
    let url = ''
    if (metricsQuery.type === 'app') {
      url = '/applications?page=1&pageSize=1000'
    } else if (metricsQuery.type === 'environment') {
      url = '/environments?page=1&pageSize=1000'
    } else if (metricsQuery.type === 'cluster') {
      url = '/clusters?page=1&pageSize=1000'
    }
    
    const { data } = await request({ url, method: 'get' })
    targetOptions.value = (data.list || []).map(item => ({
      id: item.id,
      name: item.name || item.appName || item.clusterName || item.envName
    }))
    
    if (targetOptions.value.length > 0) {
      metricsQuery.targetId = targetOptions.value[0].id
    }
  } catch (error) {
    console.error('获取目标选项失败:', error)
  }
}

// 获取指标
const fetchMetrics = async () => {
  if (!metricsQuery.targetId) return
  
  try {
    const { data } = await request({
      url: `/metrics/${metricsQuery.type}s/${metricsQuery.targetId}`,
      method: 'get',
      params: { timeRange: metricsQuery.timeRange }
    })
    
    metrics.value = {
      cpu: data.cpu || '--',
      cpuTrend: data.cpuTrend || '--',
      memory: data.memory || '--',
      memoryTrend: data.memoryTrend || '--',
      qps: data.qps || '--',
      qpsTrend: data.qpsTrend || '--',
      errorRate: data.errorRate || '--',
      errorTrend: data.errorTrend || '--'
    }
  } catch (error) {
    console.error('获取指标失败:', error)
  }
}

// 获取标签
const getMetricsLabel = () => {
  const map = {
    app: '应用名称',
    environment: '环境名称',
    cluster: '集群名称'
  }
  return map[metricsQuery.type] || '名称'
}

// 趋势样式
const getTrendClass = (trend) => {
  if (!trend || trend === '--') return ''
  if (trend.includes('↑')) return 'trend-up'
  if (trend.includes('↓')) return 'trend-down'
  return ''
}

// 监听类型变化
watch(() => metricsQuery.type, () => {
  fetchTargetOptions()
})

// ========== 日志查询 ==========
const logsQuery = reactive({
  type: 'app',
  appId: '',
  podName: '',
  level: '',
  keyword: ''
})

const logsLoading = ref(false)
const logs = ref([])

// 获取日志
const fetchLogs = async () => {
  logsLoading.value = true
  try {
    let url = ''
    let params = {
      level: logsQuery.level,
      keyword: logsQuery.keyword
    }
    
    if (logsQuery.type === 'app' && logsQuery.appId) {
      url = `/logs/apps/${logsQuery.appId}`
    } else if (logsQuery.type === 'pod' && logsQuery.podName) {
      url = `/logs/pods/${logsQuery.podName}`
    }
    
    if (!url) {
      logs.value = []
      return
    }
    
    const { data } = await request({ url, method: 'get', params })
    logs.value = data.logs || []
  } catch (error) {
    console.error('获取日志失败:', error)
    logs.value = []
  } finally {
    logsLoading.value = false
  }
}

// 清空日志
const clearLogs = () => {
  logs.value = []
}

// 日志级别类型
const getLevelType = (level) => {
  const map = {
    error: 'danger',
    warn: 'warning',
    info: 'info',
    debug: ''
  }
  return map[level?.toLowerCase()] || 'info'
}

// ========== 链路追踪 ==========
const tracesQuery = reactive({
  type: 'traceId',
  traceId: '',
  appId: ''
})

const tracesLoading = ref(false)
const traces = ref([])

// 获取链路
const fetchTraces = async () => {
  tracesLoading.value = true
  try {
    let url = ''
    
    if (tracesQuery.type === 'traceId' && tracesQuery.traceId) {
      url = `/traces/${tracesQuery.traceId}`
    } else if (tracesQuery.type === 'app' && tracesQuery.appId) {
      url = `/traces/apps/${tracesQuery.appId}`
    }
    
    if (!url) {
      traces.value = []
      return
    }
    
    const { data } = await request({ url, method: 'get' })
    traces.value = Array.isArray(data) ? data : [data]
  } catch (error) {
    console.error('获取链路失败:', error)
    traces.value = []
  } finally {
    tracesLoading.value = false
  }
}

// 查看链路详情
const viewTraceDetail = (row) => {
  console.log('查看链路详情:', row)
  // TODO: 打开详情弹窗或跳转到Jaeger/SkyWalking
}

// 初始化
onMounted(async () => {
  // 加载应用列表（用于日志和链路查询）
  try {
    const { data } = await request({
      url: '/applications?page=1&pageSize=1000',
      method: 'get'
    })
    applications.value = data.list || []
  } catch (error) {
    console.error('加载应用列表失败:', error)
  }
  
  // 加载指标目标选项
  fetchTargetOptions()
})
</script>

<style scoped lang="scss">
.monitor-dashboard {
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

.monitor-tabs {
  :deep(.el-tabs__content) {
    padding-top: 20px;
  }
}

.metrics-cards {
  margin-bottom: 20px;
  
  .metric-card {
    text-align: center;
    
    .metric-title {
      font-size: 14px;
      color: #909399;
      margin-bottom: 10px;
    }
    
    .metric-value {
      font-size: 32px;
      font-weight: bold;
      color: #303133;
      margin-bottom: 8px;
    }
    
    .metric-trend {
      font-size: 12px;
      color: #909399;
      
      &.trend-up {
        color: #f56c6c;
      }
      
      &.trend-down {
        color: #67c23a;
      }
    }
  }
}

.chart-placeholder {
  margin-top: 20px;
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed #dcdfe6;
  border-radius: 4px;
  background-color: #fafafa;
}

.logs-container {
  min-height: 400px;
  max-height: 600px;
  overflow-y: auto;
  
  .logs-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 400px;
  }
  
  .logs-list {
    .log-item {
      padding: 8px 12px;
      border-bottom: 1px solid #f0f0f0;
      font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
      font-size: 13px;
      line-height: 1.6;
      
      &:hover {
        background-color: #f5f7fa;
      }
      
      .log-time {
        color: #909399;
        margin-right: 8px;
      }
      
      .log-level {
        margin-right: 8px;
      }
      
      .log-content {
        color: #606266;
      }
      
      &.log-error {
        background-color: #fef0f0;
        
        .log-content {
          color: #f56c6c;
        }
      }
      
      &.log-warn {
        background-color: #fdf6ec;
        
        .log-content {
          color: #e6a23c;
        }
      }
    }
  }
}
</style>
