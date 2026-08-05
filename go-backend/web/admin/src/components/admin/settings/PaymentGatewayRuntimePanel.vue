<template>
  <div class="space-y-4">
    <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
      <div class="rounded-2xl border bg-muted/30 p-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Gateway Runtime</p>
            <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">支付服务商就绪状态</h3>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            :disabled="loading"
            @click="emit('refresh')"
          >
            <RefreshCw :class="['size-3.5', loading ? 'animate-spin' : '']" />
            刷新
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
                {{ gateway.label }}
              </span>
              <CheckCircle2 v-if="gatewayRuntimeStatus(gateway.value)?.production_ready" class="size-4 text-emerald-500" />
              <XCircle
                v-else-if="gatewayRuntimeStatus(gateway.value) && !gatewayRuntimeStatus(gateway.value)?.webhook_supported"
                class="size-4 text-rose-500"
              />
              <AlertTriangle v-else class="size-4 text-amber-500" />
            </div>
            <p class="mt-2 text-xs leading-relaxed text-muted-foreground">
              {{ gateway.description }}
            </p>
            <span class="mt-3 inline-flex rounded-full border px-2 py-0.5 text-[11px] font-black" :class="runtimeStatusBadgeClass(gatewayRuntimeStatus(gateway.value))">
              {{ runtimeStatusLabel(gatewayRuntimeStatus(gateway.value), loading) }}
            </span>
          </button>
        </div>
      </div>

      <div class="rounded-2xl border bg-muted/30 p-4">
        <p class="text-xs font-black uppercase tracking-widest text-muted-foreground/60">Selected Gateway</p>
        <div class="mt-3 space-y-3">
          <div>
            <div class="text-sm font-black text-foreground">{{ paymentGatewayLabel(selectedGateway) }}</div>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">
              {{ paymentGatewayDescription(selectedGateway) }}
            </p>
          </div>

          <div class="flex items-center justify-between gap-3 rounded-xl border bg-background/70 px-3 py-2.5">
            <div>
              <span class="text-xs font-bold text-foreground">后台首选测试模式</span>
              <p class="mt-0.5 text-xs text-muted-foreground">保存为后台记录；真实支付环境看 runtime。</p>
            </div>
            <Switch :model-value="testMode" aria-label="支付测试模式" @update:model-value="emit('update:testMode', $event)" />
          </div>

          <div class="flex items-center justify-between text-xs">
            <span class="text-muted-foreground">后台记录</span>
            <span
              class="rounded-full px-2 py-0.5 font-black"
              :class="testMode ? 'bg-amber-500/10 text-amber-600 dark:text-amber-300' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-300'"
            >
              {{ testMode ? '测试' : '生产' }}
            </span>
          </div>

          <div class="flex items-center justify-between text-xs">
            <span class="text-muted-foreground">Runtime Source</span>
            <span class="font-bold text-foreground">{{ runtime?.runtime_source || 'environment' }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="grid items-start gap-4 lg:grid-cols-2">
      <div class="rounded-2xl border bg-muted/30 p-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Readiness Detail</p>
            <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">生产就绪检查</h3>
          </div>
          <span class="rounded-full border px-2.5 py-1 text-[11px] font-black" :class="runtimeStatusBadgeClass(selectedRuntimeStatus)">
            {{ runtimeStatusLabel(selectedRuntimeStatus, loading) }}
          </span>
        </div>

        <div v-if="loading" class="mt-4 flex h-28 items-center justify-center text-xs text-muted-foreground">
          <RefreshCw class="mr-2 size-4 animate-spin" />
          正在检查支付运行配置
        </div>

        <div v-else-if="selectedRuntimeStatus" class="mt-4 space-y-4">
          <div class="grid gap-3 md:grid-cols-3">
            <div class="rounded-xl border bg-background/70 p-3">
              <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Environment</p>
              <p class="mt-1 text-sm font-black text-foreground">{{ selectedRuntimeStatus.environment || 'unknown' }}</p>
            </div>
            <div class="rounded-xl border bg-background/70 p-3">
              <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Credentials</p>
              <p class="mt-1 text-sm font-black" :class="selectedRuntimeStatus.configured ? 'text-emerald-500' : 'text-amber-500'">
                {{ selectedRuntimeStatus.configured ? '已配置' : '缺字段' }}
              </p>
            </div>
            <div class="rounded-xl border bg-background/70 p-3">
              <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Webhook</p>
              <p class="mt-1 text-sm font-black" :class="selectedRuntimeStatus.webhook_configured ? 'text-emerald-500' : 'text-amber-500'">
                {{ selectedRuntimeStatus.webhook_configured ? '已配置' : '缺配置' }}
              </p>
            </div>
          </div>

          <div class="rounded-xl border bg-background/70 p-3">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Callback URL</p>
              <div class="flex items-center gap-1">
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="size-8"
                  aria-label="复制支付回调地址"
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
                  探测
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
            <p class="text-xs font-black text-amber-800 dark:text-amber-100">缺失字段</p>
            <p class="mt-1 text-xs leading-relaxed text-amber-800/80 dark:text-amber-100/75">
              {{ selectedRuntimeStatus.missing.join(', ') }}
            </p>
          </div>

          <div v-if="selectedRuntimeStatus.blockers?.length" class="rounded-xl border border-rose-500/20 bg-rose-500/10 p-3">
            <p class="text-xs font-black text-rose-700 dark:text-rose-100">生产阻塞</p>
            <ul class="mt-1 space-y-1 text-xs leading-relaxed text-rose-700/85 dark:text-rose-100/80">
              <li v-for="blocker in selectedRuntimeStatus.blockers" :key="blocker">{{ blocker }}</li>
            </ul>
          </div>

          <div v-if="selectedRuntimeStatus.warnings?.length" class="rounded-xl border border-amber-500/20 bg-amber-500/10 p-3">
            <p class="text-xs font-black text-amber-800 dark:text-amber-100">运行警告</p>
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
          选择一个支付服务商后查看生产就绪检查。
        </div>
      </div>

      <PaymentGatewaySecureConfigPanel
        :selected-gateway="selectedGateway"
        :status="selectedRuntimeStatus"
        :secret-store-configured="runtime?.secret_store_configured === true"
        :can-edit="canEdit"
        @saved="emit('refresh')"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Copy,
  RefreshCw,
  ShieldCheck,
  XCircle,
} from '@lucide/vue'
import PaymentGatewaySecureConfigPanel from '@/components/admin/settings/PaymentGatewaySecureConfigPanel.vue'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import axios from '@/utils/axios'

