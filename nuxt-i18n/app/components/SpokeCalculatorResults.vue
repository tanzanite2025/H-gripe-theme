<template>
  <section class="spoke-calculator__results-shell mt-6">
    <h2 class="text-xs font-semibold uppercase tracking-[0.18em] tz-text-secondary mb-3">
      {{ t('resourcesSpokeCalculator.calculator.results.title') }}
    </h2>

    <p class="tz-description mb-4 tz-text-secondary">
      {{ t('resourcesSpokeCalculator.calculator.results.description') }}
    </p>

    <div class="grid gap-4 md:grid-cols-2">
      <!-- Front Wheel Results -->
      <div class="space-y-3">
        <div class="text-xs font-semibold tz-text-accent uppercase tracking-wide mb-2">
          {{ t('resourcesSpokeCalculator.calculator.frontWheel') }}
        </div>
        <div class="grid gap-3 grid-cols-2">
          <div class="spoke-calculator__result-card px-4 py-3">
            <div class="mb-1 flex items-center justify-between gap-2">
              <span class="tz-compact-label tz-text-muted">
                {{ t('resourcesSpokeCalculator.calculator.results.leftSide') }}
              </span>
              <span v-if="frontLeftSourceLabel" class="spoke-calculator__source-badge">{{ frontLeftSourceLabel }}</span>
            </div>
            <div class="flex items-baseline gap-1">
              <span class="text-2xl font-semibold text-[var(--tz-site-accent)]">{{ frontLeftDisplay }}</span>
              <span v-if="frontLeftDisplay !== '--'" class="text-xs tz-text-muted">
                {{ t('resourcesSpokeCalculator.calculator.results.unit') }}
              </span>
            </div>
          </div>
          <div class="spoke-calculator__result-card px-4 py-3">
            <div class="mb-1 flex items-center justify-between gap-2">
              <span class="tz-compact-label tz-text-muted">
                {{ t('resourcesSpokeCalculator.calculator.results.rightSide') }}
              </span>
              <span v-if="frontRightSourceLabel" class="spoke-calculator__source-badge">{{ frontRightSourceLabel }}</span>
            </div>
            <div class="flex items-baseline gap-1">
              <span class="text-2xl font-semibold text-[var(--tz-site-accent)]">{{ frontRightDisplay }}</span>
              <span v-if="frontRightDisplay !== '--'" class="text-xs tz-text-muted">
                {{ t('resourcesSpokeCalculator.calculator.results.unit') }}
              </span>
            </div>
          </div>
        </div>
        <div v-if="frontTopologyLengthRows.length" class="spoke-calculator__topology-lengths">
          <div class="spoke-calculator__topology-lengths-title">
            {{ t('resourcesSpokeCalculator.calculator.results.topologyDetails') }}
          </div>
          <div v-for="row in frontTopologyLengthRows" :key="row.key" class="spoke-calculator__topology-length-row">
            <span>{{ row.label }} · {{ row.count }} {{ t('resourcesSpokeCalculator.calculator.results.spokes') }}</span>
            <strong>{{ formatResultLength(row.lengthMm) }} {{ t('resourcesSpokeCalculator.calculator.results.unit') }}</strong>
          </div>
        </div>
        <div class="spoke-calculator__tension-summary">
          <div class="flex items-baseline justify-between gap-3">
            <span class="tz-compact-label tz-text-muted">
              {{ t('resourcesSpokeCalculator.calculator.results.predictedTensionRatio') }}
            </span>
            <span class="text-lg font-semibold text-[var(--tz-site-accent)]">{{ formatTensionRatio(frontResult?.tensionRatio) }}</span>
          </div>
          <div class="mt-1 tz-caption tz-text-muted">
            {{ t('resourcesSpokeCalculator.calculator.results.directionalRatio', {
              ratio: formatDirectionalTensionRatio(frontResult?.tensionRatio),
              side: lowerTensionSideLabel(frontResult?.tensionRatio),
            }) }}
          </div>
        </div>
      </div>

      <!-- Rear Wheel Results -->
      <div class="space-y-3">
        <div class="text-xs font-semibold tz-text-accent uppercase tracking-wide mb-2">
          {{ t('resourcesSpokeCalculator.calculator.rearWheel') }}
        </div>
        <div class="grid gap-3 grid-cols-2">
          <div class="spoke-calculator__result-card px-4 py-3">
            <div class="mb-1 flex items-center justify-between gap-2">
              <span class="tz-compact-label tz-text-muted">
                {{ t('resourcesSpokeCalculator.calculator.results.leftSide') }}
              </span>
              <span v-if="rearLeftSourceLabel" class="spoke-calculator__source-badge">{{ rearLeftSourceLabel }}</span>
            </div>
            <div class="flex items-baseline gap-1">
              <span class="text-2xl font-semibold text-[var(--tz-site-accent)]">{{ rearLeftDisplay }}</span>
              <span v-if="rearLeftDisplay !== '--'" class="text-xs tz-text-muted">
                {{ t('resourcesSpokeCalculator.calculator.results.unit') }}
              </span>
            </div>
          </div>
          <div class="spoke-calculator__result-card px-4 py-3">
            <div class="mb-1 flex items-center justify-between gap-2">
              <span class="tz-compact-label tz-text-muted">
                {{ t('resourcesSpokeCalculator.calculator.results.rightSide') }}
              </span>
              <span v-if="rearRightSourceLabel" class="spoke-calculator__source-badge">{{ rearRightSourceLabel }}</span>
            </div>
            <div class="flex items-baseline gap-1">
              <span class="text-2xl font-semibold text-[var(--tz-site-accent)]">{{ rearRightDisplay }}</span>
              <span v-if="rearRightDisplay !== '--'" class="text-xs tz-text-muted">
                {{ t('resourcesSpokeCalculator.calculator.results.unit') }}
              </span>
            </div>
          </div>
        </div>
        <div v-if="rearTopologyLengthRows.length" class="spoke-calculator__topology-lengths">
          <div class="spoke-calculator__topology-lengths-title">
            {{ t('resourcesSpokeCalculator.calculator.results.topologyDetails') }}
          </div>
          <div v-for="row in rearTopologyLengthRows" :key="row.key" class="spoke-calculator__topology-length-row">
            <span>{{ row.label }} · {{ row.count }} {{ t('resourcesSpokeCalculator.calculator.results.spokes') }}</span>
            <strong>{{ formatResultLength(row.lengthMm) }} {{ t('resourcesSpokeCalculator.calculator.results.unit') }}</strong>
          </div>
        </div>
        <div class="spoke-calculator__tension-summary">
          <div class="flex items-baseline justify-between gap-3">
            <span class="tz-compact-label tz-text-muted">
              {{ t('resourcesSpokeCalculator.calculator.results.predictedTensionRatio') }}
            </span>
            <span class="text-lg font-semibold text-[var(--tz-site-accent)]">{{ formatTensionRatio(rearResult?.tensionRatio) }}</span>
          </div>
          <div class="mt-1 tz-caption tz-text-muted">
            {{ t('resourcesSpokeCalculator.calculator.results.directionalRatio', {
              ratio: formatDirectionalTensionRatio(rearResult?.tensionRatio),
              side: lowerTensionSideLabel(rearResult?.tensionRatio),
            }) }}
          </div>
        </div>
      </div>
    </div>

    <div class="spoke-calculator__results-note mt-6">
      {{ t('resourcesSpokeCalculator.calculator.results.fallbackNote') }}
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import type { SpokeTensionRatio } from '~~/types/spoke'
import type { SpokeWheelResult } from '~/types/spokeCalculator'

