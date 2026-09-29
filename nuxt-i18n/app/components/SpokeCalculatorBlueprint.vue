<template>
  <div class="spoke-calculator">
    <div class="grid gap-6 items-start">
      <section class="spoke-calculator__shell">
        <h2 class="text-xs font-semibold uppercase tracking-[0.18em] tz-text-secondary mb-4">
          {{ t('resourcesSpokeCalculator.calculator.wheelSetup') }}
        </h2>

        <!-- Two-column layout: Front Wheel | Rear Wheel -->
        <div class="grid gap-6 md:grid-cols-2">
          <!-- ========== FRONT WHEEL COLUMN ========== -->
          <div class="spoke-calculator__panel space-y-4">
            <h3 class="text-sm font-semibold text-[var(--tz-site-accent)] uppercase tracking-wide">
              {{ t('resourcesSpokeCalculator.calculator.frontWheel') }}
            </h3>

            <div class="spoke-calculator__blueprint-sheet">
              <div class="spoke-calculator__blueprint-grid">
                <div class="spoke-calculator__legend-table">
                  <div class="spoke-calculator__legend-row">
                    <div class="spoke-calculator__legend-badge">4</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>{{ t('resourcesSpokeCalculator.calculator.schematic.rimOffset.label') }}</strong>
                      <span>{{ t('resourcesSpokeCalculator.calculator.schematic.rimOffset.description') }}</span>
                    </div>
                    <div class="spoke-calculator__legend-control">
                      <div class="spoke-calculator__unit-field">
                        <input
                          id="front-blueprint-rim-offset"
                          v-model.number="frontConfig.rimOffsetMm"
                          type="number"
                          min="-20"
                          max="20"
                          step="0.1"
                          placeholder="0"
                          class="spoke-calculator__control spoke-calculator__control--with-unit"
                        />
                        <span class="spoke-calculator__unit">{{ t('resourcesSpokeCalculator.calculator.results.unit') }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <SpokeCalculatorBuildSettings
              side="front"
              :config="frontConfig"
              :spoke-count-options="spokeCountOptions"
              :lacing-options="lacingOptions"
              :nipple-type-options="nippleTypeOptions"
              :rim-brand-options="rimBrandOptions"
              :rim-model-options="frontRimModelOptions"
              :hub-brand-options="hubBrandOptions"
              :hub-model-options="frontHubModelOptions"
              :spoke-head-type-options="spokeHeadTypeOptions"
              :spoke-profile-options="spokeProfileOptions"
              :interlacing-options="interlacingOptions"
            />

          </div>

          <!-- ========== REAR WHEEL COLUMN ========== -->
          <div class="spoke-calculator__panel space-y-4">
            <h3 class="text-sm font-semibold text-[var(--tz-site-accent)] uppercase tracking-wide">
              {{ t('resourcesSpokeCalculator.calculator.rearWheel') }}
            </h3>

            <div class="spoke-calculator__blueprint-sheet">
              <div class="spoke-calculator__blueprint-grid">
                <div class="spoke-calculator__legend-table">
                  <div class="spoke-calculator__legend-row">
                    <div class="spoke-calculator__legend-badge">4</div>
                    <div class="spoke-calculator__legend-heading">
                      <strong>{{ t('resourcesSpokeCalculator.calculator.schematic.rimOffset.label') }}</strong>
                      <span>{{ t('resourcesSpokeCalculator.calculator.schematic.rimOffset.description') }}</span>
                    </div>
                    <div class="spoke-calculator__legend-control">
                      <div class="spoke-calculator__unit-field">
                        <input
                          id="rear-blueprint-rim-offset"
                          v-model.number="rearConfig.rimOffsetMm"
                          type="number"
                          min="-20"
                          max="20"
                          step="0.1"
                          placeholder="0"
                          class="spoke-calculator__control spoke-calculator__control--with-unit"
                        />
                        <span class="spoke-calculator__unit">{{ t('resourcesSpokeCalculator.calculator.results.unit') }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <SpokeCalculatorBuildSettings
              side="rear"
              :config="rearConfig"
              :spoke-count-options="spokeCountOptions"
              :lacing-options="lacingOptions"
              :nipple-type-options="nippleTypeOptions"
              :rim-brand-options="rimBrandOptions"
              :rim-model-options="rearRimModelOptions"
              :hub-brand-options="hubBrandOptions"
              :hub-model-options="rearHubModelOptions"
              :spoke-head-type-options="spokeHeadTypeOptions"
              :spoke-profile-options="spokeProfileOptions"
              :interlacing-options="interlacingOptions"
            />
        </div>

        </div>

        <!-- Action row -->
        <div class="mt-6 flex flex-col gap-3 md:flex-row md:items-center md:justify-between border-t tz-border-subtle pt-4">
          <p class="tz-description tz-text-muted max-w-md">
            {{ t('resourcesSpokeCalculator.calculator.action.description') }}
          </p>
          <div class="flex items-center gap-3">
            <button
              type="button"
              class="inline-flex items-center rounded-full bg-[var(--tz-action-primary)] px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-[var(--tz-action-primary-hover)] focus:outline-none focus:ring-2 focus:ring-[color:var(--tz-site-accent)] focus:ring-offset-2 focus:ring-offset-[var(--tz-card-surface)] disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="loading"
              @click="onCalculate"
            >
              <span v-if="loading">{{ t('resourcesSpokeCalculator.calculator.action.calculating') }}</span>
              <span v-else>{{ t('resourcesSpokeCalculator.calculator.action.recalculate') }}</span>
            </button>
            <p v-if="error" class="tz-caption text-rose-400">{{ error }}</p>
          </div>
        </div>

        <SpokeCalculatorResults
          :front-result="frontResult"
          :rear-result="rearResult"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import SpokeCalculatorBuildSettings from '~/components/SpokeCalculatorBuildSettings.vue'
import SpokeCalculatorResults from '~/components/SpokeCalculatorResults.vue'
import type { HubGeometry, HubModel, RimModel } from '~/data/spoke-calculator/database'
import { useSpokeCalculator } from '~/composables/useSpokeCalculator'
import type { SpokeHeadType, SpokeWheelBuildConfig, SpokeWheelResult, SpokeWheelSide } from '~/types/spokeCalculator'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { useSpokeCalculatorCatalog } from '~/composables/useSpokeCalculatorCatalog'
import { useI18n } from '#imports'

const props = defineProps<{
  frontErd?: number | null
  rearErd?: number | null
  frontGeometry?: HubGeometry | null
  rearGeometry?: HubGeometry | null
  frontSpokeHeadType?: SpokeHeadType
  rearSpokeHeadType?: SpokeHeadType
}>()

const emit = defineEmits<{
  'update:frontErd': [value: number | null]
  'update:rearErd': [value: number | null]
  'update:frontGeometry': [value: HubGeometry]
  'update:rearGeometry': [value: HubGeometry]
  'update:frontSpokeHeadType': [value: SpokeHeadType]
  'update:rearSpokeHeadType': [value: SpokeHeadType]
}>()

// Front wheel configuration
const frontConfig = reactive<SpokeWheelBuildConfig>({
  spokeCount: 32,
  crossing: 3,
  nippleType: 'standard',
  nippleLength: 12,
	spokeHeadType: 'j_bend',
	spokeHoleDiameterMm: 2.5,
	straightPullTangentOffsetMm: 0.8,
	spokeProfile: 'round_2_0',
	targetTensionN: 0,
	alternatingDrillingOffsetMm: 0,
	interlacing: 'off',
	interlaceCompensationMm: 0.45,
  rimBrandId: null,
  rimModelId: null,
  hubBrandId: null,
  hubModelId: null,
  erd: null,
  rimOffsetMm: 0,
  leftFlange: null,
  rightFlange: null,
  leftFlangePcd: null,
  rightFlangePcd: null,
})

// Rear wheel configuration
const rearConfig = reactive<SpokeWheelBuildConfig>({
  spokeCount: 32,
  crossing: 3,
  nippleType: 'standard',
  nippleLength: 12,
	spokeHeadType: 'j_bend',
	spokeHoleDiameterMm: 2.5,
	straightPullTangentOffsetMm: 0.8,
	spokeProfile: 'round_2_0',
	targetTensionN: 0,
	alternatingDrillingOffsetMm: 0,
	interlacing: 'off',
	interlaceCompensationMm: 0.45,
  rimBrandId: null,
  rimModelId: null,
  hubBrandId: null,
  hubModelId: null,
  erd: null,
  rimOffsetMm: 0,
  leftFlange: null,
  rightFlange: null,
  leftFlangePcd: null,
  rightFlangePcd: null,
})

const { t } = useI18n()
const { rims, hubs, options: catalogOptions } = useSpokeCalculatorCatalog()
const { calculateWheel } = useSpokeCalculator()

const spokeCountOptions = computed(() => catalogOptions.value.spokeCounts)
const crossingTranslationKeys: Record<number, string> = {
  0: 'radial',
  1: 'one',
  2: 'two',
  3: 'three',
  4: 'four',
}
const lacingOptions = computed(() => catalogOptions.value.crossings.map(option => ({
  ...option,
  label: t(
    `resourcesSpokeCalculator.calculator.options.crossing.${crossingTranslationKeys[option.value] || option.value}`,
    option.label,
  ),
})))
const nippleTypeOptions = computed(() => catalogOptions.value.nippleTypes.map(option => ({
  ...option,
  label: t(
    `resourcesSpokeCalculator.calculator.options.nippleType.${option.value}`,
    option.label,
  ),
})))

const spokeHeadTypeOptions = computed(() => [
  {
    value: 'j_bend',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.jBend'),
  },
  {
    value: 'straight_pull',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.straightPull'),
  },
])

const spokeProfileOptions = computed(() => [
  {
    value: 'round_2_0',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.round20'),
  },
  {
    value: 'round_1_8',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.round18'),
  },
  {
    value: 'bladed_0_9x2_2',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.bladed0922'),
  },
])

const interlacingOptions = computed(() => [
  {
    value: 'off',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.off'),
  },
  {
    value: 'on',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.on'),
  },
])

const rimBrandOptions = computed(() => rims.value.map(brand => ({
  label: brand.name,
  value: brand.id,
})))

const hubBrandOptions = computed(() => hubs.value.map(brand => ({
  label: brand.name,
  value: brand.id,
})))

// --- Computed Models based on Brand Selection ---

// Front Rim Models
const frontRimModels = computed<RimModel[]>(() => {
  if (!frontConfig.rimBrandId) return []
  const brand = rims.value.find(b => b.id === frontConfig.rimBrandId)
  return brand ? brand.items : []
})

const frontRimModelOptions = computed(() => frontRimModels.value.map(rim => ({
  label: rim.name,
  value: rim.id,
})))

// Front Hub Models
const frontHubModels = computed<HubModel[]>(() => {
  if (!frontConfig.hubBrandId) return []
  const brand = hubs.value.find(b => b.id === frontConfig.hubBrandId)
  return brand ? brand.items : []
})

const frontHubModelOptions = computed(() => frontHubModels.value.map(hub => ({
  label: hub.name,
  value: hub.id,
})))

// Rear Rim Models
const rearRimModels = computed<RimModel[]>(() => {
  if (!rearConfig.rimBrandId) return []
  const brand = rims.value.find(b => b.id === rearConfig.rimBrandId)
  return brand ? brand.items : []
})

const rearRimModelOptions = computed(() => rearRimModels.value.map(rim => ({
  label: rim.name,
  value: rim.id,
})))

// Rear Hub Models
const rearHubModels = computed<HubModel[]>(() => {
  if (!rearConfig.hubBrandId) return []
  const brand = hubs.value.find(b => b.id === rearConfig.hubBrandId)
  return brand ? brand.items : []
})

const rearHubModelOptions = computed(() => rearHubModels.value.map(hub => ({
  label: hub.name,
  value: hub.id,
})))

const applyHubGeometry = (config: SpokeWheelBuildConfig, geometry?: HubGeometry | null) => {
  if (!geometry) return
  config.leftFlange = geometry?.leftFlange ?? null
  config.rightFlange = geometry?.rightFlange ?? null
  config.leftFlangePcd = geometry?.leftFlangePcd ?? null
  config.rightFlangePcd = geometry?.rightFlangePcd ?? null
	config.spokeHoleDiameterMm = geometry?.spokeHoleDiameter ?? config.spokeHoleDiameterMm
}

// Keep the wizard's front and rear ERD values synchronized with the matching
// calculator configs. These channels stay separate so one wheel can never
// overwrite the other wheel's ERD.
watch(
  () => props.frontErd,
  (value) => {
    if (value !== undefined && frontConfig.erd !== value) {
      frontConfig.erd = value
    }
  },
  { immediate: true },
)

watch(
  () => props.rearErd,
  (value) => {
    if (value !== undefined && rearConfig.erd !== value) {
      rearConfig.erd = value
    }
  },
  { immediate: true },
)

watch(
  () => frontConfig.erd,
  (value) => {
    if (props.frontErd !== undefined && props.frontErd !== value) {
      emit('update:frontErd', value)
    }
  },
)

watch(
  () => rearConfig.erd,
  (value) => {
    if (props.rearErd !== undefined && props.rearErd !== value) {
      emit('update:rearErd', value)
    }
  },
)

type FlangeGeometryKey = 'leftFlange' | 'rightFlange' | 'leftFlangePcd' | 'rightFlangePcd'

const flangeGeometryKeys: FlangeGeometryKey[] = [
  'leftFlange',
  'rightFlange',
  'leftFlangePcd',
  'rightFlangePcd',
]

const geometryFromConfig = (config: SpokeWheelBuildConfig): HubGeometry => ({
  leftFlange: config.leftFlange,
  rightFlange: config.rightFlange,
  leftFlangePcd: config.leftFlangePcd,
  rightFlangePcd: config.rightFlangePcd,
})

const geometryMatches = (current: HubGeometry | null | undefined, next: HubGeometry) => (
  Boolean(current)
  && flangeGeometryKeys.every(key => current?.[key] === next[key])
)

const applyExternalGeometry = (config: SpokeWheelBuildConfig, geometry?: HubGeometry | null) => {
  if (!geometry) return
  for (const key of flangeGeometryKeys) {
    if (config[key] !== geometry[key]) {
      config[key] = geometry[key]
    }
  }
}

watch(
  () => props.frontGeometry,
  geometry => applyExternalGeometry(frontConfig, geometry),
  { immediate: true, deep: true },
)

watch(
  () => props.rearGeometry,
  geometry => applyExternalGeometry(rearConfig, geometry),
  { immediate: true, deep: true },
)

watch(
  () => props.frontSpokeHeadType,
  value => {
    if (value && frontConfig.spokeHeadType !== value) {
      frontConfig.spokeHeadType = value
    }
  },
  { immediate: true },
)

watch(
  () => props.rearSpokeHeadType,
  value => {
    if (value && rearConfig.spokeHeadType !== value) {
      rearConfig.spokeHeadType = value
    }
  },
  { immediate: true },
)

watch(
  () => frontConfig.spokeHeadType,
  value => {
    if (props.frontSpokeHeadType !== undefined && props.frontSpokeHeadType !== value) {
      emit('update:frontSpokeHeadType', value)
    }
  },
)

watch(
  () => rearConfig.spokeHeadType,
  value => {
    if (props.rearSpokeHeadType !== undefined && props.rearSpokeHeadType !== value) {
      emit('update:rearSpokeHeadType', value)
    }
  },
)

watch(
  () => flangeGeometryKeys.map(key => frontConfig[key]),
  () => {
    const next = geometryFromConfig(frontConfig)
    if (!geometryMatches(props.frontGeometry, next)) {
      emit('update:frontGeometry', next)
    }
  },
)

watch(
  () => flangeGeometryKeys.map(key => rearConfig[key]),
  () => {
    const next = geometryFromConfig(rearConfig)
    if (!geometryMatches(props.rearGeometry, next)) {
      emit('update:rearGeometry', next)
    }
  },
)

// --- Watchers for Auto-Population ---

// Front Rim Change
watch(
  () => frontConfig.rimModelId,
  (newId) => {
    if (!newId) {
      frontConfig.erd = null
      return
    }
    const model = frontRimModels.value.find(m => m.id === newId)
    if (model && model.erd != null) {
      frontConfig.erd = model.erd
    }
  }
)

// Front Hub Change
watch(
  () => frontConfig.hubModelId,
  (newId) => {
    if (!newId) {
      frontConfig.leftFlange = frontConfig.rightFlange = frontConfig.leftFlangePcd = frontConfig.rightFlangePcd = null
      return
    }
    const model = frontHubModels.value.find(m => m.id === newId)
    applyHubGeometry(frontConfig, model?.front ?? null)
  }
)

// Rear Rim Change
watch(
  () => rearConfig.rimModelId,
  (newId) => {
    if (!newId) {
      rearConfig.erd = null
      return
    }
    const model = rearRimModels.value.find(m => m.id === newId)
    if (model && model.erd != null) {
      rearConfig.erd = model.erd
    }
  }
)

// Rear Hub Change
watch(
  () => rearConfig.hubModelId,
  (newId) => {
    if (!newId) {
      rearConfig.leftFlange = rearConfig.rightFlange = rearConfig.leftFlangePcd = rearConfig.rightFlangePcd = null
      return
    }
    const model = rearHubModels.value.find(m => m.id === newId)
    applyHubGeometry(rearConfig, model?.rear ?? null)
  }
)

const loading = ref(false)
const error = ref<string | null>(null)

const frontResult = ref<SpokeWheelResult | null>(null)
const rearResult = ref<SpokeWheelResult | null>(null)
const lastTrackedCalculation = ref('')
const { track: trackBehaviorEvent } = useBehaviorEvents()

const onCalculate = async () => {
  error.value = null
  loading.value = true

  try {
    const completedWheelCount = await updateResults()

    if (completedWheelCount > 0) {
      const fingerprint = JSON.stringify({
        front: frontConfig,
        rear: rearConfig,
      })

      if (fingerprint !== lastTrackedCalculation.value) {
        lastTrackedCalculation.value = fingerprint
        trackBehaviorEvent({
          eventType: 'calculator_use',
          metadata: {
            source: 'spoke_calculator',
            wheel_count: completedWheelCount,
            front_spoke_count: frontConfig.spokeCount,
            rear_spoke_count: rearConfig.spokeCount,
            front_crossing: frontConfig.crossing,
            rear_crossing: rearConfig.crossing,
            front_rim_offset_mm: frontConfig.rimOffsetMm,
            rear_rim_offset_mm: rearConfig.rimOffsetMm,
            front_rim_selected: Boolean(frontConfig.rimModelId),
            rear_rim_selected: Boolean(rearConfig.rimModelId),
            front_hub_selected: Boolean(frontConfig.hubModelId),
            rear_hub_selected: Boolean(rearConfig.hubModelId),
          },
        })
      }
    }
  } catch (e: any) {
    error.value = e?.message || t('resourcesSpokeCalculator.calculator.action.calculationFailed')
  } finally {
    loading.value = false
  }
}

const updateResults = async () => {
  const [front, rear] = await Promise.all([
    buildWheelResult(frontConfig, 'front'),
    buildWheelResult(rearConfig, 'rear'),
  ])
  frontResult.value = front
  rearResult.value = rear

  return [frontResult.value, rearResult.value].filter(result => (
    result && (result.leftLengthMm != null || result.rightLengthMm != null)
  )).length
}

const buildWheelResult = async (config: SpokeWheelBuildConfig, wheel: SpokeWheelSide): Promise<SpokeWheelResult | null> => {
  let calculated: Awaited<ReturnType<typeof calculateWheel>>
  try {
    calculated = await calculateWheel(
      config,
      wheel,
      t('resourcesSpokeCalculator.calculator.action.calculationFailed'),
    )
  } catch (requestError: unknown) {
    error.value = requestError instanceof Error
      ? requestError.message
      : t('resourcesSpokeCalculator.calculator.action.calculationFailed')
    return null
  }

  const leftLengthMm = calculated?.leftLengthMm ?? null
  const rightLengthMm = calculated?.rightLengthMm ?? null

  if (leftLengthMm == null && rightLengthMm == null) return null

  return {
    leftLengthMm,
    rightLengthMm,
    tensionRatio: calculated?.tensionRatio ?? null,
    leftSource: calculated?.leftLengthMm != null ? 'calculated' : null,
    rightSource: calculated?.rightLengthMm != null ? 'calculated' : null,
  }
}

watch(
  () => ({
    front: { ...frontConfig },
    rear: { ...rearConfig },
  }),
  () => {
    // Results are generated explicitly by the Calculate action. Avoid firing
    // an API request for every slider/input keystroke (and wasting the quota).
    frontResult.value = null
    rearResult.value = null
  },
  { deep: true }
)
</script>

<style scoped>
.spoke-calculator {
  --spoke-shell-surface: var(--tz-card-surface);
  --spoke-panel-surface: var(--tz-form-panel-surface);
  --spoke-control-surface: var(--tz-input-surface);
  --spoke-result-surface: var(--tz-surface-subtle);
  --spoke-border: var(--tz-border-subtle);
  --spoke-border-strong: var(--tz-border-strong);
  --spoke-focus-ring: var(--tz-form-control-focus-ring);
  color: var(--tz-text-primary);
}

.spoke-calculator__shell {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-shell-surface);
  box-shadow: 0 10px 26px -14px rgba(20, 32, 43, 0.12);
}

.spoke-calculator__shell {
  padding: 1.25rem;
}

.spoke-calculator__panel {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-panel-surface);
  padding: 1rem;
}

