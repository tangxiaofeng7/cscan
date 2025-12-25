<template>
  <div class="task-page">
    <!-- 操作栏 -->
    <el-card class="action-card">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>新建任务
      </el-button>
      <el-button @click="showProfileDialog">
        <el-icon><Setting /></el-icon>配置管理
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
        <el-table-column prop="profileName" label="任务配置" width="120" />
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
            <!-- 启动按钮：仅CREATED状态显示 -->
            <el-button v-if="row.status === 'CREATED'" type="success" link size="small" @click="handleStart(row)">启动</el-button>
            <!-- 编辑按钮：仅CREATED状态显示 -->
            <el-button v-if="row.status === 'CREATED'" type="warning" link size="small" @click="handleEdit(row)">编辑</el-button>
            <!-- 暂停按钮：仅STARTED状态显示 -->
            <el-button v-if="row.status === 'STARTED'" type="warning" link size="small" @click="handlePause(row)">暂停</el-button>
            <!-- 继续按钮：仅PAUSED状态显示 -->
            <el-button v-if="row.status === 'PAUSED'" type="success" link size="small" @click="handleResume(row)">继续</el-button>
            <!-- 停止按钮：STARTED/PAUSED/PENDING状态显示 -->
            <el-button v-if="['STARTED', 'PAUSED', 'PENDING'].includes(row.status)" type="danger" link size="small" @click="handleStop(row)">停止</el-button>
            <el-button type="primary" link size="small" @click="showDetail(row)">详情</el-button>
            <el-button type="info" link size="small" @click="showLogs(row)">日志</el-button>
            <el-button type="info" link size="small" @click="viewReport(row)">报告</el-button>
            <el-button 
              v-if="['SUCCESS', 'FAILURE', 'STOPPED'].includes(row.status)" 
              type="warning" 
              link 
              size="small" 
              @click="handleRetry(row)"
            >重新执行</el-button>
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
        <el-descriptions-item label="任务配置">{{ currentTask.profileName }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentTask.status)">{{ currentTask.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="进度">
          <div>
            <el-progress :percentage="currentTask.progress" :stroke-width="10" style="width: 150px" />
            <div v-if="currentTask.subTaskCount > 1" class="sub-task-info">
              子任务: {{ currentTask.subTaskDone }}/{{ currentTask.subTaskCount }}
            </div>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ currentTask.createTime }}</el-descriptions-item>
        <el-descriptions-item label="定时任务">
          <span v-if="currentTask.isCron">{{ currentTask.cronRule }}</span>
          <span v-else>否</span>
        </el-descriptions-item>
        <el-descriptions-item label="扫描目标" :span="2">
          <div style="max-height: 100px; overflow-y: auto; white-space: pre-wrap">{{ currentTask.target }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="执行结果" :span="2">
          <div style="max-height: 100px; overflow-y: auto">{{ currentTask.result || '-' }}</div>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 配置管理对话框 -->
    <el-dialog v-model="profileDialogVisible" title="任务配置管理" width="800px">
      <div style="margin-bottom: 15px">
        <el-button type="primary" size="small" @click="showProfileForm()">新建配置</el-button>
      </div>
      <el-table :data="profiles" stripe size="small" max-height="300">
        <el-table-column prop="name" label="配置名称" width="120" />
        <el-table-column prop="description" label="描述" min-width="150" />
        <el-table-column label="端口扫描" width="100">
          <template #default="{ row }">
            <el-tag v-if="parseConfig(row.config).portscan?.enable" type="success" size="small">启用</el-tag>
            <el-tag v-else type="info" size="small">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="指纹识别" width="100">
          <template #default="{ row }">
            <el-tag v-if="parseConfig(row.config).fingerprint?.enable" type="success" size="small">启用</el-tag>
            <el-tag v-else type="info" size="small">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="漏洞扫描" width="100">
          <template #default="{ row }">
            <el-tag v-if="parseConfig(row.config).pocscan?.enable" type="success" size="small">启用</el-tag>
            <el-tag v-else type="info" size="small">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showProfileForm(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteProfile(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 配置编辑对话框 -->
    <el-dialog v-model="profileFormVisible" :title="profileForm.id ? '编辑配置' : '新建配置'" width="600px">
      <el-form ref="profileFormRef" :model="profileForm" :rules="profileRules" label-width="100px">
        <el-form-item label="配置名称" prop="name">
          <el-input v-model="profileForm.name" placeholder="请输入配置名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="profileForm.description" placeholder="请输入描述" />
        </el-form-item>
        <el-divider content-position="left">任务分发</el-divider>
        <el-form-item label="任务拆分">
          <el-input-number v-model="profileForm.batchSize" :min="10" :max="1000" :step="10" />
          <span class="form-hint">每批目标数量，多Worker时自动拆分并行执行 (0=不拆分)</span>
        </el-form-item>
        <el-divider content-position="left">扫描选项</el-divider>
        <el-form-item label="端口扫描">
          <el-switch v-model="profileForm.portscanEnable" />
        </el-form-item>
        <el-form-item v-if="profileForm.portscanEnable" label="扫描工具">
          <el-radio-group v-model="profileForm.portscanTool">
            <el-radio label="naabu">Naabu (推荐)</el-radio>
            <el-radio label="masscan">Masscan</el-radio>
          </el-radio-group>
          <span class="form-hint">不支持域名</span>
        </el-form-item>
        <el-form-item v-if="profileForm.portscanEnable" label="端口范围">
          <el-select v-model="profileForm.ports" filterable allow-create default-first-option placeholder="选择或输入端口" style="width: 100%">
            <el-option label="top100 - 常用100端口" value="top100" />
            <el-option label="top1000 - 常用1000端口(包含top100)" value="top1000" />
            <el-option label="80,443,8080,8443 - Web常用" value="80,443,8080,8443" />
            <el-option label="3306,5432,1433,1521,27017,6379 - 数据库" value="3306,5432,1433,1521,27017,6379" />
            <el-option label="1-65535 - 全端口" value="1-65535" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="profileForm.portscanEnable" label="扫描速率">
          <el-input-number v-model="profileForm.portscanRate" :min="100" :max="100000" :step="100" />
          <span class="form-hint">包/秒</span>
        </el-form-item>
        <el-form-item v-if="profileForm.portscanEnable" label="端口阈值">
          <el-input-number v-model="profileForm.portThreshold" :min="0" :max="65535" :step="10" />
          <span class="form-hint">开放端口超过此数量的主机将被过滤 (0=不过滤)</span>
        </el-form-item>
        <el-form-item label="指纹识别">
          <el-switch v-model="profileForm.fingerprintEnable" />
        </el-form-item>
        <el-form-item v-if="profileForm.fingerprintEnable" label="探测工具">
          <el-checkbox v-model="profileForm.fingerprintHttpx">Httpx探测 (推荐)</el-checkbox>
        </el-form-item>
        <el-form-item v-if="profileForm.fingerprintEnable" label="附加功能">
          <el-checkbox v-model="profileForm.fingerprintIconHash">Icon Hash</el-checkbox>
          <el-checkbox v-model="profileForm.fingerprintWappalyzer">Wappalyzer指纹</el-checkbox>
          <el-checkbox v-model="profileForm.fingerprintCustomEngine">自定义指纹引擎</el-checkbox>
          <el-checkbox v-model="profileForm.fingerprintScreenshot">网页截图</el-checkbox>
          <span v-if="profileForm.fingerprintScreenshot" class="form-hint" style="margin-left: 5px">
            ({{ profileForm.fingerprintHttpx ? 'httpx截图' : 'chromedp截图，较慢' }})
          </span>
        </el-form-item>
        <el-form-item v-if="profileForm.fingerprintEnable && profileForm.fingerprintCustomEngine" label="">
          <el-alert type="info" :closable="false" show-icon style="padding: 5px 10px">
            <template #title>
              <span style="font-size: 12px">自定义指纹引擎将使用指纹管理中自定义的指纹进行识别</span>
            </template>
          </el-alert>
        </el-form-item>
        <el-form-item label="漏洞扫描">
          <el-switch v-model="profileForm.pocscanEnable" />
          <span class="form-hint">使用 Nuclei 引擎</span>
        </el-form-item>
        <el-form-item v-if="profileForm.pocscanEnable" label="自动扫描">
          <div style="display: block; width: 100%">
            <div>
              <el-checkbox v-model="profileForm.pocscanAutoScan" :disabled="profileForm.pocscanCustomOnly">自定义标签映射</el-checkbox>
              <span class="form-hint" style="margin-left: 5px">(POC管理中配置)</span>
            </div>
            <div style="margin-top: 5px">
              <el-checkbox v-model="profileForm.pocscanAutomaticScan" :disabled="profileForm.pocscanCustomOnly">Wappalyzer自动扫描</el-checkbox>
              <span class="form-hint" style="margin-left: 5px">(内置技术栈映射)</span>
            </div>
          </div>
        </el-form-item>
        <el-form-item v-if="profileForm.pocscanEnable" label="自定义POC">
          <el-checkbox v-model="profileForm.pocscanCustomOnly">只使用自定义POC</el-checkbox>
          <span class="form-hint" style="margin-left: 5px">(仅扫描POC管理中启用的自定义POC)</span>
        </el-form-item>
        <el-form-item v-if="profileForm.pocscanEnable" label="严重级别">
          <el-checkbox-group v-model="profileForm.pocscanSeverity">
            <el-checkbox label="critical">Critical</el-checkbox>
            <el-checkbox label="high">High</el-checkbox>
            <el-checkbox label="medium">Medium</el-checkbox>
            <el-checkbox label="low">Low</el-checkbox>
            <el-checkbox label="info">Info</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="profileFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveProfile">保存</el-button>
      </template>
    </el-dialog>

    <!-- 新建任务对话框 -->
    <el-dialog v-model="dialogVisible" title="新建任务" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="工作空间">
          <el-select v-model="form.workspaceId" placeholder="选择工作空间（可选）" clearable style="width: 100%">
            <el-option
              v-for="ws in workspaceStore.workspaces"
              :key="ws.id"
              :label="ws.name"
              :value="ws.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="扫描目标" prop="target">
          <el-input
            v-model="form.target"
            type="textarea"
            :rows="5"
            placeholder="每行一个目标，支持IP、IP段、域名"
          />
        </el-form-item>
        <el-form-item label="任务配置" prop="profileId">
          <el-select v-model="form.profileId" placeholder="请选择任务配置" style="width: 100%">
            <el-option
              v-for="p in profiles"
              :key="p.id"
              :label="p.name"
              :value="p.id"
            >
              <span>{{ p.name }}</span>
              <span class="option-desc">{{ p.description }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="定时任务">
          <el-switch v-model="form.isCron" />
        </el-form-item>
        <el-form-item v-if="form.isCron" label="Cron表达式" prop="cronRule">
          <el-input v-model="form.cronRule" placeholder="如: 0 0 * * * (每天0点执行)" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑任务对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑任务" width="600px">
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="扫描目标" prop="target">
          <el-input
            v-model="editForm.target"
            type="textarea"
            :rows="5"
            placeholder="每行一个目标，支持IP、IP段、域名"
          />
        </el-form-item>
        <el-form-item label="任务配置" prop="profileId">
          <el-select v-model="editForm.profileId" placeholder="请选择任务配置" style="width: 100%">
            <el-option
              v-for="p in profiles"
              :key="p.id"
              :label="p.name"
              :value="p.id"
            >
              <span>{{ p.name }}</span>
              <span class="option-desc">{{ p.description }}</span>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleUpdateTask">保存</el-button>
      </template>
    </el-dialog>

    <!-- 任务日志对话框 -->
    <el-dialog v-model="logDialogVisible" title="任务日志" width="1000px" @close="closeLogDialog">
      <div class="log-filter">
        <el-select v-model="logWorkerFilter" placeholder="筛选Worker" clearable size="small" style="width: 150px">
          <el-option label="全部Worker" value="" />
          <el-option v-for="w in logWorkers" :key="w" :label="w" :value="w" />
        </el-select>
        <el-select v-model="logSubTaskFilter" placeholder="筛选子任务" clearable size="small" style="width: 150px; margin-left: 10px">
          <el-option label="全部子任务" value="" />
          <el-option v-for="s in logSubTasks" :key="s" :label="s === 'main' ? '主任务' : `子任务 ${s}`" :value="s" />
        </el-select>
        <el-select v-model="logLevelFilter" placeholder="筛选级别" clearable size="small" style="width: 120px; margin-left: 10px">
          <el-option label="全部级别" value="" />
          <el-option label="INFO" value="INFO" />
          <el-option label="WARN" value="WARN" />
          <el-option label="ERROR" value="ERROR" />
        </el-select>
        <span class="log-stats">共 {{ filteredLogs.length }} 条日志</span>
      </div>
      <div class="log-container" ref="logContainerRef">
        <div v-if="filteredLogs.length === 0" class="log-empty">暂无日志</div>
        <div v-for="(log, index) in filteredLogs" :key="index" class="log-entry" :class="'log-' + log.level.toLowerCase()">
          <span class="log-time">{{ formatLogTime(log.timestamp) }}</span>
          <span class="log-level">[{{ log.level }}]</span>
          <span class="log-worker">{{ log.workerName }}</span>
          <span v-if="log.subTask" class="log-subtask">[{{ log.subTask === 'main' ? '主' : log.subTask }}]</span>
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
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Setting, Delete } from '@element-plus/icons-vue'
import { getTaskList, createTask, deleteTask, batchDeleteTask, getTaskProfileList, saveTaskProfile, deleteTaskProfile, retryTask, startTask, pauseTask, resumeTask, stopTask, updateTask, getTaskLogs } from '@/api/task'
import { useWorkspaceStore } from '@/stores/workspace'

const router = useRouter()
const workspaceStore = useWorkspaceStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const detailVisible = ref(false)
const profileDialogVisible = ref(false)
const profileFormVisible = ref(false)
const editDialogVisible = ref(false)
const logDialogVisible = ref(false)
const tableData = ref([])
const profiles = ref([])
const formRef = ref()
const profileFormRef = ref()
const editFormRef = ref()
const logContainerRef = ref()
const currentTask = ref({})
const selectedRows = ref([])
const autoRefresh = ref(true)
const taskLogs = ref([])
const currentLogTaskId = ref('')
const logIdSet = new Set() // 用于日志去重
const logWorkerFilter = ref('')
const logSubTaskFilter = ref('')
const logLevelFilter = ref('')
let refreshTimer = null
let logEventSource = null
let logPollingTimer = null // 日志轮询定时器

// 从日志中提取Worker列表
const logWorkers = computed(() => {
  const workers = new Set()
  taskLogs.value.forEach(log => {
    if (log.workerName) workers.add(log.workerName)
  })
  return Array.from(workers).sort()
})

// 从日志中提取子任务列表
const logSubTasks = computed(() => {
  const subTasks = new Set()
  taskLogs.value.forEach(log => {
    if (log.subTask) subTasks.add(log.subTask)
  })
  return Array.from(subTasks).sort((a, b) => {
    if (a === 'main') return -1
    if (b === 'main') return 1
    return parseInt(a) - parseInt(b)
  })
})

// 筛选后的日志
const filteredLogs = computed(() => {
  return taskLogs.value.filter(log => {
    if (logWorkerFilter.value && log.workerName !== logWorkerFilter.value) return false
    if (logSubTaskFilter.value && log.subTask !== logSubTaskFilter.value) return false
    if (logLevelFilter.value && log.level !== logLevelFilter.value) return false
    return true
  })
})

// 格式化日志时间（只显示时分秒）
function formatLogTime(timestamp) {
  if (!timestamp) return ''
  // 如果是完整时间格式，只取时分秒部分
  const match = timestamp.match(/(\d{2}:\d{2}:\d{2})/)
  return match ? match[1] : timestamp
}

// 解析日志消息，提取子任务信息并简化显示
function parseLogMessage(log) {
  let message = log.message || ''
  let subTask = 'main'
  
  // 提取 [Sub-X] 标记
  const subMatch = message.match(/^\[Sub-(\d+)\]\s*/)
  if (subMatch) {
    subTask = subMatch[1]
    message = message.replace(subMatch[0], '')
  }
  
  return {
    ...log,
    subTask,
    displayMessage: message
  }
}

const editForm = reactive({
  id: '',
  name: '',
  target: '',
  profileId: ''
})

const editRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  target: [{ required: true, message: '请输入扫描目标', trigger: 'blur' }],
  profileId: [{ required: true, message: '请选择任务配置', trigger: 'change' }]
}

const profileForm = reactive({
  id: '',
  name: '',
  description: '',
  batchSize: 50,
  portscanEnable: true,
  portscanTool: 'naabu',
  portscanRate: 1000,
  ports: '80,443,8080',
  portThreshold: 100,
  fingerprintEnable: true,
  fingerprintHttpx: true,
  fingerprintIconHash: true,
  fingerprintWappalyzer: true,
  fingerprintCustomEngine: false,
  fingerprintScreenshot: false,
  pocscanEnable: false,
  pocscanAutoScan: true,
  pocscanAutomaticScan: true,
  pocscanCustomOnly: false,
  pocscanSeverity: ['critical', 'high', 'medium']
})

const profileRules = {
  name: [{ required: true, message: '请输入配置名称', trigger: 'blur' }]
}

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const form = reactive({
  name: '',
  target: '',
  profileId: '',
  workspaceId: '',
  isCron: false,
  cronRule: ''
})

const rules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  target: [{ required: true, message: '请输入扫描目标', trigger: 'blur' }],
  profileId: [{ required: true, message: '请选择任务配置', trigger: 'change' }]
}

// 监听工作空间切换
function handleWorkspaceChanged() {
  pagination.page = 1
  loadData()
}

onMounted(() => {
  loadData()
  loadProfiles()
  // 如果默认开启自动刷新，启动定时器
  if (autoRefresh.value) {
    startAutoRefresh()
  }
  // 监听工作空间切换事件
  window.addEventListener('workspace-changed', handleWorkspaceChanged)
})

onUnmounted(() => {
  stopAutoRefresh()
  window.removeEventListener('workspace-changed', handleWorkspaceChanged)
  // 关闭SSE连接
  if (logEventSource) {
    logEventSource.close()
    logEventSource = null
  }
})

function handleAutoRefreshChange(val) {
  if (val) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

function startAutoRefresh() {
  stopAutoRefresh()
  refreshTimer = setInterval(() => {
    loadData()
  }, 30000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await getTaskList({
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

async function loadProfiles() {
  const res = await getTaskProfileList()
  if (res.code === 0) {
    profiles.value = res.list || []
  }
}

function getStatusType(status) {
  const map = {
    CREATED: 'info',
    PENDING: 'warning',
    STARTED: 'primary',
    PAUSED: 'warning',
    SUCCESS: 'success',
    FAILURE: 'danger',
    STOPPED: 'info'
  }
  return map[status] || 'info'
}

function showCreateDialog() {
  Object.assign(form, { 
    name: '', 
    target: '', 
    profileId: '', 
    workspaceId: workspaceStore.currentWorkspaceId || '',
    isCron: false, 
    cronRule: '' 
  })
  dialogVisible.value = true
}

function showDetail(row) {
  currentTask.value = row
  detailVisible.value = true
}

function showProfileDialog() {
  profileDialogVisible.value = true
}

function parseConfig(configStr) {
  try {
    return JSON.parse(configStr || '{}')
  } catch {
    return {}
  }
}

function showProfileForm(row = null) {
  if (row) {
    const config = parseConfig(row.config)
    Object.assign(profileForm, {
      id: row.id,
      name: row.name,
      description: row.description,
      batchSize: config.batchSize || 50,
      portscanEnable: config.portscan?.enable ?? true,
      portscanTool: config.portscan?.tool || 'naabu',
      portscanRate: config.portscan?.rate || 1000,
      ports: config.portscan?.ports || '80,443,8080',
      portThreshold: config.portscan?.portThreshold || 100,
      fingerprintEnable: config.fingerprint?.enable ?? true,
      fingerprintHttpx: config.fingerprint?.httpx ?? true,
      fingerprintIconHash: config.fingerprint?.iconHash ?? true,
      fingerprintWappalyzer: config.fingerprint?.wappalyzer ?? true,
      fingerprintCustomEngine: config.fingerprint?.customEngine ?? false,
      fingerprintScreenshot: config.fingerprint?.screenshot ?? false,
      pocscanEnable: config.pocscan?.enable ?? false,
      pocscanAutoScan: config.pocscan?.autoScan ?? true,
      pocscanAutomaticScan: config.pocscan?.automaticScan ?? true,
      pocscanCustomOnly: config.pocscan?.customPocOnly ?? false,
      pocscanSeverity: config.pocscan?.severity ? config.pocscan.severity.split(',') : ['critical', 'high', 'medium']
    })
  } else {
    Object.assign(profileForm, {
      id: '',
      name: '',
      description: '',
      batchSize: 50,
      portscanEnable: true,
      portscanTool: 'naabu',
      portscanRate: 1000,
      ports: '80,443,8080',
      portThreshold: 100,
      fingerprintEnable: true,
      fingerprintHttpx: true,
      fingerprintIconHash: true,
      fingerprintWappalyzer: true,
      fingerprintCustomEngine: false,
      fingerprintScreenshot: false,
      pocscanEnable: false,
      pocscanAutoScan: true,
      pocscanAutomaticScan: true,
      pocscanCustomOnly: false,
      pocscanSeverity: ['critical', 'high', 'medium']
    })
  }
  profileFormVisible.value = true
}

async function handleSaveProfile() {
  await profileFormRef.value.validate()
  const config = {
    batchSize: profileForm.batchSize,
    portscan: { 
      enable: profileForm.portscanEnable, 
      tool: profileForm.portscanTool,
      rate: profileForm.portscanRate,
      ports: profileForm.ports,
      portThreshold: profileForm.portThreshold
    },
    fingerprint: { 
      enable: profileForm.fingerprintEnable,
      httpx: profileForm.fingerprintHttpx,
      iconHash: profileForm.fingerprintIconHash,
      wappalyzer: profileForm.fingerprintWappalyzer,
      customEngine: profileForm.fingerprintCustomEngine,
      screenshot: profileForm.fingerprintScreenshot
    },
    pocscan: { 
      enable: profileForm.pocscanEnable, 
      useNuclei: true,
      autoScan: profileForm.pocscanAutoScan,
      automaticScan: profileForm.pocscanAutomaticScan,
      customPocOnly: profileForm.pocscanCustomOnly,
      severity: profileForm.pocscanSeverity.join(',')
    }
  }
  const data = {
    id: profileForm.id,
    name: profileForm.name,
    description: profileForm.description,
    config: JSON.stringify(config)
  }
  const res = await saveTaskProfile(data)
  if (res.code === 0) {
    ElMessage.success('保存成功')
    profileFormVisible.value = false
    loadProfiles()
  } else {
    ElMessage.error(res.msg)
  }
}

async function handleDeleteProfile(row) {
  await ElMessageBox.confirm('确定删除该配置吗？', '提示', { type: 'warning' })
  const res = await deleteTaskProfile({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    loadProfiles()
  }
}

async function handleSubmit() {
  await formRef.value.validate()
  submitting.value = true
  try {
    const res = await createTask(form)
    if (res.code === 0) {
      ElMessage.success('任务创建成功')
      dialogVisible.value = false
      loadData()
    } else {
      ElMessage.error(res.msg)
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row) {
  await ElMessageBox.confirm('确定删除该任务吗？', '提示', { type: 'warning' })
  const res = await deleteTask({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    loadData()
  } else {
    ElMessage.error(res.msg)
  }
}

function handleSelectionChange(rows) {
  selectedRows.value = rows
}

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return
  await ElMessageBox.confirm(`确定删除选中的 ${selectedRows.value.length} 条任务吗？`, '提示', { type: 'warning' })
  const ids = selectedRows.value.map(row => row.id)
  const res = await batchDeleteTask({ ids })
  if (res.code === 0) {
    ElMessage.success(`成功删除 ${selectedRows.value.length} 条任务`)
    selectedRows.value = []
    loadData()
  } else {
    ElMessage.error(res.msg)
  }
}

async function handleRetry(row) {
  await ElMessageBox.confirm('确定重新执行该任务吗？', '提示', { type: 'warning' })
  const res = await retryTask({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('任务已重新执行')
    loadData()
  } else {
    ElMessage.error(res.msg)
  }
}

async function handleStart(row) {
  const res = await startTask({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('任务已启动')
    loadData()
  } else {
    ElMessage.error(res.msg)
  }
}

async function handlePause(row) {
  await ElMessageBox.confirm('确定暂停该任务吗？暂停后可继续执行', '提示', { type: 'warning' })
  const res = await pauseTask({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('任务已暂停')
    loadData()
  } else {
    ElMessage.error(res.msg)
  }
}

async function handleResume(row) {
  const res = await resumeTask({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('任务已继续')
    loadData()
  } else {
    ElMessage.error(res.msg)
  }
}

async function handleStop(row) {
  await ElMessageBox.confirm('确定停止该任务吗？停止后无法继续', '提示', { type: 'warning' })
  const res = await stopTask({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('任务已停止')
    loadData()
  } else {
    ElMessage.error(res.msg)
  }
}

function viewReport(row) {
  router.push({ path: '/report', query: { taskId: row.id } })
}

function handleEdit(row) {
  Object.assign(editForm, {
    id: row.id,
    name: row.name,
    target: row.target,
    profileId: row.profileId
  })
  editDialogVisible.value = true
}

async function handleUpdateTask() {
  await editFormRef.value.validate()
  submitting.value = true
  try {
    const res = await updateTask(editForm)
    if (res.code === 0) {
      ElMessage.success('任务更新成功')
      editDialogVisible.value = false
      loadData()
    } else {
      ElMessage.error(res.msg)
    }
  } finally {
    submitting.value = false
  }
}

async function showLogs(row) {
  currentLogTaskId.value = row.taskId // 使用taskId（UUID）而不是id（MongoDB ObjectID）
  taskLogs.value = []
  logIdSet.clear()
  logDialogVisible.value = true
  await refreshLogs()
  // 连接SSE实时日志流
  connectLogStream()
  // 同时启动轮询作为备选（确保日志能更新）
  startLogPolling()
}

async function refreshLogs() {
  if (!currentLogTaskId.value) return
  try {
    const res = await getTaskLogs({ taskId: currentLogTaskId.value, limit: 500 })
    if (res.code === 0) {
      const newLogs = res.list || []
      // 使用去重逻辑
      for (const log of newLogs) {
        const logId = (log.timestamp || '') + (log.message || '')
        if (!logIdSet.has(logId)) {
          logIdSet.add(logId)
          // 解析日志消息，提取子任务信息
          taskLogs.value.push(parseLogMessage(log))
        }
      }
      // 按时间排序
      taskLogs.value.sort((a, b) => (a.timestamp || '').localeCompare(b.timestamp || ''))
      // 滚动到底部
      scrollToBottom()
    }
  } catch (err) {
    console.error('Failed to load task logs:', err)
  }
}

// 启动日志轮询
function startLogPolling() {
  if (logPollingTimer) return
  logPollingTimer = setInterval(async () => {
    if (logDialogVisible.value && currentLogTaskId.value) {
      await refreshLogs()
    }
  }, 2000) // 每2秒轮询一次
}

function scrollToBottom() {
  setTimeout(() => {
    if (logContainerRef.value) {
      logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
    }
  }, 100)
}

function connectLogStream() {
  // 关闭已有连接
  if (logEventSource) {
    logEventSource.close()
    logEventSource = null
  }
  
  if (!currentLogTaskId.value) return
  
  // 获取token用于认证
  const token = localStorage.getItem('token')
  const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
  const url = `${baseUrl}/api/v1/task/logs/stream?taskId=${currentLogTaskId.value}&token=${token}`
  
  logEventSource = new EventSource(url)
  
  logEventSource.onmessage = (event) => {
    try {
      const log = JSON.parse(event.data)
      // 使用去重逻辑
      const logId = (log.timestamp || '') + (log.message || '')
      if (!logIdSet.has(logId)) {
        logIdSet.add(logId)
        // 解析日志消息，提取子任务信息
        taskLogs.value.push(parseLogMessage(log))
        scrollToBottom()
      }
    } catch (err) {
      console.error('Failed to parse log:', err)
    }
  }
  
  logEventSource.onerror = (err) => {
    console.error('SSE connection error:', err)
    // SSE 断开后依赖轮询继续工作
  }
}

// 停止日志轮询
function stopLogPolling() {
  if (logPollingTimer) {
    clearInterval(logPollingTimer)
    logPollingTimer = null
  }
}

function closeLogDialog() {
  logDialogVisible.value = false
  currentLogTaskId.value = ''
  taskLogs.value = []
  logIdSet.clear()
  // 重置筛选条件
  logWorkerFilter.value = ''
  logSubTaskFilter.value = ''
  logLevelFilter.value = ''
  // 关闭SSE连接
  if (logEventSource) {
    logEventSource.close()
    logEventSource = null
  }
  // 停止轮询
  stopLogPolling()
}
</script>

<style lang="scss" scoped>
.task-page {
  .action-card {
    margin-bottom: 20px;
  }

  .pagination {
    margin-top: 20px;
    justify-content: flex-end;
  }
  
  .form-hint {
    margin-left: 10px;
    color: var(--text-muted);
    font-size: 12px;
  }
  
  .option-desc {
    color: var(--text-muted);
    font-size: 12px;
    margin-left: 10px;
  }
  
  .sub-task-info {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 2px;
  }
}

.log-filter {
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  
  .log-stats {
    margin-left: auto;
    color: var(--text-muted);
    font-size: 12px;
  }
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

.log-empty {
  color: var(--text-muted);
  text-align: center;
  padding: 20px;
}

.log-entry {
  padding: 2px 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.log-time {
  color: #6a9955;
  margin-right: 8px;
  font-size: 11px;
}

.log-level {
  font-weight: bold;
  margin-right: 6px;
  min-width: 45px;
  display: inline-block;
  font-size: 11px;
}

.log-worker {
  color: #569cd6;
  margin-right: 6px;
  font-size: 11px;
}

.log-subtask {
  color: #ce9178;
  margin-right: 6px;
  font-size: 11px;
}

.log-message {
  color: #d4d4d4;
}

.log-info .log-level {
  color: #4fc3f7;
}

.log-warn .log-level,
.log-warning .log-level {
  color: #ffb74d;
}

.log-error .log-level {
  color: #ef5350;
}

.log-debug .log-level {
  color: #9e9e9e;
}
</style>
