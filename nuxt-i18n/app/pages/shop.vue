<template>
  <main class="shop-page max-w-5xl mx-auto pt-0 pb-16 space-y-6">
    <section class="rounded-xl bg-white/5 p-4 text-sm text-white/80 shadow-[8px_8px_22px_rgba(0,0,0,0.92)]">
      <div class="flex flex-col gap-2">
        <div class="shop-search-row">
          <div class="shop-search-input-shell">
            <button
              v-for="keyword in quickSelectedKeywords"
              :key="keyword"
              type="button"
              class="shop-search-chip-in-input"
              @click="toggleQuickKeyword(keyword)"
            >
              <span class="shop-search-chip-in-input__label">{{ keyword }}</span>
              <span class="shop-search-chip-in-input__close" aria-hidden="true">×</span>
            </button>
            <input
              v-model="quickFreeTextQuery"
              type="text"
              :placeholder="$t('sidebar.searchProductPlaceholder', 'Enter product name...')"
              class="shop-search-input-inner"
            />
          </div>

          <button
            type="button"
            class="h-[38px] px-4 rounded-lg bg-white text-black font-semibold shadow-[8px_8px_22px_rgba(0,0,0,0.92)] hover:shadow-[10px_10px_26px_rgba(0,0,0,0.95)] transition-all"
            @click="runQuickSearch"
          >
            {{ $t('sidebar.search', 'Search') }}
          </button>

          <button
            type="button"
            class="h-[38px] px-4 rounded-lg bg-white/10 hover:bg-white/15 border border-white/20 hover:border-white/40 text-white font-semibold transition-all inline-flex items-center gap-2"
            :aria-label="$t('filter.filters', 'Filters')"
            @click="openFilters"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L15 12.414V19a1 1 0 01-.553.894l-4 2A1 1 0 019 21v-8.586L3.293 6.707A1 1 0 013 6V4z" />
            </svg>
            <span>{{ $t('filter.filters', 'Filters') }}</span>
          </button>
        </div>

        <PopularSearchChips
          v-model="quickSelectedKeywords"
          :keywords="popularSearchKeywords"
        />
      </div>
    </section>

    <section class="flex gap-4">
      <!-- 桌面端左侧分类栏 -->
      <aside class="hidden md:block w-56 flex-shrink-0">
        <section class="rounded-xl bg-white/5 p-4 text-sm text-white/80 shadow-[8px_8px_22px_rgba(0,0,0,0.92)]">
          <CategorySidebar
            :categories="categories"
            :selected="selectedCategory"
            :loading="categoriesLoading"
            :error="categoriesError"
            @select="onCategorySelect"
          />
        </section>
      </aside>

      <!-- 右侧商品列表区域 -->
      <div class="flex-1">
        <!-- 移动端分类 chips -->
        <div class="md:hidden mb-3 overflow-x-auto">
          <CategoryChips
            :categories="categories"
            :selected="selectedCategory"
            @select="onCategorySelect"
          />
        </div>

        <section class="rounded-xl bg-white/5 p-6 text-sm text-white/80 shadow-[8px_8px_22px_rgba(0,0,0,0.92)]">
          <div v-if="loading" class="flex items-center justify-center py-12">
            <p class="text-white/70 text-sm">Loading products...</p>
          </div>

          <div v-else-if="error" class="py-8 text-center text-red-300 text-sm">
            {{ error }}
          </div>

          <div v-else-if="products.length === 0" class="py-10 text-center space-y-2">
            <p class="text-white/70">No products are available yet.</p>
            <p class="text-white/40 text-xs">
              Once products are published via the Tanzanite plugin, they will appear here automatically.
            </p>
          </div>

          <div v-else class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            <div
              v-for="product in products"
              :key="product.id"
              class="group rounded-xl bg-black/40 hover:bg-black/60 transition-colors overflow-hidden flex flex-col shadow-[8px_8px_22px_rgba(0,0,0,0.92)]"
            >
              <div class="aspect-square bg-white/5">
                <img
                  v-if="product.thumbnail"
                  :src="product.thumbnail"
                  :alt="product.title"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
                <div v-else class="w-full h-full flex items-center justify-center text-white/30 text-2xl">
                  📦
                </div>
              </div>
              <div class="px-3 pt-2 pb-3 flex-1 flex flex-col">
                <h3 class="text-xs font-semibold text-white line-clamp-2 mb-1">
                  {{ product.title }}
                </h3>
                <p v-if="product.price" class="text-xs text-[#40ffaa] mb-2">
                  {{ product.price }}
                </p>
                <div class="mt-auto flex gap-1.5 items-center">
                  <button
                    type="button"
                    @click="handleAddToWishlist(product)"
                    class="w-8 h-8 flex items-center justify-center rounded-full border border-white/25 text-white/80 hover:bg-white/15 transition-colors"
                    title="Add to wishlist"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.7"
                        d="M12.1 19.3 12 19.4l-.1-.1C7.14 15.24 4 12.39 4 9.2 4 7 5.7 5.3 7.9 5.3c1.4 0 2.8.7 3.6 1.9 0.8-1.2 2.2-1.9 3.6-1.9 2.2 0 3.9 1.7 3.9 3.9 0 3.19-3.14 6.04-7.9 10.1z"
                      />
                    </svg>
                  </button>

                  <NuxtLink
                    :to="product.url"
                    class="flex-1 px-2 py-1.5 bg-white/10 hover:bg-white/20 border border-white/20 hover:border-white/40 rounded text-[11px] text-white text-center transition-all"
                  >
                    View
                  </NuxtLink>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </section>

    <section class="mt-10">
      <UserFeedbackThread
        threadKey="shop-page"
        title="Share your feedback about the Tanzanite shop"
      />
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute, useRuntimeConfig, useAsyncData } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import CategorySidebar from '~/components/CategorySidebar.vue'
import CategoryChips from '~/components/CategoryChips.vue'
import PopularSearchChips from '~/components/PopularSearchChips.vue'
import { useWishlist } from '~/composables/useWishlist'
import { useShopCategories } from '~/composables/useShopCategories'
import type { ShopCategory } from '~/composables/useShopCategories'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'
import { popularSearchKeywords } from '~/utils/popularSearchKeywords'

