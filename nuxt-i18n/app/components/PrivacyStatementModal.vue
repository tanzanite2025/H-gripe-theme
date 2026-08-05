<template>
  <div
    class="privacy-statement-modal-shell wa-drawer-shell"
  >
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-3 md:py-4 border-b border-white/10">
      <h2
        class="text-xl md:text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-[#B5FF6D] to-[#6b73ff]"
      >
        {{ $t('privacy.title', 'Privacy statement') }}
      </h2>
      <button
        type="button"
        @click="$emit('close')"
        class="tz-global-close-btn"
        aria-label="Close privacy statement"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      </button>
    </div>

    <!-- Content -->
    <div class="privacy-statement-modal-content flex-1 overflow-y-auto px-6 py-6">
      <PrivacyStatementContent />
    </div>
  </div>
</template>

<script setup lang="ts">
import PrivacyStatementContent from '~/components/PrivacyStatementContent.vue'

const emit = defineEmits<{ (e: 'close'): void }>()
</script>

<style scoped>
@media (max-width: 767px) {
  .privacy-statement-modal-shell {
    height: min(92vh, calc(var(--tz-mobile-safe-viewport-height, 100vh) - 16px));
    max-height: min(92vh, calc(var(--tz-mobile-safe-viewport-height, 100vh) - 16px));
    border-color: rgba(255, 255, 255, 0.1) !important;
    box-shadow: 0 18px 44px rgba(0, 0, 0, 0.92) !important;
  }

  @supports (height: 100svh) {
    .privacy-statement-modal-shell {
      height: min(92svh, calc(var(--tz-mobile-safe-viewport-height, 100svh) - 16px));
      max-height: min(92svh, calc(var(--tz-mobile-safe-viewport-height, 100svh) - 16px));
    }
  }

  @supports (height: 100dvh) {
    .privacy-statement-modal-shell {
      height: min(92dvh, calc(var(--tz-mobile-safe-viewport-height, 100dvh) - 16px));
      max-height: min(92dvh, calc(var(--tz-mobile-safe-viewport-height, 100dvh) - 16px));
    }
  }

  .privacy-statement-modal-content {
    padding-bottom: var(--tz-mobile-modal-safe-padding-bottom, 1.5rem);
  }
}
</style>

