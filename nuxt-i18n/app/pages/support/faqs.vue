<template>
  <div class="faqs-page">
    <header class="faqs-page__header">
      <div class="faqs-page__header-main">
        <h2 class="faqs-page__title">{{ t('faq.title') }}</h2>
        <span class="faqs-page__status-badge" aria-hidden="true">
          <span class="faqs-page__status-dot" />
          {{ t('faq.ui.categorizedSupport') }}
        </span>
      </div>
      <p class="faqs-page__intro">
        {{ t('faq.ui.pageIntro') }}
      </p>
    </header>

    <!-- 搜索框 -->
    <div class="faqs-search">
      <input
        v-model="searchQuery"
        type="text"
        :placeholder="t('faq.ui.searchPlaceholder')"
        class="faqs-search__input"
      />
      <span v-if="searchQuery" class="faqs-search__clear" @click="searchQuery = ''">✕</span>
    </div>

    <div class="faqs-layout">
      <!-- 页面分类标签 -->
      <aside class="faqs-sidebar" :aria-label="t('faq.ui.categoriesAriaLabel')">
        <div class="faqs-sidebar__label">{{ t('faq.ui.categoriesLabel') }}</div>
        <div class="faqs-tabs">
          <button
            type="button"
            class="premium-button faqs-tabs__button"
            :class="{ 'premium-button--active': activePageId === 'all' }"
            @click="activePageId = 'all'"
          >
            <span class="faqs-tabs__label">{{ t('faq.ui.all') }}</span>
            <span class="faqs-tabs__dot" aria-hidden="true" />
          </button>
          <button
            v-for="(page, index) in allPages"
            :key="page.pageId"
            type="button"
            class="premium-button faqs-tabs__button"
            :class="{ 'premium-button--active': activePageId === page.pageId }"
            @click="activePageId = page.pageId"
          >
            <span class="faqs-tabs__label">
              <span class="faqs-tabs__index">{{ formatPageIndex(index) }}.</span>
              {{ page.title || page.pageId }}
            </span>
            <span class="faqs-tabs__dot" aria-hidden="true" />
          </button>
        </div>
      </aside>

    <!-- FAQ 内容 -->
    <div v-if="filteredItems.length > 0" class="faqs-content">
      <Transition
        enter-active-class="transition-opacity duration-300 ease-out"
        leave-active-class="transition-opacity duration-200 ease-in"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
        mode="out-in"
      >
        <div :key="activePageId + searchQuery" class="faqs-content__results">
          <GlobalAllFaqsDesktopGroupedSearchResultsPanel
            :groups="displayedGroups"
            :expanded-items="expandedItems"
            @toggle-item="toggleItem"
          />
          <GlobalAllFaqsMobileGroupedSearchResultsPanel
            :groups="displayedGroups"
            :expanded-items="expandedItems"
            @toggle-item="toggleItem"
          />
        </div>
      </Transition>

    </div>

    <!-- 无结果 -->
    <div v-else class="faqs-empty">
      <p>{{ t('faq.ui.noResults', { query: searchQuery }) }}</p>
    </div>
    </div>

    <div
      v-if="hasMoreGroups && activePageId === 'all' && !searchQuery"
      class="faqs-load-more flex justify-center mt-4 mb-8"
    >
      <button
        type="button"
        class="inline-flex items-center gap-2 px-8 py-3 rounded-full text-sm font-bold bg-[var(--tz-action-primary)] text-white hover:bg-[var(--tz-action-primary-hover)] hover:shadow-[0_8px_18px_rgba(15,23,42,0.16)] transition-all"
        @click="loadMoreGroups"
      >
        {{ t('faq.ui.viewMoreContent') }}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useHead } from '#imports'
import GlobalAllFaqsDesktopGroupedSearchResultsPanel from '~/components/faq/GlobalAllFaqsDesktopGroupedSearchResultsPanel.vue'
import GlobalAllFaqsMobileGroupedSearchResultsPanel from '~/components/faq/GlobalAllFaqsMobileGroupedSearchResultsPanel.vue'
import { useGlobalAllFaqsSearchAndGroupedResults } from '~/composables/useGlobalAllFaqsSearchAndGroupedResults'

