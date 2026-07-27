<template>
  <main class="shop-page w-full pt-0 pb-16 space-y-6">
    <section class="rounded-xl bg-white/5 p-4 text-sm tz-text-secondary shadow-[8px_8px_22px_rgba(0,0,0,0.92)]">
      <form class="shop-search-form" @submit.prevent="runQuickSearch">
        <label class="shop-search-input-shell">
          <span class="sr-only">{{ $t('sidebar.searchProductPlaceholder', 'Enter product name...') }}</span>
          <input
            v-model="quickFreeTextQuery"
            type="text"
            :placeholder="$t('sidebar.searchProductPlaceholder', 'Enter product name...')"
            class="shop-search-input-inner"
          />
        </label>

        <div class="shop-price-range" :aria-label="$t('filter.priceRange', 'Price Range')">
          <span class="shop-price-range__label">{{ $t('filter.price', 'Price') }}</span>
          <label class="shop-currency-field">
            <span class="sr-only">{{ $t('filter.currency', 'Currency') }}</span>
            <select
              v-model="quickCurrency"
              :disabled="paymentCurrenciesLoading || paymentCurrencies.length === 0"
              aria-label="Currency"
            >
              <option value="" disabled>{{ currencySelectLabel }}</option>
              <option
                v-for="currency in paymentCurrencies"
                :key="currency"
                :value="currency"
              >
                {{ currency }}
              </option>
            </select>
          </label>
          <label class="shop-price-field">
            <span>{{ $t('filter.from', 'From') }}</span>
            <input
              v-model.number="quickPriceMin"
              type="number"
              min="0"
              inputmode="numeric"
              aria-label="Minimum price"
            />
          </label>
          <label class="shop-price-field">
            <span>{{ $t('filter.to', 'To') }}</span>
            <input
              v-model.number="quickPriceMax"
              type="number"
              min="0"
              inputmode="numeric"
              aria-label="Maximum price"
            />
          </label>
        </div>

        <div class="shop-search-actions">
          <button
            type="submit"
            class="shop-search-submit"
          >
            {{ $t('sidebar.search', 'Search') }}
          </button>

          <button
            type="button"
            class="shop-filter-button"
            :aria-label="$t('filter.filters', 'Filters')"
            @click="openFilters"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L15 12.414V19a1 1 0 01-.553.894l-4 2A1 1 0 019 21v-8.586L3.293 6.707A1 1 0 013 6V4z" />
            </svg>
            <span>{{ $t('filter.filters', 'Filters') }}</span>
          </button>
        </div>
      </form>
    </section>

    <teleport to="body">
      <transition name="shop-category-sheet">
        <div
          v-if="categoryFilterOpen"
          class="shop-category-sheet"
          role="dialog"
          aria-modal="true"
          :aria-label="$t('filter.categories', 'Categories')"
          @click.self="closeCategoryFilter"
        >
          <section class="shop-category-sheet__panel">
            <header class="shop-category-sheet__header">
              <span>{{ $t('filter.categories', 'Categories') }}</span>
              <button
                type="button"
                class="shop-category-sheet__close"
                aria-label="Close categories"
                @click="closeCategoryFilter"
              >
                <Icon name="lucide:x" />
              </button>
            </header>

            <ShopCategoryVerticalMenu
              :categories="categories"
              :selected="selectedCategory"
              :loading="categoriesLoading"
              :error="categoriesError"
              @select="onMobileCategorySelect"
            />
          </section>
        </div>
      </transition>
    </teleport>

    <section class="shop-catalog-layout">
      <!-- Desktop category rail: local component, no external dependency. -->
      <aside class="shop-category-rail" :aria-label="$t('shopPage.categoryRailLabel', 'Shop category navigation')">
        <ShopCategoryVerticalMenu
          :categories="categories"
          :selected="selectedCategory"
          :loading="categoriesLoading"
          :error="categoriesError"
          @select="onCategorySelect"
        />
      </aside>

      <!-- 右侧商品列表区域 -->
      <div class="shop-catalog-main">
        <section class="shop-page-product-collection-display-card shop-products-panel rounded-xl p-6 text-sm tz-text-secondary shadow-[8px_8px_22px_rgba(0,0,0,0.92)]">
          <div v-if="loading" class="shop-products-state py-12">
            <p class="tz-text-secondary text-sm">{{ $t('shopPage.products.loading', 'Loading products...') }}</p>
          </div>

          <div v-else-if="error" class="shop-products-state py-8 text-center text-red-300 text-sm">
            {{ error }}
          </div>

          <div v-else-if="products.length === 0" class="shop-products-state py-10 text-center space-y-2">
            <p class="tz-text-primary">{{ emptyProductsTitle }}</p>
            <p class="tz-text-secondary text-xs">
              {{ emptyProductsDescription }}
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
                <div v-else class="w-full h-full flex items-center justify-center tz-text-muted text-2xl">
                  📦
                </div>
              </div>
              <div class="px-3 pt-2 pb-3 flex-1 flex flex-col">
                <h3 class="text-xs font-semibold text-white line-clamp-2 mb-1">
                  {{ product.title }}
                </h3>
                <p v-if="product.priceLabel" class="text-xs text-[#40ffaa] mb-2">
                  {{ product.priceLabel }}
                </p>
                <div class="mt-auto flex gap-1.5 items-center">
                  <button
                    type="button"
                    @click="handleAddToWishlist(product)"
                    class="w-8 h-8 flex items-center justify-center rounded-full border border-white/25 tz-text-secondary hover:bg-white/15 transition-colors"
                    :title="$t('shopPage.actions.addToWishlist', 'Add to wishlist')"
                    :aria-label="$t('shopPage.actions.addToWishlist', 'Add to wishlist')"
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
                    class="flex-1 px-2 py-1.5 bg-white/10 hover:bg-white/20 border border-white/20 hover:border-white/40 rounded tz-caption text-white text-center transition-all"
                  >
                    {{ $t('shopPage.actions.view', 'View') }}
                  </NuxtLink>
                </div>
              </div>
            </div>
          </div>

          <div v-if="showPagination" class="shop-pagination">
            <button
              type="button"
              class="shop-pagination__button"
              :disabled="!canGoPrevious"
              @click="goToProductPage(productPagination.page - 1)"
            >
              <Icon name="lucide:chevron-left" />
              <span>{{ $t('common.previous', 'Previous') }}</span>
            </button>

            <span class="shop-pagination__count">
              {{ productPagination.page }} / {{ productPagination.totalPages }}
            </span>

            <button
              type="button"
              class="shop-pagination__button"
              :disabled="!canGoNext"
              @click="goToProductPage(productPagination.page + 1)"
            >
              <span>{{ $t('common.next', 'Next') }}</span>
              <Icon name="lucide:chevron-right" />
            </button>
          </div>
        </section>
      </div>
    </section>

    <section class="mt-10">
      <UserFeedbackThread
        threadKey="shop-page"
        :title="$t('shopPage.feedback.title', 'Share your feedback')"
      />
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute, useAsyncData } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import ShopCategoryVerticalMenu from '~/components/shop/ShopCategoryVerticalMenu.vue'
import { useWishlist } from '~/composables/useWishlist'
import { useShopCategories } from '~/composables/useShopCategories'
import type { ShopCategory } from '~/composables/useShopCategories'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'
import { useShopProducts } from '~/composables/useShopProducts'
import type { ShopProduct } from '~/composables/useShopProducts'
import { usePaymentCurrencies } from '~/composables/usePaymentCurrencies'

