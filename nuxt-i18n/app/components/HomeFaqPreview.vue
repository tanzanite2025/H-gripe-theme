<template>
  <section class="home-faq" :class="{ 'home-faq--wide': wide, 'home-faq--fluid': fluid }">
    <!-- Premium Card Container -->
    <div class="rounded-2xl premium-card p-3 md:p-4">
      
      <div class="home-faq__header">
        <div class="home-faq__header-main">
          <h2 class="home-faq__title tz-faq-title">{{ t('faq.title') }}</h2>
          <span class="home-faq__status-badge" aria-hidden="true">
            <span class="home-faq__status-dot"></span>
            {{ t('faq.ui.categorizedSupport') }}
          </span>
        </div>
        <p class="home-faq__subtitle tz-faq-subtitle">{{ t('faq.ui.quickAnswers') }}</p>
      </div>

      <div class="home-faq__desktop-layout">
        <aside class="home-faq__sidebar" :aria-label="t('faq.ui.categoriesAriaLabel')">
          <div class="home-faq__sidebar-label">{{ t('faq.ui.categoriesLabel') }}</div>
          <div class="home-faq__sidebar-tabs">
            <button
              v-for="(page, index) in previewPages"
              :key="page.pageId"
              type="button"
              class="home-faq__sidebar-button"
              :class="{
                'home-faq__sidebar-button--active': activePageId === page.pageId
                  || (activePageId === 'all' && index === 0),
              }"
              @click="activePageId = page.pageId"
            >
              <span>{{ formatPageIndex(index) }}. {{ page.title || page.pageId }}</span>
              <span class="home-faq__sidebar-dot"></span>
            </button>
          </div>
          <div class="home-faq__sidebar-more" aria-hidden="true">
            <span class="home-faq__more-dots">
              <span></span>
              <span></span>
              <span></span>
            </span>
            <svg class="home-faq__more-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14m0 0l-5-5m5 5l5-5" />
            </svg>
          </div>
        </aside>

        <main class="home-faq__desktop-panel">
          <section v-if="desktopGroup" :key="desktopGroup.pageId" class="home-faq__group">
            <div class="home-faq__group-title">{{ desktopGroup.pageTitle }}</div>
            <DesktopFaqMasterDetail
              class="home-faq__desktop-master-detail"
              :items="desktopGroup.items"
              :expanded-items="expandedItems"
              id-prefix="home-faq-desktop-answer"
              @toggle-item="toggleItem"
            />
          </section>
        </main>
      </div>

      <div class="home-faq__mobile-preview">
        <div class="nav-pill-tabs">
          <button
            type="button"
            class="nav-pill-item"
            :class="{ 'nav-pill-item--active': activePageId === 'all' }"
            @click="activePageId = 'all'"
          >
            {{ t('faq.ui.all') }}
          </button>
          <button
            v-for="page in previewPages"
            :key="page.pageId"
            type="button"
            class="nav-pill-item"
            :class="{ 'nav-pill-item--active': activePageId === page.pageId }"
            @click="activePageId = page.pageId"
          >
            {{ page.title || page.pageId }}
          </button>
        </div>

        <div class="home-faq__mobile-list rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] bg-slate-900/40 border border-slate-800/50">
          <div
            v-for="item in displayItems"
            :key="item.id"
            class="home-faq__item border-b border-slate-800/50 last:border-b-0"
          >
            <button type="button" class="home-faq__question group" @click="toggleItem(item.id)">
              <span class="home-faq__question-text tz-faq-question group-hover:text-sky-400 transition-colors">{{ item.question }}</span>
              <svg
                class="home-faq__chevron"
                :class="{ 'home-faq__icon--open': expandedItems.has(item.id) }"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              leave-active-class="transition-all duration-150 ease-in"
              enter-from-class="opacity-0 max-h-0"
              enter-to-class="opacity-100 max-h-[500px]"
              leave-from-class="opacity-100 max-h-[500px]"
              leave-to-class="opacity-0 max-h-0"
            >
              <SafeRichText
                v-if="expandedItems.has(item.id)"
                class="home-faq__answer tz-faq-answer bg-slate-900/30"
                :html="item.answer"
              />
            </Transition>
          </div>
        </div>
      </div>

      <!-- 查看全部链接 -->
      <div class="home-faq__footer">
        <div class="home-faq__footer-actions">
          <NuxtLink :to="localePath('/support/faqs')" class="home-faq__link">
            {{ t('faq.ui.viewAll') }}
            <svg class="home-faq__link-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </NuxtLink>
        </div>
      </div>
      
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useLocalePath } from '#imports'
import DesktopFaqMasterDetail from '~/components/faq/DesktopFaqMasterDetail.vue'
import { useFaqAccordionState } from '~/composables/useFaqAccordionState'
import { useFaqCatalog } from '~/composables/useFaqCatalog'

