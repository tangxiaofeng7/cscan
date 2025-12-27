<template>
  <div class="task-page">
    <!-- 操作栏 -->
    <el-card class="action-card">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>新建任务
      </el-button>
      <el-switch
        v-model="autoRefresh"
        style="margin-left: 20px"
        active-text="自动刷新(间隔30秒)"
        inactive-text=""
        @change="handleAutoRefreshChange"
      />
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <div style="margin-bottom: 15px">
        <el-button type="danger" :disabled="selectedRows.length === 0" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon>批量删除 ({{ selectedRows.length }})
        </el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe max-height="500" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" label="任务名称" min-width="150" />
        <el-table-column prop="target" label="扫描目标" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="progress" label="进度" width="150">
          <template #default="{ row }">
            <div>
              <el-progress :percentage="row.progress" :stroke-width="6" />
              <div v-if="row.subTaskCount > 1" class="sub-task-info">
                子任务: {{ row.subTaskDone }}/{{ row.subTaskCount }}
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="定时任务" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.isCron" type="success" size="small">{{ row.cronRule }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="160" />
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'CREATED'" type="success" link size="small" @click="handleStart(row)">启动</el-button>
            <el-button v-if="row.status === 'CREATED'" type="warning" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="row.status === 'STARTED'" type="warning" link size="small" @click="handlePause(row)">暂停</el-button>
            <el-button v-if="row.status === 'PAUSED'" type="success" link size="small" @click="handleResume(row)">继续</el-button>
            <el-button v-if="['STARTED', 'PAUSED', 'PENDING'].includes(row.status)" type="danger" link size="small" @click="handleStop(row)">停止</el-button>
            <el-button type="primary" link size="small" @click="showDetail(row)">详情</el-button>
            <el-button type="info" link size="small" @click="showLogs(row)">日志</el-button>
            <el-button type="info" link size="small" @click="viewReport(row)">报告</el-button>
            <el-button v-if="['SUCCESS', 'FAILURE', 'STOPPED'].includes(row.status)" type="warning" link size="small" @click="handleRetry(row)">重新执行</el-button>
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

    <!-- 任务详情对话框 -->
    <el-dialog v-model="detailVisible" title="任务详情" width="700px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="任务名称">{{ currentTask.name }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentTask.status)">{{ currentTask.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="进度">
          <el-progress :percentage="currentTask.progress" :stroke-width="10" style="width: 150px" />
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ currentTask.createTime }}</el-descriptions-item>
        <el-descriptions-item label="扫描目标" :span="2">
          <div style="max-height: 100px; overflow-y: auto; white-space: pre-wrap">{{ currentTask.target }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="执行结果" :span="2">
          <div style="max-height: 100px; overflow-y: auto">{{ currentTask.result || '-' }}</div>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 新建/编辑任务对话框 - Tab页布局 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑任务' : '新建任务'" width="720px" top="5vh" class="task-dialog">
      <el-tabs v-model="activeTab" class="task-tabs">
        <!-- 基本信息 Tab -->
        <el-tab-pane label="基本信息" name="basic">
          <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" class="tab-form">
            <el-form-item label="任务名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入任务名称" />
            </el-form-item>
            <el-form-item label="扫描目标" prop="target">
              <el-input v-model="form.target" type="textarea" :rows="6" placeholder="每行一个目标，支持格式:&#10;• IP: 192.168.1.1&#10;• CIDR: 192.168.1.0/24&#10;• IP范围: 192.168.1.1-192.168.1.100&#10;• 域名: example.com" />
            </el-form-item>
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="工作空间">
                  <el-select v-model="form.workspaceId" placeholder="选择工作空间" clearable style="width: 100%">
                    <el-option v-for="ws in workspaceStore.workspaces" :key="ws.id" :label="ws.name" :value="ws.id" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="所属组织">
                  <el-select v-model="form.orgId" placeholder="选择组织" clearable style="width: 100%">
                    <el-option v-for="org in organizations" :key="org.id" :label="org.name" :value="org.id" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="指定Worker">
              <el-select v-model="form.workers" multiple placeholder="不选则任意Worker执行" clearable style="width: 100%">
                <el-option v-for="w in workers" :key="w.name" :label="`${w.name} (${w.ip})`" :value="w.name" />
              </el-select>
            </el-form-item>
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="定时任务">
                  <el-switch v-model="form.isCron" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item v-if="form.isCron" label="Cron表达式">
                  <el-input v-model="form.cronRule" placeholder="0 0 * * *" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-tab-pane>

        <!-- 端口扫描 Tab -->
        <el-tab-pane name="portscan">
          <template #label>
            <span>端口扫描 <el-tag v-if="form.portscanEnable" type="success" size="small" style="margin-left:4px">开</el-tag></span>
          </template>
          <el-form label-width="100px" class="tab-form">
            <el-form-item label="启用">
              <el-switch v-model="form.portscanEnable" />
            </el-form-item>
            <template v-if="form.portscanEnable">
              <el-form-item label="扫描工具">
                <el-radio-group v-model="form.portscanTool">
                  <el-radio label="naabu">Naabu (推荐)</el-radio>
                  <el-radio label="masscan" :disabled="!availableTools.masscan">
                    Masscan <span v-if="!availableTools.masscan" class="tool-tip">(未安装)</span>
                  </el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="端口范围">
                <el-select v-model="form.ports" filterable allow-create default-first-option style="width: 100%">
                  <el-option label="top100 - 常用100端口" value="top100" />
                  <el-option label="top1000 - 常用1000端口" value="top1000" />
                  <el-option label="80,443,8080,8443 - Web常用" value="80,443,8080,8443" />
                  <el-option label="1-65535 - 全端口" value="1-65535" />
                </el-select>
              </el-form-item>
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="扫描速率">
                    <el-input-number v-model="form.portscanRate" :min="100" :max="100000" style="width:100%" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="端口阈值">
                    <el-input-number v-model="form.portThreshold" :min="0" :max="65535" style="width:100%" />
                  </el-form-item>
                </el-col>
              </el-row>
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item v-if="form.portscanTool === 'naabu'" label="扫描类型">
                    <el-radio-group v-model="form.scanType">
                      <el-radio label="c">CONNECT</el-radio>
                      <el-radio label="s">SYN</el-radio>
                    </el-radio-group>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="超时(秒)">
                    <el-input-number v-model="form.portscanTimeout" :min="5" :max="1200" style="width:100%" />
                  </el-form-item>
                </el-col>
              </el-row>
              <el-form-item label="高级选项">
                <el-checkbox v-model="form.skipHostDiscovery">跳过主机发现 (-Pn)</el-checkbox>
                <span class="form-hint">跳过主机存活检测，直接扫描端口</span>
              </el-form-item>
            </template>
          </el-form>
        </el-tab-pane>

        <!-- 端口识别 Tab -->
        <el-tab-pane name="portidentify">
          <template #label>
            <span>端口识别 <el-tag v-if="form.portidentifyEnable" type="success" size="small" style="margin-left:4px">开</el-tag></span>
          </template>
          <el-form label-width="100px" class="tab-form">
            <el-form-item label="启用">
              <el-switch v-model="form.portidentifyEnable" :disabled="!availableTools.nmap" />
              <span v-if="!availableTools.nmap" class="tool-tip" style="margin-left:10px">(Nmap 未安装)</span>
            </el-form-item>
            <template v-if="form.portidentifyEnable">
              <el-form-item label="超时(秒)">
                <el-input-number v-model="form.portidentifyTimeout" :min="5" :max="300" />
                <span class="form-hint">单个主机超时时间</span>
              </el-form-item>
              <el-form-item label="Nmap参数">
                <el-input v-model="form.portidentifyArgs" placeholder="-sV --version-intensity 5" />
              </el-form-item>
            </template>
            <el-alert v-if="!form.portidentifyEnable" type="info" :closable="false" show-icon>
              <template #title>端口识别使用 Nmap 对开放端口进行服务版本探测</template>
            </el-alert>
          </el-form>
        </el-tab-pane>

        <!-- 指纹识别 Tab -->
        <el-tab-pane name="fingerprint">
          <template #label>
            <span>指纹识别 <el-tag v-if="form.fingerprintEnable" type="success" size="small" style="margin-left:4px">开</el-tag></span>
          </template>
          <el-form label-width="100px" class="tab-form">
            <el-form-item label="启用">
              <el-switch v-model="form.fingerprintEnable" />
            </el-form-item>
            <template v-if="form.fingerprintEnable">
              <el-form-item label="探测工具">
                <el-checkbox v-model="form.fingerprintHttpx">Httpx探测 (推荐)</el-checkbox>
              </el-form-item>
              <el-form-item label="附加功能">
                <el-checkbox v-model="form.fingerprintIconHash">Icon Hash</el-checkbox>
                <el-checkbox v-model="form.fingerprintWappalyzer">Wappalyzer</el-checkbox>
                <el-checkbox v-model="form.fingerprintCustomEngine">自定义指纹</el-checkbox>
                <el-checkbox v-model="form.fingerprintScreenshot">网页截图</el-checkbox>
              </el-form-item>
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="超时(秒)">
                    <el-input-number v-model="form.fingerprintTimeout" :min="5" :max="120" style="width:100%" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="并发数">
                    <el-input-number v-model="form.fingerprintConcurrency" :min="1" :max="50" style="width:100%" />
                  </el-form-item>
                </el-col>
              </el-row>
            </template>
          </el-form>
        </el-tab-pane>

        <!-- 漏洞扫描 Tab -->
        <el-tab-pane name="pocscan">
          <template #label>
            <span>漏洞扫描 <el-tag v-if="form.pocscanEnable" type="success" size="small" style="margin-left:4px">开</el-tag></span>
          </template>
          <el-form label-width="100px" class="tab-form">
            <el-form-item label="启用">
              <el-switch v-model="form.pocscanEnable" />
              <span class="form-hint">使用 Nuclei 引擎</span>
            </el-form-item>
            <template v-if="form.pocscanEnable">
              <el-form-item label="自动扫描">
                <el-checkbox v-model="form.pocscanAutoScan" :disabled="form.pocscanCustomOnly">自定义标签映射</el-checkbox>
                <el-checkbox v-model="form.pocscanAutomaticScan" :disabled="form.pocscanCustomOnly">Wappalyzer自动扫描</el-checkbox>
              </el-form-item>
              <el-form-item label="自定义POC">
                <el-checkbox v-model="form.pocscanCustomOnly">只使用自定义POC</el-checkbox>
              </el-form-item>
              <el-form-item label="严重级别">
                <el-checkbox-group v-model="form.pocscanSeverity">
                  <el-checkbox label="critical">Critical</el-checkbox>
                  <el-checkbox label="high">High</el-checkbox>
                  <el-checkbox label="medium">Medium</el-checkbox>
                  <el-checkbox label="low">Low</el-checkbox>
                  <el-checkbox label="info">Info</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
              <el-form-item label="目标超时">
                <el-input-number v-model="form.pocscanTargetTimeout" :min="30" :max="600" />
                <span class="form-hint">秒</span>
              </el-form-item>
            </template>
          </el-form>
        </el-tab-pane>

        <!-- 高级设置 Tab -->
        <el-tab-pane label="高级设置" name="advanced">
          <el-form label-width="100px" class="tab-form">
            <el-form-item label="任务拆分">
              <el-input-number v-model="form.batchSize" :min="0" :max="1000" :step="10" />
              <span class="form-hint">每批目标数量，0=不拆分</span>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '保存' : '创建任务' }}</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 任务日志对话框 -->
    <el-dialog v-model="logDialogVisible" title="任务日志" width="1000px" @close="closeLogDialog">
      <div class="log-progress" v-if="currentLogTask">
        <div class="progress-info">
          <span class="task-name">{{ currentLogTask.name }}</span>
          <el-tag :type="getStatusType(currentLogTask.status)" size="small">{{ currentLogTask.status }}</el-tag>
        </div>
        <el-progress :percentage="currentLogTask.progress" :status="currentLogTask.status === 'SUCCESS' ? 'success' : (currentLogTask.status === 'FAILURE' ? 'exception' : '')" :stroke-width="12" />
      </div>
      <div class="log-filter">
        <el-input v-model="logSearchKeyword" placeholder="搜索日志..." clearable size="small" style="width: 180px; margin-right: 10px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="logWorkerFilter" placeholder="筛选Worker" clearable size="small" style="width: 150px">
          <el-option label="全部Worker" value="" />
          <el-option v-for="w in logWorkers" :key="w" :label="w" :value="w" />
        </el-select>
        <el-select v-model="logLevelFilter" placeholder="筛选级别" clearable size="small" style="width: 120px; margin-left: 10px">
          <el-option label="全部级别" value="" />
          <el-option label="INFO" value="INFO" />
          <el-option label="WARN" value="WARN" />
          <el-option label="ERROR" value="ERROR" />
        </el-select>
        <el-switch v-model="logAutoRefresh" size="small" active-text="自动刷新" style="margin-left: 15px" @change="handleLogAutoRefreshChange" />
        <span class="log-stats">共 {{ filteredLogs.length }} 条日志</span>
      </div>
      <div class="log-container" ref="logContainerRef">
        <div v-if="filteredLogs.length === 0" class="log-empty">暂无日志</div>
        <div v-for="(log, index) in filteredLogs" :key="index" class="log-entry" :class="'log-' + log.level.toLowerCase()">
          <span class="log-time">{{ formatLogTime(log.timestamp) }}</span>
          <span class="log-level">[{{ log.level }}]</span>
          <span class="log-worker">{{ log.workerName }}</span>
          <span class="log-message">{{ log.displayMessage }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="closeLogDialog">关闭</el-button>
        <el-button type="primary" @click="refreshLogs">刷新</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Search } from '@element-plus/icons-vue'
