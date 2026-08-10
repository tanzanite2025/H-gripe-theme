import { useRuntimeConfig, useState } from '#imports'
import { useI18n } from 'vue-i18n'
import { computed, watch } from 'vue'

export interface ShopCategory {
  id: number
  slug: string
  name: string
  image?: string
  count?: number
  isProductType?: boolean
}

export type ShopCategorySource = 'api' | 'empty' | 'error' | 'dev-fallback'

interface ShopCategoryState {
  categories: ShopCategory[]
  loading: boolean
  error: string | null
  source: ShopCategorySource
}

type ShopCategoryStateStore = Record<string, ShopCategoryState>

const fallbackCategories: ShopCategory[] = [
  { id: 1, slug: 'rims', name: 'Rims', isProductType: true },
  { id: 2, slug: 'hubs', name: 'Hubs', isProductType: true },
  { id: 3, slug: 'spokes', name: 'Spokes', isProductType: true },
  { id: 4, slug: 'accessories', name: 'Accessories', isProductType: true },
]

const createEmptyCategoryState = (): ShopCategoryState => ({
  categories: [],
  loading: false,
  error: null,
  source: 'empty',
})

const extractProductTypes = (payload: unknown): ShopCategory[] => {
  let current = payload

  for (let depth = 0; depth < 3; depth += 1) {
    if (Array.isArray(current)) {
      return current.flatMap((item) => {
        if (!item || typeof item !== 'object') return []

        const record = item as Record<string, unknown>
        const id = Number(record.id)
        const slug = String(record.slug || '').trim()
        const name = String(record.name || '').trim()
        const image = String(record.image_url || record.image || '').trim()

        if (!Number.isFinite(id) || !slug || !name || record.is_enabled === false) return []
        return [{
          id,
          slug,
          name,
          ...(image ? { image } : {}),
          isProductType: true,
        }]
      })
    }

    if (!current || typeof current !== 'object') break
    current = (current as Record<string, unknown>).data
  }

  return []
}

export const useShopCategories = () => {
  const config = useRuntimeConfig()
  const { locale } = useI18n()
  // Every storefront surface reads this store so category names, slugs, and images stay aligned.
  const stateStore = useState<ShopCategoryStateStore>('shop-categories-by-locale', () => ({}))
  const baseURL = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
  const localeCode = computed(() => String(locale.value || '').trim() || 'en')
  const stateKey = computed(() => `${baseURL}|${localeCode.value}`)
  const emptyCategoryState = createEmptyCategoryState()

  const getState = (key = stateKey.value): ShopCategoryState => {
    const existing = stateStore.value[key]
    if (existing) return existing

    const next = createEmptyCategoryState()
    stateStore.value[key] = next
    return next
  }

  const currentState = computed(() => stateStore.value[stateKey.value] || emptyCategoryState)
  const categories = computed(() => currentState.value.categories)
  const loading = computed(() => currentState.value.loading)
  const error = computed(() => currentState.value.error)
  const source = computed(() => currentState.value.source)

  const requestCategories = async (requestLocale: string): Promise<ShopCategory[]> => {
    const headers = requestLocale ? { 'Accept-Language': requestLocale } : undefined
    const response = await $fetch<unknown>(`${baseURL}/products/types`, { headers })
    return extractProductTypes(response)
  }

  const applyDevFallback = (state: ShopCategoryState, reason: string) => {
    // 仅本地开发兜底：生产环境必须暴露真实空分类或接口错误，避免示例分类污染商品事实源。
    if (!import.meta.dev) return false

    // eslint-disable-next-line no-console
    console.warn(`[shop categories] using development fallback: ${reason}`)
    state.categories = fallbackCategories
    state.source = 'dev-fallback'
    state.error = null
    return true
  }

  const waitForExistingLoad = (state: ShopCategoryState): Promise<ShopCategory[]> => (
    new Promise((resolve) => {
      const stop = watch(
        () => state.loading,
        (isLoading) => {
          if (isLoading) return
          stop()
          resolve(state.categories)
        },
        { immediate: true },
      )
    })
  )

  const loadCategories = async (): Promise<ShopCategory[]> => {
    const requestKey = stateKey.value
    const requestLocale = localeCode.value
    const state = getState(requestKey)
    if (state.loading) return waitForExistingLoad(state)

    state.loading = true
    state.error = null

    try {
      const productTypes = await requestCategories(requestLocale)
      if (productTypes.length) {
        state.categories = productTypes
        state.source = 'api'
        return state.categories
      }

      if (applyDevFallback(state, 'API returned no enabled product types')) return state.categories

      state.categories = []
      state.source = 'empty'
      return state.categories
    } catch (e: any) {
      // eslint-disable-next-line no-console
      console.error('Failed to load shop categories:', e)
      const message = e?.data?.message || e?.message || 'Failed to load categories.'

      if (applyDevFallback(state, message)) return state.categories

      state.error = message
      state.categories = []
      state.source = 'error'
      return state.categories
    } finally {
      state.loading = false
    }
  }

  const fetchCategories = async (): Promise<ShopCategory[]> => loadCategories()

  watch(localeCode, () => {
    void loadCategories()
  })

  return {
    categories,
    loading,
    error,
    source,
    fetchCategories,
    loadCategories,
  }
}
