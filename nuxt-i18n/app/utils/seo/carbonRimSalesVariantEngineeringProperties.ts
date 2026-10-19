import type { CarbonRimVariantSpecification } from '~/types/productDetail'
import type { ProductSeoPropertyValue } from './types'

// Use the SKU's own facts; never substitute fulfillment weight for rim net weight.
export function buildCarbonRimSalesVariantEngineeringProperties(specification?: CarbonRimVariantSpecification | null): ProductSeoPropertyValue[] {
  if (!specification) return []
  const properties: ProductSeoPropertyValue[] = [
    { '@type': 'PropertyValue', propertyID: 'spoke_hole_count', name: 'Spoke hole count', value: String(specification.spoke_hole_count) },
    { '@type': 'PropertyValue', propertyID: 'drilling_pattern', name: 'Drilling pattern', value: specification.drilling_pattern },
    { '@type': 'PropertyValue', propertyID: 'finish_code', name: 'Finish', value: specification.finish_code },
    { '@type': 'PropertyValue', propertyID: 'net_weight_grams', name: 'Net weight per rim', value: String(specification.net_weight_grams), unitText: 'g' },
  ]
  if (specification.erd_mm != null) {
    properties.push({ '@type': 'PropertyValue', propertyID: 'erd_mm', name: 'ERD', value: String(specification.erd_mm), unitText: 'mm' })
    if (specification.erd_measurement_notes) properties.push({ '@type': 'PropertyValue', propertyID: 'erd_measurement_notes', name: 'ERD measurement convention', value: specification.erd_measurement_notes })
  }
  return properties
}
