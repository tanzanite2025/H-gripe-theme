<template>
  <div class="tire-pressure-calculator-page">
    <NuxtLink
      class="tire-pressure-calculator-page__back-link"
      :to="localePath('/guides/tireguides/tire-pressure')"
    >
      {{ t('guidesTirePressure.calculator.backToStandards') }}
    </NuxtLink>

    <TirePressureCalculator />
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, useI18n, useLocalePath } from '#imports'
import TirePressureCalculator from '~/components/tireguides/tirepressure/TirePressureCalculator.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  path: '/guides/tireguides/tire-pressure-calculator',
  layout: 'products',
  footerLabelKey: 'products.nav.tireSizeCharts',
  footerLabelFallback: 'Tire Guides',
  feedbackThreadKey: 'guides-tire-pressure-calculator',
  feedbackTitleKey: 'guidesTirePressure.feedbackTitle',
  feedbackTitle: 'Share your feedback about the tire-pressure calculator',
  pageTitleKey: 'guidesTirePressure.calculator.title',
  pageTitle: 'Tire pressure, load, and cornering force demonstration',
})

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('guidesTirePressure')

await loadPageMessages(locale.value)
watch(locale, (nextLocale) => void loadPageMessages(nextLocale))

useHead(() => ({
  title: t('guidesTirePressure.calculator.title'),
  meta: [{ name: 'description', content: t('guidesTirePressure.calculator.description') }],
  script: [{
    key: 'tire-pressure-engineering-jsonld',
    type: 'application/ld+json',
    children: JSON.stringify({
      '@context': 'https://schema.org',
      '@graph': [
        {
          '@type': 'TechArticle',
          headline: t('guidesTirePressure.calculator.title'),
          description: t('guidesTirePressure.calculator.description'),
          version: 'v1.1-MVP-Review',
          proficiencyLevel: 'Expert',
          author: {
            '@type': 'Organization',
            name: 'Tanzanite Engineering Laboratory',
          },
          inLanguage: locale.value,
          articleBody: t('guidesTirePressure.dashboard.scopeBody'),
        },
        {
          '@type': 'SoftwareApplication',
          name: t('guidesTirePressure.dashboard.title'),
          operatingSystem: 'All',
          applicationCategory: 'EngineeringApplication',
          isAccessibleForFree: true,
          description: t('guidesTirePressure.dashboard.noticeBody'),
        },
      ],
    }),
  }],
}))
</script>

<style scoped>
.tire-pressure-calculator-page {
  width: 100%;
}

.tire-pressure-calculator-page__back-link {
  display: inline-flex;
  align-items: center;
  margin-bottom: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  font-weight: 600;
  text-decoration: none;
}

.tire-pressure-calculator-page__back-link:hover,
.tire-pressure-calculator-page__back-link:focus-visible {
  color: var(--tz-text-accent);
  text-decoration: underline;
  outline: none;
}
</style>