import { getTaskList, createTask, deleteTask, batchDeleteTask, retryTask, startTask, pauseTask, resumeTask, stopTask, updateTask, getTaskLogs, getWorkerList, saveScanConfig, getScanConfig } from '@/api/task'
import { useWorkspaceStore } from '@/stores/workspace'
import { validateTargets, formatValidationErrors } from '@/utils/target'
import request from '@/api/request'

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const detailVisible = ref(false)
const logDialogVisible = ref(false)
const tableData = ref([])
const organizations = ref([])
const workers = ref([])
const formRef = ref()
const logContainerRef = ref()
const currentTask = ref({})
const selectedRows = ref([])
const autoRefresh = ref(true)
const activeTab = ref('basic')
const isEdit = ref(false)
const taskLogs = ref([])
const currentLogTaskId = ref('')
const currentLogTask = ref(null)
const logIdSet = new Set()
const logWorkerFilter = ref('')
const logLevelFilter = ref('')
const logSearchKeyword = ref('')
const logAutoRefresh = ref(true)
let refreshTimer = null
let logEventSource = null
let logPollingTimer = null

const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const form = reactive({
  id: '',
  name: '',
  target: '',
  workspaceId: '',
  orgId: '',
  isCron: false,
  cronRule: '',
  workers: [],
  batchSize: 50,
  portscanEnable: true,
  portscanTool: 'naabu',
  portscanRate: 1000,
  ports: 'top100',
  portThreshold: 100,
  scanType: 'c',
  portscanTimeout: 60,
  skipHostDiscovery: false,
  portidentifyEnable: false,
  portidentifyTimeout: 30,
  portidentifyArgs: '',
  fingerprintEnable: true,
  fingerprintHttpx: true,
  fingerprintIconHash: true,
  fingerprintWappalyzer: true,
  fingerprintCustomEngine: false,
  fingerprintScreenshot: false,
  fingerprintTimeout: 30,
  fingerprintConcurrency: 10,
  pocscanEnable: false,
  pocscanAutoScan: true,
  pocscanAutomaticScan: true,
  pocscanCustomOnly: false,
  pocscanSeverity: ['critical', 'high', 'medium'],
  pocscanTargetTimeout: 600
})

