import { computed, watch } from 'vue'
import { useI18n, useRuntimeConfig, useState } from '#imports'

export interface ProductCategory {
  id: number
  parentId: number | null
  name: string
  slug: string
  description: string
  imageUrl: string
  depth: number
  sortOrder: number
  children: ProductCategory[]
}

export interface ProductCategoryList {
  tree: ProductCategory[]
  flat: ProductCategory[]
  maxDepth: number
}

interface ProductCategoryState {
  tree: ProductCategory[]
  flat: ProductCategory[]
  maxDepth: number
  loading: boolean
  loaded: boolean
  error: string | null
}

type ProductCategoryStateStore = Record<string, ProductCategoryState>

const createEmptyState = (): ProductCategoryState => ({
  tree: [],
  flat: [],
  maxDepth: 5,
  loading: false,
  loaded: false,
  error: null,
})

const normalizeCategory = (value: any): ProductCategory | null => {
  const id = Number(value?.id)
  const slug = String(value?.slug || '').trim()
  const name = String(value?.name || '').trim()
  if (!Number.isFinite(id) || id <= 0 || !slug || !name) return null

  const children = Array.isArray(value?.children)
    ? value.children
      .map((child: any) => normalizeCategory(child))
      .filter((child: ProductCategory | null): child is ProductCategory => Boolean(child))
    : []

  return {
    id,
    parentId: Number.isFinite(Number(value?.parent_id)) && Number(value.parent_id) > 0
      ? Number(value.parent_id)
      : null,
    name,
    slug,
    description: String(value?.description || '').trim(),
    imageUrl: String(value?.image_url || '').trim(),
    depth: Math.max(1, Number(value?.depth) || 1),
    sortOrder: Number(value?.sort_order) || 0,
    children,
  }
}

const flattenCategories = (tree: ProductCategory[]): ProductCategory[] => {
  const result: ProductCategory[] = []
  const visit = (items: ProductCategory[]) => {
    for (const item of items) {
      result.push(item)
      visit(item.children)
    }
  }
  visit(tree)
  return result
}

const extractProductCategoryList = (payload: unknown): ProductCategoryList => {
  let current: any = payload
  for (let depth = 0; depth < 3; depth += 1) {
    if (!current || typeof current !== 'object') break
    if (Array.isArray(current.tree) || Array.isArray(current.flat)) {
      const tree = Array.isArray(current.tree)
        ? current.tree
          .map((item: any) => normalizeCategory(item))
          .filter((item: ProductCategory | null): item is ProductCategory => Boolean(item))
        : []
      const flat = Array.isArray(current.flat)
        ? current.flat
          .map((item: any) => normalizeCategory(item))
          .filter((item: ProductCategory | null): item is ProductCategory => Boolean(item))
        : flattenCategories(tree)

      return {
        tree,
        flat: flat.length ? flat : flattenCategories(tree),
        maxDepth: Math.max(1, Number(current.max_depth) || 5),
      }
    }
    current = current.data
  }

  return {
    tree: [],
    flat: [],
    maxDepth: 5,
  }
}

export const useProductCategories = () => {
  const config = useRuntimeConfig()
  const { locale } = useI18n()
  const stateStore = useState<ProductCategoryStateStore>('product-categories-by-locale', () => ({}))
  const publicBaseURL = computed(() => (
    ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
  ))
  const requestBaseURL = computed(() => {
    if (import.meta.server) {
      const internalOrigin = String(
        (config as { apiInternalOrigin?: string }).apiInternalOrigin || '',
      ).replace(/\/$/, '')
      if (internalOrigin) return `${internalOrigin}/api/v1`
    }
    return publicBaseURL.value
  })
  const localeCode = computed(() => String(locale.value || '').trim() || 'en')
  const stateKey = computed(() => `${publicBaseURL.value}|${localeCode.value}`)
  const emptyState = createEmptyState()

  const getState = (key = stateKey.value): ProductCategoryState => {
    const existing = stateStore.value[key]
    if (existing) return existing
    const next = createEmptyState()
    stateStore.value[key] = next
    return next
  }

  const currentState = computed(() => stateStore.value[stateKey.value] || emptyState)
  const tree = computed(() => currentState.value.tree)
  const categories = computed(() => currentState.value.flat)
  const maxDepth = computed(() => currentState.value.maxDepth)
  const loading = computed(() => currentState.value.loading)
  const loaded = computed(() => currentState.value.loaded)
  const error = computed(() => currentState.value.error)

  const requestCategories = async (requestLocale: string): Promise<ProductCategoryList> => {
    const headers = requestLocale ? { 'Accept-Language': requestLocale } : undefined
    const response = await $fetch<unknown>(`${requestBaseURL.value}/products/categories`, { headers })
    return extractProductCategoryList(response)
  }

  const waitForExistingLoad = (state: ProductCategoryState): Promise<ProductCategory[]> => (
    new Promise((resolve) => {
      const stop = watch(
        () => state.loading,
        (isLoading) => {
          if (isLoading) return
          stop()
          resolve(state.flat)
        },
        { immediate: true },
      )
    })
  )

  const loadCategories = async (): Promise<ProductCategory[]> => {
    const state = getState()
    if (state.loading) return waitForExistingLoad(state)
    if (state.loaded) return state.flat

    state.loading = true
    state.error = null

    try {
      const result = await requestCategories(localeCode.value)
      state.tree = result.tree
      state.flat = result.flat
      state.maxDepth = result.maxDepth
      state.loaded = true
      return state.flat
    } catch (e: any) {
      console.error('Failed to load product categories:', e)
      state.error = e?.data?.message || e?.message || 'Failed to load product categories.'
      state.tree = []
      state.flat = []
      state.loaded = false
      return state.flat
    } finally {
      state.loading = false
    }
  }

  watch(localeCode, () => {
    void loadCategories()
  })

  return {
    tree,
    categories,
    maxDepth,
    loading,
    loaded,
    error,
    loadCategories,
  }
}
