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

    <section
      v-if="engineeringReferenceResult"
      class="tire-rim-engineering-workbench__global-diagnostic"
      :class="`tire-rim-engineering-workbench__global-diagnostic--${engineeringReferenceResult.verdict}`"
      aria-live="polite"
    >
      <div class="tire-rim-engineering-workbench__global-diagnostic-heading">
        <span class="tire-rim-engineering-workbench__global-diagnostic-badge">
          {{ t(`guidesTireChoose.engineering.board.status.badges.${engineeringReferenceResult.verdict}`) }}
        </span>
        <span class="tire-rim-engineering-workbench__global-diagnostic-source">
          {{ t('guidesTireChoose.engineering.board.statusSource') }}
        </span>
      </div>
      <p>{{ diagnosticDescription }}</p>
    </section>

    <section class="tire-rim-engineering-workbench__board-metrics" aria-label="Engineering safety metrics">
      <article class="tire-rim-engineering-workbench__board-metric tire-rim-engineering-workbench__board-metric--pressure">
        <div>
          <span class="tire-rim-engineering-workbench__board-metric-label">
            {{ t('guidesTireChoose.engineering.board.metrics.pressureLabel') }}
          </span>
          <strong>{{ maxPressure.bar.toFixed(1) }} bar</strong>
        </div>
        <p>{{ maxPressure.psi.toFixed(1) }} PSI · {{ t('guidesTireChoose.engineering.board.metrics.pressureNote') }}</p>
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

    <div class="tire-rim-engineering-workbench__mode-tabs" role="tablist" :aria-label="t('guidesTireChoose.engineering.modeLabel')">
      <button
        v-for="modeOption in modeOptions"
        :key="modeOption"
        type="button"
        class="tire-rim-engineering-workbench__mode-tab"
        :class="{ 'tire-rim-engineering-workbench__mode-tab--active': mode === modeOption }"
        role="tab"
        :aria-selected="mode === modeOption"
        @click="mode = modeOption"
      >
        {{ t(`guidesTireChoose.engineering.modes.${modeOption}`) }}
      </button>
    </div>

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
        <legend>{{ t('guidesTireChoose.engineering.rimInnerWidthLabel') }}</legend>
        <div class="tire-rim-engineering-workbench__width-options">
          <button
            v-for="rimWidth in TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM"
            :key="rimWidth"
            type="button"
            class="tire-rim-engineering-workbench__width-option"
            :class="{ 'tire-rim-engineering-workbench__width-option--active': rimInnerWidthMm === rimWidth }"
            :aria-pressed="rimInnerWidthMm === rimWidth"
            @click="rimInnerWidthMm = rimWidth"
          >
            {{ rimWidth }} mm
          </button>
        </div>
      </fieldset>

      <label class="tire-rim-engineering-workbench__field tire-rim-engineering-workbench__field--tire-width">
        <span>{{ t(mode === 'rim' ? 'guidesTireChoose.engineering.tireWidthLabel' : 'guidesTireChoose.engineering.tireWidthInputLabel') }}</span>
        <span class="tire-rim-engineering-workbench__input-wrap">
          <input
            v-model.number="tireWidthInput"
            type="number"
            min="18"
            max="127"
            step="1"
            inputmode="numeric"
            :aria-label="t('guidesTireChoose.engineering.tireWidthLabel')"
          />
          <span>mm</span>
        </span>
        <span class="tire-rim-engineering-workbench__input-hint">
          {{ t('guidesTireChoose.engineering.tireWidthHint') }}
        </span>
      </label>
    </div>

    <div class="tire-rim-engineering-workbench__preset-row" aria-label="Tire width presets">
      <span>{{ t('guidesTireChoose.engineering.tireWidthPresetsLabel') }}</span>
      <button
        v-for="tireWidth in TIRE_RIM_ENGINEERING_TIRE_WIDTH_PRESETS_MM"
        :key="tireWidth"
        type="button"
        class="tire-rim-engineering-workbench__preset"
        :class="{ 'tire-rim-engineering-workbench__preset--active': tireWidthInput === tireWidth }"
        @click="tireWidthInput = tireWidth"
      >
        {{ tireWidth }}C
      </button>
    </div>

    <div class="tire-rim-engineering-workbench__result-layout">
      <section
        v-if="engineeringReferenceResult"
        class="tire-rim-engineering-workbench__diagnostic"
        :class="`tire-rim-engineering-workbench__diagnostic--${engineeringReferenceResult.verdict}`"
        aria-live="polite"
      >
        <div class="tire-rim-engineering-workbench__diagnostic-badge">
          {{ t(`guidesTireChoose.engineering.verdicts.${engineeringReferenceResult.verdict}.badge`) }}
        </div>
        <h4>
          {{ t(`guidesTireChoose.engineering.verdicts.${engineeringReferenceResult.verdict}.title`) }}
        </h4>
        <p>
          {{ t(`guidesTireChoose.engineering.reasons.${engineeringReferenceResult.verdictReason}`, {
            rimWidth: engineeringReferenceResult.rimInnerWidthMm,
            tireWidth: engineeringReferenceResult.tireWidthMm,
            allowedMin: engineeringReferenceResult.allowedTireWidthRange.min,
            allowedMax: engineeringReferenceResult.allowedTireWidthRange.max,
            recommendedMin: engineeringReferenceResult.recommendedTireWidthRange.min,
            recommendedMax: engineeringReferenceResult.recommendedTireWidthRange.max,
          }) }}
        </p>

        <dl class="tire-rim-engineering-workbench__metrics">
          <div>
            <dt>{{ t('guidesTireChoose.engineering.metrics.allowedRange') }}</dt>
            <dd>{{ formatWidthRange(engineeringReferenceResult.allowedTireWidthRange) }} mm</dd>
          </div>
          <div>
            <dt>{{ t('guidesTireChoose.engineering.metrics.recommendedRange') }}</dt>
            <dd>{{ formatWidthRange(engineeringReferenceResult.recommendedTireWidthRange) }} mm</dd>
          </div>
          <div>
            <dt>{{ t('guidesTireChoose.engineering.metrics.inflatedWidth') }}</dt>
            <dd>{{ engineeringReferenceResult.inflatedTireWidthMm.toFixed(1) }} mm</dd>
          </div>
          <div>
            <dt>{{ t('guidesTireChoose.engineering.metrics.aeroTarget') }}</dt>
            <dd>{{ engineeringReferenceResult.aeroTargetOuterWidthMm.toFixed(1) }} mm</dd>
          </div>
        </dl>
        <p class="tire-rim-engineering-workbench__diagnostic-note">
          {{ t('guidesTireChoose.engineering.resultNote') }}
        </p>
      </section>

      <section v-else class="tire-rim-engineering-workbench__diagnostic tire-rim-engineering-workbench__diagnostic--empty">
        <h4>{{ t('guidesTireChoose.engineering.emptyTitle') }}</h4>
        <p>{{ t('guidesTireChoose.engineering.emptyDescription') }}</p>
      </section>

      <figure class="tire-rim-engineering-workbench__diagram">
        <figcaption>{{ t('guidesTireChoose.engineering.diagramTitle') }}</figcaption>
        <div class="tire-rim-engineering-workbench__diagram-stage">
          <div class="tire-rim-engineering-workbench__diagram-floating-badge">
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

    <div class="tire-rim-engineering-workbench__explanations">
      <article v-for="explanationKey in explanationKeys" :key="explanationKey">
        <h4>{{ t(`guidesTireChoose.engineering.explanations.${explanationKey}.title`) }}</h4>
        <p>{{ t(`guidesTireChoose.engineering.explanations.${explanationKey}.body`) }}</p>
      </article>
    </div>

    <div class="tire-rim-engineering-workbench__checklist">
      <h4>{{ t('guidesTireChoose.engineering.checklist.title') }}</h4>
      <ul>
        <li v-for="checklistKey in checklistKeys" :key="checklistKey">
          {{ t(`guidesTireChoose.engineering.checklist.items.${checklistKey}`) }}
        </li>
      </ul>
    </div>

    <section class="tire-rim-engineering-workbench__aero-board">
      <div class="tire-rim-engineering-workbench__board-section-heading">
        <h4>{{ t('guidesTireChoose.engineering.board.aero.title') }}</h4>
        <span>{{ t('guidesTireChoose.engineering.board.aero.source') }}</span>
      </div>
      <p class="tire-rim-engineering-workbench__board-section-intro">
        {{ t('guidesTireChoose.engineering.board.aero.intro') }}
      </p>
      <div class="tire-rim-engineering-workbench__aero-grid">
        <article v-for="aeroKey in aeroKeys" :key="aeroKey">
          <h5>{{ t(`guidesTireChoose.engineering.board.aero.cards.${aeroKey}.title`) }}</h5>
          <strong>{{ t(`guidesTireChoose.engineering.board.aero.cards.${aeroKey}.formula`) }}</strong>
          <p>{{ t(`guidesTireChoose.engineering.board.aero.cards.${aeroKey}.body`) }}</p>
        </article>
      </div>
      <div class="tire-rim-engineering-workbench__aero-formula">
        <strong>{{ t('guidesTireChoose.engineering.board.aero.formulaLabel') }}</strong>
        <code>W<sub>inflated</sub> ≈ W<sub>nominal</sub> + 0.4 × (W<sub>rim inner</sub> − 19)</code>
        <span>
          {{ t('guidesTireChoose.engineering.board.aero.current', {
            rimWidth: engineeringReferenceResult?.rimInnerWidthMm ?? rimInnerWidthMm,
            tireWidth: engineeringReferenceResult?.tireWidthMm ?? tireWidthInput,
            inflatedWidth: engineeringReferenceResult?.inflatedTireWidthMm.toFixed(1) ?? '—',
            aeroTarget: engineeringReferenceResult?.aeroTargetOuterWidthMm.toFixed(1) ?? '—',
          }) }}
        </span>
      </div>
    </section>

    <section class="tire-rim-engineering-workbench__traps-board">
      <div class="tire-rim-engineering-workbench__board-section-heading">
        <h4>{{ t('guidesTireChoose.engineering.board.traps.title') }}</h4>
        <span>{{ t('guidesTireChoose.engineering.board.traps.source') }}</span>
      </div>
      <div class="tire-rim-engineering-workbench__traps-grid">
        <article v-for="trapKey in trapKeys" :key="trapKey">
          <span class="tire-rim-engineering-workbench__trap-number">
            {{ t(`guidesTireChoose.engineering.board.traps.items.${trapKey}.number`) }}
          </span>
          <h5>{{ t(`guidesTireChoose.engineering.board.traps.items.${trapKey}.title`) }}</h5>
          <p>{{ t(`guidesTireChoose.engineering.board.traps.items.${trapKey}.body`) }}</p>
          <strong>{{ t(`guidesTireChoose.engineering.board.traps.items.${trapKey}.rule`) }}</strong>
        </article>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  calculateTireRimEngineeringReference,
  TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM,
  TIRE_RIM_ENGINEERING_TIRE_WIDTH_PRESETS_MM,
  type TireRimEngineeringRimSystem,
} from '~/data/tireguides/tireRimEngineeringReferenceModel'

