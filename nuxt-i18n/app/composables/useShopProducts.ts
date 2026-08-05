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
  prices: {
    regular: number
    sale: number
  }
  availability: ShopProductAvailability
  productType?: ShopProductType | null
  variants: ShopProductVariant[]
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
  isVariantOption?: boolean
  sortOrder?: number
}

export interface ShopProductVariant {
  id: number
  sku?: string
  title: string
  optionValues: Record<string, string>
  priceNumber: number
  price?: number
  salePriceNumber: number | null
  sale_price?: number | null
  availability: ShopProductAvailability
  weightGrams?: number
  isDefault: boolean
}

export interface ShopProductsResult {
  items: ShopProduct[]
  raw: unknown
}

export type ShopProductQueryParams = Record<string, string | number | boolean | undefined>

export interface ShopProductCartOptions {
  variantId?: number | null
  price?: number
  salePrice?: number | null
  sku?: string
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

const formatPriceLabel = (amount: number) => (amount > 0 ? `$${amount}` : '')

const normalizeAvailability = (value: unknown): ShopProductAvailability => {
  return value === 'out_of_stock' ? 'out_of_stock' : 'in_stock'
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

const normalizeVariant = (variant: any): ShopProductVariant | null => {
  const id = toFiniteNumber(variant?.id)
  const sku = String(variant?.sku || '').trim()
  if (!id) return null

  const regular = toFiniteNumber(variant?.price)
  const sale = toOptionalNumber(variant?.sale_price)
  const salePriceNumber = sale && sale > 0 ? sale : null
  const priceNumber = salePriceNumber ?? regular
  const optionValues = parseOptionValues(variant?.option_values)
  const title = String(variant?.title || Object.values(optionValues).join(' / ') || sku || `Option ${id}`).trim()
  const weightGrams = toOptionalPositiveNumber(variant?.weight_grams)

  return {
    id,
    ...(sku ? { sku } : {}),
    title,
    optionValues,
    priceNumber,
    price: regular,
    salePriceNumber,
    sale_price: salePriceNumber,
    availability: normalizeAvailability(variant?.availability),
    ...(weightGrams ? { weightGrams } : {}),
    isDefault: Boolean(variant?.is_default),
  }
}

export const normalizeShopProduct = (item: any): ShopProduct => {
  const id = toFiniteNumber(item?.id)
  const variants = Array.isArray(item?.variants) ? item.variants.map(normalizeVariant).filter(Boolean) as ShopProductVariant[] : []
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
    priceLabel: formatPriceLabel(priceNumber),
    prices: {
      regular,
      sale,
    },
    availability: normalizeAvailability(item?.availability),
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

export function useShopProducts() {
  const config = useRuntimeConfig()
  const { locale } = useI18n()
  const baseURL = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
  const productRequestHeaders = () => {
    const currentLocale = String(locale.value || '').trim()
    return currentLocale ? { 'Accept-Language': currentLocale } : undefined
  }

  // Legacy search source kept for existing shop and customer-service flows.
  // Recommendation baseline data must use fetchPublicShopProducts instead.
  const fetchShopProducts = async (params: ShopProductQueryParams): Promise<ShopProductsResult> => {
    const response = await $fetch<any>(`${baseURL}/customer-service/products`, {
      params,
      headers: productRequestHeaders(),
    })
    const items = extractProductItems(response).map(normalizeShopProduct)

    return {
      items,
      raw: response,
    }
  }

  const fetchPublicShopProducts = async (params: ShopProductQueryParams): Promise<ShopProductsResult> => {
    const response = await $fetch<any>(`${baseURL}/products`, {
      headers: productRequestHeaders(),
      params: {
        status: 'active',
        page_size: 12,
        ...params,
      },
    })
    const items = extractProductItems(response).map(normalizeShopProduct)

    return {
      items,
      raw: response,
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
