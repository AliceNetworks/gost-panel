import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  // 默认暗色主题，除非明确设置为 light
  const storedTheme = localStorage.getItem('theme')
  const isDark = ref(storedTheme !== 'light')

  const toggle = () => {
    isDark.value = !isDark.value
  }

  watch(isDark, (value) => {
    localStorage.setItem('theme', value ? 'dark' : 'light')
    if (value) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }, { immediate: true })

  return {
    isDark,
    toggle,
  }
})
