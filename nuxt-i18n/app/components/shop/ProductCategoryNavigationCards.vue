<template>
  <section
    class="product-category-navigation-cards"
    :class="`product-category-navigation-cards--${density}`"
    :style="productCategoryNavigationStyle"
    :aria-label="productCategoryNavigationHeadingText"
  >
    <header v-if="showHeader" class="product-category-navigation-cards__header">
      <span class="product-category-navigation-cards__title">
        {{ productCategoryNavigationHeadingText }}
      </span>
      <NuxtLink
        class="product-category-navigation-cards__all-link"
        :to="allProductCategoriesRoute"
        @click="emitProductCategoryNavigation()"
      >
        <span>{{ allProductCategoriesLabelText }}</span>
      </NuxtLink>
    </header>

    <div v-if="resolvedProductCategoryLoading" class="product-category-navigation-cards__state">
      {{ productCategoryLoadingText }}
    </div>

    <div
      v-else-if="resolvedProductCategoryError"
      class="product-category-navigation-cards__state product-category-navigation-cards__state--error"
    >
      {{ resolvedProductCategoryError }}
    </div>

    <div v-else-if="visibleProductCategories.length === 0" class="product-category-navigation-cards__state">
      {{ emptyProductCategoryText }}
    </div>

    <ul v-else class="product-category-navigation-cards__list">
      <li
        v-for="productCategory in visibleProductCategories"
        :key="productCategory.id"
        class="product-category-navigation-cards__list-item"
      >
        <NuxtLink
          class="product-category-navigation-cards__item"
          :to="buildProductCategoryNavigationRoute(productCategory)"
          @click="emitProductCategoryNavigation(productCategory)"
        >
          <span class="product-category-navigation-cards__media" aria-hidden="true">
            <StorefrontImage
              v-if="resolveProductCategoryImageSource(productCategory)"
              class="product-category-navigation-cards__image"
              :src="resolveProductCategoryImageSource(productCategory)"
              :alt="productCategory.name"
              preset="card"
              @error="handleProductCategoryImageError(productCategory.id)"
            />
            <span v-else class="product-category-navigation-cards__placeholder">
              <Icon name="lucide:image" />
            </span>
          </span>

          <span class="product-category-navigation-cards__footer">
            <span class="product-category-navigation-cards__name">{{ productCategory.name }}</span>
            <span class="product-category-navigation-cards__meta">
              <span
                v-if="formatProductCategoryProductCount(productCategory)"
                class="product-category-navigation-cards__count"
              >
                {{ formatProductCategoryProductCount(productCategory) }}
              </span>
              <Icon name="lucide:arrow-up-right" aria-hidden="true" />
            </span>
          </span>
        </NuxtLink>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { useProductCategories } from '~/composables/useProductCategories'

type ProductCategoryNavigationCardsDensity = 'compact' | 'comfortable'
type ProductCategoryNavigationRouteLocation = string | {
  path: string
  query: Record<string, string>
}
type ProductCategoryNavigationCardCategory = {
  id: number
  slug: string
  name: string
  routePath?: string
  count?: number
  isProductSpecificationTemplate?: boolean
  image?: string
  imageUrl?: string
  image_url?: string
  thumbnail?: string
  thumbnailUrl?: string
  thumbnail_url?: string
  coverImage?: string
  coverImageUrl?: string
  cover_image?: string
  featuredImage?: string
  featured_image?: string
}

const props = withDefaults(defineProps<{
  productCategories?: ProductCategoryNavigationCardCategory[]
  productCategoriesLoading?: boolean
  productCategoriesError?: string | null
  productCategoryDisplayLimit?: number
  productCategoryBasePath?: string
  productCategoryQueryParameterName?: string
  density?: ProductCategoryNavigationCardsDensity
  columns?: number
  showHeader?: boolean
  heading?: string
  allProductCategoriesLabel?: string
  loadingLabel?: string
  emptyLabel?: string
}>(), {
  productCategoryDisplayLimit: 0,
  productCategoryBasePath: '/shop',
  productCategoryQueryParameterName: 'product_category',
  density: 'comfortable',
  columns: 0,
  showHeader: true,
  heading: 'Product categories',
  loadingLabel: 'Loading categories...',
  emptyLabel: 'No categories available.',
})

