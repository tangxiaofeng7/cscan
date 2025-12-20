<template>
  <div class="user-page">
    <el-card class="action-card">
      <el-button type="primary">
        <el-icon><Plus /></el-icon>新建用户
      </el-button>
    </el-card>

    <el-card>
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="username" label="用户名" min-width="150" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'superadmin' ? 'danger' : 'primary'">
              {{ getRoleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'enable' ? 'success' : 'danger'">
              {{ row.status === 'enable' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small">编辑</el-button>
            <el-button type="warning" link size="small">重置密码</el-button>
            <el-button type="danger" link size="small" :disabled="row.role === 'superadmin'">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getUserList } from '@/api/auth'

const loading = ref(false)
const tableData = ref([])

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const res = await getUserList({ page: 1, pageSize: 100 })
    if (res.code === 0) tableData.value = res.list || []
  } finally {
    loading.value = false
  }
}

function getRoleLabel(role) {
  const map = { superadmin: '超级管理员', admin: '管理员', guest: '访客' }
  return map[role] || role
}
</script>

<style lang="scss" scoped>
.user-page {
  .action-card { margin-bottom: 20px; }
}
</style>
