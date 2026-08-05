<template>
  <section class="search-panel-c">
    <ShopProductQuickSearchForm
      class="search-panel-c__quick-search"
      density="drawer"
      :initial-query="drawerSearchQuery"
      @submit="searchProducts"
    />

    <SmartRecommendationPanel
      :categories="displayedCategoryCards"
      :categories-loading="categoriesLoading"
      @category-click="goToCategory"
      @view-all="close"
    />
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import SmartRecommendationPanel from '~/components/SmartRecommendationPanel.vue'
import ShopProductQuickSearchForm from '~/components/shop/ShopProductQuickSearchForm.vue'
import { useShopSearchSheet, type ShopSearchFiltersPayload } from '~/composables/useShopSearchSheet'
import { useShopCategories, type ShopCategory } from '~/composables/useShopCategories'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'

const emit = defineEmits<{
  (e: 'search', payload: { query: string; filters: ShopSearchFiltersPayload; chipCategorySlug?: string }): void
}>()

const createDefaultSearchFilters = (): ShopSearchFiltersPayload => ({
  priceRange: [0, 5000],
  attributes: {},
})

const drawerSearchQuery = ref('')
const lastSearchFilters = ref<ShopSearchFiltersPayload>(createDefaultSearchFilters())
const searchingProducts = ref(false)

const { presetKeywords, close } = useShopSearchSheet()
const { track: trackBehaviorEvent } = useBehaviorEvents()
const {
  categories: displayedCategoryCards,
  loading: categoriesLoading,
  source: categorySource,
  loadCategories,
} = useShopCategories()

const cloneFilters = (filters: ShopSearchFiltersPayload): ShopSearchFiltersPayload => ({
  ...filters,
  priceRange: Array.isArray(filters.priceRange)
    ? [...filters.priceRange] as [number, number]
    : undefined,
  attributes: { ...(filters.attributes || {}) },
})

const syncPresetQuery = () => {
  drawerSearchQuery.value = Array.from(new Set(
    (presetKeywords.value || [])
      .map((keyword) => String(keyword || '').trim())
      .filter(Boolean)
  )).join(' ')
}

const buildChipCategorySlug = (query: string) => {
  const normalizedPreset = (presetKeywords.value || []).map(keyword => String(keyword || '').trim().toLowerCase())
  if (normalizedPreset.includes('inner tube') || query.toLowerCase().includes('inner tube')) {
    return 'inner-tube'
  }
  return undefined
}

const goToCategory = (category: ShopCategory) => {
  const categoryId = Number(category.id)
  if (categorySource.value === 'api' && Number.isInteger(categoryId) && categoryId > 0) {
    trackBehaviorEvent({
      eventType: 'category_navigation_click',
      categoryId,
      metadata: {
        surface: 'shop_search_drawer',
        target_type: 'category',
        category_slug: category.slug,
        category_source: categorySource.value,
      },
    })
  }
  emit('search', {
    query: '',
    filters: cloneFilters(lastSearchFilters.value),
    chipCategorySlug: category.slug,
  })
}

const searchProducts = async (payload: { query: string; filters: ShopSearchFiltersPayload }) => {
  if (searchingProducts.value) return
  searchingProducts.value = true
  const query = String(payload.query || '').trim()
  const filters = cloneFilters(payload.filters || createDefaultSearchFilters())
  lastSearchFilters.value = filters

  const chipCategorySlug = buildChipCategorySlug(query)
  trackBehaviorEvent({
    eventType: 'search_submit',
    metadata: {
      surface: 'shop_search_drawer',
      query: query.slice(0, 120),
      query_length: query.length,
    },
  })

  emit('search', {
    query,
    filters: cloneFilters(filters),
    ...(chipCategorySlug ? { chipCategorySlug } : {}),
  })

  setTimeout(() => {
    searchingProducts.value = false
  }, 360)
}

onMounted(() => {
  syncPresetQuery()
  loadCategories()
})

watch(presetKeywords, () => {
  syncPresetQuery()
}, { deep: true })
</script>

<style scoped>
/* 搜索面板容器 */
.search-panel-c {
  --search-accent: #B5FF6D;
  --search-accent-soft: rgba(181, 255, 109, 0.42);
  --search-accent-muted: rgba(181, 255, 109, 0.24);
  width: 100%;
  max-width: 1540px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #000000;
  border-radius: 0;
  border: none;
  padding: 0;
  box-shadow: none;
}

.search-panel-c__quick-search {
  width: 100%;
}

@media (max-width: 768px) {
  .search-panel-c {
    gap: 10px;
  }
}
</style>
