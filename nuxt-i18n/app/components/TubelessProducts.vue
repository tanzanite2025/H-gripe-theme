<template>
  <section class="tubeless-products">
    <header class="tubeless-products__header">
      <h4 class="tubeless-products__title">Tubeless products</h4>
    </header>

    <div class="tubeless-products__body">
      <!-- Loading state -->
      <div v-if="loading" class="tubeless-products__loading">
        <span class="tubeless-products__spinner" />
        <span class="tubeless-products__loading-text">Loading tubeless products...</span>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="tubeless-products__error">
        {{ error }}
      </div>

      <!-- Empty state -->
      <div v-else-if="products.length === 0" class="tubeless-products__empty">
        <p>No tubeless products are listed yet. Coming soon.</p>
      </div>

      <!-- Product strip -->
      <div v-else class="tubeless-products__list-wrapper">
        <div
          ref="scrollContainer"
          class="tubeless-products__list scrollbar-hide"
          style="scrollbar-width: none; -ms-overflow-style: none;"
        >
          <article
            v-for="product in products"
            :key="product.id"
            class="tubeless-products__card"
          >
            <div class="tubeless-products__image">
              <img
                v-if="product.thumbnail"
                :src="product.thumbnail"
                :alt="product.title"
                loading="lazy"
              />
              <div v-else class="tubeless-products__image-placeholder">
                <span class="tubeless-products__image-icon">📦</span>
              </div>
            </div>

            <div class="tubeless-products__info">
              <h5 class="tubeless-products__name">{{ product.title }}</h5>
              <p v-if="product.priceLabel" class="tubeless-products__price">
                {{ product.priceLabel }}
              </p>

              <div class="tubeless-products__actions">
                <button
                  type="button"
                  class="tubeless-products__button tubeless-products__button--primary"
                  @click="handleAddToCart(product)"
                >
                  Add to cart
                </button>
                <NuxtLink
                  :to="product.url"
                  class="tubeless-products__button tubeless-products__button--ghost"
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
import { ref, onMounted } from 'vue'
import { useRuntimeConfig } from '#imports'
import { useCart } from '~/composables/useCart'

interface TubelessProduct {
  id: number
  title: string
  slug: string
  url: string
  thumbnail?: string
  priceNumber: number
  priceLabel: string
}

const products = ref<TubelessProduct[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const scrollContainer = ref<HTMLElement | null>(null)

const { addToCart, openCart } = useCart()

const loadTubelessProducts = async () => {
  loading.value = true
  error.value = null

  try {
    const config = useRuntimeConfig()
    const base = ((config.public as { wpApiBase?: string }).wpApiBase || '/wp-json').replace(/\/$/, '')

    // Step 1: resolve tubeless category ID by slug
    let tubelessCategoryId: number | null = null
    try {
      const categoryResponse = await $fetch<any>(`${base}/tanzanite/v1/product-categories`, {
        params: {
          per_page: 100,
        },
        credentials: 'include',
      })

      const items = Array.isArray(categoryResponse?.items) ? categoryResponse.items : []
      const tubeless = items.find((term: any) => term?.slug === 'tubeless')
      if (tubeless && typeof tubeless.term_id === 'number') {
        tubelessCategoryId = tubeless.term_id
      }
    } catch (e) {
      // If category lookup fails, we still continue and treat as empty tubeless list
      console.error('Failed to load tubeless category information', e)
    }

    // Step 2: fetch products filtered by tubeless category (if available)
    const productParams: Record<string, any> = {
      per_page: 12,
      status: 'publish',
    }

    if (tubelessCategoryId && tubelessCategoryId > 0) {
      productParams.category = tubelessCategoryId
    }

    const response = await $fetch<any>(`${base}/tanzanite/v1/products`, {
      params: productParams,
      credentials: 'include',
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
        } as TubelessProduct
      })
    } else {
      products.value = []
    }
  } catch (e: any) {
    console.error('Failed to load tubeless products:', e)
    error.value = e?.data?.message || 'Failed to load tubeless products.'
    products.value = []
  } finally {
    loading.value = false
  }
}

const handleAddToCart = (product: TubelessProduct) => {
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
  loadTubelessProducts()
})
</script>

<style scoped>
.tubeless-products {
  margin-top: 1.25rem;
  padding: 0.75rem 0.75rem 0.9rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(148, 163, 184, 0.3);
  background: radial-gradient(circle at top left, rgba(56, 189, 248, 0.12), transparent 45%),
    rgba(15, 23, 42, 0.85);
}

.tubeless-products__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.35rem;
}

.tubeless-products__title {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  color: #e5e7eb;
}

.tubeless-products__body {
  font-size: 0.85rem;
}

.tubeless-products__loading,
.tubeless-products__error,
.tubeless-products__empty {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.2rem 0.1rem;
  color: rgba(148, 163, 184, 0.9);
}

.tubeless-products__spinner {
  width: 0.8rem;
  height: 0.8rem;
  border-radius: 9999px;
  border: 2px solid rgba(148, 163, 184, 0.4);
  border-top-color: rgba(56, 189, 248, 0.9);
  animation: tubeless-spin 0.7s linear infinite;
}

.tubeless-products__list-wrapper {
  margin-top: 0.2rem;
}

.tubeless-products__list {
  display: flex;
  gap: 0.75rem;
  padding: 0.4rem 0.2rem 0.1rem;
  overflow-x: auto;
}

.tubeless-products__card {
  flex: 0 0 9.5rem;
  display: flex;
  flex-direction: column;
  border-radius: 0.75rem;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(15, 23, 42, 0.95);
  overflow: hidden;
}

.tubeless-products__image {
  position: relative;
  width: 100%;
  height: 6.5rem;
  background: rgba(15, 23, 42, 0.9);
}

.tubeless-products__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.tubeless-products__image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(148, 163, 184, 0.4);
  font-size: 1.4rem;
}

.tubeless-products__info {
  padding: 0.4rem 0.45rem 0.5rem;
}

.tubeless-products__name {
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

.tubeless-products__price {
  margin: 0 0 0.35rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: #40ffaa;
}

.tubeless-products__actions {
  display: flex;
  gap: 0.25rem;
}

.tubeless-products__button {
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

.tubeless-products__button--primary {
  background-image: linear-gradient(135deg, rgba(56, 189, 248, 0.95), rgba(59, 130, 246, 0.98));
  border-color: rgba(56, 189, 248, 0.9);
  color: #020617;
}

.tubeless-products__button--ghost {
  background: rgba(15, 23, 42, 0.9);
  border-color: rgba(148, 163, 184, 0.45);
  color: #e5e7eb;
}

/* Hide scrollbar for WebKit */
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}

@keyframes tubeless-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .tubeless-products {
    margin-top: 1rem;
    padding-inline: 0.5rem;
  }

  .tubeless-products__card {
    flex-basis: 9rem;
  }
}
</style>
