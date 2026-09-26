export type TireRimSystem = 'hookless' | 'hooked'

/**
 * Temporary UI-only demonstration cap. This is not an approved safety limit
 * and must not be used as the domain/API recommendation source.
 */
export const HOOKLESS_MAX_PRESSURE_PSI = 73
export const HOOKLESS_MAX_PRESSURE_BAR = 5
export const PSI_PER_BAR = 14.5037738

/**
 * Apply the rim-system safety ceiling to a pressure recommendation.
 * Returns null for non-numeric recommendations so callers cannot render NaN.
 */
export const capRecommendedPressurePsi = (
  pressurePsi: number,
  rimSystem: TireRimSystem,
): number | null => {
  if (!Number.isFinite(pressurePsi)) return null

  return rimSystem === 'hookless'
    ? Math.min(pressurePsi, HOOKLESS_MAX_PRESSURE_PSI)
    : pressurePsi
}

export const capRecommendedPressureBar = (
  pressureBar: number,
  rimSystem: TireRimSystem,
): number | null => {
  if (!Number.isFinite(pressureBar)) return null

  return rimSystem === 'hookless'
    ? Math.min(pressureBar, HOOKLESS_MAX_PRESSURE_BAR)
    : pressureBar
}

export const psiToBar = (pressurePsi: number): number => (
  pressurePsi / PSI_PER_BAR
)
