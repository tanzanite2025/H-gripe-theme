<template>
  <!-- Tire width -> rim internal width helper -->
  <div class="tire-rim-helper-card mt-5 rounded-2xl bg-[var(--tz-form-panel-surface)] p-3 text-center md:p-4">
    <h3 class="mb-2 text-sm font-semibold tz-text-primary">
      {{ title || t('guidesTireRimHelper.title') }}
    </h3>
    <p class="mb-3 text-xs tz-text-secondary">
      {{ description || t('guidesTireRimHelper.description') }}
    </p>
    <div class="tire-rim-helper__controls mx-auto flex w-full max-w-xl flex-col gap-3 items-stretch sm:flex-row sm:items-end">
      <div class="w-full min-w-0 sm:flex-1">
        <label class="block text-left text-xs font-medium tz-text-secondary" for="tire-width-input">
            {{ t('guidesTireRimHelper.tireWidthLabel') }}
        </label>
        <button
          id="tire-width-input"
          type="button"
          class="tire-rim-helper__picker-trigger mt-1 flex w-full items-center justify-between gap-3 rounded-md px-3 py-0 text-left text-xs outline-none transition-colors focus-visible:ring-2"
          aria-haspopup="dialog"
          :aria-expanded="tireWidthPickerOpen"
          @click="openTireWidthPicker"
        >
          <span class="min-w-0 flex-1">
            <span class="block truncate font-medium tz-text-primary">
              {{ selectedTireWidthSummary }}
            </span>
          </span>
          <Icon name="lucide:chevron-down" class="h-4 w-4 shrink-0 tz-text-secondary" aria-hidden="true" />
        </button>
      </div>

      <div class="w-full shrink-0 sm:w-52">
        <div class="tire-rim-helper__rim-system-label mb-1 text-xs font-medium tz-text-secondary">
          <span>{{ t('guidesTireRimHelper.rimSystemLabel') }}</span>
          <button
            type="button"
            class="tire-rim-helper__rim-system-help inline-flex items-center justify-center rounded-full"
            :aria-label="t('guidesTireRimHelper.rimSystemHelpButton')"
            :aria-expanded="tireRimSystemHelpOpen"
            aria-haspopup="dialog"
            @click="openTireRimSystemHelp"
          >
            <Icon name="lucide:circle-help" class="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
        <div
          class="tire-rim-helper__toggle-group inline-flex rounded-full bg-[var(--tz-form-control-surface)] p-0.5"
        >
          <button
            type="button"
            class="tire-rim-helper__toggle rounded-full px-3 py-1 tz-caption transition-colors"
            :class="{ 'tire-rim-helper__toggle--active': rimType === 'hookless' }"
            @click="rimType = 'hookless'"
          >
            {{ t('guidesTireRimHelper.hookless') }}
          </button>
          <button
            type="button"
            class="tire-rim-helper__toggle rounded-full px-3 py-1 tz-caption transition-colors"
            :class="{ 'tire-rim-helper__toggle--active': rimType === 'hooked' }"
            @click="rimType = 'hooked'"
          >
            {{ t('guidesTireRimHelper.hooked') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Suggestion / Hint Text (Moved outside flex container to ensure new line) -->
    <div class="mt-3">
      <p
        v-if="tireWidthOutOfRange"
        class="text-xs tz-text-muted"
      >
        {{ t('guidesTireRimHelper.outOfRange') }}
      </p>

      <p
        v-else-if="recommendationPending"
        class="text-xs tz-text-muted"
      >
        {{ t('guidesTireRimHelper.loading') }}
      </p>

      <p
        v-else-if="tireWidthHasNoPublishedBracket"
        class="text-xs tz-text-muted"
      >
        {{ t('guidesTireRimHelper.noPublishedBracket') }}
      </p>

      <p
        v-else-if="!tireRimSuggestion"
        class="text-xs tz-text-muted"
      >
        {{ t('guidesTireRimHelper.empty') }}
      </p>

      <div
        v-else
        class="text-xs tz-text-secondary"
      >
        <p class="tire-rim-helper__result font-semibold">
          {{ t(
            tireRimSuggestion.isEngineeringRecommended
              ? 'guidesTireRimHelper.engineeringRecommended'
              : tireRimSuggestion.isEngineeringReference
                ? 'guidesTireRimHelper.engineeringReference'
                : tireRimSuggestion.isCalculated
                  ? 'guidesTireRimHelper.calculatedReference'
                  : tireRimSuggestion.isPossibleReference
                    ? 'guidesTireRimHelper.possibleReference'
                    : 'guidesTireRimHelper.recommended',
          ) }}
          {{ formatTireRimWidthReferenceRanges(tireRimSuggestion.rimWidthRanges) }} mm
        </p>
        <p
          v-if="tireRimSuggestion.isCalculated && tireRimSuggestion.calculationRows"
          class="mt-0.5 tz-caption tz-text-muted"
        >
          {{ t('guidesTireRimHelper.calculated', tireRimSuggestion.calculationRows) }}
        </p>
        <p
          v-if="tireRimSuggestion.isPossibleReference"
          class="mt-0.5 tz-caption tz-text-muted"
        >
          {{ t('guidesTireRimHelper.possibleReferenceNote') }}
        </p>
        <p
          v-if="tireRimSuggestion.isCalculatedFromPossibleReference"
          class="mt-0.5 tz-caption tz-text-muted"
        >
          {{ t('guidesTireRimHelper.calculatedFromPossibleReferenceNote') }}
        </p>
        <p
          v-if="tireRimSuggestion.isEngineeringReference"
          class="mt-0.5 tz-caption tz-text-muted"
        >
          {{ t('guidesTireRimHelper.engineeringReferenceNote') }}
        </p>
        <div class="tire-rim-helper__physics mt-2 text-left tz-caption tz-text-muted">
          <p>
            {{ t('guidesTireRimHelper.inflatedWidth') }}
            {{ formatMetricRange(tireRimSuggestion.physical.inflatedTireWidth) }} mm
          </p>
          <p>
            {{ t('guidesTireRimHelper.aeroTargetOuterWidth') }}
            {{ formatMetricRange(tireRimSuggestion.physical.aeroTargetOuterWidth) }} mm
          </p>
          <p class="mt-0.5">
            {{ t('guidesTireRimHelper.physicsNote') }}
          </p>
        </div>
      </div>
    </div>

    <div v-if="!hideSearchButton && !tireWidthOutOfRange" class="mt-4 flex justify-center">
      <button
        type="button"
        class="tire-rim-helper__search inline-flex items-center justify-center rounded-full px-4 py-1.5 text-xs font-semibold shadow-md transition-all"
        @click="tireRimSearchSheetOpen = true"
      >
        {{ t('guidesTireRimHelper.search') }}
      </button>
    </div>

    <TireRimProductSearchSheet v-model="tireRimSearchSheetOpen" />
  </div>

  <Teleport to="body">
    <div
      v-if="tireWidthPickerOpen"
      class="tire-rim-helper__picker-mask fixed inset-0 z-[14000] flex items-center justify-center p-2 md:p-5 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
      role="presentation"
      @click.self="closeTireWidthPicker"
      @keydown.esc.stop.prevent="closeTireWidthPicker"
    >
      <section
        class="tire-rim-helper__picker tz-mobile-dialog-surface w-full max-w-2xl overflow-hidden rounded-2xl p-4"
        role="dialog"
        aria-modal="true"
        aria-labelledby="tire-width-picker-title"
        @click.stop
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 id="tire-width-picker-title" class="text-base font-semibold tz-text-primary">
              {{ t('guidesTireRimHelper.tireWidthPickerTitle') }}
            </h2>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ t('guidesTireRimHelper.tireWidthPickerDescription') }}
            </p>
          </div>
          <button
            type="button"
            class="tz-global-close-btn shrink-0"
            :aria-label="t('guidesTireRimHelper.tireWidthPickerClose')"
            @click="closeTireWidthPicker"
          >
            <Icon name="lucide:x" class="h-3.5 w-3.5" aria-hidden="true" />
          </button>
        </div>

        <div class="tire-rim-helper__picker-options mt-4 min-h-0 flex-1 overflow-y-auto pr-1">
          <p class="mb-2 text-xs font-semibold tz-text-primary">
            {{ t('guidesTireRimHelper.tireWidthPresetCSection') }}
          </p>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <button
              v-for="preset in cMarkingTireWidthPresets"
              :key="preset.id"
              type="button"
              class="tire-rim-helper__preset rounded-lg px-3 py-2 text-left transition-colors"
              :class="{ 'tire-rim-helper__preset--active': selectedTireWidthPreset?.id === preset.id }"
              @click="selectTireWidthPreset(preset)"
            >
              <span class="block text-sm font-semibold tz-text-primary">{{ preset.label }}</span>
              <span class="mt-0.5 block tz-caption tz-text-secondary">
                {{ preset.millimeters }} mm · {{ formatTireWidthPresetInches(preset) }}″
              </span>
            </button>
          </div>

          <p class="mb-2 mt-5 text-xs font-semibold tz-text-primary">
            {{ t('guidesTireRimHelper.tireWidthPresetInchSection') }}
          </p>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <button
              v-for="preset in inchMarkingTireWidthPresets"
              :key="preset.id"
              type="button"
              class="tire-rim-helper__preset rounded-lg px-3 py-2 text-left transition-colors"
              :class="{ 'tire-rim-helper__preset--active': selectedTireWidthPreset?.id === preset.id }"
              @click="selectTireWidthPreset(preset)"
            >
              <span class="block text-sm font-semibold tz-text-primary">{{ preset.label }}</span>
              <span class="mt-0.5 block tz-caption tz-text-secondary">
                {{ preset.millimeters }} mm · {{ formatTireWidthPresetInches(preset) }}″
              </span>
            </button>
          </div>

        </div>
      </section>
    </div>
  </Teleport>

  <Teleport to="body">
    <div
      v-if="tireRimSystemHelpOpen"
      class="tire-rim-helper__system-help-mask fixed inset-0 z-[14000] flex items-center justify-center p-2 md:p-5 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
      role="presentation"
      @click.self="closeTireRimSystemHelp"
      @keydown.esc.stop.prevent="closeTireRimSystemHelp"
    >
      <section
        class="tire-rim-helper__system-help tz-mobile-dialog-surface w-full max-w-5xl overflow-hidden rounded-2xl p-4 md:p-8"
        role="dialog"
        aria-modal="true"
        aria-labelledby="tire-rim-system-help-title"
        @click.stop
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 id="tire-rim-system-help-title" class="text-base font-semibold tz-text-primary md:text-lg">
              {{ t('guidesTireRimHelper.rimSystemHelpTitle') }}
            </h2>
            <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
              {{ t('guidesTireRimHelper.rimSystemHelpIntro') }}
            </p>
          </div>
          <button
            type="button"
            class="tz-global-close-btn shrink-0"
            :aria-label="t('guidesTireRimHelper.rimSystemHelpClose')"
            @click="closeTireRimSystemHelp"
          >
            <Icon name="lucide:x" class="h-3.5 w-3.5" aria-hidden="true" />
          </button>
        </div>

        <div class="tire-rim-helper__system-help-grid mt-4 grid grid-cols-2 gap-3">
          <article class="tire-rim-helper__system-help-card rounded-xl p-3">
            <h3 class="text-sm font-semibold tz-text-primary">
              {{ t('guidesTireRimHelper.rimSystemHelpHookedTitle') }}
            </h3>
            <div class="tire-rim-helper__system-help-diagram">
              <img
                class="tire-rim-helper__system-help-image"
                src="/images/guides/tireguides/choose/hooked-rim-cross-section.webp"
                :alt="t('guidesTireRimHelper.rimSystemHelpHookedDiagram')"
                loading="lazy"
                decoding="async"
              >
            </div>
            <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
              {{ t('guidesTireRimHelper.rimSystemHelpHookedBody') }}
            </p>
          </article>

          <article class="tire-rim-helper__system-help-card rounded-xl p-3">
            <h3 class="text-sm font-semibold tz-text-primary">
              {{ t('guidesTireRimHelper.rimSystemHelpHooklessTitle') }}
            </h3>
            <div class="tire-rim-helper__system-help-diagram">
              <img
                class="tire-rim-helper__system-help-image"
                src="/images/guides/tireguides/choose/hookless-rim-cross-section.webp"
                :alt="t('guidesTireRimHelper.rimSystemHelpHooklessDiagram')"
                loading="lazy"
                decoding="async"
              >
            </div>
            <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
              {{ t('guidesTireRimHelper.rimSystemHelpHooklessBody') }}
            </p>
          </article>
        </div>

        <div class="tire-rim-helper__system-help-note mt-4 rounded-xl p-3">
          <h3 class="text-sm font-semibold tz-text-primary">
            {{ t('guidesTireRimHelper.rimSystemHelpUseTitle') }}
          </h3>
          <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
            {{ t('guidesTireRimHelper.rimSystemHelpUseBody') }}
          </p>
        </div>

        <p class="mt-3 text-xs leading-relaxed tz-text-muted">
          {{ t('guidesTireRimHelper.rimSystemHelpSafety') }}
        </p>

        <div class="mt-4 border-t tz-border-subtle pt-3">
          <p class="text-[11px] font-semibold uppercase tracking-wide tz-text-muted">
            {{ t('guidesTireRimHelper.rimSystemHelpSourcesTitle') }}
          </p>
          <ul class="mt-2 space-y-1 text-xs tz-text-secondary">
            <li>
              {{ t('guidesTireRimHelper.rimSystemHelpSourceZipp') }}
            </li>
            <li>
              {{ t('guidesTireRimHelper.rimSystemHelpSourceZipp303') }}
            </li>
            <li>
              {{ t('guidesTireRimHelper.rimSystemHelpSourceEnve') }}
            </li>
            <li>
              {{ t('guidesTireRimHelper.rimSystemHelpSourceDtSwiss') }}
            </li>
          </ul>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  useTireRimWidthReferenceRecommendation,
  type RimType,
} from '~/composables/useTireRimWidthReferenceRecommendation'
import { formatTireRimWidthReferenceRanges } from '~/data/tireguides/tireRimWidthReferencePresentation'
import {
  commonTireWidthInputPresets,
  parseTireWidthInputAsMillimeters,
  stripTireWidthInputUnitSuffix,
  type TireWidthInputPreset,
  type TireWidthInputUnit,
} from '~/data/tireguides/tireRimWidthInputConversion'
import TireRimProductSearchSheet from '~/components/TireRimProductSearchSheet.vue'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireRimHelper')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const props = withDefaults(defineProps<{
  hideSearchButton?: boolean
  initialRimType?: RimType
  title?: string
  description?: string
}>(), {
  hideSearchButton: false,
  initialRimType: 'hooked',
  title: '',
  description: '',
})

