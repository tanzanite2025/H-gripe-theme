# AdvancedFilter 组件使用示例

## 基础用法

### 1. 默认配置

```vue
<template>
  <div>
    <AdvancedFilter
      v-model:filters="filters"
      @update:filters="handleFilterChange"
    />
  </div>
</template>

<script setup>
const filters = ref({
  priceRange: [0, 5000],
  inStock: true,
  preOrder: false,
  sortBy: 'newest',
  minRating: 0
})

const handleFilterChange = (newFilters) => {
  console.log('筛选条件变化:', newFilters)
  // 根据筛选条件更新商品列表
}
</script>
```

---

### 2. 自定义配置

```vue
<template>
  <AdvancedFilter
    :initial-filters="{
      priceRange: [100, 2000],
      inStock: true,
      sortBy: 'price_asc'
    }"
    :options="{
      showPriceRange: true,
      showStockFilter: true,
      showSortBy: true,
      showRating: true,
      showResetButton: true,
      priceMin: 0,
      priceMax: 10000,
      sortOptions: [
        { label: '最新', value: 'newest', i18nKey: 'filter.sort.newest' },
        { label: '价格从低到高', value: 'price_asc', i18nKey: 'filter.sort.priceLowToHigh' },
        { label: '价格从高到低', value: 'price_desc', i18nKey: 'filter.sort.priceHighToLow' }
      ]
    }"
    @update:filters="handleFilterChange"
    @reset="handleReset"
  />
</template>

<script setup>
const handleFilterChange = (filters) => {
  console.log('筛选条件:', filters)
}

const handleReset = () => {
  console.log('重置筛选')
}
</script>
```

---

### 3. 紧凑模式

```vue
<template>
  <AdvancedFilter
    :compact="true"
    :options="{
      showPriceRange: true,
      showStockFilter: true,
      showSortBy: true,
      showRating: false,
      showResetButton: false
    }"
    @update:filters="handleFilterChange"
  />
</template>
```

---

### 4. 浅色主题

```vue
<template>
  <AdvancedFilter
    theme="light"
    @update:filters="handleFilterChange"
  />
</template>
```

---

### 5. 在 Sidebar 中使用

```vue
<template>
  <div class="sidebar-column-2">
    <!-- 高级筛选器 -->
    <div class="filter-section">
      <h3 class="text-sm font-semibold text-white/40 mb-3">高级筛选</h3>
      <AdvancedFilter
        v-model:filters="filters"
        :options="{
          showPriceRange: true,
          showStockFilter: true,
          showSortBy: true,
          showRating: true,
          priceMin: 0,
          priceMax: 10000
        }"
        @update:filters="handleFilterChange"
      />
    </div>

    <!-- 联系信息 -->
    <div class="contact-section mt-6">
      <ContactInfo />
    </div>

    <!-- 社交图标 -->
    <div class="social-section mt-6">
      <SocialIcons />
    </div>
  </div>
</template>

<script setup>
const filters = ref({
  priceRange: [0, 5000],
  inStock: true,
  preOrder: false,
  sortBy: 'newest',
  minRating: 0
})

const handleFilterChange = (newFilters) => {
  // 更新商品列表
  fetchProducts(newFilters)
}

const fetchProducts = async (filters) => {
  const { data } = await useFetch('/api/products', {
    params: {
      priceMin: filters.priceRange[0],
      priceMax: filters.priceRange[1],
      inStock: filters.inStock,
      sortBy: filters.sortBy,
      minRating: filters.minRating
    }
  })
}
</script>
```

---

### 6. 在搜索页面使用

```vue
<template>
  <div class="search-page">
    <div class="search-sidebar">
      <AdvancedFilter
        :compact="true"
        theme="light"
        :options="{
          showPriceRange: true,
          showStockFilter: true,
          showSortBy: true,
          showRating: true,
          priceMin: 0,
          priceMax: 20000
        }"
        @update:filters="handleSearch"
      />
    </div>

    <div class="search-results">
      <ProductGrid :products="filteredProducts" />
    </div>
  </div>
</template>

<script setup>
const filteredProducts = ref([])

const handleSearch = async (filters) => {
  const { data } = await useFetch('/api/search', {
    params: {
      ...filters,
      priceMin: filters.priceRange[0],
      priceMax: filters.priceRange[1]
    }
  })
  
  filteredProducts.value = data.value
}
</script>
```

---

## Props 详解

### initialFilters

初始筛选条件

```typescript
{
  priceRange: [number, number]  // 价格范围 [最小值, 最大值]
  inStock: boolean               // 是否有货
  preOrder: boolean              // 是否预售
  sortBy: string                 // 排序方式
  minRating: number              // 最低评分
}
```

### options

筛选器配置选项

```typescript
{
  showPriceRange: boolean        // 显示价格范围筛选
  showStockFilter: boolean       // 显示库存状态筛选
  showSortBy: boolean            // 显示排序选择
  showRating: boolean            // 显示评分筛选
  showResetButton: boolean       // 显示重置按钮
  priceMin: number               // 价格最小值
  priceMax: number               // 价格最大值
  sortOptions: Array<{           // 排序选项
    label: string                // 显示文本
    value: string                // 值
    i18nKey: string              // 国际化键
  }>
}
```

