import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import type { ShopCategory } from '~/composables/useShopCategories'
import { useShopCategories } from '~/composables/useShopCategories'
import { useShopProducts } from '~/composables/useShopProducts'
import { useApiRequest } from '~/composables/useApiRequest'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import type {
  RecommendationAPIResult,
  RecommendationProductCard,
  RecommendationRequest,
} from '~/types/recommendation'

export type RecommendationSource =
  | 'recommendation-api'
  | 'catalog-fallback'
  | 'development-fallback'
  | 'empty'

export interface RecommendationLoadOptions {
  surface?: string
  limit?: number
  productId?: number | null
  categoryId?: number | null
  query?: string | null
  route?: string | null
  excludeProductIds?: Array<number | null | undefined>
  catalogFallback?: boolean
}

const DEFAULT_RECOMMENDATION_SURFACE = 'shop_search_drawer'
const DEFAULT_RECOMMENDATION_LIMIT = 6
const MAX_RECOMMENDATION_LIMIT = 12

const developmentProductFallbacks: RecommendationProductCard[] = [
  { id: 'fallback-carbon-wheels', title: 'Carbon Wheels', url: '/shop?keyword=Carbon%20Wheels', priceLabel: 'Search' },
  { id: 'fallback-carbon-rim', title: 'Carbon Rim', url: '/shop?keyword=Carbon%20Rim', priceLabel: 'Search' },
  { id: 'fallback-spoke', title: 'Sapim Spokes', url: '/shop?keyword=Sapim%20Spoke', priceLabel: 'Search' },
  { id: 'fallback-inner-tube', title: 'Inner Tube', url: '/shop?keyword=Inner%20Tube', priceLabel: 'Search' },
  { id: 'fallback-hub', title: 'Hub Parts', url: '/shop?keyword=Hub', priceLabel: 'Search' },
]

const developmentCategoryFallbacks: ShopCategory[] = [
  { id: -1, slug: 'carbon-wheels', name: 'Carbon Wheels' },
  { id: -2, slug: 'carbon-rim', name: 'Carbon Rim' },
  { id: -3, slug: 'spoke', name: 'Spokes' },
  { id: -4, slug: 'inner-tube', name: 'Inner Tube' },
  { id: -5, slug: 'hub', name: 'Hubs' },
]

const toPositiveInteger = (value: unknown) => {
  const numberValue = Number(value)
  if (!Number.isInteger(numberValue) || numberValue <= 0) return null
  return numberValue
}

const normalizeRecommendationLimit = (value: unknown) => {
  const limit = toPositiveInteger(value) || DEFAULT_RECOMMENDATION_LIMIT
  return Math.min(Math.max(limit, 1), MAX_RECOMMENDATION_LIMIT)
}

const normalizeRecommendationSurface = (value: unknown) => {
  const surface = String(value || '').trim()
  return (surface || DEFAULT_RECOMMENDATION_SURFACE).slice(0, 64)
}

const normalizeExcludedProductIds = (values?: Array<number | null | undefined>) => {
  const seen = new Set<number>()
  for (const value of values || []) {
    const productId = toPositiveInteger(value)
    if (productId) seen.add(productId)
  }
  return Array.from(seen)
}

