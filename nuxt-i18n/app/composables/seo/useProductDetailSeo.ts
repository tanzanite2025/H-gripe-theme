import { computed, type Ref } from 'vue'
import {
  useHead,
  useLocalePath,
  useRoute,
  useRequestURL,
  useRuntimeConfig,
} from '#imports'
import type {
  GoProduct,
  ProductBreadcrumbItem,
  ProductMediaImage,
  ProductAvailability,
  ProductVariant,
  SpecDefinition,
} from '~/types/productDetail'
import {
  displayPriceSnapshotForCurrency,
  normalizeProductCurrencyCode,
  validProductDisplayPrice,
} from '~/utils/productDetail'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import {
  buildProductPath,
  toAbsoluteSeoUrl,
} from '~/utils/seo/urls'
import {
  buildProductSeoDocument,
  resolveProductMetaDescription,
  resolveProductMetaTitle,
} from '~/utils/seo/product'
import {
  type StorefrontSeoAlternateLinkEntry,
  useStorefrontSeoRouteOverride,
} from '~/composables/seo/useStorefrontSeoLinks'
import localeRegistry from '~/i18n/locales.manifest'

export interface ProductDetailSeoOptions {
  product: Ref<GoProduct | null | undefined>
  slug: Ref<string>
  productImages: Ref<ProductMediaImage[]>
  activeVariants: Ref<ProductVariant[]>
  selectedVariant: Ref<ProductVariant | null>
  selectedAvailability: Ref<ProductAvailability>
  currentDisplayPrice: Ref<{ amount: number; currency: string }>
  variantOptionDefinitions: Ref<SpecDefinition[]>
  variantLabel: (variant: ProductVariant) => string
  displayCurrency: Ref<string>
  breadcrumbItems: Ref<ProductBreadcrumbItem[]>
}

const storefrontLocaleCodes = new Set(
  localeRegistry.map((locale) => String(locale.code || '').trim()).filter(Boolean),
)

