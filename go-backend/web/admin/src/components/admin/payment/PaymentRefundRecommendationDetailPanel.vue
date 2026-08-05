<template>
  <aside class="min-h-0 overflow-auto rounded-[24px] border border-dashed border-border/80 bg-card p-4">
    <div class="mb-4">
      <h2 class="text-sm font-black uppercase italic tracking-tight">退款建议处理</h2>
      <p class="mt-1 text-xs text-muted-foreground">建议处理与本地待处理退款分开记录，支付渠道退款需人工后续执行。</p>
    </div>

    <div v-if="recommendation" class="space-y-3">
      <dl class="grid grid-cols-2 gap-2 text-xs">
        <div class="rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Recommendation</dt>
          <dd class="mt-1 font-mono">#{{ recommendation.id }}</dd>
        </div>
        <div class="rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Status</dt>
          <dd class="mt-1"><StatusPill :status="recommendation.status" /></dd>
        </div>
        <div class="rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Provider</dt>
          <dd class="mt-1 font-mono uppercase">{{ recommendation.provider || '-' }}</dd>
        </div>
        <div class="rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Amount</dt>
          <dd class="mt-1 font-mono">{{ formatMoney(recommendation.recommended_amount, recommendation.currency) }}</dd>
        </div>
        <div class="rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Order</dt>
          <dd class="mt-1 font-mono">{{ recommendation.order_id ? `#${recommendation.order_id}` : '-' }}</dd>
        </div>
        <div class="rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Refund Draft</dt>
          <dd class="mt-1 font-mono">{{ recommendation.linked_refund_id ? `#${recommendation.linked_refund_id}` : '-' }}</dd>
        </div>
        <div class="col-span-2 rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Reason</dt>
          <dd class="mt-1 whitespace-pre-wrap">{{ recommendation.reason || '-' }}</dd>
        </div>
        <div class="col-span-2 rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Payment Reference</dt>
          <dd class="mt-1 break-all font-mono">{{ paymentReference }}</dd>
        </div>
        <div v-if="recommendation.refund_created_at" class="col-span-2 rounded-xl bg-muted/40 p-3">
          <dt class="font-black uppercase text-muted-foreground">Refund Draft Created</dt>
          <dd class="mt-1 font-mono">{{ formatDate(recommendation.refund_created_at) }}</dd>
        </div>
      </dl>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">DECISION / 处理结果</span>
        <select
          :value="decisionStatus"
          class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"
          :disabled="recommendation.status !== 'pending'"
          @change="$emit('update:decisionStatus', $event.target.value)"
        >
          <option value="pending">保持待处理</option>
          <option value="accepted">采纳建议</option>
          <option value="dismissed">驳回建议</option>
          <option value="cancelled">取消</option>
        </select>
      </label>
      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">DECISION NOTES / 决策备注</span>
        <Textarea
          :model-value="decisionNotes"
          rows="5"
          :disabled="recommendation.status !== 'pending'"
          placeholder="记录订单核验、客户沟通、是否后续创建人工退款"
          @update:model-value="$emit('update:decisionNotes', $event)"
        />
      </label>
      <Button class="w-full rounded-full font-black uppercase tracking-wider" :disabled="recommendation.status !== 'pending' || saving" @click="$emit('save-decision')">
        <CheckCircle2 class="size-4" />
        保存建议处理
      </Button>

      <Dialog v-model:open="draftDialogOpen">
        <Button
          type="button"
          variant="outline"
          class="w-full rounded-full font-black uppercase tracking-wider"
          :disabled="!canCreateDraft || draftSaving"
          @click="openDraftDialog"
        >
          <FilePlus2 class="size-4" />
          生成待处理退款
        </Button>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>生成待处理退款</DialogTitle>
            <DialogDescription>
              此操作只创建本地 pending 退款记录，不会调用 Stripe、PayPal 或其他支付渠道退款接口。
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-3">
            <label class="block space-y-1">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">AMOUNT / 金额</span>
              <Input v-model="draftAmount" inputmode="decimal" :placeholder="recommendedAmountPlaceholder" />
            </label>
            <label class="block space-y-1">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">REASON / 退款原因</span>
              <Textarea v-model="draftReason" rows="4" placeholder="默认使用风险建议原因，可填写人工审核结论" />
            </label>
            <label class="flex items-start gap-2 rounded-2xl border border-dashed border-border/80 p-3 text-xs font-bold text-muted-foreground">
              <input v-model="draftConfirmed" type="checkbox" class="mt-0.5 size-4 accent-primary" />
              我确认只生成本地待处理退款，由人工后续核对并执行支付渠道退款。
            </label>
          </div>
          <DialogFooter>
            <Button variant="outline" class="rounded-full" @click="draftDialogOpen = false">
              取消
            </Button>
            <Button class="rounded-full font-black uppercase tracking-wider" :disabled="!draftConfirmed || draftSaving" @click="submitDraft">
              <FilePlus2 class="size-4" />
              创建 pending 退款
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog v-if="recommendation.linked_refund_id" v-model:open="executionDialogOpen">
        <Button
          type="button"
          variant="outline"
          class="w-full rounded-full border-rose-500/30 text-rose-700 font-black uppercase tracking-wider hover:bg-rose-500/10"
          :disabled="executionSaving || executionCompleted"
          @click="openExecutionDialog"
        >
          <Send class="size-4" />
          {{ executionCompleted ? '渠道退款已提交' : '执行支付渠道退款' }}
        </Button>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>执行支付渠道退款</DialogTitle>
            <DialogDescription>
              此操作会调用支付渠道退款接口，并把本地 pending 退款标记为 completed。
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-3 text-xs text-muted-foreground">
            <div class="grid grid-cols-2 gap-2">
              <div class="rounded-xl bg-muted/40 p-3">
                <p class="font-black uppercase">Refund Draft</p>
                <p class="mt-1 font-mono">#{{ recommendation.linked_refund_id }}</p>
              </div>
              <div class="rounded-xl bg-muted/40 p-3">
                <p class="font-black uppercase">Amount</p>
                <p class="mt-1 font-mono">{{ formatMoney(recommendation.recommended_amount, recommendation.currency) }}</p>
              </div>
            </div>
            <label class="flex items-start gap-2 rounded-2xl border border-dashed border-rose-500/30 bg-rose-500/10 p-3 font-bold text-rose-700">
              <input v-model="executionConfirmed" type="checkbox" class="mt-0.5 size-4 accent-rose-600" />
              我确认已核对订单、金额、客户沟通和争议状态，并理解此操作会调用支付渠道退款接口。
            </label>
          </div>
          <DialogFooter>
            <Button variant="outline" class="rounded-full" @click="executionDialogOpen = false">
              取消
            </Button>
            <Button class="rounded-full bg-rose-600 font-black uppercase tracking-wider text-white hover:bg-rose-700" :disabled="!executionConfirmed || executionSaving" @click="submitExecution">
              <Send class="size-4" />
              确认执行
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <p v-if="!canCreateDraft && !recommendation.linked_refund_id" class="rounded-2xl border border-dashed border-amber-500/30 bg-amber-500/10 p-3 text-xs font-bold text-amber-700">
        当前建议缺少订单或交易关联，需先在支付风险事件中补齐映射后再生成退款。
      </p>
    </div>
    <div v-else class="rounded-2xl border border-dashed border-border/80 p-6 text-center text-sm text-muted-foreground">
      选择左侧建议后处理。
    </div>
  </aside>
