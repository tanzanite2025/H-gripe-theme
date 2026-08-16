import { ref, watch, type ComputedRef } from 'vue'
import { useShopProducts, type ShopProduct } from '~/composables/useShopProducts'
import type { WheelsetSelectionProductQuery } from '~/types/wheelsetSelectionAssistant'

export const useWheelsetSelectionProducts = (
  query: ComputedRef<WheelsetSelectionProductQuery>,
) => {
  const { fetchShopProducts } = useShopProducts()
  const products = ref<ShopProduct[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const page = ref(1)
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
      hasMore.value = Boolean(result.hasMore)
    } catch (cause: any) {
      if (requestId !== requestSequence) return
      products.value = []
      hasMore.value = false
      error.value = cause?.data?.message || cause?.message || 'Unable to load wheelset products.'
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
    hasMore,
    setPage,
    reload: load,
  }
}
