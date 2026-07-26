import { ref } from 'vue'

interface PaymentMethodCurrencySource {
  supported_currencies?: string | string[] | null
}

const extractList = <T>(payload: unknown): T[] => {
  let current = payload

  for (let depth = 0; depth < 3; depth += 1) {
    if (Array.isArray(current)) return current as T[]
    if (!current || typeof current !== 'object') return []

    const record = current as Record<string, unknown>
    if (Array.isArray(record.items)) return record.items as T[]
    current = record.data
  }

  return []
}

const splitCurrencyCodes = (value: PaymentMethodCurrencySource['supported_currencies']) => {
  if (Array.isArray(value)) return value
  if (typeof value !== 'string') return []

  return value.split(',')
}

export const usePaymentCurrencies = () => {
  const { request } = useApiRequest()

  const currencies = ref<string[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const loadCurrencies = async () => {
    loading.value = true
    error.value = null

    try {
      const response = await request<unknown>(
        '/payment/methods?enabled=true',
        { headers: { accept: 'application/json' } },
        'Unable to load payment methods'
      )
      const methods = extractList<PaymentMethodCurrencySource>(response)
      const seen = new Set<string>()
      const next: string[] = []

      for (const method of methods) {
        for (const rawCode of splitCurrencyCodes(method.supported_currencies)) {
          const code = String(rawCode || '').trim().toUpperCase()
          if (!/^[A-Z]{3}$/.test(code) || seen.has(code)) continue

          seen.add(code)
          next.push(code)
        }
      }

      currencies.value = next
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unable to load payment currencies'
      currencies.value = []
    } finally {
      loading.value = false
    }
  }

  return {
    currencies,
    loading,
    error,
    loadCurrencies,
  }
}
