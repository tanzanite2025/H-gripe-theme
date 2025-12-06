<template>
  <section 
    class="page-faq w-full"
    :class="[
      theme === 'dark' ? 'bg-transparent' : 'bg-gray-50',
      'py-4 md:py-6'
    ]"
  >
    <div class="max-w-4xl mx-auto">
      <!-- Header -->
      <div class="text-center mb-4 md:mb-5">
        <h3 
          class="text-base md:text-lg font-semibold mb-1"
          :class="theme === 'dark' ? 'text-white/90' : 'text-gray-800'"
        >
          {{ displayTitle }}
        </h3>
        <p 
          v-if="faqData?.subtitle"
          class="text-xs md:text-sm"
          :class="theme === 'dark' ? 'text-white/50' : 'text-gray-500'"
        >
          {{ faqData.subtitle }}
        </p>
      </div>

      <!-- FAQ Content -->
      <div v-if="faqData && displayCategories.length > 0" class="space-y-4">
        <!-- Category Loop -->
        <div 
          v-for="category in displayCategories" 
          :key="category.id"
          class="faq-category"
        >
          <!-- Category Header -->
          <div 
            v-if="showCategories && displayCategories.length > 1"
            class="flex items-center gap-1.5 mb-2"
          >
            <span v-if="category.icon" class="text-sm">{{ category.icon }}</span>
            <h4 
              class="text-sm font-medium"
              :class="theme === 'dark' ? 'text-white/80' : 'text-gray-700'"
            >
              {{ category.name }}
            </h4>
          </div>

          <!-- FAQ Items (Accordion) -->
          <div 
            class="space-y-1 rounded-lg overflow-hidden"
            :class="theme === 'dark' ? 'bg-white/5' : 'bg-white shadow-sm'"
          >
            <div 
              v-for="item in getCategoryItems(category)" 
              :key="item.id"
              class="faq-item border-b last:border-b-0"
              :class="theme === 'dark' ? 'border-white/10' : 'border-gray-100'"
            >
              <!-- Question (Accordion Header) -->
              <button
                type="button"
                class="w-full flex items-center justify-between gap-3 px-3 md:px-4 py-2.5 text-left transition-colors"
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
                  class="text-xs md:text-sm font-medium flex-1"
                  :class="theme === 'dark' ? 'text-white/90' : 'text-gray-800'"
                >
                  {{ item.question }}
                </span>
                <svg 
                  class="w-4 h-4 flex-shrink-0 transition-transform duration-200"
                  :class="[
                    theme === 'dark' ? 'text-white/50' : 'text-gray-400',
                    expandedItems.has(item.id) ? 'rotate-180' : ''
                  ]"
                  fill="none" 
                  stroke="currentColor" 
                  viewBox="0 0 24 24"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
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
                  class="overflow-hidden"
                >
                  <div 
                    class="px-3 md:px-4 pb-3 text-xs md:text-sm leading-relaxed"
                    :class="theme === 'dark' ? 'text-white/60' : 'text-gray-600'"
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
        class="text-center py-6"
        :class="theme === 'dark' ? 'text-white/40' : 'text-gray-400'"
      >
        <p class="text-xs">No FAQs available for this page.</p>
      </div>

      <!-- View All Link -->
      <div 
        v-if="showViewAllLink && hasMoreItems"
        class="text-center mt-4"
      >
        <NuxtLink
          :to="localePath('/support/faqs')"
          class="inline-flex items-center gap-1.5 px-4 py-2 rounded-full text-xs font-medium transition-all"
          :class="theme === 'dark' 
            ? 'bg-white/10 text-white/80 hover:bg-white/15 border border-white/20' 
            : 'bg-gray-800 text-white hover:bg-gray-700'"
        >
          View All FAQs
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useLocalePath } from '#imports'
import { getFaqData } from '~/data/faq'
import type { PageFaqProps, FaqCategory } from '~/data/faq/types'

const props = withDefaults(defineProps<PageFaqProps>(), {
  theme: 'dark',
  showCategories: true,
  showViewAllLink: false,
})

const localePath = useLocalePath()

// Get FAQ data for the page
const faqData = computed(() => getFaqData(props.pageId))

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

/* Ensure answer content links are styled */
.faq-item :deep(a) {
  color: #6b73ff;
  text-decoration: underline;
}

.faq-item :deep(a:hover) {
  color: #40ffaa;
}

/* List styling in answers */
.faq-item :deep(ul),
.faq-item :deep(ol) {
  padding-left: 1.5rem;
  margin: 0.5rem 0;
}

.faq-item :deep(li) {
  margin: 0.25rem 0;
}
</style>
