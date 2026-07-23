<template>
  <div>
    <NuxtLoadingIndicator
      color="repeating-linear-gradient(90deg, #22d3ee 0%, #8b5cf6 50%, #22d3ee 100%)"
      :height="3"
      :throttle="80"
    />
    <SiteHeader ref="siteHeaderRef" />
    <NuxtLayout>
      <SidePanel>
        <template #left>
          <SidebarContent />
        </template>
      </SidePanel>
      <!-- Render the current page inside the active layout -->
      <NuxtPage />
    </NuxtLayout>
    
    <!-- 购物车和结账弹窗 -->
    <LazyCartDrawer />
    <LazyCheckoutModal />
    <LazyShopSearchSheet />
    
    <!-- 全局聊天弹窗 -->
    <LazyWhatsAppChatModal
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
import SiteHeader from '~/components/SiteHeader.vue'
import { useChatWidget } from '~/composables/useChatWidget'

// 全局聊天状态
const { currentConversation, closeChat } = useChatWidget()

const siteHeaderRef = ref<InstanceType<typeof SiteHeader> | null>(null)
</script>