</template>

<script setup>
import { computed, defineComponent, h, ref, watch } from 'vue'
import { CheckCircle2, FilePlus2, Send } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

const StatusPill = defineComponent({
  props: { status: { type: String, default: '' } },
  setup(props) {
    const tone = computed(() => {
      if (['approved', 'accepted', 'won', 'succeeded'].includes(props.status)) return 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'
      if (['rejected', 'dismissed', 'lost', 'needs_response', 'warning_needs_response'].includes(props.status)) return 'border-rose-500/25 bg-rose-500/10 text-rose-700'
      if (['pending', 'under_review', 'processing'].includes(props.status)) return 'border-amber-500/25 bg-amber-500/10 text-amber-700'
      return 'border-border bg-muted text-muted-foreground'
    })
    return () => h('span', { class: ['inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-black uppercase tracking-wider', tone.value] }, props.status || '-')
  },
})

const props = defineProps({
  recommendation: { type: Object, default: null },
  decisionStatus: { type: String, default: 'pending' },
  decisionNotes: { type: String, default: '' },
  saving: { type: Boolean, default: false },
  draftSaving: { type: Boolean, default: false },
  executionSaving: { type: Boolean, default: false },
  executionCompleted: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:decisionStatus',
  'update:decisionNotes',
  'save-decision',
  'create-draft',
  'execute-refund',
])

