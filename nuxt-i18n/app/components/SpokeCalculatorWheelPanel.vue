<template>
  <div class="spoke-calculator__panel space-y-4">
    <h3 class="text-sm font-semibold text-[var(--tz-site-accent)] uppercase tracking-wide">
      {{ side === 'front'
        ? t('resourcesSpokeCalculator.calculator.frontWheel')
        : t('resourcesSpokeCalculator.calculator.rearWheel') }}
    </h3>

    <SpokeCalculatorBuildSettings
      :side="side"
      :config="config"
      :options="props.options"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokeCalculatorBuildSettings from '~/components/SpokeCalculatorBuildSettings.vue'
import type {
  SpokeCalculatorManualOptions,
  SpokeWheelBuildConfig,
  SpokeWheelSide,
} from '~/types/spokeCalculator'

const props = defineProps<{
  side: SpokeWheelSide
  config: SpokeWheelBuildConfig
  options: SpokeCalculatorManualOptions
}>()

const { t } = useI18n()
const side = props.side
const config = props.config
</script>

<style scoped>
.spoke-calculator__panel {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-panel-surface);
  padding: 1rem;
}
</style>
