import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import type { ShopCategory } from '~/composables/useShopCategories'
import { useShopCategories } from '~/composables/useShopCategories'
import { useShopProducts, type ShopProductReviewSummary } from '~/composables/useShopProducts'
import { useApiRequest } from '~/composables/useApiRequest'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import type {
  RecommendationAPIResult,
  RecommendationAPIReviewSummary,
  RecommendationProductCard,
  RecommendationRequest,
} from '~/types/recommendation'

export type RecommendationSource =
  | 'recommendation-api'
  | 'catalog-fill'
  | 'empty'

export interface RecommendationLoadOptions {
  surface?: string
  limit?: number
  productId?: number | null
  categoryId?: number | null
  query?: string | null
  route?: string | null
  excludeProductIds?: Array<number | null | undefined>
}

const DEFAULT_RECOMMENDATION_SURFACE = 'shop_search_drawer'
const DEFAULT_RECOMMENDATION_LIMIT = 6
const MAX_RECOMMENDATION_LIMIT = 12
const MIN_RECOMMENDATION_CARDS = 5
const CATALOG_FILL_PAGE_SIZE = 24
const MAX_CATALOG_FILL_PAGES = 8

const toPositiveInteger = (value: unknown) => {
  const numberValue = Number(value)
  if (!Number.isInteger(numberValue) || numberValue <= 0) return null
  return numberValue
}

