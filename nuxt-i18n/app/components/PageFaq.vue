<template>
  <section
    class="page-faq w-full"
    :class="[
      theme === 'dark' ? 'bg-transparent' : 'bg-white',
      'py-4 md:py-6',
    ]"
  >
    <div class="page-faq__shell w-full max-w-none mx-auto">
      <header
        v-if="displayTitle || faqData?.subtitle"
        class="page-faq__header text-center"
      >
        <h3
          v-if="displayTitle"
          class="page-faq__title tz-faq-title"
          :class="theme === 'dark' ? 'tz-text-primary' : 'text-gray-800'"
        >
          {{ displayTitle }}
        </h3>
        <p
          v-if="faqData?.subtitle"
          class="page-faq__subtitle tz-faq-subtitle max-w-2xl mx-auto"
          :class="theme === 'dark' ? 'tz-text-secondary' : 'tz-text-muted'"
        >
          {{ faqData.subtitle }}
        </p>
      </header>

      <div v-if="faqData && displayItems.length > 0" class="page-faq__content">
        <DesktopFaqMasterDetail
          class="page-faq__desktop-list"
          :items="displayItems"
          :expanded-items="expandedItems"
          id-prefix="page-faq-answer"
          @toggle-item="toggleItem"
        />

        <div class="page-faq__mobile-list">
          <div
            v-for="item in displayItems"
            :key="item.id"
            class="page-faq__item"
            :class="{ 'is-expanded': expandedItems.has(item.id) }"
          >
            <button
              type="button"
              class="page-faq__question group"
              :aria-expanded="expandedItems.has(item.id)"
              :aria-controls="answerId(item.id)"
              @click="toggleItem(item.id)"
            >
              <span
                class="page-faq__question-text tz-faq-question"
                :class="[
                  theme === 'dark' ? 'tz-text-secondary' : 'text-gray-800',
                  expandedItems.has(item.id)
                    ? 'text-[var(--tz-text-accent)]'
                    : 'group-hover:text-[var(--tz-text-accent)]',
                ]"
              >
                {{ item.question }}
              </span>
              <Icon
                name="lucide:chevron-down"
                class="page-faq__chevron"
                :class="{ 'is-expanded': expandedItems.has(item.id) }"
                aria-hidden="true"
              />
            </button>

            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              leave-active-class="transition-all duration-150 ease-in"
              enter-from-class="opacity-0 max-h-0"
              enter-to-class="opacity-100 max-h-[60rem]"
              leave-from-class="opacity-100 max-h-[60rem]"
              leave-to-class="opacity-0 max-h-0"
            >
              <div
                v-if="expandedItems.has(item.id)"
                :id="answerId(item.id)"
                class="page-faq__answer-wrap"
                role="region"
                :aria-label="item.question"
              >
                <div class="page-faq__answer tz-faq-answer">
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

      <div
        v-else
        class="page-faq__empty text-center py-12 rounded-2xl border-2 border-dashed"
        :class="theme === 'dark'
          ? 'tz-border-subtle tz-text-muted'
          : 'border-[var(--tz-form-control-border)] tz-text-muted'"
      >
        <p class="text-sm">{{ t('faq.ui.emptySection') }}</p>
      </div>

      <div
        v-if="showViewAllLink && hasMoreItems"
        class="page-faq__footer text-center mt-8"
      >
        <NuxtLink
          :to="localePath('/support/faqs')"
          class="inline-flex items-center gap-2 px-6 py-2.5 rounded-full text-sm font-bold transition-all shadow-lg hover:-translate-y-0.5"
          :class="theme === 'dark'
            ? 'tz-surface-panel tz-text-secondary hover:tz-surface-panel hover:tz-text-primary hover:shadow-md'
            : 'bg-[var(--tz-action-primary)] text-white hover:bg-[var(--tz-action-primary-hover)] hover:shadow-[0_8px_18px_rgba(15,23,42,0.16)]'"
        >
          {{ t('faq.ui.viewAll') }}
          <Icon name="lucide:arrow-right" class="w-4 h-4" aria-hidden="true" />
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useLocalePath } from '#imports'
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'
import DesktopFaqMasterDetail from '~/components/faq/DesktopFaqMasterDetail.vue'
import { usePageFaq } from '~/composables/usePageFaq'
import type { PageFaqProps } from '../data/faq/types'

