<template>
  <SiteHeader ref="siteHeaderRef" />
  <div :style="{ height: headerHeight + 'px' }" />
  <NuxtLayout>
    <SidePanel>
      <template #left>
        <SidebarContent />
      </template>
      <template #right>
        <ProductSearchResults />
      </template>
    </SidePanel>
    <!-- Render the current page inside the active layout -->
    <NuxtPage />
  </NuxtLayout>
  
  <!-- 购物车和结账弹窗 -->
  <CartDrawer />
  <CheckoutModal />
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import SidePanel from './components/SidePanel.vue'
import SidebarContent from './components/SidebarContent.vue'
import ProductSearchResults from './components/ProductSearchResults.vue'
import SiteHeader from '~/components/SiteHeader.vue'

const siteHeaderRef = ref<InstanceType<typeof SiteHeader> | null>(null)
const headerHeight = ref(0)

const updateHeaderHeight = () => {
  if (typeof window === 'undefined') return
  const inst = siteHeaderRef.value as any
  if (!inst || !inst.$el) return
  const el = inst.$el as HTMLElement
  const rect = el.getBoundingClientRect()
  // spacer 高度 = 视口顶部到 SiteHeader 底部的距离，
  // 把整个 SiteHeader 组件（包含它自己的 top 偏移）全部预留出来。
  headerHeight.value = rect.bottom
}

onMounted(() => {
  if (typeof window === 'undefined') return
  nextTick(() => {
    updateHeaderHeight()
    window.addEventListener('resize', updateHeaderHeight)
  })
})

onBeforeUnmount(() => {
  if (typeof window === 'undefined') return
  window.removeEventListener('resize', updateHeaderHeight)
})
</script>
