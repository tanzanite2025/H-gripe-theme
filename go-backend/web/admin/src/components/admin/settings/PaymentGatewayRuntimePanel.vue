<template>
  <div class="space-y-4">
    <div class="rounded-2xl border border-admin-selected-border/60 bg-admin-selected-soft/40 p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.gatewayConnection') }}</p>
          <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">{{ t('payment.paymentAccountReady') }}</h3>
          <p class="mt-1 max-w-3xl text-xs leading-relaxed text-muted-foreground">
            {{ t('payment.runtimeDescription') }}
          </p>
        </div>
        <span class="rounded-full border border-admin-selected-border/60 bg-background/80 px-2.5 py-1 text-[11px] font-black text-admin-selected">
          {{ t('payment.selectedGateway', { gateway: paymentGatewayLabel(selectedGateway) }) }}
        </span>
      </div>

      <div class="mt-4 grid gap-3 md:grid-cols-3">
        <div class="rounded-xl border bg-background/70 p-3">
            <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">1 · {{ t('payment.officialBackend') }}</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">{{ t('payment.officialBackendDescription') }}</p>
        </div>
        <div class="rounded-xl border bg-background/70 p-3">
            <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">2 · {{ t('payment.merchantCredentials') }}</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">{{ t('payment.merchantCredentialsDescription') }}</p>
        </div>
        <div class="rounded-xl border bg-background/70 p-3">
            <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">3 · {{ t('payment.webhook') }}</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">{{ t('payment.webhookDescription') }}</p>
        </div>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-2">
        <a
          v-if="selectedGatewayOption?.officialDashboardURL"
          :href="selectedGatewayOption.officialDashboardURL"
          target="_blank"
          rel="noreferrer"
          class="inline-flex items-center gap-2 rounded-md border bg-background px-3 py-2 text-xs font-black text-foreground transition-colors hover:border-admin-selected-border hover:bg-admin-selected-soft"
        >
          <ExternalLink class="size-3.5" />
          {{ t('payment.openOfficialDashboard', { gateway: paymentGatewayLabel(selectedGateway) }) }}
        </a>
        <Button type="button" size="sm" variant="outline" @click="scrollToCredentials">
          <KeyRound class="size-3.5" />
          {{ t('payment.fillCredentials') }}
        </Button>
      </div>

      <div class="mt-4 border-t border-admin-selected-border/40 pt-4">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div>
            <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.boundGateways') }}</p>
            <p class="mt-1 text-xs text-muted-foreground">{{ t('payment.boundGatewaysDescription') }}</p>
          </div>
          <span class="rounded-full border bg-background/80 px-2.5 py-1 text-[11px] font-black text-foreground">
            {{ t('payment.boundCount', { bound: adminBoundGateways.length, total: paymentGatewayOptions.length }) }}
          </span>
        </div>

        <div v-if="adminBoundGateways.length" class="mt-3 flex flex-wrap gap-2">
          <button
            v-for="gateway in adminBoundGateways"
            :key="gateway.provider"
            type="button"
            class="inline-flex items-center gap-2 rounded-lg border bg-background/80 px-3 py-2 text-left transition-colors hover:border-admin-selected-border hover:bg-background"
            :class="selectedGateway === gateway.provider ? 'border-admin-selected-border bg-background shadow-[var(--admin-control-selected-surface-shadow)]' : ''"
            @click="emit('update:selectedGateway', gateway.provider)"
          >
            <span class="text-xs font-black text-foreground">{{ paymentGatewayLabel(gateway.provider) }}</span>
            <span class="text-[11px] font-bold" :class="bindingStatusClass(gateway)">
              {{ bindingStatusLabel(gateway) }}
            </span>
          </button>
        </div>
        <p v-else class="mt-3 rounded-lg border border-dashed px-3 py-2 text-xs text-muted-foreground">
          {{ t('payment.noBoundGateways') }}
        </p>
      </div>

      <p
        v-if="selectedGateway === 'paypal'"
        class="mt-3 rounded-xl border border-sky-500/20 bg-sky-500/10 px-3 py-2 text-xs leading-relaxed text-sky-900 dark:text-sky-100"
      >
        {{ t('payment.paypalBinNote') }}
      </p>
    </div>

    <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
      <div class="rounded-2xl border bg-muted/30 p-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.gatewayRuntime') }}</p>
          <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">{{ t('payment.runtimeReady') }}</h3>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            :disabled="loading"
            @click="emit('refresh')"
          >
            <RefreshCw :class="['size-3.5', loading ? 'animate-spin' : '']" />
            {{ t('common.refresh') }}
          </Button>
        </div>

        <div class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <button
            v-for="gateway in paymentGatewayOptions"
            :key="gateway.value"
            type="button"
            class="group min-h-28 rounded-xl border p-3 text-left transition-all hover:-translate-y-0.5"
            :class="gatewayCardClass(gateway.value, selectedGateway === gateway.value)"
            @click="emit('update:selectedGateway', gateway.value)"
          >
            <div class="flex items-center justify-between gap-2">
              <span class="text-sm font-black" :class="selectedGateway === gateway.value ? 'text-admin-selected' : 'text-foreground'">
                {{ paymentGatewayLabel(gateway.value) }}
              </span>
              <CheckCircle2 v-if="gatewayRuntimeStatus(gateway.value)?.production_ready" class="size-4 text-emerald-500" />
              <XCircle
                v-else-if="gatewayRuntimeStatus(gateway.value) && !gatewayRuntimeStatus(gateway.value)?.webhook_supported"
                class="size-4 text-rose-500"
              />
              <AlertTriangle v-else class="size-4 text-amber-500" />
            </div>
            <p class="mt-2 text-xs leading-relaxed text-muted-foreground">
              {{ paymentGatewayDescription(gateway.value) }}
            </p>
            <span class="mt-3 inline-flex rounded-full border px-2 py-0.5 text-[11px] font-black" :class="runtimeStatusBadgeClass(gatewayRuntimeStatus(gateway.value))">
              {{ runtimeStatusLabel(gatewayRuntimeStatus(gateway.value), loading) }}
            </span>
          </button>
        </div>
      </div>

      <div class="rounded-2xl border bg-muted/30 p-4">
        <p class="text-xs font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.selectedGatewayLabel') }}</p>
        <div class="mt-3 space-y-3">
          <div>
            <div class="text-sm font-black text-foreground">{{ paymentGatewayLabel(selectedGateway) }}</div>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">
              {{ paymentGatewayDescription(selectedGateway) }}
            </p>
          </div>

          <div class="flex items-center justify-between text-xs">
              <span class="text-muted-foreground">{{ t('payment.credentialSource') }}</span>
            <span class="font-bold text-foreground">{{ credentialSourceLabel(runtime?.runtime_source) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="grid items-start gap-4 lg:grid-cols-2">
      <div class="rounded-2xl border bg-muted/30 p-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.readinessDetail') }}</p>
            <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">{{ t('payment.productionReadiness') }}</h3>
          </div>
          <span class="rounded-full border px-2.5 py-1 text-[11px] font-black" :class="runtimeStatusBadgeClass(selectedRuntimeStatus)">
            {{ runtimeStatusLabel(selectedRuntimeStatus, loading) }}
          </span>
        </div>

        <div v-if="loading" class="mt-4 flex h-28 items-center justify-center text-xs text-muted-foreground">
          <RefreshCw class="mr-2 size-4 animate-spin" />
          {{ t('payment.checkingRuntime') }}
        </div>

        <div v-else-if="selectedRuntimeStatus" class="mt-4 space-y-4">
          <div class="grid gap-3 md:grid-cols-3">
            <div class="rounded-xl border bg-background/70 p-3">
              <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.environment') }}</p>
              <p class="mt-1 text-sm font-black text-foreground">{{ selectedRuntimeStatus.environment || 'unknown' }}</p>
            </div>
            <div class="rounded-xl border bg-background/70 p-3">
              <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.credentials') }}</p>
              <p class="mt-1 text-sm font-black" :class="selectedRuntimeStatus.configured ? 'text-emerald-500' : 'text-amber-500'">
                {{ selectedRuntimeStatus.configured ? t('common.configured') : t('payment.missingFields') }}
              </p>
            </div>
            <div class="rounded-xl border bg-background/70 p-3">
              <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.webhookStatus') }}</p>
              <p class="mt-1 text-sm font-black" :class="selectedRuntimeStatus.webhook_configured ? 'text-emerald-500' : 'text-amber-500'">
                {{ selectedRuntimeStatus.webhook_configured ? t('common.configured') : t('payment.missingConfig') }}
              </p>
            </div>
          </div>

          <div class="rounded-xl border bg-background/70 p-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.callbackUrl') }}</p>
              <div class="flex items-center gap-1">
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="size-8"
                  :aria-label="t('payment.copyCallbackUrl')"
                  @click="copyText(selectedRuntimeStatus.callback_url)"
                >
                  <Copy class="size-3.5" />
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  :disabled="checkingCallback || !canEdit"
                  @click="checkCallback"
                >
                  <RefreshCw v-if="checkingCallback" class="size-3.5 animate-spin" />
                  <Activity v-else class="size-3.5" />
                  {{ t('common.probe') }}
                </Button>
              </div>
            </div>
            <p class="mt-1 break-all font-mono text-xs text-foreground">
              {{ selectedRuntimeStatus.callback_url }}
            </p>
          </div>

          <div
            v-if="callbackCheckResult"
            class="rounded-xl border p-3"
            :class="callbackCheckResultClass(callbackCheckResult)"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <p class="text-xs font-black">{{ callbackCheckResultTitle(callbackCheckResult) }}</p>
              <span class="font-mono text-[11px] font-black">
                HTTP {{ callbackCheckResult.status_code || 'N/A' }}
              </span>
            </div>
            <p class="mt-1 text-xs leading-relaxed">
              {{ callbackCheckResultMessage(callbackCheckResult) }}
            </p>
            <p class="mt-2 font-mono text-[11px] opacity-80">
              {{ callbackCheckResult.method || 'POST' }} · {{ callbackCheckResult.duration_ms ?? 0 }}ms
            </p>
          </div>

          <div class="grid gap-3 md:grid-cols-2">
            <FieldList title="Required Runtime Fields" :fields="selectedRuntimeStatus.required_fields" />
            <FieldList title="Configured Fields" :fields="selectedRuntimeStatus.configured_fields" configured />
          </div>

          <div v-if="selectedRuntimeStatus.missing?.length" class="rounded-xl border border-amber-500/20 bg-amber-500/10 p-3">
            <p class="text-xs font-black text-amber-800 dark:text-amber-100">{{ t('payment.missingFields') }}</p>
            <p class="mt-1 text-xs leading-relaxed text-amber-800/80 dark:text-amber-100/75">
              {{ selectedRuntimeStatus.missing.join(', ') }}
            </p>
          </div>

          <div v-if="selectedRuntimeStatus.blockers?.length" class="rounded-xl border border-rose-500/20 bg-rose-500/10 p-3">
            <p class="text-xs font-black text-rose-700 dark:text-rose-100">{{ t('payment.productionBlockers') }}</p>
            <ul class="mt-1 space-y-1 text-xs leading-relaxed text-rose-700/85 dark:text-rose-100/80">
              <li v-for="blocker in selectedRuntimeStatus.blockers" :key="blocker">{{ blocker }}</li>
            </ul>
          </div>

          <div v-if="selectedRuntimeStatus.warnings?.length" class="rounded-xl border border-amber-500/20 bg-amber-500/10 p-3">
            <p class="text-xs font-black text-amber-800 dark:text-amber-100">{{ t('payment.runtimeWarnings') }}</p>
            <ul class="mt-1 space-y-1 text-xs leading-relaxed text-amber-800/80 dark:text-amber-100/75">
              <li v-for="warning in selectedRuntimeStatus.warnings" :key="warning">{{ warning }}</li>
            </ul>
          </div>

          <a
            v-if="selectedRuntimeStatus.documentation_url"
            :href="selectedRuntimeStatus.documentation_url"
            target="_blank"
            rel="noreferrer"
            class="inline-flex items-center gap-2 text-xs font-black text-admin-selected hover:text-admin-selected/80"
          >
            <ShieldCheck class="size-3.5" />
            {{ selectedRuntimeStatus.documentation_label }}
          </a>
        </div>

        <div v-else class="mt-4 rounded-xl border border-dashed p-5 text-sm text-muted-foreground">
          {{ t('payment.noGatewaySelected') }}
        </div>
      </div>

      <div id="payment-gateway-credentials" class="scroll-mt-6">
        <PaymentGatewaySecureConfigPanel
          :selected-gateway="selectedGateway"
          :status="selectedRuntimeStatus"
          :secret-store-configured="runtime?.secret_store_configured === true"
          :can-edit="canEdit"
          @saved="emit('refresh')"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, watch } from 'vue'
import type { PropType } from 'vue'
import { toast } from 'vue-sonner'
import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Copy,
  ExternalLink,
  KeyRound,
  RefreshCw,
  ShieldCheck,
  XCircle,
} from '@lucide/vue'
import PaymentGatewaySecureConfigPanel from '@/components/admin/settings/PaymentGatewaySecureConfigPanel.vue'
import { Button } from '@/components/ui/button'
import axios from '@/utils/axios'
import { useAdminI18n } from '@/i18n'
import type {
  PaymentCallbackCheckResult,
  PaymentGatewayOption,
  PaymentGatewayRuntime,
  PaymentGatewayRuntimeStatus,
} from './settingsTypes'

