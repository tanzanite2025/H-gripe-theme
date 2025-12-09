<template>
  <main class="i18n-demo-page">
    <!-- Tailwind Carousel Preview (demo only) -->
    <div class="carousel-wrapper">
      <TWCarousel />
    </div>

    <!-- Trust Cards Section -->
    <section class="max-w-5xl mx-auto px-4 py-4 lg:py-12">
      <h2 class="text-lg lg:text-xl font-semibold text-white/90 text-center mb-4 lg:mb-6">Why Choose Us</h2>
      <TrustCards layout="row" size="lg" :show-title="false" />
    </section>

    <!-- FAQ Preview Section -->
    <div class="mt-2 lg:mt-0">
      <HomeFaqPreview :max-items-per-category="4" />
    </div>
  </main>
</template>

<script setup>
import TWCarousel from '~/components/TWCarousel.vue'
import HomeFaqPreview from '~/components/HomeFaqPreview.vue'
const { t } = useI18n()

// SEO Meta Tags
useHead({
  title: t('welcome'),
  meta: [
    { name: 'description', content: t('welcome') },
    { property: 'og:title', content: t('welcome') },
    { property: 'og:description', content: t('welcome') }
  ]
})
</script>

<style scoped>
.i18n-demo-page {
  min-height: 100vh;
  background: #020617;
  color: #fff;
  padding: 20px;
}

/* 轮播突破父级 padding，移动端满屏 */
.carousel-wrapper {
  margin-left: -20px;
  margin-right: -20px;
  width: calc(100% + 40px);
}

@media (min-width: 1024px) {
  .carousel-wrapper {
    margin-left: 0;
    margin-right: 0;
    width: 100%;
  }
}

.demo-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.demo-header h1 {
  font-size: 28px;
  margin: 0;
  background: linear-gradient(90deg, #2aa3ff, #40ffaa);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.demo-content {
  max-width: 800px;
  margin: 0 auto;
}

.breadcrumbs {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 20px;
}

.search-section {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.search-input {
  flex: 1;
  padding: 12px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
}

.search-button {
  padding: 12px 24px;
  background: rgba(42, 163, 255, 0.18);
  border: 1px solid rgba(42, 163, 255, 0.55);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.search-button:hover {
  background: rgba(42, 163, 255, 0.28);
}

.filters-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-item label {
  font-size: 13px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
}

.filter-item select {
  padding: 10px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
}

.user-actions {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 20px;
}

.action-btn {
  padding: 10px 20px;
  background: rgba(42, 163, 255, 0.18);
  border: 1px solid rgba(42, 163, 255, 0.55);
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: rgba(42, 163, 255, 0.28);
}

.points {
  padding: 10px 20px;
  background: rgba(42, 163, 255, 0.18);
  border: 1px solid rgba(42, 163, 255, 0.55);
  border-radius: 6px;
  font-size: 14px;
}

.cta-section {
  padding: 20px;
  background: rgba(42, 163, 255, 0.1);
  border: 1px solid rgba(42, 163, 255, 0.3);
  border-radius: 8px;
  text-align: center;
}

.cta-section p {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  background: linear-gradient(90deg, #ffffff 0%, #ffe8a3 45%, #ffd24d 75%, #ff9f1a 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
