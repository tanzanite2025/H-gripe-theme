<template>
  <div
    :class="[
      'h-7 w-7 rounded-full flex items-center justify-center text-[12px] font-semibold transition-colors',
      active ? 'bg-white text-slate-900' : 'border tz-border-strong/30 tz-text-muted'
    ]"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  active?: boolean
}>()
</script>
