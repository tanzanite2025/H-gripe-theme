import { ref, computed, onMounted } from 'vue'
import { useAsyncData } from '#imports'
import {
  normalizeRuntimeSocialLinks,
  useSiteSettings
} from '~/composables/usePublicSettings'
import { usePublicApiBase } from '~/composables/usePublicApiBase'

export interface SocialLinkViewModel { network: string; url: string; label: string; size: number }

type RawSocialSettings = Partial<Record<'facebook' | 'twitter' | 'instagram' | 'linkedin' | 'youtube' | 'wechat', unknown>>

const socialSettingDefinitions = [
  { network: 'facebook', label: 'Facebook' },
  { network: 'twitter', label: 'Twitter' },
  { network: 'instagram', label: 'Instagram' },
  { network: 'linkedin', label: 'LinkedIn' },
  { network: 'youtube', label: 'YouTube' },
  { network: 'wechat', label: 'WeChat' },
] as const

export function useSocialLinks() {
  const previewLinks = ref<SocialLinkViewModel[] | null>(null)
  const { siteSettings } = useSiteSettings()
  const apiBase = usePublicApiBase()
  const { data: configuredSocialSettings } = useAsyncData<RawSocialSettings | null>(
    'mytheme-social-settings',
    async () => {
      if (!apiBase.value) return null
      try {
        return await $fetch<RawSocialSettings>(`${apiBase.value}/settings/social`, {
          query: { locale: 'en' },
          headers: { accept: 'application/json' }
        })
      } catch (error) {
        console.warn('Failed to load social settings:', error)
        return null
      }
    },
    {
      default: () => null
    }
  )

  if (import.meta.client) {
    onMounted(() => {
      const globalObject = window as unknown as {
        wp?: { customize?: (id: string, cb: (setting: { get?: () => unknown; bind?: (fn: (v: unknown) => void) => void }) => void) => void }
      }
      const customize = globalObject.wp?.customize
      if (typeof customize === 'function') {
        // Try common setting id used by theme customizer for social links
        const ids = ['mytheme_social_links', 'social_links']
        ids.forEach((id) => {
          customize(id, (setting) => {
            const apply = (v: unknown) => {
              try {
                const arr = Array.isArray(v) ? v : typeof v === 'string' ? JSON.parse(v) : []
                previewLinks.value = normalize(arr)
              } catch {
                previewLinks.value = null
              }
            }
            if (typeof setting?.get === 'function') apply(setting.get())
            if (typeof setting?.bind === 'function') setting.bind((v) => apply(v))
          })
        })
      }
    })
  }

  const normalize = (items: unknown) => {
    return normalizeRuntimeSocialLinks(items)
      .map((item) => {
        const network = String(item.network || '').toLowerCase()
        const url = String(item.url || '')
        const label = 'label' in item && item.label ? String(item.label) : network.toUpperCase()
        const size = Number('size' in item && item.size ? item.size : 24) || 24
        return { network, url, label, size } as SocialLinkViewModel
      })
      .filter((x) => x.network && x.url)
  }

  const normalizeConfiguredSettings = (settings: RawSocialSettings) => {
    return socialSettingDefinitions
      .map(({ network, label }) => ({
        network,
        url: String(settings[network] ?? '').trim(),
        label,
        size: 24
      }))
      .filter((item) => item.url)
  }

  const socialLinks = computed<SocialLinkViewModel[]>(() => {
    if (previewLinks.value && previewLinks.value.length) return previewLinks.value
    if (configuredSocialSettings.value !== null) {
      return normalizeConfiguredSettings(configuredSocialSettings.value)
    }
    return normalize(siteSettings.value.socialLinks || [])
  })

  return { socialLinks }
}
