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
          <!--
            指标计算口径说明：
            - CPU使用率: K8s container_cpu_usage_seconds_total / 无K8s时用 goroutine 数×5 估算
            - 内存使用率: K8s container_memory_working_set_bytes / limit  / 无K8s时用 Go heap/512MB 估算
            - QPS: mycloud_http_requests_total 按 service=<应用名> 的 rate 聚合
            - 错误率: 5xx 响应占该应用总请求的比例
            数据源: Prometheus (优先K8s容器指标,其次mycloud自定义指标,无Prometheus时回退K8s Pod估算)
          -->
          <el-row :gutter="16" class="metrics-cards">
            <el-col :span="6">
              <el-tooltip content="容器CPU使用率 = sum(rate(container_cpu_usage_seconds_total[5m])) × 100；无K8s指标时用 goroutine×5 估算；按 app 标签过滤" placement="top">
                <el-card class="metric-card">
                  <div class="metric-title">
                    CPU使用率
                    <el-tag size="small" type="info" effect="plain" style="margin-left:4px">{{ metrics.dataSource || '--' }}</el-tag>
                  </div>
                  <div class="metric-value">{{ metrics.cpu === '--' ? '--' : metrics.cpu + '%' }}</div>
                  <div class="metric-trend" :class="getTrendClass(metrics.cpuTrend)">
                    {{ metrics.cpuTrend }}
                  </div>
                </el-card>
              </el-tooltip>
            </el-col>
            <el-col :span="6">
              <el-tooltip content="容器内存使用率 = sum(container_memory_working_set_bytes) / sum(kube_pod_container_resource_limits) × 100；无K8s指标时用 Go heap/512MB 估算；按 app 标签过滤" placement="top">
                <el-card class="metric-card">
                  <div class="metric-title">
                    内存使用率
                    <el-tag size="small" type="info" effect="plain" style="margin-left:4px">{{ metrics.dataSource || '--' }}</el-tag>
                  </div>
                  <div class="metric-value">{{ metrics.memory === '--' ? '--' : metrics.memory + '%' }}</div>
                  <div class="metric-trend" :class="getTrendClass(metrics.memoryTrend)">
                    {{ metrics.memoryTrend }}
                  </div>
                </el-card>
              </el-tooltip>
            </el-col>
            <el-col :span="6">
              <el-tooltip content="QPS = sum(rate(mycloud_http_requests_total{service=<应用名>}[5m]))；按 service 标签过滤，仅统计已部署且上报指标的应用" placement="top">
                <el-card class="metric-card">
                  <div class="metric-title">
                    请求QPS
                    <el-tag size="small" type="info" effect="plain" style="margin-left:4px">{{ metrics.dataSource || '--' }}</el-tag>
                  </div>
                  <div class="metric-value">{{ metrics.qps === '--' ? '--' : metrics.qps }}</div>
                  <div class="metric-trend" :class="getTrendClass(metrics.qpsTrend)">
                    {{ metrics.qpsTrend }}
                  </div>
                </el-card>
              </el-tooltip>
            </el-col>
            <el-col :span="6">
              <el-tooltip content="错误率 = sum(rate({service=<app>,status=~5xx}[5m])) / sum(rate({service=<app>}[5m])) x 100；按 service 标签过滤，仅统计应用的 5xx 错误占比" placement="top">
                <el-card class="metric-card">
                  <div class="metric-title">
                    错误率
                    <el-tag size="small" type="info" effect="plain" style="margin-left:4px">{{ metrics.dataSource || '--' }}</el-tag>
                  </div>
                  <div class="metric-value">{{ metrics.errorRate === '--' ? '--' : metrics.errorRate + '%' }}</div>
                  <div class="metric-trend" :class="getTrendClass(metrics.errorTrend)">
                    {{ metrics.errorTrend }}
                  </div>
                </el-card>
              </el-tooltip>
            </el-col>
          </el-row>

          <!-- 图表区：集成 Grafana -->
          <div class="chart-area">
            <div v-if="!grafanaConfig.enabled" class="chart-placeholder">
              <el-empty description="未配置 Grafana，请到 系统设置 → 集成配置 中填写 Grafana 地址" />
            </div>
            <div v-else>
              <div class="chart-toolbar">
                <span class="chart-source">Grafana 仪表盘: my-cloud-监控概览 | 时间范围: {{ metricsQuery.timeRange }} | 筛选: {{ metricsQuery.type }}={{ selectedTargetName }}</span>
                <el-link type="primary" :href="grafanaConfig.grafanaUrl" target="_blank">
                  打开 Grafana
                </el-link>
              </div>
              <iframe
                :src="grafanaIframeUrl"
                class="grafana-iframe"
                frameborder="0"
              />
            </div>
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
                <el-radio label="filter">条件筛选</el-radio>
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
              <span v-if="selectedAppName" style="margin-left: 8px; color: #909399; font-size: 12px;">
                按服务名查询: {{ selectedAppName }}
              </span>
            </el-form-item>
            <template v-if="tracesQuery.type === 'filter'">
              <el-form-item label="服务名">
                <el-select v-model="tracesQuery.serviceName" clearable placeholder="全部" style="width: 180px">
                  <el-option v-for="s in traceServices" :key="s" :label="s" :value="s" />
                </el-select>
              </el-form-item>
              <el-form-item label="耗时(ms)">
                <el-input-number v-model="tracesQuery.minDuration" :min="0" placeholder="最小" style="width: 120px" controls-position="right" />
                <span style="margin: 0 8px">-</span>
                <el-input-number v-model="tracesQuery.maxDuration" :min="0" placeholder="最大" style="width: 120px" controls-position="right" />
              </el-form-item>
              <el-form-item label="状态">
                <el-select v-model="tracesQuery.hasError" clearable placeholder="全部" style="width: 100px">
                  <el-option label="正常" :value="0" />
                  <el-option label="错误" :value="1" />
                </el-select>
              </el-form-item>
            </template>
            <el-form-item>
              <el-button type="primary" :loading="tracesLoading" @click="fetchTraces">查询</el-button>
            </el-form-item>
          </el-form>

          <el-divider />

          <!-- 统计信息 -->
          <el-row v-if="traceStats.totalTraces !== undefined" :gutter="16" class="trace-stats-row">
            <el-col :span="6">
              <div class="trace-stat-item">
                <div class="stat-label">Trace总数</div>
                <div class="stat-value">{{ traceStats.totalTraces }}</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="trace-stat-item">
                <div class="stat-label">平均耗时</div>
                <div class="stat-value">{{ formatMs(traceStats.avgDurationMs) }}</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="trace-stat-item">
                <div class="stat-label">错误率</div>
                <div class="stat-value" :class="{ 'error-rate': traceStats.errorRate > 5 }">{{ formatPercent(traceStats.errorRate) }}</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="trace-stat-item">
                <div class="stat-label">活跃服务</div>
                <div class="stat-value">{{ (traceStats.topServices || []).length }}</div>
              </div>
            </el-col>
          </el-row>

          <!-- 链路列表 -->
          <div v-loading="tracesLoading">
            <el-table :data="traces" style="width: 100%">
              <el-table-column prop="traceId" label="TraceID" width="280" />
              <el-table-column prop="serviceName" label="服务名称" width="180" />
              <el-table-column prop="operationName" label="操作" min-width="200" />
              <el-table-column label="耗时" width="140">
                <template #default="{ row }">
                  <div class="duration-cell">
                    <div class="duration-bar" :style="{ width: durationPercent(row) + '%', background: durationColor(row) }"></div>
                    <span class="duration-text">{{ row.durationMs }}ms</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="method" label="方法" width="80" />
              <el-table-column prop="startTime" label="开始时间" width="210" :formatter="formatStartTime" />
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

            <div class="pagination-wrap">
              <el-pagination
                v-model:current-page="tracesQuery.page"
                :page-size="tracesQuery.pageSize"
                :total="tracesTotal"
                layout="total, prev, pager, next"
                @current-change="fetchTraces"
              />
            </div>
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

    <!-- 链路详情弹窗 -->
    <el-dialog
      v-model="traceDetailVisible"
      title="链路详情"
      width="1000px"
      destroy-on-close
    >
      <div v-if="traceDetailData.spans.length" class="trace-detail">
        <div class="trace-info">
          <span class="trace-label">TraceID:</span>
          <span class="trace-value">{{ traceDetailData.traceId }}</span>
          <span class="trace-label">Span总数:</span>
          <span class="trace-value">{{ traceDetailData.spans.length }}</span>
          <span class="trace-label">总耗时:</span>
          <span class="trace-value">{{ totalTraceDuration }}ms</span>
        </div>

        <!-- 水瀑布图 -->
        <div class="waterfall-container">
          <div class="waterfall-header">
            <span class="wf-col-service">服务 / 操作</span>
            <span class="wf-col-duration">耗时分布</span>
            <span class="wf-col-time">耗时(ms)</span>
          </div>
          <div
            v-for="(node, index) in spanTree"
            :key="index"
            class="waterfall-row"
            :class="{ 'has-error': node.hasError }"
            :style="{ paddingLeft: (node.depth * 20 + 12) + 'px' }"
          >
            <div class="wf-col-service">
              <span class="span-service-tag" :style="{ background: serviceColor(node.serviceName) }">
                {{ node.serviceName }}
              </span>
              <span class="span-operation">{{ node.operationName }}</span>
            </div>
            <div class="wf-col-duration">
              <div
                class="wf-bar"
                :style="{
                  width: barWidth(node) + '%',
                  marginLeft: barOffset(node) + '%',
                  background: barColor(node)
                }"
                :title="node.durationMs + 'ms'"
              ></div>
            </div>
            <div class="wf-col-time">
              <span :class="{ 'error-text': node.hasError }">{{ node.durationMs }}ms</span>
              <el-tag v-if="node.statusCode >= 400" type="danger" size="small" style="margin-left: 6px">{{ node.statusCode }}</el-tag>
            </div>
          </div>
        </div>
      </div>
      <el-empty v-else description="暂无Span数据" />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import axios from 'axios'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const formatStartTime = (row) => formatTime(row.startTime)

