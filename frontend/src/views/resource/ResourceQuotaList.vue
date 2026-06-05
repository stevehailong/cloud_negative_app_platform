<template>
  <div class="resource-quota-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>资源配额管理</span>
          <el-button type="primary" @click="handleCreate">新建配额</el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" class="search-form">
        <el-form-item label="作用范围">
          <el-select v-model="searchForm.scopeType" placeholder="请选择" clearable>
            <el-option label="租户" value="tenant" />
            <el-option label="项目" value="project" />
            <el-option label="环境" value="env" />
            <el-option label="命名空间" value="namespace" />
            <el-option label="应用" value="app" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="请输入Scope ID或关键字" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="scopeType" label="作用范围" width="120">
          <template #default="{ row }">
            <el-tag>{{ getScopeTypeLabel(row.scopeType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="scopeId" label="Scope ID" width="120" />
        <el-table-column prop="cpuLimit" label="CPU限制" width="120">
          <template #default="{ row }">
            {{ row.cpuLimit || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="memoryLimit" label="内存限制" width="120">
          <template #default="{ row }">
            {{ row.memoryLimit || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="storageLimit" label="存储限制" width="120">
          <template #default="{ row }">
            {{ row.storageLimit || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="podLimit" label="Pod限制" width="100">
          <template #default="{ row }">
            {{ row.podLimit ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'Active' ? 'success' : 'info'">
              {{ row.status || 'Active' }}
            </el-tag>
          </template>
        </el-table-column>
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
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="作用范围" prop="scopeType">
          <el-select v-model="form.scopeType" placeholder="请选择作用范围">
            <el-option label="租户" value="tenant" />
            <el-option label="项目" value="project" />
            <el-option label="环境" value="env" />
            <el-option label="命名空间" value="namespace" />
            <el-option label="应用" value="app" />
          </el-select>
        </el-form-item>
        <el-form-item label="Scope ID" prop="scopeId">
          <el-input v-model="form.scopeId" placeholder="请输入Scope ID" />
        </el-form-item>
        <el-form-item label="CPU限制" prop="cpuLimit">
          <el-input v-model="form.cpuLimit" placeholder="如：4 或 4000m" />
        </el-form-item>
        <el-form-item label="内存限制" prop="memoryLimit">
          <el-input v-model="form.memoryLimit" placeholder="如：8Gi 或 8192Mi" />
        </el-form-item>
        <el-form-item label="存储限制" prop="storageLimit">
          <el-input v-model="form.storageLimit" placeholder="如：100Gi" />
        </el-form-item>
        <el-form-item label="Pod限制" prop="podLimit">
          <el-input v-model.number="form.podLimit" placeholder="如：100" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio value="Active">Active</el-radio>
            <el-radio value="Inactive">Inactive</el-radio>
          </el-radio-group>
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
const dialogTitle = ref('新建配额')
const formRef = ref(null)

const searchForm = reactive({
  scopeType: '',
  keyword: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  id: null,
  scopeType: 'tenant',
  scopeId: '',
  cpuLimit: '',
  memoryLimit: '',
  storageLimit: '',
  podLimit: null,
  status: 'Active'
})

const rules = {
  scopeType: [{ required: true, message: '请选择作用范围', trigger: 'change' }],
  scopeId: [{ required: true, message: '请输入Scope ID', trigger: 'blur' }]
}

const getScopeTypeLabel = (type) => {
  const labelMap = {
    tenant: '租户',
    project: '项目',
    env: '环境',
    namespace: '命名空间',
    app: '应用'
  }
  return labelMap[type] || type
}

const loadData = async () => {
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword
    }
    if (searchForm.scopeType) {
      params.scopeType = searchForm.scopeType
    }
    const res = await request.get('/resource-quotas', { params })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载失败')
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建配额'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑配额'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该资源配额吗？', '提示', {
      type: 'warning'
    })
    const res = await request.delete(`/resource-quotas/${row.id}`)
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
    const url = form.id ? `/resource-quotas/${form.id}` : '/resource-quotas'
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
  searchForm.scopeType = ''
  searchForm.keyword = ''
  pagination.page = 1
  loadData()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetForm = () => {
  form.id = null
  form.scopeType = 'tenant'
  form.scopeId = ''
  form.cpuLimit = ''
  form.memoryLimit = ''
  form.storageLimit = ''
  form.podLimit = null
  form.status = 'Active'
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.resource-quota-container {
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
