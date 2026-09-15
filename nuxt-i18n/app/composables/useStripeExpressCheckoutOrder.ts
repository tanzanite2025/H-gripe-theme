import type {
  StripeExpressCheckoutElementConfirmEvent,
  StripeExpressCheckoutElementShippingAddressChangeEvent,
} from '@stripe/stripe-js'
import type { CartItem } from '~~/types/cart'
import { useAuth } from '~/composables/useAuth'
import { createIdempotencyKey } from '~/utils/idempotency'
import { useStorefrontContext } from '~/composables/useStorefrontContext'

type ApiResponse<T> = T | { data?: T | { data?: T } }

interface CheckoutQuoteResponse {
  total_amount?: number | string | null
  shipping_quote?: {
    id?: string
    selected_plan?: { id?: string }
  }
}

export interface StripeExpressCheckoutOrderSession {
  orderNumber: string
  clientSecret: string
  publishableKey: string
}

interface StripeExpressCheckoutAddress {
  name: string
  address: {
    line1: string
    line2?: string | null
    city: string
    state?: string
    postal_code: string
    country: string
  }
}

interface StripeExpressCheckoutPaymentDetails {
  email?: string
  phone?: string
  address?: StripeExpressCheckoutAddress['address']
}

const unwrapApiData = <T,>(payload: ApiResponse<T> | null | undefined): T | null => {
  let current: unknown = payload
  for (let depth = 0; depth < 3; depth += 1) {
    if (!current || typeof current !== 'object') return (current as T) || null
    if (!('data' in current)) return current as T
    current = (current as { data?: unknown }).data
  }
  return null
}

const splitCustomerName = (value: string) => {
  const parts = String(value || '').trim().split(/\s+/).filter(Boolean)
  return {
    firstName: parts[0] || 'Customer',
    lastName: parts.length > 1 ? parts.slice(1).join(' ') : 'User',
  }
}

const normalizeCountryCode = (value: unknown) => String(value || '').trim().toUpperCase()

const buildOrderAddressFromStripeExpressCheckoutDetails = (
  addressDetails: StripeExpressCheckoutAddress | undefined,
  paymentDetails: StripeExpressCheckoutPaymentDetails | undefined,
  fallbackEmail: string,
) => {
  if (!addressDetails?.address?.line1 || !addressDetails.address.city || !addressDetails.address.country) {
    throw new Error('Apple Pay or Google Pay did not return a complete shipping address')
  }

  const name = splitCustomerName(addressDetails.name)
  const email = String(paymentDetails?.email || fallbackEmail || '').trim()
  const phone = String(paymentDetails?.phone || '').trim()
  if (!email || !phone) {
    throw new Error('Apple Pay or Google Pay did not return the required email and phone number')
  }

  return {
    first_name: name.firstName,
    last_name: name.lastName,
    address1: addressDetails.address.line1.trim(),
    address2: String(addressDetails.address.line2 || '').trim(),
    city: addressDetails.address.city.trim(),
    state: String(addressDetails.address.state || '').trim(),
    postal_code: String(addressDetails.address.postal_code || '').trim(),
    country: normalizeCountryCode(addressDetails.address.country),
    phone,
    email,
  }
}

