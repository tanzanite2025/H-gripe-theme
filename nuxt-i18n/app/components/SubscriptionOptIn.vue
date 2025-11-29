<template>
  <form class="space-y-3" @submit.prevent="handleSubmit">
    <div>
      <label v-if="label" class="block text-sm font-medium text-white/80 mb-1">
        {{ label }}
      </label>
      <input
        type="email"
        v-model="email"
        :placeholder="placeholder"
        class="form-input w-full"
        :disabled="loading"
        required
        autocomplete="email"
      />
    </div>

    <button
      type="submit"
      class="primary-btn w-full"
      :disabled="loading"
    >
      <span v-if="loading">{{ loadingText }}</span>
      <span v-else>{{ buttonLabel }}</span>
    </button>

    <p
      v-if="successMessage"
      class="text-sm text-emerald-400 text-center"
    >
      {{ successMessage }}
    </p>

    <p
      v-if="errorMessage"
      class="text-sm text-red-400 text-center"
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
