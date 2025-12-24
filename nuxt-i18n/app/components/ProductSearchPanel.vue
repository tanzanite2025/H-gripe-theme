<template>
  <section class="search-panel-c">
    <!-- 搜索输入行：外壳 + 已选热门 TAB 胶囊 + 输入框 -->
    <div class="search-row-c">
      <div class="search-input-shell">
        <button
          v-for="keyword in selectedKeywords"
          :key="keyword"
          type="button"
          class="search-chip-in-input"
          @click="toggleKeyword(keyword)"
        >
          <span class="search-chip-in-input__label">{{ keyword }}</span>
          <span class="search-chip-in-input__close" aria-hidden="true">×</span>
        </button>
        <input
          v-model="freeTextQuery"
          type="text"
          :placeholder="$t('sidebar.searchProductPlaceholder', 'Enter product name...')"
          class="search-input-inner"
        />
      </div>
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

    <!-- 热门搜索 TAB 区域 -->
    <PopularSearchChips
      v-model="selectedKeywords"
      :keywords="popularSearchKeywords"
    />

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
import { ref, onMounted, watch } from 'vue'
import AdvancedFilter from '~/components/AdvancedFilter.vue'
import PopularSearchChips from '~/components/PopularSearchChips.vue'
import { popularSearchKeywords } from '~/utils/popularSearchKeywords'
import { useProductAttributes } from '~/composables/useProductAttributes'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'

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

const selectedKeywords = ref<string[]>([])
const freeTextQuery = ref('')
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
const { presetKeywords } = useShopSearchSheet()

const syncProductSearchQuery = () => {
  const parts: string[] = []
  if (selectedKeywords.value.length) {
    parts.push(...selectedKeywords.value)
  }
  const free = freeTextQuery.value.trim()
  if (free) {
    parts.push(free)
  }
  productSearchQuery.value = parts.join(' ')
}

const toggleKeyword = (keyword: string) => {
  const current = [...selectedKeywords.value]
  const index = current.indexOf(keyword)
  if (index === -1) {
    current.push(keyword)
  } else {
    current.splice(index, 1)
  }
  selectedKeywords.value = current
  syncProductSearchQuery()
}

const handleFilterChange = (newFilters: ProductSearchFilters) => {
  filters.value = newFilters
  console.log('Product search filters changed:', newFilters)
}

const handleReset = () => {
  selectedKeywords.value = []
  freeTextQuery.value = ''
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
  syncProductSearchQuery()
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

  if (Array.isArray(presetKeywords.value) && presetKeywords.value.length) {
    selectedKeywords.value = [...presetKeywords.value]
    syncProductSearchQuery()
  }
})

watch(freeTextQuery, () => {
  syncProductSearchQuery()
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
    0 10px 28px -14px rgba(0,0,0,0.92),
    0 0 24px rgba(0,0,0,0.85);
}

/* 搜索行 */
.search-row-c {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 8px;
}

.search-input-shell {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  padding: 3px 4px;
  background: linear-gradient(135deg, rgba(15,23,42,0.98), rgba(15,23,42,0.96));
  border-radius: 10px;
  box-shadow:
    0 2px 6px -3px rgba(0,0,0,0.9),
    0 0 6px rgba(15,23,42,0.7);
}

.search-chip-in-input {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 500;
  border: none;
  background: #ffffff;
  color: #000000;
  cursor: pointer;
}

.search-chip-in-input__close {
  font-size: 11px;
  opacity: 0.75;
}

.search-input-inner {
  flex: 1;
  min-width: 120px;
  border: none;
  background: transparent;
  color: #ffffff;
  font-size: 13px;
  outline: none;
}

.search-input-inner::placeholder {
  color: rgba(148,163,184,0.7);
}

.btn-group {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

/* 搜索输入框原有样式合并到 search-input-shell / search-input-inner */

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
</style>
