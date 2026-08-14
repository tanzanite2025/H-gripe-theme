import {
  loadStripe,
  type Stripe,
  type StripeElements,
  type StripeExpressCheckoutElement,
  type StripeExpressCheckoutElementAvailablePaymentMethodsChangeEvent,
  type StripeExpressCheckoutElementClickEvent,
  type StripeExpressCheckoutElementConfirmEvent,
  type StripeExpressCheckoutElementReadyEvent,
  type StripeExpressCheckoutElementShippingAddressChangeEvent,
  type StripeExpressCheckoutElementShippingRateChangeEvent,
  type StripeExpressCheckoutElementOptions,
  type StripeError,
} from '@stripe/stripe-js'
import { shallowRef, ref } from 'vue'
import { storefrontFontFamily, storefrontFontStylesheetUrl } from '~/utils/storefrontFonts'

export interface StripeExpressCheckoutLineItem {
  name: string
  amount: number
}

export interface StripeExpressCheckoutShippingRate {
  id: string
  amount: number
  displayName: string
  deliveryEstimate?: {
    maximum?: {
      unit: 'hour' | 'day' | 'business_day' | 'week' | 'month'
      value: number
    }
    minimum?: {
      unit: 'hour' | 'day' | 'business_day' | 'week' | 'month'
      value: number
    }
  }
}

export interface StripeExpressCheckoutMountSession {
  publishableKey: string
  amount: number
  currency: string
  lineItems: StripeExpressCheckoutLineItem[]
  shippingRates?: StripeExpressCheckoutShippingRate[]
  allowedShippingCountries?: string[]
}

export interface StripeExpressCheckoutConfirmationResult {
  status: string
  paymentIntentId?: string
}

export interface StripeExpressCheckoutAvailablePaymentMethods {
  applePay: boolean
  googlePay: boolean
}

export interface StripeExpressCheckoutEventHandlers {
  onReady?: (event: StripeExpressCheckoutElementReadyEvent) => void
  onClick?: (event: StripeExpressCheckoutElementClickEvent) => void
  onConfirm?: (event: StripeExpressCheckoutElementConfirmEvent) => void | Promise<void>
  onShippingAddressChange?: (event: StripeExpressCheckoutElementShippingAddressChangeEvent) => void | Promise<void>
  onShippingRateChange?: (event: StripeExpressCheckoutElementShippingRateChangeEvent) => void | Promise<void>
  onAvailablePaymentMethodsChange?: (event: StripeExpressCheckoutElementAvailablePaymentMethodsChangeEvent) => void
  onError?: (error: StripeError | Error) => void
  onCancel?: () => void
}

const ZERO_DECIMAL_STRIPE_CURRENCIES = new Set([
  'BIF',
  'CLP',
  'DJF',
  'GNF',
  'JPY',
  'KMF',
  'KRW',
  'MGA',
  'PYG',
  'RWF',
  'UGX',
  'VND',
  'VUV',
  'XAF',
  'XOF',
  'XPF',
])

const normalizeCurrencyCode = (value: unknown) => {
  const currency = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(currency) ? currency : ''
}

export const convertMajorAmountToStripeMinorAmount = (amount: number, currency: string) => {
  const normalizedCurrency = normalizeCurrencyCode(currency)
  const multiplier = ZERO_DECIMAL_STRIPE_CURRENCIES.has(normalizedCurrency) ? 1 : 100
  return Math.max(0, Math.round(Number(amount || 0) * multiplier))
}

const buildStripeExpressCheckoutElementOptions = (
  session: StripeExpressCheckoutMountSession,
): StripeExpressCheckoutElementOptions => {
  const currency = normalizeCurrencyCode(session.currency)
  if (!currency) {
    throw new Error('Stripe Express Checkout currency is invalid')
  }

  return {
    buttonHeight: 48,
    buttonTheme: {
      applePay: 'white',
      googlePay: 'white',
    },
    buttonType: {
      applePay: 'buy',
      googlePay: 'buy',
    },
    emailRequired: true,
    phoneNumberRequired: true,
    shippingAddressRequired: true,
    allowedShippingCountries: session.allowedShippingCountries,
    lineItems: session.lineItems,
    shippingRates: session.shippingRates,
    paymentMethods: {
      applePay: 'auto',
      googlePay: 'auto',
      link: 'never',
      paypal: 'never',
      amazonPay: 'never',
      klarna: 'never',
    },
  }
}

const toExpressCheckoutAvailablePaymentMethods = (
  paymentMethods:
    | StripeExpressCheckoutElementAvailablePaymentMethodsChangeEvent['paymentMethods']
    | StripeExpressCheckoutElementReadyEvent['availablePaymentMethods'],
): StripeExpressCheckoutAvailablePaymentMethods => ({
  applePay: Boolean(paymentMethods?.applePay),
  googlePay: Boolean(paymentMethods?.googlePay),
})

