<template>
  <section class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-label="页面统计">
    <div
      v-for="item in items"
      :key="item.key || item.label"
      class="flex min-h-24 items-start justify-between gap-3 rounded-lg border bg-card p-4 text-card-foreground shadow-xs"
    >
      <div class="min-w-0">
        <span class="block text-xs font-medium text-muted-foreground">{{ item.label }}</span>
        <strong class="mt-2 block truncate text-2xl font-semibold tabular-nums">{{ item.value }}</strong>
      </div>
      <span v-if="item.icon" class="flex size-9 shrink-0 items-center justify-center rounded-lg" :class="toneClass(item.tone)">
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
