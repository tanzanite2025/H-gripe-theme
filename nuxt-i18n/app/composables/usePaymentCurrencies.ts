interface CurrencyPolicyPayload {
  default_order_currency?: string
  default_checkout_currency?: string
}

const normalizeCode = (value: unknown) => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

export const usePaymentCurrencies = () => {
  const { request } = useApiRequest()

  const defaultOrderCurrency = useState<string>('currency-policy-default-order', () => '')
  const loading = useState<boolean>('currency-policy-loading', () => false)
  const error = useState<string | null>('currency-policy-error', () => null)

  const loadCurrencies = async () => {
    loading.value = true
    error.value = null

    try {
      const response = await request<{ data?: CurrencyPolicyPayload }>(
        '/settings/currency-policy',
        { headers: { accept: 'application/json' } },
        'Unable to load currency policy'
      )
      const policy = response?.data || (response as unknown as CurrencyPolicyPayload)
      defaultOrderCurrency.value = normalizeCode(policy?.default_order_currency || policy?.default_checkout_currency)
      if (!defaultOrderCurrency.value) {
        throw new Error('Currency policy is incomplete')
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unable to load default order currency'
      defaultOrderCurrency.value = ''
    } finally {
      loading.value = false
    }
  }

  return {
    defaultOrderCurrency,
    loading,
    error,
    loadCurrencies,
  }
}
