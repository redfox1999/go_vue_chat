import { toast } from 'vue-sonner'

export function useToast() {
  return {
    show: (message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info', duration?: number) => {
      switch (type) {
        case 'success':
          return toast.success(message, { duration })
        case 'error':
          return toast.error(message, { duration })
        case 'warning':
          return toast.warning(message, { duration })
        case 'info':
        default:
          return toast(message, { duration })
      }
    },
    success: (message: string, duration?: number) => {
      return toast.success(message, { duration })
    },
    error: (message: string, duration?: number) => {
      return toast.error(message, { duration })
    },
    warning: (message: string, duration?: number) => {
      return toast.warning(message, { duration })
    },
    info: (message: string, duration?: number) => {
      return toast(message, { duration })
    },
    toast
  }
}
