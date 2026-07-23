<template>
  <section 
    class="page-faq w-full"
    :class="[
      theme === 'dark' ? 'bg-transparent' : 'bg-gray-50',
      'py-4 md:py-6'
    ]"
  >
    <div class="w-full max-w-none mx-auto">
      <!-- Main Page Header (Optional, if page title not sufficient) -->
      <div v-if="displayTitle || faqData?.subtitle" class="page-faq__header text-center">
        <h3 
          v-if="displayTitle"
          class="page-faq__title tz-faq-title hidden"
          :class="theme === 'dark' ? 'tz-text-primary' : 'text-gray-800'"
        >
          {{ displayTitle }}
        </h3>
        <p 
          v-if="faqData?.subtitle"
          class="page-faq__subtitle tz-faq-subtitle max-w-2xl mx-auto"
          :class="theme === 'dark' ? 'tz-text-secondary' : 'text-gray-600'"
        >
          {{ faqData.subtitle }}
        </p>
      </div>

      <!-- FAQ Content -->
      <div v-if="faqData && displayCategories.length > 0" class="space-y-6 md:space-y-7">
        <FaqCategoryAccordion
          v-for="category in displayCategories"
          :key="category.id"
          :category="category"
          :theme="theme"
          :show-categories="showCategories"
          :expanded-items="expandedItems"
          @toggle-item="toggleItem"
        />
      </div>

      <!-- Empty State -->
      <div 
        v-else
        class="text-center py-12 rounded-2xl border-2 border-dashed"
        :class="theme === 'dark' ? 'border-slate-800 tz-text-muted' : 'border-gray-200 text-gray-500'"
      >
        <p class="text-sm">No FAQs available for this section.</p>
      </div>

      <!-- View All Link -->
      <div 
        v-if="showViewAllLink && hasMoreItems"
        class="text-center mt-8"
      >
        <NuxtLink
          :to="localePath('/support/faqs')"
          class="inline-flex items-center gap-2 px-6 py-2.5 rounded-full text-sm font-bold transition-all shadow-lg hover:-translate-y-0.5"
          :class="theme === 'dark' 
            ? 'bg-slate-800 tz-text-secondary hover:bg-slate-700 hover:text-white hover:shadow-slate-900/50'
            : 'bg-gray-800 text-white hover:bg-gray-700'"
        >
          View All FAQs
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useLocalePath } from '#imports'
import FaqCategoryAccordion from '~/components/faq/FaqCategoryAccordion.vue'
import { usePageFaq } from '~/composables/usePageFaq'
import type { PageFaqProps } from '~/data/faq/types'

const props = withDefaults(defineProps<PageFaqProps>(), {
  theme: 'dark',
  showCategories: true,
  showViewAllLink: false,
})

const localePath = useLocalePath()
const {
  faqData,
  displayTitle,
  displayCategories,
  expandedItems,
  toggleItem,
  hasMoreItems
} = await usePageFaq(props)
</script>

<style scoped>
.page-faq {
  /* Smooth scrolling for anchor links */
  scroll-margin-top: 80px;
}

.page-faq__header {
  margin-bottom: 1.25rem;
}

@media (min-width: 768px) {
  .page-faq__header {
    margin-bottom: 1.5rem;
  }
}
</style>
