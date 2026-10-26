<template>
  <button
    type="button"
    class="relative grid size-12 shrink-0 place-items-center rounded-full transition-[transform,background-color,box-shadow] duration-200 hover:-translate-y-1 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 motion-reduce:transition-none motion-reduce:hover:translate-y-0 sm:size-14"
    :class="appearance === 'support'
      ? 'border border-cyan-500/10 bg-cyan-500/10 text-cyan-600 hover:bg-cyan-500/20 hover:shadow-lg focus-visible:ring-cyan-600'
      : 'bg-slate-900 text-white hover:bg-slate-700 focus-visible:ring-slate-500'"
  >
    <slot />
    <span
      v-if="badgeLabel"
      class="absolute -right-1 -top-1 grid min-h-5 min-w-5 place-items-center rounded-full border-2 border-background bg-rose-500 px-1 text-[10px] font-black leading-none text-white"
      aria-hidden="true"
    >
      {{ badgeLabel }}
    </span>
  </button>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  appearance?: 'primary' | 'support'
  badgeLabel?: string
}>(), { appearance: 'primary', badgeLabel: '' })
</script>
