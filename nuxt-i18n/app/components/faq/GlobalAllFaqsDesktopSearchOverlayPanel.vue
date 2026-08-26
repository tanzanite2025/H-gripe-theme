<template>
  <section class="global-all-faqs-desktop-search-overlay-panel">
    <header class="global-all-faqs-desktop-search-overlay-panel__header">
      <div>
        <div class="global-all-faqs-desktop-search-overlay-panel__eyebrow">
          {{ t('faq.ui.categorizedSupport') }}
        </div>
        <h2 class="global-all-faqs-desktop-search-overlay-panel__title">
          {{ t('faq.title') }}
        </h2>
        <p class="global-all-faqs-desktop-search-overlay-panel__intro">
          {{ t('faq.ui.pageIntro') }}
        </p>
      </div>
      <button
        type="button"
        class="global-all-faqs-desktop-search-overlay-panel__close"
        :aria-label="t('common.close', 'Close')"
        @click="emit('close')"
      >
        <Icon name="lucide:x" aria-hidden="true" />
      </button>
    </header>

    <div class="global-all-faqs-desktop-search-overlay-panel__search">
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
        class="global-all-faqs-desktop-search-overlay-panel__clear"
        :aria-label="t('filter.clearSearch', 'Clear search')"
        @click="emit('update:search-query', '')"
      >
        <Icon name="lucide:x" aria-hidden="true" />
      </button>
    </div>

    <div class="global-all-faqs-desktop-search-overlay-panel__body">
      <div
        v-if="pending"
        class="global-all-faqs-desktop-search-overlay-panel__loading"
      >
        {{ t('faq.ui.loadingAnswers', 'Loading FAQs...') }}
      </div>

      <template v-else-if="hasSearchQuery">
        <div class="global-all-faqs-desktop-search-overlay-panel__search-results">
          <div class="global-all-faqs-desktop-search-overlay-panel__section-heading">
            <div>
              <span class="global-all-faqs-desktop-search-overlay-panel__section-kicker">
                {{ t('faq.ui.searchPlaceholder') }}
              </span>
              <h3>{{ searchQuery }}</h3>
            </div>
            <span class="global-all-faqs-desktop-search-overlay-panel__result-count">
              {{ searchResultCount }}
            </span>
          </div>
          <GlobalAllFaqsDesktopSearchAccordion
            v-if="searchResults.length > 0"
            :items="searchResults"
            :expanded-items="expandedItems"
            @toggle-item="emit('toggle-item', $event)"
          />
          <div
            v-else
            class="global-all-faqs-desktop-search-overlay-panel__empty"
          >
            {{ t('faq.ui.noResults', { query: searchQuery }) }}
          </div>
        </div>
      </template>

      <template v-else>
        <div class="global-all-faqs-desktop-search-overlay-panel__questions-section">
          <div class="global-all-faqs-desktop-search-overlay-panel__section-heading">
              <div>
                <h3>{{ activeTopic?.label || t('faq.ui.quickAnswers') }}</h3>
              </div>
              <button
                v-if="activeTopic"
                type="button"
                class="global-all-faqs-desktop-search-overlay-panel__topic-reset"
                :aria-label="t('filter.clearSearch', 'Clear search')"
                @click="emit('select-topic', activeTopic.id)"
              >
                <Icon name="lucide:x" aria-hidden="true" />
              </button>
            </div>
            <GlobalAllFaqsDesktopSearchAccordion
              :items="topicItems"
              :expanded-items="expandedItems"
              @toggle-item="emit('toggle-item', $event)"
            />
          </div>

        <div class="global-all-faqs-desktop-search-overlay-panel__topics-section">
          <div class="global-all-faqs-desktop-search-overlay-panel__section-heading">
            <div>
              <span class="global-all-faqs-desktop-search-overlay-panel__section-kicker">
                {{ t('faq.ui.categoriesLabel') }}
              </span>
              <h3>{{ t('faq.ui.categoriesLabel') }}</h3>
            </div>
          </div>
          <div class="global-all-faqs-desktop-search-overlay-panel__topics">
            <button
              v-for="topic in featuredTopics"
              :key="topic.id"
              type="button"
              class="global-all-faqs-desktop-search-overlay-panel__topic"
              :class="{ 'is-active': activeTopicId === topic.id }"
              @click="emit('select-topic', topic.id)"
            >
              <span class="global-all-faqs-desktop-search-overlay-panel__topic-label">
                {{ topic.label }}
              </span>
              <span class="global-all-faqs-desktop-search-overlay-panel__topic-footer">
                <span>{{ topic.count }} FAQ</span>
                <Icon name="lucide:arrow-up-right" aria-hidden="true" />
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
import GlobalAllFaqsDesktopSearchAccordion from '~/components/faq/GlobalAllFaqsDesktopSearchAccordion.vue'
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
.global-all-faqs-desktop-search-overlay-panel {
  display: none;
  flex-direction: column;
  width: 90vw;
  max-width: 90vw;
  height: min(82vh, 64rem);
  max-height: calc(100vh - 2rem);
  overflow: hidden;
  border: 0;
  border-radius: 0.9rem;
  background: var(--tz-card-surface, #ffffff);
  box-shadow: 0 24px 70px rgba(20, 32, 43, 0.18);
}

.global-all-faqs-desktop-search-overlay-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.2rem 0.85rem;
  border-bottom: 1px solid rgba(20, 32, 43, 0.12);
    background: var(--tz-card-surface);
}

