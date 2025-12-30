<template>
  <div class="text-slate-100">
    <h2 class="sizecharts-section__title">Tire Size</h2>
    <div
      class="tire-size-card mt-1 rounded-2xl bg-slate-900/70 p-4 md:p-5 shadow-[4px_4px_18px_rgba(0,0,0,1)]"
    >
      <TireSizeSection @openTireProducts="$emit('openTireProducts')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TireSizeSection from '~/components/TireSizeSection.vue'

defineEmits<{
  (e: 'openTireProducts'): void
}>()
</script>