const targetValidator = (rule, value, callback) => {
  if (!value) { callback(new Error('请输入扫描目标')); return }
  const errors = validateTargets(value)
  errors.length > 0 ? callback(new Error(formatValidationErrors(errors))) : callback()
}

const rules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  target: [{ required: true, message: '请输入扫描目标', trigger: 'blur' }, { validator: targetValidator, trigger: 'blur' }]
}

const logWorkers = computed(() => {
  const set = new Set()
  taskLogs.value.forEach(log => { if (log.workerName) set.add(log.workerName) })
  return Array.from(set).sort()
})

const filteredLogs = computed(() => {
  const keyword = logSearchKeyword.value.toLowerCase()
  return taskLogs.value.filter(log => {
    if (logWorkerFilter.value && log.workerName !== logWorkerFilter.value) return false
    if (logLevelFilter.value && log.level !== logLevelFilter.value) return false
    if (keyword) {
      const msg = (log.displayMessage || log.message || '').toLowerCase()
      if (!msg.includes(keyword) && !(log.level || '').toLowerCase().includes(keyword)) return false
    }
    return true
  })
})

const availableTools = computed(() => {
  const tools = { nmap: false, masscan: false }
  for (const w of workers.value) {
    if (w.tools) {
      if (w.tools.nmap) tools.nmap = true
      if (w.tools.masscan) tools.masscan = true
    }
  }
  return tools
})

