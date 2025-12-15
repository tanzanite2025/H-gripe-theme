<template>
  <main class="min-h-screen text-white">
    <HomeHero />

    <!-- Trust Cards Section -->
    <section class="max-w-5xl mx-auto px-1 sm:px-4 pt-[3px] pb-0">
      <h2 class="text-lg lg:text-xl font-semibold text-white/90 text-center">{{ t('home.trust.title') }}</h2>
      <div class="mt-[6px] mb-[9px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
      <TrustCards layout="row" size="lg" :show-title="false" />
    </section>

    <HomeFeatures />

    <HomeFactoryStories />

    <HomeFeaturedProducts />

    <HomeTestimonials />

    <!-- FAQ Preview Section -->
    <div class="pt-[21px] pb-0">
      <HomeFaqPreview :max-items-per-category="4" wide />
    </div>

    <section class="max-w-6xl mx-auto px-1 sm:px-4 pt-[21px] pb-0">
      <ContactLocationMap variant="compact" title-tag="h2" layout="split" />
    </section>

    <HomeFinalCta />
  </main>
</template>

<script setup>
import HomeHero from '~/components/home/HomeHero.vue'
import HomeFeatures from '~/components/home/HomeFeatures.vue'
import HomeFactoryStories from '~/components/home/HomeFactoryStories.vue'
import HomeFeaturedProducts from '~/components/home/HomeFeaturedProducts.vue'
import HomeTestimonials from '~/components/home/HomeTestimonials.vue'
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
