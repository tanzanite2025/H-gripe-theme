<template>
  <section class="tire-rim-engineering-workbench" aria-labelledby="tire-rim-engineering-workbench-title">
    <header class="tire-rim-engineering-workbench__header">
      <div>
        <p class="tire-rim-engineering-workbench__eyebrow">
          {{ t('guidesTireChoose.engineering.eyebrow') }}
        </p>
        <h3 id="tire-rim-engineering-workbench-title" class="tire-rim-engineering-workbench__title">
          {{ t('guidesTireChoose.engineering.title') }}
        </h3>
        <p class="tire-rim-engineering-workbench__description">
          {{ t('guidesTireChoose.engineering.description') }}
        </p>
      </div>
      <span class="tire-rim-engineering-workbench__version">
        {{ t('guidesTireChoose.engineering.sourceBadge') }}
      </span>
    </header>

    <div class="tire-rim-engineering-workbench__controls">
      <fieldset class="tire-rim-engineering-workbench__field">
        <legend>{{ t('guidesTireChoose.engineering.rimSystemLabel') }}</legend>
        <div class="tire-rim-engineering-workbench__segmented-control">
          <button
            v-for="systemOption in rimSystemOptions"
            :key="systemOption"
            type="button"
            class="tire-rim-engineering-workbench__segment"
            :class="{ 'tire-rim-engineering-workbench__segment--active': rimSystem === systemOption }"
            :aria-pressed="rimSystem === systemOption"
            @click="rimSystem = systemOption"
          >
            {{ t(`guidesTireChoose.engineering.rimSystems.${systemOption}`) }}
          </button>
        </div>
      </fieldset>

      <fieldset class="tire-rim-engineering-workbench__field">
        <legend id="simulation-rim-inner-width-label">{{ t('guidesTireChoose.engineering.rimInnerWidthLabel') }}</legend>
        <button
          type="button"
          class="tire-rim-engineering-workbench__picker-trigger"
          aria-labelledby="simulation-rim-inner-width-label"
          aria-haspopup="dialog"
          :aria-expanded="rimInnerWidthPickerOpen"
          @click="rimInnerWidthPickerOpen = true"
        >
          <span>{{ rimInnerWidthMm }} mm</span>
          <Icon name="lucide:chevron-down" class="h-4 w-4 shrink-0" aria-hidden="true" />
        </button>
      </fieldset>

      <fieldset class="tire-rim-engineering-workbench__field">
        <legend id="simulation-tire-width-label">{{ t('guidesTireChoose.engineering.tireWidthPresetsLabel') }}</legend>
        <button
          type="button"
          class="tire-rim-engineering-workbench__picker-trigger"
          aria-labelledby="simulation-tire-width-label"
          aria-haspopup="dialog"
          :aria-expanded="tireWidthPickerOpen"
          @click="tireWidthPickerOpen = true"
        >
          <span class="min-w-0 truncate">{{ selectedSimulationTireWidthSummary }}</span>
          <Icon name="lucide:chevron-down" class="h-4 w-4 shrink-0" aria-hidden="true" />
        </button>
      </fieldset>
    </div>

    <TireRimWidthSelectionDialog
      v-model="rimInnerWidthPickerOpen"
      kind="rim"
      :selected-rim-inner-width-mm="rimInnerWidthMm"
      :description="t('guidesTireChoose.engineering.rimWidthPickerDescription')"
      @select-rim-inner-width="rimInnerWidthMm = $event"
    />
    <TireRimWidthSelectionDialog
      v-model="tireWidthPickerOpen"
      kind="tire"
      :selected-tire-width-preset-id="selectedSimulationTireWidthPresetId"
      :selected-tire-width-mm="selectedTireWidthMm"
      @select-tire-width-preset="selectSimulationTireWidthPreset"
      @select-tire-width-millimeters="selectSimulationTireWidthMillimeters"
    />

    <div class="tire-rim-engineering-workbench__result-layout">
      <section
        v-if="engineeringReferenceResult && !recommendationPending"
        class="tire-rim-engineering-workbench__diagnostic"
        :class="`tire-rim-engineering-workbench__diagnostic--${engineeringReferenceResult.verdict}`"
        aria-live="polite"
      >
        <span class="tire-rim-engineering-workbench__diagnostic-badge">
          {{ t(`guidesTireChoose.engineering.verdicts.${engineeringReferenceResult.verdict}.badge`) }}
        </span>
        <h4>{{ t(`guidesTireChoose.engineering.verdicts.${engineeringReferenceResult.verdict}.title`) }}</h4>
        <p>{{ t('guidesTireRimHelper.layered.pairing', {
          rimWidth: engineeringReferenceResult.rimInnerWidthMm,
          tireWidth: engineeringReferenceResult.tireWidthMm,
        }) }}</p>
        <p>{{ simulationPairingExplanation }}</p>
        <dl class="tire-rim-engineering-workbench__metrics">
          <div>
            <dt>{{ t('guidesTireChoose.engineering.metrics.allowedRange') }}</dt>
            <dd>{{ engineeringReferenceResult.allowedTireWidthRange.min }}–{{ engineeringReferenceResult.allowedTireWidthRange.max }} mm</dd>
          </div>
          <div v-if="engineeringReferenceResult.recommendedTireWidthRange">
            <dt>{{ t('guidesTireChoose.engineering.metrics.recommendedRange') }}</dt>
            <dd>{{ engineeringReferenceResult.recommendedTireWidthRange.min }}–{{ engineeringReferenceResult.recommendedTireWidthRange.max }} mm</dd>
          </div>
        </dl>
        <p class="tire-rim-engineering-workbench__diagnostic-note">{{ t('guidesTireChoose.engineering.diagramNote') }}</p>
      </section>
      <section v-else class="tire-rim-engineering-workbench__diagnostic tire-rim-engineering-workbench__diagnostic--empty">
        <h4>{{ t(recommendationPending ? 'guidesTireChoose.engineering.loadingTitle' : activeRecommendationError ? 'guidesTireRimHelper.requestError' : 'guidesTireChoose.engineering.emptyTitle') }}</h4>
        <p :role="activeRecommendationError || activeTireWidthOutOfRange ? 'alert' : undefined">{{ t(activeTireWidthOutOfRange ? 'guidesTireRimHelper.outOfRange' : activeRecommendationError ? 'guidesTireRimHelper.requestError' : recommendationPending ? 'guidesTireChoose.engineering.loadingDescription' : 'guidesTireChoose.engineering.emptyDescription') }}</p>
      </section>

      <figure v-if="diagramHasConfirmedInputPairing && !recommendationPending" class="tire-rim-engineering-workbench__diagram">
        <figcaption>{{ t('guidesTireChoose.engineering.diagramTitle') }}</figcaption>
        <div class="tire-rim-engineering-workbench__diagram-stage">
          <div v-if="engineeringReferenceResult?.aeroRuleApplies" class="tire-rim-engineering-workbench__diagram-floating-badge">
            {{ t('guidesTireChoose.engineering.diagram.badges.aero') }}
          </div>
          <div
            class="tire-rim-engineering-workbench__diagram-floating-badge tire-rim-engineering-workbench__diagram-floating-badge--status"
            :class="`tire-rim-engineering-workbench__diagram-floating-badge--${diagramGeometry.status}`"
          >
            {{ t(`guidesTireChoose.engineering.diagram.badges.${diagramGeometry.status}`) }}
          </div>

          <svg
            viewBox="0 0 460 280"
            preserveAspectRatio="xMidYMid meet"
            role="img"
            :aria-label="t('guidesTireChoose.engineering.diagramAlt')"
          >
            <defs>
              <pattern id="tire-rim-engineering-grid-pattern" width="20" height="20" patternUnits="userSpaceOnUse">
                <path d="M 20 0 L 0 0 0 20" fill="none" stroke="#e2e8f0" stroke-width="0.8" stroke-dasharray="2,2" />
              </pattern>
              <linearGradient id="tire-rim-engineering-carbon-gradient" x1="0%" x2="100%" y1="0%" y2="100%">
                <stop offset="0%" stop-color="#1e293b" />
                <stop offset="50%" stop-color="#334155" />
                <stop offset="100%" stop-color="#0f172a" />
              </linearGradient>
              <linearGradient id="tire-rim-engineering-rubber-gradient" x1="0%" x2="0%" y1="0%" y2="100%">
                <stop offset="0%" stop-color="#09090b" />
                <stop offset="100%" stop-color="#18181b" />
              </linearGradient>
              <radialGradient id="tire-rim-engineering-kevlar-gradient" cx="40%" cy="40%" r="60%">
                <stop offset="0%" stop-color="#fef08a" />
                <stop offset="60%" stop-color="#eab308" />
                <stop offset="100%" stop-color="#a16207" />
              </radialGradient>
              <filter id="tire-rim-engineering-danger-glow" x="-20%" y="-20%" width="140%" height="140%">
                <feGaussianBlur stdDeviation="3" result="blur" />
                <feComposite in="SourceGraphic" in2="blur" operator="over" />
              </filter>
            </defs>

            <rect x="10" y="10" width="440" height="260" rx="10" class="tire-rim-engineering-workbench__diagram-canvas" />
            <line x1="230" y1="15" x2="230" y2="265" class="tire-rim-engineering-workbench__diagram-centerline" />
            <line x1="30" y1="125" x2="430" y2="125" class="tire-rim-engineering-workbench__diagram-baseline" />

            <path
              :d="diagramGeometry.rimOuterPath"
              fill="url(#tire-rim-engineering-carbon-gradient)"
              class="tire-rim-engineering-workbench__diagram-rim"
            />
            <path
              :d="diagramGeometry.rimCavityPath"
              class="tire-rim-engineering-workbench__diagram-rim-cavity"
            />
            <path
              :d="diagramGeometry.rimTapePath"
              class="tire-rim-engineering-workbench__diagram-rim-tape"
            />
            <line
              :x1="diagramGeometry.valveHole.x1"
              :x2="diagramGeometry.valveHole.x2"
              :y1="diagramGeometry.valveHole.y1"
              :y2="diagramGeometry.valveHole.y2"
              class="tire-rim-engineering-workbench__diagram-valve-hole"
            />

            <path
              :d="diagramGeometry.tireAirChamberPath"
              class="tire-rim-engineering-workbench__diagram-air-chamber"
            />
            <path
              :d="diagramGeometry.tireInnerCasingPath"
              class="tire-rim-engineering-workbench__diagram-inner-casing"
            />
            <path
              :d="diagramGeometry.tireOuterPath"
              fill="url(#tire-rim-engineering-rubber-gradient)"
              :stroke="diagramGeometry.tireStroke"
              class="tire-rim-engineering-workbench__diagram-tire"
            />
            <g class="tire-rim-engineering-workbench__diagram-tire-sipes">
              <line
                v-for="sipe in diagramGeometry.tireSipes"
                :key="sipe.id"
                :x1="sipe.x1"
                :y1="sipe.y1"
                :x2="sipe.x2"
                :y2="sipe.y2"
              />
            </g>

            <g class="tire-rim-engineering-workbench__diagram-bead">
              <path :d="diagramGeometry.beadRubberLeftPath" />
              <circle
                :cx="diagramGeometry.beadCoreLeft.cx"
                :cy="diagramGeometry.beadCoreLeft.cy"
                r="3.5"
                class="tire-rim-engineering-workbench__diagram-kevlar-core"
              />
              <path :d="diagramGeometry.beadRubberRightPath" />
              <circle
                :cx="diagramGeometry.beadCoreRight.cx"
                :cy="diagramGeometry.beadCoreRight.cy"
                r="3.5"
                class="tire-rim-engineering-workbench__diagram-kevlar-core"
              />
            </g>

            <g class="tire-rim-engineering-workbench__diagram-stress-indicators">
              <template v-if="diagramGeometry.status === 'critical'">
                <line
                  :x1="diagramGeometry.rimInnerLeftX"
                  :y1="diagramGeometry.lipY + 10"
                  :x2="diagramGeometry.rimInnerLeftX - 18"
                  :y2="diagramGeometry.lipY - 8"
                />
                <polygon
                  :points="`${diagramGeometry.rimInnerLeftX - 18},${diagramGeometry.lipY - 8} ${diagramGeometry.rimInnerLeftX - 12},${diagramGeometry.lipY - 2} ${diagramGeometry.rimInnerLeftX - 22},${diagramGeometry.lipY - 2}`"
                />
                <text :x="diagramGeometry.rimInnerLeftX - 24" :y="diagramGeometry.lipY - 14" text-anchor="middle">
                  {{ t('guidesTireChoose.engineering.diagram.stress.outward') }}
                </text>
                <line
                  :x1="diagramGeometry.rimInnerRightX"
                  :y1="diagramGeometry.lipY + 10"
                  :x2="diagramGeometry.rimInnerRightX + 18"
                  :y2="diagramGeometry.lipY - 8"
                />
                <polygon
                  :points="`${diagramGeometry.rimInnerRightX + 18},${diagramGeometry.lipY - 8} ${diagramGeometry.rimInnerRightX + 12},${diagramGeometry.lipY - 2} ${diagramGeometry.rimInnerRightX + 22},${diagramGeometry.lipY - 2}`"
                />
                <text :x="diagramGeometry.rimInnerRightX + 24" :y="diagramGeometry.lipY - 14" text-anchor="middle">
                  {{ t('guidesTireChoose.engineering.diagram.stress.outward') }}
                </text>
              </template>
              <template v-else-if="diagramGeometry.status === 'recommended'">
                <line
                  :x1="diagramGeometry.rimInnerLeftX + 5"
                  :y1="diagramGeometry.seatY - 20"
                  :x2="diagramGeometry.rimInnerLeftX + 5"
                  :y2="diagramGeometry.seatY - 3"
                />
                <polygon :points="`${diagramGeometry.rimInnerLeftX + 5},${diagramGeometry.seatY - 1} ${diagramGeometry.rimInnerLeftX + 2},${diagramGeometry.seatY - 7} ${diagramGeometry.rimInnerLeftX + 8},${diagramGeometry.seatY - 7}`" />
                <line
                  :x1="diagramGeometry.rimInnerRightX - 5"
                  :y1="diagramGeometry.seatY - 20"
                  :x2="diagramGeometry.rimInnerRightX - 5"
                  :y2="diagramGeometry.seatY - 3"
                />
                <polygon :points="`${diagramGeometry.rimInnerRightX - 5},${diagramGeometry.seatY - 1} ${diagramGeometry.rimInnerRightX - 8},${diagramGeometry.seatY - 7} ${diagramGeometry.rimInnerRightX - 2},${diagramGeometry.seatY - 7}`" />
                <circle :cx="diagramGeometry.beadCoreLeft.cx" :cy="diagramGeometry.beadCoreLeft.cy" r="6" class="tire-rim-engineering-workbench__diagram-lock-ring" />
                <circle :cx="diagramGeometry.beadCoreRight.cx" :cy="diagramGeometry.beadCoreRight.cy" r="6" class="tire-rim-engineering-workbench__diagram-lock-ring" />
              </template>
              <template v-else>
                <line :x1="diagramGeometry.rimInnerLeftX" :y1="diagramGeometry.seatY - 14" :x2="diagramGeometry.rimInnerLeftX - 8" :y2="diagramGeometry.seatY - 14" />
                <line :x1="diagramGeometry.rimInnerRightX" :y1="diagramGeometry.seatY - 14" :x2="diagramGeometry.rimInnerRightX + 8" :y2="diagramGeometry.seatY - 14" />
              </template>
            </g>

            <g class="tire-rim-engineering-workbench__diagram-caliper">
              <line
                :x1="diagramGeometry.rimInnerLeftX"
                :y1="diagramGeometry.measurementY"
                :x2="diagramGeometry.rimInnerRightX"
                :y2="diagramGeometry.measurementY"
              />
              <polygon :points="`${diagramGeometry.rimInnerLeftX},${diagramGeometry.measurementY} ${diagramGeometry.rimInnerLeftX + 6},${diagramGeometry.measurementY - 3} ${diagramGeometry.rimInnerLeftX + 6},${diagramGeometry.measurementY + 3}`" />
              <polygon :points="`${diagramGeometry.rimInnerRightX},${diagramGeometry.measurementY} ${diagramGeometry.rimInnerRightX - 6},${diagramGeometry.measurementY - 3} ${diagramGeometry.rimInnerRightX - 6},${diagramGeometry.measurementY + 3}`" />
              <rect :x="diagramGeometry.measurementLabelX - 30" :y="diagramGeometry.measurementY - 8" width="60" height="16" rx="4" />
              <text :x="diagramGeometry.measurementLabelX" :y="diagramGeometry.measurementY + 4" text-anchor="middle">
                {{ rimInnerWidthMm.toFixed(1) }} mm
              </text>
            </g>

            <g class="tire-rim-engineering-workbench__diagram-callout tire-rim-engineering-workbench__diagram-callout--lip">
              <path :d="diagramGeometry.lipCallout.linePath" />
              <circle :cx="diagramGeometry.lipCallout.dotX" :cy="diagramGeometry.lipCallout.dotY" r="2.5" :fill="diagramGeometry.lipCallout.dotColor" />
              <text x="35" :y="diagramGeometry.lipCallout.titleY">{{ diagramGeometry.lipCallout.title }}</text>
              <text x="35" :y="diagramGeometry.lipCallout.descriptionY" class="tire-rim-engineering-workbench__diagram-callout-sub">{{ diagramGeometry.lipCallout.description }}</text>
            </g>
            <g class="tire-rim-engineering-workbench__diagram-callout tire-rim-engineering-workbench__diagram-callout--bead">
              <path :d="diagramGeometry.beadCallout.linePath" />
              <circle :cx="diagramGeometry.beadCallout.dotX" :cy="diagramGeometry.beadCallout.dotY" r="2.5" fill="#eab308" />
              <text x="420" :y="diagramGeometry.beadCallout.titleY" text-anchor="end">{{ t('guidesTireChoose.engineering.diagram.beadCore.title') }}</text>
              <text x="420" :y="diagramGeometry.beadCallout.descriptionY" text-anchor="end" class="tire-rim-engineering-workbench__diagram-callout-sub">{{ t('guidesTireChoose.engineering.diagram.beadCore.description') }}</text>
            </g>

            <text x="230" y="32" text-anchor="middle" :fill="diagramGeometry.tireStroke" class="tire-rim-engineering-workbench__diagram-tire-label">
              {{ diagramGeometry.tireLabel }}
            </text>
          </svg>
        </div>
        <p class="tire-rim-engineering-workbench__diagram-note">
          {{ diagramGeometry.footerNote }}
        </p>
      </figure>
    </div>

    <section class="tire-rim-engineering-workbench__board-metrics" aria-label="Engineering safety metrics">
      <article class="tire-rim-engineering-workbench__board-metric tire-rim-engineering-workbench__board-metric--pressure">
        <div>
          <span class="tire-rim-engineering-workbench__board-metric-label">
            {{ t('guidesTireChoose.engineering.board.metrics.pressureLabel') }}
          </span>
          <strong>{{ maxPressure ? `${maxPressure.bar.toFixed(1)} bar` : '—' }}</strong>
        </div>
        <p>{{ maxPressure ? `${maxPressure.psi.toFixed(1)} PSI` : '—' }} · {{ t(maxPressure ? 'guidesTireChoose.engineering.board.metrics.pressureNote' : 'guidesTireRimHelper.layered.noPressureNote') }}</p>
      </article>
      <article class="tire-rim-engineering-workbench__board-metric">
        <div>
          <span class="tire-rim-engineering-workbench__board-metric-label">
            {{ t('guidesTireChoose.engineering.board.metrics.tubeLabel') }}
          </span>
          <strong>{{ tubeRecommendation.model }}</strong>
        </div>
        <p>{{ tubeRecommendation.range }}</p>
      </article>
      <article class="tire-rim-engineering-workbench__board-metric">
        <div>
          <span class="tire-rim-engineering-workbench__board-metric-label">
            {{ t('guidesTireChoose.engineering.board.metrics.certificationLabel') }}
          </span>
          <strong>{{ t(`guidesTireChoose.engineering.board.metrics.certification.${rimSystem}`) }}</strong>
        </div>
        <p>{{ t(`guidesTireChoose.engineering.board.metrics.certificationNote.${rimSystem}`) }}</p>
      </article>
    </section>

    <section class="tire-rim-engineering-workbench__examples" aria-labelledby="tire-rim-engineering-examples-title">
      <div>
        <h4 id="tire-rim-engineering-examples-title">
          {{ t('guidesTireChoose.engineering.examples.title') }}
        </h4>
        <p>{{ t('guidesTireChoose.engineering.examples.description') }}</p>
      </div>
      <div class="tire-rim-engineering-workbench__examples-grid">
        <article
          v-for="exampleKey in fitmentExampleKeys"
          :key="exampleKey"
          class="tire-rim-engineering-workbench__example"
          :class="`tire-rim-engineering-workbench__example--${exampleKey}`"
        >
          <span class="tire-rim-engineering-workbench__example-badge">
            {{ t(`guidesTireChoose.engineering.verdicts.${exampleKey}.badge`) }}
          </span>
          <h5>{{ t(`guidesTireChoose.engineering.examples.${exampleKey}.title`) }}</h5>
          <p>{{ t(`guidesTireChoose.engineering.examples.${exampleKey}.description`) }}</p>
        </article>
      </div>
    </section>

    <div class="tire-rim-engineering-workbench__explanations">
      <article v-for="explanationKey in explanationKeys" :key="explanationKey">
        <h4>{{ t(`guidesTireChoose.engineering.explanations.${explanationKey}.title`) }}</h4>
        <p>{{ t(`guidesTireChoose.engineering.explanations.${explanationKey}.body`) }}</p>
      </article>
    </div>

  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import { useTireRimWidthReferenceRecommendation } from '~/composables/useTireRimWidthReferenceRecommendation'
