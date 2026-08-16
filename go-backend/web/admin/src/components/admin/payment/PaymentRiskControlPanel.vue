<template>
  <div class="flex min-h-0 flex-1 flex-col gap-3 overflow-hidden">
    <section class="shrink-0 border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Manual Protection Meaning</p>
          <h2 class="mt-1 text-base font-black tracking-tight">人工保护会改变哪一步</h2>
          <p class="mt-1 max-w-4xl text-xs leading-5 text-muted-foreground">
            这是临时覆盖层，不是 3DS 参数配置页。它只匹配新订单/新支付启动，并且必须有过期时间；自适应风险只能把认证强度提高，不能把人工保护降级。
          </p>
        </div>
        <span class="border border-sky-500/25 bg-sky-500/10 px-2 py-1 text-[10px] font-black text-sky-700">
          当前 Stripe 基础模式：{{ stripeModeLabel }}
        </span>
      </div>
      <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-3">
        <div class="border border-border/70 p-3">
          <p class="text-xs font-black">强制 3DS</p>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
            匹配范围内的 Stripe 支付至少提升到 <span class="font-mono font-semibold text-foreground">any</span>；它可能是无感验证，也可能进入银行 challenge，不等于一定出现弹窗。
          </p>
        </div>
        <div class="border border-border/70 p-3">
          <p class="text-xs font-black">暂停新支付</p>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
            在创建 PaymentIntent 前直接拒绝匹配请求；不会自动退款，也不会切换备用渠道，已经创建的支付不受影响。
          </p>
        </div>
        <div class="border border-border/70 p-3">
          <p class="text-xs font-black">实际决策顺序</p>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
            订单可支付检查 → 人工暂停 → 单笔风险预检 → BIN 限流 → 网关熔断 → 自适应 3DS → 创建 PaymentIntent。
          </p>
        </div>
      </div>
      <p class="mt-3 border-t border-dashed border-border/70 pt-3 text-[11px] leading-5 text-muted-foreground">
        当前自适应 3DS：{{ riskConfiguration?.three_ds?.adaptive_enabled ? '已开启' : '未开启' }}；
        Step-up ≥ {{ riskConfiguration?.three_ds?.step_up_risk_score ?? '-' }} 分；
        Challenge ≥ {{ riskConfiguration?.three_ds?.challenge_risk_score ?? '-' }} 分；
        30 天组合风险异常时，{{ riskConfiguration?.monitoring?.auto_step_up_enabled ? '会继续升级高风险支付' : '不会自动升级' }}。
      </p>
    </section>

    <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden max-xl:overflow-auto xl:grid-cols-[minmax(0,1fr)_minmax(460px,520px)] 2xl:grid-cols-[minmax(0,1fr)_520px]">
      <AdminTablePanel class="h-full min-h-0" :loading="loading" scroll-body>
      <template #header>
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-sm font-black uppercase tracking-wide">人工保护控制</h2>
            <p class="mt-1 text-xs text-muted-foreground">按范围临时强制 3DS 或暂停新订单/新支付启动，所有控制都必须设置过期时间。</p>
          </div>
          <Button variant="outline" size="sm" class="rounded-full" :disabled="loading" @click="refresh">
            <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          </Button>
        </div>
      </template>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>动作</TableHead>
            <TableHead>范围</TableHead>
            <TableHead>状态</TableHead>
            <TableHead>过期时间</TableHead>
            <TableHead>原因</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="control in controls"
            :key="control.id"
            class="cursor-pointer"
            :class="selectedControl?.id === control.id ? 'bg-primary/5' : ''"
            @click="selectControl(control)"
          >
            <TableCell class="font-mono text-xs">{{ actionLabel(control.action) }}</TableCell>
            <TableCell class="text-xs">{{ formatScope(control) }}</TableCell>
            <TableCell><StatusPill :status="control.status" /></TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(control.expires_at) }}</TableCell>
            <TableCell class="max-w-[260px] truncate text-xs">{{ control.reason }}</TableCell>
          </TableRow>
          <TableRow v-if="!loading && controls.length === 0">
            <TableCell colspan="5" class="h-24 text-center text-sm text-muted-foreground">暂无人工保护控制</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <aside class="h-full min-h-[460px] min-w-0 overflow-y-auto overscroll-contain rounded-[24px] border border-dashed border-border/80 bg-card p-5">
      <div class="mb-4">
        <h2 class="text-sm font-black uppercase italic tracking-tight">新增保护控制</h2>
        <p class="mt-1 text-xs text-muted-foreground">只影响匹配范围内的新订单与新支付启动，不会退款、切换通道或影响已创建支付。</p>
      </div>

      <form class="space-y-3" @submit.prevent="createControl">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">动作</span>
          <select v-model="form.action" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="force_3ds">强制 3DS（至少 any）</option>
            <option value="pause_payment">暂停新支付</option>
          </select>
          <span class="block text-[11px] text-muted-foreground">{{ actionDescription(form.action) }}</span>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">范围类型</span>
          <select v-model="form.scope_type" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="global">全局</option>
            <option value="provider">支付提供商</option>
            <option value="country">国家/地区</option>
            <option value="payment_method">支付方式</option>
          </select>
        </label>

        <label v-if="form.scope_type !== 'global'" class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">范围值</span>
          <select v-if="form.scope_type === 'provider'" v-model="form.scope_value" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option v-for="provider in providerScopeOptions" :key="provider.value" :value="provider.value">{{ provider.label }}</option>
          </select>
          <Input v-else v-model="form.scope_value" :placeholder="form.scope_type === 'country' ? '例如 US' : '例如 card'" />
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">过期时间</span>
          <Input v-model="form.expires_at" type="datetime-local" :max="maxExpiryInput" required />
          <span class="block text-[11px] text-muted-foreground">当前动作最长 {{ currentMaxDurationHours }} 小时。</span>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">原因</span>
          <Textarea v-model="form.reason" rows="4" maxlength="2000" required placeholder="记录触发背景、观察窗口或关联工单" />
        </label>

        <Button type="submit" class="w-full rounded-full font-black uppercase tracking-wider" :disabled="saving || !enabled">
          <ShieldCheck v-if="form.action === 'force_3ds'" class="size-4" />
          <AlertTriangle v-else class="size-4" />
          {{ actionSubmitLabel }}
        </Button>
      </form>

      <div v-if="!enabled" class="mt-3 rounded-xl border border-amber-500/20 bg-amber-500/10 p-3 text-xs text-amber-800">
        支付保护模块当前未启用。
      </div>

      <section v-if="selectedControl" class="mt-5 border-t border-dashed border-border/70 pt-4">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">控制详情</h3>
            <p class="mt-1 font-mono text-xs">#{{ selectedControl.id }}</p>
          </div>
          <Button
            v-if="selectedControl.active"
            variant="outline"
            size="sm"
            class="rounded-full text-xs font-black text-rose-700"
            :disabled="saving"
            @click="openRevokeConfirmation"
          >
            撤销
          </Button>
        </div>
        <dl class="mt-3 grid grid-cols-2 gap-2 text-xs">
          <div class="rounded-xl bg-muted/40 p-3">
            <dt class="font-black text-muted-foreground">动作</dt>
            <dd class="mt-1">{{ actionLabel(selectedControl.action) }}</dd>
          </div>
          <div class="rounded-xl bg-muted/40 p-3">
            <dt class="font-black text-muted-foreground">范围</dt>
            <dd class="mt-1">{{ formatScope(selectedControl) }}</dd>
          </div>
          <div class="rounded-xl bg-muted/40 p-3">
            <dt class="font-black text-muted-foreground">状态</dt>
            <dd class="mt-1">{{ statusLabel(selectedControl.status) }}</dd>
          </div>
          <div class="col-span-2 rounded-xl bg-muted/40 p-3">
            <dt class="font-black text-muted-foreground">原因</dt>
            <dd class="mt-1 whitespace-pre-wrap">{{ selectedControl.reason }}</dd>
          </div>
        </dl>

        <div class="mt-4">
          <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">操作审计</h3>
          <div v-if="auditLogs.length" class="mt-2 space-y-2">
            <div v-for="log in auditLogs" :key="log.id" class="rounded-xl bg-muted/40 p-3 text-xs">
              <div class="flex items-center justify-between gap-2">
                <span class="font-semibold">{{ log.action === 'create' ? '创建' : '撤销' }} · {{ log.username || `用户 ${log.user_id}` }}</span>
                <span class="font-mono text-muted-foreground">{{ formatDate(log.created_at) }}</span>
              </div>
              <div class="mt-1 text-muted-foreground">{{ log.ip_address || '-' }}</div>
            </div>
          </div>
          <p v-else class="mt-2 text-xs text-muted-foreground">暂无审计记录。</p>
        </div>
      </section>
    </aside>
    </section>
  </div>

  <PaymentProtectionConfirmDialog
    :open="confirmDialog.open"
    :mode="confirmDialog.mode"
    :action="confirmDialog.action"
    :scope-label="confirmDialog.scopeLabel"
    :control-id="confirmDialog.controlId"
    :saving="saving"
    @update:open="setConfirmDialogOpen"
    @confirm="confirmProtectionAction"
  />
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { AlertTriangle, RefreshCw, ShieldCheck } from '@lucide/vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import PaymentProtectionConfirmDialog from '@/components/admin/payment/PaymentProtectionConfirmDialog.vue'
import { paymentRiskApi } from '@/api/paymentRisk'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import type {
  PaymentProtectionAction,
  PaymentProtectionAuditLog,
  PaymentProtectionControl,
  PaymentProtectionControlPayload,
  PaymentProtectionPolicy,
  PaymentProtectionScopeType,
  PaymentRiskConfiguration,
} from './paymentRiskTypes'

