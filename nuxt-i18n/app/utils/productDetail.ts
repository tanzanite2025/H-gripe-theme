import type {
  GoProduct,
  ProductDisplayPrice,
  ProductSpecValue,
  ProductVariant,
} from '~/types/productDetail'

export const PRODUCT_DETAIL_HIDDEN_SPEC_SLUGS = new Set(['availability', 'sku'])

export const stripProductHtml = (value: string | null | undefined): string => {
  if (!value) return ''
  return value.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
}

export const extractProductDetailPayload = (
  response: unknown,
  expectedSlug?: unknown,
): GoProduct | null => {
  const responseRecord = response && typeof response === 'object' && !Array.isArray(response)
    ? response as Record<string, unknown>
    : null
  const candidate = responseRecord && Object.prototype.hasOwnProperty.call(responseRecord, 'data')
    ? responseRecord.data
    : response

  if (!candidate || typeof candidate !== 'object' || Array.isArray(candidate)) {
    return null
  }

  const record = candidate as Record<string, unknown>
  const id = Number(record.id)
  const name = String(record.name || '').trim()
  const slug = String(record.slug || '').trim()
  if (!Number.isInteger(id) || id <= 0 || !name || !slug) {
    return null
  }

  const requestedSlug = String(expectedSlug || '').trim()
  if (requestedSlug && slug !== requestedSlug) {
    return null
  }

  return candidate as GoProduct
}

export const normalizeProductCurrencyCode = (value: unknown) => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

export const validProductDisplayPrice = (displayPrice?: ProductDisplayPrice | null) => {
  const amount = Number(displayPrice?.amount)
  const currency = normalizeProductCurrencyCode(displayPrice?.currency)
  if (!Number.isFinite(amount) || amount <= 0 || !currency) return null
  return { amount, currency }
}

export const displayPriceSnapshotForCurrency = (
  displayPrices: ProductDisplayPrice[] | undefined,
  requestedCurrency: unknown,
) => {
  const currency = normalizeProductCurrencyCode(requestedCurrency)
  if (!currency || !Array.isArray(displayPrices)) return null
  const snapshot = displayPrices.find((price) => (
    normalizeProductCurrencyCode(price?.currency) === currency
  ))
  return validProductDisplayPrice(snapshot)
}

export const parseProductVariantOptions = (
  variant: Pick<ProductVariant, 'option_values'> | null | undefined,
): Record<string, string> => {
  if (!variant?.option_values) return {}
  if (typeof variant.option_values === 'object') {
    return Object.entries(variant.option_values).reduce<Record<string, string>>((acc, [key, raw]) => {
      const value = String(raw ?? '').trim()
      if (key && value) acc[key] = value
      return acc
    }, {})
  }

  try {
    const parsed = JSON.parse(variant.option_values)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}
    return Object.entries(parsed as Record<string, unknown>).reduce<Record<string, string>>((acc, [key, raw]) => {
      const value = String(raw ?? '').trim()
      if (key && value) acc[key] = value
      return acc
    }, {})
  } catch {
    return {}
  }
}

export const humanizeProductSpecSlug = (slug: string) => {
  return slug
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, char => char.toUpperCase())
}

export const formatProductSpecValue = (item: ProductSpecValue) => {
  const definition = item.definition
  const value = String(item.value || '').trim()
  if (!definition) return value

  if (definition.field_type === 'boolean') {
    return value === 'true' ? 'Yes' : 'No'
  }
  if (definition.unit && value) {
    return `${value} ${definition.unit}`
  }
  return value
}