.global-all-faqs-desktop-search-overlay-panel__eyebrow {
  margin-bottom: 0.3rem;
  color: var(--tz-text-accent);
  font-size: 0.6rem;
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.global-all-faqs-desktop-search-overlay-panel__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1.15rem;
  font-weight: 900;
  line-height: 1.2;
}

.global-all-faqs-desktop-search-overlay-panel__intro {
  max-width: 48rem;
  margin: 0.35rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  line-height: 1.45;
}

.global-all-faqs-desktop-search-overlay-panel__close,
.global-all-faqs-desktop-search-overlay-panel__topic-reset {
  display: inline-grid;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgba(20, 32, 43, 0.14);
  background: #f3f6f8;
  color: var(--tz-text-secondary);
  cursor: pointer;
}

.global-all-faqs-desktop-search-overlay-panel__close {
  width: 2.2rem;
  height: 2.2rem;
  border-radius: 0.6rem;
}

.global-all-faqs-desktop-search-overlay-panel__topic-reset {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 999px;
}

.global-all-faqs-desktop-search-overlay-panel__close:hover,
.global-all-faqs-desktop-search-overlay-panel__close:focus-visible,
.global-all-faqs-desktop-search-overlay-panel__topic-reset:hover,
.global-all-faqs-desktop-search-overlay-panel__topic-reset:focus-visible {
  border-color: rgba(20, 32, 43, 0.3);
  color: var(--tz-text-primary);
}

.global-all-faqs-desktop-search-overlay-panel__close :deep(svg) {
  width: 1rem;
  height: 1rem;
}

.global-all-faqs-desktop-search-overlay-panel__topic-reset :deep(svg) {
  width: 0.8rem;
  height: 0.8rem;
}

.global-all-faqs-desktop-search-overlay-panel__search {
  display: flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.6rem;
  margin: 0.8rem 1.2rem 0.75rem;
  padding: 0 0.8rem;
  border: 1px solid rgba(20, 32, 43, 0.18);
  border-radius: 0.7rem;
  background: var(--tz-input-surface, #ffffff);
  color: var(--tz-text-secondary);
}

.global-all-faqs-desktop-search-overlay-panel__search > :deep(svg) {
  width: 0.95rem;
  height: 0.95rem;
  flex: 0 0 auto;
}

.global-all-faqs-desktop-search-overlay-panel__search input {
  min-width: 0;
  flex: 1;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--tz-text-primary);
  font-size: 0.8rem;
}

.global-all-faqs-desktop-search-overlay-panel__search input::placeholder {
  color: var(--tz-text-muted);
}

.global-all-faqs-desktop-search-overlay-panel__clear {
  display: inline-grid;
  width: 1.55rem;
  height: 1.55rem;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 999px;
  background: #e4eaee;
  color: var(--tz-text-secondary);
  cursor: pointer;
}

