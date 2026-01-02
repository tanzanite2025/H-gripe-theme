import { ref, computed, onMounted } from 'vue'
import { useRuntimeConfig, useAsyncData } from '#imports'

export interface RuntimeSocialLink { network: string; url: string }
export interface ApiSocialLink extends RuntimeSocialLink { label?: string; size?: number }
export interface SocialLinkViewModel { network: string; url: string; label: string; size: number }

export function useSocialLinks() {
  const config = useRuntimeConfig()
  const normalizeBaseUrl = (value?: string) => (value ? value.replace(/\/$/, '') : '')
  const wpApiBase = normalizeBaseUrl((config.public as { wpApiBase?: string }).wpApiBase)

  const previewLinks = ref<SocialLinkViewModel[] | null>(null)

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
                previewLinks.value = normalize(arr as any)
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

  interface SiteSettingsResponse { socialLinks?: Array<RuntimeSocialLink | ApiSocialLink> }

  const { data } = useAsyncData<SiteSettingsResponse | null>('mytheme-site-settings-social', async () => {
    if (!wpApiBase) return null
    try {
      const res = await $fetch<SiteSettingsResponse>(`${wpApiBase}/tanzanite/v1/settings`, { headers: { accept: 'application/json' } })
      return res || null
    } catch {
      return null
    }
  }, { server: false, default: () => null })

  const normalize = (items: Array<RuntimeSocialLink | ApiSocialLink>) => {
    return items
      .filter((item): item is RuntimeSocialLink | ApiSocialLink => !!item && typeof item === 'object' && 'network' in item && 'url' in item)
      .map((item) => {
        const network = String(item.network || '').toLowerCase()
        const url = String(item.url || '')
        const label = 'label' in item && item.label ? String(item.label) : network.toUpperCase()
        const size = Number('size' in item && item.size ? item.size : 24) || 24
        return { network, url, label, size } as SocialLinkViewModel
      })
      .filter((x) => x.network && x.url)
  }

  const socialLinks = computed<SocialLinkViewModel[]>(() => {
    if (previewLinks.value && previewLinks.value.length) return previewLinks.value
    const arr = Array.isArray(data.value?.socialLinks) ? data.value!.socialLinks! : []
    return normalize(arr as any)
  })

  return { socialLinks }
}
