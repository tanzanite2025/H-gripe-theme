import { useApiRequest } from '~/composables/useApiRequest'

type ApiEnvelope<T> = T | { data?: T | { data?: T } }

export interface WeChatPaymentSession {
  id: string
  status: string
  amount: number
  currency: string
  payment_url?: string
  transaction_id?: string
  metadata?: Record<string, string>
}

export interface CreateWeChatOrderInput {
  orderNumber: string
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

export function useWeChatPayment() {
  const { request } = useApiRequest()

  const createWeChatOrder = async (input: CreateWeChatOrderInput): Promise<WeChatPaymentSession> => {
    const orderNumber = assertOrderNumber(input.orderNumber)
    const response = await request<ApiEnvelope<WeChatPaymentSession>>('/payment/wechat/orders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify({ order_number: orderNumber }),
    }, 'Unable to start WeChat Pay payment')

    const session = unwrapApiData<WeChatPaymentSession>(response)
    if (!session?.id) {
      throw new Error('Invalid WeChat Pay payment response')
    }
    if (!String(session.payment_url || '').trim()) {
      throw new Error('WeChat Pay QR code URL is missing')
    }
    return session
  }

  const confirmWeChatOrder = async (orderNumber: string): Promise<WeChatPaymentSession> => {
    const normalizedOrderNumber = assertOrderNumber(orderNumber)
    const response = await request<ApiEnvelope<WeChatPaymentSession>>(
      `/payment/wechat/orders/${encodeURIComponent(normalizedOrderNumber)}/confirm`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify({}),
      },
      'Unable to confirm WeChat Pay payment',
    )

    const result = unwrapApiData<WeChatPaymentSession>(response)
    if (!result?.id) {
      throw new Error('Invalid WeChat Pay confirmation response')
    }
    return result
  }

  const createWeChatQrDataUrl = async (codeUrl: string) => {
    const value = String(codeUrl || '').trim()
    if (!value) {
      throw new Error('WeChat Pay QR code URL is missing')
    }
    const qrcode = await import('qrcode')
    return qrcode.toDataURL(value, {
      errorCorrectionLevel: 'M',
      margin: 1,
      scale: 8,
      color: {
        dark: '#020617',
        light: '#ffffff',
      },
    })
  }

  return {
    createWeChatOrder,
    confirmWeChatOrder,
    createWeChatQrDataUrl,
  }
}
