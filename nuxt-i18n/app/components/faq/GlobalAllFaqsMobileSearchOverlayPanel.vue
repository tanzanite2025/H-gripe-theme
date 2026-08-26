<template>
  <section class="global-all-faqs-mobile-search-overlay-panel">
    <header class="global-all-faqs-mobile-search-overlay-panel__header">
      <div>
        <div class="global-all-faqs-mobile-search-overlay-panel__eyebrow">
          {{ t('faq.ui.categorizedSupport') }}
        </div>
        <h2 class="global-all-faqs-mobile-search-overlay-panel__title">
          {{ t('faq.title') }}
        </h2>
      </div>
      <button
        type="button"
        class="global-all-faqs-mobile-search-overlay-panel__close"
        :aria-label="t('common.close', 'Close')"
        @click="emit('close')"
      >
        <Icon name="lucide:x" aria-hidden="true" />
      </button>
    </header>

    <div class="global-all-faqs-mobile-search-overlay-panel__search">
      <Icon name="lucide:search" aria-hidden="true" />
      <input
        :value="searchQuery"
        type="search"
        :placeholder="t('faq.ui.searchPlaceholder')"
        :aria-label="t('faq.ui.searchPlaceholder')"
        @input="handleSearchInput"
      >
      <button
        v-if="searchQuery"
        type="button"
        class="global-all-faqs-mobile-search-overlay-panel__clear"
        :aria-label="t('filter.clearSearch', 'Clear search')"
        @click="emit('update:search-query', '')"
      >
        <Icon name="lucide:x" aria-hidden="true" />
      </button>
    </div>

    <div class="global-all-faqs-mobile-search-overlay-panel__body">
      <div
        v-if="pending"
        class="global-all-faqs-mobile-search-overlay-panel__loading"
      >
        {{ t('faq.ui.loadingAnswers', 'Loading FAQs...') }}
      </div>

      <template v-else-if="hasSearchQuery">
        <div class="global-all-faqs-mobile-search-overlay-panel__search-results">
          <div class="global-all-faqs-mobile-search-overlay-panel__section-heading">
            <div>
              <span class="global-all-faqs-mobile-search-overlay-panel__section-kicker">
                {{ t('faq.ui.searchPlaceholder') }}
              </span>
              <h3>{{ searchQuery }}</h3>
            </div>
            <span class="global-all-faqs-mobile-search-overlay-panel__result-count">
              {{ searchResultCount }}
            </span>
          </div>
          <GlobalAllFaqsMobileSearchAccordion
            v-if="searchResults.length > 0"
            :items="searchResults"
            :expanded-items="expandedItems"
            @toggle-item="emit('toggle-item', $event)"
          />
          <div
            v-else
            class="global-all-faqs-mobile-search-overlay-panel__empty"
          >
            {{ t('faq.ui.noResults', { query: searchQuery }) }}
          </div>
        </div>
      </template>

      <template v-else>
        <div class="global-all-faqs-mobile-search-overlay-panel__questions-section">
          <div class="global-all-faqs-mobile-search-overlay-panel__section-heading">
              <div>
                <span class="global-all-faqs-mobile-search-overlay-panel__section-kicker">
                  {{ t('faq.ui.quickAnswers') }}
                </span>
                <h3>{{ activeTopic?.label || t('faq.ui.quickAnswers') }}</h3>
              </div>
              <button
                v-if="activeTopic"
                type="button"
                class="global-all-faqs-mobile-search-overlay-panel__topic-reset"
                :aria-label="t('filter.clearSearch', 'Clear search')"
                @click="emit('select-topic', activeTopic.id)"
              >
                <Icon name="lucide:x" aria-hidden="true" />
              </button>
            </div>
            <GlobalAllFaqsMobileSearchAccordion
              :items="topicItems"
              :expanded-items="expandedItems"
              @toggle-item="emit('toggle-item', $event)"
            />
          </div>

        <div class="global-all-faqs-mobile-search-overlay-panel__topics-section">
          <div class="global-all-faqs-mobile-search-overlay-panel__section-heading">
            <div>
              <span class="global-all-faqs-mobile-search-overlay-panel__section-kicker">
                {{ t('faq.ui.categoriesLabel') }}
              </span>
              <h3>{{ t('faq.ui.categoriesLabel') }}</h3>
            </div>
          </div>
          <div class="global-all-faqs-mobile-search-overlay-panel__topics">
            <button
              v-for="topic in featuredTopics"
              :key="topic.id"
              type="button"
              class="global-all-faqs-mobile-search-overlay-panel__topic"
              :class="{ 'is-active': activeTopicId === topic.id }"
              @click="emit('select-topic', topic.id)"
            >
              <span class="global-all-faqs-mobile-search-overlay-panel__topic-label">
                {{ topic.label }}
              </span>
              <span class="global-all-faqs-mobile-search-overlay-panel__topic-count">
                {{ topic.count }} FAQ
              </span>
            </button>
          </div>
        </div>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import GlobalAllFaqsMobileSearchAccordion from '~/components/faq/GlobalAllFaqsMobileSearchAccordion.vue'
