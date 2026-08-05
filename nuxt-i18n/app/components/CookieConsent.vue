<template>
  <ClientOnly>
    <Teleport to="body">
    <!-- 第一步：底部简洁横条 -->
    <Transition name="cookie-slide">
      <div 
        v-if="showBanner && !showModal" 
        class="cookie-banner fixed bottom-0 left-0 right-0 z-[9999] bg-[rgba(0,0,0,0.78)] border-t border-white/10 shadow-[0_-10px_30px_rgba(0,0,0,0.45)]"
      >
        <div class="cookie-banner-inner max-w-5xl mx-auto px-4 py-4 flex flex-wrap items-center justify-center gap-4 sm:justify-between">
          <p class="text-sm text-white/75 text-center sm:text-left">
            {{ t('cookieConsent.banner.message') }}
            <NuxtLink 
              to="/policies/cookie"
              class="text-[#B5FF6D] hover:text-white underline decoration-white/30 underline-offset-4 transition-colors"
              @click="hideBanner"
            >
              {{ t('cookieConsent.banner.learnMore') }}
            </NuxtLink>
          </p>
          <div class="flex items-center gap-3">
            <button 
              type="button"
              class="px-4 py-2 text-sm font-medium text-white/75 bg-white/5 border border-white/10 rounded-full hover:bg-white/10 hover:text-white transition-colors"
              @click="showModal = true"
            >
              {{ t('cookieConsent.banner.customize') }}
            </button>
            <button 
              type="button"
              class="px-4 py-2 text-sm font-bold text-black bg-white rounded-full hover:bg-white/90 transition-colors shadow-[0_8px_20px_rgba(255,255,255,0.12)]"
              @click="handleAcceptAll"
            >
              {{ t('cookieConsent.banner.accept') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 第二步：完整选择弹窗 -->
    <Transition name="cookie-fade">
      <div 
        v-if="showModal" 
        class="fixed inset-0 z-[10000] flex items-center justify-center bg-black/50 p-4 tz-mobile-safe-modal-mask"
        @click.self="showModal = false"
      >
        <div class="cookie-modal-panel bg-[rgba(0,0,0,0.88)] border border-white/10 rounded-2xl shadow-2xl shadow-black/60 max-w-lg w-full max-h-[90vh] overflow-y-auto">
          <!-- Header -->
          <div class="flex items-center justify-between p-6 pb-4">
            <h2 class="text-xl font-bold text-white">{{ t('cookieConsent.modal.title') }}</h2>
            <button 
              type="button"
              class="tz-global-close-btn"
              @click="showModal = false"
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
              <NuxtLink to="/policies/cookie" class="text-[#B5FF6D] hover:text-white hover:underline" @click="hideAll">
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
                class="mt-1 w-4 h-4 accent-[#B5FF6D] rounded cursor-not-allowed"
              />
              <div>
                <label for="essential" class="text-sm font-semibold text-white">{{ t('cookieConsent.options.essential.title') }}</label>
                <p class="text-xs tz-text-muted mt-0.5">{{ t('cookieConsent.options.essential.description') }}</p>
              </div>
            </div>

            <!-- Performance Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="performance" 
                v-model="preferences.performance"
                class="mt-1 w-4 h-4 accent-[#B5FF6D] rounded cursor-pointer"
              />
              <div>
                <label for="performance" class="text-sm font-semibold text-white cursor-pointer">{{ t('cookieConsent.options.performance.title') }}</label>
                <p class="text-xs tz-text-muted mt-0.5">{{ t('cookieConsent.options.performance.description') }}</p>
              </div>
            </div>

            <!-- Preference Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="preference" 
                v-model="preferences.preference"
                class="mt-1 w-4 h-4 accent-[#B5FF6D] rounded cursor-pointer"
              />
              <div>
                <label for="preference" class="text-sm font-semibold text-white cursor-pointer">{{ t('cookieConsent.options.preference.title') }}</label>
                <p class="text-xs tz-text-muted mt-0.5">{{ t('cookieConsent.options.preference.description') }}</p>
              </div>
            </div>

            <!-- Advertising Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="advertising" 
                v-model="preferences.advertising"
                class="mt-1 w-4 h-4 accent-[#B5FF6D] rounded cursor-pointer"
              />
              <div>
                <label for="advertising" class="text-sm font-semibold text-white cursor-pointer">{{ t('cookieConsent.options.advertising.title') }}</label>
                <p class="text-xs tz-text-muted mt-0.5">{{ t('cookieConsent.options.advertising.description') }}</p>
              </div>
            </div>
          </div>

          <!-- Buttons -->
          <div class="cookie-modal-actions flex flex-wrap gap-3 p-6 pt-0">
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-medium text-white/70 bg-white/5 border border-white/10 rounded-lg hover:bg-white/10 hover:text-white transition-colors"
              @click="handleSavePreferences"
            >
              {{ t('cookieConsent.actions.savePreferences') }}
            </button>
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-medium text-white/70 bg-white/5 border border-white/10 rounded-lg hover:bg-white/10 hover:text-white transition-colors"
              @click="handleRejectAll"
            >
              {{ t('cookieConsent.actions.rejectAll') }}
            </button>
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-semibold text-black bg-white rounded-lg hover:bg-white/90 transition-colors shadow-lg shadow-white/10"
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
import { ref, onMounted } from 'vue'
import { useI18n } from '#imports'

const COOKIE_CONSENT_KEY = 'tanzanite_cookie_consent'
const { t } = useI18n()

interface CookiePreferences {
  essential: boolean
  performance: boolean
  preference: boolean
  advertising: boolean
  timestamp: number
}

const showBanner = ref(false)
const showModal = ref(false)

const preferences = ref({
  performance: false,
  preference: false,
  advertising: false
})

// 隐藏横条
const hideBanner = () => {
  showBanner.value = false
}

// 隐藏全部（横条和弹窗）
const hideAll = () => {
  showBanner.value = false
  showModal.value = false
}

// 检查是否已有保存的偏好
const checkExistingConsent = (): CookiePreferences | null => {
  if (typeof window === 'undefined') return null
  
  const stored = localStorage.getItem(COOKIE_CONSENT_KEY)
  if (!stored) return null
  
  try {
    return JSON.parse(stored) as CookiePreferences
  } catch {
    return null
  }
}

// 保存偏好到 localStorage
const saveConsent = (prefs: Omit<CookiePreferences, 'essential' | 'timestamp'>) => {
  const consent: CookiePreferences = {
    essential: true, // 始终为 true
    ...prefs,
    timestamp: Date.now()
  }
  localStorage.setItem(COOKIE_CONSENT_KEY, JSON.stringify(consent))
  showBanner.value = false
  showModal.value = false
  
  // 触发自定义事件，供其他组件监听
  window.dispatchEvent(new CustomEvent('cookie-consent-updated', { detail: consent }))
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
  const existing = checkExistingConsent()
  if (!existing) {
    // 没有保存的偏好，显示弹窗
    showBanner.value = true
  } else {
    // 恢复已保存的偏好
    preferences.value = {
      performance: existing.performance,
      preference: existing.preference,
      advertising: existing.advertising
    }
  }
})

// 暴露方法供外部调用（如用户想重新设置偏好）
defineExpose({
  show: () => { showBanner.value = true },
  hide: () => { showBanner.value = false }
})
</script>

<style>
.cookie-fade-enter-active,
.cookie-fade-leave-active {
  transition: opacity 0.3s ease;
}

.cookie-fade-enter-from,
.cookie-fade-leave-to {
  opacity: 0;
}

.cookie-slide-enter-active,
.cookie-slide-leave-active {
  transition: transform 0.3s ease, opacity 0.3s ease;
}

.cookie-slide-enter-from,
.cookie-slide-leave-to {
  transform: translateY(100%);
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
