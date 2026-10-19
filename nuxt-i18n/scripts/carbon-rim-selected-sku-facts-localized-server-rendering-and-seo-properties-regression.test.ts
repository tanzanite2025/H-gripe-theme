import assert from 'node:assert/strict'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import vuePlugin from '@vitejs/plugin-vue'
import { createServer } from 'vite'
import { createSSRApp, h } from 'vue'
import { createI18n } from 'vue-i18n'
import { renderToString } from '@vue/server-renderer'
import { buildProductSeoDocument } from '../app/utils/seo/product'
import { buildCarbonRimSalesVariantEngineeringProperties } from '../app/utils/seo/carbonRimSalesVariantEngineeringProperties'

const frontendDirectory = fileURLToPath(new URL('..', import.meta.url))
const compilationServer = await createServer({
  root: frontendDirectory, configFile: false,
  plugins: [{
    name: 'provide-locale-for-carbon-rim-selected-sku-facts', enforce: 'pre',
    resolveId(source) { if (source === '#imports') return '\0carbon-rim-facts-imports' },
    load(id) { if (id === '\0carbon-rim-facts-imports') return "export { useI18n } from 'vue-i18n'" },
  }, vuePlugin()],
  resolve: { alias: { '~': path.join(frontendDirectory, 'app') } },
  server: { middlewareMode: true, hmr: false, ws: false, watch: null },
})

try {
  const { default: selectedVariantFacts } = await compilationServer.ssrLoadModule('/app/components/shop/product-detail/CarbonRimSelectedSalesVariantFacts.vue')
  for (const locale of ['en', 'zh_cn']) {
    for (const specification of [null, { net_weight_grams: 430 }, { net_weight_grams: 435, erd_mm: 532, erd_measurement_notes: 'Nipple seat <measured>' }]) {
      const app = createSSRApp({ render: () => h(selectedVariantFacts, { specification }) })
      app.use(createI18n({ legacy: false, locale, messages: { [locale]: {} } }))
      const markup = await renderToString(app)
      if (!specification) { assert.ok(!markup.includes('<dl')); continue }
      assert.ok(markup.includes(`${specification.net_weight_grams} g`))
      assert.ok(markup.includes(locale === 'zh_cn' ? '单只车圈净重' : 'Net weight per rim'))
      assert.equal(markup.includes('532 mm'), 'erd_mm' in specification)
      assert.equal(markup.includes('ERD'), 'erd_mm' in specification)
      if ('erd_mm' in specification) assert.ok(markup.includes('Nipple seat &lt;measured&gt;'))
      assert.ok(!markup.includes('pressure'))
    }
  }
  const additionalProperties = [{ '@type': 'PropertyValue' as const, propertyID: 'internal_width_mm', name: 'Internal width', value: '25', unitText: 'mm' }, { '@type': 'PropertyValue' as const, propertyID: 'wheel_size', name: 'Wheel size', value: '700C' }, { '@type': 'PropertyValue' as const, propertyID: 'material', name: 'Material', value: 'Carbon fiber · T700 / T800' }, { '@type': 'PropertyValue' as const, propertyID: 'riding_uses', name: 'Intended riding disciplines', value: 'Road cycling / Gravel cycling' }]
  const variants = [24,28].map((holes, index) => ({ id:index+1, sku:`R45-${holes}`, price:index ? '209.00':'199.00', currency:'USD', availability:'in_stock' as const, localizedPath:`/products/r45?variant=${index+1}`, weightGrams:430+index*5, additionalProperties:buildCarbonRimSalesVariantEngineeringProperties({spoke_hole_count:holes,drilling_pattern:'uniform',finish_code:'matte-black',net_weight_grams:430+index*5,...(index ? {erd_mm:532,erd_measurement_notes:'Nipple seats'} : {})}) }))
  for (const variantInputs of [undefined, variants]) {
    const seo = buildProductSeoDocument({ name: 'Carbon rim R45', size: '700C', material: 'Carbon fiber · T700 / T800', imageUrls: ['/rim.webp'], additionalProperties, variants:variantInputs, variesBy:['spoke_hole_count','drilling_pattern','finish_code'] }, { siteOrigin: 'https://example.com', localizedPath: '/products/r45' })
    assert.deepEqual(seo.schema?.additionalProperty, additionalProperties)
    assert.equal(seo.schema?.size,'700C')
    assert.equal(seo.schema?.material,'Carbon fiber · T700 / T800')
    if (variantInputs) {
      assert.equal(seo.schema?.['@type'], 'ProductGroup')
      if (seo.schema?.['@type'] !== 'ProductGroup') throw new Error('Expected ProductGroup')
      assert.equal(seo.schema.hasVariant.length,2)
      assert.deepEqual(seo.schema.variesBy,['spoke_hole_count','drilling_pattern','finish_code'])
      for (const [index,variant] of seo.schema.hasVariant.entries()) {
        assert.equal(variant.sku,variants[index].sku)
        assert.equal(variant.size,'700C')
        assert.equal(variant.material,'Carbon fiber · T700 / T800')
        assert.equal(variant.additionalProperty?.find(property=>property.propertyID==='riding_uses')?.value,'Road cycling / Gravel cycling')
        assert.equal(variant.offers?.price,variants[index].price)
        assert.equal(variant.offers?.url,`https://example.com/products/r45?variant=${index+1}`)
        assert.deepEqual(variant.weight,{'@type':'QuantitativeValue',value:430+index*5,unitCode:'GRM'})
        assert.equal(variant.additionalProperty?.find(property=>property.propertyID==='spoke_hole_count')?.value,String(index ? 28:24))
        assert.equal(Boolean(variant.additionalProperty?.find(property=>property.propertyID==='erd_mm')),Boolean(index))
        assert.equal(variant.isVariantOf?.['@id'],seo.schema['@id'])
      }
    }
  }
  const singleVariantSeo = buildProductSeoDocument({name:'Carbon rim R45',imageUrls:['/rim.webp'],additionalProperties:[...additionalProperties,...variants[0].additionalProperties],weightGrams:430,offer:{sku:'R45-24',price:'199.00',currency:'USD',availability:'in_stock'}},{siteOrigin:'https://example.com',localizedPath:'/products/r45'})
  assert.equal(singleVariantSeo.schema?.['@type'],'Product')
  if(singleVariantSeo.schema?.['@type'] !== 'Product') throw new Error('Expected single Product')
  assert.equal(singleVariantSeo.schema.weight?.value,430)
  assert.equal(singleVariantSeo.schema.additionalProperty?.find(property=>property.propertyID==='spoke_hole_count')?.value,'24')
  assert.ok(!singleVariantSeo.schema.additionalProperty?.some(property=>property.propertyID==='erd_mm'))
  console.log('PASS carbon rim SKU facts render in English/Chinese, absent ERD stays absent, and product SEO retains engineering properties')
} finally {
  await compilationServer.close()
}