export function useStripeExpressCheckout() {
  const stripe = shallowRef<Stripe | null>(null)
  const elements = shallowRef<StripeElements | null>(null)
  const expressCheckoutElement = shallowRef<StripeExpressCheckoutElement | null>(null)
  const isReady = ref(false)
  const hasResolvedAvailability = ref(false)
  const availablePaymentMethods = ref<StripeExpressCheckoutAvailablePaymentMethods>({
    applePay: false,
    googlePay: false,
  })
  let mountGeneration = 0

  const destroy = () => {
    mountGeneration += 1
    expressCheckoutElement.value?.unmount()
    expressCheckoutElement.value = null
    elements.value = null
    stripe.value = null
    isReady.value = false
    hasResolvedAvailability.value = false
    availablePaymentMethods.value = {
      applePay: false,
      googlePay: false,
    }
  }

  const mount = async (
    container: HTMLElement,
    session: StripeExpressCheckoutMountSession,
    handlers: StripeExpressCheckoutEventHandlers = {},
  ) => {
    if (!session.publishableKey) {
      throw new Error('Stripe Express Checkout publishable key is missing')
    }
    if (!session.amount || session.amount <= 0) {
      throw new Error('Stripe Express Checkout amount must be greater than zero')
    }

    destroy()
    const currentMountGeneration = ++mountGeneration

    const loadedStripe = await loadStripe(session.publishableKey)
    if (!loadedStripe) {
      throw new Error('Unable to load Stripe Express Checkout')
    }
    if (currentMountGeneration !== mountGeneration) return

    const loadedElements = loadedStripe.elements({
      mode: 'payment',
      amount: convertMajorAmountToStripeMinorAmount(session.amount, session.currency),
      currency: normalizeCurrencyCode(session.currency).toLowerCase(),
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
      },
    })

    const mountedElement = loadedElements.create(
      'expressCheckout',
      buildStripeExpressCheckoutElementOptions(session),
    )
    if (currentMountGeneration !== mountGeneration) {
      mountedElement.destroy()
      return
    }

    mountedElement.on('ready', (event) => {
      isReady.value = true
      hasResolvedAvailability.value = true
      availablePaymentMethods.value = toExpressCheckoutAvailablePaymentMethods(event.availablePaymentMethods)
      handlers.onReady?.(event)
    })
    mountedElement.on('availablepaymentmethodschange', (event) => {
      hasResolvedAvailability.value = true
      availablePaymentMethods.value = toExpressCheckoutAvailablePaymentMethods(event.paymentMethods)
      handlers.onAvailablePaymentMethodsChange?.(event)
    })
    mountedElement.on('click', (event) => {
      handlers.onClick?.(event)
    })
    mountedElement.on('confirm', (event) => {
      void handlers.onConfirm?.(event)
    })
    mountedElement.on('shippingaddresschange', (event) => {
      void handlers.onShippingAddressChange?.(event)
    })
    mountedElement.on('shippingratechange', (event) => {
      void handlers.onShippingRateChange?.(event)
    })
    mountedElement.on('cancel', () => {
      handlers.onCancel?.()
    })
    mountedElement.on('loaderror', (event) => {
      handlers.onError?.(event.error)
    })

    mountedElement.mount(container)
    stripe.value = loadedStripe
    elements.value = loadedElements
    expressCheckoutElement.value = mountedElement
  }

  const submit = async () => {
    if (!elements.value) {
      throw new Error('Stripe Express Checkout is not ready')
    }

    const result = await elements.value.submit()
    if (result.error) {
      throw new Error(result.error.message || 'Stripe payment details could not be submitted')
    }
  }

  const confirmPayment = async (
    clientSecret: string,
    returnUrl: string,
  ): Promise<StripeExpressCheckoutConfirmationResult> => {
    if (!stripe.value || !elements.value) {
      throw new Error('Stripe Express Checkout is not ready')
    }
    if (!clientSecret) {
      throw new Error('Stripe PaymentIntent client secret is missing')
    }

    const result = await stripe.value.confirmPayment({
      elements: elements.value,
      clientSecret,
      confirmParams: {
        return_url: returnUrl,
      },
      redirect: 'if_required',
    })

    if (result.error) {
      throw new Error(result.error.message || 'Stripe Express Checkout could not be confirmed')
    }
    if (!result.paymentIntent) {
      throw new Error('Stripe did not return an Express Checkout payment result')
    }

    const paymentIntent = result.paymentIntent
    return {
      status: paymentIntent.status,
      paymentIntentId: paymentIntent.id,
    }
  }

  return {
    stripe,
    elements,
    expressCheckoutElement,
    isReady,
    hasResolvedAvailability,
    availablePaymentMethods,
    mount,
    submit,
    confirmPayment,
    destroy,
  }
}
