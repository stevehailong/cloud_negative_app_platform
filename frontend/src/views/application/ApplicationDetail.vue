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
        <el-table-column label="Ingress" width="180">
          <template #default="{ row }">
            <template v-if="getIngressInfo(row).enabled">
              <el-tag type="success" size="small">已启用</el-tag>
              <span style="margin-left:6px;font-size:12px;color:#409EFF">{{ getIngressInfo(row).host }}</span>
            </template>
            <span v-else style="color:#C0C4CC">-</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="150">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEditConfig(row)">
              编辑配置
            </el-button>
            <el-button link type="danger" size="small" @click="handleUnbindEnv(row.id)">
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
    
    <!-- 绑定环境对话框 -->
    <el-dialog
      v-model="bindEnvDialogVisible"
      title="绑定环境"
      width="600px"
    >
      <el-form
        ref="bindEnvFormRef"
        :model="bindEnvForm"
        :rules="bindEnvRules"
        label-width="120px"
      >
        <el-form-item label="选择环境" prop="envId">
          <el-select 
            v-model="bindEnvForm.envId" 
            placeholder="请选择环境" 
            style="width: 100%"
            filterable
            @change="handleEnvChange"
          >
            <el-option
              v-for="env in availableEnvironments"
              :key="env.id"
              :label="`${env.envName} (${env.envType})`"
              :value="env.id"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ env.envName }}</span>
                <el-tag size="small" style="margin-left: 10px;">{{ env.envType }}</el-tag>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="副本数" prop="replicas">
          <el-input-number v-model="bindEnvForm.replicas" :min="1" :max="100" />
        </el-form-item>
        
        <el-form-item label="CPU请求">
          <el-input v-model="bindEnvForm.cpuRequest" placeholder="如: 100m" />
        </el-form-item>
        
        <el-form-item label="CPU限制">
          <el-input v-model="bindEnvForm.cpuLimit" placeholder="如: 500m" />
        </el-form-item>
        
        <el-form-item label="内存请求">
          <el-input v-model="bindEnvForm.memoryRequest" placeholder="如: 128Mi" />
        </el-form-item>
        
        <el-form-item label="内存限制">
          <el-input v-model="bindEnvForm.memoryLimit" placeholder="如: 512Mi" />
        </el-form-item>
        
        <el-alert
          v-if="selectedEnvInfo"
          title="环境信息"
          type="info"
          :closable="false"
          style="margin-bottom: 20px;"
        >
          <div>集群: {{ selectedEnvInfo.clusterName }}</div>
          <div>命名空间: {{ selectedEnvInfo.namespace }}</div>
        </el-alert>
      </el-form>
      
      <template #footer>
        <el-button @click="bindEnvDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="bindEnvSubmitting" @click="handleSubmitBindEnv">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 配置编辑对话框 -->
    <el-dialog
      v-model="configDialogVisible"
      title="编辑配置"
      width="900px"
      destroy-on-close
    >
      <el-tabs v-model="activeConfigTab" type="border-card">
        <!-- 基础配置 -->
        <el-tab-pane label="基础配置" name="basic">
          <el-form :model="configForm" label-width="140px">
            <el-form-item label="副本数">
              <el-input-number v-model="configForm.replicas" :min="1" :max="100" />
            </el-form-item>
            <el-form-item label="CPU请求">
              <el-input v-model="configForm.cpuRequest" placeholder="如: 100m, 0.5" />
            </el-form-item>
            <el-form-item label="CPU限制">
              <el-input v-model="configForm.cpuLimit" placeholder="如: 500m, 1" />
            </el-form-item>
            <el-form-item label="内存请求">
              <el-input v-model="configForm.memoryRequest" placeholder="如: 128Mi, 1Gi" />
            </el-form-item>
            <el-form-item label="内存限制">
              <el-input v-model="configForm.memoryLimit" placeholder="如: 512Mi, 2Gi" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 环境变量 -->
        <el-tab-pane label="环境变量" name="envVars">
          <div style="margin-bottom: 15px;">
            <el-button type="primary" size="small" @click="addEnvVar">
              <el-icon><Plus /></el-icon>
              添加变量
            </el-button>
            <el-button size="small" @click="showTemplateDialog">
              <el-icon><Document /></el-icon>
              应用模板
            </el-button>
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
            <template v-if="advancedConfig.healthCheck.enabled">
              <el-form-item label="检查路径">
                <el-input v-model="advancedConfig.healthCheck.path" placeholder="/health" />
              </el-form-item>
              <el-form-item label="检查端口">
                <el-input-number v-model="advancedConfig.healthCheck.port" :min="1" :max="65535" />
              </el-form-item>
              <el-form-item label="初始延迟(秒)">
                <el-input-number v-model="advancedConfig.healthCheck.initialDelaySeconds" :min="0" />
              </el-form-item>
              <el-form-item label="检查间隔(秒)">
                <el-input-number v-model="advancedConfig.healthCheck.periodSeconds" :min="1" />
              </el-form-item>
            </template>

            <el-divider />

            <el-form-item label="启用 Ingress">
              <el-switch v-model="advancedConfig.ingress.enabled" />
            </el-form-item>
            <template v-if="advancedConfig.ingress.enabled">
              <el-form-item label="域名">
                <el-input v-model="advancedConfig.ingress.host" placeholder="example.com" />
              </el-form-item>
              <el-form-item label="路径">
                <el-input v-model="advancedConfig.ingress.path" placeholder="/" />
              </el-form-item>
              <el-form-item label="服务端口">
                <el-input-number v-model="advancedConfig.ingress.servicePort" :min="1" :max="65535" />
              </el-form-item>
            </template>
          </el-form>
        </el-tab-pane>

        <!-- JSON编辑器 -->
        <el-tab-pane label="JSON编辑器" name="json">
          <div style="margin-bottom: 10px;">
            <el-button size="small" @click="formatJson">格式化</el-button>
            <el-text type="info" size="small" style="margin-left: 10px;">
              直接编辑完整的配置JSON
            </el-text>
          </div>
          <el-input
            v-model="configJsonText"
            type="textarea"
            :rows="20"
            placeholder="JSON格式的配置"
            style="font-family: monospace;"
          />
        </el-tab-pane>
      </el-tabs>

      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveConfig">保存配置</el-button>
      </template>
    </el-dialog>

    <!-- 环境变量模板选择对话框 -->
    <el-dialog
      v-model="templateDialogVisible"
      title="选择环境变量模板"
      width="500px"
    >
      <el-radio-group v-model="selectedTemplate" style="width: 100%;">
        <el-radio label="web" style="display: block; margin: 15px 0;">
          <strong>Web应用模板</strong>
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            包含应用名称、端口、日志级别等常用变量
          </div>
        </el-radio>
        <el-radio label="microservice" style="display: block; margin: 15px 0;">
          <strong>微服务模板</strong>
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            包含服务名称、端口、数据库连接、消息队列等
          </div>
        </el-radio>
        <el-radio label="database" style="display: block; margin: 15px 0;">
          <strong>数据库模板</strong>
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            包含数据库类型、连接信息、用户凭证等
          </div>
        </el-radio>
      </el-radio-group>
      <template #footer>
        <el-button @click="templateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="applyTemplate">应用模板</el-button>
      </template>
    </el-dialog>

    <!-- 编辑应用基本信息对话框 -->
    <el-dialog
      v-model="editAppDialogVisible"
      title="编辑应用信息"
      width="700px"
      destroy-on-close
    >
      <el-form
        ref="editAppFormRef"
        :model="editAppForm"
        :rules="editAppRules"
        label-width="120px"
      >
        <el-form-item label="应用名称" prop="name">
          <el-input v-model="editAppForm.name" placeholder="请输入应用名称" />
        </el-form-item>
        <el-form-item label="应用编码" prop="code">
          <el-input v-model="editAppForm.code" placeholder="请输入应用编码" disabled />
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            应用编码创建后不可修改
          </div>
        </el-form-item>
        <el-form-item label="应用类型" prop="type">
          <el-select v-model="editAppForm.type" placeholder="请选择应用类型">
            <el-option label="Web应用" value="web" />
            <el-option label="API服务" value="api" />
            <el-option label="定时任务" value="job" />
            <el-option label="函数服务" value="function" />
          </el-select>
        </el-form-item>
        <el-form-item label="开发语言">
          <el-input v-model="editAppForm.language" placeholder="如: Java, Go, Python" />
        </el-form-item>
        <el-form-item label="开发框架">
          <el-input v-model="editAppForm.framework" placeholder="如: Spring Boot, Gin, Django" />
        </el-form-item>
        <el-form-item label="负责人">
          <el-input v-model="editAppForm.owner" placeholder="请输入负责人姓名" />
        </el-form-item>
        <el-form-item label="代码仓库" prop="repoUrl">
          <el-input v-model="editAppForm.repoUrl" placeholder="Git仓库地址" />
        </el-form-item>
        <el-form-item label="默认分支">
          <el-input v-model="editAppForm.repoBranch" placeholder="如: main, master" />
        </el-form-item>
        <el-form-item label="构建工具">
          <el-input v-model="editAppForm.buildTool" placeholder="如: maven, go build, npm" />
        </el-form-item>
        <el-form-item label="构建路径">
          <el-input v-model="editAppForm.buildPath" placeholder="如: target/*.jar, ./bin/app" />
        </el-form-item>
        <el-form-item label="Dockerfile">
          <el-input v-model="editAppForm.dockerFile" placeholder="如: ./Dockerfile, ./deploy/Dockerfile" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="editAppForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入应用描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editAppDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editAppSubmitting" @click="handleSaveAppEdit">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Plus, Link, Document } from '@element-plus/icons-vue'
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