const props = defineProps({
  selectedGateway: { type: String, default: '' },
  testMode: { type: Boolean, default: true },
  runtime: { type: Object, default: null },
  loading: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
})

const emit = defineEmits(['update:selectedGateway', 'update:testMode', 'refresh'])
const checkingCallback = ref(false)
const callbackCheckResult = ref(null)

const paymentGatewayOptions = [
  { value: 'stripe', label: 'Stripe', description: 'Cards, wallets and international card checkout.' },
  { value: 'paypal', label: 'PayPal', description: 'PayPal account checkout and express payments.' },
  { value: 'alipay', label: '支付宝', description: '适合人民币和跨境支付宝收款场景。' },
  { value: 'wechat', label: '微信支付', description: '适合微信生态内的扫码和小程序支付。' },
]

const paymentGatewayOption = (value) => paymentGatewayOptions.find((gateway) => gateway.value === value)
const paymentGatewayLabel = (value) => paymentGatewayOption(value)?.label || '未选择支付网关'
const paymentGatewayDescription = (value) => paymentGatewayOption(value)?.description || '请选择一个支付服务商后查看 runtime 检查。'
const gatewayRuntimeStatus = (value) => (props.runtime?.gateways || []).find((gateway) => gateway.provider === value)
const selectedRuntimeStatus = computed(() => gatewayRuntimeStatus(props.selectedGateway))

