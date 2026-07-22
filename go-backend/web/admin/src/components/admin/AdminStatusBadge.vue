<template>
  <Badge variant="outline" :class="toneClass">
    <slot />
  </Badge>
</template>

<script setup>
import { computed } from 'vue'
import { Badge } from '@/components/ui/badge'

const props = defineProps({
  tone: {
    type: String,
    default: 'gray'
  }
})

const toneClass = computed(() => {
  const tones = {
    blue: 'rounded-full h-4 px-2 border border-blue-500/20 bg-blue-500/10 text-blue-600 dark:text-blue-400 font-mono text-[8px] font-bold tracking-tight inline-flex items-center',
    green: 'rounded-full h-4 px-2 border border-emerald-500/20 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-mono text-[8px] font-bold tracking-tight inline-flex items-center',
    amber: 'rounded-full h-4 px-2 border border-amber-500/20 bg-amber-500/10 text-amber-600 dark:text-amber-400 font-mono text-[8px] font-bold tracking-tight inline-flex items-center',
    coral: 'rounded-full h-4 px-2 border border-rose-500/20 bg-rose-500/10 text-rose-600 dark:text-rose-400 font-mono text-[8px] font-bold tracking-tight inline-flex items-center animate-pulse',
    gray: 'rounded-full h-4 px-2 border border-border bg-muted/60 text-muted-foreground font-mono text-[8px] font-bold tracking-tight inline-flex items-center'
  }
  return tones[props.tone] || tones.gray
})
</script>
