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
             <NuxtLink :to="localePath('/shop')" class="premium-button inline-flex items-center px-6 py-3 text-sm font-medium">
                {{ t('home.featuredProducts.viewAll') }}
                <Icon name="lucide:arrow-right" class="ml-2 h-5 w-5" />
              </NuxtLink>
           </div>
         </div>

        <!-- Right Column: Category Grid -->
        <div class="lg:col-span-9">
          <ProductCategoryNavigationCards
            class="home-featured-products__categories"
            density="comfortable"
            :columns="4"
            :product-categories="productCategories"
            :product-categories-loading="productCategoriesLoading"
            :product-categories-error="productCategoriesError"
            :product-category-display-limit="4"
            :show-header="false"
          />
        </div>

      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import ProductCategoryNavigationCards from '~/components/shop/ProductCategoryNavigationCards.vue'
import { useShopCategories } from '~/composables/useShopCategories'

const { t } = useI18n()
const localePath = useLocalePath()
const {
  categories: productCategoriesState,
  loading: productCategoriesLoading,
  error: productCategoriesError,
  loadCategories,
} = useShopCategories()

if (import.meta.server) {
  await loadCategories().catch(() => [])
}

if (import.meta.client) {
  onMounted(async () => {
    if (productCategoriesState.value.length > 0) return

    await loadCategories().catch(() => [])
  })
}

const productCategories = computed(() => {
  return productCategoriesState.value
    .filter((category) => category && category.slug && category.name)
})
</script>
