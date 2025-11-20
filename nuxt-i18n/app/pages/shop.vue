<template>
  <main class="max-w-5xl mx-auto px-6 py-16 space-y-6">
    <header class="space-y-2 text-center">
      <h1 class="text-4xl font-semibold">Tanzanite Shop</h1>
      <p class="text-base text-white/70">
        Coming soon: curated product collections powered by the Tanzanite plugin.
      </p>
    </header>

    <section class="rounded-xl border border-white/10 bg-white/5 p-6 text-sm text-white/80">
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
          class="group border border-white/10 rounded-xl bg-black/40 hover:bg-black/60 transition-colors overflow-hidden flex flex-col"
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
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useWishlist } from '~/composables/useWishlist'

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

const { addToWishlist } = useWishlist()

const loadProducts = async () => {
  loading.value = true
  error.value = null
  try {
    const config = useRuntimeConfig()
    const base = ((config.public as { wpApiBase?: string }).wpApiBase || '/wp-json').replace(/\/$/, '')

    const response = await $fetch<any>(`${base}/tanzanite/v1/products`, {
      params: {
        per_page: 24,
        status: 'publish',
      },
      credentials: 'include',
    })

    if (response && Array.isArray(response.items)) {
      products.value = response.items.map((item: any) => ({
        id: item.id,
        title: item.title,
        url: item.preview_url || `/product/${item.slug || item.id}`,
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

onMounted(loadProducts)

const handleAddToWishlist = async (product: ShopProduct) => {
  if (!product?.id) return
  try {
    await addToWishlist(product.id)
  } catch (e) {
    console.error('Failed to add to wishlist from shop:', e)
  }
}
</script>
