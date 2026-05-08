<script setup lang="ts">
import { ref, watch, computed } from 'vue'

interface Props {
  open?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  open: false
})

const emit = defineEmits<{
  close: []
}>()

const isOpen = ref(props.open)

watch(() => props.open, (newVal) => {
  isOpen.value = newVal
})

const handleClose = () => {
  isOpen.value = false
  emit('close')
}

const overlayClasses = computed(() => [
  'fixed inset-0 z-50 flex items-center justify-center p-4',
  'bg-black/60 backdrop-blur-sm',
  'transition-all duration-300',
  isOpen.value ? 'opacity-100' : 'opacity-0 pointer-events-none'
])

const contentClasses = computed(() => [
  'relative w-full max-w-md max-h-[80vh] bg-background border border-border',
  'rounded-xl shadow-2xl overflow-hidden',
  'transition-all duration-300',
  isOpen.value 
    ? 'opacity-100 scale-100 translate-y-0' 
    : 'opacity-0 scale-95 translate-y-4'
])
</script>

<template>
  <Teleport to="body">
    <div 
      :class="overlayClasses"
      @click.self="handleClose"
    >
      <div :class="contentClasses">
        <slot />
      </div>
    </div>
  </Teleport>
</template>
