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
        <!-- Category Loop (Each category is a Premium Card) -->
        <div 
          v-for="category in displayCategories" 
          :key="category.id"
          class="faq-category-card rounded-2xl p-3 md:p-4"
          :class="theme === 'dark' ? 'bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)]' : 'bg-white shadow-lg'"
        >
          <!-- Category Header -->
          <div 
            v-if="showCategories"
            class="page-faq__category-header flex items-center justify-center gap-3 border-b"
            :class="theme === 'dark' ? 'border-slate-800/50' : 'border-gray-100'"
          >
            <h4 
              class="page-faq__category-title tz-faq-category-title"
              :class="theme === 'dark' ? 'tz-text-primary' : 'text-gray-800'"
            >
              {{ category.name }}
            </h4>
          </div>

          <!-- FAQ Items (Accordion Wrapper) -->
          <div 
            class="space-y-0 rounded-xl overflow-hidden border"
            :class="theme === 'dark' ? 'bg-slate-900/40 border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)]' : 'bg-white border-gray-200 shadow-sm'"
          >
            <div 
              v-for="item in getCategoryItems(category)" 
              :key="item.id"
              class="faq-item border-b last:border-b-0"
              :class="theme === 'dark' ? 'border-slate-800/50' : 'border-gray-100'"
            >
              <!-- Question (Accordion Header) -->
              <button
                type="button"
                class="w-full flex items-center justify-between gap-3 px-3 py-3 text-left transition-colors group"
                :class="[
                  theme === 'dark' 
                    ? 'hover:bg-white/5' 
                    : 'hover:bg-gray-50',
                  expandedItems.has(item.id) 
                    ? (theme === 'dark' ? 'bg-white/5' : 'bg-gray-50') 
                    : ''
                ]"
                @click="toggleItem(item.id)"
              >
                <span 
                  class="page-faq__question-text tz-faq-question flex-1 transition-colors"
                  :class="[
                    theme === 'dark' ? 'tz-text-secondary' : 'text-gray-800',
                    expandedItems.has(item.id) ? (theme === 'dark' ? 'text-sky-400' : 'text-blue-600') : 'group-hover:text-sky-400'
                  ]"
                >
                  {{ item.question }}
                </span>
                <span 
                  class="flex-shrink-0 w-6 h-6 flex items-center justify-center rounded-full transition-all duration-200"
                  :class="[
                    expandedItems.has(item.id) 
                      ? (theme === 'dark' ? 'bg-sky-500/10 text-sky-400 rotate-180' : 'bg-blue-100 text-blue-600 rotate-180')
                      : (theme === 'dark' ? 'tz-text-muted bg-transparent' : 'text-gray-500 bg-transparent')
                  ]"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                </span>
              </button>

              <!-- Answer (Accordion Content) -->
              <Transition
                enter-active-class="transition-all duration-200 ease-out"
                leave-active-class="transition-all duration-150 ease-in"
                enter-from-class="opacity-0 max-h-0"
                enter-to-class="opacity-100 max-h-[500px]"
                leave-from-class="opacity-100 max-h-[500px]"
                leave-to-class="opacity-0 max-h-0"
              >
                <div 
                  v-if="expandedItems.has(item.id)"
                  class="overflow-hidden bg-opacity-30"
                  :class="theme === 'dark' ? 'bg-slate-950/30' : 'bg-gray-50/50'"
                >
                  <div 
                    class="page-faq__answer tz-faq-answer px-4 pb-4 pt-1"
                    :class="theme === 'dark' ? 'tz-text-secondary' : 'text-gray-600'"
                    v-html="item.answer"
                  />
                </div>
              </Transition>
            </div>
          </div>
        </div>
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
import { ref, computed } from 'vue'
import { useLocalePath, useAsyncData } from '#imports'
import { fetchFaqData, getFaqData } from '~/data/faq'
import type { PageFaqProps, FaqCategory } from '~/data/faq/types'

const props = withDefaults(defineProps<PageFaqProps>(), {
  theme: 'dark',
  showCategories: true,
  showViewAllLink: false,
})

const localePath = useLocalePath()

// Get FAQ data for the page (Async from Go backend with static fallback)
const { data: asyncFaqData } = await useAsyncData(`faq-${props.pageId}`, () => fetchFaqData(props.pageId))
const faqData = computed(() => asyncFaqData.value || getFaqData(props.pageId))

// Display title (prop override or from data)
const displayTitle = computed(() => {
  return props.title || faqData.value?.title || 'Frequently Asked Questions'
})

// Track expanded accordion items
const expandedItems = ref<Set<string>>(new Set())

// Toggle accordion item
const toggleItem = (itemId: string) => {
  if (expandedItems.value.has(itemId)) {
    expandedItems.value.delete(itemId)
  } else {
    expandedItems.value.add(itemId)
  }
  // Trigger reactivity
  expandedItems.value = new Set(expandedItems.value)
}

// Get categories to display (respecting maxItems limit)
const displayCategories = computed(() => {
  if (!faqData.value?.categories) return []
  
  if (!props.maxItems) {
    return faqData.value.categories
  }

  // If maxItems is set, we need to limit total items across categories
  let remainingItems = props.maxItems
  const limitedCategories: FaqCategory[] = []

  for (const category of faqData.value.categories) {
    if (remainingItems <= 0) break

    const itemsToTake = Math.min(category.items.length, remainingItems)
    limitedCategories.push({
      ...category,
      items: category.items.slice(0, itemsToTake),
    })
    remainingItems -= itemsToTake
  }

  return limitedCategories
})

// Get items for a category (already limited by displayCategories)
const getCategoryItems = (category: FaqCategory) => {
  return category.items
}

// Check if there are more items than displayed
const hasMoreItems = computed(() => {
  if (!props.maxItems || !faqData.value?.categories) return false
  
  const totalItems = faqData.value.categories.reduce(
    (sum, cat) => sum + cat.items.length, 
    0
  )
  return totalItems > props.maxItems
})
</script>

<style scoped>
.page-faq {
  /* Smooth scrolling for anchor links */
  scroll-margin-top: 80px;
}

.page-faq__header {
  margin-bottom: 1.25rem;
}

.page-faq__category-header {
  margin-bottom: 0.75rem;
  padding-bottom: 0.55rem;
}

@media (min-width: 768px) {
  .page-faq__header {
    margin-bottom: 1.5rem;
  }

  .page-faq__category-header {
    margin-bottom: 0.9rem;
    padding-bottom: 0.7rem;
  }
}
</style>