// 当前Tab
const activeTab = ref('metrics')

// 应用列表
const applications = ref([])

// Grafana 集成配置
const grafanaConfig = ref({ grafanaUrl: '', enabled: false })

// 计算 Grafana iframe URL：嵌入 "My Cloud 监控概览" Dashboard，kiosk 模式只显示图表
const grafanaIframeUrl = computed(() => {
  if (!grafanaConfig.value.enabled) return ''
  const base = grafanaConfig.value.grafanaUrl.replace(/\/$/, '')
  const range = mapTimeRangeToGrafana(metricsQuery.timeRange)
  return `${base}/d/my-cloud-overview/my-cloud-%E7%9B%91%E6%8E%A7%E6%A6%82%E8%A7%88?kiosk=tv&theme=light&from=${range.from}&to=${range.to}`
})

// 把前端 timeRange 转换为 Grafana 的 from/to 参数
const mapTimeRangeToGrafana = (timeRange) => {
  const map = {
    '1h': 'now-1h',
    '6h': 'now-6h',
    '24h': 'now-24h',
    '7d': 'now-7d'
  }
  return { from: map[timeRange] || 'now-1h', to: 'now' }
}

// 加载 Grafana 配置
const loadGrafanaConfig = async () => {
  try {
    const res = await request({ url: '/settings/grafana-config', method: 'get' })
    grafanaConfig.value = res.data || { grafanaUrl: '', enabled: false }
  } catch (e) {
    grafanaConfig.value = { grafanaUrl: '', enabled: false }
  }
}

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
  errorTrend: '--',
  dataSource: '--'
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
    targetOptions.value = (data.list || []).map(item => {
      // 根据不同类型选择正确的名称字段
      let name = ''
      if (metricsQuery.type === 'app') {
        name = item.name
      } else if (metricsQuery.type === 'environment') {
        name = item.envName || item.name
      } else if (metricsQuery.type === 'cluster') {
        name = item.clusterName || item.name
      }
      return {
        id: item.id,
        name: name,
        code: item.code // 保存应用的code，用于K8s标签查询
      }
    })
    
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
    // 查找当前选中的应用信息
    const selectedApp = targetOptions.value.find(item => item.id === metricsQuery.targetId)
    const appName = selectedApp?.name || '' // 使用name字段，不是code
    
    // 使用完整路径，绕过 baseURL
    const response = await axios.get(`/internal/v1/metrics/apps/${metricsQuery.targetId}`, {
      params: { 
        timeRange: metricsQuery.timeRange,
        appName: appName // 传递appName用于K8s标签查询
      }
    })
    const data = response.data.data
    
    // 格式化数值，保留小数点后1位
    const formatNumber = (val) => {
      if (val === null || val === undefined || val === '--') return '--'
      if (typeof val === 'number') {
        return val.toFixed(1)
      }
      return val
    }
    
    metrics.value = {
      cpu: formatNumber(data.cpu),
      cpuTrend: data.cpuTrend || '--',
      memory: formatNumber(data.memory),
      memoryTrend: data.memoryTrend || '--',
      qps: data.qps !== undefined && data.qps !== null ? Math.round(data.qps) : '--',
      qpsTrend: data.qpsTrend || '--',
      errorRate: formatNumber(data.errorRate),
      errorTrend: data.errorTrend || '--',
      dataSource: data.data_source || data.dataSource || '--'
    }
  } catch (error) {
    console.error('获取指标失败:', error)
    // 重置为初始值
    metrics.value = {
      cpu: '--',
      cpuTrend: '--',
      memory: '--',
      memoryTrend: '--',
      qps: '--',
      qpsTrend: '--',
      errorRate: '--',
      errorTrend: '--',
      dataSource: '--'
    }
  }
}

