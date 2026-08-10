<template>
  <main class="min-h-[100dvh] text-white">
    <HomeHero />

    <!-- Unified Features Tabs (Trust & Why Us) -->
    <HomeFeaturesTabs />

    <HomeFeatures />



    <HomeFeaturedProducts />

    <HomeBrandTabs />

    <!-- FAQ Preview Section -->
    <div class="pt-[21px] pb-0">
      <div class="page-content-shell">
        <HomeFaqPreview :max-categories="4" :max-items-per-category="3" wide fluid />
      </div>
    </div>

    <section class="page-content-shell px-0 md:px-4 pt-10 pb-10 md:pt-12 md:pb-12">
      <ContactLocationMap variant="compact" title-tag="h2" layout="split" />
    </section>

    <HomeFinalCta />
  </main>
</template>

<script setup lang="ts">
import HomeHero from '~/components/home/HomeHero.vue'
import HomeFeaturesTabs from '~/components/home/HomeFeaturesTabs.vue'
import HomeFeatures from '~/components/home/HomeFeatures.vue'
import HomeBrandTabs from '~/components/home/HomeBrandTabs.vue'
import HomeFeaturedProducts from '~/components/home/HomeFeaturedProducts.vue'
import HomeFaqPreview from '~/components/HomeFaqPreview.vue'
import ContactLocationMap from '~/components/ContactLocationMap.vue'
import HomeFinalCta from '~/components/home/HomeFinalCta.vue'
</script>
