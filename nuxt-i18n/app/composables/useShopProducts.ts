import { useRuntimeConfig } from '#imports'
import { useI18n } from 'vue-i18n'
import type { CartItem } from '~~/types/cart'

export type ShopProductAvailability = 'in_stock' | 'out_of_stock'

export interface ShopProduct {
  id: number
  productId: number
  defaultVariantId: number | null
  title: string
  description?: string
  slug: string
  url: string
  sku?: string
  thumbnail?: string
  priceNumber: number
  priceLabel: string
  currency: string
  displayPriceNumber: number
  displayPriceCurrency: string
  displayPriceLabel: string
  displayPrices: ShopProductDisplayPrice[]
  prices: {
    regular: number
    sale: number
  }
  availability: ShopProductAvailability
  brand?: ShopProductBrand | null
  productType?: ShopProductType | null
  variants: ShopProductVariant[]
}

export interface ShopProductBrand {
  id?: number
  name: string
  slug: string
  logoUrl?: string
  websiteUrl?: string
}

export interface ShopProductDisplayPrice {
  amount: number
  currency: string
  rate?: number
  source?: string
  converted?: boolean
  fallback_reason?: string
}

export interface ShopProductType {
  id?: number
  name: string
  slug: string
  specDefinitions: ShopProductSpecDefinition[]
}

export interface ShopProductSpecDefinition {
  id?: number
  name: string
  slug: string
  group?: string
  fieldType?: string
  unit?: string
  isFilterable?: boolean
  isVariantOption?: boolean
  sortOrder?: number
}

export interface ShopProductVariant {
  id: number
  sku?: string
  title: string
  optionValues: Record<string, string>
  priceNumber: number
  currency: string
  price?: number
  salePriceNumber: number | null
  sale_price?: number | null
  displayPrices: ShopProductDisplayPrice[]
  availability: ShopProductAvailability
  weightGrams?: number
  isDefault: boolean
}

export interface ShopProductsResult {
  items: ShopProduct[]
  raw: unknown
  page?: number
  pageSize?: number
  total?: number
  hasMore?: boolean
  quickBuyFilters?: import('~/utils/quickBuy/types').QuickBuySpecFilter[]
}

export type ShopProductQueryParams = Record<string, string | number | boolean | undefined>

export interface ShopProductCartOptions {
  variantId?: number | null
  price?: number
  salePrice?: number | null
  sku?: string
  currency?: string
  title?: string
  thumbnail?: string
  weightGrams?: number | null
}

const toFiniteNumber = (value: unknown, fallback = 0) => {
  if (value === null || value === undefined || value === '') {
    return fallback
  }
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : fallback
}

const toOptionalNumber = (value: unknown) => {
  if (value === null || value === undefined || value === '') {
    return null
  }
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : null
}

const toOptionalPositiveNumber = (value: unknown) => {
  const numberValue = toOptionalNumber(value)
  return numberValue && numberValue > 0 ? numberValue : null
}

const normalizeCurrencyCode = (value: unknown) => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

const formatPriceLabel = (amount: number, currency = 'USD') => {
  if (amount <= 0) return ''
  const normalizedCurrency = normalizeCurrencyCode(currency) || 'USD'
  try {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: normalizedCurrency }).format(amount)
  } catch {
    return `${normalizedCurrency} ${amount}`
  }
}

const normalizeAvailability = (value: unknown): ShopProductAvailability => {
  return value === 'out_of_stock' ? 'out_of_stock' : 'in_stock'
}

const normalizeDisplayPrice = (value: any, fallbackAmount: number, fallbackCurrency: string): ShopProductDisplayPrice => {
  const amount = toFiniteNumber(value?.amount, fallbackAmount)
  const currency = normalizeCurrencyCode(value?.currency) || fallbackCurrency
  return {
    amount,
    currency,
    rate: toFiniteNumber(value?.rate, 1),
    source: value?.source ? String(value.source) : undefined,
    converted: Boolean(value?.converted),
    fallback_reason: value?.fallback_reason ? String(value.fallback_reason) : undefined,
  }
}