const normalizeBreadcrumbPath = (value: unknown) => {
  const raw = String(value || '').trim()
  if (!raw || !raw.startsWith('/') || /[?#]/.test(raw)) return ''
  if (raw === '/') return '/'
  return `/${raw.replace(/^\/+|\/+$/g, '')}`
}

const stripBreadcrumbLocale = (path: string) => {
  const normalized = normalizeBreadcrumbPath(path)
  if (!normalized || normalized === '/') return normalized

  const segments = normalized.split('/').filter(Boolean)
  const firstSegment = segments[0] || ''
  if (firstSegment && storefrontLocaleCodes.has(firstSegment)) {
    const remainder = segments.slice(1)
    return remainder.length ? `/${remainder.join('/')}` : '/'
  }
  return normalized
}

const isValidBreadcrumbItemPath = (type: string, path: string) => {
  const barePath = stripBreadcrumbLocale(path)
  if (!barePath) return false

  const segments = barePath.split('/').filter(Boolean)
  switch (type) {
    case 'home':
      return barePath === '/'
    case 'shop':
      return barePath === '/shop'
    case 'category':
      return segments.length >= 2 && segments[0] === 'shop'
    case 'product':
      return segments.length === 2 && segments[0] === 'products'
    default:
      return false
  }
}

export function useProductDetailSeo(options: ProductDetailSeoOptions) {
  const config = useRuntimeConfig()
  const requestUrl = useRequestURL()
  const route = useRoute()
  const localePath = useLocalePath()
  const siteOrigin = computed(() => {
    const value = (config.public as { siteUrl?: string }).siteUrl
    if (value && value.trim().length) {
      return value.replace(/\/$/, '')
    }
    return requestUrl.origin.replace(/\/$/, '')
  })

  const seoMediaOrigins = [
    ...(import.meta.server
      ? [
          String((config as { apiInternalOrigin?: string }).apiInternalOrigin || ''),
          String((config as { imageInternalOrigin?: string }).imageInternalOrigin || ''),
        ]
      : []),
  ]

  const localizedProductPath = computed(() => localePath(
    buildProductPath(options.product.value?.slug || options.slug.value),
  ))

  const localizedProductSeoRoutes = computed<StorefrontSeoAlternateLinkEntry[] | null>(() => {
    if (!options.product.value || !Array.isArray(options.product.value.localized_routes)) {
      return null
    }

    const routes = options.product.value.localized_routes
      .map((entry) => {
        const code = String(entry?.locale || '').trim()
        const translatedSlug = String(entry?.slug || '').trim()
        if (!code || !translatedSlug) return null
        return {
          code,
          path: localePath(buildProductPath(translatedSlug), code as any),
        }
      })
      .filter((entry): entry is StorefrontSeoAlternateLinkEntry => Boolean(entry))

    return routes.length ? routes : null
  })

  useStorefrontSeoRouteOverride(localizedProductSeoRoutes)

  const rootCanonicalUrl = computed(() => toAbsoluteSeoUrl(
    siteOrigin.value,
    localizedProductPath.value,
  ))

  // Variant selections are addressable via ?variant= and must not emit a
  // canonical link to the unselected product URL, otherwise search engines
  // collapse every variant into one duplicate document.
  const canonicalUrl = computed(() => {
    const variantID = options.selectedVariant.value?.id
    const requestedVariant = typeof route.query.variant === 'string' ? route.query.variant.trim() : ''
    const path = variantID && requestedVariant === String(variantID)
      ? `${localizedProductPath.value}?variant=${encodeURIComponent(String(variantID))}`
      : localizedProductPath.value
    return toAbsoluteSeoUrl(siteOrigin.value, path)
  })

  const metaTitle = computed(() => resolveProductMetaTitle(
    options.product.value?.meta_title,
    options.product.value?.name,
  ))

  const metaDescription = computed(() => resolveProductMetaDescription({
    metaDescription: options.product.value?.meta_description,
    shortDescription: options.product.value?.short_description,
    description: options.product.value?.description,
  }))

  const productSeoDocument = computed(() => {
    const product = options.product.value
    if (!product) return null

    const variantSeoPath = (variantId: number) => (
      `${localizedProductPath.value}?variant=${encodeURIComponent(String(variantId))}`
    )
    const variantSeoPrice = (variant: ProductVariant) => {
      const displayPrice = validProductDisplayPrice(variant.display_price)
        || displayPriceSnapshotForCurrency(variant.display_prices, options.displayCurrency.value)
      return {
        amount: displayPrice?.amount_decimal ?? String(variant.sale_price_decimal ?? variant.price_decimal ?? ''),
        currency: displayPrice?.currency
          || normalizeProductCurrencyCode(variant.currency || product.currency)
          || 'USD',
      }
    }

    const productShippingDetails = product.shipping_details
    const seoVariants = options.activeVariants.value.map((variant) => {
      const price = variantSeoPrice(variant)
      return {
        id: variant.id,
        name: options.variantLabel(variant),
        sku: variant.sku,
        price: price.amount,
        currency: price.currency,
        availability: variant.availability,
        localizedPath: variantSeoPath(variant.id),
        imageUrls: options.productImages.value.map((image) => image.url),
        shippingDetails: options.activeVariants.value.length === 1 && productShippingDetails
          ? {
              country: productShippingDetails.country,
              amount: productShippingDetails.amount,
              currency: productShippingDetails.currency,
              freeShipping: productShippingDetails.free_shipping,
              etaMinDays: productShippingDetails.eta_min_days,
              etaMaxDays: productShippingDetails.eta_max_days,
            }
          : null,
      }
    })

    return buildProductSeoDocument(
      {
        name: product.name,
        brand: product.brand?.name,
        metaTitle: product.meta_title,
        metaDescription: product.meta_description,
        shortDescription: product.short_description,
        description: product.description,
        sku: product.sku,
        imageUrls: [
          product.thumbnail,
          ...options.productImages.value.map((image) => image.url),
        ],
        offer: {
          price: String(options.currentDisplayPrice.value.amount),
          currency: options.currentDisplayPrice.value.currency,
          availability: options.selectedAvailability.value,
          sku: options.selectedVariant.value?.sku || product.sku,
        },
        aggregateRating: product.review_summary
          ? {
              ratingValue: product.review_summary.average_rating,
              reviewCount: product.review_summary.total_reviews,
            }
          : null,
        shippingDetails: product.shipping_details
          ? {
              country: product.shipping_details.country,
              amount: product.shipping_details.amount,
              currency: product.shipping_details.currency,
              freeShipping: product.shipping_details.free_shipping,
              etaMinDays: product.shipping_details.eta_min_days,
              etaMaxDays: product.shipping_details.eta_max_days,
            }
          : null,
        productGroupId: `product-${product.id}`,
        variesBy: options.variantOptionDefinitions.value.map((definition) => (
          `https://schema.org/${String(definition.slug || '').trim()}`
        )),
        variants: seoVariants,
      },
      {
        siteOrigin: siteOrigin.value,
        localizedPath: localizedProductPath.value,
        mediaOrigins: seoMediaOrigins,
      },
    )
  })

  const breadcrumbJsonLd = computed(() => {
    const itemListElement = options.breadcrumbItems.value
      .map((item, index) => {
        const name = String(item.name || '').trim()
        const path = normalizeBreadcrumbPath(item.path)
        const type = String(item.type || '').trim().toLowerCase()
        if (!name || !path || !isValidBreadcrumbItemPath(type, path)) return null

        const absolutePath = toAbsoluteSeoUrl(siteOrigin.value, path)
        if (!/^https?:\/\//i.test(absolutePath)) return null
        return {
          '@type': 'ListItem',
          position: index + 1,
          name,
          item: absolutePath,
          type,
        }
      })
      .filter((item): item is {
        '@type': string
        position: number
        name: string
        item: string
        type: string
      } => Boolean(item))

    const categoryCount = itemListElement.filter((item) => item.type === 'category').length
    const finalProductItem = itemListElement[itemListElement.length - 1]
    if (
      itemListElement.length < 2
      || itemListElement[0]?.type !== 'home'
      || itemListElement[1]?.type !== 'shop'
      || itemListElement[itemListElement.length - 1]?.type !== 'product'
      || categoryCount < 1
      || finalProductItem?.item !== rootCanonicalUrl.value
    ) {
      return null
    }

    return {
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      itemListElement: itemListElement.map(({ type: _type, ...item }) => item),
    }
  })

  useHead(() => {
    const seo = productSeoDocument.value
    const seoTitle = seo?.title || metaTitle.value
    const seoDescription = seo?.description || metaDescription.value
    const seoCanonicalUrl = canonicalUrl.value || seo?.canonicalUrl || rootCanonicalUrl.value
    const metaEntries = [
      { name: 'description', content: seoDescription },
      { property: 'og:title', content: seoTitle },
      { property: 'og:description', content: seoDescription },
      { property: 'og:type', content: 'product' },
      { property: 'og:url', content: seoCanonicalUrl },
      { name: 'twitter:card', content: 'summary_large_image' },
      { name: 'twitter:title', content: seoTitle },
      { name: 'twitter:description', content: seoDescription },
    ]

    const seoImage = seo?.images[0] || ''
    if (seoImage) {
      metaEntries.push({ property: 'og:image', content: seoImage })
      metaEntries.push({ name: 'twitter:image', content: seoImage })
    }

    const seoOffer = seo?.schema?.['@type'] === 'Product'
      ? seo.schema.offers
      : options.selectedVariant.value
        ? {
            price: String(options.currentDisplayPrice.value.amount),
            priceCurrency: options.currentDisplayPrice.value.currency,
          }
        : seo?.schema?.['@type'] === 'ProductGroup'
          ? seo.schema.hasVariant.find((variant) => variant.offers)?.offers
          : null
    if (seoOffer) {
      metaEntries.push({
        property: 'product:price:amount',
        content: String(seoOffer.price),
      })
      metaEntries.push({
        property: 'product:price:currency',
        content: seoOffer.priceCurrency,
      })
    }

    const jsonLdScripts = [
      seo?.schema ? createSeoJsonLdScript(seo.schema) : null,
      breadcrumbJsonLd.value
        ? createSeoJsonLdScript(breadcrumbJsonLd.value)
        : null,
    ].filter((script): script is NonNullable<typeof script> => Boolean(script))

    return {
      title: seoTitle,
      meta: metaEntries.filter((entry) => Object.values(entry).every((value) => {
        if (typeof value !== 'string') return true
        return value.trim().length > 0
      })),
      script: jsonLdScripts,
    }
  })

  return {
    localizedProductPath,
    canonicalUrl,
    productSeoDocument,
    metaTitle,
    metaDescription,
  }
}
