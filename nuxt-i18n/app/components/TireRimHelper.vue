<template>
  <!-- Tire width -> rim internal width helper -->
  <div class="tire-rim-helper-card mt-5 rounded-2xl bg-[var(--tz-form-panel-surface)] p-4 text-center">
    <h3 v-if="title" class="mb-2 text-sm font-semibold tz-text-primary">
      {{ title }}
    </h3>
    <p v-if="description" class="mb-3 text-xs tz-text-secondary">
      {{ description }}
    </p>
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end justify-center items-center">
      <div class="sm:w-40">
        <label class="block text-xs font-medium tz-text-secondary" for="tire-width-mm">
          Tire width (mm)
        </label>
        <input
          id="tire-width-mm"
          v-model="tireWidthInput"
          type="number"
          min="18"
          max="130"
          step="1"
          placeholder="e.g. 28 or 57"
          class="tire-rim-helper__input mt-1 w-full rounded-md bg-[var(--tz-form-control-surface)] px-2 py-1.5 text-xs tz-text-primary outline-none focus:ring-0"
        />
      </div>

      <div class="sm:w-52">
          <span class="mb-1 block text-xs font-medium tz-text-secondary">
          Rim system
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
            Hookless (TSS)
          </button>
          <button
            type="button"
            class="tire-rim-helper__toggle rounded-full px-3 py-1 tz-caption transition-colors"
            :class="{ 'tire-rim-helper__toggle--active': rimType === 'hooked' }"
            @click="rimType = 'hooked'"
          >
            Hooked (TC)
          </button>
        </div>
      </div>
    </div>

    <!-- Suggestion / Hint Text (Moved outside flex container to ensure new line) -->
    <div class="mt-3">
      <p
        v-if="!tireRimSuggestion"
        class="text-xs tz-text-muted"
      >
        Enter a tire width to see a suggested rim internal width range.
      </p>

      <div
        v-else
        class="text-xs tz-text-secondary"
      >
        <p class="tire-rim-helper__result font-semibold">
          Recommended rim internal width:
          {{ tireRimSuggestion.minRim }} - {{ tireRimSuggestion.maxRim }} mm
        </p>
        <p class="mt-0.5 tz-caption tz-text-muted">
          Sweet spot around {{ tireRimSuggestion.ideal }} mm. For aggressive or technical riding,
          stay closer to the wider end of the range.
        </p>
      </div>
    </div>

    <div v-if="!hideSearchButton" class="mt-4 flex justify-center">
      <button
        type="button"
        class="tire-rim-helper__search inline-flex items-center justify-center rounded-full px-4 py-1.5 text-xs font-semibold shadow-[0_4px_14px_rgba(0,0,0,0.9)] transition-all"
        @click="() => openShopSearch()"
      >
        Search for suitable width rims
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'

type RimType = 'hookless' | 'hooked'

const { open: openShopSearch } = useShopSearchSheet()

const props = withDefaults(defineProps<{
  hideSearchButton?: boolean
  initialRimType?: RimType
  title?: string
  description?: string
}>(), {
  hideSearchButton: false,
  initialRimType: 'hooked',
  title: 'Tire width to rim internal width helper',
  description: 'Enter your tire width in millimetres and choose the rim system. Tube-type and tubeless-ready setups are not split into two width calculators here; tubular rims use a separate tubular tire/rim system. Always cross-check with the specific recommendations from your rim and tire manufacturers.',
})

const tireWidthInput = ref<string>('')
const rimType = ref<RimType>(props.initialRimType)

interface TireRimSuggestion {
  minRim: number
  maxRim: number
  ideal: number
}

interface RimAnchor {
  tire: number
  minRim: number
  maxRim: number
}

// Anchor rows taken or inferred from DT Swiss style charts.
// You can extend these arrays with more exact rows from the PDF if needed.
const HOOKLESS_ANCHORS: RimAnchor[] = [
  // 32 mm hookless row on chart: 23–25 mm
  { tire: 32, minRim: 23, maxRim: 25 },
  // Very wide hookless tyre example (~102 mm): 36–40 mm bucket
  { tire: 102, minRim: 36, maxRim: 40 },
]

const HOOKED_ANCHORS: RimAnchor[] = [
  // 30 mm hooked row on chart: 18–22 mm
  { tire: 30, minRim: 18, maxRim: 22 },
]

