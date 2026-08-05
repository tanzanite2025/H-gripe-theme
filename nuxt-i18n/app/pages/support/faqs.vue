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
      <!-- 按页面分组显示 -->
      <Transition
        enter-active-class="transition-opacity duration-300 ease-out"
        leave-active-class="transition-opacity duration-200 ease-in"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
        mode="out-in"
      >
      <div :key="activePageId + searchQuery"> <!-- Add key to force re-render on search/tab change for transition -->
        <div
          v-for="group in displayedGroups"
          :key="group.pageId"
          class="faqs-group rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-3 md:p-4 mb-8"
        >
          <div class="faqs-group__header flex items-center justify-center border-b border-slate-800/50">
            <h3 class="faqs-group__title tz-faq-category-title tz-text-primary text-center">
              {{ group.pageTitle }}
            </h3>
          </div>
          
          <!-- FAQ 条目 -->
          <div class="faqs-list faqs-list--mobile rounded-xl overflow-hidden bg-slate-900/40 border border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
            <div
              v-for="item in group.items"
              :key="item.id"
              class="faqs-item border-b border-slate-800/50 last:border-b-0"
              :class="{ 'faqs-item--expanded': expandedItems.has(item.id) }"
            >
              <button
                type="button"
                class="faqs-item__button w-full flex items-center gap-3 px-3 py-3 text-left transition-colors group hover:bg-white/5"
                :class="{ 'bg-white/5': expandedItems.has(item.id) }"
                @click="toggleItem(item.id)"
              >
                <span class="faqs-item__category flex-shrink-0 px-2.5 py-1 rounded-full bg-slate-800 tz-text-muted tz-micro-label uppercase font-bold tracking-wider border border-slate-700">
                  {{ item.category }}
                </span>
                <span 
                  class="faqs-item__question tz-faq-question flex-1 tz-text-primary group-hover:text-sky-400 transition-colors"
                   :class="{ 'text-sky-400': expandedItems.has(item.id) }"
                >
                  {{ item.question }}
                </span>
                <span
                  class="faqs-item__icon flex-shrink-0 w-5 h-5 flex items-center justify-center text-slate-500 transition-transform duration-200"
                  :class="{ 'rotate-180 text-sky-400': expandedItems.has(item.id) }"
                >
                  <svg class="faqs-item__chevron w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                  <span class="faqs-item__plus" aria-hidden="true">+</span>
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
                <div v-if="expandedItems.has(item.id)" class="faqs-item__answer-wrap overflow-hidden bg-slate-950/30">
                  <div class="faqs-item__answer tz-faq-answer px-4 pb-4 pt-1 tz-text-secondary">
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

          <div class="faqs-list faqs-list--desktop rounded-xl overflow-hidden bg-slate-900/40 border border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
            <div
              v-for="(columnItems, columnIndex) in group.itemColumns"
              :key="columnIndex"
              class="faqs-list__column"
            >
              <div
                v-for="item in columnItems"
                :key="item.id"
                class="faqs-item border-b border-slate-800/50 last:border-b-0"
                :class="{ 'faqs-item--expanded': expandedItems.has(item.id) }"
              >
                <button
                  type="button"
                  class="faqs-item__button w-full flex items-center gap-3 px-3 py-3 text-left transition-colors group hover:bg-white/5"
                  :class="{ 'bg-white/5': expandedItems.has(item.id) }"
                  @click="toggleItem(item.id)"
                >
                  <span class="faqs-item__category flex-shrink-0 px-2.5 py-1 rounded-full bg-slate-800 tz-text-muted tz-micro-label uppercase font-bold tracking-wider border border-slate-700">
                    {{ item.category }}
                  </span>
                  <span
                    class="faqs-item__question tz-faq-question flex-1 tz-text-primary group-hover:text-sky-400 transition-colors"
                     :class="{ 'text-sky-400': expandedItems.has(item.id) }"
                  >
                    {{ item.question }}
                  </span>
                  <span
                    class="faqs-item__icon flex-shrink-0 w-5 h-5 flex items-center justify-center text-slate-500 transition-transform duration-200"
                    :class="{ 'rotate-180 text-sky-400': expandedItems.has(item.id) }"
                  >
                    <svg class="faqs-item__chevron w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                    <span class="faqs-item__plus" aria-hidden="true">+</span>
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
                  <div v-if="expandedItems.has(item.id)" class="faqs-item__answer-wrap overflow-hidden bg-slate-950/30">
                    <div class="faqs-item__answer tz-faq-answer px-4 pb-4 pt-1 tz-text-secondary">
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
      </div>
      </Transition>

      <!-- View More Button -->
      <!-- Only show if we are in 'All' tab AND we have hidden groups AND not searching -->
      <div 
        v-if="hasMoreGroups && activePageId === 'all' && !searchQuery" 
        class="flex justify-center mt-4 mb-8"
      >
        <button
          type="button"
          class="inline-flex items-center gap-2 px-8 py-3 rounded-full text-sm font-bold bg-slate-800 tz-text-secondary hover:bg-slate-700 hover:text-white hover:shadow-lg transition-all"
          @click="loadMoreGroups"
        >
          {{ t('faq.ui.viewMoreContent') }}
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 无结果 -->
    <div v-else class="faqs-empty">
      <p>{{ t('faq.ui.noResults', { query: searchQuery }) }}</p>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useHead, useAsyncData, useRoute } from '#imports'
