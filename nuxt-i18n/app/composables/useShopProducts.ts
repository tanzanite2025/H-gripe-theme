import { useRuntimeConfig } from '#imports'
import type { CartItem } from '~/types/cart'

export interface ShopProduct {
  id: number
  productId: number
  defaultVariantId: number | null
  title: string
  description?: string
  slug: string
  url: string
  thumbnail?: string
  priceNumber: number
  priceLabel: string
  prices: {
    regular: number
    sale: number
  }
  stockQuantity: number | null
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
  stockQuantity?: number | null
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

export const normalizeShopProduct = (item: any): ShopProduct => {
  const id = toFiniteNumber(item?.id)
  const variants = Array.isArray(item?.variants) ? item.variants : []
  const defaultVariant = variants.find((variant: any) => variant?.is_default) || variants[0] || null
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
  const images = Array.isArray(item?.images) ? item.images : []
  const thumbnail = item?.thumbnail || item?.featured_image || images[0]?.url || undefined
  const stock =
    typeof item?.stock === 'object'
      ? toOptionalNumber(item?.stock?.quantity)
      : toOptionalNumber(item?.stock ?? defaultVariant?.stock)

  return {
    id,
    productId: id,
    defaultVariantId: toOptionalPositiveNumber(item?.default_variant_id ?? defaultVariant?.id),
    title: String(item?.title || item?.name || ''),
    description: item?.excerpt || item?.short_description || item?.description || undefined,
    slug,
    url: `/shop/${slug}`,
    thumbnail,
    priceNumber,
    priceLabel: formatPriceLabel(priceNumber),
    prices: {
      regular,
      sale,
    },
    stockQuantity: stock,
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
  const baseURL = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')

  const fetchShopProducts = async (params: ShopProductQueryParams): Promise<ShopProductsResult> => {
    const response = await $fetch<any>(`${baseURL}/customer-service/products`, { params })
    const items = extractProductItems(response).map(normalizeShopProduct)

    return {
      items,
      raw: response,
    }
  }

  const fetchFeaturedShopProducts = async (
    params: ShopProductQueryParams = {}
  ): Promise<ShopProductsResult> => {
    const response = await $fetch<any>(`${baseURL}/products`, {
      params: {
        status: 'active',
        featured: true,
        page_size: 4,
        ...params,
      },
    })
    const items = extractProductItems(response).map(normalizeShopProduct)

    return {
      items,
      raw: response,
    }
  }

  const toCartItem = (
    product: ShopProduct,
    options: ShopProductCartOptions = {}
  ): Omit<CartItem, 'quantity'> => {
    const variantId =
      options.variantId === undefined ? product.defaultVariantId : options.variantId
    const price = options.price ?? product.priceNumber
    const salePrice =
      options.salePrice === undefined
        ? product.prices.sale > 0 ? product.prices.sale : null
        : options.salePrice
    const thumbnail = options.thumbnail ?? product.thumbnail
    const stockQuantity =
      options.stockQuantity === undefined ? product.stockQuantity : options.stockQuantity
    const title = options.title ?? product.title

    return {
      id: variantId || product.id,
      product_id: product.productId,
      variant_id: variantId,
      title,
      name: title,
      slug: product.slug,
      sku: options.sku,
      price,
      sale_price: salePrice,
      image: thumbnail,
      thumbnail,
      maxStock: stockQuantity ?? undefined,
      stock: stockQuantity ?? undefined,
    }
  }

  return {
    baseURL,
    fetchFeaturedShopProducts,
    fetchShopProducts,
    toCartItem,
  }
}
