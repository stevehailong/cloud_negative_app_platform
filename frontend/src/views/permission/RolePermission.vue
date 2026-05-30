<template>
  <div class="role-permission">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色权限管理</span>
        </div>
      </template>

      <el-row :gutter="20">
        <!-- 左侧角色列表 -->
        <el-col :span="6">
          <el-card shadow="never">
            <template #header>角色列表</template>
            <el-menu :default-active="selectedRoleId" @select="handleRoleSelect">
              <el-menu-item
                v-for="role in roles"
                :key="role.id"
                :index="String(role.id)"
              >
                {{ role.name }} ({{ role.code }})
              </el-menu-item>
            </el-menu>
          </el-card>
        </el-col>

        <!-- 右侧权限列表 -->
        <el-col :span="18">
          <el-card shadow="never">
            <template #header>
              <div class="permission-header">
                <span>权限配置</span>
                <el-button
                  type="primary"
                  :disabled="!selectedRoleId"
                  @click="handleSave"
                >
                  保存
                </el-button>
              </div>
            </template>

            <div v-if="!selectedRoleId" class="empty-tip">
              请选择左侧角色查看权限
            </div>

            <div v-else>
              <el-collapse v-model="activeCollapse" accordion>
                <el-collapse-item
                  v-for="(perms, resourceType) in groupedPermissions"
                  :key="resourceType"
                  :name="resourceType"
                >
                  <template #title>
                    <div class="collapse-title">
                      <span>{{ resourceTypeMap[resourceType] || resourceType }}</span>
                      <span class="count">
                        ({{ checkedCount(resourceType) }}/{{ perms.length }})
                      </span>
                    </div>
                  </template>
                  <el-checkbox-group v-model="selectedPermissions">
                    <div class="permission-grid">
                      <el-checkbox
                        v-for="perm in perms"
                        :key="perm.id"
                        :label="perm.id"
                        class="permission-item"
                      >
                        <div class="perm-info">
                          <div class="perm-name">{{ perm.name }}</div>
                          <div class="perm-detail">
                            <el-tag size="small" type="info">{{ perm.httpMethod }}</el-tag>
                            <span class="perm-path">{{ perm.path }}</span>
                          </div>
                        </div>
                      </el-checkbox>
                    </div>
                  </el-checkbox-group>
                </el-collapse-item>
              </el-collapse>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const roles = ref([])
const allPermissions = ref([])
const selectedRoleId = ref('')
const selectedPermissions = ref([])
const activeCollapse = ref('')

const resourceTypeMap = {
  application: '应用管理',
  component: '组件管理',
  user: '用户管理',
  role: '角色管理',
  permission: '权限管理',
  deployment: '部署管理',
  pipeline: '流水线管理',
  environment: '环境管理',
  cluster: '集群管理',
  project: '项目管理',
  monitor: '监控管理'
}

// 按资源类型分组权限
const groupedPermissions = computed(() => {
  const groups = {}
  allPermissions.value.forEach(perm => {
    const type = perm.resourceType || 'other'
    if (!groups[type]) {
      groups[type] = []
    }
    groups[type].push(perm)
  })
  return groups
})

// 统计每个资源类型已选中的权限数
const checkedCount = (resourceType) => {
  const perms = groupedPermissions.value[resourceType] || []
  return perms.filter(p => selectedPermissions.value.includes(p.id)).length
}

// 加载角色列表
const loadRoles = async () => {
  try {
    const { data } = await request.get('/roles/')
    roles.value = data || []
  } catch (error) {
    ElMessage.error('加载角色列表失败')
  }
}

// 加载所有权限
const loadAllPermissions = async () => {
  try {
    const { data } = await request.get('/permissions/', {
      params: { page: 1, pageSize: 1000 }
    })
    allPermissions.value = data.items || []
  } catch (error) {
    ElMessage.error('加载权限列表失败')
  }
}

// 加载角色的权限
const loadRolePermissions = async (roleId) => {
  try {
    const { data } = await request.get(`/roles/${roleId}/permissions/`)
    selectedPermissions.value = (data || []).map(p => p.id)
  } catch (error) {
    ElMessage.error('加载角色权限失败')
  }
}

// 选择角色
const handleRoleSelect = (roleId) => {
  selectedRoleId.value = roleId
  loadRolePermissions(roleId)
}

// 保存权限配置
const handleSave = async () => {
  try {
    await request.post(`/roles/${selectedRoleId.value}/permissions/`, {
      permissionIds: selectedPermissions.value
    })
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '保存失败')
  }
}

onMounted(() => {
  loadRoles()
  loadAllPermissions()
})
</script>

<style scoped>
.role-permission {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.permission-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.empty-tip {
  text-align: center;
  padding: 60px 0;
  color: #999;
  font-size: 14px;
}

.collapse-title {
  display: flex;
  align-items: center;
  font-weight: 600;
  font-size: 15px;
}

.count {
  margin-left: 10px;
  font-size: 13px;
  color: #409eff;
  font-weight: normal;
}

.permission-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 15px;
  padding: 10px;
}

.permission-item {
  margin: 0;
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  transition: all 0.3s;
}

.permission-item:hover {
  border-color: #409eff;
  background-color: #f5f7fa;
}

.perm-info {
  margin-left: 8px;
}

.perm-name {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 5px;
}

.perm-detail {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #909399;
}

.perm-path {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.el-menu {
  border: none;
}

.el-menu-item {
  font-size: 14px;
}
</style>
