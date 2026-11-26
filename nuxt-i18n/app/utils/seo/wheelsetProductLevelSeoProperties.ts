import type { WheelsetProductSpecification, WheelsetWheelProductSpecification } from '~/types/productDetail'
import type { ProductSeoPropertyValue, ProductSeoQuantitativeValue } from './types'

export interface WheelsetProductLevelSeoProperties {
  material?: string
  weightGrams?: number
  additionalProperties: ProductSeoPropertyValue[]
}

const cleanText = (value: string | null | undefined): string => String(value || '').trim()

const quantitativeValue = (value: number | null | undefined, unitCode: string): ProductSeoQuantitativeValue | null => {
  const normalized = Number(value)
  if (!Number.isFinite(normalized) || normalized <= 0) return null
  return { '@type': 'QuantitativeValue', value: normalized, unitCode }
}

export function buildWheelsetProductLevelSeoProperties(
  specification?: WheelsetProductSpecification | null,
): WheelsetProductLevelSeoProperties {
  if (!specification) return { additionalProperties: [] }

  const additionalProperties: ProductSeoPropertyValue[] = []
  const appendText = (propertyID: string, name: string, value: string | null | undefined) => {
    const normalized = cleanText(value)
    if (normalized) additionalProperties.push({ '@type': 'PropertyValue', propertyID, name, value: normalized })
  }
  const appendTexts = (propertyID: string, name: string, values: string[] | number[] | null | undefined) => {
    for (const value of values || []) appendText(propertyID, name, String(value))
  }
  const appendQuantity = (propertyID: string, name: string, value: number | null | undefined, unitCode: string) => {
    const normalized = quantitativeValue(value, unitCode)
    if (normalized) additionalProperties.push({ '@type': 'PropertyValue', propertyID, name, value: normalized })
  }

  const appendWheel = (
    position: 'front' | 'rear',
    label: 'Front wheel' | 'Rear wheel',
    wheel: WheelsetWheelProductSpecification,
  ) => {
    appendText(`${position}-wheel-size`, `${label} size`, wheel.wheel_size)
    appendTexts(`${position}-spoke-hole-count`, `${label} spoke hole count`, wheel.spoke_hole_counts)
    appendTexts(`${position}-lacing-pattern`, `${label} lacing pattern`, wheel.lacing_patterns)
    appendQuantity(`${position}-bead-seat-diameter-mm`, `${label} bead seat diameter (BSD)`, wheel.bead_seat_diameter_mm, 'MMT')
    appendText(`${position}-wheel-material`, `${label} material`, wheel.material)
    appendText(`${position}-brake-interface`, `${label} brake interface`, wheel.brake_interface)
    appendText(`${position}-rim-profile`, `${label} rim profile`, wheel.rim_profile)
    appendText(`${position}-tire-system`, `${label} tire system`, wheel.tire_system)
    appendQuantity(`${position}-rim-depth-mm`, `${label} rim depth`, wheel.rim_depth_mm, 'MMT')
    appendQuantity(`${position}-rim-internal-width-mm`, `${label} rim internal width`, wheel.rim_internal_width_mm, 'MMT')
    appendQuantity(`${position}-rim-external-width-mm`, `${label} rim external width`, wheel.rim_external_width_mm, 'MMT')
    appendText(`${position}-hub-interface`, `${label} hub interface`, wheel.hub_interface)
    appendText(`${position}-axle-standard`, `${label} axle standard`, wheel.axle_standard)
    if (position === 'rear') appendText('rear-freehub-body', 'Rear freehub body', wheel.freehub_body)
    appendText(`${position}-spoke-type`, `${label} spoke type`, wheel.spoke_type)
    appendQuantity(`${position}-wheel-weight-g`, `${label} weight`, wheel.weight_g, 'GRM')
    appendQuantity(`${position}-maximum-tire-pressure-bar`, `${label} maximum tire pressure`, wheel.max_tire_pressure_bar, 'BAR')
  }

  appendWheel('front', 'Front wheel', specification.front_wheel)
  appendWheel('rear', 'Rear wheel', specification.rear_wheel)
  appendQuantity('maximum-system-weight-kg', 'Maximum system weight', specification.max_system_weight_kg, 'KGM')
  appendTexts('certification', 'Certification', specification.certifications)

  const frontMaterial = cleanText(specification.front_wheel?.material)
  const rearMaterial = cleanText(specification.rear_wheel?.material)
  const netWeight = quantitativeValue(specification.net_weight_g, 'GRM')

  return {
    ...(frontMaterial && frontMaterial === rearMaterial ? { material: frontMaterial } : {}),
    ...(netWeight ? { weightGrams: netWeight.value } : {}),
    additionalProperties,
  }
}
