<template>
  <main class="min-h-[100dvh] tz-text-primary">
    <div class="site-header-layout-spacer" aria-hidden="true"></div>
    <HomeHero />

    <HomeDeferredSection
      min-height="8rem"
      module-id="components/home/HomeRideCategoryStrip.vue"
      :loader="loadHomeRideCategoryStrip"
      :root-margin="sectionMountRootMargin"
    />

    <HomeDeferredSection
      min-height="38rem"
      module-id="components/home/HomeMainProductCategories.vue"
      :loader="loadHomeMainProductCategories"
      :root-margin="sectionMountRootMargin"
    />

    <HomeDeferredSection
      min-height="14rem"
      module-id="components/home/HomeStorePicksGuide.vue"
      :loader="loadHomeStorePicksGuide"
      :root-margin="sectionMountRootMargin"
    />

    <HomeDeferredSection
      min-height="32rem"
      module-id="components/home/HomePurchasePath.vue"
      :loader="loadHomePurchasePath"
      :root-margin="sectionMountRootMargin"
    />

    <HomeDeferredSection
      min-height="54rem"
      module-id="components/home/HomeFeaturedProducts.vue"
      :loader="loadHomeFeaturedProducts"
      :root-margin="sectionMountRootMargin"
    />

    <HomeDeferredSection
      min-height="36rem"
      module-id="components/home/HomeShopWithConfidence.vue"
      :loader="loadHomeShopWithConfidence"
      :root-margin="sectionMountRootMargin"
    />

    <HomeDeferredSection
      min-height="30rem"
      module-id="components/home/HomeFeaturesTabs.vue"
      :loader="loadHomeFeaturesTabs"
      :root-margin="sectionMountRootMargin"
    />

    <!-- FAQ Preview Section -->
    <div class="pt-[21px] pb-0">
      <div class="page-content-shell">
        <HomeDeferredSection
          min-height="36rem"
          module-id="components/HomeFaqPreview.vue"
          :loader="loadHomeFaqPreview"
          :root-margin="sectionMountRootMargin"
          :max-categories="4"
          :max-items-per-category="3"
          wide
          fluid
        />
      </div>
    </div>

    <HomeDeferredSection
      min-height="24rem"
      module-id="components/home/HomeFinalCta.vue"
      :loader="loadHomeFinalCta"
      :root-margin="sectionMountRootMargin"
    />
  </main>
</template>

<script setup lang="ts">
import { onMounted, type Component } from 'vue'
import HomeDeferredSection from '~/components/home/HomeDeferredSection.vue'
import HomeHero from '~/components/home/HomeHero.vue'
import { STOREFRONT_NEAR_FOLD_HYDRATION_OPTIONS } from '~/utils/storefrontLoadingPolicy'

type DeferredHomeSectionLoader = () => Promise<{ default: Component }>

const sectionMountRootMargin = STOREFRONT_NEAR_FOLD_HYDRATION_OPTIONS.rootMargin
const loadHomeRideCategoryStrip: DeferredHomeSectionLoader = () => import('~/components/home/HomeRideCategoryStrip.vue')
const loadHomeMainProductCategories: DeferredHomeSectionLoader = () => import('~/components/home/HomeMainProductCategories.vue')
const loadHomeStorePicksGuide: DeferredHomeSectionLoader = () => import('~/components/home/HomeStorePicksGuide.vue')
const loadHomePurchasePath: DeferredHomeSectionLoader = () => import('~/components/home/HomePurchasePath.vue')
const loadHomeFeaturedProducts: DeferredHomeSectionLoader = () => import('~/components/home/HomeFeaturedProducts.vue')
const loadHomeShopWithConfidence: DeferredHomeSectionLoader = () => import('~/components/home/HomeShopWithConfidence.vue')
const loadHomeFeaturesTabs: DeferredHomeSectionLoader = () => import('~/components/home/HomeFeaturesTabs.vue')
const loadHomeFaqPreview: DeferredHomeSectionLoader = () => import('~/components/HomeFaqPreview.vue')
const loadHomeFinalCta: DeferredHomeSectionLoader = () => import('~/components/home/HomeFinalCta.vue')

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
