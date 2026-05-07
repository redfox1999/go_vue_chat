<script setup lang="ts">
import { ref } from 'vue'
import { useDark } from '@/composables/useDark'

const { isDark, toggleDark } = useDark()

const count = ref(0)
const showAlert = ref(false)
const alertMessage = ref('')

const increment = () => {
  count.value++
  showAlert.value = true
  alertMessage.value = `计数器增加了！当前值: ${count.value}`
  setTimeout(() => { showAlert.value = false }, 2000)
}

const decrement = () => {
  count.value--
  showAlert.value = true
  alertMessage.value = `计数器减少了！当前值: ${count.value}`
  setTimeout(() => { showAlert.value = false }, 2000)
}

const reset = () => {
  count.value = 0
  showAlert.value = true
  alertMessage.value = '计数器已重置！'
  setTimeout(() => { showAlert.value = false }, 2000)
}
</script>

<template>
  <div class="min-h-screen bg-background py-12 px-4">
    <div class="max-w-2xl mx-auto">
      <div class="flex justify-between items-center mb-8">
        <div class="text-center">
          <h1 class="text-3xl font-bold mb-2">Shadcn Vue 组件测试</h1>
          <p class="text-muted-foreground">使用 shadcn-vue 构建的现代 UI</p>
        </div>
        <Button 
          variant="outline" 
          size="icon" 
          @click="toggleDark"
          class="rounded-full"
        >
          <svg v-if="!isDark" xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="5"></circle>
            <line x1="12" y1="1" x2="12" y2="3"></line>
            <line x1="12" y1="21" x2="12" y2="23"></line>
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
            <line x1="1" y1="12" x2="3" y2="12"></line>
            <line x1="21" y1="12" x2="23" y2="12"></line>
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
          </svg>
        </Button>
      </div>

      <div class="flex justify-center gap-3 mb-8">
        <Button @click="increment">+1 增加</Button>
        <Button variant="secondary" @click="decrement">-1 减少</Button>
        <Button variant="destructive" @click="reset">重置</Button>
      </div>

      <Card class="mb-8">
        <CardHeader>
          <CardTitle>计数器卡片</CardTitle>
          <p class="text-sm text-muted-foreground mt-1">实时显示当前计数</p>
        </CardHeader>
        <CardContent class="text-center py-8">
          <div class="text-6xl font-bold text-primary mb-4">{{ count }}</div>
          <p class="text-muted-foreground">当前计数器值</p>
          <div class="flex justify-center gap-3 mt-6">
            <Button @click="increment">点击增加</Button>
            <Button variant="outline" @click="reset">重置计数</Button>
          </div>
        </CardContent>
      </Card>

      <Transition name="fade">
        <Alert v-if="showAlert" class="mb-8">
          <div class="flex items-center gap-2">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>{{ alertMessage }}</span>
          </div>
        </Alert>
      </Transition>

      <Card>
        <CardContent class="space-y-4">
          <h3 class="text-lg font-semibold">手动输入</h3>
          <div class="space-y-3">
            <label class="block text-sm font-medium">输入数字</label>
            <Input v-model="count" type="number" placeholder="请输入数字..." />
            <div class="flex items-center justify-between">
              <span class="text-sm text-muted-foreground">当前值</span>
              <span class="text-lg font-semibold text-primary">{{ count }}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <div class="mt-8 text-center">
        <p class="text-sm text-muted-foreground">当前主题：{{ isDark ? '暗黑模式' : '亮色模式' }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
