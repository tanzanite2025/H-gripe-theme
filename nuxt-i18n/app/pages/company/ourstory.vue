<template>
  <div class="ourstory-page w-full">
    <section id="story" class="w-full max-w-none">
      <div class="ourstory-container relative overflow-hidden rounded-2xl px-4 pb-6 pt-5 md:rounded-3xl md:px-10 md:pb-10 md:pt-8">
        <div class="relative z-10">
          <header class="ourstory-header">
            <p class="ourstory-eyebrow">{{ t('company.nav.ourStory', 'Our Story') }}</p>
            <h1 class="ourstory-title">{{ t('company.nav.ourStory', 'Our Story') }}</h1>
            <p class="ourstory-intro">
              From a belief in full control to an in-house approach, our story has been shaped by the way we design, build, and support high-performance cycling components.
            </p>
          </header>

          <figure class="ourstory-hero group">
            <div class="ourstory-hero__frame">
              <div class="ourstory-image__scrim absolute inset-0 z-10"></div>
              <img
                class="h-full w-full object-cover transition-transform duration-700 group-hover:scale-105"
                src="/company/ourstory/ourstory/tanzanite-ourstory.webp"
                alt="our engineers and riders discussing product development"
                loading="lazy"
              />
            </div>
            <figcaption class="ourstory-hero__caption">
              Engineers and riders shaping the next stage of the journey.
            </figcaption>
          </figure>

          <ol class="ourstory-timeline" aria-label="Our Story milestones">
            <li
              v-for="(milestone, index) in storyMilestones"
              :key="milestone.title"
              class="ourstory-timeline__item"
            >
              <div class="ourstory-timeline__rail" aria-hidden="true">
                <span class="ourstory-timeline__marker">{{ String(index + 1).padStart(2, '0') }}</span>
              </div>

              <article class="ourstory-timeline__content">
                <p class="ourstory-timeline__phase">
                  Phase {{ String(index + 1).padStart(2, '0') }}
                </p>
                <h2 class="ourstory-timeline__title">{{ milestone.title }}</h2>
                <p class="ourstory-timeline__body">{{ milestone.body }}</p>
              </article>
            </li>
          </ol>

          <div class="ourstory-cta">
            <div class="min-w-0">
              <p class="ourstory-cta__eyebrow">
                See the work behind the story
              </p>
              <p class="ourstory-cta__body">
                Explore the factory, manufacturing flow, and quality systems that support our products.
              </p>
            </div>

            <NuxtLink
              class="ourstory-factory-link group inline-flex shrink-0 items-center gap-2 rounded-full px-5 py-2.5 text-sm font-medium tz-text-primary transition-all"
              :to="factoryTabTo"
            >
              <span>{{ t('company.ourStory.story.factoryButton') }}</span>
              <svg class="ourstory-factory-link__icon h-4 w-4 transition-transform group-hover:translate-x-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </NuxtLink>
          </div>
        </div>
      </div>
    </section>

    <div class="ourstory-feedback mt-4 mb-6 w-full max-w-none px-0">
      <UserFeedbackThread
        threadKey="company-ourstory"
        title="Share your feedback about Our Story"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useHead, definePageMeta, navigateTo, useI18n, useLocalePath, useRoute } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import {
  companyAboutTabs,
  isPageSubNavigationTabId,
  type CompanyAboutTabId,
} from '~/utils/pageSubNavigation'

const { t } = useI18n()
const localePath = useLocalePath()
const route = useRoute()

const storyParagraphs = computed(() => {
  const body = t('company.ourStory.story.body')
  return body.split('\n\n').filter(Boolean)
})

const storyMilestones = computed(() => {
  const titles = [
    'The Belief',
    'The Problem We Saw',
    'Our First Focus',
    'Complete Systems',
    'Built In-house',
    'Long-term Partnerships',
  ]

  return storyParagraphs.value.map((body, index) => ({
    title: titles[index] || `Phase ${String(index + 1).padStart(2, '0')}`,
    body,
  }))
})

const factoryTabTo = computed(() => {
  return `${localePath('/company/about')}#factory`
})

const getCompanyAboutTabFromHash = (hash: string): CompanyAboutTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  return isPageSubNavigationTabId(companyAboutTabs, raw) ? raw : null
}

watch(
  () => route.hash,
  (hash) => {
    const tab = getCompanyAboutTabFromHash(hash)
    if (!tab) return

    navigateTo(`${localePath('/company/about')}#${tab}`, { replace: true })
  },
  { immediate: true }
)

definePageMeta({
  layout: 'products',
})

