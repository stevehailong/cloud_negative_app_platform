<template>
  <div class="template-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>环境模板管理</span>
          <el-button type="primary" @click="handleCreate">新建模板</el-button>
        </div>
      </template>

      <!-- 搜索区域 -->
      <el-form :inline="true" class="search-form">
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="请输入模板名称或编码" clearable />
        </el-form-item>
        <el-form-item label="模板类型">
          <el-select v-model="searchForm.templateType" placeholder="请选择" clearable>
            <el-option label="Helm" value="helm" />
            <el-option label="Kustomize" value="kustomize" />
            <el-option label="YAML" value="yaml" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadTemplates">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table :data="tableData" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="templateCode" label="模板编码" width="150" />
        <el-table-column prop="templateName" label="模板名称" width="200" />
        <el-table-column prop="templateType" label="模板类型" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.templateType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="repoUrl" label="仓库地址" show-overflow-tooltip width="200" />
        <el-table-column prop="chartName" label="Chart名称" width="150" />
        <el-table-column prop="chartVersion" label="版本" width="100" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
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
        @size-change="loadTemplates"
        @current-change="loadTemplates"
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
        <el-form-item label="模板编码" prop="templateCode">
          <el-input v-model="form.templateCode" placeholder="请输入模板编码" />
        </el-form-item>
        <el-form-item label="模板名称" prop="templateName">
          <el-input v-model="form.templateName" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="模板类型" prop="templateType">
          <el-select v-model="form.templateType" placeholder="请选择模板类型">
            <el-option label="Helm" value="helm" />
            <el-option label="Kustomize" value="kustomize" />
            <el-option label="YAML" value="yaml" />
          </el-select>
        </el-form-item>
        <el-form-item label="仓库地址" prop="repoUrl">
          <el-input v-model="form.repoUrl" placeholder="如：https://charts.bitnami.com/bitnami" />
        </el-form-item>
        <el-form-item label="Chart名称" prop="chartName">
          <el-input v-model="form.chartName" placeholder="如：nginx" />
        </el-form-item>
        <el-form-item label="Chart版本" prop="chartVersion">
          <el-input v-model="form.chartVersion" placeholder="如：15.0.0" />
        </el-form-item>
        <el-form-item label="Values配置" prop="valuesYaml">
          <el-input 
            v-model="form.valuesYaml" 
            type="textarea" 
            :rows="10" 
            placeholder="请输入values.yaml内容" 
          />
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
const dialogVisible = ref(false)
const dialogTitle = ref('新建模板')
const formRef = ref(null)

const searchForm = reactive({
  keyword: '',
  templateType: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  id: null,
  templateCode: '',
  templateName: '',
  templateType: 'helm',
  repoUrl: '',
  chartName: '',
  chartVersion: '',
  valuesYaml: '',
  description: '',
  status: 1
})

const rules = {
  templateCode: [{ required: true, message: '请输入模板编码', trigger: 'blur' }],
  templateName: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  templateType: [{ required: true, message: '请选择模板类型', trigger: 'change' }]
}

const loadTemplates = async () => {
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword
    }
    if (searchForm.templateType) {
      params.templateType = searchForm.templateType
    }
    const res = await request.get('/env-templates', { params })
    if (res.data.code === 0) {
      tableData.value = res.data.data.list || []
      pagination.total = res.data.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载模板列表失败')
  }
}

const handleCreate = () => {
  dialogTitle.value = '新建模板'
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑模板'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该模板吗？', '提示', {
      type: 'warning'
    })
    const res = await request.delete(`/env-templates/${row.id}`)
    if (res.data.code === 0) {
      ElMessage.success('删除成功')
      loadTemplates()
    } else {
      ElMessage.error(res.data.message || '删除失败')
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
    const url = form.id ? `/env-templates/${form.id}` : '/env-templates'
    const method = form.id ? 'put' : 'post'
    const res = await request[method](url, form)
    if (res.data.code === 0) {
      ElMessage.success(form.id ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadTemplates()
    } else {
      ElMessage.error(res.data.message || '操作失败')
    }
  } catch (error) {
    console.error('表单验证失败', error)
  }
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.templateType = ''
  pagination.page = 1
  loadTemplates()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetForm = () => {
  form.id = null
  form.templateCode = ''
  form.templateName = ''
  form.templateType = 'helm'
  form.repoUrl = ''
  form.chartName = ''
  form.chartVersion = ''
  form.valuesYaml = ''
  form.description = ''
  form.status = 1
}

onMounted(() => {
  loadTemplates()
})
</script>

<style scoped>
.template-container {
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
