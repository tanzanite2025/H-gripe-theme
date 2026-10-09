export type TireRimEngineeringRimSystem = 'hookless' | 'hooked'

export type TireRimEngineeringVerdict = 'critical' | 'recommended' | 'reference'

export type TireRimEngineeringVerdictReason =
  | 'below_minimum'
  | 'above_maximum'
  | 'recommended_window'
  | 'reference_window'
  | 'transitional_rim'

export interface TireRimEngineeringWidthRange {
  min: number
  max: number
}

export interface TireRimEngineeringRimRule {
  rimSystem: TireRimEngineeringRimSystem
  rimInnerWidthMm: number
  allowedTireWidthRange: TireRimEngineeringWidthRange
  recommendedTireWidthRange: TireRimEngineeringWidthRange
  defaultVerdict?: 'recommended' | 'reference'
}

export interface TireRimEngineeringReferenceCalculation {
  verdict: TireRimEngineeringVerdict
  verdictReason: TireRimEngineeringVerdictReason
  rimSystem: TireRimEngineeringRimSystem
  rimInnerWidthMm: number
  tireWidthMm: number
  allowedTireWidthRange: TireRimEngineeringWidthRange
  recommendedTireWidthRange: TireRimEngineeringWidthRange
  inflatedTireWidthMm: number
  aeroTargetOuterWidthMm: number
}

export const TIRE_RIM_ENGINEERING_REFERENCE_MODEL_VERSION = 'tire-rim-engineering-reference-v1'

export const TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM = [19, 21, 23, 25, 28, 30] as const

export const TIRE_RIM_ENGINEERING_TIRE_WIDTH_PRESETS_MM = [23, 25, 28, 30, 32, 35, 38, 40, 45, 50, 55, 60] as const

const createTireRimEngineeringWidthRange = (
  min: number,
  max: number,
): TireRimEngineeringWidthRange => ({ min, max })

const hooklessTireRimEngineeringRules: Record<number, TireRimEngineeringRimRule> = {
  19: {
    rimSystem: 'hookless',
    rimInnerWidthMm: 19,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(28, 32),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(28, 28),
    defaultVerdict: 'reference',
  },
  21: {
    rimSystem: 'hookless',
    rimInnerWidthMm: 21,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(28, 32),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(28, 30),
  },
  23: {
    rimSystem: 'hookless',
    rimInnerWidthMm: 23,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(28, 32),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(28, 32),
  },
  25: {
    rimSystem: 'hookless',
    rimInnerWidthMm: 25,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(29, 34),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(30, 32),
  },
  28: {
    rimSystem: 'hookless',
    rimInnerWidthMm: 28,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(38, 50),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(40, 45),
  },
  30: {
    rimSystem: 'hookless',
    rimInnerWidthMm: 30,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(42, 60),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(48, 57),
  },
}

const hookedTireRimEngineeringRules: Record<number, TireRimEngineeringRimRule> = {
  19: {
    rimSystem: 'hooked',
    rimInnerWidthMm: 19,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(23, 35),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(23, 28),
  },
  21: {
    rimSystem: 'hooked',
    rimInnerWidthMm: 21,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(23, 40),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(25, 30),
  },
  23: {
    rimSystem: 'hooked',
    rimInnerWidthMm: 23,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(23, 45),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(28, 32),
  },
  25: {
    rimSystem: 'hooked',
    rimInnerWidthMm: 25,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(23, 45),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(26, 35),
  },
  28: {
    rimSystem: 'hooked',
    rimInnerWidthMm: 28,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(29, 60),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(36, 60),
  },
  30: {
    rimSystem: 'hooked',
    rimInnerWidthMm: 30,
    allowedTireWidthRange: createTireRimEngineeringWidthRange(33, 60),
    recommendedTireWidthRange: createTireRimEngineeringWidthRange(46, 60),
  },
}

export const tireRimEngineeringReferenceRules: Record<
  TireRimEngineeringRimSystem,
  Record<number, TireRimEngineeringRimRule>
> = {
  hookless: hooklessTireRimEngineeringRules,
  hooked: hookedTireRimEngineeringRules,
}

const isWithinTireRimEngineeringWidthRange = (
  value: number,
  range: TireRimEngineeringWidthRange,
): boolean => value >= range.min && value <= range.max

const roundTireRimEngineeringDisplayMetric = (value: number): number => (
  Math.round(value * 10) / 10
)

export const calculateTireRimEngineeringReference = (
  rimSystem: TireRimEngineeringRimSystem,
  rimInnerWidthMm: number,
  tireWidthMm: number,
): TireRimEngineeringReferenceCalculation | null => {
  if (!Number.isInteger(rimInnerWidthMm) || !Number.isInteger(tireWidthMm)) return null

  const rule = tireRimEngineeringReferenceRules[rimSystem][rimInnerWidthMm]
  if (!rule) return null

  const isBelowMinimum = tireWidthMm < rule.allowedTireWidthRange.min
  const isAboveMaximum = tireWidthMm > rule.allowedTireWidthRange.max
  const isRecommended = isWithinTireRimEngineeringWidthRange(
    tireWidthMm,
    rule.recommendedTireWidthRange,
  )

  let verdict: TireRimEngineeringVerdict = 'reference'
  let verdictReason: TireRimEngineeringVerdictReason = 'reference_window'

  if (isBelowMinimum) {
    verdict = 'critical'
    verdictReason = 'below_minimum'
  } else if (isAboveMaximum) {
    verdict = 'critical'
    verdictReason = 'above_maximum'
  } else if (rule.defaultVerdict === 'reference') {
    verdict = 'reference'
    verdictReason = 'transitional_rim'
  } else if (isRecommended) {
    verdict = 'recommended'
    verdictReason = 'recommended_window'
  }

  const inflatedTireWidthMm = tireWidthMm + 0.4 * (rimInnerWidthMm - 19)

  return {
    verdict,
    verdictReason,
    rimSystem,
    rimInnerWidthMm,
    tireWidthMm,
    allowedTireWidthRange: rule.allowedTireWidthRange,
    recommendedTireWidthRange: rule.recommendedTireWidthRange,
    inflatedTireWidthMm: roundTireRimEngineeringDisplayMetric(inflatedTireWidthMm),
    aeroTargetOuterWidthMm: roundTireRimEngineeringDisplayMetric(inflatedTireWidthMm * 1.05),
  }
}
