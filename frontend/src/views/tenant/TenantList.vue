<template>
  <div class="tenant-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>租户管理</span>
          <el-button type="primary" @click="handleCreate">新建租户</el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="请输入租户名称或编码" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadTenants">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantCode" label="租户编码" width="150" />
        <el-table-column prop="tenantName" label="租户名称" width="200" />
        <el-table-column prop="contactEmail" label="联系邮箱" width="200" show-overflow-tooltip />
        <el-table-column prop="contactPhone" label="联系电话" width="150" />
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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="handleViewProjects(row)">项目</el-button>
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
        @size-change="loadTenants"
        @current-change="loadTenants"
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
        <el-form-item label="租户编码" prop="tenantCode">
          <el-input 
            v-model="form.tenantCode" 
            placeholder="请输入租户编码，如：company-a" 
            :disabled="!!form.id"
          />
          <div class="form-tip">租户编码创建后不可修改</div>
        </el-form-item>
        <el-form-item label="租户名称" prop="tenantName">
          <el-input v-model="form.tenantName" placeholder="请输入租户名称" />
        </el-form-item>
        <el-form-item label="联系邮箱" prop="contactEmail">
          <el-input v-model="form.contactEmail" placeholder="请输入联系邮箱" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contactPhone">
          <el-input v-model="form.contactPhone" placeholder="请输入联系电话" />
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

    <!-- 租户项目列表对话框 -->
    <el-dialog v-model="projectsDialogVisible" title="租户项目" width="800px">
      <el-table :data="projects" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="projectCode" label="项目编码" width="150" />
        <el-table-column prop="projectName" label="项目名称" width="200" />
        <el-table-column prop="visibility" label="可见性" width="100">
          <template #default="{ row }">
            <el-tag :type="row.visibility === 'public' ? 'success' : 'info'">
              {{ row.visibility === 'public' ? '公开' : '私有' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
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
const projects = ref([])
const dialogVisible = ref(false)
const projectsDialogVisible = ref(false)
const dialogTitle = ref('新建租户')
const formRef = ref(null)

const searchForm = reactive({
  keyword: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  id: null,
  tenantCode: '',
  tenantName: '',
  contactEmail: '',
  contactPhone: '',
  status: 1
})

const rules = {
  tenantCode: [
    { required: true, message: '请输入租户编码', trigger: 'blur' },
    { pattern: /^[a-z0-9-]+$/, message: '租户编码只能包含小写字母、数字和连字符', trigger: 'blur' }
  ],
  tenantName: [{ required: true, message: '请输入租户名称', trigger: 'blur' }],
  contactEmail: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ]
}

const loadTenants = async () => {
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword
    }
    const res = await request.get('/tenants', { params })
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error('加载租户列表失败', error)
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建租户'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑租户'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除该租户吗？删除后租户下的所有项目和数据将无法恢复！', 
      '警告', 
      {
        type: 'warning',
        confirmButtonText: '确定删除',
        cancelButtonText: '取消'
      }
    )
    await request.delete(`/tenants/${row.id}`)
    ElMessage.success('删除成功')
    loadTenants()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败', error)
    }
  }
}

const handleViewProjects = async (row) => {
  try {
    const res = await request.get('/projects', { 
      params: { tenantId: row.id, page: 1, pageSize: 1000 } 
    })
    projects.value = res.data.list || []
    projectsDialogVisible.value = true
  } catch (error) {
    console.error('加载项目列表失败', error)
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const url = form.id ? `/tenants/${form.id}` : '/tenants'
    const method = form.id ? 'put' : 'post'
    await request[method](url, form)
    ElMessage.success(form.id ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadTenants()
  } catch (error) {
    if (error.message) {
      // API错误已经在拦截器中处理
      console.error('操作失败', error)
    } else {
      // 表单验证失败
      console.error('表单验证失败', error)
    }
  }
}

const handleReset = () => {
  searchForm.keyword = ''
  pagination.page = 1
  loadTenants()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetForm = () => {
  form.id = null
  form.tenantCode = ''
  form.tenantName = ''
  form.contactEmail = ''
  form.contactPhone = ''
  form.status = 1
}

onMounted(() => {
  loadTenants()
})
</script>

<style scoped>
.tenant-container {
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

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
