import { useApiRequest } from '~/composables/useApiRequest'

type ApiEnvelope<T> = T | { data?: T | { data?: T } }

export interface PayPalPaymentSession {
  id: string
  status: string
  amount: number
  currency: string
  payment_url?: string
  transaction_id?: string
  metadata?: Record<string, string>
}

export interface CreatePayPalOrderInput {
  orderNumber: string
  returnUrl?: string
  cancelUrl?: string
}

export interface CapturePayPalOrderInput {
  orderNumber: string
  paypalOrderId: string
}

const unwrapApiData = <T>(payload: ApiEnvelope<T> | null | undefined): T | null => {
  let current: unknown = payload

  for (let depth = 0; depth < 3; depth += 1) {
    if (!current || typeof current !== 'object') {
      return (current as T) || null
    }
    if (!('data' in current)) {
      return current as T
    }
    current = (current as { data?: unknown }).data
  }

  return null
}

const assertOrderNumber = (orderNumber: string) => {
  const value = String(orderNumber || '').trim()
  if (!value) {
    throw new Error('Order number is required')
  }
  return value
}

export function usePayPalPayment() {
  const { request } = useApiRequest()

  const createPayPalOrder = async (input: CreatePayPalOrderInput): Promise<PayPalPaymentSession> => {
    const orderNumber = assertOrderNumber(input.orderNumber)
    const response = await request<ApiEnvelope<PayPalPaymentSession>>('/payment/paypal/orders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify({
        order_number: orderNumber,
        return_url: input.returnUrl || '',
        cancel_url: input.cancelUrl || '',
      }),
    }, 'Unable to start PayPal payment')

    const session = unwrapApiData<PayPalPaymentSession>(response)
    if (!session?.id) {
      throw new Error('Invalid PayPal payment response')
    }

    return session
  }

  const capturePayPalOrder = async (input: CapturePayPalOrderInput): Promise<PayPalPaymentSession> => {
    const orderNumber = assertOrderNumber(input.orderNumber)
    const paypalOrderId = String(input.paypalOrderId || '').trim()
    if (!paypalOrderId) {
      throw new Error('PayPal order id is required')
    }

    const response = await request<ApiEnvelope<PayPalPaymentSession>>(
      `/payment/paypal/orders/${encodeURIComponent(paypalOrderId)}/capture`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify({ order_number: orderNumber }),
      },
      'Unable to capture PayPal payment',
    )

    const result = unwrapApiData<PayPalPaymentSession>(response)
    if (!result?.id) {
      throw new Error('Invalid PayPal capture response')
    }

    return result
  }

  const redirectToPayPal = (session: PayPalPaymentSession) => {
    const paymentUrl = String(session.payment_url || '').trim()
    if (!paymentUrl) {
      throw new Error('PayPal approval URL is missing')
    }
    window.location.assign(paymentUrl)
  }

  return {
    createPayPalOrder,
    capturePayPalOrder,
    redirectToPayPal,
  }
}