.spoke-calculator__control {
  display: block;
  width: 100%;
  min-width: 0;
  border: 1px solid var(--spoke-border) !important;
  border-radius: 0.5rem;
  background-color: var(--spoke-control-surface) !important;
  background-image: none !important;
  color: var(--tz-text-primary) !important;
  padding: 0.75rem 0.875rem;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease;
}

.spoke-calculator__control:focus,
.spoke-calculator__control:focus-visible {
  outline: none;
  border-color: var(--spoke-border-strong) !important;
  box-shadow: 0 0 0 1px var(--spoke-focus-ring) !important;
}

.spoke-calculator__control:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.spoke-calculator__control--with-unit {
  flex: 1 1 auto;
  border: 0 !important;
  border-radius: 0;
  background: transparent !important;
  box-shadow: none !important;
  padding-right: 0.75rem;
}

.spoke-calculator__unit-field {
  display: flex;
  width: 100%;
  align-items: stretch;
  gap: 0;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background-color: var(--spoke-control-surface);
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease;
}

.spoke-calculator__unit-field:focus-within {
  border-color: var(--spoke-border-strong);
  box-shadow: 0 0 0 1px var(--spoke-focus-ring);
}

.spoke-calculator__unit-field .spoke-calculator__control:focus,
.spoke-calculator__unit-field .spoke-calculator__control:focus-visible {
  box-shadow: none !important;
}

