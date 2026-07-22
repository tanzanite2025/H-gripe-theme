<template>
  <section class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-label="页面统计">
    <div
      v-for="item in items"
      :key="item.key || item.label"
      class="group relative flex min-h-24 items-start justify-between gap-3 rounded-[24px] border border-dashed border-border/80 bg-card p-4 text-card-foreground shadow-xs overflow-hidden transition-all hover:border-primary/40"
    >
      <div class="uds-glow-bg opacity-30 group-hover:opacity-100 transition-opacity" />
      <div class="relative z-10 min-w-0">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ item.label }}</span>
        <strong class="mt-1.5 block truncate text-2xl font-black italic tracking-tighter tabular-nums text-foreground">{{ item.value }}</strong>
      </div>
      <span v-if="item.icon" class="relative z-10 flex size-9 shrink-0 items-center justify-center rounded-full" :class="toneClass(item.tone)">
        <component :is="item.icon" class="size-4" />
      </span>
    </div>
  </section>
</template>

<script setup>
defineProps({
  items: {
    type: Array,
    required: true
  }
})

const toneClass = (tone) => {
  const tones = {
    blue: 'bg-blue-50 text-blue-700',
    green: 'bg-emerald-50 text-emerald-700',
    amber: 'bg-amber-50 text-amber-700',
    coral: 'bg-rose-50 text-rose-700',
    gray: 'bg-muted text-muted-foreground'
  }
  return tones[tone] || tones.gray
}
</script>
