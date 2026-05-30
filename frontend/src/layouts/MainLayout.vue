<template>
  <div class="main-layout">
    <el-container>
      <el-aside width="200px">
        <div class="logo">
          <h3>My Cloud</h3>
        </div>
        <el-menu
          :default-active="activeMenu"
          router
          unique-opened
        >
          <el-menu-item index="/dashboard">
            <el-icon><HomeFilled /></el-icon>
            <span>工作台</span>
          </el-menu-item>
          
          <el-sub-menu index="/org-project">
            <template #title>
              <el-icon><FolderOpened /></el-icon>
              <span>组织与项目</span>
            </template>
            <el-menu-item index="/tenants">
              <el-icon><OfficeBuilding /></el-icon>
              <span>租户管理</span>
            </el-menu-item>
            <el-menu-item index="/projects">
              <el-icon><FolderOpened /></el-icon>
              <span>项目管理</span>
            </el-menu-item>
          </el-sub-menu>
          
          <el-menu-item index="/applications">
            <el-icon><Grid /></el-icon>
            <span>应用管理</span>
          </el-menu-item>
          
          <el-menu-item index="/pipelines">
            <el-icon><Connection /></el-icon>
            <span>流水线</span>
          </el-menu-item>
          
          <el-menu-item index="/app-deployments">
            <el-icon><Box /></el-icon>
            <span>应用部署</span>
          </el-menu-item>

          <el-menu-item index="/releases">
            <el-icon><Upload /></el-icon>
            <span>发布管理</span>
          </el-menu-item>
          
          <el-sub-menu index="/env-infra">
            <template #title>
              <el-icon><Cpu /></el-icon>
              <span>环境与基础设施</span>
            </template>
            <el-menu-item index="/environments">
              <el-icon><Cpu /></el-icon>
              <span>环境管理</span>
            </el-menu-item>
            <el-menu-item index="/env-templates">
              <el-icon><Document /></el-icon>
              <span>环境模板</span>
            </el-menu-item>
            <el-menu-item index="/config-maps">
              <el-icon><Document /></el-icon>
              <span>ConfigMap</span>
            </el-menu-item>
            <el-menu-item index="/secrets">
              <el-icon><Lock /></el-icon>
              <span>Secret</span>
            </el-menu-item>
            <el-menu-item index="/clusters">
              <el-icon><Files /></el-icon>
              <span>集群管理</span>
            </el-menu-item>
          </el-sub-menu>
          
          <el-menu-item index="/monitors">
            <el-icon><Monitor /></el-icon>
            <span>监控中心</span>
          </el-menu-item>

          <el-sub-menu index="/system">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>系统管理</span>
            </template>
            <el-menu-item index="/users">
              <el-icon><User /></el-icon>
              <span>用户管理</span>
            </el-menu-item>
            <el-menu-item index="/permissions">
              <el-icon><Lock /></el-icon>
              <span>权限管理</span>
            </el-menu-item>
            <el-menu-item index="/role-permission">
              <el-icon><Key /></el-icon>
              <span>角色权限</span>
            </el-menu-item>
            <el-menu-item index="/audit-logs">
              <el-icon><Document /></el-icon>
              <span>审计日志</span>
            </el-menu-item>
          </el-sub-menu>
          
          <el-menu-item index="/settings">
            <el-icon><Tools /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-container>
        <el-header>
          <div class="header-content">
            <div class="breadcrumb">
              <el-breadcrumb separator="/">
                <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
                <el-breadcrumb-item v-if="currentRoute.meta.title">
                  {{ currentRoute.meta.title }}
                </el-breadcrumb-item>
              </el-breadcrumb>
            </div>
            
            <div class="user-info">
              <el-dropdown @command="handleCommand">
                <span class="el-dropdown-link">
                  <el-avatar :size="32" :src="userStore.userInfo?.avatar">
                    {{ userStore.userInfo?.username?.charAt(0).toUpperCase() }}
                  </el-avatar>
                  <span class="username">{{ userStore.userInfo?.username }}</span>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                    <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-header>
        
        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessageBox } from 'element-plus'
import { 
  HomeFilled, 
  FolderOpened, 
  OfficeBuilding,
  Grid, 
  Connection, 
  Box, 
  Cpu, 
  Document,
  Files, 
  Monitor, 
  Setting, 
  User, 
  Lock, 
  Key, 
  Tools,
  Upload
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const currentRoute = computed(() => route)
const activeMenu = computed(() => route.path)

// 获取用户信息 - 只在有token时调用
if (!userStore.userInfo && userStore.token) {
  userStore.fetchUserInfo()
}

const handleCommand = (command) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      type: 'warning'
    }).then(() => {
      userStore.doLogout()
    }).catch(() => {})
  } else if (command === 'profile') {
    router.push('/settings')
  }
}
</script>

<style scoped lang="scss">
.main-layout {
  height: 100vh;
  
  .el-container {
    height: 100%;
  }
  
  .el-aside {
    background-color: #001529;
    
    .logo {
      height: 60px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      font-size: 20px;
      font-weight: bold;
      border-bottom: 1px solid rgba(255, 255, 255, 0.1);
      background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    }
    
    .el-menu {
      border-right: none;
      background-color: #001529;
      
      :deep(.el-menu-item) {
        color: rgba(255, 255, 255, 0.65);
        transition: all 0.3s;
        
        &:hover,
        &.is-active {
          background-color: #1890ff;
          color: #fff;
        }
        
        .el-icon {
          color: rgba(255, 255, 255, 0.65);
          transition: all 0.3s;
        }
        
        &:hover .el-icon,
        &.is-active .el-icon {
          color: #fff;
        }
      }
      
      :deep(.el-sub-menu__title) {
        color: rgba(255, 255, 255, 0.85) !important;
        transition: all 0.3s;
        
        &:hover {
          background-color: rgba(255, 255, 255, 0.08) !important;
          color: #fff !important;
        }
        
        .el-icon {
          color: rgba(255, 255, 255, 0.85);
          transition: all 0.3s;
        }
        
        &:hover .el-icon {
          color: #fff;
        }
      }
      
      :deep(.el-sub-menu.is-opened) {
        > .el-sub-menu__title {
          color: #fff !important;
          background-color: rgba(255, 255, 255, 0.05) !important;
          
          .el-icon {
            color: #fff;
          }
        }
      }
      
      :deep(.el-menu--inline) {
        background-color: #000c17 !important;
        
        .el-menu-item {
          padding-left: 56px !important;
          
          &:hover,
          &.is-active {
            background-color: rgba(24, 144, 255, 0.8) !important;
          }
        }
      }
    }
  }
  
  .el-header {
    background-color: #fff;
    border-bottom: 1px solid #f0f0f0;
    padding: 0 20px;
    
    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      height: 100%;
    }
    
    .user-info {
      .el-dropdown-link {
        display: flex;
        align-items: center;
        cursor: pointer;
        
        .username {
          margin-left: 8px;
        }
      }
    }
  }
  
  .el-main {
    background-color: #f0f2f5;
    padding: 20px;
  }
}
</style>
