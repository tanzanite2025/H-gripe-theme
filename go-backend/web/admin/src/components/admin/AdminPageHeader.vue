<template>
  <header class="uds-header-box px-4 py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
    <div class="uds-glow-bg" />
    <div class="relative z-10 min-w-0">
      <span class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/60 block">MODULE CONTROL / 模块控制</span>
      <h1 class="truncate text-lg font-black tracking-tighter italic uppercase text-foreground">{{ title }}</h1>
      <p v-if="description" class="mt-0.5 text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="relative z-10 flex shrink-0 flex-wrap items-center gap-2">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  description: {
    type: String,
    default: ''
  }
})
</script>
