import { loadStripe, type Stripe, type StripeElements, type StripePaymentElement } from '@stripe/stripe-js'
import { shallowRef } from 'vue'
import { storefrontFontFamily, storefrontFontStylesheetUrl } from '~/utils/storefrontFonts'

export interface StripePaymentSession {
  clientSecret: string
  publishableKey: string
}

export interface StripeConfirmationResult {
  status: string
  paymentIntentId?: string
}

export function useStripePayment() {
  const stripe = shallowRef<Stripe | null>(null)
  const elements = shallowRef<StripeElements | null>(null)
  const paymentElement = shallowRef<StripePaymentElement | null>(null)

  const destroy = () => {
    paymentElement.value?.unmount()
    paymentElement.value = null
    elements.value = null
    stripe.value = null
  }

  const mount = async (container: HTMLElement, session: StripePaymentSession) => {
    if (!session.clientSecret || !session.publishableKey) {
      throw new Error('Stripe payment is not configured')
    }

    destroy()

    const loadedStripe = await loadStripe(session.publishableKey)
    if (!loadedStripe) {
      throw new Error('Unable to load Stripe')
    }

    const loadedElements = loadedStripe.elements({
      clientSecret: session.clientSecret,
      fonts: [
        {
          cssSrc: storefrontFontStylesheetUrl(),
        },
      ],
      appearance: {
        theme: 'night',
        variables: {
          colorPrimary: '#b5ff6d',
          colorBackground: '#080a0c',
          colorText: '#f5f7f8',
          colorDanger: '#fb7185',
          borderRadius: '10px',
          fontFamily: storefrontFontFamily,
        },
        rules: {
          '.Input': {
            border: '1px solid rgba(255,255,255,0.14)',
            boxShadow: 'none',
          },
          '.Input:focus': {
            border: '1px solid rgba(181,255,109,0.75)',
            boxShadow: '0 0 0 1px rgba(181,255,109,0.18)',
          },
          '.Label': {
            color: 'rgba(245,247,248,0.72)',
            fontSize: '13px',
          },
        },
      },
    })
    const mountedElement = loadedElements.create('payment', {
      layout: 'tabs',
    })

    mountedElement.mount(container)
    stripe.value = loadedStripe
    elements.value = loadedElements
    paymentElement.value = mountedElement
  }

  const confirm = async (returnUrl: string): Promise<StripeConfirmationResult> => {
    if (!stripe.value || !elements.value) {
      throw new Error('Stripe payment form is not ready')
    }

    const result = await stripe.value.confirmPayment({
      elements: elements.value,
      confirmParams: {
        return_url: returnUrl,
      },
      redirect: 'if_required',
    })

    if (result.error) {
      throw new Error(result.error.message || 'Stripe payment could not be confirmed')
    }

    if (!result.paymentIntent) {
      throw new Error('Stripe did not return a payment result')
    }

    return {
      status: result.paymentIntent.status,
      paymentIntentId: result.paymentIntent.id,
    }
  }

  return {
    stripe,
    elements,
    paymentElement,
    mount,
    confirm,
    destroy,
  }
}
