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
    <PageFaq
      v-if="resolvedFaqData && resolvedFaqPageId"
      :page-id="resolvedFaqPageId"
      :data="resolvedFaqData"
      :heading-tag="faqLookupRoutePath === innerTubeGuideCanonicalRoutePath ? 'h2' : 'h3'"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from '#imports'
import StorefrontFaqLoadingAndFailurePlaceholder from '~/components/faq/StorefrontFaqLoadingAndFailurePlaceholder.vue'
import { useStorefrontFaqAsyncDataWithLastSuccessfulResponse } from '~/composables/useStorefrontFaqAsyncDataWithLastSuccessfulResponse'
import PageFaq from '~/components/PageFaq.vue'
import { innerTubeGuideCanonicalRoutePath } from '~/utils/seo/innerTubeGuidePublishedLocaleIndexingPolicy'
import { fetchFaqDataByRoutePath, getPageFaqId, resolveFaqRouteLookupPath, shouldAutoInsertFaqForRoute } from '~/data/faq'

const route = useRoute()
const { locale } = useI18n()

const shouldLoadFaqSlot = computed(() => shouldAutoInsertFaqForRoute(route.path))
const faqLookupRoutePath = computed(() => resolveFaqRouteLookupPath(route.path))

const { data: faqData, pending, error, refresh } = await useStorefrontFaqAsyncDataWithLastSuccessfulResponse(
  () => `faq-slot-${locale.value}-${shouldLoadFaqSlot.value ? faqLookupRoutePath.value : 'disabled'}`,
  () => shouldLoadFaqSlot.value ? fetchFaqDataByRoutePath(faqLookupRoutePath.value) : Promise.resolve(null),
)

const resolvedFaqData = computed(() => faqData.value)
const resolvedFaqPageId = computed(() => resolvedFaqData.value ? getPageFaqId(resolvedFaqData.value) : '')
</script>

<style scoped>
.page-faq-slot {
  width: 100%;
  margin-top: clamp(1.25rem, 2vw, 2rem);
  padding: 0;
}
</style>
