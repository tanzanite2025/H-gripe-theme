import { useApiRequest } from '~/composables/useApiRequest'

type ApiEnvelope<T> = T | { data?: T | { data?: T } }

export interface AlipayPaymentSession {
  id: string
  status: string
  amount: number
  currency: string
  payment_url?: string
  transaction_id?: string
  metadata?: Record<string, string>
}

export interface CreateAlipayOrderInput {
  orderNumber: string
  returnUrl?: string
  cancelUrl?: string
  idempotencyKey?: string
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

export function useAlipayPayment() {
  const { request } = useApiRequest()

  const createAlipayOrder = async (input: CreateAlipayOrderInput): Promise<AlipayPaymentSession> => {
    const orderNumber = assertOrderNumber(input.orderNumber)
    const idempotencyKey = String(input.idempotencyKey || '').trim()
    const response = await request<ApiEnvelope<AlipayPaymentSession>>('/payment/alipay/orders', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        ...(idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {}),
      },
      body: JSON.stringify({
        order_number: orderNumber,
        return_url: input.returnUrl || '',
        cancel_url: input.cancelUrl || '',
      }),
    }, 'Unable to start Alipay payment')

    const session = unwrapApiData<AlipayPaymentSession>(response)
    if (!session?.id) {
      throw new Error('Invalid Alipay payment response')
    }
    return session
  }

  const confirmAlipayOrder = async (orderNumber: string, idempotencyKey?: string): Promise<AlipayPaymentSession> => {
    const normalizedOrderNumber = assertOrderNumber(orderNumber)
    const normalizedIdempotencyKey = String(idempotencyKey || '').trim()
    const response = await request<ApiEnvelope<AlipayPaymentSession>>(
      `/payment/alipay/orders/${encodeURIComponent(normalizedOrderNumber)}/confirm`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
          ...(normalizedIdempotencyKey ? { 'Idempotency-Key': normalizedIdempotencyKey } : {}),
        },
        body: JSON.stringify({}),
      },
      'Unable to confirm Alipay payment',
    )

    const result = unwrapApiData<AlipayPaymentSession>(response)
    if (!result?.id) {
      throw new Error('Invalid Alipay confirmation response')
    }
    return result
  }

  const redirectToAlipay = (session: AlipayPaymentSession) => {
    const paymentUrl = String(session.payment_url || '').trim()
    if (!paymentUrl) {
      throw new Error('Alipay checkout URL is missing')
    }
    window.location.assign(paymentUrl)
  }

  return {
    createAlipayOrder,
    confirmAlipayOrder,
    redirectToAlipay,
  }
}
