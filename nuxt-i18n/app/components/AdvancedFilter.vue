<template>
  <div 
    class="advanced-filter"
    :class="[
      compact ? 'compact' : '',
      theme === 'light' ? 'theme-light' : 'theme-dark'
    ]"
  >
    <!-- 顶部行：价格范围 + 属性下拉（Color / Diameter / Brake） -->
    <div
      v-if="options.showPriceRange || attributeFilters.length"
      class="filter-section filter-top-row"
    >
      <!-- 价格范围 -->
      <div v-if="options.showPriceRange" class="price-range-inline">
        <h4 class="filter-label hidden md:block">
          {{ $t('filter.price', 'Price') }}
        </h4>
        <div class="price-range-container">
          <div class="price-inputs">
            <div class="price-input-wrapper">
              <span class="price-prefix">$</span>
              <input
                type="number"
                class="price-input"
                v-model.number="localFilters.priceRange[0]"
                :min="options.priceMin"
                :max="options.priceMax"
                step="1"
                @change="handlePriceChange"
                @blur="handlePriceChange"
              />
            </div>
            <span class="price-separator">-</span>
            <div class="price-input-wrapper">
              <span class="price-prefix">$</span>
              <input
                type="number"
                class="price-input"
                v-model.number="localFilters.priceRange[1]"
                :min="options.priceMin"
                :max="options.priceMax"
                step="1"
                @change="handlePriceChange"
                @blur="handlePriceChange"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 动态属性下拉（Color / Diameter / Brake） -->
      <div v-if="attributeFilters.length" class="attribute-top-row">
        <div
          v-for="attr in attributeFilters"
          :key="attr.id"
          class="attribute-inline-row"
        >
          <button
            type="button"
            class="attribute-toggle"
            @click="toggleGroupExpanded(getAttributeKey(attr))"
          >
            <span class="attribute-label">
              {{ attr.name }}
              <span v-if="getAttributeSummary(attr) !== 'All'" class="ml-0.5">({{ getAttributeSummary(attr) }})</span>
              <span v-else class="hidden md:inline ml-0.5">({{ getAttributeSummary(attr) }})</span>
            </span>
            <span
              class="attribute-toggle-icon"
              :class="{ open: isGroupExpanded(getAttributeKey(attr)) }"
            >
              ▼
            </span>
          </button>

          <transition name="attribute-dropdown">
            <div
              v-if="isGroupExpanded(getAttributeKey(attr))"
              class="attribute-dropdown"
            >
              <div class="checkbox-group">
                <label
                  v-for="value in attr.values"
                  :key="value.id"
                  class="checkbox-item"
                >
                  <input
                    type="checkbox"
                    class="checkbox-input"
                    :value="value.slug"
                    :checked="isAttributeSelected(getAttributeKey(attr), value.slug)"
                    @change="handleAttributeCheckboxChange(getAttributeKey(attr), value.slug, $event)"
                  />
                  <span class="checkbox-label">
                    <span
                      v-if="attr.type === 'color' && value.value"
                      class="inline-block w-4 h-4 rounded mr-2 align-middle border border-white/20"
                      :style="{ backgroundColor: value.value as string }"
                    />
                    {{ formatAttributeValueLabel(attr.slug, value.name) }}
                  </span>
                </label>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>

    <!-- 排序方式 -->
    <div v-if="options.showSortBy" class="filter-section">
      <h4 class="filter-label">
        {{ $t('filter.sortBy', 'Sort By') }}
      </h4>
      <select
        v-model="localFilters.sortBy"
        @change="handleFilterChange"
        class="sort-select"
      >
        <option
          v-for="option in sortOptions"
          :key="option.value"
          :value="option.value"
        >
          {{ $t(option.i18nKey, option.label) }}
        </option>
      </select>
    </div>

    <!-- 评分筛选 -->
    <div v-if="options.showRating" class="filter-section">
      <h4 class="filter-label">
        {{ $t('filter.rating', 'Rating') }}
      </h4>
      <div class="rating-group">
        <label
          v-for="rating in [5, 4, 3, 2, 1]"
          :key="rating"
          class="rating-item"
        >
          <input
            type="radio"
            :value="rating"
            v-model="localFilters.minRating"
            @change="handleFilterChange"
            class="rating-input"
          />
          <span class="rating-stars">
            <span v-for="i in 5" :key="i" class="star" :class="{ filled: i <= rating }">
              ⭐
            </span>
          </span>
          <span class="rating-text">{{ $t('filter.andUp', '& Up') }}</span>
        </label>
      </div>
    </div>

    <!-- 重置按钮 -->
    <div v-if="options.showResetButton" class="filter-actions">
      <button
        @click="handleReset"
        class="reset-button"
      >
        {{ $t('filter.reset', 'Reset Filters') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'

// 类型定义
interface FilterState {
  priceRange: [number, number]
  inStock: boolean
  preOrder: boolean
  sortBy: string
  minRating?: number
  attributes?: Record<string, string[]>
  [key: string]: any
}

interface FilterOptions {
  showPriceRange?: boolean
  showStockFilter?: boolean
  showSortBy?: boolean
  showRating?: boolean
  showResetButton?: boolean
  priceMin?: number
  priceMax?: number
  sortOptions?: Array<{ label: string; value: string; i18nKey: string }>
}

interface AttributeValueOption {
  id: number
  attribute_id: number
  name: string
  slug: string
  value?: string | null
  sort_order?: number
  is_enabled?: boolean
  meta?: Record<string, any>
}

interface AttributeFilterConfig {
  id: number
  name: string
  slug: string
  type: string
  values: AttributeValueOption[]
}

interface Props {
  // 初始筛选条件
  initialFilters?: Partial<FilterState>
  
  // 可用的筛选选项
  options?: FilterOptions
  
  // 动态属性筛选配置（例如颜色属性组）
  attributeFilters?: AttributeFilterConfig[]
  
  // 样式配置
  compact?: boolean
  theme?: 'dark' | 'light'
  
  // 防抖延迟（毫秒）
  debounceDelay?: number
}

// Props
const props = withDefaults(defineProps<Props>(), {
  initialFilters: () => ({
    priceRange: [0, 5000],
    inStock: true,
    preOrder: false,
    sortBy: 'newest',
    minRating: 0,
    attributes: {},
  }),
  options: () => ({
    showPriceRange: true,
    showStockFilter: true,
    showSortBy: true,
    showRating: false,
    showResetButton: true,
    priceMin: 0,
    priceMax: 10000,
    sortOptions: [
      { label: 'Newest', value: 'newest', i18nKey: 'filter.sort.newest' },
      { label: 'Price: Low to High', value: 'price_asc', i18nKey: 'filter.sort.priceLowToHigh' },
      { label: 'Price: High to Low', value: 'price_desc', i18nKey: 'filter.sort.priceHighToLow' },
      { label: 'Most Popular', value: 'popular', i18nKey: 'filter.sort.popular' },
      { label: 'Best Rating', value: 'rating', i18nKey: 'filter.sort.rating' }
    ]
  }),
  compact: false,
  theme: 'dark',
  debounceDelay: 300
})

// Emits
const emit = defineEmits<{
  'update:filters': [filters: FilterState]
  'reset': []
}>()

// 本地筛选状态
const localFilters = ref<FilterState>({
  priceRange: props.initialFilters?.priceRange || [props.options.priceMin || 0, props.options.priceMax || 5000],
  inStock: props.initialFilters?.inStock ?? true,
  preOrder: props.initialFilters?.preOrder ?? false,
  sortBy: props.initialFilters?.sortBy || 'newest',
  minRating: props.initialFilters?.minRating || 0,
  attributes: props.initialFilters?.attributes || {},
})

const attributeFilters = computed(() => props.attributeFilters || [])

// 为属性下拉展开状态生成稳定的 key（即使后端 slug 为空也可用）
const getAttributeKey = (attr: AttributeFilterConfig): string => {
  if (attr && typeof attr.slug === 'string' && attr.slug.length > 0) {
    return attr.slug
  }
  return `attr-${attr.id}`
}

// 属性下拉展开状态
const expandedGroups = ref<Record<string, boolean>>({})

const ensureExpandedDefaults = () => {
  const map: Record<string, boolean> = { ...expandedGroups.value }
  let changed = false

  attributeFilters.value.forEach((attr) => {
    const key = getAttributeKey(attr)
    if (!key) return
    if (typeof map[key] === 'undefined') {
      map[key] = false // 默认收起
      changed = true
    }
  })

  if (changed) {
    expandedGroups.value = map
  }
}

const isGroupExpanded = (key: string): boolean => {
  if (!key) return false
  return !!expandedGroups.value[key]
}

const toggleGroupExpanded = (key: string) => {
  if (!key) return
  expandedGroups.value = {
    ...expandedGroups.value,
    [key]: !expandedGroups.value[key],
  }
}

// 初始化属性筛选：默认每个属性组全选
const initializeAttributeSelections = (force = false): boolean => {
  const attrs = attributeFilters.value
  if (!attrs || !attrs.length) return false

  const currentAttributes: Record<string, string[]> = {
    ...(localFilters.value.attributes || {}),
  }

  let changed = false

  attrs.forEach((attr) => {
    const key = getAttributeKey(attr)
    if (!key) return

    const existing = currentAttributes[key]
    if (!force && Array.isArray(existing) && existing.length) {
      return
    }

    const allValues = (attr.values || [])
      .filter((v) => v && typeof v.slug === 'string' && v.slug.length > 0)
      .map((v) => v.slug)

    currentAttributes[key] = allValues
    changed = true
  })

  if (changed) {
    localFilters.value = {
      ...localFilters.value,
      attributes: currentAttributes,
    }
  }

  return changed
}

// 当属性配置加载或变化时，如果该组还没有显式选择，则默认全选
watch(
  attributeFilters,
  () => {
    ensureExpandedDefaults()
    const changed = initializeAttributeSelections(true)
    if (changed) {
      emit('update:filters', { ...localFilters.value })
    }
  },
  { immediate: true, deep: true },
)

// 排序选项
const sortOptions = computed(() => {
  return props.options.sortOptions || [
    { label: 'Newest', value: 'newest', i18nKey: 'filter.sort.newest' },
    { label: 'Price: Low to High', value: 'price_asc', i18nKey: 'filter.sort.priceLowToHigh' },
    { label: 'Price: High to Low', value: 'price_desc', i18nKey: 'filter.sort.priceHighToLow' },
    { label: 'Most Popular', value: 'popular', i18nKey: 'filter.sort.popular' }
  ]
})

// 滑块范围样式
const sliderRangeStyle = computed(() => {
  const min = props.options.priceMin || 0
  const max = props.options.priceMax || 10000
  const range = max - min
  
  const leftPercent = ((localFilters.value.priceRange[0] - min) / range) * 100
  const rightPercent = ((localFilters.value.priceRange[1] - min) / range) * 100
  
  return {
    left: `${leftPercent}%`,
    right: `${100 - rightPercent}%`
  }
})

// 防抖的筛选变化处理
const debouncedEmit = useDebounceFn((filters: FilterState) => {
  emit('update:filters', filters)
}, props.debounceDelay)

// 价格变化处理
const handlePriceChange = () => {
  // 确保最小值不大于最大值
  if (localFilters.value.priceRange[0] > localFilters.value.priceRange[1]) {
    const temp = localFilters.value.priceRange[0]
    localFilters.value.priceRange[0] = localFilters.value.priceRange[1]
    localFilters.value.priceRange[1] = temp
  }
  
  debouncedEmit({ ...localFilters.value })
}

// 筛选变化处理
const handleFilterChange = () => {
  emit('update:filters', { ...localFilters.value })
}

// 属性筛选：检查某个属性值是否已被选中
const isAttributeSelected = (attrSlug: string, valueSlug: string): boolean => {
  const attributes = localFilters.value.attributes || {}
  const selected = attributes[attrSlug] || []
  return selected.includes(valueSlug)
}

// 属性按钮上的选中汇总：All / 选中数量（0 时显示 0）
const getAttributeSummary = (attr: AttributeFilterConfig): string => {
  const key = getAttributeKey(attr)
  if (!key) return 'All'

  const values = (attr.values || []).filter((v) => v && v.is_enabled !== false)
  const total = values.length
  if (!total) return 'All'

  const attributes = localFilters.value.attributes || {}
  const selected = Array.isArray(attributes[key]) ? attributes[key] : []

  const selectedCount = values.reduce((count, v) => {
    return count + (selected.includes(v.slug) ? 1 : 0)
  }, 0)

  // 调试信息：在本地环境观察实际选中数量与总数（仅在浏览器端打印）
  if (typeof window !== 'undefined') {
    // eslint-disable-next-line no-console
    console.log('[AdvancedFilter:getAttributeSummary]', key, {
      total,
      selectedCount,
      selectedSlugs: selected,
    })
  }

  // 全选：显示 All
  if (selectedCount >= total && total > 0) return 'All'

  // 一个都没选：显示 0
  if (selectedCount <= 0) return '0'

  // 其它情况显示具体选中数量
  return String(selectedCount)
}

// 属性值显示文案格式化（例如直径组：12inch -> 12"）
const formatAttributeValueLabel = (attrSlug: string, valueName: string): string => {
  if (attrSlug === 'diameter' && typeof valueName === 'string') {
    return valueName.replace(/inch/gi, '"')
  }
  return valueName
}

// 属性筛选：切换选中状态
const toggleAttributeSelection = (attrSlug: string, valueSlug: string, checked: boolean) => {
  const currentAttributes = { ...(localFilters.value.attributes || {}) }
  const currentValues = Array.isArray(currentAttributes[attrSlug])
    ? [...currentAttributes[attrSlug]]
    : []

  const index = currentValues.indexOf(valueSlug)

  if (checked && index === -1) {
    currentValues.push(valueSlug)
  } else if (!checked && index !== -1) {
    currentValues.splice(index, 1)
  }

  currentAttributes[attrSlug] = currentValues

  localFilters.value = {
    ...localFilters.value,
    attributes: currentAttributes,
  }

  handleFilterChange()
}

const handleAttributeCheckboxChange = (attrSlug: string, valueSlug: string, event: Event) => {
  const target = event.target as HTMLInputElement | null
  toggleAttributeSelection(attrSlug, valueSlug, !!target?.checked)
}

// 重置筛选
const handleReset = () => {
  localFilters.value = {
    priceRange: [props.options.priceMin || 0, props.options.priceMax || 5000],
    inStock: true,
    preOrder: false,
    sortBy: 'newest',
    minRating: 0,
    attributes: {},
  }

  // 重置后默认每个属性组全选
  initializeAttributeSelections(true)

  emit('reset')
  emit('update:filters', { ...localFilters.value })
}

// 监听 initialFilters 变化
watch(() => props.initialFilters, (newFilters) => {
  if (newFilters) {
    localFilters.value = {
      ...localFilters.value,
      ...newFilters
    }
  }
}, { deep: true })
</script>

<style scoped>
.advanced-filter {
  width: 100%;
  box-sizing: border-box;
}

/* 筛选区块 */
.filter-section {
  margin-bottom: 0.5rem; /* 行与行之间再紧凑一点 */
}

.filter-section:last-of-type {
  margin-bottom: 0;
}

.filter-label {
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: rgba(255, 255, 255, 0.9);
}

/* 顶部行：价格 + 属性下拉 同一行展示（大屏），小屏自动换行 */
.filter-top-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 0.75rem 1rem;
}

.filter-top-row .price-range-inline {
  flex: 1 1 260px;
  min-width: 220px;
}

.attribute-top-row {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.75rem;
}

/* Price range 行内布局：标题 + 输入框一行展示 */
.price-range-inline {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.price-range-inline .filter-label {
  margin-bottom: 0;
  flex-shrink: 0;
}

.price-range-inline .price-range-container {
  flex: 1;
}

/* 紧凑模式 */
.compact .filter-section {
  margin-bottom: 0.75rem;
}

.compact .filter-label {
  font-size: 0.8125rem;
  margin-bottom: 0.375rem;
}

/* 浅色主题 */
.theme-light .filter-label {
  color: rgba(0, 0, 0, 0.9);
}

/* 价格范围 */
.price-range-container {
  padding: 0.25rem 0;
}

.price-inputs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.price-input-wrapper {
  display: flex;
  align-items: center;
  padding: 0 0.5rem;
  height: 2rem;
  border-radius: 0.5rem;
  background: linear-gradient(135deg, rgba(15,23,42,0.98), rgba(15,23,42,0.95));
  border: none;
  box-shadow:
    0 2px 5px -4px rgba(0,0,0,0.9),
    0 0 5px rgba(15,23,42,0.7);
}

.price-prefix {
  font-size: 0.75rem;
  color: #ffffff;
  margin-right: 0.15rem;
}

.price-input {
  width: 3rem;
  background: transparent;
  border: none;
  outline: none;
  font-size: 0.8125rem;
  color: #ffffff;
}

.price-separator {
  color: #ffffff;
}

/* Brake / Color / Diameter 组：标题 + 选项一行显示 */
.attribute-inline-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.25rem; /* 组与组之间再稍微缩小一点 */
  position: relative;
}

.attribute-inline-row .filter-label {
  margin-bottom: 0;
  flex-shrink: 0;
}

.attribute-inline-row .checkbox-group {
  flex: 1;
}

/* 属性下拉开关按钮 */
.attribute-toggle {
  border: none;
  background: linear-gradient(135deg, rgba(15,23,42,0.98), rgba(15,23,42,0.96));
  color: #ffffff;
  cursor: pointer;
  padding: 0.4rem 0.75rem;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border-radius: 0.5rem;
  font-size: 0.8125rem;
  transition: background-color 0.15s ease, color 0.15s ease, transform 0.1s ease;
  box-shadow:
    0 2px 6px -4px rgba(0,0,0,0.9),
    0 0 8px rgba(15,23,42,0.85);
}

.attribute-toggle-icon {
  display: inline-block;
  font-size: 0.65rem;
  color: #ffffff;
  transition: transform 0.15s ease;
}

.attribute-label {
  line-height: 1;
}

.attribute-toggle:hover {
  background: linear-gradient(135deg, rgba(31,41,55,0.98), rgba(15,23,42,0.96));
  color: #ffffff;
}

.attribute-toggle-icon.open {
  transform: rotate(180deg);
}

/* 属性下拉浮层 */
.attribute-dropdown {
  position: absolute;
  top: 100%;
  right: 0; /* 统一从容器右侧对齐，避免在右边按钮时向外溢出 */
  left: auto;
  margin-top: 0.25rem;
  z-index: 20;
  box-sizing: border-box;
  min-width: min(220px, 80vw);
  max-width: min(320px, 80vw);
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  background: rgba(15, 23, 42, 0.98);
  border: 1px solid rgba(148, 163, 184, 0.35);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.75);
}

