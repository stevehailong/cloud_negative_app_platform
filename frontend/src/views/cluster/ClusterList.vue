<template>
  <div class="cluster-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>集群管理</span>
          <el-button type="primary" @click="handleCreate">新建集群</el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="请输入集群名称或编码" clearable />
        </el-form-item>
        <el-form-item label="集群类型">
          <el-select v-model="searchForm.clusterType" placeholder="请选择" clearable>
            <el-option label="Kubernetes" value="kubernetes" />
            <el-option label="Docker Swarm" value="docker-swarm" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadClusters">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="clusterCode" label="集群编码" width="150" />
        <el-table-column prop="clusterName" label="集群名称" width="200" />
        <el-table-column prop="clusterType" label="集群类型" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.clusterType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="apiServer" label="API Server" show-overflow-tooltip width="200" />
        <el-table-column prop="version" label="版本" width="120" />
        <el-table-column prop="region" label="区域" width="120" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '正常' : '异常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleViewMonitor(row)">监控</el-button>
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="handleViewNodes(row)">节点</el-button>
            <el-button link type="primary" @click="handleViewNamespaces(row)">命名空间</el-button>
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
        @size-change="loadClusters"
        @current-change="loadClusters"
        class="pagination"
      />
    </el-card>

    <!-- 新建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="700px"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="集群编码" prop="clusterCode">
          <el-input v-model="form.clusterCode" placeholder="请输入集群编码，如：k8s-prod-01" />
        </el-form-item>
        <el-form-item label="集群名称" prop="clusterName">
          <el-input v-model="form.clusterName" placeholder="请输入集群名称" />
        </el-form-item>
        <el-form-item label="集群类型" prop="clusterType">
          <el-select v-model="form.clusterType" placeholder="请选择集群类型">
            <el-option label="Kubernetes" value="kubernetes" />
            <el-option label="Docker Swarm" value="docker-swarm" />
          </el-select>
        </el-form-item>
        <el-form-item label="API Server" prop="apiServer">
          <el-input v-model="form.apiServer" placeholder="https://kubernetes.default.svc" />
        </el-form-item>
        <el-form-item label="Kubeconfig" prop="kubeconfig">
          <el-input 
            v-model="form.kubeconfig" 
            type="textarea" 
            :rows="6" 
            placeholder="请粘贴kubeconfig内容（可选）" 
          />
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="form.version" placeholder="如：v1.28.0" />
        </el-form-item>
        <el-form-item label="区域" prop="region">
          <el-input v-model="form.region" placeholder="如：beijing" />
        </el-form-item>
        <el-form-item label="可用区" prop="zone">
          <el-input v-model="form.zone" placeholder="如：zone-a" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">正常</el-radio>
            <el-radio :value="0">异常</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 节点列表对话框 -->
    <el-dialog v-model="nodesDialogVisible" title="集群节点" width="900px">
      <el-table :data="nodes" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="nodeName" label="节点名称" width="150" />
        <el-table-column prop="nodeIp" label="IP地址" width="150" />
        <el-table-column prop="nodeRole" label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="row.nodeRole === 'master' ? 'danger' : ''">
              {{ row.nodeRole }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cpuCores" label="CPU核数" width="100" />
        <el-table-column prop="memoryGb" label="内存(GB)" width="100" />
        <el-table-column prop="osImage" label="操作系统" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? 'Ready' : 'NotReady' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 命名空间列表对话框 -->
    <el-dialog v-model="namespacesDialogVisible" title="命名空间" width="800px">
      <el-table :data="namespaces" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="namespaceName" label="命名空间" width="150" />
        <el-table-column prop="projectId" label="项目ID" width="100" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'warning'">
              {{ row.status === 1 ? 'Active' : 'Terminating' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'

const router = useRouter()
const tableData = ref([])
const nodes = ref([])
const namespaces = ref([])
const dialogVisible = ref(false)
const nodesDialogVisible = ref(false)
const namespacesDialogVisible = ref(false)
const dialogTitle = ref('新建集群')
const formRef = ref(null)

const searchForm = reactive({
  keyword: '',
  clusterType: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  id: null,
  clusterCode: '',
  clusterName: '',
  clusterType: 'kubernetes',
  apiServer: '',
  kubeconfig: '',
  version: '',
  region: '',
  zone: '',
  description: '',
  status: 1
})

const rules = {
  clusterCode: [{ required: true, message: '请输入集群编码', trigger: 'blur' }],
  clusterName: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
  clusterType: [{ required: true, message: '请选择集群类型', trigger: 'change' }],
  apiServer: [{ required: true, message: '请输入API Server地址', trigger: 'blur' }]
}

const loadClusters = async () => {
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword
    }
    if (searchForm.clusterType) {
      params.clusterType = searchForm.clusterType
    }
    const res = await request.get('/clusters', { params })
    if (res.data.code === 0) {
      tableData.value = res.data.data.list || []
      pagination.total = res.data.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载集群列表失败')
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建集群'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑集群'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该集群吗？', '提示', {
      type: 'warning'
    })
    const res = await request.delete(`/clusters/${row.id}`)
    if (res.data.code === 0) {
      ElMessage.success('删除成功')
      loadClusters()
    } else {
      ElMessage.error(res.data.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleViewNodes = async (row) => {
  try {
    const res = await request.get('/nodes', { 
      params: { clusterId: row.id, page: 1, pageSize: 1000 } 
    })
    if (res.data.code === 0) {
      nodes.value = res.data.data.list || []
      nodesDialogVisible.value = true
    }
  } catch (error) {
    ElMessage.error('加载节点列表失败')
  }
}

const handleViewNamespaces = async (row) => {
  try {
    const res = await request.get('/namespaces', { 
      params: { clusterId: row.id, page: 1, pageSize: 1000 } 
    })
    if (res.data.code === 0) {
      namespaces.value = res.data.data.list || []
      namespacesDialogVisible.value = true
    }
  } catch (error) {
    ElMessage.error('加载命名空间列表失败')
  }
}

const handleViewMonitor = (row) => {
  router.push(`/clusters/${row.id}/monitor`)
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const url = form.id ? `/clusters/${form.id}` : '/clusters'
    const method = form.id ? 'put' : 'post'
    const res = await request[method](url, form)
    if (res.data.code === 0) {
      ElMessage.success(form.id ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadClusters()
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } catch (error) {
    console.error('表单验证失败', error)
  }
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.clusterType = ''
  pagination.page = 1
  loadClusters()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetForm = () => {
  form.id = null
  form.clusterCode = ''
  form.clusterName = ''
  form.clusterType = 'kubernetes'
  form.apiServer = ''
  form.kubeconfig = ''
  form.version = ''
  form.region = ''
  form.zone = ''
  form.description = ''
  form.status = 1
}

onMounted(() => {
  loadClusters()
})
</script>

<style scoped>
.cluster-container {
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
