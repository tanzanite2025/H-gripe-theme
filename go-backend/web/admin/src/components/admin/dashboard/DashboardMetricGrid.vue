<template>
  <section class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-label="关键指标">
    <button
      v-for="metric in metricCards"
      :key="metric.key"
      type="button"
      class="group relative flex min-h-31 flex-col justify-between rounded-[24px] border border-dashed border-border/80 bg-card p-4 text-left text-card-foreground overflow-hidden shadow-xs transition-all hover:-translate-y-px hover:border-primary/40 focus-visible:outline-none"
      @click="emit('navigate', metric.path)"
    >
      <div class="uds-glow-bg opacity-40 group-hover:opacity-100 transition-opacity" />
      <div class="relative z-10 flex items-start justify-between gap-3">
        <div class="min-w-0">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ metric.label }}</span>
          <strong class="mt-1 block truncate text-2xl font-black italic tracking-tighter tabular-nums text-foreground">{{ metric.value }}</strong>
        </div>
        <span class="flex size-9 shrink-0 items-center justify-center rounded-full" :class="metricToneClass(metric.tone)">
          <component :is="metric.icon" class="size-4" />
        </span>
      </div>
      <div class="relative z-10 flex items-center justify-between gap-2 text-[10px] font-mono text-muted-foreground/75">
        <span class="uppercase tracking-widest font-black text-[9px] opacity-60">{{ metric.detailLabel }}</span>
        <strong class="font-bold tabular-nums text-foreground/80">{{ metric.detailValue }}</strong>
      </div>
    </button>
  </section>
</template>

<script setup>
defineProps({
  metricCards: { type: Array, default: () => [] },
  metricToneClass: { type: Function, required: true },
})

const emit = defineEmits(['navigate'])
</script>