.spoke-calculator__unit {
  display: inline-flex;
  flex: 0 0 auto;
  min-width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-left: 1px solid var(--spoke-border);
  padding: 0 0.75rem;
  color: var(--tz-text-muted);
  font-size: var(--tz-type-caption);
  line-height: 1;
  white-space: nowrap;
}

.spoke-calculator label {
  color: var(--tz-text-secondary) !important;
}

.spoke-calculator__blueprint-sheet {
  display: grid;
  gap: 0.75rem;
  padding: 0.875rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.625rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(248, 250, 252, 0.96)),
    var(--spoke-shell-surface);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.85),
    0 12px 28px rgba(20, 32, 43, 0.08);
}

.spoke-calculator__blueprint-grid {
  display: grid;
  gap: 0.85rem;
}

.spoke-calculator__legend-table {
  display: grid;
  gap: 0.55rem;
  padding: 0.75rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.625rem;
  background: rgba(255, 255, 255, 0.7);
}

.spoke-calculator__legend-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) minmax(0, 1.55fr);
  gap: 0.75rem;
  align-items: start;
  padding: 0.65rem 0.7rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-panel-surface);
}

.spoke-calculator__legend-row--split {
  align-items: center;
}

.spoke-calculator__legend-badge {
  display: inline-flex;
  width: 1.8rem;
  height: 1.8rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(5, 150, 105, 0.28);
  border-radius: 50%;
  background: rgba(5, 150, 105, 0.08);
  color: var(--tz-text-accent);
  font-size: 0.8rem;
  font-weight: 800;
  line-height: 1;
}

