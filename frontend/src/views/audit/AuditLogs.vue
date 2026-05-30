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

// 模拟数据生成
const generateMockData = () => {
  const actions = ['create', 'update', 'delete', 'view']
  const resourceTypes = ['application', 'cluster', 'environment', 'project', 'user']
  const methods = ['GET', 'POST', 'PUT', 'DELETE']
  const users = ['admin', 'developer', 'tester']
  
  return Array.from({ length: 50 }, (_, i) => {
    const action = actions[Math.floor(Math.random() * actions.length)]
    const resourceType = resourceTypes[Math.floor(Math.random() * resourceTypes.length)]
    const method = methods[Math.floor(Math.random() * methods.length)]
    const username = users[Math.floor(Math.random() * users.length)]
    
    return {
      id: i + 1,
      userId: Math.floor(Math.random() * 10) + 1,
      username,
      action,
      resourceType,
      resourceId: Math.floor(Math.random() * 100) + 1,
      resourceName: `${resourceType}-${i + 1}`,
      method,
      path: `/api/v1/${resourceType}s/${action === 'view' ? '' : Math.floor(Math.random() * 100)}`,
      ipAddress: `192.168.1.${Math.floor(Math.random() * 255)}`,
      userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      requestBody: action !== 'view' ? JSON.stringify({ name: `test-${i}` }) : null,
      responseCode: Math.random() > 0.1 ? 200 : 403,
      responseMessage: Math.random() > 0.1 ? 'success' : 'permission denied',
      durationMs: Math.floor(Math.random() * 500) + 10,
      createTime: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000)
    }
  })
}

const mockData = generateMockData()

const fetchData = async () => {
  loading.value = true
  try {
    // 实际项目中应该调用真实API
    // const res = await request.get('/audit-logs', { params })
    
    // 模拟过滤
    let filtered = [...mockData]
    
    if (searchForm.username) {
      filtered = filtered.filter(log => log.username.includes(searchForm.username))
    }
    if (searchForm.action) {
      filtered = filtered.filter(log => log.action === searchForm.action)
    }
    if (searchForm.resourceType) {
      filtered = filtered.filter(log => log.resourceType === searchForm.resourceType)
    }
    if (dateRange.value && dateRange.value.length === 2) {
      filtered = filtered.filter(log => {
        const logTime = new Date(log.createTime)
        return logTime >= dateRange.value[0] && logTime <= dateRange.value[1]
      })
    }
    
    // 分页
    const start = (pagination.page - 1) * pagination.pageSize
    const end = start + pagination.pageSize
    tableData.value = filtered.slice(start, end)
    pagination.total = filtered.length
  } catch (error) {
    console.error('获取数据失败', error)
    ElMessage.error('获取审计日志失败')
  } finally {
    loading.value = false
  }
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

const handleExport = () => {
  ElMessage.info('导出功能开发中')
}

const viewDetail = (row) => {
  currentLog.value = row
  detailVisible.value = true
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
