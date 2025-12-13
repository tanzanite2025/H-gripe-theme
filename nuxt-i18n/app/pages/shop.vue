<template>
  <main class="shop-page max-w-5xl mx-auto pt-0 pb-16 space-y-6">
    <ProductSearchPanel @search="handleSearch" />

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
import { ref, onMounted } from 'vue'
import ProductSearchPanel from '~/components/ProductSearchPanel.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import CategorySidebar from '~/components/CategorySidebar.vue'
import CategoryChips from '~/components/CategoryChips.vue'
import { useWishlist } from '~/composables/useWishlist'
import { useShopCategories } from '~/composables/useShopCategories'
import type { ShopCategory } from '~/composables/useShopCategories'

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

const products = ref<ShopProduct[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// 商品心愿单
const { addToWishlist } = useWishlist()

// 商品分类
const { categories, loading: categoriesLoading, error: categoriesError, loadCategories } = useShopCategories()
const selectedCategory = ref<ShopCategory | null>(null)

interface ProductSearchFiltersPayload {
  priceRange: [number, number]
  attributes?: Record<string, string[]>
  categoryId?: number | null
}

interface ProductSearchPayload {
  query: string
  filters: ProductSearchFiltersPayload
}

const currentSearch = ref<ProductSearchPayload | null>(null)

const buildProductQueryParams = (payload?: ProductSearchPayload) => {
  const params: Record<string, any> = {
    per_page: 24,
    status: 'publish',
  }

  if (payload) {
    const keyword = payload.query?.trim()
    if (keyword) {
      params.keyword = keyword
    }

    const priceRange = payload.filters?.priceRange
    if (Array.isArray(priceRange) && priceRange.length === 2) {
      const [min, max] = priceRange
      params.price_min = min
      params.price_max = max
    }

    const attrs = payload.filters?.attributes || {}
    if (attrs && typeof attrs === 'object') {
      params.attributes = attrs
    }

    const categoryId = payload.filters?.categoryId
    if (typeof categoryId === 'number' && categoryId > 0) {
      params.category = categoryId
    }
  }

  return params
}

const loadProducts = async (payload?: ProductSearchPayload) => {
  loading.value = true
  error.value = null
  try {
    const config = useRuntimeConfig()
    const base = ((config.public as { wpApiBase?: string }).wpApiBase || '/wp-json').replace(/\/$/, '')

    const params = buildProductQueryParams(payload || undefined)

    const response = await $fetch<any>(`${base}/tanzanite/v1/products`, {
      params,
      credentials: 'include',
    })

    if (response && Array.isArray(response.items)) {
      products.value = response.items.map((item: any) => ({
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
    } else {
      products.value = []
    }
  } catch (e: any) {
    console.error('Failed to load shop products:', e)
    error.value = e?.data?.message || 'Failed to load products.'
    products.value = []
  } finally {
    loading.value = false
  }
}

const handleSearch = (payload: ProductSearchPayload) => {
  const next: ProductSearchPayload = {
    ...payload,
    filters: {
      ...payload.filters,
      categoryId: selectedCategory.value?.id ?? null,
    },
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
    filters: {
      ...base.filters,
      categoryId: category?.id ?? null,
    },
  }

  currentSearch.value = next
  loadProducts(next)
}

onMounted(() => {
  loadCategories()
  loadProducts()
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

@media (max-width: 400px) {
  .shop-page {
    padding-inline: 0;
  }
}
</style>
