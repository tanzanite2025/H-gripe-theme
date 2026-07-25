import type {
  ShopProduct,
  ShopProductSpecDefinition,
  ShopProductVariant
} from '~/composables/useShopProducts'

export interface ProductConfigConfirmOption {
  key: string
  label: string
  value: string
  unit?: string
}

export interface ProductConfigConfirmSelection {
  variant_id: number | null
  variant_title: string
  sku: string
  options: ProductConfigConfirmOption[]
  stock: number | null
  weight_grams: number | null
  price: string
  price_value: number
}

export interface ProductConfigConfirmMetadata {
  product: {
    id: number
    variant_id: number | null
    title: string
    slug: string
    sku: string
    url: string
    thumbnail: string
    price: string
    price_value: number
  }
  selections: ProductConfigConfirmSelection
  note: string
}

type ProductLike = Partial<ShopProduct> & Record<string, any>
type VariantLike = Partial<ShopProductVariant> & Record<string, any>

const toFiniteNumber = (value: unknown, fallback = 0) => {
  if (value === null || value === undefined || value === '') return fallback
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : fallback
}

const toNullablePositiveNumber = (value: unknown) => {
  const numberValue = toFiniteNumber(value, 0)
  return numberValue > 0 ? numberValue : null
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
    return parseOptionValues(JSON.parse(value))
  } catch {
    return {}
  }
}

const specDefinitionMap = (product: ProductLike) => {
  const definitions = product.productType?.specDefinitions
    || product.product_type?.spec_definitions
    || []

  return (Array.isArray(definitions) ? definitions : []).reduce<Map<string, ShopProductSpecDefinition>>((acc, raw) => {
    const slug = String(raw?.slug || '').trim()
    const name = String(raw?.name || slug).trim()
    if (!slug || !name) return acc
    acc.set(slug, {
      id: toFiniteNumber(raw?.id),
      name,
      slug,
      group: raw?.group ? String(raw.group) : undefined,
      fieldType: raw?.fieldType || raw?.field_type,
      unit: raw?.unit ? String(raw.unit) : undefined,
      isVariantOption: Boolean(raw?.isVariantOption ?? raw?.is_variant_option),
      sortOrder: toFiniteNumber(raw?.sortOrder ?? raw?.sort_order),
    })
    return acc
  }, new Map())
}

const humanizeKey = (value: string) => {
  return value
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, char => char.toUpperCase())
}

const variantOptionRows = (product: ProductLike, variant: VariantLike | null): ProductConfigConfirmOption[] => {
  const rawOptions = variant?.optionValues || variant?.option_values || {}
  const optionValues = parseOptionValues(rawOptions)
  const definitions = specDefinitionMap(product)

  return Object.entries(optionValues).map(([key, value]) => {
    const definition = definitions.get(key)
    return {
      key,
      label: definition?.name || humanizeKey(key),
      value,
      unit: definition?.unit,
    }
  })
}

export const findProductConfigVariant = (
  product: ProductLike,
  variantId?: number | string | null
): VariantLike | null => {
  const variants = Array.isArray(product?.variants) ? product.variants as VariantLike[] : []
  if (variants.length === 0) return null

  const normalizedVariantId = toNullablePositiveNumber(variantId)
  if (normalizedVariantId) {
    const matched = variants.find(variant => Number(variant?.id) === normalizedVariantId)
    if (matched) return matched
  }

  const defaultVariantId = toNullablePositiveNumber(product?.defaultVariantId ?? product?.default_variant_id)
  if (defaultVariantId) {
    const matchedDefaultId = variants.find(variant => Number(variant?.id) === defaultVariantId)
    if (matchedDefaultId) return matchedDefaultId
  }

  return variants.find(variant => Boolean(variant?.isDefault ?? variant?.is_default)) || variants[0]
}

export const buildProductConfigConfirmMetadata = (
  product: ProductLike,
  variant: VariantLike | null = null
): ProductConfigConfirmMetadata => {
  const selectedVariant = variant || findProductConfigVariant(product)
  const variantId = toNullablePositiveNumber(selectedVariant?.id ?? product?.defaultVariantId ?? product?.default_variant_id)
  const productId = toFiniteNumber(product?.id ?? product?.productId ?? product?.product_id)
  const title = String(product?.title || product?.name || 'Product').trim()
  const slug = String(product?.slug || '').trim()
  const sku = String(selectedVariant?.sku || product?.sku || '').trim()
  const priceValue = toFiniteNumber(
    selectedVariant?.priceNumber
      ?? selectedVariant?.price_value
      ?? selectedVariant?.sale_price
      ?? selectedVariant?.price
      ?? product?.priceValue
      ?? product?.priceNumber
      ?? product?.price_value
      ?? product?.prices?.sale
      ?? product?.prices?.regular
  )
  const price = String(
    selectedVariant?.price
      || product?.price
      || product?.priceLabel
      || (priceValue > 0 ? `$${priceValue}` : '')
  )
  const thumbnail = String(
    product?.thumbnail
      || product?.image
      || product?.featured_image
      || ''
  )
  const url = String(product?.url || (slug ? `/shop/${slug}` : '')).trim()
  const optionRows = variantOptionRows(product, selectedVariant)
  const stock = toNullablePositiveNumber(selectedVariant?.stockQuantity ?? selectedVariant?.stock)
  const weightGrams = toNullablePositiveNumber(selectedVariant?.weightGrams ?? selectedVariant?.weight_grams)
  const variantTitle = String(selectedVariant?.title || optionRows.map(item => item.value).join(' / ') || '').trim()

  return {
    product: {
      id: productId,
      variant_id: variantId,
      title,
      slug,
      sku,
      url,
      thumbnail,
      price,
      price_value: priceValue,
    },
    selections: {
      variant_id: variantId,
      variant_title: variantTitle,
      sku,
      options: optionRows,
      stock,
      weight_grams: weightGrams,
      price,
      price_value: priceValue,
    },
    note: 'Customer requested staff confirmation for this product configuration.',
  }
}
