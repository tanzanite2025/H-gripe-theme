<template>
  <section v-if="pending" class="page-faq-slot" aria-busy="true">
    <div class="page-faq-slot__loader">
      <span class="page-faq-slot__loader-dot" aria-hidden="true" />
      <span class="page-faq-slot__loader-text">LOAD</span>
    </div>
  </section>

  <PageFaq
    v-else-if="resolvedFaqData"
    :page-id="resolvedFaqData.pageId"
    :data="resolvedFaqData"
    :show-categories="true"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAsyncData, useRoute } from '#imports'
import PageFaq from '~/components/PageFaq.vue'
import { fetchFaqDataByRoutePath, normalizeFaqRoutePath } from '~/data/faq'

const route = useRoute()

const normalizedRoutePath = computed(() => normalizeFaqRoutePath(route.path))

const { data: faqData, pending } = await useAsyncData(
  () => `faq-slot-${normalizedRoutePath.value}`,
  () => fetchFaqDataByRoutePath(normalizedRoutePath.value),
  { watch: [normalizedRoutePath] }
)

const resolvedFaqData = computed(() => faqData.value)
</script>

<style scoped>
.page-faq-slot {
  width: 100%;
  padding: 1rem 0;
}

.page-faq-slot__loader {
  display: flex;
  min-height: 5.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.7rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(148, 163, 184, 0.14);
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.12), transparent 36%),
    rgba(15, 23, 42, 0.62);
  color: var(--tz-text-secondary);
}

.page-faq-slot__loader-dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 999px;
  background: #38bdf8;
  box-shadow: 0 0 18px rgba(56, 189, 248, 0.75);
  animation: page-faq-slot-pulse 0.95s ease-in-out infinite alternate;
}

.page-faq-slot__loader-text {
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.22em;
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
</style>
