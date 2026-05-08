<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue'
import Toast from './components/ui/Toast.vue'

const hasRenderError = ref(false)

onErrorCaptured((err) => {
  console.error('[App] Render error:', err)
  hasRenderError.value = true
  return false
})
</script>

<template>
  <Toast />
  <div v-if="hasRenderError" class="flex items-center justify-center h-screen bg-background">
    <div class="text-center">
      <p class="text-red-500 mb-4">页面加载出错</p>
      <button @click="hasRenderError = false" class="px-4 py-2 bg-indigo-500 text-white rounded">重试</button>
    </div>
  </div>
  <router-view v-else />
</template>
