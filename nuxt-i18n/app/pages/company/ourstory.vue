<template>
  <div class="company-page">

    <h1 class="company-page__title company-page__title--sr-only">Our Story</h1>

    <section
      id="story"
      class="company-section"
    >
      <h2 class="company-section__title">{{ t('company.ourStory.story.title') }}</h2>
      <div class="pt-1 pb-3 text-center">
        <NuxtLink
          class="company-tabs__item inline-flex items-center"
          :to="factoryTabTo"
        >
          {{ t('company.ourStory.story.factoryButton') }}
        </NuxtLink>
      </div>
      <div class="pb-6">
        <div class="aspect-[2/1] w-full overflow-hidden rounded-2xl bg-slate-950">
          <img
            class="h-full w-full object-cover"
            src="/company/ourstory/ourstory/tanzanite-ourstory.webp"
            alt="Tanzanite Our Story"
            loading="lazy"
          />
        </div>
      </div>
      <ul class="flex flex-col gap-3">
        <li
          v-for="(paragraph, index) in storyParagraphs"
          :key="index"
        >
          <p class="company-section__body m-0">
            {{ paragraph }}
          </p>
        </li>
      </ul>
    </section>

    <div class="company-feedback">
      <UserFeedbackThread
        threadKey="company-ourstory"
        title="Share your feedback about Our Story and the factory"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHead, definePageMeta, useI18n, useLocalePath } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

const { t } = useI18n()
const localePath = useLocalePath()

const storyParagraphs = computed(() => {
  const body = t('company.ourStory.story.body')
  return body.split('\n\n').filter(Boolean)
})

const factoryTabTo = computed(() => {
  return `${localePath('/company/about')}#factory`
})

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Our Story',
})
</script>

<style scoped>
.company-page {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.company-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__title--sr-only {
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

.company-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.company-tabs {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 4px 16px;
  margin: 0 -16px 1rem;
  max-width: calc(100% + 32px);
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  touch-action: pan-x;
}

.company-tabs::-webkit-scrollbar {
  display: none;
}

.company-tabs__item {
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  padding: 8px 18px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  backdrop-filter: blur(4px);
  box-shadow: 6px 8px 18px -12px rgba(0, 0, 0, 0.85);
}

.company-tabs__item:active {
  transform: scale(0.96);
}

.company-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.company-tabs__item--active {
  background: #ffffff;
  color: #0f172a;
  border: none;
  font-weight: 600;
  box-shadow: 8px 10px 22px -10px rgba(0, 0, 0, 0.9);
}

.company-section {
  margin-top: 0;
}

.company-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
  font-weight: 600;
  color: #e5e7eb;
  text-align: center;
}

.company-section__body {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.95);
  text-align: center;
}

.company-section__list {
  margin: 0;
  padding-left: 1.1rem;
  list-style-type: disc;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-section__list li + li {
  margin-top: 0.25rem;
}

.factory-flow-grid {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.factory-flow-card {
  position: relative;
  aspect-ratio: 1 / 1;
  border-radius: 0.9rem;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: #020617;
  box-shadow: 0 0 25px rgba(15, 23, 42, 0.9);
}

.factory-flow-card__image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform-origin: center;
  transition: transform 0.35s ease;
}

.factory-flow-card:hover .factory-flow-card__image {
  transform: scale(1.05);
}

.factory-flow-card__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding: 0.75rem 0.8rem 0.65rem;
  background: linear-gradient(
    to top,
    rgba(15, 23, 42, 0.96) 0%,
    rgba(15, 23, 42, 0.9) 35%,
    rgba(15, 23, 42, 0.6) 55%,
    transparent 100%
  );
  color: #f9fafb;
}

.factory-flow-card__step {
  margin: 0 0 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: #22d3ee;
}

.factory-flow-card__step-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: radial-gradient(circle at 30% 30%, #a5f3fc, #22d3ee 55%, #0ea5e9 100%);
  box-shadow: 0 0 9px rgba(56, 189, 248, 0.9);
}

.factory-flow-card__title {
  margin: 0 0 0.1rem;
  font-size: 0.9rem;
  font-weight: 600;
}

.factory-flow-card__text {
  margin: 0;
  font-size: 0.78rem;
  color: rgba(226, 232, 240, 0.9);
}

.company-section--values {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--timeline {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--cta {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-values {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

.company-values__item {
  padding: 0.75rem 0.9rem;
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.24);
}

.company-values__title {
  margin: 0 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-values__body {
  margin: 0;
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-timeline {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.company-timeline__item {
  display: grid;
  grid-template-columns: auto 1fr;
  column-gap: 0.75rem;
  row-gap: 0.25rem;
  align-items: flex-start;
}

.company-timeline__year {
  font-size: 0.85rem;
  font-weight: 600;
  color: #e5e7eb;
  min-width: 3.5rem;
}

.company-timeline__content {
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-feedback {
  margin-top: 2.5rem;
}

@media (min-width: 768px) {
  .company-page__title {
    font-size: 1.75rem;
  }

  .company-values {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .company-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .company-tabs {
    justify-content: flex-start;
  }

  .factory-flow-grid {
    grid-template-columns: 1fr;
  }
}
</style>