type TireRimEngineeringWorkbenchMode = 'rim' | 'tire'

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
  valveHole: {
    x1: number
    x2: number
    y1: number
    y2: number
  }
  tireAirChamberPath: string
  tireInnerCasingPath: string
  tireOuterPath: string
  tireStroke: string
  tireSipes: Array<{
    id: string
    x1: number
    y1: number
    x2: number
    y2: number
  }>
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
  beadCallout: {
    linePath: string
    dotX: number
    dotY: number
    titleY: number
    descriptionY: number
  }
  tireLabel: string
  footerNote: string
  tireTopY: number
  tireWidth: number
}

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireChoose')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const modeOptions: readonly TireRimEngineeringWorkbenchMode[] = ['rim', 'tire']
const rimSystemOptions: readonly TireRimEngineeringRimSystem[] = ['hookless', 'hooked']
const explanationKeys = ['rimSystem', 'nominalWidth', 'aeroProjection'] as const
const checklistKeys = ['exactSpecifications', 'pressureLimits', 'installationCheck'] as const
const aeroKeys = ['bulbTire', 'ruleOf105', 'modernSweetSpot'] as const
const trapKeys = ['nonTleOnHookless', 'emptyTpuInflation', 'oldSealantResidue', 'co2ColdBrittle', 'metalLever', 'beadPinch'] as const