const normalizeDisplayPrices = (value: any, fallbackCurrency: string): ShopProductDisplayPrice[] => {
  const list = Array.isArray(value) ? value : []
  const seen = new Set<string>()
  return list
    .map((item: any) => normalizeDisplayPrice(item, 0, fallbackCurrency))
    .filter(item => item.amount > 0 && normalizeCurrencyCode(item.currency))
    .filter((item) => {
      if (seen.has(item.currency)) return false
      seen.add(item.currency)
      return true
    })
}

const parseOptionValues = (value: unknown): Record<string, string> => {
  if (!value) return {}
  if (typeof value === 'object' && !Array.isArray(value)) {
    return Object.entries(value as Record<string, unknown>).reduce<Record<string, string>>((acc, [key, raw]) => {
      const stringValue = String(raw ?? '').trim()
      if (key && stringValue) {
        acc[key] = stringValue
      }
      return acc
    }, {})
  }
  if (typeof value !== 'string') return {}
  try {
    const parsed = JSON.parse(value)
    return parseOptionValues(parsed)
  } catch {
    return {}
  }
}

const normalizeSpecDefinitions = (item: any): ShopProductSpecDefinition[] => {
  const definitions: any[] = Array.isArray(item?.product_type?.spec_definitions)
    ? item.product_type.spec_definitions
    : []

  return definitions
    .map((definition: any): ShopProductSpecDefinition | null => {
      const id = toOptionalPositiveNumber(definition?.id)
      const slug = String(definition?.slug || '').trim()
      const name = String(definition?.name || slug || '').trim()
      if (!slug || !name) return null

      return {
        ...(id ? { id } : {}),
        name,
        slug,
        group: definition?.group ? String(definition.group) : undefined,
        fieldType: definition?.field_type ? String(definition.field_type) : undefined,
        unit: definition?.unit ? String(definition.unit) : undefined,
        isFilterable: Boolean(definition?.is_filterable),
        isVariantOption: Boolean(definition?.is_variant_option),
        sortOrder: toFiniteNumber(definition?.sort_order),
      }
    })
    .filter((definition): definition is ShopProductSpecDefinition => Boolean(definition))
}

const normalizeProductType = (item: any): ShopProductType | null => {
  if (!item?.product_type) return null
  const id = toOptionalPositiveNumber(item.product_type.id)
  const slug = String(item.product_type.slug || '').trim()
  const name = String(item.product_type.name || slug || '').trim()
  if (!slug || !name) return null

  return {
    ...(id ? { id } : {}),
    name,
    slug,
    specDefinitions: normalizeSpecDefinitions(item),
  }
}

const normalizeVariant = (variant: any, fallbackCurrency = 'USD'): ShopProductVariant | null => {
  const id = toFiniteNumber(variant?.id)
  const sku = String(variant?.sku || '').trim()
  if (!id) return null

  const regular = toFiniteNumber(variant?.price)
  const sale = toOptionalNumber(variant?.sale_price)
  const salePriceNumber = sale && sale > 0 ? sale : null
  const priceNumber = salePriceNumber ?? regular
  const variantCurrency = normalizeCurrencyCode(variant?.currency) || normalizeCurrencyCode(fallbackCurrency) || 'USD'
  const displayPrices = normalizeDisplayPrices(variant?.display_prices, variantCurrency)
  const optionValues = parseOptionValues(variant?.option_values)
  const title = String(variant?.title || Object.values(optionValues).join(' / ') || sku || `Option ${id}`).trim()
  const weightGrams = toOptionalPositiveNumber(variant?.weight_grams)

  return {
    id,
    ...(sku ? { sku } : {}),
    title,
    optionValues,
    priceNumber,
    currency: variantCurrency,
    price: regular,
    salePriceNumber,
    sale_price: salePriceNumber,
    displayPrices,
    availability: normalizeAvailability(variant?.availability),
    ...(weightGrams ? { weightGrams } : {}),
    isDefault: Boolean(variant?.is_default),
  }
}

