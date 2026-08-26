<template>
  <div class="flex justify-center pt-0 pb-0 w-full">
    <div class="membership-and-points-modal-card sidebar-panel leverandpoint-shell tz-mobile-dialog-surface tz-surface-card w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] rounded-2xl border tz-border-subtle backdrop-blur-xl shadow-[0_18px_44px_rgb(15_23_42_/_0.16)] relative overflow-hidden flex flex-col" role="region" aria-label="Membership Levels and Points">
      <button class="tz-global-close-btn absolute right-2 top-2 z-50 pointer-events-auto" type="button" @click="emit('close')">×</button>
      <div class="flex-1 flex py-4 px-0 md:p-4 md:px-5 pointer-events-auto overflow-hidden box-border">
        <div class="w-full h-full overflow-hidden pt-6">
          <MembershipAndPointsTabs variant="modal" class="h-full" />
        </div>
      </div>
      <div class="membership-and-points-modal-actions flex flex-col items-center justify-center py-3 pb-4 pointer-events-auto gap-3">
        <div class="flex flex-wrap gap-2 md:gap-3 items-center justify-center">
          <button
            class="membership-and-points-view-full-button h-10 px-[18px] rounded-full inline-flex items-center justify-center tz-surface-subtle tz-text-primary border tz-border-strong text-sm pointer-events-auto hover:tz-surface-muted transition-all shadow-[0_4px_12px_-4px_rgb(15_23_42_/_0.12)]"
            type="button"
            @click="handleMemberCenter"
          >
            {{ $t('member.viewAll', 'View Full') }}
          </button>
          <button
            class="h-10 px-[18px] rounded-full inline-flex items-center justify-center tz-surface-card tz-text-secondary border tz-border-strong text-sm pointer-events-auto shadow-[0_2px_6px_-3px_rgb(15_23_42_/_0.1)] hover:tz-surface-subtle hover:shadow-[0_4px_12px_-4px_rgb(15_23_42_/_0.14)] transition-all"
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
          class="fixed inset-0 z-[12000] flex items-end justify-center p-0 md:p-4 pointer-events-none tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
        >
          <div class="wa-drawer-backdrop" @click="closePrivacy"></div>
          <div class="relative z-10 pointer-events-none w-full max-w-[1400px]">
            <LazyPrivacyStatementModal class="pointer-events-auto" @close="closePrivacy" />
          </div>
        </div>
      </Transition>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onBeforeUnmount } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import MembershipAndPointsTabs from '~/components/MembershipAndPointsTabs.vue'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'

const emit = defineEmits<{
  (event: 'close'): void
}>()
const { t: $t } = useI18n()
const localePath = useLocalePath()
const showPrivacyModal = ref<boolean>(false)

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
const handlePrivacy = (): void => {
  showPrivacyModal.value = true
}

const closePrivacy = (): void => {
  showPrivacyModal.value = false
}

// Member Center - 跳转到会员中心页面
const handleMemberCenter = (): void => {
const target = localePath('/resources/membershipandpoints')
  if (typeof window !== 'undefined' && target) {
    window.location.href = String(target)
  }
}
</script>

<style src="~/assets/css/components/whatsapp-mobile-drawer.css"></style>

<style scoped>
.membership-and-points-modal-card {
  background-color: var(--tz-card-surface);
}

/* 自定义滚动条样式 */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: var(--tz-surface-muted);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--tz-border-strong);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: var(--tz-site-accent);
}

/* Firefox 滚动条样式 */
.custom-scrollbar {
  scrollbar-width: thin;
  scrollbar-color: var(--tz-border-strong) var(--tz-surface-muted);
}

@media (max-width: 767px) {
  .leverandpoint-shell {
    height: calc(var(--tz-mobile-safe-viewport-height, 100vh) - var(--tz-mobile-dialog-inset, 2px) * 2);
    max-height: calc(var(--tz-mobile-safe-viewport-height, 100vh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  }

  .membership-and-points-modal-actions {
    padding-top: 0.25rem;
    padding-bottom: var(--tz-mobile-modal-safe-padding-bottom, 0.75rem);
    gap: 0.25rem;
  }
}
</style>
