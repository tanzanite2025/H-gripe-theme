// This module only normalizes the user's display unit. Recommendation,
// interpolation, and derived metrics remain owned by the Go tire/rim domain.
export type TireWidthInputUnit = 'mm' | 'inch'
export type TireWidthInputValue = string | number
export type TireWidthInputPresetCategory = 'c-marking' | 'inch-marking'

export interface TireWidthInputPreset {
  id: string
  category: TireWidthInputPresetCategory
  label: string
  millimeters: number
  inches: number
  inputValue: string
  inputUnit: TireWidthInputUnit
}

const MILLIMETERS_PER_INCH = 25.4

const MIN_COMMON_TIRE_WIDTH_INCHES = 0.7
const MAX_COMMON_TIRE_WIDTH_INCHES = 5

// These are common tire sidewall markings, shown as a picker convenience.
// The backend remains the only source of engineering recommendations.
export const commonTireWidthInputPresets: readonly TireWidthInputPreset[] = [
  { id: '23c', category: 'c-marking', label: '23C', millimeters: 23, inches: 0.91, inputValue: '23', inputUnit: 'mm' },
  { id: '25c', category: 'c-marking', label: '25C', millimeters: 25, inches: 0.98, inputValue: '25', inputUnit: 'mm' },
  { id: '28c', category: 'c-marking', label: '28C', millimeters: 28, inches: 1.1, inputValue: '28', inputUnit: 'mm' },
  { id: '30c', category: 'c-marking', label: '30C', millimeters: 30, inches: 1.18, inputValue: '30', inputUnit: 'mm' },
  { id: '32c', category: 'c-marking', label: '32C', millimeters: 32, inches: 1.26, inputValue: '32', inputUnit: 'mm' },
  { id: '35c', category: 'c-marking', label: '35C', millimeters: 35, inches: 1.38, inputValue: '35', inputUnit: 'mm' },
  { id: '38c', category: 'c-marking', label: '38C', millimeters: 38, inches: 1.5, inputValue: '38', inputUnit: 'mm' },
  { id: '40c', category: 'c-marking', label: '40C', millimeters: 40, inches: 1.57, inputValue: '40', inputUnit: 'mm' },
  { id: '45c', category: 'c-marking', label: '45C', millimeters: 45, inches: 1.77, inputValue: '45', inputUnit: 'mm' },
  { id: '50c', category: 'c-marking', label: '50C', millimeters: 50, inches: 1.97, inputValue: '50', inputUnit: 'mm' },
  { id: '55c', category: 'c-marking', label: '55C', millimeters: 55, inches: 2.17, inputValue: '55', inputUnit: 'mm' },
  { id: '60c', category: 'c-marking', label: '60C', millimeters: 60, inches: 2.36, inputValue: '60', inputUnit: 'mm' },
  { id: '1-75-inch', category: 'inch-marking', label: '1.75″', millimeters: 44, inches: 1.75, inputValue: '1.75', inputUnit: 'inch' },
  { id: '2-0-inch', category: 'inch-marking', label: '2.0″', millimeters: 51, inches: 2, inputValue: '2.0', inputUnit: 'inch' },
  { id: '2-1-inch', category: 'inch-marking', label: '2.1″', millimeters: 53, inches: 2.1, inputValue: '2.1', inputUnit: 'inch' },
  { id: '2-25-inch', category: 'inch-marking', label: '2.25″', millimeters: 57, inches: 2.25, inputValue: '2.25', inputUnit: 'inch' },
  { id: '2-4-inch', category: 'inch-marking', label: '2.4″', millimeters: 61, inches: 2.4, inputValue: '2.4', inputUnit: 'inch' },
  { id: '2-5-inch', category: 'inch-marking', label: '2.5″', millimeters: 64, inches: 2.5, inputValue: '2.5', inputUnit: 'inch' },
  { id: '2-6-inch', category: 'inch-marking', label: '2.6″', millimeters: 66, inches: 2.6, inputValue: '2.6', inputUnit: 'inch' },
  { id: '2-8-inch', category: 'inch-marking', label: '2.8″', millimeters: 71, inches: 2.8, inputValue: '2.8', inputUnit: 'inch' },
  { id: '3-0-inch', category: 'inch-marking', label: '3.0″', millimeters: 76, inches: 3, inputValue: '3.0', inputUnit: 'inch' },
  { id: '4-0-inch', category: 'inch-marking', label: '4.0″', millimeters: 102, inches: 4, inputValue: '4.0', inputUnit: 'inch' },
  { id: '4-5-inch', category: 'inch-marking', label: '4.5″', millimeters: 114, inches: 4.5, inputValue: '4.5', inputUnit: 'inch' },
  { id: '5-0-inch', category: 'inch-marking', label: '5.0″', millimeters: 127, inches: 5, inputValue: '5.0', inputUnit: 'inch' },
]

const getExplicitTireWidthInputUnit = (input: TireWidthInputValue): TireWidthInputUnit | null => {
  const normalizedInput = String(input).trim().toLowerCase()
  if (/(?:mm|c)$/.test(normalizedInput)) return 'mm'
  if (/(?:inch|in|["'″])$/.test(normalizedInput)) return 'inch'
  return null
}

export const stripTireWidthInputUnitSuffix = (input: TireWidthInputValue): string => (
  String(input)
    .trim()
    .replace(/\s*(?:mm|c|inch|in|["'″])$/i, '')
    .trim()
)

export const inferTireWidthInputUnit = (
  input: TireWidthInputValue,
  fallbackUnit: TireWidthInputUnit = 'mm',
): TireWidthInputUnit => {
  const explicitUnit = getExplicitTireWidthInputUnit(input)
  if (explicitUnit !== null) return explicitUnit

  const numericInput = Number(stripTireWidthInputUnitSuffix(input))
  if (fallbackUnit === 'mm'
    && Number.isFinite(numericInput)
    && numericInput >= MIN_COMMON_TIRE_WIDTH_INCHES
    && numericInput <= MAX_COMMON_TIRE_WIDTH_INCHES) {
    return 'inch'
  }

  return fallbackUnit
}

export const isLikelyInchTireWidthInput = (
  input: TireWidthInputValue,
  currentUnit: TireWidthInputUnit,
): boolean => {
  return currentUnit === 'mm' && inferTireWidthInputUnit(input, currentUnit) === 'inch'
}

export const parseTireWidthInputAsMillimeters = (
  input: TireWidthInputValue,
  unit: TireWidthInputUnit,
): number | null => {
  const normalizedInput = stripTireWidthInputUnitSuffix(input)
  if (!normalizedInput) return null

  const numericInput = Number(normalizedInput)
  if (!Number.isFinite(numericInput)) return null

  const effectiveUnit = getExplicitTireWidthInputUnit(input) || unit
  return effectiveUnit === 'inch'
    ? Math.round(numericInput * MILLIMETERS_PER_INCH)
    : numericInput
}

export const convertTireWidthInputBetweenUnits = (
  input: TireWidthInputValue,
  fromUnit: TireWidthInputUnit,
  toUnit: TireWidthInputUnit,
): string => {
  if (fromUnit === toUnit) return String(input)

  const widthInMillimeters = parseTireWidthInputAsMillimeters(input, fromUnit)
  if (widthInMillimeters === null) return String(input)

  return toUnit === 'inch'
    ? String(Number((widthInMillimeters / MILLIMETERS_PER_INCH).toFixed(2)))
    : String(widthInMillimeters)
}
