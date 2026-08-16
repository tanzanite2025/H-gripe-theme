<template>
  <div class="global-all-faqs-desktop-grouped-search-results-panel">
    <div
      v-for="group in groups"
      :key="group.pageId"
      class="global-all-faqs-desktop-grouped-search-results-panel__group"
    >
      <div class="global-all-faqs-desktop-grouped-search-results-panel__group-header">
        <h3 class="tz-faq-category-title tz-text-primary">
          {{ group.pageTitle }}
        </h3>
      </div>

      <div class="global-all-faqs-desktop-grouped-search-results-panel__grid">
        <div
          v-for="(columnItems, columnIndex) in itemColumns(group.items)"
          :key="columnIndex"
          class="global-all-faqs-desktop-grouped-search-results-panel__column"
        >
          <div
            v-for="item in columnItems"
            :key="item.id"
            class="global-all-faqs-desktop-grouped-search-results-panel__item"
            :class="{ 'is-expanded': expandedItems.has(item.id) }"
          >
            <button
              type="button"
              class="global-all-faqs-desktop-grouped-search-results-panel__item-button"
              :class="{ 'is-expanded': expandedItems.has(item.id) }"
              @click="emit('toggle-item', item.id)"
            >
              <span class="global-all-faqs-desktop-grouped-search-results-panel__category">
                {{ item.category }}
              </span>
              <span class="global-all-faqs-desktop-grouped-search-results-panel__question">
                {{ item.question }}
              </span>
              <span
                class="global-all-faqs-desktop-grouped-search-results-panel__icon"
                :class="{ 'is-expanded': expandedItems.has(item.id) }"
                aria-hidden="true"
              >
                +
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
                class="global-all-faqs-desktop-grouped-search-results-panel__answer-wrap"
              >
                <div class="global-all-faqs-desktop-grouped-search-results-panel__answer">
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
  </div>
</template>

<script setup lang="ts">
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'
import type { GlobalAllFaqFlatItem, GlobalAllFaqsDisplayGroup } from '~/data/faq'

defineProps<{
  groups: GlobalAllFaqsDisplayGroup[]
  expandedItems: ReadonlySet<string>
}>()

const emit = defineEmits<{
  'toggle-item': [itemId: string]
}>()

const itemColumns = (items: GlobalAllFaqFlatItem[]): [GlobalAllFaqFlatItem[], GlobalAllFaqFlatItem[]] => {
  const columns: [GlobalAllFaqFlatItem[], GlobalAllFaqFlatItem[]] = [[], []]

  items.forEach((item, index) => {
    columns[index % 2]?.push(item)
  })

  return columns
}
</script>

<style scoped>
.global-all-faqs-desktop-grouped-search-results-panel {
  display: none;
  min-width: 0;
}

.global-all-faqs-desktop-grouped-search-results-panel__group {
  margin-bottom: 0.85rem;
}

.global-all-faqs-desktop-grouped-search-results-panel__group:last-child {
  margin-bottom: 0;
}

.global-all-faqs-desktop-grouped-search-results-panel__group-header {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  margin-bottom: 0.65rem;
  padding-bottom: 0.55rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.global-all-faqs-desktop-grouped-search-results-panel__group-header h3 {
  margin: 0;
  color: #94a3b8 !important;
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: 0.16em;
  line-height: 1.2;
  text-align: left;
  text-transform: uppercase;
}

.global-all-faqs-desktop-grouped-search-results-panel__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.625rem;
  align-items: start;
}

.global-all-faqs-desktop-grouped-search-results-panel__column {
  display: grid;
  gap: 0.625rem;
  align-content: start;
}

.global-all-faqs-desktop-grouped-search-results-panel__item {
  position: relative;
  align-self: start;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 1rem;
  background: #000000;
}

.global-all-faqs-desktop-grouped-search-results-panel__item::before {
  position: absolute;
  inset: 0 auto 0 0;
  width: 0;
  background: #B5FF6D;
  box-shadow: 0 0 10px #B5FF6D;
  content: '';
  transition: width 0.3s ease;
}

.global-all-faqs-desktop-grouped-search-results-panel__item.is-expanded {
  border-color: rgba(255, 255, 255, 0.3);
}

.global-all-faqs-desktop-grouped-search-results-panel__item.is-expanded::before {
  width: 3px;
}

.global-all-faqs-desktop-grouped-search-results-panel__item-button {
  position: relative;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  min-width: 0;
  padding: 1rem;
  border: 0;
  background: transparent;
  color: #ffffff;
  text-align: left;
  cursor: pointer;
}

.global-all-faqs-desktop-grouped-search-results-panel__item-button:hover,
.global-all-faqs-desktop-grouped-search-results-panel__item-button.is-expanded {
  background: rgba(255, 255, 255, 0.025);
}

.global-all-faqs-desktop-grouped-search-results-panel__category {
  max-width: 9rem;
  overflow: hidden;
  padding: 0.25rem 0.65rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.72);
  color: #94a3b8;
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.global-all-faqs-desktop-grouped-search-results-panel__question {
  min-width: 0;
  color: #ffffff;
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1.45;
}

.global-all-faqs-desktop-grouped-search-results-panel__icon {
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  font-size: 1.05rem;
  font-weight: 900;
  line-height: 1;
  transition: transform 0.2s ease, color 0.2s ease;
}

.global-all-faqs-desktop-grouped-search-results-panel__icon.is-expanded {
  color: #B5FF6D;
  transform: rotate(45deg);
}

.global-all-faqs-desktop-grouped-search-results-panel__answer-wrap {
  overflow: hidden;
  background: transparent;
}

.global-all-faqs-desktop-grouped-search-results-panel__answer {
  margin: 0 1rem 1rem;
  padding: 0.8rem 0 0;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  color: #cbd5e1;
  font-size: 0.78rem;
  line-height: 1.7;
}

@media (min-width: 768px) and (max-width: 1280px) {
  .global-all-faqs-desktop-grouped-search-results-panel__item-button {
    grid-template-columns: minmax(0, 1fr) auto;
    row-gap: 0.55rem;
  }

  .global-all-faqs-desktop-grouped-search-results-panel__category {
    grid-column: 1 / -1;
    max-width: 100%;
    justify-self: start;
  }
}

@media (min-width: 768px) {
  .global-all-faqs-desktop-grouped-search-results-panel {
    display: block;
  }
}
</style>