.global-all-faqs-desktop-search-overlay-panel__clear :deep(svg) {
  width: 0.75rem;
  height: 0.75rem;
}

.global-all-faqs-desktop-search-overlay-panel__body {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  padding: 0 1.2rem 1.2rem;
}

.global-all-faqs-desktop-search-overlay-panel__search-results,
.global-all-faqs-desktop-search-overlay-panel__questions-section {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.global-all-faqs-desktop-search-overlay-panel__search-results
  > :deep(.global-all-faqs-desktop-search-accordion),
.global-all-faqs-desktop-search-overlay-panel__questions-section
  > :deep(.global-all-faqs-desktop-search-accordion) {
  flex: 1 1 auto;
  min-height: 0;
}

.global-all-faqs-desktop-search-overlay-panel__search-results {
  flex: 1 1 auto;
  padding-bottom: 0.15rem;
}

.global-all-faqs-desktop-search-overlay-panel__questions-section {
  flex: 1 1 auto;
  padding-bottom: 0.2rem;
}

.global-all-faqs-desktop-search-overlay-panel__topics-section {
  flex: 0 0 13rem;
  min-height: 13rem;
  overflow: hidden;
  margin-top: auto;
  padding-top: 0.65rem;
  padding-bottom: 0.35rem;
}

.global-all-faqs-desktop-search-overlay-panel__section-heading {
  display: flex;
  min-height: 2rem;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.55rem;
  padding-bottom: 0.45rem;
}

.global-all-faqs-desktop-search-overlay-panel__section-kicker {
  display: block;
  margin-bottom: 0.18rem;
  color: var(--tz-text-muted);
  font-size: 0.55rem;
  font-weight: 900;
  letter-spacing: 0.13em;
  line-height: 1.2;
  text-transform: uppercase;
}

.global-all-faqs-desktop-search-overlay-panel__section-heading h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 900;
  line-height: 1.25;
}

.global-all-faqs-desktop-search-overlay-panel__result-count {
  display: inline-grid;
  min-width: 1.7rem;
  height: 1.7rem;
  place-items: center;
  border: 1px solid rgba(20, 32, 43, 0.16);
  border-radius: 999px;
  color: var(--tz-text-accent);
  font-size: 0.62rem;
  font-weight: 900;
}

.global-all-faqs-desktop-search-overlay-panel__topics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
}

.global-all-faqs-desktop-search-overlay-panel__topic {
  display: grid;
  min-width: 0;
  height: 4.1rem;
  gap: 0.8rem;
  padding: 0.75rem;
  border: 1px solid transparent;
  border-radius: 0.7rem;
  background: var(--tz-card-surface, #ffffff);
  color: var(--tz-text-primary);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.global-all-faqs-desktop-search-overlay-panel__topic:hover,
.global-all-faqs-desktop-search-overlay-panel__topic.is-active {
  border-color: rgba(5, 150, 105, 0.48);
  background:
    linear-gradient(0deg, rgba(5, 150, 105, 0.075), rgba(5, 150, 105, 0.075)),
    var(--tz-card-surface, #ffffff);
}

.global-all-faqs-desktop-search-overlay-panel__topic-label {
  min-width: 0;
  overflow: hidden;
  font-size: 0.74rem;
  font-weight: 800;
  line-height: 1.35;
  text-overflow: ellipsis;
}

.global-all-faqs-desktop-search-overlay-panel__topic-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  color: var(--tz-text-muted);
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.global-all-faqs-desktop-search-overlay-panel__topic-footer :deep(svg) {
  width: 0.85rem;
  height: 0.85rem;
  color: var(--tz-text-accent);
}

.global-all-faqs-desktop-search-overlay-panel__empty,
.global-all-faqs-desktop-search-overlay-panel__loading {
  display: grid;
  min-height: 10rem;
  place-items: center;
  border: 1px dashed rgba(20, 32, 43, 0.18);
  border-radius: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.5;
  text-align: center;
}

.global-all-faqs-desktop-search-overlay-panel__loading {
  color: var(--tz-text-accent);
}

@media (min-width: 768px) {
  .global-all-faqs-desktop-search-overlay-panel {
    display: flex;
  }
}
</style>
