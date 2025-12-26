<template>
  <div class="login-container">
    <!-- 主题切换按钮 -->
    <div class="theme-switch" @click="themeStore.toggleTheme">
      <el-icon v-if="themeStore.isDark"><Sunny /></el-icon>
      <el-icon v-else><Moon /></el-icon>
    </div>
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
import { useThemeStore } from '@/stores/theme'
import { Sunny, Moon } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const themeStore = useThemeStore()
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
  background: #e8ecf1;
  position: relative;
  transition: background 0.3s;
}

.theme-switch {
  position: absolute;
  top: 20px;
  right: 20px;
  cursor: pointer;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: #fff;
  border: 1px solid #d0d5dd;
  color: #667085;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s;
  
  &:hover {
    transform: scale(1.1);
    border-color: #409eff;
    color: #409eff;
  }
  
  .el-icon {
    font-size: 20px;
  }
}

.login-box {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border: 1px solid #e4e7ed;
  transition: all 0.3s;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;

  h1 {
    font-size: 32px;
    color: #1a1a2e;
    margin: 0 0 10px;
    letter-spacing: 4px;
    font-weight: 600;
  }

  p {
    color: #606266;
    margin: 0;
    font-size: 14px;
  }
}

.login-form {
  :deep(.el-input__wrapper) {
    background: #fff;
    border: 1px solid #dcdfe6;
    box-shadow: none;
    border-radius: 8px;
    
    &:hover {
      border-color: #c0c4cc;
    }
    
    &.is-focus {
      border-color: #409eff;
    }
  }
  
  :deep(.el-input__inner) {
    color: #303133;
    
    &::placeholder {
      color: #a8abb2;
    }
  }
  
  :deep(.el-input__prefix) {
    color: #909399;
  }

  .login-btn {
    width: 100%;
    height: 44px;
    background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%);
    border: none;
    border-radius: 8px;
    font-size: 16px;
    font-weight: 500;
    letter-spacing: 2px;
    
    &:hover {
      background: linear-gradient(90deg, #66b1ff 0%, #409eff 100%);
    }
  }
}
</style>

<!-- 暗黑模式样式 -->
<style lang="scss">
html.dark {
  .login-container {
    background: #0d0d0d;
  }
  
  .theme-switch {
    background: #252525;
    border-color: #333;
    color: #fbbf24;
    
    &:hover {
      border-color: #fbbf24;
    }
  }
  
  .login-box {
    background: #1a1a1a;
    border-color: #2a2a2a;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  }
  
  .login-header {
    h1 {
      color: #fff;
    }
    
    p {
      color: #888;
    }
  }
  
  .login-form {
    .el-input__wrapper {
      background: #252525 !important;
      border-color: #333 !important;
      
      &:hover {
        border-color: #444 !important;
      }
      
      &.is-focus {
        border-color: #409eff !important;
      }
    }
    
    .el-input__inner {
      color: #fff !important;
      
      &::placeholder {
        color: #666 !important;
      }
    }
    
    .el-input__prefix {
      color: #666 !important;
    }
  }
}
</style>
