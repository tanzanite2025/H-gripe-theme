import { computed, useAsyncData, useI18n, useRuntimeConfig } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import {
  HOME_HERO_VISUAL_SHOWCASE_IMAGE_DIMENSION,
  HOME_HERO_VISUAL_SHOWCASE_MAXIMUM_ITEM_COUNT,
  type HomeHeroVisualShowcaseApiEnvelope,
  type HomeHeroVisualShowcaseApiItem,
  type HomeHeroVisualShowcaseItem,
} from '~/types/homeHeroVisualShowcase'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'
import { HOME_HERO_SHOWCASE_SSR_TIMEOUT_MS } from '~/utils/storefrontLoadingPolicy'

const isAbortError = (error: unknown) => (
  error instanceof DOMException && error.name === 'AbortError'
) || (
  error instanceof Error && error.name === 'AbortError'
)

const numericValue = (value: unknown, fallback: number): number => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

const createEmptyHomeHeroVisualShowcaseSlot = (
  index: number,
  locale: string,
): HomeHeroVisualShowcaseItem => ({
  id: `empty-home-hero-slot-${index + 1}`,
  showcaseKey: 'home-hero',
  locale,
  src: '',
  altText: '',
  title: '',
  caption: '',
  width: HOME_HERO_VISUAL_SHOWCASE_IMAGE_DIMENSION,
  height: HOME_HERO_VISUAL_SHOWCASE_IMAGE_DIMENSION,
  desktopOrder: index + 1,
})

const fillHomeHeroVisualShowcaseSlots = (
  configuredItems: HomeHeroVisualShowcaseItem[],
  locale: string,
): HomeHeroVisualShowcaseItem[] => {
  const itemsByDesktopOrder = new Map(
    configuredItems.map((item) => [item.desktopOrder, item]),
  )

  return Array.from(
    { length: HOME_HERO_VISUAL_SHOWCASE_MAXIMUM_ITEM_COUNT },
    (_, index) => itemsByDesktopOrder.get(index + 1) ?? createEmptyHomeHeroVisualShowcaseSlot(index, locale),
  )
}

const normalizeShowcaseItem = (
  raw: HomeHeroVisualShowcaseApiItem,
  index: number,
  locale: string,
  mediaContext: ReturnType<typeof createStorefrontMediaContext>,
): HomeHeroVisualShowcaseItem | null => {
  const src = normalizeStorefrontMediaUrl(raw.image_url || raw.thumbnail_url, mediaContext)
  if (!src) return null

  const width = numericValue(raw.width, 0)
  const height = numericValue(raw.height, 0)
  if (
    width !== HOME_HERO_VISUAL_SHOWCASE_IMAGE_DIMENSION
    || height !== HOME_HERO_VISUAL_SHOWCASE_IMAGE_DIMENSION
  ) {
    return null
  }

  return {
    id: String(raw.id || `api-home-hero-${index + 1}`),
    showcaseKey: String(raw.showcase_key || 'home-hero'),
    locale: String(raw.locale || locale),
    src,
    altText: String(raw.alt_text || raw.title || 'Wheelset manufacturing and inspection'),
    title: String(raw.title || 'Wheelset manufacturing'),
    caption: String(raw.caption || ''),
    width,
    height,
    desktopOrder: Math.max(1, numericValue(raw.desktop_order, index + 1)),
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

  const configuredItems = computed(() => {
    const apiItems = data.value?.data?.items
    if (
      !Array.isArray(apiItems)
      || data.value?.data?.fallback === true
      || apiItems.length > HOME_HERO_VISUAL_SHOWCASE_MAXIMUM_ITEM_COUNT
    ) {
      return []
    }

    return apiItems
      .map((item, index) => normalizeShowcaseItem(item, index, locale.value, mediaContext))
      .filter((item): item is HomeHeroVisualShowcaseItem => Boolean(item))
      .sort((left, right) => left.desktopOrder - right.desktopOrder)
  })

  const items = computed(() => fillHomeHeroVisualShowcaseSlots(
    error.value ? [] : configuredItems.value,
    locale.value,
  ))

  const source = computed<'configured' | 'empty' | 'error' | 'loading'>(() => {
    if (pending.value && !data.value) return 'loading'
    if (error.value) return 'error'
    return configuredItems.value.length > 0 ? 'configured' : 'empty'
  })

  return {
    homeHeroVisualShowcaseItems: items,
    homeHeroVisualShowcaseSource: source,
    homeHeroVisualShowcasePending: pending,
    homeHeroVisualShowcaseError: error,
    refreshHomeHeroVisualShowcase: refresh,
  }
}