import { fetchAllFaqData, resolvePageFaqDataList } from '~/data/faq'

definePageMeta({
  layout: 'support',
})

const { locale, t } = useI18n()
const route = useRoute()

useHead({
  title: () => t('faq.ui.allFaqsMetaTitle'),
})

// 获取所有 FAQ 数据 (从 Go API 获取)
const { data: asyncAllPages } = await useAsyncData(
  () => `faqs-all-${locale.value}`,
  () => fetchAllFaqData(),
  { watch: [locale] }
)
const allPages = computed(() => resolvePageFaqDataList(asyncAllPages.value || []))

// 搜索和筛选
const searchQuery = ref('')
const activePageId = ref<string>('all')
const expandedItems = ref<Set<string>>(new Set())
const applyingDeepLink = ref(false)

// Pagination state
const visibleGroupsCount = ref(3)

// Watchers to reset pagination
watch(activePageId, () => {
  if (applyingDeepLink.value) return
  expandedItems.value = new Set<string>()

  if (activePageId.value === 'all') {
    visibleGroupsCount.value = 3
  } else {
    // Single page view usually not paginated, show all
    visibleGroupsCount.value = 999 
  }
})

watch(searchQuery, () => {
  if (applyingDeepLink.value) return
  expandedItems.value = new Set<string>()

  if (searchQuery.value) {
    // Search results should show all matches
    visibleGroupsCount.value = 999
  } else if (activePageId.value === 'all') {
     visibleGroupsCount.value = 3
  }
})

// 切换展开状态
const toggleItem = (itemId: string) => {
  expandedItems.value = expandedItems.value.has(itemId) ? new Set<string>() : new Set([itemId])
}

// 扁平化所有 FAQ 条目
interface FlatFaqItem {
  id: string
  pageId: string
  pageTitle: string
  category: string
  question: string
  answer: string
  answerImageUrl?: string
  answerImageAlt?: string
  answerImageWidth?: number
  answerImageHeight?: number
  tags?: string[]
}

const allItems = computed<FlatFaqItem[]>(() => {
  const items: FlatFaqItem[] = []
  for (const page of allPages.value) {
    for (const category of page.categories) {
      for (const item of category.items) {
        items.push({
          id: `${page.pageId}-${item.id}`,
          pageId: page.pageId,
          pageTitle: page.title || page.pageId,
          category: category.name,
          question: item.question,
          answer: item.answer,
          answerImageUrl: item.answerImageUrl,
          answerImageAlt: item.answerImageAlt,
          answerImageWidth: item.answerImageWidth,
          answerImageHeight: item.answerImageHeight,
          tags: item.tags,
        })
      }
    }
  }
  return items
})

const queryValue = (value: unknown) => {
  if (Array.isArray(value)) return String(value[0] || '').trim()
  return String(value || '').trim()
}

const applyDeepLinkQuery = async () => {
  if (!allPages.value.length) return

  const requestedPageId = queryValue(route.query.page)
  const requestedFaqId = queryValue(route.query.faq)
  const target = requestedFaqId
    ? allItems.value.find((item) => {
        if (requestedPageId && item.pageId !== requestedPageId) return false
        return item.id === requestedFaqId || item.id === `${item.pageId}-${requestedFaqId}`
      })
    : null
  const pageId = target?.pageId
    || (requestedPageId && allPages.value.some((page) => page.pageId === requestedPageId) ? requestedPageId : '')

  applyingDeepLink.value = true
  try {
    activePageId.value = pageId || 'all'
    if (activePageId.value === 'all') {
      const pageIndex = target ? allPages.value.findIndex((page) => page.pageId === target.pageId) : -1
      visibleGroupsCount.value = Math.max(3, pageIndex + 1)
    } else {
      visibleGroupsCount.value = 999
    }

    await nextTick()
    expandedItems.value = target ? new Set<string>([target.id]) : new Set<string>()
  } finally {
    applyingDeepLink.value = false
  }
}

