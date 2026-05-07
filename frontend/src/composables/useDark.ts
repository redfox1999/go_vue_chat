import { ref, onMounted, watch } from 'vue'

const isDark = ref(false)

export function useDark() {
  const toggleDark = () => {
    isDark.value = !isDark.value
  }

  const applyDarkMode = (dark: boolean) => {
    if (dark) {
      document.documentElement.classList.add('dark')
      localStorage.setItem('theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark')
      localStorage.setItem('theme', 'light')
    }
  }

  onMounted(() => {
    const savedTheme = localStorage.getItem('theme')
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    
    if (savedTheme) {
      isDark.value = savedTheme === 'dark'
    } else if (prefersDark) {
      isDark.value = true
    }
    
    applyDarkMode(isDark.value)
  })

  watch(isDark, (newVal) => {
    applyDarkMode(newVal)
  })

  return {
    isDark,
    toggleDark,
  }
}
