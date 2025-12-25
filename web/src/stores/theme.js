import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  // 从 localStorage 读取主题设置，默认为深色
  const isDark = ref(localStorage.getItem('theme') !== 'light')

  function toggleTheme() {
    isDark.value = !isDark.value
  }

  function setTheme(dark) {
    isDark.value = dark
  }

  // 监听主题变化，更新 DOM 和 localStorage
  // Element Plus 官方暗黑模式通过 html.dark 类名控制
  watch(isDark, (val) => {
    localStorage.setItem('theme', val ? 'dark' : 'light')
    if (val) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }, { immediate: true })

  return {
    isDark,
    toggleTheme,
    setTheme
  }
})
