<template>
  <div class="pipeline-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2>流水线管理</h2>
        <span class="subtitle">CI/CD流水线配置与执行</span>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新建流水线
        </el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <el-card class="filter-card" shadow="never">
      <el-form :inline="true" :model="queryParams">
        <el-form-item label="流水线名称">
          <el-input
            v-model="queryParams.pipeline_name"
            placeholder="请输入流水线名称"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 流水线列表 -->
    <el-card class="table-card" shadow="never">
      <el-table
        v-loading="loading"
        :data="pipelineList"
        style="width: 100%"
      >
        <el-table-column prop="pipelineName" label="流水线名称" min-width="200" />
        <el-table-column prop="pipelineCode" label="流水线编码" width="180" />
        <el-table-column label="关联应用" width="150">
          <template #default="{ row }">
            {{ getAppName(row.appId) }}
          </template>
        </el-table-column>
        <el-table-column prop="pipelineType" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.pipelineType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ciTool" label="CI工具" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ row.ciTool }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'" size="small">
              {{ row.enabled === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="360" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewRuns(row)">运行记录</el-button>
            <el-button type="primary" link size="small" @click="handleTrigger(row)">执行</el-button>
            <el-button type="success" link size="small" @click="handleDeploy(row)">部署上线</el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleQuery"
          @current-change="handleQuery"
        />
      </div>
    </el-card>

    <!-- 新建/编辑流水线对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="流水线名称" prop="pipelineName">
          <el-input v-model="formData.pipelineName" placeholder="请输入流水线名称" />
        </el-form-item>
        <el-form-item label="流水线编码" prop="pipelineCode">
          <el-input
            v-model="formData.pipelineCode"
            placeholder="请输入流水线编码，如 user-center-ci"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="关联应用" prop="appId">
          <el-select
            v-model="formData.appId"
            filterable
            placeholder="请选择关联应用"
            style="width: 100%"
            @change="handleAppChange"
          >
            <el-option
              v-for="app in applicationList"
              :key="app.id"
              :label="`${app.id} - ${app.name}`"
              :value="app.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="流水线类型" prop="pipelineType">
          <el-select v-model="formData.pipelineType" placeholder="请选择类型" style="width: 100%">
            <el-option label="CI（构建）" value="ci" />
            <el-option label="CD（部署）" value="cd" />
            <el-option label="CI/CD（完整）" value="ci_cd" />
          </el-select>
        </el-form-item>
        <el-form-item label="CI工具" prop="ciTool">
          <el-select v-model="formData.ciTool" placeholder="请选择CI工具" style="width: 100%">
            <el-option label="Jenkins" value="jenkins" />
            <el-option label="Tekton" value="tekton" />
          </el-select>
        </el-form-item>
        <el-form-item label="Git仓库" prop="repoUrl">
          <el-select
            v-model="formData.repoUrl"
            filterable
            remote
            :remote-method="searchGitlabProjects"
            :loading="loadingProjects"
            placeholder="搜索GitLab项目"
            style="width: 100%"
            @change="handleProjectChange"
          >
            <el-option
              v-for="p in gitlabProjects"
              :key="p.id"
              :label="p.name_with_namespace"
              :value="p.http_url_to_repo"
            />
          </el-select>
          <div v-if="!gitlabAvailable" class="form-tip warn">GitLab未配置，请手动输入仓库地址或在系统设置中配置GitLab</div>
        </el-form-item>
        <el-form-item label="默认分支" prop="defaultBranch">
          <el-select
            v-model="formData.defaultBranch"
            filterable
            :loading="loadingBranches"
            placeholder="选择分支"
            style="width: 100%"
            :disabled="!selectedProjectId"
          >
            <el-option
              v-for="b in gitlabBranches"
              :key="b.name"
              :label="b.name"
              :value="b.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch
            v-model="formData.enabled"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- 运行记录抽屉 -->
    <el-drawer
      v-model="runsDrawerVisible"
      :title="`运行记录 - ${currentPipeline?.pipelineName || ''}`"
      size="70%"
      direction="rtl"
    >
      <PipelineRuns v-if="runsDrawerVisible && currentPipeline" :pipeline-id="currentPipeline.id" />
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPipelineList, createPipeline, updatePipeline, deletePipeline, triggerPipeline, deployPipeline } from '@/api/pipeline'
import { getApplicationList } from '@/api/application'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'
import PipelineRuns from './components/PipelineRuns.vue'

// 查询参数
const queryParams = reactive({
  page: 1,
  pageSize: 10,
  pipeline_name: ''
})

// 数据
const loading = ref(false)
const pipelineList = ref([])
const total = ref(0)
const applicationList = ref([])

// 获取应用列表
const fetchApplicationList = async () => {
  try {
    const { data } = await getApplicationList({ page: 1, pageSize: 100 })
    applicationList.value = data.list || []
  } catch (error) {
    console.error('获取应用列表失败:', error)
  }
}

// 根据appId获取应用名称
const getAppName = (appId) => {
  const app = applicationList.value.find(a => a.id === appId)
  return app ? app.name : (appId ? `#${appId}` : '-')
}

// 选中关联应用后自动填充仓库信息
const handleAppChange = (appId) => {
  const app = applicationList.value.find(a => a.id === appId)
  if (app && app.repoUrl) {
    formData.repoUrl = app.repoUrl
    if (app.repoBranch) {
      formData.defaultBranch = app.repoBranch
    }
    // 尝试搜索GitLab项目以加载分支列表
    if (gitlabAvailable.value) {
      const projectName = app.repoUrl.split('/').pop()?.replace('.git', '') || ''
      if (projectName) {
        searchGitlabProjects(projectName)
      }
    }
  }
}

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('新建流水线')
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)
const runsDrawerVisible = ref(false)
const currentPipeline = ref(null)

