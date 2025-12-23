<template>
  <div class="vul-page">
    <!-- 搜索区域 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="目标">
          <el-input v-model="searchForm.authority" placeholder="IP:端口" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="危害等级">
          <el-select v-model="searchForm.severity" placeholder="全部" clearable style="width: 120px">
            <el-option label="严重" value="critical" />
            <el-option label="高危" value="high" />
            <el-option label="中危" value="medium" />
            <el-option label="低危" value="low" />
            <el-option label="信息" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源">
          <el-select v-model="searchForm.source" placeholder="全部" clearable style="width: 120px">
            <el-option label="Nuclei" value="nuclei" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="danger" :disabled="selectedRows.length === 0" @click="handleBatchDelete">
            批量删除 ({{ selectedRows.length }})
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <el-table 
        :data="tableData" 
        v-loading="loading" 
        stripe
        max-height="500"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="authority" label="目标" min-width="150" />
        <el-table-column prop="url" label="URL" min-width="250" show-overflow-tooltip />
        <el-table-column prop="pocFile" label="POC" min-width="200" show-overflow-tooltip />
        <el-table-column prop="severity" label="危害等级" width="100">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)" size="small">
              {{ getSeverityLabel(row.severity) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="100" />
        <el-table-column prop="createTime" label="发现时间" width="160" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showDetail(row)">详情</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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

    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" title="漏洞详情" width="800px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="目标">{{ currentVul.authority }}</el-descriptions-item>
        <el-descriptions-item label="危害等级">
          <el-tag :type="getSeverityType(currentVul.severity)">
            {{ getSeverityLabel(currentVul.severity) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="URL" :span="2">{{ currentVul.url }}</el-descriptions-item>
        <el-descriptions-item label="POC文件" :span="2">{{ currentVul.pocFile }}</el-descriptions-item>
        <el-descriptions-item label="来源">{{ currentVul.source }}</el-descriptions-item>
        <el-descriptions-item label="发现时间">{{ currentVul.createTime }}</el-descriptions-item>
        <!-- 知识库信息 -->
        <el-descriptions-item label="CVE编号" v-if="currentVul.cveId">{{ currentVul.cveId }}</el-descriptions-item>
        <el-descriptions-item label="CWE编号" v-if="currentVul.cweId">{{ currentVul.cweId }}</el-descriptions-item>
        <el-descriptions-item label="CVSS评分" v-if="currentVul.cvssScore">{{ currentVul.cvssScore }}</el-descriptions-item>
        <el-descriptions-item label="扫描次数" v-if="currentVul.scanCount">{{ currentVul.scanCount }}</el-descriptions-item>
        <el-descriptions-item label="首次发现" v-if="currentVul.firstSeenTime">{{ currentVul.firstSeenTime }}</el-descriptions-item>
        <el-descriptions-item label="最近发现" v-if="currentVul.lastSeenTime">{{ currentVul.lastSeenTime }}</el-descriptions-item>
        <el-descriptions-item label="修复建议" :span="2" v-if="currentVul.remediation">
          <pre class="result-pre">{{ currentVul.remediation }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="参考链接" :span="2" v-if="currentVul.references && currentVul.references.length">
          <div v-for="(ref, idx) in currentVul.references" :key="idx">
            <a :href="ref" target="_blank" rel="noopener">{{ ref }}</a>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="验证结果" :span="2">
          <pre class="result-pre">{{ currentVul.result }}</pre>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 证据链 -->
      <template v-if="currentVul.evidence">
        <el-divider content-position="left">证据链</el-divider>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="匹配器名称" v-if="currentVul.evidence.matcherName">
            {{ currentVul.evidence.matcherName }}
          </el-descriptions-item>
          <el-descriptions-item label="提取结果" v-if="currentVul.evidence.extractedResults && currentVul.evidence.extractedResults.length">
            <el-tag v-for="(item, idx) in currentVul.evidence.extractedResults" :key="idx" style="margin-right: 5px; margin-bottom: 5px;">
              {{ item }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="cURL命令" v-if="currentVul.evidence.curlCommand">
            <pre class="result-pre">{{ currentVul.evidence.curlCommand }}</pre>
          </el-descriptions-item>
          <el-descriptions-item label="请求内容" v-if="currentVul.evidence.request">
            <pre class="result-pre">{{ currentVul.evidence.request }}</pre>
          </el-descriptions-item>
          <el-descriptions-item label="响应内容" v-if="currentVul.evidence.response">
            <pre class="result-pre">{{ currentVul.evidence.response }}</pre>
            <el-tag v-if="currentVul.evidence.responseTruncated" type="warning" size="small">响应已截断</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'
import { useWorkspaceStore } from '@/stores/workspace'

const workspaceStore = useWorkspaceStore()
const loading = ref(false)
const tableData = ref([])
const detailVisible = ref(false)
const currentVul = ref({})
const selectedRows = ref([])

const searchForm = reactive({
  authority: '',
  severity: '',
  source: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 监听工作空间切换
function handleWorkspaceChanged() {
  pagination.page = 1
  loadData()
}

onMounted(() => {
  loadData()
  window.addEventListener('workspace-changed', handleWorkspaceChanged)
})

onUnmounted(() => {
  window.removeEventListener('workspace-changed', handleWorkspaceChanged)
})

async function loadData() {
  loading.value = true
  try {
    const res = await request.post('/vul/list', {
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize,
      workspaceId: workspaceStore.currentWorkspaceId || ''
    })
    if (res.code === 0) {
      tableData.value = res.list || []
      pagination.total = res.total
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  Object.assign(searchForm, { authority: '', severity: '', source: '' })
  handleSearch()
}

function handleSelectionChange(rows) {
  selectedRows.value = rows
}

function getSeverityType(severity) {
  const map = { critical: 'danger', high: 'danger', medium: 'warning', low: 'info', info: '' }
  return map[severity] || ''
}

function getSeverityLabel(severity) {
  const map = { critical: '严重', high: '高危', medium: '中危', low: '低危', info: '信息' }
  return map[severity] || severity
}

async function showDetail(row) {
  try {
    const res = await request.post('/vul/detail', { id: row.id })
    if (res.code === 0 && res.data) {
      currentVul.value = res.data
    } else {
      currentVul.value = row
    }
  } catch (e) {
    currentVul.value = row
  }
  detailVisible.value = true
}

async function handleDelete(row) {
  await ElMessageBox.confirm('确定删除该漏洞记录吗？', '提示', { type: 'warning' })
  const res = await request.post('/vul/delete', { id: row.id })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    loadData()
  } else {
    ElMessage.error(res.msg || '删除失败')
  }
}

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return
  await ElMessageBox.confirm(`确定删除选中的 ${selectedRows.value.length} 条漏洞记录吗？`, '提示', { type: 'warning' })
  const ids = selectedRows.value.map(row => row.id)
  const res = await request.post('/vul/batchDelete', { ids })
  if (res.code === 0) {
    ElMessage.success(res.msg || '删除成功')
    selectedRows.value = []
    loadData()
  } else {
    ElMessage.error(res.msg || '删除失败')
  }
}
</script>

<style lang="scss" scoped>
.vul-page {
  .search-card {
    margin-bottom: 20px;
  }

  .pagination {
    margin-top: 20px;
    justify-content: flex-end;
  }

  .result-pre {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 300px;
    overflow: auto;
    background: #1e1e1e;
    color: #d4d4d4;
    padding: 12px;
    border-radius: 6px;
    font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
    font-size: 13px;
    line-height: 1.5;
    border: 1px solid #3c3c3c;
  }

  // 浅色主题适配
  :deep(.el-dialog) {
    .result-pre {
      background: #1e1e1e;
      color: #d4d4d4;
      border: 1px solid #3c3c3c;
    }
  }
}

// 证据链代码块样式
:deep(.el-descriptions) {
  .el-descriptions__cell {
    .result-pre {
      background: #1e1e1e;
      color: #d4d4d4;
    }
  }
}
</style>