import type { TireRimEngineeringCalculation } from '~/data/tireguides/tireRimWidthReferencePresentation'
import type { TireRimEngineeringRimSystem } from '~/data/tireguides/tireRimEngineeringReferenceUiOptions'
import { commonTireWidthInputPresets, type TireWidthInputPreset } from '~/data/tireguides/tireRimWidthInputConversion'
import TireRimWidthSelectionDialog from '~/components/tireguides/TireRimWidthSelectionDialog.vue'

interface TireRimEngineeringDiagramGeometry {
  centerX: number
  lipY: number
  seatY: number
  status: 'critical' | 'recommended' | 'reference'
  rimInnerLeftX: number
  rimInnerRightX: number
  rimOuterPath: string
  rimCavityPath: string
  rimTapePath: string
  valveHole: { x1: number; x2: number; y1: number; y2: number }
  tireAirChamberPath: string
  tireInnerCasingPath: string
  tireOuterPath: string
  tireStroke: string
  tireSipes: Array<{ id: string; x1: number; y1: number; x2: number; y2: number }>
  beadRubberLeftPath: string
  beadRubberRightPath: string
  beadCoreLeft: { cx: number; cy: number }
  beadCoreRight: { cx: number; cy: number }
  measurementY: number
  measurementLabelX: number
  lipCallout: {
    linePath: string
    dotX: number
    dotY: number
    dotColor: string
    title: string
    description: string
    titleY: number
    descriptionY: number
  }
  beadCallout: { linePath: string; dotX: number; dotY: number; titleY: number; descriptionY: number }
  tireLabel: string
  footerNote: string
  tireTopY: number
  tireWidth: number
}

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireChoose')
const { loadPageMessages: loadTireRimHelperMessages } = usePageMessages('guidesTireRimHelper')

