<template>
  <div class="sidebar-content w-full h-full flex flex-col p-3 md:p-4 max-md:pb-2">
    <!-- Dock 导航 -->
    <div class="shrink-0 dock-nav flex gap-1 p-1.5 rounded-3xl bg-white/[0.03] border border-white/5 mb-3">
      <NuxtLink 
        :to="localePath('/')"
        class="dock-item active"
        @click="closeSidebar"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
        <span>Home</span>
      </NuxtLink>
      <NuxtLink 
        :to="localePath('/shop')"
        class="dock-item"
        @click="closeSidebar"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
        </svg>
        <span>Shop</span>
      </NuxtLink>
      <NuxtLink 
        :to="localePath('/picture-warehouse')"
        class="dock-item"
        @click="closeSidebar"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <span>Picture</span>
      </NuxtLink>
    </div>

    <!-- 可滚动内容区 -->
    <div class="flex-1 overflow-y-auto flex flex-col gap-2 md:gap-4 pr-1 scrollbar-hide">
      <!-- Quick Access -->
      <div>
        <div class="space-y-2">
          <!-- Payment -->
          <NuxtLink 
            :to="localePath('/support/payment')"
            class="hybrid-card card-indigo"
            @click="closeSidebar"
          >
            <div class="hybrid-icon">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <div class="hybrid-title">Payment Methods</div>
              <div class="hybrid-desc max-md:truncate">Wallet, Cards & Crypto</div>
            </div>
            <svg class="hybrid-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </NuxtLink>
          <!-- Shipping -->
          <NuxtLink 
            :to="localePath('/support/shipping')"
            class="hybrid-card card-teal"
            @click="closeSidebar"
          >
            <div class="hybrid-icon">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <div class="hybrid-title">Shipping Info</div>
              <div class="hybrid-desc max-md:truncate">Track & Trace</div>
            </div>
            <svg class="hybrid-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </NuxtLink>
          <!-- Warranty -->
          <NuxtLink 
            :to="localePath('/support/warranty')"
            class="hybrid-card card-sky"
            @click="closeSidebar"
          >
            <div class="hybrid-icon">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <div class="hybrid-title">Warranty Policy</div>
              <div class="hybrid-desc max-md:truncate">Protection Plan</div>
            </div>
            <svg class="hybrid-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </NuxtLink>
          <!-- After Sales -->
          <NuxtLink 
            :to="localePath('/support/after-sales')"
            class="hybrid-card card-pink"
            @click="closeSidebar"
          >
            <div class="hybrid-icon">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <div class="hybrid-title">After Sales</div>
              <div class="hybrid-desc max-md:truncate">24/7 Support</div>
            </div>
            <svg class="hybrid-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </NuxtLink>

          <NuxtLink
            :to="localePath('/policies')"
            class="hybrid-card card-teal"
            @click="closeSidebar"
          >
            <div class="hybrid-icon">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m-7 4h8a2 2 0 002-2V6a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <div class="hybrid-title">Policies</div>
              <div class="hybrid-desc max-md:truncate">cookie, refund-return, terms and more</div>
            </div>
            <svg class="hybrid-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </NuxtLink>

          <NuxtLink
            :to="localePath('/company/about')"
            class="hybrid-card card-sky"
            @click="closeSidebar"
          >
            <div class="hybrid-icon">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 21h8m-4-4v4m-7-7h14M5 10h14M5 6h14M6 10v11m12-11v11" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <div class="hybrid-title">About us</div>
              <div class="hybrid-desc max-md:truncate">Factory Facilities Detection OEM/ODM</div>
            </div>
            <svg class="hybrid-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </NuxtLink>

          <NuxtLink
            :to="localePath('/blog')"
            class="hybrid-card card-indigo"
            @click="closeSidebar"
          >
            <div class="hybrid-icon">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="1.5"
                  d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10l6 6v8a2 2 0 01-2 2z"
                />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M14 4v6h6" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <div class="hybrid-title">{{ $t('products.nav.wheelsbuildBlog') }}</div>
              <div class="hybrid-desc max-md:truncate">
                {{ $t('sidebar.wheelsbuildBlog.desc') }}
              </div>
            </div>
            <svg class="hybrid-arrow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </NuxtLink>
        </div>
      </div>

      <!-- Tools & Resources -->
      <div>
        <div class="flex flex-wrap gap-2">
          <!-- Support -->
          <NuxtLink
            v-for="item in supportLinks"
            :key="item.id"
            :to="localePath(item.to)"
            class="hybrid-btn btn-teal"
            @click="closeSidebar"
          >
            {{ $t(item.labelKey, item.fallback) }}
          </NuxtLink>
          <!-- Guides -->
          <NuxtLink
            v-for="item in guidesLinks"
            :key="item.id"
            :to="localePath(item.to)"
            class="hybrid-btn btn-sky"
            @click="closeSidebar"
          >
            {{ $t(item.labelKey, item.fallback) }}
          </NuxtLink>
          <NuxtLink
            v-for="item in guidesNavLinks"
            :key="item.id"
            :to="localePath(item.to)"
            class="hybrid-btn btn-sky"
            @click="closeSidebar"
          >
            {{ item.label }}
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject } from 'vue'
import { useLocalePath } from '#imports'

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

