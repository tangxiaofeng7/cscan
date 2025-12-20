<template>
  <div class="asset-page">
    <!-- 搜索和统计区域 -->
    <el-card class="search-card">
      <!-- Tab切换 -->
      <el-tabs v-model="activeTab" class="search-tabs">
        <el-tab-pane label="语法查询" name="syntax">
          <el-input v-model="searchForm.query" placeholder="输入搜索语法，如: port=80 && service=http" style="width: 100%" />
          <div class="syntax-hints">
            <span class="hint-title">语法示例：</span>
            <span class="hint-item" @click="searchForm.query = 'port=80'">port=80</span>
            <span class="hint-item" @click="searchForm.query = 'port=80 && service=http'">port=80 && service=http</span>
            <span class="hint-item" @click="searchForm.query = 'title=&quot;后台管理&quot;'">title="后台管理"</span>
            <span class="hint-item" @click="searchForm.query = 'app=nginx && port=443'">app=nginx && port=443</span>
          </div>
        </el-tab-pane>
        <el-tab-pane label="快捷查询" name="quick">
          <el-form :model="searchForm" inline>
            <el-form-item label="主机">
              <el-input v-model="searchForm.host" placeholder="IP/域名" clearable style="width: 120px" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input v-model.number="searchForm.port" placeholder="端口号" clearable style="width: 90px" />
            </el-form-item>
            <el-form-item label="服务">
              <el-input v-model="searchForm.service" placeholder="服务" clearable style="width: 90px" />
            </el-form-item>
            <el-form-item label="标题">
              <el-input v-model="searchForm.title" placeholder="标题" clearable style="width: 120px" />
            </el-form-item>
            <el-form-item label="应用">
              <el-input v-model="searchForm.app" placeholder="指纹" clearable style="width: 100px" />
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="统计信息" name="stat">
          <div class="stat-panel">
            <div class="stat-column">
              <div class="stat-title">Port</div>
              <div v-for="item in stat.topPorts" :key="'port-'+item.name" class="stat-item" @click="quickFilter('port', item.name)">
                <span class="stat-count">{{ item.count }}</span>
                <span class="stat-name">{{ item.name }}</span>
              </div>
            </div>
            <div class="stat-column">
              <div class="stat-title">Service</div>
              <div v-for="item in stat.topService" :key="'svc-'+item.name" class="stat-item" @click="quickFilter('service', item.name)">
                <span class="stat-count">{{ item.count }}</span>
                <span class="stat-name">{{ item.name }}</span>
              </div>
            </div>
            <div class="stat-column">
              <div class="stat-title">App</div>
              <div v-for="item in stat.topApp" :key="'app-'+item.name" class="stat-item" @click="quickFilter('app', item.name)">
                <span class="stat-count">{{ item.count }}</span>
                <span class="stat-name">{{ item.name }}</span>
              </div>
            </div>
            <div class="stat-column">
              <div class="stat-title">Title</div>
              <div v-for="item in stat.topTitle" :key="'title-'+item.name" class="stat-item" @click="quickFilter('title', item.name)">
                <span class="stat-count">{{ item.count }}</span>
                <span class="stat-name">{{ item.name }}</span>
              </div>
            </div>
            <div class="stat-column filter-column">
              <div class="filter-options">
                <el-checkbox v-model="searchForm.excludeCdn">不看CDN/Cloud资产</el-checkbox>
                <el-checkbox v-model="searchForm.onlyNew">只看新资产</el-checkbox>
                <el-checkbox v-model="searchForm.onlyUpdated">只看有更新</el-checkbox>
                <el-checkbox v-model="searchForm.sortByUpdate">按更新时间排序</el-checkbox>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <div class="search-actions">
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <div class="table-header">
        <span class="total-info">共 {{ pagination.total }} 条记录，当前显示 {{ tableData.length }} 条</span>
        <el-button type="danger" size="small" :disabled="selectedRows.length === 0" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon>批量删除 ({{ selectedRows.length }})
        </el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe size="small" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" />
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column label="资产-组织" min-width="160">
          <template #default="{ row }">
            <div class="asset-cell">
              <a :href="getAssetUrl(row)" target="_blank" class="asset-link">{{ row.authority }}</a>
              <el-icon v-if="row.authority" class="link-icon"><Link /></el-icon>
            </div>
            <div class="org-text">{{ row.org || '暂无' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="IP" width="140">
          <template #default="{ row }">
            <div>{{ row.host }}</div>
            <div v-if="row.location && row.location !== 'IANA'" class="location-text">{{ row.location }}</div>
          </template>
        </el-table-column>
        <el-table-column label="端口-协议" width="120">
          <template #default="{ row }">
            <span>{{ row.port }}</span>
            <span v-if="row.service" class="service-text"> {{ row.service }}</span>
          </template>
        </el-table-column>
        <el-table-column label="标题-指纹" min-width="200">
          <template #default="{ row }">
            <div class="title-text" :title="row.title">{{ row.title || '-' }}</div>
            <div class="app-tags-container">
              <el-tooltip 
                v-for="app in (row.app || [])" 
                :key="app" 
                :content="getAppTooltip(app)"
                placement="top"
              >
                <el-tag 
                  size="small" 
                  :type="getAppTagType(app)" 
                  class="app-tag"
                  :class="{ 'clickable-tag': isCustomFingerprint(app) }"
                  @click="handleAppTagClick(app)"
                >
                  {{ getAppName(app) }}
                </el-tag>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="指纹信息" min-width="320">
          <template #default="{ row }">
            <div class="fingerprint-info">
              <el-tabs v-if="row.httpHeader || row.httpStatus || row.httpBody || row.banner" type="border-card" class="fingerprint-tabs">
                <el-tab-pane label="Header">
                  <pre v-if="row.httpHeader || row.httpStatus" class="tab-content">{{ formatHeaderWithStatus(row) }}</pre>
                  <pre v-else-if="row.banner" class="tab-content">{{ row.banner }}</pre>
                  <span v-else class="no-data">-</span>
                </el-tab-pane>
                <el-tab-pane label="Body">
                  <pre v-if="row.httpBody" class="tab-content">{{ row.httpBody }}</pre>
                  <span v-else class="no-data">无内容</span>
                </el-tab-pane>
              </el-tabs>
              <span v-else class="no-data">-</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Web截屏" width="100" align="center">
          <template #default="{ row }">
            <el-image 
              v-if="row.screenshot" 
              :src="getScreenshotUrl(row.screenshot)" 
              :preview-src-list="[getScreenshotUrl(row.screenshot)]"
              :z-index="9999"
              :preview-teleported="true"
              :hide-on-click-modal="true"
              fit="cover"
              class="screenshot-img"
            />
            <span v-else class="no-screenshot">-</span>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="100">
          <template #default="{ row }">
            <div class="update-time">{{ formatTime(row.updateTime) }}</div>
            <el-tag v-if="row.isNew" type="success" size="small" effect="dark" class="mark-tag">新</el-tag>
            <el-tag v-if="row.isUpdated" type="warning" size="small" effect="dark" class="mark-tag" style="cursor: pointer" @click="showHistory(row)">更新</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        class="pagination"
        @size-change="loadData"
        @current-change="loadData"
      />
    </el-card>

    <!-- 历史记录对话框 -->
    <el-dialog v-model="historyVisible" title="资产扫描历史" width="800px">
      <div v-if="currentAsset" class="history-current">
        <span>当前资产: </span>
        <strong>{{ currentAsset.authority }}</strong>
        <span style="margin-left: 15px; color: #909399">{{ currentAsset.host }}:{{ currentAsset.port }}</span>
      </div>
      <el-table :data="historyList" v-loading="historyLoading" stripe size="small" max-height="400">
        <el-table-column prop="createTime" label="扫描时间" width="160" />
        <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
        <el-table-column prop="service" label="服务" width="80" />
        <el-table-column prop="httpStatus" label="状态码" width="80" />
        <el-table-column label="应用" min-width="150">
          <template #default="{ row }">
            <el-tag v-for="app in (row.app || []).slice(0, 3)" :key="app" size="small" type="success" style="margin-right: 4px">
              {{ app }}
            </el-tag>
            <span v-if="(row.app || []).length > 3" style="color: #909399; font-size: 12px">+{{ (row.app || []).length - 3 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="截图" width="80" align="center">
          <template #default="{ row }">
            <el-image 
              v-if="row.screenshot" 
              :src="getScreenshotUrl(row.screenshot)" 
              :preview-src-list="[getScreenshotUrl(row.screenshot)]"
              :z-index="9999"
              :preview-teleported="true"
              fit="cover"
              style="width: 50px; height: 40px; border-radius: 4px"
            />
            <span v-else style="color: #c0c4cc">-</span>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="!historyLoading && historyList.length === 0" class="history-empty">
        暂无历史记录
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Link } from '@element-plus/icons-vue'
import { getAssetList, getAssetStat, deleteAsset, batchDeleteAsset, getAssetHistory } from '@/api/asset'
import { useWorkspaceStore } from '@/stores/workspace'

const workspaceStore = useWorkspaceStore()
const loading = ref(false)
const tableData = ref([])
const selectedRows = ref([])
const activeTab = ref('quick')
const historyVisible = ref(false)
const historyLoading = ref(false)
const historyList = ref([])
const currentAsset = ref(null)

const stat = reactive({
  total: 0,
  newCount: 0,
  updatedCount: 0,
  topPorts: [],
  topService: [],
  topApp: [],
  topTitle: []
})

const searchForm = reactive({
  query: '',
  host: '',
  port: null,
  service: '',
  title: '',
  app: '',
  onlyNew: false,
  onlyUpdated: false,
  excludeCdn: false,
  sortByUpdate: true
})

const pagination = reactive({
  page: 1,
  pageSize: 50,
  total: 0
})

// 监听工作空间切换
function handleWorkspaceChanged() {
  pagination.page = 1
  loadData()
  loadStat()
}

onMounted(() => {
  loadData()
  loadStat()
  window.addEventListener('workspace-changed', handleWorkspaceChanged)
})

onUnmounted(() => {
  window.removeEventListener('workspace-changed', handleWorkspaceChanged)
})

async function loadData() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      onlyNew: searchForm.onlyNew,
      onlyUpdated: searchForm.onlyUpdated,
      excludeCdn: searchForm.excludeCdn,
      sortByUpdate: searchForm.sortByUpdate,
      workspaceId: workspaceStore.currentWorkspaceId || ''
    }

    // 根据当前Tab决定使用哪种查询方式
    if (activeTab.value === 'syntax' && searchForm.query) {
      params.query = searchForm.query
    } else {
      params.host = searchForm.host
      params.port = searchForm.port
      params.service = searchForm.service
      params.title = searchForm.title
      params.app = searchForm.app
    }

    const res = await getAssetList(params)
    if (res.code === 0) {
      tableData.value = res.list || []
      pagination.total = res.total
    }
  } finally {
    loading.value = false
  }
}