import type {
  GlobalAllFaqFlatItem,
  GlobalAllFaqSearchTopic,
} from '~/data/faq'

const props = defineProps<{
  pending: boolean
  searchQuery: string
  featuredTopics: GlobalAllFaqSearchTopic[]
  activeTopicId: string
  activeTopic: GlobalAllFaqSearchTopic | null
  topicItems: GlobalAllFaqFlatItem[]
  searchResults: GlobalAllFaqFlatItem[]
  searchResultCount: number
  expandedItems: ReadonlySet<string>
}>()

const emit = defineEmits<{
  'update:search-query': [value: string]
  'select-topic': [topicId: string]
  'toggle-item': [itemId: string]
  close: []
}>()

const { t } = useI18n()
const hasSearchQuery = computed(() => props.searchQuery.trim().length > 0)

const handleSearchInput = (event: Event) => {
  emit('update:search-query', (event.target as HTMLInputElement).value)
}
</script>

<style scoped>
.global-all-faqs-mobile-search-overlay-panel {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-height: calc(100vh - 1rem);
  overflow: hidden;
  border: 0;
  border-radius: 0.8rem;
  background: var(--tz-card-surface, #ffffff);
  box-shadow: 0 20px 54px rgba(20, 32, 43, 0.18);
}

.global-all-faqs-mobile-search-overlay-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.7rem;
  padding: 0.8rem 0.75rem 0.65rem;
  border-bottom: 1px solid rgba(20, 32, 43, 0.12);
    background: var(--tz-card-surface);
}

.global-all-faqs-mobile-search-overlay-panel__eyebrow {
  margin-bottom: 0.2rem;
  color: var(--tz-text-accent);
  font-size: 0.51rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.global-all-faqs-mobile-search-overlay-panel__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.92rem;
  font-weight: 900;
  line-height: 1.2;
}

.global-all-faqs-mobile-search-overlay-panel__close,
.global-all-faqs-mobile-search-overlay-panel__topic-reset {
  display: inline-grid;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgba(20, 32, 43, 0.14);
  background: #f3f6f8;
  color: var(--tz-text-secondary);
  cursor: pointer;
}

.global-all-faqs-mobile-search-overlay-panel__close {
  width: 2rem;
  height: 2rem;
  border-radius: 0.55rem;
}

.global-all-faqs-mobile-search-overlay-panel__topic-reset {
  width: 1.55rem;
  height: 1.55rem;
  border-radius: 999px;
}

.global-all-faqs-mobile-search-overlay-panel__close :deep(svg) {
  width: 0.88rem;
  height: 0.88rem;
}

.global-all-faqs-mobile-search-overlay-panel__topic-reset :deep(svg) {
  width: 0.72rem;
  height: 0.72rem;
}

.global-all-faqs-mobile-search-overlay-panel__search {
  display: flex;
  min-height: 2.45rem;
  align-items: center;
  gap: 0.48rem;
  margin: 0.6rem 0.65rem 0.55rem;
  padding: 0 0.65rem;
  border: 1px solid rgba(20, 32, 43, 0.18);
  border-radius: 0.65rem;
  background: var(--tz-input-surface, #ffffff);
  color: var(--tz-text-secondary);
}

.global-all-faqs-mobile-search-overlay-panel__search > :deep(svg) {
  width: 0.86rem;
  height: 0.86rem;
  flex: 0 0 auto;
}

.global-all-faqs-mobile-search-overlay-panel__search input {
  min-width: 0;
  flex: 1;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--tz-text-primary);
  font-size: 0.76rem;
}

