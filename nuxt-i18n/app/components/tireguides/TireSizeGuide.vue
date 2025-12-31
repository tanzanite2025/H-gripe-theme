<template>
  <div class="text-slate-100">
    <h2 class="sizecharts-section__title">Tire Size</h2>
    <div class="mt-1">
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
