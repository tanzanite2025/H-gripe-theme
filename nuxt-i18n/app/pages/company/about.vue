<template>
  <div class="about-page">
    <header class="support-hero">
      <div class="support-hero__inner">
        <h1 class="support-page__title">About Tanzanite</h1>
        <p class="support-page__intro">
          Project overview and brand story will be published here soon. Replace this copy with company background,
          mission, and team details whenever they are ready.
        </p>
      </div>
    </header>

    <section class="support-placeholder">
      <div class="support-placeholder__card">
        <p>
          This placeholder keeps the <code>/company/about</code> route aligned with the Support layout spacing so future
          content can drop in without extra CSS tweaks.
        </p>
        <p>
          You can safely edit this file to connect dynamic content from your custom plugin at any time.
        </p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta } from '#imports'

definePageMeta({
  layout: 'products',
})
</script>

<style scoped>
.support-hero {
  margin-top: 0;
  padding: 1.5rem 1.5rem 0.75rem;
  background: transparent;
}

.support-hero__inner {
  max-width: 960px;
  margin: 0 auto;
}

.support-page__title {
  margin: 0 0 0.5rem;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #f9fafb;
}

.support-page__intro {
  margin: 0;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-placeholder {
  padding: 2rem 1.5rem 3rem;
}

.support-placeholder__card {
  max-width: 760px;
  margin: 0 auto;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.6);
  padding: 1.75rem;
  font-size: 0.95rem;
  color: rgba(226, 232, 240, 0.9);
  line-height: 1.6;
}

@media (max-width: 768px) {
  .support-hero {
    padding: 1.25rem 1.25rem 0.75rem;
  }

  .support-page__title {
    font-size: 1.5rem;
  }

  .support-placeholder {
    padding-inline: 1.25rem;
  }

  .support-placeholder__card {
    padding: 1.5rem;
  }
}
</style>