.global-all-faqs-mobile-search-overlay-panel__search input::placeholder {
  color: var(--tz-text-muted);
}

.global-all-faqs-mobile-search-overlay-panel__clear {
  display: inline-grid;
  width: 1.4rem;
  height: 1.4rem;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 999px;
  background: #e4eaee;
  color: var(--tz-text-secondary);
  cursor: pointer;
}

.global-all-faqs-mobile-search-overlay-panel__clear :deep(svg) {
  width: 0.7rem;
  height: 0.7rem;
}

.global-all-faqs-mobile-search-overlay-panel__body {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  padding: 0 0.65rem 0.75rem;
}

.global-all-faqs-mobile-search-overlay-panel__search-results,
.global-all-faqs-mobile-search-overlay-panel__questions-section {
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.global-all-faqs-mobile-search-overlay-panel__search-results {
  flex: 1 1 auto;
  padding-bottom: 0.15rem;
}

.global-all-faqs-mobile-search-overlay-panel__questions-section {
  flex: 1 1 auto;
  padding-bottom: 0.2rem;
}

.global-all-faqs-mobile-search-overlay-panel__topics-section {
  flex: 0 0 11.75rem;
  min-height: 11.75rem;
  overflow: hidden;
  margin-top: auto;
  padding-top: 0.55rem;
  padding-bottom: 0.35rem;
}

.global-all-faqs-mobile-search-overlay-panel__section-heading {
  display: flex;
  min-height: 1.9rem;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
  margin-bottom: 0.5rem;
  padding-bottom: 0.4rem;
}

.global-all-faqs-mobile-search-overlay-panel__section-kicker {
  display: block;
  margin-bottom: 0.16rem;
  color: var(--tz-text-muted);
  font-size: 0.5rem;
  font-weight: 900;
  letter-spacing: 0.11em;
  line-height: 1.2;
  text-transform: uppercase;
}

.global-all-faqs-mobile-search-overlay-panel__section-heading h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  font-weight: 900;
  line-height: 1.25;
}

.global-all-faqs-mobile-search-overlay-panel__result-count {
  display: inline-grid;
  min-width: 1.55rem;
  height: 1.55rem;
  place-items: center;
  border: 1px solid rgba(20, 32, 43, 0.16);
  border-radius: 999px;
  color: var(--tz-text-accent);
  font-size: 0.57rem;
  font-weight: 900;
}

.global-all-faqs-mobile-search-overlay-panel__topics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.45rem;
}

.global-all-faqs-mobile-search-overlay-panel__topic {
  display: grid;
  min-width: 0;
  height: 3.7rem;
  align-content: space-between;
  gap: 0.55rem;
  padding: 0.65rem;
  border: 1px solid transparent;
  border-radius: 0.65rem;
  background: var(--tz-card-surface, #ffffff);
  color: var(--tz-text-primary);
  text-align: left;
  cursor: pointer;
}

.global-all-faqs-mobile-search-overlay-panel__topic.is-active {
  border-color: rgba(5, 150, 105, 0.48);
  background:
    linear-gradient(0deg, rgba(5, 150, 105, 0.075), rgba(5, 150, 105, 0.075)),
    var(--tz-card-surface, #ffffff);
}

.global-all-faqs-mobile-search-overlay-panel__topic-label {
  min-width: 0;
  overflow: hidden;
  font-size: 0.67rem;
  font-weight: 800;
  line-height: 1.35;
  text-overflow: ellipsis;
}

.global-all-faqs-mobile-search-overlay-panel__topic-count {
  color: var(--tz-text-muted);
  font-size: 0.54rem;
  font-weight: 800;
  letter-spacing: 0.07em;
  text-transform: uppercase;
}

.global-all-faqs-mobile-search-overlay-panel__empty,
.global-all-faqs-mobile-search-overlay-panel__loading {
  display: grid;
  min-height: 9rem;
  place-items: center;
  border: 1px dashed rgba(20, 32, 43, 0.18);
  border-radius: 0.7rem;
  color: var(--tz-text-secondary);
  font-size: 0.74rem;
  line-height: 1.5;
  text-align: center;
}

.global-all-faqs-mobile-search-overlay-panel__loading {
  color: var(--tz-text-accent);
}

@media (min-width: 768px) {
  .global-all-faqs-mobile-search-overlay-panel {
    display: none;
  }
}
</style>
