<template>
  <div class="project-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>项目管理</span>
          <el-button type="primary" @click="handleCreate">新建项目</el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="请输入项目名称或编码" clearable />
        </el-form-item>
        <el-form-item label="所属租户">
          <el-select v-model="searchForm.tenantId" placeholder="请选择" clearable>
            <el-option 
              v-for="tenant in tenants" 
              :key="tenant.id" 
              :label="tenant.tenantName" 
              :value="tenant.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadProjects">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="projectCode" label="项目编码" width="150" />
        <el-table-column prop="projectName" label="项目名称" width="200" />
        <el-table-column prop="tenantId" label="租户ID" width="100" />
        <el-table-column prop="visibility" label="可见性" width="100">
          <template #default="{ row }">
            <el-tag :type="row.visibility === 'public' ? 'success' : 'info'">
              {{ row.visibility === 'public' ? '公开' : '私有' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="handleViewMembers(row)">成员</el-button>
            <el-button link type="primary" @click="handleViewApps(row)">应用</el-button>
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
        @size-change="loadProjects"
        @current-change="loadProjects"
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
        <el-form-item label="项目编码" prop="projectCode">
          <el-input v-model="form.projectCode" placeholder="请输入项目编码，如：my-project" />
        </el-form-item>
        <el-form-item label="项目名称" prop="projectName">
          <el-input v-model="form.projectName" placeholder="请输入项目名称" />
        </el-form-item>
        <el-form-item label="所属租户" prop="tenantId">
          <el-select v-model="form.tenantId" placeholder="请选择租户" style="width: 100%">
            <el-option 
              v-for="tenant in tenants" 
              :key="tenant.id" 
              :label="tenant.tenantName" 
              :value="tenant.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="可见性" prop="visibility">
          <el-radio-group v-model="form.visibility">
            <el-radio value="private">私有</el-radio>
            <el-radio value="public">公开</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入项目描述" />
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

    <!-- 项目成员管理对话框 -->
    <el-dialog
      v-model="membersDialogVisible"
      :title="`${currentProject?.projectName} - 成员管理`"
      width="800px"
    >
      <el-card class="member-add-card" shadow="never">
        <template #header>
          <span>添加成员</span>
        </template>
        <el-form :inline="true">
          <el-form-item label="选择用户">
            <el-select v-model="memberForm.userId" placeholder="请选择用户" filterable style="width: 200px">
              <el-option
                v-for="user in users"
                :key="user.id"
                :label="`${user.nickname || user.username}`"
                :value="user.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="角色">
            <el-select v-model="memberForm.roleCode" style="width: 150px">
              <el-option label="负责人" value="owner" />
              <el-option label="维护者" value="maintainer" />
              <el-option label="开发者" value="developer" />
              <el-option label="访客" value="reporter" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleAddMember">添加</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-table :data="members" border stripe style="margin-top: 20px">
        <el-table-column prop="userId" label="用户ID" width="80" />
        <el-table-column label="用户名" width="150">
          <template #default="{ row }">
            {{ users.find(u => u.id === row.userId)?.username || row.userId }}
          </template>
        </el-table-column>
        <el-table-column label="昵称" width="150">
          <template #default="{ row }">
            {{ users.find(u => u.id === row.userId)?.nickname || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag 
              :type="row.roleCode === 'owner' ? 'danger' : row.roleCode === 'maintainer' ? 'warning' : row.roleCode === 'developer' ? 'primary' : 'info'"
            >
              {{ row.roleCode === 'owner' ? '负责人' : row.roleCode === 'maintainer' ? '维护者' : row.roleCode === 'developer' ? '开发者' : '访客' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="加入时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" @click="handleRemoveMember(row.userId)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 项目应用对话框 -->
    <el-dialog
      v-model="appsDialogVisible"
      :title="`${currentProject?.projectName} - 应用列表`"
      width="900px"
    >
      <div style="margin-bottom: 15px;">
        <el-button 
          type="primary" 
          @click="$router.push('/applications?projectId=' + currentProject?.id)"
        >
          前往应用管理
        </el-button>
        <span style="margin-left: 15px; color: #909399;">
          共 {{ projectApps.length }} 个应用
        </span>
      </div>

      <el-table :data="projectApps" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="code" label="应用编码" width="150" />
        <el-table-column prop="name" label="应用名称" width="150" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag 
              :type="row.type === 'web' ? 'primary' : row.type === 'api' ? 'success' : row.type === 'job' ? 'warning' : 'info'"
            >
              {{ row.type === 'web' ? 'Web应用' : row.type === 'api' ? 'API服务' : row.type === 'job' ? '定时任务' : row.type === 'function' ? '函数服务' : row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="language" label="语言" width="100" />
        <el-table-column prop="framework" label="框架" width="120" />
        <el-table-column prop="owner" label="负责人" width="100" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import { formatTime as formatDate } from '@/utils/time'

const tableData = ref([])
const tenants = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新建项目')
const formRef = ref(null)

const searchForm = reactive({
  keyword: '',
  tenantId: null
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  id: null,
  projectCode: '',
  projectName: '',
  tenantId: null,
  visibility: 'private',
  description: '',
  status: 1
})

const rules = {
  projectCode: [{ required: true, message: '请输入项目编码', trigger: 'blur' }],
  projectName: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
  tenantId: [{ required: true, message: '请选择租户', trigger: 'change' }]
}

const loadProjects = async () => {
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword
    }
    if (searchForm.tenantId) {
      params.tenantId = searchForm.tenantId
    }
    const res = await request.get('/projects', { params })
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error('加载项目列表失败', error)
  }
}

const loadTenants = async () => {
  try {
    const res = await request.get('/tenants', { params: { page: 1, pageSize: 1000 } })
    tenants.value = res.data.list || []
  } catch (error) {
    console.error('加载租户列表失败', error)
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建项目'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑项目'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该项目吗？删除后项目下的所有数据将无法恢复！', '警告', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
    await request.delete(`/projects/${row.id}`)
    ElMessage.success('删除成功')
    loadProjects()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败', error)
    }
  }
}

const membersDialogVisible = ref(false)
const members = ref([])
const currentProject = ref(null)
const users = ref([])
const memberForm = reactive({
  userId: null,
  roleCode: 'developer'
})

const handleViewMembers = async (row) => {
  currentProject.value = row
  await loadUsers()
  await loadProjectMembers(row.id)
  membersDialogVisible.value = true
}

const loadUsers = async () => {
  try {
    const res = await request.get('/users', { params: { page: 1, pageSize: 1000 } })
    users.value = res.data.list || []
  } catch (error) {
    console.error('加载用户列表失败', error)
  }
}

const loadProjectMembers = async (projectId) => {
  try {
    const res = await request.get(`/project-members/${projectId}`)
    members.value = res.data || []
  } catch (error) {
    console.error('加载项目成员失败', error)
  }
}

const handleAddMember = async () => {
  if (!memberForm.userId) {
    ElMessage.warning('请选择用户')
    return
  }
  try {
    await request.post(`/project-members/${currentProject.value.id}`, memberForm)
    ElMessage.success('添加成员成功')
    await loadProjectMembers(currentProject.value.id)
    memberForm.userId = null
    memberForm.roleCode = 'developer'
  } catch (error) {
    console.error('添加成员失败', error)
  }
}

const handleRemoveMember = async (userId) => {
  try {
    await ElMessageBox.confirm('确定要移除该成员吗?', '提示', {
      type: 'warning'
    })
    await request.delete(`/project-members/${currentProject.value.id}/${userId}`)
    ElMessage.success('移除成员成功')
    await loadProjectMembers(currentProject.value.id)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('移除成员失败', error)
    }
  }
}

const appsDialogVisible = ref(false)
const projectApps = ref([])

const handleViewApps = async (row) => {
  currentProject.value = row
  await loadProjectApps(row.id)
  appsDialogVisible.value = true
}

const loadProjectApps = async (projectId) => {
  try {
    const res = await request.get('/applications', { 
      params: { projectId, page: 1, pageSize: 1000 } 
    })
    projectApps.value = res.data.list || []
  } catch (error) {
    console.error('加载项目应用失败', error)
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const url = form.id ? `/projects/${form.id}` : '/projects'
    const method = form.id ? 'put' : 'post'
    await request[method](url, form)
    ElMessage.success(form.id ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadProjects()
  } catch (error) {
    if (error.message) {
      console.error('操作失败', error)
    } else {
      console.error('表单验证失败', error)
    }
  }
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.tenantId = null
  pagination.page = 1
  loadProjects()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetForm = () => {
  form.id = null
  form.projectCode = ''
  form.projectName = ''
  form.tenantId = null
  form.visibility = 'private'
  form.description = ''
  form.status = 1
}

onMounted(() => {
  loadProjects()
  loadTenants()
})
</script>

<style scoped>
.project-container {
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

.member-add-card {
  margin-bottom: 20px;
}

.member-add-card :deep(.el-card__header) {
  padding: 12px 20px;
}

.member-add-card :deep(.el-card__body) {
  padding: 15px 20px;
}
</style>
