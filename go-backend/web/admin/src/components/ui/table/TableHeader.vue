<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <thead
    data-slot="table-header"
    :class="cn('[&_tr]:border-b [&_tr]:border-dashed [&_tr]:border-border/70 [&_tr]:bg-muted/30', props.class)"
  >
    <slot />
  </thead>
</template>