const draftDialogOpen = ref(false)
const draftAmount = ref('')
const draftReason = ref('')
const draftConfirmed = ref(false)
const executionDialogOpen = ref(false)
const executionConfirmed = ref(false)

const paymentReference = computed(() => {
  const item = props.recommendation || {}
  return item.provider_payment_id || item.payment_intent_id || item.charge_id || '-'
})

const canCreateDraft = computed(() => {
  const item = props.recommendation
  return Boolean(
    item &&
    !item.linked_refund_id &&
    ['pending', 'accepted'].includes(item.status) &&
    item.order_id &&
    item.transaction_id,
  )
})

const recommendedAmountPlaceholder = computed(() => {
  const amount = Number(props.recommendation?.recommended_amount || 0)
  return amount > 0 ? amount.toFixed(2) : '输入退款金额'
})

const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const formatMoney = (amount, currency = '') => {
  const value = Number(amount || 0)
  const normalizedCurrency = String(currency || '').trim().toUpperCase()
  try {
    if (!normalizedCurrency) throw new Error('missing currency')
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: normalizedCurrency }).format(value)
  } catch {
    return `${normalizedCurrency || '币种缺失'} ${value.toFixed(2)}`
  }
}

const resetDraftForm = () => {
  const amount = Number(props.recommendation?.recommended_amount || 0)
  draftAmount.value = amount > 0 ? amount.toFixed(2) : ''
  draftReason.value = ''
  draftConfirmed.value = false
}

const openDraftDialog = () => {
  if (!canCreateDraft.value) return
  resetDraftForm()
  draftDialogOpen.value = true
}

const submitDraft = () => {
  if (!canCreateDraft.value || !draftConfirmed.value) return
  emit('create-draft', {
    amount: Number(draftAmount.value || 0),
    reason: draftReason.value.trim(),
    decision_notes: props.decisionNotes,
    confirm: true,
  })
}

const openExecutionDialog = () => {
  executionConfirmed.value = false
  executionDialogOpen.value = true
}

const submitExecution = () => {
  if (!props.recommendation?.linked_refund_id || !executionConfirmed.value) return
  emit('execute-refund', {
    refund_id: props.recommendation.linked_refund_id,
    confirm: true,
  })
}

watch(() => props.recommendation?.id, () => {
  draftDialogOpen.value = false
  executionDialogOpen.value = false
  resetDraftForm()
  executionConfirmed.value = false
})

watch(() => props.draftSaving, (saving, previous) => {
  if (!saving && previous) {
    draftDialogOpen.value = false
    draftConfirmed.value = false
  }
})

watch(() => props.executionSaving, (saving, previous) => {
  if (!saving && previous) {
    executionDialogOpen.value = false
    executionConfirmed.value = false
  }
})
</script>
