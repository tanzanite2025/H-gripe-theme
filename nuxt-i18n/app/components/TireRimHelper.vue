<template>
  <!-- Tire width -> rim internal width helper -->
  <div class="tire-rim-helper-card mt-5 rounded-2xl bg-[var(--tz-form-panel-surface)] p-4 text-center">
    <h3 class="mb-2 text-sm font-semibold tz-text-primary">
      {{ title || t('guidesTireRimHelper.title') }}
    </h3>
    <p class="mb-3 text-xs tz-text-secondary">
      {{ description || t('guidesTireRimHelper.description') }}
    </p>
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end justify-center items-center">
      <div class="sm:w-40">
        <label class="block text-xs font-medium tz-text-secondary" for="tire-width-mm">
          {{ t('guidesTireRimHelper.tireWidthLabel') }}
        </label>
        <input
          id="tire-width-mm"
          v-model="tireWidthInput"
          type="number"
          min="18"
          max="130"
          step="1"
          :placeholder="t('guidesTireRimHelper.tireWidthPlaceholder')"
          class="tire-rim-helper__input mt-1 w-full rounded-md bg-[var(--tz-form-control-surface)] px-2 py-1.5 text-xs tz-text-primary outline-none focus:ring-0"
        />
      </div>

      <div class="sm:w-52">
          <span class="mb-1 block text-xs font-medium tz-text-secondary">
          {{ t('guidesTireRimHelper.rimSystemLabel') }}
        </span>
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
        v-if="hooklessSafetyWarning"
        class="tire-rim-helper__safety-warning text-base font-bold"
        role="alert"
      >
        {{ t('guidesTireRimHelper.hooklessSafetyWarning') }}
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
          {{ t('guidesTireRimHelper.recommended') }}
          {{ tireRimSuggestion.minRim }} - {{ tireRimSuggestion.maxRim }} mm
        </p>
        <p class="mt-0.5 tz-caption tz-text-muted">
          {{ t('guidesTireRimHelper.sweetSpot', { width: tireRimSuggestion.ideal }) }}
        </p>
      </div>
    </div>

    <div v-if="!hideSearchButton && !hooklessSafetyWarning" class="mt-4 flex justify-center">
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
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import { useTireRimRecommendation, type RimType } from '~/composables/useTireRimRecommendation'
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
const tireRimSearchSheetOpen = ref(false)
const rimType = ref<RimType>(props.initialRimType)
const { hooklessSafetyWarning, tireRimSuggestion } = useTireRimRecommendation(
  tireWidthInput,
  rimType,
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

.tire-rim-helper__toggle-group {
  background-color: var(--tz-form-control-surface) !important;
  background-image: none !important;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.06);
}

.tire-rim-helper__result {
  color: var(--tz-site-accent);
}

.tire-rim-helper__safety-warning {
  color: #b91c1c;
}

.tire-rim-helper__toggle {
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
