import type { Stripe, StripeElements, StripePaymentElement } from '@stripe/stripe-js'
import { shallowRef } from 'vue'
import { useI18n } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import { storefrontFontFamilyForLocale, storefrontFontStylesheetUrl } from '~/utils/storefrontFonts'
import { createStripeInstance } from '~/utils/security/stripeClient'

export interface StripePaymentSession {
  clientSecret: string
  publishableKey: string
  orderNumber?: string
}

export interface StripePaymentBillingDetails {
  name?: string
  email?: string
  phone?: string
  address: {
    line1: string
    line2?: string
    city: string
    state?: string
    postal_code: string
    country: string
  }
}

export interface StripeConfirmationResult {
  status: string
  paymentIntentId?: string
}

type ApiResponse<T> = T | { data?: T | { data?: T } }

const unwrapApiData = <T,>(payload: ApiResponse<T> | null | undefined): T | null => {
  let current: unknown = payload
  for (let depth = 0; depth < 3; depth += 1) {
    if (!current || typeof current !== 'object') return (current as T) || null
    if (!('data' in current)) return current as T
    current = (current as { data?: unknown }).data
  }
  return null
}

export function useStripePayment() {
  const { locale } = useI18n()
  const { request } = useApiRequest()
  const stripe = shallowRef<Stripe | null>(null)
  const elements = shallowRef<StripeElements | null>(null)
  const paymentElement = shallowRef<StripePaymentElement | null>(null)

  const destroy = () => {
    paymentElement.value?.unmount()
    paymentElement.value = null
    elements.value = null
    stripe.value = null
  }

  const mount = async (
    container: HTMLElement,
    session: StripePaymentSession,
    billingDetails?: StripePaymentBillingDetails,
  ) => {
    if (!session.clientSecret || !session.publishableKey) {
      throw new Error('Stripe payment is not configured')
    }

    destroy()

    const loadedStripe = await createStripeInstance(session.publishableKey)

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
          colorPrimary: '#059669',
          colorBackground: '#080a0c',
          colorText: '#f5f7f8',
          colorDanger: '#fb7185',
          borderRadius: '10px',
          fontFamily: storefrontFontFamilyForLocale(locale.value),
        },
        rules: {
          '.Input': {
            border: '1px solid rgba(255,255,255,0.14)',
            boxShadow: 'none',
          },
          '.Input:focus': {
            border: '1px solid rgba(5, 150, 105,0.75)',
            boxShadow: '0 0 0 1px rgba(5, 150, 105,0.18)',
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
      fields: {
        billingDetails: {
          name: 'auto',
          email: 'auto',
          phone: 'auto',
          address: {
            line1: 'auto',
            line2: 'auto',
            city: 'auto',
            state: 'auto',
            postalCode: 'auto',
            country: 'auto',
          },
        },
      },
      ...(billingDetails ? { defaultValues: { billingDetails } } : {}),
    })

    mountedElement.mount(container)
    stripe.value = loadedStripe
    elements.value = loadedElements
    paymentElement.value = mountedElement
  }

  const loadPublishableKey = async () => {
    const response = await request<ApiResponse<{ publishable_key?: string; publishableKey?: string }>>(
      '/payment/stripe/express-checkout/config',
      { method: 'GET', headers: { Accept: 'application/json' } },
      'Stripe payment is not configured',
    )
    const config = unwrapApiData<{ publishable_key?: string; publishableKey?: string }>(response)
    const publishableKey = String(config?.publishableKey || config?.publishable_key || '').trim()
    if (!publishableKey) {
      throw new Error('Stripe publishable key is missing')
    }
    return publishableKey
  }

  const retrievePaymentIntent = async (
    clientSecret: string,
    publishableKey: string,
  ): Promise<StripeConfirmationResult> => {
    if (!clientSecret || !publishableKey) {
      throw new Error('Stripe payment recovery data is incomplete')
    }

    const loadedStripe = await createStripeInstance(publishableKey)
    const result = await loadedStripe.retrievePaymentIntent(clientSecret)
    if (result.error) {
      throw new Error(result.error.message || 'Stripe payment could not be recovered')
    }
    if (!result.paymentIntent) {
      throw new Error('Stripe did not return a payment result')
    }

    return {
      status: result.paymentIntent.status,
      paymentIntentId: result.paymentIntent.id,
    }
  }

  const confirm = async (
    returnUrl: string,
    billingDetails?: StripePaymentBillingDetails,
  ): Promise<StripeConfirmationResult> => {
    if (!stripe.value || !elements.value) {
      throw new Error('Stripe payment form is not ready')
    }

    const result = await stripe.value.confirmPayment({
      elements: elements.value,
      confirmParams: {
        return_url: returnUrl,
        ...(billingDetails ? { payment_method_data: { billing_details: billingDetails } } : {}),
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
    loadPublishableKey,
    retrievePaymentIntent,
    destroy,
  }
}
