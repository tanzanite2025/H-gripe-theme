import type { SpokeGaugeOption, SpokesProductSpecification } from '~/types/productDetail'
import type { ProductSeoPropertyValue, ProductSeoQuantitativeValue } from './types'

export interface SpokesProductLevelSeoProperties {
  model?: string
  material?: string
  color?: string
  additionalProperties: ProductSeoPropertyValue[]
}

export interface SpokesProductLevelSeoIdentifiers {
  mpn?: string
  gtin?: string
}

const cleanText = (value: string | null | undefined): string => String(value || '').trim()

export function buildSpokesProductLevelSeoIdentifiers(
  mpn: string | null | undefined,
  gtin: string | null | undefined,
): SpokesProductLevelSeoIdentifiers {
  const normalizedMPN = cleanText(mpn)
  const normalizedGTIN = cleanText(gtin)
  return {
    ...(normalizedMPN ? { mpn: normalizedMPN } : {}),
    ...(normalizedGTIN ? { gtin: normalizedGTIN } : {}),
  }
}

const quantitativeValue = (value: number | null | undefined, unitCode: string): ProductSeoQuantitativeValue | null => {
  const normalized = Number(value)
  if (!Number.isFinite(normalized) || normalized <= 0) return null
  return { '@type': 'QuantitativeValue', value: normalized, unitCode }
}

const propertyIdPart = (value: string): string => value
  .trim()
  .toLowerCase()
  .replace(/[^a-z0-9]+/g, '-')
  .replace(/^-+|-+$/g, '') || 'unspecified'

export function buildSpokesProductLevelSeoProperties(
  specification?: SpokesProductSpecification | null,
): SpokesProductLevelSeoProperties {
  if (!specification) return { additionalProperties: [] }

  const additionalProperties: ProductSeoPropertyValue[] = []
  const appendText = (propertyID: string, name: string, value: string | null | undefined) => {
    const normalized = cleanText(value)
    if (normalized) additionalProperties.push({ '@type': 'PropertyValue', propertyID, name, value: normalized })
  }
  const appendQuantity = (propertyID: string, name: string, value: number | null | undefined, unitCode: string) => {
    const normalized = quantitativeValue(value, unitCode)
    if (normalized) additionalProperties.push({ '@type': 'PropertyValue', propertyID, name, value: normalized })
  }
  const appendTexts = (propertyID: string, name: string, values: string[] | null | undefined) => {
    for (const value of values || []) appendText(propertyID, name, value)
  }

  const model = cleanText(specification.model_name) || undefined
  const colors = (specification.color_options || []).map(cleanText).filter(Boolean)
  if (colors.length > 1) appendTexts('color-option', 'Color option', colors)

  appendText('manufacturer-family', 'Manufacturer family', specification.manufacturer_family)
  appendText('manufacturer-description', 'Manufacturer description', specification.manufacturer_description)

  for (const [index, gauge] of (specification.gauge_options || []).entries()) {
    appendSpokeGaugeOptionProperties(gauge, index, appendText, appendQuantity)
  }

  const strength = specification.middle_strength_n_per_mm2
  if (strength) {
    const minimum = Number(strength.minimum_n_per_mm2)
    const maximum = Number(strength.maximum_n_per_mm2)
    if (Number.isFinite(minimum) && minimum > 0 && Number.isFinite(maximum) && maximum >= minimum) {
      const value = minimum === maximum ? String(minimum) : `${minimum} - ${maximum}`
      additionalProperties.push({
        '@type': 'PropertyValue',
        propertyID: 'middle-strength-n-per-mm2',
        name: 'Middle strength',
        value,
        unitText: 'N/mm²',
      })
    }
  }

  appendTexts('manufacturer-option', 'Manufacturer option', specification.manufacturer_options)
  appendTexts('recommended-use', 'Manufacturer recommended use', specification.recommended_uses)
  appendTexts('manufacturer-claim', 'Manufacturer claim', specification.manufacturer_claims)

  return {
    ...(model ? { model } : {}),
    ...(cleanText(specification.material) ? { material: cleanText(specification.material) } : {}),
    ...(colors.length === 1 ? { color: colors[0] } : {}),
    additionalProperties,
  }
}

function appendSpokeGaugeOptionProperties(
  gauge: SpokeGaugeOption,
  index: number,
  appendText: (propertyID: string, name: string, value: string | null | undefined) => void,
  appendQuantity: (propertyID: string, name: string, value: number | null | undefined, unitCode: string) => void,
): void {
  const gaugeName = cleanText(gauge.gauge)
  const gaugeKey = `${index + 1}-${propertyIdPart(gaugeName)}`
  if (!gaugeName) return

  appendText(`gauge-${gaugeKey}`, 'Gauge', gaugeName)
  const measurements: Array<[keyof SpokeGaugeOption, string, string]> = [
    ['diameter_mm', 'Diameter', 'diameter'],
    ['hub_end_diameter_mm', 'Hub-end diameter', 'hub-end-diameter'],
    ['middle_diameter_mm', 'Middle diameter', 'middle-diameter'],
    ['middle_width_mm', 'Middle width', 'middle-width'],
    ['middle_thickness_mm', 'Middle thickness', 'middle-thickness'],
    ['thread_end_diameter_mm', 'Thread-end diameter', 'thread-end-diameter'],
  ]
  for (const [field, label, property] of measurements) {
    appendQuantity(`gauge-${gaugeKey}-${property}-mm`, `${gaugeName} ${label}`, gauge[field] as number | undefined, 'MMT')
  }
  appendQuantity(`gauge-${gaugeKey}-length-min-mm`, `${gaugeName} minimum length`, gauge.length_min_mm, 'MMT')
  appendQuantity(`gauge-${gaugeKey}-length-max-mm`, `${gaugeName} maximum length`, gauge.length_max_mm, 'MMT')

  const reference = gauge.weight_reference
  if (reference) {
    const spokeCount = Number(reference.spoke_count)
    const lengthMM = Number(reference.length_mm)
    const totalWeightG = Number(reference.total_weight_g)
    if ([spokeCount, lengthMM, totalWeightG].every(Number.isFinite) && spokeCount > 0 && lengthMM > 0 && totalWeightG > 0) {
      appendQuantity(
        `gauge-${gaugeKey}-weight-reference-g`,
        `${gaugeName} reference weight (${spokeCount} spokes at ${lengthMM} mm)`,
        totalWeightG,
        'GRM',
      )
    }
  }
}
