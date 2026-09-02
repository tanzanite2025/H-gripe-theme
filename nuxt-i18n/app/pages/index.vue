<template>
  <main class="min-h-[100dvh] tz-text-primary">
    <div class="site-header-layout-spacer" aria-hidden="true"></div>
    <HomeHero />

    <LazyHomeRideCategoryStrip :hydrate-on-visible="sectionHydrationOptions" />

    <LazyHomeMainProductCategories :hydrate-on-visible="sectionHydrationOptions" />

    <LazyHomeStorePicksGuide :hydrate-on-visible="sectionHydrationOptions" />

    <LazyHomePurchasePath :hydrate-on-visible="sectionHydrationOptions" />

    <LazyHomeFeaturedProducts :hydrate-on-visible="sectionHydrationOptions" />

    <LazyHomeShopWithConfidence :hydrate-on-visible="sectionHydrationOptions" />

    <LazyHomeFeaturesTabs :hydrate-on-visible="sectionHydrationOptions" />

    <!-- FAQ Preview Section -->
    <div class="pt-[21px] pb-0">
      <div class="page-content-shell">
        <LazyHomeFaqPreview
          :hydrate-on-visible="sectionHydrationOptions"
          :max-categories="4"
          :max-items-per-category="3"
          wide
          fluid
        />
      </div>
    </div>

    <LazyHomeFinalCta :hydrate-on-visible="sectionHydrationOptions" />
  </main>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import HomeHero from '~/components/home/HomeHero.vue'
import { STOREFRONT_NEAR_FOLD_HYDRATION_OPTIONS } from '~/utils/storefrontLoadingPolicy'

const sectionHydrationOptions = STOREFRONT_NEAR_FOLD_HYDRATION_OPTIONS

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
