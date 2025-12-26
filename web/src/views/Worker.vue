<template>
  <div class="worker-page">
    <el-card class="action-card">
      <el-button type="primary" @click="loadData" :loading="loading">
        <el-icon><Refresh /></el-icon>刷新状态
      </el-button>
      <span v-if="loading" style="margin-left: 10px; color: #909399; font-size: 12px;">正在查询Worker实时状态...</span>
      <el-switch 
        v-model="autoRefresh" 
        active-text="自动刷新(10s)" 
        style="margin-left: 15px"
        @change="toggleAutoRefresh"
      />
    </el-card>

    <el-card style="margin-bottom: 20px">
      <el-table :data="tableData" v-loading="loading" stripe max-height="500">
        <el-table-column prop="name" label="Worker名称" min-width="150">
          <template #default="{ row }">
            <span 
              class="editable-name" 
              @click="openRenameDialog(row)"
              :title="'点击修改名称'"
            >
              {{ row.name }}
              <el-icon class="edit-icon"><Edit /></el-icon>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP地址" width="140">
          <template #default="{ row }">
            {{ row.ip || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="cpuLoad" label="CPU负载" width="120">
          <template #default="{ row }">
            <el-progress :percentage="Math.round(row.cpuLoad)" :stroke-width="10" :color="getLoadColor(row.cpuLoad)" />
          </template>
        </el-table-column>
        <el-table-column prop="memUsed" label="内存使用" width="120">
          <template #default="{ row }">
            <el-progress :percentage="Math.round(row.memUsed)" :stroke-width="10" :color="getLoadColor(row.memUsed)" />
          </template>
        </el-table-column>
        <el-table-column prop="taskCount" label="已执行任务" width="100" />
        <el-table-column prop="runningCount" label="正在执行" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.runningCount > 0" type="warning">{{ row.runningCount }}</el-tag>
            <span v-else>0</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'running' ? 'success' : 'danger'">
              {{ row.status === 'running' ? '运行中' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updateTime" label="最后响应" width="160" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-popconfirm
              title="确定要重启该Worker吗？"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="restartWorker(row.name)"
            >
              <template #reference>
                <el-button size="small" type="warning" :icon="RefreshRight" :disabled="row.status !== 'running'">重启</el-button>
              </template>
            </el-popconfirm>
            <el-popconfirm
              title="确定要删除该Worker吗？这将停止Worker并清除其数据"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="deleteWorker(row.name)"
            >
              <template #reference>
                <el-button size="small" type="danger" :icon="Delete">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && tableData.length === 0" description="暂无Worker节点" />
    </el-card>

    <!-- 重命名对话框 -->
    <el-dialog v-model="renameDialogVisible" title="修改Worker名称" width="400px">
      <el-form :model="renameForm" label-width="80px">
        <el-form-item label="原名称">
          <el-input v-model="renameForm.oldName" disabled />
        </el-form-item>
        <el-form-item label="新名称">
          <el-input v-model="renameForm.newName" placeholder="请输入新的Worker名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renameDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRename" :loading="renameLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 实时日志 -->
    <el-card>
      <template #header>
        <div class="log-header">
          <span>Worker运行日志</span>
          <div class="log-filters">
            <el-select 
              v-model="filterWorker" 
              placeholder="筛选Worker" 
              clearable 
              size="small"
              style="width: 150px; margin-right: 10px"
            >
              <el-option label="全部Worker" value="" />
              <el-option 
                v-for="worker in tableData" 
                :key="worker.name" 
                :label="worker.name" 
                :value="worker.name" 
              />
            </el-select>
            <el-select 
              v-model="filterLevel" 
              placeholder="筛选级别" 
              clearable 
              size="small"
              style="width: 120px; margin-right: 10px"
            >
              <el-option label="全部级别" value="" />
              <el-option label="INFO" value="INFO" />
              <el-option label="WARN" value="WARN" />
              <el-option label="ERROR" value="ERROR" />
              <el-option label="DEBUG" value="DEBUG" />
            </el-select>
            <el-switch v-model="autoScroll" active-text="自动滚动" style="margin-right: 15px" />
            <el-button size="small" @click="clearLogs">清空</el-button>
            <el-button size="small" :type="isConnected ? 'success' : 'danger'" @click="toggleConnection">
              {{ isConnected ? '自动刷新中' : '已暂停' }}
            </el-button>
          </div>
        </div>
      </template>
      <div ref="logContainer" class="log-container">
        <div v-for="(log, index) in filteredLogs" :key="index" class="log-item" :class="'log-' + log.level?.toLowerCase()">
          <span class="log-time">{{ log.timestamp }}</span>
          <span class="log-level">[{{ log.level }}]</span>
          <span class="log-worker">[{{ log.workerName }}]</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
        <div v-if="filteredLogs.length === 0 && logs.length > 0" class="log-empty">没有匹配的日志</div>
        <div v-if="logs.length === 0" class="log-empty">暂无日志，请确保Worker启动时指定了Redis地址参数 -r</div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch, reactive, computed } from 'vue'
import { Refresh, Delete, Edit, RefreshRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(false)
const tableData = ref([])
const logs = ref([])
const logContainer = ref(null)
const autoScroll = ref(true)
const isConnected = ref(false)
const autoRefresh = ref(true)
const filterWorker = ref('')
const filterLevel = ref('')
let pollingTimer = null
let workerRefreshTimer = null
let logIdSet = new Set() // 用于去重

// 筛选后的日志
const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    // 筛选 Worker
    if (filterWorker.value && log.workerName !== filterWorker.value) {
      return false
    }
    // 筛选级别
    if (filterLevel.value && log.level?.toUpperCase() !== filterLevel.value) {
      return false
    }
    return true
  })
})

