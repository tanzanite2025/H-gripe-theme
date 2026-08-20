import { useRuntimeConfig, useState } from '#imports'
import { useI18n } from 'vue-i18n'
import { computed, watch } from 'vue'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'

export interface ShopCategory {
  id: number
  slug: string
  name: string
  image?: string
  count?: number
  isProductSpecificationTemplate?: boolean
}

export type ShopCategorySource = 'api' | 'empty' | 'error'

interface ShopCategoryState {
  categories: ShopCategory[]
  loading: boolean
  loaded: boolean
  error: string | null
  source: ShopCategorySource
}

type ShopCategoryStateStore = Record<string, ShopCategoryState>

const createEmptyCategoryState = (): ShopCategoryState => ({
  categories: [],
  loading: false,
  loaded: false,
  error: null,
  source: 'empty',
})

const extractProductSpecificationTemplates = (
  payload: unknown,
  mediaContext: ReturnType<typeof createStorefrontMediaContext>,
): ShopCategory[] => {
  let current = payload

  for (let depth = 0; depth < 3; depth += 1) {
    if (Array.isArray(current)) {
      return current.flatMap((item) => {
        if (!item || typeof item !== 'object') return []

        const record = item as Record<string, unknown>
        const id = Number(record.id)
        const slug = String(record.slug || '').trim()
        const name = String(record.name || '').trim()
        const image = normalizeStorefrontMediaUrl(
          record.image_url || record.image,
          mediaContext,
        )

        if (!Number.isFinite(id) || !slug || !name || record.is_enabled === false) return []
        return [{
          id,
          slug,
          name,
          ...(image ? { image } : {}),
          isProductSpecificationTemplate: true,
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
  const mediaContext = createStorefrontMediaContext(config)
  const { locale } = useI18n()
  // Every storefront surface reads this store so category names, slugs, and images stay aligned.
  const stateStore = useState<ShopCategoryStateStore>('shop-categories-by-locale', () => ({}))
  const publicBaseURL = computed(() => ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, ''))
  const internalApiOrigin = import.meta.server
    ? String((config as { apiInternalOrigin?: string }).apiInternalOrigin || '').replace(/\/$/, '')
    : ''
  const requestBaseURL = computed(() => {
    if (internalApiOrigin) return `${internalApiOrigin}/api/v1`
    return publicBaseURL.value
  })
  const localeCode = computed(() => String(locale.value || '').trim() || 'en')
  const stateKey = computed(() => `${publicBaseURL.value}|${localeCode.value}`)
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
    const response = await $fetch<unknown>(`${requestBaseURL.value}/products/specification-templates`, { headers })
    return extractProductSpecificationTemplates(response, mediaContext)
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
    if (state.loaded) return state.categories

    state.loading = true
    state.error = null

    try {
      const productSpecificationTemplates = await requestCategories(requestLocale)
      if (productSpecificationTemplates.length) {
        state.categories = productSpecificationTemplates
        state.loaded = true
        state.source = 'api'
        return state.categories
      }

      state.categories = []
      state.loaded = true
      state.source = 'empty'
      return state.categories
    } catch (e: any) {
      // eslint-disable-next-line no-console
      console.error('Failed to load shop categories:', e)
      const message = e?.data?.message || e?.message || 'Failed to load categories.'

      state.error = message
      state.categories = []
      state.loaded = false
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