export const normalizeShopProduct = (item: any, fallbackCurrency = 'USD'): ShopProduct => {
  const id = toFiniteNumber(item?.id)
  const variants = Array.isArray(item?.variants)
    ? item.variants.map((variant: any) => normalizeVariant(variant, fallbackCurrency)).filter(Boolean) as ShopProductVariant[]
    : []
  const defaultVariant = variants.find((variant) => variant.isDefault) || variants[0] || null
  const regular = toFiniteNumber(
    item?.prices?.regular,
    toFiniteNumber(defaultVariant?.price, toFiniteNumber(item?.price))
  )
  const sale = toFiniteNumber(
    item?.prices?.sale,
    toFiniteNumber(defaultVariant?.sale_price, toFiniteNumber(item?.sale_price))
  )
  const priceNumber = sale > 0 ? sale : regular > 0 ? regular : 0
  const productCurrency = normalizeCurrencyCode(defaultVariant?.currency || item?.currency) || normalizeCurrencyCode(fallbackCurrency) || 'USD'
  const displayPrice = normalizeDisplayPrice(item?.display_price, priceNumber, productCurrency)
  const displayPrices = normalizeDisplayPrices(item?.display_prices, productCurrency)
  const slug = String(item?.slug || id)
  const media = Array.isArray(item?.media) ? item.media : []
  const imageMedia = media.filter((mediaItem: any) => {
    return mediaItem?.media_type === 'image' && mediaItem?.url && mediaItem?.is_visible !== false
  })
  const primaryMediaImage =
    imageMedia.find((mediaItem: any) => mediaItem?.is_primary || mediaItem?.role === 'primary') ||
    imageMedia[0]
  const thumbnail = item?.thumbnail || item?.featured_image || primaryMediaImage?.url || undefined
  return {
    id,
    productId: id,
    defaultVariantId: toOptionalPositiveNumber(item?.default_variant_id ?? defaultVariant?.id),
    title: String(item?.title || item?.name || ''),
    description: item?.excerpt || item?.short_description || item?.description || undefined,
    slug,
    url: String(item?.preview_url || item?.url || `/shop/${slug}`),
    sku: defaultVariant?.sku || item?.sku || undefined,
    thumbnail,
    priceNumber,
    priceLabel: formatPriceLabel(displayPrice.amount, displayPrice.currency),
    currency: productCurrency,
    displayPriceNumber: displayPrice.amount,
    displayPriceCurrency: displayPrice.currency,
    displayPriceLabel: formatPriceLabel(displayPrice.amount, displayPrice.currency),
    displayPrices,
    prices: {
      regular,
      sale,
    },
    availability: normalizeAvailability(item?.availability),
    brand: item?.brand?.name
      ? {
          id: toOptionalPositiveNumber(item.brand.id) || undefined,
          name: String(item.brand.name),
          slug: String(item.brand.slug || ''),
          logoUrl: item.brand.logo_url ? String(item.brand.logo_url) : undefined,
          websiteUrl: item.brand.website_url ? String(item.brand.website_url) : undefined,
        }
      : null,
    productType: normalizeProductType(item),
    variants,
  }
}

const extractProductItems = (response: any): any[] => {
  if (Array.isArray(response?.items)) return response.items
  if (Array.isArray(response?.data)) return response.data
  if (Array.isArray(response)) return response
  return []
}

const extractPagination = (response: any, fallbackPageSize: number): Pick<ShopProductsResult, 'page' | 'pageSize' | 'total' | 'hasMore'> => ({
  page: toFiniteNumber(response?.page, 1),
  pageSize: toFiniteNumber(response?.page_size, fallbackPageSize),
  total: toFiniteNumber(response?.total, 0),
  hasMore: Boolean(response?.has_more),
})

