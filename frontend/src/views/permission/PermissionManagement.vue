<template>
  <div class="permission-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>权限管理</span>
          <el-button type="primary" @click="handleCreate">创建权限</el-button>
        </div>
      </template>

      <el-form :inline="true" :model="queryForm" class="query-form">
        <el-form-item label="资源类型">
          <el-select v-model="queryForm.resourceType" placeholder="全部" clearable>
            <el-option label="应用" value="application"></el-option>
            <el-option label="组件" value="component"></el-option>
            <el-option label="用户" value="user"></el-option>
            <el-option label="角色" value="role"></el-option>
            <el-option label="权限" value="permission"></el-option>
            <el-option label="部署" value="deployment"></el-option>
            <el-option label="流水线" value="pipeline"></el-option>
            <el-option label="环境" value="environment"></el-option>
            <el-option label="集群" value="cluster"></el-option>
            <el-option label="项目" value="project"></el-option>
            <el-option label="监控" value="monitor"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadPermissions">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="permissions" border>
        <el-table-column prop="code" label="权限编码" width="200"></el-table-column>
        <el-table-column prop="name" label="权限名称" width="150"></el-table-column>
        <el-table-column prop="resourceType" label="资源类型" width="120"></el-table-column>
        <el-table-column prop="httpMethod" label="HTTP方法" width="120"></el-table-column>
        <el-table-column prop="path" label="API路径" min-width="250"></el-table-column>
        <el-table-column prop="description" label="描述" min-width="200"></el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="queryForm.page"
        v-model:page-size="queryForm.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadPermissions"
        @current-change="loadPermissions"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="权限编码" prop="code">
          <el-input v-model="form.code" :disabled="!!form.id" placeholder="如: app:view"></el-input>
        </el-form-item>
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="form.name" placeholder="如: 查看应用"></el-input>
        </el-form-item>
        <el-form-item label="资源类型">
          <el-select v-model="form.resourceType" placeholder="请选择">
            <el-option label="应用" value="application"></el-option>
            <el-option label="组件" value="component"></el-option>
            <el-option label="用户" value="user"></el-option>
            <el-option label="角色" value="role"></el-option>
            <el-option label="权限" value="permission"></el-option>
            <el-option label="部署" value="deployment"></el-option>
            <el-option label="流水线" value="pipeline"></el-option>
            <el-option label="环境" value="environment"></el-option>
            <el-option label="集群" value="cluster"></el-option>
            <el-option label="项目" value="project"></el-option>
            <el-option label="监控" value="monitor"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="HTTP方法">
          <el-input v-model="form.httpMethod" placeholder="如: GET 或 GET,POST"></el-input>
        </el-form-item>
        <el-form-item label="API路径">
          <el-input v-model="form.path" placeholder="如: /api/v1/applications/*"></el-input>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea"></el-input>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
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

const queryForm = reactive({
  page: 1,
  pageSize: 10,
  resourceType: ''
})

const permissions = ref([])
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('创建权限')
const formRef = ref(null)

const form = reactive({
  id: null,
  code: '',
  name: '',
  resourceType: '',
  httpMethod: '',
  path: '',
  description: '',
  status: 1
})

const rules = {
  code: [{ required: true, message: '请输入权限编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }]
}

const loadPermissions = async () => {
  try {
    const params = {
      page: queryForm.page,
      pageSize: queryForm.pageSize
    }
    if (queryForm.resourceType) {
      params.resourceType = queryForm.resourceType
    }
    const { data } = await request.get('/permissions/', { params })
    permissions.value = data.items || []
    total.value = data.total || 0
  } catch (error) {
    ElMessage.error('加载权限列表失败')
  }
}

const resetQuery = () => {
  queryForm.resourceType = ''
  queryForm.page = 1
  loadPermissions()
}

const handleCreate = () => {
  dialogTitle.value = '创建权限'
  Object.assign(form, {
    id: null,
    code: '',
    name: '',
    resourceType: '',
    httpMethod: '',
    path: '',
    description: '',
    status: 1
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑权限'
  Object.assign(form, { ...row })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      if (form.id) {
        await request.put(`/permissions/${form.id}/`, form)
        ElMessage.success('更新成功')
      } else {
        await request.post('/permissions/', form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadPermissions()
    } catch (error) {
      ElMessage.error(error.response?.data?.message || '操作失败')
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该权限吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await request.delete(`/permissions/${row.id}/`)
      ElMessage.success('删除成功')
      loadPermissions()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

onMounted(() => {
  loadPermissions()
})
</script>

<style scoped>
.permission-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.query-form {
  margin-bottom: 20px;
}

.el-pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
