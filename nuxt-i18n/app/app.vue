<template>
  <div>
    <SiteHeader ref="siteHeaderRef" />
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
    
    <!-- 全局聊天弹窗 -->
    <WhatsAppChatModal
      v-if="currentConversation"
      :conversation="currentConversation"
      @close="closeChat"
    />
    
    <!-- Cookie 同意弹窗 -->
    <CookieConsent />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SidePanel from './components/SidePanel.vue'
import SidebarContent from './components/SidebarContent.vue'
import ProductSearchResults from './components/ProductSearchResults.vue'
import SiteHeader from '~/components/SiteHeader.vue'
import WhatsAppChatModal from '~/components/WhatsAppChatModal.vue'
import { useChatWidget } from '~/composables/useChatWidget'

// 全局聊天状态
const { currentConversation, closeChat } = useChatWidget()

const siteHeaderRef = ref<InstanceType<typeof SiteHeader> | null>(null)
</script>
