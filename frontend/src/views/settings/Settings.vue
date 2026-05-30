<template>
  <div class="settings-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2>系统设置</h2>
        <span class="subtitle">平台配置与参数管理</span>
      </div>
    </div>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" class="settings-tabs">
      <!-- 基本设置 -->
      <el-tab-pane label="基本设置" name="basic">
        <el-card shadow="never">
          <el-form :model="basicForm" label-width="150px">
            <el-form-item label="平台名称">
              <el-input v-model="basicForm.platformName" placeholder="云原生应用研发交付平台" style="width: 400px" />
            </el-form-item>
            <el-form-item label="平台简称">
              <el-input v-model="basicForm.platformShortName" placeholder="My Cloud" style="width: 400px" />
            </el-form-item>
            <el-form-item label="平台Logo">
              <el-upload
                class="logo-uploader"
                action="#"
                :show-file-list="false"
                :auto-upload="false"
                :on-change="handleLogoChange"
                accept=".png,.jpg,.jpeg,.gif,.svg"
              >
                <img v-if="basicForm.platformLogo" :src="basicForm.platformLogo" class="logo" />
                <el-icon v-else class="logo-uploader-icon"><Plus /></el-icon>
              </el-upload>
              <div class="form-tip">建议尺寸：200x60px，支持PNG/JPG/GIF/SVG格式，最大5MB</div>
            </el-form-item>
            <el-form-item label="联系邮箱">
              <el-input v-model="basicForm.contactEmail" placeholder="support@example.com" style="width: 400px" />
            </el-form-item>
            <el-form-item label="技术支持电话">
              <el-input v-model="basicForm.supportPhone" placeholder="400-xxx-xxxx" style="width: 400px" />
            </el-form-item>
            <el-form-item label="备案号">
              <el-input v-model="basicForm.icp" placeholder="京ICP备xxxxxxxx号" style="width: 400px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="saveBasicSettings">保存</el-button>
              <el-button @click="resetBasicForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 安全设置 -->
      <el-tab-pane label="安全设置" name="security">
        <el-card shadow="never">
          <el-form :model="securityForm" label-width="180px">
            <el-divider content-position="left">认证配置</el-divider>
            <el-form-item label="Session超时时间">
              <el-input-number v-model="securityForm.sessionTimeout" :min="5" :max="1440" />
              <span class="unit-text">分钟</span>
            </el-form-item>
            <el-form-item label="密码最小长度">
              <el-input-number v-model="securityForm.passwordMinLength" :min="6" :max="20" />
              <span class="unit-text">位</span>
            </el-form-item>
            <el-form-item label="密码复杂度要求">
              <el-checkbox-group v-model="securityForm.passwordComplexity">
                <el-checkbox label="uppercase">大写字母</el-checkbox>
                <el-checkbox label="lowercase">小写字母</el-checkbox>
                <el-checkbox label="number">数字</el-checkbox>
                <el-checkbox label="special">特殊字符</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="登录失败锁定">
              <el-switch v-model="securityForm.loginLockEnabled" />
              <span class="unit-text">连续失败 {{ securityForm.loginLockAttempts }} 次锁定</span>
            </el-form-item>
            <el-form-item label="登录失败次数" v-if="securityForm.loginLockEnabled">
              <el-input-number v-model="securityForm.loginLockAttempts" :min="3" :max="10" />
              <span class="unit-text">次</span>
            </el-form-item>
            <el-form-item label="锁定时长" v-if="securityForm.loginLockEnabled">
              <el-input-number v-model="securityForm.loginLockDuration" :min="5" :max="120" />
              <span class="unit-text">分钟</span>
            </el-form-item>

            <el-divider content-position="left">API安全</el-divider>
            <el-form-item label="API限流">
              <el-switch v-model="securityForm.apiRateLimitEnabled" />
            </el-form-item>
            <el-form-item label="限流阈值" v-if="securityForm.apiRateLimitEnabled">
              <el-input-number v-model="securityForm.apiRateLimit" :min="10" :max="10000" />
              <span class="unit-text">请求/分钟</span>
            </el-form-item>
            <el-form-item label="IP白名单">
              <el-input
                v-model="securityForm.ipWhitelist"
                type="textarea"
                :rows="4"
                placeholder="每行一个IP地址或CIDR，例如：192.168.1.1 或 192.168.1.0/24"
                style="width: 500px"
              />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :loading="saving" @click="saveSecuritySettings">保存</el-button>
              <el-button @click="resetSecurityForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 通知设置 -->
      <el-tab-pane label="通知设置" name="notification">
        <el-card shadow="never">
          <el-form :model="notificationForm" label-width="150px">
            <el-divider content-position="left">邮件配置</el-divider>
            <el-form-item label="SMTP服务器">
              <el-input v-model="notificationForm.smtpHost" placeholder="smtp.example.com" style="width: 400px" />
            </el-form-item>
            <el-form-item label="SMTP端口">
              <el-input-number v-model="notificationForm.smtpPort" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item label="发件人邮箱">
              <el-input v-model="notificationForm.smtpFrom" placeholder="noreply@example.com" style="width: 400px" />
            </el-form-item>
            <el-form-item label="SMTP用户名">
              <el-input v-model="notificationForm.smtpUsername" placeholder="smtp账号" style="width: 400px" />
            </el-form-item>
            <el-form-item label="SMTP密码">
              <el-input
                v-model="notificationForm.smtpPassword"
                type="password"
                placeholder="smtp密码"
                show-password
                style="width: 400px"
              />
            </el-form-item>
            <el-form-item label="启用SSL/TLS">
              <el-switch v-model="notificationForm.smtpSSL" />
            </el-form-item>
            <el-form-item>
              <el-button @click="testEmail">测试邮件发送</el-button>
            </el-form-item>

            <el-divider content-position="left">企业微信配置</el-divider>
            <el-form-item label="企业ID">
              <el-input v-model="notificationForm.wechatCorpId" placeholder="企业微信CorpID" style="width: 400px" />
            </el-form-item>
            <el-form-item label="应用AgentId">
              <el-input v-model="notificationForm.wechatAgentId" placeholder="应用AgentId" style="width: 400px" />
            </el-form-item>
            <el-form-item label="应用Secret">
              <el-input
                v-model="notificationForm.wechatSecret"
                type="password"
                placeholder="应用Secret"
                show-password
                style="width: 400px"
              />
            </el-form-item>

            <el-divider content-position="left">钉钉配置</el-divider>
            <el-form-item label="机器人Webhook">
              <el-input
                v-model="notificationForm.dingtalkWebhook"
                placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx"
                style="width: 500px"
              />
            </el-form-item>
            <el-form-item label="签名密钥">
              <el-input
                v-model="notificationForm.dingtalkSecret"
                type="password"
                placeholder="机器人签名密钥"
                show-password
                style="width: 400px"
              />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :loading="saving" @click="saveNotificationSettings">保存</el-button>
              <el-button @click="resetNotificationForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 集成配置 -->
      <el-tab-pane label="集成配置" name="integration">
        <el-card shadow="never">
          <el-form :model="integrationForm" label-width="150px">
            <el-divider content-position="left">GitLab配置</el-divider>
            <el-form-item label="GitLab地址">
              <el-input v-model="integrationForm.gitlabUrl" placeholder="https://gitlab.example.com" style="width: 400px" />
            </el-form-item>
            <el-form-item label="访问Token">
              <el-input
                v-model="integrationForm.gitlabToken"
                type="password"
                placeholder="GitLab Personal Access Token"
                show-password
                style="width: 400px"
              />
            </el-form-item>
            <el-form-item label="">
              <el-button :loading="testingGitlab" @click="testGitlabConnection">
                测试连接
              </el-button>
              <span v-if="gitlabTestResult" :class="['test-result', gitlabTestResult.success ? 'success' : 'error']">
                {{ gitlabTestResult.message }}
              </span>
            </el-form-item>

            <el-divider content-position="left">Harbor配置</el-divider>
            <el-form-item label="Harbor地址">
              <el-input v-model="integrationForm.harborUrl" placeholder="https://harbor.example.com" style="width: 400px" />
            </el-form-item>
            <el-form-item label="Harbor用户名">
              <el-input v-model="integrationForm.harborUsername" placeholder="admin" style="width: 400px" />
            </el-form-item>
            <el-form-item label="Harbor密码">
              <el-input
                v-model="integrationForm.harborPassword"
                type="password"
                placeholder="Harbor密码"
                show-password
                style="width: 400px"
              />
            </el-form-item>

            <el-divider content-position="left">Prometheus配置</el-divider>
            <el-form-item label="Prometheus地址">
              <el-input v-model="integrationForm.prometheusUrl" placeholder="http://prometheus:9090" style="width: 400px" />
            </el-form-item>

            <el-divider content-position="left">Grafana配置</el-divider>
            <el-form-item label="Grafana地址">
              <el-input v-model="integrationForm.grafanaUrl" placeholder="http://grafana:3000" style="width: 400px" />
            </el-form-item>
            <el-form-item label="API Key">
              <el-input
                v-model="integrationForm.grafanaApiKey"
                type="password"
                placeholder="Grafana API Key"
                show-password
                style="width: 400px"
              />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :loading="saving" @click="saveIntegrationSettings">保存</el-button>
              <el-button @click="resetIntegrationForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 系统信息 -->
      <el-tab-pane label="系统信息" name="system">
        <el-card shadow="never">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="平台版本">{{ systemInfo.version }}</el-descriptions-item>
            <el-descriptions-item label="构建时间">{{ systemInfo.buildTime }}</el-descriptions-item>
            <el-descriptions-item label="Go版本">{{ systemInfo.goVersion }}</el-descriptions-item>
            <el-descriptions-item label="Git Commit">{{ systemInfo.gitCommit }}</el-descriptions-item>
            <el-descriptions-item label="启动时间">{{ systemInfo.startTime }}</el-descriptions-item>
            <el-descriptions-item label="运行时长">{{ systemInfo.uptime }}</el-descriptions-item>
            <el-descriptions-item label="操作系统">{{ systemInfo.os }}</el-descriptions-item>
            <el-descriptions-item label="架构">{{ systemInfo.arch }}</el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">服务状态</el-divider>
          <el-table :data="systemInfo.services" style="width: 100%">
            <el-table-column prop="name" label="服务名称" width="250" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'running' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'running' ? '运行中' : '已停止' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="version" label="版本" width="150" />
            <el-table-column prop="url" label="访问地址" />
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'

