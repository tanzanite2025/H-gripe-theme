<template>
  <section
    class="rounded-xl border border-white/10 bg-black/40 p-6 space-y-4 text-sm text-white/80 product-search-panel"
  >
    <div>
      <h2 class="text-base font-semibold text-white mb-3">Product Search</h2>
      <div class="flex flex-col md:flex-row gap-2 md:items-center">
        <input
          v-model="productSearchQuery"
          type="text"
          :placeholder="$t('sidebar.searchProductPlaceholder', 'Enter product name...')"
          class="flex-1 h-9 px-3 py-2 border border-white/20 rounded-lg bg-white/[0.05] text-white text-[13px] box-border transition-all duration-200 placeholder:text-white/50 focus:outline-none focus:border-[#6b73ff] focus:bg-white/[0.08]"
        />
        <div class="flex gap-2 w-full md:w-auto">
          <button
            type="button"
            class="flex-1 md:flex-none h-9 px-4 min-w-[120px] border-none rounded-lg bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black text-[13px] font-semibold cursor-pointer box-border transition-all duration-200 hover:shadow-[0_0_20px_rgba(107,115,255,0.5)] hover:-translate-y-0.5"
            @click="searchProducts"
          >
            {{ $t('sidebar.searchProducts', 'Search Products') }}
          </button>
          <button
            type="button"
            class="flex-1 md:flex-none h-9 px-4 min-w-[120px] border border-white/30 rounded-lg bg-transparent text-white/80 text-[12px] font-medium cursor-pointer box-border transition-all duration-200 hover:bg-white/10 hover:text-white"
            @click="handleReset"
          >
            {{ $t('filter.reset', 'Reset Filters') }}
          </button>
        </div>
      </div>
    </div>

    <div class="w-full border-t border-white/10 my-3"></div>

    <AdvancedFilter
      v-model:filters="filters"
      :options="{
        showPriceRange: true,
        showStockFilter: true,
        showSortBy: false,
        showRating: false,
        showResetButton: false,
        priceMin: 0,
        priceMax: 10000,
      }"
      @update:filters="handleFilterChange"
    />
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AdvancedFilter from '~/components/AdvancedFilter.vue'

interface ProductSearchFilters {
  priceRange: [number, number]
  inStock: boolean
  preOrder: boolean
  sortBy: string
  minRating?: number
  [key: string]: any
}

const productSearchQuery = ref('')

const filters = ref<ProductSearchFilters>({
  priceRange: [0, 5000],
  inStock: true,
  preOrder: false,
  sortBy: 'newest',
  minRating: 0,
})

const searchingProducts = ref(false)

const handleFilterChange = (newFilters: ProductSearchFilters) => {
  filters.value = newFilters
  console.log('Product search filters changed:', newFilters)
}

const handleReset = () => {
  console.log('Product search filters reset')
}

const searchProducts = async () => {
  if (searchingProducts.value) return
  searchingProducts.value = true
  const query = productSearchQuery.value.trim()
  console.log('Product search query:', query || '(empty)')
  console.log('Product search filters:', filters.value)
  // TODO: 在此根据 query + filters 触发实际的商品搜索逻辑或向父组件发事件
  setTimeout(() => {
    searchingProducts.value = false
  }, 360)
}
</script>

<style scoped>
</style>
