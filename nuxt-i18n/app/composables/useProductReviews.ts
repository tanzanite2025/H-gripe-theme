import { computed, ref } from 'vue'
import { useI18n } from '#imports'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import type { ShopProductReviewSummary } from '~/composables/useShopProducts'

export interface ProductReview {
  id: number
  productId: number
  rating: number
  title: string
  content: string
  images: string[]
  pros: string
  cons: string
  verified: boolean
  featured: boolean
  helpfulCount: number
  replyContent: string
  createdAt: string
}

export interface ProductReviewPagination {
  page: number
  pageSize: number
  total: number
  totalPage: number
}

interface LoadProductReviewsOptions {
  initialSummary?: ShopProductReviewSummary | null
  pageSize?: number
  refreshSummary?: boolean
}

const toFiniteNumber = (value: unknown, fallback = 0) => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

const parseReviewImages = (value: unknown): string[] => {
  if (Array.isArray(value)) {
    return value.map(item => String(item || '').trim()).filter(Boolean)
  }
  if (typeof value !== 'string' || !value.trim()) return []

  try {
    const parsed = JSON.parse(value)
    return Array.isArray(parsed)
      ? parsed.map(item => String(item || '').trim()).filter(Boolean)
      : []
  } catch {
    return []
  }
}

const normalizeReviewSummary = (value: any): ShopProductReviewSummary | null => {
  if (!value || typeof value !== 'object') return null

  return {
    productId: Math.max(0, Math.floor(toFiniteNumber(value.product_id))),
    totalReviews: Math.max(0, Math.floor(toFiniteNumber(value.total_reviews))),
    averageRating: Math.min(5, Math.max(0, toFiniteNumber(value.average_rating))),
    rating5Count: Math.max(0, Math.floor(toFiniteNumber(value.rating_5_count))),
    rating4Count: Math.max(0, Math.floor(toFiniteNumber(value.rating_4_count))),
    rating3Count: Math.max(0, Math.floor(toFiniteNumber(value.rating_3_count))),
    rating2Count: Math.max(0, Math.floor(toFiniteNumber(value.rating_2_count))),
    rating1Count: Math.max(0, Math.floor(toFiniteNumber(value.rating_1_count))),
  }
}

const normalizeReview = (value: any): ProductReview | null => {
  const id = Math.floor(toFiniteNumber(value?.id))
  if (!id) return null

  return {
    id,
    productId: Math.floor(toFiniteNumber(value?.product_id)),
    rating: Math.min(5, Math.max(1, Math.floor(toFiniteNumber(value?.rating, 1)))),
    title: String(value?.title || '').trim(),
    content: String(value?.content || '').trim(),
    images: parseReviewImages(value?.images),
    pros: String(value?.pros || '').trim(),
    cons: String(value?.cons || '').trim(),
    verified: Boolean(value?.verified),
    featured: Boolean(value?.featured),
    helpfulCount: Math.max(0, Math.floor(toFiniteNumber(value?.helpful_count))),
    replyContent: String(value?.reply_content || '').trim(),
    createdAt: String(value?.created_at || ''),
  }
}