definePageMeta({
  layout: 'products',
})

const route = useRoute()
const { t } = useI18n()
const { fetchShopProducts } = useShopProducts()

const SHOP_PRODUCTS_PAGE_SIZE = 24
const quickFreeTextQuery = ref('')
const quickSearchQuery = ref('')
const quickPriceMin = ref(0)
const quickPriceMax = ref(5000)
const quickCurrency = ref('')
const categoryFilterOpen = ref(false)
const currentProductPage = ref(1)

// 商品心愿单
const { addToWishlist } = useWishlist()

// 商品分类
const { categories, loading: categoriesLoading, error: categoriesError, loadCategories } = useShopCategories()
const selectedCategory = ref<ShopCategory | null>(null)

interface ProductSearchFiltersPayload {
  priceRange: [number, number]
  currency?: string
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

const { pendingSearch, presetCategorySlug } = useShopSearchSheet()
const {
  currencies: paymentCurrencies,
  loading: paymentCurrenciesLoading,
  loadCurrencies: loadPaymentCurrencies,
} = usePaymentCurrencies()

const currencySelectLabel = computed(() => {
  if (paymentCurrenciesLoading.value) return '...'
  return t('filter.currency', 'Currency')
})

const openFilters = () => {
  categoryFilterOpen.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'shop-category-filter' } }))
  }
}

const closeCategoryFilter = () => {
  categoryFilterOpen.value = false
}

