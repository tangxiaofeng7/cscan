<template>
  <div class="online-search-page">
    <!-- 搜索区域 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="数据源">
          <el-select v-model="searchForm.source" style="width: 120px">
            <el-option label="Fofa" value="fofa" />
            <el-option label="Hunter" value="hunter" />
            <el-option label="Quake" value="quake" />
          </el-select>
        </el-form-item>
        <el-form-item label="查询语句" style="flex: 1">
          <el-input
            v-model="searchForm.query"
            placeholder="输入查询语句，如: ip=1.1.1.1 或 domain=example.com"
            style="width: 400px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="数量">
          <el-select v-model="searchForm.size" style="width: 100px">
            <el-option :value="10" label="10" />
            <el-option :value="50" label="50" />
            <el-option :value="100" label="100" />
            <el-option :value="500" label="500" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSearch">搜索</el-button>
          <el-button @click="handleImport" :disabled="!tableData.length">导入资产</el-button>
          <el-button @click="showConfigDialog">API配置</el-button>
        </el-form-item>
      </el-form>

      <!-- 快捷查询 -->
      <div class="quick-search">
        <span class="label">快捷查询：</span>
        <el-tag
          v-for="item in quickQueries"
          :key="item.query"
          class="quick-tag"
          @click="applyQuickQuery(item)"
        >
          {{ item.label }}
        </el-tag>
      </div>
    </el-card>

    <!-- 结果表格 -->
    <el-card class="result-card">
      <template #header>
        <div class="card-header">
          <span>搜索结果</span>
          <span v-if="total > 0" class="total">共 {{ total }} 条</span>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" stripe max-height="600">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="ip" label="IP" width="140" />
        <el-table-column prop="port" label="端口" width="80" />
        <el-table-column prop="protocol" label="协议" width="80" />
        <el-table-column prop="domain" label="域名" min-width="150" show-overflow-tooltip />
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="server" label="Server" width="120" show-overflow-tooltip />
        <el-table-column prop="product" label="产品" width="120" show-overflow-tooltip />
        <el-table-column prop="country" label="国家" width="80" />
        <el-table-column prop="city" label="城市" width="100" />
        <el-table-column prop="icp" label="ICP备案" width="150" show-overflow-tooltip />
      </el-table>

      <el-pagination
        v-if="total > 0"
        v-model:current-page="searchForm.page"
        :page-size="searchForm.size"
        :total="total"
        layout="total, prev, pager, next"
        class="pagination"
        @current-change="handleSearch"
      />
    </el-card>

    <!-- API配置对话框 -->
    <el-dialog v-model="configDialogVisible" title="API配置" width="600px">
      <el-tabs v-model="configTab">
        <el-tab-pane label="Fofa" name="fofa">
          <el-form label-width="80px">
            <el-form-item label="Email">
              <el-input v-model="apiConfigs.fofa.key" placeholder="Fofa账号邮箱" />
            </el-form-item>
            <el-form-item label="API Key">
              <el-input v-model="apiConfigs.fofa.secret" placeholder="Fofa API Key" show-password />
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="Hunter" name="hunter">
          <el-form label-width="80px">
            <el-form-item label="API Key">
              <el-input v-model="apiConfigs.hunter.key" placeholder="Hunter API Key" show-password />
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="Quake" name="quake">
          <el-form label-width="80px">
            <el-form-item label="API Key">
              <el-input v-model="apiConfigs.quake.key" placeholder="Quake API Key" show-password />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>

    <!-- 语法帮助 -->
    <el-card class="help-card">
      <template #header>语法帮助</template>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="Fofa" name="fofa">
          <div class="syntax-help">
            <p><code>ip="1.1.1.1"</code> - 搜索指定IP</p>
            <p><code>domain="example.com"</code> - 搜索指定域名</p>
            <p><code>title="后台"</code> - 搜索标题包含关键词</p>
            <p><code>body="content"</code> - 搜索正文包含关键词</p>
            <p><code>port="80"</code> - 搜索指定端口</p>
            <p><code>icp="京ICP备"</code> - 搜索ICP备案</p>
            <p><code>org="公司名"</code> - 搜索组织</p>
            <p>组合查询: <code>ip="1.1.1.1" && port="80"</code></p>
          </div>
        </el-tab-pane>
        <el-tab-pane label="Hunter" name="hunter">
          <div class="syntax-help">
            <p><code>ip="1.1.1.1"</code> - 搜索指定IP</p>
            <p><code>domain.suffix="example.com"</code> - 搜索域名后缀</p>
            <p><code>web.title="后台"</code> - 搜索网页标题</p>
            <p><code>icp.name="公司名"</code> - 搜索ICP主体</p>
            <p><code>icp.number="京ICP备"</code> - 搜索ICP备案号</p>
            <p><code>port="443"</code> - 搜索指定端口</p>
          </div>
        </el-tab-pane>
        <el-tab-pane label="Quake" name="quake">
          <div class="syntax-help">
            <p><code>ip:"1.1.1.1"</code> - 搜索指定IP</p>
            <p><code>domain:"example.com"</code> - 搜索指定域名</p>
            <p><code>title:"后台"</code> - 搜索标题</p>
            <p><code>service:"http"</code> - 搜索服务类型</p>
            <p><code>port:"80"</code> - 搜索指定端口</p>
            <p><code>country:"CN"</code> - 搜索国家</p>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const activeTab = ref('fofa')
