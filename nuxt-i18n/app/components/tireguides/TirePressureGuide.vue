<template>
  <div class="tire-pressure-guide">
    <h2 class="sizecharts-section__title">Tire Pressure</h2>
    <div class="tire-pressure-card">
      <TirePressureSection @openTireProducts="$emit('openTireProducts')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TirePressureSection from '~/components/TirePressureSection.vue'

defineEmits<{
  (e: 'openTireProducts'): void
}>()
</script>
