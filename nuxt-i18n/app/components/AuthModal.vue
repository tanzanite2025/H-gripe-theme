<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[13000] flex items-end md:items-center justify-center p-0 md:p-4"
        aria-modal="true"
        role="dialog"
        @keydown.esc.prevent="close"
      >
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm" @click="close"></div>
        <Transition name="slide-up">
          <div
            class="relative w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh] bg-black border-2 border-[#6b73ff] rounded-t-3xl md:rounded-2xl shadow-[0_0_30px_rgba(107,115,255,0.3)] text-white flex flex-col pointer-events-auto overflow-hidden"
          >
            <button
              class="absolute right-4 top-4 w-9 h-9 rounded-full border border-white/20 hover:bg-white/10 flex items-center justify-center"
              type="button"
              aria-label="Close"
              @click="close"
            >
              ×
            </button>

            <div class="flex-1 w-full overflow-y-auto px-4 md:px-12 pt-16 pb-10">
              <div class="w-full max-w-[520px] mx-auto">
                <div v-if="!completionState" class="space-y-6">
                  <div class="flex justify-center gap-2">
                    <button
                      type="button"
                      class="px-5 py-2 rounded-full text-sm font-semibold transition-all"
                      :class="mode === 'login'
                        ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black'
                        : 'border border-white/20 text-white/70'"
                      @click="setMode('login')"
                    >
                      {{ $t('auth.signIn', 'Sign in') }}
                    </button>
                    <button
                      type="button"
                      class="px-5 py-2 rounded-full text-sm font-semibold transition-all"
                      :class="mode === 'register'
                        ? 'bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black'
                        : 'border border-white/20 text-white/70'"
                      @click="setMode('register')"
                    >
                      {{ $t('auth.signUp', 'Sign up') }}
                    </button>
                  </div>

                  <div class="space-y-4">
                    <div class="text-center text-sm text-white/70">
                      {{ mode === 'login'
                        ? $t('auth.welcomeBack', 'Welcome back! Choose a method to sign in:')
                        : $t('auth.joinToday', 'Create an account in seconds:') }}
                    </div>

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
                </div>

                <div v-else class="space-y-6 text-center">
                  <div class="flex justify-center">
                    <div class="w-16 h-16 rounded-full bg-white/10 flex items-center justify-center text-3xl text-[#40ffaa]">
                      ✓
                    </div>
                  </div>
                  <div class="space-y-2">
                    <h3 class="text-2xl font-semibold">{{ completionState.title }}</h3>
                    <p class="text-white/70">{{ completionState.message }}</p>
                  </div>
                  <button type="button" class="primary-btn w-full" @click="handleCompletionCta">
                    {{ completionState.ctaLabel }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import { useAuth } from '~/composables/useAuth'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  defaultMode: { type: String as () => 'login' | 'register', default: 'login' }
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
  (event: 'success', payload: { type: 'login' | 'register' }): void
  (event: 'mode-change', value: 'login' | 'register'): void
}>()

const { t: $t } = useI18n()
const auth = useAuth()

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
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease, opacity 0.3s ease;
}
.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(64px);
}

.form-input {
  width: 100%;
  height: 2.75rem;
  padding: 0 0.75rem;
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: white;
}

.form-input::placeholder {
  color: rgba(255, 255, 255, 0.4);
}

.form-input:focus {
  outline: none;
  border-color: #6b73ff;
}

.primary-btn {
  height: 2.75rem;
  border-radius: 0.5rem;
  background: linear-gradient(to right, #40ffaa, #6b73ff);
  color: black;
  font-weight: 600;
  transition: filter 0.2s ease;
}

.primary-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.primary-btn:not(:disabled):hover {
  filter: brightness(1.1);
}

.social-btn {
  width: 3rem;
  height: 3rem;
  border-radius: 9999px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: white;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
}

.social-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}
</style>