definePageMeta({
  layout: 'support',
  footerLabelKey: 'support.nav.faqs',
  footerLabelFallback: 'All FAQs',
})

const { t } = useI18n()

useHead({
  title: () => t('faq.ui.allFaqsMetaTitle'),
})

const {
  allPages,
  searchQuery,
  activePageId,
  expandedItems,
  filteredItems,
  displayedGroups,
  hasMoreGroups,
  toggleItem,
  loadMoreGroups,
} = await useGlobalAllFaqsSearchAndGroupedResults()

const formatPageIndex = (index: number) => String(index + 1).padStart(2, '0')
</script>

<style scoped>
.faqs-page {
  width: 100%; /* Use full available width (parent handles padding) */
  max-width: none;
  margin: 0 auto;
  padding: 0;
}

.faqs-page__title,
.faqs-page__status-badge,
.faqs-sidebar__label,
.faqs-tabs__index,
.faqs-tabs__dot {
  display: none;
}

.faqs-page__intro {
  margin: 0 auto 1.5rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
  max-width: 600px;
  text-align: center;
}

.faqs-search {
  position: relative;
  margin-bottom: 2rem;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.faqs-search__input {
  width: 100%;
  padding: 0.8rem 2.5rem 0.8rem 1.25rem;
  border-radius: 9999px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  background: var(--tz-input-surface);
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  transition: all 0.2s;
}

.faqs-search__input::placeholder {
  color: var(--tz-text-muted);
}

.faqs-search__input:focus {
  outline: none;
  border-color: var(--tz-form-control-focus-border);
  background: var(--tz-input-surface);
  box-shadow: 0 0 0 4px var(--tz-form-control-focus-ring);
}

.faqs-search__clear {
  position: absolute;
  right: 1.25rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--tz-text-muted);
  cursor: pointer;
  font-size: 0.85rem;
}

.faqs-search__clear:hover {
  color: #e2e8f0;
}

.faqs-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 2.5rem;
  justify-content: center;
}

.faqs-empty {
  text-align: center;
  padding: 4rem 1rem;
  color: var(--tz-text-muted);
  font-size: 1rem;
  background: var(--tz-form-panel-surface);
  border-radius: 1rem;
  border: 1px dashed rgba(148, 163, 184, 0.2);
}

