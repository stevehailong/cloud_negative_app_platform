<template>
  <div class="user-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-input
            v-model="searchKeyword"
            placeholder="搜索用户名、邮箱或姓名"
            style="width: 300px"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #append>
              <el-button :icon="Search" @click="handleSearch" />
            </template>
          </el-input>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="userList"
        border
        stripe
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="realName" label="真实姓名" width="120" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="department" label="部门" width="120">
          <template #default="{ row }">
            {{ row.department || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="position" label="职位" width="120">
          <template #default="{ row }">
            {{ row.position || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="角色" width="200">
          <template #default="{ row }">
            <el-tag
              v-for="role in row.roles"
              :key="role.id"
              size="small"
              style="margin-right: 5px"
            >
              {{ role.name }}
            </el-tag>
            <el-tag v-if="!row.roles || row.roles.length === 0" type="info" size="small">
              未分配
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="handleAssignRoles(row)"
            >
              分配角色
            </el-button>
            <el-button
              :type="row.status === 1 ? 'warning' : 'success'"
              size="small"
              link
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchUserList"
          @current-change="fetchUserList"
        />
      </div>
    </el-card>

    <!-- 分配角色对话框 -->
    <el-dialog
      v-model="roleDialogVisible"
      title="分配角色"
      width="500px"
    >
      <el-form label-width="80px">
        <el-form-item label="用户名">
          <span>{{ currentUser?.username }}</span>
        </el-form-item>
        <el-form-item label="真实姓名">
          <span>{{ currentUser?.realName || '-' }}</span>
        </el-form-item>
        <el-form-item label="选择角色">
          <el-checkbox-group v-model="selectedRoleIds">
            <el-checkbox
              v-for="role in allRoles"
              :key="role.id"
              :label="role.id"
            >
              {{ role.name }} ({{ role.code }})
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmAssignRoles" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getUserList, updateUserStatus, assignRoles, getUserRoles, getRoleList } from '@/api/user'
import { formatTime } from '@/utils/time'

const loading = ref(false)
const submitting = ref(false)
const searchKeyword = ref('')
const userList = ref([])
const allRoles = ref([])
const currentUser = ref(null)
const selectedRoleIds = ref([])
const roleDialogVisible = ref(false)

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 获取用户列表
const fetchUserList = async () => {
  loading.value = true
  try {
    const res = await getUserList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchKeyword.value
    })
    
    // 为每个用户获取角色信息
    const users = res.data.list || []
    for (const user of users) {
      try {
        const rolesRes = await getUserRoles(user.id)
        user.roles = rolesRes.data || []
      } catch (err) {
        user.roles = []
      }
    }
    
    userList.value = users
    pagination.total = res.data.total
  } catch (error) {
    ElMessage.error(error.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchUserList()
}

// 切换用户状态
const handleToggleStatus = async (user) => {
  const action = user.status === 1 ? '禁用' : '启用'
  const newStatus = user.status === 1 ? 0 : 1
  
  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.username}" 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await updateUserStatus(user.id, newStatus)
    ElMessage.success(`${action}成功`)
    fetchUserList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || `${action}失败`)
    }
  }
}

// 打开分配角色对话框
const handleAssignRoles = async (user) => {
  currentUser.value = user
  roleDialogVisible.value = true
  
  // 获取用户当前角色
  try {
    const res = await getUserRoles(user.id)
    selectedRoleIds.value = (res.data || []).map(role => role.id)
  } catch (error) {
    selectedRoleIds.value = []
  }
}

// 确认分配角色
const handleConfirmAssignRoles = async () => {
  submitting.value = true
  try {
    await assignRoles(currentUser.value.id, selectedRoleIds.value)
    ElMessage.success('角色分配成功')
    roleDialogVisible.value = false
    fetchUserList()
  } catch (error) {
    ElMessage.error(error.message || '角色分配失败')
  } finally {
    submitting.value = false
  }
}

// 获取所有角色列表
const fetchRoleList = async () => {
  try {
    const res = await getRoleList()
    allRoles.value = res.data || []
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  }
}

onMounted(() => {
  fetchUserList()
  fetchRoleList()
})
</script>

<style scoped>
.user-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-checkbox) {
  display: block;
  margin: 10px 0;
}
</style>