// 监听工具可用性，自动关闭不可用的功能
watch(availableTools, (tools) => {
  if (!tools.nmap && form.portidentifyEnable) {
    form.portidentifyEnable = false
  }
  if (!tools.masscan && form.portscanTool === 'masscan') {
    form.portscanTool = 'naabu'
  }
}, { immediate: true })

function formatLogTime(timestamp) {
  if (!timestamp) return ''
  const match = timestamp.match(/(\d{2}:\d{2}:\d{2})/)
  return match ? match[1] : timestamp
}

function parseLogMessage(log) {
  let message = log.message || '', subTask = 'main'
  const subMatch = message.match(/^\[Sub-(\d+)\]\s*/)
  if (subMatch) { subTask = subMatch[1]; message = message.replace(subMatch[0], '') }
  return { ...log, subTask, displayMessage: message }
}

onMounted(() => {
  loadData()
  loadOrganizations()
  loadWorkers()
  if (autoRefresh.value) startAutoRefresh()
  window.addEventListener('workspace-changed', () => { pagination.page = 1; loadData() })
})

onUnmounted(() => {
  stopAutoRefresh()
  if (logEventSource) { logEventSource.close(); logEventSource = null }
})

function handleAutoRefreshChange(val) { val ? startAutoRefresh() : stopAutoRefresh() }
function startAutoRefresh() { stopAutoRefresh(); refreshTimer = setInterval(() => loadData(), 30000) }
function stopAutoRefresh() { if (refreshTimer) { clearInterval(refreshTimer); refreshTimer = null } }