// Guides 快捷链接
const guidesLinks = [
  { id: 'wheelset-buyers', labelKey: 'products.nav.wheelsetBuyersGuide', to: '/guides/wheelset-buyers', fallback: 'Wheelset guide' },
  { id: 'tire-size-charts', labelKey: 'products.nav.tireSizeCharts', to: '/guides/sizecharts', fallback: 'Tire Size' },
]

// Guides 快捷链接
const guidesNavLinks = [
  { id: 'technical', label: 'Technical', to: '/guides/technical' },
  { id: 'our-story', label: 'Our Story', to: '/company/ourstory' },
  { id: 'membership', label: 'Membership', to: '/membershipandpoints' },
]
</script>

<style scoped>
/* 隐藏滚动条 */
.scrollbar-hide::-webkit-scrollbar { display: none; }
.scrollbar-hide { -ms-overflow-style: none; scrollbar-width: none; }

/* Dock 导航 */
.dock-nav {
  background: radial-gradient(circle at top left, rgba(45,55,72,0.9), rgba(25,32,52,0.94));
  border-radius: 24px;
  border: none;
  box-shadow: 0 14px 32px -14px rgba(0,0,0,0.9);
}
.dock-item {
  flex: 1;
  height: 40px;
  border-radius: 18px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  color: rgba(248,250,252,0.9);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  font-size: 10px;
  font-weight: 600;
  cursor: pointer;
}
.dock-item:hover {
  background: radial-gradient(circle at top, rgba(15,23,42,0.9), rgba(15,23,42,0.8));
  box-shadow: 0 16px 28px -12px rgba(0,0,0,0.9);
}
.dock-item.active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

/* 区块标题 */
.section-label {
  font-size: 10px;
  color: rgba(255,255,255,0.4);
  text-transform: uppercase;
  letter-spacing: 1.5px;
  font-weight: 700;
  margin-bottom: 10px;
  margin-left: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.section-label .dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}

/* Hybrid 卡片 */
.hybrid-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: radial-gradient(circle at top left, rgba(48,62,86,0.96), rgba(28,38,60,0.96));
  border-radius: 16px;
  border: none;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  text-decoration: none;
  box-shadow: 0 16px 36px -14px rgba(0,0,0,0.95);
}
.hybrid-card::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  background: var(--card-color);
  opacity: 0.85;
  transition: all 0.3s ease;
  box-shadow: none;
}
.hybrid-card:hover {
  background: linear-gradient(90deg, rgba(31,41,55,0.98), rgba(15,23,42,0.99));
  transform: translateY(-2px);
  box-shadow: 0 20px 42px -14px rgba(0,0,0,0.95);
}
.hybrid-card:active {
  transform: translateY(2px) scale(0.99);
  box-shadow: 0 12px 28px -12px rgba(0,0,0,0.9);
}
.hybrid-card:hover::after {
  opacity: 1;
  width: 6px;
  box-shadow: none;
}

