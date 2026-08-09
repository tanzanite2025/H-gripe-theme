import { computed } from 'vue'

interface CurrencyPolicyPayload {
  primary_currency?: unknown
  display_currencies?: unknown[]
}

const normalizeCode = (value: unknown) => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

const normalizeCodes = (values: unknown) => {
  const seen = new Set<string>()
  const list = Array.isArray(values) ? values : []
  return list
    .map(normalizeCode)
    .filter(Boolean)
    .filter((code) => {
      if (seen.has(code)) return false
      seen.add(code)
      return true
    })
}

export const useDisplayCurrencies = () => {
  const { request } = useApiRequest()

  const primaryCurrency = useState<string>('display-currency-primary-currency', () => 'USD')
  const displayCurrencies = useState<string[]>('display-currency-list', () => [])
  const loading = useState<boolean>('display-currency-loading', () => false)
  const error = useState<string | null>('display-currency-error', () => null)
  const primaryPricingCurrency = computed(() => primaryCurrency.value || 'USD')
  const primaryDisplayCurrency = primaryPricingCurrency

  const loadCurrencies = async () => {
    loading.value = true
    error.value = null

    try {
      const response = await request<{ data?: CurrencyPolicyPayload }>(
        '/settings/currency-policy',
        { headers: { accept: 'application/json' } },
        'Unable to load display currencies',
      )
      const policy = response?.data || (response as unknown as CurrencyPolicyPayload)
      primaryCurrency.value = normalizeCode(policy?.primary_currency) || 'USD'
      displayCurrencies.value = normalizeCodes(policy?.display_currencies)
        .filter(code => code !== primaryCurrency.value)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unable to load display currencies'
      displayCurrencies.value = []
    } finally {
      loading.value = false
    }
  }

  return {
    primaryCurrency,
    displayCurrencies,
    primaryPricingCurrency,
    primaryDisplayCurrency,
    loading,
    error,
    loadCurrencies,
  }
}