const bindEnvDialogVisible = ref(false)
const bindEnvFormRef = ref(null)
const bindEnvSubmitting = ref(false)
const availableEnvironments = ref([])
const selectedEnvInfo = ref(null)
const bindEnvForm = reactive({
  appId: null,
  envId: null,
  replicas: 1,
  cpuRequest: '100m',
  cpuLimit: '500m',
  memoryRequest: '128Mi',
  memoryLimit: '512Mi',
  configJson: '{}'
})

const bindEnvRules = {
  envId: [{ required: true, message: '请选择环境', trigger: 'change' }],
  replicas: [{ required: true, message: '请输入副本数', trigger: 'blur' }]
}

// 编辑应用基本信息相关状态
const editAppDialogVisible = ref(false)
const editAppFormRef = ref(null)
const editAppSubmitting = ref(false)
const editAppForm = reactive({
  id: null,
  name: '',
  code: '',
  type: '',
  language: '',
  framework: '',
  owner: '',
  repoUrl: '',
  repoBranch: '',
  buildTool: '',
  buildPath: '',
  dockerFile: '',
  description: ''
})

const editAppRules = {
  name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入应用编码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择应用类型', trigger: 'change' }],
  repoUrl: [{ required: true, message: '请输入代码仓库地址', trigger: 'blur' }]
}

