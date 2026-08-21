<template>
  <main class="min-h-[100dvh] text-white">
    <div class="site-header-layout-spacer" aria-hidden="true"></div>
    <HomeHero />

    <HomeRideCategoryStrip />

    <HomeMainProductCategories />

    <HomeStorePicksGuide />

    <LazyHomePurchasePath />

    <Suspense>
      <AsyncHomeFeaturedProducts />
      <template #fallback>
        <HomeFeaturedProductsSkeleton />
      </template>
    </Suspense>

    <HomeShopWithConfidence />

    <HomeFeaturesTabs />

    <!-- FAQ Preview Section -->
    <div class="pt-[21px] pb-0">
      <div class="page-content-shell">
        <HomeFaqPreview :max-categories="4" :max-items-per-category="3" wide fluid />
      </div>
    </div>

    <HomeFinalCta />
  </main>
</template>

<script setup lang="ts">
import { defineAsyncComponent, onMounted } from 'vue'
import HomeHero from '~/components/home/HomeHero.vue'
import HomeRideCategoryStrip from '~/components/home/HomeRideCategoryStrip.vue'
import HomeMainProductCategories from '~/components/home/HomeMainProductCategories.vue'
import HomeStorePicksGuide from '~/components/home/HomeStorePicksGuide.vue'
import HomeShopWithConfidence from '~/components/home/HomeShopWithConfidence.vue'
import HomeFeaturesTabs from '~/components/home/HomeFeaturesTabs.vue'
import HomeFeaturedProductsSkeleton from '~/components/home/HomeFeaturedProductsSkeleton.vue'
import HomeFaqPreview from '~/components/HomeFaqPreview.vue'
import HomeFinalCta from '~/components/home/HomeFinalCta.vue'

const AsyncHomeFeaturedProducts = defineAsyncComponent(
  () => import('~/components/home/HomeFeaturedProducts.vue'),
)

const homepageLegacySectionHashes = new Set(['#home-buying-path', '#featured-products'])

onMounted(() => {
  if (!homepageLegacySectionHashes.has(window.location.hash)) return

  window.history.replaceState(
    window.history.state,
    '',
    `${window.location.pathname}${window.location.search}`,
  )

  requestAnimationFrame(() => {
    window.scrollTo({ top: 0, left: 0, behavior: 'auto' })
  })
})
</script>
