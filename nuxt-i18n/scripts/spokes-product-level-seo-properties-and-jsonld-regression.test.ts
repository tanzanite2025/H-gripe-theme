import assert from 'node:assert/strict'
import type { SpokesProductSpecification } from '../app/types/productDetail'
import { buildProductSeoDocument } from '../app/utils/seo/product'
import {
  buildSpokesProductLevelSeoIdentifiers,
  buildSpokesProductLevelSeoProperties,
} from '../app/utils/seo/spokesProductLevelSeoProperties'

const specification: SpokesProductSpecification = {
  schema_version: 1,
  model_name: 'CX-Ray',
  manufacturer_family: 'Aero spokes',
  material: '18/8 stainless steel',
  gauge_options: [{
    gauge: '14G',
    hub_end_diameter_mm: 2,
    middle_width_mm: 2.2,
    middle_thickness_mm: 0.9,
    length_min_mm: 141,
    length_max_mm: 310,
    weight_reference: { spoke_count: 64, length_mm: 260, total_weight_g: 279 },
  }],
  middle_strength_n_per_mm2: { minimum_n_per_mm2: 1600, maximum_n_per_mm2: 1600 },
  color_options: ['Black oxidized', 'Silver'],
  recommended_uses: ['Road racing', 'Triathlon', 'Downhill biking'],
  manufacturer_claims: ['Aerodynamic bladed spoke'],
  source_urls: ['https://sapim.eu/wp-content/uploads/2026/03/CX-Ray_hr-scaled.png'],
}

const properties = buildSpokesProductLevelSeoProperties(specification)
assert.equal(properties.model, 'CX-Ray')
assert.equal(properties.material, '18/8 stainless steel')
assert.equal(properties.color, undefined)
assert.deepEqual(properties.additionalProperties.find((property) => property.propertyID === 'gauge-1-14g-hub-end-diameter-mm')?.value, {
  '@type': 'QuantitativeValue', value: 2, unitCode: 'MMT',
})
assert.deepEqual(properties.additionalProperties.find((property) => property.propertyID === 'gauge-1-14g-length-min-mm')?.value, {
  '@type': 'QuantitativeValue', value: 141, unitCode: 'MMT',
})
assert.deepEqual(properties.additionalProperties.find((property) => property.propertyID === 'gauge-1-14g-weight-reference-g')?.value, {
  '@type': 'QuantitativeValue', value: 279, unitCode: 'GRM',
})
assert.equal(properties.additionalProperties.find((property) => property.propertyID === 'middle-strength-n-per-mm2')?.unitText, 'N/mm²')
assert.ok(!properties.additionalProperties.some((property) => property.propertyID.includes('source')))

const productDocument = buildProductSeoDocument({
  name: 'SAPIM CX-Ray stainless steel spoke',
  model: properties.model,
  material: properties.material,
  additionalProperties: properties.additionalProperties,
  mpn: '',
  gtin: '',
  imageUrls: ['/sapim-cx-ray.webp'],
}, { siteOrigin: 'https://example.com', localizedPath: '/products/sapim-cx-ray' })
assert.equal(productDocument.schema?.['@type'], 'Product')
if (productDocument.schema?.['@type'] !== 'Product') throw new Error('Expected a spoke Product schema')
assert.equal(productDocument.schema.model, 'CX-Ray')
assert.equal(productDocument.schema.material, '18/8 stainless steel')
assert.ok(!Object.hasOwn(productDocument.schema, 'mpn'))
assert.ok(!Object.hasOwn(productDocument.schema, 'gtin'))
assert.equal(productDocument.schema.offers, undefined)

const spokesIdentifiers = buildSpokesProductLevelSeoIdentifiers(' SAPIM-CXRAY-001 ', '4006381333931')
assert.deepEqual(spokesIdentifiers, { mpn: 'SAPIM-CXRAY-001', gtin: '4006381333931' })
const identifiedProductDocument = buildProductSeoDocument({
  name: 'SAPIM CX-Ray stainless steel spoke',
  model: properties.model,
  mpn: spokesIdentifiers.mpn,
  gtin: spokesIdentifiers.gtin,
  imageUrls: ['/sapim-cx-ray.webp'],
}, { siteOrigin: 'https://example.com', localizedPath: '/products/sapim-cx-ray' })
assert.equal(identifiedProductDocument.schema?.mpn, 'SAPIM-CXRAY-001')
assert.equal(identifiedProductDocument.schema?.gtin, '4006381333931')

const productGroupDocument = buildProductSeoDocument({
  name: 'SAPIM CX-Ray stainless steel spoke',
  model: properties.model,
  material: properties.material,
  mpn: spokesIdentifiers.mpn,
  gtin: spokesIdentifiers.gtin,
  additionalProperties: properties.additionalProperties,
  imageUrls: ['/sapim-cx-ray.webp'],
  variants: [
    { name: '260 mm', sku: 'CX-RAY-260', price: '1.00', currency: 'USD', availability: 'in_stock', localizedPath: '/products/sapim-cx-ray?variant=260' },
    { name: '262 mm', sku: 'CX-RAY-262', price: '1.00', currency: 'USD', availability: 'in_stock', localizedPath: '/products/sapim-cx-ray?variant=262' },
  ],
}, { siteOrigin: 'https://example.com', localizedPath: '/products/sapim-cx-ray' })
assert.equal(productGroupDocument.schema?.['@type'], 'ProductGroup')
if (productGroupDocument.schema?.['@type'] !== 'ProductGroup') throw new Error('Expected a spoke ProductGroup schema')
assert.equal(productGroupDocument.schema.model, 'CX-Ray')
assert.equal(productGroupDocument.schema.material, '18/8 stainless steel')
assert.equal(productGroupDocument.schema.mpn, 'SAPIM-CXRAY-001')
assert.equal(productGroupDocument.schema.gtin, '4006381333931')
assert.ok(productGroupDocument.schema.additionalProperty?.some((property) => property.propertyID === 'middle-strength-n-per-mm2'))
assert.equal(productGroupDocument.schema.hasVariant[0].model, 'CX-Ray')
assert.equal(productGroupDocument.schema.hasVariant[0].mpn, undefined)
assert.equal(productGroupDocument.schema.hasVariant[0].gtin, undefined)
assert.ok(productGroupDocument.schema.hasVariant[0].additionalProperty?.some((property) => property.propertyID === 'gauge-1-14g-middle-width-mm'))

const normalizedGaugeLabels = buildSpokesProductLevelSeoProperties({
  ...specification,
  gauge_options: [{ gauge: 'A/B', diameter_mm: 2 }, { gauge: 'A-B', diameter_mm: 1.8 }],
})
const gaugeDiameterPropertyIDs = normalizedGaugeLabels.additionalProperties
  .filter((property) => property.propertyID.endsWith('diameter-mm'))
  .map((property) => property.propertyID)
assert.deepEqual(gaugeDiameterPropertyIDs, ['gauge-1-a-b-diameter-mm', 'gauge-2-a-b-diameter-mm'])

const singleColor = buildSpokesProductLevelSeoProperties({ ...specification, color_options: ['Black oxidized'] })
assert.equal(singleColor.color, 'Black oxidized')
assert.ok(!singleColor.additionalProperties.some((property) => property.propertyID === 'color-option'))

console.log('PASS spoke Product JSON-LD preserves verified product identifiers and engineering facts without fabricating variant identifiers or SKU offers')
