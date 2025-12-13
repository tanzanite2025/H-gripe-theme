<template>
  <section class="search-panel-c">
    <!-- 搜索输入行 -->
    <div class="search-row-c">
      <input
        v-model="productSearchQuery"
        type="text"
        :placeholder="$t('sidebar.searchProductPlaceholder', 'Enter product name...')"
        class="search-input-c"
      />
      <div class="btn-group">
        <button
          type="button"
          class="search-btn-c primary"
          @click="searchProducts"
        >
          {{ $t('sidebar.search', 'Search') }}
        </button>
        <button
          type="button"
          class="search-btn-c secondary"
          @click="handleReset"
        >
          {{ $t('filter.resetShort', 'Reset') }}
        </button>
      </div>
    </div>

    <!-- 分隔线 -->
    <div class="filter-divider-c"></div>

    <!-- 高级筛选 -->
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
/* 搜索面板容器 */
.search-panel-c {
  background: radial-gradient(circle at top left, rgba(41,55,80,0.96), rgba(15,23,42,0.98));
  border-radius: 18px;
  border: none;
  padding: 14px;
  box-shadow:
    0 8px 24px -12px rgba(0,0,0,0.9),
    0 0 18px rgba(15,23,42,0.9);
}

/* 搜索行 */
.search-row-c {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.btn-group {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

/* 搜索输入框 */
.search-input-c {
  flex: 1;
  min-width: 0;
  height: 38px;
  padding: 0 14px;
  background: linear-gradient(135deg, rgba(15,23,42,0.98), rgba(15,23,42,0.96));
  border: none;
  border-radius: 10px;
  color: #ffffff;
  font-size: 13px;
  outline: none;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow:
    0 2px 6px -3px rgba(0,0,0,0.9),
    0 0 6px rgba(15,23,42,0.7);
}
.search-input-c::placeholder { color: rgba(148,163,184,0.7); }
.search-input-c:focus {
  background: linear-gradient(135deg, rgba(15,23,42,0.98), rgba(15,23,42,0.98));
  box-shadow:
    0 0 0 1px rgba(45,212,191,0.75),
    0 0 14px rgba(45,212,191,0.35);
}

/* 搜索按钮 */
.search-btn-c {
  height: 38px;
  padding: 0 16px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}
.search-btn-c.primary {
  background: #ffffff;
  color: #000000;
  box-shadow:
    8px 8px 22px rgba(0,0,0,0.92);
}
.search-btn-c.primary:hover {
  transform: translateY(-1px);
  box-shadow:
    10px 10px 26px rgba(0,0,0,0.95);
}
.search-btn-c.secondary {
  background: linear-gradient(135deg, rgba(15,23,42,0.98), rgba(15,23,42,0.96));
  color: #ffffff;
  box-shadow:
    0 2px 6px -4px rgba(0,0,0,0.9),
    0 0 6px rgba(15,23,42,0.7);
}
.search-btn-c.secondary:hover {
  background: linear-gradient(135deg, rgba(31,41,55,0.98), rgba(15,23,42,0.96));
  color: #ffffff;
}

/* 分隔线 */
.filter-divider-c {
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent);
  margin: 12px 0;
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
