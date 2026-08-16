import assert from 'node:assert/strict'
import { buildProductSeoDocument } from '../../app/utils/seo/product'
import { createSeoJsonLdScript } from '../../app/utils/seo/jsonLd'

const seo = buildProductSeoDocument(
  {
    name: 'Demo Product',
    brand: 'HACK-GRIPE',
    metaTitle: 'Demo Product | Brand',
    metaDescription: '<p>Clean <strong>description</strong>.</p>',
    imageUrls: ['/media/demo.jpg', '', null, '/media/demo.jpg'],
    sku: 'SKU-001',
    offer: {
      price: 123.45,
      currency: 'usd',
      availability: 'in_stock',
    },
    aggregateRating: {
      ratingValue: 4.6,
      reviewCount: 12,
    },
    shippingDetails: {
      country: 'us',
      amount: 0,
      currency: 'usd',
      freeShipping: true,
      etaMinDays: 3,
      etaMaxDays: 7,
    },
  },
  {
    siteOrigin: 'https://example.com',
    localizedPath: '/zh_cn/shop/demo-product',
  },
)

assert.equal(seo.title, 'Demo Product | Brand')
assert.equal(seo.description, 'Clean description.')
assert.deepEqual(seo.images, ['https://example.com/media/demo.jpg'])
assert.equal(seo.canonicalUrl, 'https://example.com/zh_cn/shop/demo-product')
assert.equal(seo.schema?.name, 'Demo Product')
assert.deepEqual(seo.schema?.brand, { '@type': 'Brand', name: 'HACK-GRIPE' })
if (seo.schema?.['@type'] !== 'Product') {
  throw new Error('Expected a Product schema for a single configuration.')
}
assert.deepEqual(seo.schema.image, ['https://example.com/media/demo.jpg'])
assert.equal(seo.schema.sku, 'SKU-001')
assert.equal(seo.schema.offers?.price, 123.45)
assert.equal(seo.schema.offers?.priceCurrency, 'USD')
assert.equal(seo.schema.offers?.availability, 'https://schema.org/InStock')
assert.deepEqual(seo.schema.aggregateRating, {
  '@type': 'AggregateRating',
  ratingValue: 4.6,
  reviewCount: 12,
  ratingCount: 12,
  bestRating: 5,
  worstRating: 1,
})
assert.deepEqual(seo.schema.offers?.shippingDetails, {
  '@type': 'OfferShippingDetails',
  shippingRate: { '@type': 'MonetaryAmount', value: 0, currency: 'USD' },
  shippingDestination: { '@type': 'DefinedRegion', addressCountry: 'US' },
  deliveryTime: {
    '@type': 'ShippingDeliveryTime',
    transitTime: {
      '@type': 'QuantitativeValue',
      minValue: 3,
      maxValue: 7,
      unitCode: 'DAY',
    },
  },
})

const groupSeo = buildProductSeoDocument(
  {
    name: 'Variant Product',
    productGroupId: 'product-42',
    variesBy: ['https://schema.org/color'],
    imageUrls: ['/media/variant.jpg'],
    variants: [
      {
        id: 1,
        name: 'Variant Product Red',
        sku: 'RED-001',
        price: 99,
        currency: 'USD',
        availability: 'in_stock',
        localizedPath: '/shop/variant-product?variant=1',
      },
      {
        id: 2,
        name: 'Variant Product Blue',
        sku: 'BLUE-001',
        price: 109,
        currency: 'USD',
        availability: 'out_of_stock',
        localizedPath: '/shop/variant-product?variant=2',
      },
    ],
  },
  {
    siteOrigin: 'https://example.com',
    localizedPath: '/shop/variant-product',
  },
)

