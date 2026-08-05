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

export const useSmartRecommendations = () => {
  const recommendationsLoading = ref(false)
  const recommendedProducts = ref<RecommendationProductCard[]>([])
  const recommendationRequestId = ref('')
  const recommendationAlgorithmVersion = ref('')
  const recommendationSource = ref<RecommendationSource>('empty')
  const { request } = useApiRequest()
  const route = useRoute()
  const { locale } = useI18n()
  const { anonymousId, sessionId, ensureIdentity } = useBehaviorEvents()
  const { fetchPublicShopProducts } = useShopProducts()
  const {
    categories,
    loading: categoriesLoading,
    source: categorySource,
    loadCategories,
  } = useShopCategories()

  const displayedProductCards = computed<RecommendationProductCard[]>(() => {
    const cards = recommendedProducts.value
      .filter((product) => product && product.title && product.url)
      .slice(0, 6)

    if (cards.length || !import.meta.dev) return cards
    return developmentProductFallbacks
  })

  const displayedCategoryCards = computed<ShopCategory[]>(() => {
    const cards = categories.value
      .filter((category) => category && category.slug && category.name)
      .slice(0, 6)

    if (cards.length || !import.meta.dev) return cards
    return developmentCategoryFallbacks
  })

  const applyCatalogFallback = async () => {
    const result = await fetchPublicShopProducts({
      featured: false,
      page_size: 6,
      status: 'active',
    })
    recommendedProducts.value = result.items.slice(0, 6).map((product) => ({
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

  const loadBaselineRecommendations = async () => {
    if (recommendationsLoading.value) return

    recommendationsLoading.value = true
    recommendationRequestId.value = ''
    recommendationAlgorithmVersion.value = ''
    recommendationSource.value = 'empty'
    try {
      ensureIdentity()
      const requestBody: RecommendationRequest = {
        surface: 'shop_search_drawer',
        locale: String(locale.value || 'en'),
        anonymous_id: anonymousId.value || undefined,
        session_id: sessionId.value || undefined,
        context: {
          route: route.fullPath,
        },
        limit: 6,
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
      await applyCatalogFallback()
    } catch (error) {
      // Keep the search drawer useful if the new recommendation endpoint is
      // unavailable during rollout.
      try {
        await applyCatalogFallback()
        recommendationAlgorithmVersion.value = 'public-catalog-fallback'
      } catch (fallbackError) {
        // eslint-disable-next-line no-console
        console.error('Failed to load baseline recommendations:', error, fallbackError)
        recommendedProducts.value = []
        recommendationSource.value = import.meta.dev ? 'development-fallback' : 'empty'
      }
    } finally {
      recommendationsLoading.value = false
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