await Promise.all([loadPageMessages(locale.value), loadTireRimHelperMessages(locale.value)])

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
  void loadTireRimHelperMessages(nextLocale)
})

const rimSystemOptions: readonly TireRimEngineeringRimSystem[] = ['hookless', 'hooked']
const explanationKeys = ['rimSystem', 'nominalWidth', 'aeroProjection'] as const
const fitmentExampleKeys = ['critical', 'reference', 'recommended'] as const

const rimSystem = ref<TireRimEngineeringRimSystem>('hookless')
const rimInnerWidthMm = ref<number>(25)
const selectedTireWidthMm = ref<number>(28)
const rimInnerWidthPickerOpen = ref(false)
const tireWidthPickerOpen = ref(false)
const selectedSimulationTireWidthPresetId = ref<string | undefined>('28c')
const selectedSimulationTireWidthSummary = computed(() => {
  const preset = commonTireWidthInputPresets.find(option => option.id === selectedSimulationTireWidthPresetId.value)
  const tireWidthMm = selectedTireWidthMm.value
  const inches = String(Number((tireWidthMm / 25.4).toFixed(2)))
  return `${preset?.label || `${tireWidthMm} mm`} · ${tireWidthMm} mm · ${inches}″`
})

const selectSimulationTireWidthPreset = (preset: TireWidthInputPreset): void => {
  selectedSimulationTireWidthPresetId.value = preset.id
  selectedTireWidthMm.value = preset.millimeters
}

