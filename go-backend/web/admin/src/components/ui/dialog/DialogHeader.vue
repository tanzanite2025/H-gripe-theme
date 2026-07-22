<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <div
    data-slot="dialog-header"
    :class="cn('gap-1 flex flex-col border-b border-dashed border-border/70 pb-3', props.class)"
  >
    <slot />
  </div>
</template>