const props = defineProps<{
  frontResult: SpokeWheelResult | null
  rearResult: SpokeWheelResult | null
}>()

const { t } = useI18n()

const formatResultLength = (value: number | null | undefined) => (
  value == null ? '--' : value.toFixed(2)
)

const formatTensionRatio = (value: SpokeTensionRatio | null | undefined) => (
  value == null ? '--' : `${(value.lowerToHigher * 100).toFixed(2)}%`
)

const formatDirectionalTensionRatio = (value: SpokeTensionRatio | null | undefined) => {
  if (!value) return '--'
  if (value.leftToRight <= 1) {
    return `${(value.leftToRight * 100).toFixed(2)}% : 100.00%`
  }
  return `100.00% : ${(value.rightToLeft * 100).toFixed(2)}%`
}

const lowerTensionSideLabel = (value: SpokeTensionRatio | null | undefined) => {
  if (!value || value.lowerSide === 'balanced') {
    return t('resourcesSpokeCalculator.calculator.results.balanced')
  }
  return value.lowerSide === 'left'
    ? t('resourcesSpokeCalculator.calculator.results.leftSide')
    : t('resourcesSpokeCalculator.calculator.results.rightSide')
}

const resultSourceLabel = (source: SpokeWheelResult['leftSource'] | undefined) => {
  if (source === 'calculated') return t('resourcesSpokeCalculator.calculator.results.calculated')
  return ''
}