const tireWidthInput = ref<string>('')
const tireWidthInputUnit = ref<TireWidthInputUnit>('mm')
const tireWidthPickerOpen = ref(false)
const tireRimSystemHelpOpen = ref(false)
const tireRimSearchSheetOpen = ref(false)
const rimType = ref<RimType>(props.initialRimType)
const {
  parsedTireWidth,
  tireRimSuggestion,
  recommendationPending,
  tireWidthOutOfRange,
  tireWidthHasNoPublishedBracket,
} = useTireRimWidthReferenceRecommendation(
  tireWidthInput,
  rimType,
  tireWidthInputUnit,
)

const cMarkingTireWidthPresets = computed(() => (
  commonTireWidthInputPresets.filter(preset => preset.category === 'c-marking')
))

const inchMarkingTireWidthPresets = computed(() => (
  commonTireWidthInputPresets.filter(preset => preset.category === 'inch-marking')
))

const formatTireWidthInches = (widthInMillimeters: number): string => (
  String(Number((widthInMillimeters / 25.4).toFixed(2)))
)

const formatTireWidthPresetInches = (preset: TireWidthInputPreset): string => (
  String(Number(preset.inches.toFixed(2)))
)

const selectedTireWidthPreset = computed(() => {
  const currentWidth = parsedTireWidth.value
  if (currentWidth === null) return undefined

  const currentInput = stripTireWidthInputUnitSuffix(tireWidthInput.value)
  return commonTireWidthInputPresets.find(preset => (
    preset.millimeters === currentWidth
      && preset.inputUnit === tireWidthInputUnit.value
      && preset.inputValue === currentInput
  ))
})