const tireRimSuggestion = computed<TireRimSuggestion | null>(() => {
  const raw = Number(tireWidthInput.value)
  if (!Number.isFinite(raw)) return null

  const width = Math.round(raw)
  // Allow a reasonably wide range; DT Swiss charts go roughly 20–127 mm.
  if (width < 18 || width > 130) return null

  // 1) If we have an explicit anchor for this tyre width, prefer that.
  const anchors = rimType.value === 'hookless' ? HOOKLESS_ANCHORS : HOOKED_ANCHORS
  const anchor = anchors.find((a) => a.tire === width)
  if (anchor) {
    const ideal = Math.round((anchor.minRim + anchor.maxRim) / 2)
    return {
      minRim: anchor.minRim,
      maxRim: anchor.maxRim,
      ideal,
    }
  }

  // 2) Otherwise fall back to an approximate guideline loosely calibrated
  // against DT Swiss style charts.
  // For Hookless we use a small hard-coded bucket table; for Hooked we keep a
  // simple ratio-based guideline tuned to match typical points.
  let minRim: number
  let maxRim: number
  let ideal: number

  if (rimType.value === 'hookless') {
    // Hookless (TSS): buckets roughly following the DT Swiss TSS chart.
    // Key anchors: 32 mm -> 23–25 mm, 40–45 mm -> ~28–30 mm, 102 mm -> ~36–40 mm.
    if (width <= 30) {
      minRim = 23
      maxRim = 25
    } else if (width <= 33) {
      // 32 mm row on chart
      minRim = 23
      maxRim = 25
    } else if (width <= 40) {
      // mid-width gravel tyres
      minRim = 25
      maxRim = 30
    } else if (width <= 50) {
      minRim = 28
      maxRim = 30
    } else if (width <= 60) {
      minRim = 30
      maxRim = 35
    } else if (width <= 80) {
      minRim = 35
      maxRim = 40
    } else {
      // very wide tyres, keep in the largest practical bucket
      minRim = 36
      maxRim = 40
    }
    ideal = Math.round((minRim + maxRim) / 2)
  } else {
    // Hooked (TC): calibrated so 30 mm tyre -> ~18–22 mm inner width.
    minRim = Math.round(width * 0.6)
    maxRim = Math.round(width * 0.74)
    ideal = Math.round(width * 0.67)
  }

  return {
    minRim,
    maxRim,
    ideal,
  }
})
</script>

<style scoped>
.tire-rim-helper-card {
  border: 1px solid rgba(181, 255, 109, 0.14);
  background-color: var(--tz-form-panel-surface) !important;
  background-image: none !important;
  box-shadow: 3px 3px 10px rgba(0, 0, 0, 0.78);
}

.tire-rim-helper__input {
  border: 1px solid var(--tz-form-control-border);
  background-color: var(--tz-form-control-surface) !important;
  background-image: none !important;
  box-shadow: 2px 2px 6px rgba(0, 0, 0, 0.72);
}

.tire-rim-helper__input:focus {
  border-color: var(--tz-brand-primary);
}

.tire-rim-helper__toggle-group {
  background-color: var(--tz-form-control-surface) !important;
  background-image: none !important;
  box-shadow: 2px 2px 6px rgba(0, 0, 0, 0.72);
}

.tire-rim-helper__result {
  color: var(--tz-brand-primary);
}

.tire-rim-helper__toggle {
  border: 1px solid rgba(181, 255, 109, 0.35);
  background-color: #171d13;
  background-image: none;
  color: var(--tz-text-secondary);
}

.tire-rim-helper__toggle:hover,
.tire-rim-helper__toggle:focus-visible {
  border-color: rgba(181, 255, 109, 0.85);
  color: #f8fff1;
  outline: none;
}

.tire-rim-helper__toggle--active {
  border-color: var(--tz-brand-primary);
  background-color: var(--tz-brand-primary);
  background-image: none;
  color: #0b1020;
  box-shadow: 0 0 0 1px rgba(181, 255, 109, 0.15), 0 4px 14px rgba(181, 255, 109, 0.16);
}

.tire-rim-helper__toggle--active:hover,
.tire-rim-helper__toggle--active:focus-visible {
  border-color: var(--tz-brand-primary-hover);
  background-color: var(--tz-brand-primary-hover);
  background-image: none;
  color: #0b1020;
}

.tire-rim-helper__search {
  border: 1px solid var(--tz-brand-primary);
  background: var(--tz-brand-primary);
  color: #0b1020;
}

.tire-rim-helper__search:hover,
.tire-rim-helper__search:focus-visible {
  border-color: var(--tz-brand-primary-hover);
  background: var(--tz-brand-primary-hover);
  box-shadow: 0 8px 22px -6px rgba(181, 255, 109, 0.55);
  transform: translateY(-1px);
  outline: none;
}
</style>