async function loadData() {
  loading.value = true
  try {
    const res = await getTaskList({ page: pagination.page, pageSize: pagination.pageSize, workspaceId: workspaceStore.currentWorkspaceId || '' })
    if (res.code === 0) { tableData.value = res.list || []; pagination.total = res.total }
  } finally { loading.value = false }
}

async function loadOrganizations() {
  try {
    const res = await request.post('/organization/list', { page: 1, pageSize: 100 })
    const data = res.data || res
    if (data.code === 0) organizations.value = (data.list || []).filter(org => org.status === 'enable')
  } catch (e) { console.error('Failed to load organizations:', e) }
}

async function loadWorkers() {
  try {
    const res = await getWorkerList()
    const data = res.data || res
    if (data.code === 0) workers.value = (data.list || []).filter(w => w.status === 'running')
  } catch (e) { console.error('Failed to load workers:', e) }
}

function getStatusType(status) {
  const map = { CREATED: 'info', PENDING: 'warning', STARTED: 'primary', PAUSED: 'warning', SUCCESS: 'success', FAILURE: 'danger', STOPPED: 'info' }
  return map[status] || 'info'
}

function resetForm() {
  Object.assign(form, {
    id: '', name: '', target: '', workspaceId: '', orgId: '', isCron: false, cronRule: '', workers: [],
    batchSize: 50, portscanEnable: true, portscanTool: 'naabu', portscanRate: 1000, ports: 'top100',
    portThreshold: 100, scanType: 'c', portscanTimeout: 60, skipHostDiscovery: false, portidentifyEnable: false, portidentifyTimeout: 30,
    portidentifyArgs: '', fingerprintEnable: true, fingerprintHttpx: true, fingerprintIconHash: true,
    fingerprintWappalyzer: true, fingerprintCustomEngine: false, fingerprintScreenshot: false,
    fingerprintTimeout: 30, fingerprintConcurrency: 10, pocscanEnable: false, pocscanAutoScan: true,
    pocscanAutomaticScan: true, pocscanCustomOnly: false, pocscanSeverity: ['critical', 'high', 'medium'],
    pocscanTargetTimeout: 600
  })
}

async function showCreateDialog() {
  loadWorkers()
  isEdit.value = false
  resetForm()
  // 加载用户上次保存的扫描配置
  try {
    const res = await getScanConfig()
    if (res.code === 0 && res.config) {
      const config = JSON.parse(res.config)
      applyConfig(config)
    }
  } catch (e) { console.error('加载扫描配置失败:', e) }
  let wsId = workspaceStore.currentWorkspaceId
  if (wsId === 'all' || !wsId) {
    const defaultWs = workspaceStore.workspaces.find(ws => ws.name === '默认工作空间')
    wsId = defaultWs ? defaultWs.id : (workspaceStore.workspaces.length > 0 ? workspaceStore.workspaces[0].id : '')
  }
  form.workspaceId = wsId
  activeTab.value = 'basic'
  dialogVisible.value = true
}