const selectSimulationTireWidthMillimeters = (widthMm: number): void => {
  selectedSimulationTireWidthPresetId.value = undefined
  selectedTireWidthMm.value = widthMm
}

const selectedTireWidthRequestValue = computed(() => String(selectedTireWidthMm.value))

const {
  tireRimSuggestion: simulatedPairingSuggestion,
  recommendationPending,
  recommendationError: activeRecommendationError,
  tireWidthOutOfRange: activeTireWidthOutOfRange,
} = useTireRimWidthReferenceRecommendation(
  selectedTireWidthRequestValue,
  rimSystem,
  undefined,
  rimInnerWidthMm,
)

const engineeringReferenceResult = computed<TireRimEngineeringCalculation | null>(() => (
  simulatedPairingSuggestion.value?.engineering ?? null
))
const diagramHasConfirmedInputPairing = computed(() => engineeringReferenceResult.value !== null)
const simulationPairingExplanation = computed(() => {
  const result = engineeringReferenceResult.value
  if (!result) return ''
  return t(`guidesTireChoose.engineering.reasons.${result.verdictReason}`, {
    rimWidth: result.rimInnerWidthMm,
    tireWidth: result.tireWidthMm,
    allowedMin: result.allowedTireWidthRange.min,
    allowedMax: result.allowedTireWidthRange.max,
    recommendedMin: result.recommendedTireWidthRange?.min ?? '—',
    recommendedMax: result.recommendedTireWidthRange?.max ?? '—',
  })
})

const maxPressure = computed(() => engineeringReferenceResult.value?.maximumPressure ?? null)

const tubeRecommendationRange = computed(() => {
  const recommendation = engineeringReferenceResult.value?.tubeRecommendation
  if (!recommendation) return '—'
  return recommendation.band === 'fat'
    ? `${recommendation.widthMinMm}–${recommendation.widthMaxMm} mm`
    : `${recommendation.widthMinMm}–${recommendation.widthMaxMm}C`
})

const tubeRecommendation = computed(() => {
  const recommendation = engineeringReferenceResult.value?.tubeRecommendation
  if (!recommendation) return { model: '—', range: '—' }
  const band = recommendation.band
  if (band === 'fat') {
    return { model: t('guidesTireChoose.engineering.board.metrics.tubes.fatModel'), range: tubeRecommendationRange.value }
  }
  if (band === 'wide') {
    return { model: t('guidesTireChoose.engineering.board.metrics.tubes.wideModel'), range: tubeRecommendationRange.value }
  }
  if (band === 'mid') {
    return { model: t('guidesTireChoose.engineering.board.metrics.tubes.midModel'), range: tubeRecommendationRange.value }
  }
  return {
    model: t('guidesTireChoose.engineering.board.metrics.tubes.roadModel'),
    range: tubeRecommendationRange.value,
  }
})