.spoke-calculator__legend-heading {
  display: grid;
  min-width: 0;
  gap: 0.08rem;
}

.spoke-calculator__legend-heading strong {
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 750;
  line-height: 1.2;
}

.spoke-calculator__legend-heading span {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  line-height: 1.35;
}

.spoke-calculator__legend-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
  min-width: 0;
}

.spoke-calculator__legend-field {
  display: grid;
  gap: 0.28rem;
  color: var(--tz-text-secondary);
  font-size: 0.65rem;
  font-weight: 700;
}

.spoke-calculator__legend-field > span {
  color: var(--tz-text-secondary);
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.spoke-calculator__legend-control {
  min-width: 0;
}

@media (max-width: 767px) {
  .spoke-calculator__panel {
    padding: 0.5rem 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .spoke-calculator__panel + .spoke-calculator__panel {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--spoke-border);
  }

  .spoke-calculator__blueprint-sheet {
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .spoke-calculator__legend-table {
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .spoke-calculator__legend-row {
    grid-template-columns: 1fr;
    padding: 0.4rem;
    gap: 0.4rem;
  }

  .spoke-calculator__legend-table {
    gap: 0.45rem;
  }

  .spoke-calculator__legend-fields {
    grid-template-columns: 1fr;
    gap: 0.4rem;
  }

  .spoke-wheel {
    gap: 0.25rem;
  }

  .spoke-wheel__header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.1rem;
  }

  .spoke-wheel__title {
    text-align: left;
  }
}
</style>