export function useShopProducts() {
  const config = useRuntimeConfig()
  const { locale } = useI18n()
  const { displayCurrency, countryCode, baseCurrency } = useStorefrontContext()
  const baseURL = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
  const productRequestHeaders = () => {
    const currentLocale = String(locale.value || '').trim()
    const headers: Record<string, string> = {}
    if (currentLocale) headers['Accept-Language'] = currentLocale
    if (displayCurrency.value) headers['X-Display-Currency'] = displayCurrency.value
    if (countryCode.value && countryCode.value !== 'ZZ') headers['X-Market-Country'] = countryCode.value
    return Object.keys(headers).length ? headers : undefined
  }

  // Legacy search source kept for existing shop and customer-service flows.
  // Recommendation baseline data must use fetchPublicShopProducts instead.
  const fetchShopProducts = async (params: ShopProductQueryParams): Promise<ShopProductsResult> => {
    const response = await $fetch<any>(`${baseURL}/customer-service/products`, {
      params,
      headers: productRequestHeaders(),
    })
    const items = extractProductItems(response).map((item: any) => normalizeShopProduct(item, baseCurrency.value))

    return {
      items,
      raw: response,
      ...extractPagination(response, toFiniteNumber(params.per_page ?? params.page_size, 12)),
    }
  }

  const fetchPublicShopProducts = async (params: ShopProductQueryParams): Promise<ShopProductsResult> => {
    const response = await $fetch<any>(`${baseURL}/products`, {
      headers: productRequestHeaders(),
      params: {
        status: 'active',
        page_size: 12,
        currency: displayCurrency.value,
        country: countryCode.value !== 'ZZ' ? countryCode.value : undefined,
        ...params,
      },
    })
    const items = extractProductItems(response).map((item: any) => normalizeShopProduct(item, baseCurrency.value))

    return {
      items,
      raw: response,
      ...extractPagination(response, toFiniteNumber(params.page_size ?? params.per_page, 12)),
    }
  }

  const fetchFeaturedShopProducts = async (
    params: ShopProductQueryParams = {}
  ): Promise<ShopProductsResult> => {
    return fetchPublicShopProducts({
      featured: true,
      page_size: 4,
      ...params,
    })
  }

  const toCartItem = (
    product: ShopProduct,
    options: ShopProductCartOptions = {}
  ): Omit<CartItem, 'quantity'> => {
    const variantId =
      options.variantId === undefined ? product.defaultVariantId : options.variantId
    const selectedVariant = variantId
      ? product.variants.find(variant => Number(variant.id) === Number(variantId)) || null
      : product.variants.find(variant => variant.isDefault) || product.variants[0] || null
    const price = options.price ?? product.priceNumber
    const salePrice =
      options.salePrice === undefined
        ? product.prices.sale > 0 ? product.prices.sale : null
        : options.salePrice
    const thumbnail = options.thumbnail ?? product.thumbnail
    const title = options.title ?? product.title
    const sku = options.sku ?? selectedVariant?.sku ?? product.sku
    const currency = normalizeCurrencyCode(options.currency || selectedVariant?.currency || product.currency) || baseCurrency.value || 'USD'
    const weightGrams = options.weightGrams ?? selectedVariant?.weightGrams ?? null

    return {
      id: variantId || product.id,
      product_id: product.productId,
      variant_id: variantId,
      title,
      name: title,
      slug: product.slug,
      sku: sku || undefined,
      price,
      currency,
      sale_price: salePrice,
      image: thumbnail,
      thumbnail,
      weight_grams: weightGrams || undefined,
    }
  }

  return {
    baseURL,
    fetchPublicShopProducts,
    fetchFeaturedShopProducts,
    fetchShopProducts,
    toCartItem,
  }
}