const StatusPill = defineComponent({
  props: { status: { type: String, default: '' } },
  setup(statusProps) {
    const tone = computed(() => ({
      active: 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700',
      expired: 'border-amber-500/25 bg-amber-500/10 text-amber-700',
      revoked: 'border-border bg-muted text-muted-foreground',
    }[statusProps.status] || 'border-border bg-muted text-muted-foreground'))
    return () => h('span', { class: ['inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-black uppercase tracking-wider', tone.value] }, statusLabel(statusProps.status))
  },
})

const loading = ref(false)
const saving = ref(false)
const enabled = ref(false)
const riskConfiguration = ref<PaymentRiskConfiguration | null>(null)
const stripeMode = ref('automatic')
const stripeModeLabel = computed(() => ({
  automatic: 'automatic（由 Stripe / 系统判断）',
  any: 'any（要求 3DS，可能无感或挑战）',
  challenge: 'challenge（进入银行挑战路径）',
}[stripeMode.value] || stripeMode.value))
const protectionPolicy = ref<PaymentProtectionPolicy>({
  max_control_duration_hours: 168,
  max_pause_payment_duration_hours: 24,
  max_global_pause_payment_duration_hours: 2,
})
const controls = ref<PaymentProtectionControl[]>([])
const selectedControl = ref<PaymentProtectionControl | null>(null)
const auditLogs = ref<PaymentProtectionAuditLog[]>([])
const pendingCreatePayload = ref<PaymentProtectionControlPayload | null>(null)
const confirmDialog = reactive<{
  open: boolean
  mode: 'create' | 'revoke'
  action: PaymentProtectionAction
  scopeLabel: string
  controlId: string | number
}>({
  open: false,
  mode: 'create',
  action: 'force_3ds',
  scopeLabel: '',
  controlId: '',
})
const form = reactive<{
  action: PaymentProtectionAction
  scope_type: PaymentProtectionScopeType
  scope_value: string
  expires_at: string
  reason: string
}>({
  action: 'force_3ds',
  scope_type: 'global',
  scope_value: 'stripe',
  expires_at: defaultExpiry(),
  reason: '',
})
const providerScopeOptions = [
  { value: 'stripe', label: 'Stripe' },
  { value: 'paypal', label: 'PayPal' },
  { value: 'alipay', label: 'Alipay' },
  { value: 'wechat', label: 'WeChat Pay' },
]

