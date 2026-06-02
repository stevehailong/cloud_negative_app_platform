<template>
  <div class="pipeline-runs">
    <div class="runs-header">
      <h3>{{ pipelineName }} - 执行记录</h3>
    </div>

    <!-- 筛选栏 -->
    <el-form :inline="true" :model="queryParams" class="filter-form">
      <el-form-item label="状态">
        <el-select
          v-model="queryParams.status"
          placeholder="全部状态"
          clearable
          style="width: 150px"
          @change="handleQuery"
        >
          <el-option label="成功" value="success" />
          <el-option label="失败" value="failed" />
          <el-option label="运行中" value="running" />
          <el-option label="等待中" value="pending" />
        </el-select>
      </el-form-item>
      <el-form-item label="Git分支">
        <el-input
          v-model="queryParams.git_branch"
          placeholder="请输入分支名"
          clearable
          style="width: 150px"
          @clear="handleQuery"
        />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="Search" @click="handleQuery">
          查询
        </el-button>
        <el-button :icon="Refresh" @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <!-- 执行记录列表 -->
    <el-table
      v-loading="loading"
      :data="runList"
      style="width: 100%"
    >
      <el-table-column prop="runNo" label="执行编号" width="200" />
      
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusTag(row.status)" size="small">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column prop="triggerType" label="触发方式" width="100">
        <template #default="{ row }">
          <el-tag size="small" effect="plain">
            {{ getTriggerText(row.triggerType) }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="Git信息" min-width="200">
        <template #default="{ row }">
          <div class="git-info">
            <div class="git-branch">
              <el-icon><Share /></el-icon>
              <span>{{ row.gitBranch || '-' }}</span>
            </div>
            <div class="git-commit" v-if="row.gitCommit">
              <el-icon><DocumentCopy /></el-icon>
              <span>{{ row.gitCommit.substring(0, 8) }}</span>
            </div>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column label="构建镜像" min-width="250">
        <template #default="{ row }">
          <div v-if="row.imageUrl" style="display: flex; align-items: center; gap: 8px;">
            <el-text size="small" style="font-family: monospace; flex: 1;" truncated>
              {{ row.imageUrl }}
            </el-text>
            <el-button 
              size="small" 
              text
              :icon="DocumentCopy" 
              @click="copyToClipboard(row.imageUrl)"
              title="复制镜像地址"
            />
          </div>
          <el-text v-else-if="row.status === 'running'" type="info" size="small">构建中...</el-text>
          <el-text v-else type="info" size="small">-</el-text>
        </template>
      </el-table-column>
      
      <el-table-column label="执行时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.startTime) }}
        </template>
      </el-table-column>
      
      <el-table-column prop="durationSeconds" label="耗时" width="100">
        <template #default="{ row }">
          {{ formatDuration(row.durationSeconds) }}
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button
            type="primary"
            link
            :icon="View"
            @click="handleViewDetail(row)"
          >
            详情
          </el-button>
          <el-button
            v-if="row.logUrl"
            type="primary"
            link
            :icon="Document"
            @click="handleViewLog(row)"
          >
            日志
          </el-button>
          <el-button
            v-if="row.status === 'running'"
            type="danger"
            link
            :icon="Close"
            @click="handleStop(row)"
          >
            停止
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.pageSize"
        :page-sizes="[10, 20, 50]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @size-change="handleQuery"
        @current-change="handleQuery"
      />
    </div>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailVisible"
      title="执行详情"
      width="800px"
      destroy-on-close
    >
      <el-descriptions :column="2" border v-if="currentRun">
        <el-descriptions-item label="执行编号">
          {{ currentRun.runNo }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusTag(currentRun.status)">
            {{ getStatusText(currentRun.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="触发方式">
          {{ getTriggerText(currentRun.triggerType) }}
        </el-descriptions-item>
        <el-descriptions-item label="Git分支">
          {{ currentRun.gitBranch || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="Git Commit">
          {{ currentRun.gitCommit ? currentRun.gitCommit.substring(0, 8) : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="构建镜像" :span="2" v-if="currentRun.imageUrl || currentRun.status === 'success'">
          <div v-if="currentRun.imageUrl" style="display: flex; align-items: center; gap: 10px;">
            <el-link :href="`#`" type="primary" style="font-family: monospace;">
              {{ currentRun.imageUrl }}
            </el-link>
            <el-button 
              size="small" 
              :icon="DocumentCopy" 
              @click="copyToClipboard(currentRun.imageUrl)"
              title="复制镜像地址"
            >
              复制
            </el-button>
          </div>
          <el-text v-else type="warning">构建中，镜像地址生成后将显示在此处</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="执行时间">
          {{ formatTime(currentRun.startTime) }}
        </el-descriptions-item>
        <el-descriptions-item label="结束时间">
          {{ formatTime(currentRun.endTime) }}
        </el-descriptions-item>
        <el-descriptions-item label="耗时">
          {{ formatDuration(currentRun.durationSeconds) }}
        </el-descriptions-item>
        <el-descriptions-item label="日志地址" :span="2">
          <el-link
            v-if="currentRun.logUrl"
            :href="currentRun.logUrl"
            target="_blank"
            type="primary"
          >
            {{ currentRun.logUrl }}
          </el-link>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="失败原因" :span="2" v-if="currentRun.status === 'failed' && currentRun.errorMessage">
          <el-text type="danger">{{ currentRun.errorMessage }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Refresh,
  View,
  Document,
  Close,
  Share,
  DocumentCopy
} from '@element-plus/icons-vue'
import { getPipelineRunList, stopPipelineRun } from '@/api/pipeline'
import { formatTime, formatDuration } from '@/utils/time'

const props = defineProps({
  pipelineId: {
    type: Number,
    required: true
  },
  pipelineName: {
    type: String,
    default: ''
  }
})

// 查询参数
const queryParams = reactive({
  page: 1,
  pageSize: 10,
  pipeline_id: props.pipelineId,
  status: '',
  git_branch: ''
})

// 数据
const loading = ref(false)
const runList = ref([])
const total = ref(0)

// 详情
const detailVisible = ref(false)
const currentRun = ref(null)

// 获取状态标签
const getStatusTag = (status) => {
  const tagMap = {
    'success': 'success',
    'failed': 'danger',
    'running': 'warning',
    'pending': 'info'
  }
  return tagMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const textMap = {
    'success': '成功',
    'failed': '失败',
    'running': '运行中',
    'pending': '等待中'
  }
  return textMap[status] || status
}

// 获取触发方式文本
const getTriggerText = (type) => {
  const textMap = {
    'manual': '手动触发',
    'webhook': '自动触发(Webhook)',
    'scheduled': '定时触发',
    'api': 'API触发'
  }
  return textMap[type] || '手动触发'
}

// 获取执行记录列表
const fetchRunList = async () => {
  loading.value = true
  try {
    const { data } = await getPipelineRunList(queryParams)
    runList.value = data.list || []
    total.value = data.total || 0
  } catch (error) {
    console.error('获取执行记录失败:', error)
  } finally {
    loading.value = false
  }
}

// 查询
const handleQuery = () => {
  queryParams.page = 1
  fetchRunList()
}

// 重置
const handleReset = () => {
  Object.assign(queryParams, {
    page: 1,
    pageSize: 10,
    pipeline_id: props.pipelineId,
    status: '',
    git_branch: ''
  })
  fetchRunList()
}

// 查看详情
const handleViewDetail = (row) => {
  currentRun.value = row
  detailVisible.value = true
}

// 复制到剪贴板
const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制镜像地址到剪贴板')
  } catch (err) {
    // 降级方案：使用传统方法
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    try {
      document.execCommand('copy')
      ElMessage.success('已复制镜像地址到剪贴板')
    } catch (e) {
      ElMessage.error('复制失败')
    }
    document.body.removeChild(textarea)
  }
}

// 查看日志
const JENKINS_BASE_URL = 'http://localhost:9090'

const handleViewLog = (row) => {
  if (row.logUrl) {
    // logUrl 格式: /jenkins/job/xxx/1/console → 转换为 Jenkins 外部地址
    let url = row.logUrl
    if (url.startsWith('/jenkins/')) {
      url = JENKINS_BASE_URL + url.replace('/jenkins', '')
    } else if (url.startsWith('/')) {
      url = JENKINS_BASE_URL + url
    }
    window.open(url, '_blank')
  }
}

// 停止执行
const handleStop = (row) => {
  ElMessageBox.confirm(
    '确定要停止该流水线执行吗？',
    '停止确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await stopPipelineRun(row.id)
      ElMessage.success('已停止执行')
      fetchRunList()
    } catch (error) {
      console.error('停止失败:', error)
    }
  }).catch(() => {})
}

// 初始化
onMounted(() => {
  fetchRunList()
})
</script>

<style scoped lang="scss">
.pipeline-runs {
  .runs-header {
    margin-bottom: 20px;
    
    h3 {
      margin: 0;
      font-size: 18px;
      font-weight: 500;
      color: #303133;
    }
  }
  
  .filter-form {
    margin-bottom: 16px;
  }
  
  .git-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 13px;
    
    .git-branch,
    .git-commit {
      display: flex;
      align-items: center;
      gap: 4px;
      color: #606266;
      
      .el-icon {
        font-size: 14px;
      }
    }
  }
  
  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 20px;
  }
}
</style>