/* 卡片颜色变体 */
.card-indigo {
  --card-color: #818cf8;
  --card-color-alpha: rgba(129, 140, 248, 0.5);
  --card-color-dim: rgba(129, 140, 248, 0.2);
  --card-bg: rgba(129, 140, 248, 0.1);
}
.card-teal {
  --card-color: #2dd4bf;
  --card-color-alpha: rgba(45, 212, 191, 0.5);
  --card-color-dim: rgba(45, 212, 191, 0.2);
  --card-bg: rgba(45, 212, 191, 0.1);
}
.card-sky {
  --card-color: #38bdf8;
  --card-color-alpha: rgba(56, 189, 248, 0.5);
  --card-color-dim: rgba(56, 189, 248, 0.2);
  --card-bg: rgba(56, 189, 248, 0.1);
}
.card-pink {
  --card-color: #8b5cf6;
  --card-color-alpha: rgba(139, 92, 246, 0.55);
  --card-color-dim: rgba(139, 92, 246, 0.25);
  --card-bg: rgba(139, 92, 246, 0.12);
}

.hybrid-icon {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  background: var(--card-bg);
  color: var(--card-color);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
  border: 1px solid rgba(255,255,255,0.05);
  flex-shrink: 0;
}
.hybrid-card:hover .hybrid-icon {
  background: var(--card-color);
  color: #000;
  box-shadow: 0 14px 24px -12px rgba(0,0,0,0.88);
  transform: scale(1.1) rotate(-5deg);
}

.hybrid-title {
  font-size: 14px;
  font-weight: 700;
  color: #e5e5e5;
  margin-bottom: 2px;
}
.hybrid-card:hover .hybrid-title { color: #fff; }

.hybrid-desc {
  font-size: 11px;
  color: #888;
  font-weight: 500;
}
.hybrid-card:hover .hybrid-desc { color: var(--card-color); }

.hybrid-arrow {
  width: 16px;
  height: 16px;
  margin-left: auto;
  margin-right: 8px;
  color: #fff;
  transition: all 0.3s;
  flex-shrink: 0;
  opacity: 1;
}
.hybrid-card:hover .hybrid-arrow {
  color: #fff;
  transform: translateX(4px);
}

/* 工具按钮 */
.hybrid-btn {
  background: linear-gradient(135deg, rgba(28,38,58,0.95), rgba(24,33,52,0.93)),
    radial-gradient(circle at top left, var(--btn-bg), transparent 65%);
  border: none;
  height: 32px;
  padding: 0 12px;
  border-radius: 8px;
  font-size: 11px;
  font-weight: 600;
  color: #cbd5f5;
  transition: all 0.2s ease;
  flex-grow: 1;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 12px 26px -12px rgba(0,0,0,0.9);
}
.hybrid-btn:hover {
  background: linear-gradient(135deg, var(--btn-bg), rgba(15,23,42,0.96));
  color: var(--btn-color);
  box-shadow: 0 14px 30px -12px rgba(0,0,0,0.95);
  transform: translateY(-1px);
}

/* 按钮颜色变体 */
.btn-teal {
  --btn-color: #2dd4bf;
  --btn-bg: rgba(45,212,191,0.1);
  --btn-dim: rgba(45,212,191,0.2);
}
.btn-sky {
  --btn-color: #38bdf8;
  --btn-bg: rgba(56,189,248,0.1);
  --btn-dim: rgba(56,189,248,0.2);
}
.btn-indigo {
  --btn-color: #818cf8;
  --btn-bg: rgba(129,140,248,0.1);
  --btn-dim: rgba(129,140,248,0.2);
}

@media (max-width: 767px) {
  .hybrid-card {
    gap: 10px;
    padding: 10px 12px;
    border-radius: 14px;
  }

  .hybrid-icon {
    width: 34px;
    height: 34px;
    border-radius: 10px;
  }

  .hybrid-title {
    font-size: 13px;
    margin-bottom: 1px;
  }

  .hybrid-desc {
    font-size: 10px;
  }

  .hybrid-arrow {
    margin-right: 6px;
  }
}
</style>