const formatDate = (value: unknown): string => value ? new Date(value as string | number | Date).toLocaleString('zh-CN') : '-'
const statusLabel = (status?: string): string => ({ active: '生效中', expired: '已过期', revoked: '已撤销' }[status || ''] || '未知')
const actionLabel = (action?: string): string => ({ force_3ds: '强制 3DS（至少 any）', pause_payment: '暂停新支付' }[action || ''] || action || '-')
const actionDescription = (action?: string): string => ({
  force_3ds: '匹配范围内的新 Stripe 支付至少提升到 any；可能无感完成，也可能进入银行 challenge。',
  pause_payment: '匹配范围内的新订单创建与新支付启动会被拒绝，已创建或已完成的支付不受影响。',
}[action || ''] || '')
const actionSubmitLabel = computed(() => form.action === 'pause_payment' ? '启用暂停新支付' : '启用强制 3DS')
const currentMaxDurationHours = computed(() => {
  if (form.action !== 'pause_payment') return numericPolicyValue('max_control_duration_hours', 168)
  if (form.scope_type === 'global') return numericPolicyValue('max_global_pause_payment_duration_hours', 2)
  return numericPolicyValue('max_pause_payment_duration_hours', 24)
})
const maxExpiryInput = computed(() => expiryInputValue(currentMaxDurationHours.value))
const formatScope = (control: { scope_type?: string; scope_value?: string } | null): string => {
  if (!control || control.scope_type === 'global') return '全局'
  const labels = { provider: '提供商', country: '国家/地区', payment_method: '支付方式' }
  return `${labels[control.scope_type as keyof typeof labels] || control.scope_type}: ${control.scope_value || ''}`
}

function numericPolicyValue(key: string, fallback: number): number {
  const value = Number(protectionPolicy.value?.[key])
  return Number.isFinite(value) && value > 0 ? value : fallback
}

