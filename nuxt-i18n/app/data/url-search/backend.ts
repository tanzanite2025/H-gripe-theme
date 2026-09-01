import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'
import { useApiRequest } from '~/composables/useApiRequest'
import type { StorefrontURLSearchProfile } from './types'

function getBackendUrlSearchLocale() {
  try {
    const { locale } = useI18n()
    return normalizeStorefrontLocaleCode(locale.value) || 'en'
  } catch {
    return 'en'
  }
}

export async function fetchAllUrlSearchData(): Promise<StorefrontURLSearchProfile[]> {
  try {
    const { request } = useApiRequest()
    const response = await request<{ items?: StorefrontURLSearchProfile[] }>(
      '/storefront/url-search-index',
      {
        query: { locale: getBackendUrlSearchLocale() },
        headers: { accept: 'application/json' },
      },
      'Failed to fetch storefront URL search index',
    )
    if (Array.isArray(response?.items)) return response.items
  } catch (error) {
    console.error('Failed to fetch storefront URL search index from Go backend:', error)
  }

  return []
}
