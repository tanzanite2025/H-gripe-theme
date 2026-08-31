<template>
  <section v-if="category" class="shop-page w-full pt-0 pb-16 space-y-6">
    <section class="rounded-xl tz-surface-card p-4 text-sm tz-text-secondary shadow-[8px_8px_22px_rgb(15_23_42_/_0.08)]">
      <ShopProductQuickSearchForm
        density="page"
        show-filter-button
        :initial-query="currentSearch?.query || ''"
        :initial-price-range="currentSearch?.filters?.priceRange || defaultSearchPriceRange"
        @submit="handleSearch"
        @filter-click="openCategorySidebar"
      />
    </section>

    <teleport to="body">
      <transition name="shop-category-sidebar">
        <div
          v-if="categorySidebarOpen"
          class="shop-category-sidebar"
          role="dialog"
          aria-modal="true"
          :aria-label="$t('filter.categories', 'Categories')"
          @click.self="closeCategorySidebar"
        >
          <section class="shop-category-sidebar__panel">
            <header class="shop-category-sidebar__header">
              <span>{{ $t('filter.categories', 'Categories') }}</span>
              <button
                type="button"
                class="shop-category-sidebar__close"
                aria-label="Close categories"
                @click="closeCategorySidebar"
              >
                <Icon name="lucide:x" />
              </button>
            </header>

            <ShopCategoryVerticalMenu
              :categories="categories"
              :selected="category"
              :loading="loadingCategories"
              :error="categoriesError"
              @select="onMobileCategorySelect"
            />
          </section>
        </div>
      </transition>
    </teleport>

    <section class="shop-catalog-layout">
      <aside class="shop-category-rail" :aria-label="$t('shopPage.categoryRailLabel', 'Shop category navigation')">
        <ShopCategoryVerticalMenu
          :categories="categories"
          :selected="category"
          :loading="loadingCategories"
          :error="categoriesError"
          @select="onCategorySelect"
        />
      </aside>

      <div class="shop-catalog-main">
        <section class="shop-page-product-collection-display-card shop-products-panel rounded-xl p-2 text-sm tz-text-secondary shadow-[8px_8px_22px_rgb(15_23_42_/_0.08)] md:p-6">
          <section class="product-category-page__products" aria-live="polite">
            <div v-if="pending" class="shop-products-state py-12">
              <p class="tz-text-secondary text-sm">{{ t('shopPage.products.loading', 'Loading products...') }}</p>
            </div>
            <div v-else-if="error" class="shop-products-state py-8 text-center text-red-300 text-sm">
              {{ error }}
            </div>
            <div v-else-if="products.length === 0" class="shop-products-state py-10 text-center space-y-2">
              <p class="tz-text-primary">
                {{ t('shopPage.products.empty.categoryTitle', 'No products found in this category.') }}
              </p>
              <p class="tz-text-secondary text-xs">
                {{ t('shopPage.products.empty.categoryDescription', 'Products will appear here automatically once they are published.') }}
              </p>
            </div>
            <div v-else class="tz-product-card-grid content-start">
              <ShopProductDisplayCard
                v-for="product in products"
                :key="product.id"
                :product="product"
                show-wishlist-action
                show-view-action
              />
            </div>
          </section>
        </section>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  createError,
  useAsyncData,
  useHead,
  useI18n,
  useLocalePath,
  useRouter,
  useRoute,
  useRequestURL,
  useRuntimeConfig,
} from '#imports'
import ShopProductQuickSearchForm from '~/components/shop/ShopProductQuickSearchForm.vue'
import ShopCategoryVerticalMenu from '~/components/shop/ShopCategoryVerticalMenu.vue'
import {
  useProductCategories,
  type ProductCategory,
} from '~/composables/useProductCategories'
import {
  useShopProducts,
  type ShopProduct,
  type ShopProductsResult,
} from '~/composables/useShopProducts'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'
import { useShopSearchSheet, type ShopSearchFiltersPayload, type ShopSearchPayload } from '~/composables/useShopSearchSheet'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { toAbsoluteSeoUrl } from '~/utils/seo/urls'