useHead(() => ({
  title: t('company.nav.ourStory', 'Our Story'),
}))
</script>

<style scoped>
.ourstory-container {
  border: 1px solid rgba(255, 255, 255, 0.1);
  background:
    radial-gradient(circle at 86% 0%, rgba(181, 255, 109, 0.08), transparent 34rem),
    linear-gradient(180deg, rgba(255, 255, 255, 0.035), rgba(255, 255, 255, 0.012)),
    var(--tz-card-surface);
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.46);
}

.ourstory-header {
  max-width: 52rem;
  margin: 0 auto 1.75rem;
  text-align: center;
}

.ourstory-eyebrow,
.ourstory-timeline__phase,
.ourstory-cta__eyebrow {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.18em;
  line-height: 1.4;
  text-transform: uppercase;
}

.ourstory-title {
  margin: 0.35rem 0 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.85rem, 4vw, 3rem);
  font-weight: 800;
  line-height: 1.08;
}

.ourstory-intro {
  max-width: 42rem;
  margin: 0.9rem auto 0;
  color: var(--tz-text-secondary);
  font-size: 1rem;
  line-height: 1.7;
}

.ourstory-hero {
  margin: 0 auto;
}

.ourstory-hero__frame {
  position: relative;
  aspect-ratio: 21 / 9;
  width: 100%;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.9rem;
  box-shadow: 0 20px 38px rgba(0, 0, 0, 0.32);
}

.ourstory-image__scrim {
  background: linear-gradient(to top, rgba(0, 0, 0, 0.42), transparent);
}

.ourstory-hero__caption {
  margin-top: 0.65rem;
  color: var(--tz-text-muted);
  font-size: 0.78rem;
  line-height: 1.5;
  text-align: center;
}

.ourstory-timeline {
  position: relative;
  display: grid;
  gap: 0;
  max-width: 58rem;
  margin: 2.5rem auto 0;
  padding: 0;
  list-style: none;
}

.ourstory-timeline::before {
  position: absolute;
  top: 1.2rem;
  bottom: 1.2rem;
  left: 1.05rem;
  width: 1px;
  content: '';
  background: linear-gradient(
    to bottom,
    rgba(181, 255, 109, 0.55),
    rgba(255, 255, 255, 0.12) 42%,
    rgba(181, 255, 109, 0.35)
  );
}

.ourstory-timeline__item {
  position: relative;
  display: grid;
  grid-template-columns: 2.15rem minmax(0, 1fr);
  gap: 1rem;
}

.ourstory-timeline__rail {
  position: relative;
  z-index: 1;
}

.ourstory-timeline__marker {
  display: inline-flex;
  width: 2.15rem;
  height: 2.15rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(181, 255, 109, 0.45);
  border-radius: 9999px;
  background: #111116;
  color: var(--tz-text-accent);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.ourstory-timeline__content {
  min-width: 0;
  padding: 0 0 1.65rem;
}

.ourstory-timeline__item:not(:first-child) .ourstory-timeline__content {
  padding-top: 0.2rem;
}

.ourstory-timeline__title {
  margin: 0.3rem 0 0.55rem;
  color: var(--tz-text-primary);
  font-size: 1.2rem;
  font-weight: 700;
  line-height: 1.25;
}

.ourstory-timeline__body {
  max-width: 48rem;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.98rem;
  line-height: 1.75;
}

.ourstory-cta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.25rem;
  max-width: 58rem;
  margin: 0.25rem auto 0;
  padding-top: 1.35rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.ourstory-cta__body {
  max-width: 42rem;
  margin: 0.35rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  line-height: 1.6;
}

.ourstory-factory-link {
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(17, 17, 22, 0.86);
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.36);
}

.ourstory-factory-link:hover {
  border-color: rgba(181, 255, 109, 0.36);
  background: rgba(26, 26, 31, 0.92);
}

.ourstory-factory-link__icon {
  color: var(--tz-text-accent);
}

@media (max-width: 640px) {
  .ourstory-header {
    margin-bottom: 1.35rem;
  }

  .ourstory-intro {
    font-size: 0.92rem;
    line-height: 1.6;
  }

  .ourstory-hero__frame {
    aspect-ratio: 4 / 3;
  }

  .ourstory-timeline {
    margin-top: 2rem;
  }

  .ourstory-timeline__body {
    font-size: 0.92rem;
    line-height: 1.65;
  }

  .ourstory-cta {
    align-items: flex-start;
    flex-direction: column;
  }

  .ourstory-factory-link {
    align-self: flex-start;
  }
}
</style>
