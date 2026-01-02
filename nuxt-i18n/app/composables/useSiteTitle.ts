import { ref, computed, onMounted } from 'vue'
import { useRuntimeConfig, useAsyncData } from '#imports'

export function useSiteTitle() {
  const config = useRuntimeConfig()

  const normalizeBaseUrl = (value?: string) => (value ? value.replace(/\/$/, '') : '')
  const wpApiBase = normalizeBaseUrl((config.public as { wpApiBase?: string }).wpApiBase)

  const previewSiteTitle = ref('')

  if (import.meta.client) {
    onMounted(() => {
      const globalObject = window as unknown as {
        wp?: { customize?: (id: string, cb: (setting: { get?: () => unknown; bind?: (fn: (v: unknown) => void) => void }) => void) => void }
      }
      const customize = globalObject.wp?.customize
      if (typeof customize === 'function') {
        customize('blogname', (setting) => {
          const apply = (v: unknown) => { if (typeof v === 'string') previewSiteTitle.value = v }
          if (typeof setting?.get === 'function') apply(setting.get())
          if (typeof setting?.bind === 'function') setting.bind((v) => apply(v))
        })
      }
    })
  }

  interface SiteSettingsResponse { siteTitle?: string }

  const { data } = useAsyncData<SiteSettingsResponse | null>('mytheme-site-settings-title', async () => {
    if (!wpApiBase) return null
    try {
      const res = await $fetch<SiteSettingsResponse>(`${wpApiBase}/tanzanite/v1/settings`, { headers: { accept: 'application/json' } })
      return res || null
    } catch {
      return null
    }
  }, { server: false, default: () => null })

  const siteTitle = computed(() => {
    const fromPreview = previewSiteTitle.value.trim()
    if (fromPreview) return fromPreview
    const fromApi = (data.value?.siteTitle || '').toString().trim()
    if (fromApi) return fromApi
    const fromEnv = ((config.public as { siteTitle?: string }).siteTitle || '').trim()
    return fromEnv || 'Tanzanite'
  })

  return { siteTitle }
}
