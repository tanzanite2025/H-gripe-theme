import { computed, type Ref } from 'vue'
import { useAsyncData } from '#imports'
import { usePublicApiBase } from '~/composables/usePublicApiBase'

export interface AnalyticsSettings {
  googleAnalytics: string
  googleTagManager: string
}

type RawAnalyticsSettings = {
  google_analytics?: unknown
  google_tag_manager?: unknown
}

export const normalizeAnalyticsSettings = (
  raw: RawAnalyticsSettings | null | undefined,
): AnalyticsSettings => ({
  googleAnalytics: typeof raw?.google_analytics === 'string' ? raw.google_analytics.trim() : '',
  googleTagManager: typeof raw?.google_tag_manager === 'string' ? raw.google_tag_manager.trim() : '',
})

export async function fetchAnalyticsSettings(
  apiBase: string,
  locale = 'en',
): Promise<AnalyticsSettings> {
  if (!apiBase) {
    return { googleAnalytics: '', googleTagManager: '' }
  }

  const result = await $fetch<RawAnalyticsSettings>(`${apiBase}/analytics`, {
    params: { locale },
    headers: { accept: 'application/json' },
  })
  return normalizeAnalyticsSettings(result)
}

export function useAnalyticsSettings(locale: Ref<string> | string = 'en') {
  const apiBase = usePublicApiBase()
  const localeCode = computed(() => typeof locale === 'string' ? locale : locale.value || 'en')

  const { data, pending, error } = useAsyncData<AnalyticsSettings | null>(
    () => `mytheme-analytics-settings-${localeCode.value}`,
    async () => {
      try {
        return await fetchAnalyticsSettings(apiBase.value, localeCode.value)
      } catch (fetchError) {
        console.warn('Failed to load analytics settings:', fetchError)
        return null
      }
    },
    {
      default: () => null,
      watch: [localeCode],
    },
  )

  const analyticsSettings = computed<AnalyticsSettings>(() => data.value ?? {
    googleAnalytics: '',
    googleTagManager: '',
  })

  return { analyticsSettings, pending, error }
}