// 重命名相关
const renameDialogVisible = ref(false)
const renameLoading = ref(false)
const renameForm = reactive({
  oldName: '',
  newName: ''
})

onMounted(() => {
  loadData()
  startPolling()
  startWorkerRefresh()
})

onUnmounted(() => {
  stopPolling()
  stopWorkerRefresh()
})

watch(filteredLogs, () => {
  if (autoScroll.value) {
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  }
}, { deep: true })

async function loadData() {
  loading.value = true
  try {
    const res = await request.post('/worker/list')
    if (res.code === 0) tableData.value = res.list || []
  } finally {
    loading.value = false
  }
}

function startWorkerRefresh() {
  if (workerRefreshTimer) return
  // 每10秒自动刷新Worker列表（因为每次查询需要约1.5秒等待Worker响应）
  workerRefreshTimer = setInterval(() => {
    if (autoRefresh.value && !loading.value) {
      loadData()
    }
  }, 10000)
}

function stopWorkerRefresh() {
  if (workerRefreshTimer) {
    clearInterval(workerRefreshTimer)
    workerRefreshTimer = null
  }
}

function toggleAutoRefresh(val) {
  if (val) {
    startWorkerRefresh()
  } else {
    stopWorkerRefresh()
  }
}

function startPolling() {
  if (pollingTimer) return
  console.log('[Polling] Starting...')
  isConnected.value = true
  // 立即获取一次
  fetchLogsHistory()
  // 每2秒轮询一次
  pollingTimer = setInterval(fetchLogsHistory, 2000)
}

function stopPolling() {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
  isConnected.value = false
}

async function fetchLogsHistory() {
  try {
    const res = await request.post('/worker/logs/history', { limit: 200 })
    if (res.code === 0 && res.list && res.list.length > 0) {
      // 找出新日志（使用timestamp+message作为唯一标识）
      let hasNew = false
      for (const log of res.list) {
        const logId = (log.timestamp || '') + (log.message || '')
        if (!logIdSet.has(logId)) {
          logIdSet.add(logId)
          logs.value.push(log)
          hasNew = true
        }
      }
      // 限制日志数量和去重集合大小
      if (logs.value.length > 1000) {
        const removed = logs.value.splice(0, logs.value.length - 500)
        removed.forEach(l => logIdSet.delete((l.timestamp || '') + (l.message || '')))
      }
    }
  } catch (e) {
    console.error('[Polling] Fetch logs error:', e)
    isConnected.value = false
  }
}

