<template>
  <div 
    class="advanced-filter"
    :class="[
      compact ? 'compact' : '',
      theme === 'light' ? 'theme-light' : 'theme-dark'
    ]"
  >
    <!-- 价格范围筛选 -->
    <div v-if="options.showPriceRange" class="filter-section">
      <h4 class="filter-label">
        {{ $t('filter.priceRange', 'Price Range') }}
      </h4>
      <div class="price-range-container">
        <!-- 价格显示 -->
        <div class="price-display">
          <span class="price-value">${{ localFilters.priceRange[0] }}</span>
          <span class="price-separator">-</span>
          <span class="price-value">${{ localFilters.priceRange[1] }}</span>
        </div>
        
        <!-- 双滑块 -->
        <div class="slider-container">
          <input
            type="range"
            :min="options.priceMin"
            :max="options.priceMax"
            v-model.number="localFilters.priceRange[0]"
            @input="handlePriceChange"
            class="slider slider-min"
          />
          <input
            type="range"
            :min="options.priceMin"
            :max="options.priceMax"
            v-model.number="localFilters.priceRange[1]"
            @input="handlePriceChange"
            class="slider slider-max"
          />
          <div class="slider-track">
            <div 
              class="slider-range"
              :style="sliderRangeStyle"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 库存状态筛选 -->
    <div v-if="options.showStockFilter" class="filter-section">
      <h4 class="filter-label">
        {{ $t('filter.stockStatus', 'Stock Status') }}
      </h4>
      <div class="checkbox-group">
        <label class="checkbox-item">
          <input
            type="checkbox"
            v-model="localFilters.inStock"
            @change="handleFilterChange"
            class="checkbox-input"
          />
          <span class="checkbox-label">
            {{ $t('filter.inStock', 'In Stock') }}
          </span>
        </label>
        <label class="checkbox-item">
          <input
            type="checkbox"
            v-model="localFilters.preOrder"
            @change="handleFilterChange"
            class="checkbox-input"
          />
          <span class="checkbox-label">
            {{ $t('filter.preOrder', 'Pre-order') }}
          </span>
        </label>
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

interface Props {
  // 初始筛选条件
  initialFilters?: Partial<FilterState>
  
  // 可用的筛选选项
  options?: FilterOptions
  
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
    minRating: 0
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
  minRating: props.initialFilters?.minRating || 0
})

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

// 重置筛选
const handleReset = () => {
  localFilters.value = {
    priceRange: [props.options.priceMin || 0, props.options.priceMax || 5000],
    inStock: true,
    preOrder: false,
    sortBy: 'newest',
    minRating: 0
  }
  
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
}

/* 筛选区块 */
.filter-section {
  margin-bottom: 1.5rem;
}

.filter-section:last-of-type {
  margin-bottom: 0;
}

.filter-label {
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: rgba(255, 255, 255, 0.9);
}

/* 紧凑模式 */
.compact .filter-section {
  margin-bottom: 1rem;
}

.compact .filter-label {
  font-size: 0.8125rem;
  margin-bottom: 0.5rem;
}

/* 浅色主题 */
.theme-light .filter-label {
  color: rgba(0, 0, 0, 0.9);
}

/* 价格范围 */
.price-range-container {
  padding: 0.5rem 0;
}

.price-display {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  font-size: 1rem;
  font-weight: 600;
  color: #40ffaa;
}

.price-separator {
  color: rgba(255, 255, 255, 0.4);
}

/* 滑块容器 */
.slider-container {
  position: relative;
  height: 2.5rem;
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
}

.slider-range {
  position: absolute;
  height: 100%;
  background: linear-gradient(to right, #40ffaa, #6b73ff);
  border-radius: 2px;
}

.slider {
  position: absolute;
  width: 100%;
  height: 4px;
  background: transparent;
  pointer-events: none;
  -webkit-appearance: none;
  appearance: none;
}

.slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #6b73ff;
  cursor: pointer;
  pointer-events: auto;
  border: 2px solid #000;
  box-shadow: 0 0 0 2px rgba(107, 115, 255, 0.2);
  transition: all 0.2s;
}

.slider::-webkit-slider-thumb:hover {
  transform: scale(1.2);
  box-shadow: 0 0 0 4px rgba(107, 115, 255, 0.3);
}

.slider::-moz-range-thumb {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #6b73ff;
  cursor: pointer;
  pointer-events: auto;
  border: 2px solid #000;
  box-shadow: 0 0 0 2px rgba(107, 115, 255, 0.2);
  transition: all 0.2s;
}

.slider::-moz-range-thumb:hover {
  transform: scale(1.2);
  box-shadow: 0 0 0 4px rgba(107, 115, 255, 0.3);
}

/* 复选框组 */
.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 0.5rem;
  transition: background-color 0.2s;
}

.checkbox-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.checkbox-input {
  width: 1.125rem;
  height: 1.125rem;
  cursor: pointer;
  accent-color: #6b73ff;
}

.checkbox-label {
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.7);
  user-select: none;
}

.theme-light .checkbox-label {
  color: rgba(0, 0, 0, 0.7);
}

/* 排序选择器 */
.sort-select {
  width: 100%;
  padding: 0.75rem;
  background-color: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.sort-select:hover {
  border-color: #6b73ff;
  background-color: rgba(255, 255, 255, 0.08);
}

.sort-select:focus {
  outline: none;
  border-color: #6b73ff;
  box-shadow: 0 0 0 3px rgba(107, 115, 255, 0.1);
}

.theme-light .sort-select {
  background-color: rgba(0, 0, 0, 0.05);
  border-color: rgba(0, 0, 0, 0.1);
  color: rgba(0, 0, 0, 0.9);
}

/* 评分组 */
.rating-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.rating-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 0.5rem;
  transition: background-color 0.2s;
}

.rating-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.rating-input {
  width: 1rem;
  height: 1rem;
  cursor: pointer;
  accent-color: #6b73ff;
}

.rating-stars {
  display: flex;
  gap: 0.125rem;
}

.star {
  font-size: 0.875rem;
  filter: grayscale(100%);
  opacity: 0.3;
}

.star.filled {
  filter: grayscale(0%);
  opacity: 1;
}

.rating-text {
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.7);
}

/* 重置按钮 */
.filter-actions {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.reset-button {
  width: 100%;
  padding: 0.75rem;
  background-color: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 0.5rem;
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
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
</style>
