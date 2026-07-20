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
        <HomeFaqPreview :max-items-per-category="4" wide fluid />
      </div>
    </div>

    <section class="page-content-shell px-0 md:px-4 pt-[21px] pb-0">
      <ContactLocationMap variant="compact" title-tag="h2" layout="split" />
    </section>

    <HomeFinalCta />
  </main>
</template>

<script setup>
import HomeHero from '~/components/home/HomeHero.vue'
import HomeFeaturesTabs from '~/components/home/HomeFeaturesTabs.vue'
import HomeFeatures from '~/components/home/HomeFeatures.vue'
import HomeBrandTabs from '~/components/home/HomeBrandTabs.vue'
import HomeFeaturedProducts from '~/components/home/HomeFeaturedProducts.vue'
import HomeFaqPreview from '~/components/HomeFaqPreview.vue'
import ContactLocationMap from '~/components/ContactLocationMap.vue'
import HomeFinalCta from '~/components/home/HomeFinalCta.vue'
const { t } = useI18n()

// SEO Meta Tags
useHead({
  title: t('home.meta.title'),
  meta: [
    { name: 'description', content: t('home.meta.description') },
    { property: 'og:title', content: t('home.meta.title') },
    { property: 'og:description', content: t('home.meta.description') }
  ]
})
</script>
