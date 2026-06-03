<template>
  <div class="application-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>应用管理</span>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            新建应用
          </el-button>
        </div>
      </template>
      
      <div class="search-bar">
        <el-select v-model="searchForm.projectId" placeholder="选择项目" clearable style="width: 200px; margin-right: 10px">
          <el-option
            v-for="project in projects"
            :key="project.id"
            :label="project.projectName"
            :value="project.id"
          />
        </el-select>
        <el-input
          v-model="searchForm.keyword"
          placeholder="搜索应用名称或编码"
          style="width: 300px"
          clearable
          @clear="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
      </div>
      
      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%; margin-top: 20px"
      >
        <el-table-column prop="name" label="应用名称" width="150" />
        <el-table-column prop="code" label="应用编码" width="150" />
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="language" label="语言" width="100" />
        <el-table-column prop="framework" label="框架" width="120" />
        <el-table-column prop="owner" label="负责人" width="100" />
        <el-table-column prop="createdBy" label="创建人" width="100" />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="200">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row.id)">
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
      width="650px"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="应用名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入应用名称" />
        </el-form-item>
        
        <el-form-item label="应用编码" prop="code">
          <el-input v-model="formData.code" placeholder="请输入应用编码" />
        </el-form-item>
        
        <el-form-item label="项目" prop="projectId">
          <el-select v-model="formData.projectId" placeholder="请选择项目" filterable>
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.projectName"
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="应用类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择应用类型">
            <el-option label="Web应用" value="web" />
            <el-option label="API服务" value="api" />
            <el-option label="定时任务" value="job" />
            <el-option label="函数服务" value="function" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="开发语言" prop="language">
          <el-select v-model="formData.language" placeholder="请选择开发语言">
            <el-option label="Java" value="java" />
            <el-option label="Go" value="go" />
            <el-option label="Python" value="python" />
            <el-option label="Node.js" value="nodejs" />
          </el-select>
        </el-form-item>
        
        <el-divider content-position="left">代码仓库信息</el-divider>

        <el-form-item label="代码仓库" prop="repoUrl">
          <el-input v-model="formData.repoUrl" placeholder="请输入Git仓库地址，如 https://gitlab.com/group/project.git">
            <template #prefix>
              <el-icon><Link /></el-icon>
            </template>
          </el-input>
          <div class="form-tip">应用关联的代码仓库地址，创建流水线时将自动填充</div>
        </el-form-item>

        <el-form-item label="默认分支">
          <el-input v-model="formData.repoBranch" placeholder="请输入默认分支，如 main 或 master" />
        </el-form-item>

        <el-form-item label="框架">
          <el-input v-model="formData.framework" placeholder="请输入开发框架，如 Spring Boot、Gin" />
        </el-form-item>

        <el-form-item label="构建工具">
          <el-input v-model="formData.buildTool" placeholder="请输入构建工具，如 maven、gradle、go build" />
        </el-form-item>

        <el-form-item label="Dockerfile">
          <el-input v-model="formData.dockerFile" placeholder="Dockerfile路径，如 ./Dockerfile" />
        </el-form-item>

        <el-divider content-position="left">其他信息</el-divider>

        <el-form-item label="负责人">
          <el-input v-model="formData.owner" placeholder="请输入负责人姓名" />
        </el-form-item>
        
        <el-form-item label="描述">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入应用描述"
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Link } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const router = useRouter()

const loading = ref(false)
const tableData = ref([])
const projects = ref([])
const searchForm = reactive({
  keyword: '',
  projectId: null
})
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const dialogVisible = ref(false)
const dialogTitle = ref('新建应用')
const formRef = ref(null)
const submitting = ref(false)
const formData = reactive({
  id: null,
  name: '',
  code: '',
  projectId: null,
  type: '',
  language: '',
  framework: '',
  repoUrl: '',
  repoBranch: '',
  buildTool: '',
  dockerFile: '',
  owner: '',
  description: ''
})

const rules = {
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入应用编码', trigger: 'blur' }],
  projectId: [{ required: true, message: '请选择项目', trigger: 'change' }],
  type: [{ required: true, message: '请选择应用类型', trigger: 'change' }],
  language: [{ required: true, message: '请选择开发语言', trigger: 'change' }],
  repoUrl: [{ required: true, message: '请输入代码仓库地址', trigger: 'blur' }]
}

const loadProjects = async () => {
  try {
    const res = await request.get('/projects', { params: { page: 1, pageSize: 1000 } })
    projects.value = res.data.list || []
  } catch (error) {
    console.error('加载项目列表失败', error)
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    if (searchForm.keyword) {
      params.keyword = searchForm.keyword
    }
    if (searchForm.projectId) {
      params.projectId = searchForm.projectId
    }
    const res = await request.get('/applications', { params })
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
  dialogTitle.value = '新建应用'
  Object.assign(formData, {
    id: null,
    name: '',
    code: '',
    projectId: null,
    type: '',
    language: '',
    framework: '',
    repoUrl: '',
    repoBranch: '',
    buildTool: '',
    dockerFile: '',
    owner: '',
    description: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑应用'
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    submitting.value = true
    if (formData.id) {
      await request.put(`/applications/${formData.id}`, formData)
      ElMessage.success('更新成功')
    } else {
      await request.post('/applications', formData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    const msg = error?.response?.data?.message || error?.message || '操作失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该应用吗？', '提示', {
      type: 'warning'
    })
    await request.delete(`/applications/${id}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败', error)
    }
  }
}

const viewDetail = (id) => {
  router.push(`/applications/${id}`)
}

onMounted(() => {
  loadProjects()
  fetchData()
})
</script>

<style scoped lang="scss">
.application-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .search-bar {
    display: flex;
    gap: 10px;
  }

  .form-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 4px;
  }
}
</style>
