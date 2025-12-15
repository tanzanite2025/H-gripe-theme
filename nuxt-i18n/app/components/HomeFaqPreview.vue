<template>
  <section class="home-faq" :class="{ 'home-faq--wide': wide }">
    <div class="home-faq__header">
      <h2 class="home-faq__title">Frequently Asked Questions</h2>
      <div class="mt-[6px] mb-[3px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
      <p class="home-faq__subtitle">Quick answers to common questions</p>
    </div>

    <!-- 分类标签 -->
    <div class="home-faq__tabs">
      <button
        v-for="page in allPages"
        :key="page.pageId"
        type="button"
        class="home-faq__tab"
        :class="{ 'home-faq__tab--active': activePageId === page.pageId }"
        @click="activePageId = page.pageId"
      >
        {{ page.title || page.pageId }}
      </button>
    </div>

    <!-- FAQ 条目 -->
    <div class="home-faq__content">
      <div 
        v-for="item in displayItems" 
        :key="item.id"
        class="home-faq__item"
      >
        <button
          type="button"
          class="home-faq__question"
          @click="toggleItem(item.id)"
        >
          <span class="home-faq__question-text">{{ item.question }}</span>
          <svg 
            class="home-faq__icon"
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
          <div v-if="expandedItems.has(item.id)" class="home-faq__answer" v-html="item.answer" />
        </Transition>
      </div>
    </div>

    <!-- 查看全部链接 - 在新窗口打开 -->
    <div class="home-faq__footer">
      <NuxtLink to="/support/faqs" target="_blank" class="home-faq__link">
        View All FAQs
        <svg class="home-faq__link-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </NuxtLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getAllFaqData } from '~/data/faq'

interface Props {
  maxItemsPerCategory?: number
  defaultCategory?: string
  wide?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxItemsPerCategory: 3,
  defaultCategory: '',
  wide: false,
})

const wide = computed(() => props.wide)

// 获取所有 FAQ 数据
const allPages = computed(() => getAllFaqData())

// 当前选中的分类
const activePageId = ref<string>('')

// 展开的条目
const expandedItems = ref<Set<string>>(new Set())

// 初始化默认分类
onMounted(() => {
  if (props.defaultCategory) {
    activePageId.value = props.defaultCategory
  } else if (allPages.value.length > 0) {
    const firstPage = allPages.value[0]
    if (firstPage) {
      activePageId.value = firstPage.pageId
    }
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

// 扁平化并限制条目数量
interface FlatItem {
  id: string
  question: string
  answer: string
}

const displayItems = computed<FlatItem[]>(() => {
  const currentPage = allPages.value.find(p => p.pageId === activePageId.value)
  if (!currentPage) return []

  const items: FlatItem[] = []
  let count = 0

  for (const category of currentPage.categories) {
    for (const item of category.items) {
      if (count >= props.maxItemsPerCategory) break
      items.push({
        id: `${currentPage.pageId}-${item.id}`,
        question: item.question,
        answer: item.answer,
      })
      count++
    }
    if (count >= props.maxItemsPerCategory) break
  }

  return items
})
</script>

<style scoped>
.home-faq {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.home-faq--wide {
  max-width: 1152px;
  padding: 2rem 0.25rem;
}

@media (min-width: 640px) {
  .home-faq--wide {
    padding: 2rem 1rem;
  }
}

.home-faq__header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.home-faq__title {
  margin: 0 0 0.35rem;
  font-size: 1.25rem;
  font-weight: 600;
  color: #f9fafb;
}

.home-faq__subtitle {
  margin: 0;
  font-size: 0.875rem;
  color: rgba(148, 163, 184, 0.8);
}

.home-faq__tabs {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.35rem;
  margin-bottom: 1rem;
}

.home-faq__tab {
  border: none;
  border-radius: 9999px;
  padding: 0.35rem 0.9rem;
  font-size: 0.78rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
}

.home-faq__tab:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.home-faq__tab--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #ffffff;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

.home-faq__content {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 0.75rem;
  overflow: hidden;
}

.home-faq__item {
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.home-faq__item:last-child {
  border-bottom: none;
}

.home-faq__question {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: transparent;
  border: none;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s;
}

.home-faq__question:hover {
  background: rgba(255, 255, 255, 0.05);
}

.home-faq__question-text {
  flex: 1;
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.9);
}

.home-faq__icon {
  flex-shrink: 0;
  width: 1rem;
  height: 1rem;
  color: rgba(255, 255, 255, 0.5);
  transition: transform 0.2s;
}

.home-faq__icon--open {
  transform: rotate(180deg);
}

.home-faq__answer {
  padding: 0 1rem 1rem 1rem;
  font-size: 0.8rem;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.6);
  overflow: hidden;
}

.home-faq__answer :deep(ul),
.home-faq__answer :deep(ol) {
  padding-left: 1.25rem;
  margin: 0.35rem 0;
}

.home-faq__answer :deep(li) {
  margin: 0.2rem 0;
}

.home-faq__answer :deep(strong) {
  color: rgba(255, 255, 255, 0.8);
}

.home-faq__footer {
  text-align: center;
  margin-top: 1rem;
}

.home-faq__link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.5rem 1.25rem;
  border-radius: 9999px;
  font-size: 0.8rem;
  font-weight: 600;
  color: #ffffff;
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
  text-decoration: none;
  transition: all 0.2s ease;
}

.home-faq__link:hover {
  transform: translateY(-1px);
  box-shadow:
    0 6px 18px -10px rgba(59, 130, 246, 0.9),
    0 0 18px rgba(45, 212, 191, 0.7);
}

.home-faq__link-icon {
  width: 0.875rem;
  height: 0.875rem;
}
</style>
