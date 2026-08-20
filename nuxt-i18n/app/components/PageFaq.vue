<template>
  <section 
    class="page-faq w-full"
    :class="[
      theme === 'dark' ? 'bg-transparent' : 'bg-gray-50',
      'py-4 md:py-6'
    ]"
  >
    <div class="page-faq__shell w-full max-w-none mx-auto">
      <!-- Main Page Header (Optional, if page title not sufficient) -->
      <div v-if="displayTitle || faqData?.subtitle" class="page-faq__header text-center">
        <div class="page-faq__header-main">
        <h3 
          v-if="displayTitle"
          class="page-faq__title tz-faq-title"
          :class="theme === 'dark' ? 'tz-text-primary' : 'text-gray-800'"
        >
          {{ displayTitle }}
        </h3>
          <span class="page-faq__status-badge" aria-hidden="true">
            <span class="page-faq__status-dot" />
            {{ t('faq.ui.categorizedSupport') }}
          </span>
        </div>
        <p 
          v-if="faqData?.subtitle"
          class="page-faq__subtitle tz-faq-subtitle max-w-2xl mx-auto"
          :class="theme === 'dark' ? 'tz-text-secondary' : 'text-gray-600'"
        >
          {{ faqData.subtitle }}
        </p>
      </div>

      <!-- FAQ Content -->
      <div v-if="faqData && displayCategories.length > 0">
        <div class="page-faq__desktop-layout">
          <aside class="page-faq__sidebar" :aria-label="t('faq.ui.categoriesAriaLabel')">
            <div class="page-faq__sidebar-label">{{ t('faq.ui.categoriesLabel') }}</div>
            <button
              v-for="(category, index) in displayCategories"
              :key="category.id"
              type="button"
              class="page-faq__sidebar-button"
              :class="{ 'page-faq__sidebar-button--active': activeCategoryId === category.id }"
              @click="selectCategory(category.id)"
            >
              <span class="page-faq__sidebar-text">
                {{ formatCategoryIndex(index) }}. {{ category.name }}
              </span>
              <span class="page-faq__sidebar-dot" aria-hidden="true" />
            </button>
          </aside>

          <main class="page-faq__desktop-panel">
            <FaqCategoryAccordion
              v-if="activeCategory"
              :category="activeCategory"
              :page-title="displayTitle"
              :theme="theme"
              :show-categories="showCategories"
              :expanded-items="expandedItems"
              @toggle-item="toggleItem"
            />
          </main>
        </div>

        <div class="page-faq__mobile-list space-y-6 md:space-y-7">
          <FaqCategoryAccordion
            v-for="category in displayCategories"
            :key="category.id"
            :category="category"
            :page-title="displayTitle"
            :theme="theme"
            :show-categories="showCategories"
            :expanded-items="expandedItems"
            @toggle-item="toggleItem"
          />
        </div>
      </div>

      <!-- Empty State -->
      <div 
        v-else
        class="text-center py-12 rounded-2xl border-2 border-dashed"
        :class="theme === 'dark' ? 'border-slate-800 tz-text-muted' : 'border-gray-200 text-gray-500'"
      >
        <p class="text-sm">{{ t('faq.ui.emptySection') }}</p>
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
          {{ t('faq.ui.viewAll') }}
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useLocalePath } from '#imports'
import FaqCategoryAccordion from '~/components/faq/FaqCategoryAccordion.vue'
import { usePageFaq } from '~/composables/usePageFaq'
import type { PageFaqProps } from '../data/faq/types'

const props = withDefaults(defineProps<PageFaqProps>(), {
  theme: 'dark',
  showCategories: true,
  showViewAllLink: false,
})

const { t } = useI18n()
const localePath = useLocalePath()
const {
  faqData,
  displayTitle,
  displayCategories,
  expandedItems,
  toggleItem,
  resetExpandedItems,
  hasMoreItems
} = await usePageFaq(props)

const activeCategoryId = ref('')

watch(
  displayCategories,
  (categories) => {
    if (!categories.some((category) => category.id === activeCategoryId.value)) {
      activeCategoryId.value = categories[0]?.id || ''
      resetExpandedItems()
    }
  },
  { immediate: true }
)

