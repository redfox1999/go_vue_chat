<script setup lang="ts">
import { computed, inject, onMounted, ref } from 'vue'
import { cn } from '@/lib/utils'

interface Props {
  value: string
  class?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false
})

const emit = defineEmits<{
  (e: 'select', value: string): void
}>()

const itemRef = ref<HTMLDivElement | null>(null)
const registerOption = inject<(val: string, label: string) => void>('selectRegisterOption')

onMounted(() => {
  if (registerOption && itemRef.value) {
    const label = itemRef.value.textContent?.trim() || props.value
    registerOption(props.value, label)
  }
})

const handleClick = () => {
  if (!props.disabled) {
    emit('select', props.value)
  }
}
</script>

<template>
  <div
    ref="itemRef"
    role="option"
    :data-value="value"
    :class="cn(
      'flex items-center w-full rounded-md px-3 py-2 text-sm outline-none transition-colors cursor-pointer data-[disabled]:pointer-events-none data-[disabled]:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 focus-visible:ring-offset-popover',
      'hover:bg-accent hover:text-accent-foreground',
      props.class
    )"
    @click="handleClick"
  >
    <slot />
  </div>
</template>