async function loadStat() {
  const res = await getAssetStat({ workspaceId: workspaceStore.currentWorkspaceId || '' })
  if (res.code === 0) {
    stat.total = res.totalAsset || 0
    stat.newCount = res.newCount || 0
    stat.updatedCount = res.updatedCount || 0
    stat.topPorts = res.topPorts || []
    stat.topService = res.topService || []
    stat.topApp = res.topApp || []
    stat.topTitle = res.topTitle || []
  }
}

function quickFilter(field, value) {
  // 端口需要转换为数字
  if (field === 'port') {
    searchForm[field] = parseInt(value, 10)
  } else {
    searchForm[field] = value
  }
  activeTab.value = 'quick'
  handleSearch()
}

function getAssetUrl(row) {
  if (row.service === 'https' || row.port === 443) {
    return `https://${row.host}:${row.port}`
  }
  return `http://${row.host}:${row.port}`
}

function formatHeader(header) {
  if (!header) return ''
  // 限制显示行数，最多显示10行
  const lines = header.split('\n').slice(0, 10)
  return lines.join('\n')
}

function formatHeaderWithStatus(row) {
  let result = ''
  // 添加状态行
  if (row.httpStatus) {
    const statusText = getStatusText(row.httpStatus)
    result = `HTTP/1.1 ${row.httpStatus} ${statusText}\n`
  }
  // 添加Header内容
  if (row.httpHeader) {
    result += row.httpHeader
  }
  return result
}