const diagramGeometry = computed<TireRimEngineeringDiagramGeometry>(() => {
  const centerX = 230
  const rimWidth = rimInnerWidthMm.value
  const tireWidth = selectedTireWidthMm.value
  const result = engineeringReferenceResult.value
  const status = result?.verdict ?? 'reference'
  // Keep the illustration inside its fixed 460 px viewBox when the same
  // calculator is used for 65–100 mm snow/fat-bike rims. Road and gravel
  // widths retain the original scale; only the wide envelope scales down.
  const tireToRimWidthDifference = Math.max(0, tireWidth - rimWidth)
  const visualWidthEstimate = tireWidth + tireToRimWidthDifference * 1.2 + 16
  const pixelsPerMillimetre = Math.min(4.2, 400 / visualWidthEstimate)
  const halfRimWidth = (rimWidth / 2) * pixelsPerMillimetre
  const rimInnerLeftX = centerX - halfRimWidth
  const rimInnerRightX = centerX + halfRimWidth

  // These fixed coordinates define only the illustration's layout. Fitment
  // status and all engineering measurements come from the backend response.
  const lipY = 118
  const seatY = 142
  const humpY = 138
  const dropY = 168
  const bottomY = 245
  const isHookless = rimSystem.value === 'hookless'
  const leftLipOuterX = isHookless ? rimInnerLeftX - 12 : rimInnerLeftX - 8.5
  const rightLipOuterX = isHookless ? rimInnerRightX + 12 : rimInnerRightX + 8.5
  const leftHumpX = rimInnerLeftX + 16
  const rightHumpX = rimInnerRightX - 16
  const leftDropX = centerX - 22
  const rightDropX = centerX + 22
  const halfAeroWidth = (rimWidth / 2 + 3) * pixelsPerMillimetre
  const leftAeroX = centerX - halfAeroWidth
  const rightAeroX = centerX + halfAeroWidth
  const noseHalfWidth = 18
  const noseLeftX = centerX - noseHalfWidth
  const noseRightX = centerX + noseHalfWidth
  const rimOuterPath = isHookless
    ? `M ${noseLeftX},${bottomY - 5}
       C ${noseLeftX - 6},225 ${leftAeroX - 2},195 ${leftAeroX},162
       C ${leftAeroX + 1},145 ${leftLipOuterX},130 ${leftLipOuterX},${lipY + 3}
       Q ${leftLipOuterX},${lipY} ${leftLipOuterX + 3},${lipY}
       L ${rimInnerLeftX - 1},${lipY}
       Q ${rimInnerLeftX},${lipY} ${rimInnerLeftX},${lipY + 2}
       L ${rimInnerLeftX},${seatY}
       L ${leftHumpX - 4},${seatY - 1}
       Q ${leftHumpX},${humpY} ${leftHumpX + 4},${seatY + 2}
       L ${leftDropX},${dropY}
       L ${rightDropX},${dropY}
       L ${rightHumpX - 4},${seatY + 2}
       Q ${rightHumpX},${humpY} ${rightHumpX + 4},${seatY - 1}
       L ${rimInnerRightX},${seatY}
       L ${rimInnerRightX},${lipY + 2}
       Q ${rimInnerRightX},${lipY} ${rimInnerRightX + 1},${lipY}
       L ${rightLipOuterX - 3},${lipY}
       Q ${rightLipOuterX},${lipY} ${rightLipOuterX},${lipY + 3}
       C ${rightLipOuterX},130 ${rightAeroX - 1},145 ${rightAeroX},162
       C ${rightAeroX + 2},195 ${noseRightX + 6},225 ${noseRightX},${bottomY - 5}
       Q ${noseRightX - 2},${bottomY} ${centerX},${bottomY}
       Q ${noseLeftX + 2},${bottomY} ${noseLeftX},${bottomY - 5} Z`
    : `M ${noseLeftX},${bottomY - 5}
       C ${noseLeftX - 6},225 ${leftAeroX - 2},195 ${leftAeroX},162
       C ${leftAeroX + 1},145 ${leftLipOuterX},130 ${leftLipOuterX},${lipY + 3}
       Q ${leftLipOuterX},${lipY - 1} ${leftLipOuterX + 3},${lipY - 1}
       L ${rimInnerLeftX - 1},${lipY - 1}
       Q ${rimInnerLeftX + 2},${lipY - 1} ${rimInnerLeftX + 6.5},${lipY + 4}
       Q ${rimInnerLeftX + 7.5},${lipY + 8} ${rimInnerLeftX + 4.5},${lipY + 11}
       Q ${rimInnerLeftX - 1},${lipY + 13} ${rimInnerLeftX},${lipY + 18}
       L ${rimInnerLeftX},${seatY}
       L ${leftHumpX - 4},${seatY - 1}
       Q ${leftHumpX},${humpY} ${leftHumpX + 4},${seatY + 2}
       L ${leftDropX},${dropY}
       L ${rightDropX},${dropY}
       L ${rightHumpX - 4},${seatY + 2}
       Q ${rightHumpX},${humpY} ${rightHumpX + 4},${seatY - 1}
       L ${rimInnerRightX},${seatY}
       L ${rimInnerRightX},${lipY + 18}
       Q ${rimInnerRightX + 1},${lipY + 13} ${rimInnerRightX - 4.5},${lipY + 11}
       Q ${rimInnerRightX - 7.5},${lipY + 8} ${rimInnerRightX - 6.5},${lipY + 4}
       Q ${rimInnerRightX - 2},${lipY - 1} ${rimInnerRightX + 1},${lipY - 1}
       L ${rightLipOuterX - 3},${lipY - 1}
       Q ${rightLipOuterX},${lipY - 1} ${rightLipOuterX},${lipY + 3}
       C ${rightLipOuterX},130 ${rightAeroX - 1},145 ${rightAeroX},162
       C ${rightAeroX + 2},195 ${noseRightX + 6},225 ${noseRightX},${bottomY - 5}
       Q ${noseRightX - 2},${bottomY} ${centerX},${bottomY}
       Q ${noseLeftX + 2},${bottomY} ${noseLeftX},${bottomY - 5} Z`

  const cavityTopY = dropY + 8
  const cavityBottomY = bottomY - 12
  const rimCavityPath = `M ${centerX - 20},${cavityTopY}
    L ${centerX + 20},${cavityTopY}
    C ${rightAeroX - 8},180 ${rightAeroX - 6},208 ${noseRightX - 6},${cavityBottomY - 4}
    Q ${centerX},${cavityBottomY} ${noseLeftX + 6},${cavityBottomY - 4}
    C ${leftAeroX + 6},208 ${leftAeroX + 8},180 ${centerX - 20},${cavityTopY} Z`
  const rimTapePath = `M ${leftHumpX - 4},${seatY} L ${leftDropX},${dropY} L ${rightDropX},${dropY} L ${rightHumpX + 4},${seatY}`
  const valveHole = { x1: centerX - 5, x2: centerX + 5, y1: dropY + 1, y2: dropY + 1 }

  const tireBulge = Math.max(8, tireToRimWidthDifference * pixelsPerMillimetre * 0.8)
  const tireHeight = Math.min(105, tireWidth * 2.35)
  const tireTopY = Math.max(38, seatY - tireHeight)
  const shoulderY = tireTopY + 18
  const beadCoreLeft = { cx: rimInnerLeftX + 5, cy: seatY - 7 }
  const beadCoreRight = { cx: rimInnerRightX - 5, cy: seatY - 7 }
  const beadRubberLeftPath = `M ${rimInnerLeftX},${seatY} L ${rimInnerLeftX},${seatY - 14} L ${rimInnerLeftX + 10},${seatY - 12} L ${rimInnerLeftX + 10},${seatY} Z`
  const beadRubberRightPath = `M ${rimInnerRightX},${seatY} L ${rimInnerRightX},${seatY - 14} L ${rimInnerRightX - 10},${seatY - 12} L ${rimInnerRightX - 10},${seatY} Z`
  const leftBulgeX = rimInnerLeftX - tireBulge
  const rightBulgeX = rimInnerRightX + tireBulge
  const midBulgeY = (seatY + tireTopY) / 2 + 10
  const outerTreadPath = `M ${rimInnerLeftX},${seatY - 12}
    C ${leftBulgeX},${midBulgeY + 15} ${leftBulgeX - 2},${shoulderY} ${centerX - 28},${tireTopY + 3}
    Q ${centerX},${tireTopY} ${centerX + 28},${tireTopY + 3}
    C ${rightBulgeX + 2},${shoulderY} ${rightBulgeX},${midBulgeY + 15} ${rimInnerRightX},${seatY - 12}
    L ${rimInnerRightX + 3},${seatY - 12}
    C ${rightBulgeX + 5},${midBulgeY + 15} ${rightBulgeX + 7},${shoulderY} ${centerX + 30},${tireTopY - 3}
    Q ${centerX},${tireTopY - 6} ${centerX - 30},${tireTopY - 3}
    C ${leftBulgeX - 7},${shoulderY} ${leftBulgeX - 5},${midBulgeY + 15} ${rimInnerLeftX - 3},${seatY - 12} Z`
  const innerCasingPath = `M ${rimInnerLeftX + 2},${seatY - 10}
    C ${leftBulgeX + 5},${midBulgeY + 12} ${leftBulgeX + 4},${shoulderY + 6} ${centerX - 22},${tireTopY + 7}
    Q ${centerX},${tireTopY + 5} ${centerX + 22},${tireTopY + 7}
    C ${rightBulgeX - 4},${shoulderY + 6} ${rightBulgeX - 5},${midBulgeY + 12} ${rimInnerRightX - 2},${seatY - 10}`
  const airChamberPath = `M ${rimInnerLeftX + 3},${seatY - 8}
    C ${leftBulgeX + 7},${midBulgeY + 12} ${leftBulgeX + 6},${shoulderY + 8} ${centerX - 20},${tireTopY + 9}
    Q ${centerX},${tireTopY + 7} ${centerX + 20},${tireTopY + 9}
    C ${rightBulgeX - 6},${shoulderY + 8} ${rightBulgeX - 7},${midBulgeY + 12} ${rimInnerRightX - 3},${seatY - 8}
    L ${rimInnerRightX - 4},${seatY + 1}
    L ${rightHumpX},${seatY + 1}
    L ${rightDropX},${dropY - 1}
    L ${leftDropX},${dropY - 1}
    L ${leftHumpX},${seatY + 1}
    L ${rimInnerLeftX + 4},${seatY + 1} Z`
  const tireSipes = [
    { id: 'left-outer', x1: centerX - 38, y1: tireTopY + 8, x2: centerX - 46, y2: tireTopY + 18 },
    { id: 'left-inner', x1: centerX - 26, y1: tireTopY + 5, x2: centerX - 32, y2: tireTopY + 13 },
    { id: 'right-inner', x1: centerX + 26, y1: tireTopY + 5, x2: centerX + 32, y2: tireTopY + 13 },
    { id: 'right-outer', x1: centerX + 38, y1: tireTopY + 8, x2: centerX + 46, y2: tireTopY + 18 },
  ]

  const lipAnchorX = isHookless ? rimInnerLeftX : rimInnerLeftX + 6
  const lipAnchorY = isHookless ? lipY + 4 : lipY + 6
  const lipCallout = {
    linePath: `M ${lipAnchorX},${lipAnchorY} L ${leftLipOuterX - 25},${lipY - 25} L 35,${lipY - 25}`,
    dotX: lipAnchorX,
    dotY: lipAnchorY,
    dotColor: isHookless ? '#0284c7' : '#2563eb',
    title: isHookless
      ? t('guidesTireChoose.engineering.diagram.lip.tssTitle')
      : t('guidesTireChoose.engineering.diagram.lip.tcTitle'),
    description: isHookless
      ? t('guidesTireChoose.engineering.diagram.lip.tssDescription')
      : t('guidesTireChoose.engineering.diagram.lip.tcDescription'),
    titleY: lipY - 30,
    descriptionY: lipY - 14,
  }
  const beadCallout = {
    linePath: `M ${beadCoreRight.cx},${beadCoreRight.cy} L ${rightLipOuterX + 25},${seatY - 25} L 420,${seatY - 25}`,
    dotX: beadCoreRight.cx,
    dotY: beadCoreRight.cy,
    titleY: seatY - 30,
    descriptionY: seatY - 14,
  }
  const tireLabel = t(`guidesTireChoose.engineering.diagram.tireLabels.${status}`, { tireWidth })
  const footerNote = result
    ? t(`guidesTireChoose.engineering.diagram.footer.${status}`, {
      rimWidth,
      tireWidth,
      allowedMin: result.allowedTireWidthRange.min,
      allowedMax: result.allowedTireWidthRange.max,
      recommendedMin: result.recommendedTireWidthRange?.min,
      recommendedMax: result.recommendedTireWidthRange?.max,
    })
    : t('guidesTireChoose.engineering.diagram.footer.empty')

  return {
    centerX,
    lipY,
    seatY,
    status,
    rimInnerLeftX,
    rimInnerRightX,
    rimOuterPath,
    rimCavityPath,
    rimTapePath,
    valveHole,
    tireAirChamberPath: airChamberPath,
    tireInnerCasingPath: innerCasingPath,
    tireOuterPath: outerTreadPath,
    tireStroke: status === 'recommended' ? '#059669' : status === 'reference' ? '#d97706' : '#dc2626',
    tireSipes,
    beadRubberLeftPath,
    beadRubberRightPath,
    beadCoreLeft,
    beadCoreRight,
    measurementY: 140,
    measurementLabelX: centerX,
    lipCallout,
    beadCallout,
    tireLabel,
    footerNote,
    tireTopY,
    tireWidth,
  }
})
</script>