watch(
  [allPages, () => route.query.page, () => route.query.faq],
  () => {
    void applyDeepLinkQuery()
  },
  { immediate: true }
)

// 筛选后的条目
const filteredItems = computed(() => {
  let items = allItems.value

  // 按页面筛选
  if (activePageId.value !== 'all') {
    items = items.filter(item => item.pageId === activePageId.value)
  }

  // 按搜索词筛选
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    items = items.filter(item => 
      item.question.toLowerCase().includes(query) ||
      item.answer.toLowerCase().includes(query) ||
      item.category.toLowerCase().includes(query) ||
      (item.tags && item.tags.some(tag => tag.toLowerCase().includes(query)))
    )
  }

  return items
})

// 按页面分组
type FaqDisplayGroup = {
  pageId: string
  pageTitle: string
  items: FlatFaqItem[]
  itemColumns: [FlatFaqItem[], FlatFaqItem[]]
}

const groupedItems = computed<FaqDisplayGroup[]>(() => {
  const groups: Record<string, FaqDisplayGroup> = {}

  for (const item of filteredItems.value) {
    if (!groups[item.pageId]) {
      groups[item.pageId] = {
        pageId: item.pageId,
        pageTitle: item.pageTitle,
        items: [],
        itemColumns: [[], []],
      }
    }
    const group = groups[item.pageId]
    if (group) {
      group.items.push(item)
      group.itemColumns[group.items.length % 2 === 1 ? 0 : 1]?.push(item)
    }
  }

  return Object.values(groups)
})

// Compute displayed groups based on pagination
const displayedGroups = computed(() => {
  return groupedItems.value.slice(0, visibleGroupsCount.value)
})

const hasMoreGroups = computed(() => {
  return groupedItems.value.length > visibleGroupsCount.value
})

const loadMoreGroups = () => {
  visibleGroupsCount.value += 3
}

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
.faqs-tabs__dot,
.faqs-item__plus {
  display: none;
}

.faqs-page__intro {
  margin: 0 auto 1.5rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
  max-width: 600px;
  text-align: center;
}

 .faqs-list--desktop {
  display: none;
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
  background: rgba(30, 41, 59, 0.5);
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  transition: all 0.2s;
}

.faqs-search__input::placeholder {
  color: var(--tz-text-muted);
}