function getStatusText(status) {
  const statusMap = {
    '200': 'OK',
    '201': 'Created',
    '204': 'No Content',
    '301': 'Moved Permanently',
    '302': 'Found',
    '304': 'Not Modified',
    '400': 'Bad Request',
    '401': 'Unauthorized',
    '403': 'Forbidden',
    '404': 'Not Found',
    '500': 'Internal Server Error',
    '502': 'Bad Gateway',
    '503': 'Service Unavailable'
  }
  return statusMap[status] || ''
}

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

function formatTime(time) {
  if (!time) return '-'
  // 只显示日期部分
  return time.split(' ')[0] || time.substring(0, 10)
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  Object.assign(searchForm, {
    query: '',
    host: '',
    port: null,
    service: '',
    title: '',
    app: '',
    onlyNew: false,
    onlyUpdated: false,
    excludeCdn: false,
    sortByUpdate: true
  })
  handleSearch()
  loadStat()
}

async function handleDelete(row) {
  await ElMessageBox.confirm('确定删除该资产吗？', '提示', { type: 'warning' })
  const res = await deleteAsset({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    loadData()
    loadStat()
  }
}

function handleSelectionChange(rows) {
  selectedRows.value = rows
}

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return
  await ElMessageBox.confirm(`确定删除选中的 ${selectedRows.value.length} 条资产吗？`, '提示', { type: 'warning' })
  const ids = selectedRows.value.map(row => row.id)
  const res = await batchDeleteAsset({ ids })
  if (res.code === 0) {
    ElMessage.success(`成功删除 ${selectedRows.value.length} 条资产`)
    selectedRows.value = []
    loadData()
    loadStat()
  } else {
    ElMessage.error(res.msg)
  }
}

