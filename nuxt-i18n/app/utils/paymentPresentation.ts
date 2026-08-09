import type { CheckoutPaymentOption } from '~/types/payment'

export type StorefrontPaymentMethod = 'card' | 'paypal' | 'alipay' | 'wechat'

export interface PaymentLogoAsset {
  src: string
  alt: string
  className?: string
}

export interface PaymentPresentation {
  method: StorefrontPaymentMethod
  titleKey: string
  title: string
  descriptionKey: string
  description: string
  ctaKey: string
  cta: string
  logos: PaymentLogoAsset[]
}

export const storefrontPaymentMethodOrder: StorefrontPaymentMethod[] = ['card', 'paypal', 'alipay', 'wechat']

export const normalizeStorefrontPaymentMethod = (value?: string | null): StorefrontPaymentMethod | '' => {
  const method = String(value || '').trim().toLowerCase()
  if (['stripe', 'card', 'credit_card', 'credit-card', 'cards'].includes(method)) return 'card'
  if (method === 'paypal') return 'paypal'
  if (method === 'alipay') return 'alipay'
  if (['wechat', 'wechatpay', 'wechat_pay', 'we-chat-pay'].includes(method)) return 'wechat'
  return ''
}

export const paymentMethodFromOption = (option: CheckoutPaymentOption): StorefrontPaymentMethod | '' => {
  const keys = [option.id, option.code, option.provider]
    .map(value => String(value || '').trim().toLowerCase())
    .filter(Boolean)

  for (const key of keys) {
    const method = normalizeStorefrontPaymentMethod(key)
    if (method) return method
  }

  return ''
}

export const paymentPresentation = (method: StorefrontPaymentMethod): PaymentPresentation => {
  switch (method) {
    case 'paypal':
      return {
        method,
        titleKey: 'checkout.payment.paypal.optionTitle',
        title: 'PayPal',
        descriptionKey: 'checkout.payment.paypal.description',
        description: 'Pay with a PayPal account or supported wallet.',
        ctaKey: 'checkout.payment.paypal.cta',
        cta: 'Continue to PayPal',
        logos: [{ src: '/icons/payment/paypal.svg', alt: 'PayPal' }],
      }
    case 'alipay':
      return {
        method,
        titleKey: 'checkout.payment.alipay.optionTitle',
        title: 'Alipay',
        descriptionKey: 'checkout.payment.alipay.description',
        description: 'Pay through Alipay.',
        ctaKey: 'checkout.payment.alipay.cta',
        cta: 'Continue to Alipay',
        logos: [{ src: '/icons/payment/alipay.svg?v=6', alt: 'Alipay', className: 'payment-logo--alipay' }],
      }
    case 'wechat':
      return {
        method,
        titleKey: 'checkout.payment.wechat.optionTitle',
        title: 'WeChat Pay',
        descriptionKey: 'checkout.payment.wechat.description',
        description: 'Scan a WeChat Pay QR code to complete payment.',
        ctaKey: 'checkout.payment.wechat.cta',
        cta: 'Continue to WeChat Pay',
        logos: [{ src: '/icons/payment/wechatpay.svg', alt: 'WeChat Pay' }],
      }
    default:
      return {
        method: 'card',
        titleKey: 'checkout.payment.card.title',
        title: 'Credit / Debit cards',
        descriptionKey: 'checkout.payment.card.description',
        description: 'Secure card checkout powered by Stripe.',
        ctaKey: 'checkout.payment.card.cta',
        cta: 'Continue to card checkout',
        logos: [
          { src: '/icons/payment/visa.svg', alt: 'Visa' },
          { src: '/icons/payment/mastercard.svg', alt: 'Mastercard' },
          { src: '/icons/payment/amex.svg', alt: 'American Express' },
        ],
      }
  }
}

export const isPaymentOptionAvailable = (option: CheckoutPaymentOption) =>
  option.enabled !== false && option.available === true