// 当前选中的目标名称（用于图表区标注）
const selectedTargetName = computed(() => {
  const item = targetOptions.value.find(i => i.id === metricsQuery.targetId)
  return item?.name || item?.code || metricsQuery.targetId || '--'
})
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
      url = `/internal/v1/logs/apps/${logsQuery.appId}`
    } else if (logsQuery.type === 'pod' && logsQuery.podName) {
      url = `/internal/v1/logs/pods/${logsQuery.podName}`
    }
    
    if (!url) {
      logs.value = []
      return
    }
    
    const response = await axios.get(url, { params })
    logs.value = response.data.data?.logs || []
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
  type: 'filter',
  traceId: '',
  appId: '',
  serviceName: '',
  minDuration: undefined,
  maxDuration: undefined,
  hasError: undefined,
  page: 1,
  pageSize: 20
})

const tracesLoading = ref(false)
const traces = ref([])
const tracesTotal = ref(0)
const traceServices = ref([])
const traceStats = ref({})

// 根据选中的 appId 获取应用名称
const selectedAppName = computed(() => {
  if (!tracesQuery.appId) return ''
  const app = applications.value.find(a => a.id === tracesQuery.appId)
  return app ? app.name : ''
})

// 加载服务列表
const loadTraceServices = async () => {
  try {
    const { data } = await request({ url: '/traces/services/list', method: 'get' })
    traceServices.value = data.services || []
  } catch (e) {
    // 静默失败
  }
}