const syncQuickSearchQuery = () => {
  quickSearchQuery.value = quickFreeTextQuery.value.trim()
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
  selectedCategory.value && !selectedCategory.value.isProductType
    ? selectedCategory.value.name || selectedCategory.value.slug
    : null,
  payload?.chipCategorySlug && !categories.value.some(category => category.slug === payload.chipCategorySlug)
    ? categorySlugToKeyword(payload.chipCategorySlug)
    : null,
])

const normalizedQuickPriceRange = (): [number, number] => {
  const fallbackMin = DEFAULT_QUICK_FILTERS.priceRange[0]
  const fallbackMax = DEFAULT_QUICK_FILTERS.priceRange[1]
  const rawMin = Number(quickPriceMin.value)
  const rawMax = Number(quickPriceMax.value)
  const min = Number.isFinite(rawMin) ? Math.max(0, rawMin) : fallbackMin
  const max = Number.isFinite(rawMax) ? Math.max(0, rawMax) : fallbackMax

  return min <= max ? [min, max] : [max, min]
}

const buildQuickFilters = (): ProductSearchFiltersPayload => {
  const filters: ProductSearchFiltersPayload = {
    priceRange: normalizedQuickPriceRange(),
    attributes: {},
  }

  if (quickCurrency.value && paymentCurrencies.value.includes(quickCurrency.value)) {
    filters.currency = quickCurrency.value
  }

  return filters
}

const runQuickSearch = () => {
  syncQuickSearchQuery()
  handleSearch({
    query: quickSearchQuery.value,
    filters: buildQuickFilters(),
  })
}

const buildProductQueryParams = (payload?: ProductSearchPayload) => {
  const params: Record<string, any> = {
    page: currentProductPage.value,
    per_page: SHOP_PRODUCTS_PAGE_SIZE,
    status: 'active',
  }

  if (selectedCategory.value?.slug && selectedCategory.value.isProductType) {
    params.product_type = selectedCategory.value.slug
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

    if (payload.filters?.currency) {
      params.currency = payload.filters.currency
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
    const params = buildProductQueryParams(currentSearch.value || undefined)
    Object.assign(params, route.query)

    return fetchShopProducts(params)
  }
)

watch(() => route.query, () => {
  refresh()
}, { deep: true })

const products = computed<ShopProduct[]>(() => {
  const response = asyncData.value
  if (response && Array.isArray(response.items)) {
    return response.items
  }
  return []
})

const productPagination = computed(() => {
  const raw = asyncData.value?.raw as any
  const meta = raw?.meta && typeof raw.meta === 'object' ? raw.meta : {}
  const page = Math.max(1, Number(meta.page || currentProductPage.value || 1))
  const perPage = Math.max(1, Number(meta.per_page || SHOP_PRODUCTS_PAGE_SIZE))
  const total = Math.max(0, Number(meta.total || products.value.length || 0))
  const totalPages = Math.max(1, Number(meta.total_pages || Math.ceil(total / perPage) || 1))

  return {
    page,
    perPage,
    total,
    totalPages,
  }
})
const showPagination = computed(() => productPagination.value.totalPages > 1)
const canGoPrevious = computed(() => productPagination.value.page > 1)
const canGoNext = computed(() => productPagination.value.page < productPagination.value.totalPages)

const loading = computed(() => pending.value)
const error = computed(() => asyncError.value?.message || null)
const emptyProductsTitle = computed(() => {
  if (selectedCategory.value) {
    return t('shopPage.products.empty.categoryTitle', { category: selectedCategory.value.name })
  }

  return t('shopPage.products.empty.title')
})
const emptyProductsDescription = computed(() => {
  if (selectedCategory.value) {
    return t('shopPage.products.empty.categoryDescription')
  }

  return t('shopPage.products.empty.description')
})

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

  currentProductPage.value = 1
  currentSearch.value = next
  loadProducts(next)
}

const onCategorySelect = (category: ShopCategory | null) => {
  selectedCategory.value = category

  const base: ProductSearchPayload =
    currentSearch.value || ({
      query: '',
      filters: buildQuickFilters(),
    } as ProductSearchPayload)

  const next: ProductSearchPayload = {
    ...base,
  }

  currentProductPage.value = 1
  currentSearch.value = next
  loadProducts(next)
}

const goToProductPage = (page: number) => {
  const totalPages = productPagination.value.totalPages
  const nextPage = Math.min(Math.max(1, page), totalPages)
  if (nextPage === productPagination.value.page && nextPage === currentProductPage.value) return

  currentProductPage.value = nextPage
  loadProducts(currentSearch.value || undefined)
}

