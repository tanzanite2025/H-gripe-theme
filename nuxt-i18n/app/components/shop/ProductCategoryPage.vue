<template>
  <main v-if="category" class="product-category-page">
    <nav
      v-if="breadcrumbItems.length"
      class="product-category-breadcrumb"
      aria-label="Breadcrumb"
    >
      <ol class="product-category-breadcrumb__list">
        <li
          v-for="(item, index) in breadcrumbItems"
          :key="item.id"
          class="product-category-breadcrumb__item"
        >
          <span v-if="index > 0" class="product-category-breadcrumb__separator" aria-hidden="true">/</span>
          <NuxtLink
            v-if="item.path && index < breadcrumbItems.length - 1"
            :to="item.path"
            class="product-category-breadcrumb__link"
          >
            {{ item.name }}
          </NuxtLink>
          <span v-else class="product-category-breadcrumb__current">{{ item.name }}</span>
        </li>
      </ol>
    </nav>

    <header class="product-category-page__header">
      <p class="product-category-page__eyebrow">{{ t('filter.categories', 'Categories') }}</p>
      <h1>{{ category.name }}</h1>
      <SafeRichText
        v-if="categoryIntro"
        class="product-category-page__description tz-rich-text"
        :html="categoryIntro"
      />
    </header>

    <nav
      v-if="category.children.length"
      class="product-category-page__children"
      :aria-label="t('shopCategoryMenu.children', 'Subcategories')"
    >
      <NuxtLink
        v-for="child in category.children"
        :key="child.id"
        :to="categoryPath(child)"
        class="product-category-page__child-link"
      >
        <span>{{ child.name }}</span>
        <Icon name="lucide:arrow-up-right" aria-hidden="true" />
      </NuxtLink>
    </nav>

    <section class="product-category-page__products" aria-live="polite">
      <div v-if="pending" class="product-category-page__state">
        {{ t('shopPage.products.loading', 'Loading products...') }}
      </div>
      <div v-else-if="error" class="product-category-page__state product-category-page__state--error">
        {{ error }}
      </div>
      <div v-else-if="products.length === 0" class="product-category-page__state">
        {{ t('shopPage.products.empty.categoryTitle', 'No products found in this category.') }}
      </div>
      <div v-else class="tz-product-card-grid">
        <ShopProductDisplayCard
          v-for="product in products"
          :key="product.id"
          :product="product"
          show-wishlist-action
          show-view-action
        />
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  createError,
  useAsyncData,
  useHead,
  useI18n,
  useLocalePath,
  useRoute,
  useRequestURL,
  useRuntimeConfig,
} from '#imports'
import {
  useProductCategories,
  type ProductCategory,
} from '~/composables/useProductCategories'
import {
  useShopProducts,
  type ShopProduct,
  type ShopProductsResult,
} from '~/composables/useShopProducts'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { toAbsoluteSeoUrl } from '~/utils/seo/urls'

const props = withDefaults(defineProps<{
  categoryRoutePath?: string
}>(), {
  categoryRoutePath: '',
})

const route = useRoute()
const { t } = useI18n()
const localePath = useLocalePath()
const requestUrl = useRequestURL()
const config = useRuntimeConfig()
const {
  categories,
  loadCategories,
} = useProductCategories()
const {
  fetchPublicShopProducts,
} = useShopProducts()

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

const categoryMap = computed(() => new Map(
  categories.value.map((item) => [item.id, item]),
))

const ancestorCategories = computed<ProductCategory[]>(() => {
  const current = category.value
  if (!current) return []

  const result: ProductCategory[] = []
  const visited = new Set<number>()
  let parent: ProductCategory | undefined = current
  while (parent && !visited.has(parent.id)) {
    visited.add(parent.id)
    result.unshift(parent)
    parent = parent.parentId ? categoryMap.value.get(parent.parentId) : undefined
  }
  return result
})

const breadcrumbItems = computed(() => [
  {
    id: 'home',
    name: t('breadcrumbs.home', 'Home'),
    path: '/',
  },
  {
    id: 'shop',
    name: t('products.nav.shop', 'Shop'),
    path: '/shop',
  },
  ...ancestorCategories.value.map((item) => ({
    id: `category:${item.id}`,
    name: item.name,
    path: item.routePath,
  })),
])

const categoryPath = (item: ProductCategory) => {
  if (item.routePath) return item.routePath
  const base = category.value?.routePath || '/shop'
  return `${base.replace(/\/+$/, '')}/${encodeURIComponent(item.slug)}`
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

const categoryIntro = computed(() => (
  String(category.value?.intro || category.value?.description || '').trim()
))

const fallbackMetaDescription = computed(() => {
  const text = categoryIntro.value.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()
  if (text.length <= 320) return text
  return `${text.slice(0, 317)}...`
})

const { data: productData, pending, error: asyncError } = await useAsyncData<ShopProductsResult>(
  productDataKey,
  () => fetchPublicShopProducts({
    page: 1,
    page_size: 24,
    status: 'active',
    product_category: categorySlug.value,
  }),
  {
    watch: [requestedRoutePath, categorySlug],
  },
)

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
</script>

<style scoped>
.product-category-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 1.5rem 1rem 4rem;
  color: var(--tz-text-primary);
}

.product-category-breadcrumb {
  overflow-x: auto;
}

.product-category-breadcrumb__list {
  display: flex;
  min-width: max-content;
  align-items: center;
  gap: 0.45rem;
  margin: 0;
  padding: 0;
  list-style: none;
  color: var(--tz-text-secondary);
  font-size: 0.82rem;
}

.product-category-breadcrumb__item {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  white-space: nowrap;
}

.product-category-breadcrumb__separator {
  color: var(--tz-text-muted);
}

.product-category-breadcrumb__link {
  color: inherit;
  text-decoration: none;
}

.product-category-breadcrumb__link:hover,
.product-category-breadcrumb__link:focus-visible {
  color: var(--tz-site-accent);
}

.product-category-breadcrumb__current {
  color: var(--tz-text-primary);
  font-weight: 700;
}

.product-category-page__header {
  max-width: 56rem;
}

.product-category-page__eyebrow {
  margin: 0 0 0.4rem;
  color: var(--tz-site-accent);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.product-category-page__header h1 {
  margin: 0;
  font-size: clamp(1.8rem, 3vw, 3rem);
  font-weight: 700;
}

.product-category-page__description {
  max-width: 48rem;
  margin: 0.8rem 0 0;
  color: var(--tz-text-secondary);
  line-height: 1.65;
}

.product-category-page__children {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
  gap: 0.75rem;
}

.product-category-page__child-link {
  display: flex;
  min-height: 3.25rem;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  padding: 0.75rem 0.9rem;
  color: var(--tz-text-primary);
  text-decoration: none;
  transition: border-color 0.18s ease, color 0.18s ease, transform 0.18s ease;
}

.product-category-page__child-link:hover,
.product-category-page__child-link:focus-visible {
  border-color: var(--tz-site-accent);
  color: var(--tz-site-accent);
  transform: translateY(-1px);
}

.product-category-page__child-link :deep(svg) {
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
}

.product-category-page__products {
  min-width: 0;
}

.product-category-page__state {
  min-height: 14rem;
  display: grid;
  place-items: center;
  color: var(--tz-text-secondary);
  text-align: center;
}

.product-category-page__state--error {
  color: #b91c1c;
}

@media (min-width: 768px) {
  .product-category-page {
    padding-inline: 1.5rem;
  }
}
</style>
