<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      :title="pageTitle"
      :description="pageDescription"
    />

    <div class="min-h-0 flex-1 overflow-auto">
      <div class="space-y-6">
        <PaymentGatewayRuntimePanel
          v-model:selected-gateway="selectedGateway"
          :runtime="paymentRuntime"
          :loading="loadingPaymentRuntime"
          :can-edit="canEdit"
          :providers="visibleProviders"
          @refresh="fetchPaymentRuntime"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import PaymentGatewayRuntimePanel from '@/components/admin/settings/PaymentGatewayRuntimePanel.vue'
import { useAdminI18n } from '@/i18n'
import { getPaymentChannelLabel, normalizePaymentChannelKey } from '@/lib/paymentChannels'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'
import type { PaymentGatewayRuntime } from '@/components/admin/settings/settingsTypes'

const props = withDefaults(defineProps<{
  provider?: string
}>(), {
  provider: '',
})

const authStore = useAuthStore()
const { t } = useAdminI18n()

const paymentRuntime = ref<PaymentGatewayRuntime | null>(null)
const loadingPaymentRuntime = ref(false)
const selectedGateway = ref('stripe')

const canEdit = computed(() => authStore.hasPermission('settings:edit'))
const providerKey = computed(() => {
  return normalizePaymentChannelKey(String(props.provider || ''))
})
const providerLabel = computed(() => getPaymentChannelLabel(providerKey.value))
const visibleProviders = computed(() => (providerKey.value ? [providerKey.value] : []))
const pageTitle = computed(() => providerKey.value ? `${providerLabel.value} 接入` : t('settings.paymentTitle'))
const pageDescription = computed(() => (
  providerKey.value
    ? `管理 ${providerLabel.value} 收款凭据、回调和运行状态。`
    : t('settings.paymentDescription')
))

const fetchPaymentRuntime = async () => {
  loadingPaymentRuntime.value = true
  try {
    const response = await axios.get('/api/admin/settings/payment-runtime')
    paymentRuntime.value = response.data?.data || response.data || null
  } catch (error) {
    console.error('Failed to fetch payment runtime:', error)
    paymentRuntime.value = null
  } finally {
    loadingPaymentRuntime.value = false
  }
}

watch([paymentRuntime, providerKey], ([runtime, provider]) => {
  if (provider) {
    selectedGateway.value = provider
    return
  }

  const firstGateway = runtime?.gateways?.find((gateway) => gateway.provider)?.provider || ''
  if (!firstGateway) return

  const availableGateways = new Set((runtime?.gateways || []).map((gateway) => String(gateway.provider || '').trim()).filter(Boolean))
  if (!selectedGateway.value || !availableGateways.has(selectedGateway.value)) {
    selectedGateway.value = firstGateway
  }
}, { immediate: true })

onMounted(fetchPaymentRuntime)
</script>
