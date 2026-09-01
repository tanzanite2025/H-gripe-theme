import { computed, useAsyncData } from '#imports'
import type { Ref } from 'vue'
import { useApiRequest } from '~/composables/useApiRequest'
import { WEBSITE_NAME_DEFAULTS } from '~/data/websiteNameDefaults.generated'
import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'

export interface WebsiteNameSettings {
  locale: string
  status: string
  intro: string
  eyebrow: string
  title: string
  body: string
  note: string
}

type RawWebsiteNameSettings = Record<string, unknown>

const asString = (value: unknown, fallback = '') => {
  const result = typeof value === 'string' ? value : value === null || value === undefined ? '' : String(value)
  return result.trim() || fallback
}

export const defaultWebsiteNameSettings = (locale: unknown): WebsiteNameSettings => {
  const normalizedLocale = normalizeStorefrontLocaleCode(locale)
  const defaultLocale = normalizedLocale === 'zh_cn' ? 'zh_cn' : 'en'

  return {
    locale: normalizedLocale || defaultLocale,
    ...WEBSITE_NAME_DEFAULTS[defaultLocale],
  }
}

const normalizeWebsiteNameSettings = (
  raw: RawWebsiteNameSettings | null | undefined,
  locale: unknown,
): WebsiteNameSettings => {
  const fallback = defaultWebsiteNameSettings(locale)

  return {
    locale: normalizeStorefrontLocaleCode(raw?.locale) || fallback.locale,
    status: asString(raw?.status, fallback.status),
    intro: asString(raw?.intro, fallback.intro),
    eyebrow: asString(raw?.eyebrow, fallback.eyebrow),
    title: asString(raw?.title, fallback.title),
    body: asString(raw?.body, fallback.body),
    note: asString(raw?.note, fallback.note),
  }
}

export function useWebsiteNameSettings(locale: Ref<string> | string) {
  const { request } = useApiRequest()
  const localeValue = computed(() => String(typeof locale === 'string' ? locale : locale.value || 'en'))

  const { data, pending, error } = useAsyncData<WebsiteNameSettings>(
    () => `mytheme-website-name-${localeValue.value}`,
    async () => {
      const fallback = defaultWebsiteNameSettings(localeValue.value)

      try {
        const raw = await request<RawWebsiteNameSettings>(
          '/settings/website-name',
          {
            query: { locale: localeValue.value },
            headers: { accept: 'application/json' },
          },
          'Failed to load website name settings',
        )
        return normalizeWebsiteNameSettings(raw, localeValue.value)
      } catch (fetchError) {
        console.warn('Failed to load why this name settings:', fetchError)
        return fallback
      }
    },
    {
      default: () => defaultWebsiteNameSettings(localeValue.value),
      watch: [localeValue],
    },
  )

  const websiteNameSettings = computed(() =>
    data.value || defaultWebsiteNameSettings(localeValue.value),
  )

  return {
    websiteNameSettings,
    pending,
    error,
  }
}