const emit = defineEmits<{
  navigate: []
  categoryNavigate: [productCategory: ProductCategoryNavigationCardCategory]
}>()

const { t } = useI18n()
const localePath = useLocalePath()
const brokenProductCategoryImageIds = ref<number[]>([])
const {
  tree: loadedProductCategories,
  loading: loadedProductCategoriesLoading,
  error: loadedProductCategoriesError,
  loadCategories: loadProductCategories,
} = useProductCategories()

const hasExternallyProvidedProductCategories = computed(() => {
  return Array.isArray(props.productCategories)
})

const resolvedProductCategories = computed(() => {
  return hasExternallyProvidedProductCategories.value
    ? props.productCategories || []
    : loadedProductCategories.value
})

const resolvedProductCategoryLoading = computed(() => {
  return hasExternallyProvidedProductCategories.value
    ? Boolean(props.productCategoriesLoading)
    : loadedProductCategoriesLoading.value
})

const resolvedProductCategoryError = computed(() => {
  return hasExternallyProvidedProductCategories.value
    ? props.productCategoriesError || null
    : loadedProductCategoriesError.value
})

const visibleProductCategories = computed(() => {
  const categories = resolvedProductCategories.value
  const displayLimit = props.productCategoryDisplayLimit
  if (!displayLimit || displayLimit <= 0) return categories
  return categories.slice(0, displayLimit)
})

const normalizedProductCategoryColumns = computed(() => {
  const columns = Number(props.columns)
  if (!Number.isInteger(columns) || columns <= 0) return 0
  return Math.min(columns, 8)
})

const productCategoryNavigationStyle = computed(() => {
  if (!normalizedProductCategoryColumns.value) return undefined
  return {
    '--product-category-navigation-columns': String(normalizedProductCategoryColumns.value),
  }
})

const productCategoryNavigationHeadingText = computed(() => {
  return props.heading || (t('productCategoryNavigationCards.heading', 'Product categories') as string)
})

const allProductCategoriesLabelText = computed(() => {
  return props.allProductCategoriesLabel || (t('productCategoryNavigationCards.all', 'View spokes, tires, and more') as string)
})

const productCategoryLoadingText = computed(() => {
  return props.loadingLabel || (t('productCategoryNavigationCards.loading', 'Loading categories...') as string)
})

const emptyProductCategoryText = computed(() => {
  return props.emptyLabel || (t('productCategoryNavigationCards.empty', 'No categories available.') as string)
})

const localizedProductCategoryBasePath = computed(() => {
  return localePath(props.productCategoryBasePath || '/shop')
})

const allProductCategoriesRoute = computed(() => localizedProductCategoryBasePath.value)

const buildProductCategoryNavigationRoute = (
  productCategory: ProductCategoryNavigationCardCategory,
): ProductCategoryNavigationRouteLocation => {
  if (productCategory.routePath) return productCategory.routePath

  const queryParameterName = props.productCategoryQueryParameterName
  if (!queryParameterName) return localizedProductCategoryBasePath.value

  return {
    path: localizedProductCategoryBasePath.value,
    query: {
      [queryParameterName]: productCategory.slug,
    },
  }
}

const formatProductCategoryProductCount = (productCategory: ProductCategoryNavigationCardCategory) => {
  if (typeof productCategory.count !== 'number' || productCategory.count < 0) return ''
  return productCategory.count === 1 ? '1 product' : `${productCategory.count} products`
}

