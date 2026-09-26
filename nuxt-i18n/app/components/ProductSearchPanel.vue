<template>
  <section class="search-panel-c">
    <ShopProductQuickSearchForm
      class="search-panel-c__quick-search"
      density="drawer"
      :initial-query="drawerSearchQuery"
      @submit="searchProducts"
    />
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import ShopProductQuickSearchForm from '~/components/shop/ShopProductQuickSearchForm.vue'
import { useShopSearchSheet, type ShopSearchFiltersPayload } from '~/composables/useShopSearchSheet'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { useStorefrontContext } from '~/composables/useStorefrontContext'
import { defaultPriceRangeForCurrency } from '~/utils/money'

const emit = defineEmits<{
  (e: 'search', payload: { query: string; filters: ShopSearchFiltersPayload; chipCategorySlug?: string }): void
}>()
const { displayCurrency } = useStorefrontContext()

const createDefaultSearchFilters = (): ShopSearchFiltersPayload => ({
  priceRange: defaultPriceRangeForCurrency(displayCurrency.value),
  attributes: {},
})

const drawerSearchQuery = ref('')
const lastSearchFilters = ref<ShopSearchFiltersPayload>(createDefaultSearchFilters())
const searchingProducts = ref(false)

const { presetKeywords } = useShopSearchSheet()
const { track: trackBehaviorEvent } = useBehaviorEvents()

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
})

watch(presetKeywords, () => {
  syncPresetQuery()
}, { deep: true })
</script>

<style scoped>
/* 搜索面板容器 */
.search-panel-c {
  --search-accent: #059669;
  --search-accent-soft: rgba(5, 150, 105, 0.42);
  --search-accent-muted: rgba(5, 150, 105, 0.24);
  width: 100%;
  max-width: 1540px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--tz-card-surface);
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
