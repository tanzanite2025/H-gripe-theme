export interface OrderChatItem {
  id: number
  product_id: number | null
  variant_id: number | null
  title: string
  sku: string
  quantity: number
  currency: string
  price_minor: number
  total_minor: number
  attributes: unknown
}

export interface OrderChatMetadata {
  order_number: string
  title: string
  status: string
  payment_status: string
  shipping_status: string
  fulfillment_mode: string
  production_status: string
  signature_required: boolean
  production_started_at: string
  production_completed_at: string
  total_minor: number
  currency: string
  url: string
  thumbnail: string
  item_count: number
  items: OrderChatItem[]
  note: string
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

const parseAttributes = (value: unknown) => {
  if (!value) return null
  if (typeof value === 'object') return value
  if (typeof value !== 'string') return null
  try {
    return JSON.parse(value)
  } catch {
    return value
  }
}

const normalizeOrderItems = (order: Record<string, any>): OrderChatItem[] => {
  const items = Array.isArray(order?.items) ? order.items : []

  return items.map((item: Record<string, any>) => {
    const quantity = toFiniteNumber(item?.quantity, 1)
    const priceMinor = toFiniteNumber(item?.price_minor)
    const totalMinor = toFiniteNumber(item?.total_minor, priceMinor * quantity)

    return {
      id: toFiniteNumber(item?.id),
      product_id: toNullablePositiveNumber(item?.product_id ?? item?.productId),
      variant_id: toNullablePositiveNumber(item?.variant_id ?? item?.variantId),
      title: String(item?.product_name || item?.name || item?.title || 'Product').trim(),
      sku: String(item?.sku || '').trim(),
      fulfillment_mode: String(item?.fulfillment_mode || '').trim(),
      quantity,
      currency: String(item?.currency || '').trim().toUpperCase(),
      price_minor: priceMinor,
      total_minor: totalMinor,
      attributes: parseAttributes(item?.attributes),
    }
  })
}

export const buildOrderChatMetadata = (order: Record<string, any>): OrderChatMetadata => {
  const items = normalizeOrderItems(order)
  const orderNumber = String(order?.order_number || order?.orderNumber || '').trim()
  const title = String(order?.title || (orderNumber ? `Order #${orderNumber}` : 'Order')).trim()
  const totalMinor = toFiniteNumber(order?.total_minor)

  return {
    order_number: orderNumber,
    title,
    status: String(order?.status || '').trim(),
    payment_status: String(order?.payment_status || '').trim(),
    shipping_status: String(order?.shipping_status || '').trim(),
    fulfillment_mode: String(order?.fulfillment_mode || '').trim(),
    production_status: String(order?.production_status || '').trim(),
    signature_required: Boolean(order?.signature_required),
    production_started_at: String(order?.production_started_at || '').trim(),
    production_completed_at: String(order?.production_completed_at || '').trim(),
    total_minor: totalMinor,
    currency: String(order?.currency || '').trim().toUpperCase(),
    url: String(order?.url || '').trim(),
    thumbnail: String(order?.thumbnail || '').trim(),
    item_count: toFiniteNumber(order?.item_count, items.reduce((sum, item) => sum + item.quantity, 0)),
    items,
    note: 'Customer asked staff to confirm this order and its purchased configuration.',
  }
}