async function showHistory(row) {
  currentAsset.value = row
  historyVisible.value = true
  historyLoading.value = true
  historyList.value = []
  try {
    const res = await getAssetHistory({ assetId: row.id, limit: 20 })
    if (res.code === 0) {
      historyList.value = res.list || []
    } else {
      ElMessage.error(res.msg || '获取历史记录失败')
    }
  } catch (e) {
    ElMessage.error('获取历史记录失败')
  } finally {
    historyLoading.value = false
  }
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
      'custom': '自定义指纹'
    }
    
    // 处理组合来源，如 httpx+wappalyzer+custom(ID1,ID2)
    if (source.includes('+')) {
      const parts = source.split('+')
      const mappedParts = parts.map(part => {
        // 处理 custom(ID1,ID2) 格式
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
    
    // 处理单一来源 custom(ID1,ID2) 格式
    if (source.startsWith('custom(')) {
      const ids = source.match(/custom\(([^)]+)\)/)
      if (ids) {
        const idList = ids[1].split(',').map(id => id.trim())
        return `自定义指纹 (${idList.length}个ID: ${idList.join(', ')})`
      }
      return '自定义指纹'
    }
    
    // 处理旧格式 custom-ID
    if (source.startsWith('custom-')) {
      const id = source.substring(7)
      return `自定义指纹 (ID: ${id})`
    }
    
    return sourceMap[source] || source
  }
  return '未知来源'
}