const activeTab = ref('basic')
const saving = ref(false)
const testingGitlab = ref(false)
const gitlabTestResult = ref(null)

// 基本设置
const basicForm = reactive({
  platformName: '云原生应用研发交付平台',
  platformShortName: 'My Cloud',
  platformLogo: '',
  contactEmail: 'support@example.com',
  supportPhone: '400-xxx-xxxx',
  icp: ''
})

// 安全设置
const securityForm = reactive({
  sessionTimeout: 30,
  passwordMinLength: 8,
  passwordComplexity: ['lowercase', 'number'],
  loginLockEnabled: true,
  loginLockAttempts: 5,
  loginLockDuration: 30,
  apiRateLimitEnabled: true,
  apiRateLimit: 1000,
  ipWhitelist: ''
})

// 通知设置
const notificationForm = reactive({
  smtpHost: '',
  smtpPort: 465,
  smtpFrom: '',
  smtpUsername: '',
  smtpPassword: '',
  smtpSSL: true,
  wechatCorpId: '',
  wechatAgentId: '',
  wechatSecret: '',
  dingtalkWebhook: '',
  dingtalkSecret: ''
})

// 集成配置
const integrationForm = reactive({
  gitlabUrl: '',
  gitlabToken: '',
  harborUrl: '',
  harborUsername: '',
  harborPassword: '',
  prometheusUrl: 'http://prometheus:9090',
  grafanaUrl: 'http://grafana:3000',
  grafanaApiKey: ''
})

