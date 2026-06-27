<template>
  <section class="category-strip">
    <header class="category-strip__header">
      <h4 class="category-strip__title">{{ title }}</h4>
    </header>

    <div class="category-strip__body">
      <!-- Loading state -->
      <div v-if="loading" class="category-strip__loading">
        <span class="category-strip__spinner" />
        <span class="category-strip__loading-text">Loading {{ loadingLabel }}...</span>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="category-strip__error">
        {{ error }}
      </div>

      <!-- Empty state -->
      <div v-else-if="products.length === 0" class="category-strip__empty">
        <p>{{ resolvedEmptyMessage }}</p>
      </div>

      <!-- Product strip -->
      <div v-else class="category-strip__list-wrapper">
        <div
          ref="scrollContainer"
          class="category-strip__list scrollbar-hide"
          style="scrollbar-width: none; -ms-overflow-style: none;"
        >
          <article
            v-for="product in products"
            :key="product.id"
            class="category-strip__card"
          >
            <div class="category-strip__image">
              <img
                v-if="product.thumbnail"
                :src="product.thumbnail"
                :alt="product.title"
                loading="lazy"
              />
              <div v-else class="category-strip__image-placeholder">
                <span class="category-strip__image-icon">📦</span>
              </div>
            </div>

            <div class="category-strip__info">
              <h5 class="category-strip__name">{{ product.title }}</h5>
              <p v-if="product.priceLabel" class="category-strip__price">
                {{ product.priceLabel }}
              </p>

              <div class="category-strip__actions">
                <button
                  v-if="showAddToCart"
                  type="button"
                  class="category-strip__button category-strip__button--primary"
                  @click="handleAddToCart(product)"
                >
                  Add to cart
                </button>
                <NuxtLink
                  :to="product.url"
                  class="category-strip__button category-strip__button--ghost"
                >
                  View
                </NuxtLink>
              </div>
            </div>
          </article>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRuntimeConfig } from '#imports'
import { useCart } from '~/composables/useCart'

interface StripProduct {
  id: number
  title: string
  slug: string
  url: string
  thumbnail?: string
  priceNumber: number
  priceLabel: string
}

const props = defineProps<{
  categorySlug: string
  title: string
  perPage?: number
  emptyMessage?: string
  showAddToCart?: boolean
}>()

const products = ref<StripProduct[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const scrollContainer = ref<HTMLElement | null>(null)

const { addToCart, openCart } = useCart()

const perPage = computed(() => props.perPage ?? 12)
const showAddToCart = computed(() => props.showAddToCart ?? true)

const loadingLabel = computed(() => props.title.toLowerCase())
const resolvedEmptyMessage = computed(
  () => props.emptyMessage || 'No products are listed yet. Coming soon.'
)

const categoryKeyword = computed(() => props.categorySlug.replace(/[-_]+/g, ' ').trim())

const loadProducts = async () => {
  if (!categoryKeyword.value) {
    products.value = []
    return
  }

  loading.value = true
  error.value = null

  try {
    const config = useRuntimeConfig()
    const base = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')

    const productParams: Record<string, any> = {
      keyword: categoryKeyword.value,
      per_page: perPage.value,
      status: 'active',
    }

    const response = await $fetch<any>(`${base}/customer-service/products`, {
      params: productParams,
    })

    if (response && Array.isArray(response.items)) {
      products.value = response.items.map((item: any) => {
        const sale = Number(item?.prices?.sale || 0)
        const regular = Number(item?.prices?.regular || 0)
        const priceNumber = sale > 0 ? sale : regular > 0 ? regular : 0

        let priceLabel = ''
        if (priceNumber > 0) {
          priceLabel = `$${priceNumber}`
        }

        return {
          id: item.id,
          title: item.title,
          slug: item.slug || String(item.id),
          url: `/shop/${item.slug || item.id}`,
          thumbnail: item.thumbnail,
          priceNumber,
          priceLabel,
        } as StripProduct
      })
    } else {
      products.value = []
    }
  } catch (e: any) {
    console.error('Failed to load products:', e)
    error.value = e?.data?.message || 'Failed to load products.'
    products.value = []
  } finally {
    loading.value = false
  }
}

const handleAddToCart = (product: StripProduct) => {
  if (!product || !product.id) return

  const price = Number(product.priceNumber || 0)

  const result = addToCart({
    id: product.id,
    title: product.title,
    slug: product.slug,
    price,
    thumbnail: product.thumbnail,
  })

  if (result?.success) {
    openCart()
  }
}

onMounted(() => {
  loadProducts()
})

watch(
  () => props.categorySlug,
  () => {
    loadProducts()
  }
)
</script>

<style scoped>
.category-strip {
  margin-top: 1.25rem;
  padding: 0.75rem 0.75rem 0.9rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(148, 163, 184, 0.3);
  background: radial-gradient(circle at top left, rgba(56, 189, 248, 0.12), transparent 45%),
    rgba(15, 23, 42, 0.85);
}

.category-strip__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.35rem;
}

.category-strip__title {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  color: #e5e7eb;
}

.category-strip__body {
  font-size: 0.85rem;
}

.category-strip__loading,
.category-strip__error,
.category-strip__empty {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.2rem 0.1rem;
  color: rgba(148, 163, 184, 0.9);
}

.category-strip__spinner {
  width: 0.8rem;
  height: 0.8rem;
  border-radius: 9999px;
  border: 2px solid rgba(148, 163, 184, 0.4);
  border-top-color: rgba(56, 189, 248, 0.9);
  animation: strip-spin 0.7s linear infinite;
}

.category-strip__list-wrapper {
  margin-top: 0.2rem;
}

.category-strip__list {
  display: flex;
  gap: 0.75rem;
  padding: 0.4rem 0.2rem 0.1rem;
  overflow-x: auto;
}

.category-strip__card {
  flex: 0 0 9.5rem;
  display: flex;
  flex-direction: column;
  border-radius: 0.75rem;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(15, 23, 42, 0.95);
  overflow: hidden;
}

.category-strip__image {
  position: relative;
  width: 100%;
  height: 6.5rem;
  background: rgba(15, 23, 42, 0.9);
}

.category-strip__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.category-strip__image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(148, 163, 184, 0.4);
  font-size: 1.4rem;
}

.category-strip__info {
  padding: 0.4rem 0.45rem 0.5rem;
}

.category-strip__name {
  margin: 0 0 0.2rem;
  font-size: 0.78rem;
  font-weight: 500;
  color: #f9fafb;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.category-strip__price {
  margin: 0 0 0.35rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: #40ffaa;
}

.category-strip__actions {
  display: flex;
  gap: 0.25rem;
}

.category-strip__button {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.4rem;
  border-radius: 9999px;
  font-size: 0.72rem;
  font-weight: 500;
  border-width: 1px;
  border-style: solid;
  cursor: pointer;
  text-decoration: none;
}

.category-strip__button--primary {
  background-image: linear-gradient(135deg, rgba(56, 189, 248, 0.95), rgba(59, 130, 246, 0.98));
  border-color: rgba(56, 189, 248, 0.9);
  color: #020617;
}

.category-strip__button--ghost {
  background: rgba(15, 23, 42, 0.9);
  border-color: rgba(148, 163, 184, 0.45);
  color: #e5e7eb;
}

/* Hide scrollbar for WebKit */
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}

@keyframes strip-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .category-strip {
    margin-top: 1rem;
    padding-inline: 0.5rem;
  }

  .category-strip__card {
    flex-basis: 9rem;
  }
}
</style>