// 加载统计信息
const loadTraceStats = async () => {
  try {
    const { data } = await request({ url: '/traces/stats', method: 'get' })
    traceStats.value = data
  } catch (e) {
    // 静默失败
  }
}

// 格式化毫秒
const formatMs = (val) => {
  if (val === null || val === undefined) return '--'
  const n = Number(val)
  if (n < 1000) return n.toFixed(1) + 'ms'
  return (n / 1000).toFixed(2) + 's'
}

const formatPercent = (val) => {
  if (val === null || val === undefined) return '--'
  return Number(val).toFixed(1) + '%'
}

// 获取链路
const fetchTraces = async () => {
  tracesLoading.value = true
  try {
    let url = ''
    let params = {}

    if (tracesQuery.type === 'traceId' && tracesQuery.traceId) {
      url = `/traces/${tracesQuery.traceId}`
    } else if (tracesQuery.type === 'app' && tracesQuery.appId) {
      url = `/traces/apps/${tracesQuery.appId}`
      params.page = tracesQuery.page
      params.pageSize = tracesQuery.pageSize
      // 传递应用名称作为 serviceName，以便按服务名筛选 Trace
      if (selectedAppName.value) {
        params.serviceName = selectedAppName.value
      }
    } else {
      url = '/traces'
      params = {
        page: tracesQuery.page,
        pageSize: tracesQuery.pageSize
      }
      if (tracesQuery.serviceName) params.serviceName = tracesQuery.serviceName
      if (tracesQuery.minDuration) params.minDuration = tracesQuery.minDuration
      if (tracesQuery.maxDuration) params.maxDuration = tracesQuery.maxDuration
      if (tracesQuery.hasError !== undefined && tracesQuery.hasError !== null && tracesQuery.hasError !== '') {
        params.hasError = tracesQuery.hasError
      }
    }

    const { data } = await request({ url, method: 'get', params })

    if (tracesQuery.type === 'traceId' && data.spans) {
      const rootSpan = data.spans[0]
      traces.value = rootSpan
        ? [{ ...rootSpan, spanCount: data.total || data.spans.length }]
        : []
      tracesTotal.value = traces.value.length
    } else if (data.list) {
      traces.value = data.list.map(item => ({
        ...item,
        durationMs: item.durationMs,
        spanCount: 1
      }))
      tracesTotal.value = data.total || 0
    } else {
      traces.value = []
      tracesTotal.value = 0
    }
  } catch (error) {
    console.error('获取链路失败:', error)
    traces.value = []
  } finally {
    tracesLoading.value = false
  }
}

