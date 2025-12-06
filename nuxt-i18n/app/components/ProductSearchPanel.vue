<template>
  <section
    class="rounded-xl border border-white/10 bg-black/40 p-4 space-y-3 text-sm text-white/80 product-search-panel"
  >
    <div>
      <div class="flex flex-col md:flex-row gap-1.5 md:items-center">
        <input
          v-model="productSearchQuery"
          type="text"
          :placeholder="$t('sidebar.searchProductPlaceholder', 'Enter product name...')"
          class="flex-1 h-9 px-3 py-2 border border-white/20 rounded-lg bg-white/[0.05] text-white text-[13px] box-border transition-all duration-200 placeholder:text-white/50 focus:outline-none focus:border-[#6b73ff] focus:bg-white/[0.08]"
        />
        <div class="flex gap-2 w-full md:w-auto">
          <button
            type="button"
            class="flex-1 md:flex-none h-9 px-4 min-w-[80px] md:min-w-[120px] border-none rounded-lg bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black text-[13px] font-semibold cursor-pointer box-border transition-all duration-200 hover:shadow-[0_0_20px_rgba(107,115,255,0.5)] hover:-translate-y-0.5"
            @click="searchProducts"
          >
            {{ $t('sidebar.search', 'Search') }}
          </button>
          <button
            type="button"
            class="flex-1 md:flex-none h-9 px-4 min-w-[80px] md:min-w-[120px] border border-white/30 rounded-lg bg-transparent text-white/80 text-[12px] font-medium cursor-pointer box-border transition-all duration-200 hover:bg-white/10 hover:text-white"
            @click="handleReset"
          >
            {{ $t('filter.resetShort', 'Reset') }}
          </button>
        </div>
      </div>
    </div>

    <div class="w-full border-t border-white/10 my-2"></div>

    <AdvancedFilter
      class="sidebar-advanced-filter"
      :key="filterResetKey"
      v-model:filters="filters"
      :options="{
        showPriceRange: true,
        showStockFilter: false,
        showSortBy: false,
        showRating: false,
        showResetButton: false,
        priceMin: 0,
        priceMax: 10000,
      }"
      :attribute-filters="colorAttributes"
      @update:filters="handleFilterChange"
    />
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AdvancedFilter from '~/components/AdvancedFilter.vue'
import { useProductAttributes } from '~/composables/useProductAttributes'

interface ProductSearchFilters {
  priceRange: [number, number]
  inStock: boolean
  preOrder: boolean
  sortBy: string
  minRating?: number
  attributes?: Record<string, string[]>
  [key: string]: any
}

const emit = defineEmits<{
  (e: 'search', payload: { query: string; filters: ProductSearchFilters }): void
}>()

const productSearchQuery = ref('')

const filters = ref<ProductSearchFilters>({
  priceRange: [0, 5000],
  inStock: true,
  preOrder: false,
  sortBy: 'newest',
  minRating: 0,
  attributes: {},
})

const searchingProducts = ref(false)

const filterResetKey = ref(0)

const { colorAttributes, loadFilterableColorAttributes } = useProductAttributes()

const handleFilterChange = (newFilters: ProductSearchFilters) => {
  filters.value = newFilters
  console.log('Product search filters changed:', newFilters)
}

const handleReset = () => {
  productSearchQuery.value = ''
  filters.value = {
    priceRange: [0, 5000],
    inStock: true,
    preOrder: false,
    sortBy: 'newest',
    minRating: 0,
    attributes: {},
  }
  filterResetKey.value += 1
  console.log('Product search filters reset')

  // 重置后立即触发一次搜索，让父级重新请求商品列表
  emit('search', {
    query: '',
    filters: { ...filters.value },
  })
}

const searchProducts = async () => {
  if (searchingProducts.value) return
  searchingProducts.value = true
  const query = productSearchQuery.value.trim()
  console.log('Product search query:', query || '(empty)')
  console.log('Product search filters:', filters.value)

  emit('search', {
    query,
    filters: { ...filters.value },
  })

  setTimeout(() => {
    searchingProducts.value = false
  }, 360)
}

onMounted(() => {
  loadFilterableColorAttributes()
})
</script>

<style scoped>
.product-search-panel {
	width: 100%;
	max-width: 100%;
	box-sizing: border-box;
}

@media (max-width: 768px) {
  /* 移动端隐藏 Color、Diameter、Brake 下拉筛选按钮 */
  :deep(.sidebar-advanced-filter .attribute-top-row),
  :deep(.sidebar-advanced-filter .attribute-inline-row),
  :deep(.sidebar-advanced-filter [class*="attribute-filter"]) {
    display: none !important;
  }
}
</style>
