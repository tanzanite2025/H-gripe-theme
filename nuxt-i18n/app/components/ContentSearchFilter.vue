<template>
  <div 
    class="content-search-filter"
    :class="[
      compact ? 'compact' : '',
      theme === 'light' ? 'theme-light' : 'theme-dark'
    ]"
  >
    <!-- 搜索框 -->
    <div class="filter-section">
      <div class="search-container">
        <div class="search-input-wrapper">
          <svg class="search-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            v-model="localFilters.searchText"
            @input="handleSearchChange"
            :placeholder="$t('filter.searchPlaceholder', 'Search...')"
            class="search-input"
          />
          <button
            v-if="localFilters.searchText"
            @click="clearSearch"
            class="clear-button"
            :aria-label="$t('filter.clearSearch', 'Clear search')"
          >
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 分类筛选 -->
    <div v-if="options.showCategories" class="filter-section">
      <h4 class="filter-label">
        {{ $t('filter.categories', 'Categories') }}
      </h4>
      <div class="category-list">
        <button
          v-for="category in categories"
          :key="category.id"
          @click="toggleCategory(category.id)"
          class="category-item"
          :class="{ active: localFilters.selectedCategories.includes(category.id) }"
        >
          <span class="category-name">{{ category.name }}</span>
          <span v-if="category.count" class="category-count">{{ category.count }}</span>
        </button>
      </div>
    </div>

    <!-- 标签筛选 -->
    <div v-if="options.showTags && tags.length > 0" class="filter-section">
      <h4 class="filter-label">
        {{ $t('filter.tags', 'Tags') }}
      </h4>
      <div class="tag-list">
        <button
          v-for="tag in tags"
          :key="tag.id"
          @click="toggleTag(tag.id)"
          class="tag-item"
          :class="{ active: localFilters.selectedTags.includes(tag.id) }"
        >
          {{ tag.name }}
        </button>
      </div>
    </div>

    <!-- 排序选项 -->
    <div v-if="options.showSortBy" class="filter-section">
      <h4 class="filter-label">
        {{ $t('filter.sortBy', 'Sort By') }}
      </h4>
      <div class="sort-options">
        <button
          v-for="option in sortOptions"
          :key="option.value"
          @click="handleSortChange(option.value)"
          class="sort-option"
          :class="{ active: localFilters.sortBy === option.value }"
        >
          {{ option.label }}
        </button>
      </div>
    </div>

    <!-- 日期范围筛选（可选，用于 Blog） -->
    <div v-if="options.showDateRange" class="filter-section">
      <h4 class="filter-label">
        {{ $t('filter.dateRange', 'Date Range') }}
      </h4>
      <div class="date-range-container">
        <input
          type="date"
          v-model="localFilters.dateFrom"
          @change="handleDateChange"
          class="date-input"
          :placeholder="$t('filter.from', 'From')"
        />
        <span class="date-separator">-</span>
        <input
          type="date"
          v-model="localFilters.dateTo"
          @change="handleDateChange"
          class="date-input"
          :placeholder="$t('filter.to', 'To')"
        />
      </div>
    </div>

    <!-- 重置按钮 - 移动端隐藏 -->
    <div class="filter-actions hidden md:block">
      <button
        @click="resetFilters"
        class="reset-button"
        :disabled="!hasActiveFilters"
      >
        {{ $t('filter.reset', 'Reset Filters') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

// Props
interface Props {
  // 筛选选项配置
  options?: {
    showCategories?: boolean
    showTags?: boolean
    showSortBy?: boolean
    showDateRange?: boolean
  }
  // 分类列表
  categories?: Array<{
    id: string | number
    name: string
    count?: number
  }>
  // 标签列表
  tags?: Array<{
    id: string | number
    name: string
  }>
  // 排序选项
  sortOptions?: Array<{
    value: string
    label: string
  }>
  // 紧凑模式
  compact?: boolean
  // 主题
  theme?: 'light' | 'dark'
  // 初始筛选值
  initialFilters?: {
    searchText?: string
    selectedCategories?: Array<string | number>
    selectedTags?: Array<string | number>
    sortBy?: string
    dateFrom?: string
    dateTo?: string
  }
}

const props = withDefaults(defineProps<Props>(), {
  options: () => ({
    showCategories: true,
    showTags: true,
    showSortBy: true,
    showDateRange: false
  }),
  categories: () => [],
  tags: () => [],
  sortOptions: () => [
    { value: 'relevance', label: 'Relevance' },
    { value: 'newest', label: 'Newest' },
    { value: 'popular', label: 'Most Popular' }
  ],
  compact: false,
  theme: 'dark',
  initialFilters: () => ({
    searchText: '',
    selectedCategories: [],
    selectedTags: [],
    sortBy: 'relevance',
    dateFrom: '',
    dateTo: ''
  })
})

// Emits
const emit = defineEmits<{
  (e: 'update:filters', filters: any): void
  (e: 'search', searchText: string): void
}>()

// 本地筛选状态
const localFilters = ref({
  searchText: props.initialFilters?.searchText || '',
  selectedCategories: props.initialFilters?.selectedCategories || [],
  selectedTags: props.initialFilters?.selectedTags || [],
  sortBy: props.initialFilters?.sortBy || 'relevance',
  dateFrom: props.initialFilters?.dateFrom || '',
  dateTo: props.initialFilters?.dateTo || ''
})

// 是否有激活的筛选
const hasActiveFilters = computed(() => {
  return (
    localFilters.value.searchText !== '' ||
    localFilters.value.selectedCategories.length > 0 ||
    localFilters.value.selectedTags.length > 0 ||
    localFilters.value.dateFrom !== '' ||
    localFilters.value.dateTo !== ''
  )
})

// 搜索变化处理
const handleSearchChange = () => {
  emit('search', localFilters.value.searchText)
  emitFilters()
}

// 清除搜索
const clearSearch = () => {
  localFilters.value.searchText = ''
  handleSearchChange()
}

// 切换分类
const toggleCategory = (categoryId: string | number) => {
  const index = localFilters.value.selectedCategories.indexOf(categoryId)
  if (index > -1) {
    localFilters.value.selectedCategories.splice(index, 1)
  } else {
    localFilters.value.selectedCategories.push(categoryId)
  }
  emitFilters()
}

// 切换标签
const toggleTag = (tagId: string | number) => {
  const index = localFilters.value.selectedTags.indexOf(tagId)
  if (index > -1) {
    localFilters.value.selectedTags.splice(index, 1)
  } else {
    localFilters.value.selectedTags.push(tagId)
  }
  emitFilters()
}

// 排序变化
const handleSortChange = (sortValue: string) => {
  localFilters.value.sortBy = sortValue
  emitFilters()
}

// 日期变化
const handleDateChange = () => {
  emitFilters()
}

// 重置筛选
const resetFilters = () => {
  localFilters.value = {
    searchText: '',
    selectedCategories: [],
    selectedTags: [],
    sortBy: 'relevance',
    dateFrom: '',
    dateTo: ''
  }
  emitFilters()
}

// 发送筛选更新
const emitFilters = () => {
  emit('update:filters', { ...localFilters.value })
}

// 监听初始筛选值变化
watch(() => props.initialFilters, (newVal) => {
  if (newVal) {
    localFilters.value = {
      searchText: newVal.searchText || '',
      selectedCategories: newVal.selectedCategories || [],
      selectedTags: newVal.selectedTags || [],
      sortBy: newVal.sortBy || 'relevance',
      dateFrom: newVal.dateFrom || '',
      dateTo: newVal.dateTo || ''
    }
  }
}, { deep: true })
</script>

<style scoped>
.content-search-filter {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 1rem;
}

.content-search-filter.compact {
  gap: 0.5rem;
  padding: 0.25rem 0.75rem;
}

@media (min-width: 768px) {
  .content-search-filter.compact {
    gap: 1rem;
    padding: 0.75rem;
  }
}

/* 主题样式 */
.content-search-filter.theme-dark {
  --text-primary: rgba(255, 255, 255, 0.9);
  --text-secondary: rgba(255, 255, 255, 0.6);
  --bg-primary: rgba(255, 255, 255, 0.05);
  --bg-hover: rgba(255, 255, 255, 0.1);
  --bg-active: linear-gradient(to right, #40ffaa, #6b73ff);
  --border-color: rgba(255, 255, 255, 0.2);
}

.content-search-filter.theme-light {
  --text-primary: rgba(0, 0, 0, 0.9);
  --text-secondary: rgba(0, 0, 0, 0.6);
  --bg-primary: rgba(0, 0, 0, 0.05);
  --bg-hover: rgba(0, 0, 0, 0.1);
  --bg-active: linear-gradient(to right, #40ffaa, #6b73ff);
  --border-color: rgba(0, 0, 0, 0.2);
}

/* 筛选区块 */
.filter-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.filter-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

/* 搜索框 */
.search-container {
  width: 100%;
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  width: 1.25rem;
  height: 1.25rem;
  color: var(--text-secondary);
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 0.625rem 2.5rem 0.625rem 2.5rem;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 9999px;
  color: var(--text-primary);
  font-size: 0.875rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #6b73ff;
  background: var(--bg-hover);
}

.search-input::placeholder {
  color: var(--text-secondary);
}

.clear-button {
  position: absolute;
  right: 0.75rem;
  width: 1.25rem;
  height: 1.25rem;
  padding: 0;
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  transition: color 0.2s;
}

.clear-button:hover {
  color: var(--text-primary);
}

.clear-button svg {
  width: 100%;
  height: 100%;
}

/* 分类列表 */
.category-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5rem;
}

@media (min-width: 768px) {
  .category-list {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
  }
}

.category-item {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 0.875rem;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  color: var(--text-primary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

@media (min-width: 768px) {
  .category-item {
    flex: 0 0 auto;
  }
}

.category-item:hover {
  background: var(--bg-hover);
}

.category-item.active {
  background: var(--bg-active);
  border-color: transparent;
  color: #000;
  font-weight: 600;
}

.category-name {
  flex: 1;
}

.category-count {
  font-size: 0.75rem;
  opacity: 0.7;
}

/* 标签列表 */
.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.tag-item {
  padding: 0.375rem 0.875rem;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 9999px;
  color: var(--text-primary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.tag-item:hover {
  background: var(--bg-hover);
}

.tag-item.active {
  background: var(--bg-active);
  border-color: transparent;
  color: #000;
  font-weight: 600;
}

/* 排序选项 */
.sort-options {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.sort-option {
  padding: 0.625rem 0.875rem;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  color: var(--text-primary);
  font-size: 0.875rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  flex: 0 0 auto;
  white-space: nowrap;
}

.sort-option:hover {
  background: var(--bg-hover);
}

.sort-option.active {
  background: var(--bg-active);
  border-color: transparent;
  color: #000;
  font-weight: 600;
}

/* 日期范围 */
.date-range-container {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.date-input {
  flex: 1;
  padding: 0.625rem 0.75rem;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  color: var(--text-primary);
  font-size: 0.875rem;
  transition: all 0.2s;
}

.date-input:focus {
  outline: none;
  border-color: #6b73ff;
  background: var(--bg-hover);
}

.date-separator {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

/* 操作按钮 */
.filter-actions {
  gap: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border-color);
}

@media (min-width: 768px) {
  .filter-actions {
    display: flex;
  }
}

.reset-button {
  flex: 1;
  padding: 0.625rem 1rem;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  color: var(--text-primary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.reset-button:hover:not(:disabled) {
  background: var(--bg-hover);
}

.reset-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
