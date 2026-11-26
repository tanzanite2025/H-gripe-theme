import assert from 'node:assert/strict'
import type { WheelsetProductSpecification } from '../app/types/productDetail'
import { buildProductSeoDocument } from '../app/utils/seo/product'
import { buildWheelsetProductLevelSeoProperties } from '../app/utils/seo/wheelsetProductLevelSeoProperties'

const specification: WheelsetProductSpecification = {
  schema_version: 1,
  front_wheel: {
    wheel_size: '700C',
    spoke_hole_counts: [24, 28],
    lacing_patterns: ['2x', '3x'],
    bead_seat_diameter_mm: 622,
    material: 'Carbon fiber',
    brake_interface: 'Center Lock',
    rim_depth_mm: 45,
    rim_internal_width_mm: 25,
    rim_external_width_mm: 32,
    weight_g: 730,
    max_tire_pressure_bar: 5.5,
  },
  rear_wheel: {
    wheel_size: '650B',
    spoke_hole_counts: [28, 32],
    lacing_patterns: ['3x'],
    bead_seat_diameter_mm: 584,
    material: 'Carbon fiber',
    brake_interface: '6-bolt',
    rim_depth_mm: 35,
    rim_internal_width_mm: 27,
    rim_external_width_mm: 34,
    freehub_body: 'XDR',
    weight_g: 860,
    max_tire_pressure_bar: 4.2,
  },
  net_weight_g: 1590,
  max_system_weight_kg: 120,
  certifications: ['UCI approval'],
}

const properties = buildWheelsetProductLevelSeoProperties(specification)
assert.equal(properties.material, 'Carbon fiber')
assert.equal(properties.weightGrams, 1590)
assert.deepEqual(
  properties.additionalProperties.filter((property) => property.propertyID.endsWith('-wheel-size')),
  [
    { '@type': 'PropertyValue', propertyID: 'front-wheel-size', name: 'Front wheel size', value: '700C' },
    { '@type': 'PropertyValue', propertyID: 'rear-wheel-size', name: 'Rear wheel size', value: '650B' },
  ],
)
assert.deepEqual(
  properties.additionalProperties.find((property) => property.propertyID === 'front-rim-depth-mm')?.value,
  { '@type': 'QuantitativeValue', value: 45, unitCode: 'MMT' },
)
assert.deepEqual(
  properties.additionalProperties.find((property) => property.propertyID === 'rear-wheel-weight-g')?.value,
  { '@type': 'QuantitativeValue', value: 860, unitCode: 'GRM' },
)
assert.deepEqual(
  properties.additionalProperties.find((property) => property.propertyID === 'rear-freehub-body')?.value,
  'XDR',
)

const document = buildProductSeoDocument({
  name: 'Mixed diameter wheelset',
  mpn: 'FACTORY-WHEELSET-001',
  gtin: '00012345678905',
  imageUrls: ['/wheelset.webp'],
  material: properties.material,
  weightGrams: properties.weightGrams,
  additionalProperties: properties.additionalProperties,
  offer: { price: '1299.00', currency: 'USD', availability: 'in_stock' },
}, { siteOrigin: 'https://example.com', localizedPath: '/products/mixed-diameter-wheelset' })
assert.equal(document.schema?.['@type'], 'Product')
if (document.schema?.['@type'] !== 'Product') throw new Error('Expected a wheelset Product schema')
assert.equal(document.schema.mpn, 'FACTORY-WHEELSET-001')
assert.equal(document.schema.gtin, '00012345678905')
assert.deepEqual(document.schema.weight, { '@type': 'QuantitativeValue', value: 1590, unitCode: 'GRM' })
assert.equal(document.schema.additionalProperty?.find((property) => property.propertyID === 'front-wheel-size')?.value, '700C')
assert.equal(document.schema.additionalProperty?.find((property) => property.propertyID === 'rear-wheel-size')?.value, '650B')

const documentWithoutIdentifiers = buildProductSeoDocument({
  name: 'Wheelset without optional GTIN',
  mpn: 'FACTORY-WHEELSET-002',
  gtin: '',
  imageUrls: ['/wheelset.webp'],
}, { siteOrigin: 'https://example.com', localizedPath: '/products/wheelset-without-gtin' })
assert.equal(documentWithoutIdentifiers.schema?.['@type'], 'Product')
if (documentWithoutIdentifiers.schema?.['@type'] !== 'Product') throw new Error('Expected a Product schema without GTIN')
assert.equal(documentWithoutIdentifiers.schema.mpn, 'FACTORY-WHEELSET-002')
assert.ok(!Object.hasOwn(documentWithoutIdentifiers.schema, 'gtin'))

const mixedMaterials = buildWheelsetProductLevelSeoProperties({
  ...specification,
  rear_wheel: { ...specification.rear_wheel, material: 'Aluminum' },
})
assert.equal(mixedMaterials.material, undefined)

const emptyOptionalFacts = buildWheelsetProductLevelSeoProperties({
  schema_version: 1,
  front_wheel: { wheel_size: '700C', spoke_hole_counts: [24], lacing_patterns: ['2x'] },
  rear_wheel: { wheel_size: '700C', spoke_hole_counts: [28], lacing_patterns: ['3x'] },
})
assert.equal(emptyOptionalFacts.weightGrams, undefined)
assert.ok(!emptyOptionalFacts.additionalProperties.some((property) => property.propertyID.includes('rim-depth')))
assert.ok(!emptyOptionalFacts.additionalProperties.some((property) => property.propertyID === 'certification'))

console.log('PASS WHEELSET JSON-LD keeps front/rear engineering facts independent and omits absent values')