const mode = ref<TireRimEngineeringWorkbenchMode>('rim')
const rimSystem = ref<TireRimEngineeringRimSystem>('hookless')
const rimInnerWidthMm = ref<number>(25)
const tireWidthInput = ref<number>(28)

const engineeringReferenceResult = computed(() => calculateTireRimEngineeringReference(
  rimSystem.value,
  rimInnerWidthMm.value,
  tireWidthInput.value,
))

const formatWidthRange = ({ min, max }: { min: number; max: number }): string => (
  min === max ? String(min) : `${min}–${max}`
)

const maxPressure = computed(() => {
  const inputWidth = Number(tireWidthInput.value)
  const tireWidth = Number.isFinite(inputWidth) && inputWidth > 0 ? inputWidth : 28
  if (rimSystem.value === 'hookless') {
    if (tireWidth <= 24) return { bar: 5.5, psi: 80 }
    if (tireWidth <= 29) return { bar: 5, psi: 72.5 }
    if (tireWidth <= 34) return { bar: 4.5, psi: 65 }
    if (tireWidth <= 39) return { bar: 4, psi: 58 }
    if (tireWidth <= 44) return { bar: 3.5, psi: 51 }
    return { bar: 3, psi: 43.5 }
  }
  if (tireWidth <= 24) return { bar: 8, psi: 116 }
  if (tireWidth <= 26) return { bar: 7.5, psi: 110 }
  if (tireWidth <= 29) return { bar: 6.5, psi: 95 }
  if (tireWidth <= 34) return { bar: 5.5, psi: 80 }
  if (tireWidth <= 39) return { bar: 4.5, psi: 65 }
  if (tireWidth <= 45) return { bar: 3.5, psi: 50 }
  return { bar: 2.8, psi: 40 }
})

