export interface ProductChatMetadata {
  kind: 'product_reference'
  product_id: number | null
  variant_id: number | null
  title: string
  slug: string
  sku: string
  url: string
  thumbnail: string
  price: string
  price_value: number
}

const toFiniteNumber = (value: unknown, fallback = 0) => {
  if (value === null || value === undefined || value === '') return fallback
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : fallback
}

const toNullablePositiveNumber = (value: unknown) => {
  const numberValue = toFiniteNumber(value, 0)
  return numberValue > 0 ? numberValue : null
}

export const buildProductChatMetadata = (product: Record<string, any>): ProductChatMetadata => {
  const productId = toNullablePositiveNumber(product?.id ?? product?.product_id ?? product?.productId)
  const variantId = toNullablePositiveNumber(
    product?.selectedVariantId
      ?? product?.variant_id
      ?? product?.variantId
      ?? product?.defaultVariantId
      ?? product?.default_variant_id
  )
  const title = String(product?.title || product?.name || 'Product').trim()
  const slug = String(product?.slug || '').trim()
  const sku = String(product?.sku || '').trim()
  const url = String(product?.url || (slug ? `/shop/${slug}` : '')).trim()
  const thumbnail = String(product?.thumbnail || product?.image || product?.featured_image || '').trim()
  const priceValue = toFiniteNumber(
    product?.priceValue
      ?? product?.priceNumber
      ?? product?.price_value
      ?? product?.prices?.sale
      ?? product?.prices?.regular
  )
  const price = String(
    product?.price
      || product?.priceLabel
      || (priceValue > 0 ? `$${priceValue}` : '')
  ).trim()

  return {
    kind: 'product_reference',
    product_id: productId,
    variant_id: variantId,
    title,
    slug,
    sku,
    url,
    thumbnail,
    price,
    price_value: priceValue,
  }
}
