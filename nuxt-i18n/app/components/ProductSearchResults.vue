<template>
  <div class="product-search-results w-full h-full flex flex-col gap-4">
    <!-- 搜索结果标题 -->
    <div class="search-header mb-4">
      <h2 class="text-2xl font-bold text-white mb-2">
        {{ $t('products.searchResults', '搜索结果') }}
      </h2>
      <div class="text-sm text-white/60">
        <span v-if="searchQuery">
          {{ $t('products.searchFor', '搜索') }}: <span class="text-[#40ffaa] font-semibold">"{{ searchQuery }}"</span>
        </span>
        <span v-else>
          {{ $t('products.allProducts', '所有商品') }}
        </span>
        <span class="ml-2 text-white/40">
          ({{ products.length }} {{ $t('products.items', '件') }})
        </span>
      </div>
    </div>

    <!-- 商品列表容器 -->
    <div class="products-container flex-1 overflow-y-auto">
      <!-- 加载中 -->
      <div v-if="loading" class="flex items-center justify-center h-full">
        <div class="text-white/60">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-[#6b73ff] mx-auto mb-4"></div>
          <p>{{ $t('products.loading', '加载中...') }}</p>
        </div>
      </div>

      <!-- 商品网格 -->
      <div v-else-if="products.length > 0" class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <!-- 商品卡片占位 -->
        <div 
          v-for="product in products" 
          :key="product.id"
          class="product-card p-3 rounded-lg border border-white/10 bg-white/5 hover:bg-white/10 transition-all cursor-pointer"
        >
          <div class="aspect-square bg-white/5 rounded-lg mb-2 flex items-center justify-center">
            <span class="text-2xl">📦</span>
          </div>
          <h3 class="text-xs font-semibold text-white mb-1 truncate">{{ product.name }}</h3>
          <p class="text-[10px] text-white/60 mb-2 line-clamp-2">{{ product.description }}</p>
          <div class="flex items-center justify-between gap-1">
            <span class="text-sm font-bold text-[#40ffaa]">${{ product.price }}</span>
            <span v-if="product.inStock" class="text-[10px] text-green-400">{{ $t('filter.inStock') }}</span>
            <span v-else class="text-[10px] text-orange-400">{{ $t('filter.preOrder') }}</span>
          </div>
        </div>
      </div>

      <!-- 无结果 -->
      <div v-else class="flex flex-col items-center justify-center h-full text-white/60">
        <span class="text-6xl mb-4">🔍</span>
        <p class="text-lg mb-2">{{ $t('products.noResults', '未找到商品') }}</p>
        <p class="text-sm">{{ $t('products.tryAdjustFilters', '请尝试调整筛选条件') }}</p>
      </div>
    </div>

    <!-- 加载更多按钮 -->
    <div v-if="hasMore && !loading" class="load-more">
      <button 
        class="w-full h-10 px-4 py-2 border border-white/20 rounded-lg bg-white/5 text-white text-sm font-medium cursor-pointer transition-all duration-200 hover:bg-white/10 hover:border-[#6b73ff]"
        @click="loadMore"
      >
        {{ $t('products.loadMore', '加载更多') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  searchQuery: {
    type: String,
    default: ''
  },
  filters: {
    type: Object,
    default: () => ({
      priceRange: [0, 5000],
      inStock: true,
      preOrder: false,
      sortBy: 'newest',
      minRating: 0
    })
  }
})

// 商品数据（模拟）
const products = ref([])
const loading = ref(true) // 初始状态为加载中
const hasMore = ref(true)

// 加载商品
const loadProducts = async () => {
  loading.value = true
  
  // 模拟 API 调用
  await new Promise(resolve => setTimeout(resolve, 1000))
  
  // 模拟商品数据
  const mockProducts = Array.from({ length: 6 }, (_, i) => ({
    id: i + 1,
    name: `Product ${i + 1}`,
    description: 'This is a product description',
    price: Math.floor(Math.random() * (props.filters.priceRange[1] - props.filters.priceRange[0]) + props.filters.priceRange[0]),
    inStock: Math.random() > 0.3
  }))
  
  products.value = mockProducts
  loading.value = false
}

// 加载更多
const loadMore = async () => {
  loading.value = true
  await new Promise(resolve => setTimeout(resolve, 800))
  
  const moreProducts = Array.from({ length: 4 }, (_, i) => ({
    id: products.value.length + i + 1,
    name: `Product ${products.value.length + i + 1}`,
    description: 'This is a product description',
    price: Math.floor(Math.random() * (props.filters.priceRange[1] - props.filters.priceRange[0]) + props.filters.priceRange[0]),
    inStock: Math.random() > 0.3
  }))
  
  products.value = [...products.value, ...moreProducts]
  loading.value = false
  
  // 模拟没有更多数据
  if (products.value.length >= 20) {
    hasMore.value = false
  }
}

// 监听筛选条件变化
watch(() => [props.searchQuery, props.filters], () => {
  loadProducts()
}, { immediate: true, deep: true })
</script>

<style scoped>
.products-container::-webkit-scrollbar {
  width: 6px;
}

.products-container::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.products-container::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.products-container::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