const tubeRecommendation = computed(() => {
  const tireWidth = Number(tireWidthInput.value)
  if (tireWidth >= 38) {
    return {
      model: t('guidesTireChoose.engineering.board.metrics.tubes.wideModel'),
      range: t('guidesTireChoose.engineering.board.metrics.tubes.wideRange'),
    }
  }
  if (tireWidth >= 29) {
    return {
      model: t('guidesTireChoose.engineering.board.metrics.tubes.midModel'),
      range: t('guidesTireChoose.engineering.board.metrics.tubes.midRange'),
    }
  }
  return {
    model: t('guidesTireChoose.engineering.board.metrics.tubes.roadModel'),
    range: t('guidesTireChoose.engineering.board.metrics.tubes.roadRange'),
  }
})

const diagnosticDescription = computed(() => {
  const result = engineeringReferenceResult.value
  if (!result) return t('guidesTireChoose.engineering.board.status.empty')
  const messageKey = result.verdict === 'critical'
    ? result.verdictReason === 'below_minimum' ? 'criticalBelow' : 'criticalAbove'
    : result.verdict === 'recommended' ? 'recommended' : 'reference'
  return t(`guidesTireChoose.engineering.board.status.${messageKey}`, {
    rimWidth: result.rimInnerWidthMm,
    tireWidth: result.tireWidthMm,
    allowedMin: result.allowedTireWidthRange.min,
    allowedMax: result.allowedTireWidthRange.max,
    recommendedMin: result.recommendedTireWidthRange.min,
    recommendedMax: result.recommendedTireWidthRange.max,
    maxBar: maxPressure.value.bar.toFixed(1),
  })
})

