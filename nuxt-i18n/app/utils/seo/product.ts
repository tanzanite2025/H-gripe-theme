import { toAbsoluteSeoUrl } from './urls'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'
import type {
  ProductSeoContext,
  ProductSeoDocument,
  ProductSeoInput,
  ProductSeoAggregateRating,
  ProductSeoProductGroupSchema,
  ProductSeoSchema,
  ProductSeoShippingDetails,
  ProductSeoStructuredData,
  ProductSeoVariantInput,
} from './types'

const stripHtml = (value: string | null | undefined): string => {
  if (!value) return ''
  return value.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
}

const cleanText = (value: string | null | undefined): string => String(value || '').trim()

const truncateText = (value: string, maxLength: number): string => {
  if (value.length <= maxLength) return value
  return `${value.slice(0, Math.max(0, maxLength - 3))}...`
}

const normalizeCurrency = (value: string | null | undefined): string => {
  const code = cleanText(value).toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

const normalizeImages = (
  imageUrls: Array<string | null | undefined> | null | undefined,
  context: ProductSeoContext,
): string[] => {
  const mediaContext = createStorefrontMediaContext({
    public: { siteUrl: context.siteOrigin },
    additionalOrigins: context.mediaOrigins,
  })
  const images = (imageUrls || [])
    .map((image) => cleanText(image))
    .filter(Boolean)
    .map((image) => normalizeStorefrontMediaUrl(image, mediaContext))
    .map((image) => toAbsoluteSeoUrl(context.siteOrigin, image))
    .filter((image) => /^https?:\/\//i.test(image))

  return [...new Set(images)]
}

const buildAggregateRating = (
  input: ProductSeoInput['aggregateRating'],
): ProductSeoAggregateRating | undefined => {
  const ratingValue = Number(input?.ratingValue)
  const reviewCount = Number(input?.reviewCount)
  if (
    !Number.isFinite(ratingValue)
    || ratingValue < 1
    || ratingValue > 5
    || !Number.isInteger(reviewCount)
    || reviewCount <= 0
  ) {
    return undefined
  }

  return {
    '@type': 'AggregateRating',
    ratingValue: Math.round(ratingValue * 100) / 100,
    reviewCount,
    ratingCount: reviewCount,
    bestRating: 5,
    worstRating: 1,
  }
}

const buildShippingDetails = (
  input: ProductSeoInput['shippingDetails'],
): ProductSeoShippingDetails | undefined => {
  const country = cleanText(input?.country).toUpperCase()
  const currency = normalizeCurrency(input?.currency)
  const amount = Number(input?.amount)
  const etaMinDays = Number(input?.etaMinDays)
  const etaMaxDays = Number(input?.etaMaxDays)
  if (
    !/^[A-Z]{2}$/.test(country)
    || !currency
    || !Number.isFinite(amount)
    || amount < 0
    || !Number.isInteger(etaMinDays)
    || !Number.isInteger(etaMaxDays)
    || etaMinDays < 1
    || etaMaxDays < etaMinDays
  ) {
    return undefined
  }

  return {
    '@type': 'OfferShippingDetails',
    shippingRate: {
      '@type': 'MonetaryAmount',
      value: amount,
      currency,
    },
    shippingDestination: {
      '@type': 'DefinedRegion',
      addressCountry: country,
    },
    deliveryTime: {
      '@type': 'ShippingDeliveryTime',
      transitTime: {
        '@type': 'QuantitativeValue',
        minValue: etaMinDays,
        maxValue: etaMaxDays,
        unitCode: 'DAY',
      },
    },
  }
}

export const resolveProductMetaTitle = (
  metaTitle: string | null | undefined,
  productName: string | null | undefined,
): string => cleanText(metaTitle) || cleanText(productName) || 'Product'

export const resolveProductMetaDescription = (input: ProductSeoInput): string => {
  const explicitDescription = stripHtml(input.metaDescription)
  if (explicitDescription) return truncateText(explicitDescription, 160)

  const fallback = stripHtml(input.shortDescription || input.description)
  return truncateText(fallback, 160)
}

export const buildProductJsonLd = (
  input: ProductSeoInput,
  context: ProductSeoContext,
): ProductSeoSchema | null => {
  const productName = cleanText(input.name)
  if (!productName) return null

  const canonicalUrl = toAbsoluteSeoUrl(context.siteOrigin, context.localizedPath)
  const description = resolveProductMetaDescription(input)
  const images = normalizeImages(input.imageUrls, context)
  if (!images.length) return null

  const productSku = cleanText(input.offer?.sku) || cleanText(input.sku)
  const price = Number(input.offer?.price)
  const currency = normalizeCurrency(input.offer?.currency)
  const availability = input.offer?.availability
  const aggregateRating = buildAggregateRating(input.aggregateRating)
  const shippingDetails = buildShippingDetails(input.shippingDetails)

  const schema: ProductSeoSchema = {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: productName,
    image: images,
    url: canonicalUrl,
  }

  const brand = cleanText(input.brand)
  if (brand) {
    schema.brand = {
      '@type': 'Brand',
      name: brand,
    }
  }

  if (description) schema.description = description
  if (productSku) schema.sku = productSku

  if (Number.isFinite(price) && price > 0 && currency && availability) {
    schema.offers = {
      '@type': 'Offer',
      price,
      priceCurrency: currency,
      availability: availability === 'in_stock'
        ? 'https://schema.org/InStock'
        : 'https://schema.org/OutOfStock',
      url: canonicalUrl,
    }
    if (shippingDetails) {
      schema.offers.shippingDetails = shippingDetails
    }
  }
  if (aggregateRating) schema.aggregateRating = aggregateRating

  return schema
}

const variantOfferInput = (variant: ProductSeoVariantInput) => ({
  price: variant.price,
  currency: variant.currency,
  availability: variant.availability,
  sku: variant.sku,
  shippingDetails: variant.shippingDetails,
})

export const buildProductGroupJsonLd = (
  input: ProductSeoInput,
  context: ProductSeoContext,
): ProductSeoStructuredData | null => {
  const productName = cleanText(input.name)
  const variants = (input.variants || []).filter((variant) => (
    cleanText(variant.name || variant.sku)
    && cleanText(variant.localizedPath)
  ))
  if (!productName || variants.length < 2) {
    return buildProductJsonLd(input, context)
  }

  const description = resolveProductMetaDescription(input)
  const images = normalizeImages(input.imageUrls, context)
  const group: ProductSeoProductGroupSchema = {
    '@context': 'https://schema.org',
    '@type': 'ProductGroup',
    name: productName,
    url: toAbsoluteSeoUrl(context.siteOrigin, context.localizedPath),
    productGroupID: cleanText(input.productGroupId) || productName,
    image: images,
    hasVariant: [],
  }

  const brand = cleanText(input.brand)
  if (brand) {
    group.brand = {
      '@type': 'Brand',
      name: brand,
    }
  }
  if (description) group.description = description
  const aggregateRating = buildAggregateRating(input.aggregateRating)
  if (aggregateRating) group.aggregateRating = aggregateRating
  const variesBy = (input.variesBy || []).map(cleanText).filter(Boolean)
  if (variesBy.length) group.variesBy = [...new Set(variesBy)]

  const variantUrls = new Set<string>()
  group.hasVariant = variants
    .map((variant) => {
      const variantName = cleanText(variant.name) || `${productName} ${cleanText(variant.sku)}`
      const variantSchema = buildProductJsonLd(
        {
          ...input,
          name: variantName,
          sku: variant.sku,
          imageUrls: variant.imageUrls?.length ? variant.imageUrls : input.imageUrls,
          offer: variantOfferInput(variant),
          shippingDetails: variant.shippingDetails,
          variants: null,
        },
        {
          ...context,
          localizedPath: cleanText(variant.localizedPath),
        },
      )
      return variantSchema
    })
    .filter((variant): variant is ProductSeoSchema => {
      if (!variant?.offers || !variant.url || variantUrls.has(variant.url)) return false
      variantUrls.add(variant.url)
      return true
    })

  if (!group.image.length) {
    const firstVariantImages = group.hasVariant.find((variant) => variant.image?.length)?.image
    if (firstVariantImages?.length) group.image = firstVariantImages
  }

  return group.hasVariant.length >= 2 && group.image.length ? group : buildProductJsonLd(input, context)
}

export const buildProductSeoDocument = (
  input: ProductSeoInput,
  context: ProductSeoContext,
): ProductSeoDocument => {
  const title = resolveProductMetaTitle(input.metaTitle, input.name)
  const description = resolveProductMetaDescription(input)
  const canonicalUrl = toAbsoluteSeoUrl(context.siteOrigin, context.localizedPath)
  const images = normalizeImages(input.imageUrls, context)
  const schema = buildProductGroupJsonLd(input, context)

  return {
    title,
    description,
    canonicalUrl,
    images,
    schema,
  }
}
