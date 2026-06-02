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
          <el-input 
            v-model="form.namespace" 
            placeholder="如: my-app-dev"
            @input="validateNamespaceFormat"
          >
            <template #append>
              <el-tooltip 
                content="命名空间是环境在K8s集群中的唯一标识，用于资源隔离"
                placement="top"
              >
                <el-icon><InfoFilled /></el-icon>
              </el-tooltip>
            </template>
          </el-input>
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            只能包含小写字母、数字和短横线(-)，必须以字母或数字开头和结尾，1-63字符
          </div>
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
        <el-form-item label="环境模板" prop="templateId">
          <el-select v-model="form.templateId" placeholder="请选择模板(可选)" clearable>
            <el-option 
              v-for="template in templates" 
              :key="template.id" 
              :label="template.templateName" 
              :value="template.id" 
            >
              <span style="float: left">{{ template.templateName }}</span>
              <span style="float: right; color: #8492a6; font-size: 13px">{{ template.templateType }}</span>
            </el-option>
          </el-select>
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            选择模板后将应用模板的默认配置
          </div>
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

    <!-- 绑定应用列表对话框 -->
    <el-dialog
      v-model="bindingsDialogVisible"
      title="绑定的应用列表"
      width="800px"
    >
      <el-alert
        v-if="currentEnv"
        :title="`环境：${currentEnv.envName} (${currentEnv.envType}) - ${currentEnv.namespace}`"
        type="info"
        :closable="false"
        style="margin-bottom: 20px;"
      />
      
      <el-table
        v-loading="bindingsLoading"
        :data="bindingsList"
        style="width: 100%"
      >
        <el-table-column prop="appId" label="应用ID" width="100" />
        <el-table-column prop="appName" label="应用名称" width="180" />
        <el-table-column label="资源配置" min-width="200">
          <template #default="{ row }">
            <div style="font-size: 12px;">
              <div>副本: {{ row.replicas }}</div>
              <div>CPU: {{ row.cpuRequest }} / {{ row.cpuLimit }}</div>
              <div>内存: {{ row.memoryRequest }} / {{ row.memoryLimit }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="配置状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.configStatus === 'ready' ? 'success' : 'warning'">
              {{ row.configStatus === 'ready' ? '已配置' : '待配置' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEditConfig(row)">
              编辑配置
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div v-if="bindingsList.length === 0 && !bindingsLoading" style="text-align: center; padding: 40px; color: #909399;">
        暂无应用绑定到此环境
      </div>
      
      <template #footer>
        <el-button @click="bindingsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 配置编辑对话框 -->
    <el-dialog
      v-model="configDialogVisible"
      title="编辑环境配置"
      width="900px"
      @close="handleConfigDialogClose"
    >
      <el-alert
        v-if="currentBinding"
        :title="`应用${currentBinding.appId} - ${currentEnv?.envName} (${currentEnv?.namespace})`"
        type="info"
        :closable="false"
        style="margin-bottom: 20px;"
      />

      <el-tabs v-model="activeConfigTab">
        <!-- 基础配置 -->
        <el-tab-pane label="基础配置" name="basic">
          <el-form :model="configForm" label-width="120px">
            <el-form-item label="副本数">
              <el-input-number v-model="configForm.replicas" :min="1" :max="10" />
            </el-form-item>
            <el-form-item label="CPU Request">
              <el-input v-model="configForm.cpuRequest" placeholder="如: 100m" style="width: 200px" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">
                推荐: 100m-500m
              </span>
            </el-form-item>
            <el-form-item label="CPU Limit">
              <el-input v-model="configForm.cpuLimit" placeholder="如: 500m" style="width: 200px" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">
                推荐: 500m-2000m
              </span>
            </el-form-item>
            <el-form-item label="Memory Request">
              <el-input v-model="configForm.memoryRequest" placeholder="如: 128Mi" style="width: 200px" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">
                推荐: 128Mi-512Mi
              </span>
            </el-form-item>
            <el-form-item label="Memory Limit">
              <el-input v-model="configForm.memoryLimit" placeholder="如: 512Mi" style="width: 200px" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">
                推荐: 512Mi-2Gi
              </span>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 环境变量 -->
        <el-tab-pane label="环境变量" name="env">
          <div style="margin-bottom: 15px;">
            <el-button size="small" @click="addEnvVar">添加环境变量</el-button>
            <el-button size="small" @click="loadEnvTemplate">使用模板</el-button>
          </div>
          <el-table :data="envVars" border style="width: 100%">
            <el-table-column label="变量名" width="200">
              <template #default="{ row, $index }">
                <el-input v-model="row.key" placeholder="如: APP_NAME" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="变量值">
              <template #default="{ row, $index }">
                <el-input v-model="row.value" placeholder="变量值" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="说明" width="180">
              <template #default="{ row, $index }">
                <el-input v-model="row.description" placeholder="可选" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" align="center">
              <template #default="{ row, $index }">
                <el-button link type="danger" size="small" @click="removeEnvVar($index)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 高级配置 -->
        <el-tab-pane label="高级配置" name="advanced">
          <el-form :model="advancedConfig" label-width="140px">
            <el-form-item label="健康检查">
              <el-switch v-model="advancedConfig.healthCheck.enabled" />
            </el-form-item>
            <el-form-item v-if="advancedConfig.healthCheck.enabled" label="检查路径">
              <el-input v-model="advancedConfig.healthCheck.path" placeholder="/health" style="width: 300px" />
            </el-form-item>
            <el-form-item v-if="advancedConfig.healthCheck.enabled" label="检查端口">
              <el-input-number v-model="advancedConfig.healthCheck.port" :min="1" :max="65535" />
            </el-form-item>
            
            <el-divider />
            
            <el-form-item label="Ingress">
              <el-switch v-model="advancedConfig.ingress.enabled" />
            </el-form-item>
            <el-form-item v-if="advancedConfig.ingress.enabled" label="域名">
              <el-input v-model="advancedConfig.ingress.host" placeholder="app.example.com" style="width: 400px" />
            </el-form-item>
            <el-form-item v-if="advancedConfig.ingress.enabled" label="路径">
              <el-input v-model="advancedConfig.ingress.path" placeholder="/" style="width: 200px" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- JSON编辑 -->
        <el-tab-pane label="JSON编辑" name="json">
          <div style="margin-bottom: 10px; color: #909399; font-size: 12px;">
            直接编辑完整的JSON配置（高级用户）
          </div>
          <el-input
            v-model="jsonConfig"
            type="textarea"
            :rows="15"
            placeholder="输入JSON配置"
            style="font-family: monospace;"
          />
          <div style="margin-top: 10px;">
            <el-button size="small" @click="formatJson">格式化</el-button>
            <el-button size="small" @click="validateJson">验证</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>

      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="configSaving" @click="handleSaveConfig">
          保存配置
        </el-button>
      </template>
    </el-dialog>

    <!-- 环境变量模板选择对话框 -->
    <el-dialog
      v-model="templateDialogVisible"
      title="选择环境变量模板"
      width="600px"
    >
      <el-radio-group v-model="selectedTemplate" style="width: 100%;">
        <el-radio :value="'web'" style="display: block; margin-bottom: 15px;">
          <div style="margin-left: 25px;">
            <div style="font-weight: 500;">Web应用模板</div>
            <div style="color: #909399; font-size: 12px;">包含端口、日志级别等常用配置</div>
          </div>
        </el-radio>
        <el-radio :value="'microservice'" style="display: block; margin-bottom: 15px;">
          <div style="margin-left: 25px;">
            <div style="font-weight: 500;">微服务模板</div>
            <div style="color: #909399; font-size: 12px;">包含服务发现、注册中心等配置</div>
          </div>
        </el-radio>
        <el-radio :value="'database'" style="display: block; margin-bottom: 15px;">
          <div style="margin-left: 25px;">
            <div style="font-weight: 500;">数据库应用模板</div>
            <div style="color: #909399; font-size: 12px;">包含数据库连接相关配置</div>
          </div>
        </el-radio>
      </el-radio-group>
      <template #footer>
        <el-button @click="templateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="applyTemplate">应用模板</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { InfoFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'

const tableData = ref([])
const projects = ref([])
const clusters = ref([])
const templates = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新建环境')
const formRef = ref(null)

// 绑定应用列表相关
const bindingsDialogVisible = ref(false)
const bindingsLoading = ref(false)
const bindingsList = ref([])
const currentEnv = ref(null)

// 配置编辑相关
const configDialogVisible = ref(false)
const activeConfigTab = ref('basic')
const currentBinding = ref(null)
const configSaving = ref(false)
const templateDialogVisible = ref(false)
const selectedTemplate = ref('web')

const configForm = reactive({
  replicas: 1,
  cpuRequest: '100m',
  cpuLimit: '500m',
  memoryRequest: '128Mi',
  memoryLimit: '512Mi'
})

const envVars = ref([])

const advancedConfig = reactive({
  healthCheck: {
    enabled: false,
    path: '/health',
    port: 8080
  },
  ingress: {
    enabled: false,
    host: '',
    path: '/'
  }
})

const jsonConfig = ref('')

// 环境变量模板
const envTemplates = {
  web: [
    { key: 'APP_NAME', value: '', description: '应用名称' },
    { key: 'APP_PORT', value: '8080', description: '应用端口' },
    { key: 'LOG_LEVEL', value: 'info', description: '日志级别' },
    { key: 'NODE_ENV', value: 'production', description: '运行环境' }
  ],
  microservice: [
    { key: 'SERVICE_NAME', value: '', description: '服务名称' },
    { key: 'SERVICE_PORT', value: '8080', description: '服务端口' },
    { key: 'REGISTRY_URL', value: '', description: '注册中心地址' },
    { key: 'CONFIG_SERVER', value: '', description: '配置中心地址' },
    { key: 'LOG_LEVEL', value: 'info', description: '日志级别' }
  ],
  database: [
    { key: 'DB_HOST', value: '', description: '数据库主机' },
    { key: 'DB_PORT', value: '3306', description: '数据库端口' },
    { key: 'DB_NAME', value: '', description: '数据库名' },
    { key: 'DB_USER', value: '', description: '数据库用户' },
    { key: 'DB_PASSWORD', value: '', description: '数据库密码' },
    { key: 'DB_POOL_SIZE', value: '10', description: '连接池大小' }
  ]
}


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
  templateId: null,
  description: '',
  status: 1,
  configJson: ''
})

const rules = {
  envCode: [{ required: true, message: '请输入环境编码', trigger: 'blur' }],
  envName: [{ required: true, message: '请输入环境名称', trigger: 'blur' }],
  envType: [{ required: true, message: '请选择环境类型', trigger: 'change' }],
  clusterId: [{ required: true, message: '请选择集群', trigger: 'change' }],
  namespace: [
    { required: true, message: '请输入命名空间', trigger: 'blur' },
    { 
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
      message: '只能包含小写字母、数字和短横线(-)，且必须以字母或数字开头和结尾',
      trigger: 'blur'
    },
    {
      min: 1,
      max: 63,
      message: '长度必须在1-63个字符之间',
      trigger: 'blur'
    }
  ],
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
    if (res.code === 0 || res.code === 200) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载环境列表失败')
  }
}