const defaultFormData = {
  pipelineName: '',
  pipelineCode: '',
  appId: undefined,
  pipelineType: 'ci',
  ciTool: 'jenkins',
  repoUrl: '',
  defaultBranch: 'main',
  enabled: 1
}

const formData = reactive({ ...defaultFormData })

const formRules = {
  pipelineName: [{ required: true, message: '请输入流水线名称', trigger: 'blur' }],
  pipelineCode: [{ required: true, message: '请输入流水线编码', trigger: 'blur' }],
  appId: [{ required: true, message: '请输入关联应用ID', trigger: 'blur' }],
  pipelineType: [{ required: true, message: '请选择流水线类型', trigger: 'change' }],
  ciTool: [{ required: true, message: '请选择CI工具', trigger: 'change' }]
}

let editingId = null

// GitLab相关状态
const gitlabAvailable = ref(true)
const gitlabProjects = ref([])
const gitlabBranches = ref([])
const loadingProjects = ref(false)
const loadingBranches = ref(false)
const selectedProjectId = ref(null)

// 搜索GitLab项目
const searchGitlabProjects = async (query) => {
  if (!query || query.length < 2) return
  loadingProjects.value = true
  try {
    const res = await request({ url: '/gitlab/projects', method: 'get', params: { search: query } })
    gitlabProjects.value = res.data || []
    gitlabAvailable.value = true
  } catch (error) {
    gitlabAvailable.value = false
    gitlabProjects.value = []
  } finally {
    loadingProjects.value = false
  }
}

// 选择项目后加载分支
const handleProjectChange = (repoUrl) => {
  const project = gitlabProjects.value.find(p => p.http_url_to_repo === repoUrl)
  if (project) {
    selectedProjectId.value = project.id
    formData.defaultBranch = project.default_branch || 'main'
    loadBranches(project.id)
  }
}

// 加载分支
const loadBranches = async (projectId) => {
  loadingBranches.value = true
  try {
    const res = await request({ url: `/gitlab/projects/${projectId}/branches`, method: 'get' })
    gitlabBranches.value = res.data || []
  } catch (error) {
    gitlabBranches.value = []
  } finally {
    loadingBranches.value = false
  }
}

// 检查GitLab是否可用
const checkGitlabAvailability = async () => {
  try {
    await request({ url: '/gitlab/test', method: 'post' })
    gitlabAvailable.value = true
  } catch {
    gitlabAvailable.value = false
  }
}

