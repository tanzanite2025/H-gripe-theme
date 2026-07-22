<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <th
    data-slot="table-head"
    :class="cn('text-muted-foreground/60 h-9 px-3 first:pl-4 last:pr-4 text-left align-middle text-[10px] font-black uppercase tracking-widest whitespace-nowrap [&:has([role=checkbox])]:pr-0', props.class)"
  >
    <slot />
  </th>
</template>
