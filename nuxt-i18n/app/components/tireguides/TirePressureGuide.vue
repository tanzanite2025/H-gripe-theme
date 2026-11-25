<template>
  <div class="tire-pressure-guide">
    <InteractiveToolPageHeader
      :title="t('guidesTirePressure.title')"
    />
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
      <TirePressureSection />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, useI18n } from '#imports'
import TirePressureSection from '~/components/TirePressureSection.vue'
import InteractiveToolPageHeader from '~/components/InteractiveToolPageHeader.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import localeManifest from '~/i18n/locales.manifest'

const { locale, t } = useI18n()
const { siteTitle } = useSiteTitle()
const { canonicalUrl } = useStorefrontSeoLinks()
const { loadPageMessages } = usePageMessages('guidesTirePressure')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  script: [{
    key: 'dynamic-tire-pressure-calculator-web-application-schema',
    id: 'dynamic-tire-pressure-calculator-web-application-schema',
    ...createSeoJsonLdScript({
      '@context': 'https://schema.org',
      '@type': 'WebApplication',
      '@id': `${canonicalUrl.value}#dynamic-tire-pressure-calculator`,
      url: canonicalUrl.value,
      mainEntityOfPage: canonicalUrl.value,
      name: t('guidesTirePressure.title'),
      description: t('guidesTirePressure.description'),
      inLanguage: localeManifest.find(entry => entry.code === locale.value)?.iso || locale.value,
      applicationCategory: 'EngineeringApplication',
      operatingSystem: 'Web',
      isAccessibleForFree: true,
      author: { '@type': 'Person', name: 'wangguangen' },
      ...(siteTitle.value.trim() ? {
        publisher: { '@type': 'Organization', name: siteTitle.value.trim() },
      } : {}),
    }),
  }],
}))

</script>
