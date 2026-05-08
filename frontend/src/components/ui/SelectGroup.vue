<script setup lang="ts">
import { cn } from '@/lib/utils'

interface Props {
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
})
</script>

<template>
  <div :class="cn('space-y-1', props.class)">
    <slot />
  </div>
</template>
