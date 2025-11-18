<template>
  <div class="flex gap-4 h-full w-full justify-center items-stretch overflow-hidden max-md:flex-col max-md:gap-0">
    <!-- 桌面端：第一栏 - 分类导航 -->
    <div class="flex-1 h-full min-w-0 max-w-full flex flex-col border-r border-white/[0.08] p-4 max-md:hidden">
      <!-- 社交图标 - 固定在顶部 -->
      <div class="mb-4 pb-4 border-b border-white/10">
        <SocialIcons :items="demoSocialLinks" />
      </div>
      
      <!-- 第一栏内容（待添加分类树） - 可滚动区域 -->
      <div class="flex-1 overflow-y-auto">
        <div class="text-sm text-white text-center p-4">
          {{ $t('sidebar.categoriesPlaceholder', '商品分类树（待实现）') }}
        </div>
      </div>
    </div>
    
    <!-- 桌面端：第二栏 - 筛选器 + 其他内容 -->
    <div class="sidebar-column-2 flex-1 h-full min-w-0 max-w-full flex flex-col items-center justify-start gap-3 overflow-y-auto p-4 max-md:hidden">
      <!-- 商品搜索 -->
      <div class="w-full">
        <h3 class="text-sm font-semibold text-white mb-3">{{ $t('sidebar.productSearch', '商品搜索') }}</h3>
        <div class="flex flex-col gap-2">
          <!-- 商品名称搜索 -->
          <input 
            class="w-full h-9 px-3 py-2 border border-white/20 rounded-lg bg-white/[0.05] text-white text-[13px] box-border transition-all duration-200 placeholder:text-white/50 focus:outline-none focus:border-[#6b73ff] focus:bg-white/[0.08]" 
            type="text" 
            :placeholder="$t('sidebar.searchProductPlaceholder', '输入商品名称...')" 
            v-model="productSearchQuery"
          >
          <!-- 搜索商品按钮 -->
          <button 
            class="w-full h-9 px-4 py-2 border-none rounded-lg bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black text-[13px] font-semibold cursor-pointer box-border transition-all duration-200 hover:shadow-[0_0_20px_rgba(107,115,255,0.5)] hover:-translate-y-0.5" 
            type="button" 
            @click="searchProducts"
          >
            {{ $t('sidebar.searchProducts', '搜索商品') }}
          </button>
        </div>
      </div>

      <!-- 分隔线 -->
      <div class="w-full border-t border-white/10 my-2"></div>

      <!-- 高级筛选器 -->
      <div class="w-full">
        <h3 class="text-sm font-semibold text-white mb-3">{{ $t('filter.title', 'Advanced Filters') }}</h3>
        <AdvancedFilter
          v-model:filters="filters"
          :options="{
            showPriceRange: true,
            showStockFilter: true,
            showSortBy: true,
            showRating: false,
            showResetButton: true,
            priceMin: 0,
            priceMax: 10000
          }"
          @update:filters="handleFilterChange"
          @reset="handleReset"
        />
      </div>

    </div>

    <!-- 移动端：分页显示 -->
    <div class="hidden min-w-0 max-w-full overflow-hidden max-md:flex max-md:flex-col max-md:w-full max-md:h-full">
      <!-- 分页标签 -->
      <div class="page-tabs flex gap-2 p-2 flex-shrink-0">
        <button
          @click="currentPage = 1"
          class="flex-1 py-2.5 px-4 text-sm font-semibold transition-all rounded-lg border-none cursor-pointer outline-none"
          :class="currentPage === 1 ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black shadow-[0_0_15px_rgba(107,115,255,0.4)]' : 'bg-white/5 text-white hover:bg-white/10'"
        >
          {{ $t('sidebar.page1', 'Category') }}
        </button>
        <button
          @click="currentPage = 2"
          class="flex-1 py-2.5 px-4 text-sm font-semibold transition-all rounded-lg border-none cursor-pointer outline-none"
          :class="currentPage === 2 ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black shadow-[0_0_15px_rgba(107,115,255,0.4)]' : 'bg-white/5 text-white hover:bg-white/10'"
        >
          {{ $t('sidebar.page2', 'Filter & More') }}
        </button>
      </div>

      <!-- 分页内容容器 -->
      <div class="page-content flex-1 min-h-0 max-h-full overflow-y-auto">
        <!-- 第 1 页：分类导航 -->
        <div v-show="currentPage === 1" class="p-4 h-full min-w-0 max-w-full">
          <div class="flex flex-col gap-4 h-full">
            <!-- 社交图标 -->
            <div class="social-icons-section">
              <SocialIcons :items="demoSocialLinks" />
            </div>
            
            <!-- 商品分类（占位，后续实现） -->
            <div class="product-categories">
              <h3 class="text-xs font-semibold text-white mb-2 uppercase tracking-wide">{{ $t('sidebar.categories', 'PRODUCT CATEGORIES') }}</h3>
              <div class="text-sm text-white p-4 text-center border border-white/10 rounded-lg">
                {{ $t('sidebar.categoriesPlaceholder', '商品分类树（待实现）') }}
              </div>
            </div>
          </div>
        </div>

        <!-- 第 2 页：筛选器 + 其他内容 -->
        <div v-show="currentPage === 2" class="p-4 h-full min-w-0 max-w-full">
          <div class="flex flex-col gap-4">
            <!-- 商品搜索 -->
            <div class="product-search-section">
              <h3 class="text-sm font-semibold text-white mb-3">{{ $t('sidebar.productSearch', 'Product Search') }}</h3>
              <div class="flex flex-col gap-2">
                <!-- 商品名称搜索 -->
                <input 
                  class="w-full h-9 px-3 py-2 border border-white/20 rounded-lg bg-white/[0.05] text-white text-[13px] box-border transition-all duration-200 placeholder:text-white/50 focus:outline-none focus:border-[#6b73ff] focus:bg-white/[0.08]" 
                  type="text" 
                  :placeholder="$t('sidebar.searchProductPlaceholder', '输入商品名称...')" 
                  v-model="productSearchQuery"
                >
                <!-- 搜索商品按钮 -->
                <button 
                  class="w-full h-9 px-4 py-2 border-none rounded-lg bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black text-[13px] font-semibold cursor-pointer box-border transition-all duration-200 hover:shadow-[0_0_20px_rgba(107,115,255,0.5)] hover:-translate-y-0.5" 
                  type="button" 
                  @click="searchProducts"
                >
                  {{ $t('sidebar.searchProducts', '搜索商品') }}
                </button>
              </div>
            </div>

            <!-- 分隔线 -->
            <div class="border-t border-white/10"></div>

            <!-- 高级筛选器 -->
            <div class="filter-section">
              <h3 class="text-sm font-semibold text-white/40 mb-3">{{ $t('filter.title', 'Advanced Filters') }}</h3>
              <AdvancedFilter
                v-model:filters="filters"
                :options="{
                  showPriceRange: true,
                  showStockFilter: true,
                  showSortBy: true,
                  showRating: false,
                  showResetButton: true,
                  priceMin: 0,
                  priceMax: 10000
                }"
                :compact="true"
                @update:filters="handleFilterChange"
                @reset="handleReset"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject } from 'vue'
