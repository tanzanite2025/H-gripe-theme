<template>
  <section class="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-[repeat(auto-fit,minmax(170px,1fr))]" aria-label="页面统计">
    <div
      v-for="item in items"
      :key="item.key || item.label"
      class="group relative flex min-h-14 items-center justify-between gap-3 overflow-hidden rounded-lg border border-dashed border-border/80 bg-card px-3 py-2.5 text-card-foreground shadow-xs transition-all hover:border-primary/40"
    >
      <div class="uds-glow-bg opacity-30 group-hover:opacity-100 transition-opacity" />
      <div class="relative z-10 flex min-w-0 flex-1 items-baseline gap-2">
        <span class="min-w-0 truncate text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ item.label }}</span>
        <strong class="shrink-0 truncate text-xl font-black tabular-nums text-foreground">{{ item.value }}</strong>
      </div>
      <span v-if="item.icon" class="relative z-10 flex size-8 shrink-0 items-center justify-center rounded-full" :class="toneClass(item.tone)">
        <component :is="item.icon" class="size-4" />
      </span>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

type AdminStatTone = 'blue' | 'green' | 'amber' | 'coral' | 'gray'

interface AdminStatItem {
  key?: string
  label: string
  value: string | number
  icon?: Component
  tone?: string
}

defineProps<{
  items: AdminStatItem[]
}>()

const toneClass = (tone?: string) => {
  const tones: Record<AdminStatTone, string> = {
    blue: 'bg-blue-50 text-blue-700',
    green: 'bg-emerald-50 text-emerald-700',
    amber: 'bg-amber-50 text-amber-700',
    coral: 'bg-rose-50 text-rose-700',
    gray: 'bg-muted text-muted-foreground'
  }
  return tone && tone in tones ? tones[tone as AdminStatTone] : tones.gray
}
</script>
