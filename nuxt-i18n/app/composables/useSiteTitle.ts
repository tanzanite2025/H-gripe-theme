import { ref, computed, onMounted } from 'vue'
import { useRuntimeConfig } from '#imports'
import { useSiteSettings } from '~/composables/usePublicSettings'

export function useSiteTitle() {
  const config = useRuntimeConfig()

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

  const { siteSettings } = useSiteSettings()

  const siteTitle = computed(() => {
    const fromPreview = previewSiteTitle.value.trim()
    if (fromPreview) return fromPreview
    const fromApi = (siteSettings.value.siteTitle || '').toString().trim()
    if (fromApi) return fromApi
    const fromEnv = ((config.public as { siteTitle?: string }).siteTitle || '').trim()
    return fromEnv || 'Tanzanite'
  })

  return { siteTitle }
}
