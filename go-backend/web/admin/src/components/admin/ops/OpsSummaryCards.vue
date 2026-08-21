<template>
  <section class="grid gap-3 md:grid-cols-4">
    <div
      v-for="item in items"
      :key="item.key"
      class="rounded-2xl border border-dashed border-border/80 bg-card p-3"
    >
      <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">
        {{ item.label }}
      </p>
      <p class="mt-2 text-2xl font-black" :class="valueClass(item.tone)">
        {{ item.value }}
      </p>
      <p class="mt-1 text-[10px] text-muted-foreground">
        {{ item.detail }}
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
export interface OpsSummaryCardItem {
  key: string
  label: string
  value: number | string
  detail: string
  tone?: 'default' | 'green' | 'primary' | 'amber' | 'coral'
}

defineProps<{
  items: readonly OpsSummaryCardItem[]
}>()

const valueClass = (tone: OpsSummaryCardItem['tone']): string => ({
  green: 'text-emerald-600',
  primary: 'text-primary',
  amber: 'text-amber-600',
  coral: 'text-rose-600',
}[tone || 'default'] || '')
</script>
