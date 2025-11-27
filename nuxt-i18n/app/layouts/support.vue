<template>
  <div class="layout layout--support">
    <main class="layout-main">
      <!-- Support hero / banner -->
      <section class="support-hero">
        <div class="support-hero__inner">
          <h1 class="support-hero__title">Support</h1>
        </div>
      </section>

      <!-- Top horizontal Support navigation -->
      <SupportTopNav />

      <!-- Support page content -->
      <section class="support-content">
        <div class="support-content__inner">
          <slot />
        </div>
      </section>
    </main>

    <AppFooter />
    <GradientDockMenu />
  </div>
</template>

<script setup lang="ts">
import AppFooter from '~/components/AppFooter.vue'
import GradientDockMenu from '~/components/GradientDockMenu.vue'
import SupportTopNav from '~/components/SupportTopNav.vue'
</script>

<style scoped>
.layout--support {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: radial-gradient(circle at top, #020617 0, #020617 24%, #000000 100%);
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.support-hero {
  margin-top: 2.75rem;
  padding: 3rem 1.5rem 1.25rem;
  background: radial-gradient(circle at top left, #0f172a 0, #020617 40%, #020617 100%);
}

.support-hero__inner {
  max-width: 960px;
  margin: 0 auto;
}

.support-hero__title {
  margin: 0 0 0.5rem;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #f9fafb;
}

.support-hero__subtitle {
  margin: 0;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-content {
  padding: 2rem 1.5rem 3rem;
}

.support-content__inner {
  max-width: 960px;
  margin: 0 auto;
}

@media (max-width: 768px) {
  .support-hero {
    margin-top: 3.25rem;
    padding: 2rem 1.25rem 0.75rem;
  }

  .support-hero__title {
    font-size: 1.5rem;
  }

  .support-content {
    padding-inline: 1.25rem;
  }
}
</style>