const loadProjects = async () => {
  try {
    const res = await request.get('/projects', { params: { page: 1, pageSize: 1000 } })
    if (res.code === 0 || res.code === 200) {
      projects.value = res.data.list || []
      console.log('加载项目列表成功，共', projects.value.length, '个项目')
    }
  } catch (error) {
    console.error('加载项目列表异常', error)
    // 拦截器已经显示过错误消息了，这里不再重复显示
  }
}

const loadClusters = async () => {
  try {
    const res = await request.get('/clusters', { params: { page: 1, pageSize: 1000 } })
    if (res.code === 0 || res.code === 200) {
      clusters.value = res.data.list || []
      console.log('加载集群列表成功，共', clusters.value.length, '个集群')
    }
  } catch (error) {
    console.error('加载集群列表异常', error)
    // 拦截器已经显示过错误消息了，这里不再重复显示
  }
}

const loadTemplates = async () => {
  try {
    const res = await request.get('/env-templates', { params: { page: 1, pageSize: 1000 } })
    if (res.code === 0 || res.code === 200) {
      templates.value = res.data.list || []
      console.log('加载模板列表成功，共', templates.value.length, '个模板')
    }
  } catch (error) {
    console.error('加载模板列表异常', error)
    // 拦截器已经显示过错误消息了，这里不再重复显示
  }
}

