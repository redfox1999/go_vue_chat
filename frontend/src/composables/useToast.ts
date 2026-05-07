import { ref } from 'vue'

interface ToastRef {
  show: (message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number) => void
}

const toastRef = ref<ToastRef | null>(null)

export function useToast() {
  return {
    show: (message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info', duration = 3000) => {
      if (toastRef.value) {
        toastRef.value.show(message, type, duration)
      }
    },
    success: (message: string, duration?: number) => {
      toastRef.value?.show(message, 'success', duration)
    },
    error: (message: string, duration?: number) => {
      toastRef.value?.show(message, 'error', duration)
    },
    warning: (message: string, duration?: number) => {
      toastRef.value?.show(message, 'warning', duration)
    },
    info: (message: string, duration?: number) => {
      toastRef.value?.show(message, 'info', duration)
    }
  }
}

export function registerToast(ref: ToastRef) {
  toastRef.value = ref
}