const onMobileCategorySelect = (category: ShopCategory | null) => {
  onCategorySelect(category)
  closeCategoryFilter()
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
  await Promise.all([
    loadCategories(),
    loadPaymentCurrencies(),
  ])

  if (!quickCurrency.value && paymentCurrencies.value.length > 0) {
    quickCurrency.value = paymentCurrencies.value[0]
  }

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

watch(paymentCurrencies, (currencies) => {
  if (!currencies.length) {
    quickCurrency.value = ''
    return
  }

  if (!quickCurrency.value || !currencies.includes(quickCurrency.value)) {
    quickCurrency.value = currencies[0]
  }
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
  --shop-catalog-left-gutter: clamp(0.5rem, 1.2vw, 1.5rem);
  --shop-category-rail-width: clamp(10rem, 16vw, 18rem);
  --shop-catalog-gap: clamp(1rem, 2vw, 2.5rem);
  --shop-products-min-height: clamp(22rem, 48vh, 34rem);
  padding-inline: 1.5rem;
}

.shop-catalog-layout {
  width: 100vw;
  margin-inline: calc(50% - 50vw);
  display: grid;
  grid-template-columns: var(--shop-category-rail-width) minmax(0, 1fr);
  gap: var(--shop-catalog-gap);
  align-items: stretch;
  padding-inline: var(--shop-catalog-left-gutter) clamp(1rem, 2.5vw, 3rem);
  box-sizing: border-box;
}

.shop-category-rail {
  min-height: var(--shop-products-min-height);
  display: flex;
  align-items: center;
  align-self: stretch;
  overflow-y: auto;
  scrollbar-width: none;
}

.shop-category-rail::-webkit-scrollbar {
  display: none;
}

.shop-catalog-main {
  min-width: 0;
}

.shop-products-panel {
  min-height: var(--shop-products-min-height);
  display: flex;
  flex-direction: column;
}

.shop-products-state {
  flex: 1;
  display: flex;
  min-height: 14rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

@media (min-width: 769px) {
  .shop-catalog-main {
    grid-column: 2;
  }
}

@media (max-width: 768px) {
  .shop-page {
    padding-inline: 0;
  }

  .shop-catalog-layout {
    width: auto;
    margin-inline: 0;
    display: block;
    padding-inline: 0;
  }

  .shop-category-rail {
    display: none;
  }

  .shop-products-panel {
    min-height: 18rem;
  }
}

.shop-search-form {
  display: grid;
  grid-template-columns: minmax(14rem, 1fr) minmax(18rem, 24rem) auto;
  gap: 8px;
  align-items: stretch;
}

.shop-search-input-shell {
  min-width: 0;
  display: flex;
  align-items: center;
  padding: 0 14px;
  background: linear-gradient(135deg, rgba(15,23,42,0.98), rgba(15,23,42,0.96));
  border-radius: 10px;
  box-shadow:
    0 2px 6px -3px rgba(0,0,0,0.9),
    0 0 6px rgba(15,23,42,0.7);
}

.shop-search-input-inner {
  width: 100%;
  min-width: 0;
  height: 38px;
  border: none;
  background: transparent;
  color: #ffffff;
  font-size: 13px;
  outline: none;
}

.shop-search-input-inner::placeholder {
  color: var(--tz-text-muted);
}

.shop-price-range {
  display: grid;
  grid-template-columns: auto minmax(5.5rem, 0.72fr) minmax(0, 1fr) minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  min-width: 0;
  min-height: 38px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.62);
  padding: 4px 8px;
}

.shop-price-range__label {
  color: rgba(226, 232, 240, 0.72);
  font-size: var(--tz-type-micro-label);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
}

.shop-currency-field,
.shop-price-field {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 5px;
  border-radius: 8px;
  background: rgba(2, 6, 23, 0.34);
  padding: 4px 7px;
}

.shop-currency-field select {
  width: 100%;
  min-width: 0;
  border: none;
  background: transparent;
  color: #ffffff;
  font-size: 13px;
  font-weight: 800;
  outline: none;
}

.shop-currency-field select:disabled {
  color: rgba(148, 163, 184, 0.64);
  cursor: not-allowed;
}

.shop-currency-field option {
  background: #020617;
  color: #ffffff;
}

.shop-price-field span {
  color: rgba(203, 213, 225, 0.78);
  font-size: var(--tz-type-micro-label);
  font-weight: 750;
  white-space: nowrap;
}

.shop-price-field input {
  width: 100%;
  min-width: 0;
  border: none;
  background: transparent;
  color: #ffffff;
  font-size: 13px;
  outline: none;
}

.shop-price-field input::-webkit-outer-spin-button,
.shop-price-field input::-webkit-inner-spin-button {
  margin: 0;
  appearance: none;
}

.shop-search-actions {
  display: flex;
  gap: 8px;
}

.shop-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-top: auto;
  padding-top: 18px;
}

.shop-pagination__button {
  display: inline-flex;
  min-height: 34px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
  padding: 0 12px;
  font-size: var(--tz-type-caption);
  font-weight: 800;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease,
    opacity 0.18s ease;
}

.shop-pagination__button:hover:not(:disabled) {
  border-color: rgba(255, 255, 255, 0.34);
  background: rgba(255, 255, 255, 0.14);
}

.shop-pagination__button:disabled {
  cursor: not-allowed;
  opacity: 0.42;
}

.shop-pagination__button :deep(svg) {
  width: 15px;
  height: 15px;
}

.shop-pagination__count {
  min-width: 4.5rem;
  text-align: center;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: var(--tz-type-caption);
  font-weight: 850;
  color: rgba(226, 232, 240, 0.82);
}

.shop-search-submit,
.shop-filter-button {
  min-height: 38px;
  border-radius: 10px;
  padding: 0 16px;
  font-weight: 800;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease,
    color 0.18s ease,
    box-shadow 0.18s ease;
}

.shop-search-submit {
  background: #ffffff;
  color: #000000;
  box-shadow: 8px 8px 22px rgba(0,0,0,0.92);
}

.shop-search-submit:hover {
  box-shadow: 10px 10px 26px rgba(0,0,0,0.95);
}

.shop-filter-button {
  display: none;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
}

.shop-filter-button:hover {
  border-color: rgba(255, 255, 255, 0.4);
  background: rgba(255, 255, 255, 0.15);
}

.shop-category-sheet {
  position: fixed;
  inset: 0;
  z-index: 1700;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: transparent;
  padding: 0 max(0.75rem, env(safe-area-inset-right)) max(0.75rem, env(safe-area-inset-bottom)) max(0.75rem, env(safe-area-inset-left));
}

.shop-category-sheet__panel {
  width: min(100%, 30rem);
  max-height: min(78dvh, 680px);
  overflow-y: auto;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 22px 22px 16px 16px;
  background:
    radial-gradient(circle at top left, rgba(64, 255, 170, 0.12), transparent 42%),
    linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(2, 6, 23, 0.98));
  padding: 14px;
  box-shadow: 0 30px 70px -28px rgba(0, 0, 0, 1);
}