const handleCreate = async () => {
  dialogTitle.value = '新建环境'
  resetForm()
  // 打开对话框前重新加载项目、集群和模板列表
  await Promise.all([loadProjects(), loadClusters(), loadTemplates()])
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

const handleViewBindings = async (row) => {
  currentEnv.value = row
  bindingsDialogVisible.value = true
  bindingsLoading.value = true
  bindingsList.value = []
  
  try {
    const res = await request.get('/app-env-bindings', {
      params: { envId: row.id, page: 1, pageSize: 100 }
    })
    
    if (res.code === 0) {
      bindingsList.value = res.data.list || []
      
      // 如果有应用ID，查询应用名称
      if (bindingsList.value.length > 0) {
        // 批量查询应用信息（这里简化处理，实际可以调用批量查询接口）
        for (const binding of bindingsList.value) {
          try {
            const appRes = await request.get(`/applications/${binding.appId}`)
            if (appRes.code === 0 && appRes.data.application) {
              binding.appName = appRes.data.application.name
            } else {
              binding.appName = `应用${binding.appId}`
            }
          } catch (error) {
            binding.appName = `应用${binding.appId}`
          }
        }
      }
    } else {
      ElMessage.error(res.message || '加载绑定列表失败')
    }
  } catch (error) {
    console.error('加载绑定列表失败', error)
    ElMessage.error('加载绑定列表失败')
  } finally {
    bindingsLoading.value = false
  }
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 编辑配置
const handleEditConfig = (row) => {
  currentBinding.value = row
  
  // 填充基础配置
  configForm.replicas = row.replicas || 1
  configForm.cpuRequest = row.cpuRequest || '100m'
  configForm.cpuLimit = row.cpuLimit || '500m'
  configForm.memoryRequest = row.memoryRequest || '128Mi'
  configForm.memoryLimit = row.memoryLimit || '512Mi'
  
  // 解析现有配置
  try {
    if (row.configJson && row.configJson !== '{}') {
      const config = JSON.parse(row.configJson)
      
      // 填充环境变量
      if (config.env) {
        envVars.value = Object.entries(config.env).map(([key, value]) => ({
          key,
          value: String(value),
          description: ''
        }))
      } else {
        envVars.value = []
      }
      
      // 填充高级配置
      if (config.healthCheck) {
        Object.assign(advancedConfig.healthCheck, config.healthCheck)
      }
      if (config.ingress) {
        Object.assign(advancedConfig.ingress, config.ingress)
      }
      
      // 填充JSON编辑器
      jsonConfig.value = JSON.stringify(config, null, 2)
    } else {
      // 默认配置
      envVars.value = []
      advancedConfig.healthCheck.enabled = false
      advancedConfig.ingress.enabled = false
      jsonConfig.value = JSON.stringify({
        env: {},
        healthCheck: { enabled: false },
        ingress: { enabled: false }
      }, null, 2)
    }
  } catch (error) {
    console.error('解析配置失败', error)
    envVars.value = []
    jsonConfig.value = '{}'
  }
  
  activeConfigTab.value = 'basic'
  configDialogVisible.value = true
}

// 添加环境变量
const addEnvVar = () => {
  envVars.value.push({
    key: '',
    value: '',
    description: ''
  })
}

// 删除环境变量
const removeEnvVar = (index) => {
  envVars.value.splice(index, 1)
}

// 加载环境变量模板
const loadEnvTemplate = () => {
  templateDialogVisible.value = true
}

// 应用模板
const applyTemplate = () => {
  const template = envTemplates[selectedTemplate.value]
  if (template) {
    envVars.value = JSON.parse(JSON.stringify(template))
    ElMessage.success('模板已应用')
  }
  templateDialogVisible.value = false
}

// 格式化JSON
const formatJson = () => {
  try {
    const obj = JSON.parse(jsonConfig.value)
    jsonConfig.value = JSON.stringify(obj, null, 2)
    ElMessage.success('格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

// 验证JSON
const validateJson = () => {
  try {
    JSON.parse(jsonConfig.value)
    ElMessage.success('JSON格式正确')
  } catch (error) {
    ElMessage.error('JSON格式错误: ' + error.message)
  }
}

// 构建完整配置
const buildFullConfig = () => {
  // 如果用户在JSON标签，直接使用JSON编辑器的内容
  if (activeConfigTab.value === 'json') {
    try {
      return JSON.parse(jsonConfig.value)
    } catch (error) {
      throw new Error('JSON格式错误: ' + error.message)
    }
  }
  
  // 否则从各个标签构建配置
  const config = {}
  
  // 环境变量
  if (envVars.value.length > 0) {
    config.env = {}
    envVars.value.forEach(item => {
      if (item.key) {
        config.env[item.key] = item.value
      }
    })
  }
  
  // 健康检查
  if (advancedConfig.healthCheck.enabled) {
    config.healthCheck = {
      enabled: true,
      path: advancedConfig.healthCheck.path,
      port: advancedConfig.healthCheck.port
    }
  }
  
  // Ingress
  if (advancedConfig.ingress.enabled) {
    config.ingress = {
      enabled: true,
      host: advancedConfig.ingress.host,
      path: advancedConfig.ingress.path
    }
  }
  
  return config
}

// 保存配置
const handleSaveConfig = async () => {
  try {
    configSaving.value = true
    
    // 构建配置
    const config = buildFullConfig()
    
    // 准备提交数据
    const data = {
      replicas: configForm.replicas,
      cpuRequest: configForm.cpuRequest,
      cpuLimit: configForm.cpuLimit,
      memoryRequest: configForm.memoryRequest,
      memoryLimit: configForm.memoryLimit,
      configJson: JSON.stringify(config)
    }
    
    // 提交更新
    const res = await request.put(`/app-env-bindings/${currentBinding.value.id}`, data)
    
    if (res.code === 0) {
      ElMessage.success('配置保存成功')
      configDialogVisible.value = false
      // 刷新绑定列表
      if (currentEnv.value) {
        handleViewBindings(currentEnv.value)
      }
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error) {
    if (error.message) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('保存失败')
    }
  } finally {
    configSaving.value = false
  }
}

// 关闭配置对话框
const handleConfigDialogClose = () => {
  currentBinding.value = null
  envVars.value = []
  jsonConfig.value = ''
}


// 验证namespace格式（实时检查）
const validateNamespaceFormat = () => {
  const ns = form.namespace
  if (!ns) return
  
  // 检查保留名称
  const reserved = ['default', 'kube-system', 'kube-public', 'kube-node-lease']
  if (reserved.includes(ns)) {
    ElMessage.warning('不能使用保留的命名空间名称')
    return
  }
  
  if (ns.startsWith('kube-') || ns.startsWith('kubernetes-')) {
    ElMessage.warning('不能使用以 kube- 或 kubernetes- 开头的命名空间')
    return
  }
  
  if (ns.includes('--')) {
    ElMessage.warning('命名空间不能包含连续的短横线')
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const url = form.id ? `/environments/${form.id}` : '/environments'
    const method = form.id ? 'put' : 'post'
    const res = await request[method](url, form)
    if (res.code === 0 || res.code === 200) {
      ElMessage.success(form.id ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadEnvironments()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error) {
    console.error('表单提交失败', error)
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
  form.templateId = null
  form.description = ''
  form.status = 1
  form.configJson = ''
}

onMounted(() => {
  loadEnvironments()
  loadProjects()
  loadClusters()
  loadTemplates()
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