// 根据来源返回标签类型
// 多来源合并时使用紫色(primary)表示高可信度
function getAppTagType(app) {
  if (!app) return 'info'
  // 三个来源都匹配 - 紫色(高可信度)
  if (app.includes('[httpx+wappalyzer+custom(')) return 'danger'
  // 两个来源匹配 - 紫色
  if (app.includes('[httpx+wappalyzer]')) return 'primary'
  if (app.includes('[httpx+custom(')) return 'danger'
  if (app.includes('[wappalyzer+custom(')) return 'danger'
  // 单一来源
  if (app.includes('[httpx]')) return 'success'
  if (app.includes('[wappalyzer]')) return 'success'
  if (app.includes('[builtin]')) return 'warning'
  if (app.includes('custom(') || app.includes('[custom-')) return 'danger'
  return 'info'
}

// 判断是否包含自定义指纹
function isCustomFingerprint(app) {
  return app && (app.includes('custom(') || app.includes('[custom-'))
}

// 获取自定义指纹的ID列表
function getCustomFingerprintIds(app) {
  if (!app) return []
  
  // 匹配 custom(ID1,ID2,ID3) 格式，支持在任意位置
  const match = app.match(/custom\(([^)]+)\)/)
  if (match) {
    return match[1].split(',').map(id => id.trim())
  }
  
  // 旧格式兼容: [custom-ID]
  const oldFormatMatch = app.match(/\[custom-([^\]]+)\]$/)
  if (oldFormatMatch) {
    return [oldFormatMatch[1]]
  }
  
  return []
}

// 获取自定义指纹的ID（兼容旧代码，返回第一个ID）
function getCustomFingerprintId(app) {
  const ids = getCustomFingerprintIds(app)
  return ids.length > 0 ? ids[0] : ''
}

// 获取tooltip内容
function getAppTooltip(app) {
  if (!app) return ''
  
  // 先获取来源信息
  const sourceInfo = getAppSource(app)
  
  // 如果包含自定义指纹，添加点击复制提示
  if (isCustomFingerprint(app)) {
    const ids = getCustomFingerprintIds(app)
    if (ids.length > 0) {
      return `${sourceInfo}\n点击复制指纹ID`
    }
  }
  
  return sourceInfo
}

// 处理指纹标签点击
function handleAppTagClick(app) {
  if (isCustomFingerprint(app)) {
    const ids = getCustomFingerprintIds(app)
    if (ids.length > 0) {
      const textToCopy = ids.length > 1 ? ids.join(',') : ids[0]
      navigator.clipboard.writeText(textToCopy).then(() => {
        if (ids.length > 1) {
          ElMessage.success(`已复制${ids.length}个指纹ID: ${textToCopy}`)
        } else {
          ElMessage.success(`已复制指纹ID: ${textToCopy}`)
        }
      }).catch(() => {
        // 降级方案
        const textarea = document.createElement('textarea')
        textarea.value = textToCopy
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
        if (ids.length > 1) {
          ElMessage.success(`已复制${ids.length}个指纹ID: ${textToCopy}`)
        } else {
          ElMessage.success(`已复制指纹ID: ${textToCopy}`)
        }
      })
    }
  }
}
</script>

