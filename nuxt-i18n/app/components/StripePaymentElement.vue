<template>
  <section class="stripe-payment-element space-y-3" aria-live="polite">
    <div
      ref="paymentContainer"
      class="min-h-[160px] rounded-xl border border-white/10 bg-black/20 p-3"
      :class="{ 'opacity-60': isConfirming }"
    />

    <p v-if="errorMessage" class="rounded-lg border border-rose-300/25 bg-rose-300/10 px-3 py-2 text-xs text-rose-100">
      {{ errorMessage }}
    </p>

    <button
      type="button"
      class="w-full rounded-xl bg-white px-4 py-3 text-sm font-semibold text-slate-900 transition hover:brightness-95 disabled:cursor-not-allowed disabled:opacity-50"
      :disabled="!isReady || isConfirming || disabled"
      @click="confirmPayment"
    >
      {{ isConfirming ? confirmingLabel : confirmLabel }}
    </button>
  </section>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useStripePayment, type StripeConfirmationResult, type StripePaymentSession } from '~/composables/useStripePayment'

const props = defineProps<{
  session: StripePaymentSession
  confirmLabel: string
  confirmingLabel: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  (event: 'confirmed', result: StripeConfirmationResult): void
  (event: 'error', message: string): void
}>()

const paymentContainer = ref<HTMLElement | null>(null)
const isReady = ref(false)
const isConfirming = ref(false)
const errorMessage = ref('')
const { mount, confirm, destroy } = useStripePayment()

const mountPaymentElement = async () => {
  if (!paymentContainer.value) return

  isReady.value = false
  errorMessage.value = ''
  try {
    await mount(paymentContainer.value, props.session)
    isReady.value = true
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unable to load secure payment form'
    errorMessage.value = message
    emit('error', message)
  }
}

const confirmPayment = async () => {
  if (!isReady.value || isConfirming.value) return

  isConfirming.value = true
  errorMessage.value = ''
  try {
    const result = await confirm(window.location.href)
    emit('confirmed', result)
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Stripe payment could not be confirmed'
    errorMessage.value = message
    emit('error', message)
  } finally {
    isConfirming.value = false
  }
}

watch(
  () => [props.session.clientSecret, props.session.publishableKey],
  async () => {
    await nextTick()
    await mountPaymentElement()
  },
  { immediate: true },
)

onMounted(() => {
  void mountPaymentElement()
})

onBeforeUnmount(() => {
  destroy()
})
</script>
