<template>
  <div class="sidebar-container">
    <!-- 搜索区域 -->
    <div class="sidebar-search">
      <ProductSearchPanel />
    </div>

    <!-- 手风琴区域 -->
    <div class="accordion">
      <!-- Product related -->
      <div 
        class="accordion-item" 
        :class="{ active: activeSection === 'product' }"
      >
        <div class="accordion-header" @click="toggleSection('product')">
          <div class="accordion-title">
            Product
            <span class="accordion-badge">{{ productCategories.length + productLinks.length }}</span>
          </div>
          <div class="accordion-arrow">▼</div>
        </div>
        <div class="accordion-content">
          <div class="accordion-body">
            <div class="btn-grid">
              <button
                v-for="cat in productCategories"
                :key="cat.id"
                type="button"
                class="sidebar-btn"
                @click="handleCategoryClick(cat.id)"
              >
                {{ cat.label }}
              </button>
              <NuxtLink
                v-for="item in productLinks"
                :key="item.id"
                :to="localePath(item.to)"
                class="sidebar-btn"
                @click="closeSidebar"
              >
                {{ $t(item.labelKey, item.fallback) }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <!-- Support -->
      <div 
        class="accordion-item" 
        :class="{ active: activeSection === 'support' }"
      >
        <div class="accordion-header" @click="toggleSection('support')">
          <div class="accordion-title">
            {{ $t('footer.menus.support', 'Support') }}
            <span class="accordion-badge">{{ supportLinks.length }}</span>
          </div>
          <div class="accordion-arrow">▼</div>
        </div>
        <div class="accordion-content">
          <div class="accordion-body">
            <div class="btn-grid">
              <NuxtLink
                v-for="item in supportLinks"
                :key="item.id"
                :to="localePath(item.to)"
                class="sidebar-btn"
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
        class="accordion-item" 
        :class="{ active: activeSection === 'brand' }"
      >
        <div class="accordion-header" @click="toggleSection('brand')">
          <div class="accordion-title">
            Brand
            <span class="accordion-badge">{{ brandCategories.length }}</span>
          </div>
          <div class="accordion-arrow">▼</div>
        </div>
        <div class="accordion-content">
          <div class="accordion-body">
            <div class="btn-grid">
              <button
                v-for="brand in brandCategories"
                :key="brand.id"
                type="button"
                class="sidebar-btn"
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
        class="accordion-item" 
        :class="{ active: activeSection === 'guides' }"
      >
        <div class="accordion-header" @click="toggleSection('guides')">
          <div class="accordion-title">
            Guides
            <span class="accordion-badge">{{ guidesLinks.length + guidesNavLinks.length }}</span>
          </div>
          <div class="accordion-arrow">▼</div>
        </div>
        <div class="accordion-content">
          <div class="accordion-body">
            <div class="btn-grid">
              <NuxtLink
                v-for="item in guidesLinks"
                :key="item.id"
                :to="localePath(item.to)"
                class="sidebar-btn"
                @click="closeSidebar"
              >
                {{ $t(item.labelKey, item.fallback) }}
              </NuxtLink>
              <NuxtLink
                v-for="item in guidesNavLinks"
                :key="item.id"
                :to="localePath(item.to)"
                class="sidebar-btn"
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

// 手风琴状态
const activeSection = ref<string | null>('product')

const toggleSection = (section: string) => {
  activeSection.value = activeSection.value === section ? null : section
}

// 获取侧边栏控制方法
const sidePanel = inject<{ closeLeft?: () => void }>('sidePanel', {})

const closeSidebar = () => {
  sidePanel.closeLeft?.()
}

// 商品分类按钮
const productCategories = [
  { id: 'carbon-rim', label: 'Carbon Rim' },
  { id: 'carbon-wheels', label: 'Carbon Wheels' },
  { id: 'hub', label: 'Hub' },
  { id: 'spoke', label: 'Spoke' },
  { id: 'tools', label: 'Tools' },
  { id: 'tire', label: 'Tire' },
]

// 分类点击处理（暂时为空，等商品准备好后实现）
const handleCategoryClick = (categoryId: string) => {
  // TODO: 实现分类跳转逻辑
  console.log('Category clicked:', categoryId)
}

// 商品快捷链接
const productLinks = [
  { id: 'shop', labelKey: 'products.nav.allProducts', to: '/shop', fallback: 'All products' },
]

// 支持快捷链接
const supportLinks = [
  { id: 'warranty-check', labelKey: 'support.nav.warrantyCheck', to: '/support/warranty-check', fallback: 'Warranty Check' },
  { id: 'spoke-calculator', labelKey: 'support.nav.spokeCalculator', to: '/spoke-calculator', fallback: 'Spoke Calculator' },
  { id: 'shipping', labelKey: 'support.nav.shipping', to: '/support/shipping', fallback: 'Shipping' },
  { id: 'payment', labelKey: 'support.nav.payment', to: '/support/payment', fallback: 'Payment' },
  { id: 'after-sales', labelKey: 'support.nav.afterSales', to: '/support/after-sales', fallback: 'After Sales' },
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
  { id: 'picture-warehouse', label: 'Picture Warehouse', to: '/picture-warehouse' },
]
</script>

<style scoped>
.sidebar-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 8px;
}