<style scoped>
.tire-rim-engineering-workbench {
  display: grid;
  gap: 1.25rem;
  margin: 0 0 2rem;
  padding: 1.25rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1.25rem;
  background: var(--tz-card-surface);
  box-shadow: 0 12px 28px rgb(15 23 42 / 0.08);
  text-align: left;
}

.tire-rim-engineering-workbench__header {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
  justify-content: space-between;
}


.tire-rim-engineering-workbench__eyebrow,
.tire-rim-engineering-workbench__version {
  margin: 0 0 0.35rem;
  color: var(--tz-text-muted);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.tire-rim-engineering-workbench__version {
  flex: 0 0 auto;
  margin: 0;
  padding: 0.3rem 0.55rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  font-family: var(--tz-font-ui);
  letter-spacing: 0;
  text-transform: none;
}

.tire-rim-engineering-workbench__title,
.tire-rim-engineering-workbench__diagnostic h4,
.tire-rim-engineering-workbench__diagram figcaption,
.tire-rim-engineering-workbench__explanations h4 {
  margin: 0;
  color: var(--tz-text-primary);
  font-weight: 700;
}

.tire-rim-engineering-workbench__title {
  font-size: 1.1rem;
}

.tire-rim-engineering-workbench__board-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.7rem;
}

.tire-rim-engineering-workbench__board-metric {
  display: grid;
  gap: 0.7rem;
  min-width: 0;
  padding: 0.8rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.8rem;
  background: var(--tz-form-panel-surface);
}

