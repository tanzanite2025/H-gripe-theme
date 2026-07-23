<template>
  <section class="home-faq" :class="{ 'home-faq--wide': wide, 'home-faq--fluid': fluid }">
    <!-- Premium Card Container -->
    <div class="rounded-2xl premium-card p-3 md:p-4">
      
      <div class="home-faq__header">
        <h2 class="home-faq__title tz-faq-title">Frequently Asked Questions</h2>
        <div class="mt-[6px] mb-[3px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
        <p class="home-faq__subtitle tz-faq-subtitle">Quick answers to common questions</p>
      </div>

      <!-- 分类标签 -->
      <div class="nav-pill-tabs">
        <button
          v-for="page in allPages"
          :key="page.pageId"
          type="button"
          class="nav-pill-item"
          :class="{ 'nav-pill-item--active': activePageId === page.pageId }"
          @click="activePageId = page.pageId"
        >
          {{ page.title || page.pageId }}
        </button>
      </div>

      <!-- FAQ 条目 Container -->
      <div class="home-faq__content rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] bg-slate-900/40 border border-slate-800/50">
        <div 
          v-for="(item, index) in displayItems" 
          :key="item.id"
          class="home-faq__item border-b border-slate-800/50 last:border-b-0"
        >
          <button
            type="button"
            class="home-faq__question group"
            @click="toggleItem(item.id)"
          >
            <span class="home-faq__question-text tz-faq-question group-hover:text-sky-400 transition-colors">{{ item.question }}</span>
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
            <div v-if="expandedItems.has(item.id)" class="home-faq__answer tz-faq-answer bg-slate-900/30" v-html="item.answer" />
          </Transition>
        </div>
      </div>

      <!-- 查看全部链接 -->
      <div class="home-faq__footer">
        <NuxtLink to="/support/faqs" target="_blank" class="home-faq__link">
          View All FAQs
          <svg class="home-faq__link-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </NuxtLink>
      </div>
      
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
  fluid?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxItemsPerCategory: 3,
  defaultCategory: '',
  wide: false,
  fluid: false,
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

.home-faq__title {
  margin: 0 0 0.35rem;
  color: var(--tz-text-primary);
}

.home-faq__subtitle {
  margin: 0;
  color: var(--tz-text-secondary);
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
  color: #2dd4bf;
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

.home-faq__link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1.5rem;
  border-radius: 9999px;
  font-size: 0.85rem;
  font-weight: 700;
  color: #ffffff;
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  text-decoration: none;
  transition: all 0.2s ease;
  letter-spacing: 0.025em;
}

.home-faq__link:hover {
  transform: translateY(-1px);
  box-shadow:
    0 8px 16px -4px rgba(59, 130, 246, 0.6),
    0 0 20px rgba(45, 212, 191, 0.4);
}

.home-faq__link-icon {
  width: 1rem;
  height: 1rem;
}
</style>
