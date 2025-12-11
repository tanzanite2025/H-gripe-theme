<template>
  <div
    :class="[
      'h-7 w-7 rounded-full flex items-center justify-center text-[12px] font-semibold transition-colors',
      active ? 'bg-white text-slate-900' : 'border border-white/30 text-white/70'
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