const selectedTireWidthSummary = computed(() => {
  const currentInput = tireWidthInput.value.trim()
  const currentWidth = parsedTireWidth.value
  if (!currentInput || currentWidth === null) {
    return t('guidesTireRimHelper.tireWidthPickerPlaceholder')
  }

  const primaryLabel = selectedTireWidthPreset.value?.label
    || (tireWidthInputUnit.value === 'inch'
      ? `${stripTireWidthInputUnitSuffix(currentInput)}″`
      : `${currentWidth} mm`)

  return `${primaryLabel} · ${currentWidth} mm · ${formatTireWidthInches(currentWidth)}″`
})

const openTireWidthPicker = (): void => {
  tireWidthPickerOpen.value = true
}

const closeTireWidthPicker = (): void => {
  tireWidthPickerOpen.value = false
}

const openTireRimSystemHelp = (): void => {
  tireRimSystemHelpOpen.value = true
}

const closeTireRimSystemHelp = (): void => {
  tireRimSystemHelpOpen.value = false
}

const handleTireRimHelperEscapeKey = (event: KeyboardEvent): void => {
  if (event.key !== 'Escape') return

  if (tireRimSystemHelpOpen.value) {
    closeTireRimSystemHelp()
    return
  }

  if (tireWidthPickerOpen.value) {
    closeTireWidthPicker()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleTireRimHelperEscapeKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleTireRimHelperEscapeKey)
})

