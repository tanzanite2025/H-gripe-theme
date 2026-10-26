<template>
  <div
    v-if="shouldLoadFaqSlot && (pending || error || resolvedFaqData)"
    class="page-faq-slot"
    :aria-busy="pending"
  >
    <StorefrontFaqLoadingAndFailurePlaceholder
      :loading="pending"
      :failed="Boolean(error)"
      :has-saved-answers="Boolean(resolvedFaqData)"
      @retry="refresh()"
    />
    <LazyPageFaq
      v-if="resolvedFaqData && resolvedFaqPageId"
      :hydrate-on-visible="{ rootMargin: '0px 0px 300px 0px' }"
      :page-id="resolvedFaqPageId"
      :data="resolvedFaqData"
      :render-faq-structured-data="false"
      :heading-tag="faqLookupRoutePath === innerTubeGuideCanonicalRoutePath || faqLookupRoutePath === '/company/about/appearance' || faqLookupRoutePath === '/company/about/hole-patterns' || faqLookupRoutePath === '/tools/freehub-cassette-fitment-engine' ? 'h2' : 'h3'"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHead, useNuxtApp, useRoute } from '#imports'
import StorefrontFaqLoadingAndFailurePlaceholder from '~/components/faq/StorefrontFaqLoadingAndFailurePlaceholder.vue'
import { useStorefrontFaqAsyncDataWithLastSuccessfulResponse } from '~/composables/useStorefrontFaqAsyncDataWithLastSuccessfulResponse'
import { innerTubeGuideCanonicalRoutePath } from '~/utils/seo/innerTubeGuidePublishedLocaleIndexingPolicy'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import { getStorefrontLocaleLanguageTag } from '~/utils/storefrontLocales'
import { createFaqPageJsonLd, createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { fetchFaqData, fetchFaqDataByRoutePath, getPageFaqId, resolveFaqRouteLookupPath, shouldAutoInsertFaqForRoute } from '~/data/faq'

const route = useRoute()
const { locale } = useI18n()
const { canonicalUrl } = useStorefrontSeoLinks()

if (import.meta.server) {
  useNuxtApp().hook('app:rendered', ({ ssrContext }) => {
    // Nuxt collects stylesheet dependencies before processing lazy hydration.
    // Remove this subtree's registrations before that result also caches JS
    // preload hints. Eager consumers still preload these shared dependencies
    // through their own static imports; SSR HTML and inline styles are retained.
    for (const faqSubtreeModuleId of [
      'components/PageFaq.vue',
      'components/FaqAnswerContent.vue',
      'components/faq/DesktopFaqMasterDetail.vue',
      'components/SafeRichText.vue',
    ]) {
      ssrContext?.modules?.delete(faqSubtreeModuleId)
    }
  })
}

const shouldLoadFaqSlot = computed(() => shouldAutoInsertFaqForRoute(route.path))
const faqLookupRoutePath = computed(() => resolveFaqRouteLookupPath(route.path))
const configuredFaqContentPageId = computed(() => (
  typeof route.meta.faqPageId === 'string' ? route.meta.faqPageId.trim() : ''
))

const { data: faqData, pending, error, refresh } = await useStorefrontFaqAsyncDataWithLastSuccessfulResponse(
  () => `faq-slot-${locale.value}-${shouldLoadFaqSlot.value ? configuredFaqContentPageId.value || faqLookupRoutePath.value : 'disabled'}`,
  () => shouldLoadFaqSlot.value
    ? configuredFaqContentPageId.value
      ? fetchFaqData(configuredFaqContentPageId.value)
      : fetchFaqDataByRoutePath(faqLookupRoutePath.value)
    : Promise.resolve(null),
)

const resolvedFaqData = computed(() => faqData.value)
const resolvedFaqPageId = computed(() => resolvedFaqData.value ? getPageFaqId(resolvedFaqData.value) : '')

// Keep ownership of the SSR head entry even before the FAQ UI hydrates, so a
// client navigation replaces the previous page's schema rather than leaving it.
const pageFaqStructuredData = computed(() => shouldLoadFaqSlot.value && resolvedFaqData.value
  ? createFaqPageJsonLd({
    canonicalUrl: canonicalUrl.value,
    pageId: resolvedFaqPageId.value,
    inLanguage: getStorefrontLocaleLanguageTag(locale.value),
    visibleQuestionsAndAnswers: resolvedFaqData.value.items,
  })
  : null)

useHead(() => ({
  script: pageFaqStructuredData.value
    ? [{ ...createSeoJsonLdScript(pageFaqStructuredData.value), key: 'storefront-page-faq-structured-data' }]
    : [],
}))
</script>

<style scoped>
.page-faq-slot {
  width: 100%;
  margin-top: clamp(1.25rem, 2vw, 2rem);
  padding: 0;
}
</style>
