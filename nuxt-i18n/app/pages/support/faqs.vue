<template>
  <div class="faqs-page">
    <h2 class="faqs-page__title">All FAQs</h2>
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
        class="faqs-tabs__item"
        :class="{ 'faqs-tabs__item--active': activePageId === 'all' }"
        @click="activePageId = 'all'"
      >
        All
      </button>
      <button
        v-for="page in allPages"
        :key="page.pageId"
        type="button"
        class="faqs-tabs__item"
        :class="{ 'faqs-tabs__item--active': activePageId === page.pageId }"
        @click="activePageId = page.pageId"
      >
        {{ page.title || page.pageId }}
      </button>
    </div>

    <!-- FAQ 内容 -->
    <div v-if="filteredItems.length > 0" class="faqs-content">
      <!-- 按页面分组显示 -->
      <div 
        v-for="group in groupedItems" 
        :key="group.pageId"
        class="faqs-group"
      >
        <h3 class="faqs-group__title">
          {{ group.pageTitle }}
        </h3>
        
        <!-- FAQ 条目 -->
        <div class="faqs-items">
          <div 
            v-for="item in group.items" 
            :key="item.id"
            class="faqs-item"
          >
            <button
              type="button"
              class="faqs-item__question"
              @click="toggleItem(item.id)"
            >
              <span class="faqs-item__category">{{ item.category }}</span>
              <span class="faqs-item__text">{{ item.question }}</span>
              <svg 
                class="faqs-item__icon"
                :class="{ 'faqs-item__icon--open': expandedItems.has(item.id) }"
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
              <div v-if="expandedItems.has(item.id)" class="faqs-item__answer" v-html="item.answer" />
            </Transition>
          </div>
        </div>
      </div>
    </div>

    <!-- 无结果 -->
    <div v-else class="faqs-empty">
      <p>No FAQs found matching "{{ searchQuery }}"</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { getAllFaqData } from '~/data/faq'

definePageMeta({
  layout: 'support',
})

useHead({
  title: "All FAQs",
})

// 获取所有 FAQ 数据
const allPages = computed(() => getAllFaqData())

// 搜索和筛选
const searchQuery = ref('')
const activePageId = ref<string>('all')
const expandedItems = ref<Set<string>>(new Set())

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
</script>

<style scoped>
.faqs-page {
  max-width: 900px;
}

.faqs-page__title {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  font-weight: 600;
  color: #f9fafb;
}

.faqs-page__intro {
  margin: 0 0 1rem;
  font-size: 0.875rem;
  color: rgba(148, 163, 184, 0.9);
}

.faqs-search {
  position: relative;
  margin-bottom: 1rem;
}

.faqs-search__input {
  width: 100%;
  padding: 0.6rem 2.5rem 0.6rem 1rem;
  border-radius: 9999px;
  border: 1px solid rgba(148, 163, 184, 0.4);
  background: rgba(15, 23, 42, 0.85);
  color: #e5e7eb;
  font-size: 0.875rem;
}

.faqs-search__input::placeholder {
  color: rgba(148, 163, 184, 0.6);
}

.faqs-search__input:focus {
  outline: none;
  border-color: rgba(56, 189, 248, 0.8);
}

.faqs-search__clear {
  position: absolute;
  right: 1rem;
  top: 50%;
  transform: translateY(-50%);
  color: rgba(148, 163, 184, 0.6);
  cursor: pointer;
  font-size: 0.75rem;
}

.faqs-search__clear:hover {
  color: #e5e7eb;
}

.faqs-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-bottom: 1rem;
}

.faqs-tabs__item {
  border: none;
  border-radius: 9999px;
  padding: 0.25rem 0.65rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.9);
  background: rgba(15, 23, 42, 0.85);
  border: 1px solid rgba(148, 163, 184, 0.3);
  cursor: pointer;
  transition: all 0.15s;
}

.faqs-tabs__item:hover {
  border-color: rgba(148, 163, 184, 0.6);
}

.faqs-tabs__item--active {
  background: rgba(56, 189, 248, 0.15);
  color: #e5e7eb;
  border-color: rgba(56, 189, 248, 0.8);
}

.faqs-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.faqs-group__title {
  margin: 0 0 0.5rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: #e5e7eb;
}

.faqs-items {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 0.75rem;
  overflow: hidden;
}

.faqs-item {
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.faqs-item:last-child {
  border-bottom: none;
}

.faqs-item__question {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 0.75rem;
  background: transparent;
  border: none;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s;
}

.faqs-item__question:hover {
  background: rgba(255, 255, 255, 0.05);
}

.faqs-item__category {
  flex-shrink: 0;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  background: rgba(107, 115, 255, 0.2);
  color: #a5b4fc;
  font-size: 0.65rem;
  font-weight: 500;
}

.faqs-item__text {
  flex: 1;
  font-size: 0.8rem;
  color: rgba(255, 255, 255, 0.9);
}

.faqs-item__icon {
  flex-shrink: 0;
  width: 1rem;
  height: 1rem;
  color: rgba(255, 255, 255, 0.5);
  transition: transform 0.2s;
}

.faqs-item__icon--open {
  transform: rotate(180deg);
}

.faqs-item__answer {
  padding: 0 0.75rem 0.75rem 0.75rem;
  font-size: 0.8rem;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.6);
  overflow: hidden;
}

.faqs-item__answer :deep(ul),
.faqs-item__answer :deep(ol) {
  padding-left: 1.25rem;
  margin: 0.35rem 0;
}

.faqs-item__answer :deep(li) {
  margin: 0.2rem 0;
}

.faqs-item__answer :deep(strong) {
  color: rgba(255, 255, 255, 0.8);
}

.faqs-empty {
  text-align: center;
  padding: 2rem;
  color: rgba(148, 163, 184, 0.6);
  font-size: 0.875rem;
}
</style>
