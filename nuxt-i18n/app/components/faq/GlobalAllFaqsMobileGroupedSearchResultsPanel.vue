<template>
  <div class="global-all-faqs-mobile-grouped-search-results-panel">
    <div
      v-for="group in groups"
      :key="group.pageId"
      class="global-all-faqs-mobile-grouped-search-results-panel__group"
    >
      <div class="global-all-faqs-mobile-grouped-search-results-panel__group-header">
        <h3 class="tz-faq-category-title tz-text-primary">
          {{ group.pageTitle }}
        </h3>
      </div>

      <div class="global-all-faqs-mobile-grouped-search-results-panel__list">
        <div
          v-for="item in group.items"
          :key="item.id"
          class="global-all-faqs-mobile-grouped-search-results-panel__item"
          :class="{ 'is-expanded': expandedItems.has(item.id) }"
        >
          <button
            type="button"
            class="global-all-faqs-mobile-grouped-search-results-panel__item-button"
            :class="{ 'is-expanded': expandedItems.has(item.id) }"
            @click="emit('toggle-item', item.id)"
          >
            <span class="global-all-faqs-mobile-grouped-search-results-panel__category">
              {{ item.category }}
            </span>
            <span class="global-all-faqs-mobile-grouped-search-results-panel__question">
              {{ item.question }}
            </span>
            <span
              class="global-all-faqs-mobile-grouped-search-results-panel__icon"
              :class="{ 'is-expanded': expandedItems.has(item.id) }"
              aria-hidden="true"
            >
              <svg
                class="global-all-faqs-mobile-grouped-search-results-panel__chevron"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 9l-7 7-7-7"
                />
              </svg>
            </span>
          </button>

          <Transition
            enter-active-class="transition-all duration-200 ease-out"
            leave-active-class="transition-all duration-150 ease-in"
            enter-from-class="opacity-0 max-h-0"
            enter-to-class="opacity-100 max-h-none"
            leave-from-class="opacity-100 max-h-none"
            leave-to-class="opacity-0 max-h-0"
          >
            <div
              v-if="expandedItems.has(item.id)"
              class="global-all-faqs-mobile-grouped-search-results-panel__answer-wrap"
            >
              <div class="global-all-faqs-mobile-grouped-search-results-panel__answer">
                <FaqAnswerContent
                  :answer="item.answer"
                  :image-url="item.answerImageUrl"
                  :image-alt="item.answerImageAlt"
                  :image-width="item.answerImageWidth"
                  :image-height="item.answerImageHeight"
                />
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'
import type { GlobalAllFaqsDisplayGroup } from '~/data/faq'

defineProps<{
  groups: GlobalAllFaqsDisplayGroup[]
  expandedItems: ReadonlySet<string>
}>()

const emit = defineEmits<{
  'toggle-item': [itemId: string]
}>()
</script>

<style scoped>
.global-all-faqs-mobile-grouped-search-results-panel {
  display: block;
  min-width: 0;
}

.global-all-faqs-mobile-grouped-search-results-panel__group {
  margin-bottom: 2rem;
  padding: 0.75rem;
  border-radius: 1rem;
  background: var(--tz-card-surface);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.6);
}

.global-all-faqs-mobile-grouped-search-results-panel__group-header {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(30, 41, 59, 0.5);
}

.global-all-faqs-mobile-grouped-search-results-panel__group-header h3 {
  margin: 0;
  text-align: center;
}

.global-all-faqs-mobile-grouped-search-results-panel__list {
  overflow: hidden;
  border: 1px solid rgba(30, 41, 59, 0.5);
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.4);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
}

.global-all-faqs-mobile-grouped-search-results-panel__item {
  border-bottom: 1px solid rgba(30, 41, 59, 0.5);
}

.global-all-faqs-mobile-grouped-search-results-panel__item:last-child {
  border-bottom: 0;
}

.global-all-faqs-mobile-grouped-search-results-panel__item-button {
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 0;
  background: transparent;
  color: var(--tz-text-primary);
  text-align: left;
  cursor: pointer;
  transition: background 0.2s ease;
}

.global-all-faqs-mobile-grouped-search-results-panel__item-button:hover,
.global-all-faqs-mobile-grouped-search-results-panel__item-button.is-expanded {
  background: rgba(255, 255, 255, 0.05);
}

.global-all-faqs-mobile-grouped-search-results-panel__category {
  flex-shrink: 0;
  max-width: 36%;
  overflow: hidden;
  padding: 0.25rem 0.55rem;
  border: 1px solid rgba(51, 65, 85, 1);
  border-radius: 999px;
  background: rgba(30, 41, 59, 0.85);
  color: var(--tz-text-muted);
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.global-all-faqs-mobile-grouped-search-results-panel__question {
  min-width: 0;
  flex: 1;
  color: var(--tz-text-primary);
  font-size: 0.84rem;
  line-height: 1.45;
}

.global-all-faqs-mobile-grouped-search-results-panel__icon {
  display: inline-flex;
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  color: var(--tz-text-muted);
  transition: transform 0.2s ease, color 0.2s ease;
}

.global-all-faqs-mobile-grouped-search-results-panel__icon.is-expanded {
  color: #B5FF6D;
  transform: rotate(180deg);
}

.global-all-faqs-mobile-grouped-search-results-panel__chevron {
  width: 1rem;
  height: 1rem;
}

.global-all-faqs-mobile-grouped-search-results-panel__answer-wrap {
  overflow: hidden;
  background: rgba(15, 23, 42, 0.3);
}

.global-all-faqs-mobile-grouped-search-results-panel__answer {
  padding: 0.25rem 1rem 1rem;
  color: var(--tz-text-secondary);
  line-height: 1.7;
}

@media (min-width: 768px) {
  .global-all-faqs-mobile-grouped-search-results-panel {
    display: none;
  }
}
</style>
