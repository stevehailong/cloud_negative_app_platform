<template>
  <div class="environment-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>环境管理</span>
          <el-button type="primary" @click="handleCreate">新建环境</el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="请输入环境名称或编码" clearable />
        </el-form-item>
        <el-form-item label="环境类型">
          <el-select v-model="searchForm.envType" placeholder="请选择" clearable>
            <el-option label="开发环境" value="dev" />
            <el-option label="测试环境" value="test" />
            <el-option label="预发环境" value="staging" />
            <el-option label="生产环境" value="prod" />
            <el-option label="预览环境" value="preview" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属项目">
          <el-select v-model="searchForm.projectId" placeholder="请选择" clearable>
            <el-option 
              v-for="project in projects" 
              :key="project.id" 
              :label="project.projectName" 
              :value="project.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadEnvironments">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="envCode" label="环境编码" width="150" />
        <el-table-column prop="envName" label="环境名称" width="150" />
        <el-table-column prop="envType" label="环境类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getEnvTypeColor(row.envType)">
              {{ getEnvTypeName(row.envType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="150" />
        <el-table-column prop="clusterId" label="集群ID" width="100" />
        <el-table-column prop="projectId" label="项目ID" width="100" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="handleViewBindings(row)">绑定应用</el-button>
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
        @size-change="loadEnvironments"
        @current-change="loadEnvironments"
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
        <el-form-item label="环境编码" prop="envCode">
          <el-input v-model="form.envCode" placeholder="请输入环境编码，如：dev-001" />
        </el-form-item>
        <el-form-item label="环境名称" prop="envName">
          <el-input v-model="form.envName" placeholder="请输入环境名称" />
        </el-form-item>
        <el-form-item label="环境类型" prop="envType">
          <el-select v-model="form.envType" placeholder="请选择环境类型">
            <el-option label="开发环境" value="dev" />
            <el-option label="测试环境" value="test" />
            <el-option label="预发环境" value="staging" />
            <el-option label="生产环境" value="prod" />
            <el-option label="预览环境" value="preview" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属集群" prop="clusterId">
          <el-select v-model="form.clusterId" placeholder="请选择集群">
            <el-option 
              v-for="cluster in clusters" 
              :key="cluster.id" 
              :label="cluster.clusterName" 
              :value="cluster.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="命名空间" prop="namespace">
          <el-input v-model="form.namespace" placeholder="请输入K8s命名空间" />
        </el-form-item>
        <el-form-item label="所属项目" prop="projectId">
          <el-select v-model="form.projectId" placeholder="请选择项目">
            <el-option 
              v-for="project in projects" 
              :key="project.id" 
              :label="project.projectName" 
              :value="project.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
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
const projects = ref([])
const clusters = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新建环境')
const formRef = ref(null)

const searchForm = reactive({
  keyword: '',
  envType: '',
  projectId: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  id: null,
  envCode: '',
  envName: '',
  envType: '',
  clusterId: null,
  namespace: '',
  projectId: null,
  description: '',
  status: 1,
  configJson: ''
})

const rules = {
  envCode: [{ required: true, message: '请输入环境编码', trigger: 'blur' }],
  envName: [{ required: true, message: '请输入环境名称', trigger: 'blur' }],
  envType: [{ required: true, message: '请选择环境类型', trigger: 'change' }],
  clusterId: [{ required: true, message: '请选择集群', trigger: 'change' }],
  namespace: [{ required: true, message: '请输入命名空间', trigger: 'blur' }],
  projectId: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

const getEnvTypeName = (type) => {
  const map = {
    dev: '开发环境',
    test: '测试环境',
    staging: '预发环境',
    prod: '生产环境',
    preview: '预览环境'
  }
  return map[type] || type
}

const getEnvTypeColor = (type) => {
  const map = {
    dev: 'info',
    test: 'warning',
    staging: '',
    prod: 'danger',
    preview: 'success'
  }
  return map[type] || ''
}

const loadEnvironments = async () => {
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword
    }
    if (searchForm.projectId) {
      params.projectId = searchForm.projectId
    }
    const res = await request.get('/environments', { params })
    if (res.data.code === 0) {
      tableData.value = res.data.data.list || []
      pagination.total = res.data.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载环境列表失败')
  }
}

const loadProjects = async () => {
  try {
    const res = await request.get('/projects', { params: { page: 1, pageSize: 1000 } })
    if (res.data.code === 0) {
      projects.value = res.data.data.list || []
    }
  } catch (error) {
    console.error('加载项目列表失败', error)
  }
}

const loadClusters = async () => {
  try {
    const res = await request.get('/clusters', { params: { page: 1, pageSize: 1000 } })
    if (res.data.code === 0) {
      clusters.value = res.data.data.list || []
    }
  } catch (error) {
    console.error('加载集群列表失败', error)
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建环境'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑环境'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该环境吗？', '提示', {
      type: 'warning'
    })
    const res = await request.delete(`/environments/${row.id}`)
    if (res.data.code === 0) {
      ElMessage.success('删除成功')
      loadEnvironments()
    } else {
      ElMessage.error(res.data.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleViewBindings = (row) => {
  ElMessage.info('应用绑定功能开发中')
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const url = form.id ? `/environments/${form.id}` : '/environments'
    const method = form.id ? 'put' : 'post'
    const res = await request[method](url, form)
    if (res.data.code === 0) {
      ElMessage.success(form.id ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadEnvironments()
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } catch (error) {
    console.error('表单验证失败', error)
  }
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.envType = ''
  searchForm.projectId = ''
  pagination.page = 1
  loadEnvironments()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetForm = () => {
  form.id = null
  form.envCode = ''
  form.envName = ''
  form.envType = ''
  form.clusterId = null
  form.namespace = ''
  form.projectId = null
  form.description = ''
  form.status = 1
  form.configJson = ''
}

onMounted(() => {
  loadEnvironments()
  loadProjects()
  loadClusters()
})
</script>

<style scoped>
.environment-container {
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