export function useProductReviews() {
  const { locale } = useI18n()
  const publicApiBase = usePublicApiBase()
  const summary = ref<ShopProductReviewSummary | null>(null)
  const reviews = ref<ProductReview[]>([])
  const pagination = ref<ProductReviewPagination>({
    page: 1,
    pageSize: 5,
    total: 0,
    totalPage: 0,
  })
  const isLoading = ref(false)
  const isLoadingMore = ref(false)
  const error = ref('')
  const summaryError = ref('')
  const productId = ref(0)
  let requestSequence = 0

  const requestHeaders = computed(() => {
    const currentLocale = String(locale.value || '').trim()
    return currentLocale ? { 'Accept-Language': currentLocale } : undefined
  })

  const extractData = (response: any) => response?.data ?? response

  const fetchSummary = async (targetProductId: number) => {
    const response = await $fetch<any>(
      `${publicApiBase.value}/reviews/summary/${encodeURIComponent(String(targetProductId))}`,
      { headers: requestHeaders.value },
    )
    const normalized = normalizeReviewSummary(extractData(response))
    if (normalized) summary.value = normalized
  }

  const fetchPage = async (targetProductId: number, page: number, pageSize: number) => {
    const response = await $fetch<any>(`${publicApiBase.value}/reviews`, {
      headers: requestHeaders.value,
      params: {
        product_id: targetProductId,
        page,
        page_size: pageSize,
      },
    })
    const data = response && typeof response === 'object' ? response : {}
    const nextReviews = Array.isArray(data?.data)
      ? data.data.map(normalizeReview).filter(Boolean) as ProductReview[]
      : []
    const rawPagination = data?.pagination || {}

    return {
      reviews: nextReviews,
      pagination: {
        page: Math.max(1, Math.floor(toFiniteNumber(rawPagination.page, page))),
        pageSize: Math.max(1, Math.floor(toFiniteNumber(rawPagination.page_size, pageSize))),
        total: Math.max(0, Math.floor(toFiniteNumber(rawPagination.total))),
        totalPage: Math.max(0, Math.floor(toFiniteNumber(rawPagination.total_page))),
      } satisfies ProductReviewPagination,
    }
  }

  const loadProductReviews = async (
    nextProductId: number,
    options: LoadProductReviewsOptions = {},
  ) => {
    const targetProductId = Math.floor(Number(nextProductId))
    if (!targetProductId) return

    const requestId = requestSequence + 1
    requestSequence = requestId
    const pageSize = Math.max(1, Math.min(20, Math.floor(options.pageSize || 5)))
    productId.value = targetProductId
    summary.value = options.initialSummary ?? null
    reviews.value = []
    pagination.value = { page: 1, pageSize, total: 0, totalPage: 0 }
    error.value = ''
    summaryError.value = ''
    isLoading.value = true

    const summaryTask = options.refreshSummary
      ? fetchSummary(targetProductId).catch((reason: unknown) => {
          if (requestId === requestSequence) {
            summaryError.value = reason instanceof Error ? reason.message : 'Unable to load rating summary'
          }
        })
      : Promise.resolve()

    try {
      const [pageResult] = await Promise.all([
        fetchPage(targetProductId, 1, pageSize),
        summaryTask,
      ])
      if (requestId !== requestSequence) return
      reviews.value = pageResult.reviews
      pagination.value = pageResult.pagination
    } catch (reason: unknown) {
      if (requestId !== requestSequence) return
      error.value = reason instanceof Error ? reason.message : 'Unable to load reviews'
    } finally {
      if (requestId === requestSequence) isLoading.value = false
    }
  }

  const loadMoreProductReviews = async () => {
    if (
      !productId.value
      || isLoading.value
      || isLoadingMore.value
      || !hasMore.value
    ) {
      return
    }

    const requestId = requestSequence
    isLoadingMore.value = true
    try {
      const nextPage = pagination.value.page + 1
      const result = await fetchPage(productId.value, nextPage, pagination.value.pageSize)
      if (requestId !== requestSequence) return
      reviews.value = [...reviews.value, ...result.reviews]
      pagination.value = result.pagination
    } catch (reason: unknown) {
      if (requestId === requestSequence) {
        error.value = reason instanceof Error ? reason.message : 'Unable to load more reviews'
      }
    } finally {
      if (requestId === requestSequence) isLoadingMore.value = false
    }
  }

  const refreshProductReviewSummary = async () => {
    if (!productId.value) return
    try {
      await fetchSummary(productId.value)
      summaryError.value = ''
    } catch (reason: unknown) {
      summaryError.value = reason instanceof Error ? reason.message : 'Unable to load rating summary'
    }
  }

  const hasMore = computed(() => (
    pagination.value.totalPage > 0 && pagination.value.page < pagination.value.totalPage
  ))

  return {
    summary,
    reviews,
    pagination,
    isLoading,
    isLoadingMore,
    error,
    summaryError,
    hasMore,
    loadProductReviews,
    loadMoreProductReviews,
    refreshProductReviewSummary,
  }
}