interface Props {
  maxItemsPerCategory?: number
  maxCategories?: number
  preferredPageIds?: string[]
  defaultCategory?: string
  wide?: boolean
  fluid?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxItemsPerCategory: 3,
  maxCategories: 4,
  preferredPageIds: () => [
    'support-payment',
    'company-ourstory',
    'company-oem-odm',
    'guides-wheelset-buyers',
  ],
  defaultCategory: '',
  wide: false,
  fluid: false,
})

const wide = computed(() => props.wide)
const localePath = useLocalePath()

// 获取所有 FAQ 数据
const { t } = useI18n()
const { allPages } = await useFaqCatalog()
const previewPages = computed(() => {
  const pageById = new Map(allPages.value.map(page => [page.pageId, page]))
  const curatedPages = props.preferredPageIds
    .map(pageId => pageById.get(pageId))
    .filter((page): page is typeof allPages.value[number] => Boolean(page))

  const selectedIds = new Set(curatedPages.map(page => page.pageId))
  const fallbackPages = allPages.value.filter(page => !selectedIds.has(page.pageId))

  return [...curatedPages, ...fallbackPages].slice(0, props.maxCategories)
})

// 当前选中的分类
const activePageId = ref<string>('all')

// 展开的条目
const {
  expandedItems,
  toggleItem,
  resetExpandedItems,
} = useFaqAccordionState()

// 初始化默认分类
watch(allPages, (pages) => {
  if (props.defaultCategory) {
    activePageId.value = props.defaultCategory
  } else if (activePageId.value !== 'all' && !pages.some((page) => page.pageId === activePageId.value)) {
    activePageId.value = 'all'
  }
}, { immediate: true })

watch(activePageId, () => {
  resetExpandedItems()
})

// 扁平化并限制条目数量
interface FlatItem {
  id: string
  category: string
  pageTitle: string
  question: string
  answer: string
  answerImageUrl?: string
  answerImageAlt?: string
  answerImageWidth?: number
  answerImageHeight?: number
}

const categoryPriorityByPageId: Record<string, string[]> = {
  'support-payment': ['security', 'payment-methods', 'billing', 'troubleshooting'],
}

const pageItems = (page: typeof allPages.value[number]): FlatItem[] => {
  const items: FlatItem[] = []
  const categoryPriority = categoryPriorityByPageId[page.pageId] || []
  const categories = [...page.categories]
    .filter(category => category.items.length > 0)
    .sort((a, b) => {
      const aIndex = categoryPriority.indexOf(a.id)
      const bIndex = categoryPriority.indexOf(b.id)

      if (aIndex === -1 && bIndex === -1) return 0
      if (aIndex === -1) return 1
      if (bIndex === -1) return -1
      return aIndex - bIndex
    })

  let itemIndex = 0

  while (items.length < props.maxItemsPerCategory && categories.some(category => category.items[itemIndex])) {
    for (const category of categories) {
      const item = category.items[itemIndex]
      if (!item) continue
      if (items.length >= props.maxItemsPerCategory) return items

      items.push({
        id: `${page.pageId}-${category.id}-${item.id}`,
        category: category.name,
        pageTitle: page.title || page.pageId,
        question: item.question,
        answer: item.answer,
        answerImageUrl: item.answerImageUrl,
        answerImageAlt: item.answerImageAlt,
        answerImageWidth: item.answerImageWidth,
        answerImageHeight: item.answerImageHeight,
      })
    }

    itemIndex++
  }

  return items
}

