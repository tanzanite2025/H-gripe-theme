<template>
  <div
    class="faq-category-card rounded-2xl p-3 md:p-4"
    :class="theme === 'dark' ? 'bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)]' : 'bg-white shadow-lg'"
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
      class="space-y-0 rounded-xl overflow-hidden border"
      :class="theme === 'dark' ? 'bg-slate-900/40 border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)]' : 'bg-white border-gray-200 shadow-sm'"
    >
      <div
        v-for="item in category.items"
        :key="item.id"
        class="faq-item border-b last:border-b-0"
        :class="theme === 'dark' ? 'border-slate-800/50' : 'border-gray-100'"
      >
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
            class="overflow-hidden bg-opacity-30"
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
</template>

<script setup lang="ts">
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'
import type { FaqCategory } from '~/data/faq/types'

defineProps<{
  category: FaqCategory
  theme: 'light' | 'dark'
  showCategories: boolean
  expandedItems: Set<string>
}>()

defineEmits<{
  'toggle-item': [itemId: string]
}>()
</script>

<style scoped>
.page-faq__category-header {
  margin-bottom: 0.75rem;
  padding-bottom: 0.55rem;
}

@media (min-width: 768px) {
  .page-faq__category-header {
    margin-bottom: 0.9rem;
    padding-bottom: 0.7rem;
  }
}
</style>
