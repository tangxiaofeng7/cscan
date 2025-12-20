import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userId = ref(localStorage.getItem('userId') || '')
  const username = ref(localStorage.getItem('username') || '')
  const role = ref(localStorage.getItem('role') || '')
  const workspaceId = ref(localStorage.getItem('workspaceId') || '')

  const isLoggedIn = computed(() => !!token.value)
  const isSuperAdmin = computed(() => role.value === 'superadmin')

  async function login(loginForm) {
    const res = await loginApi(loginForm)
    if (res.code === 0) {
      token.value = res.token
      userId.value = res.userId
      username.value = res.username
      role.value = res.role
      workspaceId.value = res.workspaceId

      localStorage.setItem('token', res.token)
      localStorage.setItem('userId', res.userId)
      localStorage.setItem('username', res.username)
      localStorage.setItem('role', res.role)
      localStorage.setItem('workspaceId', res.workspaceId)
    }
    return res
  }

  function logout() {
    token.value = ''
    userId.value = ''
    username.value = ''
    role.value = ''
    workspaceId.value = ''

    localStorage.removeItem('token')
    localStorage.removeItem('userId')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
    localStorage.removeItem('workspaceId')
  }

  function setWorkspace(id) {
    workspaceId.value = id
    localStorage.setItem('workspaceId', id)
  }

  return {
    token,
    userId,
    username,
    role,
    workspaceId,
    isLoggedIn,
    isSuperAdmin,
    login,
    logout,
    setWorkspace
  }
})