const normalizeRecommendationLimit = (value: unknown) => {
  const limit = toPositiveInteger(value) || DEFAULT_RECOMMENDATION_LIMIT
  return Math.min(Math.max(limit, MIN_RECOMMENDATION_CARDS), MAX_RECOMMENDATION_LIMIT)
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

const isProductDetailUrl = (value: unknown) => {
  const url = String(value || '').trim().replace(/\/+$/, '')
  return /(?:^|\/)shop\/[^/?#]+$/.test(url)
}

const extractProductSlugFromUrl = (value: unknown) => {
  const url = String(value || '').trim().replace(/\/+$/, '')
  const match = url.match(/(?:^|\/)shop\/([^/?#]+)$/)
  return match?.[1] ? decodeURIComponent(match[1]) : ''
}

const normalizeRecommendationReviewSummary = (
  value: RecommendationAPIReviewSummary | null | undefined,
  fallbackProductId: number,
): ShopProductReviewSummary | null => {
  if (!value || typeof value !== 'object') return null

  const toCount = (count: unknown) => Math.max(0, Math.floor(Number(count) || 0))
  const averageRating = Number(value.average_rating)

  return {
    productId: toPositiveInteger(value.product_id) || fallbackProductId,
    totalReviews: toCount(value.total_reviews),
    averageRating: Number.isFinite(averageRating)
      ? Math.min(5, Math.max(0, averageRating))
      : 0,
    rating5Count: toCount(value.rating_5_count),
    rating4Count: toCount(value.rating_4_count),
    rating3Count: toCount(value.rating_3_count),
    rating2Count: toCount(value.rating_2_count),
    rating1Count: toCount(value.rating_1_count),
  }
}

const createRecommendationFallbackCard = (
  item: RecommendationAPIResult['items'][number],
): RecommendationProductCard | null => {
  const productId = toPositiveInteger(item?.product_id)
  const title = String(item?.title || '').trim()
  const url = String(item?.url || '').trim()
  if (!productId || !title || !isProductDetailUrl(url)) {
    return null
  }

  const priceLabel = String(item?.price_label || '').trim()
  return {
    id: productId,
    productId,
    defaultVariantId: null,
    title,
    slug: extractProductSlugFromUrl(url) || String(productId),
    url,
    thumbnail: item?.thumbnail || undefined,
    priceNumber: 0,
    priceLabel,
    currency: 'USD',
    weightGrams: toPositiveInteger(item?.weight_grams) || undefined,
    displayPriceNumber: 0,
    displayPriceCurrency: 'USD',
    displayPriceLabel: priceLabel,
    displayPrices: [],
    prices: {
      regular: 0,
      sale: 0,
    },
    availability: 'in_stock',
    reviewSummary: normalizeRecommendationReviewSummary(item.review_summary, productId),
    variants: [],
    slot: item?.slot || undefined,
    reason: item?.reason || undefined,
  }
}

const isRecommendationProduct = (product: RecommendationProductCard | null | undefined) => {
  return Boolean(
    product
    && toPositiveInteger(product.id)
    && String(product.title || '').trim()
    && isProductDetailUrl(product.url)
  )
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
      .filter(isRecommendationProduct)
      .slice(0, recommendationDisplayLimit.value)

    return cards
  })

  const displayedCategoryCards = computed<ShopCategory[]>(() => {
    return categories.value
      .filter((category) => category && category.slug && category.name)
      .slice(0, 6)
  })

  const fillRecommendationSlots = async (
    existingCards: RecommendationProductCard[],
    excludeProductIds: number[],
    targetCount: number,
    loadId: number,
  ) => {
    if (existingCards.length >= targetCount) return existingCards

    const seenProductIds = new Set<number>([
      ...excludeProductIds,
      ...existingCards
        .map((product) => toPositiveInteger(product.id))
        .filter((productId): productId is number => Boolean(productId)),
    ])
    const filledCards: RecommendationProductCard[] = []
    let page = 1

    while (
      existingCards.length + filledCards.length < targetCount
      && page <= MAX_CATALOG_FILL_PAGES
    ) {
      const result = await fetchPublicShopProducts({
        page,
        page_size: CATALOG_FILL_PAGE_SIZE,
        status: 'active',
      })
      if (loadId !== activeRecommendationLoadId) return existingCards

      for (const product of result.items) {
        const productId = toPositiveInteger(product.id)
        const title = String(product.title || '').trim()
        const url = String(product.url || '').trim() || (product.slug ? `/shop/${product.slug}` : '')
        if (!productId || seenProductIds.has(productId) || !title || !isProductDetailUrl(url)) {
          continue
        }

        seenProductIds.add(productId)
        filledCards.push({
          ...product,
          slot: 'catalog_fill',
          reason: 'fill_recommendation_slots',
        })

        if (existingCards.length + filledCards.length >= targetCount) break
      }

      if (
        existingCards.length + filledCards.length >= targetCount
        || !result.hasMore
        || result.items.length === 0
      ) {
        break
      }
      page += 1
    }

    return [...existingCards, ...filledCards].slice(0, targetCount)
  }

  const resolveRecommendationCards = (
    items: RecommendationAPIResult['items'],
  ) => {
    return items
      .map(createRecommendationFallbackCard)
      .filter((card): card is RecommendationProductCard => Boolean(card))
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
    recommendedProducts.value = []
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

      const algorithmCards = resolveRecommendationCards(items)
      if (loadId !== activeRecommendationLoadId) return

      recommendedProducts.value = algorithmCards
      recommendationRequestId.value = String(payload?.request_id || '')
      recommendationAlgorithmVersion.value = String(payload?.algorithm_version || '')

      if (recommendedProducts.value.length) {
        if (recommendedProducts.value.length >= MIN_RECOMMENDATION_CARDS) {
          recommendationSource.value = 'recommendation-api'
          return
        }
      }

      const filledCards = await fillRecommendationSlots(
        algorithmCards,
        excludeProductIds,
        MIN_RECOMMENDATION_CARDS,
        loadId,
      )
      if (loadId !== activeRecommendationLoadId) return

      recommendedProducts.value = filledCards
      recommendationSource.value = filledCards.length > algorithmCards.length
        ? 'catalog-fill'
        : algorithmCards.length
          ? 'recommendation-api'
          : 'empty'
    } catch (error) {
      if (loadId !== activeRecommendationLoadId) return

      try {
        const filledCards = await fillRecommendationSlots(
          [],
          excludeProductIds,
          MIN_RECOMMENDATION_CARDS,
          loadId,
        )
        if (loadId !== activeRecommendationLoadId) return

        recommendedProducts.value = filledCards
        recommendationSource.value = filledCards.length ? 'catalog-fill' : 'empty'
      } catch (fallbackError) {
        if (loadId !== activeRecommendationLoadId) return

        // eslint-disable-next-line no-console
        console.error('Failed to load recommendations:', error, fallbackError)
        recommendedProducts.value = []
        recommendationSource.value = 'empty'
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
