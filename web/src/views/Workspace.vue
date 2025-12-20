<template>
  <div class="workspace-page">
    <el-card class="action-card">
      <el-button type="primary" @click="showDialog()">
        <el-icon><Plus /></el-icon>新建工作空间
      </el-button>
    </el-card>

    <el-card>
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="250" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'enable' ? 'success' : 'danger'">
              {{ row.status === 'enable' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="160" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showDialog(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑工作空间' : '新建工作空间'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const tableData = ref([])
const formRef = ref()

const form = reactive({ id: '', name: '', description: '' })
const rules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const res = await request.post('/workspace/list', { page: 1, pageSize: 100 })
    if (res.code === 0) tableData.value = res.list || []
  } finally {
    loading.value = false
  }
}

function showDialog(row = null) {
  if (row) {
    Object.assign(form, { id: row.id, name: row.name, description: row.description })
  } else {
    Object.assign(form, { id: '', name: '', description: '' })
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value.validate()
  submitting.value = true
  try {
    const res = await request.post('/workspace/save', form)
    if (res.code === 0) {
      ElMessage.success(form.id ? '更新成功' : '创建成功')
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
  await ElMessageBox.confirm('确定删除该工作空间吗？', '提示', { type: 'warning' })
  const res = await request.post('/workspace/delete', { id: row.id })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    loadData()
  }
}
</script>

<style lang="scss" scoped>
.workspace-page {
  .action-card { margin-bottom: 20px; }
}
</style>