// 耗时占比和颜色
const maxDurationInList = computed(() => {
  if (!traces.value.length) return 1
  return Math.max(...traces.value.map(t => t.durationMs || 0), 1)
})

const durationPercent = (row) => {
  return Math.min(((row.durationMs || 0) / maxDurationInList.value) * 100, 100)
}

const durationColor = (row) => {
  if (row.hasError) return '#f56c6c'
  const pct = durationPercent(row)
  if (pct > 80) return '#e6a23c'
  return '#409eff'
}

// ===== 水瀑布图相关 =====
const traceDetailVisible = ref(false)
const traceDetailData = ref({ traceId: '', spans: [] })

// 服务颜色映射
const serviceColors = {}
const colorPalette = [
  '#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399',
  '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7',
  '#DDA0DD', '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E9'
]
let colorIndex = 0
const serviceColor = (name) => {
  if (!serviceColors[name]) {
    serviceColors[name] = colorPalette[colorIndex % colorPalette.length]
    colorIndex++
  }
  return serviceColors[name]
}

// 构建 Span 树
const buildSpanTree = (spans) => {
  if (!spans || spans.length === 0) return []

  const spanMap = {}
  spans.forEach(s => { spanMap[s.spanId] = { ...s, children: [], depth: 0 } })

  const roots = []
  spans.forEach(s => {
    const node = spanMap[s.spanId]
    if (s.parentSpanId && spanMap[s.parentSpanId]) {
      spanMap[s.parentSpanId].children.push(node)
    } else {
      roots.push(node)
    }
  })

  // BFS 计算深度，按开始时间排序
  const flatList = []
  const traverse = (nodes, depth) => {
    nodes.sort((a, b) => {
      if (a.startTime < b.startTime) return -1
      if (a.startTime > b.startTime) return 1
      return 0
    })
    nodes.forEach(node => {
      node.depth = depth
      flatList.push(node)
      if (node.children.length > 0) {
        traverse(node.children, depth + 1)
      }
    })
  }
  traverse(roots, 0)

  return flatList
}

const spanTree = computed(() => buildSpanTree(traceDetailData.value.spans))

const totalTraceDuration = computed(() => {
  const spans = traceDetailData.value.spans
  if (!spans.length) return 0
  const startTimes = spans.map(s => new Date(s.startTime).getTime()).filter(t => !isNaN(t))
  const endTimes = spans.map(s => s.endTime ? new Date(s.endTime).getTime() : null).filter(t => t && !isNaN(t))
  if (startTimes.length === 0) return 0
  const minStart = Math.min(...startTimes)
  const maxEnd = endTimes.length > 0 ? Math.max(...endTimes) : Math.max(...startTimes)
  return maxEnd - minStart
})

const barWidth = (node) => {
  if (totalTraceDuration.value <= 0) return 0
  return Math.min((node.durationMs / totalTraceDuration.value) * 100, 100)
}

const barOffset = (node) => {
  if (totalTraceDuration.value <= 0 || !node.startTime) return 0
  const spans = traceDetailData.value.spans
  const startTimes = spans.map(s => new Date(s.startTime).getTime()).filter(t => !isNaN(t))
  if (startTimes.length === 0) return 0
  const minStart = Math.min(...startTimes)
  const nodeStart = new Date(node.startTime).getTime()
  return ((nodeStart - minStart) / totalTraceDuration.value) * 100
}

