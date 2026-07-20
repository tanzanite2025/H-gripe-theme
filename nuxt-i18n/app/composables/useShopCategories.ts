import { useRuntimeConfig } from '#imports'
import { ref } from 'vue'

export interface ShopCategory {
  id: number
  slug: string
  name: string
  count?: number
  isProductType?: boolean
}

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

        if (!Number.isFinite(id) || !slug || !name || record.is_enabled === false) return []
        return [{ id, slug, name, isProductType: true }]
      })
    }

    if (!current || typeof current !== 'object') break
    current = (current as Record<string, unknown>).data
  }

  return []
}

export const useShopCategories = () => {
  const config = useRuntimeConfig()
  const baseURL = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
  const categories = ref<ShopCategory[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 开发环境兜底：当后端分类接口尚未就绪时，提供一组示例分类，保证 UI 能正常工作
  const fallbackCategories: ShopCategory[] = [
    { id: 1, slug: 'rims', name: 'Rims' },
    { id: 2, slug: 'hubs', name: 'Hubs' },
    { id: 3, slug: 'spokes', name: 'Spokes' },
    { id: 4, slug: 'accessories', name: 'Accessories' },
  ]

  const loadCategories = async () => {
    if (loading.value) return

    loading.value = true
    error.value = null

    try {
      const response = await $fetch<unknown>(`${baseURL}/products/types`)
      const productTypes = extractProductTypes(response)
      categories.value = productTypes.length ? productTypes : fallbackCategories
    } catch (e: any) {
      // eslint-disable-next-line no-console
      console.error('Failed to load shop categories:', e)
      error.value = e?.data?.message || e?.message || 'Failed to load categories.'
      categories.value = fallbackCategories
    } finally {
      loading.value = false
    }
  }

  return {
    categories,
    loading,
    error,
    loadCategories,
  }
}
