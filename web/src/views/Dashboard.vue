<template>
  <div class="workspace">
    <!-- 顶部欢迎区 -->
    <div class="welcome-section">
      <div class="welcome-left">
        <h2>Hello, {{ userStore.username }}!</h2>
        <span class="role-tag">{{ userStore.isSuperAdmin ? '超级管理员' : '普通用户' }}</span>
      </div>
      <div class="welcome-right">
        <span class="time">{{ currentTime }}</span>
      </div>
    </div>

    <el-row :gutter="20">
      <!-- 左侧主区域 -->
      <el-col :span="16">
        <!-- 漏洞统计 -->
        <div class="dark-card">
          <div class="card-header">
            <span class="card-title">漏洞统计</span>
            <el-icon class="card-icon"><Warning /></el-icon>
          </div>
          <div class="vul-stats">
            <div class="vul-stat-item">
              <span class="vul-count total">{{ vulStats.total }}</span>
              <span class="vul-label">总数</span>
            </div>
            <div class="vul-stat-item">
              <span class="vul-count critical">{{ vulStats.critical }}</span>
              <span class="vul-label">严重</span>
            </div>
            <div class="vul-stat-item">
              <span class="vul-count high">{{ vulStats.high }}</span>
              <span class="vul-label">高危</span>
            </div>
            <div class="vul-stat-item">
              <span class="vul-count medium">{{ vulStats.medium }}</span>
              <span class="vul-label">中危</span>
            </div>
            <div class="vul-stat-item">
              <span class="vul-count low">{{ vulStats.low }}</span>
              <span class="vul-label">低危</span>
            </div>
            <div class="vul-stat-item">
              <span class="vul-count info">{{ vulStats.info }}</span>
              <span class="vul-label">信息</span>
            </div>
          </div>
          <div class="vul-chart-section">
            <div class="chart-title">漏洞分类统计</div>
            <div v-if="vulStats.total === 0" class="no-data">暂无漏洞数据</div>
            <div v-else ref="vulCategoryChartRef" class="chart-container"></div>
          </div>
          <div class="view-all-btn" @click="$router.push('/vul')">查看全部</div>
        </div>

        <!-- 资产统计 -->
        <div class="dark-card">
          <div class="card-header">
            <span class="card-title">资产统计</span>
            <el-icon class="card-icon"><Monitor /></el-icon>
          </div>
          <div class="asset-stats">
            <div class="asset-stat-item">
              <span class="asset-count">{{ assetStats.totalAsset }}</span>
              <span class="asset-label">资产</span>
            </div>
            <div class="asset-stat-item">
              <span class="asset-count">{{ assetStats.totalHost }}</span>
              <span class="asset-label">主机</span>
            </div>
            <div class="asset-stat-item">
              <span class="asset-count">{{ assetStats.totalService }}</span>
              <span class="asset-label">服务</span>
            </div>
            <div class="asset-stat-item">
              <span class="asset-count">{{ assetStats.totalApp }}</span>
              <span class="asset-label">应用</span>
            </div>
          </div>
          <div class="asset-category-section">
            <div class="chart-title">资产分类</div>
            <div class="category-list">
              <div v-for="item in assetCategories" :key="item.name" class="category-item">
                <span class="category-name">{{ item.name }}</span>
                <div class="category-bar">
                  <div class="category-bar-fill" :style="{ width: item.percent + '%' }"></div>
                </div>
                <span class="category-count">{{ item.count }}</span>
              </div>
            </div>
          </div>
          <div class="view-all-btn" @click="$router.push('/asset')">查看全部</div>
        </div>

        <!-- 任务效率统计 -->
        <div class="dark-card">
          <div class="card-header">
            <span class="card-title">任务统计 (近30天)</span>
            <el-icon class="card-icon"><List /></el-icon>
          </div>
          <div class="task-stats">
            <div class="task-stat-item">
              <span class="task-count">{{ taskStats.total }}</span>
              <span class="task-label">总任务</span>
            </div>
            <div class="task-stat-item">
              <span class="task-count success">{{ taskStats.completed }}</span>
              <span class="task-label">已完成</span>
            </div>
            <div class="task-stat-item">
              <span class="task-count warning">{{ taskStats.running }}</span>
              <span class="task-label">进行中</span>
            </div>
            <div class="task-stat-item">
              <span class="task-count error">{{ taskStats.failed }}</span>
              <span class="task-label">失败</span>
            </div>
          </div>
          <div class="task-chart-section">
            <div ref="taskTrendChartRef" class="chart-container"></div>
          </div>
        </div>
      </el-col>

      <!-- 右侧信息区 -->
      <el-col :span="8">
        <!-- 安全评分 -->
        <div class="dark-card score-card">
          <div class="card-header">
            <span class="card-title">安全评分</span>
            <el-icon class="card-icon"><Aim /></el-icon>
          </div>
          <div class="score-display">
            <div class="score-circle" :class="getScoreClass(securityScore)">
              <span class="score-value">{{ securityScore }}</span>
            </div>
            <div class="score-level">{{ getScoreLevel(securityScore) }}</div>
          </div>
          <div class="score-bar">
            <div class="score-bar-fill" :style="{ width: securityScore + '%', background: getScoreColor(securityScore) }"></div>
          </div>
        </div>

        <!-- 最新漏洞 -->
        <div class="dark-card">
          <div class="card-header">
            <span class="card-title">最新漏洞</span>
            <div class="header-stats">
              <span class="stat-item"><strong>{{ recentVulStats.week }}</strong> 近7天</span>
              <span class="stat-item"><strong>{{ recentVulStats.month }}</strong> 近30天</span>
            </div>
          </div>
          <div class="recent-vul-list">
            <div v-for="vul in recentVuls" :key="vul.id" class="vul-item">
              <el-tag :type="getSeverityType(vul.severity)" size="small" class="vul-severity">
                {{ vul.severity }}
              </el-tag>
              <span class="vul-name" :title="vul.name">{{ vul.name }}</span>
              <span class="vul-time">{{ vul.time }}</span>
            </div>
            <div v-if="recentVuls.length === 0" class="no-data">暂无漏洞</div>
          </div>
          <div class="view-all-btn" @click="$router.push('/vul')">查看全部</div>
        </div>

        <!-- Worker状态 -->
        <div class="dark-card">
          <div class="card-header">
            <span class="card-title">Worker状态</span>
            <el-icon class="card-icon"><Connection /></el-icon>
          </div>
          <div class="worker-stats">
            <div class="worker-stat-item">
              <div class="worker-indicator online"></div>
              <span class="worker-label">在线</span>
              <span class="worker-count">{{ workerStats.online }}</span>
            </div>
            <div class="worker-stat-item">
              <div class="worker-indicator offline"></div>
              <span class="worker-label">离线</span>
              <span class="worker-count">{{ workerStats.offline }}</span>
            </div>
          </div>
          <div class="view-all-btn" @click="$router.push('/worker')">管理Worker</div>
        </div>

        <!-- 快捷操作 -->
        <div class="dark-card">
          <div class="card-header">
            <span class="card-title">快捷操作</span>
          </div>
          <div class="quick-actions">
            <div class="action-btn" @click="$router.push('/task')">
              <el-icon><Plus /></el-icon>
              <span>新建任务</span>
            </div>
            <div class="action-btn" @click="$router.push('/asset')">
              <el-icon><Monitor /></el-icon>
              <span>资产管理</span>
            </div>
            <div class="action-btn" @click="$router.push('/poc')">
              <el-icon><Aim /></el-icon>
              <span>POC管理</span>
            </div>
            <div class="action-btn" @click="$router.push('/online-search')">
              <el-icon><Search /></el-icon>
              <span>在线搜索</span>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { useWorkspaceStore } from '@/stores/workspace'
