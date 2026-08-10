import { useRuntimeConfig } from '#imports'
import { useI18n } from 'vue-i18n'
import { ref } from 'vue'

export interface ShopCategory {
  id: number
  slug: string
  name: string
  image?: string
  count?: number
  isProductType?: boolean
}

export type ShopCategorySource = 'api' | 'empty' | 'error' | 'dev-fallback'

const fallbackCategories: ShopCategory[] = [
  { id: 1, slug: 'rims', name: 'Rims' },
  { id: 2, slug: 'hubs', name: 'Hubs' },
  { id: 3, slug: 'spokes', name: 'Spokes' },
  { id: 4, slug: 'accessories', name: 'Accessories' },
]

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
  const baseURL = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
  const categories = ref<ShopCategory[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const source = ref<ShopCategorySource>('empty')

  const categoryRequestHeaders = () => {
    const currentLocale = String(locale.value || '').trim()
    return currentLocale ? { 'Accept-Language': currentLocale } : undefined
  }

  const applyDevFallback = (reason: string) => {
    // 仅本地开发兜底：生产环境必须暴露真实空分类或接口错误，避免示例分类污染商品事实源。
    if (!import.meta.dev) return false

    // eslint-disable-next-line no-console
    console.warn(`[shop categories] using development fallback: ${reason}`)
    categories.value = fallbackCategories
    source.value = 'dev-fallback'
    error.value = null
    return true
  }

  const fetchCategories = async (): Promise<ShopCategory[]> => {
    const response = await $fetch<unknown>(`${baseURL}/products/types`, {
      headers: categoryRequestHeaders(),
    })
    return extractProductTypes(response)
  }

  const loadCategories = async () => {
    if (loading.value) return

    loading.value = true
    error.value = null

    try {
      const productTypes = await fetchCategories()
      if (productTypes.length) {
        categories.value = productTypes
        source.value = 'api'
        return
      }

      if (applyDevFallback('API returned no enabled product types')) return

      categories.value = []
      source.value = 'empty'
    } catch (e: any) {
      // eslint-disable-next-line no-console
      console.error('Failed to load shop categories:', e)
      const message = e?.data?.message || e?.message || 'Failed to load categories.'

      if (applyDevFallback(message)) return

      error.value = message
      categories.value = []
      source.value = 'error'
    } finally {
      loading.value = false
    }
  }

  return {
    categories,
    loading,
    error,
    source,
    fetchCategories,
    loadCategories,
  }
}
