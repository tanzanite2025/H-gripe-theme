<template>
  <div class="global-all-faqs-desktop-search-accordion">
    <div class="global-all-faqs-desktop-search-accordion__list">
      <div
        v-for="item in items"
        :key="item.id"
        class="global-all-faqs-desktop-search-accordion__item"
        :class="{ 'is-expanded': expandedItems.has(item.id) }"
      >
        <button
          type="button"
          class="global-all-faqs-desktop-search-accordion__button"
          :class="{ 'is-expanded': expandedItems.has(item.id) }"
          :aria-expanded="expandedItems.has(item.id)"
          :aria-controls="answerId(item.id)"
          @click="emit('toggle-item', item.id)"
        >
          <span class="global-all-faqs-desktop-search-accordion__main">
            <span class="global-all-faqs-desktop-search-accordion__meta">
              <span class="global-all-faqs-desktop-search-accordion__category">
                {{ item.category }}
              </span>
              <span class="global-all-faqs-desktop-search-accordion__page">
                {{ item.pageTitle }}
              </span>
            </span>
            <span class="global-all-faqs-desktop-search-accordion__question">
              {{ item.question }}
            </span>
          </span>
          <Icon
            name="lucide:plus"
            class="global-all-faqs-desktop-search-accordion__icon"
            :class="{ 'is-expanded': expandedItems.has(item.id) }"
            aria-hidden="true"
          />
        </button>
      </div>
    </div>

    <aside class="global-all-faqs-desktop-search-accordion__detail">
      <Transition
        mode="out-in"
        enter-active-class="transition-all duration-300 ease-out"
        leave-active-class="transition-all duration-200 ease-in"
        enter-from-class="opacity-0 translate-y-2"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div
          v-if="selectedItem"
          :key="selectedItem.id"
          :id="answerId(selectedItem.id)"
          class="global-all-faqs-desktop-search-accordion__detail-content"
          role="region"
          :aria-label="selectedItem.question"
        >
          <div class="global-all-faqs-desktop-search-accordion__detail-meta">
            <span class="global-all-faqs-desktop-search-accordion__category">
              {{ selectedItem.category }}
            </span>
            <span class="global-all-faqs-desktop-search-accordion__page">
              {{ selectedItem.pageTitle }}
            </span>
          </div>
          <h4 class="global-all-faqs-desktop-search-accordion__detail-question">
            {{ selectedItem.question }}
          </h4>
          <div class="global-all-faqs-desktop-search-accordion__answer">
            <FaqAnswerContent
              :answer="selectedItem.answer"
              :image-url="selectedItem.answerImageUrl"
              :image-alt="selectedItem.answerImageAlt"
              :image-width="selectedItem.answerImageWidth"
              :image-height="selectedItem.answerImageHeight"
            />
          </div>
        </div>
        <div
          v-else
          key="empty"
          class="global-all-faqs-desktop-search-accordion__detail-empty"
        >
          <Icon name="lucide:mouse-pointer-click" aria-hidden="true" />
          <span>{{ t('faq.ui.selectQuestion', 'Select a question to view the answer.') }}</span>
        </div>
      </Transition>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'
import type { GlobalAllFaqFlatItem } from '~/data/faq'

const props = defineProps<{
  items: GlobalAllFaqFlatItem[]
  expandedItems: ReadonlySet<string>
}>()

const emit = defineEmits<{
  'toggle-item': [itemId: string]
}>()

const { t } = useI18n()
const selectedItem = computed(() => (
  props.items.find(item => props.expandedItems.has(item.id)) || null
))

const answerId = (itemId: string) => (
  `global-faq-desktop-answer-${itemId.replace(/[^a-zA-Z0-9_-]/g, '-')}`
)
</script>

<style scoped>
.global-all-faqs-desktop-search-accordion {
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
  align-items: stretch;
  gap: 0.75rem;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  min-height: 0;
}

.global-all-faqs-desktop-search-accordion__list {
  display: flex;
  flex-direction: column;
  min-width: 0;
  max-width: 100%;
  gap: 0.55rem;
  max-height: 100%;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.global-all-faqs-desktop-search-accordion__item {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  flex: 0 0 auto;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: 0.75rem;
  background: var(--tz-card-surface, #111116);
  transition: border-color 0.2s ease, background 0.2s ease;
}

.global-all-faqs-desktop-search-accordion__item:hover,
.global-all-faqs-desktop-search-accordion__item.is-expanded {
  border-color: rgba(181, 255, 109, 0.52);
  background:
    linear-gradient(0deg, rgba(181, 255, 109, 0.045), rgba(181, 255, 109, 0.045)),
    var(--tz-card-surface, #111116);
}

.global-all-faqs-desktop-search-accordion__button {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 1rem;
  width: 100%;
  min-width: 0;
  align-items: center;
  padding: 0.85rem 1rem;
  border: 0;
  background: transparent;
  color: #ffffff;
  text-align: left;
  cursor: pointer;
}

.global-all-faqs-desktop-search-accordion__main {
  display: grid;
  min-width: 0;
  gap: 0.35rem;
}

.global-all-faqs-desktop-search-accordion__meta {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.5rem;
  color: #64748b;
  font-size: 0.57rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.global-all-faqs-desktop-search-accordion__category {
  max-width: 14rem;
  overflow: hidden;
  color: #B5FF6D;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.global-all-faqs-desktop-search-accordion__page {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.global-all-faqs-desktop-search-accordion__question {
  overflow: hidden;
  min-width: 0;
  color: #ffffff;
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.global-all-faqs-desktop-search-accordion__icon {
  width: 1.1rem;
  height: 1.1rem;
  flex: 0 0 auto;
  color: #94a3b8;
  transition: transform 0.25s ease, color 0.25s ease;
}

.global-all-faqs-desktop-search-accordion__icon.is-expanded {
  color: #B5FF6D;
  transform: rotate(45deg);
}

.global-all-faqs-desktop-search-accordion__detail {
  display: flex;
  min-width: 0;
  min-height: 0;
  max-width: 100%;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: 0.75rem;
  background: var(--tz-card-surface, #111116);
}

.global-all-faqs-desktop-search-accordion__detail-content,
.global-all-faqs-desktop-search-accordion__detail-empty {
  display: flex;
  width: 100%;
  min-width: 0;
  flex-direction: column;
  padding: 1rem 1.1rem 1.1rem;
}

.global-all-faqs-desktop-search-accordion__detail-content {
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.global-all-faqs-desktop-search-accordion__detail-meta {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.65rem;
  color: #64748b;
  font-size: 0.57rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.global-all-faqs-desktop-search-accordion__detail-question {
  margin: 0;
  color: #ffffff;
  font-size: 0.95rem;
  font-weight: 900;
  line-height: 1.45;
}

.global-all-faqs-desktop-search-accordion__detail-empty {
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  color: #64748b;
  font-size: 0.78rem;
  line-height: 1.5;
  text-align: center;
}

.global-all-faqs-desktop-search-accordion__detail-empty :deep(svg) {
  width: 1.35rem;
  height: 1.35rem;
  color: #B5FF6D;
}

.global-all-faqs-desktop-search-accordion__answer {
  min-width: 0;
  overflow-wrap: anywhere;
  margin-top: 1rem;
  padding-top: 0.85rem;
  color: #cbd5e1;
  font-size: 0.77rem;
  line-height: 1.7;
}

@media (max-width: 980px) {
  .global-all-faqs-desktop-search-accordion {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }
}
</style>