export const useSmartRecommendations = () => {
  const recommendationsLoading = ref(false)
  const recommendedProducts = ref<RecommendationProductCard[]>([])
  const recommendationRequestId = ref('')
  const recommendationAlgorithmVersion = ref('')
  const recommendationSource = ref<RecommendationSource>('empty')
  const recommendationDisplayLimit = ref(DEFAULT_RECOMMENDATION_LIMIT)
  const { request } = useApiRequest()
  const route = useRoute()
  const { locale } = useI18n()
  const { anonymousId, sessionId, ensureIdentity } = useBehaviorEvents()
  const { fetchPublicShopProducts } = useShopProducts()
  let activeRecommendationLoadId = 0
  const {
    categories,
    loading: categoriesLoading,
    source: categorySource,
    loadCategories,
  } = useShopCategories()

  const displayedProductCards = computed<RecommendationProductCard[]>(() => {
    const cards = recommendedProducts.value
      .filter((product) => product && product.title && product.url)
      .slice(0, recommendationDisplayLimit.value)

    if (cards.length || !import.meta.dev) return cards
    return developmentProductFallbacks.slice(0, recommendationDisplayLimit.value)
  })

  const displayedCategoryCards = computed<ShopCategory[]>(() => {
    const cards = categories.value
      .filter((category) => category && category.slug && category.name)
      .slice(0, 6)

    if (cards.length || !import.meta.dev) return cards
    return developmentCategoryFallbacks
  })

  const applyCatalogFallback = async (
    limit = DEFAULT_RECOMMENDATION_LIMIT,
    excludeProductIds: number[] = [],
    loadId = activeRecommendationLoadId
  ) => {
    const excluded = new Set(excludeProductIds)
    const result = await fetchPublicShopProducts({
      featured: false,
      page_size: Math.min(MAX_RECOMMENDATION_LIMIT, limit + excluded.size),
      status: 'active',
    })
    if (loadId !== activeRecommendationLoadId) return

    recommendedProducts.value = result.items
      .filter((product) => !excluded.has(Number(product.id)))
      .slice(0, limit)
      .map((product) => ({
        id: product.id,
        title: product.title,
        url: product.url || `/shop/${product.slug}`,
        thumbnail: product.thumbnail,
        priceLabel: product.priceLabel,
        slot: 'catalog_fallback',
        reason: 'public_catalog_fallback',
      }))
    recommendationSource.value = recommendedProducts.value.length
      ? 'catalog-fallback'
      : import.meta.dev
        ? 'development-fallback'
        : 'empty'
  }

  const loadBaselineRecommendations = async (options: RecommendationLoadOptions = {}) => {
    const loadId = activeRecommendationLoadId + 1
    activeRecommendationLoadId = loadId
    const limit = normalizeRecommendationLimit(options.limit)
    const productId = toPositiveInteger(options.productId)
    const categoryId = toPositiveInteger(options.categoryId)
    const excludeProductIds = normalizeExcludedProductIds([
      ...(options.excludeProductIds || []),
      productId,
    ])
    const context: NonNullable<RecommendationRequest['context']> = {
      route: String(options.route || route.fullPath || '').slice(0, 1024),
    }
    if (productId) context.product_id = productId
    if (categoryId) context.category_id = categoryId
    const query = String(options.query || '').trim()
    if (query) context.query = query.slice(0, 256)

    recommendationsLoading.value = true
    recommendationRequestId.value = ''
    recommendationAlgorithmVersion.value = ''
    recommendationSource.value = 'empty'
    recommendationDisplayLimit.value = limit
    try {
      ensureIdentity()
      const requestBody: RecommendationRequest = {
        surface: normalizeRecommendationSurface(options.surface),
        locale: String(locale.value || 'en'),
        anonymous_id: anonymousId.value || undefined,
        session_id: sessionId.value || undefined,
        context,
        limit,
        exclude_product_ids: excludeProductIds.length ? excludeProductIds : undefined,
      }
      const response = await request<{ data?: RecommendationAPIResult } | RecommendationAPIResult>(
        '/recommendations',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
          },
          body: JSON.stringify(requestBody),
        },
        'Failed to load recommendations'
      )
      const payload = (response as { data?: RecommendationAPIResult })?.data || response as RecommendationAPIResult
      const items = Array.isArray(payload?.items) ? payload.items : []
      if (loadId !== activeRecommendationLoadId) return

      recommendedProducts.value = items
        .filter((item) => Number(item?.product_id) > 0 && item?.title && item?.url)
        .map((item) => ({
          id: Number(item.product_id),
          title: String(item.title),
          url: String(item.url),
          thumbnail: item.thumbnail || undefined,
          priceLabel: item.price_label || undefined,
          slot: item.slot || undefined,
          reason: item.reason || undefined,
        }))
      recommendationRequestId.value = String(payload?.request_id || '')
      recommendationAlgorithmVersion.value = String(payload?.algorithm_version || '')

      if (recommendedProducts.value.length) {
        recommendationSource.value = 'recommendation-api'
        return
      }

      // A successful empty response still gets a public catalog fallback while
      // the recommendation candidate set is being configured.
      if (options.catalogFallback === false) {
        recommendationSource.value = 'empty'
        return
      }
      await applyCatalogFallback(limit, excludeProductIds, loadId)
    } catch (error) {
      if (loadId !== activeRecommendationLoadId) return

      // Keep the search drawer useful if the new recommendation endpoint is
      // unavailable during rollout.
      try {
        if (options.catalogFallback === false) {
          recommendedProducts.value = []
          recommendationSource.value = 'empty'
          return
        }
        await applyCatalogFallback(limit, excludeProductIds, loadId)
        if (loadId !== activeRecommendationLoadId) return
        recommendationAlgorithmVersion.value = 'public-catalog-fallback'
      } catch (fallbackError) {
        if (loadId !== activeRecommendationLoadId) return

        // eslint-disable-next-line no-console
        console.error('Failed to load baseline recommendations:', error, fallbackError)
        recommendedProducts.value = []
        recommendationSource.value = import.meta.dev ? 'development-fallback' : 'empty'
      }
    } finally {
      if (loadId === activeRecommendationLoadId) {
        recommendationsLoading.value = false
      }
    }
  }

  const load = async () => {
    await Promise.all([
      loadCategories(),
      loadBaselineRecommendations(),
    ])
  }

  return {
    displayedProductCards,
    displayedCategoryCards,
    recommendationsLoading,
    categoriesLoading,
    categorySource,
    load,
    loadBaselineRecommendations,
    recommendationRequestId,
    recommendationAlgorithmVersion,
    recommendationSource,
  }
}
