<template>
  <div class="report-page">
    <!-- 报告头部 -->
    <el-card class="report-header" v-if="reportData">
      <div class="header-content">
        <div class="title-section">
          <h2>{{ reportData.taskName }}</h2>
          <el-tag :type="getStatusType(reportData.status)" size="large">{{ reportData.status }}</el-tag>
        </div>
        <div class="action-section">
          <el-button type="primary" @click="handleExport" :loading="exporting">
            <el-icon><Download /></el-icon>导出Excel
          </el-button>
          <el-button @click="goBack">
            <el-icon><Back /></el-icon>返回
          </el-button>
        </div>
      </div>
      <el-descriptions :column="4" border class="task-info">
        <el-descriptions-item label="扫描目标">
          <div class="target-text">{{ reportData.target }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ reportData.createTime }}</el-descriptions-item>
        <el-descriptions-item label="资产数量">
          <span class="stat-number">{{ reportData.assetCount }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="漏洞数量">
          <span class="stat-number danger">{{ reportData.vulCount }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 统计概览 -->
    <el-row :gutter="20" class="stats-row" v-if="reportData">
      <!-- 漏洞统计 -->
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>漏洞等级分布</template>
          <div class="vul-stats">
            <div class="vul-item critical">
              <span class="label">Critical</span>
              <span class="count">{{ reportData.vulStats?.critical || 0 }}</span>
            </div>
            <div class="vul-item high">
              <span class="label">High</span>
              <span class="count">{{ reportData.vulStats?.high || 0 }}</span>
            </div>
            <div class="vul-item medium">
              <span class="label">Medium</span>
              <span class="count">{{ reportData.vulStats?.medium || 0 }}</span>
            </div>
            <div class="vul-item low">
              <span class="label">Low</span>
              <span class="count">{{ reportData.vulStats?.low || 0 }}</span>
            </div>
            <div class="vul-item info">
              <span class="label">Info</span>
              <span class="count">{{ reportData.vulStats?.info || 0 }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      <!-- 端口统计 -->
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>Top 端口</template>
          <div class="top-list">
            <div v-for="item in topPorts" :key="item.name" class="top-item">
              <span class="name">{{ item.name }}</span>
              <span class="count">{{ item.count }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      <!-- 服务统计 -->
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>Top 服务</template>
          <div class="top-list">
            <div v-for="item in topServices" :key="item.name" class="top-item">
              <span class="name">{{ item.name || '-' }}</span>
              <span class="count">{{ item.count }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      <!-- 应用统计 -->
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>Top 应用</template>
          <div class="top-list">
            <div v-for="item in topApps" :key="item.name" class="top-item">
              <span class="name">{{ item.name }}</span>
              <span class="count">{{ item.count }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 资产列表 -->
    <el-card class="data-card" v-if="reportData">
      <template #header>
        <div class="card-header">
          <span>资产列表 ({{ reportData.assets?.length || 0 }})</span>
          <el-input v-model="assetSearch" placeholder="搜索资产..." style="width: 250px" clearable />
        </div>
      </template>
      <el-table :data="filteredAssets" stripe max-height="400">
        <el-table-column prop="authority" label="地址" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="asset-cell">
              <span class="authority">{{ row.authority }}</span>
              <el-tag v-if="row.httpStatus" size="small" :type="getHttpStatusType(row.httpStatus)">
                {{ row.httpStatus }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="service" label="服务" width="100" />
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column label="应用" min-width="200">
          <template #default="{ row }">
            <div class="app-tags">
              <el-tooltip 
                v-for="app in (row.app || [])" 
                :key="app" 
                :content="getAppSource(app)"
                placement="top"
              >
                <el-tag 
                  size="small" 
                  :type="getAppTagType(app)"
                  class="app-tag"
                >
                  {{ getAppName(app) }}
                </el-tag>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="server" label="Server" width="120" show-overflow-tooltip />
        <el-table-column label="截图" width="100">
          <template #default="{ row }">
            <el-image 
              v-if="row.screenshot" 
              :src="getScreenshotUrl(row.screenshot)" 
              :preview-src-list="[getScreenshotUrl(row.screenshot)]"
              :z-index="9999"
              :preview-teleported="true"
              fit="cover" 
              style="width: 60px; height: 40px; cursor: pointer; border-radius: 4px"
            >
              <template #error>
                <div class="image-error">
                  <el-icon><Picture /></el-icon>
                </div>
              </template>
            </el-image>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 漏洞列表 -->
    <el-card class="data-card" v-if="reportData && reportData.vuls?.length > 0">
      <template #header>
        <div class="card-header">
          <span>漏洞列表 ({{ reportData.vuls?.length || 0 }})</span>
          <el-input v-model="vulSearch" placeholder="搜索漏洞..." style="width: 250px" clearable />
        </div>
      </template>
      <el-table :data="filteredVuls" stripe max-height="400">
        <el-table-column prop="severity" label="等级" width="100">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="authority" label="目标" width="180" show-overflow-tooltip />
        <el-table-column prop="url" label="URL" min-width="250" show-overflow-tooltip />
        <el-table-column prop="pocFile" label="POC" min-width="200" show-overflow-tooltip />
        <el-table-column prop="result" label="结果" min-width="200" show-overflow-tooltip />
        <el-table-column prop="createTime" label="发现时间" width="160" />
      </el-table>
    </el-card>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading" :size="40"><Loading /></el-icon>
      <p>加载报告中...</p>
    </div>
    
    <!-- 无数据状态 -->
    <div v-if="!loading && reportData && reportData.assetCount === 0 && reportData.vulCount === 0" class="empty-container">
      <el-empty description="暂无扫描结果">
        <template #description>
          <p>该任务暂无扫描结果</p>
          <p style="color: #909399; font-size: 12px;">可能原因：任务未完成、目标无开放端口、或扫描配置问题</p>
        </template>
      </el-empty>
    </div>
    
    <!-- 任务不存在 -->
    <div v-if="!loading && !reportData" class="empty-container">
      <el-empty description="任务不存在或加载失败" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Download, Back, Loading, Picture } from '@element-plus/icons-vue'
import { getReportDetail, exportReport } from '@/api/report'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const exporting = ref(false)
const reportData = ref(null)
const assetSearch = ref('')
const vulSearch = ref('')

const topPorts = computed(() => {
  return (reportData.value?.topPorts || []).slice(0, 5)
})

const topServices = computed(() => {
  return (reportData.value?.topServices || []).slice(0, 5)
})

const topApps = computed(() => {
  return (reportData.value?.topApps || []).slice(0, 5)
})

const filteredAssets = computed(() => {
  const assets = reportData.value?.assets || []
  if (!assetSearch.value) return assets
  const keyword = assetSearch.value.toLowerCase()
  return assets.filter(a => 
    a.authority?.toLowerCase().includes(keyword) ||
    a.title?.toLowerCase().includes(keyword) ||
    a.service?.toLowerCase().includes(keyword) ||
    (a.app || []).some(app => app.toLowerCase().includes(keyword))
  )
})

const filteredVuls = computed(() => {
  const vuls = reportData.value?.vuls || []
  if (!vulSearch.value) return vuls
  const keyword = vulSearch.value.toLowerCase()
  return vuls.filter(v => 
    v.authority?.toLowerCase().includes(keyword) ||
    v.url?.toLowerCase().includes(keyword) ||
    v.pocFile?.toLowerCase().includes(keyword) ||
    v.severity?.toLowerCase().includes(keyword)
  )
})

onMounted(() => {
  const taskId = route.query.taskId
  if (taskId) {
    loadReport(taskId)
  } else {
    ElMessage.error('缺少任务ID')
    router.push('/task')
  }
})

async function loadReport(taskId) {
  loading.value = true
  try {
    console.log('Loading report for taskId:', taskId)
    const res = await getReportDetail({ taskId })
    console.log('Report response:', res)
    if (res.code === 0) {
      reportData.value = res.data
      console.log('Report data:', res.data)
    } else {
      ElMessage.error(res.msg || '加载报告失败')
    }
  } catch (e) {
    console.error('Load report error:', e)
    ElMessage.error('加载报告失败')
  } finally {
    loading.value = false
  }
}

async function handleExport() {
  if (!reportData.value) return
  exporting.value = true
  try {
    const res = await exportReport({ taskId: reportData.value.taskId })
    // 创建下载链接
    const blob = new Blob([res], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `report_${reportData.value.taskName}_${new Date().toISOString().slice(0,10)}.xlsx`
    link.click()
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (e) {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

function goBack() {
  router.push('/task')
}

function getStatusType(status) {
  const map = { SUCCESS: 'success', FAILURE: 'danger', STARTED: 'primary', PENDING: 'warning', CREATED: 'info', STOPPED: 'info' }
  return map[status] || 'info'
}

function getHttpStatusType(status) {
  if (status?.startsWith('2')) return 'success'
  if (status?.startsWith('3')) return 'warning'
  if (status?.startsWith('4') || status?.startsWith('5')) return 'danger'
  return 'info'
}

function getSeverityType(severity) {
  const map = { critical: 'danger', high: 'danger', medium: 'warning', low: 'info', info: '' }
  return map[severity?.toLowerCase()] || 'info'
}

// 获取应用名称（去除来源标识）
function getAppName(app) {
  if (!app) return ''
  const idx = app.indexOf('[')
  return idx > 0 ? app.substring(0, idx) : app
}

// 获取应用来源（用于tooltip显示）
function getAppSource(app) {
  if (!app) return ''
  const match = app.match(/\[([^\]]+)\]$/)
  if (match) {
    const source = match[1]
    const sourceMap = {
      'httpx': 'httpx识别',
      'wappalyzer': 'Wappalyzer识别',
      'custom': '自定义指纹',
      'builtin': '内置指纹'
    }
    
    if (source.includes('+')) {
      const parts = source.split('+')
      const mappedParts = parts.map(part => {
        if (part.startsWith('custom(')) {
          const ids = part.match(/custom\(([^)]+)\)/)
          if (ids) {
            const idList = ids[1].split(',').map(id => id.trim())
            return `自定义指纹(${idList.length}个ID: ${idList.join(', ')})`
          }
          return '自定义指纹'
        }
        return sourceMap[part] || part
      })
      return mappedParts.join(' + ')
    }
    
    if (source.startsWith('custom(')) {
      const ids = source.match(/custom\(([^)]+)\)/)
      if (ids) {
        const idList = ids[1].split(',').map(id => id.trim())
        return `自定义指纹 (${idList.length}个ID: ${idList.join(', ')})`
      }
      return '自定义指纹'
    }
    
    if (source.startsWith('custom-')) {
      const id = source.substring(7)
      return `自定义指纹 (ID: ${id})`
    }
    
    return sourceMap[source] || source
  }
  return '未知来源'
}

// 根据来源返回标签类型
function getAppTagType(app) {
  if (!app) return 'info'
  if (app.includes('[httpx+wappalyzer+custom(')) return 'danger'
  if (app.includes('[httpx+wappalyzer]')) return 'primary'
  if (app.includes('[httpx+custom(')) return 'danger'
  if (app.includes('[wappalyzer+custom(')) return 'danger'
  if (app.includes('[httpx]')) return 'success'
  if (app.includes('[wappalyzer]')) return 'success'
  if (app.includes('[builtin]')) return 'warning'
  if (app.includes('custom(') || app.includes('[custom-')) return 'danger'
  return 'info'
}

// 获取截图URL
function getScreenshotUrl(screenshot) {
  if (!screenshot) return ''
  // 如果是base64格式
  if (screenshot.startsWith('data:') || screenshot.startsWith('/9j/') || screenshot.startsWith('iVBOR')) {
    if (!screenshot.startsWith('data:')) {
      return `data:image/png;base64,${screenshot}`
    }
    return screenshot
  }
  // 如果是文件路径
  return `/api/screenshot/${screenshot}`
}
</script>

<style lang="scss" scoped>
.report-page {
  padding: 20px;

  .report-header {
    margin-bottom: 20px;

    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;

      .title-section {
        display: flex;
        align-items: center;
        gap: 15px;

        h2 {
          margin: 0;
          font-size: 24px;
        }
      }

      .action-section {
        display: flex;
        gap: 10px;
      }
    }

    .target-text {
      max-width: 300px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .stat-number {
      font-size: 18px;
      font-weight: bold;
      color: var(--el-color-primary);

      &.danger {
        color: var(--el-color-danger);
      }
    }
  }

  .stats-row {
    margin-bottom: 20px;

    .stat-card {
      height: 250px;

      .vul-stats {
        .vul-item {
          display: flex;
          justify-content: space-between;
          padding: 8px 12px;
          margin-bottom: 5px;
          border-radius: 4px;

          &.critical { background: rgba(245, 108, 108, 0.1); .count { color: #f56c6c; } }
          &.high { background: rgba(230, 162, 60, 0.1); .count { color: #e6a23c; } }
          &.medium { background: rgba(64, 158, 255, 0.1); .count { color: #409eff; } }
          &.low { background: rgba(103, 194, 58, 0.1); .count { color: #67c23a; } }
          &.info { background: rgba(144, 147, 153, 0.1); .count { color: #909399; } }

          .count {
            font-weight: bold;
          }
        }
      }

      .top-list {
        .top-item {
          display: flex;
          justify-content: space-between;
          padding: 6px 0;
          border-bottom: 1px solid var(--el-border-color-lighter);

          &:last-child {
            border-bottom: none;
          }

          .name {
            max-width: 150px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          .count {
            color: var(--el-color-primary);
            font-weight: bold;
          }
        }
      }
    }
  }

  .data-card {
    margin-bottom: 20px;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .asset-cell {
      display: flex;
      align-items: center;
      gap: 8px;

      .authority {
        font-family: monospace;
      }
    }

    .app-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 300px;
    color: var(--el-text-color-secondary);
  }
  
  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 0;
  }

  .app-tag {
    margin: 0;
    flex-shrink: 0;
  }

  .image-error {
    width: 60px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--el-fill-color-light);
    border-radius: 4px;
    color: var(--el-text-color-secondary);
  }
}
</style>
