import { ref, computed, onMounted } from 'vue'
import {
  normalizeRuntimeSocialLinks,
  useSiteSettings
} from '~/composables/usePublicSettings'

export interface SocialLinkViewModel { network: string; url: string; label: string; size: number }

export function useSocialLinks() {
  const previewLinks = ref<SocialLinkViewModel[] | null>(null)
  const { siteSettings } = useSiteSettings()

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

  const socialLinks = computed<SocialLinkViewModel[]>(() => {
    if (previewLinks.value && previewLinks.value.length) return previewLinks.value
    if (!siteSettings.value.socialLinks) throw new Error("[CRITICAL] socialLinks missing");
    return normalize(siteSettings.value.socialLinks)
  })

  return { socialLinks }
}