// 应用配置到表单
function applyConfig(config) {
  Object.assign(form, {
    batchSize: config.batchSize || 50,
    portscanEnable: config.portscan?.enable ?? true,
    portscanTool: config.portscan?.tool || 'naabu',
    portscanRate: config.portscan?.rate || 1000,
    ports: config.portscan?.ports || 'top100',
    portThreshold: config.portscan?.portThreshold || 100,
    scanType: config.portscan?.scanType || 'c',
    portscanTimeout: config.portscan?.timeout || 60,
    skipHostDiscovery: config.portscan?.skipHostDiscovery ?? false,
    portidentifyEnable: config.portidentify?.enable ?? false,
    portidentifyTimeout: config.portidentify?.timeout || 30,
    portidentifyArgs: config.portidentify?.args || '',
    fingerprintEnable: config.fingerprint?.enable ?? true,
    fingerprintHttpx: config.fingerprint?.httpx ?? true,
    fingerprintIconHash: config.fingerprint?.iconHash ?? true,
    fingerprintWappalyzer: config.fingerprint?.wappalyzer ?? true,
    fingerprintCustomEngine: config.fingerprint?.customEngine ?? false,
    fingerprintScreenshot: config.fingerprint?.screenshot ?? false,
    fingerprintTimeout: config.fingerprint?.targetTimeout || 30,
    fingerprintConcurrency: config.fingerprint?.concurrency || 10,
    pocscanEnable: config.pocscan?.enable ?? false,
    pocscanAutoScan: config.pocscan?.autoScan ?? true,
    pocscanAutomaticScan: config.pocscan?.automaticScan ?? true,
    pocscanCustomOnly: config.pocscan?.customPocOnly ?? false,
    pocscanSeverity: config.pocscan?.severity ? config.pocscan.severity.split(',') : ['critical', 'high', 'medium'],
    pocscanTargetTimeout: config.pocscan?.targetTimeout || 600
  })
}

function showDetail(row) { currentTask.value = row; detailVisible.value = true }

function handleEdit(row) {
  loadWorkers()
  isEdit.value = true
  resetForm()
  Object.assign(form, { id: row.id, name: row.name, target: row.target, workspaceId: row.workspaceId || '' })
  // 解析已保存的配置
  if (row.config) {
    try {
      const config = JSON.parse(row.config)
      applyConfig(config)
    } catch (e) { console.error('Parse config error:', e) }
  }
  activeTab.value = 'basic'
  dialogVisible.value = true
}

function buildConfig() {
  return {
    batchSize: form.batchSize,
    portscan: { enable: form.portscanEnable, tool: form.portscanTool, rate: form.portscanRate, ports: form.ports, portThreshold: form.portThreshold, scanType: form.scanType, timeout: form.portscanTimeout, skipHostDiscovery: form.skipHostDiscovery },
    portidentify: { enable: form.portidentifyEnable, timeout: form.portidentifyTimeout, args: form.portidentifyArgs },
    fingerprint: { enable: form.fingerprintEnable, httpx: form.fingerprintHttpx, iconHash: form.fingerprintIconHash, wappalyzer: form.fingerprintWappalyzer, customEngine: form.fingerprintCustomEngine, screenshot: form.fingerprintScreenshot, targetTimeout: form.fingerprintTimeout, concurrency: form.fingerprintConcurrency },
    pocscan: { enable: form.pocscanEnable, useNuclei: true, autoScan: form.pocscanAutoScan, automaticScan: form.pocscanAutomaticScan, customPocOnly: form.pocscanCustomOnly, severity: form.pocscanSeverity.join(','), targetTimeout: form.pocscanTargetTimeout }
  }
}

// 扫描配置字段列表（用于监听变化自动保存）
const scanConfigFields = [
  'batchSize', 'portscanEnable', 'portscanTool', 'portscanRate', 'ports', 'portThreshold', 'scanType', 'portscanTimeout', 'skipHostDiscovery',
  'portidentifyEnable', 'portidentifyTimeout', 'portidentifyArgs',
  'fingerprintEnable', 'fingerprintHttpx', 'fingerprintIconHash', 'fingerprintWappalyzer', 'fingerprintCustomEngine', 'fingerprintScreenshot', 'fingerprintTimeout', 'fingerprintConcurrency',
  'pocscanEnable', 'pocscanAutoScan', 'pocscanAutomaticScan', 'pocscanCustomOnly', 'pocscanSeverity', 'pocscanTargetTimeout'
]

