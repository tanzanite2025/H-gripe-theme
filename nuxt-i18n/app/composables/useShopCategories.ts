import { ref, computed } from 'vue'

export interface ShopCategory {
  id: number
  slug: string
  name: string
  count?: number
}

export const useShopCategories = () => {
  const config = useRuntimeConfig()
  const apiBase = computed(() => {
    const base = (config.public as { apiBase?: string }).apiBase || '/api/v1'
    return base.replace(/\/$/, '')
  })

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
      // TODO: Implement /categories endpoint in Go backend.
      // For now, immediately resolve to fallback categories to prevent 404 errors.
      categories.value = fallbackCategories
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
