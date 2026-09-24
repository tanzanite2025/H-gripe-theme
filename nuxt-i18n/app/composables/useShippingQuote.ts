import { ref } from 'vue'
import type { CartItem } from '~~/types/cart'

type ApiResponse<T> = T | { data?: T }

export interface ShippingQuoteItemInput {
  product_id: number
  variant_id?: number | null
  quantity: number
  configuration?: unknown
}

export interface ShippingQuoteRequest {
  country: string
  postal_code?: string
  currency: string
  display_currency?: string
  shipping_quote_id?: string
  selected_quote_plan_id?: string
  items: ShippingQuoteItemInput[]
}

export interface CheckoutQuoteRequest {
  shipping_address: {
    first_name?: string
    last_name?: string
    company?: string
    address1?: string
    address2?: string
    city?: string
    state?: string
    postal_code?: string
    country: string
    phone?: string
    email?: string
  }
  display_currency?: string
  payment_method?: string
  shipping_quote_id?: string
  selected_quote_plan_id?: string
  coupon_code?: string
}

export interface ShippingDisplayPrice {
  amount: number
  currency: string
  quote_currency?: string
  rate?: number
  source?: string
  converted?: boolean
  fallback_reason?: string
}

export interface ShippingQuoteItemResult {
  product_id: number
  variant_id?: number | null
  product_specification_template_id?: number | null
  template_id: number
  template_name: string
  packaging_rule_id?: number | null
  packaging_rule_name?: string
  quantity: number
  unit_price_minor: number | string
  amount_minor: number | string
  weight_grams: number
  packaging_weight_grams: number
  charge_weight_grams: number
  shipping_fee_minor: number | string
  free_shipping: boolean
}

export interface ShippingQuoteLeg {
  group_key: string
  item_indexes: number[]
  allocation_basis: string
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
  base_fee_minor: number | string
  fuel_surcharge_minor: number | string
  remote_surcharge_minor: number | string
  shipping_fee_minor: number | string
  display_price?: ShippingDisplayPrice | null
  display_prices?: ShippingDisplayPrice[]
  free_shipping: boolean
  eta_min_days: number
  eta_max_days: number
  sort_order: number
}

export interface ShippingQuotePlan {
  id: string
  currency: string
  shipping_fee_minor: number | string
  display_price?: ShippingDisplayPrice | null
  display_prices?: ShippingDisplayPrice[]
  free_shipping: boolean
  eta_min_days: number
  eta_max_days: number
  legs: ShippingQuoteLeg[]
}

export interface ShippingQuoteResult {
  id: string
  rate_version: string
  expires_at: string
  shipping_fee_minor: number | string
  free_shipping: boolean
  currency?: string
  display_price?: ShippingDisplayPrice | null
  display_prices?: ShippingDisplayPrice[]
  display_currency?: string
  items?: ShippingQuoteItemResult[]
  plans: ShippingQuotePlan[]
  selected_plan?: ShippingQuotePlan | null
}

export interface CheckoutQuoteResult {
  currency: string
  subtotal_minor: number | string
  shipping_fee_minor: number | string
  tax_minor: number | string
  member_discount_minor: number | string
  coupon_discount_minor: number | string
  discount_minor: number | string
  total_minor: number | string
  coupon_code?: string
  payment_currency?: string
  payment_amount_minor?: number | string
  shipping_quote?: ShippingQuoteResult | null
}

const unwrapApiData = <T>(payload: ApiResponse<T> | null | undefined): T | null => {
  let current: unknown = payload
  for (let depth = 0; depth < 3; depth += 1) {
    if (!current || typeof current !== 'object' || Array.isArray(current)) {
      return (current as T) || null
    }
    if (!('data' in current)) return current as T
    current = (current as { data?: unknown }).data
  }
  return null
}

const cartItemToQuoteItem = (item: CartItem): ShippingQuoteItemInput | null => {
  const productId = Number(item.product_id || item.id || 0)
  if (!productId) return null

  const variantId = Number(item.variant_id || 0)
  return {
    product_id: productId,
    variant_id: variantId > 0 ? variantId : null,
    quantity: Math.max(1, Number(item.quantity || 1)),
    configuration: { schema_version: 1, selections: item.selected_options || [] },
  }
}

export const useShippingQuote = () => {
  const { request } = useApiRequest()

  const quote = ref<ShippingQuoteResult | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  let requestVersion = 0

  const quoteCart = async (payload: ShippingQuoteRequest) => {
    const currentRequestVersion = ++requestVersion
    const country = payload.country.trim().toUpperCase()
    if (!country) {
      quote.value = null
      error.value = 'Shipping country is required'
      isLoading.value = false
      return null
    }
    const currency = payload.currency.trim().toUpperCase()
    if (!currency) {
      quote.value = null
      error.value = 'Shipping quote currency is required'
      isLoading.value = false
      return null
    }
    const displayCurrency = payload.display_currency?.trim().toUpperCase()

    const items = payload.items
      .map((item) => ({
        product_id: Number(item.product_id || 0),
        variant_id: Number(item.variant_id || 0) > 0 ? Number(item.variant_id) : null,
        quantity: Math.max(1, Number(item.quantity || 1)),
        ...(item.configuration !== undefined ? { configuration: item.configuration } : {}),
      }))
      .filter((item) => item.product_id > 0)

    if (!items.length) {
      quote.value = null
      error.value = 'Shipping quote requires at least one item'
      isLoading.value = false
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
          ...(payload.postal_code?.trim() ? { postal_code: payload.postal_code.trim() } : {}),
          currency,
          ...(displayCurrency ? { display_currency: displayCurrency } : {}),
          ...(payload.shipping_quote_id?.trim() ? { shipping_quote_id: payload.shipping_quote_id.trim() } : {}),
          ...(payload.selected_quote_plan_id?.trim() ? { selected_quote_plan_id: payload.selected_quote_plan_id.trim() } : {}),
          items,
        }),
      })
      const data = unwrapApiData<ShippingQuoteResult>(response)
      if (!data) throw new Error('Invalid shipping quote response')
      if (currentRequestVersion === requestVersion) quote.value = data
      return data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unable to refresh shipping quote'
      if (currentRequestVersion === requestVersion) {
        quote.value = null
        error.value = message
      }
      return null
    } finally {
      if (currentRequestVersion === requestVersion) isLoading.value = false
    }
  }

  const quoteCheckout = async (payload: CheckoutQuoteRequest): Promise<CheckoutQuoteResult> => {
    const response = await request<ApiResponse<CheckoutQuoteResult>>('/checkout/quote', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify(payload),
    })
    const data = unwrapApiData<CheckoutQuoteResult>(response)
    if (!data) throw new Error('Invalid checkout quote response')
    return data
  }

  const quoteCartItems = (items: CartItem[], country: string, currency: string, displayCurrency?: string, postalCode?: string) => {
    return quoteCart({
      country,
      postal_code: postalCode,
      currency,
      display_currency: displayCurrency,
      items: items.map(cartItemToQuoteItem).filter((item): item is ShippingQuoteItemInput => Boolean(item)),
    })
  }

  const reset = () => {
    requestVersion += 1
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
    quoteCheckout,
    reset,
  }
}
