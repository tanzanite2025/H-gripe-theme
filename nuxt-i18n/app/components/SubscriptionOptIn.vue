<template>
  <form class="space-y-4" @submit.prevent="handleSubmit">
    <div>
      <label v-if="label" class="block text-xs font-medium text-slate-400 mb-2 tracking-wide uppercase text-center">
        {{ label }}
      </label>
      <div class="relative">
        <input
          type="email"
          v-model="email"
          :placeholder="placeholder"
          class="w-full h-12 pl-5 pr-4 rounded-full bg-[#0b1020]/60 border border-white/10 text-white placeholder:text-slate-500 text-sm focus:border-sky-500/50 focus:outline-none focus:ring-1 focus:ring-sky-500/50 transition-all"
          :disabled="loading"
          required
          autocomplete="email"
        />
      </div>
    </div>

    <button
      type="submit"
      class="w-full h-12 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-[#0b1020] font-bold text-sm shadow-lg shadow-[#40ffaa]/10 hover:shadow-[#40ffaa]/30 active:scale-[0.98] transition-all flex items-center justify-center disabled:opacity-70 disabled:cursor-not-allowed"
      :disabled="loading"
    >
      <span v-if="loading" class="flex items-center gap-2">
        <svg class="animate-spin h-4 w-4 text-[#0b1020]" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        {{ loadingText }}
      </span>
      <span v-else>{{ buttonLabel }}</span>
    </button>

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
const props = withDefaults(
  defineProps<{
    label?: string
    placeholder?: string
    buttonLabel?: string
    loadingText?: string
    /**
     * 相对于 wpApiBase 的订阅接口路径。
     * 默认指向 \"/tanz/v1/subscribe\"。
     */
    endpointPath?: string
  }>(),
  {
    label: '',
    placeholder: 'Enter your email',
    buttonLabel: 'Subscribe',
    loadingText: 'Subscribing...',
    endpointPath: '/tanz/v1/subscribe',
  }
)

const emit = defineEmits<{
  (e: 'subscribed', payload: any): void
}>()

const config = useRuntimeConfig()

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
    const base = (config.public as { wpApiBase?: string }).wpApiBase || '/wp-json'
    const wpBase = base.replace(/\/$/, '')
    const endpoint = `${wpBase}${props.endpointPath}`

    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email: value }),
    })

    const data = await response.json().catch(() => ({}))

    if (!response.ok || (data && data.success === false)) {
      throw new Error(data?.message || '订阅失败，请稍后重试')
    }

    successMessage.value = data?.message || '订阅成功，请前往邮箱确认'
    email.value = ''

    emit('subscribed', data)
  } catch (error: any) {
    console.error('Subscription failed', error)
    errorMessage.value = error?.message || '订阅失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
