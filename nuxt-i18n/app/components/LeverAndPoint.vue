<template>
  <div class="flex justify-center pt-0 pb-0 w-full">
    <div class="sidebar-panel leverandpoint-shell w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] rounded-2xl border-2 border-[#6b73ff]/40 bg-slate-950/80 backdrop-blur-xl shadow-[0_0_30px_rgba(107,115,255,0.6)] relative overflow-hidden flex flex-col" role="region" aria-label="Membership Levels and Points">
      <!-- 背景装饰，与聊天欢迎页一致 -->
      <div class="absolute inset-x-0 top-0 h-[200px] bg-gradient-to-br from-indigo-600/20 to-teal-600/20 blur-3xl pointer-events-none z-0"></div>
      <button class="absolute right-2 top-2 z-50 w-7 h-7 inline-flex items-center justify-center border border-[rgba(124,117,255,0.6)] rounded-md bg-[rgba(30,27,75,0.6)] text-[#e8e9ff] pointer-events-auto hover:brightness-110 transition-all" type="button" @click="$emit('close')">×</button>
      <div class="flex-1 flex p-4 px-5 pointer-events-auto overflow-hidden box-border">
        <div class="w-full h-full overflow-hidden pt-6">
          <MembershipAndPointsTabs variant="modal" class="h-full" />
        </div>
      </div>
      <div class="flex flex-col items-center justify-center py-3 pb-4 pointer-events-auto gap-3">
        <div class="flex flex-wrap gap-2 md:gap-3 items-center justify-center">
          <button
            class="h-10 px-[18px] rounded-full inline-flex items-center justify-center bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black text-sm font-semibold pointer-events-auto hover:brightness-110 transition-all shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95)]"
            type="button"
            @click="handleMemberCenter"
          >
            {{ $t('member.viewAll', 'Member Center') }}
          </button>
          <button
            class="h-10 px-[18px] rounded-full inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white text-sm font-semibold pointer-events-auto shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] hover:bg-[linear-gradient(135deg,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] hover:shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95),0_0_8px_rgba(0,0,0,0.9)] transition-all"
            @click="handleSelectProducts"
          >
            Products
          </button>
          <button
            class="h-10 px-[18px] rounded-full inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white text-sm font-semibold pointer-events-auto shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] hover:bg-[linear-gradient(135deg,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] hover:shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95),0_0_8px_rgba(0,0,0,0.9)] transition-all"
            @click="handleViewCart"
          >
            Cart
          </button>
          <button
            class="h-10 px-[18px] rounded-full inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white text-sm font-semibold pointer-events-auto shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] hover:bg-[linear-gradient(135deg,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] hover:shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95),0_0_8px_rgba(0,0,0,0.9)] transition-all"
            @click="handleWishlist"
          >
            Wishlist
          </button>
          <button
            class="h-10 px-[18px] rounded-full inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white text-sm font-semibold pointer-events-auto shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] hover:bg-[linear-gradient(135deg,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] hover:shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95),0_0_8px_rgba(0,0,0,0.9)] transition-all"
            type="button"
            @click="handlePrivacy"
          >
            {{ $t('privacy.button', 'Privacy statement') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Privacy Statement Modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-300 ease-out"
        leave-active-class="transition duration-300 ease-in"
        enter-from-class="translate-y-full opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-y-full opacity-0"
      >
        <div
          v-if="showPrivacyModal"
          class="fixed inset-0 z-[12000] flex items-end justify-center p-0 md:p-4 pointer-events-none"
        >
          <div class="pointer-events-none w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh]">
            <PrivacyStatementModal class="pointer-events-auto" @close="closePrivacy" />
          </div>
        </div>
      </Transition>
    </Teleport>

    <WishlistDrawer v-model="wishlistDrawerVisible" variant="bottom" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch, onBeforeUnmount } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { useCart } from '~/composables/useCart'
import WishlistDrawer from '~/components/WishlistDrawer.vue'
import PrivacyStatementModal from '~/components/PrivacyStatementModal.vue'
import MembershipAndPointsTabs from '~/components/MembershipAndPointsTabs.vue'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'

const emit = defineEmits(['close'])
const cart = useCart()
const { t: $t } = useI18n()
const localePath = useLocalePath()
const showPrivacyModal = ref(false)
const wishlistDrawerVisible = ref(false)

const SIDEBAR_TOKEN_MODAL = 'lever-modal'
const SIDEBAR_TOKEN_PRIVACY = 'lever-privacy'

onMounted(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_MODAL, true)
})

watch(showPrivacyModal, (open) => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_PRIVACY, open)
}, { immediate: true })

onBeforeUnmount(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_MODAL, false)
  setSidebarHandlesHidden(SIDEBAR_TOKEN_PRIVACY, false)
})

// Products - 在新标签页打开 Shop 页面
const handleSelectProducts = () => {
  try {
    const target = localePath('/shop')
    if (typeof window !== 'undefined' && target) {
      window.open(String(target), '_blank')
    }
  } catch (e) {
    console.error('Failed to open shop page:', e)
  }
}

// Cart - 打开购物车弹窗（不关闭当前 LeverAndPoint），使用专用底部停靠模式
const handleViewCart = () => {
  cart.openCartFromLever()
}

// Privacy statement
const handlePrivacy = () => {
  showPrivacyModal.value = true
}

const closePrivacy = () => {
  showPrivacyModal.value = false
}

// Wishlist - 心愿单抽屉
const handleWishlist = () => {
  wishlistDrawerVisible.value = true
}

// Member Center - 跳转到会员中心页面
const handleMemberCenter = () => {
  const target = localePath('/company/membershipandpoints')
  if (typeof window !== 'undefined' && target) {
    window.location.href = String(target)
  }
}
</script>

<style scoped>
/* 自定义滚动条样式 */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(107, 115, 255, 0.5);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(107, 115, 255, 0.7);
}

/* Firefox 滚动条样式 */
.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: rgba(107, 115, 255, 0.5) rgba(255, 255, 255, 0.05);
}

@media (max-width: 767px) {
  .leverandpoint-shell {
    height: min(95vh, calc(100vh - 16px));
    max-height: min(95vh, calc(100vh - 16px));
  }

  @supports (height: 100svh) {
    .leverandpoint-shell {
      height: min(95svh, calc(100svh - 16px));
      max-height: min(95svh, calc(100svh - 16px));
    }
  }

  @supports (height: 100dvh) {
    .leverandpoint-shell {
      height: min(95dvh, calc(100dvh - 16px));
      max-height: min(95dvh, calc(100dvh - 16px));
    }
  }
}
</style>
