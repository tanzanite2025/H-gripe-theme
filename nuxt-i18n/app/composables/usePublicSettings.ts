import { computed } from 'vue'
import { useAsyncData, useRuntimeConfig } from '#imports'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'

export interface RuntimeSocialLink {
  network: string
  url: string
}

export interface ApiSocialLink extends RuntimeSocialLink {
  label?: string
}

export interface SiteSettingsResponse {
  siteTitle?: string
  siteDescription?: string
  siteLogo?: string
  siteLogoWidth?: number
  siteLogoHeight?: number
  siteFavicon?: string
  contactEmail?: string
  contactPhone?: string
  socialLinks?: ApiSocialLink[]
}

type RawSettings = Record<string, unknown>

const asString = (value: unknown) => {
  if (typeof value === 'string') return value
  if (value === null || value === undefined) return ''
  return String(value)
}

const asPositiveInt = (value: unknown) => {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? Math.round(parsed) : undefined
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
    .map((item) => ({
      network: item.network,
      url: item.url,
      label: item.label,
    }))
}

const normalizeSiteSettings = (
  raw: RawSettings,
  mediaContext: ReturnType<typeof createStorefrontMediaContext>,
): SiteSettingsResponse => {
  const siteTitle = asString(raw.site_name).trim()
  const siteLogo = normalizeStorefrontMediaUrl(
    asString(raw.site_logo).trim(),
    mediaContext,
  )
  const siteFavicon = normalizeStorefrontMediaUrl(
    asString(raw.siteFavicon ?? raw.site_favicon).trim(),
    mediaContext,
  )

  return {
    siteTitle,
    siteDescription: asString(raw.siteDescription ?? raw.site_description),
    siteLogo: siteLogo === '/images/logo.png' ? '' : siteLogo,
    siteLogoWidth: asPositiveInt(raw.siteLogoWidth ?? raw.site_logo_width),
    siteLogoHeight: asPositiveInt(raw.siteLogoHeight ?? raw.site_logo_height),
    siteFavicon,
    contactEmail: asString(raw.contactEmail ?? raw.contact_email),
    contactPhone: asString(raw.contactPhone ?? raw.contact_phone),
    socialLinks: normalizeRuntimeSocialLinks(raw.socialLinks ?? raw.social_links)
  }
}

export function useSiteSettings() {
  const runtimeConfig = useRuntimeConfig()
  const apiBase = usePublicApiBase()
  const mediaContext = createStorefrontMediaContext(runtimeConfig)

  const { data } = useAsyncData<SiteSettingsResponse | null>(
    'mytheme-site-settings',
    async () => {
      if (!apiBase.value) return null
      try {
        const result = await $fetch<RawSettings>(`${apiBase.value}/settings/site`, {
          headers: { accept: 'application/json' }
        })
        return result ? normalizeSiteSettings(result, mediaContext) : null
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
