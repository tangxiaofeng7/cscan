<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>CSCAN</h1>
        <p>资产安全扫描平台</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" class="login-form">
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            prefix-icon="User"
            size="large"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            prefix-icon="Lock"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="login-btn"
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function handleLogin() {
  await formRef.value.validate()
  loading.value = true
  try {
    const res = await userStore.login(form)
    if (res.code === 0) {
      ElMessage.success('登录成功')
      router.push('/dashboard')
    } else {
      ElMessage.error(res.msg || '登录失败')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0d0d0d;
}

.login-box {
  width: 400px;
  padding: 40px;
  background: #1a1a1a;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
  border: 1px solid #2a2a2a;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;

  h1 {
    font-size: 32px;
    color: #fff;
    margin: 0 0 10px;
    letter-spacing: 4px;
  }

  p {
    color: #888;
    margin: 0;
  }
}

.login-form {
  :deep(.el-input__wrapper) {
    background: #252525;
    border: 1px solid #333;
    box-shadow: none;
    
    &:hover, &:focus {
      border-color: #409eff;
    }
  }
  
  :deep(.el-input__inner) {
    color: #fff;
    
    &::placeholder {
      color: #666;
    }
  }
  
  :deep(.el-input__prefix) {
    color: #666;
  }

  .login-btn {
    width: 100%;
    background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%);
    border: none;
    
    &:hover {
      background: linear-gradient(90deg, #66b1ff 0%, #409eff 100%);
    }
  }
}
</style>
