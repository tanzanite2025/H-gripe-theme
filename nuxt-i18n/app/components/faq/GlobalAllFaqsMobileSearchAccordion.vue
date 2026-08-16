<template>
  <div class="global-all-faqs-mobile-search-accordion">
    <div
      v-for="item in items"
      :key="item.id"
      class="global-all-faqs-mobile-search-accordion__item"
      :class="{ 'is-expanded': expandedItems.has(item.id) }"
    >
      <button
        type="button"
        class="global-all-faqs-mobile-search-accordion__button"
        :class="{ 'is-expanded': expandedItems.has(item.id) }"
        :aria-expanded="expandedItems.has(item.id)"
        :aria-controls="answerId(item.id)"
        @click="emit('toggle-item', item.id)"
      >
        <span class="global-all-faqs-mobile-search-accordion__main">
          <span class="global-all-faqs-mobile-search-accordion__category">
            {{ item.category }}
          </span>
          <span class="global-all-faqs-mobile-search-accordion__question">
            {{ item.question }}
          </span>
        </span>
        <Icon
          name="lucide:chevron-down"
          class="global-all-faqs-mobile-search-accordion__icon"
          :class="{ 'is-expanded': expandedItems.has(item.id) }"
          aria-hidden="true"
        />
      </button>

      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        leave-active-class="transition-all duration-200 ease-in"
        enter-from-class="opacity-0 max-h-0"
        enter-to-class="opacity-100 max-h-[60rem]"
        leave-from-class="opacity-100 max-h-[60rem]"
        leave-to-class="opacity-0 max-h-0"
      >
        <div
          v-if="expandedItems.has(item.id)"
          :id="answerId(item.id)"
          class="global-all-faqs-mobile-search-accordion__answer-wrap"
          role="region"
          :aria-label="item.question"
        >
          <div class="global-all-faqs-mobile-search-accordion__answer">
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
</template>

<script setup lang="ts">
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'
import type { GlobalAllFaqFlatItem } from '~/data/faq'

defineProps<{
  items: GlobalAllFaqFlatItem[]
  expandedItems: ReadonlySet<string>
}>()

const emit = defineEmits<{
  'toggle-item': [itemId: string]
}>()

const answerId = (itemId: string) => (
  `global-faq-mobile-answer-${itemId.replace(/[^a-zA-Z0-9_-]/g, '-')}`
)
</script>

<style scoped>
.global-all-faqs-mobile-search-accordion {
  display: grid;
  gap: 0.45rem;
}

.global-all-faqs-mobile-search-accordion__item {
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: 0.7rem;
  background: var(--tz-card-surface, #111116);
}

.global-all-faqs-mobile-search-accordion__item.is-expanded {
  border-color: rgba(181, 255, 109, 0.48);
}

.global-all-faqs-mobile-search-accordion__button {
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: center;
  gap: 0.7rem;
  padding: 0.75rem 0.7rem;
  border: 0;
  background: transparent;
  color: #ffffff;
  text-align: left;
  cursor: pointer;
}

.global-all-faqs-mobile-search-accordion__main {
  display: grid;
  min-width: 0;
  flex: 1;
  gap: 0.28rem;
}

.global-all-faqs-mobile-search-accordion__category {
  min-width: 0;
  overflow: hidden;
  color: #B5FF6D;
  font-size: 0.56rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.global-all-faqs-mobile-search-accordion__question {
  min-width: 0;
  color: #ffffff;
  font-size: 0.76rem;
  line-height: 1.45;
}

.global-all-faqs-mobile-search-accordion__icon {
  width: 0.9rem;
  height: 0.9rem;
  flex: 0 0 auto;
  color: #94a3b8;
  transition: transform 0.25s ease, color 0.25s ease;
}

.global-all-faqs-mobile-search-accordion__icon.is-expanded {
  color: #B5FF6D;
  transform: rotate(180deg);
}

.global-all-faqs-mobile-search-accordion__answer-wrap {
  max-height: 60rem;
  overflow: hidden;
  background:
    linear-gradient(0deg, rgba(255, 255, 255, 0.025), rgba(255, 255, 255, 0.025)),
    var(--tz-card-surface, #111116);
}

.global-all-faqs-mobile-search-accordion__answer {
  padding: 0.25rem 0.7rem 0.8rem;
  color: #cbd5e1;
  font-size: 0.73rem;
  line-height: 1.65;
}
</style>