const selectTireWidthPreset = (preset: TireWidthInputPreset): void => {
  tireWidthInput.value = preset.inputValue
  tireWidthInputUnit.value = preset.inputUnit
  closeTireWidthPicker()
}

const formatMetricRange = ({ min, max }: { min: number; max: number }) => (
  min === max ? min.toFixed(1) : `${min.toFixed(1)}–${max.toFixed(1)}`
)
</script>

<style scoped>
.tire-rim-helper-card {
  border: 1px solid rgba(5, 150, 105, 0.14);
  background-color: var(--tz-form-panel-surface) !important;
  background-image: none !important;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
}

.tire-rim-helper__input {
  border: 1px solid var(--tz-form-control-border);
  background-color: var(--tz-form-control-surface) !important;
  background-image: none !important;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.06);
}

.tire-rim-helper__input:focus {
  border-color: var(--tz-site-accent);
}

.tire-rim-helper__picker-trigger {
  height: 42px;
  min-height: 42px;
  padding-top: 0;
  padding-bottom: 0;
  border: 1px solid var(--tz-form-control-border);
  background-color: var(--tz-form-control-surface) !important;
  background-image: none !important;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.06);
}

.tire-rim-helper__picker-trigger:hover,
.tire-rim-helper__picker-trigger:focus-visible {
  border-color: var(--tz-site-accent);
  outline: none;
}

