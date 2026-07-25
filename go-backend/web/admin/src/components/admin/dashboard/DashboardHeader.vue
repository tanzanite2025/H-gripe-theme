<template>
  <header class="uds-header-box px-4 py-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
    <div class="uds-glow-bg" />
    <div class="relative z-10">
      <span class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/60 block">OVERVIEW / 综合监控</span>
      <h1 class="text-lg font-black tracking-tighter italic uppercase text-foreground">仪表板</h1>
    </div>
    <div class="relative z-10 flex items-center gap-1.5 text-[10px] font-mono font-bold text-muted-foreground/75 bg-background/70 backdrop-blur-xs px-3 py-1 rounded-full border border-dashed border-border/80">
      <CalendarDays class="size-3.5 text-primary" />
      <span>{{ currentDate }}</span>
    </div>
  </header>
</template>

<script setup>
import { CalendarDays } from '@lucide/vue'

defineProps({
  currentDate: { type: String, required: true },
})
</script>