### compact

紧凑模式，减少间距和字体大小

```typescript
compact?: boolean  // 默认: false
```

### theme

主题颜色

```typescript
theme?: 'dark' | 'light'  // 默认: 'dark'
```

### debounceDelay

防抖延迟（毫秒），用于价格滑块

```typescript
debounceDelay?: number  // 默认: 300
```

---

## Events

### update:filters

筛选条件变化时触发

```typescript
(filters: FilterState) => void
```

### reset

重置筛选时触发

```typescript
() => void
```

---

## 国际化支持

组件使用 `$t()` 进行国际化，需要在语言文件中添加以下键：

```json
{
  "filter": {
    "priceRange": "价格范围",
    "stockStatus": "库存状态",
    "sortBy": "排序方式",
    "rating": "评分",
    "inStock": "有货",
    "preOrder": "预售",
    "andUp": "及以上",
    "reset": "重置筛选",
    "sort": {
      "newest": "最新",
      "priceLowToHigh": "价格从低到高",
      "priceHighToLow": "价格从高到低",
      "popular": "最受欢迎",
      "rating": "评分最高"
    }
  }
}
```

---

## 样式自定义

### CSS 变量

```css
.advanced-filter {
  --filter-primary-color: #6b73ff;
  --filter-secondary-color: #40ffaa;
  --filter-bg-color: rgba(255, 255, 255, 0.05);
  --filter-border-color: rgba(255, 255, 255, 0.1);
  --filter-text-color: rgba(255, 255, 255, 0.9);
}
```

### 覆盖样式

```vue
<style>
.advanced-filter .filter-label {
  color: #your-color;
  font-size: 1rem;
}

.advanced-filter .slider::-webkit-slider-thumb {
  background: #your-color;
}
</style>
```

---

## 完整示例：商店页面

```vue
<template>
  <div class="shop-page">
    <!-- Sidebar -->
    <div class="sidebar">
      <!-- 第一栏：分类导航 -->
      <div class="sidebar-column-1">
        <CategoryTree @category-click="handleCategoryClick" />
      </div>

      <!-- 第二栏：筛选器 -->
      <div class="sidebar-column-2">
        <AdvancedFilter
          v-model:filters="filters"
          :options="filterOptions"
          @update:filters="handleFilterChange"
          @reset="handleReset"
        />
        
        <ContactInfo class="mt-6" />
        <SocialIcons class="mt-6" />
      </div>
    </div>

    <!-- 右侧弹窗 -->
    <SidePanel
      :open="sidePanelOpen"
      :category="selectedCategory"
      :filters="filters"
      @close="sidePanelOpen = false"
    >
      <ProductList :products="filteredProducts" />
    </SidePanel>
  </div>
</template>

<script setup>
const filters = ref({
  priceRange: [0, 5000],
  inStock: true,
  preOrder: false,
  sortBy: 'newest',
  minRating: 0
})

const filterOptions = {
  showPriceRange: true,
  showStockFilter: true,
  showSortBy: true,
  showRating: true,
  showResetButton: true,
  priceMin: 0,
  priceMax: 10000,
  sortOptions: [
    { label: 'Newest', value: 'newest', i18nKey: 'filter.sort.newest' },
    { label: 'Price: Low to High', value: 'price_asc', i18nKey: 'filter.sort.priceLowToHigh' },
    { label: 'Price: High to Low', value: 'price_desc', i18nKey: 'filter.sort.priceHighToLow' },
    { label: 'Most Popular', value: 'popular', i18nKey: 'filter.sort.popular' }
  ]
}

const sidePanelOpen = ref(false)
const selectedCategory = ref(null)
const filteredProducts = ref([])

const handleCategoryClick = (category) => {
  selectedCategory.value = category
  sidePanelOpen.value = true
  fetchProducts()
}

const handleFilterChange = (newFilters) => {
  filters.value = newFilters
  fetchProducts()
}

const handleReset = () => {
  fetchProducts()
}

const fetchProducts = async () => {
  const { data } = await useFetch('/api/products', {
    params: {
      category: selectedCategory.value?.slug,
      priceMin: filters.value.priceRange[0],
      priceMax: filters.value.priceRange[1],
      inStock: filters.value.inStock,
      preOrder: filters.value.preOrder,
      sortBy: filters.value.sortBy,
      minRating: filters.value.minRating
    }
  })
  
  filteredProducts.value = data.value
}
</script>
```

---

## 注意事项

1. **防抖处理**：价格滑块使用防抖，避免频繁触发 API 请求
2. **双向绑定**：使用 `v-model:filters` 实现双向绑定
3. **响应式**：组件完全响应式，适配桌面和移动端
4. **可复用**：可在任何页面使用，不依赖特定父组件
5. **国际化**：支持多语言，使用 `$t()` 函数
6. **主题**：支持深色和浅色主题
7. **性能**：使用 `@vueuse/core` 的 `useDebounceFn` 优化性能

---

**创建日期**: 2025-11-17  
**维护者**: Tanzanite Team