definePageMeta({
  layout: 'products',
})

interface ShopProduct {
  id: number
  title: string
  url: string
  thumbnail?: string
  price?: string
}

const route = useRoute()

const quickSelectedKeywords = ref<string[]>([])
const quickFreeTextQuery = ref('')
const quickSearchQuery = ref('')

// 商品心愿单
const { addToWishlist } = useWishlist()

// 商品分类
const { categories, loading: categoriesLoading, error: categoriesError, loadCategories } = useShopCategories()
const selectedCategory = ref<ShopCategory | null>(null)

interface ProductSearchFiltersPayload {
  priceRange: [number, number]
  attributes?: Record<string, string[]>
}

interface ProductSearchPayload {
  query: string
  filters: ProductSearchFiltersPayload
  chipCategorySlug?: string
}

const currentSearch = ref<ProductSearchPayload | null>(null)

const DEFAULT_QUICK_FILTERS: ProductSearchFiltersPayload = {
  priceRange: [0, 5000],
  attributes: {},
}

const { open: openShopSearchSheet, pendingSearch, presetCategorySlug } = useShopSearchSheet()

const openFilters = () => {
  openShopSearchSheet()
}

const syncQuickSearchQuery = () => {
  const parts: string[] = []
  if (quickSelectedKeywords.value.length) {
    parts.push(...quickSelectedKeywords.value)
  }
  const free = quickFreeTextQuery.value.trim()
  if (free) {
    parts.push(free)
  }
  quickSearchQuery.value = parts.join(' ')
}

const toggleQuickKeyword = (keyword: string) => {
  const current = [...quickSelectedKeywords.value]
  const index = current.indexOf(keyword)
  if (index === -1) {
    current.push(keyword)
  } else {
    current.splice(index, 1)
  }
  quickSelectedKeywords.value = current
  syncQuickSearchQuery()
}

const categorySlugToKeyword = (slug: string) => slug.replace(/[-_]+/g, ' ').trim()

const joinUniqueSearchParts = (parts: Array<string | null | undefined>) => {
  const seen = new Set<string>()
  const normalized: string[] = []

  for (const part of parts) {
    const value = String(part || '').trim()
    if (!value) continue

    const key = value.toLowerCase()
    if (seen.has(key)) continue

    seen.add(key)
    normalized.push(value)
  }

  return normalized.join(' ')
}

const buildProductKeyword = (payload?: ProductSearchPayload) => joinUniqueSearchParts([
  payload?.query,
  selectedCategory.value?.name || selectedCategory.value?.slug,
  payload?.chipCategorySlug ? categorySlugToKeyword(payload.chipCategorySlug) : null,
])

const runQuickSearch = () => {
  handleSearch({
    query: quickSearchQuery.value,
    filters: { ...DEFAULT_QUICK_FILTERS },
  })
}

