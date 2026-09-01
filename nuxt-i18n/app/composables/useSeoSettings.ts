import { computed, useAsyncData, useI18n } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'

export interface SeoSettings {
  metaTitle: string
  metaDescription: string
}

type RawSeoSettings = {
  meta_title?: unknown
  meta_description?: unknown
}

const asString = (value: unknown) => (typeof value === 'string' ? value.trim() : '')

const normalizeSeoSettings = (raw: RawSeoSettings | null | undefined): SeoSettings => ({
  metaTitle: asString(raw?.meta_title),
  metaDescription: asString(raw?.meta_description),
})

export function useSeoSettings() {
  const { locale } = useI18n()
  const { request } = useApiRequest()

  const { data, pending, error } = useAsyncData<SeoSettings | null>(
    () => `mytheme-seo-home-${locale.value || 'en'}`,
    async () => {
      try {
        const result = await request<RawSeoSettings>('/seo/home', {
          params: { locale: locale.value || 'en' },
          headers: { accept: 'application/json' },
        }, 'Failed to load home SEO settings')
        return normalizeSeoSettings(result)
      } catch (fetchError) {
        console.warn('Failed to load home SEO settings:', fetchError)
        return null
      }
    },
    {
      default: () => null,
      watch: [locale],
    },
  )

  const seoSettings = computed<SeoSettings>(() => data.value ?? {
    metaTitle: '',
    metaDescription: '',
  })

  return { seoSettings, pending, error }
}