.shop-category-sheet__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  color: #ffffff;
  font-size: 14px;
  font-weight: 850;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.shop-category-sheet__close {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

.shop-category-sheet__close :deep(svg) {
  width: 18px;
  height: 18px;
}

.shop-category-sheet__panel :deep(.shop-category-menu) {
  width: 100%;
}

.shop-category-sheet__panel :deep(.shop-category-menu__list) {
  gap: 0.9rem;
}

.shop-category-sheet-enter-active,
.shop-category-sheet-leave-active {
  transition: opacity 0.2s ease;
}

.shop-category-sheet-enter-active .shop-category-sheet__panel,
.shop-category-sheet-leave-active .shop-category-sheet__panel {
  transition: transform 0.22s ease;
}

.shop-category-sheet-enter-from,
.shop-category-sheet-leave-to {
  opacity: 0;
}

.shop-category-sheet-enter-from .shop-category-sheet__panel,
.shop-category-sheet-leave-to .shop-category-sheet__panel {
  transform: translateY(18px);
}

@media (max-width: 400px) {
  .shop-page {
    padding-inline: 0;
  }
}

@media (max-width: 768px) {
  .shop-search-form {
    grid-template-columns: 1fr;
  }

  .shop-price-range {
    grid-template-columns: auto minmax(5.5rem, 0.8fr) minmax(0, 1fr) minmax(0, 1fr);
  }

  .shop-search-actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  .shop-filter-button {
    display: inline-flex;
  }
}

@media (max-width: 430px) {
  .shop-price-range {
    grid-template-columns: minmax(0, 0.9fr) minmax(0, 1fr) minmax(0, 1fr);
  }

  .shop-price-range__label {
    grid-column: 1 / -1;
  }
}
</style>