// 获取流水线列表
const fetchPipelineList = async () => {
  loading.value = true
  try {
    const { data } = await getPipelineList(queryParams)
    pipelineList.value = data.list || []
    total.value = data.total || 0
  } catch (error) {
    console.error('获取流水线列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 查询
const handleQuery = () => {
  queryParams.page = 1
  fetchPipelineList()
}

// 重置
const handleReset = () => {
  queryParams.pipeline_name = ''
  queryParams.page = 1
  fetchPipelineList()
}

// 新建
const handleCreate = () => {
  isEdit.value = false
  dialogTitle.value = '新建流水线'
  editingId = null
  Object.assign(formData, defaultFormData)
  dialogVisible.value = true
}

// 编辑
const handleEdit = (row) => {
  isEdit.value = true
  dialogTitle.value = '编辑流水线'
  editingId = row.id

  // 解析configJson获取repoUrl和defaultBranch
  let repoUrl = ''
  let defaultBranch = 'main'
  if (row.configJson) {
    try {
      const config = JSON.parse(row.configJson)
      repoUrl = config.repoUrl || ''
      defaultBranch = config.defaultBranch || 'main'
    } catch (e) {}
  }

  Object.assign(formData, {
    pipelineName: row.pipelineName,
    pipelineCode: row.pipelineCode,
    appId: row.appId,
    pipelineType: row.pipelineType,
    ciTool: row.ciTool,
    repoUrl: repoUrl,
    defaultBranch: defaultBranch,
    enabled: row.enabled
  })

  // 如果有repoUrl，搜索项目并加载分支
  if (repoUrl && gitlabAvailable.value) {
    searchGitlabProjects(repoUrl.split('/').pop()?.replace('.git', '') || '')
  }

  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate()
  submitLoading.value = true
  try {
    // 将repoUrl和defaultBranch打入configJson
    const submitData = {
      pipelineName: formData.pipelineName,
      pipelineCode: formData.pipelineCode,
      appId: formData.appId,
      pipelineType: formData.pipelineType,
      ciTool: formData.ciTool,
      enabled: formData.enabled,
      configJson: JSON.stringify({
        repoUrl: formData.repoUrl,
        defaultBranch: formData.defaultBranch
      })
    }
    if (isEdit.value) {
      await updatePipeline(editingId, submitData)
      ElMessage.success('更新成功')
    } else {
      await createPipeline(submitData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchPipelineList()
  } catch (error) {
    console.error('提交失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// 删除
const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除流水线「${row.pipelineName}」吗？`, '确认删除', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deletePipeline(row.id)
      ElMessage.success('删除成功')
      fetchPipelineList()
    } catch (error) {
      console.error('删除失败:', error)
    }
  }).catch(() => {})
}

// 触发执行
// 查看运行记录
const handleViewRuns = (row) => {
  currentPipeline.value = row
  runsDrawerVisible.value = true
}

const handleTrigger = (row) => {
  ElMessageBox.confirm(`确定要执行流水线「${row.pipelineName}」吗？`, '确认执行', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'info'
  }).then(async () => {
    try {
      await triggerPipeline(row.id, { triggerType: 'manual' })
      ElMessage.success('流水线已触发执行')
      fetchPipelineList()
    } catch (error) {
      ElMessage.error('触发执行失败: ' + (error.message || '未知错误'))
    }
  }).catch(() => {})
}

// 手动部署上线（创建发布工单）
const handleDeploy = (row) => {
  ElMessageBox.confirm(
    `将为流水线「${row.pipelineName}」的最新构建制品创建发布工单，您可以在发布管理中选择部署策略(滚动发布/金丝雀发布/蓝绿发布)并审批执行`,
    '创建发布工单',
    {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      const res = await deployPipeline(row.id)
      ElMessage.success(res.data?.message || '发布工单已创建，请前往发布管理选择部署策略并审批')
    } catch (error) {
      ElMessage.error('创建发布工单失败: ' + (error.message || '未知错误'))
    }
  }).catch(() => {})
}

// 初始化
onMounted(() => {
  fetchPipelineList()
  fetchApplicationList()
  checkGitlabAvailability()
})
</script>

<style scoped lang="scss">
.pipeline-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  
  .header-left {
    h2 {
      margin: 0 0 8px 0;
      font-size: 24px;
      font-weight: 500;
      color: #303133;
    }
    
    .subtitle {
      font-size: 14px;
      color: #909399;
    }
  }
}

.filter-card {
  margin-bottom: 20px;
  
  :deep(.el-card__body) {
    padding: 16px;
  }
}

.table-card {
  :deep(.el-card__body) {
    padding: 0;
  }
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  padding: 20px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  &.warn {
    color: #e6a23c;
  }
}
</style>
