<template>
  <div class="worker-page">
    <el-card class="action-card">
      <el-button type="primary" @click="loadData">
        <el-icon><Refresh /></el-icon>刷新
      </el-button>
    </el-card>

    <el-card style="margin-bottom: 20px">
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="name" label="Worker名称" min-width="200" />
        <el-table-column prop="cpuLoad" label="CPU负载" width="120">
          <template #default="{ row }">
            <el-progress :percentage="row.cpuLoad" :stroke-width="10" :color="getLoadColor(row.cpuLoad)" />
          </template>
        </el-table-column>
        <el-table-column prop="memUsed" label="内存使用" width="120">
          <template #default="{ row }">
            <el-progress :percentage="row.memUsed" :stroke-width="10" :color="getLoadColor(row.memUsed)" />
          </template>
        </el-table-column>
        <el-table-column prop="taskCount" label="执行任务数" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'running' ? 'success' : 'danger'">
              {{ row.status === 'running' ? '运行中' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updateTime" label="最后心跳" width="160" />
      </el-table>

      <el-empty v-if="!loading && tableData.length === 0" description="暂无Worker节点" />
    </el-card>

    <!-- 实时日志 -->
    <el-card>
      <template #header>
        <div class="log-header">
          <span>Worker运行日志</span>
          <div>
            <el-switch v-model="autoScroll" active-text="自动滚动" style="margin-right: 15px" />
            <el-button size="small" @click="clearLogs">清空</el-button>
            <el-button size="small" :type="isConnected ? 'success' : 'danger'" @click="toggleConnection">
              {{ isConnected ? '自动刷新中' : '已暂停' }}
            </el-button>
          </div>
        </div>
      </template>
      <div ref="logContainer" class="log-container">
        <div v-for="(log, index) in logs" :key="index" class="log-item" :class="'log-' + log.level?.toLowerCase()">
          <span class="log-time">{{ log.timestamp }}</span>
          <span class="log-level">[{{ log.level }}]</span>
          <span class="log-worker">[{{ log.workerName }}]</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
        <div v-if="logs.length === 0" class="log-empty">暂无日志，请确保Worker启动时指定了Redis地址参数 -r</div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const tableData = ref([])
const logs = ref([])
const logContainer = ref(null)
const autoScroll = ref(true)
const isConnected = ref(false)
let pollingTimer = null
let logIdSet = new Set() // 用于去重

onMounted(() => {
  loadData()
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})

watch(logs, () => {
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
</script>

<style lang="scss" scoped>
.worker-page {
  .action-card { margin-bottom: 20px; }

  .log-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
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
}
</style>
