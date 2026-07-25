export interface OrderChatItem {
  id: number
  product_id: number | null
  variant_id: number | null
  title: string
  sku: string
  quantity: number
  price: number
  total: number
  attributes: unknown
}

export interface OrderChatMetadata {
  order_id: number
  order_number: string
  title: string
  status: string
  payment_status: string
  shipping_status: string
  total: number
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
    const price = toFiniteNumber(item?.price)
    const total = toFiniteNumber(item?.total ?? item?.line_total ?? item?.subtotal, price * quantity)

    return {
      id: toFiniteNumber(item?.id),
      product_id: toNullablePositiveNumber(item?.product_id ?? item?.productId),
      variant_id: toNullablePositiveNumber(item?.variant_id ?? item?.variantId),
      title: String(item?.product_name || item?.name || item?.title || 'Product').trim(),
      sku: String(item?.sku || '').trim(),
      quantity,
      price,
      total,
      attributes: parseAttributes(item?.attributes),
    }
  })
}

export const buildOrderChatMetadata = (order: Record<string, any>): OrderChatMetadata => {
  const items = normalizeOrderItems(order)
  const orderID = toFiniteNumber(order?.id ?? order?.order_id)
  const orderNumber = String(order?.order_number || order?.orderNumber || '').trim()
  const title = String(order?.title || `Order #${orderNumber || orderID}`).trim()
  const total = toFiniteNumber(order?.total ?? order?.total_amount)

  return {
    order_id: orderID,
    order_number: orderNumber,
    title,
    status: String(order?.status || '').trim(),
    payment_status: String(order?.payment_status || '').trim(),
    shipping_status: String(order?.shipping_status || '').trim(),
    total,
    currency: String(order?.currency || 'USD').trim(),
    url: String(order?.url || (orderID ? `/orders/${orderID}` : '')).trim(),
    thumbnail: String(order?.thumbnail || '').trim(),
    item_count: toFiniteNumber(order?.item_count, items.reduce((sum, item) => sum + item.quantity, 0)),
    items,
    note: 'Customer asked staff to confirm this order and its purchased configuration.',
  }
}
