<template>
  <section class="bg-transparent text-white py-8 sm:py-12 lg:py-20">
    <div class="page-content-shell px-0 md:px-6">
      
      <div class="grid lg:grid-cols-12 gap-5 sm:gap-10 lg:gap-16 items-start">
        
        <!-- Left Column: Header & Context -->
        <div class="lg:col-span-3 lg:sticky lg:top-32 self-start space-y-3 sm:space-y-6">
          <div class="inline-flex items-center gap-2 rounded-full border border-[#B5FF6D]/30 bg-[#B5FF6D]/10 px-3 py-1 text-xs font-medium uppercase tracking-wider text-[#B5FF6D]">
            <span>Our Collection</span>
          </div>

          <h2 class="text-2xl sm:text-3xl font-bold text-white leading-tight">
             {{ t('home.featuredProducts.title') }}
          </h2>
          
          <p class="text-base tz-text-secondary leading-relaxed max-w-md">
            {{ t('home.featuredProducts.subtitle') }}
          </p>

          <div class="pt-2 sm:pt-4">
              <NuxtLink :to="localePath('/shop')" class="inline-flex items-center rounded-full border border-white/15 bg-white/5 px-6 py-3 text-sm font-medium text-white transition-[background-color,color,transform] duration-200 hover:-translate-y-px hover:bg-white hover:text-black">
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
            product-category-query-parameter-name="product_category"
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
import { useProductCategories } from '~/composables/useProductCategories'

const { t } = useI18n()
const localePath = useLocalePath()
const {
  tree: productCategoriesState,
  loading: productCategoriesLoading,
  error: productCategoriesError,
  loadCategories,
} = useProductCategories()

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