// 系统信息
const systemInfo = ref({
  version: 'v1.0.0',
  buildTime: '2026-05-28 12:00:00',
  goVersion: 'go1.22',
  gitCommit: 'abc123',
  startTime: '2026-05-28 08:00:00',
  uptime: '4小时30分',
  os: 'linux',
  arch: 'amd64',
  services: [
    { name: 'gateway', status: 'running', version: 'v1.0.0', url: 'http://localhost:8080' },
    { name: 'auth-service', status: 'running', version: 'v1.0.0', url: 'http://localhost:8081' },
    { name: 'project-service', status: 'running', version: 'v1.0.0', url: 'http://localhost:8082' },
    { name: 'application-service', status: 'running', version: 'v1.0.0', url: 'http://localhost:8083' },
    { name: 'pipeline-service', status: 'running', version: 'v1.0.0', url: 'http://localhost:8084' }
  ]
})

// Logo 上传处理
const handleLogoChange = async (uploadFile) => {
  const file = uploadFile.raw
  // 校验文件类型
  const allowedTypes = ['image/png', 'image/jpeg', 'image/gif', 'image/svg+xml']
  if (!allowedTypes.includes(file.type)) {
    ElMessage.error('仅支持PNG/JPG/GIF/SVG格式')
    return
  }
  // 校验文件大小
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过5MB')
    return
  }

  // 上传文件
  const formData = new FormData()
  formData.append('file', file)
  try {
    const res = await request({
      url: '/upload/',
      method: 'post',
      data: formData,
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    basicForm.platformLogo = res.data.url
    ElMessage.success('Logo上传成功')
  } catch (error) {
    console.error('上传失败:', error)
    ElMessage.error('Logo上传失败')
  }
}

// 保存基本设置
const saveBasicSettings = async () => {
  saving.value = true
  try {
    await request({
      url: '/settings/basic/',
      method: 'put',
      data: basicForm
    })
    ElMessage.success('基本设置保存成功')
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

// 保存安全设置
const saveSecuritySettings = async () => {
  saving.value = true
  try {
    await request({
      url: '/settings/security/',
      method: 'put',
      data: {
        ...securityForm,
        passwordComplexity: JSON.stringify(securityForm.passwordComplexity),
        loginLockEnabled: String(securityForm.loginLockEnabled),
        apiRateLimitEnabled: String(securityForm.apiRateLimitEnabled),
        sessionTimeout: String(securityForm.sessionTimeout),
        passwordMinLength: String(securityForm.passwordMinLength),
        loginLockAttempts: String(securityForm.loginLockAttempts),
        loginLockDuration: String(securityForm.loginLockDuration),
        apiRateLimit: String(securityForm.apiRateLimit)
      }
    })
    ElMessage.success('安全设置保存成功')
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

// 保存通知设置
const saveNotificationSettings = async () => {
  saving.value = true
  try {
    await request({
      url: '/settings/notification/',
      method: 'put',
      data: {
        ...notificationForm,
        smtpPort: String(notificationForm.smtpPort),
        smtpSSL: String(notificationForm.smtpSSL)
      }
    })
    ElMessage.success('通知设置保存成功')
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

// 保存集成配置
const saveIntegrationSettings = async () => {
  saving.value = true
  try {
    await request({
      url: '/settings/integration/',
      method: 'put',
      data: integrationForm
    })
    // 同步更新pipeline-service的GitLab客户端
    if (integrationForm.gitlabUrl && integrationForm.gitlabToken) {
      try {
        await request({
          url: '/gitlab/client',
          method: 'put',
          data: {
            gitlabUrl: integrationForm.gitlabUrl,
            gitlabToken: integrationForm.gitlabToken
          }
        })
      } catch (e) {
        // 非致命错误，GitLab服务可能未就绪
        console.warn('同步GitLab客户端失败:', e)
      }
    }
    ElMessage.success('集成配置保存成功')
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

// 测试GitLab连接
const testGitlabConnection = async () => {
  if (!integrationForm.gitlabUrl || !integrationForm.gitlabToken) {
    ElMessage.warning('请先填写GitLab地址和Token')
    return
  }
  testingGitlab.value = true
  gitlabTestResult.value = null
  try {
    // 先动态更新客户端配置
    await request({
      url: '/gitlab/client',
      method: 'put',
      data: {
        gitlabUrl: integrationForm.gitlabUrl,
        gitlabToken: integrationForm.gitlabToken
      }
    })
    // 再测试连接
    const res = await request({ url: '/gitlab/test', method: 'post' })
    gitlabTestResult.value = {
      success: true,
      message: `连接成功！用户: ${res.data.name || res.data.username}`
    }
  } catch (error) {
    gitlabTestResult.value = {
      success: false,
      message: error.response?.data?.message || '连接失败'
    }
  } finally {
    testingGitlab.value = false
  }
}

// 测试邮件
const testEmail = async () => {
  try {
    await request({
      url: '/settings/notification/test-email',
      method: 'post',
      data: {
        to: basicForm.contactEmail
      }
    })
    ElMessage.success('测试邮件已发送')
  } catch (error) {
    console.error('测试失败:', error)
  }
}

// 重置表单
const resetBasicForm = () => {
  loadGroupSettings('basic')
}

const resetSecurityForm = () => {
  loadGroupSettings('security')
}

const resetNotificationForm = () => {
  loadGroupSettings('notification')
}

const resetIntegrationForm = () => {
  loadGroupSettings('integration')
}

// 加载指定分组的设置
const loadGroupSettings = async (group) => {
  try {
    const res = await request({ url: `/settings/${group}/`, method: 'get' })
    const data = res.data
    if (!data) return

    if (group === 'basic') {
      Object.keys(basicForm).forEach(key => {
        if (data[key] !== undefined && data[key] !== '') {
          basicForm[key] = data[key]
        }
      })
    } else if (group === 'security') {
      if (data.sessionTimeout) securityForm.sessionTimeout = Number(data.sessionTimeout)
      if (data.passwordMinLength) securityForm.passwordMinLength = Number(data.passwordMinLength)
      if (data.passwordComplexity) {
        try { securityForm.passwordComplexity = JSON.parse(data.passwordComplexity) } catch (e) {}
      }
      if (data.loginLockEnabled !== undefined) securityForm.loginLockEnabled = data.loginLockEnabled === 'true'
      if (data.loginLockAttempts) securityForm.loginLockAttempts = Number(data.loginLockAttempts)
      if (data.loginLockDuration) securityForm.loginLockDuration = Number(data.loginLockDuration)
      if (data.apiRateLimitEnabled !== undefined) securityForm.apiRateLimitEnabled = data.apiRateLimitEnabled === 'true'
      if (data.apiRateLimit) securityForm.apiRateLimit = Number(data.apiRateLimit)
      if (data.ipWhitelist !== undefined) securityForm.ipWhitelist = data.ipWhitelist
    } else if (group === 'notification') {
      Object.keys(notificationForm).forEach(key => {
        if (data[key] !== undefined) {
          if (key === 'smtpPort') {
            notificationForm[key] = Number(data[key])
          } else if (key === 'smtpSSL') {
            notificationForm[key] = data[key] === 'true'
          } else {
            notificationForm[key] = data[key]
          }
        }
      })
    } else if (group === 'integration') {
      Object.keys(integrationForm).forEach(key => {
        if (data[key] !== undefined) {
          integrationForm[key] = data[key]
        }
      })
    }
  } catch (error) {
    // 首次加载可能404，忽略
  }
}

// 加载所有设置
const loadSettings = async () => {
  await Promise.all([
    loadGroupSettings('basic'),
    loadGroupSettings('security'),
    loadGroupSettings('notification'),
    loadGroupSettings('integration')
  ])
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped lang="scss">
.settings-container {
  padding: 20px;
}

.page-header {
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

.settings-tabs {
  :deep(.el-tabs__content) {
    padding-top: 20px;
  }
}

.unit-text {
  margin-left: 8px;
  color: #909399;
  font-size: 14px;
}

.form-tip {
  margin-top: 5px;
  font-size: 12px;
  color: #909399;
}

.test-result {
  margin-left: 12px;
  font-size: 13px;
  &.success {
    color: #67c23a;
  }
  &.error {
    color: #f56c6c;
  }
}

.logo-uploader {
  :deep(.el-upload) {
    border: 1px dashed #d9d9d9;
    border-radius: 6px;
    cursor: pointer;
    position: relative;
    overflow: hidden;
    transition: border-color 0.3s;
    
    &:hover {
      border-color: #409eff;
    }
  }
  
  .logo-uploader-icon {
    font-size: 28px;
    color: #8c939d;
    width: 200px;
    height: 60px;
    line-height: 60px;
    text-align: center;
  }
  
  .logo {
    width: 200px;
    height: 60px;
    display: block;
    object-fit: contain;
  }
}
</style>
