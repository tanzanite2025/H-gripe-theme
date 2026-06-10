<template>
  <div class="faqs-page">
    <h2 class="sr-only">All FAQs</h2>
    <p class="faqs-page__intro">
      Browse common questions and quick answers about orders, products, and service.
    </p>

    <!-- 搜索框 -->
    <div class="faqs-search">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Search FAQs..."
        class="faqs-search__input"
      />
      <span v-if="searchQuery" class="faqs-search__clear" @click="searchQuery = ''">✕</span>
    </div>

    <!-- 页面分类标签 -->
    <div class="faqs-tabs">
      <button
        type="button"
        class="premium-button"
        :class="{ 'premium-button--active': activePageId === 'all' }"
        @click="activePageId = 'all'"
      >
        All
      </button>
      <button
        v-for="page in allPages"
        :key="page.pageId"
        type="button"
        class="premium-button"
        :class="{ 'premium-button--active': activePageId === page.pageId }"
        @click="activePageId = page.pageId"
      >
        {{ page.title || page.pageId }}
      </button>
    </div>

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
          class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-3 md:p-4 mb-8"
        >
          <div class="flex items-center justify-center mb-6 pb-4 border-b border-slate-800/50">
            <h3 class="text-xl font-bold uppercase tracking-wider text-slate-200 text-center">
              {{ group.pageTitle }}
            </h3>
          </div>
          
          <!-- FAQ 条目 -->
          <div class="rounded-xl overflow-hidden bg-slate-900/40 border border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
            <div 
              v-for="item in group.items" 
              :key="item.id"
              class="border-b border-slate-800/50 last:border-b-0"
            >
              <button
                type="button"
                class="w-full flex items-center gap-3 px-3 py-3 text-left transition-colors group hover:bg-white/5"
                :class="{ 'bg-white/5': expandedItems.has(item.id) }"
                @click="toggleItem(item.id)"
              >
                <span class="flex-shrink-0 px-2.5 py-1 rounded-full bg-slate-800 text-slate-300 text-[10px] uppercase font-bold tracking-wider border border-slate-700">
                  {{ item.category }}
                </span>
                <span 
                  class="flex-1 text-sm font-semibold text-slate-200 group-hover:text-sky-400 transition-colors"
                   :class="{ 'text-sky-400': expandedItems.has(item.id) }"
                >
                  {{ item.question }}
                </span>
                <span 
                  class="flex-shrink-0 w-5 h-5 flex items-center justify-center text-slate-500 transition-transform duration-200"
                  :class="{ 'rotate-180 text-sky-400': expandedItems.has(item.id) }"
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
                enter-to-class="opacity-100 max-h-[500px]"
                leave-from-class="opacity-100 max-h-[500px]"
                leave-to-class="opacity-0 max-h-0"
              >
                <div v-if="expandedItems.has(item.id)" class="overflow-hidden bg-slate-950/30">
                  <div class="px-4 pb-4 pt-1 text-sm leading-relaxed text-slate-400" v-html="item.answer" />
                </div>
              </Transition>
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
          class="inline-flex items-center gap-2 px-8 py-3 rounded-full text-sm font-bold bg-slate-800 text-slate-200 hover:bg-slate-700 hover:text-white hover:shadow-lg transition-all"
          @click="loadMoreGroups"
        >
          View More Content
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 无结果 -->
    <div v-else class="faqs-empty">
      <p>No FAQs found matching "{{ searchQuery }}"</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useHead, useAsyncData } from '#imports'
import { fetchAllFaqData, getAllFaqData } from '~/data/faq'

definePageMeta({
  layout: 'support',
})

useHead({
  title: "All FAQs",
})

// 获取所有 FAQ 数据 (从 Go API 获取)
const { data: asyncAllPages } = await useAsyncData('faqs-all', () => fetchAllFaqData())
const allPages = computed(() => asyncAllPages.value || getAllFaqData())

// 搜索和筛选
const searchQuery = ref('')
const activePageId = ref<string>('all')
const expandedItems = ref<Set<string>>(new Set())

// Pagination state
const visibleGroupsCount = ref(3)

// Watchers to reset pagination
watch(activePageId, () => {
  if (activePageId.value === 'all') {
    visibleGroupsCount.value = 3
  } else {
    // Single page view usually not paginated, show all
    visibleGroupsCount.value = 999 
  }
})

watch(searchQuery, () => {
  if (searchQuery.value) {
    // Search results should show all matches
    visibleGroupsCount.value = 999
  } else if (activePageId.value === 'all') {
     visibleGroupsCount.value = 3
  }
})

// 切换展开状态
const toggleItem = (itemId: string) => {
  if (expandedItems.value.has(itemId)) {
    expandedItems.value.delete(itemId)
  } else {
    expandedItems.value.add(itemId)
  }
  expandedItems.value = new Set(expandedItems.value)
}

// 扁平化所有 FAQ 条目
interface FlatFaqItem {
  id: string
  pageId: string
  pageTitle: string
  category: string
  question: string
  answer: string
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
          tags: item.tags,
        })
      }
    }
  }
  return items
})

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
const groupedItems = computed(() => {
  const groups: Record<string, { pageId: string; pageTitle: string; items: FlatFaqItem[] }> = {}
  
  for (const item of filteredItems.value) {
    if (!groups[item.pageId]) {
      groups[item.pageId] = {
        pageId: item.pageId,
        pageTitle: item.pageTitle,
        items: [],
      }
    }
    const group = groups[item.pageId]
    if (group) {
      group.items.push(item)
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
</script>

<style scoped>
.faqs-page {
  width: 100%; /* Use full available width (parent handles padding) */
  max-width: 960px;
  margin: 0 auto;
  padding: 0;
}

.faqs-page__title {
  margin: 0 0 0.5rem;
  font-size: 1.5rem;
  font-weight: 700;
  color: #f1f5f9;
  text-align: center;
}

.faqs-page__intro {
  margin: 0 auto 1.5rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
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
  background: rgba(30, 41, 59, 0.5);
  color: #e2e8f0;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.faqs-search__input::placeholder {
  color: rgba(148, 163, 184, 0.5);
}

.faqs-search__input:focus {
  outline: none;
  border-color: rgba(45, 212, 191, 0.5);
  background: rgba(30, 41, 59, 0.8);
  box-shadow: 0 0 0 4px rgba(45, 212, 191, 0.1);
}

.faqs-search__clear {
  position: absolute;
  right: 1.25rem;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(148, 163, 184, 0.6);
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



/* FAQ Item inner styles */
:deep(.faqs-item__answer ul),
:deep(.faqs-item__answer ol) {
  padding-left: 1.25rem;
  margin: 0.5rem 0;
}

:deep(.faqs-item__answer li) {
  margin: 0.25rem 0;
}

:deep(.faqs-item__answer strong) {
  color: #e2e8f0;
}

:deep(.faqs-item__answer a) {
  color: #38bdf8;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.faqs-empty {
  text-align: center;
  padding: 4rem 1rem;
  color: rgba(148, 163, 184, 0.6);
  font-size: 1rem;
  background: rgba(30, 41, 59, 0.3);
  border-radius: 1rem;
  border: 1px dashed rgba(148, 163, 184, 0.2);
}
</style>