import SocialIcons from './SocialIcons.vue'

const productSearchQuery = ref('')
const router = useRouter()

// 演示社交链接数据（即使没有 URL 也显示图标）
const demoSocialLinks = ref([
  { url: '#', label: 'Facebook', network: 'facebook' },
  { url: '#', label: 'Instagram', network: 'instagram' },
  { url: '#', label: 'X', network: 'x' },
  { url: '#', label: 'Reddit', network: 'reddit' },
  { url: '#', label: 'YouTube', network: 'youtube' },
  { url: '#', label: 'Pinterest', network: 'pinterest' }
])

// 移动端分页状态
const currentPage = ref(1)

// 获取父组件 SidePanel 暴露的方法
const sidePanel = inject('sidePanel', null)

// 高级筛选器状态
const filters = ref({
  priceRange: [0, 5000],
  inStock: true,
  preOrder: false,
  sortBy: 'newest',
  minRating: 0
})

// 筛选条件变化处理
const handleFilterChange = (newFilters) => {
  console.log('筛选条件变化:', newFilters)
  filters.value = newFilters
  // TODO: 根据筛选条件更新商品列表
}

// 重置筛选
const handleReset = () => {
  console.log('重置筛选')
  // TODO: 重置商品列表
}

// 搜索商品
const searchingProducts = ref(false)
const searchProducts = async () => {
  if (searchingProducts.value) return
  searchingProducts.value = true
  
  const query = productSearchQuery.value.trim()
  
  console.log('搜索商品:', query || '(无关键词，仅使用筛选条件)')
  console.log('当前筛选条件:', filters.value)
  
  // 打开右侧面板显示搜索结果
  if (sidePanel && typeof sidePanel.openRight === 'function') {
    sidePanel.openRight()
  }
  
  // TODO: 在右侧面板中显示商品搜索结果
  // 可以通过 provide/inject 传递搜索参数给右侧面板
  // 搜索参数包括：
  // - query: 商品名称（可为空）
  // - priceRange: [min, max]
  // - inStock: boolean
  // - sortBy: string
  
  setTimeout(() => { searchingProducts.value = false }, 360)
}
</script>

<style scoped>
/* 自定义滚动条样式（Tailwind 不支持） */
.page-content::-webkit-scrollbar,
.sidebar-column-2::-webkit-scrollbar {
  width: 6px;
}

.page-content::-webkit-scrollbar-track,
.sidebar-column-2::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.page-content::-webkit-scrollbar-thumb,
.sidebar-column-2::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.page-content::-webkit-scrollbar-thumb:hover,
.sidebar-column-2::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* 移动端触摸滚动优化 */
@media (max-width: 768px) {
  .page-content {
    -webkit-overflow-scrolling: touch;
  }
}
</style>