@media (min-width: 768px) {
  .faqs-page {
    max-width: min(100rem, calc(100vw - 5rem));
    padding: 2rem 0 0;
    background-color: var(--tz-card-surface);
    background-image: radial-gradient(rgba(20, 32, 43, 0.04) 1px, transparent 0);
    background-size: 24px 24px;
    color: var(--tz-text-primary);
  }

  .faqs-page__header {
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid rgba(20, 32, 43, 0.12);
  }

  .faqs-page__header-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .faqs-page__title {
    display: block;
    margin: 0;
    color: var(--tz-text-primary);
    font-size: 1.25rem;
    font-style: italic;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .faqs-page__status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
    padding: 0.18rem 0.65rem;
    border-radius: 999px;
    border: 1px solid rgba(4, 120, 87, 0.28);
    background: rgba(5, 150, 105, 0.2);
    color: var(--tz-text-accent);
    font-size: 0.58rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .faqs-page__status-dot {
    width: 0.38rem;
    height: 0.38rem;
    border-radius: 999px;
    background: #059669;
    box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.22);
    animation: faqs-page-status-pulse 1s ease-in-out infinite alternate;
  }

  .faqs-page__intro {
    max-width: none;
    margin: 0.25rem 0 0;
    color: var(--tz-text-secondary);
    font-size: 0.78rem;
    text-align: left;
  }

  .faqs-search {
    max-width: none;
    margin-bottom: 1.5rem;
  }

  .faqs-search__input {
    padding: 0.78rem 2.5rem 0.78rem 1rem;
    border-radius: 0.9rem;
    border-color: rgba(20, 32, 43, 0.14);
    background: var(--tz-card-surface);
    font-size: 0.82rem;
  }

  .faqs-search__input:focus {
    border-color: var(--tz-form-control-focus-border);
    background: var(--tz-input-surface);
    box-shadow: 0 0 0 4px var(--tz-form-control-focus-ring);
  }

  .faqs-layout {
    display: grid;
    grid-template-columns: clamp(15rem, 21vw, 18rem) minmax(0, 1fr);
    gap: 1.5rem;
    align-items: stretch;
  }

  .faqs-sidebar {
    position: sticky;
    top: calc(112px + 1rem);
    display: grid;
    align-content: start;
    min-width: 0;
    max-width: 100%;
    gap: 0.25rem;
    box-sizing: border-box;
    padding: 0.75rem;
    border: 1px solid rgba(20, 32, 43, 0.14);
    border-radius: 1rem;
    background: var(--tz-card-surface);
    overflow: hidden;
  }

  .faqs-sidebar__label {
    display: block;
    padding: 0.25rem 0.75rem 0.35rem;
    color: var(--tz-text-muted);
    font-size: 0.58rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .faqs-tabs {
    display: grid;
    min-width: 0;
    width: 100%;
    gap: 0.25rem;
    justify-content: stretch;
    margin-bottom: 0;
  }

  .faqs-tabs__button.premium-button {
    display: flex;
    justify-content: space-between;
    width: 100%;
    min-width: 0;
    max-width: 100%;
    min-height: 2.35rem;
    box-sizing: border-box;
    padding: 0.68rem 0.85rem;
    border: 0;
    border-radius: 0.75rem;
    background: transparent !important;
    box-shadow: none !important;
    color: var(--tz-text-secondary);
    font-size: 0.74rem;
    font-weight: 900;
    line-height: 1.25;
    text-align: left;
    text-transform: uppercase;
  }

  .faqs-tabs__button.premium-button:hover {
    color: var(--tz-text-primary);
    background: rgba(20, 32, 43, 0.04) !important;
    transform: none;
  }

  .faqs-tabs__button.premium-button--active {
    color: #ffffff !important;
    background: var(--tz-text-primary) !important;
    box-shadow: 0 8px 18px rgba(20, 32, 43, 0.16) !important;
  }

  .faqs-tabs__label {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .faqs-tabs__index {
    display: inline;
  }

  .faqs-tabs__dot {
    display: block;
    width: 0.42rem;
    height: 0.42rem;
    flex: 0 0 auto;
    border-radius: 999px;
    background: #cbd5e1;
  }

  .faqs-tabs__button.premium-button--active .faqs-tabs__dot {
    background: #059669;
    box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.2);
  }

  .faqs-content {
    display: flex;
    min-width: 0;
    flex-direction: column;
  }

  .faqs-content__results {
    height: 100%;
    min-height: 0;
  }

  .faqs-content__results :deep(.global-all-faqs-desktop-grouped-search-results-panel) {
    height: 100%;
    min-height: 100%;
  }

  .faqs-content__results :deep(.desktop-faq-master-detail__list),
  .faqs-content__results :deep(.desktop-faq-master-detail__detail-content) {
    max-height: none;
    overflow: visible;
  }

  .faqs-load-more {
    margin-left: clamp(15rem, 21vw, 18rem);
  }

  .faqs-empty {
    grid-column: 2;
    background: var(--tz-card-surface);
    border-color: rgba(20, 32, 43, 0.14);
  }
}

@media (min-width: 768px) and (max-width: 980px) {
  .faqs-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .faqs-sidebar {
    position: static;
  }

  .faqs-load-more {
    margin-left: 0;
  }

  .faqs-tabs {
    grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
  }

  .faqs-empty {
    grid-column: auto;
  }
}

@keyframes faqs-page-status-pulse {
  from {
    opacity: 0.45;
    transform: scale(0.86);
  }

  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
