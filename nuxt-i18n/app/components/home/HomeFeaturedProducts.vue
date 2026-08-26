<template>
  <section id="featured-products" class="bg-transparent py-8 tz-text-primary sm:py-12 lg:py-20">
    <div class="page-content-shell px-0 md:px-6">
      <div class="grid items-start gap-5 sm:gap-10 lg:grid-cols-12 lg:gap-16">
        <div class="self-start space-y-3 sm:space-y-6 lg:sticky lg:top-32 lg:col-span-3">
          <div
            class="inline-flex items-center gap-2 rounded-full border border-[#059669]/30 bg-[#059669]/10 px-3 py-1 text-[11px] font-medium uppercase tracking-[0.16em] text-[#059669]"
          >
            <Icon name="lucide:star" class="h-3.5 w-3.5" aria-hidden="true" />
            {{ t('home.featuredProducts.eyebrow') }}
          </div>

          <div>
            <h2 class="text-2xl font-bold leading-tight tz-text-primary sm:text-3xl">
              {{ t('home.featuredProducts.heading') }}
            </h2>
          </div>

          <NuxtLink
            :to="wheelsetShopPath"
            class="premium-button inline-flex w-full items-center justify-center sm:w-auto"
          >
            <Icon name="lucide:arrow-right" class="mr-2 h-4 w-4" aria-hidden="true" />
            {{ t('home.featuredProducts.shopAll') }}
          </NuxtLink>
        </div>

        <div class="min-w-0 lg:col-span-9">
          <HomeFeaturedWheelsets
            :products="featuredProducts"
            :pending="featuredProductsPending"
            :error="featuredProductsError || null"
            :has-more="featuredProductsHasMore"
            :total="featuredProductsTotal"
            :shop-path="wheelsetShopPath"
            @retry="retryFeaturedProducts"
          />

          <HomeWheelsetStandards class="mt-6 lg:mt-8" />
        </div>
      </div>

      <HomeBrandPhotoCarousel class="mt-8 lg:mt-12" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAsyncData, useI18n, useLocalePath } from '#imports'
import HomeBrandPhotoCarousel from '~/components/home/HomeBrandPhotoCarousel.vue'
import HomeFeaturedWheelsets from '~/components/home/HomeFeaturedWheelsets.vue'
import HomeWheelsetStandards from '~/components/home/HomeWheelsetStandards.vue'
import { useShopProducts, type ShopProductsResult } from '~/composables/useShopProducts'

const { t } = useI18n()
const localePath = useLocalePath()
const { fetchFeaturedShopProducts } = useShopProducts()
const WHEELSET_PRODUCT_CATEGORY_SLUG = 'wheelset'
const HOME_FEATURED_WHEELSETS_PAGE_SIZE = 12

const wheelsetShopPath = localePath({
  path: '/shop',
  query: { product_category: WHEELSET_PRODUCT_CATEGORY_SLUG },
})

const {
  data: featuredProductsData,
  pending: featuredProductsPending,
  error: featuredProductsError,
  refresh: refreshFeaturedProducts,
} = await useAsyncData<ShopProductsResult>(
  'home-featured-wheelsets',
  async () => {
    return fetchFeaturedShopProducts({
      page_size: HOME_FEATURED_WHEELSETS_PAGE_SIZE,
      product_category: WHEELSET_PRODUCT_CATEGORY_SLUG,
    })
  },
  {
    default: () => ({
      items: [],
      raw: null,
      page: 1,
      pageSize: HOME_FEATURED_WHEELSETS_PAGE_SIZE,
      total: 0,
      hasMore: false,
    }),
  },
)

const featuredProducts = computed(() => featuredProductsData.value?.items ?? [])
const featuredProductsTotal = computed(() => featuredProductsData.value?.total ?? featuredProducts.value.length)
const featuredProductsHasMore = computed(() => featuredProductsData.value?.hasMore === true)

const retryFeaturedProducts = () => {
  void refreshFeaturedProducts()
}
</script>

<style scoped>
#featured-products {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}
</style>
