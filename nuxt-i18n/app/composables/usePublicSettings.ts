import { computed } from 'vue'
import { useAsyncData } from '#imports'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import type { QuickBuyConfig } from '~/utils/quickBuy/types'

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
  brandTitle?: string
  siteDescription?: string
  siteLogo?: string
  contactEmail?: string
  contactPhone?: string
  socialLinks?: ApiSocialLink[]
}

export type QuickBuyConfigProp = QuickBuyConfig

type RawSettings = Record<string, unknown>

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

const normalizeSiteSettings = (raw: RawSettings): SiteSettingsResponse => {
  const brandTitle = asString(raw.brandTitle ?? raw.brand_title)
  const legacySiteTitle = asString(raw.siteTitle ?? raw.site_name)
  const siteTitle = brandTitle || legacySiteTitle
  const siteLogo = asString(raw.siteLogo ?? raw.site_logo).trim()

  return {
    siteTitle,
    brandTitle: brandTitle || siteTitle,
    siteDescription: asString(raw.siteDescription ?? raw.site_description),
    siteLogo: siteLogo === '/images/logo.png' ? '' : siteLogo,
    contactEmail: asString(raw.contactEmail ?? raw.contact_email),
    contactPhone: asString(raw.contactPhone ?? raw.contact_phone),
    socialLinks: normalizeRuntimeSocialLinks(raw.socialLinks ?? raw.social_links)
  }
}

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
  const apiBase = usePublicApiBase()

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
      default: () => null
    }
  )

  const siteSettings = computed<SiteSettingsResponse>(() => data.value ?? {})

  return { siteSettings }
}

export function useQuickBuySettings() {
  const apiBase = usePublicApiBase()

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