function toggleConnection() {
  if (isConnected.value) {
    stopPolling()
  } else {
    startPolling()
  }
}

async function clearLogs() {
  try {
    // 清空服务端Redis中的历史日志
    const res = await request.post('/worker/logs/clear')
    if (res.code === 0) {
      // 清空本地日志
      logs.value = []
      logIdSet.clear()
    }
  } catch (e) {
    console.error('Clear logs error:', e)
  }
}

function getLoadColor(value) {
  if (value < 50) return '#67C23A'
  if (value < 80) return '#E6A23C'
  return '#F56C6C'
}

async function deleteWorker(workerName) {
  try {
    const res = await request.post('/worker/delete', { name: workerName })
    if (res.code === 0) {
      ElMessage.success('Worker已删除，停止信号已发送')
      loadData()
    } else {
      ElMessage.error(res.msg || '删除失败')
    }
  } catch (e) {
    ElMessage.error('删除失败: ' + e.message)
  }
}

async function restartWorker(workerName) {
  try {
    const res = await request.post('/worker/restart', { name: workerName })
    if (res.code === 0) {
      ElMessage.success('重启命令已发送')
      // 延迟刷新，等待Worker重启
      setTimeout(() => loadData(), 3000)
    } else {
      ElMessage.error(res.msg || '重启失败')
    }
  } catch (e) {
    ElMessage.error('重启失败: ' + e.message)
  }
}

function openRenameDialog(row) {
  renameForm.oldName = row.name
  renameForm.newName = row.name
  renameDialogVisible.value = true
}

async function submitRename() {
  if (!renameForm.newName.trim()) {
    ElMessage.warning('请输入新的Worker名称')
    return
  }
  if (renameForm.newName === renameForm.oldName) {
    renameDialogVisible.value = false
    return
  }

  renameLoading.value = true
  try {
    const res = await request.post('/worker/rename', {
      oldName: renameForm.oldName,
      newName: renameForm.newName.trim()
    })
    if (res.code === 0) {
      ElMessage.success('重命名成功')
      renameDialogVisible.value = false
      loadData()
    } else {
      ElMessage.error(res.msg || '重命名失败')
    }
  } catch (e) {
    ElMessage.error('重命名失败: ' + e.message)
  } finally {
    renameLoading.value = false
  }
}
</script>

<style lang="scss" scoped>
.worker-page {
  .action-card { margin-bottom: 20px; }

  .log-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
    
    .log-filters {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 5px;
    }
  }

  .log-container {
    height: 400px;
    overflow-y: auto;
    background: #1e1e1e;
    border-radius: 4px;
    padding: 10px;
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 12px;
  }

  .log-item {
    padding: 2px 0;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;

    .log-time {
      color: #6a9955;
      margin-right: 10px;
    }

    .log-level {
      display: inline-block;
      width: 60px;
      margin-right: 8px;
      font-weight: bold;
    }

    .log-worker {
      color: #569cd6;
      margin-right: 8px;
    }

    .log-message {
      color: #d4d4d4;
    }

    &.log-info .log-level { color: #4ec9b0; }
    &.log-warn .log-level { color: #dcdcaa; }
    &.log-error .log-level { color: #f14c4c; }
    &.log-debug .log-level { color: #9cdcfe; }
  }

  .log-empty {
    color: #6a6a6a;
    text-align: center;
    padding: 50px;
  }

  .editable-name {
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    
    &:hover {
      color: #409eff;
      
      .edit-icon {
        opacity: 1;
      }
    }
    
    .edit-icon {
      opacity: 0;
      font-size: 14px;
      transition: opacity 0.2s;
    }
  }
}
</style>
