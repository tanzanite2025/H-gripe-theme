<template>
  <form class="subscription-opt-in space-y-2" @submit.prevent="handleSubmit">
    <div ref="turnstileContainer" class="sr-only" aria-hidden="true"></div>
    <label v-if="label" class="block text-xs font-medium tz-text-secondary mb-2 tracking-wide uppercase text-center">
      {{ label }}
    </label>
    <div class="subscription-opt-in__control">
      <input
        type="email"
        v-model="email"
        :placeholder="placeholder"
        class="subscription-opt-in__input rounded-full bg-[var(--tz-input-surface)] border-none tz-text-primary placeholder:tz-text-muted text-sm shadow-md focus:shadow-md focus:outline-none transition-all"
        :disabled="loading"
        required
        autocomplete="email"
      />
      <button
        type="submit"
        class="subscription-opt-in__button flex shrink-0 items-center justify-center rounded-full bg-white text-sm font-semibold text-[#0b1020] shadow-md transition-all hover:shadow-md active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-70"
        :disabled="loading"
      >
        <span v-if="loading" class="flex items-center gap-1.5">
          <svg class="animate-spin h-3.5 w-3.5 text-[#0b1020]" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ loadingText }}
        </span>
        <span v-else>{{ buttonLabel }}</span>
      </button>
    </div>

    <p
      v-if="successMessage"
      class="text-xs text-emerald-600 text-center mt-2 font-medium"
    >
      {{ successMessage }}
    </p>

    <p
      v-if="errorMessage"
      class="text-xs text-red-400 text-center mt-2"
    >
      {{ errorMessage }}
    </p>
  </form>
</template>

<script setup lang="ts">
import { loadTurnstileScript } from '~/utils/security/trustedScriptUrl'

interface SubscriptionSubmitResponse {
  message?: string
  data?: unknown
  error?: string
  success?: boolean
}

type TurnstileApi = {
  render: (
    element: HTMLElement,
    options: {
      sitekey: string
      size?: 'invisible'
      action?: string
      callback?: (token: string) => void
      'error-callback'?: () => void
      'expired-callback'?: () => void
    },
  ) => string | number
  execute: (widgetId: string | number) => void
  reset: (widgetId: string | number) => void
}

type WindowWithTurnstile = Window & {
  turnstile?: TurnstileApi
}

const props = withDefaults(
  defineProps<{
    label?: string
    placeholder?: string
    buttonLabel?: string
    loadingText?: string
    endpointPath?: string
  }>(),
  {
    label: '',
    placeholder: 'Enter your email',
    buttonLabel: 'Subscribe',
    loadingText: 'Subscribing...',
    endpointPath: '/subscriptions',
  }
)

const emit = defineEmits<{
  (e: 'subscribed', payload: SubscriptionSubmitResponse): void
}>()

const { locale } = useI18n()
const { request } = useApiRequest()
const runtimeConfig = useRuntimeConfig()

const email = ref('')
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')
const turnstileContainer = ref<HTMLElement | null>(null)
let turnstileWidgetId: string | number | null = null
let turnstileLoadPromise: Promise<void> | null = null
let pendingTurnstile: { resolve: (token: string) => void; reject: (error: Error) => void } | null = null

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const loadTurnstile = () => {
  if (!import.meta.client || typeof window === 'undefined') return Promise.resolve()

  const browserWindow = window as WindowWithTurnstile
  if (browserWindow.turnstile) return Promise.resolve()
  if (turnstileLoadPromise) return turnstileLoadPromise

  turnstileLoadPromise = loadTurnstileScript().then(() => undefined)
  return turnstileLoadPromise
}

const ensureTurnstileWidget = async () => {
  const siteKey = String(runtimeConfig.public?.turnstileSiteKey || '').trim()
  if (!siteKey) {
    throw new Error('Verification challenge is not configured')
  }
  if (!import.meta.client || typeof window === 'undefined') return null
  if (!turnstileContainer.value) {
    throw new Error('Verification challenge is not ready')
  }

  await loadTurnstile()

  const browserWindow = window as WindowWithTurnstile
  if (!browserWindow.turnstile) {
    throw new Error('Verification challenge is unavailable')
  }

  if (turnstileWidgetId === null) {
    turnstileWidgetId = browserWindow.turnstile.render(turnstileContainer.value, {
      sitekey: siteKey,
      size: 'invisible',
      action: 'newsletter',
      callback: token => {
        pendingTurnstile?.resolve(token)
        pendingTurnstile = null
      },
      'error-callback': () => {
        pendingTurnstile?.reject(new Error('Verification challenge failed'))
        pendingTurnstile = null
      },
      'expired-callback': () => {
        pendingTurnstile?.reject(new Error('Verification challenge expired'))
        pendingTurnstile = null
      },
    })
  }

  return turnstileWidgetId
}

const executeTurnstileChallenge = async (): Promise<string> => {
  if (!import.meta.client || typeof window === 'undefined') return ''

  const id = await ensureTurnstileWidget()
  if (id === null) return ''

  return new Promise<string>((resolve, reject) => {
    const browserWindow = window as WindowWithTurnstile
    if (!browserWindow.turnstile) {
      reject(new Error('Verification challenge is unavailable'))
      return
    }

    pendingTurnstile = { resolve, reject }
    browserWindow.turnstile.reset(id)
    browserWindow.turnstile.execute(id)
  })
}

async function handleSubmit() {
  successMessage.value = ''
  errorMessage.value = ''

  const value = email.value.trim()
  if (!value || !emailPattern.test(value)) {
    errorMessage.value = '请输入有效的邮箱地址'
    return
  }

  loading.value = true

  try {
    const captchaToken = await executeTurnstileChallenge()
    const data = await request<SubscriptionSubmitResponse>(props.endpointPath, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        accept: 'application/json',
      },
      body: JSON.stringify({
        email: value,
        source: 'website',
        locale: locale.value,
        captcha_token: captchaToken || '',
      }),
    }, 'Subscription failed, please try again later')

    if (data && data.success === false) {
      throw new Error(data?.message || data?.error || '订阅失败，请稍后重试')
    }

    successMessage.value = data?.message || '订阅成功，请前往邮箱确认'
    email.value = ''

    emit('subscribed', data)
  } catch (error: unknown) {
    console.error('Subscription failed', error)
    errorMessage.value = error instanceof Error ? error.message : '订阅失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.subscription-opt-in {
  min-width: 0;
}

.subscription-opt-in__control {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.5rem;
}

.subscription-opt-in__input,
.subscription-opt-in__button {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  height: 2.25rem;
}

.subscription-opt-in__input {
  display: block;
  padding: 0 1rem;
}

.subscription-opt-in__button {
  padding: 0 1.25rem;
  white-space: nowrap;
}

@media (min-width: 640px) {
  .subscription-opt-in__control {
    flex-direction: row;
  }

  .subscription-opt-in__input {
    flex: 1 1 auto;
  }

  .subscription-opt-in__button {
    width: auto;
  }
}
</style>
