import { computed, useAsyncData, useI18n, usePublicApiBase, useRuntimeConfig } from '#imports'
import type {
  HomeMainProductCategoriesApiEnvelope,
  HomeMainProductCategoryApiItem,
  HomeMainProductCategoryItem,
} from '~/types/homeMainProductCategories'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'

const defaultCategoryImageWidth = 1600
const defaultCategoryImageHeight = 900

const numericValue = (value: unknown, fallback: number): number => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

const normalizeCategoryItem = (
  raw: HomeMainProductCategoryApiItem,
  index: number,
  mediaContext: ReturnType<typeof createStorefrontMediaContext>,
): HomeMainProductCategoryItem | null => {
  const src = normalizeStorefrontMediaUrl(raw.image_url || raw.thumbnail_url, mediaContext)
  if (!src) return null

  return {
    id: String(raw.id || `api-home-main-product-category-${index + 1}`),
    src,
    altText: String(raw.alt_text || raw.title || 'Featured store pick'),
    title: String(raw.title || 'Featured store pick'),
    caption: String(raw.caption || ''),
    width: Math.max(1, numericValue(raw.width, defaultCategoryImageWidth)),
    height: Math.max(1, numericValue(raw.height, defaultCategoryImageHeight)),
    desktopOrder: Math.max(1, numericValue(raw.desktop_order, index + 1)),
    targetUrl: String(raw.target_url || '').trim() || undefined,
    targetLabel: String(raw.target_label || '').trim() || undefined,
  }
}

export async function useHomeMainProductCategories() {
  const { locale } = useI18n()
  const apiBase = usePublicApiBase()
  const runtimeConfig = useRuntimeConfig()
  const mediaContext = createStorefrontMediaContext(runtimeConfig)
  const requestKey = computed(() => `home-main-product-categories-${locale.value}`)

  const {
    data,
    pending,
    error,
    refresh,
  } = await useAsyncData<HomeMainProductCategoriesApiEnvelope | null>(
    requestKey,
    () => $fetch<HomeMainProductCategoriesApiEnvelope>(
      `${apiBase.value}/visual-showcases/home-main-product-categories`,
      {
        query: { locale: locale.value },
        headers: { accept: 'application/json' },
      },
    ),
    { default: () => null },
  )

  const configuredItems = computed(() => (
    Array.isArray(data.value?.data?.items)
      ? data.value.data.items
        .map((item, index) => normalizeCategoryItem(item, index, mediaContext))
        .filter((item): item is HomeMainProductCategoryItem => Boolean(item))
        .sort((left, right) => left.desktopOrder - right.desktopOrder)
      : []
  ))

  const configuredCount = computed(() => numericValue(data.value?.data?.configured_count, 0))

  return {
    homeMainProductCategoryItems: configuredItems,
    homeMainProductCategoriesIsConfigured: computed(() => configuredCount.value > 0),
    homeMainProductCategoriesIsLocaleFallback: computed(() => Boolean(data.value?.data?.fallback)),
    homeMainProductCategoriesPending: pending,
    homeMainProductCategoriesError: error,
    refreshHomeMainProductCategories: refresh,
  }
}