// 防抖保存配置
let saveConfigTimer = null
function debounceSaveConfig() {
  if (saveConfigTimer) clearTimeout(saveConfigTimer)
  saveConfigTimer = setTimeout(() => {
    const config = buildConfig()
    saveScanConfig({ config: JSON.stringify(config) }).catch(e => console.error('自动保存配置失败:', e))
  }, 500)
}

// 监听扫描配置变化，自动保存（仅在新建任务对话框打开且非编辑模式时）
watch(
  () => scanConfigFields.map(f => form[f]),
  () => {
    if (dialogVisible.value && !isEdit.value) {
      debounceSaveConfig()
    }
  },
  { deep: true }
)

async function handleSubmit() {
  await formRef.value.validate()
  submitting.value = true
  try {
    const config = buildConfig()
    const configStr = JSON.stringify(config)
    const data = { name: form.name, target: form.target, workspaceId: form.workspaceId, orgId: form.orgId, isCron: form.isCron, cronRule: form.cronRule, workers: form.workers, config: configStr }
    let res
    if (isEdit.value) {
      res = await updateTask({ id: form.id, ...data })
    } else {
      res = await createTask(data)
    }
    if (res.code === 0) {
      ElMessage.success(isEdit.value ? '任务更新成功' : '任务创建成功')
      dialogVisible.value = false
      loadData()
    } else { ElMessage.error(res.msg) }
  } finally { submitting.value = false }
}

async function handleDelete(row) {
  await ElMessageBox.confirm('确定删除该任务吗？', '提示', { type: 'warning' })
  const res = await deleteTask({ id: row.id })
  res.code === 0 ? (ElMessage.success('删除成功'), loadData()) : ElMessage.error(res.msg)
}

function handleSelectionChange(rows) { selectedRows.value = rows }

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return
  await ElMessageBox.confirm(`确定删除选中的 ${selectedRows.value.length} 条任务吗？`, '提示', { type: 'warning' })
  const res = await batchDeleteTask({ ids: selectedRows.value.map(row => row.id) })
  res.code === 0 ? (ElMessage.success('删除成功'), selectedRows.value = [], loadData()) : ElMessage.error(res.msg)
}

async function handleRetry(row) {
  await ElMessageBox.confirm('确定重新执行该任务吗？', '提示', { type: 'warning' })
  const res = await retryTask({ id: row.id })
  res.code === 0 ? (ElMessage.success('任务已重新执行'), loadData()) : ElMessage.error(res.msg)
}

async function handleStart(row) {
  const res = await startTask({ id: row.id })
  res.code === 0 ? (ElMessage.success('任务已启动'), loadData()) : ElMessage.error(res.msg)
}

async function handlePause(row) {
  await ElMessageBox.confirm('确定暂停该任务吗？', '提示', { type: 'warning' })
  const res = await pauseTask({ id: row.id })
  res.code === 0 ? (ElMessage.success('任务已暂停'), loadData()) : ElMessage.error(res.msg)
}

async function handleResume(row) {
  const res = await resumeTask({ id: row.id })
  res.code === 0 ? (ElMessage.success('任务已继续'), loadData()) : ElMessage.error(res.msg)
}

async function handleStop(row) {
  await ElMessageBox.confirm('确定停止该任务吗？', '提示', { type: 'warning' })
  const res = await stopTask({ id: row.id })
  res.code === 0 ? (ElMessage.success('任务已停止'), loadData()) : ElMessage.error(res.msg)
}

function viewReport(row) { router.push({ path: '/report', query: { taskId: row.id } }) }

async function showLogs(row) {
  currentLogTaskId.value = row.taskId
  currentLogTask.value = { ...row }
  taskLogs.value = []
  logIdSet.clear()
  logDialogVisible.value = true
  await refreshLogs()
  if (logAutoRefresh.value) { connectLogStream(); startLogPolling() }
}