const props = withDefaults(defineProps<PageFaqProps>(), {
  theme: 'light',
  showViewAllLink: false,
})

const { t } = useI18n()
const localePath = useLocalePath()
const {
  faqData,
  displayTitle,
  displayItems,
  expandedItems,
  toggleItem,
  hasMoreItems,
} = await usePageFaq(props)

const answerId = (itemId: string) => (
  `page-faq-answer-${itemId.replace(/[^a-zA-Z0-9_-]/g, '-')}`
)
</script>

<style scoped>
.page-faq {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
  color: var(--tz-text-primary);
}

.page-faq__shell {
  width: 100%;
}

.page-faq__header {
  margin-bottom: 1.25rem;
}

.page-faq__title {
  margin: 0;
}

.page-faq__subtitle {
  margin: 0.35rem auto 0;
}

.page-faq__desktop-list {
  display: none !important;
}

.page-faq__mobile-list {
  overflow: hidden;
  border: 1px solid rgba(20, 32, 43, 0.12);
  border-radius: 1rem;
  background: var(--tz-card-surface);
  box-shadow: 0 4px 16px rgba(20, 32, 43, 0.08);
}

.page-faq__item {
  border-bottom: 1px solid rgba(20, 32, 43, 0.1);
}

.page-faq__item:last-child {
  border-bottom: 0;
}

.page-faq__question {
  display: flex;
  width: 100%;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.95rem 1rem;
  border: 0;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.page-faq__question:hover,
.page-faq__item.is-expanded .page-faq__question {
  background: rgba(20, 32, 43, 0.04);
}

.page-faq__question-text {
  min-width: 0;
  flex: 1;
  font-size: 1rem;
  line-height: 1.45;
}

.page-faq__chevron {
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
  color: var(--tz-text-muted);
  transition: transform 0.2s ease, color 0.2s ease;
}

.page-faq__chevron.is-expanded {
  color: var(--tz-text-accent);
  transform: rotate(180deg);
}

.page-faq__answer-wrap {
  max-height: 60rem;
  overflow: hidden;
  background:
    linear-gradient(0deg, rgba(20, 32, 43, 0.025), rgba(20, 32, 43, 0.025)),
    var(--tz-card-surface);
}

.page-faq__answer {
  padding: 0.25rem 1rem 1rem;
  color: var(--tz-text-secondary);
  line-height: 1.7;
}

.page-faq__empty {
  margin-top: 1rem;
}

@media (min-width: 768px) {
  .page-faq {
    padding: 2rem 0 2.25rem;
    background-color: var(--tz-card-surface);
    background-image: radial-gradient(rgba(20, 32, 43, 0.04) 1px, transparent 0);
    background-size: 24px 24px;
  }

  .page-faq__shell {
    max-width: min(100rem, calc(100vw - 5rem));
  }

  .page-faq__header {
    margin-bottom: 1.5rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid rgba(20, 32, 43, 0.12);
    text-align: left;
  }

  .page-faq__title {
    color: var(--tz-text-primary) !important;
    font-size: 1.25rem;
    font-style: italic;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .page-faq__subtitle {
    max-width: none;
    margin: 0.25rem 0 0;
    color: var(--tz-text-secondary) !important;
    font-size: 0.78rem;
    text-align: left;
  }

  .page-faq__mobile-list {
    display: none;
  }

  .page-faq__desktop-list {
    display: grid !important;
    min-height: 26rem;
  }
}
</style>
