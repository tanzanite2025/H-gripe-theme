import { ref, watch, type ComputedRef } from 'vue'
import { useI18n } from '#imports'
import { useShopProducts, type ShopProduct } from '~/composables/useShopProducts'
import type { WheelsetSelectionProductQuery } from '~/types/wheelsetSelectionAssistant'
import { toUserFacingApiError } from '~/utils/storefrontApiFailures'

export const useWheelsetSelectionProducts = (
  query: ComputedRef<WheelsetSelectionProductQuery>,
) => {
  const { t } = useI18n()
  const { fetchShopProducts } = useShopProducts()
  const products = ref<ShopProduct[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const page = ref(1)
  const total = ref(0)
  const hasMore = ref(false)
  let requestSequence = 0

  const load = async () => {
    const requestId = ++requestSequence
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
        per_page: 12,
      })
      if (requestId !== requestSequence) return
      products.value = result.items
      const reportedTotal = Number(result.total)
      total.value = reportedTotal > 0 ? reportedTotal : result.items.length
      hasMore.value = Boolean(result.hasMore)
    } catch (cause: any) {
      if (requestId !== requestSequence) return
      products.value = []
      total.value = 0
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
    }
  }

  const setPage = (nextPage: number) => {
    page.value = Math.max(1, nextPage)
  }

  watch(query, () => {
    page.value = 1
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
    hasMore,
    setPage,
    reload: load,
  }
}