const configDialogVisible = ref(false)
const configTab = ref('fofa')

const searchForm = reactive({
  source: 'fofa',
  query: '',
  page: 1,
  size: 50
})

const apiConfigs = reactive({
  fofa: { key: '', secret: '' },
  hunter: { key: '', secret: '' },
  quake: { key: '', secret: '' }
})

onMounted(() => loadConfigs())

const quickQueries = [
  { label: 'IP搜索', query: 'ip="1.1.1.1"' },
  { label: '域名搜索', query: 'domain="example.com"' },
  { label: '标题搜索', query: 'title="后台管理"' },
  { label: 'ICP搜索', query: 'icp="京ICP备"' },
  { label: '端口搜索', query: 'port="3389"' },
]

async function handleSearch() {
  if (!searchForm.query) {
    ElMessage.warning('请输入查询语句')
    return
  }

  loading.value = true
  try {
    const res = await request.post('/onlineapi/search', {
      platform: searchForm.source,
      query: searchForm.query,
      page: searchForm.page,
      pageSize: searchForm.size
    })
    if (res.code === 0) {
      tableData.value = res.list || []
      total.value = res.total || 0
    } else {
      ElMessage.error(res.msg || '搜索失败')
    }
  } finally {
    loading.value = false
  }
}

function applyQuickQuery(item) {
  searchForm.query = item.query
}

async function handleImport() {
  await ElMessageBox.confirm(`确定将 ${tableData.value.length} 条数据导入到资产库吗？`, '提示')
  
  const res = await request.post('/onlineapi/import', { assets: tableData.value })
  
  if (res.code === 0) {
    ElMessage.success(res.msg || '导入成功')
  } else {
    ElMessage.error(res.msg || '导入失败')
  }
}

async function loadConfigs() {
  const res = await request.post('/onlineapi/config/list', {})
  if (res.code === 0 && res.list) {
    res.list.forEach(item => {
      if (apiConfigs[item.platform]) {
        apiConfigs[item.platform].key = item.key
        apiConfigs[item.platform].secret = item.secret
      }
    })
  }
}

function showConfigDialog() {
  configDialogVisible.value = true
}

async function saveConfig() {
  const platform = configTab.value
  const config = apiConfigs[platform]
  const res = await request.post('/onlineapi/config/save', {
    platform,
    key: config.key,
    secret: config.secret
  })
  if (res.code === 0) {
    ElMessage.success('保存成功')
  } else {
    ElMessage.error(res.msg || '保存失败')
  }
}
</script>

<style lang="scss" scoped>
.online-search-page {
  .search-card {
    margin-bottom: 20px;

    .quick-search {
      margin-top: 10px;
      padding-top: 10px;
      border-top: 1px solid #eee;

      .label {
        color: #666;
        margin-right: 10px;
      }

      .quick-tag {
        cursor: pointer;
        margin-right: 8px;

        &:hover {
          background: #409EFF;
          color: #fff;
        }
      }
    }
  }

  .result-card {
    margin-bottom: 20px;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .total {
        color: #999;
        font-size: 14px;
      }
    }

    .pagination {
      margin-top: 20px;
      justify-content: flex-end;
    }
  }

  .help-card {
    .syntax-help {
      p {
        margin: 8px 0;
        line-height: 1.6;

        code {
          background: var(--el-fill-color-light);
          padding: 2px 6px;
          border-radius: 4px;
          color: #e83e8c;
        }
      }
    }
  }
}
</style>