const props = withDefaults(defineProps<{
  categoryRoutePath?: string
}>(), {
  categoryRoutePath: '',
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const localePath = useLocalePath()
const requestUrl = useRequestURL()
const config = useRuntimeConfig()
const overlayBackStack = useOverlayBackStack()
const {
  categories,
  loadCategories,
  loading: loadingCategories,
  error: categoriesError,
} = useProductCategories()
const {
  fetchPublicShopProducts,
} = useShopProducts()
const { pendingSearch } = useShopSearchSheet()

const defaultSearchPriceRange: [number, number] = [0, 5000]
const currentSearch = ref<ShopSearchPayload | null>(null)
const categorySidebarOpen = ref(false)

const normalizePath = (value: unknown) => {
  const path = (String(value || '').split(/[?#]/, 1)[0] || '').trim()
  if (!path) return '/'
  const normalized = `/${path.replace(/^\/+|\/+$/g, '')}`
  return normalized === '//' ? '/' : normalized
}

const requestedRoutePath = computed(() => normalizePath(
  props.categoryRoutePath || route.path,
))

await loadCategories()

const category = computed<ProductCategory | null>(() => (
  categories.value.find((item) => normalizePath(item.routePath) === requestedRoutePath.value) || null
))

if (!category.value) {
  throw createError({
    statusCode: 404,
    statusMessage: 'Product category not found',
  })
}

const categoryPath = (item: ProductCategory) => {
  if (item.routePath) return item.routePath
  const base = category.value?.routePath || '/shop'
  return `${base.replace(/\/+$/, '')}/${encodeURIComponent(item.slug)}`
}

const createDefaultSearchFilters = (): ShopSearchFiltersPayload => ({
  priceRange: [...defaultSearchPriceRange] as [number, number],
  attributes: {},
})

const cloneSearchFilters = (filters?: ShopSearchFiltersPayload): ShopSearchFiltersPayload => ({
  priceRange: Array.isArray(filters?.priceRange)
    ? [...filters!.priceRange] as [number, number]
    : [...defaultSearchPriceRange] as [number, number],
  currency: filters?.currency,
  attributes: { ...(filters?.attributes || {}) },
})

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

const buildProductKeyword = (payload?: ShopSearchPayload) => joinUniqueSearchParts([payload?.query])

const buildProductQueryParams = (payload?: ShopSearchPayload) => {
  const params: Record<string, any> = {
    page: 1,
    page_size: 24,
    status: 'active',
    product_category: categorySlug.value,
  }

  if (payload) {
    const keyword = buildProductKeyword(payload)
    if (keyword) params.keyword = keyword

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

const onCategorySelect = async (nextCategory: ProductCategory | null) => {
  const targetPath = nextCategory ? categoryPath(nextCategory) : localePath('/shop')
  if (normalizePath(targetPath) === requestedRoutePath.value) return

  await router.replace(targetPath)
}

const openCategorySidebar = () => {
  categorySidebarOpen.value = true
  overlayBackStack.open('shop-category-sidebar', () => {
    categorySidebarOpen.value = false
  })
}

const closeCategorySidebar = () => {
  void overlayBackStack.close('shop-category-sidebar')
  categorySidebarOpen.value = false
}

const handleExternalCategorySidebarOpen = () => {
  openCategorySidebar()
}

const onMobileCategorySelect = async (nextCategory: ProductCategory | null) => {
  await onCategorySelect(nextCategory)
  closeCategorySidebar()
}

const loadProducts = async () => {
  await refresh()
}

const handleSearch = (payload: ShopSearchPayload) => {
  currentSearch.value = {
    ...payload,
    filters: cloneSearchFilters(payload.filters),
  }
  loadProducts()
}

const categorySlug = computed(() => category.value?.slug || '')
const productDataKey = computed(() => (
  `product-category:${requestedRoutePath.value}:${categorySlug.value}`
))

const siteOrigin = computed(() => {
  const configured = String((config.public as { siteUrl?: string }).siteUrl || '').trim()
  return (configured || requestUrl.origin).replace(/\/$/, '')
})

const canonicalUrl = computed(() => toAbsoluteSeoUrl(
  siteOrigin.value,
  category.value?.routePath || requestedRoutePath.value,
))

const { data: productData, pending, error: asyncError, refresh } = await useAsyncData<ShopProductsResult>(
  productDataKey,
  () => fetchPublicShopProducts(buildProductQueryParams(currentSearch.value || undefined)),
  {
    watch: [requestedRoutePath, categorySlug],
  },
)

const fallbackMetaDescription = computed(() => {
  const text = String(category.value?.intro || category.value?.description || '').replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()
  if (text.length <= 320) return text
  return `${text.slice(0, 317)}...`
})

const products = computed<ShopProduct[]>(() => productData.value?.items || [])
const error = computed(() => asyncError.value?.message || '')

const itemList = computed(() => ({
  '@type': 'ItemList',
  numberOfItems: products.value.length,
  itemListElement: products.value.map((product, index) => ({
    '@type': 'ListItem',
    position: index + 1,
    name: product.title,
    url: toAbsoluteSeoUrl(siteOrigin.value, localePath(product.url)),
  })),
}))

const categoryMetaDescription = computed(() => (
  category.value?.metaDescription?.trim() || fallbackMetaDescription.value
))

const categorySchema = computed(() => ({
  '@context': 'https://schema.org',
  '@type': 'CollectionPage',
  name: category.value?.name || t('products.nav.shop', 'Shop'),
  description: categoryMetaDescription.value || undefined,
  url: canonicalUrl.value,
  mainEntity: itemList.value,
}))

useHead(() => {
  const title = category.value?.metaTitle?.trim()
    || category.value?.name
    || t('products.nav.shop', 'Shop')
  const description = categoryMetaDescription.value
  return {
    title,
    meta: [
      { name: 'description', content: description },
      { property: 'og:title', content: title },
      { property: 'og:description', content: description },
      { property: 'og:type', content: 'website' },
      { property: 'og:url', content: canonicalUrl.value },
      { name: 'twitter:card', content: 'summary' },
      { name: 'twitter:title', content: title },
      { name: 'twitter:description', content: description },
    ].filter((entry) => entry.content),
    script: [createSeoJsonLdScript(categorySchema.value)],
  }
})

onMounted(() => {
  if (typeof window !== 'undefined') {
    window.addEventListener(
      'ui:product-category-sidebar-open',
      handleExternalCategorySidebarOpen,
    )
  }

  const initialPending = pendingSearch.value
  if (initialPending) {
    pendingSearch.value = null
    handleSearch(initialPending as unknown as ShopSearchPayload)
    return
  }

  loadProducts()
})

onBeforeUnmount(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener(
      'ui:product-category-sidebar-open',
      handleExternalCategorySidebarOpen,
    )
  }
})

watch(pendingSearch, (payload) => {
  if (!payload) return
  pendingSearch.value = null
  handleSearch(payload as unknown as ShopSearchPayload)
})
</script>

<style scoped>
.shop-page {
  --shop-catalog-left-gutter: clamp(0.5rem, 1.2vw, 1.5rem);
  --shop-category-rail-width: clamp(16rem, 20vw, 21rem);
  --shop-catalog-gap: clamp(1rem, 2vw, 2.5rem);
  padding-inline: 1.5rem;
}

.shop-page-product-collection-display-card {
  background-color: var(--tz-card-surface);
}

.shop-category-sidebar {
  display: none;
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
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.shop-products-panel {
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

.product-category-page__products {
  min-width: 0;
}

@media (min-width: 768px) {
  .shop-catalog-layout {
    width: 100%;
    margin-inline: 0;
    padding-inline: 0;
  }

  .shop-catalog-main {
    grid-column: 2;
  }

  .shop-category-rail {
    align-items: flex-start;
    box-sizing: border-box;
    padding: 1.5rem 0;
  }

  .shop-category-rail :deep(.shop-category-menu) {
    width: 100%;
    height: 100%;
    min-height: 0;
  }

  .shop-category-rail :deep(.shop-category-menu__state) {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    padding: 0 0.75rem;
    text-align: center;
  }
}

@media (min-width: 1024px) {
  .shop-page {
    --shop-category-rail-width: clamp(16rem, 24vw, 22rem);
    --shop-catalog-gap: clamp(1.25rem, 2vw, 2.5rem);
  }

  .shop-category-rail {
    align-items: stretch;
    overflow: hidden;
    padding: 1.25rem 0;
  }
}

@media (max-width: 768px) {
  .shop-page {
    padding-inline: 0;
  }

  .shop-category-sidebar {
    position: fixed;
    inset: 0;
    z-index: 1700;
    display: flex;
    background: rgb(15 23 42 / 0.2);
  }

  .shop-category-sidebar__panel {
    display: flex;
    width: min(86vw, 22rem);
    min-width: 17rem;
    height: 100%;
    flex-direction: column;
    overflow: hidden;
    border-right: 1px solid var(--tz-border-strong);
    background: var(--tz-card-surface);
    box-shadow: 24px 0 60px -28px rgb(15 23 42 / 0.16);
  }

  .shop-category-sidebar__header {
    display: flex;
    min-height: 3.5rem;
    flex: 0 0 auto;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--tz-border-subtle);
    color: var(--tz-text-primary);
    font-size: 14px;
    font-weight: 850;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .shop-category-sidebar__close {
    display: inline-flex;
    width: 34px;
    height: 34px;
    flex: 0 0 auto;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--tz-border-strong);
    border-radius: 999px;
    background: var(--tz-surface-subtle);
    color: var(--tz-text-primary);
  }

  .shop-category-sidebar__close:hover,
  .shop-category-sidebar__close:focus-visible {
    border-color: rgba(5, 150, 105, 0.64);
    background: var(--tz-site-accent-soft-surface);
  }

  .shop-category-sidebar__close:focus-visible {
    outline: 2px solid rgba(5, 150, 105, 0.72);
    outline-offset: 2px;
  }

  .shop-category-sidebar__close :deep(svg) {
    width: 18px;
    height: 18px;
  }

  .shop-category-sidebar__panel :deep(.shop-category-menu) {
    width: 100%;
    height: 100%;
    overflow-y: auto;
    padding: 0.75rem 0.75rem calc(1rem + env(safe-area-inset-bottom));
    scrollbar-width: thin;
    scrollbar-color: rgba(5, 150, 105, 0.46) transparent;
  }

  .shop-category-sidebar__panel :deep(.shop-category-menu__list) {
    gap: 0.45rem;
  }

  .shop-category-sidebar-enter-active,
  .shop-category-sidebar-leave-active {
    transition: opacity 0.2s ease;
  }

  .shop-category-sidebar-enter-active .shop-category-sidebar__panel,
  .shop-category-sidebar-leave-active .shop-category-sidebar__panel {
    transition: transform 0.22s ease;
  }

  .shop-category-sidebar-enter-from,
  .shop-category-sidebar-leave-to {
    opacity: 0;
  }

  .shop-category-sidebar-enter-from .shop-category-sidebar__panel,
  .shop-category-sidebar-leave-to .shop-category-sidebar__panel {
    transform: translateX(-100%);
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

@media (max-width: 400px) {
  .shop-page {
    padding-inline: 0;
  }
}
</style>
