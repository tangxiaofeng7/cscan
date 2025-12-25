<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '220px'" class="aside">
      <div class="logo">
        <img src="/logo.png" alt="logo" />
        <span v-show="!isCollapse">CSCAN</span>
      </div>
      <el-menu
        :default-active="$route.path"
        :collapse="isCollapse"
        router
        :background-color="themeStore.isDark ? '#1d1e1f' : '#fff'"
        :text-color="themeStore.isDark ? '#a3a6ad' : '#606266'"
        :active-text-color="themeStore.isDark ? '#fff' : '#409eff'"
        :unique-opened="true"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>工作台</template>
        </el-menu-item>
        
        <!-- 扫描管理分组 -->
        <el-sub-menu index="scan">
          <template #title>
            <el-icon><Cpu /></el-icon>
            <span>扫描管理</span>
          </template>
          <el-menu-item index="/asset">
            <el-icon><Monitor /></el-icon>
            <template #title>资产管理</template>
          </el-menu-item>
          <el-menu-item index="/task">
            <el-icon><List /></el-icon>
            <template #title>任务管理</template>
          </el-menu-item>
          <el-menu-item index="/vul">
            <el-icon><Warning /></el-icon>
            <template #title>漏洞管理</template>
          </el-menu-item>
          <el-menu-item index="/online-search">
            <el-icon><Search /></el-icon>
            <template #title>在线搜索</template>
          </el-menu-item>
        </el-sub-menu>
        
        <!-- 策略管理分组 -->
        <el-sub-menu index="policy">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>策略管理</span>
          </template>
          <el-menu-item index="/poc">
            <el-icon><Aim /></el-icon>
            <template #title>POC管理</template>
          </el-menu-item>
          <el-menu-item index="/fingerprint">
            <el-icon><Stamp /></el-icon>
            <template #title>指纹管理</template>
          </el-menu-item>

        </el-sub-menu>
        
        <!-- 系统管理分组 -->
        <el-sub-menu index="system">
          <template #title>
            <el-icon><Tools /></el-icon>
            <span>系统管理</span>
          </template>
          <el-menu-item index="/worker">
            <el-icon><Connection /></el-icon>
            <template #title>Worker管理</template>
          </el-menu-item>
          <el-menu-item index="/workspace">
            <el-icon><Folder /></el-icon>
            <template #title>工作空间</template>
          </el-menu-item>
          <el-menu-item index="/user">
            <el-icon><User /></el-icon>
            <template #title>用户管理</template>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>

    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <!-- 工作空间选择器 -->
          <el-select 
            v-model="workspaceStore.currentWorkspaceId" 
            placeholder="默认工作空间" 
            clearable
            style="width: 160px; margin-right: 16px;"
            @change="handleWorkspaceChange"
          >
            <el-option 
              v-for="ws in workspaceStore.workspaces" 
              :key="ws.id" 
              :label="ws.name" 
              :value="ws.id" 
            />
          </el-select>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ $route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <!-- 主题切换 -->
          <div class="theme-switch" @click="themeStore.toggleTheme">
            <el-icon v-if="themeStore.isDark"><Sunny /></el-icon>
            <el-icon v-else><Moon /></el-icon>
          </div>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" icon="User" />
              <span class="username">{{ userStore.username }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容区 -->
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'
import { useWorkspaceStore } from '@/stores/workspace'
import { Setting, Sunny, Moon, Cpu, Tools } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const themeStore = useThemeStore()
const workspaceStore = useWorkspaceStore()
const isCollapse = ref(false)

onMounted(() => {
  workspaceStore.loadWorkspaces()
})

function handleWorkspaceChange(val) {
  workspaceStore.setCurrentWorkspace(val)
  // 触发页面刷新数据
  window.dispatchEvent(new CustomEvent('workspace-changed', { detail: val }))
}

function handleCommand(command) {
  if (command === 'logout') {
    userStore.logout()
    router.push('/login')
  }
}
</script>

<style lang="scss" scoped>
.layout-container {
  height: 100vh;
}

.aside {
  background: var(--bg-secondary);
  transition: width 0.3s, background 0.3s;
  overflow: hidden;
  box-shadow: 2px 0 8px var(--shadow-color);
  border-right: 1px solid var(--border-color);

  .logo {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-primary);
    font-size: 18px;
    font-weight: 600;
    letter-spacing: 2px;
    border-bottom: 1px solid var(--border-color);

    img {
      width: 36px;
      height: 36px;
      margin-right: 10px;
    }
  }

  .el-menu {
    border-right: none;
    background: transparent !important;
    
    .el-menu-item {
      margin: 4px 8px;
      border-radius: 8px;
      transition: all 0.3s;
      
      &:hover {
        background: var(--bg-hover) !important;
      }
      
      &.is-active {
        background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%) !important;
        color: #fff !important;
      }
    }
    
    .el-sub-menu {
      .el-sub-menu__title {
        margin: 4px 8px;
        border-radius: 8px;
        
        &:hover {
          background: var(--bg-hover) !important;
        }
      }
      
      .el-menu-item {
        padding-left: 50px !important;
        min-width: auto;
      }
    }
  }
}

.header {
  background: var(--bg-secondary);
  box-shadow: 0 1px 4px var(--shadow-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.3s;

  .header-left {
    display: flex;
    align-items: center;

    .collapse-btn {
      font-size: 20px;
      cursor: pointer;
      margin-right: 20px;
      color: var(--text-secondary);
      transition: color 0.3s;
      
      &:hover {
        color: var(--primary-color);
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;
    
    .theme-switch {
      width: 36px;
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 8px;
      cursor: pointer;
      color: var(--text-secondary);
      transition: all 0.3s;
      
      &:hover {
        background: var(--bg-hover);
        color: var(--primary-color);
      }
      
      .el-icon {
        font-size: 18px;
      }
    }
    
    .user-info {
      display: flex;
      align-items: center;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: 8px;
      transition: background 0.3s;
      
      &:hover {
        background: var(--bg-hover);
      }

      .username {
        margin-left: 8px;
        color: var(--text-secondary);
      }
    }
  }
}

.main {
  background: var(--bg-primary);
  padding: 20px;
  overflow-y: auto;
  transition: background 0.3s;
}
</style>