assert.equal(groupSeo.schema?.['@type'], 'ProductGroup')
if (groupSeo.schema?.['@type'] !== 'ProductGroup') {
  throw new Error('Expected a ProductGroup schema for multiple variants.')
}
assert.equal(groupSeo.schema.productGroupID, 'product-42')
assert.equal(groupSeo.schema.hasVariant.length, 2)
assert.deepEqual(groupSeo.schema.image, ['https://example.com/media/variant.jpg'])
assert.deepEqual(groupSeo.schema.hasVariant[0]?.image, ['https://example.com/media/variant.jpg'])
assert.equal(groupSeo.schema.hasVariant[0]?.offers?.url, 'https://example.com/shop/variant-product?variant=1')
assert.equal(groupSeo.schema.hasVariant[1]?.offers?.availability, 'https://schema.org/OutOfStock')
assert.equal(groupSeo.schema.hasVariant[0]?.offers?.shippingDetails, undefined)

const incompleteVariantGroupSeo = buildProductSeoDocument(
  {
    name: 'Partially Sellable Variant Product',
    imageUrls: ['/media/partial.jpg'],
    offer: {
      price: 150,
      currency: 'USD',
      availability: 'in_stock',
    },
    variants: [
      {
        id: 1,
        name: 'Complete Variant',
        sku: 'COMPLETE-001',
        price: 99,
        currency: 'USD',
        availability: 'in_stock',
        localizedPath: '/shop/partial?variant=1',
      },
      {
        id: 2,
        name: 'Missing Price Variant',
        sku: 'MISSING-PRICE-001',
        price: 0,
        currency: 'USD',
        availability: 'in_stock',
        localizedPath: '/shop/partial?variant=2',
      },
    ],
  },
  {
    siteOrigin: 'https://example.com',
    localizedPath: '/shop/partial',
  },
)
assert.equal(incompleteVariantGroupSeo.schema?.['@type'], 'Product')
if (incompleteVariantGroupSeo.schema?.['@type'] !== 'Product') {
  throw new Error('Expected fallback Product schema when fewer than two variants have complete Offers.')
}
assert.equal(incompleteVariantGroupSeo.schema.offers?.url, 'https://example.com/shop/partial')

assert.equal(
  buildProductSeoDocument(
    { name: 'No Image Product', offer: { price: 10, currency: 'USD', availability: 'in_stock' } },
    { siteOrigin: 'https://example.com', localizedPath: '/shop/no-image-product' },
  ).schema,
  null,
)

const noOfferSeo = buildProductSeoDocument(
  { name: 'No Offer Product', imageUrls: ['/media/no-offer.jpg'] },
  { siteOrigin: 'https://example.com', localizedPath: '/shop/no-offer-product' },
)
assert.equal(noOfferSeo.schema?.['@type'], 'Product')
if (noOfferSeo.schema?.['@type'] !== 'Product') {
  throw new Error('Expected Product schema without an incomplete Offer.')
}
assert.equal(noOfferSeo.schema.offers, undefined)

const noReviewSchema = buildProductSeoDocument(
  {
    name: 'No Review Product',
    imageUrls: ['/media/no-review.jpg'],
    aggregateRating: { ratingValue: 0, reviewCount: 0 },
  },
  {
    siteOrigin: 'https://example.com',
    localizedPath: '/shop/no-review-product',
  },
)
assert.equal(noReviewSchema.schema?.['@type'], 'Product')
assert.equal(noReviewSchema.schema?.aggregateRating, undefined)

const dangerousValue = '</script><script>alert(1)</script>'
const script = createSeoJsonLdScript({ name: dangerousValue })
assert.equal(script.type, 'application/ld+json')
assert.ok(script.textContent)
assert.equal(JSON.parse(String(script.textContent)).name, dangerousValue)
assert.equal(String(script.textContent).includes('</script>'), false)

assert.equal(
  buildProductSeoDocument(
    { name: '', imageUrls: ['/media/empty-name.jpg'], offer: { price: 10, currency: 'USD', availability: 'in_stock' } },
    { siteOrigin: 'https://example.com', localizedPath: '/shop/empty' },
  ).schema,
  null,
)

console.log('SEO product output checks passed.')