.sidebar-search {
  flex-shrink: 0;
}

/* 手风琴容器 */
.accordion {
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
  overflow-y: auto;
}

/* 手风琴项 */
.accordion-item {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.03) 0%, rgba(255, 255, 255, 0.01) 100%);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 14px;
  overflow: hidden;
  transition: all 0.25s ease;
}

.accordion-item:hover {
  border-color: rgba(107, 115, 255, 0.25);
}

.accordion-item.active {
  border-color: rgba(107, 115, 255, 0.4);
  background: linear-gradient(135deg, rgba(107, 115, 255, 0.08) 0%, rgba(107, 115, 255, 0.02) 100%);
  box-shadow: 0 4px 20px rgba(107, 115, 255, 0.1);
}

/* 手风琴标题 */
.accordion-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  cursor: pointer;
  -webkit-user-select: none;
  user-select: none;
  transition: all 0.15s ease;
}

.accordion-header:hover {
  background: rgba(255, 255, 255, 0.02);
}

.accordion-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  letter-spacing: 0.01em;
}

.accordion-icon {
  font-size: 18px;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
}

.accordion-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  font-size: 11px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.9);
  background: linear-gradient(135deg, rgba(107, 115, 255, 0.7) 0%, rgba(64, 255, 170, 0.5) 100%);
  border-radius: 10px;
  margin-left: 4px;
  box-shadow: 0 2px 8px rgba(107, 115, 255, 0.3);
}

.accordion-arrow {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.4);
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  transition: all 0.25s ease;
  font-size: 10px;
}

.accordion-item:hover .accordion-arrow {
  background: rgba(107, 115, 255, 0.15);
  color: rgba(255, 255, 255, 0.6);
}

.accordion-item.active .accordion-arrow {
  transform: rotate(180deg);
  background: rgba(107, 115, 255, 0.3);
  color: #fff;
}

/* 手风琴内容 */
.accordion-content {
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.accordion-item.active .accordion-content {
  max-height: 400px;
}

.accordion-body {
  padding: 4px 16px 16px;
}

/* 按钮网格 */
.btn-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(95px, 1fr));
  gap: 8px;
}

/* 按钮样式 */
.sidebar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 10px 14px;
  font-size: 12px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.75);
  background: linear-gradient(135deg, rgba(107, 115, 255, 0.12) 0%, rgba(107, 115, 255, 0.06) 100%);
  border: 1px solid rgba(107, 115, 255, 0.2);
  border-radius: 10px;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.sidebar-btn:hover {
  background: linear-gradient(135deg, rgba(107, 115, 255, 0.25) 0%, rgba(107, 115, 255, 0.15) 100%);
  border-color: rgba(107, 115, 255, 0.5);
  color: #fff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(107, 115, 255, 0.2);
}

.sidebar-btn:active {
  transform: translateY(0);
}

.sidebar-btn__icon {
  font-size: 13px;
}

/* 响应式 */
@media (max-width: 400px) {
  .sidebar-container {
    padding: 6px;
  }

  .btn-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .accordion-header {
    padding: 12px 14px;
  }

  .accordion-body {
    padding: 4px 14px 14px;
  }

  .sidebar-btn {
    padding: 9px 10px;
    font-size: 11px;
  }
}
</style>
