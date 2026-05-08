<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick, provide } from 'vue'
import { cn } from '@/lib/utils'

interface Props {
  modelValue?: string | number
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const isOpen = ref(false)
const triggerRef = ref<HTMLButtonElement | null>(null)
const contentRef = ref<HTMLDivElement | null>(null)
const optionsMap = ref<Map<string, string>>(new Map())
const displayText = ref('请选择')

const value = computed({
  get: () => String(props.modelValue ?? ''),
  set: (newValue) => emit('update:modelValue', newValue),
})

const updateDisplayText = () => {
  displayText.value = optionsMap.value.get(value.value) || value.value || '请选择'
}

// 提供注册选项的方法给子组件
provide('selectRegisterOption', (val: string, label: string) => {
  optionsMap.value.set(val, label)
  updateDisplayText()
})

const toggleOpen = () => {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    nextTick(() => {
      updateOptionsMap()
      updateContentPosition()
    })
  }
}

const updateOptionsMap = () => {
  if (!contentRef.value) return
  const options = contentRef.value.querySelectorAll('[role="option"]')
  optionsMap.value.clear()
  options.forEach(option => {
    const optValue = option.getAttribute('data-value')
    if (optValue) {
      optionsMap.value.set(optValue, option.textContent?.trim() || optValue)
    }
  })
  updateDisplayText()
}

const updateContentPosition = () => {
  if (!triggerRef.value || !contentRef.value) return
  const rect = triggerRef.value.getBoundingClientRect()
  contentRef.value.style.top = `${rect.bottom + 2}px`
  contentRef.value.style.left = `${rect.left}px`
  contentRef.value.style.width = `${rect.width}px`
}

const handleDocumentClick = (e: MouseEvent) => {
  if (contentRef.value && !contentRef.value.contains(e.target as Node)) {
    if (triggerRef.value && !triggerRef.value.contains(e.target as Node)) {
      isOpen.value = false
    }
  }
}

const handleContentClick = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  const option = target.closest('[role="option"]') as HTMLElement
  if (option) {
    const optionValue = option.getAttribute('data-value')
    if (optionValue) {
      value.value = optionValue
      isOpen.value = false
    }
  }
}

onMounted(() => {
  // 延迟 100ms 确保 Teleport 内容已渲染
  setTimeout(() => {
    updateOptionsMap()
  }, 100)
})

watch(isOpen, (newVal) => {
  if (newVal) {
    document.addEventListener('click', handleDocumentClick)
  } else {
    document.removeEventListener('click', handleDocumentClick)
  }
})

watch(() => props.modelValue, () => {
  if (isOpen.value) {
    nextTick(updateOptionsMap)
  }
})
</script>

<template>
  <div class="relative inline-flex flex-col w-full">
    <button
      ref="triggerRef"
      role="combobox"
      :aria-expanded="isOpen"
      :class="cn(
        'flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-foreground ring-offset-background transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
        props.class
      )"
      @click.stop="toggleOpen"
    >
      <span class="truncate">{{ displayText }}</span>
      <svg 
        xmlns="http://www.w3.org/2000/svg" 
        class="w-4 h-4 ml-2 shrink-0 transition-transform"
        :class="{ 'rotate-180': isOpen }"
        viewBox="0 0 24 24" 
        fill="none" 
        stroke="currentColor" 
        stroke-width="2"
      >
        <path d="M6 9l6 6 6-6"></path>
      </svg>
    </button>
    
    <Teleport to="body">
      <div
        ref="contentRef"
        :class="[
          'fixed z-[100] overflow-hidden rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-lg transition-all duration-200',
          isOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none absolute -top-9999'
        ]"
        @click="handleContentClick"
      >
        <slot />
      </div>
    </Teleport>
  </div>
</template>
