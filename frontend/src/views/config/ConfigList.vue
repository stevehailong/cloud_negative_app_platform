<template>
  <div class="config-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>配置管理</span>
          <el-button type="primary" @click="handleCreate">新建配置</el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" class="search-form">
        <el-form-item label="应用ID">
          <el-input v-model="searchForm.appId" placeholder="请输入应用ID" clearable />
        </el-form-item>
        <el-form-item label="环境ID">
          <el-input v-model="searchForm.envId" placeholder="请输入环境ID" clearable />
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="请输入配置Key或描述" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="appId" label="应用ID" width="100" />
        <el-table-column prop="envId" label="环境ID" width="100" />
        <el-table-column prop="configKey" label="配置Key" width="200" show-overflow-tooltip />
        <el-table-column prop="configValue" label="配置值" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ truncateValue(row.configValue) }}
          </template>
        </el-table-column>
        <el-table-column prop="valueType" label="值类型" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.valueType || 'string' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="80" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <!-- 新建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="应用ID" prop="appId">
          <el-input v-model="form.appId" placeholder="请输入应用ID" />
        </el-form-item>
        <el-form-item label="环境ID" prop="envId">
          <el-input v-model="form.envId" placeholder="请输入环境ID" />
        </el-form-item>
        <el-form-item label="配置Key" prop="configKey">
          <el-input v-model="form.configKey" placeholder="请输入配置Key" />
        </el-form-item>
        <el-form-item label="配置值" prop="configValue">
          <el-input
            v-model="form.configValue"
            type="textarea"
            :rows="4"
            placeholder="请输入配置值"
          />
        </el-form-item>
        <el-form-item label="值类型" prop="valueType">
          <el-select v-model="form.valueType" placeholder="请选择值类型">
            <el-option label="字符串" value="string" />
            <el-option label="数字" value="number" />
            <el-option label="布尔" value="boolean" />
            <el-option label="JSON" value="json" />
            <el-option label="YAML" value="yaml" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="form.version" placeholder="如：v1.0.0" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新建配置')
const formRef = ref(null)

const searchForm = reactive({
  appId: '',
  envId: '',
  keyword: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  id: null,
  appId: '',
  envId: '',
  configKey: '',
  configValue: '',
  valueType: 'string',
  version: '',
  description: ''
})

const rules = {
  appId: [{ required: true, message: '请输入应用ID', trigger: 'blur' }],
  envId: [{ required: true, message: '请输入环境ID', trigger: 'blur' }],
  configKey: [{ required: true, message: '请输入配置Key', trigger: 'blur' }],
  configValue: [{ required: true, message: '请输入配置值', trigger: 'blur' }]
}

const truncateValue = (value) => {
  if (!value) return '-'
  return value.length > 100 ? value.substring(0, 100) + '...' : value
}

const loadData = async () => {
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword
    }
    if (searchForm.appId) {
      params.appId = searchForm.appId
    }
    if (searchForm.envId) {
      params.envId = searchForm.envId
    }
    const res = await request.get('/app-configs', { params })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载失败')
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建配置'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑配置'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该配置吗？', '提示', {
      type: 'warning'
    })
    const res = await request.delete(`/configs/${row.id}`)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      loadData()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const url = form.id ? `/configs/${form.id}` : '/configs'
    const method = form.id ? 'put' : 'post'
    const res = await request[method](url, form)
    if (res.code === 0) {
      ElMessage.success(form.id ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadData()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error) {
    console.error('表单验证失败', error)
  }
}

const handleReset = () => {
  searchForm.appId = ''
  searchForm.envId = ''
  searchForm.keyword = ''
  pagination.page = 1
  loadData()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetForm = () => {
  form.id = null
  form.appId = ''
  form.envId = ''
  form.configKey = ''
  form.configValue = ''
  form.valueType = 'string'
  form.version = ''
  form.description = ''
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.config-container {
  padding: 20px;
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