const resolveProductCategoryImageSource = (productCategory: ProductCategoryNavigationCardCategory) => {
  if (brokenProductCategoryImageIds.value.includes(productCategory.id)) return ''

  const candidateSources = [
    productCategory.image,
    productCategory.imageUrl,
    productCategory.image_url,
    productCategory.thumbnail,
    productCategory.thumbnailUrl,
    productCategory.thumbnail_url,
    productCategory.coverImage,
    productCategory.coverImageUrl,
    productCategory.cover_image,
    productCategory.featuredImage,
    productCategory.featured_image,
  ]

  const source = candidateSources.find((value): value is string => {
    return typeof value === 'string' && value.trim().length > 0
  })

  return source ? source.trim() : ''
}

const handleProductCategoryImageError = (productCategoryId: number) => {
  if (brokenProductCategoryImageIds.value.includes(productCategoryId)) return
  brokenProductCategoryImageIds.value = [...brokenProductCategoryImageIds.value, productCategoryId]
}

const emitProductCategoryNavigation = (productCategory?: ProductCategoryNavigationCardCategory) => {
  if (productCategory) {
    emit('categoryNavigate', productCategory)
  }
  emit('navigate')
}

if (import.meta.server && !hasExternallyProvidedProductCategories.value) {
  await loadProductCategories().catch(() => [])
}

onMounted(() => {
  if (!hasExternallyProvidedProductCategories.value) {
    void loadProductCategories()
  }
})
</script>

