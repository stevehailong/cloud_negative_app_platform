<template>
  <div class="config-map-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>ConfigMap 配置管理</span>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            新建ConfigMap
          </el-button>
        </div>
      </template>
      
      <div class="search-bar">
        <el-select v-model="searchForm.envId" placeholder="选择环境" clearable style="width: 200px; margin-right: 10px">
          <el-option
            v-for="env in environments"
            :key="env.id"
            :label="env.envName"
            :value="env.id"
          />
        </el-select>
        <el-input
          v-model="searchForm.namespace"
          placeholder="搜索命名空间"
          style="width: 200px"
          clearable
          @clear="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">搜索</el-button>
      </div>
      
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%; margin-top: 20px"
      >
        <el-table-column prop="name" label="名称" width="180" />
        <el-table-column prop="namespace" label="命名空间" width="150" />
        <el-table-column prop="envId" label="环境ID" width="100" />
        <el-table-column label="同步状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getSyncStatusType(row.syncStatus)">
              {{ getSyncStatusLabel(row.syncStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column label="最后同步时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.lastSyncTime) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="200">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">
              详情
            </el-button>
            <el-button link type="primary" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button link type="danger" @click="handleDelete(row.id)">
              删除
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
    
    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="700px"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入ConfigMap名称" />
        </el-form-item>
        
        <el-form-item label="环境" prop="envId">
          <el-select v-model="formData.envId" placeholder="请选择环境" filterable style="width: 100%">
            <el-option
              v-for="env in environments"
              :key="env.id"
              :label="env.envName"
              :value="env.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="命名空间" prop="namespace">
          <el-input v-model="formData.namespace" placeholder="如: default" />
        </el-form-item>
        
        <el-form-item label="配置数据" prop="data">
          <el-input
            v-model="formData.data"
            type="textarea"
            :rows="8"
            placeholder='JSON格式配置数据, 如: {"key1": "value1", "key2": "value2"}'
          />
        </el-form-item>
        
        <el-form-item label="描述">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="2"
            placeholder="请输入描述"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
    
    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailVisible"
      title="ConfigMap详情"
      width="800px"
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="名称" :span="2">{{ currentItem.name }}</el-descriptions-item>
        <el-descriptions-item label="命名空间">{{ currentItem.namespace }}</el-descriptions-item>
        <el-descriptions-item label="环境ID">{{ currentItem.envId }}</el-descriptions-item>
        <el-descriptions-item label="同步状态">
          <el-tag :type="getSyncStatusType(currentItem.syncStatus)">
            {{ getSyncStatusLabel(currentItem.syncStatus) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="最后同步时间">
          {{ formatTime(currentItem.lastSyncTime) }}
        </el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="配置数据" :span="2">
          <pre v-if="currentItem.data" style="margin: 0">{{ formatJSON(currentItem.data) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="同步消息" :span="2">
          {{ currentItem.syncMessage || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(currentItem.createTime) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(currentItem.updateTime) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const loading = ref(false)
const tableData = ref([])
const environments = ref([])
const searchForm = reactive({
  envId: null,
  namespace: ''
})
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const dialogVisible = ref(false)
const dialogTitle = ref('新建ConfigMap')
const formRef = ref(null)
const submitting = ref(false)
const formData = reactive({
  id: null,
  name: '',
  envId: null,
  namespace: '',
  data: '',
  description: ''
})

const detailVisible = ref(false)
const currentItem = ref({})

const rules = {
  name: [{ required: true, message: '请输入ConfigMap名称', trigger: 'blur' }],
  envId: [{ required: true, message: '请选择环境', trigger: 'change' }],
  namespace: [{ required: true, message: '请输入命名空间', trigger: 'blur' }],
  data: [
    { required: true, message: '请输入配置数据', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        try {
          JSON.parse(value)
          callback()
        } catch (e) {
          callback(new Error('配置数据必须是有效的JSON格式'))
        }
      },
      trigger: 'blur'
    }
  ]
}

const loadEnvironments = async () => {
  try {
    const res = await request.get('/environments', { params: { page: 1, pageSize: 1000 } })
    environments.value = res.data.list || []
  } catch (error) {
    console.error('加载环境列表失败', error)
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    if (searchForm.envId) {
      params.envId = searchForm.envId
    }
    if (searchForm.namespace) {
      params.namespace = searchForm.namespace
    }
    const res = await request.get('/config-maps', { params })
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error('获取数据失败', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchData()
}

const showCreateDialog = () => {
  dialogTitle.value = '新建ConfigMap'
  Object.assign(formData, {
    id: null,
    name: '',
    envId: null,
    namespace: '',
    data: '',
    description: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑ConfigMap'
  Object.assign(formData, {
    id: row.id,
    name: row.name,
    envId: row.envId,
    namespace: row.namespace,
    data: row.data,
    description: row.description
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    submitting.value = true
    if (formData.id) {
      await request.put(`/config-maps/${formData.id}`, formData)
      ElMessage.success('更新成功')
    } else {
      await request.post('/config-maps', formData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    if (error.message) {
      console.error('操作失败', error)
    }
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该ConfigMap吗？', '提示', {
      type: 'warning'
    })
    await request.delete(`/config-maps/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败', error)
    }
  }
}

const viewDetail = (row) => {
  currentItem.value = row
  detailVisible.value = true
}

const getSyncStatusType = (status) => {
  const typeMap = {
    pending: 'info',
    synced: 'success',
    failed: 'danger'
  }
  return typeMap[status] || 'info'
}

const getSyncStatusLabel = (status) => {
  const labelMap = {
    pending: '待同步',
    synced: '已同步',
    failed: '同步失败'
  }
  return labelMap[status] || status
}

const formatJSON = (str) => {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

onMounted(() => {
  loadEnvironments()
  fetchData()
})
</script>

<style scoped lang="scss">
.config-map-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .search-bar {
    display: flex;
    gap: 10px;
  }
  
  pre {
    background: #f5f7fa;
    padding: 10px;
    border-radius: 4px;
    font-size: 12px;
    line-height: 1.5;
    max-height: 400px;
    overflow: auto;
  }
}
</style>
