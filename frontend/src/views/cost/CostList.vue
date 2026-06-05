<template>
  <div class="cost-container">
    <!-- 概览卡片 -->
    <el-row :gutter="20" class="overview-row">
      <el-col :span="6">
        <el-card class="overview-card">
          <div class="overview-label">本月总成本</div>
          <div class="overview-value">{{ overview.totalCost || '0.00' }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="overview-card">
          <div class="overview-label">CPU成本</div>
          <div class="overview-value">{{ overview.cpuCost || '0.00' }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="overview-card">
          <div class="overview-label">内存成本</div>
          <div class="overview-value">{{ overview.memoryCost || '0.00' }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="overview-card">
          <div class="overview-label">存储成本</div>
          <div class="overview-value">{{ overview.storageCost || '0.00' }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>成本记录</span>
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
        <el-table-column prop="clusterId" label="集群ID" width="120" />
        <el-table-column prop="namespace" label="命名空间" width="150" show-overflow-tooltip />
        <el-table-column prop="projectId" label="项目ID" width="100" />
        <el-table-column prop="appId" label="应用ID" width="100" />
        <el-table-column prop="costDate" label="日期" width="120" />
        <el-table-column prop="cpuCost" label="CPU成本" width="120">
          <template #default="{ row }">
            {{ row.cpuCost ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="memoryCost" label="内存成本" width="120">
          <template #default="{ row }">
            {{ row.memoryCost ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="storageCost" label="存储成本" width="120">
          <template #default="{ row }">
            {{ row.storageCost ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="networkCost" label="网络成本" width="120">
          <template #default="{ row }">
            {{ row.networkCost ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="totalCost" label="总成本" width="120">
          <template #default="{ row }">
            <strong>{{ row.totalCost ?? '-' }}</strong>
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

const handleReset = () => {
  searchForm.clusterId = ''
  searchForm.projectId = ''
  searchForm.appId = ''
  searchForm.dateRange = null
  pagination.page = 1
  loadData()
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