import * as echarts from 'echarts'
import { getAssetStat } from '@/api/asset'
import { getTaskList } from '@/api/task'
import request from '@/api/request'

const userStore = useUserStore()
const workspaceStore = useWorkspaceStore()
const currentTime = ref('')

// 漏洞统计
const vulStats = reactive({
  total: 0,
  critical: 0,
  high: 0,
  medium: 0,
  low: 0,
  info: 0
})

// 资产统计
const assetStats = reactive({
  totalAsset: 0,
  totalHost: 0,
  totalService: 0,
  totalApp: 0
})

// 资产分类
const assetCategories = ref([])

// 任务统计
const taskStats = reactive({
  total: 0,
  completed: 0,
  running: 0,
  failed: 0
})

// 最新漏洞
const recentVuls = ref([])
const recentVulStats = reactive({
  week: 0,
  month: 0
})

// Worker状态
const workerStats = reactive({
  online: 0,
  offline: 0
})

// 安全评分
const securityScore = computed(() => {
  if (vulStats.total === 0) return 100
  const criticalWeight = vulStats.critical * 25
  const highWeight = vulStats.high * 15
  const mediumWeight = vulStats.medium * 5
  const lowWeight = vulStats.low * 2
  const totalWeight = criticalWeight + highWeight + mediumWeight + lowWeight
  const score = Math.max(0, 100 - totalWeight)
  return Math.round(score)
})

