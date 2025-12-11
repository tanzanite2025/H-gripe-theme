<template>
  <div class="product-feedback-page">
    <header class="support-hero">
      <div class="support-hero__inner">
        <p class="support-hero__eyebrow">{{ t('feedbackForm.pageEyebrow') }}</p>
        <h1 class="support-page__title support-page__title--sr-only">
          {{ t('feedbackForm.pageTitle') }}
        </h1>
        <p class="support-page__intro">
          {{ t('feedbackForm.pageIntro') }}
        </p>
      </div>
    </header>

    <section class="feedback-section">
      <SuggestionForm />
    </section>

    <section class="support-section">
      <PageFaq 
        page-id="support-product-feedback"
        theme="dark"
        :show-categories="true"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SuggestionForm from '~/components/feedback/SuggestionForm.vue'
import PageFaq from '~/components/PageFaq.vue'

const { t } = useI18n()

definePageMeta({
  layout: 'support',
})

useHead({
  title: t('feedbackForm.pageTitle'),
})
</script>

<style scoped>
.product-feedback-page {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.support-hero {
  margin-top: 0;
  padding: 0 0 1rem;
}

.support-hero__inner {
  max-width: 960px;
}

.support-hero__eyebrow {
  font-size: 0.85rem;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: rgba(129, 140, 248, 0.9);
  margin-bottom: 0.65rem;
}

.support-page__title {
  margin: 0 0 0.75rem;
  font-size: clamp(1.75rem, 4vw, 2.5rem);
  font-weight: 700;
  color: #f9fafb;
}

.support-page__title--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

.support-page__intro {
  margin: 0;
  font-size: 1rem;
  color: rgba(148, 163, 184, 0.9);
  max-width: 720px;
}

.feedback-section {
  width: 100%;
  max-width: 960px;
}

@media (max-width: 768px) {
  .feedback-section {
    max-width: none;
  }

  :deep(.feedback-card) {
    width: 100%;
    border-radius: 20px;
    padding: 1.25rem;
  }
}
</style>
