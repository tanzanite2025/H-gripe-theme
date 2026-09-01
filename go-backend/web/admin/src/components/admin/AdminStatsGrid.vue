<template>
  <section
    :class="[
      'grid grid-cols-1 gap-2 sm:grid-cols-2',
      compact
        ? 'xl:grid-cols-[repeat(8,minmax(0,1fr))]'
        : 'xl:grid-cols-[repeat(auto-fit,minmax(170px,1fr))]'
    ]"
    aria-label="页面统计"
  >
    <div
      v-for="item in items"
      :key="item.key || item.label"
      :class="[
        'group relative flex min-h-14 items-center justify-between overflow-hidden rounded-lg border border-dashed border-border/80 bg-card text-card-foreground shadow-xs transition-all hover:border-primary/40',
        compact ? 'gap-2 px-2.5 py-2' : 'gap-3 px-3 py-2.5'
      ]"
    >
      <div :class="['relative z-10 flex min-w-0 flex-1 items-baseline', compact ? 'gap-1.5' : 'gap-2']">
        <span :class="['min-w-0 truncate font-black uppercase tracking-widest text-muted-foreground/60', compact ? 'text-[9px]' : 'text-[10px]']">{{ item.label }}</span>
        <strong :class="['shrink-0 truncate font-black tabular-nums text-foreground', compact ? 'text-lg' : 'text-xl']">{{ item.value }}</strong>
      </div>
      <span
        v-if="item.icon"
        :class="[
          'relative z-10 flex shrink-0 items-center justify-center rounded-full',
          compact ? 'size-7' : 'size-8',
          toneClass(item.tone)
        ]"
      >
        <component :is="item.icon" :class="compact ? 'size-3.5' : 'size-4'" />
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
  compact?: boolean
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