/* 下拉动效 */
.attribute-dropdown-enter-active,
.attribute-dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.attribute-dropdown-enter-from,
.attribute-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* 滑块容器 */
.slider-container {
  position: relative;
  height: 2rem;
  display: flex;
  align-items: center;
}

.slider-track {
  position: absolute;
  width: 100%;
  height: 4px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  pointer-events: none;
  cursor: pointer;
  transition: all 0.2s;
}

/* 复选框组 */
.advanced-filter .checkbox-group {
  display: flex;
  flex-wrap: wrap;
  column-gap: 0.75rem; /* 横向间距稍大一点 */
  row-gap: 0.25rem;    /* 纵向间距保持较小，整体高度不会太高 */
}

.advanced-filter .checkbox-item {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  cursor: pointer;
  padding: 0.3rem 0.45rem;
  border-radius: 0.5rem;
  transition: background-color 0.2s;
}

.advanced-filter .checkbox-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.advanced-filter .checkbox-input {
  width: 1.125rem;
  height: 1.125rem;
  cursor: pointer;
  accent-color: #6b73ff;
}

.advanced-filter .checkbox-label {
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.7);
  user-select: none;
}

.advanced-filter.theme-light .checkbox-label {
  color: rgba(0, 0, 0, 0.7);
}

.reset-button:hover {
  background-color: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
  color: rgba(255, 255, 255, 0.9);
}

.reset-button:active {
  transform: scale(0.98);
}

.theme-light .reset-button {
  background-color: rgba(0, 0, 0, 0.05);
  border-color: rgba(0, 0, 0, 0.2);
  color: rgba(0, 0, 0, 0.7);
}

.theme-light .reset-button:hover {
  background-color: rgba(0, 0, 0, 0.1);
  border-color: rgba(0, 0, 0, 0.3);
  color: rgba(0, 0, 0, 0.9);
}

@media (max-width: 768px) {
	/* 在移动端让 Color / Diameter / Brake 等属性按钮自然流式排列，避免强制 50% 宽度导致溢出 */
	.attribute-top-row {
		flex-wrap: wrap;
		justify-content: flex-start;
		gap: 0.5rem;
	}

	.attribute-inline-row {
		flex: 0 1 auto; /* 自适应宽度 */
		width: auto;
	}

	/* 移动端：缩小价格输入框宽度，大约为桌面的约一半 */
	.price-input {
		width: 2.5rem;
	}
}
</style>