.tire-rim-engineering-workbench__board-metric--pressure {
  border-color: rgb(220 38 38 / 0.42);
}

.tire-rim-engineering-workbench__board-metric-label {
  display: block;
  color: var(--tz-text-muted);
  font-size: 0.64rem;
  font-weight: 800;
}

.tire-rim-engineering-workbench__board-metric strong {
  display: block;
  margin-top: 0.25rem;
  color: var(--tz-text-primary);
  font-family: var(--tz-font-ui);
  font-size: 0.9rem;
}

.tire-rim-engineering-workbench__board-metric--pressure strong {
  color: #dc2626;
}

.tire-rim-engineering-workbench__board-metric p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.64rem;
  line-height: 1.45;
}

.tire-rim-engineering-workbench__description,
.tire-rim-engineering-workbench__diagnostic p,
.tire-rim-engineering-workbench__explanations p,
.tire-rim-engineering-workbench__diagram p {
  margin: 0.45rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.6;
}

.tire-rim-engineering-workbench__segmented-control {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: center;
}

.tire-rim-engineering-workbench__segment {
  border: 1px solid var(--tz-form-control-border);
  border-radius: 999px;
  background: var(--tz-form-control-surface);
  color: var(--tz-text-secondary);
  cursor: pointer;
  font-size: 0.72rem;
  font-weight: 600;
  transition: background-color 160ms ease, border-color 160ms ease, color 160ms ease;
}

.tire-rim-engineering-workbench__segment {
  padding: 0.42rem 0.65rem;
}

.tire-rim-engineering-workbench__segment:hover,
.tire-rim-engineering-workbench__segment:focus-visible {
  border-color: var(--tz-site-accent);
  color: var(--tz-text-primary);
  outline: none;
}

.tire-rim-engineering-workbench__segment--active {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent);
  color: #fff;
}

.tire-rim-engineering-workbench__controls {
  display: grid;
  grid-template-columns: minmax(12rem, 0.6fr) repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
}

.tire-rim-engineering-workbench__picker-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  width: 100%;
  height: 42px;
  padding: 0 0.75rem;
  border: 1px solid var(--tz-form-control-border);
  border-radius: 0.375rem;
  background: var(--tz-form-control-surface);
  color: var(--tz-text-primary);
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.06);
  font-size: 0.75rem;
  text-align: left;
  cursor: pointer;
}

.tire-rim-engineering-workbench__picker-trigger:hover,
.tire-rim-engineering-workbench__picker-trigger:focus-visible {
  border-color: var(--tz-site-accent);
  outline: none;
}

.tire-rim-engineering-workbench__field {
  min-width: 0;
  margin: 0;
  padding: 0;
  border: 0;
}

.tire-rim-engineering-workbench__field legend {
  display: block;
  margin-bottom: 0.4rem;
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  font-weight: 700;
}

.tire-rim-engineering-workbench__result-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.85rem;
  align-items: stretch;
}

.tire-rim-engineering-workbench__diagnostic,
.tire-rim-engineering-workbench__diagram,
.tire-rim-engineering-workbench__explanations article {
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.9rem;
  background: var(--tz-form-panel-surface);
}

.tire-rim-engineering-workbench__diagnostic {
  padding: 1rem;
}

.tire-rim-engineering-workbench__diagnostic--recommended {
  border-color: rgb(5 150 105 / 0.4);
  background: rgb(236 253 245 / 0.7);
}

.tire-rim-engineering-workbench__diagnostic--reference {
  border-color: rgb(2 132 199 / 0.35);
  background: rgb(240 249 255 / 0.7);
}

.tire-rim-engineering-workbench__diagnostic--critical {
  border-color: rgb(220 38 38 / 0.45);
  background: rgb(254 242 242 / 0.8);
}

.tire-rim-engineering-workbench__diagnostic--empty {
  display: grid;
  align-content: center;
  min-height: 15rem;
}

.tire-rim-engineering-workbench__diagnostic-badge {
  display: inline-flex;
  padding: 0.3rem 0.55rem;
  border-radius: 999px;
  background: var(--tz-text-primary);
  color: #fff;
  font-size: 0.66rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.tire-rim-engineering-workbench__diagnostic--recommended .tire-rim-engineering-workbench__diagnostic-badge {
  background: #047857;
}

.tire-rim-engineering-workbench__diagnostic--reference .tire-rim-engineering-workbench__diagnostic-badge {
  background: #0369a1;
}

.tire-rim-engineering-workbench__diagnostic--critical .tire-rim-engineering-workbench__diagnostic-badge {
  background: #b91c1c;
}

.tire-rim-engineering-workbench__diagnostic h4 {
  margin-top: 0.65rem;
  font-size: 0.98rem;
}

.tire-rim-engineering-workbench__metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
  margin: 0.9rem 0 0;
}

.tire-rim-engineering-workbench__metrics div {
  padding: 0.55rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.65rem;
  background: var(--tz-card-surface);
}

.tire-rim-engineering-workbench__metrics dt {
  color: var(--tz-text-muted);
  font-size: 0.62rem;
}

.tire-rim-engineering-workbench__metrics dd {
  margin: 0.2rem 0 0;
  color: var(--tz-text-primary);
  font-family: var(--tz-font-ui);
  font-size: 0.78rem;
  font-weight: 700;
}

.tire-rim-engineering-workbench__metrics small {
  display: block;
  margin-top: 0.15rem;
  color: var(--tz-text-muted);
  font-size: 0.62rem;
}

.tire-rim-engineering-workbench__diagnostic-note {
  padding-top: 0.6rem;
  border-top: 1px solid var(--tz-border-subtle);
  font-size: 0.68rem !important;
}

.tire-rim-engineering-workbench__diagram {
  display: grid;
  grid-column: 1 / -1;
  grid-template-rows: auto 1fr auto;
  min-width: 0;
  margin: 0;
  padding: 0.75rem;
}

.tire-rim-engineering-workbench__diagram figcaption {
  font-size: 0.8rem;
}

.tire-rim-engineering-workbench__diagram-stage {
  position: relative;
  min-width: 0;
  min-height: 280px;
  margin-top: 0.35rem;
  padding: 1.35rem 0.5rem 0.5rem;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.85rem;
  background: var(--tz-form-panel-surface);
}

.tire-rim-engineering-workbench__diagram-floating-badge {
  position: absolute;
  z-index: 1;
  top: 0.15rem;
  left: 0.35rem;
  max-width: 48%;
  padding: 0.22rem 0.42rem;
  border: 0;
  border-radius: 0.25rem;
  background: rgb(15 23 42 / 0.88);
  color: #fff;
  font-family: var(--tz-font-ui);
  font-size: 0.52rem;
  font-weight: 800;
  line-height: 1.25;
}

.tire-rim-engineering-workbench__diagram-floating-badge--status {
  right: 0.35rem;
  left: auto;
  color: #fff;
  text-align: right;
}

