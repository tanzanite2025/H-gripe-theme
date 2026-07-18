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
    blue: 'border-blue-200 bg-blue-50 text-blue-700',
    green: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    amber: 'border-amber-200 bg-amber-50 text-amber-700',
    coral: 'border-rose-200 bg-rose-50 text-rose-700',
    gray: 'border-border bg-muted text-muted-foreground'
  }
  return tones[props.tone] || tones.gray
})
</script>
