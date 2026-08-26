import { ref, computed, onMounted } from 'vue'
import { useAsyncData } from '#imports'
import { useSiteSettings } from '~/composables/usePublicSettings'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import {
  normalizeSocialLinkItems,
  socialSettingDefinitions,
  type SocialNetwork,
} from '~/utils/socialLinks'

import type { SocialLinkViewModel } from '~/utils/socialLinks'

type RawSocialSettings = Partial<Record<SocialNetwork, unknown>>

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
                previewLinks.value = normalizeSocialLinkItems(arr)
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

  const normalizeConfiguredSettings = (settings: RawSocialSettings) => {
    return normalizeSocialLinkItems(socialSettingDefinitions.map(({ network, label }) => ({
      network,
      url: String(settings[network] ?? '').trim(),
      label,
    })))
  }

  const socialLinks = computed<SocialLinkViewModel[]>(() => {
    if (previewLinks.value && previewLinks.value.length) return previewLinks.value
    if (configuredSocialSettings.value !== null) {
      return normalizeConfiguredSettings(configuredSocialSettings.value)
    }
    return normalizeSocialLinkItems(siteSettings.value.socialLinks || [])
  })

  return { socialLinks }
}
