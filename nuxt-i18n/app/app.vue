<template>
  <SiteHeader />
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

<script setup>
import SidePanel from './components/SidePanel.vue'
import SidebarContent from './components/SidebarContent.vue'
import ProductSearchResults from './components/ProductSearchResults.vue'
import SiteHeader from '~/components/SiteHeader.vue'
</script>
