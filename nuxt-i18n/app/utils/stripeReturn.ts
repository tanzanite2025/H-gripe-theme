export const STRIPE_RETURN_PATH = '/checkout/stripe/return'

export interface StripeReturnSession {
  orderNumber: string
  publishableKey: string
  clientSecret?: string
}

const storageKey = (orderNumber: string) => `checkout:stripe:${orderNumber}`

export const saveStripeReturnSession = (session: StripeReturnSession) => {
  if (typeof window === 'undefined' || !session.orderNumber || !session.publishableKey) return

  window.sessionStorage.setItem(storageKey(session.orderNumber), JSON.stringify(session))
}

export const readStripeReturnSession = (orderNumber: string): StripeReturnSession | null => {
  if (typeof window === 'undefined' || !orderNumber) return null

  const raw = window.sessionStorage.getItem(storageKey(orderNumber))
  if (!raw) return null

  try {
    const session = JSON.parse(raw) as Partial<StripeReturnSession>
    if (!session.orderNumber || !session.publishableKey) return null
    return {
      orderNumber: String(session.orderNumber),
      publishableKey: String(session.publishableKey),
      clientSecret: session.clientSecret ? String(session.clientSecret) : undefined,
    }
  } catch {
    return null
  }
}

export const clearStripeReturnSession = (orderNumber: string) => {
  if (typeof window === 'undefined' || !orderNumber) return

  window.sessionStorage.removeItem(storageKey(orderNumber))
}
