import { ref } from 'vue'
import type { CartItem } from '~/types/cart'

type ApiResponse<T> = T | { data?: T }

export interface ShippingQuoteItemInput {
  product_id: number
  variant_id?: number | null
  quantity: number
}

export interface ShippingQuoteRequest {
  country: string
  currency?: string
  items: ShippingQuoteItemInput[]
}

export interface ShippingQuoteItemResult {
  product_id: number
  variant_id?: number | null
  product_type_id?: number | null
  template_id: number
  template_name: string
  packaging_rule_id?: number | null
  packaging_rule_name?: string
  quantity: number
  unit_price: number
  amount: number
  weight_grams: number
  packaging_weight_grams: number
  charge_weight_grams: number
  shipping_fee: number
  free_shipping: boolean
}

export interface ShippingQuoteOption {
  carrier_id: number
  carrier_name: string
  carrier_code: string
  carrier_service_id: number
  service_code: string
  service_name: string
  route_name?: string
  template_id: number
  template_name: string
  currency?: string
  billing_mode: string
  actual_weight_grams: number
  volumetric_weight_grams: number
  charge_weight_grams: number
  billable_weight_grams: number
  base_fee: number
  fuel_surcharge: number
  remote_surcharge: number
  shipping_fee: number
  free_shipping: boolean
  eta_min_days: number
  eta_max_days: number
  sort_order: number
}

export interface ShippingQuoteResult {
  shipping_fee: number
  free_shipping: boolean
  currency?: string
  source?: string
  items?: ShippingQuoteItemResult[]
  options?: ShippingQuoteOption[]
  selected_option?: ShippingQuoteOption | null
}

const unwrapApiData = <T>(payload: ApiResponse<T> | null | undefined): T | null => {
  if (!payload || typeof payload !== 'object') {
    return (payload as T) || null
  }
  if ('data' in payload && payload.data !== undefined) {
    return payload.data as T
  }
  return payload as T
}

const cartItemToQuoteItem = (item: CartItem): ShippingQuoteItemInput | null => {
  const productId = Number(item.product_id || item.id || 0)
  if (!productId) return null

  const variantId = Number(item.variant_id || 0)
  return {
    product_id: productId,
    variant_id: variantId > 0 ? variantId : null,
    quantity: Math.max(1, Number(item.quantity || 1)),
  }
}

export const useShippingQuote = () => {
  const { request } = useApiRequest()

  const quote = ref<ShippingQuoteResult | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const quoteCart = async (payload: ShippingQuoteRequest) => {
    const country = payload.country.trim().toUpperCase()
    if (!country) {
      quote.value = null
      error.value = 'Shipping country is required'
      return null
    }

    const items = payload.items
      .map((item) => ({
        product_id: Number(item.product_id || 0),
        variant_id: Number(item.variant_id || 0) > 0 ? Number(item.variant_id) : null,
        quantity: Math.max(1, Number(item.quantity || 1)),
      }))
      .filter((item) => item.product_id > 0)

    if (!items.length) {
      quote.value = null
      error.value = 'Shipping quote requires at least one item'
      return null
    }

    isLoading.value = true
    error.value = null
    try {
      const response = await request<ApiResponse<ShippingQuoteResult>>('/shipping/quote', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify({
          country,
          currency: payload.currency?.trim().toUpperCase() || 'USD',
          items,
        }),
      })
      const data = unwrapApiData<ShippingQuoteResult>(response)
      if (!data) throw new Error('Invalid shipping quote response')
      quote.value = data
      return data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unable to refresh shipping quote'
      quote.value = null
      error.value = message
      return null
    } finally {
      isLoading.value = false
    }
  }

  const quoteCartItems = (items: CartItem[], country: string, currency = 'USD') => {
    return quoteCart({
      country,
      currency,
      items: items.map(cartItemToQuoteItem).filter((item): item is ShippingQuoteItemInput => Boolean(item)),
    })
  }

  const reset = () => {
    quote.value = null
    error.value = null
    isLoading.value = false
  }

  return {
    quote,
    isLoading,
    error,
    quoteCart,
    quoteCartItems,
    reset,
  }
}
