import { computed, useAsyncData, useI18n, useRuntimeConfig } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import { homeHeroVisualShowcaseFallback } from '~/data/homeHeroVisualShowcaseFallback'
import type {
  HomeHeroVisualShowcaseApiEnvelope,
  HomeHeroVisualShowcaseApiItem,
  HomeHeroVisualShowcaseItem,
} from '~/types/homeHeroVisualShowcase'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'

const HOME_HERO_SHOWCASE_SSR_TIMEOUT_MS = 900

const isAbortError = (error: unknown) => (
  error instanceof DOMException && error.name === 'AbortError'
) || (
  error instanceof Error && error.name === 'AbortError'
)

const numericValue = (value: unknown, fallback: number): number => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

const normalizeShowcaseItem = (
  raw: HomeHeroVisualShowcaseApiItem,
  index: number,
  locale: string,
  mediaContext: ReturnType<typeof createStorefrontMediaContext>,
): HomeHeroVisualShowcaseItem | null => {
  const src = normalizeStorefrontMediaUrl(raw.image_url || raw.thumbnail_url, mediaContext)
  if (!src) return null

  return {
    id: String(raw.id || `api-home-hero-${index + 1}`),
    showcaseKey: String(raw.showcase_key || 'home-hero'),
    locale: String(raw.locale || locale),
    src,
    altText: String(raw.alt_text || raw.title || 'Wheelset manufacturing and inspection'),
    title: String(raw.title || 'Wheelset manufacturing'),
    caption: String(raw.caption || ''),
    width: Math.max(1, numericValue(raw.width, 900)),
    height: Math.max(1, numericValue(raw.height, 1200)),
    desktopOrder: Math.max(1, numericValue(raw.desktop_order, index + 1)),
    targetUrl: String(raw.target_url || '').trim() || undefined,
    targetLabel: String(raw.target_label || '').trim() || undefined,
  }
}

export async function useHomeHeroVisualShowcase() {
  const { locale } = useI18n()
  const { request } = useApiRequest()
  const runtimeConfig = useRuntimeConfig()
  const mediaContext = createStorefrontMediaContext(runtimeConfig)
  const requestKey = computed(() => `home-hero-visual-showcase-${locale.value}`)

  const {
    data,
    pending,
    error,
    refresh,
  } = await useAsyncData<HomeHeroVisualShowcaseApiEnvelope | null>(
    requestKey,
    async () => {
      const abortController = import.meta.server && typeof AbortController !== 'undefined'
        ? new AbortController()
        : null
      const timeoutHandle = abortController
        ? setTimeout(() => abortController.abort(), HOME_HERO_SHOWCASE_SSR_TIMEOUT_MS)
        : null

      try {
        return await request<HomeHeroVisualShowcaseApiEnvelope>(
          '/visual-showcases/home-hero',
          {
            query: { locale: locale.value },
            headers: { accept: 'application/json' },
            ...(abortController ? { signal: abortController.signal } : {}),
          },
          'Failed to load home hero visual showcase',
        )
      } catch (fetchError) {
        if (abortController && isAbortError(fetchError)) {
          return null
        }
        throw fetchError
      } finally {
        if (timeoutHandle) {
          clearTimeout(timeoutHandle)
        }
      }
    },
    { default: () => null },
  )

  const configuredItems = computed(() => (
    Array.isArray(data.value?.data?.items)
      ? data.value.data.items
        .map((item, index) => normalizeShowcaseItem(item, index, locale.value, mediaContext))
        .filter((item): item is HomeHeroVisualShowcaseItem => Boolean(item))
        .sort((left, right) => left.desktopOrder - right.desktopOrder)
      : []
  ))

  const items = computed(() => (
    configuredItems.value.length >= 8 && !error.value
      ? configuredItems.value
      : homeHeroVisualShowcaseFallback
  ))

  const source = computed<'configured' | 'locale-fallback' | 'built-in-fallback' | 'error' | 'loading'>(() => {
    if (pending.value && !data.value) return 'loading'
    if (error.value) return 'error'
    if (configuredItems.value.length >= 8) {
      return data.value?.data?.fallback ? 'locale-fallback' : 'configured'
    }
    return 'built-in-fallback'
  })

  return {
    homeHeroVisualShowcaseItems: items,
    homeHeroVisualShowcaseSource: source,
    homeHeroVisualShowcasePending: pending,
    homeHeroVisualShowcaseError: error,
    refreshHomeHeroVisualShowcase: refresh,
  }
}
