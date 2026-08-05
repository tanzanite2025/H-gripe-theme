<template>
  <Teleport to="body">
    <Transition :name="props.embedded ? 'wa-drawer' : 'fade'">
      <div
        v-if="modelValue"
        ref="modalRef"
        :class="props.embedded ? 'wa-drawer-mask' : 'fixed inset-0 z-[13000] flex items-center justify-center p-0 md:p-4 tz-mobile-safe-modal-mask'"
        aria-modal="true"
        role="dialog"
        @keydown.esc.prevent="close"
        @click.self="!props.embedded && close()"
      >
        <!-- Backdrop -->
        <!-- Embedded (Mobile Drawer): md:hidden via wa-drawer-backdrop -->
        <!-- Standalone: Visible (bg-black/80) -->
        <div
          v-if="props.embedded"
          class="wa-drawer-backdrop md:hidden"
          @click="close"
        ></div>
        <div
          v-else
          class="absolute inset-0 bg-black/80 backdrop-blur-sm"
          @click="close"
        ></div>

        <!-- Shell -->
        <div
          id="auth-modal-shell"
          :class="[
            props.embedded 
              ? 'wa-drawer-shell auth-modal-shell--drawer'
              : 'auth-modal__panel auth-modal-shell auth-modal-shell--standalone relative w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh] rounded-2xl backdrop-blur-xl border text-white flex flex-col pointer-events-auto overflow-hidden'
          ]"
        >
          <!-- Background Decoration matches other drawers if embedded, or keep original if standalone -->
          <div v-if="props.embedded" class="absolute inset-x-0 top-0 h-[200px] bg-[radial-gradient(circle_at_top,rgba(181,255,109,0.06),transparent_66%)] blur-3xl pointer-events-none z-0"></div>

          <!-- Close Button -->
          <button
            v-if="!props.embedded"
            class="tz-global-close-btn absolute right-4 top-4 z-20"
            type="button"
            :aria-label="t('authModal.actions.close')"
            @click="close"
          >
            <Icon name="lucide:x" class="h-3.5 w-3.5" />
          </button>
          
          <button
            v-else
            class="tz-global-close-btn absolute right-4 top-4 z-20"
            type="button"
            :aria-label="t('authModal.actions.close')"
            @click="close"
          >
            <Icon name="lucide:x" class="h-3.5 w-3.5" />
          </button>

          <!-- Body -->
          <div class="auth-modal__body flex-1 w-full overflow-y-auto px-4 md:px-12 pt-10 pb-6 relative z-10 custom-scrollbar">
            <div class="w-full max-w-[520px] mx-auto">
              <!-- 登录 / 注册 表单状态 -->
              <div v-if="!completionState" class="space-y-6">
                <!-- 顶部模式切换按钮 -->
                <div class="flex justify-center gap-2">
                  <button
                    type="button"
                    class="px-5 py-2 rounded-full text-sm font-semibold transition-all"
                    :class="mode === 'login'
                      ? 'bg-[#B5FF6D] text-[#050505] shadow-[0_12px_26px_-14px_rgba(0,0,0,1)]'
                      : 'bg-white/10 text-white/70 shadow-[0_8px_20px_-12px_rgba(0,0,0,1)]'"
                    @click="setMode('login')"
                  >
                    {{ t('authModal.actions.signIn') }}
                  </button>
                  <button
                    type="button"
                    class="px-5 py-2 rounded-full text-sm font-semibold transition-all"
                    :class="mode === 'register'
                      ? 'bg-[#B5FF6D] text-[#050505] shadow-[0_12px_26px_-14px_rgba(0,0,0,1)]'
                      : 'bg-white/10 text-white/70 shadow-[0_8px_20px_-12px_rgba(0,0,0,1)]'"
                    @click="setMode('register')"
                  >
                    {{ t('authModal.actions.signUp') }}
                  </button>
                </div>

                <!-- Privacy Notice -->
                <div class="mt-3 px-4 py-3 rounded-xl bg-white/[0.055] backdrop-blur-sm shadow-[0_4px_16px_rgba(0,0,0,0.4)]">
                  <p class="text-center text-xs tz-text-secondary leading-relaxed">
                    <Icon name="lucide:shield-check" class="inline-block mr-1 h-3.5 w-3.5 text-[#B5FF6D]" />
                    {{ t('authModal.privacyNotice') }}
                  </p>
                </div>

                <div class="space-y-4">
                  <!-- 顶部说明文字 -->
                  <div class="text-center text-sm tz-text-secondary">
                    {{ mode === 'login'
                      ? t('authModal.intro.login')
                      : t('authModal.intro.register') }}
                  </div>

                  <!-- 社交登录按钮 -->
                  <div class="flex justify-center gap-3">
                    <button 
                      type="button" 
                      class="social-btn" 
                      :aria-label="t('authModal.actions.continueWithGoogle')"
                      :disabled="googleAuthLoading"
                      @click="handleGoogleLogin"
                    >
                      <span v-if="googleAuthLoading" class="animate-spin">
                        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none">
                          <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" opacity="0.25"/>
                          <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                        </svg>
                      </span>
                      <svg v-else viewBox="0 0 48 48" class="w-5 h-5"><path fill="#FFC107" d="M43.611 20.083H42V20H24v8h11.303C33.565 32.664 29.177 36 24 36c-6.627 0-12-5.373-12-12s5.373-12 12-12c3.059 0 5.842 1.156 7.961 3.039l5.657-5.657C33.797 6.053 29.139 4 24 4 12.955 4 4 12.955 4 24s8.955 20 20 20 20-9 20-20c0-1.341-.138-2.651-.389-3.917z"/><path fill="#FF3D00" d="M6.306 14.691l6.571 4.819C14.655 15.108 19 12 24 12c3.059 0 5.842 1.156 7.961 3.039l5.657-5.657C33.797 6.053 29.139 4 24 4 15.322 4 8.135 9.069 6.306 14.691z"/><path fill="#4CAF50" d="M24 44c5.114 0 9.725-1.961 13.261-5.174l-6.132-5.198C29.16 34.488 26.715 35.5 24 35.5c-5.139 0-9.479-3.335-11.029-8.014l-6.57 5.055C8.122 38.897 15.348 44 24 44z"/><path fill="#1976D2" d="M43.611 20.083H42V20H24v8h11.303c-.685 2.316-2.172 4.285-4.134 5.628l.003-.001 6.132 5.198C39.846 35.896 44 30.5 44 24c0-1.341-.138-2.651-.389-3.917z"/></svg>
                    </button>
                    <p v-if="googleAuthError" class="text-red-400 text-xs text-center mt-1">{{ googleAuthError }}</p>
                  </div>

                  <div class="flex items-center gap-2 tz-text-muted text-xs uppercase tracking-[0.2em] justify-center">
                    <span class="flex-1 h-px bg-white/10"></span>
                    <span>{{ t('authModal.divider.email') }}</span>
                    <span class="flex-1 h-px bg-white/10"></span>
                  </div>

                  <!-- 登录表单 -->
                  <form v-if="mode === 'login'" @submit.prevent="handleLogin" class="space-y-3">
                    <div>
                      <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('authModal.fields.email') }}</label>
                      <input
                        type="text"
                        v-model="loginForm.username"
                        required
                        class="form-input"
                        autocomplete="email"
                      />
                    </div>
                    <div>
                      <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('authModal.fields.password') }}</label>
                      <input
                        type="password"
                        v-model="loginForm.password"
                        required
                        class="form-input"
                        autocomplete="current-password"
                      />
                    </div>
                    <label class="flex items-center gap-2 cursor-pointer text-sm tz-text-secondary">
                      <input type="checkbox" v-model="loginForm.remember" class="w-4 h-4" />
                      {{ t('authModal.fields.rememberMe') }}
                    </label>
                    <button type="submit" :disabled="loginForm.loading" class="primary-btn w-full">
                      {{ loginForm.loading ? t('authModal.actions.signingIn') : t('authModal.actions.signIn') }}
                    </button>
                    <p v-if="loginForm.error" class="text-red-400 text-sm text-center">{{ loginForm.error }}</p>
                    <p class="text-center text-sm tz-text-secondary">
                      {{ t('authModal.switch.noAccount') }}
                      <button type="button" class="underline-offset-4 underline" @click="setMode('register')">
                        {{ t('authModal.switch.signUpHere') }}
                      </button>
                    </p>
                  </form>

                  <!-- 注册表单 -->
                  <form v-else @submit.prevent="handleRegister" class="space-y-3">
                    <div>
                      <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('authModal.fields.username') }}</label>
                      <input type="text" v-model="registerForm.username" required class="form-input" autocomplete="username" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('authModal.fields.email') }}</label>
                      <input type="email" v-model="registerForm.email" required class="form-input" autocomplete="email" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium tz-text-secondary mb-1">{{ t('authModal.fields.password') }}</label>
                      <input type="password" v-model="registerForm.password" required class="form-input" autocomplete="new-password" />
                    </div>
                    <button type="submit" :disabled="registerForm.loading" class="primary-btn w-full">
                      {{ registerForm.loading ? t('authModal.actions.signingUp') : t('authModal.actions.signUp') }}
                    </button>
                    <p v-if="registerForm.error" class="text-red-400 text-sm text-center">{{ registerForm.error }}</p>
                    <p class="text-center text-sm tz-text-secondary">
                      {{ t('authModal.switch.hasAccount') }}
                      <button type="button" class="underline-offset-4 underline" @click="setMode('login')">
                        {{ t('authModal.switch.signInHere') }}
                      </button>
                    </p>
                  </form>
                </div>

                <div v-if="completionState" class="space-y-6 text-center">
                  <div class="flex justify-center">
                    <div class="w-16 h-16 rounded-full bg-white/10 flex items-center justify-center text-3xl text-[#B5FF6D]">
                      &#10003;
                    </div>
                  </div>
                  <div class="space-y-2">
                    <h3 class="text-2xl font-semibold">{{ completionTitle }}</h3>
                    <p class="tz-text-secondary">{{ completionMessage }}</p>
                  </div>
                  <button type="button" class="primary-btn w-full" @click="handleCompletionCta">
                    {{ completionCtaLabel }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { useI18n } from '#imports'
import { useAuth } from '~/composables/useAuth'
import { useGoogleAuth } from '~/composables/useGoogleAuth'
import { z } from 'zod'
import { useFocusTrap } from '@vueuse/integrations/useFocusTrap'

const { t } = useI18n()

const authSchema = z.object({
  email: z.string().email(t('authModal.validation.invalidEmail')),
  password: z.string().min(8, t('authModal.validation.passwordMin')).regex(/[A-Z]/, t('authModal.validation.passwordUppercase'))
})

const props = defineProps({
  defaultMode: { type: String as () => 'login' | 'register', default: 'login' },
  embedded: { type: Boolean, default: false },
  placement: { type: String as () => 'auto' | 'center' | 'bottom', default: 'auto' }
})

const modelValue = defineModel<boolean>({ default: false })

const emit = defineEmits<{
  (event: 'success', payload: { type: 'login' | 'register' }): void
  (event: 'mode-change', value: 'login' | 'register'): void
}>()

const auth = useAuth()

const containerPlacementClass = computed(() => {
  if (props.embedded) {
    return 'items-end z-[12000] pointer-events-none'
  }
  switch (props.placement) {
    case 'center':
      return 'items-center z-[13000]'
    case 'bottom':
      return 'items-end z-[13000]'
    default:
      return 'items-end md:items-center z-[13000]'
  }
})

const mode = ref<'login' | 'register'>(props.defaultMode)
const loginForm = ref({ username: '', password: '', remember: false, loading: false, error: '' })
const registerForm = ref({ username: '', email: '', password: '', loading: false, error: '' })
type CompletionState = {
  type: 'login' | 'register'
  title: string
  message: string
  ctaLabel: string
}
const completionState = ref<CompletionState | null>(null)
const completionTitle = computed(() => completionState.value?.title || '')
const completionMessage = computed(() => completionState.value?.message || '')
const completionCtaLabel = computed(() => completionState.value?.ctaLabel || '')

watch(() => props.defaultMode, (val) => {
  mode.value = val
})

const modalRef = ref<HTMLElement | null>(null)
const { activate, deactivate } = useFocusTrap(modalRef)

watch(() => modelValue.value, async (isOpen) => {
  if (isOpen) {
    // delay activation to wait for the DOM element to mount
    setTimeout(() => activate(), 50)
  } else {
    deactivate()
    resetForms()
  }
})

const resetForms = () => {
  loginForm.value = { username: '', password: '', remember: false, loading: false, error: '' }
  registerForm.value = { username: '', email: '', password: '', loading: false, error: '' }
  completionState.value = null
}

const close = () => {
  modelValue.value = false
}

const setMode = (next: 'login' | 'register') => {
  mode.value = next
  emit('mode-change', next)
}

// ============ Google Sign-In Logic ============
const googleAuth = useGoogleAuth()
const googleAuthLoading = ref(false)
const googleAuthError = ref<string | null>(null)

// 处理 Google 登录响应
const handleGoogleCredentialResponse = async (response: { credential: string }) => {
  googleAuthLoading.value = true
  googleAuthError.value = null
  
  try {
    // 发送 ID Token 到后端验证
    await auth.loginWithGoogle(response.credential)
    await auth.ensureSession?.()
    
    // 登录成功
    completionState.value = {
      type: 'login',
      title: t('authModal.completion.loginTitle'),
      message: t('authModal.completion.googleLoginMessage'),
      ctaLabel: t('authModal.completion.loginCta')
    }
  } catch (err) {
    googleAuthError.value = err instanceof Error ? err.message : t('authModal.errors.googleLoginFailed')
    console.error('[AuthModal] Google login failed:', err)
  } finally {
    googleAuthLoading.value = false
  }
}

// 点击 Google 按钮
const handleGoogleLogin = async () => {
  googleAuthError.value = null
  googleAuthLoading.value = true
  
  try {
    const initialized = await googleAuth.initialize(handleGoogleCredentialResponse)
    if (initialized) {
      googleAuth.prompt()
    } else {
      googleAuthError.value = googleAuth.error.value || t('authModal.errors.googleInitFailed')
    }
  } catch (err) {
    googleAuthError.value = err instanceof Error ? err.message : t('authModal.errors.googleInitFailed')
  } finally {
    // 注意：loading 状态将在 handleGoogleCredentialResponse 中关闭
    // 如果用户关闭弹窗，需要手动关闭 loading
    setTimeout(() => {
      if (googleAuthLoading.value && !completionState.value) {
        googleAuthLoading.value = false
      }
    }, 10000) // 10 秒超时
  }
}

const handleLogin = async () => {
  loginForm.value.error = ''
  
  const validation = authSchema.safeParse({
    email: loginForm.value.username,
    password: loginForm.value.password
  })
  
  if (!validation.success) {
    loginForm.value.error = validation.error.issues[0]?.message || t('authModal.errors.loginFailed')
    return
  }

  loginForm.value.loading = true
  try {
    await auth.login({
      username: loginForm.value.username,
      password: loginForm.value.password,
      remember: loginForm.value.remember
    })
    await auth.ensureSession?.()
    completionState.value = {
      type: 'login',
      title: t('authModal.completion.loginTitle'),
      message: t('authModal.completion.loginMessage'),
      ctaLabel: t('authModal.completion.loginCta')
    }
  } catch (error) {
    loginForm.value.error = error instanceof Error ? error.message : t('authModal.errors.loginFailed')
  } finally {
    loginForm.value.loading = false
  }
}

const handleRegister = async () => {
  registerForm.value.error = ''
  
  const validation = authSchema.safeParse({
    email: registerForm.value.email,
    password: registerForm.value.password
  })
  
  if (!validation.success) {
    registerForm.value.error = validation.error.issues[0]?.message || t('authModal.errors.registrationFailed')
    return
  }

  registerForm.value.loading = true
  try {
    await auth.register({
      username: registerForm.value.username,
      email: registerForm.value.email,
      password: registerForm.value.password
    })
    await auth.ensureSession?.()
    completionState.value = {
      type: 'register',
      title: t('authModal.completion.registerTitle'),
      message: t('authModal.completion.registerMessage'),
      ctaLabel: t('authModal.completion.registerCta')
    }
  } catch (error) {
    registerForm.value.error = error instanceof Error ? error.message : t('authModal.errors.registrationFailed')
  } finally {
    registerForm.value.loading = false
  }
}

const handleCompletionCta = async () => {
  if (!completionState.value) return
  await auth.ensureSession?.()
  emit('success', { type: completionState.value.type })
  completionState.value = null
  close()
}
</script>

<style scoped>
/* Standard styles for non-embedded mode */
.auth-modal-shell {
  height: min(90vh, var(--tz-mobile-safe-viewport-height, 90vh));
  max-height: min(80vh, var(--tz-mobile-safe-viewport-height, 80vh));
}

.auth-modal-shell--standalone,
.auth-modal-shell--drawer {
  background: linear-gradient(135deg, #141416 0%, #050505 100%) !important;
  background-image: linear-gradient(135deg, #141416 0%, #050505 100%) !important;
  border-color: rgba(255, 255, 255, 0.14) !important;
  box-shadow: 0 22px 70px -36px rgba(0, 0, 0, 1) !important;
}

.auth-modal-shell--drawer::before,
.auth-modal-shell--drawer::after {
  display: none !important;
}

@supports (height: 100dvh) {
  .auth-modal-shell {
    height: min(90dvh, var(--tz-mobile-safe-viewport-height, 90dvh));
    max-height: min(80dvh, var(--tz-mobile-safe-viewport-height, 80dvh));
  }
}

@media (min-width: 768px) {
  .auth-modal-shell {
    height: 700px;
    max-height: 85vh;
  }

  @supports (height: 100dvh) {
    .auth-modal-shell {
      height: min(700px, 85dvh);
    }
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Custom Scrollbar for Auth Body */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 99px;
}

.auth-modal-shell--standalone .form-input,
.auth-modal-shell--drawer .form-input {
  width: 100%;
  height: 2.6rem;
  padding: 0 0.85rem;
  border-radius: 0.75rem;
  background-color: transparent !important;
  background-image: none !important;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-color: rgba(255, 255, 255, 0.18) !important;
  box-shadow: none;
  color: #f5f5f5 !important;
}

.auth-modal-shell--standalone .form-input::placeholder,
.auth-modal-shell--drawer .form-input::placeholder {
  color: rgba(255, 255, 255, 0.4);
}

.auth-modal-shell--standalone .form-input:focus,
.auth-modal-shell--drawer .form-input:focus {
  outline: none;
  border-color: rgba(255, 255, 255, 0.78) !important;
  background-color: transparent !important;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.28);
}

:global(#auth-modal-shell input.form-input) {
  background-color: transparent !important;
  background-image: none !important;
  border-color: rgba(255, 255, 255, 0.34) !important;
  color: #f5f5f5 !important;
}

:global(#auth-modal-shell input.form-input:focus) {
  border-color: rgba(255, 255, 255, 0.78) !important;
  background-color: transparent !important;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.28) !important;
}

.primary-btn {
  height: 2.75rem;
  border-radius: 9999px;
  background: #ffffff;
  color: #050505;
  font-weight: 600;
  box-shadow:
    0 12px 30px -18px rgba(0, 0, 0, 1),
    0 0 18px rgba(255, 255, 255, 0.12);
  transition:
    filter 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.15s ease;
}

.primary-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.primary-btn:not(:disabled):hover {
  filter: brightness(0.96);
  box-shadow:
    0 10px 26px -18px rgba(0, 0, 0, 1),
    0 0 20px rgba(255, 255, 255, 0.16);
  transform: translateY(-1px);
}

.social-btn {
  width: 3rem;
  height: 3rem;
  border-radius: 9999px;
  background: linear-gradient(135deg, rgba(18, 18, 20, 0.98), rgba(7, 7, 8, 0.96));
  border: none;
  color: #e5e7eb;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    0 6px 18px -12px rgba(0, 0, 0, 1),
    0 0 10px rgba(0, 0, 0, 0.78);
  transition:
    background 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.15s ease;
}

.social-btn:hover {
  background: linear-gradient(135deg, rgba(28, 28, 30, 0.98), rgba(10, 10, 11, 0.98));
  box-shadow:
    0 8px 20px -12px rgba(0, 0, 0, 1),
    0 0 14px rgba(181, 255, 109, 0.12);
  transform: translateY(-1px);
}

@media (max-width: 420px) {
  .auth-modal__panel {
    height: min(94vh, var(--tz-mobile-safe-viewport-height, 94vh));
    max-height: min(94vh, var(--tz-mobile-safe-viewport-height, 94vh));
    border-radius: 24px;
  }

  .auth-modal__body {
    padding: 2.5rem 1.25rem var(--tz-mobile-modal-safe-padding-bottom, 1.25rem);
  }

  .auth-modal__body .space-y-6 {
    gap: 1rem;
  }

  .social-btn {
    width: 2.75rem;
    height: 2.75rem;
  }
}
</style>
