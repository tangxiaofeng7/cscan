import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useWorkspaceStore } from '@/stores/workspace'
import router from '@/router'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    const workspaceStore = useWorkspaceStore()
    if (userStore.token) {
      config.headers['Authorization'] = `Bearer ${userStore.token}`
    }
    // 优先使用 workspaceStore 的当前工作空间ID
    const workspaceId = workspaceStore.currentWorkspaceId || userStore.workspaceId
    if (workspaceId) {
      config.headers['X-Workspace-Id'] = workspaceId
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code === 401) {
      const userStore = useUserStore()
      userStore.logout()
      router.push('/login')
      ElMessage.error('登录已过期，请重新登录')
      return Promise.reject(new Error('Unauthorized'))
    }
    return res
  },
  error => {
    ElMessage.error(error.message || '请求失败')
    return Promise.reject(error)
  }
)

export default request
