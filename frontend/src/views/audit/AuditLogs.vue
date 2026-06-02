<template>
  <div class="audit-logs">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>审计日志</span>
          <el-button type="primary" size="small" @click="handleExport">
            <el-icon><Download /></el-icon>
            导出日志
          </el-button>
        </div>
      </template>
      
      <div class="search-bar">
        <el-form :inline="true" :model="searchForm">
          <el-form-item label="用户名">
            <el-input v-model="searchForm.username" placeholder="请输入用户名" clearable style="width: 150px" />
          </el-form-item>
          <el-form-item label="操作类型">
            <el-select v-model="searchForm.action" placeholder="请选择" clearable style="width: 120px">
              <el-option label="创建" value="create" />
              <el-option label="更新" value="update" />
              <el-option label="删除" value="delete" />
              <el-option label="查看" value="view" />
            </el-select>
          </el-form-item>
          <el-form-item label="资源类型">
            <el-select v-model="searchForm.resourceType" placeholder="请选择" clearable style="width: 150px">
              <el-option label="应用" value="application" />
              <el-option label="集群" value="cluster" />
              <el-option label="环境" value="environment" />
              <el-option label="项目" value="project" />
              <el-option label="用户" value="user" />
            </el-select>
          </el-form-item>
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              style="width: 360px"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>
      
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
      >
        <el-table-column prop="username" label="用户" width="120" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action)" size="small">
              {{ getActionLabel(row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="资源类型" width="120">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.resourceType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resourceName" label="资源名称" width="150" show-overflow-tooltip />
        <el-table-column label="HTTP方法" width="100">
          <template #default="{ row }">
            <el-tag :type="getMethodType(row.method)" size="small">{{ row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="请求路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="ipAddress" label="IP地址" width="140" />
        <el-table-column label="响应码" width="100">
          <template #default="{ row }">
            <el-tag :type="row.responseCode === 200 ? 'success' : 'danger'" size="small">
              {{ row.responseCode }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="durationMs" label="耗时(ms)" width="100" />
        <el-table-column label="操作时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="100">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top: 20px; justify-content: flex-end"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </el-card>
    
    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailVisible"
      title="审计日志详情"
      width="900px"
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户ID">{{ currentLog.userId }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ currentLog.username }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">
          <el-tag :type="getActionType(currentLog.action)">
            {{ getActionLabel(currentLog.action) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="资源类型">
          <el-tag type="info">{{ currentLog.resourceType }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="资源ID">{{ currentLog.resourceId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="资源名称">{{ currentLog.resourceName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="HTTP方法">
          <el-tag :type="getMethodType(currentLog.method)">{{ currentLog.method }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="响应码">
          <el-tag :type="currentLog.responseCode === 200 ? 'success' : 'danger'">
            {{ currentLog.responseCode }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="请求路径" :span="2">{{ currentLog.path }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentLog.ipAddress }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ currentLog.durationMs }}ms</el-descriptions-item>
        <el-descriptions-item label="操作时间" :span="2">{{ formatTime(currentLog.createTime) }}</el-descriptions-item>
        <el-descriptions-item label="User-Agent" :span="2">
          <el-text truncated style="max-width: 100%">{{ currentLog.userAgent || '-' }}</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="请求体" :span="2">
          <pre v-if="currentLog.requestBody" style="margin: 0; max-height: 200px; overflow: auto">{{ formatJSON(currentLog.requestBody) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="响应消息" :span="2">
          {{ currentLog.responseMessage || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const loading = ref(false)
const tableData = ref([])
const searchForm = reactive({
  username: '',
  action: '',
  resourceType: ''
})
const dateRange = ref([])
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const detailVisible = ref(false)
const currentLog = ref({})

const fetchData = async () => {
  loading.value = true
  try {
    // 构建查询参数
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    
    // 添加搜索条件
    if (searchForm.username) {
      params.username = searchForm.username
    }
    if (searchForm.action) {
      params.action = searchForm.action
    }
    if (searchForm.resourceType) {
      params.resourceType = searchForm.resourceType
    }
    
    // 处理时间范围
    if (dateRange.value && dateRange.value.length === 2) {
      // 格式化为 YYYY-MM-DD HH:mm:ss
      params.startTime = formatDateTime(dateRange.value[0])
      params.endTime = formatDateTime(dateRange.value[1])
    }
    
    // 调用真实API
    const res = await request.get('/audit-logs', { params })
    
    if (res.code === 0) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    } else {
      ElMessage.error(res.message || '获取审计日志失败')
      tableData.value = []
      pagination.total = 0
    }
  } catch (error) {
    console.error('获取审计日志失败', error)
    ElMessage.error('获取审计日志失败')
    tableData.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

const viewDetail = async (row) => {
  try {
    // 获取详细信息
    const res = await request.get(`/audit-logs/${row.id}`)
    if (res.code === 0) {
      currentLog.value = res.data || row
      detailVisible.value = true
    } else {
      ElMessage.error(res.message || '获取日志详情失败')
    }
  } catch (error) {
    console.error('获取日志详情失败', error)
    // 如果获取详情失败，仍然显示列表中的数据
    currentLog.value = row
    detailVisible.value = true
  }
}

// 格式化日期时间为字符串
const formatDateTime = (date) => {
  if (!date) return ''
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

const handleSearch = () => {
  pagination.page = 1
  fetchData()
}

const handleReset = () => {
  searchForm.username = ''
  searchForm.action = ''
  searchForm.resourceType = ''
  dateRange.value = []
  pagination.page = 1
  fetchData()
}

const handleExport = async () => {
  try {
    // 构建导出参数（与查询参数相同）
    const params = {
      page: 1,
      pageSize: 10000  // 导出最多10000条记录
    }
    
    if (searchForm.username) {
      params.username = searchForm.username
    }
    if (searchForm.action) {
      params.action = searchForm.action
    }
    if (searchForm.resourceType) {
      params.resourceType = searchForm.resourceType
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.startTime = formatDateTime(dateRange.value[0])
      params.endTime = formatDateTime(dateRange.value[1])
    }
    
    ElMessage.info('正在导出审计日志...')
    
    // 获取数据
    const res = await request.get('/audit-logs/export', { 
      params,
      responseType: 'blob'  // 接收二进制数据
    })
    
    // 创建下载链接
    const blob = new Blob([res], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    
    link.href = url
    link.download = `audit_logs_${formatDateTime(new Date()).replace(/[: ]/g, '_')}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('导出审计日志失败', error)
    ElMessage.error('导出失败')
  }
}

const getActionType = (action) => {
  const typeMap = {
    create: 'success',
    update: 'primary',
    delete: 'danger',
    view: 'info'
  }
  return typeMap[action] || ''
}

const getActionLabel = (action) => {
  const labelMap = {
    create: '创建',
    update: '更新',
    delete: '删除',
    view: '查看'
  }
  return labelMap[action] || action
}

const getMethodType = (method) => {
  const typeMap = {
    GET: 'info',
    POST: 'success',
    PUT: 'warning',
    DELETE: 'danger'
  }
  return typeMap[method] || ''
}

const formatJSON = (str) => {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.audit-logs {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .search-bar {
    margin-bottom: 20px;
  }
  
  pre {
    background: #f5f7fa;
    padding: 10px;
    border-radius: 4px;
    font-size: 12px;
    line-height: 1.5;
  }
}
</style>
