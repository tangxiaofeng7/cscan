import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '@/api/request'

export const useWorkspaceStore = defineStore('workspace', () => {
  const workspaces = ref([])
  const currentWorkspaceId = ref(localStorage.getItem('currentWorkspaceId') || 'all')
  const loading = ref(false)

  // 获取有效的工作空间ID
  const effectiveWorkspaceId = computed(() => {
    // 'all' 表示全部空间，传空字符串给后端
    if (currentWorkspaceId.value === 'all' || !currentWorkspaceId.value) {
      return ''
    }
    return currentWorkspaceId.value
  })

  // 加载工作空间列表
  async function loadWorkspaces() {
    if (loading.value) return
    loading.value = true
    try {
      const res = await request.post('/workspace/list', { page: 1, pageSize: 100 })
      if (res.code === 0) {
        workspaces.value = res.list || []
        // 如果当前选中的工作空间不存在且不是 'all'，重置为 'all'
        if (currentWorkspaceId.value && 
            currentWorkspaceId.value !== 'all' && 
            !workspaces.value.find(w => w.id === currentWorkspaceId.value)) {
          currentWorkspaceId.value = 'all'
          localStorage.setItem('currentWorkspaceId', 'all')
        }
      }
    } finally {
      loading.value = false
    }
  }

  // 设置当前工作空间
  function setCurrentWorkspace(id) {
    currentWorkspaceId.value = id || 'all'
    localStorage.setItem('currentWorkspaceId', currentWorkspaceId.value)
  }

  // 获取当前工作空间名称
  function getCurrentWorkspaceName() {
    if (!currentWorkspaceId.value || currentWorkspaceId.value === 'all') {
      return '全部空间'
    }
    const ws = workspaces.value.find(w => w.id === currentWorkspaceId.value)
    return ws ? ws.name : '全部空间'
  }

  return {
    workspaces,
    currentWorkspaceId,
    effectiveWorkspaceId,
    loading,
    loadWorkspaces,
    setCurrentWorkspace,
    getCurrentWorkspaceName
  }
})
