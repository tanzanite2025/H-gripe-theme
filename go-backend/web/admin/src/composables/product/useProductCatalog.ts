import { computed, reactive, ref } from 'vue'
import productApi from '@/api/products'

interface ProductStats {
  total?: number
  active?: number
  low_stock?: number
  out_of_stock?: number
}

type ProductRecord = Record<string, any>

export const useProductCatalog = () => {
  const loading = ref(false)
  const products = ref<ProductRecord[]>([])
  const selectedProducts = ref<ProductRecord[]>([])
  const stats = ref<ProductStats>({})
  const filters = reactive({ search: '', status: 'all', locale: 'all', featured: 'all' })
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

  const selectionState = computed(() => {
    if (products.value.length === 0 || selectedProducts.value.length === 0) return false
    return selectedProducts.value.length === products.value.length ? true : 'indeterminate'
  })

  const buildFilterParams = () => ({
    ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
    ...(filters.status !== 'all' ? { status: filters.status } : {}),
    ...(filters.locale !== 'all' ? { locale: filters.locale } : {}),
    ...(filters.featured !== 'all' ? { featured: filters.featured } : {})
  })

  const fetchStats = async () => {
    try {
      stats.value = await productApi.stats()
    } catch (error) {
      console.error('Failed to fetch product stats:', error)
    }
  }

  const fetchProducts = async () => {
    loading.value = true
    try {
      const data = await productApi.list({ page: pagination.page, page_size: pagination.pageSize, ...buildFilterParams() })
      products.value = data.products || []
      pagination.total = data.pagination?.total || 0
      selectedProducts.value = []
    } catch (error) {
      console.error('Failed to fetch products:', error)
    } finally {
      loading.value = false
    }
  }

  const refreshProducts = () => Promise.all([fetchProducts(), fetchStats()])

  const applyFilters = () => {
    pagination.page = 1
    fetchProducts()
  }

  const resetFilters = () => {
    Object.assign(filters, { search: '', status: 'all', locale: 'all', featured: 'all' })
    pagination.page = 1
    fetchProducts()
  }

  const updatePage = (page: number) => {
    pagination.page = page
    fetchProducts()
  }

  const updatePageSize = (pageSize: number) => {
    pagination.pageSize = pageSize
    pagination.page = 1
    fetchProducts()
  }

  const isSelected = (productId: number | string) => selectedProducts.value.some((product: any) => product.id === productId)

  const toggleAllProducts = (checked: boolean | 'indeterminate') => {
    selectedProducts.value = checked === true ? [...products.value] : []
  }

  const toggleProduct = (product: any, checked: boolean | 'indeterminate') => {
    if (checked === true && !isSelected(product.id)) {
      selectedProducts.value = [...selectedProducts.value, product]
    } else if (checked !== true) {
      selectedProducts.value = selectedProducts.value.filter((selected: any) => selected.id !== product.id)
    }
  }

  return {
    loading,
    products,
    selectedProducts,
    stats,
    filters,
    pagination,
    selectionState,
    fetchStats,
    fetchProducts,
    refreshProducts,
    applyFilters,
    resetFilters,
    updatePage,
    updatePageSize,
    isSelected,
    toggleAllProducts,
    toggleProduct
  }
}

export default useProductCatalog
