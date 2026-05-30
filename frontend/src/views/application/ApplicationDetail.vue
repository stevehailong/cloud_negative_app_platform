<template>
  <div class="application-detail">
    <el-page-header @back="goBack" :title="'应用详情'">
      <template #content>
        <div class="header-content">
          <span class="app-name">{{ appInfo.name }}</span>
          <el-tag :type="getTypeColor(appInfo.type)" size="large">{{ appInfo.type }}</el-tag>
        </div>
      </template>
    </el-page-header>
    
    <!-- 基本信息 -->
    <el-card class="info-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>基本信息</span>
          <el-button type="primary" size="small" @click="handleEdit">
            <el-icon><Edit /></el-icon>
            编辑
          </el-button>
        </div>
      </template>
      
      <el-descriptions :column="2" border>
        <el-descriptions-item label="应用名称">{{ appInfo.name }}</el-descriptions-item>
        <el-descriptions-item label="应用编码">{{ appInfo.code }}</el-descriptions-item>
        <el-descriptions-item label="应用类型">
          <el-tag :type="getTypeColor(appInfo.type)">{{ getTypeLabel(appInfo.type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="开发语言">{{ appInfo.language || '-' }}</el-descriptions-item>
        <el-descriptions-item label="开发框架">{{ appInfo.framework || '-' }}</el-descriptions-item>
        <el-descriptions-item label="负责人">{{ appInfo.owner || '-' }}</el-descriptions-item>
        <el-descriptions-item label="代码仓库" :span="2">
          <el-link v-if="appInfo.repoUrl" :href="appInfo.repoUrl" target="_blank" type="primary">
            {{ appInfo.repoUrl }}
          </el-link>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="默认分支">{{ appInfo.repoBranch || '-' }}</el-descriptions-item>
        <el-descriptions-item label="构建工具">{{ appInfo.buildTool || '-' }}</el-descriptions-item>
        <el-descriptions-item label="构建路径">{{ appInfo.buildPath || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Dockerfile">{{ appInfo.dockerFile || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建人">{{ appInfo.createdBy || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(appInfo.createTime) }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ appInfo.description || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
    
    <!-- 组件管理 -->
    <el-card class="component-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>组件管理</span>
          <el-button type="primary" size="small" @click="showAddComponent">
            <el-icon><Plus /></el-icon>
            添加组件
          </el-button>
        </div>
      </template>
      
      <el-table
        v-loading="componentLoading"
        :data="components"
        style="width: 100%"
      >
        <el-table-column prop="name" label="组件名称" width="150" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getComponentTypeColor(row.type)">{{ getComponentTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="image" label="镜像" min-width="200" show-overflow-tooltip />
        <el-table-column prop="port" label="端口" width="80" />
        <el-table-column prop="replicas" label="副本数" width="80" />
        <el-table-column label="资源限制" width="150">
          <template #default="{ row }">
            CPU: {{ row.cpu || '-' }}<br/>
            内存: {{ row.memory || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewComponentDetail(row)">
              详情
            </el-button>
            <el-button link type="primary" @click="handleEditComponent(row)">
              编辑
            </el-button>
            <el-button link type="danger" @click="handleDeleteComponent(row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 环境绑定 -->
    <el-card class="env-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>环境绑定</span>
          <el-button type="primary" size="small" @click="showBindEnv">
            <el-icon><Link /></el-icon>
            绑定环境
          </el-button>
        </div>
      </template>
      
      <el-table
        v-loading="envLoading"
        :data="envBindings"
        style="width: 100%"
      >
        <el-table-column prop="envName" label="环境名称" width="150" />
        <el-table-column prop="envType" label="环境类型" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.envType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="150" />
        <el-table-column prop="clusterName" label="集群" width="150" />
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
        <el-table-column label="操作" fixed="right" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewEnvConfig(row)">
              配置
            </el-button>
            <el-button link type="danger" @click="handleUnbindEnv(row.id)">
              解绑
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 添加/编辑组件对话框 -->
    <el-dialog
      v-model="componentDialogVisible"
      :title="componentDialogTitle"
      width="700px"
    >
      <el-form
        ref="componentFormRef"
        :model="componentForm"
        :rules="componentRules"
        label-width="100px"
      >
        <el-form-item label="组件名称" prop="name">
          <el-input v-model="componentForm.name" placeholder="请输入组件名称" />
        </el-form-item>
        
        <el-form-item label="组件类型" prop="type">
          <el-select v-model="componentForm.type" placeholder="请选择组件类型">
            <el-option label="前端服务" value="frontend" />
            <el-option label="后端服务" value="backend" />
            <el-option label="数据库" value="database" />
            <el-option label="缓存" value="cache" />
            <el-option label="消息队列" value="queue" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="版本" prop="version">
          <el-input v-model="componentForm.version" placeholder="如: v1.0.0" />
        </el-form-item>
        
        <el-form-item label="镜像地址" prop="image">
          <el-input v-model="componentForm.image" placeholder="如: nginx:latest" />
        </el-form-item>
        
        <el-form-item label="端口">
          <el-input-number v-model="componentForm.port" :min="1" :max="65535" />
        </el-form-item>
        
        <el-form-item label="副本数">
          <el-input-number v-model="componentForm.replicas" :min="1" :max="100" />
        </el-form-item>
        
        <el-form-item label="CPU限制">
          <el-input v-model="componentForm.cpu" placeholder="如: 500m 或 1" />
        </el-form-item>
        
        <el-form-item label="内存限制">
          <el-input v-model="componentForm.memory" placeholder="如: 512Mi 或 1Gi" />
        </el-form-item>
        
        <el-form-item label="环境变量">
          <el-input
            v-model="componentForm.envVars"
            type="textarea"
            :rows="3"
            placeholder='JSON格式, 如: {"ENV": "prod", "PORT": "8080"}'
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="componentDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="componentSubmitting" @click="handleSubmitComponent">
          确定
        </el-button>
      </template>
    </el-dialog>
    
    <!-- 组件详情对话框 -->
    <el-dialog
      v-model="componentDetailVisible"
      title="组件详情"
      width="800px"
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="组件名称">{{ currentComponent.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag :type="getComponentTypeColor(currentComponent.type)">
            {{ getComponentTypeLabel(currentComponent.type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="版本">{{ currentComponent.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="端口">{{ currentComponent.port || '-' }}</el-descriptions-item>
        <el-descriptions-item label="副本数">{{ currentComponent.replicas || '-' }}</el-descriptions-item>
        <el-descriptions-item label="镜像" :span="2">
          <el-text truncated style="max-width: 500px">{{ currentComponent.image || '-' }}</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="CPU限制">{{ currentComponent.cpu || '-' }}</el-descriptions-item>
        <el-descriptions-item label="内存限制">{{ currentComponent.memory || '-' }}</el-descriptions-item>
        <el-descriptions-item label="环境变量" :span="2">
          <pre v-if="currentComponent.envVars" style="margin: 0">{{ formatJSON(currentComponent.envVars) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="ConfigMaps" :span="2">
          <pre v-if="currentComponent.configMaps" style="margin: 0">{{ formatJSON(currentComponent.configMaps) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="Secrets" :span="2">
          <pre v-if="currentComponent.secrets" style="margin: 0">{{ formatJSON(currentComponent.secrets) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="存储卷" :span="2">
          <pre v-if="currentComponent.volumes" style="margin: 0">{{ formatJSON(currentComponent.volumes) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Plus, Link } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/time'

const router = useRouter()
const route = useRoute()

const appInfo = ref({})
const components = ref([])
const envBindings = ref([])
const componentLoading = ref(false)
const envLoading = ref(false)

const componentDialogVisible = ref(false)
const componentDialogTitle = ref('添加组件')
const componentFormRef = ref(null)
const componentSubmitting = ref(false)
const componentForm = reactive({
  id: null,
  applicationId: null,
  name: '',
  type: '',
  version: '',
  image: '',
  port: 8080,
  replicas: 1,
  cpu: '',
  memory: '',
  envVars: ''
})

const componentDetailVisible = ref(false)
const currentComponent = ref({})

const componentRules = {
  name: [{ required: true, message: '请输入组件名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择组件类型', trigger: 'change' }],
  image: [{ required: true, message: '请输入镜像地址', trigger: 'blur' }]
}

// 加载应用信息
const loadAppInfo = async () => {
  try {
    const appId = route.params.id
    const res = await request.get(`/applications/${appId}`)
    appInfo.value = res.data.application || {}
    components.value = res.data.components || []
  } catch (error) {
    console.error('加载应用信息失败', error)
    ElMessage.error('加载应用信息失败')
  }
}

// 加载环境绑定
const loadEnvBindings = async () => {
  envLoading.value = true
  try {
    const appId = route.params.id
    const res = await request.get(`/app-env-bindings`, {
      params: { applicationId: appId, page: 1, pageSize: 100 }
    })
    envBindings.value = res.data.list || []
  } catch (error) {
    console.error('加载环境绑定失败', error)
  } finally {
    envLoading.value = false
  }
}

const goBack = () => {
  router.back()
}

const handleEdit = () => {
  router.push(`/applications/${appInfo.value.id}/edit`)
}

// 显示添加组件对话框
const showAddComponent = () => {
  componentDialogTitle.value = '添加组件'
  Object.assign(componentForm, {
    id: null,
    applicationId: appInfo.value.id,
    name: '',
    type: '',
    version: '',
    image: '',
    port: 8080,
    replicas: 1,
    cpu: '',
    memory: '',
    envVars: ''
  })
  componentDialogVisible.value = true
}

// 编辑组件
const handleEditComponent = (row) => {
  componentDialogTitle.value = '编辑组件'
  Object.assign(componentForm, row)
  componentDialogVisible.value = true
}

// 提交组件表单
const handleSubmitComponent = async () => {
  if (!componentFormRef.value) return
  
  try {
    await componentFormRef.value.validate()
    componentSubmitting.value = true
    
    if (componentForm.id) {
      await request.put(`/components/${componentForm.id}`, componentForm)
      ElMessage.success('更新成功')
    } else {
      await request.post('/components', componentForm)
      ElMessage.success('添加成功')
    }
    
    componentDialogVisible.value = false
    loadAppInfo()
  } catch (error) {
    if (error.message) {
      console.error('操作失败', error)
    }
  } finally {
    componentSubmitting.value = false
  }
}

// 删除组件
const handleDeleteComponent = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该组件吗？', '提示', {
      type: 'warning'
    })
    await request.delete(`/components/${id}`)
    ElMessage.success('删除成功')
    loadAppInfo()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败', error)
    }
  }
}

// 查看组件详情
const viewComponentDetail = (row) => {
  currentComponent.value = row
  componentDetailVisible.value = true
}

// 绑定环境
const showBindEnv = () => {
  ElMessage.info('环境绑定功能开发中')
}

// 查看环境配置
const viewEnvConfig = (row) => {
  router.push(`/app-env-config/${row.id}`)
}

// 解绑环境
const handleUnbindEnv = async (id) => {
  try {
    await ElMessageBox.confirm('确定要解绑该环境吗？', '提示', {
      type: 'warning'
    })
    await request.delete(`/app-env-bindings/${id}`)
    ElMessage.success('解绑成功')
    loadEnvBindings()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('解绑失败', error)
    }
  }
}

// 工具方法
const getTypeColor = (type) => {
  const colorMap = {
    web: 'primary',
    api: 'success',
    job: 'warning',
    function: 'info'
  }
  return colorMap[type] || ''
}

const getTypeLabel = (type) => {
  const labelMap = {
    web: 'Web应用',
    api: 'API服务',
    job: '定时任务',
    function: '函数服务'
  }
  return labelMap[type] || type
}

const getComponentTypeColor = (type) => {
  const colorMap = {
    frontend: 'primary',
    backend: 'success',
    database: 'warning',
    cache: 'info',
    queue: 'danger',
    other: ''
  }
  return colorMap[type] || ''
}

const getComponentTypeLabel = (type) => {
  const labelMap = {
    frontend: '前端服务',
    backend: '后端服务',
    database: '数据库',
    cache: '缓存',
    queue: '消息队列',
    other: '其他'
  }
  return labelMap[type] || type
}

const formatJSON = (str) => {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

onMounted(() => {
  loadAppInfo()
  loadEnvBindings()
})
</script>

<style scoped lang="scss">
.application-detail {
  padding: 20px;
  
  .header-content {
    display: flex;
    align-items: center;
    gap: 10px;
    
    .app-name {
      font-size: 18px;
      font-weight: bold;
    }
  }
  
  .info-card,
  .component-card,
  .env-card {
    margin-top: 20px;
  }
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  pre {
    background: #f5f7fa;
    padding: 10px;
    border-radius: 4px;
    font-size: 12px;
    line-height: 1.5;
  }
}
</style>
