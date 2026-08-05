<template>
  <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden xl:grid-cols-[minmax(0,1fr)_380px]">
    <AdminTablePanel :loading="loading" scroll-body>
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

    <aside class="min-h-0 overflow-auto rounded-[24px] border border-dashed border-border/80 bg-card p-4">
      <div class="mb-4">
        <h2 class="text-sm font-black uppercase italic tracking-tight">新增保护控制</h2>
        <p class="mt-1 text-xs text-muted-foreground">只影响匹配范围内的新订单与新支付启动，不会退款、切换通道或影响已创建支付。</p>
      </div>

      <form class="space-y-3" @submit.prevent="createControl">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">动作</span>
          <select v-model="form.action" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="force_3ds">强制 3DS</option>
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

<script setup>
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
const protectionPolicy = ref({
  max_control_duration_hours: 168,
  max_pause_payment_duration_hours: 24,
  max_global_pause_payment_duration_hours: 2,
})
const controls = ref([])
const selectedControl = ref(null)
const auditLogs = ref([])
const pendingCreatePayload = ref(null)
const confirmDialog = reactive({
  open: false,
  mode: 'create',
  action: 'force_3ds',
  scopeLabel: '',
  controlId: '',
})
const form = reactive({
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

const formatDate = (value) => value ? new Date(value).toLocaleString('zh-CN') : '-'
const statusLabel = (status) => ({ active: '生效中', expired: '已过期', revoked: '已撤销' }[status] || '未知')
const actionLabel = (action) => ({ force_3ds: '强制 3DS', pause_payment: '暂停新支付' }[action] || action || '-')
const actionDescription = (action) => ({
  force_3ds: '匹配范围内的新 Stripe 支付将要求更强的银行认证。',
  pause_payment: '匹配范围内的新订单创建与新支付启动会被拒绝，已创建或已完成的支付不受影响。',
}[action] || '')
const actionSubmitLabel = computed(() => form.action === 'pause_payment' ? '启用暂停新支付' : '启用强制 3DS')
const currentMaxDurationHours = computed(() => {
  if (form.action !== 'pause_payment') return numericPolicyValue('max_control_duration_hours', 168)
  if (form.scope_type === 'global') return numericPolicyValue('max_global_pause_payment_duration_hours', 2)
  return numericPolicyValue('max_pause_payment_duration_hours', 24)
})
const maxExpiryInput = computed(() => expiryInputValue(currentMaxDurationHours.value))
const formatScope = (control) => {
  if (!control || control.scope_type === 'global') return '全局'
  const labels = { provider: '提供商', country: '国家/地区', payment_method: '支付方式' }
  return `${labels[control.scope_type] || control.scope_type}: ${control.scope_value}`
}

function numericPolicyValue(key, fallback) {
  const value = Number(protectionPolicy.value?.[key])
  return Number.isFinite(value) && value > 0 ? value : fallback
}

function defaultExpiry(hours = 24) {
  const value = new Date(Date.now() + hours * 60 * 60 * 1000)
  value.setMinutes(0, 0, 0)
  return toDateTimeLocalInputValue(value)
}

function expiryInputValue(hours) {
  return toDateTimeLocalInputValue(new Date(Date.now() + hours * 60 * 60 * 1000))
}

function toDateTimeLocalInputValue(value) {
  return value.toISOString().slice(0, 16)
}

function clampExpiryToPolicy() {
  const maxTime = new Date(maxExpiryInput.value)
  const current = new Date(form.expires_at)
  if (Number.isNaN(current.getTime()) || current > maxTime) {
    form.expires_at = defaultExpiry(currentMaxDurationHours.value)
  }
}

const refresh = async () => {
  loading.value = true
  try {
    const payload = await paymentRiskApi.listProtectionControls(true)
    enabled.value = payload.enabled !== false
    protectionPolicy.value = { ...protectionPolicy.value, ...(payload.policy || {}) }
    controls.value = payload.controls || []
    if (selectedControl.value) {
      selectedControl.value = controls.value.find((control) => control.id === selectedControl.value.id) || null
    }
  } finally {
    loading.value = false
  }
}

const selectControl = async (control) => {
  selectedControl.value = control
  const payload = await paymentRiskApi.listProtectionControlAudit(control.id)
  auditLogs.value = payload.logs || []
}

const createControl = () => {
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

const submitCreateControl = async () => {
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

const openRevokeConfirmation = () => {
  if (!selectedControl.value) return
  confirmDialog.mode = 'revoke'
  confirmDialog.action = selectedControl.value.action
  confirmDialog.scopeLabel = formatScope(selectedControl.value)
  confirmDialog.controlId = selectedControl.value.id
  confirmDialog.open = true
}

const setConfirmDialogOpen = (open) => {
  confirmDialog.open = open
  if (!open && !saving.value) {
    pendingCreatePayload.value = null
  }
}

const revokeControl = async () => {
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

const confirmProtectionAction = async () => {
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
