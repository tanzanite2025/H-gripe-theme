<template>
  <ClientOnly>
    <Teleport to="body">
    <!-- 第一步：底部简洁横条。它是 fixed overlay，不参与文档流。 -->
    <div
      v-if="showBanner && !showModal"
      class="cookie-banner fixed left-0 right-0 z-[9999] border-t border-slate-200 bg-white shadow-[0_-10px_30px_rgba(20,32,43,0.12)]"
    >
        <div class="cookie-banner-inner max-w-5xl mx-auto px-4 py-4 flex flex-wrap items-center justify-center gap-4 sm:justify-between">
          <p class="text-sm tz-text-secondary text-center sm:text-left">
            {{ t('cookieConsent.banner.message') }}
            <NuxtLink 
              to="/policies/cookie"
              class="tz-text-accent hover:text-[var(--tz-text-accent-hover)] hover:underline underline decoration-[var(--tz-site-accent)] underline-offset-4 transition-colors"
              @click="hideBanner"
            >
              {{ t('cookieConsent.modal.policyLink') }}
            </NuxtLink>
          </p>
          <div class="flex items-center gap-3">
            <button 
              type="button"
              class="px-4 py-2 text-sm font-medium tz-text-muted bg-slate-50 border border-slate-200 rounded-full hover:bg-slate-100 hover:text-slate-900 transition-colors"
              @click="openCookieModal"
            >
              {{ t('cookieConsent.banner.customize') }}
            </button>
            <button 
              type="button"
              class="px-4 py-2 text-sm font-bold text-white bg-[var(--tz-action-primary)] rounded-full hover:bg-[var(--tz-action-primary-hover)] transition-colors shadow-[0_8px_20px_rgba(15,23,42,0.16)]"
              @click="handleAcceptAll"
            >
              {{ t('cookieConsent.banner.accept') }}
            </button>
          </div>
        </div>
    </div>

    <!-- 第二步：完整选择弹窗 -->
    <Transition name="cookie-fade">
      <div 
        v-if="showModal" 
        class="tz-standard-modal-mask fixed inset-0 z-[10000] flex items-center justify-center p-4 tz-mobile-safe-modal-mask"
        @click.self="closeCookieModal"
      >
        <div class="cookie-modal-panel tz-standard-modal-surface bg-white max-w-lg w-full max-h-[90vh] overflow-y-auto">
          <!-- Header -->
          <div class="flex items-center justify-between p-6 pb-4">
            <h2 class="text-xl font-bold tz-text-primary">{{ t('cookieConsent.modal.title') }}</h2>
            <button 
              type="button"
              class="tz-global-close-btn"
              @click="closeCookieModal"
              :aria-label="t('cookieConsent.modal.closeAriaLabel')"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Description -->
          <div class="px-6 pb-4">
            <p class="text-sm tz-text-secondary leading-relaxed">
              {{ t('cookieConsent.modal.descriptionBeforePolicy') }}
              <NuxtLink to="/policies/cookie" class="tz-text-primary hover:underline" @click="hideAll">
                {{ t('cookieConsent.modal.policyLink') }}
              </NuxtLink>{{ t('cookieConsent.modal.descriptionAfterPolicy') }}
            </p>
          </div>

          <!-- Cookie Options -->
          <div class="px-6 pb-6 space-y-4">
            <!-- Essential Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="essential" 
                checked 
                disabled
                class="mt-1 w-4 h-4 accent-[#059669] rounded cursor-not-allowed"
              />
              <div>
                <label for="essential" class="text-sm font-semibold tz-text-primary">{{ t('cookieConsent.options.essential.title') }}</label>
                <p class="text-xs tz-text-muted mt-0.5">{{ t('cookieConsent.options.essential.description') }}</p>
              </div>
            </div>

            <!-- Performance Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="performance" 
                v-model="preferences.performance"
                class="mt-1 w-4 h-4 accent-[#059669] rounded cursor-pointer"
              />
              <div>
                <label for="performance" class="text-sm font-semibold tz-text-primary cursor-pointer">{{ t('cookieConsent.options.performance.title') }}</label>
                <p class="text-xs tz-text-muted mt-0.5">{{ t('cookieConsent.options.performance.description') }}</p>
              </div>
            </div>

            <!-- Preference Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="preference" 
                v-model="preferences.preference"
                class="mt-1 w-4 h-4 accent-[#059669] rounded cursor-pointer"
              />
              <div>
                <label for="preference" class="text-sm font-semibold tz-text-primary cursor-pointer">{{ t('cookieConsent.options.preference.title') }}</label>
                <p class="text-xs tz-text-muted mt-0.5">{{ t('cookieConsent.options.preference.description') }}</p>
              </div>
            </div>

            <!-- Advertising Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="advertising" 
                v-model="preferences.advertising"
                class="mt-1 w-4 h-4 accent-[#059669] rounded cursor-pointer"
              />
              <div>
                <label for="advertising" class="text-sm font-semibold tz-text-primary cursor-pointer">{{ t('cookieConsent.options.advertising.title') }}</label>
                <p class="text-xs tz-text-muted mt-0.5">{{ t('cookieConsent.options.advertising.description') }}</p>
              </div>
            </div>
          </div>

          <!-- Buttons -->
          <div class="cookie-modal-actions flex flex-wrap gap-3 p-6 pt-0">
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-medium tz-text-muted bg-slate-50 border border-slate-200 rounded-lg hover:bg-slate-100 hover:text-slate-900 transition-colors"
              @click="handleSavePreferences"
            >
              {{ t('cookieConsent.actions.savePreferences') }}
            </button>
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-medium tz-text-muted bg-slate-50 border border-slate-200 rounded-lg hover:bg-slate-100 hover:text-slate-900 transition-colors"
              @click="handleRejectAll"
            >
              {{ t('cookieConsent.actions.rejectAll') }}
            </button>
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-semibold text-white bg-[var(--tz-action-primary)] rounded-lg hover:bg-[var(--tz-action-primary-hover)] transition-colors shadow-[0_8px_18px_rgba(15,23,42,0.16)]"
              @click="handleAcceptAll"
            >
              {{ t('cookieConsent.actions.acceptAll') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
    </Teleport>
  </ClientOnly>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from '#imports'
import {
  COOKIE_CONSENT_KEY,
  COOKIE_CONSENT_UPDATED_EVENT,
  readCookieConsent,
  type CookieConsentPreferences,
} from '~/utils/cookieConsent'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'
const { t } = useI18n()
const overlayBackStack = useOverlayBackStack()

const showBanner = ref(false)
const showModal = ref(false)
const COOKIE_BANNER_REVEAL_DELAY_MS = 60_000
let bannerRevealTimer: number | null = null

const clearBannerRevealTimer = () => {
  if (bannerRevealTimer === null || typeof window === 'undefined') return

  window.clearTimeout(bannerRevealTimer)
  bannerRevealTimer = null
}

const revealBannerAfterDelay = () => {
  bannerRevealTimer = null
  showBanner.value = true
}

const scheduleBannerReveal = () => {
  if (bannerRevealTimer !== null || typeof window === 'undefined') return

  bannerRevealTimer = window.setTimeout(revealBannerAfterDelay, COOKIE_BANNER_REVEAL_DELAY_MS)
}

const closeCookieModalState = () => {
  showModal.value = false
}

const openCookieModal = () => {
  showModal.value = true
  overlayBackStack.open('cookie-consent', closeCookieModalState)
}

const closeCookieModal = () => {
  void overlayBackStack.close('cookie-consent')
  closeCookieModalState()
}

const preferences = ref({
  performance: false,
  preference: false,
  advertising: false
})

// 隐藏横条
const hideBanner = () => {
  clearBannerRevealTimer()
  showBanner.value = false
}

// 隐藏全部（横条和弹窗）
const hideAll = () => {
  clearBannerRevealTimer()
  showBanner.value = false
  closeCookieModal()
}

// 检查是否已有保存的偏好
// 保存偏好到 localStorage
const saveConsent = (prefs: Omit<CookieConsentPreferences, 'essential' | 'timestamp'>) => {
  const consent: CookieConsentPreferences = {
    essential: true, // 始终为 true
    ...prefs,
    timestamp: Date.now()
  }
  localStorage.setItem(COOKIE_CONSENT_KEY, JSON.stringify(consent))
  clearBannerRevealTimer()
  showBanner.value = false
  closeCookieModal()
  
  // 触发自定义事件，供其他组件监听
  window.dispatchEvent(new CustomEvent(COOKIE_CONSENT_UPDATED_EVENT, { detail: consent }))
}

// 接受全部
const handleAcceptAll = () => {
  saveConsent({
    performance: true,
    preference: true,
    advertising: true
  })
}

// 拒绝全部（只保留必要 Cookie）
const handleRejectAll = () => {
  saveConsent({
    performance: false,
    preference: false,
    advertising: false
  })
}

// 保存当前选择
const handleSavePreferences = () => {
  saveConsent({
    performance: preferences.value.performance,
    preference: preferences.value.preference,
    advertising: preferences.value.advertising
  })
}

onMounted(() => {
  const existing = readCookieConsent()
  if (!existing) {
    // 首屏不插入提示，延迟显示，避免影响 LCP 候选元素。
    scheduleBannerReveal()
  } else {
    // 恢复已保存的偏好
    preferences.value = {
      performance: existing.performance,
      preference: existing.preference,
      advertising: existing.advertising
    }
  }

  // The banner is fixed and must never change document geometry.
})

onBeforeUnmount(() => {
  clearBannerRevealTimer()
})

// 暴露方法供外部调用（如用户想重新设置偏好）
defineExpose({
  show: () => {
    clearBannerRevealTimer()
    showBanner.value = true
  },
  hide: () => {
    clearBannerRevealTimer()
    showBanner.value = false
  }
})
</script>

<style>
.cookie-banner {
  bottom: var(--tz-bottom-dock-height, 4.5rem);
}

.cookie-fade-enter-active,
.cookie-fade-leave-active {
  transition: opacity 0.3s ease;
}

.cookie-fade-enter-from,
.cookie-fade-leave-to {
  opacity: 0;
}

@media (max-width: 767px) {
  .cookie-banner-inner {
    padding-bottom: var(--tz-mobile-modal-safe-padding-bottom, 1rem);
  }

  .cookie-modal-panel {
    max-height: min(90vh, var(--tz-mobile-safe-viewport-height, 90vh));
  }

  .cookie-modal-actions {
    padding-bottom: var(--tz-mobile-modal-safe-padding-bottom, 1.5rem);
  }
}

</style>
