import type {
  Stripe,
  StripeConstructor,
  StripeConstructorOptions,
} from '@stripe/stripe-js'
import { loadStripeScript } from '~/utils/security/trustedScriptUrl'

const getStripeConstructor = (): StripeConstructor => {
  const stripeConstructor = window.Stripe
  if (!stripeConstructor) {
    throw new Error('Stripe.js is not available')
  }
  return stripeConstructor
}

export const createStripeInstance = async (
  publishableKey: string,
  options?: StripeConstructorOptions,
): Promise<Stripe> => {
  if (typeof window === 'undefined') {
    throw new Error('Stripe.js can only load in the browser')
  }

  if (!window.Stripe) {
    await loadStripeScript()
  }

  return getStripeConstructor()(publishableKey, options)
}