const activeCategory = computed(() => {
  return displayCategories.value.find((category) => category.id === activeCategoryId.value) || displayCategories.value[0] || null
})

const selectCategory = (categoryId: string) => {
  if (activeCategoryId.value === categoryId) return

  activeCategoryId.value = categoryId
  resetExpandedItems()
}

const formatCategoryIndex = (index: number) => String(index + 1).padStart(2, '0')
</script>

<style scoped>
.page-faq {
  /* Smooth scrolling for anchor links */
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}

.page-faq__header {
  margin-bottom: 1.25rem;
}

.page-faq__title,
.page-faq__status-badge,
.page-faq__desktop-layout {
  display: none;
}

@media (min-width: 768px) {
  .page-faq {
    padding: 2rem 0 2.25rem;
    background-color: #000000;
    background-image: radial-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 0);
    background-size: 24px 24px;
  }

  .page-faq__shell {
    max-width: min(100rem, calc(100vw - 5rem));
  }

  .page-faq__header {
    margin-bottom: 1.5rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    text-align: left;
  }

  .page-faq__header-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .page-faq__title {
    display: block;
    color: #ffffff !important;
    font-size: 1.25rem;
    font-style: italic;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .page-faq__status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
    padding: 0.18rem 0.65rem;
    border-radius: 999px;
    border: 1px solid rgba(181, 255, 109, 0.32);
    background: rgba(6, 78, 59, 0.36);
    color: #B5FF6D;
    font-size: 0.58rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .page-faq__status-dot {
    width: 0.38rem;
    height: 0.38rem;
    border-radius: 999px;
    background: #B5FF6D;
    box-shadow: 0 0 12px rgba(181, 255, 109, 0.8);
    animation: page-faq-status-pulse 1s ease-in-out infinite alternate;
  }

  .page-faq__subtitle {
    max-width: none;
    margin: 0.25rem 0 0;
    color: #94a3b8 !important;
    font-size: 0.78rem;
    text-align: left;
  }

  .page-faq__desktop-layout {
    display: grid;
    grid-template-columns: minmax(11rem, 0.2fr) minmax(0, 0.8fr);
    gap: 1.5rem;
    align-items: start;
  }

  .page-faq__mobile-list {
    display: none;
  }

  .page-faq__sidebar {
    position: sticky;
    top: calc(112px + 1rem);
    display: grid;
    min-width: 0;
    max-width: 100%;
    gap: 0.25rem;
    box-sizing: border-box;
    padding: 0.75rem;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 1rem;
    background: #000000;
  }

  .page-faq__sidebar-label {
    padding: 0.25rem 0.75rem 0.35rem;
    color: #64748b;
    font-size: 0.58rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .page-faq__sidebar-button {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    width: 100%;
    min-width: 0;
    max-width: 100%;
    min-height: 2.35rem;
    box-sizing: border-box;
    padding: 0.68rem 0.85rem;
    border: 0;
    border-radius: 0.75rem;
    background: transparent;
    color: #94a3b8;
    text-align: left;
    font-size: 0.74rem;
    font-weight: 900;
    line-height: 1.25;
    text-transform: uppercase;
    cursor: pointer;
    transition: background 0.2s ease, color 0.2s ease, transform 0.2s ease;
  }

  .page-faq__sidebar-button:hover {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.04);
  }

  .page-faq__sidebar-button--active {
    color: #000000;
    background: #ffffff;
    box-shadow: 0 8px 18px rgba(0, 0, 0, 0.35);
  }

  .page-faq__sidebar-text {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .page-faq__sidebar-dot {
    width: 0.42rem;
    height: 0.42rem;
    flex-shrink: 0;
    border-radius: 999px;
    background: #334155;
  }

  .page-faq__sidebar-button--active .page-faq__sidebar-dot {
    background: #B5FF6D;
    box-shadow: 0 0 10px rgba(181, 255, 109, 0.7);
  }

  .page-faq__desktop-panel {
    min-width: 0;
    min-height: 26rem;
  }
}

@keyframes page-faq-status-pulse {
  from {
    opacity: 0.45;
    transform: scale(0.86);
  }

  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