const diagramGeometry = computed<TireRimEngineeringDiagramGeometry>(() => {
  const centerX = 230
  const pixelsPerMillimetre = 4.2
  const rimWidth = rimInnerWidthMm.value
  const parsedTireWidth = Number(tireWidthInput.value)
  const tireWidth = Number.isFinite(parsedTireWidth) && parsedTireWidth > 0
    ? Math.max(18, Math.min(127, parsedTireWidth))
    : 28
  const result = engineeringReferenceResult.value
  const status = result?.verdict ?? 'reference'
  const halfRimWidth = (rimWidth / 2) * pixelsPerMillimetre
  const rimInnerLeftX = centerX - halfRimWidth
  const rimInnerRightX = centerX + halfRimWidth

  // These anchors intentionally match the original HTML workbench. Keeping
  // the same coordinate system preserves the detailed callouts and section
  // relationships when the diagram is moved into Vue.
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

  const tireBulge = Math.max(8, (tireWidth - rimWidth) * 3.4)
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
      recommendedMin: result.recommendedTireWidthRange.min,
      recommendedMax: result.recommendedTireWidthRange.max,
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
.tire-rim-engineering-workbench__explanations h4,
.tire-rim-engineering-workbench__checklist h4 {
  margin: 0;
  color: var(--tz-text-primary);
  font-weight: 700;
}

.tire-rim-engineering-workbench__title {
  font-size: 1.1rem;
}

.tire-rim-engineering-workbench__global-diagnostic {
  padding: 0.9rem 1rem;
  border: 1px solid rgb(220 38 38 / 0.42);
  border-radius: 0.9rem;
  background: rgb(254 242 242 / 0.85);
}

.tire-rim-engineering-workbench__global-diagnostic--recommended {
  border-color: rgb(5 150 105 / 0.42);
  background: rgb(236 253 245 / 0.85);
}

.tire-rim-engineering-workbench__global-diagnostic--reference {
  border-color: rgb(217 119 6 / 0.42);
  background: rgb(255 251 235 / 0.9);
}

.tire-rim-engineering-workbench__global-diagnostic-heading {
  display: flex;
  gap: 0.7rem;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
}

.tire-rim-engineering-workbench__global-diagnostic-badge {
  display: inline-flex;
  padding: 0.32rem 0.55rem;
  border-radius: 0.35rem;
  background: #b91c1c;
  color: #fff;
  font-family: var(--tz-font-ui);
  font-size: 0.64rem;
  font-weight: 900;
}

.tire-rim-engineering-workbench__global-diagnostic--recommended .tire-rim-engineering-workbench__global-diagnostic-badge {
  background: #047857;
}

.tire-rim-engineering-workbench__global-diagnostic--reference .tire-rim-engineering-workbench__global-diagnostic-badge {
  background: #b45309;
}

.tire-rim-engineering-workbench__global-diagnostic-source {
  color: var(--tz-text-muted);
  font-family: var(--tz-font-ui);
  font-size: 0.62rem;
  font-weight: 800;
}

.tire-rim-engineering-workbench__global-diagnostic p {
  margin: 0.6rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.76rem;
  line-height: 1.65;
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
.tire-rim-engineering-workbench__checklist li,
.tire-rim-engineering-workbench__diagram p {
  margin: 0.45rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.6;
}

.tire-rim-engineering-workbench__mode-tabs,
.tire-rim-engineering-workbench__segmented-control,
.tire-rim-engineering-workbench__width-options,
.tire-rim-engineering-workbench__preset-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: center;
}

.tire-rim-engineering-workbench__mode-tabs {
  padding: 0.25rem;
  border-radius: 999px;
  background: var(--tz-form-panel-surface);
}

.tire-rim-engineering-workbench__mode-tab,
.tire-rim-engineering-workbench__segment,
.tire-rim-engineering-workbench__width-option,
.tire-rim-engineering-workbench__preset {
  border: 1px solid var(--tz-form-control-border);
  border-radius: 999px;
  background: var(--tz-form-control-surface);
  color: var(--tz-text-secondary);
  cursor: pointer;
  font-size: 0.72rem;
  font-weight: 600;
  transition: background-color 160ms ease, border-color 160ms ease, color 160ms ease;
}

.tire-rim-engineering-workbench__mode-tab {
  flex: 1 1 14rem;
  padding: 0.55rem 0.8rem;
}

.tire-rim-engineering-workbench__segment,
.tire-rim-engineering-workbench__width-option,
.tire-rim-engineering-workbench__preset {
  padding: 0.42rem 0.65rem;
}

.tire-rim-engineering-workbench__mode-tab:hover,
.tire-rim-engineering-workbench__mode-tab:focus-visible,
.tire-rim-engineering-workbench__segment:hover,
.tire-rim-engineering-workbench__segment:focus-visible,
.tire-rim-engineering-workbench__width-option:hover,
.tire-rim-engineering-workbench__width-option:focus-visible,
.tire-rim-engineering-workbench__preset:hover,
.tire-rim-engineering-workbench__preset:focus-visible {
  border-color: var(--tz-site-accent);
  color: var(--tz-text-primary);
  outline: none;
}

.tire-rim-engineering-workbench__mode-tab--active,
.tire-rim-engineering-workbench__segment--active,
.tire-rim-engineering-workbench__width-option--active,
.tire-rim-engineering-workbench__preset--active {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent);
  color: #fff;
}

.tire-rim-engineering-workbench__controls {
  display: grid;
  grid-template-columns: minmax(12rem, 0.8fr) minmax(18rem, 1.4fr) minmax(10rem, 0.7fr);
  gap: 0.85rem;
}

.tire-rim-engineering-workbench__field {
  min-width: 0;
  margin: 0;
  padding: 0;
  border: 0;
}

.tire-rim-engineering-workbench__field legend,
.tire-rim-engineering-workbench__field > span:first-child,
.tire-rim-engineering-workbench__preset-row > span {
  display: block;
  margin-bottom: 0.4rem;
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  font-weight: 700;
}

.tire-rim-engineering-workbench__input-wrap {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.tire-rim-engineering-workbench__input-wrap input {
  width: 100%;
  min-width: 0;
  padding: 0.46rem 0.55rem;
  border: 1px solid var(--tz-form-control-border);
  border-radius: 0.5rem;
  background: var(--tz-form-control-surface);
  color: var(--tz-text-primary);
  font-size: 0.8rem;
}

.tire-rim-engineering-workbench__input-wrap input:focus {
  border-color: var(--tz-site-accent);
  outline: none;
}

.tire-rim-engineering-workbench__input-wrap > span {
  color: var(--tz-text-muted);
  font-size: 0.72rem;
}

.tire-rim-engineering-workbench__input-hint {
  display: block;
  margin-top: 0.35rem;
  color: var(--tz-text-muted);
  font-size: 0.66rem;
  line-height: 1.4;
}

.tire-rim-engineering-workbench__preset-row {
  padding-top: 0.25rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.tire-rim-engineering-workbench__preset-row > span {
  margin: 0 0.15rem 0 0;
}

.tire-rim-engineering-workbench__result-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(18rem, 0.9fr);
  gap: 0.85rem;
  align-items: stretch;
}

.tire-rim-engineering-workbench__diagnostic,
.tire-rim-engineering-workbench__diagram,
.tire-rim-engineering-workbench__explanations article,
.tire-rim-engineering-workbench__checklist {
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

.tire-rim-engineering-workbench__explanations article,
.tire-rim-engineering-workbench__checklist {
  padding: 0.8rem;
}

.tire-rim-engineering-workbench__explanations h4,
.tire-rim-engineering-workbench__checklist h4 {
  font-size: 0.78rem;
}

.tire-rim-engineering-workbench__explanations p,
.tire-rim-engineering-workbench__checklist li {
  font-size: 0.7rem;
}

.tire-rim-engineering-workbench__checklist ul {
  display: grid;
  gap: 0.2rem;
  margin: 0.45rem 0 0;
  padding-left: 1.1rem;
}

.tire-rim-engineering-workbench__aero-board,
.tire-rim-engineering-workbench__traps-board {
  padding: 1rem;
  border: 1px dashed var(--tz-border-subtle);
  border-radius: 0.95rem;
  background: var(--tz-form-panel-surface);
}

.tire-rim-engineering-workbench__board-section-heading {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0.55rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.tire-rim-engineering-workbench__board-section-heading h4 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 900;
}

.tire-rim-engineering-workbench__board-section-heading span {
  color: var(--tz-text-muted);
  font-family: var(--tz-font-ui);
  font-size: 0.58rem;
  font-weight: 800;
  text-align: right;
}

.tire-rim-engineering-workbench__board-section-intro {
  margin: 0.65rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  line-height: 1.6;
}

.tire-rim-engineering-workbench__aero-grid,
.tire-rim-engineering-workbench__traps-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.7rem;
  margin-top: 0.8rem;
}

.tire-rim-engineering-workbench__aero-grid article,
.tire-rim-engineering-workbench__traps-grid article {
  min-width: 0;
  padding: 0.75rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.tire-rim-engineering-workbench__aero-grid h5,
.tire-rim-engineering-workbench__traps-grid h5 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.72rem;
  line-height: 1.4;
}

.tire-rim-engineering-workbench__aero-grid strong {
  display: block;
  margin-top: 0.45rem;
  color: var(--tz-text-primary);
  font-family: var(--tz-font-ui);
  font-size: 0.66rem;
  line-height: 1.45;
}

.tire-rim-engineering-workbench__aero-grid p,
.tire-rim-engineering-workbench__traps-grid p {
  margin: 0.45rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.66rem;
  line-height: 1.55;
}

.tire-rim-engineering-workbench__aero-formula {
  display: flex;
  gap: 0.7rem;
  align-items: baseline;
  flex-wrap: wrap;
  margin-top: 0.8rem;
  padding: 0.7rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.7rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-secondary);
  font-size: 0.66rem;
}

.tire-rim-engineering-workbench__aero-formula strong,
.tire-rim-engineering-workbench__aero-formula code {
  color: var(--tz-text-primary);
  font-family: var(--tz-font-ui);
  font-weight: 900;
}

.tire-rim-engineering-workbench__aero-formula code {
  color: #0284c7;
}

.tire-rim-engineering-workbench__aero-formula span {
  flex: 1 1 100%;
  line-height: 1.5;
}

.tire-rim-engineering-workbench__trap-number {
  display: inline-flex;
  padding: 0.2rem 0.38rem;
  border-radius: 0.25rem;
  background: #fee2e2;
  color: #b91c1c;
  font-family: var(--tz-font-ui);
  font-size: 0.56rem;
  font-weight: 900;
}

.tire-rim-engineering-workbench__traps-grid article strong {
  display: block;
  margin-top: 0.55rem;
  padding: 0.45rem;
  border-left: 3px solid #0284c7;
  border-radius: 0.3rem;
  background: rgb(239 246 255 / 0.8);
  color: #075985;
  font-size: 0.62rem;
  line-height: 1.45;
}

@media (max-width: 900px) {
  .tire-rim-engineering-workbench__controls,
  .tire-rim-engineering-workbench__result-layout,
  .tire-rim-engineering-workbench__board-metrics {
    grid-template-columns: 1fr;
  }

  .tire-rim-engineering-workbench__aero-grid,
  .tire-rim-engineering-workbench__traps-grid {
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