// 配置编辑相关状态
const configDialogVisible = ref(false)
const activeConfigTab = ref('basic')
const currentBinding = ref(null)
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
    port: 8080,
    initialDelaySeconds: 30,
    periodSeconds: 10
  },
  ingress: {
    enabled: false,
    host: '',
    path: '/',
    servicePort: 80
  }
})
const configJsonText = ref('{}')
const templateDialogVisible = ref(false)
const selectedTemplate = ref('web')

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
    { key: 'DB_HOST', value: 'mysql', description: '数据库主机' },
    { key: 'DB_PORT', value: '3306', description: '数据库端口' },
    { key: 'REDIS_HOST', value: 'redis', description: 'Redis主机' },
    { key: 'REDIS_PORT', value: '6379', description: 'Redis端口' },
    { key: 'MESSAGE_QUEUE_URL', value: '', description: '消息队列地址' }
  ],
  database: [
    { key: 'DB_TYPE', value: 'mysql', description: '数据库类型' },
    { key: 'DB_NAME', value: '', description: '数据库名' },
    { key: 'DB_USER', value: '', description: '数据库用户' },
    { key: 'DB_PASSWORD', value: '', description: '数据库密码' },
    { key: 'DB_ROOT_PASSWORD', value: '', description: 'Root密码' }
  ]
}


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
  // 加载当前应用信息到表单
  editAppForm.id = appInfo.value.id
  editAppForm.name = appInfo.value.name
  editAppForm.code = appInfo.value.code
  editAppForm.type = appInfo.value.type
  editAppForm.language = appInfo.value.language || ''
  editAppForm.framework = appInfo.value.framework || ''
  editAppForm.owner = appInfo.value.owner || ''
  editAppForm.repoUrl = appInfo.value.repoUrl || ''
  editAppForm.repoBranch = appInfo.value.repoBranch || ''
  editAppForm.buildTool = appInfo.value.buildTool || ''
  editAppForm.buildPath = appInfo.value.buildPath || ''
  editAppForm.dockerFile = appInfo.value.dockerFile || ''
  editAppForm.description = appInfo.value.description || ''
  
  editAppDialogVisible.value = true
}

