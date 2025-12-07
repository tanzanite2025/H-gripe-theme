<template>
  <div class="w-full h-full flex flex-col gap-2 md:gap-4 p-2">
    <!-- 搜索区域（移动端隐藏） -->
    <div class="shrink-0 hidden md:block">
      <ProductSearchPanel />
    </div>

    <!-- 首页按钮 -->
    <NuxtLink 
      :to="localePath('/')"
      class="shrink-0 flex items-center justify-center gap-2 px-4 py-2 md:py-3 bg-gradient-to-r from-cyan-500/20 to-cyan-400/10 border border-cyan-500/50 rounded-xl text-white font-semibold text-sm hover:from-cyan-500/30 hover:to-cyan-400/20 hover:border-cyan-400 hover:shadow-[0_0_15px_rgba(6,182,212,0.3)] transition-all"
      @click="closeSidebar"
    >
      <svg class="w-4 h-4 md:w-5 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
      Home
    </NuxtLink>

    <!-- Shop 按钮 -->
    <NuxtLink 
      :to="localePath('/shop')"
      class="shrink-0 flex items-center justify-center gap-2 px-4 py-2 md:py-3 bg-gradient-to-r from-cyan-500/20 to-cyan-400/10 border border-cyan-500/50 rounded-xl text-white font-semibold text-sm hover:from-cyan-500/30 hover:to-cyan-400/20 hover:border-cyan-400 hover:shadow-[0_0_15px_rgba(6,182,212,0.3)] transition-all"
      @click="closeSidebar"
    >
      <svg class="w-4 h-4 md:w-5 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
      </svg>
      Shop
    </NuxtLink>

    <!-- Picture Warehouse 按钮 -->
    <NuxtLink 
      :to="localePath('/picture-warehouse')"
      class="shrink-0 flex items-center justify-center gap-2 px-4 py-2 md:py-3 bg-gradient-to-r from-cyan-500/20 to-cyan-400/10 border border-cyan-500/50 rounded-xl text-white font-semibold text-sm hover:from-cyan-500/30 hover:to-cyan-400/20 hover:border-cyan-400 hover:shadow-[0_0_15px_rgba(6,182,212,0.3)] transition-all"
      @click="closeSidebar"
    >
      <svg class="w-4 h-4 md:w-5 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>
      Picture
    </NuxtLink>

    <!-- 信任卡片区 -->
    <div class="shrink-0">
      <TrustCards layout="grid" size="sm" />
    </div>

    <!-- 分隔线 -->
    <div class="border-t border-white/10"></div>

    <!-- 手风琴区域 -->
    <div class="flex flex-col gap-2.5 flex-1 overflow-y-auto">
      <!-- Support -->
      <div 
        class="bg-black/50 rounded-xl overflow-hidden transition-all duration-200 shadow-[inset_0_1px_0_0_rgba(255,255,255,0.6)]"
        :class="activeSections.has('support') ? 'shadow-[inset_0_1px_0_0_rgba(255,255,255,0.1),0_0_20px_rgba(16,185,129,0.15)]' : 'hover:shadow-[inset_0_1px_0_0_rgba(255,255,255,0.1),0_4px_12px_rgba(0,0,0,0.3)]'"
      >
        <button 
          type="button"
          class="w-full flex items-center justify-between px-4 py-3.5 cursor-pointer select-none hover:bg-white/[0.02] transition-colors"
          @click="toggleSection('support')"
        >
          <div class="flex items-center gap-2.5 text-sm font-semibold text-white/90">
            {{ $t('footer.menus.support', 'Support') }}
            <span class="inline-flex items-center justify-center min-w-5 h-5 px-1.5 text-[11px] font-bold text-black bg-gradient-to-br from-emerald-400 to-cyan-400 rounded-full shadow-sm">
              {{ supportLinks.length }}
            </span>
          </div>
          <div 
            class="w-6 h-6 flex items-center justify-center text-[10px] rounded-md transition-all duration-200 shadow-[0_0_8px_rgba(255,255,255,0.5)]"
            :class="activeSections.has('support') ? 'rotate-180 bg-gradient-to-br from-emerald-400 to-cyan-400 text-black shadow-[0_0_12px_rgba(16,185,129,0.8)]' : 'bg-emerald-500/20 text-emerald-400'"
          >▼</div>
        </button>
        <div 
          class="overflow-hidden transition-all duration-300"
          :class="activeSections.has('support') ? 'max-h-96' : 'max-h-0'"
        >
          <div class="px-4 pb-4 pt-1">
            <div class="grid grid-cols-2 gap-2">
              <NuxtLink
                v-for="item in supportLinks"
                :key="item.id"
                :to="localePath(item.to)"
                class="flex items-center justify-center px-3 py-2.5 text-xs font-semibold text-white bg-black/50 rounded-full cursor-pointer transition-all shadow-[0_0_0_1px_rgba(255,255,255,0.3)] hover:text-white hover:-translate-y-0.5 hover:shadow-[0_0_0_1px_rgba(255,255,255,0.5),0_0_12px_rgba(255,255,255,0.3)] active:translate-y-0 no-underline"
                @click="closeSidebar"
              >
                {{ $t(item.labelKey, item.fallback) }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <!-- Brand -->
      <div 
        class="bg-black/50 rounded-xl overflow-hidden transition-all duration-200 shadow-[inset_0_1px_0_0_rgba(255,255,255,0.6)]"
        :class="activeSections.has('brand') ? 'shadow-[inset_0_1px_0_0_rgba(255,255,255,0.1),0_0_20px_rgba(16,185,129,0.15)]' : 'hover:shadow-[inset_0_1px_0_0_rgba(255,255,255,0.1),0_4px_12px_rgba(0,0,0,0.3)]'"
      >
        <button 
          type="button"
          class="w-full flex items-center justify-between px-4 py-3.5 cursor-pointer select-none hover:bg-white/[0.02] transition-colors"
          @click="toggleSection('brand')"
        >
          <div class="flex items-center gap-2.5 text-sm font-semibold text-white/90">
            Brand
            <span class="inline-flex items-center justify-center min-w-5 h-5 px-1.5 text-[11px] font-bold text-black bg-gradient-to-br from-emerald-400 to-cyan-400 rounded-full shadow-sm">
              {{ brandCategories.length }}
            </span>
          </div>
          <div 
            class="w-6 h-6 flex items-center justify-center text-[10px] rounded-md transition-all duration-200 shadow-[0_0_8px_rgba(255,255,255,0.5)]"
            :class="activeSections.has('brand') ? 'rotate-180 bg-gradient-to-br from-emerald-400 to-cyan-400 text-black shadow-[0_0_12px_rgba(16,185,129,0.8)]' : 'bg-emerald-500/20 text-emerald-400'"
          >▼</div>
        </button>
        <div 
          class="overflow-hidden transition-all duration-300"
          :class="activeSections.has('brand') ? 'max-h-96' : 'max-h-0'"
        >
          <div class="px-4 pb-4 pt-1">
            <div class="grid grid-cols-2 gap-2">
              <button
                v-for="brand in brandCategories"
                :key="brand.id"
                type="button"
                class="flex items-center justify-center px-3 py-2.5 text-xs font-semibold text-white bg-black/50 rounded-full cursor-pointer transition-all shadow-[0_0_0_1px_rgba(255,255,255,0.3)] hover:text-white hover:-translate-y-0.5 hover:shadow-[0_0_0_1px_rgba(255,255,255,0.5),0_0_12px_rgba(255,255,255,0.3)] active:translate-y-0"
                @click="handleBrandClick(brand.id)"
              >
                {{ brand.label }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Guides -->
      <div 
        class="bg-black/50 rounded-xl overflow-hidden transition-all duration-200 shadow-[inset_0_1px_0_0_rgba(255,255,255,0.6)]"
        :class="activeSections.has('guides') ? 'shadow-[inset_0_1px_0_0_rgba(255,255,255,0.1),0_0_20px_rgba(16,185,129,0.15)]' : 'hover:shadow-[inset_0_1px_0_0_rgba(255,255,255,0.1),0_4px_12px_rgba(0,0,0,0.3)]'"
      >
        <button 
          type="button"
          class="w-full flex items-center justify-between px-4 py-3.5 cursor-pointer select-none hover:bg-white/[0.02] transition-colors"
          @click="toggleSection('guides')"
        >
          <div class="flex items-center gap-2.5 text-sm font-semibold text-white/90">
            Guides
            <span class="inline-flex items-center justify-center min-w-5 h-5 px-1.5 text-[11px] font-bold text-black bg-gradient-to-br from-emerald-400 to-cyan-400 rounded-full shadow-sm">
              {{ guidesLinks.length + guidesNavLinks.length }}
            </span>
          </div>
          <div 
            class="w-6 h-6 flex items-center justify-center text-[10px] rounded-md transition-all duration-200 shadow-[0_0_8px_rgba(255,255,255,0.5)]"
            :class="activeSections.has('guides') ? 'rotate-180 bg-gradient-to-br from-emerald-400 to-cyan-400 text-black shadow-[0_0_12px_rgba(16,185,129,0.8)]' : 'bg-emerald-500/20 text-emerald-400'"
          >▼</div>
        </button>
        <div 
          class="overflow-hidden transition-all duration-300"
          :class="activeSections.has('guides') ? 'max-h-96' : 'max-h-0'"
        >
          <div class="px-4 pb-4 pt-1">
            <div class="grid grid-cols-2 gap-2">
              <NuxtLink
                v-for="item in guidesLinks"
                :key="item.id"
                :to="localePath(item.to)"
                class="flex items-center justify-center px-3 py-2.5 text-xs font-semibold text-white bg-black/50 rounded-full cursor-pointer transition-all shadow-[0_0_0_1px_rgba(255,255,255,0.3)] hover:text-white hover:-translate-y-0.5 hover:shadow-[0_0_0_1px_rgba(255,255,255,0.5),0_0_12px_rgba(255,255,255,0.3)] active:translate-y-0 no-underline"
                @click="closeSidebar"
              >
                {{ $t(item.labelKey, item.fallback) }}
              </NuxtLink>
              <NuxtLink
                v-for="item in guidesNavLinks"
                :key="item.id"
                :to="localePath(item.to)"
                class="flex items-center justify-center px-3 py-2.5 text-xs font-semibold text-white bg-black/50 rounded-full cursor-pointer transition-all shadow-[0_0_0_1px_rgba(255,255,255,0.3)] hover:text-white hover:-translate-y-0.5 hover:shadow-[0_0_0_1px_rgba(255,255,255,0.5),0_0_12px_rgba(255,255,255,0.3)] active:translate-y-0 no-underline"
                @click="closeSidebar"
              >
                {{ item.label }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject } from 'vue'
import { useLocalePath } from '#imports'
import ProductSearchPanel from '~/components/ProductSearchPanel.vue'

const localePath = useLocalePath()

// 手风琴状态（支持多个同时展开）
const activeSections = ref<Set<string>>(new Set(['support']))

const toggleSection = (section: string) => {
  if (activeSections.value.has(section)) {
    activeSections.value.delete(section)
  } else {
    activeSections.value.add(section)
  }
  // 触发响应式更新
  activeSections.value = new Set(activeSections.value)
}

// 获取侧边栏控制方法
const sidePanel = inject<{ closeLeft?: () => void }>('sidePanel', {})

const closeSidebar = () => {
  sidePanel.closeLeft?.()
}

// 支持快捷链接
const supportLinks = [
  { id: 'warranty-check', labelKey: 'support.nav.warrantyCheck', to: '/support/warranty-check', fallback: 'Warranty Check' },
  { id: 'spoke-calculator', labelKey: 'support.nav.spokeCalculator', to: '/spoke-calculator', fallback: 'Spoke Calculator' },
  { id: 'test-report', labelKey: 'support.nav.testReport', to: '/support/test-report', fallback: 'Test Report' },
  { id: 'faqs', labelKey: 'support.nav.faqs', to: '/support/faqs', fallback: 'All FAQs' },
]

// 品牌按钮
const brandCategories = [
  { id: 'sapim', label: 'Sapim' },
  { id: 'dt-swiss', label: 'DT Swiss' },
  { id: 'pillar', label: 'Pillar' },
]

// 品牌点击处理（暂时为空，等商品准备好后实现）
const handleBrandClick = (brandId: string) => {
  // TODO: 实现品牌跳转逻辑
  console.log('Brand clicked:', brandId)
}

// Guides 快捷链接
const guidesLinks = [
  { id: 'wheelset-buyers', labelKey: 'products.nav.wheelsetBuyersGuide', to: '/guides/wheelset-buyers', fallback: 'Wheelset guide' },
  { id: 'tire-size-charts', labelKey: 'products.nav.tireSizeCharts', to: '/guides/sizecharts', fallback: 'Tire Size' },
]

// Guides 快捷链接（带跳转逻辑）
const guidesNavLinks = [
  { id: 'technical', label: 'Technical', to: '/guides/technical' },
  { id: 'our-story', label: 'Our Story', to: '/company/about' },
  { id: 'membership', label: 'Membership', to: '/company/membershipandpoints' },
]
</script>

