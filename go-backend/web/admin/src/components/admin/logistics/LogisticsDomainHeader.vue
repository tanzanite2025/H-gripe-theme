<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  theme: 'orange' | 'emerald'
  badge: string
  domainLabel: string
  title: string
  description: string
  activeLabel: string
}>()

const accentClass = computed(() => props.theme === 'orange' ? 'bg-orange-500' : 'bg-emerald-500')
</script>

<template>
  <section class="overflow-hidden rounded-[28px] bg-slate-950 px-6 py-6 text-white shadow-sm">
    <div class="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
      <div class="space-y-2">
        <div class="flex items-center gap-3">
          <span :class="accentClass" class="rounded-full px-3 py-1 text-xs font-black tracking-[0.16em]">{{ badge }}</span>
          <span class="text-xs font-semibold text-slate-400">{{ domainLabel }}</span>
        </div>
        <h1 class="text-2xl font-black tracking-tight">{{ title }}</h1>
        <p class="max-w-3xl text-sm leading-6 text-slate-300">{{ description }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-white/20 bg-white/5 px-4 py-3">
        <p class="text-[11px] font-bold uppercase tracking-[0.14em] text-slate-400">当前模块</p>
        <p class="mt-1 text-sm font-black">{{ activeLabel }}</p>
      </div>
    </div>
  </section>
</template>