const props = withDefaults(defineProps<{
  selectedGateway?: string
  runtime?: PaymentGatewayRuntime | null
  loading?: boolean
  canEdit?: boolean
}>(), {
  selectedGateway: '',
  runtime: null,
  loading: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:selectedGateway', value: string): void
  (event: 'refresh'): void
}>()
const { t } = useAdminI18n()
const checkingCallback = ref(false)
const callbackCheckResult = ref<PaymentCallbackCheckResult | null>(null)

const paymentGatewayOptions: PaymentGatewayOption[] = [
  {
    value: 'stripe',
    label: 'Stripe',
    description: 'Cards, wallets and international card checkout.',
    officialDashboardURL: 'https://dashboard.stripe.com/apikeys',
  },
  {
    value: 'paypal',
    label: 'PayPal',
    description: 'PayPal account checkout and express payments.',
    officialDashboardURL: 'https://developer.paypal.com/dashboard/applications',
  },
  {
    value: 'alipay',
    label: '支付宝',
    description: '适合人民币和跨境支付宝收款场景。',
    officialDashboardURL: 'https://open.alipay.com/',
  },
  {
    value: 'wechat',
    label: '微信支付',
    description: '适合微信生态内的扫码和小程序支付。',
    officialDashboardURL: 'https://pay.weixin.qq.com/',
  },
]

const paymentGatewayOption = (value: string): PaymentGatewayOption | undefined => paymentGatewayOptions.find((gateway) => gateway.value === value)
const paymentGatewayLabel = (value: string): string => {
  const key = `payment.${value}`
  return t(key, {}, paymentGatewayOption(value)?.label || t('payment.selectGateway'))
}
const paymentGatewayDescription = (value: string): string => {
  const key = `payment.${value}Description`
  return t(key, {}, t('payment.runtimeCheckDescription'))
}
const selectedGatewayOption = computed(() => paymentGatewayOption(props.selectedGateway))
const gatewayRuntimeStatus = (value: string): PaymentGatewayRuntimeStatus | undefined => (props.runtime?.gateways || []).find((gateway) => gateway.provider === value)
const selectedRuntimeStatus = computed(() => gatewayRuntimeStatus(props.selectedGateway))
const adminBoundGateways = computed(() => (
  (props.runtime?.gateways || []).filter((gateway) => gateway.admin_config_configured === true)
))
const credentialSourceLabel = (source?: string): string => {
  if (source === 'admin-encrypted') return t('payment.adminEncrypted')
  if (source === 'environment') return t('payment.serverEnvironment')
  return source || t('payment.credentialNotConfigured')
}

const scrollToCredentials = (): void => {
  if (typeof document === 'undefined') return
  document.getElementById('payment-gateway-credentials')?.scrollIntoView({
    behavior: 'smooth',
    block: 'start',
  })
}

const bindingStatusLabel = (status?: PaymentGatewayRuntimeStatus | null): string => {
  if (!status) return t('common.unknown')
  if (status.admin_config_readable === false) return t('payment.credentialUnavailable')
  if (status.production_ready) return t('payment.productionReady')
  return t('payment.pendingConfig')
}

const bindingStatusClass = (status?: PaymentGatewayRuntimeStatus | null): string => {
  if (!status) return 'text-muted-foreground'
  if (status.admin_config_readable === false) return 'text-rose-600 dark:text-rose-300'
  if (status.production_ready) return 'text-emerald-600 dark:text-emerald-300'
  return 'text-amber-600 dark:text-amber-300'
}

const runtimeStatusLabel = (status?: PaymentGatewayRuntimeStatus | null, loading = false): string => {
  if (loading) return t('payment.checking')
  if (!status) return t('common.unknown')
  if (status.production_ready) return t('payment.productionReady')
  if (!status.webhook_supported) return t('payment.locked')
  if (status.configured || status.webhook_configured) return t('payment.pendingConfig')
  return t('payment.missingConfig')
}

const runtimeStatusBadgeClass = (status?: PaymentGatewayRuntimeStatus | null): string => {
  if (!status) return 'border-border bg-muted text-muted-foreground'
  if (status.production_ready) return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
  if (!status.webhook_supported) return 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-200'
  if (status.configured || status.webhook_configured) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-200'
  return 'border-border bg-muted text-muted-foreground'
}

const gatewayCardClass = (value: string, isSelected: boolean): string => {
  const status = gatewayRuntimeStatus(value)
  if (isSelected) return 'border-admin-selected-border bg-admin-selected-soft shadow-[var(--admin-control-selected-surface-shadow)]'
  if (status?.production_ready) return 'border-emerald-500/20 bg-emerald-500/5 hover:border-emerald-500/35'
  if (status && !status.webhook_supported) return 'border-rose-500/20 bg-rose-500/5 hover:border-rose-500/35'
  if (status?.configured || status?.webhook_configured) return 'border-amber-500/20 bg-amber-500/5 hover:border-amber-500/35'
  return 'border-border bg-background/70 hover:border-admin-selected-border hover:bg-admin-selected-soft'
}

const copyText = async (value?: string): Promise<void> => {
  if (!value || typeof navigator === 'undefined' || !navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(value)
    toast.success(t('payment.callbackUrlCopied'))
  } catch (error) {
    console.error('Failed to copy payment callback URL:', error)
    toast.error(t('payment.copyFailed'))
  }
}

interface ErrorLike {
  response?: { data?: { message?: string } }
}

const checkCallback = async (): Promise<void> => {
  if (!props.selectedGateway || !selectedRuntimeStatus.value?.callback_url) return
  checkingCallback.value = true
  callbackCheckResult.value = null
  try {
    const response = await axios.post(`/api/admin/settings/payment-runtime/${props.selectedGateway}/callback-check`)
    const result = response.data?.data || response.data
    callbackCheckResult.value = result
    if (result?.route_reachable && result?.expected_signature_failure) {
      toast.success(t('payment.probeSuccess'))
    } else if (result?.route_reachable) {
      toast.warning(t('payment.probeWarning'))
    } else {
      toast.error(t('payment.probeFailure'))
    }
  } catch (error) {
    console.error('Failed to probe payment callback URL:', error)
    callbackCheckResult.value = null
    toast.error((error as ErrorLike)?.response?.data?.message || t('payment.probeError'))
  } finally {
    checkingCallback.value = false
  }
}

const callbackCheckResultClass = (result: PaymentCallbackCheckResult): string => {
  if (result?.route_reachable && result?.expected_signature_failure) {
    return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-800 dark:text-emerald-100'
  }
  if (result?.route_reachable) {
    return 'border-amber-500/20 bg-amber-500/10 text-amber-800 dark:text-amber-100'
  }
  return 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-100'
}

const callbackCheckResultTitle = (result: PaymentCallbackCheckResult): string => {
  if (result?.route_reachable && result?.expected_signature_failure) return t('payment.callbackReachable')
  if (result?.route_reachable) return t('payment.callbackResponded')
  return t('payment.callbackUnavailable')
}

const callbackCheckResultMessage = (result: PaymentCallbackCheckResult): string => {
  if (!result?.transport_reachable) return result?.error || t('payment.callbackTransportError')
  if (!result?.route_reachable) return t('payment.callbackRouteMissed')
  if (result?.expected_signature_failure) return t('payment.callbackExpectedSignature')
  return t('payment.callbackUnexpectedStatus')
}

watch(() => props.selectedGateway, () => {
  callbackCheckResult.value = null
})

const FieldList = defineComponent({
  props: {
    title: { type: String, required: true },
    fields: { type: Array as PropType<string[]>, default: () => [] },
    configured: { type: Boolean, default: false },
  },
  setup(props) {
    return () => h('div', { class: 'rounded-xl border bg-background/70 p-3' }, [
      h('p', { class: 'text-[11px] font-black uppercase tracking-widest text-muted-foreground/60' }, props.title),
      h('div', { class: 'mt-2 flex flex-wrap gap-1.5' }, [
        ...(props.fields || []).map((field) =>
          h('span', {
            key: field,
            class: props.configured
              ? 'rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2 py-0.5 font-mono text-[11px] font-bold text-emerald-700 dark:text-emerald-200'
              : 'rounded-full border bg-muted px-2 py-0.5 font-mono text-[11px] font-bold text-foreground',
          }, String(field))
        ),
        props.fields?.length ? null : h('span', { class: 'text-xs text-muted-foreground' }, t('common.none')),
      ]),
    ])
  },
})
</script>
