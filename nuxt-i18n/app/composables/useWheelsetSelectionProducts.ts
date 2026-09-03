import { ref, watch, type ComputedRef } from 'vue'
import { useI18n } from '#imports'
import { useShopProducts, type ShopProduct } from '~/composables/useShopProducts'
import type { WheelsetSelectionProductQuery } from '~/types/wheelsetSelectionAssistant'
import { toUserFacingApiError } from '~/utils/storefrontApiFailures'

export const useWheelsetSelectionProducts = (
  query: ComputedRef<WheelsetSelectionProductQuery>,
  enabled?: ComputedRef<boolean>,
) => {
  const { t } = useI18n()
  const { fetchShopProducts } = useShopProducts()
  const products = ref<ShopProduct[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const page = ref(1)
  const total = ref(0)
  const totalIsExact = ref(false)
  const hasMore = ref(false)
  let requestSequence = 0
  let activeRequestController: AbortController | null = null

  const load = async () => {
    if (enabled && !enabled.value) return

    const requestId = ++requestSequence
    activeRequestController?.abort()
    const requestController = new AbortController()
    activeRequestController = requestController
    loading.value = true
    error.value = null
    try {
      const current = query.value
      const result = await fetchShopProducts({
        product_category: current.category_slug,
        keyword: current.keyword,
        attributes: Object.keys(current.spec_filters).length
          ? JSON.stringify(current.spec_filters)
          : undefined,
        page: page.value,
        per_page: 6,
        compact: true,
      }, {
        signal: requestController.signal,
      })
      if (requestId !== requestSequence) return
      products.value = result.items
      const reportedTotal = Number(result.total)
      const reportedPage = Math.max(1, Number(result.page) || page.value)
      const reportedPageSize = Math.max(1, Number(result.pageSize) || 6)
      const knownResultCount = Math.max(0, (reportedPage - 1) * reportedPageSize + result.items.length)
      const hasReportedTotal = reportedTotal > 0
      const nextPageAvailable = Boolean(result.hasMore)
      total.value = hasReportedTotal ? reportedTotal : knownResultCount
      totalIsExact.value = hasReportedTotal || !nextPageAvailable
      hasMore.value = nextPageAvailable
    } catch (cause: any) {
      if (requestId !== requestSequence) return
      if (cause?.name === 'AbortError') return
      products.value = []
      total.value = 0
      totalIsExact.value = false
      hasMore.value = false
      error.value = toUserFacingApiError(
        cause,
        t(
          'wheelsetSelectionAssistant.results.unavailable',
          'Matching products are temporarily unavailable. Please try again.',
        ),
      )
    } finally {
      if (requestId === requestSequence) loading.value = false
      if (activeRequestController === requestController) {
        activeRequestController = null
      }
    }
  }

  const setPage = (nextPage: number) => {
    page.value = Math.max(1, nextPage)
  }

  watch([query, enabled || (() => true)], () => {
    if (enabled && !enabled.value) return
    if (page.value !== 1) {
      page.value = 1
      return
    }
    void load()
  }, { deep: true, immediate: true })
  watch(page, () => {
    void load()
  })

  return {
    products,
    loading,
    error,
    page,
    total,
    totalIsExact,
    hasMore,
    setPage,
    reload: load,
  }
}
