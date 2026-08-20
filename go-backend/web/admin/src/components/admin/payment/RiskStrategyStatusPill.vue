<template>
  <span
    class="inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-black uppercase tracking-wider"
    :class="toneClass"
  >
    {{ status || '-' }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  status?: string
}>(), {
  status: '',
})

const toneClass = computed(() => {
  const status = String(props.status || '').toLowerCase()
  if (['approved', 'accepted', 'won', 'succeeded', 'resolved'].includes(status)) {
    return 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'
  }
  if (['rejected', 'dismissed', 'lost', 'needs_response', 'warning_needs_response', 'waiting_for_seller_response', 'open'].includes(status)) {
    return 'border-rose-500/25 bg-rose-500/10 text-rose-700'
  }
  if (['pending', 'under_review', 'processing', 'waiting_for_buyer_response'].includes(status)) {
    return 'border-amber-500/25 bg-amber-500/10 text-amber-700'
  }
  return 'border-border bg-muted text-muted-foreground'
})
</script>