// 保存应用编辑
const handleSaveAppEdit = async () => {
  try {
    await editAppFormRef.value.validate()
    editAppSubmitting.value = true
    
    const updateData = {
      name: editAppForm.name,
      type: editAppForm.type,
      language: editAppForm.language,
      framework: editAppForm.framework,
      owner: editAppForm.owner,
      repoUrl: editAppForm.repoUrl,
      repoBranch: editAppForm.repoBranch,
      buildTool: editAppForm.buildTool,
      buildPath: editAppForm.buildPath,
      dockerFile: editAppForm.dockerFile,
      description: editAppForm.description
    }
    
    await request.put(`/applications/${editAppForm.id}`, updateData)
    
    ElMessage.success('应用信息更新成功')
    editAppDialogVisible.value = false
    loadAppInfo()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('保存应用信息失败', error)
      ElMessage.error('保存应用信息失败')
    }
  } finally {
    editAppSubmitting.value = false
  }
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
const showBindEnv = async () => {
  // 重置表单
  Object.assign(bindEnvForm, {
    appId: appInfo.value.id,
    envId: null,
    replicas: 1,
    cpuRequest: '100m',
    cpuLimit: '500m',
    memoryRequest: '128Mi',
    memoryLimit: '512Mi',
    configJson: '{}'
  })
  selectedEnvInfo.value = null
  
  // 加载可用环境列表
  try {
    const res = await request.get('/environments', {
      params: { page: 1, pageSize: 100 }
    })
    availableEnvironments.value = res.data.list || []
    bindEnvDialogVisible.value = true
  } catch (error) {
    console.error('加载环境列表失败', error)
    ElMessage.error('加载环境列表失败')
  }
}

// 环境选择变化
const handleEnvChange = (envId) => {
  const env = availableEnvironments.value.find(e => e.id === envId)
  if (env) {
    selectedEnvInfo.value = {
      clusterName: env.clusterName || '未知',
      namespace: env.namespace
    }
  }
}

