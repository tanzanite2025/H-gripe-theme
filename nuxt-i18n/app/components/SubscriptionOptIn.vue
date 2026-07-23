<template>
  <form class="space-y-2" @submit.prevent="handleSubmit">
    <label v-if="label" class="block text-xs font-medium tz-text-secondary mb-2 tracking-wide uppercase text-center">
      {{ label }}
    </label>
    <div class="flex gap-2">
      <input
        type="email"
        v-model="email"
        :placeholder="placeholder"
        class="flex-1 h-9 px-4 rounded-full bg-[#0b1020]/60 border-none text-white placeholder:text-slate-500 text-sm shadow-[0_0_0_1px_rgba(0,0,0,0.6)] focus:shadow-[0_0_0_2px_rgba(0,0,0,0.9)] focus:outline-none transition-all"
        :disabled="loading"
        required
        autocomplete="email"
      />
      <button
        type="submit"
        class="h-9 px-5 rounded-full bg-white text-[#0b1020] font-semibold text-sm shadow-[0_4px_14px_rgba(0,0,0,0.45)] hover:shadow-[0_5px_16px_rgba(0,0,0,0.5)] active:scale-[0.98] transition-all flex items-center justify-center disabled:opacity-70 disabled:cursor-not-allowed whitespace-nowrap"
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
      class="text-xs text-emerald-400 text-center mt-2 font-medium"
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
interface SubscriptionSubmitResponse {
  message?: string
  data?: unknown
  error?: string
  success?: boolean
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

const email = ref('')
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

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
