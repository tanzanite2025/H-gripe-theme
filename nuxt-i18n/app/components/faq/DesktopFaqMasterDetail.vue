<template>
  <div class="desktop-faq-master-detail">
    <div class="desktop-faq-master-detail__list">
      <div
        v-for="item in items"
        :key="item.id"
        class="desktop-faq-master-detail__item"
        :class="{ 'is-expanded': expandedItems.has(item.id) }"
      >
        <button
          type="button"
          class="desktop-faq-master-detail__button"
          :class="{ 'is-expanded': expandedItems.has(item.id) }"
          :aria-expanded="expandedItems.has(item.id)"
          :aria-controls="answerId(item.id)"
          :title="item.question"
          @click="emit('toggle-item', item.id)"
        >
          <span class="desktop-faq-master-detail__main">
            <span
              v-if="item.category || item.pageTitle"
              class="desktop-faq-master-detail__meta"
            >
              <span
                v-if="item.category"
                class="desktop-faq-master-detail__category"
              >
                {{ item.category }}
              </span>
              <span
                v-if="item.pageTitle"
                class="desktop-faq-master-detail__page"
              >
                {{ item.pageTitle }}
              </span>
            </span>
            <span class="desktop-faq-master-detail__question">
              {{ item.question }}
            </span>
          </span>
          <Icon
            name="lucide:plus"
            class="desktop-faq-master-detail__icon"
            :class="{ 'is-expanded': expandedItems.has(item.id) }"
            aria-hidden="true"
          />
        </button>
      </div>
    </div>

    <aside class="desktop-faq-master-detail__detail">
      <Transition
        mode="out-in"
        enter-active-class="transition-all duration-300 ease-out"
        leave-active-class="transition-all duration-200 ease-in"
        enter-from-class="opacity-0 translate-y-2"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <section
          v-if="selectedItem"
          :key="selectedItem.id"
          :id="answerId(selectedItem.id)"
          class="desktop-faq-master-detail__detail-content"
          role="region"
          :aria-label="selectedItem.question"
        >
          <div
            v-if="selectedItem.category || selectedItem.pageTitle"
            class="desktop-faq-master-detail__detail-meta"
          >
            <span
              v-if="selectedItem.category"
              class="desktop-faq-master-detail__category"
            >
              {{ selectedItem.category }}
            </span>
            <span
              v-if="selectedItem.pageTitle"
              class="desktop-faq-master-detail__page"
            >
              {{ selectedItem.pageTitle }}
            </span>
          </div>
          <h4 class="desktop-faq-master-detail__detail-question">
            {{ selectedItem.question }}
          </h4>
          <div class="desktop-faq-master-detail__answer">
            <FaqAnswerContent
              :answer="selectedItem.answer"
              :image-url="selectedItem.answerImageUrl"
              :image-alt="selectedItem.answerImageAlt"
              :image-width="selectedItem.answerImageWidth"
              :image-height="selectedItem.answerImageHeight"
            />
          </div>
        </section>
        <div
          v-else
          key="empty"
          class="desktop-faq-master-detail__detail-empty"
        >
          <Icon name="lucide:mouse-pointer-click" aria-hidden="true" />
          <span>{{ emptyMessage || t('faq.ui.selectQuestion', 'Select a question to view the answer.') }}</span>
        </div>
      </Transition>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import FaqAnswerContent from '~/components/FaqAnswerContent.vue'

interface DesktopFaqMasterDetailItem {
  id: string
  category?: string
  pageTitle?: string
  question: string
  answer: string
  answerImageUrl?: string
  answerImageAlt?: string
  answerImageWidth?: number
  answerImageHeight?: number
}

const props = withDefaults(defineProps<{
  items: DesktopFaqMasterDetailItem[]
  expandedItems: ReadonlySet<string>
  idPrefix?: string
  emptyMessage?: string
}>(), {
  idPrefix: 'desktop-faq-answer',
  emptyMessage: '',
})

const emit = defineEmits<{
  'toggle-item': [itemId: string]
}>()

const { t } = useI18n()

