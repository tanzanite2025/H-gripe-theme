import { computed } from 'vue'
import { useAsyncData, useCookie, useRequestHeaders } from '#imports'
import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'

export interface StorefrontContextPayload {
  country?: { code?: string; source?: string }
  market?: {
    code?: string
    default_locale?: string
    supported_locales?: string[]
    default_currency?: string
    display_currencies?: string[]
  }
  locale?: { requested?: string; resolved?: string; fallback?: string; source?: string }
  currency?: { requested?: string; resolved?: string; base?: string; source?: string }
}

type StorefrontContextResponse = StorefrontContextPayload | { data?: StorefrontContextPayload }

const normalizeCurrencyCode = (value: unknown) => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

export function useStorefrontContext() {
  const apiBase = useApiBase()
  const displayCurrencyCookie = useCookie<string | null>('display_currency', {
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 365,
  })
  const requestHeaders = import.meta.server ? useRequestHeaders(['accept-language', 'cookie']) : {}

  const { data, pending, error, refresh } = useAsyncData<StorefrontContextPayload | null>(
    'storefront-market-context',
    async () => {
      try {
        const response = await $fetch<StorefrontContextResponse>(`${apiBase.value}/storefront/context`, {
          headers: {
            accept: 'application/json',
            ...requestHeaders,
          },
        })
        const context = 'data' in Object(response) ? (response as { data?: StorefrontContextPayload }).data : response as StorefrontContextPayload
        // SSR HTML may be cached by URL. Do not emit a Set-Cookie while rendering
        // a cached page; the client can persist the resolved preference instead.
        if (import.meta.client && context?.currency?.resolved) {
          displayCurrencyCookie.value = context.currency.resolved
        }
        return context || null
      } catch (err) {
        console.warn('Failed to load storefront context:', err)
        return null
      }
    },
    {
      default: () => null,
    },
  )

  const countryCode = computed(() => String(data.value?.country?.code || 'ZZ').toUpperCase())
  const marketCode = computed(() => String(data.value?.market?.code || 'GLOBAL').toUpperCase())
  const resolvedLocale = computed(() => normalizeStorefrontLocaleCode(data.value?.locale?.resolved) || 'en')
  const baseCurrency = computed(() => normalizeCurrencyCode(data.value?.currency?.base) || normalizeCurrencyCode(data.value?.market?.default_currency) || 'USD')
  const displayCurrency = computed(() => normalizeCurrencyCode(data.value?.currency?.resolved) || normalizeCurrencyCode(displayCurrencyCookie.value) || baseCurrency.value)
  const displayCurrencies = computed(() => {
    const values = data.value?.market?.display_currencies || []
    return values.map(normalizeCurrencyCode).filter(Boolean)
  })

  const setDisplayCurrency = async (currency: string) => {
    const code = normalizeCurrencyCode(currency)
    if (!code) return
    displayCurrencyCookie.value = code
    await refresh()
  }

  return {
    context: data,
    pending,
    error,
    refresh,
    countryCode,
    marketCode,
    resolvedLocale,
    baseCurrency,
    displayCurrency,
    displayCurrencies,
    setDisplayCurrency,
  }
}
