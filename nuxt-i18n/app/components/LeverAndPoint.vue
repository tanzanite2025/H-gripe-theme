<template>
  <div class="flex justify-center pt-0 pb-0 w-full">
    <div class="membership-and-points-modal-card sidebar-panel leverandpoint-shell w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] rounded-2xl border border-white/10 backdrop-blur-xl shadow-[0_18px_44px_rgba(0,0,0,0.92)] relative overflow-hidden flex flex-col" role="region" aria-label="Membership Levels and Points">
      <button class="tz-global-close-btn absolute right-2 top-2 z-50 pointer-events-auto" type="button" @click="$emit('close')">×</button>
      <div class="flex-1 flex py-4 px-0 md:p-4 md:px-5 pointer-events-auto overflow-hidden box-border">
        <div class="w-full h-full overflow-hidden pt-6">
          <MembershipAndPointsTabs variant="modal" class="h-full" />
        </div>
      </div>
      <div class="membership-and-points-modal-actions flex flex-col items-center justify-center py-3 pb-4 pointer-events-auto gap-3">
        <div class="flex flex-wrap gap-2 md:gap-3 items-center justify-center">
          <button
            class="membership-and-points-view-full-button h-10 px-[18px] rounded-full inline-flex items-center justify-center bg-white text-black text-sm pointer-events-auto hover:bg-white/90 transition-all shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95)]"
            type="button"
            @click="handleMemberCenter"
          >
            {{ $t('member.viewAll', 'View Full') }}
          </button>
          <button
            class="h-10 px-[18px] rounded-full inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white text-sm pointer-events-auto shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] hover:bg-[linear-gradient(135deg,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] hover:shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95),0_0_8px_rgba(0,0,0,0.9)] transition-all"
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
          class="fixed inset-0 z-[12000] flex items-end justify-center p-0 md:p-4 pointer-events-none tz-mobile-safe-modal-mask"
        >
          <div class="wa-drawer-backdrop" @click="closePrivacy"></div>
          <div class="relative z-10 pointer-events-none w-full max-w-[1400px]">
            <PrivacyStatementModal class="pointer-events-auto" @close="closePrivacy" />
          </div>
        </div>
      </Transition>
    </Teleport>

  </div>
</template>

<script setup>
import { ref, onMounted, watch, onBeforeUnmount } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import PrivacyStatementModal from '~/components/PrivacyStatementModal.vue'
import MembershipAndPointsTabs from '~/components/MembershipAndPointsTabs.vue'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'

const emit = defineEmits(['close'])
const { t: $t } = useI18n()
const localePath = useLocalePath()
const showPrivacyModal = ref(false)

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

// Privacy statement
const handlePrivacy = () => {
  showPrivacyModal.value = true
}

const closePrivacy = () => {
  showPrivacyModal.value = false
}

// Member Center - 跳转到会员中心页面
const handleMemberCenter = () => {
  const target = localePath('/membershipandpoints')
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
    height: var(--tz-mobile-safe-viewport-height, 100vh);
    max-height: var(--tz-mobile-safe-viewport-height, 100vh);
  }

  .membership-and-points-modal-actions {
    padding-top: 0.25rem;
    padding-bottom: var(--tz-mobile-modal-safe-padding-bottom, 0.75rem);
    gap: 0.25rem;
  }
}
</style>
