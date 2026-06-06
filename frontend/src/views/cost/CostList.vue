<template>
  <div class="cost-container">
    <!-- 计算说明 -->
    <el-alert type="info" :closable="false" show-icon style="margin-bottom: 16px;">
      <template #title>
        成本估算基于 Prometheus 集群指标，仅供参考
      </template>
      <template #default>
        <div style="font-size: 12px; color: #909399; line-height: 1.8;">
          CPU = 核数 × $0.03/核/小时 × 24h &nbsp;|&nbsp;
          内存 = GB × $0.01/GB/小时 × 24h &nbsp;|&nbsp;
          存储 = PVC容量GB × $0.10/GB/月 ÷ 30 &nbsp;|&nbsp;
          网络 = 日流量GB × $0.01/GB
        </div>
      </template>
    </el-alert>

    <!-- 概览卡片 -->
    <el-row :gutter="20" class="overview-row">
      <el-col :span="6">
        <el-card class="overview-card">
          <div class="overview-label">总成本 (USD)</div>
          <div class="overview-value">${{ Number(overview.totalCost || 0).toFixed(2) }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="overview-card">
          <div class="overview-label">CPU成本</div>
          <div class="overview-value">${{ Number(overview.totalCpuCost || 0).toFixed(2) }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="overview-card">
          <div class="overview-label">内存成本</div>
          <div class="overview-value">${{ Number(overview.totalMemoryCost || 0).toFixed(2) }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="overview-card">
          <div class="overview-label">存储成本</div>
          <div class="overview-value">${{ Number(overview.totalStorageCost || 0).toFixed(2) }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>成本记录</span>
          <el-button size="small" type="primary" @click="syncCost">同步最新数据</el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" class="search-form">
        <el-form-item label="集群ID">
          <el-input v-model="searchForm.clusterId" placeholder="请输入集群ID" clearable />
        </el-form-item>
        <el-form-item label="项目ID">
          <el-input v-model="searchForm.projectId" placeholder="请输入项目ID" clearable />
        </el-form-item>
        <el-form-item label="应用ID">
          <el-input v-model="searchForm.appId" placeholder="请输入应用ID" clearable />
        </el-form-item>
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="clusterId" label="集群ID" width="100" />
        <el-table-column prop="namespace" label="命名空间" width="150" show-overflow-tooltip />
        <el-table-column prop="projectName" label="项目" width="120">
          <template #default="{ row }">
            {{ row.projectName || (row.projectId ? '项目-'+row.projectId : '-') }}
          </template>
        </el-table-column>
        <el-table-column prop="appName" label="应用" width="120">
          <template #default="{ row }">
            {{ row.appName || (row.appId ? '应用-'+row.appId : '-') }}
          </template>
        </el-table-column>
        <el-table-column label="日期" width="120">
          <template #default="{ row }">
            {{ formatDate(row.costDate) }}
          </template>
        </el-table-column>
        <el-table-column label="CPU成本" width="130">
          <template #default="{ row }">
            <el-tooltip content="CPU核数 × $0.03/核/小时 × 24小时" placement="top">
              <span>{{ formatCost(row.cpuCost) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="内存成本" width="130">
          <template #default="{ row }">
            <el-tooltip content="内存GB × $0.01/GB/小时 × 24小时" placement="top">
              <span>{{ formatCost(row.memoryCost) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="存储成本" width="130">
          <template #default="{ row }">
            <el-tooltip content="PVC容量GB × $0.10/GB/月 ÷ 30天" placement="top">
              <span>{{ formatCost(row.storageCost) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="网络成本" width="130">
          <template #default="{ row }">
            <el-tooltip content="日流量GB × $0.01/GB" placement="top">
              <span>{{ formatCost(row.networkCost) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="总成本" width="120">
          <template #default="{ row }">
            <strong>{{ formatCost(row.totalCost) }}</strong>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadData"
        @current-change="loadData"
        class="pagination"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const tableData = ref([])
const overview = ref({})

const searchForm = reactive({
  clusterId: '',
  projectId: '',
  appId: '',
  dateRange: null
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const loadOverview = async () => {
  try {
    const res = await request.get('/costs/overview')
    if (res.code === 0) {
      overview.value = res.data || {}
    }
  } catch (error) {
    console.error('加载概览数据失败', error)
  }
}

const loadData = async () => {
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    if (searchForm.clusterId) {
      params.clusterId = searchForm.clusterId
    }
    if (searchForm.projectId) {
      params.projectId = searchForm.projectId
    }
    if (searchForm.appId) {
      params.appId = searchForm.appId
    }
    if (searchForm.dateRange && searchForm.dateRange.length === 2) {
      params.startDate = searchForm.dateRange[0]
      params.endDate = searchForm.dateRange[1]
    }
    const res = await request.get('/costs', { params })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载失败')
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const formatCost = (val) => {
  if (val == null) return '-'
  const num = Number(val)
  if (num === 0) return '$0'
  if (num < 0.01) return '$' + num.toFixed(4)  // 小于1分钱显示4位
  return '$' + num.toFixed(2)
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const cleaned = String(dateStr).replace('T', ' ').split('+')[0].split('.')[0]
  const parts = cleaned.split(' ')[0].split('-')
  if (parts.length === 3) return `${parts[0]}-${parts[1].padStart(2, '0')}-${parts[2].padStart(2, '0')}`
  return cleaned
}

const handleReset = () => {
  searchForm.clusterId = ''
  searchForm.projectId = ''
  searchForm.appId = ''
  searchForm.dateRange = null
  pagination.page = 1
  loadData()
}

const syncCost = async () => {
  try {
    const today = new Date().toISOString().split('T')[0]
    await request.post('/costs/sync', { clusterId: 1, costDate: today })
    ElMessage.success('同步完成')
    loadOverview()
    loadData()
  } catch (error) {
    ElMessage.error('同步失败: ' + (error.response?.data?.message || error.message))
  }
}

onMounted(() => {
  loadOverview()
  loadData()
})
</script>

<style scoped>
.cost-container {
  padding: 20px;
}

.overview-row {
  margin-bottom: 20px;
}

.overview-card {
  text-align: center;
}

.overview-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.overview-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
