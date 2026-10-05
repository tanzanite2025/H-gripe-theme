<template>
  <section class="inner-tube-fitment-guide w-full min-w-0 rounded-2xl bg-[var(--tz-card-surface)] p-4 shadow-md md:p-5">
    <div class="sizecharts-section__step-header sizecharts-section__step-header--compact">
      <span class="sizecharts-section__step-badge">4</span>
      <div>
        <h3 class="sizecharts-section__step-title">
          {{ t('guidesTireInnerTube.fitment.title') }}
        </h3>
        <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
          {{ t('guidesTireInnerTube.fitment.description') }}
        </p>
      </div>
    </div>

    <div class="mt-5 grid min-w-0 gap-5 xl:grid-cols-[minmax(0,0.92fr)_minmax(0,1.08fr)] xl:items-start">
      <div class="min-w-0 space-y-4">
        <div class="rounded-xl border tz-border-subtle tz-surface-subtle p-3">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
            <span class="text-xs font-bold uppercase tracking-wider tz-text-primary">
              {{ t('guidesTireInnerTube.fitment.controls.modeLabel') }}
            </span>
            <div class="inline-flex rounded-full border tz-border-subtle p-1">
              <button
                type="button"
                class="rounded-full px-3 py-1.5 text-[11px] font-bold transition-colors"
                :class="fitmentMode === 'automatic' ? 'bg-[var(--tz-action-primary)] text-[var(--tz-action-primary-foreground)]' : 'tz-text-secondary hover:tz-text-primary'"
                @click="setInnerTubeFitmentMode('automatic')"
              >
                {{ t('guidesTireInnerTube.fitment.controls.automaticMode') }}
              </button>
              <button
                type="button"
                class="rounded-full px-3 py-1.5 text-[11px] font-bold transition-colors"
                :class="fitmentMode === 'manual' ? 'bg-[var(--tz-action-primary)] text-[var(--tz-action-primary-foreground)]' : 'tz-text-secondary hover:tz-text-primary'"
                @click="setInnerTubeFitmentMode('manual')"
              >
                {{ t('guidesTireInnerTube.fitment.controls.manualMode') }}
              </button>
            </div>
          </div>

          <label class="block text-xs font-semibold tz-text-secondary" for="inner-tube-rim-depth-input">
            {{ t('guidesTireInnerTube.fitment.controls.rimDepth') }}
          </label>
          <div class="mt-2 flex flex-wrap items-center gap-3">
            <input
              id="inner-tube-rim-depth-range"
              v-model.number="rimDepth"
              :aria-label="t('guidesTireInnerTube.fitment.controls.rimDepth')"
              class="h-2 min-w-[10rem] flex-1 cursor-pointer accent-[var(--tz-action-primary)]"
              type="range"
              :min="rimDepthMinimum"
              :max="rimDepthMaximum"
              step="1"
              @change="() => requestCurrentInnerTubeFitmentResult()"
            >
            <div class="flex items-center gap-2">
              <input
                id="inner-tube-rim-depth-input"
                v-model="rimDepthInputValue"
                class="w-[5.5rem] rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] px-2.5 py-1.5 text-center font-mono text-xs font-bold tz-text-primary"
                type="number"
                inputmode="numeric"
                step="1"
                :min="rimDepthMinimum"
                :max="rimDepthMaximum"
                :aria-describedby="'inner-tube-rim-depth-hint'"
                @change="commitInnerTubeRimDepthInput"
                @keydown.enter.prevent="commitInnerTubeRimDepthInput"
              >
              <span class="font-mono text-xs font-bold tz-text-primary">mm</span>
            </div>
          </div>
          <p id="inner-tube-rim-depth-hint" class="mt-2 text-[11px] leading-relaxed tz-text-muted">
            {{ t('guidesTireInnerTube.fitment.controls.rimDepthRange', { min: rimDepthMinimum, max: rimDepthMaximum }) }}
          </p>
          <div class="mt-3 flex flex-wrap gap-1.5">
            <button
              v-for="preset in rimDepthPresets"
              :key="preset"
              type="button"
              class="rounded-full border px-2.5 py-1 text-[11px] font-semibold transition-colors"
              :class="rimDepth === preset ? 'border-[var(--tz-action-primary)] bg-[var(--tz-action-primary)] text-[var(--tz-action-primary-foreground)]' : 'tz-border-subtle tz-text-secondary hover:tz-text-primary'"
              @click="setInnerTubeRimDepthPreset(preset)"
            >
              {{ preset }} mm
            </button>
          </div>
          <div class="mt-3 rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] p-2.5">
            <div class="text-[11px] font-bold tz-text-primary">
              {{ t('guidesTireInnerTube.fitment.controls.measurementNoteTitle') }}
            </div>
            <p class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
              {{ t('guidesTireInnerTube.fitment.controls.measurementNote', { tolerance: rimDepthMeasurementTolerance }) }}
            </p>
          </div>

          <div class="mt-3 grid min-w-0 gap-3 sm:grid-cols-2">
            <div class="min-w-0 rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] p-2.5">
              <label class="block text-[11px] font-bold tz-text-primary" for="inner-tube-pump-head-grip-depth-input">
                {{ t('guidesTireInnerTube.fitment.controls.pumpHeadGripDepth') }}
              </label>
              <div class="mt-2 flex flex-wrap items-center gap-2">
                <input
                  id="inner-tube-pump-head-grip-depth-range"
                  v-model.number="pumpHeadGripDepth"
                  :aria-label="t('guidesTireInnerTube.fitment.controls.pumpHeadGripDepth')"
                  class="h-2 min-w-[7rem] flex-1 cursor-pointer accent-[var(--tz-action-primary)]"
                  type="range"
                  :min="pumpHeadGripDepthMinimum"
                  :max="pumpHeadGripDepthMaximum"
                  step="1"
                  @change="() => requestCurrentInnerTubeFitmentResult()"
                >
                <div class="flex items-center gap-1.5">
                  <input
                    id="inner-tube-pump-head-grip-depth-input"
                    v-model="pumpHeadGripDepthInputValue"
                    class="w-[4.5rem] rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] px-2 py-1.5 text-center font-mono text-xs font-bold tz-text-primary"
                    type="number"
                    inputmode="numeric"
                    step="1"
                    :min="pumpHeadGripDepthMinimum"
                    :max="pumpHeadGripDepthMaximum"
                    @change="commitInnerTubePumpHeadGripDepthInput"
                    @keydown.enter.prevent="commitInnerTubePumpHeadGripDepthInput"
                  >
                  <span class="font-mono text-xs font-bold tz-text-primary">mm</span>
                </div>
              </div>
              <p class="mt-1.5 text-[10px] leading-relaxed tz-text-muted">
                {{ t('guidesTireInnerTube.fitment.controls.pumpHeadGripDepthRange', { min: pumpHeadGripDepthMinimum, max: pumpHeadGripDepthMaximum }) }}
              </p>
            </div>

            <div class="min-w-0 rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] p-2.5">
              <label class="block text-[11px] font-bold tz-text-primary" for="inner-tube-rim-depth-uncertainty-input">
                {{ t('guidesTireInnerTube.fitment.controls.rimDepthUncertainty') }}
              </label>
              <div class="mt-2 flex flex-wrap items-center gap-2">
                <input
                  id="inner-tube-rim-depth-uncertainty-range"
                  v-model.number="rimDepthUncertainty"
                  :aria-label="t('guidesTireInnerTube.fitment.controls.rimDepthUncertainty')"
                  class="h-2 min-w-[7rem] flex-1 cursor-pointer accent-[var(--tz-action-primary)]"
                  type="range"
                  min="0"
                  :max="rimDepthUncertaintyMaximum"
                  step="1"
                  @change="() => requestCurrentInnerTubeFitmentResult()"
                >
                <div class="flex items-center gap-1.5">
                  <input
                    id="inner-tube-rim-depth-uncertainty-input"
                    v-model="rimDepthUncertaintyInputValue"
                    class="w-[4.5rem] rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] px-2 py-1.5 text-center font-mono text-xs font-bold tz-text-primary"
                    type="number"
                    inputmode="numeric"
                    step="1"
                    min="0"
                    :max="rimDepthUncertaintyMaximum"
                    @change="commitInnerTubeRimDepthUncertaintyInput"
                    @keydown.enter.prevent="commitInnerTubeRimDepthUncertaintyInput"
                  >
                  <span class="font-mono text-xs font-bold tz-text-primary">mm</span>
                </div>
              </div>
              <p class="mt-1.5 text-[10px] leading-relaxed tz-text-muted">
                {{ t('guidesTireInnerTube.fitment.controls.rimDepthUncertaintyRange', { max: rimDepthUncertaintyMaximum }) }}
              </p>
            </div>
          </div>
        </div>

        <div v-if="fitmentMode === 'manual'" class="grid gap-3 sm:grid-cols-2">
          <div class="rounded-xl border tz-border-subtle tz-surface-subtle p-3">
            <div class="text-xs font-semibold tz-text-secondary">
              {{ t('guidesTireInnerTube.fitment.controls.baseValveLength') }}
            </div>
            <div class="mt-2 flex flex-wrap gap-1.5">
              <button
                v-for="length in baseValveLengthOptions"
                :key="length"
                type="button"
                class="rounded-full border px-2.5 py-1 text-[11px] font-semibold transition-colors"
                :class="manualValveLength === length ? 'border-[var(--tz-action-primary)] bg-[var(--tz-action-primary)] text-[var(--tz-action-primary-foreground)]' : 'tz-border-subtle tz-text-secondary hover:tz-text-primary'"
                @click="selectManualInnerTubeValveLength(length)"
              >
                {{ length }} mm
              </button>
            </div>
          </div>

          <div class="rounded-xl border tz-border-subtle tz-surface-subtle p-3">
            <div class="text-xs font-semibold tz-text-secondary">
              {{ t('guidesTireInnerTube.fitment.controls.extenderLength') }}
            </div>
            <div class="mt-2 flex flex-wrap gap-1.5">
              <button
                v-for="length in extenderLengthOptions"
                :key="length"
                type="button"
                class="rounded-full border px-2.5 py-1 text-[11px] font-semibold transition-colors"
                :class="manualExtenderLength === length ? 'border-[var(--tz-action-primary)] bg-[var(--tz-action-primary)] text-[var(--tz-action-primary-foreground)]' : 'tz-border-subtle tz-text-secondary hover:tz-text-primary'"
                @click="selectManualInnerTubeExtenderLength(length)"
              >
                {{ length === 0 ? t('guidesTireInnerTube.fitment.controls.noExtender') : `+${length} mm` }}
              </button>
            </div>
          </div>
        </div>

        <div v-if="activeInnerTubeFitmentResult" class="inner-tube-fitment-status rounded-xl border p-3" :class="fitmentStatusClass" :aria-busy="innerTubeFitmentRequestPending">
          <div class="flex items-start gap-2">
            <span class="text-lg leading-none" aria-hidden="true">{{ fitmentStatusIcon }}</span>
            <div class="min-w-0">
              <div class="text-sm font-bold">
                {{ t(`guidesTireInnerTube.fitment.status.${fitmentStatus}.title`) }}
              </div>
              <p class="mt-1 text-xs leading-relaxed">
              {{ t(`guidesTireInnerTube.fitment.status.${fitmentStatus}.description`, {
                clearance: formatInnerTubeMillimetres(exposedValveLength),
                totalLength: `${totalAssemblyLength} mm`,
                gripDepth: pumpHeadGripDepthResult,
                preferredExposure,
              }) }}
              </p>
            </div>
          </div>
        </div>

        <p v-if="innerTubeFitmentMatrixPending || innerTubeFitmentRequestPending" class="mt-3 text-xs tz-text-muted" aria-live="polite">
          {{ t('guidesTireInnerTube.fitment.loading') }}
        </p>
        <p v-if="innerTubeFitmentMatrixError || innerTubeFitmentRequestError" class="mt-3 text-xs text-red-600" role="alert">
          {{ innerTubeFitmentRequestError || t('guidesTireInnerTube.fitment.error') }}
        </p>

        <div v-if="activeInnerTubeFitmentResult" class="grid min-w-0 grid-cols-1 gap-3 sm:grid-cols-2">
          <div class="min-w-0 rounded-xl border tz-border-subtle tz-surface-subtle p-3">
            <div class="text-[11px] font-semibold tz-text-muted">
              {{ t('guidesTireInnerTube.fitment.metrics.passageDepth') }}
            </div>
            <div class="mt-1 font-mono text-xl font-bold tz-text-primary">
              {{ formatInnerTubeMillimetres(passageDepth) }}
            </div>
            <div class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
              {{ t('guidesTireInnerTube.fitment.metrics.passageFormula', { rimDepth, offset: outerLipOffset }) }}
            </div>
          </div>
          <div class="min-w-0 rounded-xl border tz-border-subtle tz-surface-subtle p-3">
            <div class="text-[11px] font-semibold tz-text-muted">
              {{ t('guidesTireInnerTube.fitment.metrics.totalValveLength') }}
            </div>
            <div class="mt-1 font-mono text-xl font-bold tz-text-primary">
              {{ totalAssemblyLength }} mm
            </div>
            <div class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
              {{ activeExtenderLength > 0
                ? t('guidesTireInnerTube.fitment.metrics.totalValveWithExtender', { valve: activeValveLength, extender: activeExtenderLength })
                : t('guidesTireInnerTube.fitment.metrics.totalValveWithoutExtender') }}
            </div>
            <div class="mt-2 grid grid-cols-2 gap-2 border-t tz-border-subtle pt-2 text-[11px]">
              <div>
                <div class="tz-text-muted">{{ t('guidesTireInnerTube.fitment.metrics.minimumRequiredLength') }}</div>
                <div class="mt-0.5 font-mono font-bold tz-text-primary">{{ minimumRequiredLength }} mm</div>
              </div>
              <div>
                <div class="tz-text-muted">{{ t('guidesTireInnerTube.fitment.metrics.preferredMinimumLength') }}</div>
                <div class="mt-0.5 font-mono font-bold tz-text-primary">{{ preferredMinimumLength }} mm</div>
              </div>
            </div>
            <div class="mt-2 text-[11px] font-semibold" :class="fitmentStatusTextClass">
              {{ t('guidesTireInnerTube.fitment.metrics.lengthMargins', {
                minimumMargin: formatInnerTubeSignedMillimetres(minimumLengthMargin),
                preferredMargin: formatInnerTubeSignedMillimetres(preferredLengthMargin),
              }) }}
            </div>
            <div class="mt-2 grid grid-cols-2 gap-2 border-t tz-border-subtle pt-2 text-[11px]">
              <div>
                <div class="tz-text-muted">{{ t('guidesTireInnerTube.fitment.metrics.pumpHeadGripDepth') }}</div>
                <div class="mt-0.5 font-mono font-bold tz-text-primary">{{ pumpHeadGripDepthResult }} mm</div>
              </div>
              <div>
                <div class="tz-text-muted">{{ t('guidesTireInnerTube.fitment.metrics.preferredExposure') }}</div>
                <div class="mt-0.5 font-mono font-bold tz-text-primary">{{ formatInnerTubeMillimetres(preferredExposure) }}</div>
              </div>
            </div>
          </div>
          <div class="min-w-0 rounded-xl border tz-border-subtle tz-surface-subtle p-3">
            <div class="text-[11px] font-semibold tz-text-muted">
              {{ t('guidesTireInnerTube.fitment.metrics.exposedLength') }}
            </div>
            <div class="mt-1 font-mono text-xl font-bold" :class="fitmentStatusTextClass">
              {{ formatInnerTubeMillimetres(exposedValveLength) }}
            </div>
            <div class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
              {{ t('guidesTireInnerTube.fitment.metrics.exposedFormula') }}
            </div>
          </div>
          <div class="min-w-0 rounded-xl border tz-border-subtle tz-surface-subtle p-3">
            <div class="text-[11px] font-semibold tz-text-muted">
              {{ t('guidesTireInnerTube.fitment.metrics.recommendation') }}
            </div>
            <div class="mt-1 text-sm font-bold tz-text-primary">
              {{ activeRecommendation
                ? t(`guidesTireInnerTube.fitment.recommendations.${activeRecommendation.recommendation_key}`, {
                    valve: activeRecommendation.valve_length_mm,
                    extender: activeRecommendation.extender_length_mm,
                  })
                : t('guidesTireInnerTube.fitment.loading') }}
            </div>
            <div class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
              {{ fitmentMode === 'automatic'
                ? t('guidesTireInnerTube.fitment.metrics.automaticRecommendation')
                : t('guidesTireInnerTube.fitment.metrics.manualSelection') }}
            </div>
            <div v-if="activeInnerTubeFitmentResult" class="mt-2 rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] p-2 text-[11px] leading-relaxed tz-text-primary">
              {{ t(`guidesTireInnerTube.fitment.recommendationReasons.${activeRecommendationReasonKey}`, {
                minimumLength: minimumRequiredLength,
                totalLength: totalAssemblyLength,
              }) }}
            </div>
          </div>
        </div>

      </div>

      <div class="min-w-0 space-y-4">
        <div v-if="activeInnerTubeFitmentResult" class="overflow-hidden rounded-xl border tz-border-subtle tz-surface-subtle p-3">
          <div class="flex flex-wrap items-end justify-between gap-2">
            <div>
              <h4 class="text-sm font-bold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.visuals.title') }}
              </h4>
              <p class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
                {{ t('guidesTireInnerTube.fitment.visuals.description') }}
              </p>
            </div>
            <span class="rounded-full border tz-border-subtle px-2.5 py-1 text-[10px] font-bold tz-text-muted">
              {{ formatInnerTubeMillimetres(exposedValveLength) }}
            </span>
          </div>

          <div class="mt-3 flex min-h-[320px] items-center justify-start overflow-x-auto overflow-y-hidden rounded-lg bg-[var(--tz-card-surface)] sm:justify-center">
            <svg
            class="block h-auto max-h-[360px] min-w-[520px] w-full"
            viewBox="0 0 600 370"
            role="img"
            :aria-label="t('guidesTireInnerTube.fitment.visuals.ariaLabel')"
          >
            <defs>
              <linearGradient id="inner-tube-rim-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stop-color="#cbd5e1" />
                <stop offset="50%" stop-color="#f1f5f9" />
                <stop offset="100%" stop-color="#cbd5e1" />
              </linearGradient>
              <linearGradient id="inner-tube-valve-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stop-color="#1e293b" />
                <stop offset="50%" stop-color="#64748b" />
                <stop offset="100%" stop-color="#0f172a" />
              </linearGradient>
              <linearGradient id="inner-tube-extender-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stop-color="#2563eb" />
                <stop offset="50%" stop-color="#60a5fa" />
                <stop offset="100%" stop-color="#1d4ed8" />
              </linearGradient>
            </defs>

            <line x1="215" :y1="innerTubeSvgBaseY" x2="385" :y2="innerTubeSvgBaseY" stroke="var(--tz-site-accent)" stroke-width="1.5" stroke-dasharray="4,4" />
            <line x1="215" :y1="innerTubeSvgOuterLipY" x2="385" :y2="innerTubeSvgOuterLipY" stroke="#94a3b8" stroke-width="1" stroke-dasharray="3,3" />

            <rect x="250" :y="innerTubeSvgRimExitY" width="100" :height="innerTubeSvgRimHeight" fill="url(#inner-tube-rim-gradient)" stroke="#64748b" stroke-width="2" rx="4" />

            <polygon points="284,300 316,300 310,293 290,293" fill="#0f172a" stroke="#000000" stroke-width="1" />

            <rect x="296" :y="innerTubeSvgValveStemTopY" width="8" :height="innerTubeSvgValveHeight" fill="url(#inner-tube-valve-gradient)" stroke="#09090b" stroke-width="1" rx="2" />

            <template v-if="activeExtenderLength > 0">
              <rect x="295" :y="innerTubeSvgExtenderTopY" width="10" :height="innerTubeSvgExtenderHeight" fill="url(#inner-tube-extender-gradient)" stroke="#1d4ed8" stroke-width="1" rx="2" />
              <line x1="293" :y1="innerTubeSvgValveStemTopY" x2="307" :y2="innerTubeSvgValveStemTopY" stroke="#09090b" stroke-width="2" />
            </template>

            <rect x="294" :y="innerTubeSvgTopY - 12" width="12" height="12" fill="#fbbf24" stroke="#78350f" stroke-width="1" rx="2" />
            <line x1="300" :y1="innerTubeSvgTopY - 12" x2="300" :y2="innerTubeSvgTopY - 20" stroke="#78350f" stroke-width="2" />

            <line x1="370" :y1="innerTubeSvgRimExitY" x2="370" :y2="innerTubeSvgTopY" :stroke="fitmentStatusSvgColor" stroke-width="1.5" stroke-dasharray="4,4" />
            <line x1="364" :y1="innerTubeSvgTopY" x2="376" :y2="innerTubeSvgTopY" :stroke="fitmentStatusSvgColor" stroke-width="2" />
            <line x1="364" :y1="innerTubeSvgRimExitY" x2="376" :y2="innerTubeSvgRimExitY" :stroke="fitmentStatusSvgColor" stroke-width="2" />

            <text x="20" y="39" class="inner-tube-fitment-svg-label inner-tube-fitment-svg-label--strong">
              {{ t('guidesTireInnerTube.fitment.visuals.valveStem', { valve: activeValveLength }) }}
            </text>
            <text v-if="activeExtenderLength > 0" x="20" y="58" class="inner-tube-fitment-svg-label inner-tube-fitment-svg-label--extender">
              {{ t('guidesTireInnerTube.fitment.visuals.extender', { extender: activeExtenderLength }) }}
            </text>

            <line x1="182" :y1="innerTubeSvgRimLabelY" x2="240" :y2="innerTubeSvgRimLabelY" stroke="#94a3b8" stroke-width="1" stroke-dasharray="3,3" />
            <text x="20" :y="innerTubeSvgRimLabelY - 4" class="inner-tube-fitment-svg-value">
              {{ formatInnerTubeMillimetres(passageDepth) }}
            </text>
            <text x="20" :y="innerTubeSvgRimLabelY + 13" class="inner-tube-fitment-svg-label">
              {{ t('guidesTireInnerTube.fitment.visuals.passageLabel', { rimDepth }) }}
            </text>

            <line x1="160" y1="330" x2="184" y2="330" stroke="var(--tz-site-accent)" stroke-width="1.5" stroke-dasharray="4,3" />
            <text x="20" y="334" class="inner-tube-fitment-svg-label inner-tube-fitment-svg-label--accent">
              {{ t('guidesTireInnerTube.fitment.visuals.rimBase') }}
            </text>
            <line x1="160" y1="350" x2="184" y2="350" stroke="#94a3b8" stroke-width="1" stroke-dasharray="3,3" />
            <text x="20" y="354" class="inner-tube-fitment-svg-label inner-tube-fitment-svg-label--muted">
              {{ t('guidesTireInnerTube.fitment.visuals.outerLip') }}
            </text>

            <text x="400" :y="innerTubeSvgMeasurementMidY - 3" class="inner-tube-fitment-svg-value" :fill="fitmentStatusSvgColor">
              {{ t('guidesTireInnerTube.fitment.visuals.exposed', { clearance: formatInnerTubeMillimetres(exposedValveLength) }) }}
            </text>
            <text x="400" :y="innerTubeSvgMeasurementMidY + 16" class="inner-tube-fitment-svg-label">
              {{ t(`guidesTireInnerTube.fitment.status.${fitmentStatus}.title`) }}
            </text>
            </svg>
          </div>
        </div>
        <div v-else-if="!innerTubeFitmentMatrixError && !innerTubeFitmentRequestError" class="rounded-xl border tz-border-subtle tz-surface-subtle p-6 text-center text-xs tz-text-muted">
          {{ t('guidesTireInnerTube.fitment.loading') }}
        </div>

        <div v-if="uncertaintyReview" class="rounded-xl border tz-border-subtle tz-surface-subtle p-3">
          <div class="flex flex-wrap items-start justify-between gap-2">
            <div>
              <h4 class="text-sm font-bold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.uncertainty.title') }}
              </h4>
              <p class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
                {{ t('guidesTireInnerTube.fitment.uncertainty.description', {
                  lower: uncertaintyReview.lower_bound.rim_depth_mm,
                  upper: uncertaintyReview.upper_bound.rim_depth_mm,
                  uncertainty: uncertaintyReview.rim_depth_uncertainty_mm,
                }) }}
              </p>
            </div>
            <span class="rounded-full border px-2.5 py-1 text-[10px] font-bold" :class="`inner-tube-fitment-text--${uncertaintyReview.worst_case_status}`">
              {{ t(`guidesTireInnerTube.fitment.status.${uncertaintyReview.worst_case_status}.title`) }}
            </span>
          </div>
          <div class="mt-3 grid min-w-0 gap-2 sm:grid-cols-2">
            <div class="min-w-0 rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] p-2.5">
              <div class="text-[11px] font-bold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.uncertainty.lowerBound', { rimDepth: uncertaintyReview.lower_bound.rim_depth_mm }) }}
              </div>
              <div class="mt-1 grid grid-cols-2 gap-2 text-[11px]">
                <div>
                  <div class="tz-text-muted">{{ t('guidesTireInnerTube.fitment.uncertainty.exposedLength') }}</div>
                  <div class="mt-0.5 font-mono font-bold" :class="`inner-tube-fitment-text--${uncertaintyReview.lower_bound.status}`">
                    {{ formatInnerTubeMillimetres(uncertaintyReview.lower_bound.effective_exposure_mm) }}
                  </div>
                </div>
                <div>
                  <div class="tz-text-muted">{{ t('guidesTireInnerTube.fitment.uncertainty.minimumMargin') }}</div>
                  <div class="mt-0.5 font-mono font-bold tz-text-primary">
                    {{ formatInnerTubeSignedMillimetres(uncertaintyReview.lower_bound.minimum_length_margin_mm) }}
                  </div>
                </div>
              </div>
              <div class="mt-2 text-[11px] font-semibold" :class="`inner-tube-fitment-text--${uncertaintyReview.lower_bound.status}`">
                {{ t(`guidesTireInnerTube.fitment.status.${uncertaintyReview.lower_bound.status}.title`) }}
              </div>
            </div>
            <div class="min-w-0 rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] p-2.5">
              <div class="text-[11px] font-bold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.uncertainty.upperBound', { rimDepth: uncertaintyReview.upper_bound.rim_depth_mm }) }}
              </div>
              <div class="mt-1 grid grid-cols-2 gap-2 text-[11px]">
                <div>
                  <div class="tz-text-muted">{{ t('guidesTireInnerTube.fitment.uncertainty.exposedLength') }}</div>
                  <div class="mt-0.5 font-mono font-bold" :class="`inner-tube-fitment-text--${uncertaintyReview.upper_bound.status}`">
                    {{ formatInnerTubeMillimetres(uncertaintyReview.upper_bound.effective_exposure_mm) }}
                  </div>
                </div>
                <div>
                  <div class="tz-text-muted">{{ t('guidesTireInnerTube.fitment.uncertainty.minimumMargin') }}</div>
                  <div class="mt-0.5 font-mono font-bold tz-text-primary">
                    {{ formatInnerTubeSignedMillimetres(uncertaintyReview.upper_bound.minimum_length_margin_mm) }}
                  </div>
                </div>
              </div>
              <div class="mt-2 text-[11px] font-semibold" :class="`inner-tube-fitment-text--${uncertaintyReview.upper_bound.status}`">
                {{ t(`guidesTireInnerTube.fitment.status.${uncertaintyReview.upper_bound.status}.title`) }}
              </div>
            </div>
          </div>
          <div class="mt-3 rounded-lg border p-2.5 text-[11px] leading-relaxed" :class="uncertaintyReview.is_status_stable ? 'border-emerald-200 bg-emerald-50 text-emerald-800' : 'border-amber-200 bg-amber-50 text-amber-800'">
            {{ uncertaintyReview.is_status_stable
              ? t('guidesTireInnerTube.fitment.uncertainty.stable', { status: t(`guidesTireInnerTube.fitment.status.${uncertaintyReview.lower_bound.status}.title`) })
              : t('guidesTireInnerTube.fitment.uncertainty.unstable', { status: t(`guidesTireInnerTube.fitment.status.${uncertaintyReview.worst_case_status}.title`) }) }}
          </div>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-2 rounded-xl border tz-border-subtle tz-surface-subtle p-3">
          <div>
            <div class="text-xs font-bold tz-text-primary">
              {{ t('guidesTireInnerTube.fitment.share.title') }}
            </div>
            <p class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
              {{ t('guidesTireInnerTube.fitment.share.description') }}
            </p>
          </div>
          <button
            type="button"
            class="rounded-lg bg-[var(--tz-action-primary)] px-3 py-2 text-[11px] font-bold text-[var(--tz-action-primary-foreground)] transition-opacity hover:opacity-90"
            @click="copyInnerTubeFitmentShareLink"
          >
            {{ sharedCalculationCopied ? t('guidesTireInnerTube.fitment.share.copied') : t('guidesTireInnerTube.fitment.share.copy') }}
          </button>
          <p v-if="sharedCalculationCopyError" class="basis-full text-[11px] text-red-600" role="alert">
            {{ t('guidesTireInnerTube.fitment.share.error') }}
          </p>
        </div>

        <div v-if="activeInnerTubeFitmentResult?.alternatives?.length" class="rounded-xl border tz-border-subtle tz-surface-subtle p-3">
          <div class="flex flex-wrap items-end justify-between gap-2">
            <div>
              <h4 class="text-sm font-bold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.alternatives.title') }}
              </h4>
              <p class="mt-1 text-[11px] leading-relaxed tz-text-secondary">
                {{ t('guidesTireInnerTube.fitment.alternatives.description') }}
              </p>
            </div>
            <span class="rounded-full border tz-border-subtle px-2.5 py-1 text-[10px] font-bold tz-text-muted">
              {{ t('guidesTireInnerTube.fitment.alternatives.referenceBadge') }}
            </span>
          </div>
          <div class="mt-3 grid min-w-0 gap-2 sm:grid-cols-2">
            <button
              v-for="alternative in activeInnerTubeFitmentResult.alternatives"
              :key="`${alternative.valve_length_mm}-${alternative.extender_length_mm}`"
              type="button"
              class="min-w-0 rounded-lg border p-2.5 text-left transition-colors hover:tz-surface-muted"
              :class="alternative.is_automatic_recommendation ? 'border-[var(--tz-action-primary)] tz-surface-muted' : 'tz-border-subtle bg-[var(--tz-card-surface)]'"
              @click="selectInnerTubeFitmentAlternative(alternative)"
            >
              <div class="flex flex-wrap items-center justify-between gap-1.5">
                <span class="text-xs font-bold tz-text-primary">
                  {{ alternative.valve_length_mm }} mm
                  <template v-if="alternative.extender_length_mm > 0">
                    + {{ alternative.extender_length_mm }} mm
                  </template>
                  {{ t('guidesTireInnerTube.fitment.alternatives.assemblySuffix') }}
                </span>
                <span v-if="alternative.is_automatic_recommendation" class="rounded-full bg-[var(--tz-action-primary)] px-2 py-0.5 text-[10px] font-bold text-[var(--tz-action-primary-foreground)]">
                  {{ t('guidesTireInnerTube.fitment.alternatives.automaticBadge') }}
                </span>
              </div>
              <div class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[11px] tz-text-secondary">
                <span>{{ t('guidesTireInnerTube.fitment.alternatives.total', { total: alternative.total_assembly_length_mm }) }}</span>
                <span :class="`inner-tube-fitment-text--${alternative.status}`">
                  {{ t(`guidesTireInnerTube.fitment.status.${alternative.status}.title`) }} · {{ formatInnerTubeMillimetres(alternative.effective_exposure_mm) }}
                </span>
              </div>
              <div class="mt-1 text-[10px] tz-text-muted">
                {{ t('guidesTireInnerTube.fitment.alternatives.margin', { margin: formatInnerTubeSignedMillimetres(alternative.minimum_length_margin_mm) }) }}
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="innerTubeFitmentMetadata" class="mt-5 rounded-xl border tz-border-subtle bg-[var(--tz-card-surface)] p-3">
      <div class="flex flex-wrap items-end justify-between gap-2">
        <div>
          <h4 class="text-sm font-bold tz-text-primary">
            {{ t('guidesTireInnerTube.fitment.matrix.title') }}
          </h4>
          <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
            {{ t('guidesTireInnerTube.fitment.matrix.description') }}
          </p>
        </div>
        <span class="rounded-full border tz-border-subtle px-2.5 py-1 text-[10px] font-bold tz-text-muted">
          {{ t('guidesTireInnerTube.fitment.matrix.referenceBadge') }}
        </span>
      </div>
      <div class="mt-3 overflow-x-auto rounded-lg border tz-border-subtle">
        <table class="min-w-[820px] w-full text-center text-xs tz-text-secondary">
          <thead class="tz-surface-muted">
            <tr>
              <th class="px-3 py-2 text-left font-semibold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.matrix.rimDepth') }}
              </th>
              <th class="px-3 py-2 font-semibold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.matrix.passageDepth') }}
              </th>
              <th class="px-3 py-2 font-semibold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.matrix.minimumRequiredLength') }}
              </th>
              <th
                v-for="length in baseValveLengthOptions"
                :key="length"
                class="px-3 py-2 font-semibold tz-text-primary"
              >
                {{ length }} mm
              </th>
              <th class="px-3 py-2 font-semibold tz-text-primary">
                {{ t('guidesTireInnerTube.fitment.matrix.recommendation') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y tz-border-subtle">
            <tr
              v-for="row in fitmentMatrixRows"
              :key="row.rimDepth"
              class="cursor-pointer transition-colors hover:tz-surface-subtle"
              :class="rimDepth === row.rimDepth ? 'tz-surface-subtle' : ''"
              @click="setInnerTubeRimDepthPreset(row.rimDepth)"
            >
              <td class="px-3 py-2 text-left font-semibold tz-text-primary">
                {{ row.rimDepth }} mm
              </td>
              <td class="px-3 py-2 font-mono">
                {{ formatInnerTubeMillimetres(row.passageDepth) }}
              </td>
              <td class="px-3 py-2 font-mono">
                {{ row.minimumRequiredLength }} mm
              </td>
              <td
                v-for="cell in row.clearances"
                :key="cell.valveLength"
                class="px-3 py-2 font-mono"
                :class="getInnerTubeClearanceClass(cell.status)"
              >
                {{ formatInnerTubeMillimetres(cell.clearance) }}
              </td>
              <td class="px-3 py-2 text-left font-semibold tz-text-primary">
                {{ t(`guidesTireInnerTube.fitment.recommendations.${row.recommendationKey}`, {
                  valve: row.recommendationValveLength,
                  extender: row.recommendationExtenderLength,
                }) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div v-else-if="!innerTubeFitmentMatrixError && !innerTubeFitmentRequestError" class="mt-5 rounded-xl border tz-border-subtle tz-surface-subtle p-6 text-center text-xs tz-text-muted">
      {{ t('guidesTireInnerTube.fitment.loading') }}
    </div>

    <div class="mt-5 grid min-w-0 gap-4 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)] lg:items-start">
      <div class="min-w-0 rounded-xl border tz-border-subtle tz-surface-subtle p-3">
        <h4 class="text-sm font-bold tz-text-primary">
          {{ t('guidesTireInnerTube.fitment.structure.title') }}
        </h4>
        <div class="mt-3 flex flex-wrap gap-1.5">
          <button
            v-for="structure in innerTubeValveStructureOptions"
            :key="structure"
            type="button"
            class="rounded-full border px-2.5 py-1.5 text-[11px] font-semibold transition-colors"
            :class="activeValveStructure === structure ? 'border-[var(--tz-action-primary)] bg-[var(--tz-action-primary)] text-[var(--tz-action-primary-foreground)]' : 'tz-border-subtle tz-text-secondary hover:tz-text-primary'"
            @click="selectInnerTubeValveStructure(structure)"
          >
            {{ t(`guidesTireInnerTube.fitment.structure.options.${structure}.label`) }}
          </button>
        </div>
        <div class="mt-3 rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] p-2">
          <svg
            class="h-auto w-full"
            viewBox="0 0 360 180"
            role="img"
            :aria-label="t(`guidesTireInnerTube.fitment.structure.options.${activeValveStructure}.title`)"
          >
            <path d="M40 135 Q180 108 320 135 L320 148 Q180 121 40 148 Z" fill="#0f172a" stroke="#000000" stroke-width="1.5" />
            <text x="48" y="165" class="inner-tube-fitment-svg-label">
              {{ t('guidesTireInnerTube.fitment.visuals.tubeBody') }}
            </text>
            <path d="M150 123 Q180 113 210 123 L205 132 Q180 125 155 132 Z" fill="#334155" stroke="#1e293b" stroke-width="1" />
            <rect x="176" y="48" width="8" height="78" fill="url(#inner-tube-valve-gradient)" stroke="#09090b" stroke-width="1" rx="1" />
            <rect x="173" y="30" width="14" height="18" :fill="innerTubeStructureSvgCoreColor" stroke="#78350f" stroke-width="1.2" rx="1" />
            <line x1="180" y1="30" x2="180" y2="20" stroke="#78350f" stroke-width="2" />
            <template v-if="activeValveStructure === 'rvcExtender'">
              <rect x="172" y="18" width="16" height="12" fill="url(#inner-tube-extender-gradient)" stroke="#1d4ed8" stroke-width="1" rx="1" />
              <rect x="174" y="4" width="12" height="14" fill="#fbbf24" stroke="#78350f" stroke-width="1" rx="1" />
            </template>
            <template v-if="activeValveStructure === 'rootProtection'">
              <ellipse cx="180" cy="125" rx="22" ry="5" fill="none" stroke="#38bdf8" stroke-width="4" />
            </template>
            <line x1="205" y1="84" x2="255" y2="84" stroke="var(--tz-site-accent)" stroke-width="1" stroke-dasharray="3,3" />
            <text x="260" y="88" class="inner-tube-fitment-svg-label inner-tube-fitment-svg-label--accent">
              {{ t(`guidesTireInnerTube.fitment.structure.options.${activeValveStructure}.label`) }}
            </text>
          </svg>
        </div>
        <div class="mt-3 rounded-lg border tz-border-subtle bg-[var(--tz-card-surface)] p-3">
          <div class="text-xs font-bold tz-text-primary">
            {{ t(`guidesTireInnerTube.fitment.structure.options.${activeValveStructure}.title`) }}
          </div>
          <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
            {{ t(`guidesTireInnerTube.fitment.structure.options.${activeValveStructure}.description`) }}
          </p>
          <p class="mt-2 text-[11px] font-semibold leading-relaxed tz-text-primary">
            {{ t(`guidesTireInnerTube.fitment.structure.options.${activeValveStructure}.rule`) }}
          </p>
        </div>
      </div>

      <div class="min-w-0">
        <h4 class="text-sm font-bold tz-text-primary">
          {{ t('guidesTireInnerTube.fitment.pitfalls.title') }}
        </h4>
        <div class="mt-3 grid min-w-0 gap-3 sm:grid-cols-2">
          <article
            v-for="pitfall in innerTubeFitmentPitfalls"
            :key="pitfall"
            class="min-w-0 rounded-xl border tz-border-subtle p-3"
          >
            <div class="text-xs font-bold tz-text-primary">
              {{ t(`guidesTireInnerTube.fitment.pitfalls.items.${pitfall}.title`) }}
            </div>
            <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
              {{ t(`guidesTireInnerTube.fitment.pitfalls.items.${pitfall}.description`) }}
            </p>
            <p class="mt-2 text-[11px] font-semibold leading-relaxed tz-text-primary">
              {{ t(`guidesTireInnerTube.fitment.pitfalls.items.${pitfall}.rule`) }}
            </p>
          </article>
        </div>
      </div>
    </div>

    <p class="mt-4 text-[11px] leading-relaxed tz-text-muted">
      {{ t('guidesTireInnerTube.fitment.disclaimer') }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n, useHead, useRoute, useRouter } from '#imports'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import {
  useInnerTubeValveFitmentCalculator,
  type InnerTubeValveFitmentAlternative,
  type InnerTubeValveFitmentMode,
  type InnerTubeValveFitmentResult,
  type InnerTubeValveFitmentStatus,
  type InnerTubeValveStructure,
} from '~/composables/useInnerTubeValveFitmentCalculator'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const { canonicalUrl } = useStorefrontSeoLinks()
const {
  metadata: innerTubeFitmentMetadata,
  defaultResult: defaultInnerTubeFitmentResult,
  matrixPending: innerTubeFitmentMatrixPending,
  matrixError: innerTubeFitmentMatrixError,
  solveInnerTubeValveFitmentWithBackend,
} = await useInnerTubeValveFitmentCalculator()

const innerTubeFitmentSchemaLanguage = computed(() => locale.value === 'zh_cn' ? 'zh-CN' : locale.value)

const innerTubeValveStructureOptions: readonly InnerTubeValveStructure[] = [
  'removableCore',
  'fixedCore',
  'rvcExtender',
  'rootProtection',
]
const innerTubeFitmentPitfalls = ['shortValve', 'overTightenedLocknut', 'fixedCoreExtender', 'carbonHole'] as const

const fitmentMode = ref<InnerTubeValveFitmentMode>('automatic')
const rimDepth = ref(50)
const rimDepthInputValue = ref(String(rimDepth.value))
const manualValveLength = ref(60)
const manualExtenderLength = ref(0)
const pumpHeadGripDepth = ref(innerTubeFitmentMetadata.value?.default_pump_head_grip_depth_mm ?? 15)
const pumpHeadGripDepthInputValue = ref(String(pumpHeadGripDepth.value))
const rimDepthUncertainty = ref(innerTubeFitmentMetadata.value?.default_rim_depth_uncertainty_mm ?? 2)
const rimDepthUncertaintyInputValue = ref(String(rimDepthUncertainty.value))
const activeValveStructure = ref<InnerTubeValveStructure>('removableCore')
const activeInnerTubeFitmentResult = ref<InnerTubeValveFitmentResult | null>(defaultInnerTubeFitmentResult.value)
const innerTubeFitmentRequestPending = ref(false)
const innerTubeFitmentRequestError = ref<string | null>(null)
const sharedCalculationCopied = ref(false)
const sharedCalculationCopyError = ref(false)
let innerTubeFitmentRequestSequence = 0

const formatInnerTubeMillimetres = (value: number) => `${value.toFixed(1)} mm`
const formatInnerTubeSignedMillimetres = (value: number) => `${value >= 0 ? '+' : ''}${value.toFixed(1)} mm`

const rimDepthMinimum = computed(() => innerTubeFitmentMetadata.value?.rim_depth_min_mm ?? 20)
const rimDepthMaximum = computed(() => innerTubeFitmentMetadata.value?.rim_depth_max_mm ?? 100)
const rimDepthPresets = computed(() => innerTubeFitmentMetadata.value?.preset_rim_depths_mm ?? [])
const baseValveLengthOptions = computed(() => innerTubeFitmentMetadata.value?.base_valve_length_options_mm ?? [])
const extenderLengthOptions = computed(() => innerTubeFitmentMetadata.value?.extender_length_options_mm ?? [])
const pumpHeadGripDepthMinimum = computed(() => innerTubeFitmentMetadata.value?.pump_head_grip_depth_min_mm ?? 10)
const pumpHeadGripDepthMaximum = computed(() => innerTubeFitmentMetadata.value?.pump_head_grip_depth_max_mm ?? 30)
const rimDepthUncertaintyMaximum = computed(() => innerTubeFitmentMetadata.value?.rim_depth_uncertainty_max_mm ?? 5)
const rimDepthMeasurementTolerance = computed(() => rimDepthUncertainty.value)
const outerLipOffset = computed(() => activeInnerTubeFitmentResult.value?.outer_lip_offset_mm ?? innerTubeFitmentMetadata.value?.outer_lip_offset_mm ?? 0)
const activeRecommendation = computed(() => activeInnerTubeFitmentResult.value?.recommendation ?? null)
const activeRecommendationReasonKey = computed(() => activeInnerTubeFitmentResult.value?.recommendation_reason_key ?? activeRecommendation.value?.recommendation_reason_key ?? 'manualCombination')
const activeValveLength = computed(() => activeInnerTubeFitmentResult.value?.valve_length_mm ?? 0)
const activeExtenderLength = computed(() => activeInnerTubeFitmentResult.value?.extender_length_mm ?? 0)
const totalAssemblyLength = computed(() => activeInnerTubeFitmentResult.value?.total_assembly_length_mm ?? 0)
const passageDepth = computed(() => activeInnerTubeFitmentResult.value?.passage_depth_mm ?? 0)
const exposedValveLength = computed(() => activeInnerTubeFitmentResult.value?.effective_exposure_mm ?? 0)
const minimumRequiredLength = computed(() => activeInnerTubeFitmentResult.value?.minimum_required_length_mm ?? 0)
const preferredMinimumLength = computed(() => activeInnerTubeFitmentResult.value?.preferred_minimum_length_mm ?? 0)
const minimumLengthMargin = computed(() => activeInnerTubeFitmentResult.value?.minimum_length_margin_mm ?? (totalAssemblyLength.value - minimumRequiredLength.value))
const preferredLengthMargin = computed(() => activeInnerTubeFitmentResult.value?.preferred_length_margin_mm ?? (totalAssemblyLength.value - preferredMinimumLength.value))
const pumpHeadGripDepthResult = computed(() => activeInnerTubeFitmentResult.value?.pump_head_grip_depth_mm ?? pumpHeadGripDepth.value)
const preferredExposure = computed(() => activeInnerTubeFitmentResult.value?.preferred_exposure_mm ?? 0)
const uncertaintyReview = computed(() => activeInnerTubeFitmentResult.value?.uncertainty_review ?? null)
const fitmentStatus = computed<InnerTubeValveFitmentStatus>(() => activeInnerTubeFitmentResult.value?.status ?? 'unsafe')
const fitmentStatusClass = computed(() => `inner-tube-fitment-status--${fitmentStatus.value}`)
const fitmentStatusTextClass = computed(() => `inner-tube-fitment-text--${fitmentStatus.value}`)
const fitmentStatusSvgColor = computed(() => {
  if (fitmentStatus.value === 'optimal') return '#059669'
  if (fitmentStatus.value === 'marginal') return '#b45309'
  return '#dc2626'
})
const fitmentStatusIcon = computed(() => {
  if (fitmentStatus.value === 'optimal') return '✅'
  if (fitmentStatus.value === 'marginal') return '⚠️'
  return '🚨'
})

const innerTubeSvgBaseY = 300
const innerTubeSvgScale = 1.65
const innerTubeSvgOuterLipY = computed(() => innerTubeSvgBaseY + (outerLipOffset.value * innerTubeSvgScale))
const innerTubeSvgRimExitY = computed(() => innerTubeSvgBaseY - (passageDepth.value * innerTubeSvgScale))
const innerTubeSvgRimHeight = computed(() => passageDepth.value * innerTubeSvgScale)
const innerTubeSvgRimLabelY = computed(() => innerTubeSvgRimExitY.value + (innerTubeSvgRimHeight.value / 2))
const innerTubeSvgValveStemTopY = computed(() => innerTubeSvgBaseY - (activeValveLength.value * innerTubeSvgScale))
const innerTubeSvgValveHeight = computed(() => activeValveLength.value * innerTubeSvgScale)
const innerTubeSvgExtenderTopY = computed(() => innerTubeSvgValveStemTopY.value - (activeExtenderLength.value * innerTubeSvgScale))
const innerTubeSvgExtenderHeight = computed(() => activeExtenderLength.value * innerTubeSvgScale)
const innerTubeSvgTopY = computed(() => activeExtenderLength.value > 0
  ? innerTubeSvgExtenderTopY.value
  : innerTubeSvgValveStemTopY.value)
const innerTubeSvgMeasurementMidY = computed(() => (innerTubeSvgRimExitY.value + innerTubeSvgTopY.value) / 2)
const innerTubeStructureSvgCoreColor = computed(() => {
  if (activeValveStructure.value === 'fixedCore') return '#dc2626'
  if (activeValveStructure.value === 'rvcExtender') return '#2563eb'
  if (activeValveStructure.value === 'rootProtection') return '#38bdf8'
  return '#d97706'
})

const fitmentMatrixRows = computed(() => (innerTubeFitmentMetadata.value?.rows ?? []).map(row => ({
  rimDepth: row.rim_depth_mm,
  passageDepth: row.passage_depth_mm,
  minimumRequiredLength: row.minimum_required_length_mm,
  recommendationKey: row.recommended_result.recommendation_key,
  recommendationValveLength: row.recommended_result.recommendation.valve_length_mm,
  recommendationExtenderLength: row.recommended_result.recommendation.extender_length_mm,
  clearances: row.clearances.map(cell => ({
    valveLength: cell.valve_length_mm,
    clearance: cell.effective_exposure_mm,
    status: cell.status,
  })),
})))

const getInnerTubeClearanceClass = (status: InnerTubeValveFitmentStatus) => {
  return `inner-tube-fitment-cell--${status}`
}

watch(rimDepth, (value) => {
  rimDepthInputValue.value = String(value)
})

watch(pumpHeadGripDepth, (value) => {
  pumpHeadGripDepthInputValue.value = String(value)
})

watch(rimDepthUncertainty, (value) => {
  rimDepthUncertaintyInputValue.value = String(value)
})

const getInnerTubeShareQueryValue = (queryKey: string) => {
  const queryValue = route.query[queryKey]
  if (Array.isArray(queryValue)) return queryValue[0] ?? null
  return queryValue ?? null
}

const normalizeInnerTubeShareInteger = (value: string | null, minimum: number, maximum: number, fallback: number) => {
  if (value === null) return fallback
  const parsedValue = Number(value)
  if (!Number.isFinite(parsedValue)) return fallback
  return Math.min(maximum, Math.max(minimum, Math.round(parsedValue)))
}

const restoreInnerTubeFitmentStateFromShareUrl = () => {
  const sharedRimDepth = getInnerTubeShareQueryValue('rim_depth')
  const sharedFitmentMode = getInnerTubeShareQueryValue('fitment_mode')
  const sharedValveLength = getInnerTubeShareQueryValue('valve_length')
  const sharedExtenderLength = getInnerTubeShareQueryValue('extender_length')
  const sharedPumpHeadGripDepth = getInnerTubeShareQueryValue('pump_head_grip_depth')
  const sharedRimDepthUncertainty = getInnerTubeShareQueryValue('rim_depth_uncertainty')

  if (sharedRimDepth !== null) {
    rimDepth.value = normalizeInnerTubeShareInteger(sharedRimDepth, rimDepthMinimum.value, rimDepthMaximum.value, rimDepth.value)
  }
  if (sharedFitmentMode === 'automatic' || sharedFitmentMode === 'manual') {
    fitmentMode.value = sharedFitmentMode
  }

  const supportedValveLengths = baseValveLengthOptions.value.length > 0 ? baseValveLengthOptions.value : [40, 48, 60, 80]
  const supportedExtenderLengths = extenderLengthOptions.value.length > 0 ? extenderLengthOptions.value : [0, 20, 30, 40, 60]
  const parsedValveLength = normalizeInnerTubeShareInteger(sharedValveLength, Math.min(...supportedValveLengths), Math.max(...supportedValveLengths), manualValveLength.value)
  const parsedExtenderLength = normalizeInnerTubeShareInteger(sharedExtenderLength, Math.min(...supportedExtenderLengths), Math.max(...supportedExtenderLengths), manualExtenderLength.value)
  if (supportedValveLengths.includes(parsedValveLength)) manualValveLength.value = parsedValveLength
  if (supportedExtenderLengths.includes(parsedExtenderLength)) manualExtenderLength.value = parsedExtenderLength

  pumpHeadGripDepth.value = normalizeInnerTubeShareInteger(
    sharedPumpHeadGripDepth,
    pumpHeadGripDepthMinimum.value,
    pumpHeadGripDepthMaximum.value,
    pumpHeadGripDepth.value,
  )
  rimDepthUncertainty.value = normalizeInnerTubeShareInteger(
    sharedRimDepthUncertainty,
    0,
    rimDepthUncertaintyMaximum.value,
    rimDepthUncertainty.value,
  )
  rimDepthInputValue.value = String(rimDepth.value)
  pumpHeadGripDepthInputValue.value = String(pumpHeadGripDepth.value)
  rimDepthUncertaintyInputValue.value = String(rimDepthUncertainty.value)

  return [
    sharedRimDepth,
    sharedFitmentMode,
    sharedValveLength,
    sharedExtenderLength,
    sharedPumpHeadGripDepth,
    sharedRimDepthUncertainty,
  ].some(value => value !== null)
}

const initialInnerTubeFitmentStateWasRestored = restoreInnerTubeFitmentStateFromShareUrl()
if (initialInnerTubeFitmentStateWasRestored) {
  activeInnerTubeFitmentResult.value = null
}

const updateInnerTubeFitmentShareUrl = async () => {
  if (!import.meta.client) return
  await router.replace({
    query: {
      ...route.query,
      rim_depth: String(rimDepth.value),
      fitment_mode: fitmentMode.value,
      valve_length: String(manualValveLength.value),
      extender_length: String(manualExtenderLength.value),
      pump_head_grip_depth: String(pumpHeadGripDepth.value),
      rim_depth_uncertainty: String(rimDepthUncertainty.value),
    },
  })
}

const commitInnerTubePumpHeadGripDepthInput = () => {
  const parsedPumpHeadGripDepth = Number(pumpHeadGripDepthInputValue.value)
  if (!Number.isFinite(parsedPumpHeadGripDepth)) {
    pumpHeadGripDepthInputValue.value = String(pumpHeadGripDepth.value)
    return
  }
  const normalizedPumpHeadGripDepth = Math.min(
    pumpHeadGripDepthMaximum.value,
    Math.max(pumpHeadGripDepthMinimum.value, Math.round(parsedPumpHeadGripDepth)),
  )
  pumpHeadGripDepth.value = normalizedPumpHeadGripDepth
  pumpHeadGripDepthInputValue.value = String(normalizedPumpHeadGripDepth)
  void requestCurrentInnerTubeFitmentResult()
}

const commitInnerTubeRimDepthUncertaintyInput = () => {
  const parsedRimDepthUncertainty = Number(rimDepthUncertaintyInputValue.value)
  if (!Number.isFinite(parsedRimDepthUncertainty)) {
    rimDepthUncertaintyInputValue.value = String(rimDepthUncertainty.value)
    return
  }
  const normalizedRimDepthUncertainty = Math.min(
    rimDepthUncertaintyMaximum.value,
    Math.max(0, Math.round(parsedRimDepthUncertainty)),
  )
  rimDepthUncertainty.value = normalizedRimDepthUncertainty
  rimDepthUncertaintyInputValue.value = String(normalizedRimDepthUncertainty)
  void requestCurrentInnerTubeFitmentResult()
}

const commitInnerTubeRimDepthInput = () => {
  const parsedRimDepth = Number(rimDepthInputValue.value)
  if (!Number.isFinite(parsedRimDepth)) {
    rimDepthInputValue.value = String(rimDepth.value)
    return
  }
  const normalizedRimDepth = Math.min(
    rimDepthMaximum.value,
    Math.max(rimDepthMinimum.value, Math.round(parsedRimDepth)),
  )
  rimDepth.value = normalizedRimDepth
  rimDepthInputValue.value = String(normalizedRimDepth)
  void requestCurrentInnerTubeFitmentResult()
}

const requestCurrentInnerTubeFitmentResult = async (options: { updateShareUrl?: boolean } = {}) => {
  if (options.updateShareUrl !== false) void updateInnerTubeFitmentShareUrl()
  const requestSequence = ++innerTubeFitmentRequestSequence
  activeInnerTubeFitmentResult.value = null
  innerTubeFitmentRequestPending.value = true
  innerTubeFitmentRequestError.value = null
  sharedCalculationCopied.value = false
  sharedCalculationCopyError.value = false
  try {
    const result = await solveInnerTubeValveFitmentWithBackend({
      rim_depth_mm: rimDepth.value,
      mode: fitmentMode.value,
      pump_head_grip_depth_mm: pumpHeadGripDepth.value,
      rim_depth_uncertainty_mm: rimDepthUncertainty.value,
      ...(fitmentMode.value === 'manual'
        ? {
            valve_length_mm: manualValveLength.value,
            extender_length_mm: manualExtenderLength.value,
          }
        : {}),
    })
    if (requestSequence === innerTubeFitmentRequestSequence) {
      activeInnerTubeFitmentResult.value = result
    }
  } catch (error) {
    if (requestSequence === innerTubeFitmentRequestSequence) {
      innerTubeFitmentRequestError.value = error instanceof Error ? error.message : 'Calculation unavailable'
    }
  } finally {
    if (requestSequence === innerTubeFitmentRequestSequence) {
      innerTubeFitmentRequestPending.value = false
    }
  }
}

const setInnerTubeFitmentMode = (mode: InnerTubeValveFitmentMode) => {
  fitmentMode.value = mode
  void requestCurrentInnerTubeFitmentResult()
}

const setInnerTubeRimDepthPreset = (depth: number) => {
  rimDepth.value = depth
  void requestCurrentInnerTubeFitmentResult()
}

const selectManualInnerTubeValveLength = (length: number) => {
  manualValveLength.value = length
  void requestCurrentInnerTubeFitmentResult()
}

const selectManualInnerTubeExtenderLength = (length: number) => {
  manualExtenderLength.value = length
  void requestCurrentInnerTubeFitmentResult()
}

const selectInnerTubeFitmentAlternative = (alternative: InnerTubeValveFitmentAlternative) => {
  manualValveLength.value = alternative.valve_length_mm
  manualExtenderLength.value = alternative.extender_length_mm
  fitmentMode.value = 'manual'
  void requestCurrentInnerTubeFitmentResult()
}

const selectInnerTubeValveStructure = (structure: InnerTubeValveStructure) => {
  activeValveStructure.value = structure
}

const copyInnerTubeFitmentShareLink = async () => {
  if (!import.meta.client) return
  sharedCalculationCopied.value = false
  sharedCalculationCopyError.value = false
  try {
    await updateInnerTubeFitmentShareUrl()
    const shareUrl = new URL(router.currentRoute.value.fullPath, window.location.origin).toString()
    let copied = false
    if (navigator.clipboard?.writeText) {
      try {
        await navigator.clipboard.writeText(shareUrl)
        copied = true
      } catch {
        copied = false
      }
    }
    if (!copied) {
      const shareUrlTextarea = document.createElement('textarea')
      shareUrlTextarea.value = shareUrl
      shareUrlTextarea.setAttribute('readonly', '')
      shareUrlTextarea.style.position = 'fixed'
      shareUrlTextarea.style.opacity = '0'
      document.body.appendChild(shareUrlTextarea)
      try {
        shareUrlTextarea.select()
        copied = document.execCommand('copy')
      } finally {
        document.body.removeChild(shareUrlTextarea)
      }
    }
    if (!copied) throw new Error('clipboard copy failed')
    sharedCalculationCopied.value = true
    window.setTimeout(() => {
      sharedCalculationCopied.value = false
    }, 2400)
  } catch {
    sharedCalculationCopyError.value = true
  }
}

onMounted(() => {
  void requestCurrentInnerTubeFitmentResult({ updateShareUrl: false })
})

const innerTubeFitmentJsonLd = computed(() => ({
  '@context': 'https://schema.org',
  '@type': 'TechArticle',
  '@id': `${canonicalUrl.value}#inner-tube-valve-fitment-engine`,
  url: canonicalUrl.value,
  headline: t('guidesTireInnerTube.fitment.title'),
  description: t('guidesTireInnerTube.fitment.description'),
  proficiencyLevel: 'Expert',
  inLanguage: innerTubeFitmentSchemaLanguage.value,
  author: { '@type': 'Organization', name: 'Tanzanite Engineering Laboratory' },
  articleBody: t('guidesTireInnerTube.fitment.description'),
  hasPart: {
    '@type': 'Dataset',
    name: t('guidesTireInnerTube.fitment.matrix.title'),
    description: t('guidesTireInnerTube.fitment.matrix.description'),
    isAccessibleForFree: true,
    ...(innerTubeFitmentMetadata.value?.rows?.length ? { numberOfItems: innerTubeFitmentMetadata.value.rows.length } : {}),
    variableMeasured: ['rim depth', 'valve length', 'pump-head grip depth', 'rim-depth uncertainty', 'minimum required total length', 'effective valve exposure', 'length margin'],
  },
}))

useHead(() => ({
  script: [createSeoJsonLdScript(innerTubeFitmentJsonLd.value)],
}))
</script>

<style scoped>
.inner-tube-fitment-svg-label,
.inner-tube-fitment-svg-value {
  font-family: var(--tz-font-ui);
  font-size: 10px;
  font-weight: 700;
  fill: var(--tz-text-secondary);
}

.inner-tube-fitment-svg-value {
  font-size: 13px;
  font-weight: 900;
  fill: var(--tz-text-primary);
}

.inner-tube-fitment-svg-label--strong {
  fill: var(--tz-text-primary);
}

.inner-tube-fitment-svg-label--accent {
  fill: var(--tz-site-accent);
}

.inner-tube-fitment-svg-label--extender {
  fill: #2563eb;
}

.inner-tube-fitment-svg-label--muted {
  fill: var(--tz-text-muted);
}

.inner-tube-fitment-status {
  color: var(--tz-text-primary);
}

.inner-tube-fitment-status--optimal {
  border-color: color-mix(in srgb, var(--tz-site-accent) 42%, transparent);
  background: color-mix(in srgb, var(--tz-site-accent) 9%, transparent);
}

.inner-tube-fitment-status--marginal {
  border-color: color-mix(in srgb, #d97706 44%, transparent);
  background: color-mix(in srgb, #d97706 10%, transparent);
}

.inner-tube-fitment-status--unsafe {
  border-color: color-mix(in srgb, #dc2626 44%, transparent);
  background: color-mix(in srgb, #dc2626 9%, transparent);
}

.inner-tube-fitment-text--optimal {
  color: var(--tz-site-accent);
}

.inner-tube-fitment-text--marginal {
  color: #b45309;
}

.inner-tube-fitment-text--unsafe {
  color: #dc2626;
}

.inner-tube-fitment-cell--optimal {
  color: var(--tz-site-accent);
  font-weight: 700;
}

.inner-tube-fitment-cell--marginal {
  color: #b45309;
  font-weight: 700;
}

.inner-tube-fitment-cell--unsafe {
  color: #dc2626;
  font-weight: 700;
}
</style>
