<template>
  <ClientOnly>
    <Teleport to="body">
    <!-- 第一步：底部简洁横条 -->
    <Transition name="cookie-slide">
      <div 
        v-if="showBanner && !showModal" 
        class="fixed bottom-0 left-0 right-0 z-[9999] bg-slate-900/95 backdrop-blur-sm border-t border-slate-700 shadow-lg"
      >
        <div class="max-w-5xl mx-auto px-4 py-4 flex flex-wrap items-center justify-center gap-4 sm:justify-between">
          <p class="text-sm text-slate-300 text-center sm:text-left">
            We use cookies for a better experience. 
            <NuxtLink 
              to="/policies/cookie"
              class="text-cyan-400 hover:text-cyan-300 underline"
              @click="hideBanner"
            >
              Learn more
            </NuxtLink>
          </p>
          <div class="flex items-center gap-3">
            <button 
              type="button"
              class="px-4 py-2 text-sm font-medium text-slate-300 bg-slate-800 border border-slate-600 rounded-full hover:bg-slate-700 hover:text-white transition-colors"
              @click="showModal = true"
            >
              Customize
            </button>
            <button 
              type="button"
              class="px-4 py-2 text-sm font-medium text-black bg-gradient-to-r from-cyan-400 to-cyan-500 rounded-full hover:from-cyan-500 hover:to-cyan-600 transition-all shadow-lg shadow-cyan-500/25"
              @click="handleAcceptAll"
            >
              Accept
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 第二步：完整选择弹窗 -->
    <Transition name="cookie-fade">
      <div 
        v-if="showModal" 
        class="fixed inset-0 z-[10000] flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
        @click.self="showModal = false"
      >
        <div class="bg-slate-900 border border-slate-700 rounded-2xl shadow-2xl shadow-cyan-500/10 max-w-lg w-full max-h-[90vh] overflow-y-auto">
          <!-- Header -->
          <div class="flex items-center justify-between p-6 pb-4">
            <h2 class="text-xl font-bold text-white">Your Cookie Preferences</h2>
            <button 
              type="button"
              class="text-slate-400 hover:text-white transition-colors"
              @click="showModal = false"
              aria-label="Close"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Description -->
          <div class="px-6 pb-4">
            <p class="text-sm text-slate-400 leading-relaxed">
              We use cookies to improve your experience on this website. You may choose which types 
              of cookies to allow and change your preferences at any time. Disabling cookies may 
              impact your experience on this website. You can learn more by viewing our 
              <NuxtLink to="/policies/cookie" class="text-cyan-400 hover:text-cyan-300 hover:underline" @click="hideAll">
                Cookie Policy
              </NuxtLink>.
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
                class="mt-1 w-4 h-4 accent-cyan-500 rounded cursor-not-allowed"
              />
              <div>
                <label for="essential" class="text-sm font-semibold text-white">Essential Cookies</label>
                <p class="text-xs text-slate-500 mt-0.5">Cookies required to enable basic website functionality.</p>
              </div>
            </div>

            <!-- Performance Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="performance" 
                v-model="preferences.performance"
                class="mt-1 w-4 h-4 accent-cyan-500 rounded cursor-pointer"
              />
              <div>
                <label for="performance" class="text-sm font-semibold text-white cursor-pointer">Performance Cookies</label>
                <p class="text-xs text-slate-500 mt-0.5">Cookies used to understand how the website is being used.</p>
              </div>
            </div>

            <!-- Preference Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="preference" 
                v-model="preferences.preference"
                class="mt-1 w-4 h-4 accent-cyan-500 rounded cursor-pointer"
              />
              <div>
                <label for="preference" class="text-sm font-semibold text-white cursor-pointer">Preference Cookies</label>
                <p class="text-xs text-slate-500 mt-0.5">Cookies that are used to enhance the functionality of the website.</p>
              </div>
            </div>

            <!-- Advertising Cookies -->
            <div class="flex items-start gap-3">
              <input 
                type="checkbox" 
                id="advertising" 
                v-model="preferences.advertising"
                class="mt-1 w-4 h-4 accent-cyan-500 rounded cursor-pointer"
              />
              <div>
                <label for="advertising" class="text-sm font-semibold text-white cursor-pointer">Advertising Cookies</label>
                <p class="text-xs text-slate-500 mt-0.5">Cookies used to deliver advertising that is more relevant to your interests.</p>
              </div>
            </div>
          </div>

          <!-- Buttons -->
          <div class="flex flex-wrap gap-3 p-6 pt-0">
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-medium text-slate-300 bg-slate-800 border border-slate-600 rounded-lg hover:bg-slate-700 hover:text-white transition-colors"
              @click="handleSavePreferences"
            >
              Save Preferences
            </button>
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-medium text-slate-300 bg-slate-800 border border-slate-600 rounded-lg hover:bg-slate-700 hover:text-white transition-colors"
              @click="handleRejectAll"
            >
              Reject All Cookies
            </button>
            <button 
              type="button"
              class="px-4 py-2.5 text-sm font-medium text-black bg-gradient-to-r from-cyan-400 to-cyan-500 rounded-lg hover:from-cyan-500 hover:to-cyan-600 transition-all shadow-lg shadow-cyan-500/25"
              @click="handleAcceptAll"
            >
              Accept All Cookies
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

const COOKIE_CONSENT_KEY = 'tanzanite_cookie_consent'

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
</style>
