import { computed } from 'vue'
import { useAsyncData, useRuntimeConfig } from '#imports'

export interface RuntimeSocialLink {
  network: string
  url: string
}

export interface ApiSocialLink extends RuntimeSocialLink {
  label?: string
  size?: number
}

export interface SiteSettingsResponse {
  siteTitle?: string
  siteDescription?: string
  siteLogo?: string
  socialLinks?: ApiSocialLink[]
}

export interface QuickBuyConfigProp {
  steps?: unknown[]
  storeApiBase?: string
  cartUrl?: string
  checkoutUrl?: string
  taxonomy?: string
  buttonText?: string
  enabled?: boolean
  successMessage?: string
  requireLogin?: boolean
}

type RawSettings = Record<string, unknown>

const normalizeBaseUrl = (value?: string) => (value ? value.replace(/\/$/, '') : '')

const asString = (value: unknown) => {
  if (typeof value === 'string') return value
  if (value === null || value === undefined) return ''
  return String(value)
}

const asBoolean = (value: unknown) => {
  if (typeof value === 'boolean') return value
  if (typeof value === 'string') return value.toLowerCase() === 'true'
  return Boolean(value)
}

const parseArray = (value: unknown) => {
  if (Array.isArray(value)) return value
  if (typeof value !== 'string' || !value.trim()) return []
  try {
    const parsed = JSON.parse(value)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

const isSocialLink = (value: unknown): value is ApiSocialLink => {
  if (!value || typeof value !== 'object') return false
  const record = value as Record<string, unknown>
  return typeof record.network === 'string' && typeof record.url === 'string'
}

export const normalizeRuntimeSocialLinks = (value: unknown): ApiSocialLink[] => {
  return parseArray(value)
    .filter(isSocialLink)
    .map((item) => {
      const size = Number(item.size)
      return {
        network: item.network,
        url: item.url,
        label: item.label,
        size: Number.isFinite(size) && size > 0 ? size : undefined
      }
    })
}

const normalizeSiteSettings = (raw: RawSettings): SiteSettingsResponse => ({
  siteTitle: asString(raw.siteTitle || raw.site_name),
  siteDescription: asString(raw.siteDescription || raw.site_description),
  siteLogo: asString(raw.siteLogo || raw.site_logo),
  socialLinks: normalizeRuntimeSocialLinks(raw.socialLinks || raw.social_links)
})

const normalizeQuickBuySettings = (raw: RawSettings): QuickBuyConfigProp => ({
  enabled: raw.enabled === undefined ? undefined : asBoolean(raw.enabled),
  buttonText: asString(raw.buttonText || raw.button_text),
  successMessage: asString(raw.successMessage || raw.success_message),
  requireLogin: raw.requireLogin === undefined && raw.require_login === undefined
    ? undefined
    : asBoolean(raw.requireLogin ?? raw.require_login),
  steps: parseArray(raw.steps),
  storeApiBase: asString(raw.storeApiBase || raw.store_api_base),
  cartUrl: asString(raw.cartUrl || raw.cart_url),
  checkoutUrl: asString(raw.checkoutUrl || raw.checkout_url),
  taxonomy: asString(raw.taxonomy)
})

export function useSiteSettings() {
  const config = useRuntimeConfig()
  const apiBase = computed(() => normalizeBaseUrl((config.public as { apiBase?: string }).apiBase || '/api/v1'))

  const { data } = useAsyncData<SiteSettingsResponse | null>(
    'mytheme-site-settings',
    async () => {
      if (!apiBase.value) return null
      try {
        const result = await $fetch<RawSettings>(`${apiBase.value}/settings/site`, {
          headers: { accept: 'application/json' }
        })
        return result ? normalizeSiteSettings(result) : null
      } catch (error) {
        console.warn('Failed to load site settings:', error)
        return null
      }
    },
    {
      server: false,
      default: () => null
    }
  )

  const siteSettings = computed<SiteSettingsResponse>(() => data.value ?? {})

  return { siteSettings }
}

export function useQuickBuySettings() {
  const config = useRuntimeConfig()
  const apiBase = computed(() => normalizeBaseUrl((config.public as { apiBase?: string }).apiBase || '/api/v1'))

  const { data } = useAsyncData<QuickBuyConfigProp | null>(
    'mytheme-quick-buy',
    async () => {
      if (!apiBase.value) return null
      try {
        const result = await $fetch<RawSettings>(`${apiBase.value}/settings/quick-buy`, {
          headers: { accept: 'application/json' }
        })
        return result ? normalizeQuickBuySettings(result) : null
      } catch (error) {
        console.warn('Failed to load quick buy config:', error)
        return null
      }
    },
    {
      server: false,
      default: () => null
    }
  )

  const quickBuySettings = computed<QuickBuyConfigProp | null>(() => data.value)

  return { quickBuySettings }
}
