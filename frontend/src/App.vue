<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Toast from './components/ui/Toast.vue'
import { registerToast } from './composables/useToast'

const toastRef = ref<InstanceType<typeof Toast> | null>(null)

onMounted(() => {
  if (toastRef.value) {
    registerToast(toastRef.value)
  }
})
</script>

<template>
  <router-view />
  <Toast ref="toastRef" />
</template>