.tire-rim-engineering-workbench__diagram-floating-badge--critical {
  background: rgb(220 38 38 / 0.94);
}

.tire-rim-engineering-workbench__diagram-floating-badge--recommended {
  background: rgb(5 150 105 / 0.94);
}

.tire-rim-engineering-workbench__diagram-floating-badge--reference {
  background: rgb(217 119 6 / 0.94);
}

.tire-rim-engineering-workbench__diagram svg {
  display: block;
  width: 100%;
  height: 280px;
  margin: 0;
  overflow: visible;
}

.tire-rim-engineering-workbench__diagram-canvas {
  fill: url(#tire-rim-engineering-grid-pattern);
  stroke: var(--tz-border-subtle);
  stroke-dasharray: 3 3;
}

.tire-rim-engineering-workbench__diagram-centerline {
  stroke: #cbd5e1;
  stroke-dasharray: 3 3;
}

.tire-rim-engineering-workbench__diagram-baseline {
  stroke: #e2e8f0;
  stroke-width: 1;
}

.tire-rim-engineering-workbench__diagram-rim {
  stroke: #0f172a;
  stroke-width: 2;
}

.tire-rim-engineering-workbench__diagram-rim-cavity {
  fill: #f8fafc;
  stroke: #475569;
  stroke-width: 1.2;
  stroke-dasharray: 2 2;
}

.tire-rim-engineering-workbench__diagram-rim-tape {
  fill: none;
  stroke: #0284c7;
  stroke-linecap: round;
  stroke-width: 3;
}

.tire-rim-engineering-workbench__diagram-valve-hole {
  stroke: #94a3b8;
  stroke-width: 3;
}

.tire-rim-engineering-workbench__diagram-air-chamber {
  fill: #f1f5f9;
  fill-opacity: 0.9;
  stroke: #64748b;
  stroke-dasharray: 3 2;
  stroke-width: 1.2;
}

.tire-rim-engineering-workbench__diagram-inner-casing {
  fill: none;
  stroke: #27272a;
  stroke-width: 2;
}

.tire-rim-engineering-workbench__diagram-tire {
  stroke-width: 2.5;
}

.tire-rim-engineering-workbench__diagram-tire-sipes {
  fill: none;
  stroke: #52525b;
  stroke-linecap: round;
  stroke-width: 1.2;
}

.tire-rim-engineering-workbench__diagram-bead path {
  fill: #18181b;
  stroke: #09090b;
  stroke-width: 1;
}

.tire-rim-engineering-workbench__diagram-kevlar-core {
  fill: url(#tire-rim-engineering-kevlar-gradient);
  stroke: #78350f;
  stroke-width: 0.8;
}

.tire-rim-engineering-workbench__diagram-stress-indicators {
  fill: #dc2626;
  stroke: #dc2626;
  stroke-width: 2.5;
  stroke-dasharray: 2 2;
}

.tire-rim-engineering-workbench__diagram-stress-indicators text {
  stroke: none;
  fill: #dc2626;
  font-family: var(--tz-font-ui);
  font-size: 9px;
  font-weight: 900;
}

.tire-rim-engineering-workbench__diagram-lock-ring {
  fill: none;
  stroke: #059669;
  stroke-width: 1.5;
  stroke-dasharray: 2 2;
}

.tire-rim-engineering-workbench__diagram-caliper line {
  stroke: #0284c7;
  stroke-dasharray: 3 2;
  stroke-width: 1.5;
}

.tire-rim-engineering-workbench__diagram-caliper polygon {
  fill: #0284c7;
}

.tire-rim-engineering-workbench__diagram-caliper rect {
  fill: #f0f9ff;
  stroke: #bae6fd;
  stroke-width: 1;
}

.tire-rim-engineering-workbench__diagram-caliper text {
  fill: #0369a1;
  stroke: none;
  font-family: var(--tz-font-ui);
  font-size: 10px;
  font-weight: 900;
}

.tire-rim-engineering-workbench__diagram-callout path {
  fill: none;
  stroke: #475569;
  stroke-width: 1.2;
}

.tire-rim-engineering-workbench__diagram-callout text {
  fill: #0f172a;
  stroke: none;
  font-family: var(--tz-font-ui);
  font-size: 9px;
  font-weight: 900;
}

.tire-rim-engineering-workbench__diagram-callout-sub {
  fill: #64748b !important;
  font-size: 8px !important;
  font-weight: 700 !important;
}

.tire-rim-engineering-workbench__diagram-callout--bead text:first-of-type {
  fill: #854d0e;
}

.tire-rim-engineering-workbench__diagram-tire-label {
  font-family: var(--tz-font-ui);
  font-size: 12px;
  font-weight: 900;
  stroke: none;
}

.tire-rim-engineering-workbench__diagram-note {
  min-height: 2.5em;
}

.tire-rim-engineering-workbench__explanations {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.tire-rim-engineering-workbench__explanations article {
  padding: 0.8rem;
}

.tire-rim-engineering-workbench__explanations h4 {
  font-size: 0.78rem;
}

.tire-rim-engineering-workbench__explanations p {
  font-size: 0.7rem;
}

.tire-rim-engineering-workbench__examples {
  display: grid;
  gap: 0.7rem;
  padding: 1rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.95rem;
  background: var(--tz-form-panel-surface);
}

.tire-rim-engineering-workbench__examples h4 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.85rem;
  font-weight: 800;
}

.tire-rim-engineering-workbench__examples > div:first-child > p {
  margin: 0.3rem 0 0;
  color: var(--tz-text-muted);
  font-size: 0.7rem;
  line-height: 1.5;
}

.tire-rim-engineering-workbench__examples-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.7rem;
}

.tire-rim-engineering-workbench__example {
  min-width: 0;
  padding: 0.8rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.tire-rim-engineering-workbench__example--critical {
  border-color: rgb(220 38 38 / 0.42);
}

.tire-rim-engineering-workbench__example--reference {
  border-color: rgb(2 132 199 / 0.35);
}

.tire-rim-engineering-workbench__example--recommended {
  border-color: rgb(5 150 105 / 0.4);
}

.tire-rim-engineering-workbench__example-badge {
  color: var(--tz-text-muted);
  font-size: 0.62rem;
  font-weight: 800;
}

.tire-rim-engineering-workbench__example h5 {
  margin: 0.45rem 0 0;
  color: var(--tz-text-primary);
  font-size: 0.73rem;
  line-height: 1.4;
}

.tire-rim-engineering-workbench__example p {
  margin: 0.4rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  line-height: 1.5;
}

@media (max-width: 900px) {
  .tire-rim-engineering-workbench__controls,
  .tire-rim-engineering-workbench__result-layout,
  .tire-rim-engineering-workbench__board-metrics {
    grid-template-columns: 1fr;
  }

  .tire-rim-engineering-workbench__examples-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 680px) {
  .tire-rim-engineering-workbench {
    padding: 0.85rem;
  }

  .tire-rim-engineering-workbench__header,
  .tire-rim-engineering-workbench__explanations {
    grid-template-columns: 1fr;
    display: grid;
  }

  .tire-rim-engineering-workbench__version {
    justify-self: start;
  }
}
</style>