async function refreshLogs() {
  if (!currentLogTaskId.value) return
  try {
    const task = tableData.value.find(t => t.id === currentLogTask.value?.id)
    if (task) currentLogTask.value = { ...task }
    const res = await getTaskLogs({ taskId: currentLogTaskId.value, limit: 500 })
    if (res.code === 0) {
      for (const log of (res.list || [])) {
        const logId = (log.timestamp || '') + (log.message || '')
        if (!logIdSet.has(logId)) { logIdSet.add(logId); taskLogs.value.push(parseLogMessage(log)) }
      }
      taskLogs.value.sort((a, b) => (a.timestamp || '').localeCompare(b.timestamp || ''))
      scrollToBottom()
    }
  } catch (err) { console.error('Failed to load task logs:', err) }
}

function startLogPolling() {
  if (logPollingTimer || !logAutoRefresh.value) return
  logPollingTimer = setInterval(async () => {
    if (logDialogVisible.value && currentLogTaskId.value && logAutoRefresh.value) { await loadData(); await refreshLogs() }
  }, 2000)
}

function handleLogAutoRefreshChange(val) {
  if (val) { startLogPolling(); connectLogStream() }
  else { stopLogPolling(); if (logEventSource) { logEventSource.close(); logEventSource = null } }
}

function scrollToBottom() {
  setTimeout(() => { if (logContainerRef.value) logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight }, 100)
}

function connectLogStream() {
  if (logEventSource) { logEventSource.close(); logEventSource = null }
  if (!currentLogTaskId.value) return
  const token = localStorage.getItem('token')
  const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
  logEventSource = new EventSource(`${baseUrl}/api/v1/task/logs/stream?taskId=${currentLogTaskId.value}&token=${token}`)
  logEventSource.onmessage = (event) => {
    try {
      const log = JSON.parse(event.data)
      const logId = (log.timestamp || '') + (log.message || '')
      if (!logIdSet.has(logId)) { logIdSet.add(logId); taskLogs.value.push(parseLogMessage(log)); scrollToBottom() }
    } catch (err) { console.error('Failed to parse log:', err) }
  }
  logEventSource.onerror = () => {}
}

function stopLogPolling() { if (logPollingTimer) { clearInterval(logPollingTimer); logPollingTimer = null } }

function closeLogDialog() {
  logDialogVisible.value = false
  currentLogTaskId.value = ''
  currentLogTask.value = null
  taskLogs.value = []
  logIdSet.clear()
  logWorkerFilter.value = ''
  logLevelFilter.value = ''
  if (logEventSource) { logEventSource.close(); logEventSource = null }
  stopLogPolling()
}
</script>

<style lang="scss" scoped>
.task-page {
  .action-card { margin-bottom: 20px; }
  .pagination { margin-top: 20px; justify-content: flex-end; }
  .form-hint { margin-left: 10px; color: var(--el-text-color-secondary); font-size: 12px; }
  .sub-task-info { font-size: 11px; color: var(--el-text-color-secondary); margin-top: 2px; }
  .tool-tip { color: #f56c6c; font-size: 12px; }
}

.task-dialog {
  :deep(.el-dialog__body) { padding: 10px 20px 0; }
}

.task-tabs {
  :deep(.el-tabs__header) { margin-bottom: 15px; }
  :deep(.el-tabs__item) { font-size: 14px; }
}

.tab-form {
  min-height: 320px;
  padding: 10px 0;
}

.dialog-footer {
  padding-top: 10px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.log-progress {
  margin-bottom: 15px;
  padding: 12px 15px;
  background-color: var(--el-fill-color-light);
  border-radius: 6px;
  .progress-info {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 10px;
    .task-name { font-weight: 500; font-size: 14px; }
  }
}

.log-filter {
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  .log-stats { margin-left: auto; color: var(--el-text-color-secondary); font-size: 12px; }
}

.log-container {
  max-height: 450px;
  overflow-y: auto;
  background-color: #1e1e1e;
  border-radius: 4px;
  padding: 10px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.6;
}

.log-empty { color: var(--el-text-color-secondary); text-align: center; padding: 20px; }
.log-entry { padding: 2px 0; white-space: pre-wrap; word-break: break-all; }
.log-time { color: #6a9955; margin-right: 8px; font-size: 11px; }
.log-level { font-weight: bold; margin-right: 6px; min-width: 45px; display: inline-block; font-size: 11px; }
.log-worker { color: #569cd6; margin-right: 6px; font-size: 11px; }
.log-message { color: #d4d4d4; }
.log-info .log-level { color: #4fc3f7; }
.log-warn .log-level, .log-warning .log-level { color: #ffb74d; }
.log-error .log-level { color: #ef5350; }
</style>