export function useStripeExpressCheckoutOrder() {
  const auth = useAuth()
  const { displayCurrency } = useStorefrontContext()

  const loadStripeExpressCheckoutPublishableKey = async () => {
    const response = await auth.request<ApiResponse<{ publishable_key?: string; publishableKey?: string }>>(
      '/payment/stripe/express-checkout/config',
      { method: 'GET', headers: { Accept: 'application/json' } },
      'Stripe Express Checkout is not configured',
    )
    const config = unwrapApiData<{ publishable_key?: string; publishableKey?: string }>(response)
    const publishableKey = String(config?.publishableKey || config?.publishable_key || '').trim()
    if (!publishableKey) {
      throw new Error('Stripe Express Checkout publishable key is missing')
    }
    return publishableKey
  }

  const createLocalOrderFromStripeExpressCheckoutConfirmation = async (
    confirmationEvent: StripeExpressCheckoutElementConfirmEvent,
    cartItems: CartItem[],
    ensureCartReady?: () => Promise<void>,
    idempotencyKey?: string,
  ) => {
    const session = await auth.ensureSession()
    if (!session) {
      throw new Error('Please sign in before using Apple Pay or Google Pay')
    }

    if (ensureCartReady) {
      await ensureCartReady()
    }

    const shippingAddress = buildOrderAddressFromStripeExpressCheckoutDetails(
      confirmationEvent.shippingAddress,
      confirmationEvent.billingDetails,
      String(session.email || ''),
    )
    const billingDetails = confirmationEvent.billingDetails
    let billingAddress = shippingAddress
    if (billingDetails?.address) {
      try {
        billingAddress = buildOrderAddressFromStripeExpressCheckoutDetails(
          {
            name: billingDetails.name,
            address: billingDetails.address,
          },
          billingDetails,
          shippingAddress.email,
        )
      } catch {
        billingAddress = shippingAddress
      }
    }

    const quoteResponse = await auth.request<ApiResponse<CheckoutQuoteResponse>>('/checkout/quote', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({ shipping_address: shippingAddress, display_currency: String(displayCurrency.value || '').trim().toUpperCase() }),
    }, 'Express Checkout quote refresh failed')
    const quote = unwrapApiData<CheckoutQuoteResponse>(quoteResponse)
    const expectedTotal = Number(quote?.total_amount)
    if (!Number.isFinite(expectedTotal)) {
      throw new Error('Express Checkout quote did not include a valid total')
    }
    const shippingQuoteID = String(quote?.shipping_quote?.id || '').trim()
    const selectedQuotePlanID = String(quote?.shipping_quote?.selected_plan?.id || '').trim()
    if (!shippingQuoteID || !selectedQuotePlanID) {
      throw new Error('Express Checkout quote did not include a shipping plan')
    }

    const response = await auth.request<ApiResponse<{ order_number?: string }>>('/orders', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        ...(String(idempotencyKey || '').trim() ? { 'Idempotency-Key': String(idempotencyKey || '').trim() } : {}),
      },
      body: JSON.stringify({
        items: cartItems.map(item => ({
          product_id: Number(item.product_id || item.id || 0),
          variant_id: item.variant_id || null,
          quantity: Math.max(1, Number(item.quantity || 1)),
        })),
        shipping_address: shippingAddress,
        billing_address: billingAddress,
        payment_method: 'card',
        shipping_method: 'standard',
        shipping_quote_id: shippingQuoteID,
        selected_quote_plan_id: selectedQuotePlanID,
        expected_total: Number(expectedTotal.toFixed(2)),
        display_currency: String(displayCurrency.value || '').trim().toUpperCase(),
      }),
    }, 'Express Checkout order creation failed')
    const order = unwrapApiData<{ order_number?: string }>(response)
    if (!order?.order_number) {
      throw new Error('Express Checkout order creation failed')
    }
    return {
      orderNumber: order.order_number,
      shippingAddress,
    }
  }

  const createStripePaymentIntentForExpressCheckoutOrder = async (orderNumber: string, idempotencyKey?: string) => {
    const response = await auth.request<ApiResponse<{
      client_secret?: string
      clientSecret?: string
      publishable_key?: string
      publishableKey?: string
    }>>('/payment/stripe/payment-intents', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        ...(String(idempotencyKey || '').trim() ? { 'Idempotency-Key': String(idempotencyKey || '').trim() } : {}),
      },
      body: JSON.stringify({ order_number: orderNumber }),
    }, 'Stripe Express Checkout PaymentIntent creation failed')
    const payment = unwrapApiData<{
      client_secret?: string
      clientSecret?: string
      publishable_key?: string
      publishableKey?: string
    }>(response)
    const clientSecret = String(payment?.clientSecret || payment?.client_secret || '').trim()
    const publishableKey = String(payment?.publishableKey || payment?.publishable_key || '').trim()
    if (!clientSecret) {
      throw new Error('Stripe Express Checkout PaymentIntent response is incomplete')
    }
    return { clientSecret, publishableKey }
  }

  const createStripeExpressCheckoutOrderAndPaymentSession = async (
    confirmationEvent: StripeExpressCheckoutElementConfirmEvent,
    cartItems: CartItem[],
    ensureCartReady?: () => Promise<void>,
  ): Promise<StripeExpressCheckoutOrderSession> => {
    const idempotencyKey = createIdempotencyKey('stripe-express-checkout')
    const order = await createLocalOrderFromStripeExpressCheckoutConfirmation(
      confirmationEvent,
      cartItems,
      ensureCartReady,
      idempotencyKey,
    )
    const payment = await createStripePaymentIntentForExpressCheckoutOrder(order.orderNumber, idempotencyKey)
    const publishableKey = payment.publishableKey || await loadStripeExpressCheckoutPublishableKey()
    return {
      orderNumber: order.orderNumber,
      clientSecret: payment.clientSecret,
      publishableKey,
    }
  }

  const buildExpressCheckoutShippingQuoteAddress = (
    shippingEvent: StripeExpressCheckoutElementShippingAddressChangeEvent,
    fallbackEmail: string,
    fallbackPhone: string,
  ) => {
    const name = splitCustomerName(shippingEvent.name)
    return {
      first_name: name.firstName,
      last_name: name.lastName,
      address1: '',
      city: shippingEvent.address.city,
      state: shippingEvent.address.state,
      postal_code: shippingEvent.address.postal_code,
      country: normalizeCountryCode(shippingEvent.address.country),
      phone: fallbackPhone,
      email: fallbackEmail,
    }
  }

  return {
    loadStripeExpressCheckoutPublishableKey,
    createStripeExpressCheckoutOrderAndPaymentSession,
    buildExpressCheckoutShippingQuoteAddress,
  }
}
