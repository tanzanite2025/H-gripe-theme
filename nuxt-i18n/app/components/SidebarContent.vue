<template>
  <div class="w-full h-full flex flex-col gap-2 md:gap-3 p-2 md:p-3">
    <!-- 搜索区域（移动端隐藏） -->
    <div class="shrink-0 hidden md:block">
      <ProductSearchPanel />
    </div>

    <!-- 导航按钮组 - 分段控制样式 -->
    <div class="shrink-0 bg-white/[0.03] p-1 rounded-xl flex gap-1">
      <NuxtLink 
        :to="localePath('/')"
        class="flex-1 h-9 rounded-lg bg-gradient-to-r from-emerald-500/15 via-cyan-500/15 to-emerald-500/15 text-white text-xs font-medium flex items-center justify-center gap-1.5 shadow-[0_0_12px_rgba(64,255,170,0.1)]"
        @click="closeSidebar"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
        Home
      </NuxtLink>
      <NuxtLink 
        :to="localePath('/shop')"
        class="flex-1 h-9 rounded-lg text-white/90 text-xs font-medium flex items-center justify-center gap-1.5 hover:text-white hover:bg-white/[0.06] transition-all"
        @click="closeSidebar"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
        </svg>
        Shop
      </NuxtLink>
      <NuxtLink 
        :to="localePath('/company/picture-warehouse')"
        class="flex-1 h-9 rounded-lg text-white/90 text-xs font-medium flex items-center justify-center gap-1.5 hover:text-white hover:bg-white/[0.06] transition-all"
        @click="closeSidebar"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        Picture
      </NuxtLink>
    </div>

    <!-- 渐变分隔线 -->
    <div class="h-px bg-gradient-to-r from-transparent via-white/10 to-transparent"></div>

    <!-- Quick Access - 彩色横条卡片 -->
    <div class="shrink-0">
      <div class="text-[10px] text-white/40 uppercase tracking-wider mb-1.5 md:mb-2 px-0.5">Quick Access</div>
      <div class="space-y-0.5 md:space-y-1">
        <!-- Payment -->
        <NuxtLink 
          :to="localePath('/support/payment')"
          class="group relative overflow-hidden flex items-center gap-2.5 px-3 h-11 rounded-lg bg-gradient-to-r from-emerald-500/[0.35] to-emerald-500/[0.25] hover:from-emerald-500/[0.42] hover:to-emerald-500/[0.32] transition-all"
          @click="closeSidebar"
        >
          <div class="absolute right-0 top-1/2 -translate-y-1/2 w-20 h-20 bg-gradient-to-br from-emerald-400/20 to-transparent rounded-full blur-2xl opacity-50 group-hover:opacity-100 transition-opacity"></div>
          <svg class="w-4 h-4 text-emerald-400 shrink-0 relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
          </svg>
          <span class="text-xs text-white font-medium transition-colors relative z-10">Payment Methods</span>
          <svg class="w-3.5 h-3.5 text-emerald-500/40 group-hover:text-emerald-400 ml-auto transition-colors relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </NuxtLink>
        <!-- Shipping -->
        <NuxtLink 
          :to="localePath('/support/shipping')"
          class="group relative overflow-hidden flex items-center gap-2.5 px-3 h-11 rounded-lg bg-gradient-to-r from-blue-500/[0.35] to-blue-500/[0.25] hover:from-blue-500/[0.42] hover:to-blue-500/[0.32] transition-all"
          @click="closeSidebar"
        >
          <div class="absolute right-0 top-1/2 -translate-y-1/2 w-20 h-20 bg-gradient-to-br from-blue-400/20 to-transparent rounded-full blur-2xl opacity-50 group-hover:opacity-100 transition-opacity"></div>
          <svg class="w-4 h-4 text-blue-400 shrink-0 relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
          </svg>
          <span class="text-xs text-white font-medium transition-colors relative z-10">Shipping Info</span>
          <svg class="w-3.5 h-3.5 text-blue-500/40 group-hover:text-blue-400 ml-auto transition-colors relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </NuxtLink>
        <!-- Warranty -->
        <NuxtLink 
          :to="localePath('/support/warranty')"
          class="group relative overflow-hidden flex items-center gap-2.5 px-3 h-11 rounded-lg bg-gradient-to-r from-cyan-500/[0.35] to-cyan-500/[0.25] hover:from-cyan-500/[0.42] hover:to-cyan-500/[0.32] transition-all"
          @click="closeSidebar"
        >
          <div class="absolute right-0 top-1/2 -translate-y-1/2 w-20 h-20 bg-gradient-to-br from-cyan-400/20 to-transparent rounded-full blur-2xl opacity-50 group-hover:opacity-100 transition-opacity"></div>
          <svg class="w-4 h-4 text-cyan-400 shrink-0 relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <span class="text-xs text-white font-medium transition-colors relative z-10">Warranty Policy</span>
          <svg class="w-3.5 h-3.5 text-cyan-500/40 group-hover:text-cyan-400 ml-auto transition-colors relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </NuxtLink>
        <!-- After Sales -->
        <NuxtLink 
          :to="localePath('/support/after-sales')"
          class="group relative overflow-hidden flex items-center gap-2.5 px-3 h-11 rounded-lg bg-gradient-to-r from-purple-500/[0.35] to-purple-500/[0.25] hover:from-purple-500/[0.42] hover:to-purple-500/[0.32] transition-all"
          @click="closeSidebar"
        >
          <div class="absolute right-0 top-1/2 -translate-y-1/2 w-20 h-20 bg-gradient-to-br from-purple-400/20 to-transparent rounded-full blur-2xl opacity-50 group-hover:opacity-100 transition-opacity"></div>
          <svg class="w-4 h-4 text-purple-400 shrink-0 relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
          <span class="text-xs text-white font-medium transition-colors relative z-10">After Sales Support</span>
          <svg class="w-3.5 h-3.5 text-purple-500/40 group-hover:text-purple-400 ml-auto transition-colors relative z-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </NuxtLink>
      </div>
    </div>

    <!-- 渐变分隔线 -->
    <div class="h-px bg-gradient-to-r from-transparent via-white/10 to-transparent"></div>

    <!-- Support & Tools -->
    <div class="shrink-0">
      <div class="text-[10px] text-white/40 uppercase tracking-wider mb-1.5 md:mb-2 px-0.5">Support & Tools</div>
      <div class="flex flex-wrap gap-1 md:gap-1.5">
        <NuxtLink
          v-for="item in supportLinks"
          :key="item.id"
          :to="localePath(item.to)"
          class="px-3 h-9 rounded-lg bg-white/[0.08] text-[11px] font-medium text-white hover:bg-white/[0.15] transition-all flex items-center justify-center"
          @click="closeSidebar"
        >
          {{ $t(item.labelKey, item.fallback) }}
        </NuxtLink>
      </div>
    </div>

    <!-- 渐变分隔线 -->
    <div class="h-px bg-gradient-to-r from-transparent via-white/10 to-transparent"></div>

    <!-- Guides & Resources -->
    <div class="shrink-0">
      <div class="text-[10px] text-white/40 uppercase tracking-wider mb-1.5 md:mb-2 px-0.5">Guides & Resources</div>
      <div class="flex flex-wrap gap-1 md:gap-1.5">
        <NuxtLink
          v-for="item in guidesLinks"
          :key="item.id"
          :to="localePath(item.to)"
          class="px-3 h-9 rounded-lg bg-white/[0.08] text-[11px] font-medium text-white hover:bg-white/[0.15] transition-all flex items-center justify-center"
          @click="closeSidebar"
        >
          {{ $t(item.labelKey, item.fallback) }}
        </NuxtLink>
        <NuxtLink
          v-for="item in guidesNavLinks"
          :key="item.id"
          :to="localePath(item.to)"
          class="px-3 h-9 rounded-lg bg-white/[0.08] text-[11px] font-medium text-white hover:bg-white/[0.15] transition-all flex items-center justify-center"
          @click="closeSidebar"
        >
          {{ item.label }}
        </NuxtLink>
      </div>
    </div>

    <!-- 渐变分隔线 -->
    <div class="h-px bg-gradient-to-r from-transparent via-white/10 to-transparent"></div>

    <!-- Shop by Brand -->
    <div class="shrink-0">
      <div class="text-[10px] text-white/40 uppercase tracking-wider mb-1.5 md:mb-2 px-0.5">Shop by Brand</div>
      <div class="flex gap-1.5 md:gap-2">
        <button
          v-for="brand in brandCategories"
          :key="brand.id"
          type="button"
          class="flex-1 h-9 rounded-lg bg-white/[0.08] text-[11px] font-semibold text-white hover:bg-white/[0.15] transition-all flex items-center justify-center"
          @click="handleBrandClick(brand.id)"
        >
          {{ brand.label }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject } from 'vue'
import { useLocalePath } from '#imports'
import ProductSearchPanel from '~/components/ProductSearchPanel.vue'

const localePath = useLocalePath()

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

// 品牌点击处理
const handleBrandClick = (brandId: string) => {
  // TODO: 实现品牌跳转逻辑
  console.log('Brand clicked:', brandId)
  closeSidebar()
}

// Guides 快捷链接
const guidesLinks = [
  { id: 'wheelset-buyers', labelKey: 'products.nav.wheelsetBuyersGuide', to: '/guides/wheelset-buyers', fallback: 'Wheelset guide' },
  { id: 'tire-size-charts', labelKey: 'products.nav.tireSizeCharts', to: '/guides/sizecharts', fallback: 'Tire Size' },
]

// Guides 快捷链接
const guidesNavLinks = [
  { id: 'technical', label: 'Technical', to: '/guides/technical' },
  { id: 'our-story', label: 'Our Story', to: '/company/about' },
  { id: 'membership', label: 'Membership', to: '/company/membershipandpoints' },
]
</script>

