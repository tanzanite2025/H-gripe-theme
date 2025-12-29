<template>
  <!-- Tire width -> rim internal width helper -->
  <div class="mt-5 rounded-2xl bg-slate-900/70 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)] text-center">
    <h3 class="mb-2 text-sm font-semibold text-slate-100">
      Tire width to rim internal width helper
    </h3>
    <p class="mb-3 text-xs text-slate-400">
      Enter your tire width in millimetres to see an approximate compatible rim internal width
      range based on common ETRTO-style guidelines. Use the rim type toggle to see how hookless vs
      hooked shifts the suggested range. Always cross-check with brand charts and the specific
      recommendations from your rim and tire manufacturers.
    </p>
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end justify-center items-center">
      <div class="sm:w-40">
        <label class="block text-xs font-medium text-slate-300" for="tire-width-mm">
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
          class="mt-1 w-full rounded-md border border-slate-600/80 bg-slate-950/80 px-2 py-1.5 text-xs text-slate-100 shadow-[2px_2px_6px_rgba(0,0,0,0.85)] outline-none focus:border-sky-400 focus:ring-0"
        />
      </div>

      <div class="sm:w-52">
        <span class="mb-1 block text-xs font-medium text-slate-300">
          Rim type
        </span>
        <div
          class="inline-flex rounded-full bg-slate-800/80 p-0.5 shadow-[2px_2px_6px_rgba(0,0,0,0.85)]"
        >
          <button
            type="button"
            class="rounded-full px-3 py-1 text-[11px] transition-colors"
            :class="rimType === 'hookless'
              ? 'bg-sky-400 text-slate-900'
              : 'text-slate-200 hover:text-slate-50'"
            @click="rimType = 'hookless'"
          >
            Hookless (TSS)
          </button>
          <button
            type="button"
            class="rounded-full px-3 py-1 text-[11px] transition-colors"
            :class="rimType === 'hooked'
              ? 'bg-sky-400 text-slate-900'
              : 'text-slate-200 hover:text-slate-50'"
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
        class="text-xs text-slate-500"
      >
        Enter a tire width between about 18 and 80 mm to see a suggested rim internal width range.
      </p>

      <div
        v-else
        class="text-xs text-slate-200"
      >
        <p class="font-semibold text-sky-300">
          Recommended rim internal width:
          {{ tireRimSuggestion.minRim }} - {{ tireRimSuggestion.maxRim }} mm
        </p>
        <p class="mt-0.5 text-[11px] text-slate-400">
          Sweet spot around {{ tireRimSuggestion.ideal }} mm. For aggressive or technical riding,
          stay closer to the wider end of the range.
        </p>
      </div>
    </div>

    <div v-if="!hideSearchButton" class="mt-4 flex justify-center">
      <button
        type="button"
        class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-4 py-1.5 text-xs font-semibold text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] transition-all"
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

const tireWidthInput = ref<string>('')
const rimType = ref<RimType>('hooked')

const { open: openShopSearch } = useShopSearchSheet()

const props = withDefaults(defineProps<{
  hideSearchButton?: boolean
}>(), {
  hideSearchButton: false
})

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
