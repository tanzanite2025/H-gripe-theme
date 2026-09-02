<template>
  <SidePanel>
    <template #left>
      <LazyAccountSidebarPanel />
    </template>
  </SidePanel>
  <component
    :is="cartDrawerComponent"
    v-if="isCartOpen && cartDrawerComponent"
  />
  <component
    :is="checkoutModalComponent"
    v-if="isCheckoutOpen && checkoutModalComponent"
  />
  <component
    :is="shopSearchSheetComponent"
    v-if="isShopSearchSheetOpen && shopSearchSheetComponent"
  />
  <component
    :is="pagesSearchOverlayComponent"
    v-if="isPagesSearchOpen && pagesSearchOverlayComponent"
  />
  <component
    :is="globalProductDetailBottomSheetComponent"
    v-if="isGlobalProductDetailBottomSheetOpen && globalProductDetailBottomSheetComponent"
  />
  <component
    :is="whatsAppChatModalComponent"
    v-if="currentConversation && whatsAppChatModalComponent"
    :conversation="currentConversation"
    @close="closeChat"
  />
</template>

<script setup lang="ts">
import {
  defineAsyncComponent,
  shallowRef,
  watch,
  type Component,
  type ShallowRef,
} from 'vue'
import { useChatWidget } from '~/composables/useChatWidget'
import { useCart } from '~/composables/useCart'
import { useGlobalProductDetailBottomSheet } from '~/composables/useGlobalProductDetailBottomSheet'
import { usePagesSearchOverlayState } from '~/composables/usePagesSearchOverlayState'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'
import SidePanel from '~/components/SidePanel.vue'

const { currentConversation, closeChat } = useChatWidget()
const { isCartOpen, isCheckoutOpen } = useCart()
const { isGlobalProductDetailBottomSheetOpen } = useGlobalProductDetailBottomSheet()
const { isPagesSearchOpen } = usePagesSearchOverlayState()
const { isOpen: isShopSearchSheetOpen } = useShopSearchSheet()

const createDeferredOverlay = (
  loader: () => Promise<{ default: Component }>,
) => {
  const component = shallowRef(null) as ShallowRef<Component | null>

  const load = () => {
    if (component.value) return
    component.value = defineAsyncComponent(async () => (await loader()).default)
  }

  return { component, load }
}

const {
  component: cartDrawerComponent,
  load: loadCartDrawer,
} = createDeferredOverlay(() => import('./CartDrawer.vue'))
const {
  component: checkoutModalComponent,
  load: loadCheckoutModal,
} = createDeferredOverlay(() => import('./CheckoutModal.vue'))
const {
  component: shopSearchSheetComponent,
  load: loadShopSearchSheet,
} = createDeferredOverlay(() => import('./ShopSearchSheet.vue'))
const {
  component: pagesSearchOverlayComponent,
  load: loadPagesSearchOverlay,
} = createDeferredOverlay(() => import('./search/GlobalPagesSearchOverlay.vue'))
const {
  component: globalProductDetailBottomSheetComponent,
  load: loadGlobalProductDetailBottomSheet,
} = createDeferredOverlay(() => import('./global/product-detail/GlobalProductDetailBottomSheet.vue'))
const {
  component: whatsAppChatModalComponent,
  load: loadWhatsAppChatModal,
} = createDeferredOverlay(() => import('./WhatsAppChatModal.vue'))

watch(isCartOpen, (open) => {
  if (open) loadCartDrawer()
}, { immediate: true })

watch(isCheckoutOpen, (open) => {
  if (open) loadCheckoutModal()
}, { immediate: true })

watch(isShopSearchSheetOpen, (open) => {
  if (open) loadShopSearchSheet()
}, { immediate: true })

watch(isPagesSearchOpen, (open) => {
  if (open) loadPagesSearchOverlay()
}, { immediate: true })

watch(isGlobalProductDetailBottomSheetOpen, (open) => {
  if (open) loadGlobalProductDetailBottomSheet()
}, { immediate: true })

watch(currentConversation, (conversation) => {
  if (conversation) loadWhatsAppChatModal()
}, { immediate: true })
</script>
