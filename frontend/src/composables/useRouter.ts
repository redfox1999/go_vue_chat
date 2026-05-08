import { ref } from 'vue'

const currentPath = ref(window.location.pathname)

export function navigateTo(path: string) {
  window.history.pushState(null, '', path)
  currentPath.value = path
}

export function useRouter() {
  return {
    currentPath,
    navigateTo
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('popstate', () => {
    currentPath.value = window.location.pathname
  })
}
