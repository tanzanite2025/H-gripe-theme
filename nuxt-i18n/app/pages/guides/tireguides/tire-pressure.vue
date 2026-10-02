<template>
  <TirePressureGuide />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, useI18n } from '#imports'
import TirePressureGuide from '~/components/tireguides/TirePressureGuide.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'products.nav.tireSizeCharts',
  footerLabelFallback: 'Tire Guides',
})

const { locale, t } = useI18n()
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
