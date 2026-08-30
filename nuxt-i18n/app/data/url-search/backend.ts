import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'
import type { StorefrontURLSearchProfile } from './types'

function getBackendUrlSearchLocale() {
  try {
    const { locale } = useI18n()
    return normalizeStorefrontLocaleCode(locale.value) || 'en'
  } catch {
    return 'en'
  }
}

function getUrlSearchApiBase() {
  const config = useRuntimeConfig()
  const publicApiBase = (config.public as { apiBase?: string }).apiBase || '/api/v1'

  if (import.meta.dev && import.meta.client) {
    return '/api/v1'
  }

  if (import.meta.dev && import.meta.server) {
    const internalApiOrigin = (config as { apiInternalOrigin?: string }).apiInternalOrigin
      ?.replace(/\/$/, '')
    if (internalApiOrigin) return `${internalApiOrigin}/api/v1`
  }

  return publicApiBase
}

export async function fetchAllUrlSearchData(): Promise<StorefrontURLSearchProfile[]> {
  try {
    const response = await $fetch<{ items?: StorefrontURLSearchProfile[] }>(`${getUrlSearchApiBase()}/storefront/url-search-index`, {
      query: { locale: getBackendUrlSearchLocale() },
    })
    if (Array.isArray(response?.items)) return response.items
  } catch (error) {
    console.error('Failed to fetch storefront URL search index from Go backend:', error)
  }

  return []
}