const barColor = (node) => {
  if (node.hasError) return '#f56c6c'
  return serviceColor(node.serviceName)
}

const viewTraceDetail = async (row) => {
  if (!row.traceId) return
  try {
    // 重置颜色索引，保证同一 trace 颜色一致
    Object.keys(serviceColors).forEach(k => delete serviceColors[k])
    colorIndex = 0

    const { data } = await request({ url: `/traces/${row.traceId}`, method: 'get' })
    traceDetailData.value = {
      traceId: data.traceId || row.traceId,
      spans: data.spans || []
    }
    traceDetailVisible.value = true
  } catch (error) {
    console.error('获取链路详情失败:', error)
  }
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

  // 加载 Grafana 配置
  loadGrafanaConfig()

  // 加载指标目标选项
  fetchTargetOptions()

  // 加载链路追踪服务列表和统计
  loadTraceServices()
  loadTraceStats()
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

.chart-area {
  margin-top: 20px;
  min-height: 600px;
}

.chart-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #fafafa;
  border: 1px solid #ebeef5;
  border-bottom: none;
  border-radius: 4px 4px 0 0;

  .chart-source {
    color: #909399;
    font-size: 12px;
  }
}

.grafana-iframe {
  width: 100%;
  height: 600px;
  border: 1px solid #ebeef5;
  border-radius: 0 0 4px 4px;
  background: #fff;
}

.chart-placeholder {
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

.trace-detail {
  .trace-info {
    margin-bottom: 16px;
    padding: 12px;
    background: #f5f7fa;
    border-radius: 4px;

    .trace-label {
      font-size: 13px;
      color: #909399;
      margin-right: 8px;
    }

    .trace-value {
      font-size: 13px;
      color: #303133;
      font-weight: 500;
      margin-right: 24px;
    }
  }
}

// 统计信息
.trace-stats-row {
  margin-bottom: 16px;

  .trace-stat-item {
    text-align: center;
    padding: 12px;
    background: #f5f7fa;
    border-radius: 4px;

    .stat-label {
      font-size: 12px;
      color: #909399;
      margin-bottom: 4px;
    }
    .stat-value {
      font-size: 20px;
      font-weight: 600;
      color: #303133;
      &.error-rate {
        color: #f56c6c;
      }
    }
  }
}

// 耗时列
.duration-cell {
  display: flex;
  align-items: center;
  position: relative;

  .duration-bar {
    height: 6px;
    border-radius: 3px;
    min-width: 4px;
    margin-right: 8px;
  }
  .duration-text {
    font-size: 12px;
    white-space: nowrap;
  }
}

// 分页
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

// 水瀑布图
.waterfall-container {
  font-size: 13px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  overflow: auto;
  max-height: 500px;
}

.waterfall-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
  font-weight: 500;
  color: #606266;
  position: sticky;
  top: 0;
  z-index: 1;
}

.waterfall-row {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  border-bottom: 1px solid #f0f0f0;
  min-height: 36px;

  &:hover {
    background: #f5f7fa;
  }

  &.has-error {
    background: #fef0f0;
    &:hover {
      background: #fde2e2;
    }
  }
}

.wf-col-service {
  flex: 0 0 320px;
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;

  .span-service-tag {
    display: inline-block;
    padding: 1px 6px;
    border-radius: 3px;
    color: #fff;
    font-size: 11px;
    white-space: nowrap;
    flex-shrink: 0;
  }
  .span-operation {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #606266;
  }
}

.wf-col-duration {
  flex: 1;
  min-width: 100px;
  display: flex;
  align-items: center;
  padding: 0 12px;

  .wf-bar {
    height: 14px;
    border-radius: 3px;
    min-width: 2px;
    opacity: 0.8;
    transition: opacity 0.2s;
    &:hover {
      opacity: 1;
    }
  }
}

.wf-col-time {
  flex: 0 0 120px;
  text-align: right;
  font-size: 12px;
  color: #606266;

  .error-text {
    color: #f56c6c;
    font-weight: 500;
  }
}
</style>
