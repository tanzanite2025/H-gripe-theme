<template>
  <div
    class="faq-category-card rounded-2xl p-3 md:p-4"
    :class="theme === 'dark' ? 'bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)]' : 'bg-white shadow-lg'"
  >
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

    <div
      class="faq-list faq-list--mobile space-y-0 rounded-xl overflow-hidden border"
      :class="theme === 'dark' ? 'bg-slate-900/40 border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)]' : 'bg-white border-gray-200 shadow-sm'"
    >
      <div
        v-for="item in category.items"
        :key="item.id"
        class="faq-item border-b last:border-b-0"
        :class="[
          theme === 'dark' ? 'border-slate-800/50' : 'border-gray-100',
          expandedItems.has(item.id) ? 'faq-item--expanded' : ''
        ]"
      >
        <button
          type="button"
          class="faq-item__button w-full flex items-center justify-between gap-3 px-3 py-3 text-left transition-colors group"
          :class="[
            theme === 'dark'
              ? 'hover:bg-white/5'
              : 'hover:bg-gray-50',
            expandedItems.has(item.id)
              ? (theme === 'dark' ? 'bg-white/5' : 'bg-gray-50')
              : ''
          ]"
          @click="$emit('toggle-item', item.id)"
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
            class="faq-item__icon flex-shrink-0 w-6 h-6 flex items-center justify-center rounded-full transition-all duration-200"
            :class="[
              expandedItems.has(item.id)
                ? (theme === 'dark' ? 'bg-sky-500/10 text-sky-400 rotate-180' : 'bg-blue-100 text-blue-600 rotate-180')
                : (theme === 'dark' ? 'tz-text-muted bg-transparent' : 'text-gray-500 bg-transparent')
            ]"
          >
            <svg class="faq-item__chevron w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
            <span class="faq-item__plus" aria-hidden="true">+</span>
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
            class="faq-item__answer-wrap overflow-hidden bg-opacity-30"
            :class="theme === 'dark' ? 'bg-slate-950/30' : 'bg-gray-50/50'"
          >
            <div
              class="page-faq__answer tz-faq-answer px-4 pb-4 pt-1"
              :class="theme === 'dark' ? 'tz-text-secondary' : 'text-gray-600'"
            >
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

    <div
      class="faq-list faq-list--desktop"
      :class="theme === 'dark' ? 'bg-slate-900/40 border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)]' : 'bg-white border-gray-200 shadow-sm'"
    >
      <div
        v-for="(columnItems, columnIndex) in faqColumns"
        :key="columnIndex"
        class="faq-list__column"
      >
        <div
          v-for="item in columnItems"
          :key="item.id"
          class="faq-item border-b last:border-b-0"
          :class="[
            theme === 'dark' ? 'border-slate-800/50' : 'border-gray-100',
            expandedItems.has(item.id) ? 'faq-item--expanded' : ''
          ]"
        >
          <button
            type="button"
            class="faq-item__button w-full flex items-center justify-between gap-3 px-3 py-3 text-left transition-colors group"
            :class="[
              theme === 'dark'
                ? 'hover:bg-white/5'
                : 'hover:bg-gray-50',
              expandedItems.has(item.id)
                ? (theme === 'dark' ? 'bg-white/5' : 'bg-gray-50')
                : ''
            ]"
            @click="$emit('toggle-item', item.id)"
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
              class="faq-item__icon flex-shrink-0 w-6 h-6 flex items-center justify-center rounded-full transition-all duration-200"
              :class="[
                expandedItems.has(item.id)
                  ? (theme === 'dark' ? 'bg-sky-500/10 text-sky-400 rotate-180' : 'bg-blue-100 text-blue-600 rotate-180')
                  : (theme === 'dark' ? 'tz-text-muted bg-transparent' : 'text-gray-500 bg-transparent')
              ]"
            >
              <svg class="faq-item__chevron w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
              <span class="faq-item__plus" aria-hidden="true">+</span>
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
              class="faq-item__answer-wrap overflow-hidden bg-opacity-30"
              :class="theme === 'dark' ? 'bg-slate-950/30' : 'bg-gray-50/50'"
            >
              <div
                class="page-faq__answer tz-faq-answer px-4 pb-4 pt-1"
                :class="theme === 'dark' ? 'tz-text-secondary' : 'text-gray-600'"
              >
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
import { computed } from 'vue'
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'
import type { FaqCategory, FaqItem } from '~/data/faq/types'

const props = defineProps<{
  category: FaqCategory
  theme: 'light' | 'dark'
  showCategories: boolean
  expandedItems: Set<string>
}>()

defineEmits<{
  'toggle-item': [itemId: string]
}>()

const faqColumns = computed<[FaqItem[], FaqItem[]]>(() => {
  const columns: [FaqItem[], FaqItem[]] = [[], []]

  props.category.items.forEach((item, index) => {
    columns[index % 2]?.push(item)
  })

  return columns
})
</script>

<style scoped>
.faq-item__plus {
  display: none;
}

.faq-list--desktop {
  display: none;
}

.page-faq__category-header {
  margin-bottom: 0.75rem;
  padding-bottom: 0.55rem;
}

@media (min-width: 768px) {
  .faq-category-card {
    padding: 0 !important;
    border-radius: 0;
    background: transparent !important;
    box-shadow: none !important;
  }

  .page-faq__category-header {
    display: none;
  }

  .faq-list--mobile {
    display: none;
  }

  .faq-list--desktop {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.625rem;
    align-items: start;
    overflow: visible !important;
    border: 0 !important;
    border-radius: 0;
    background: transparent !important;
    box-shadow: none !important;
  }

  .faq-list__column {
    display: grid;
    gap: 0.625rem;
    align-content: start;
  }

  .faq-item {
    position: relative;
    align-self: start;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.1) !important;
    border-radius: 1rem;
    background: #000000;
  }

  .faq-item::before {
    position: absolute;
    inset: 0 auto 0 0;
    width: 0;
    content: '';
    background: #B5FF6D;
    box-shadow: 0 0 10px #B5FF6D;
    transition: width 0.3s ease;
  }

  .faq-item--expanded {
    border-color: rgba(255, 255, 255, 0.3) !important;
  }

  .faq-item--expanded::before {
    width: 3px;
  }

  .faq-item__button {
    position: relative;
    padding: 1rem !important;
    background: transparent !important;
  }

  .faq-item__button:hover {
    background: rgba(255, 255, 255, 0.025) !important;
  }

  .page-faq__question-text {
    color: #ffffff !important;
    font-size: 0.82rem;
    font-weight: 800;
    line-height: 1.45;
  }

  .faq-item__icon {
    width: 1.35rem !important;
    height: 1.35rem !important;
    border-radius: 0;
    background: transparent !important;
    color: #94a3b8 !important;
    transform: none !important;
  }

  .faq-item--expanded .faq-item__icon {
    color: #B5FF6D !important;
    transform: rotate(45deg) !important;
  }

  .faq-item__chevron {
    display: none;
  }

  .faq-item__plus {
    display: block;
    font-size: 1.05rem;
    font-weight: 900;
    line-height: 1;
  }

  .faq-item__answer-wrap {
    background: transparent !important;
  }

  .page-faq__answer {
    margin: 0 1rem 1rem;
    padding: 0.8rem 0 0 !important;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    color: #cbd5e1 !important;
    font-size: 0.78rem;
    line-height: 1.7;
  }
}
</style>