const desktopGroup = computed(() => {
  const page = activePageId.value === 'all'
    ? previewPages.value[0]
    : previewPages.value.find(item => item.pageId === activePageId.value)

  if (!page) return null

  return {
    pageId: page.pageId,
    pageTitle: page.title || page.pageId,
    items: pageItems(page),
  }
})

const displayItems = computed<FlatItem[]>(() => {
  const page = activePageId.value === 'all'
    ? previewPages.value[0]
    : previewPages.value.find(item => item.pageId === activePageId.value)
  return page ? pageItems(page) : []
})

const formatPageIndex = (index: number) => String(index + 1).padStart(2, '0')
</script>

<style scoped>
.home-faq {
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 2rem 0;
}

.home-faq--wide {
  max-width: 1200px;
  padding: 2rem 0.5rem;
}

@media (min-width: 640px) {
  .home-faq--wide {
    padding: 2rem 1rem;
  }
}

.home-faq--fluid {
  max-width: none;
  padding-inline: 0;
}

.home-faq__header {
  text-align: center;
  margin-bottom: 2rem;
}

.home-faq__header-main {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

.home-faq__title {
  margin: 0 0 0.35rem;
  color: var(--tz-text-primary);
}

.home-faq__subtitle {
  margin: 0;
  color: var(--tz-text-secondary);
}

.home-faq__desktop-layout,
.home-faq__status-badge,
.home-faq__group-title,
.home-faq__category,
.home-faq__plus {
  display: none;
}

.home-faq__mobile-preview {
  display: block;
}

/* Global classes .nav-pill-tabs and .nav-pill-item are used now */

.home-faq__question {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: transparent;
  border: none;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s;
}

.home-faq__question:hover {
  background: rgba(255, 255, 255, 0.03);
}

.home-faq__question-text {
  flex: 1;
  color: var(--tz-text-secondary);
}

.home-faq__icon {
  flex-shrink: 0;
  width: 1.1rem;
  height: 1.1rem;
  color: var(--tz-text-muted);
  transition: transform 0.2s, color 0.2s;
}

.home-faq__question:hover .home-faq__icon {
  color: var(--tz-text-secondary);
}

.home-faq__icon--open {
  transform: rotate(180deg);
  color: #B5FF6D;
}

.home-faq__chevron {
  display: block;
  width: 1.1rem;
  height: 1.1rem;
}

.home-faq__answer {
  padding: 0 1.25rem 1.25rem 1.25rem;
  color: var(--tz-text-secondary);
  overflow: hidden;
}

.home-faq__footer {
  text-align: center;
  margin-top: 2rem;
}

.home-faq__footer-actions {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.home-faq__link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1.5rem;
  border-radius: 9999px;
  font-size: 0.85rem;
  font-weight: 700;
  color: #000000;
  background: #B5FF6D;
  border: none;
  box-shadow: 0 4px 12px rgba(181, 255, 109, 0.22);
  text-decoration: none;
  transition: all 0.2s ease;
  letter-spacing: 0.025em;
}

.home-faq__link:hover {
  transform: translateY(-1px);
  background: #c8ff91;
  box-shadow: 0 8px 16px -4px rgba(181, 255, 109, 0.38);
}

.home-faq__link-icon {
  width: 1rem;
  height: 1rem;
}

@media (min-width: 768px) {
  .home-faq {
    max-width: min(100rem, calc(100vw - 5rem));
    padding: 2rem 0 2.25rem;
  }

  .home-faq > .premium-card {
    padding: 0 !important;
    border-radius: 0 !important;
    background: transparent !important;
    box-shadow: none !important;
  }

  .home-faq__header {
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    text-align: left;
  }

  .home-faq__header-main {
    justify-content: space-between;
  }

  .home-faq__title {
    display: block;
    margin: 0;
    color: #ffffff;
    font-size: 1.45rem;
    font-style: italic;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .home-faq__subtitle {
    max-width: none;
    margin: 0.25rem 0 0;
    color: #94a3b8;
    font-size: 0.9rem;
    text-align: left;
  }

  .home-faq__status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
    padding: 0.18rem 0.65rem;
    border: 1px solid rgba(181, 255, 109, 0.32);
    border-radius: 999px;
    background: rgba(6, 78, 59, 0.36);
    color: #B5FF6D;
    font-size: 0.68rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .home-faq__status-dot {
    width: 0.38rem;
    height: 0.38rem;
    border-radius: 999px;
    background: #B5FF6D;
    box-shadow: 0 0 12px rgba(181, 255, 109, 0.8);
    animation: home-faq-status-pulse 1s ease-in-out infinite alternate;
  }

  .home-faq__desktop-layout {
    display: grid !important;
    grid-template-columns: minmax(15.5rem, 0.24fr) minmax(0, 0.76fr);
    gap: 1.5rem;
    align-items: start;
  }

  .home-faq__mobile-preview {
    display: none !important;
    height: 0 !important;
    overflow: hidden !important;
    visibility: hidden !important;
  }

  .home-faq__sidebar {
    position: sticky;
  top: calc(112px + 1rem);
    box-sizing: border-box;
    min-width: 0;
    max-width: 100%;
    overflow: hidden;
    display: grid;
    gap: 0.25rem;
    padding: 0.75rem;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 1rem;
    background: #000000;
  }

  .home-faq__sidebar-label {
    padding: 0.25rem 0.75rem 0.35rem;
    color: #64748b;
    font-size: 0.68rem;
    font-weight: 900;
    line-height: 1.2;
    text-transform: uppercase;
  }

  .home-faq__sidebar-tabs {
    display: grid;
    gap: 0.25rem;
  }

  .home-faq__sidebar-more {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.35rem;
    min-height: 1.45rem;
    margin-top: 0.15rem;
    color: #B5FF6D;
    opacity: 0.78;
  }

  .home-faq__more-dots {
    display: inline-flex;
    align-items: center;
    gap: 0.18rem;
  }

  .home-faq__more-dots span {
    width: 0.24rem;
    height: 0.24rem;
    border-radius: 999px;
    background: currentColor;
  }

  .home-faq__more-arrow {
    width: 0.85rem;
    height: 0.85rem;
  }

  .home-faq__sidebar-button {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    box-sizing: border-box;
    width: 100%;
    min-width: 0;
    max-width: 100%;
    min-height: 2.65rem;
    padding: 0.75rem 0.9rem;
    border: 0;
    border-radius: 0.75rem;
    background: transparent;
    color: #94a3b8;
    text-align: left;
    font-size: 0.86rem;
    font-weight: 900;
    line-height: 1.25;
    text-transform: uppercase;
    cursor: pointer;
    transition: background 0.2s ease, color 0.2s ease;
  }

  .home-faq__sidebar-button:hover {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.04);
  }

  .home-faq__sidebar-button--active {
    color: #000000;
    background: #ffffff;
    box-shadow: 0 8px 18px rgba(0, 0, 0, 0.35);
  }

  .home-faq__sidebar-button > span:first-child {
    min-width: 0;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow-wrap: anywhere;
    white-space: normal;
  }

  .home-faq__sidebar-dot {
    width: 0.42rem;
    height: 0.42rem;
    flex-shrink: 0;
    border-radius: 999px;
    background: #334155;
  }

  .home-faq__sidebar-button--active .home-faq__sidebar-dot {
    background: #B5FF6D;
    box-shadow: 0 0 10px rgba(181, 255, 109, 0.7);
  }

  .home-faq__desktop-panel {
    min-width: 0;
    min-height: 24rem;
  }

  .home-faq__desktop-master-detail {
    min-height: 24rem;
  }

  .home-faq__group {
    margin-bottom: 0.85rem;
  }

  .home-faq__group:last-child {
    margin-bottom: 0;
  }

  .home-faq__group-title {
    display: block;
    margin-bottom: 0.65rem;
    padding-bottom: 0.55rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    color: #94a3b8;
    font-size: 0.86rem;
    font-weight: 900;
    letter-spacing: 0.16em;
    text-transform: uppercase;
  }

  .home-faq__content {
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

  .home-faq__column {
    display: grid;
    gap: 0.625rem;
    align-content: start;
  }

  .home-faq__item {
    position: relative;
    align-self: start;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.1) !important;
    border-radius: 1rem;
    background: #000000;
  }

  .home-faq__item::before {
    position: absolute;
    inset: 0 auto 0 0;
    width: 0;
    content: '';
    background: #B5FF6D;
    box-shadow: 0 0 10px #B5FF6D;
    transition: width 0.3s ease;
  }

  .home-faq__item:has(.home-faq__icon--open) {
    border-color: rgba(255, 255, 255, 0.3) !important;
  }

  .home-faq__item:has(.home-faq__icon--open)::before {
    width: 3px;
  }

  .home-faq__question {
    position: relative;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    justify-content: stretch;
    padding: 1.05rem 1rem !important;
    gap: 0.75rem;
    background: transparent !important;
  }

  .home-faq__question:hover {
    background: rgba(255, 255, 255, 0.025) !important;
  }

  .home-faq__category {
    display: block;
    max-width: 10rem;
    overflow: hidden;
    flex-shrink: 0;
    padding: 0.25rem 0.65rem;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 999px;
    background: rgba(15, 23, 42, 0.72);
    color: #94a3b8;
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    line-height: 1.2;
    text-overflow: ellipsis;
    text-transform: uppercase;
    white-space: nowrap;
  }

  .home-faq__question-text {
    min-width: 0;
    color: #ffffff !important;
    font-size: 0.96rem;
    font-weight: 800;
    line-height: 1.45;
  }

  .home-faq__icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.5rem;
    height: 1.5rem;
    border-radius: 0;
    background: transparent;
    color: #94a3b8;
    transform: none;
  }

  @media (max-width: 1280px) {
    .home-faq__desktop-layout {
      grid-template-columns: minmax(14rem, 0.26fr) minmax(0, 0.74fr);
      gap: 1.25rem;
    }

    .home-faq__question {
      grid-template-columns: minmax(0, 1fr) auto;
      row-gap: 0.55rem;
    }

    .home-faq__category {
      grid-column: 1 / -1;
      max-width: 100%;
      justify-self: start;
    }

    .home-faq__question-text {
      grid-column: 1;
    }

    .home-faq__icon {
      grid-column: 2;
    }
  }

  .home-faq__chevron {
    display: none;
  }

  .home-faq__plus {
    display: block;
    font-size: 1.2rem;
    font-weight: 900;
    line-height: 1;
  }

  .home-faq__icon--open {
    color: #B5FF6D;
    transform: rotate(45deg);
  }

  .home-faq__answer {
    margin: 0 1rem 1rem;
    padding: 0.8rem 0 0 !important;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    color: #cbd5e1;
    font-size: 0.9rem;
    line-height: 1.7;
  }

  .home-faq__footer {
    margin-top: 2rem;
  }
}

@keyframes home-faq-status-pulse {
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
