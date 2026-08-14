<template>
  <main class="shop-page w-full pt-0 pb-16 space-y-6">
    <section class="rounded-xl bg-white/5 p-4 text-sm tz-text-secondary shadow-[8px_8px_22px_rgba(0,0,0,0.92)]">
      <ShopProductQuickSearchForm
        density="page"
        show-filter-button
        @submit="handleSearch"
        @filter-click="openFilters"
      />
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
            <ShopProductDisplayCard
              v-for="product in products"
              :key="product.id"
              :product="product"
              show-wishlist-action
              show-view-action
              @wishlist="handleAddToWishlist"
            />
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

            <span class="shop-pagination__count" :aria-label="$t('common.currentPage', 'Current page')">
              {{ productPagination.page }}
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

    <ProductRecommendations
      surface="shop_index_bottom"
      :title="$t('recommendations.shopTitle', 'Recommended products')"
      :category-id="shopRecommendationCategoryId"
      :query="shopRecommendationQuery"
      :exclude-product-ids="visibleProductIds"
      :limit="8"
    />

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
import { useRoute, useRouter, useAsyncData } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import ShopProductQuickSearchForm from '~/components/shop/ShopProductQuickSearchForm.vue'
import ShopCategoryVerticalMenu from '~/components/shop/ShopCategoryVerticalMenu.vue'
import ShopProductDisplayCard from '~/components/shop/ShopProductDisplayCard.vue'
import ProductRecommendations from '~/components/shop/ProductRecommendations.vue'
import { useWishlist } from '~/composables/useWishlist'
import { useShopCategories } from '~/composables/useShopCategories'
import type { ShopCategory } from '~/composables/useShopCategories'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'
import { useShopProducts } from '~/composables/useShopProducts'
import type { ShopProduct } from '~/composables/useShopProducts'

definePageMeta({
  layout: 'products',
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { fetchShopProducts } = useShopProducts()

const SHOP_PRODUCTS_PAGE_SIZE = 24
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

const createDefaultSearchFilters = (): ProductSearchFiltersPayload => ({
  priceRange: [0, 5000],
  attributes: {},
})

const { pendingSearch, presetCategorySlug } = useShopSearchSheet()

const routeProductTypeSlug = computed(() => {
  const value = route.query.product_type
  const raw = Array.isArray(value) ? value[0] : value
  return String(raw || '').trim()
})

const syncSelectedCategoryFromRoute = () => {
  const slug = routeProductTypeSlug.value
  if (!slug) {
    if (selectedCategory.value?.isProductType) {
      selectedCategory.value = null
    }
    return
  }

  const match = categories.value.find(category => category.slug === slug)
  if (match) {
    selectedCategory.value = match
    return
  }

  if (selectedCategory.value?.isProductType) {
    selectedCategory.value = null
  }
}

const replaceProductTypeRoute = async (category: ShopCategory | null) => {
  const nextSlug = category?.isProductType ? category.slug : ''
  if (routeProductTypeSlug.value === nextSlug) return false

  const nextQuery: Record<string, any> = { ...route.query }
  if (nextSlug) {
    nextQuery.product_type = nextSlug
  } else {
    delete nextQuery.product_type
  }

  await router.replace({
    path: route.path,
    query: nextQuery,
  })
  return true
}

if (import.meta.server) {
  await loadCategories().catch(() => [])
  syncSelectedCategoryFromRoute()
}

const openFilters = () => {
  categoryFilterOpen.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'shop-category-filter' } }))
  }
}

const closeCategoryFilter = () => {
  categoryFilterOpen.value = false
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
    const routeQuery = { ...route.query }
    delete routeQuery.currency
    Object.assign(params, routeQuery)

    return fetchShopProducts(params)
  }
)

watch(() => route.query, () => {
  syncSelectedCategoryFromRoute()
  refresh()
}, { deep: true })

const products = computed<ShopProduct[]>(() => {
  const response = asyncData.value
  if (response && Array.isArray(response.items)) {
    return response.items
  }
  return []
})

const visibleProductIds = computed(() => products.value.map(product => product.id).filter(id => Number.isInteger(id) && id > 0))
const shopRecommendationCategoryId = computed(() => {
  const categoryId = Number(selectedCategory.value?.id)
  return Number.isInteger(categoryId) && categoryId > 0 ? categoryId : null
})
const shopRecommendationQuery = computed(() => buildProductKeyword(currentSearch.value || undefined))

const productPagination = computed(() => {
  const raw = asyncData.value?.raw as any
  const page = Math.max(1, currentProductPage.value || 1)
  const hasMore = raw?.has_more === true

  return {
    page,
    hasMore,
  }
})
const showPagination = computed(() => productPagination.value.page > 1 || productPagination.value.hasMore)
const canGoPrevious = computed(() => productPagination.value.page > 1)
const canGoNext = computed(() => productPagination.value.hasMore)

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

const onCategorySelect = async (category: ShopCategory | null) => {
  selectedCategory.value = category

  const base: ProductSearchPayload =
    currentSearch.value || ({
      query: '',
      filters: createDefaultSearchFilters(),
    } as ProductSearchPayload)

  const next: ProductSearchPayload = {
    ...base,
  }

  currentProductPage.value = 1
  currentSearch.value = next
  const routeChanged = await replaceProductTypeRoute(category)
  if (!routeChanged) {
    loadProducts(next)
  }
}

const goToProductPage = (page: number) => {
  const nextPage = Math.max(1, page)
  if (nextPage === productPagination.value.page && nextPage === currentProductPage.value) return

  currentProductPage.value = nextPage
  loadProducts(currentSearch.value || undefined)
}

const onMobileCategorySelect = async (category: ShopCategory | null) => {
  await onCategorySelect(category)
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
  await loadCategories()
  syncSelectedCategoryFromRoute()

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
  syncSelectedCategoryFromRoute()
  applyPresetCategoryFromSlug()

  handleSearch(payload as unknown as ProductSearchPayload)
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
  font-family: 'StorefrontSystem';
  font-size: var(--tz-type-caption);
  font-weight: 850;
  color: rgba(226, 232, 240, 0.82);
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
    radial-gradient(circle at top left, rgba(181, 255, 109, 0.12), transparent 42%),
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

</style>
