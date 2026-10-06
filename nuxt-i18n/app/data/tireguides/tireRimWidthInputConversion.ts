// This module only normalizes the user's display unit. Recommendation,
// interpolation, and derived metrics remain owned by the Go tire/rim domain.
export type TireWidthInputUnit = 'mm' | 'inch'

const MILLIMETERS_PER_INCH = 25.4

export const parseTireWidthInputAsMillimeters = (
  input: string,
  unit: TireWidthInputUnit,
): number | null => {
  const normalizedInput = input.trim()
  if (!normalizedInput) return null

  const numericInput = Number(normalizedInput)
  if (!Number.isFinite(numericInput)) return null

  return unit === 'inch'
    ? Math.round(numericInput * MILLIMETERS_PER_INCH)
    : numericInput
}

export const convertTireWidthInputBetweenUnits = (
  input: string,
  fromUnit: TireWidthInputUnit,
  toUnit: TireWidthInputUnit,
): string => {
  if (fromUnit === toUnit) return input

  const widthInMillimeters = parseTireWidthInputAsMillimeters(input, fromUnit)
  if (widthInMillimeters === null) return input

  return toUnit === 'inch'
    ? String(Number((widthInMillimeters / MILLIMETERS_PER_INCH).toFixed(2)))
    : String(widthInMillimeters)
}