.tire-rim-helper__rim-system-label {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
}

.tire-rim-helper__rim-system-help {
  width: 1.5rem;
  height: 1.5rem;
  flex: 0 0 auto;
  border: 1px solid var(--tz-site-accent);
  background: var(--tz-site-accent);
  color: var(--tz-action-primary-foreground);
  box-shadow: 0 3px 10px rgba(5, 150, 105, 0.28);
  transition: color 160ms ease, background-color 160ms ease, border-color 160ms ease, box-shadow 160ms ease, transform 160ms ease;
}

.tire-rim-helper__rim-system-help:hover,
.tire-rim-helper__rim-system-help:focus-visible {
  border-color: var(--tz-site-accent-hover);
  background: var(--tz-site-accent-hover);
  color: var(--tz-action-primary-foreground);
  box-shadow: 0 5px 14px rgba(5, 150, 105, 0.36);
  outline: 2px solid color-mix(in srgb, var(--tz-site-accent) 45%, transparent);
  outline-offset: 2px;
  transform: scale(1.06);
}

.tire-rim-helper__picker-mask {
  background: rgba(15, 23, 42, 0.35);
  backdrop-filter: blur(3px);
}

.tire-rim-helper__picker {
  display: flex;
  flex-direction: column;
  max-height: min(760px, 92vh);
  border: 1px solid var(--tz-form-control-border);
  background: var(--tz-form-panel-surface);
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.24);
}