const runtimeStatusLabel = (status, loading) => {
  if (loading) return '检查中'
  if (!status) return '未知'
  if (status.production_ready) return '生产就绪'
  if (!status.webhook_supported) return '锁定'
  if (status.configured || status.webhook_configured) return '需补配置'
  return '缺配置'
}

const runtimeStatusBadgeClass = (status) => {
  if (!status) return 'border-border bg-muted text-muted-foreground'
  if (status.production_ready) return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
  if (!status.webhook_supported) return 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-200'
  if (status.configured || status.webhook_configured) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-200'
  return 'border-border bg-muted text-muted-foreground'
}

const gatewayCardClass = (value, isSelected) => {
  const status = gatewayRuntimeStatus(value)
  if (isSelected) return 'border-admin-selected-border bg-admin-selected-soft shadow-[var(--admin-control-selected-surface-shadow)]'
  if (status?.production_ready) return 'border-emerald-500/20 bg-emerald-500/5 hover:border-emerald-500/35'
  if (status && !status.webhook_supported) return 'border-rose-500/20 bg-rose-500/5 hover:border-rose-500/35'
  if (status?.configured || status?.webhook_configured) return 'border-amber-500/20 bg-amber-500/5 hover:border-amber-500/35'
  return 'border-border bg-background/70 hover:border-admin-selected-border hover:bg-admin-selected-soft'
}

const copyText = async (value) => {
  if (!value || typeof navigator === 'undefined' || !navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(value)
    toast.success('支付回调地址已复制')
  } catch (error) {
    console.error('Failed to copy payment callback URL:', error)
    toast.error('复制失败，请手动复制')
  }
}

const checkCallback = async () => {
  if (!props.selectedGateway || !selectedRuntimeStatus.value?.callback_url) return
  checkingCallback.value = true
  callbackCheckResult.value = null
  try {
    const response = await axios.post(`/api/admin/settings/payment-runtime/${props.selectedGateway}/callback-check`)
    const result = response.data?.data || response.data
    callbackCheckResult.value = result
    if (result?.route_reachable && result?.expected_signature_failure) {
      toast.success('支付回调路由可达')
    } else if (result?.route_reachable) {
      toast.warning('支付回调有响应，但返回状态需检查')
    } else {
      toast.error('支付回调未命中可用路由')
    }
  } catch (error) {
    console.error('Failed to probe payment callback URL:', error)
    callbackCheckResult.value = null
    toast.error(error?.response?.data?.message || '支付回调探测失败')
  } finally {
    checkingCallback.value = false
  }
}

const callbackCheckResultClass = (result) => {
  if (result?.route_reachable && result?.expected_signature_failure) {
    return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-800 dark:text-emerald-100'
  }
  if (result?.route_reachable) {
    return 'border-amber-500/20 bg-amber-500/10 text-amber-800 dark:text-amber-100'
  }
  return 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-100'
}

const callbackCheckResultTitle = (result) => {
  if (result?.route_reachable && result?.expected_signature_failure) return '回调路由可达'
  if (result?.route_reachable) return '回调有响应'
  return '回调不可达'
}

const callbackCheckResultMessage = (result) => {
  if (!result?.transport_reachable) return result?.error || '没有收到 HTTP 响应，请检查域名、CDN/WAF、Nginx/Caddy 和防火墙。'
  if (!result?.route_reachable) return '已收到 HTTP 响应，但没有命中支付 webhook 路由，请检查反向代理路径和后端路由。'
  if (result?.expected_signature_failure) return '已命中支付 webhook 路由；本次探测没有真实签名，返回签名失败属于预期。'
  return '已命中支付 webhook 路由，但状态码不是预期的签名失败，请检查服务端日志。'
}

watch(() => props.selectedGateway, () => {
  callbackCheckResult.value = null
})

const FieldList = defineComponent({
  props: {
    title: { type: String, required: true },
    fields: { type: Array, default: () => [] },
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
        props.fields?.length ? null : h('span', { class: 'text-xs text-muted-foreground' }, '暂无'),
      ]),
    ])
  },
})
</script>