// 提交环境绑定
const handleSubmitBindEnv = async () => {
  if (!bindEnvFormRef.value) return
  
  try {
    await bindEnvFormRef.value.validate()
    bindEnvSubmitting.value = true
    
    await request.post('/app-env-bindings', bindEnvForm)
    ElMessage.success('绑定成功')
    
    bindEnvDialogVisible.value = false
    loadEnvBindings()
  } catch (error) {
    if (error.message) {
      console.error('绑定失败', error)
    }
  } finally {
    bindEnvSubmitting.value = false
  }
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

// 编辑配置
const handleEditConfig = (row) => {
  currentBinding.value = row
  activeConfigTab.value = 'basic'
  
  // 加载现有配置
  configForm.replicas = row.replicas || 1
  configForm.cpuRequest = row.cpuRequest || '100m'
  configForm.cpuLimit = row.cpuLimit || '500m'
  configForm.memoryRequest = row.memoryRequest || '128Mi'
  configForm.memoryLimit = row.memoryLimit || '512Mi'
  
  // 解析config_json
  try {
    const config = row.configJson ? JSON.parse(row.configJson) : {}
    
    // 加载环境变量
    if (config.envVars && Array.isArray(config.envVars)) {
      envVars.value = config.envVars.map(v => ({ ...v }))
    } else {
      envVars.value = []
    }
    
    // 加载高级配置
    if (config.healthCheck) {
      Object.assign(advancedConfig.healthCheck, config.healthCheck)
    } else {
      advancedConfig.healthCheck = {
        enabled: false,
        path: '/health',
        port: 8080,
        initialDelaySeconds: 30,
        periodSeconds: 10
      }
    }
    
    // 加载 Ingress 配置（兼容两种格式）
    // 格式1（后端存储格式）：{ ingressEnabled: true, ingressHost: "app-1.local", ingressPath: "/", containerPort: 9891 }
    // 格式2（前端内部格式）：{ ingress: { enabled: true, host: "app-1.local", path: "/", servicePort: 80 } }
    if (config.ingressEnabled !== undefined || config.ingressHost !== undefined) {
      // 后端存储的扁平格式
      advancedConfig.ingress = {
        enabled: config.ingressEnabled === true,
        host: config.ingressHost || '',
        path: config.ingressPath || '/',
        servicePort: config.containerPort || 80
      }
    } else if (config.ingress && config.ingress.enabled !== undefined) {
      // 前端内部嵌套格式
      Object.assign(advancedConfig.ingress, config.ingress)
    } else {
      advancedConfig.ingress = {
        enabled: false,
        host: '',
        path: '/',
        servicePort: 80
      }
    }
    
    configJsonText.value = JSON.stringify(config, null, 2)
  } catch (error) {
    console.error('解析配置失败', error)
    envVars.value = []
    configJsonText.value = '{}'
  }
  
  configDialogVisible.value = true
}

// 添加环境变量
const addEnvVar = () => {
  envVars.value.push({ key: '', value: '', description: '' })
}

// 删除环境变量
const removeEnvVar = (index) => {
  envVars.value.splice(index, 1)
}

// 显示模板对话框
const showTemplateDialog = () => {
  templateDialogVisible.value = true
}

// 应用模板
const applyTemplate = () => {
  const template = envTemplates[selectedTemplate.value]
  if (template) {
    envVars.value = template.map(v => ({ ...v }))
    ElMessage.success('模板已应用')
  }
  templateDialogVisible.value = false
}

// 从绑定的 configJson 中提取 Ingress 信息
const getIngressInfo = (row) => {
  try {
    const config = typeof row.configJson === 'string' ? JSON.parse(row.configJson) : (row.configJson || {})
    // 兼容扁平格式和后端存储格式
    if (config.ingressEnabled !== undefined || config.ingressHost !== undefined) {
      return {
        enabled: config.ingressEnabled === true,
        host: config.ingressHost || ''
      }
    }
    if (config.ingress && config.ingress.enabled) {
      return { enabled: true, host: config.ingress.host || '' }
    }
  } catch (e) {
    // ignore parse errors
  }
  return { enabled: false, host: '' }
}

// 格式化JSON
const formatJson = () => {
  try {
    const obj = JSON.parse(configJsonText.value)
    configJsonText.value = JSON.stringify(obj, null, 2)
    ElMessage.success('格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误')
  }
}

// 构建完整配置（使用后端存储的扁平格式，确保与部署服务兼容）
const buildFullConfig = () => {
  const config = {
    envVars: envVars.value.filter(v => v.key),
    healthCheck: advancedConfig.healthCheck.enabled ? advancedConfig.healthCheck : undefined
  }
  
  // Ingress 使用后端识别的扁平字段名
  if (advancedConfig.ingress.enabled) {
    config.ingressEnabled = true
    config.ingressHost = advancedConfig.ingress.host
    config.ingressPath = advancedConfig.ingress.path
    config.containerPort = advancedConfig.ingress.servicePort
  }
  
  return config
}

// 保存配置
const handleSaveConfig = async () => {
  try {
    let finalConfig
    
    // 如果在JSON编辑器标签页，使用JSON文本
    if (activeConfigTab.value === 'json') {
      try {
        finalConfig = JSON.parse(configJsonText.value)
      } catch (error) {
        ElMessage.error('JSON格式错误，请检查')
        return
      }
    } else {
      // 否则从表单构建配置
      finalConfig = buildFullConfig()
    }
    
    // 准备更新数据
    const updateData = {
      replicas: configForm.replicas,
      cpuRequest: configForm.cpuRequest,
      cpuLimit: configForm.cpuLimit,
      memoryRequest: configForm.memoryRequest,
      memoryLimit: configForm.memoryLimit,
      configJson: JSON.stringify(finalConfig)
    }
    
    await request.put(`/app-env-bindings/${currentBinding.value.id}`, updateData)
    
    ElMessage.success('配置保存成功')
    configDialogVisible.value = false
    loadEnvBindings()
  } catch (error) {
    console.error('保存配置失败', error)
    ElMessage.error('保存配置失败')
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
