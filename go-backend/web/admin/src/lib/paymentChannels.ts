export const paymentChannelLabels = {
  stripe: 'Stripe',
  paypal: 'PayPal',
  wechat: '微信支付',
  alipay: '支付宝',
} as const

export type PaymentChannelKey = keyof typeof paymentChannelLabels

export const normalizePaymentChannelKey = (value: string): PaymentChannelKey | '' => {
  const key = String(value || '').trim().toLowerCase()
  return key in paymentChannelLabels ? key as PaymentChannelKey : ''
}

export const getPaymentChannelLabel = (value: string, fallback = ''): string => {
  const key = normalizePaymentChannelKey(value)
  return key ? paymentChannelLabels[key] : fallback
}