.tire-rim-helper__picker-options {
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
}

.tire-rim-helper__system-help-mask {
  background: rgba(15, 23, 42, 0.38);
  backdrop-filter: blur(3px);
}

.tire-rim-helper__system-help {
  max-height: min(960px, 94vh);
  border: 1px solid var(--tz-form-control-border);
  background: var(--tz-form-panel-surface);
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.24);
}

.tire-rim-helper__system-help-card {
  border: 1px solid var(--tz-form-control-border);
  background: var(--tz-form-control-surface);
}

.tire-rim-helper__system-help-diagram {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.75rem;
  background: #171717;
}

.tire-rim-helper__system-help-image {
  display: block;
  width: 100%;
  max-width: 20rem;
  height: auto;
}

.tire-rim-helper__system-help-note {
  border: 1px solid rgba(5, 150, 105, 0.24);
  background: rgba(5, 150, 105, 0.06);
}

@media (max-width: 767px) {
  .tire-rim-helper__picker {
    max-height: min(
      760px,
      calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2)
    );
  }

  .tire-rim-helper__system-help {
    max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  }
}

.tire-rim-helper__preset {
  border: 1px solid var(--tz-form-control-border);
  background: var(--tz-form-control-surface);
}

.tire-rim-helper__preset:hover,
.tire-rim-helper__preset:focus-visible,
.tire-rim-helper__preset--active {
  border-color: var(--tz-site-accent);
  background: color-mix(in srgb, var(--tz-site-accent) 10%, var(--tz-form-control-surface));
  outline: none;
}

.tire-rim-helper__toggle-group {
  display: inline-flex;
  height: 42px;
  align-items: stretch;
  background-color: var(--tz-form-control-surface) !important;
  background-image: none !important;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.06);
}

.tire-rim-helper__result {
  color: var(--tz-site-accent);
}

.tire-rim-helper__physics {
  border-top: 1px solid var(--tz-form-control-border);
  padding-top: 0.5rem;
  line-height: 1.55;
}

.tire-rim-helper__toggle {
  display: inline-flex;
  height: 100%;
  align-items: center;
  padding-top: 0;
  padding-bottom: 0;
  line-height: 1.25;
  border: 1px solid rgba(5, 150, 105, 0.35);
  background-color: var(--tz-surface-muted);
  background-image: none;
  color: var(--tz-text-secondary);
}

.tire-rim-helper__toggle:hover,
.tire-rim-helper__toggle:focus-visible {
  border-color: rgba(5, 150, 105, 0.85);
  color: var(--tz-text-primary);
  outline: none;
}

.tire-rim-helper__toggle--active {
  border-color: var(--tz-site-accent);
  background-color: var(--tz-site-accent);
  background-image: none;
  color: #ffffff;
  box-shadow: 0 0 0 1px rgba(5, 150, 105, 0.15), 0 4px 14px rgba(5, 150, 105, 0.16);
}

.tire-rim-helper__toggle--active:hover,
.tire-rim-helper__toggle--active:focus-visible {
  border-color: var(--tz-site-accent-hover);
  background-color: var(--tz-site-accent-hover);
  background-image: none;
  color: #ffffff;
}

.tire-rim-helper__search {
  border: 1px solid var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
}

.tire-rim-helper__search:hover,
.tire-rim-helper__search:focus-visible {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
  box-shadow: 0 8px 22px -6px rgb(15 23 42 / 0.24);
  transform: translateY(-1px);
  outline: none;
}
</style>
