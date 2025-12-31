<template>
  <div class="tire-pressure-guide">
    <h2 class="sizecharts-section__title">Tire Pressure</h2>
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
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