const buildProductQueryParams = (payload?: ProductSearchPayload) => {
  const params: Record<string, any> = {
    per_page: 24,
    status: 'active',
  }

  if (payload) {
    const keyword = buildProductKeyword(payload)
    if (keyword) {
      params.keyword = keyword
    }

    const priceRange = payload.filters?.priceRange
    if (Array.isArray(priceRange) && priceRange.length === 2) {
      const [min, max] = priceRange
      params.price_min = min
      params.price_max = max
    }

    const attrs = payload.filters?.attributes
    if (attrs && typeof attrs === 'object') {
      params.attributes = JSON.stringify(attrs)
    }
  }

  return params
}

const { data: asyncData, pending, error: asyncError, refresh } = await useAsyncData(
  'shop-products',
  () => {
    const config = useRuntimeConfig()
    const base = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
    const params = buildProductQueryParams(currentSearch.value || undefined)
    Object.assign(params, route.query)

    return $fetch<any>(`${base}/customer-service/products`, { params })
  }
)

watch(() => route.query, () => {
  refresh()
}, { deep: true })

const products = computed<ShopProduct[]>(() => {
  const response = asyncData.value
  if (response && Array.isArray(response.items)) {
    return response.items.map((item: any) => ({
      id: item.id,
      title: item.title,
      url: `/shop/${item.slug || item.id}`,
      thumbnail: item.thumbnail,
      price:
        item.prices?.sale > 0
          ? `$${item.prices.sale}`
          : item.prices?.regular > 0
          ? `$${item.prices.regular}`
          : '',
    }))
  }
  return []
})

const loading = computed(() => pending.value)
const error = computed(() => asyncError.value?.message || null)

const loadProducts = async (payload?: ProductSearchPayload) => {
  await refresh()
}

const handleSearch = (payload: ProductSearchPayload) => {
  if (payload.chipCategorySlug && Array.isArray(categories.value) && categories.value.length) {
    const match = categories.value.find(cat => cat.slug === payload.chipCategorySlug)
    if (match) {
      selectedCategory.value = match
    }
  }

  const next: ProductSearchPayload = {
    ...payload,
  }

  currentSearch.value = next
  loadProducts(next)
}

const onCategorySelect = (category: ShopCategory | null) => {
  selectedCategory.value = category

  const base: ProductSearchPayload =
    currentSearch.value || ({
      query: '',
      filters: {
        priceRange: [0, 5000],
        attributes: {},
      },
    } as ProductSearchPayload)

  const next: ProductSearchPayload = {
    ...base,
  }

  currentSearch.value = next
  loadProducts(next)
}

const applyPresetCategoryFromSlug = () => {
  const slug = presetCategorySlug.value
  if (!slug || !Array.isArray(categories.value) || !categories.value.length) return

  const match = categories.value.find((cat) => cat.slug === slug)
  if (match) {
    selectedCategory.value = match
  }

  // 只用于入口预设，消费一次后清空，避免影响后续手动选择
  presetCategorySlug.value = null
}

onMounted(async () => {
  await loadCategories()

  // 页面首次挂载时，如果是从 Inner tube 等入口过来，先根据 slug 预设分类
  applyPresetCategoryFromSlug()

  const initialPending = pendingSearch.value
  if (initialPending) {
    pendingSearch.value = null
    handleSearch(initialPending as unknown as ProductSearchPayload)
    return
  }

  loadProducts()
})

watch(pendingSearch, async (payload) => {
  if (!payload) return
  pendingSearch.value = null

  // 确保分类已加载，再根据 slug 预设分类
  if (!categories.value.length) {
    await loadCategories()
  }
  applyPresetCategoryFromSlug()

  handleSearch(payload as unknown as ProductSearchPayload)
})

watch(quickFreeTextQuery, () => {
  syncQuickSearchQuery()
})

const handleAddToWishlist = async (product: ShopProduct) => {
  if (!product?.id) return
  try {
    await addToWishlist(product.id)
  } catch (e) {
    console.error('Failed to add to wishlist from shop:', e)
  }
}
</script>

<style scoped>
.shop-page {
  padding-inline: 1.5rem;
}

@media (max-width: 768px) {
  .shop-page {
    padding-inline: 0;
  }
}

.shop-search-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.shop-search-input-shell {
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

.shop-search-chip-in-input {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border-radius: 9999px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 500;
  border: none;
  background: #ffffff;
  color: #000000;
  cursor: pointer;
}

.shop-search-chip-in-input__close {
  font-size: 11px;
  opacity: 0.75;
}

.shop-search-input-inner {
  flex: 1;
  min-width: 120px;
  border: none;
  background: transparent;
  color: #ffffff;
  font-size: 13px;
  outline: none;
}

.shop-search-input-inner::placeholder {
  color: rgba(148,163,184,0.7);
}

@media (max-width: 400px) {
  .shop-page {
    padding-inline: 0;
  }
}
</style>
