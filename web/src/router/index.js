import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '工作台', icon: 'Odometer' }
      },
      {
        path: 'asset',
        name: 'Asset',
        component: () => import('@/views/Asset.vue'),
        meta: { title: '资产管理', icon: 'Monitor' }
      },
      {
        path: 'task',
        name: 'Task',
        component: () => import('@/views/Task.vue'),
        meta: { title: '任务管理', icon: 'List' }
      },
      {
        path: 'vul',
        name: 'Vul',
        component: () => import('@/views/Vul.vue'),
        meta: { title: '漏洞管理', icon: 'Warning' }
      },
      {
        path: 'online-search',
        name: 'OnlineSearch',
        component: () => import('@/views/OnlineSearch.vue'),
        meta: { title: '在线搜索', icon: 'Search' }
      },
      {
        path: 'workspace',
        name: 'Workspace',
        component: () => import('@/views/Workspace.vue'),
        meta: { title: '工作空间', icon: 'Folder' }
      },
      {
        path: 'worker',
        name: 'Worker',
        component: () => import('@/views/Worker.vue'),
        meta: { title: 'Worker管理', icon: 'Connection' }
      },
      {
        path: 'poc',
        name: 'Poc',
        component: () => import('@/views/Poc.vue'),
        meta: { title: 'POC管理', icon: 'Aim' }
      },
      {
        path: 'fingerprint',
        name: 'Fingerprint',
        component: () => import('@/views/Fingerprint.vue'),
        meta: { title: '指纹管理', icon: 'Stamp' }
      },
      {
        path: 'report',
        name: 'Report',
        component: () => import('@/views/Report.vue'),
        meta: { title: '扫描报告', icon: 'Document', hidden: true }
      },
      {
        path: 'user',
        name: 'User',
        component: () => import('@/views/User.vue'),
        meta: { title: '用户管理', icon: 'User', roles: ['superadmin'] }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth !== false && !userStore.token) {
    next('/login')
  } else if (to.path === '/login' && userStore.token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
