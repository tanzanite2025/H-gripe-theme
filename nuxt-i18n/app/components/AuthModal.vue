<template>
  <Teleport to="body">
    <Transition :name="props.embedded ? 'whatsapp-product-drawer' : 'fade'">
      <div
        v-if="modelValue"
        :class="[
          'fixed inset-0 flex justify-center p-0 md:p-4',
          props.embedded ? 'items-end z-[12000] pointer-events-none' : containerPlacementClass
        ]"
        aria-modal="true"
        role="dialog"
        @keydown.esc.prevent="close"
      >
        <!-- 非 embedded 模式：黑色蒙版 -->
        <div
          v-if="!props.embedded"
          class="absolute inset-0 bg-black/80 backdrop-blur-sm"
          @click="close"
        ></div>

        <!-- 弹窗卡片：对齐 Checkout 弹窗的暗色玻璃风格 -->
        <div
          :class="[
            'auth-modal__panel auth-modal-shell relative w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh] rounded-t-3xl md:rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(15,23,42,0.98),rgba(0,0,0,1))] backdrop-blur-xl border-2 border-[#6b73ff]/40 shadow-[0_0_30px_rgba(107,115,255,0.6)] text-white flex flex-col pointer-events-auto overflow-hidden',
            props.embedded ? 'rounded-2xl' : ''
          ]"
        >
          <!-- 右上角关闭按钮 -->
          <button
            class="absolute right-4 top-4 w-9 h-9 rounded-full border border-white/20 hover:bg-white/10 flex items-center justify-center"
            type="button"
            aria-label="Close"
            @click="close"
          >
            x
          </button>

          <!-- 主体内容 -->
          <div class="auth-modal__body flex-1 w-full overflow-y-auto px-4 md:px-12 pt-16 pb-10">
            <div class="w-full max-w-[520px] mx-auto">
              <!-- 登录 / 注册 表单状态 -->
              <div v-if="!completionState" class="space-y-6">
                <!-- 顶部模式切换按钮 -->
                <div class="flex justify-center gap-2">
                  <button
                    type="button"
                    class="px-5 py-2 rounded-full text-sm font-semibold transition-all"
                    :class="mode === 'login'
                      ? 'bg-[linear-gradient(135deg,#4efce7_0%,#60a5fa_100%)] text-slate-950 shadow-[0_12px_26px_-14px_rgba(15,23,42,1)]'
                      : 'bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white/80 shadow-[0_8px_20px_-12px_rgba(0,0,0,1)]'"
                    @click="setMode('login')"
                  >
                    {{ $t('auth.signIn', 'Sign in') }}
                  </button>
                  <button
                    type="button"
                    class="px-5 py-2 rounded-full text-sm font-semibold transition-all"
                    :class="mode === 'register'
                      ? 'bg-[linear-gradient(135deg,#4efce7_0%,#60a5fa_100%)] text-slate-950 shadow-[0_12px_26px_-14px_rgba(15,23,42,1)]'
                      : 'bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white/80 shadow-[0_8px_20px_-12px_rgba(0,0,0,1)]'"
                    @click="setMode('register')"
                  >
                    {{ $t('auth.signUp', 'Sign up') }}
                  </button>
                </div>

                <div class="space-y-4">
                  <!-- 顶部说明文字 -->
                  <div class="text-center text-sm text-white/70">
                    {{ mode === 'login'
                      ? $t('auth.welcomeBack', 'Welcome back! Choose a method to sign in:')
                      : $t('auth.joinToday', 'Create an account in seconds:') }}
                  </div>

                  <!-- 社交登录按钮 -->
                  <div class="flex justify-center gap-3">
                    <button type="button" class="social-btn" aria-label="Continue with Google">
                      <svg viewBox="0 0 48 48" class="w-5 h-5"><path fill="#FFC107" d="M43.611 20.083H42V20H24v8h11.303C33.565 32.664 29.177 36 24 36c-6.627 0-12-5.373-12-12s5.373-12 12-12c3.059 0 5.842 1.156 7.961 3.039l5.657-5.657C33.797 6.053 29.139 4 24 4 12.955 4 4 12.955 4 24s8.955 20 20 20 20-9 20-20c0-1.341-.138-2.651-.389-3.917z"/><path fill="#FF3D00" d="M6.306 14.691l6.571 4.819C14.655 15.108 19 12 24 12c3.059 0 5.842 1.156 7.961 3.039l5.657-5.657C33.797 6.053 29.139 4 24 4 15.322 4 8.135 9.069 6.306 14.691z"/><path fill="#4CAF50" d="M24 44c5.114 0 9.725-1.961 13.261-5.174l-6.132-5.198C29.16 34.488 26.715 35.5 24 35.5c-5.139 0-9.479-3.335-11.029-8.014l-6.57 5.055C8.122 38.897 15.348 44 24 44z"/><path fill="#1976D2" d="M43.611 20.083H42V20H24v8h11.303c-.685 2.316-2.172 4.285-4.134 5.628l.003-.001 6.132 5.198C39.846 35.896 44 30.5 44 24c0-1.341-.138-2.651-.389-3.917z"/></svg>
                    </button>
                    <button type="button" class="social-btn" aria-label="Continue with X (Twitter)">
                      <svg viewBox="0 0 24 24" class="w-5 h-5" fill="currentColor">
                        <path d="M18.244 2h3.308l-7.227 8.26L22 22h-6.146l-4.807-6.266L5.484 22H2.174l7.73-8.838L2 2h6.277l4.353 5.724L18.244 2z" />
                      </svg>
                    </button>
                    <button type="button" class="social-btn" aria-label="Continue with Facebook">
                      <svg viewBox="0 0 24 24" class="w-5 h-5" fill="currentColor">
                        <path d="M22 12.073C22 6.505 17.523 2 12 2S2 6.505 2 12.073c0 4.991 3.657 9.128 8.438 9.878v-6.987H7.898V12.07h2.54V9.797c0-2.506 1.492-3.89 3.777-3.89 1.094 0 2.238.195 2.238.195v2.463h-1.261c-1.243 0-1.631.771-1.631 1.562v1.941h2.773l-.443 2.894h-2.33v6.987C18.343 21.201 22 17.064 22 12.073z" />
                      </svg>
                    </button>
                  </div>

                  <div class="flex items-center gap-2 text-white/40 text-xs uppercase tracking-[0.2em] justify-center">
                    <span class="flex-1 h-px bg-white/10"></span>
                    <span>{{ $t('auth.orWithEmail', 'or with email') }}</span>
                    <span class="flex-1 h-px bg-white/10"></span>
                  </div>

                  <!-- 登录表单 -->
                  <form v-if="mode === 'login'" @submit.prevent="handleLogin" class="space-y-3">
                    <div>
                      <label class="block text-sm font-medium text-white/80 mb-1">{{ $t('auth.email', 'Email') }}</label>
                      <input
                        type="text"
                        v-model="loginForm.username"
                        required
                        class="form-input"
                        autocomplete="email"
                      />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-white/80 mb-1">{{ $t('auth.password', 'Password') }}</label>
                      <input
                        type="password"
                        v-model="loginForm.password"
                        required
                        class="form-input"
                        autocomplete="current-password"
                      />
                    </div>
                    <label class="flex items-center gap-2 cursor-pointer text-sm text-white/70">
                      <input type="checkbox" v-model="loginForm.remember" class="w-4 h-4" />
                      {{ $t('auth.rememberMe', 'Remember me') }}
                    </label>
                    <button type="submit" :disabled="loginForm.loading" class="primary-btn w-full">
                      {{ loginForm.loading ? $t('auth.signingIn', 'Signing in...') : $t('auth.signIn', 'Sign in') }}
                    </button>
                    <p v-if="loginForm.error" class="text-red-400 text-sm text-center">{{ loginForm.error }}</p>
                    <p class="text-center text-sm text-white/60">
                      {{ $t('auth.dontHaveAccount', "Don't have an account?") }}
                      <button type="button" class="underline-offset-4 underline" @click="setMode('register')">
                        {{ $t('auth.signUpHere', 'Sign up here') }}
                      </button>
                    </p>
                  </form>

                  <!-- 注册表单 -->
                  <form v-else @submit.prevent="handleRegister" class="space-y-3">
                    <div>
                      <label class="block text-sm font-medium text-white/80 mb-1">{{ $t('auth.username', 'Username') }}</label>
                      <input type="text" v-model="registerForm.username" required class="form-input" autocomplete="username" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-white/80 mb-1">{{ $t('auth.email', 'Email') }}</label>
                      <input type="email" v-model="registerForm.email" required class="form-input" autocomplete="email" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-white/80 mb-1">{{ $t('auth.password', 'Password') }}</label>
                      <input type="password" v-model="registerForm.password" required class="form-input" autocomplete="new-password" />
                    </div>
                    <button type="submit" :disabled="registerForm.loading" class="primary-btn w-full">
                      {{ registerForm.loading ? $t('auth.signingUp', 'Signing up...') : $t('auth.signUp', 'Sign up') }}
                    </button>
                    <p v-if="registerForm.error" class="text-red-400 text-sm text-center">{{ registerForm.error }}</p>
                    <p class="text-center text-sm text-white/60">
                      {{ $t('auth.alreadyHaveAccount', 'Already have an account?') }}
                      <button type="button" class="underline-offset-4 underline" @click="setMode('login')">
                        {{ $t('auth.signInHere', 'Sign in here') }}
                      </button>
                    </p>
                  </form>
                </div>

                <div v-if="completionState" class="space-y-6 text-center">
                  <div class="flex justify-center">
                    <div class="w-16 h-16 rounded-full bg-white/10 flex items-center justify-center text-3xl text-[#40ffaa]">
                      &#10003;
                    </div>
                  </div>
                  <div class="space-y-2">
                    <h3 class="text-2xl font-semibold">{{ completionState?.title }}</h3>
                    <p class="text-white/70">{{ completionState?.message }}</p>
                  </div>
                  <button type="button" class="primary-btn w-full" @click="handleCompletionCta">
                    {{ completionState?.ctaLabel }}
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
import { computed, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { useAuth } from '~/composables/useAuth'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  defaultMode: { type: String as () => 'login' | 'register', default: 'login' },
  embedded: { type: Boolean, default: false },
  placement: { type: String as () => 'auto' | 'center' | 'bottom', default: 'auto' }
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
  (event: 'success', payload: { type: 'login' | 'register' }): void
  (event: 'mode-change', value: 'login' | 'register'): void
}>()

