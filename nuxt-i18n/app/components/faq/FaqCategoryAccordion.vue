<template>
  <div
    class="faq-category-card rounded-2xl p-3 md:p-4"
    :class="theme === 'dark' ? 'bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)]' : 'bg-white shadow-lg'"
  >
    <div
      v-if="showCategories"
      class="page-faq__category-header flex items-center justify-center gap-3 border-b"
      :class="theme === 'dark' ? 'border-white/10' : 'border-gray-100'"
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
      :class="theme === 'dark' ? 'bg-[var(--tz-card-surface)] border-white/10 shadow-[0_4px_16px_rgba(0,0,0,0.5)]' : 'bg-white border-gray-200 shadow-sm'"
    >
      <div
        v-for="item in category.items"
        :key="item.id"
        class="faq-item border-b last:border-b-0"
        :class="[
          theme === 'dark' ? 'border-white/10' : 'border-gray-100',
          expandedItems.has(item.id) ? 'faq-item--expanded' : '',
        ]"
      >
        <button
          type="button"
          class="faq-item__button w-full flex items-center justify-between gap-3 px-3 py-3 text-left transition-colors group"
          :class="[
            theme === 'dark' ? 'hover:bg-white/5' : 'hover:bg-gray-50',
            expandedItems.has(item.id)
              ? (theme === 'dark' ? 'bg-white/5' : 'bg-gray-50')
              : '',
          ]"
          @click="emit('toggle-item', item.id)"
        >
          <span
            class="page-faq__question-text tz-faq-question flex-1 transition-colors"
            :class="[
              theme === 'dark' ? 'tz-text-secondary' : 'text-gray-800',
              expandedItems.has(item.id)
                ? 'text-[var(--tz-text-accent)]'
                : 'group-hover:text-[var(--tz-text-accent)]',
            ]"
          >
            {{ item.question }}
          </span>
          <span
            class="faq-item__icon flex-shrink-0 w-6 h-6 flex items-center justify-center rounded-full transition-all duration-200"
            :class="[
              expandedItems.has(item.id)
                ? 'bg-[rgba(181,255,109,0.12)] text-[var(--tz-text-accent)] rotate-180'
                : (theme === 'dark' ? 'tz-text-muted bg-transparent' : 'text-gray-500 bg-transparent'),
            ]"
          >
            <svg class="faq-item__chevron w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
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
            class="faq-item__answer-wrap overflow-hidden bg-opacity-30"
            :class="theme === 'dark' ? 'faq-item__answer-wrap--dark' : 'bg-gray-50/50'"
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

    <DesktopFaqMasterDetail
      class="faq-list faq-list--desktop"
      :items="desktopItems"
      :expanded-items="expandedItems"
      :id-prefix="`page-faq-${category.id}`"
      @toggle-item="emit('toggle-item', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'
import DesktopFaqMasterDetail from '~/components/faq/DesktopFaqMasterDetail.vue'
import type { FaqCategory } from '~/data/faq/types'

const props = defineProps<{
  category: FaqCategory
  pageTitle?: string
  theme: 'light' | 'dark'
  showCategories: boolean
  expandedItems: ReadonlySet<string>
}>()

const emit = defineEmits<{
  'toggle-item': [itemId: string]
}>()

const desktopItems = computed(() => (
  props.category.items.map(item => ({
    ...item,
    category: props.category.name,
    pageTitle: props.pageTitle,
  }))
))
</script>

<style scoped>
.faq-list--desktop {
  display: none !important;
}

.faq-category-card {
  padding: 1rem !important;
  border-radius: 1.25rem !important;
}

.page-faq__category-header {
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
}

.page-faq__category-title {
  font-size: 1rem;
  font-weight: 900;
  line-height: 1.2;
}

.faq-list--mobile {
  border-radius: 1rem !important;
}

.faq-list--mobile .faq-item__button {
  padding: 0.9rem 1rem !important;
}

.faq-list--mobile .page-faq__question-text {
  font-size: 1rem !important;
  line-height: 1.45;
}

.faq-list--mobile .faq-item__chevron {
  width: 1.1rem;
  height: 1.1rem;
}

@media (min-width: 768px) {
  .faq-category-card {
    padding: 0 !important;
    border-radius: 0;
    background: transparent !important;
    box-shadow: none !important;
  }

  .page-faq__category-header,
  .faq-list--mobile {
    display: none;
  }

  .faq-list--desktop {
    display: grid !important;
    min-height: 24rem;
  }
}

@media (max-width: 767.98px) {
  .faq-item__answer-wrap--dark,
  .faq-list--mobile .faq-item__answer-wrap {
    background:
      linear-gradient(0deg, rgba(255, 255, 255, 0.025), rgba(255, 255, 255, 0.025)),
      var(--tz-card-surface, #111116) !important;
  }
}
</style>
