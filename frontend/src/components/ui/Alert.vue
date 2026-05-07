<script setup lang="ts">
import { cn } from '@/lib/utils'

interface Props {
  variant?: 'default' | 'destructive'
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
})

const variants = {
  default: 'bg-background text-foreground',
  destructive: 'bg-destructive/10 text-destructive border-destructive/20',
}
</script>

<template>
  <div :class="cn('relative w-full rounded-lg border p-4', variants[props.variant], props.class)">
    <slot />
  </div>
</template>
