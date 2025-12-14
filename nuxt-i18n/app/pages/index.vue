<template>
  <main class="min-h-screen bg-neutral-950 text-white">
    <HomeHero />

    <!-- Trust Cards Section -->
    <section class="max-w-5xl mx-auto px-4 py-4 lg:py-12">
      <h2 class="text-lg lg:text-xl font-semibold text-white/90 text-center mb-4 lg:mb-6">{{ t('home.trust.title') }}</h2>
      <TrustCards layout="row" size="lg" :show-title="false" />
    </section>

    <!-- FAQ Preview Section -->
    <div class="mt-2 lg:mt-0">
      <HomeFaqPreview :max-items-per-category="4" />
    </div>

    <section class="max-w-5xl mx-auto px-4 pb-10 lg:pb-16">
      <ContactLocationMap variant="compact" title-tag="h2" />
    </section>
  </main>
</template>

<script setup>
import HomeHero from '~/components/home/HomeHero.vue'
import HomeFaqPreview from '~/components/HomeFaqPreview.vue'
import ContactLocationMap from '~/components/ContactLocationMap.vue'
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