interface TopologyLengthRow {
  key: string
  label: string
  count: number
  lengthMm: number
}

const topologyTypeLabel = (type: 'leading' | 'trailing' | 'nondrive') => {
  if (type === 'leading') return t('resourcesSpokeCalculator.calculator.results.leading')
  if (type === 'trailing') return t('resourcesSpokeCalculator.calculator.results.trailing')
  return t('resourcesSpokeCalculator.calculator.results.nondrive')
}

const buildTopologyLengthRows = (result: SpokeWheelResult | null): TopologyLengthRow[] => {
  if (!result || result.distribution === 'symmetric_1to1' || result.spokeLengths.length === 0) return []
  const groups = new Map<string, { side: 'A' | 'B'; type: 'leading' | 'trailing' | 'nondrive'; physicalSide: 'left' | 'right'; lengths: number[] }>()
  for (const spoke of result.spokeLengths) {
    const key = `${spoke.side}-${spoke.type}-${spoke.physicalSide}`
    const group = groups.get(key) || {
      side: spoke.side,
      type: spoke.type,
      physicalSide: spoke.physicalSide,
      lengths: [],
    }
    group.lengths.push(spoke.lengthMm)
    groups.set(key, group)
  }
  return [...groups.entries()].map(([key, group]) => ({
    key,
    label: `${group.side} · ${topologyTypeLabel(group.type)} (${group.physicalSide === 'left'
      ? t('resourcesSpokeCalculator.calculator.results.leftSide')
      : t('resourcesSpokeCalculator.calculator.results.rightSide')})`,
    count: group.lengths.length,
    lengthMm: group.lengths.reduce((sum, value) => sum + value, 0) / group.lengths.length,
  }))
}

const frontLeftDisplay = computed(() => formatResultLength(props.frontResult?.leftLengthMm))
const frontRightDisplay = computed(() => formatResultLength(props.frontResult?.rightLengthMm))
const rearLeftDisplay = computed(() => formatResultLength(props.rearResult?.leftLengthMm))
const rearRightDisplay = computed(() => formatResultLength(props.rearResult?.rightLengthMm))
const frontTopologyLengthRows = computed(() => buildTopologyLengthRows(props.frontResult))
const rearTopologyLengthRows = computed(() => buildTopologyLengthRows(props.rearResult))

const frontLeftSourceLabel = computed(() => resultSourceLabel(props.frontResult?.leftSource))
const frontRightSourceLabel = computed(() => resultSourceLabel(props.frontResult?.rightSource))
const rearLeftSourceLabel = computed(() => resultSourceLabel(props.rearResult?.leftSource))
const rearRightSourceLabel = computed(() => resultSourceLabel(props.rearResult?.rightSource))
</script>

<style scoped>
.spoke-calculator__results-shell {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-shell-surface);
  box-shadow: 0 10px 26px -14px rgba(20, 32, 43, 0.12);
  padding: 1.25rem;
}

.spoke-calculator__result-card {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-result-surface);
}

.spoke-calculator__tension-summary {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-result-surface);
  padding: 0.75rem 1rem;
}

.spoke-calculator__topology-lengths {
  display: grid;
  gap: 0.35rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-result-surface);
  padding: 0.65rem 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
}

.spoke-calculator__topology-lengths-title {
  color: var(--tz-text-muted);
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.spoke-calculator__topology-length-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
}

.spoke-calculator__topology-length-row strong {
  flex: 0 0 auto;
  color: var(--tz-text-primary);
  font-weight: 800;
}

.spoke-calculator__source-badge {
  flex: 0 0 auto;
  border: 1px solid var(--spoke-border);
  border-radius: 999px;
  padding: 0.125rem 0.375rem;
  color: var(--tz-text-muted);
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
}

.spoke-calculator__results-note {
  padding: 0.95rem 1rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-result-surface);
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  line-height: 1.55;
}

@media (max-width: 767px) {
  .spoke-calculator__results-shell {
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }
}
</style>