<style scoped>
.product-category-navigation-cards {
  --product-category-navigation-accent: var(--tz-site-accent, #059669);

  display: flex;
  height: 100%;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  gap: 12px;
  color: var(--tz-text-primary);
}

.product-category-navigation-cards__header {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.product-category-navigation-cards__title {
  min-width: 0;
  color: var(--tz-text-primary);
  font-size: 13px;
  font-weight: 850;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.product-category-navigation-cards__title::before {
  display: inline-block;
  width: 7px;
  height: 7px;
  margin-right: 8px;
  border-radius: 999px;
  background: var(--product-category-navigation-accent);
  content: "";
  vertical-align: 1px;
}

.product-category-navigation-cards__all-link {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
  max-width: 100%;
  padding: 6px 14px;
  border: 1px solid rgba(5, 150, 105, 0.48);
  border-radius: 999px;
  background: rgba(5, 150, 105, 0.18);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
  color: var(--tz-text-primary);
  font-size: 12px;
  font-weight: 850;
  letter-spacing: 0;
  line-height: 1.15;
  text-decoration: none;
  text-transform: uppercase;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease;
}

.product-category-navigation-cards__all-link span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-category-navigation-cards__all-link:hover,
.product-category-navigation-cards__all-link:focus-visible {
  border-color: rgba(5, 150, 105, 0.75);
  background: rgba(5, 150, 105, 0.28);
  color: var(--tz-text-primary);
}

.product-category-navigation-cards__list {
  display: grid;
  grid-template-columns: repeat(var(--product-category-navigation-columns, auto-fit), minmax(min(12rem, 100%), 1fr));
  gap: 12px;
  min-height: 0;
  margin: 0;
  padding: 0;
  list-style: none;
}

.product-category-navigation-cards--compact .product-category-navigation-cards__list {
  flex: 1 1 auto;
  grid-template-columns: repeat(var(--product-category-navigation-columns, 4), minmax(0, 1fr));
  gap: 8px;
  overflow: hidden;
}

.product-category-navigation-cards__list-item {
  min-width: 0;
}

.product-category-navigation-cards__item {
  position: relative;
  display: flex;
  min-width: 0;
  aspect-ratio: 1 / 1;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(20, 32, 43, 0.12);
  border-radius: 14px;
  background: var(--tz-card-surface);
  box-shadow: 0 8px 20px rgba(20, 32, 43, 0.1);
  color: var(--tz-text-primary);
  text-decoration: none;
  transition: border-color 0.18s ease;
}

.product-category-navigation-cards__item:hover,
.product-category-navigation-cards__item:focus-visible {
  border-color: rgba(5, 150, 105, 0.62);
}

.product-category-navigation-cards__item:focus-visible {
  outline: 2px solid rgba(5, 150, 105, 0.58);
  outline-offset: 2px;
}

.product-category-navigation-cards__media {
  position: relative;
  display: block;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  border-bottom: 1px solid rgba(20, 32, 43, 0.08);
  background: var(--tz-form-panel-surface);
}

.product-category-navigation-cards__image {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition:
    filter 0.18s ease,
    transform 0.22s ease;
}

.product-category-navigation-cards__item:hover .product-category-navigation-cards__image,
.product-category-navigation-cards__item:focus-visible .product-category-navigation-cards__image {
  filter: none;
  transform: none;
}

.product-category-navigation-cards__placeholder {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  color: var(--tz-text-muted);
}

.product-category-navigation-cards__placeholder :deep(svg) {
  width: 28px;
  height: 28px;
}

.product-category-navigation-cards__footer {
  position: relative;
  display: flex;
  min-height: 74px;
  flex: 0 0 auto;
  flex-direction: column;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px;
  background: var(--tz-card-surface);
}

.product-category-navigation-cards--compact .product-category-navigation-cards__footer {
  min-height: 48px;
  gap: 5px;
  padding: 8px;
}

.product-category-navigation-cards--compact .product-category-navigation-cards__footer::before {
  left: 8px;
  width: 22px;
}

.product-category-navigation-cards--compact .product-category-navigation-cards__name {
  font-size: 13px;
}

.product-category-navigation-cards--compact .product-category-navigation-cards__count {
  font-size: 11px;
}

.product-category-navigation-cards__footer::before {
  position: absolute;
  left: 12px;
  top: 0;
  width: 28px;
  height: 2px;
  border-radius: 999px;
  background: var(--product-category-navigation-accent);
  content: "";
}

.product-category-navigation-cards__name {
  min-width: 0;
  display: -webkit-box;
  overflow: hidden;
  color: inherit;
  font-size: 16px;
  font-weight: 850;
  line-height: 1.2;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.product-category-navigation-cards__meta {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.product-category-navigation-cards__count {
  color: var(--tz-text-muted);
  font-size: 12px;
  font-weight: 750;
  line-height: 1;
  white-space: nowrap;
}

.product-category-navigation-cards__meta :deep(svg) {
  width: 14px;
  height: 14px;
  margin-left: auto;
  color: var(--tz-text-secondary);
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.product-category-navigation-cards__item:hover .product-category-navigation-cards__meta :deep(svg),
.product-category-navigation-cards__item:focus-visible .product-category-navigation-cards__meta :deep(svg) {
  color: var(--tz-text-primary);
  transform: none;
}

.product-category-navigation-cards__state {
  min-height: 44px;
  display: flex;
  align-items: center;
  border: 1px solid rgba(20, 32, 43, 0.12);
  border-radius: 8px;
  background: var(--tz-card-surface);
  padding: 0 12px;
  color: var(--tz-text-secondary);
  font-size: 14px;
  font-weight: 700;
  line-height: 1.35;
}

.product-category-navigation-cards__state--error {
  border-color: rgba(248, 113, 113, 0.38);
  color: #fecaca;
}

@media (max-width: 1100px) {
  .product-category-navigation-cards--compact .product-category-navigation-cards__list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .product-category-navigation-cards {
    height: auto;
    gap: 10px;
  }

  .product-category-navigation-cards__list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .product-category-navigation-cards--compact .product-category-navigation-cards__list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .product-category-navigation-cards__footer {
    min-height: 66px;
    padding: 10px;
  }

  .product-category-navigation-cards__footer::before {
    left: 10px;
  }

  .product-category-navigation-cards__name {
    font-size: 14px;
  }
}
</style>
