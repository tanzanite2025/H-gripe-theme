<template>
  <section class="bg-transparent text-white py-8 sm:py-12 lg:py-20">
    <div class="page-content-shell px-0 md:px-6">
      
      <div class="grid lg:grid-cols-12 gap-5 sm:gap-10 lg:gap-16 items-start">
        
        <!-- Left Column: Header & Context -->
        <div class="lg:col-span-3 lg:sticky lg:top-32 self-start space-y-3 sm:space-y-6">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-300 text-xs font-medium uppercase tracking-wider">
            <span>Our Collection</span>
          </div>

          <h2 class="text-2xl sm:text-3xl font-bold text-white leading-tight">
             {{ t('home.featuredProducts.title') }}
          </h2>
          
          <p class="text-base tz-text-secondary leading-relaxed max-w-md">
            {{ t('home.featuredProducts.subtitle') }}
          </p>

          <div class="pt-2 sm:pt-4">
             <NuxtLink to="/shop" class="premium-button inline-flex items-center px-6 py-3 text-sm font-medium">
                {{ t('home.featuredProducts.viewAll') }}
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" class="ml-2 h-5 w-5"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3" /></svg>
             </NuxtLink>
          </div>
        </div>

        <!-- Right Column: Product Grid -->
        <div class="lg:col-span-9">
          <div v-if="cards.length > 0" class="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
            <NuxtLink
              v-for="card in cards"
              :key="card.key"
              :to="card.url"
              class="group block overflow-hidden rounded-2xl premium-card relative hover:shadow-2xl hover:shadow-black/40 transition-all duration-500"
            >
              <!-- Image Aspect -->
              <div
                class="relative bg-[var(--tz-card-surface)] overflow-hidden"
                :class="card.kind === 'category' ? 'aspect-square' : 'aspect-[4/3]'"
              >
                 <img
                   v-if="card.thumbnail && !brokenCardImageKeys.includes(card.key)"
                   :src="card.thumbnail"
                   :alt="card.title"
                   class="absolute inset-0 h-full w-full object-cover transition-transform duration-700 ease-out group-hover:scale-110"
                   loading="lazy"
                   @error="handleCardImageError(card.key)"
                 />
                 <!-- Placeholder Gradient / Image Slot -->
                  <div
                    v-else
                    class="absolute inset-0 bg-[var(--tz-card-surface)] group-hover:scale-110 transition-transform duration-700 ease-out"
                 ></div>
                 
                 <!-- Gradient Overlay -->
                 <div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/40 to-transparent"></div>
                 
                 <div class="absolute bottom-0 inset-x-0 p-5">
                    <span
                      v-if="card.category"
                      class="inline-flex items-center mb-2 px-2.5 py-1 rounded-md bg-black/35 border border-white/15 text-[11px] font-medium uppercase tracking-wide text-white/80"
                    >
                      {{ card.category }}
                    </span>
                    <h3 class="text-lg font-bold text-white mb-1 group-hover:text-[#B5FF6D] transition-colors">{{ card.title }}</h3>
                    <p class="tz-text-secondary text-sm line-clamp-2 mb-3">{{ card.description }}</p>
                    <div
                      v-if="card.price"
                      class="inline-block px-3 py-1 rounded-lg bg-white/10 backdrop-blur text-xs font-medium text-white/90 border border-white/10 group-hover:bg-white/15 group-hover:border-white/20 transition-colors"
                    >
                       {{ card.price }}
                    </div>
                 </div>
              </div>
            </NuxtLink>
          </div>
          <p v-else class="rounded-xl border border-white/10 bg-white/5 p-6 text-sm tz-text-secondary">
            {{ t('home.featuredProducts.subtitle') }}
          </p>
        </div>

      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAsyncData, useI18n, useLocalePath, useState } from '#imports'
import { useShopProducts } from '~/composables/useShopProducts'
import type { ShopProduct } from '~/composables/useShopProducts'
import { useShopCategories } from '~/composables/useShopCategories'
import type { ShopCategory } from '~/composables/useShopCategories'

const { t } = useI18n()
const localePath = useLocalePath()
const { fetchFeaturedShopProducts } = useShopProducts()
const { fetchCategories } = useShopCategories()
const productCategoriesState = useState<ShopCategory[]>('home-product-categories', () => [])
const brokenCardImageKeys = ref<string[]>([])

interface FeaturedProductCard {
  key: string
  title: string
  description: string
  price: string
  url: string
  thumbnail?: string
  category?: string
  kind: 'product' | 'category'
}

const { data: featuredProductsData } = await useAsyncData(
  'home-featured-products',
  () => fetchFeaturedShopProducts({
    page_size: 4,
    status: 'active',
  }),
  {
    default: () => ({
      items: [],
      raw: null,
    }),
  }
)

if (import.meta.server) {
  productCategoriesState.value = await fetchCategories().catch(() => [])
}

if (import.meta.client) {
  onMounted(async () => {
    if (productCategoriesState.value.length > 0) return

    const categories = await fetchCategories().catch(() => [])
    if (categories.length > 0) {
      productCategoriesState.value = categories
    }
  })
}

const featuredProducts = computed<ShopProduct[]>(() => {
  const items = featuredProductsData.value?.items
  return Array.isArray(items) ? items : []
})

const productCategories = computed(() => {
  return Array.isArray(productCategoriesState.value) ? productCategoriesState.value : []
})

const dynamicCards = computed<FeaturedProductCard[]>(() =>
  featuredProducts.value
    .filter((product) => Boolean(product.productType?.slug && product.productType.name))
    .slice(0, 4)
    .map((product) => ({
      key: `product-${product.id}`,
      title: product.title,
      description: product.description || t('home.featuredProducts.subtitle'),
      price: product.priceLabel,
      url: product.url,
      thumbnail: product.thumbnail,
      category: product.productType?.name,
      kind: 'product',
    }))
)

const categoryCards = computed<FeaturedProductCard[]>(() => {
  return productCategories.value.slice(0, 4).map((category) => ({
    key: `category-${category.id}`,
    title: category.name,
    description: t('home.featuredProducts.subtitle'),
    price: '',
    url: localePath({
      path: '/shop',
      query: {
        product_type: category.slug,
      },
    }),
    thumbnail: category.image,
    category: category.name,
    kind: 'category',
  }))
})

const cards = computed(() => {
  return categoryCards.value.length > 0 ? categoryCards.value : dynamicCards.value
})

const handleCardImageError = (key: string): void => {
  if (brokenCardImageKeys.value.includes(key)) return
  brokenCardImageKeys.value = [...brokenCardImageKeys.value, key]
}
</script>
