import { computed, type Ref } from 'vue'

export type RimType = 'hookless' | 'hooked'

export interface TireRimSuggestion {
  minRim: number
  maxRim: number
  ideal: number
}

interface RimAnchor {
  tire: number
  minRim: number
  maxRim: number
}

const HOOKLESS_MIN_TIRE_WIDTH = 28

// Anchor rows taken or inferred from DT Swiss style charts.
const HOOKLESS_ANCHORS: RimAnchor[] = [
  { tire: 32, minRim: 23, maxRim: 25 },
  { tire: 102, minRim: 36, maxRim: 40 },
]

const HOOKED_ANCHORS: RimAnchor[] = [
  { tire: 30, minRim: 18, maxRim: 22 },
]

export const useTireRimRecommendation = (
  tireWidthInput: Ref<string>,
  rimType: Ref<RimType>,
) => {
  const parsedTireWidth = computed(() => {
    const input = String(tireWidthInput.value).trim()
    if (!input) return null

    const raw = Number(input)
    return Number.isFinite(raw) ? raw : null
  })

  const hooklessSafetyWarning = computed(() => (
    rimType.value === 'hookless'
    && parsedTireWidth.value !== null
    && parsedTireWidth.value < HOOKLESS_MIN_TIRE_WIDTH
  ))

  const tireRimSuggestion = computed<TireRimSuggestion | null>(() => {
    const raw = parsedTireWidth.value
    if (raw === null || hooklessSafetyWarning.value) return null

    const width = Math.round(raw)
    if (width < 18 || width > 130) return null

    const anchors = rimType.value === 'hookless' ? HOOKLESS_ANCHORS : HOOKED_ANCHORS
    const anchor = anchors.find((item) => item.tire === width)
    if (anchor) {
      const ideal = Math.round((anchor.minRim + anchor.maxRim) / 2)
      return {
        minRim: anchor.minRim,
        maxRim: anchor.maxRim,
        ideal,
      }
    }

    let minRim: number
    let maxRim: number
    let ideal: number

    if (rimType.value === 'hookless') {
      if (width <= 30) {
        minRim = 23
        maxRim = 25
      } else if (width <= 33) {
        minRim = 23
        maxRim = 25
      } else if (width <= 40) {
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
        minRim = 36
        maxRim = 40
      }
      ideal = Math.round((minRim + maxRim) / 2)
    } else {
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

  return {
    parsedTireWidth,
    hooklessSafetyWarning,
    tireRimSuggestion,
  }
}