const selectedItem = computed(() => (
  props.items.find(item => props.expandedItems.has(item.id)) || null
))

const answerId = (itemId: string) => (
  `${props.idPrefix}-${itemId.replace(/[^a-zA-Z0-9_-]/g, '-')}`
)
</script>

<style scoped>
.desktop-faq-master-detail {
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
  align-items: stretch;
  gap: 0.75rem;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  min-height: 0;
  box-sizing: border-box;
}

.desktop-faq-master-detail__list {
  display: flex;
  min-width: 0;
  max-width: 100%;
  max-height: 100%;
  flex-direction: column;
  gap: 0.55rem;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.desktop-faq-master-detail__item {
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

.desktop-faq-master-detail__item:hover,
.desktop-faq-master-detail__item.is-expanded {
  border-color: rgba(181, 255, 109, 0.52);
  background:
    linear-gradient(0deg, rgba(181, 255, 109, 0.045), rgba(181, 255, 109, 0.045)),
    var(--tz-card-surface, #111116);
}

.desktop-faq-master-detail__button {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 1rem;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  align-items: center;
  box-sizing: border-box;
  padding: 0.85rem 1rem;
  border: 0;
  background: transparent;
  color: #ffffff;
  text-align: left;
  cursor: pointer;
}

.desktop-faq-master-detail__button:focus-visible {
  outline: 2px solid #B5FF6D;
  outline-offset: -2px;
}

.desktop-faq-master-detail__main {
  display: grid;
  min-width: 0;
  max-width: 100%;
  gap: 0.35rem;
}

.desktop-faq-master-detail__meta,
.desktop-faq-master-detail__detail-meta {
  display: flex;
  min-width: 0;
  max-width: 100%;
  align-items: center;
  gap: 0.5rem;
  overflow: hidden;
  color: #64748b;
  font-size: 0.57rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.desktop-faq-master-detail__category {
  max-width: 14rem;
  overflow: hidden;
  color: #B5FF6D;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.desktop-faq-master-detail__page {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.desktop-faq-master-detail__question {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #ffffff;
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1.45;
}

.desktop-faq-master-detail__icon {
  width: 1.1rem;
  height: 1.1rem;
  flex: 0 0 auto;
  color: #94a3b8;
  transition: transform 0.25s ease, color 0.25s ease;
}

.desktop-faq-master-detail__icon.is-expanded {
  color: #B5FF6D;
  transform: rotate(45deg);
}

.desktop-faq-master-detail__detail {
  display: flex;
  min-width: 0;
  min-height: 14rem;
  max-width: 100%;
  box-sizing: border-box;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: 0.75rem;
  background: var(--tz-card-surface, #111116);
}

.desktop-faq-master-detail__detail-content,
.desktop-faq-master-detail__detail-empty {
  display: flex;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  flex-direction: column;
  box-sizing: border-box;
  padding: 1rem 1.1rem 1.1rem;
}

.desktop-faq-master-detail__detail-content {
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.desktop-faq-master-detail__detail-meta {
  margin-bottom: 0.65rem;
}

.desktop-faq-master-detail__detail-question {
  margin: 0;
  color: #ffffff;
  font-size: 0.95rem;
  font-weight: 900;
  line-height: 1.45;
}

.desktop-faq-master-detail__detail-empty {
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  color: #64748b;
  font-size: 0.78rem;
  line-height: 1.5;
  text-align: center;
}

.desktop-faq-master-detail__detail-empty :deep(svg) {
  width: 1.35rem;
  height: 1.35rem;
  color: #B5FF6D;
}

.desktop-faq-master-detail__answer {
  min-width: 0;
  margin-top: 1rem;
  overflow-wrap: anywhere;
  color: #cbd5e1;
  font-size: 0.77rem;
  line-height: 1.7;
}

@media (max-width: 980px) {
  .desktop-faq-master-detail {
    grid-template-columns: minmax(0, 1fr);
  }

  .desktop-faq-master-detail__category {
    max-width: min(14rem, 48vw);
  }
}
</style>
