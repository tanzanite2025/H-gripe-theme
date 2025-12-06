<template>
  <div class="w-full h-full p-[5px] box-border">
    <div class="w-full h-full flex flex-col gap-4">
      <!-- 搜索区域 -->
      <div class="sidebar-search">
        <ProductSearchPanel />
      </div>

      <!-- 快捷链接区域 -->
      <div class="flex-1 overflow-y-auto space-y-4">
        <!-- Products 区块 -->
        <div class="sidebar-section">
          <h3 class="sidebar-section__title">
            <span class="sidebar-section__icon">🛒</span>
            {{ $t('footer.menus.products', 'Products') }}
          </h3>
          <nav class="sidebar-section__links">
            <NuxtLink
              v-for="item in productLinks"
              :key="item.id"
              :to="localePath(item.to)"
              class="sidebar-link"
              @click="closeSidebar"
            >
              {{ $t(item.labelKey, item.fallback) }}
            </NuxtLink>
          </nav>
        </div>

        <!-- Support 区块 -->
        <div class="sidebar-section">
          <h3 class="sidebar-section__title">
            <span class="sidebar-section__icon">🛠️</span>
            {{ $t('footer.menus.support', 'Support') }}
          </h3>
          <nav class="sidebar-section__links">
            <NuxtLink
              v-for="item in supportLinks"
              :key="item.id"
              :to="localePath(item.to)"
              class="sidebar-link"
              @click="closeSidebar"
            >
              <span v-if="item.icon" class="sidebar-link__icon">{{ item.icon }}</span>
              {{ $t(item.labelKey, item.fallback) }}
            </NuxtLink>
          </nav>
        </div>
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

// 商品快捷链接
const productLinks = [
  { id: 'shop', labelKey: 'products.nav.shop', to: '/shop', fallback: 'All Products' },
  { id: 'wheelset-buyers', labelKey: 'products.nav.wheelsetBuyersGuide', to: '/guides/wheelset-buyers', fallback: 'Wheelset Buyers Guide' },
  { id: 'tire-size-charts', labelKey: 'products.nav.tireSizeCharts', to: '/guides/sizecharts', fallback: 'Tire Size Charts' },
  { id: 'spoke-calculator', labelKey: 'support.nav.spokeCalculator', to: '/spoke-calculator', fallback: 'Spoke Calculator' },
]

// 支持快捷链接
const supportLinks = [
  { id: 'warranty-check', labelKey: 'support.nav.warrantyCheck', to: '/support/warranty-check', fallback: 'Warranty Check', icon: '🛡️' },
  { id: 'shipping', labelKey: 'support.nav.shipping', to: '/support/shipping', fallback: 'Shipping' },
  { id: 'payment', labelKey: 'support.nav.payment', to: '/support/payment', fallback: 'Payment' },
  { id: 'after-sales', labelKey: 'support.nav.afterSales', to: '/support/after-sales', fallback: 'After Sales' },
  { id: 'faqs', labelKey: 'support.nav.faqs', to: '/support/faqs', fallback: 'All FAQs' },
]
</script>

<style scoped>
.sidebar-section {
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.sidebar-section__title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin: 0 0 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.sidebar-section__icon {
  font-size: 1rem;
}

.sidebar-section__links {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  border-radius: 0.5rem;
  transition: all 0.15s ease;
}

.sidebar-link:hover {
  background: rgba(107, 115, 255, 0.15);
  color: #fff;
}

.sidebar-link__icon {
  font-size: 0.9rem;
}
</style>