// 图表引用
const vulCategoryChartRef = ref()
const taskTrendChartRef = ref()
let vulCategoryChart, taskTrendChart

// 时间更新
let timeInterval
function updateTime() {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 监听工作空间切换
function handleWorkspaceChanged() {
  loadAllData()
}

async function loadAllData() {
  await Promise.all([
    loadAssetStats(),
    loadVulStats(),
    loadTaskStats(),
    loadWorkerStats()
  ])
  initCharts()
}

onMounted(async () => {
  updateTime()
  timeInterval = setInterval(updateTime, 1000)
  await loadAllData()
  window.addEventListener('workspace-changed', handleWorkspaceChanged)
})

onUnmounted(() => {
  clearInterval(timeInterval)
  vulCategoryChart?.dispose()
  taskTrendChart?.dispose()
  window.removeEventListener('workspace-changed', handleWorkspaceChanged)
})

async function loadAssetStats() {
  try {
    const res = await getAssetStat()
    if (res.code === 0) {
      assetStats.totalAsset = res.totalAsset || 0
      assetStats.totalHost = res.totalHost || 0
      assetStats.totalService = res.topService?.length || 0
      assetStats.totalApp = res.topApp?.length || 0
      
      // 处理资产分类
      const categories = []
      const maxCount = Math.max(...(res.topService || []).map(i => i.count), 1)
      ;(res.topService || []).slice(0, 6).forEach(item => {
        categories.push({
          name: item.name,
          count: item.count,
          percent: Math.round((item.count / maxCount) * 100)
        })
      })
      assetCategories.value = categories
    }
  } catch (e) {
    console.error('Failed to load asset stats:', e)
  }
}

async function loadVulStats() {
  try {
    const res = await request.post('/vul/list', { page: 1, pageSize: 10 })
    if (res.code === 0) {
      vulStats.total = res.total || 0
      // 统计各级别漏洞
      const list = res.list || []
      vulStats.critical = list.filter(v => v.severity === 'critical').length
      vulStats.high = list.filter(v => v.severity === 'high').length
      vulStats.medium = list.filter(v => v.severity === 'medium').length
      vulStats.low = list.filter(v => v.severity === 'low').length
      vulStats.info = list.filter(v => v.severity === 'info').length
      
      // 最新漏洞
      recentVuls.value = list.slice(0, 5).map(v => ({
        id: v.id,
        name: v.vulName || v.pocFile || 'Unknown',
        severity: v.severity || 'info',
        time: formatTime(v.createTime)
      }))
      
      recentVulStats.week = Math.min(res.total, 10)
      recentVulStats.month = res.total
    }
  } catch (e) {
    console.error('Failed to load vul stats:', e)
  }
}

async function loadTaskStats() {
  try {
    const res = await getTaskList({ page: 1, pageSize: 100 })
    if (res.code === 0) {
      const list = res.list || []
      taskStats.total = list.length
      taskStats.completed = list.filter(t => t.status === 'completed').length
      taskStats.running = list.filter(t => t.status === 'running' || t.status === 'pending').length
      taskStats.failed = list.filter(t => t.status === 'failed').length
    }
  } catch (e) {
    console.error('Failed to load task stats:', e)
  }
}

async function loadWorkerStats() {
  try {
    const res = await request.post('/worker/list')
    if (res.code === 0) {
      const list = res.list || []
      workerStats.online = list.filter(w => w.status === 'running').length
      workerStats.offline = list.filter(w => w.status === 'offline').length
    }
  } catch (e) {
    console.error('Failed to load worker stats:', e)
  }
}

function initCharts() {
  // 漏洞分类图表
  if (vulCategoryChartRef.value && vulStats.total > 0) {
    vulCategoryChart = echarts.init(vulCategoryChartRef.value)
    vulCategoryChart.setOption({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'item', backgroundColor: '#1d1e1f', borderColor: '#333', textStyle: { color: '#fff' } },
      series: [{
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '50%'],
        data: [
          { value: vulStats.critical, name: '严重', itemStyle: { color: '#f56c6c' } },
          { value: vulStats.high, name: '高危', itemStyle: { color: '#e6a23c' } },
          { value: vulStats.medium, name: '中危', itemStyle: { color: '#409eff' } },
          { value: vulStats.low, name: '低危', itemStyle: { color: '#67c23a' } },
          { value: vulStats.info, name: '信息', itemStyle: { color: '#909399' } }
        ].filter(d => d.value > 0),
        label: { color: '#a3a6ad', fontSize: 12 },
        labelLine: { lineStyle: { color: '#444' } }
      }]
    })
  }

  // 任务趋势图表
  if (taskTrendChartRef.value) {
    taskTrendChart = echarts.init(taskTrendChartRef.value)
    const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
    taskTrendChart.setOption({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'axis', backgroundColor: '#1d1e1f', borderColor: '#333', textStyle: { color: '#fff' } },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: days, axisLine: { lineStyle: { color: '#444' } }, axisLabel: { color: '#a3a6ad' } },
      yAxis: { type: 'value', axisLine: { lineStyle: { color: '#444' } }, axisLabel: { color: '#a3a6ad' }, splitLine: { lineStyle: { color: '#333' } } },
      series: [
        { name: '完成', type: 'line', smooth: true, data: [3, 5, 2, 8, 4, 6, 3], itemStyle: { color: '#67c23a' }, areaStyle: { color: 'rgba(103, 194, 58, 0.1)' } },
        { name: '失败', type: 'line', smooth: true, data: [1, 0, 1, 2, 0, 1, 0], itemStyle: { color: '#f56c6c' }, areaStyle: { color: 'rgba(245, 108, 108, 0.1)' } }
      ]
    })
  }
}