.faqs-search__input:focus {
  outline: none;
  border-color: rgba(181, 255, 109, 0.5);
  background: rgba(30, 41, 59, 0.8);
  box-shadow: 0 0 0 4px rgba(181, 255, 109, 0.1);
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

.faqs-group__header {
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
}

.faqs-empty {
  text-align: center;
  padding: 4rem 1rem;
  color: var(--tz-text-muted);
  font-size: 1rem;
  background: rgba(30, 41, 59, 0.3);
  border-radius: 1rem;
  border: 1px dashed rgba(148, 163, 184, 0.2);
}

@media (min-width: 768px) {
  .faqs-page {
    max-width: min(100rem, calc(100vw - 5rem));
    padding: 2rem 0 0;
    background-color: #000000;
    background-image: radial-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 0);
    background-size: 24px 24px;
    color: #ffffff;
  }

  .faqs-page__header {
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
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
    color: #ffffff;
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
    border: 1px solid rgba(181, 255, 109, 0.32);
    background: rgba(6, 78, 59, 0.36);
    color: #B5FF6D;
    font-size: 0.58rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .faqs-page__status-dot {
    width: 0.38rem;
    height: 0.38rem;
    border-radius: 999px;
    background: #B5FF6D;
    box-shadow: 0 0 12px rgba(181, 255, 109, 0.8);
    animation: faqs-page-status-pulse 1s ease-in-out infinite alternate;
  }

  .faqs-page__intro {
    max-width: none;
    margin: 0.25rem 0 0;
    color: #94a3b8;
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
    border-color: rgba(255, 255, 255, 0.15);
    background: #000000;
    font-size: 0.82rem;
  }

  .faqs-search__input:focus {
    border-color: rgba(181, 255, 109, 0.55);
    background: #000000;
    box-shadow: 0 0 0 4px rgba(181, 255, 109, 0.1);
  }

  .faqs-layout {
    display: grid;
    grid-template-columns: minmax(11rem, 0.2fr) minmax(0, 0.8fr);
    gap: 1.5rem;
    align-items: start;
  }

  .faqs-sidebar {
    position: sticky;
    top: calc(var(--site-header-offset, 112px) + 1rem);
    display: grid;
    gap: 0.25rem;
    padding: 0.75rem;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 1rem;
    background: #000000;
  }

  .faqs-sidebar__label {
    display: block;
    padding: 0.25rem 0.75rem 0.35rem;
    color: #64748b;
    font-size: 0.58rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .faqs-tabs {
    display: grid;
    gap: 0.25rem;
    justify-content: stretch;
    margin-bottom: 0;
  }

  .faqs-tabs__button.premium-button {
    justify-content: space-between;
    width: 100%;
    min-height: 2.35rem;
    padding: 0.68rem 0.85rem;
    border: 0;
    border-radius: 0.75rem;
    background: transparent !important;
    box-shadow: none !important;
    color: #94a3b8;
    font-size: 0.74rem;
    font-weight: 900;
    line-height: 1.25;
    text-align: left;
    text-transform: uppercase;
  }

  .faqs-tabs__button.premium-button:hover {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.04) !important;
    transform: none;
  }

  .faqs-tabs__button.premium-button--active {
    color: #000000 !important;
    background: #ffffff !important;
    box-shadow: 0 8px 18px rgba(0, 0, 0, 0.35) !important;
  }

  .faqs-tabs__label {
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
    flex-shrink: 0;
    border-radius: 999px;
    background: #334155;
  }

  .faqs-tabs__button.premium-button--active .faqs-tabs__dot {
    background: #B5FF6D;
    box-shadow: 0 0 10px rgba(181, 255, 109, 0.7);
  }

  .faqs-content {
    min-width: 0;
    min-height: 30rem;
  }

  .faqs-group {
    margin-bottom: 0.85rem !important;
    padding: 0 !important;
    border-radius: 0 !important;
    background: transparent !important;
    box-shadow: none !important;
  }

  .faqs-group__header {
    justify-content: flex-start !important;
    margin-bottom: 0.65rem;
    padding-bottom: 0.55rem;
    border-color: rgba(255, 255, 255, 0.1) !important;
  }

  .faqs-group__title {
    color: #94a3b8 !important;
    font-size: 0.72rem;
    font-weight: 900;
    letter-spacing: 0.16em;
    text-align: left;
  }

  .faqs-list--mobile {
    display: none;
  }

  .faqs-list--desktop {
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

  .faqs-list__column {
    display: grid;
    gap: 0.625rem;
    align-content: start;
  }

  .faqs-item {
    position: relative;
    align-self: start;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.1) !important;
    border-radius: 1rem;
    background: #000000;
  }

  .faqs-item::before {
    position: absolute;
    inset: 0 auto 0 0;
    width: 0;
    content: '';
    background: #B5FF6D;
    box-shadow: 0 0 10px #B5FF6D;
    transition: width 0.3s ease;
  }

  .faqs-item--expanded {
    border-color: rgba(255, 255, 255, 0.3) !important;
  }

  .faqs-item--expanded::before {
    width: 3px;
  }

  .faqs-item__button {
    position: relative;
    padding: 1rem !important;
    background: transparent !important;
  }

  .faqs-item__button:hover {
    background: rgba(255, 255, 255, 0.025) !important;
  }

  .faqs-item__category {
    max-width: 9rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    border-color: rgba(255, 255, 255, 0.12) !important;
    background: rgba(15, 23, 42, 0.72) !important;
    color: #94a3b8 !important;
    font-size: 0.62rem;
  }

  .faqs-item__question {
    color: #ffffff !important;
    font-size: 0.82rem;
    font-weight: 800;
    line-height: 1.45;
  }

  .faqs-item__icon {
    width: 1.35rem !important;
    height: 1.35rem !important;
    border-radius: 0;
    background: transparent !important;
    color: #94a3b8 !important;
    transform: none !important;
  }

  .faqs-item--expanded .faqs-item__icon {
    color: #B5FF6D !important;
    transform: rotate(45deg) !important;
  }

  .faqs-item__chevron {
    display: none;
  }

  .faqs-item__plus {
    display: block;
    font-size: 1.05rem;
    font-weight: 900;
    line-height: 1;
  }

  .faqs-item__answer-wrap {
    background: transparent !important;
  }

  .faqs-item__answer {
    margin: 0 1rem 1rem;
    padding: 0.8rem 0 0 !important;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    color: #cbd5e1 !important;
    font-size: 0.78rem;
    line-height: 1.7;
  }

  .faqs-empty {
    grid-column: 2;
    background: #000000;
    border-color: rgba(255, 255, 255, 0.15);
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