<style lang="scss" scoped>
.asset-page {
  .search-card {
    margin-bottom: 15px;
    
    .search-tabs {
      :deep(.el-tabs__header) {
        margin-bottom: 10px;
      }
      
      .syntax-hints {
        margin-top: 8px;
        font-size: 12px;
        color: #909399;
        
        .hint-title {
          margin-right: 10px;
        }
        
        .hint-item {
          display: inline-block;
          padding: 2px 8px;
          margin-right: 10px;
          background: #f5f7fa;
          border-radius: 3px;
          color: #606266;
          cursor: pointer;
          
          &:hover {
            background: #e6f7ff;
            color: #409eff;
          }
        }
      }
    }

    .search-actions {
      margin-top: 10px;
      text-align: right;
    }
  }

  .stat-panel {
    display: flex;
    gap: 30px;
    
    .stat-column {
      min-width: 140px;
      
      .stat-title {
        font-weight: bold;
        color: #303133;
        margin-bottom: 8px;
        padding-bottom: 5px;
        border-bottom: 2px solid #409eff;
      }
      
      .stat-item {
        display: flex;
        align-items: center;
        padding: 3px 0;
        cursor: pointer;
        
        &:hover {
          background: #f5f7fa;
        }
        
        .stat-count {
          display: inline-block;
          min-width: 30px;
          padding: 1px 6px;
          margin-right: 8px;
          background: #409eff;
          color: #fff;
          border-radius: 3px;
          font-size: 12px;
          text-align: center;
        }
        
        .stat-name {
          color: #409eff;
          font-size: 13px;
        }
      }
    }
    
    .filter-column {
      margin-left: auto;
      
      .filter-options {
        display: flex;
        flex-direction: column;
        gap: 8px;
      }
    }
  }

  .table-card {
    .table-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 10px;
      
      .total-info {
        color: #606266;
        font-size: 13px;
      }
    }

    .asset-cell {
      display: flex;
      align-items: center;
      
      .asset-link {
        color: #409eff;
        text-decoration: none;
        
        &:hover {
          text-decoration: underline;
        }
      }
      
      .link-icon {
        margin-left: 4px;
        font-size: 12px;
        color: #409eff;
      }
    }
    
    .org-text, .location-text {
      color: #909399;
      font-size: 12px;
    }

    .service-text {
      color: #67c23a;
      font-size: 12px;
    }
    
    .title-text {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .app-tags-container {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
      max-height: 80px;
      overflow-y: auto;
    }

    .app-tag {
      margin: 0;
      flex-shrink: 0;
    }

    .clickable-tag {
      cursor: pointer;
      
      &:hover {
        opacity: 0.8;
        transform: scale(1.05);
      }
    }

    .mark-tag {
      margin-right: 4px;
      margin-top: 2px;
    }

    .fingerprint-info {
      .fingerprint-tabs {
        :deep(.el-tabs__header) {
          margin-bottom: 0;
          background: #f5f7fa;
        }
        
        :deep(.el-tabs__nav) {
          border: none;
        }
        
        :deep(.el-tabs__item) {
          padding: 0 10px;
          height: 28px;
          line-height: 28px;
          font-size: 12px;
        }
        
        :deep(.el-tabs__content) {
          padding: 0;
        }
        
        :deep(.el-tab-pane) {
          padding: 0;
        }
      }
      
      .tab-content {
        margin: 0;
        padding: 8px;
        background: #fafafa;
        font-size: 11px;
        line-height: 1.5;
        max-height: 160px;
        overflow-y: auto;
        white-space: pre-wrap;
        word-break: break-all;
        color: #606266;
        font-family: Consolas, Monaco, monospace;
      }
      
      .no-data {
        display: block;
        padding: 10px;
        color: #c0c4cc;
        font-size: 12px;
        text-align: center;
      }
    }

    .screenshot-img {
      width: 80px;
      height: 60px;
      border-radius: 4px;
      cursor: pointer;
      border: 1px solid #ebeef5;
    }

    .no-screenshot {
      color: #c0c4cc;
    }

    .update-time {
      font-size: 12px;
      color: #606266;
    }

    .pagination {
      margin-top: 15px;
      justify-content: flex-end;
    }
  }

  .history-current {
    margin-bottom: 15px;
    padding: 10px;
    background: #f5f7fa;
    border-radius: 4px;
  }

  .history-empty {
    text-align: center;
    padding: 40px;
    color: #909399;
  }
}
</style>