function formatTime(timeStr) {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now - date
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前'
  return Math.floor(diff / 86400000) + '天前'
}

function getSeverityType(severity) {
  const map = { critical: 'danger', high: 'warning', medium: '', low: 'info', info: 'success' }
  return map[severity] || 'info'
}

function getScoreClass(score) {
  if (score >= 80) return 'excellent'
  if (score >= 60) return 'good'
  if (score >= 40) return 'warning'
  return 'danger'
}

function getScoreLevel(score) {
  if (score >= 80) return '安全'
  if (score >= 60) return '良好'
  if (score >= 40) return '警告'
  return '危险'
}

function getScoreColor(score) {
  if (score >= 80) return 'linear-gradient(90deg, #67c23a, #85ce61)'
  if (score >= 60) return 'linear-gradient(90deg, #409eff, #66b1ff)'
  if (score >= 40) return 'linear-gradient(90deg, #e6a23c, #ebb563)'
  return 'linear-gradient(90deg, #f56c6c, #f78989)'
}
</script>

<style lang="scss" scoped>
.workspace {
  min-height: 100%;
}

.welcome-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  margin-bottom: 20px;
  
  .welcome-left {
    h2 {
      color: var(--text-primary);
      font-size: 24px;
      font-weight: 600;
      margin: 0 0 8px 0;
    }
    
    .role-tag {
      display: inline-block;
      padding: 4px 12px;
      background: rgba(64, 158, 255, 0.2);
      color: var(--primary-color);
      border-radius: 20px;
      font-size: 12px;
    }
  }
  
  .welcome-right {
    .time {
      color: var(--text-secondary);
      font-size: 14px;
    }
  }
}

.dark-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
    
    .card-title {
      color: var(--text-primary);
      font-size: 16px;
      font-weight: 600;
    }
    
    .card-icon {
      color: var(--text-secondary);
      font-size: 18px;
    }
    
    .header-stats {
      display: flex;
      gap: 16px;
      
      .stat-item {
        color: var(--text-secondary);
        font-size: 13px;
        
        strong {
          color: var(--text-primary);
          margin-right: 4px;
        }
      }
    }
  }
}

// 漏洞统计
.vul-stats {
  display: flex;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid var(--border-color);
  
  .vul-stat-item {
    text-align: center;
    
    .vul-count {
      display: block;
      font-size: 28px;
      font-weight: 700;
      margin-bottom: 4px;
      
      &.total { color: var(--text-primary); }
      &.critical { color: var(--danger-color); }
      &.high { color: var(--warning-color); }
      &.medium { color: var(--primary-color); }
      &.low { color: var(--success-color); }
      &.info { color: var(--info-color); }
    }
    
    .vul-label {
      color: var(--text-secondary);
      font-size: 13px;
    }
  }
}

.vul-chart-section, .task-chart-section {
  .chart-title {
    color: var(--text-secondary);
    font-size: 13px;
    margin: 16px 0 12px;
  }
  
  .chart-container {
    height: 200px;
  }
}

// 资产统计
.asset-stats {
  display: flex;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid var(--border-color);
  
  .asset-stat-item {
    text-align: center;
    
    .asset-count {
      display: block;
      font-size: 28px;
      font-weight: 700;
      color: var(--text-primary);
      margin-bottom: 4px;
    }
    
    .asset-label {
      color: var(--text-secondary);
      font-size: 13px;
    }
  }
}

