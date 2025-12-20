import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '@/api/request'

export const useWorkspaceStore = defineStore('workspace', () => {
  const workspaces = ref([])
  const currentWorkspaceId = ref(localStorage.getItem('currentWorkspaceId') || '')
  const loading = ref(false)

  // 加载工作空间列表
  async function loadWorkspaces() {
    if (loading.value) return
    loading.value = true
    try {
      const res = await request.post('/workspace/list', { page: 1, pageSize: 100 })
      if (res.code === 0) {
        workspaces.value = res.list || []
        // 如果当前选中的工作空间不存在，清空选择
        if (currentWorkspaceId.value && !workspaces.value.find(w => w.id === currentWorkspaceId.value)) {
          currentWorkspaceId.value = ''
          localStorage.removeItem('currentWorkspaceId')
        }
      }
    } finally {
      loading.value = false
    }
  }

  // 设置当前工作空间
  function setCurrentWorkspace(id) {
    currentWorkspaceId.value = id
    if (id) {
      localStorage.setItem('currentWorkspaceId', id)
    } else {
      localStorage.removeItem('currentWorkspaceId')
    }
  }

  // 获取当前工作空间名称
  function getCurrentWorkspaceName() {
    if (!currentWorkspaceId.value) return '默认工作空间'
    const ws = workspaces.value.find(w => w.id === currentWorkspaceId.value)
    return ws ? ws.name : '默认工作空间'
  }

  return {
    workspaces,
    currentWorkspaceId,
    loading,
    loadWorkspaces,
    setCurrentWorkspace,
    getCurrentWorkspaceName
  }
})
