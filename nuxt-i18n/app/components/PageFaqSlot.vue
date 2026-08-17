<template>
  <section
    v-if="shouldLoadFaqSlot && pending"
    class="page-faq-slot"
    aria-busy="true"
    aria-live="polite"
  >
    <div class="page-faq-slot__skeleton">
      <header class="page-faq-slot__skeleton-header">
        <span class="page-faq-slot__loader-dot" aria-hidden="true" />
        <div class="min-w-0">
          <p class="page-faq-slot__eyebrow">FAQ</p>
          <p class="page-faq-slot__loader-text">{{ t('faq.ui.loadingAnswers') }}</p>
        </div>
      </header>

      <div class="page-faq-slot__skeleton-list" aria-hidden="true">
        <div
          v-for="index in 3"
          :key="index"
          class="page-faq-slot__skeleton-item"
        >
          <span class="page-faq-slot__skeleton-line page-faq-slot__skeleton-line--question" />
          <span class="page-faq-slot__skeleton-line page-faq-slot__skeleton-line--answer" />
        </div>
      </div>
    </div>
  </section>

  <PageFaq
    v-else-if="resolvedFaqData && resolvedFaqPageId"
    :page-id="resolvedFaqPageId"
    :data="resolvedFaqData"
    :show-categories="true"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAsyncData, useRoute } from '#imports'
import PageFaq from '~/components/PageFaq.vue'
import { fetchFaqDataByRoutePath, getPageFaqId, resolveFaqRouteLookupPath, shouldAutoInsertFaqForRoute } from '~/data/faq'

const route = useRoute()
const { locale, t } = useI18n()

const shouldLoadFaqSlot = computed(() => shouldAutoInsertFaqForRoute(route.path))
const faqLookupRoutePath = computed(() => resolveFaqRouteLookupPath(route.path))

const { data: faqData, pending } = await useAsyncData(
  () => `faq-slot-${locale.value}-${shouldLoadFaqSlot.value ? faqLookupRoutePath.value : 'disabled'}`,
  () => shouldLoadFaqSlot.value ? fetchFaqDataByRoutePath(faqLookupRoutePath.value) : Promise.resolve(null),
  { watch: [locale, shouldLoadFaqSlot, faqLookupRoutePath] }
)

const resolvedFaqData = computed(() => faqData.value)
const resolvedFaqPageId = computed(() => resolvedFaqData.value ? getPageFaqId(resolvedFaqData.value) : '')
</script>

<style scoped>
.page-faq-slot {
  width: 100%;
  padding: clamp(0.75rem, 1vw, 1.2rem) 0;
}

.page-faq-slot__skeleton {
  position: relative;
  overflow: hidden;
  min-height: 9.75rem;
  padding: clamp(0.9rem, 1.4vw, 1.25rem);
  border-radius: 1.25rem;
  border: 1px solid rgba(148, 163, 184, 0.14);
  background:
    radial-gradient(circle at top left, rgba(181, 255, 109, 0.08), transparent 36%),
    rgba(15, 23, 42, 0.62);
  color: var(--tz-text-secondary);
}

.page-faq-slot__skeleton::after {
  position: absolute;
  inset: 0;
  content: '';
  pointer-events: none;
  background: linear-gradient(100deg, transparent 18%, rgba(255, 255, 255, 0.055) 36%, transparent 54%);
  transform: translateX(-100%);
  animation: page-faq-slot-shimmer 1.65s ease-in-out infinite;
}

.page-faq-slot__skeleton-header {
  display: flex;
  align-items: center;
  gap: 0.7rem;
}

.page-faq-slot__loader-dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 999px;
  background: var(--tz-brand-primary, #b5ff6d);
  box-shadow: 0 0 18px rgba(181, 255, 109, 0.65);
  animation: page-faq-slot-pulse 0.95s ease-in-out infinite alternate;
}

.page-faq-slot__eyebrow {
  margin: 0;
  font-size: var(--tz-type-micro-label);
  font-weight: 900;
  letter-spacing: 0.22em;
  color: var(--tz-text-muted);
}

.page-faq-slot__loader-text {
  margin: 0.12rem 0 0;
  font-size: var(--tz-type-caption);
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--tz-text-secondary);
}

.page-faq-slot__skeleton-list {
  display: grid;
  gap: 0.7rem;
  margin-top: 1rem;
}

.page-faq-slot__skeleton-item {
  display: grid;
  gap: 0.45rem;
  padding: 0.8rem 0.9rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(2, 6, 23, 0.22);
}

.page-faq-slot__skeleton-line {
  display: block;
  height: 0.55rem;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.18);
}

.page-faq-slot__skeleton-line--question {
  width: min(72%, 30rem);
}

.page-faq-slot__skeleton-line--answer {
  width: min(48%, 18rem);
  opacity: 0.7;
}

@keyframes page-faq-slot-pulse {
  from {
    opacity: 0.35;
    transform: scale(0.86);
  }

  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes page-faq-slot-shimmer {
  0% {
    transform: translateX(-100%);
  }

  58%,
  100% {
    transform: translateX(100%);
  }
}

@media (max-width: 640px) {
  .page-faq-slot__skeleton {
    min-height: 8.5rem;
    border-radius: 1rem;
  }

  .page-faq-slot__skeleton-item {
    padding: 0.72rem 0.75rem;
  }
}
</style>