.asset-category-section {
  .chart-title {
    color: var(--text-secondary);
    font-size: 13px;
    margin: 16px 0 12px;
  }
  
  .category-list {
    .category-item {
      display: flex;
      align-items: center;
      margin-bottom: 12px;
      
      .category-name {
        width: 100px;
        color: var(--text-secondary);
        font-size: 13px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      
      .category-bar {
        flex: 1;
        height: 8px;
        background: var(--border-color);
        border-radius: 4px;
        margin: 0 12px;
        overflow: hidden;
        
        .category-bar-fill {
          height: 100%;
          background: linear-gradient(90deg, var(--primary-color), #66b1ff);
          border-radius: 4px;
          transition: width 0.3s;
        }
      }
      
      .category-count {
        width: 40px;
        text-align: right;
        color: var(--text-primary);
        font-size: 13px;
      }
    }
  }
}

// 任务统计
.task-stats {
  display: flex;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid var(--border-color);
  
  .task-stat-item {
    text-align: center;
    
    .task-count {
      display: block;
      font-size: 24px;
      font-weight: 700;
      color: var(--text-primary);
      margin-bottom: 4px;
      
      &.success { color: var(--success-color); }
      &.warning { color: var(--warning-color); }
      &.error { color: var(--danger-color); }
    }
    
    .task-label {
      color: var(--text-secondary);
      font-size: 13px;
    }
  }
}

// 安全评分
.score-card {
  .score-display {
    text-align: center;
    padding: 20px 0;
    
    .score-circle {
      width: 100px;
      height: 100px;
      border-radius: 50%;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      margin-bottom: 12px;
      
      &.excellent {
        background: rgba(103, 194, 58, 0.15);
        border: 3px solid var(--success-color);
      }
      &.good {
        background: rgba(64, 158, 255, 0.15);
        border: 3px solid var(--primary-color);
      }
      &.warning {
        background: rgba(230, 162, 60, 0.15);
        border: 3px solid var(--warning-color);
      }
      &.danger {
        background: rgba(245, 108, 108, 0.15);
        border: 3px solid var(--danger-color);
      }
      
      .score-value {
        font-size: 32px;
        font-weight: 700;
        color: var(--text-primary);
      }
    }
    
    .score-level {
      color: var(--text-secondary);
      font-size: 14px;
    }
  }
  
  .score-bar {
    height: 6px;
    background: var(--border-color);
    border-radius: 3px;
    overflow: hidden;
    
    .score-bar-fill {
      height: 100%;
      border-radius: 3px;
      transition: width 0.5s;
    }
  }
}

// 最新漏洞
.recent-vul-list {
  .vul-item {
    display: flex;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid var(--border-light);
    
    &:last-child {
      border-bottom: none;
    }
    
    .vul-severity {
      flex-shrink: 0;
      margin-right: 12px;
    }
    
    .vul-name {
      flex: 1;
      color: var(--text-primary);
      font-size: 13px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    
    .vul-time {
      color: var(--text-muted);
      font-size: 12px;
      margin-left: 12px;
    }
  }
}

// Worker状态
.worker-stats {
  display: flex;
  justify-content: space-around;
  padding: 16px 0;
  
  .worker-stat-item {
    display: flex;
    align-items: center;
    gap: 8px;
    
    .worker-indicator {
      width: 10px;
      height: 10px;
      border-radius: 50%;
      
      &.online { background: var(--success-color); box-shadow: 0 0 8px rgba(103, 194, 58, 0.5); }
      &.offline { background: var(--info-color); }
    }
    
    .worker-label {
      color: var(--text-secondary);
      font-size: 13px;
    }
    
    .worker-count {
      color: var(--text-primary);
      font-size: 16px;
      font-weight: 600;
    }
  }
}

// 快捷操作
.quick-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  
  .action-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 16px;
    background: var(--bg-tertiary);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.3s;
    
    &:hover {
      background: var(--bg-hover);
      transform: translateY(-2px);
    }
    
    .el-icon {
      font-size: 24px;
      color: var(--primary-color);
      margin-bottom: 8px;
    }
    
    span {
      color: var(--text-secondary);
      font-size: 13px;
    }
  }
}

// 通用
.view-all-btn {
  text-align: center;
  padding: 12px 0 0;
  color: var(--primary-color);
  font-size: 13px;
  cursor: pointer;
  
  &:hover {
    color: #66b1ff;
  }
}

.no-data {
  text-align: center;
  color: var(--text-muted);
  padding: 40px 0;
  font-size: 14px;
}
</style>