const { t: $t } = useI18n()
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

watch(() => props.defaultMode, (val) => {
  mode.value = val
})

watch(() => props.modelValue, (isOpen) => {
  if (!isOpen) {
    resetForms()
  }
})

const resetForms = () => {
  loginForm.value = { username: '', password: '', remember: false, loading: false, error: '' }
  registerForm.value = { username: '', email: '', password: '', loading: false, error: '' }
  completionState.value = null
}

const close = () => {
  emit('update:modelValue', false)
}

const setMode = (next: 'login' | 'register') => {
  mode.value = next
  emit('mode-change', next)
}

const handleLogin = async () => {
  loginForm.value.error = ''
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
      title: $t('auth.loginSuccessTitle', '登录成功'),
      message: $t('auth.loginSuccessMessage', 'Your account data has been synced, click below to continue.'),
      ctaLabel: $t('auth.loginSuccessCta', '好的，返回')
    }
  } catch (error) {
    loginForm.value.error = error instanceof Error ? error.message : 'Login failed'
  } finally {
    loginForm.value.loading = false
  }
}

const handleRegister = async () => {
  registerForm.value.error = ''
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
      title: $t('auth.registerSuccessTitle', '注册成功'),
      message: $t('auth.registerSuccessMessage', '账户已创建，点击下方按钮一键登录并返回。'),
      ctaLabel: $t('auth.registerSuccessCta', '一键登录')
    }
  } catch (error) {
    registerForm.value.error = error instanceof Error ? error.message : 'Registration failed'
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
.auth-modal-shell {
  height: 90vh;
  max-height: 80vh;
}

@supports (height: 100dvh) {
  .auth-modal-shell {
    height: 90dvh;
    max-height: 80dvh;
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

.whatsapp-product-drawer-enter-active,
.whatsapp-product-drawer-leave-active {
  transition: transform 0.3s ease-out, opacity 0.3s ease-out;
}

.whatsapp-product-drawer-enter-from,
.whatsapp-product-drawer-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.whatsapp-product-drawer-enter-to,
.whatsapp-product-drawer-leave-from {
  transform: translateY(0%);
  opacity: 1;
}

.form-input {
  width: 100%;
  height: 2.6rem;
  padding: 0 0.85rem;
  border-radius: 0.75rem;
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(15, 23, 42, 0.96));
  border: none;
  box-shadow:
    0 2px 6px -3px rgba(0, 0, 0, 0.9),
    0 0 8px rgba(15, 23, 42, 0.8);
  color: #e5e7eb;
}

.form-input::placeholder {
  color: rgba(255, 255, 255, 0.4);
}

.form-input:focus {
  outline: none;
  border-color: rgba(56, 189, 248, 0.9);
  box-shadow:
    0 0 0 1px rgba(56, 189, 248, 0.95),
    0 0 12px rgba(15, 23, 42, 0.9);
}

.primary-btn {
  height: 2.75rem;
  border-radius: 9999px;
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #0b1120;
  font-weight: 600;
  box-shadow:
    0 12px 30px -18px rgba(15, 23, 42, 1),
    0 0 18px rgba(37, 99, 235, 0.6);
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
  filter: brightness(1.02);
  box-shadow:
    0 10px 26px -18px rgba(15, 23, 42, 1),
    0 0 20px rgba(56, 189, 248, 0.7);
  transform: translateY(-1px);
}

.social-btn {
  width: 3rem;
  height: 3rem;
  border-radius: 9999px;
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(15, 23, 42, 0.96));
  border: none;
  color: #e5e7eb;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    0 6px 18px -12px rgba(0, 0, 0, 1),
    0 0 10px rgba(15, 23, 42, 0.9);
  transition:
    background 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.15s ease;
}

.social-btn:hover {
  background: linear-gradient(135deg, rgba(31, 41, 55, 0.98), rgba(15, 23, 42, 0.98));
  box-shadow:
    0 8px 20px -12px rgba(0, 0, 0, 1),
    0 0 14px rgba(15, 23, 42, 0.95);
  transform: translateY(-1px);
}

@media (max-width: 420px) {
  .auth-modal__panel {
    height: 94vh;
    max-height: 94vh;
    border-radius: 24px;
  }

  .auth-modal__body {
    padding: 2.5rem 1.25rem 1.25rem;
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