function defaultExpiry(hours = 24): string {
  const value = new Date(Date.now() + hours * 60 * 60 * 1000)
  value.setMinutes(0, 0, 0)
  return toDateTimeLocalInputValue(value)
}

function expiryInputValue(hours: number): string {
  return toDateTimeLocalInputValue(new Date(Date.now() + hours * 60 * 60 * 1000))
}

function toDateTimeLocalInputValue(value: Date): string {
  return value.toISOString().slice(0, 16)
}

function clampExpiryToPolicy() {
  const maxTime = new Date(maxExpiryInput.value)
  const current = new Date(form.expires_at)
  if (Number.isNaN(current.getTime()) || current > maxTime) {
    form.expires_at = defaultExpiry(currentMaxDurationHours.value)
  }
}

const refresh = async (): Promise<void> => {
  loading.value = true
  try {
    const payload = await paymentRiskApi.listProtectionControls(true)
    enabled.value = payload.enabled !== false
    protectionPolicy.value = { ...protectionPolicy.value, ...(payload.policy || {}) }
    controls.value = payload.controls || []
    try {
      const summary = await paymentRiskApi.getSummary()
      riskConfiguration.value = summary.configuration || null
      stripeMode.value = summary.gateway_runtime?.gateways?.find((gateway) => gateway.provider === 'stripe')?.three_ds_mode || 'automatic'
    } catch (error) {
      console.error('Failed to fetch payment risk policy summary:', error)
      riskConfiguration.value = null
      stripeMode.value = 'automatic'
    }
    if (selectedControl.value) {
      selectedControl.value = controls.value.find((control) => control.id === selectedControl.value.id) || null
    }
  } finally {
    loading.value = false
  }
}

const selectControl = async (control: PaymentProtectionControl): Promise<void> => {
  selectedControl.value = control
  const payload = await paymentRiskApi.listProtectionControlAudit(control.id)
  auditLogs.value = payload.logs || []
}

const createControl = (): void => {
  if (!form.reason.trim()) {
    toast.error('请填写保护原因')
    return
  }
  const expiresAt = new Date(form.expires_at)
  if (Number.isNaN(expiresAt.getTime())) {
    toast.error('请填写有效的过期时间')
    return
  }
  const payload = {
    action: form.action,
    scope_type: form.scope_type,
    scope_value: form.scope_type === 'global' ? '' : form.scope_value.trim(),
    reason: form.reason.trim(),
    expires_at: expiresAt.toISOString(),
    confirm: true,
  }
  pendingCreatePayload.value = payload
  confirmDialog.mode = 'create'
  confirmDialog.action = form.action
  confirmDialog.scopeLabel = formatScope(payload)
  confirmDialog.controlId = ''
  confirmDialog.open = true
}

const submitCreateControl = async (): Promise<void> => {
  if (!pendingCreatePayload.value) return
  saving.value = true
  try {
    await paymentRiskApi.createProtectionControl(pendingCreatePayload.value)
    toast.success('人工保护已启用')
    confirmDialog.open = false
    pendingCreatePayload.value = null
    form.reason = ''
    form.expires_at = defaultExpiry(currentMaxDurationHours.value)
    await refresh()
  } finally {
    saving.value = false
  }
}

const openRevokeConfirmation = (): void => {
  if (!selectedControl.value) return
  confirmDialog.mode = 'revoke'
  confirmDialog.action = selectedControl.value.action
  confirmDialog.scopeLabel = formatScope(selectedControl.value)
  confirmDialog.controlId = selectedControl.value.id
  confirmDialog.open = true
}

const setConfirmDialogOpen = (open: boolean): void => {
  confirmDialog.open = open
  if (!open && !saving.value) {
    pendingCreatePayload.value = null
  }
}

const revokeControl = async (): Promise<void> => {
  if (!selectedControl.value) return
  saving.value = true
  try {
    await paymentRiskApi.revokeProtectionControl(selectedControl.value.id)
    toast.success('人工保护已撤销')
    confirmDialog.open = false
    await refresh()
    if (selectedControl.value) await selectControl(selectedControl.value)
  } finally {
    saving.value = false
  }
}

const confirmProtectionAction = async (): Promise<void> => {
  if (confirmDialog.mode === 'revoke') {
    await revokeControl()
    return
  }
  await submitCreateControl()
}

watch(() => form.scope_type, (value) => {
  if (value === 'global') form.scope_value = ''
  if (value === 'provider' && !providerScopeOptions.some((provider) => provider.value === form.scope_value)) form.scope_value = 'stripe'
  clampExpiryToPolicy()
})

watch(() => form.action, () => {
  clampExpiryToPolicy()
})

watch(currentMaxDurationHours, () => {
  clampExpiryToPolicy()
})

onMounted(refresh)
</script>
