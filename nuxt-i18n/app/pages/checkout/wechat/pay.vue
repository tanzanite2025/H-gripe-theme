<template>
  <section class="wechat-pay-page">
    <div class="wechat-pay-panel">
      <Icon
        :name="statusIcon"
        class="wechat-pay-icon"
        :class="status === 'success' ? 'text-emerald-300' : status === 'error' ? 'text-rose-300' : 'text-sky-300'"
      />
      <h1>{{ title }}</h1>
      <p>{{ message }}</p>
      <p v-if="orderNumber" class="wechat-pay-order">{{ orderNumber }}</p>

      <div v-if="qrDataUrl && status !== 'success'" class="wechat-pay-qr-wrap">
        <img :src="qrDataUrl" :alt="t('checkout.wechatPay.qrAlt')" class="wechat-pay-qr" />
      </div>

      <div class="wechat-pay-actions">
        <button
          v-if="status !== 'success'"
          class="wechat-pay-button wechat-pay-button--primary"
          type="button"
          :disabled="isPolling"
          @click="pollPayment"
        >
          {{ isPolling ? t('checkout.wechatPay.actions.checking') : t('checkout.wechatPay.actions.checkNow') }}
        </button>
        <NuxtLink class="wechat-pay-button" :to="localePath('/')">
          {{ t('checkout.paypalReturn.actions.continueShopping') }}
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n, useLocalePath, useRoute } from '#imports'
import { ApiRequestError } from '~/composables/useApiRequest'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { useWeChatPayment, type WeChatPaymentSession } from '~/composables/useWeChatPayment'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const auth = useAuth()
const { clearCart } = useCart()
const { createWeChatOrder, confirmWeChatOrder, createWeChatQrDataUrl } = useWeChatPayment()

const status = ref<'loading' | 'waiting' | 'success' | 'error'>('loading')
const message = ref(t('checkout.wechatPay.messages.loading'))
const qrDataUrl = ref('')
const isPolling = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const firstQueryValue = (value: unknown) => Array.isArray(value) ? String(value[0] || '') : String(value || '')
const orderNumber = computed(() => firstQueryValue(route.query.order_number).trim())
const storageKey = computed(() => orderNumber.value ? `checkout:wechat:${orderNumber.value}` : '')

const title = computed(() => {
  if (status.value === 'success') return t('checkout.wechatPay.title.success')
  if (status.value === 'error') return t('checkout.wechatPay.title.error')
  if (status.value === 'waiting') return t('checkout.wechatPay.title.waiting')
  return t('checkout.wechatPay.title.loading')
})

const statusIcon = computed(() => {
  if (status.value === 'success') return 'lucide:circle-check'
  if (status.value === 'error') return 'lucide:circle-alert'
  if (status.value === 'waiting') return 'lucide:qr-code'
  return 'lucide:loader-circle'
})

const loadStoredSession = (): WeChatPaymentSession | null => {
  if (!import.meta.client || !storageKey.value) return null
  const raw = window.sessionStorage.getItem(storageKey.value)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as WeChatPaymentSession
    return parsed?.payment_url ? parsed : null
  } catch {
    window.sessionStorage.removeItem(storageKey.value)
    return null
  }
}

const storeSession = (session: WeChatPaymentSession) => {
  if (!import.meta.client || !storageKey.value) return
  window.sessionStorage.setItem(storageKey.value, JSON.stringify(session))
}

const clearStoredSession = () => {
  if (!import.meta.client || !storageKey.value) return
  window.sessionStorage.removeItem(storageKey.value)
}

const isPaidStatus = (value: string) => String(value || '').toUpperCase() === 'SUCCESS'

const completePayment = () => {
  clearStoredSession()
  clearCart()
  status.value = 'success'
  message.value = t('checkout.wechatPay.messages.success')
  stopPolling()
}

const pollPayment = async (): Promise<boolean> => {
  if (!orderNumber.value || isPolling.value || status.value === 'success') return status.value === 'success'
  isPolling.value = true
  try {
    const result = await confirmWeChatOrder(
      orderNumber.value,
      `wechat-confirm-${orderNumber.value}`,
    )
    if (isPaidStatus(result.status)) {
      completePayment()
      return true
    }
    status.value = 'waiting'
    message.value = t('checkout.wechatPay.messages.pending')
    return false
  } catch (error) {
    if (error instanceof ApiRequestError && error.status === 409) {
      status.value = 'waiting'
      message.value = t('checkout.wechatPay.messages.pending')
      return false
    }
    status.value = 'error'
    message.value = error instanceof Error ? error.message : t('checkout.wechatPay.messages.failed')
    stopPolling()
    return false
  } finally {
    isPolling.value = false
  }
}

const stopPolling = () => {
  if (!pollTimer) return
  clearInterval(pollTimer)
  pollTimer = null
}

const startPolling = () => {
  stopPolling()
  pollTimer = setInterval(() => {
    void pollPayment()
  }, 5000)
}

onMounted(async () => {
  try {
    const user = await auth.ensureSession()
    if (!user) {
      throw new Error(t('checkout.wechatPay.messages.loginRequired'))
    }
    if (!orderNumber.value) {
      throw new Error(t('checkout.wechatPay.messages.missingData'))
    }

    const session = loadStoredSession() || await createWeChatOrder({
      orderNumber: orderNumber.value,
      idempotencyKey: `wechat-create-${orderNumber.value}`,
    })
    storeSession(session)
    qrDataUrl.value = await createWeChatQrDataUrl(session.payment_url || '')
    status.value = 'waiting'
    message.value = t('checkout.wechatPay.messages.scan')
    const isPaymentComplete = await pollPayment()
    if (!isPaymentComplete) {
      startPolling()
    }
  } catch (error) {
    status.value = 'error'
    message.value = error instanceof Error ? error.message : t('checkout.wechatPay.messages.failed')
    stopPolling()
  }
})

onUnmounted(stopPolling)
</script>

<style scoped>
.wechat-pay-page {
  min-height: 72vh;
  display: grid;
  place-items: center;
  padding: 8rem 1.25rem 4rem;
  background: #000;
  color: #fff;
}

.wechat-pay-panel {
  width: min(100%, 520px);
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.86);
  padding: 2rem;
  text-align: center;
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.42);
}

.wechat-pay-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 1rem;
}

.wechat-pay-panel h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
}

.wechat-pay-panel p {
  margin: 0.75rem 0 0;
  color: rgba(255, 255, 255, 0.72);
}

.wechat-pay-order {
  font-family: var(--tz-font-ui);
  color: #fff !important;
}

.wechat-pay-qr-wrap {
  display: inline-flex;
  padding: 0.75rem;
  margin-top: 1.5rem;
  border-radius: 8px;
  background: #fff;
}

.wechat-pay-qr {
  width: min(64vw, 260px);
  aspect-ratio: 1 / 1;
}

.wechat-pay-actions {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1.5rem;
}

.wechat-pay-button {
  border: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 8px;
  padding: 0.7rem 1rem;
  color: #fff;
  font-size: 0.875rem;
  font-weight: 700;
}

.wechat-pay-button:disabled {
  opacity: 0.6;
  cursor: wait;
}

.wechat-pay-button--primary {
  background: #fff;
  color: #020617;
}
</style>
