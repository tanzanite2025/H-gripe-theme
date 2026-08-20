<template>
  <section v-if="isRefundReviewCase" class="space-y-3 border-t border-dashed pt-5">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="text-sm font-black uppercase text-foreground">退款审批</h3>
        <p class="mt-1 text-xs text-muted-foreground">
          退款审批、待处理退款与支付渠道执行分开记录。
        </p>
      </div>
      <span class="status-mark" :class="refundReviewStatusClass">
        {{ refundReviewStatusLabel }}
      </span>
    </div>

    <div v-if="!canEditRefundReview && !canDecideRefundReview && !refundReview" class="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
      售后单进入“处理中”后才可提交退款审批。
    </div>

    <div v-else class="space-y-4 rounded-lg border bg-muted/15 p-4">
      <div class="grid gap-4 sm:grid-cols-[minmax(0,1fr)_140px]">
        <label class="block space-y-1.5">
          <span class="field-label">PROPOSED AMOUNT / 申请金额</span>
          <Input
            v-model="refundForm.amount"
            type="number"
            min="0"
            step="0.01"
            :max="refundMaximumAmount || undefined"
            :disabled="!canEditRefundReview || refundSubmitting"
          />
          <span class="text-[11px] text-muted-foreground">
            可审批上限 {{ formatMoney(record?.refund_review_maximum_amount, record?.refund_review_currency) }}
          </span>
        </label>
        <div class="space-y-1.5">
          <span class="field-label">CURRENCY / 币种</span>
          <div class="currency-box font-mono">
            {{ record?.refund_review_currency || '-' }}
          </div>
        </div>
      </div>

      <label v-if="canEditRefundReview" class="block space-y-1.5">
        <span class="field-label">REQUEST NOTES / 申请说明</span>
        <Textarea
          v-model="refundForm.requestNotes"
          rows="3"
          maxlength="2000"
          :disabled="refundSubmitting"
          placeholder="说明退款依据、商品处理结果或客户沟通结论"
        />
      </label>
      <div v-else class="rounded-md border border-dashed p-3 text-xs text-muted-foreground">
        <span class="font-bold text-foreground">申请说明：</span>{{ refundReview?.request_notes || '-' }}
      </div>

      <div v-if="refundReview" class="grid gap-3 text-xs sm:grid-cols-2">
        <div class="rounded-md border border-dashed p-3">
          <span class="field-label">CREATED BY / 提交人</span>
          <p class="mt-1 font-bold">{{ refundReview.creator_name || operatorFallback(refundReview.created_by) }}</p>
        </div>
        <div class="rounded-md border border-dashed p-3">
          <span class="field-label">REVIEWED BY / 审批人</span>
          <p class="mt-1 font-bold">{{ refundReview.reviewer_name || operatorFallback(refundReview.reviewed_by_id) }}</p>
        </div>
      </div>

      <div v-if="canEditRefundReview" class="flex justify-end">
        <Button type="button" :disabled="refundSubmitting || !canSaveRefundReview" @click="saveRefundReview">
          <Save :class="['size-4', refundSubmitting ? 'animate-pulse' : '']" />
          {{ refundSubmitting ? '保存中' : refundReview ? '更新审批草稿' : '提交审批草稿' }}
        </Button>
      </div>

      <div v-if="canDecideRefundReview" class="space-y-3 border-t border-dashed pt-4">
        <label class="block space-y-1.5">
          <span class="field-label">DECISION NOTES / 审批说明</span>
          <Textarea
            v-model="refundForm.decisionNotes"
            rows="3"
            maxlength="2000"
            :disabled="refundSubmitting"
            placeholder="填写批准、拒绝或取消的明确原因"
          />
        </label>
        <div class="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" :disabled="refundSubmitting || !canDecide" @click="decideRefundReview('rejected')">
            <X class="size-4" />
            拒绝退款
          </Button>
          <Button type="button" variant="outline" :disabled="refundSubmitting || !canDecide" @click="decideRefundReview('cancelled')">
            取消审批
          </Button>
          <Button type="button" :disabled="refundSubmitting || !canDecide" @click="decideRefundReview('approved')">
            <Check class="size-4" />
            批准退款
          </Button>
        </div>
      </div>

      <div v-if="canCreatePendingRefund" class="space-y-3 border-t border-dashed pt-4">
        <div class="rounded-md border border-dashed bg-background/70 p-3 text-xs text-muted-foreground">
          审批金额 {{ formatMoney(refundReview?.proposed_amount, refundReview?.currency) }} 将写入本地待处理退款单。
        </div>
        <label class="flex items-start gap-2 rounded-md border border-dashed p-3 text-xs font-bold text-muted-foreground">
          <input v-model="refundForm.draftConfirmed" type="checkbox" class="mt-0.5 size-4 accent-primary" :disabled="refundSubmitting" />
          我确认仅生成本地 pending 退款，尚未调用支付渠道退款接口。
        </label>
        <div class="flex justify-end">
          <Button type="button" :disabled="refundSubmitting || !refundForm.draftConfirmed" @click="createPendingRefund">
            <FilePlus2 :class="['size-4', refundSubmitting ? 'animate-pulse' : '']" />
            {{ refundSubmitting ? '创建中' : '生成待处理退款' }}
          </Button>
        </div>
      </div>

      <div v-else-if="refundReview?.linked_refund_id" class="rounded-md border border-dashed bg-background/70 p-3 text-xs text-muted-foreground">
        <span class="font-bold text-foreground">待处理退款：</span>#{{ refundReview.linked_refund_id }}。支付渠道退款仍需在后续步骤中单独确认执行。
      </div>

      <div v-if="refundReview && !canDecideRefundReview" class="rounded-md border border-dashed p-3 text-xs text-muted-foreground">
        <span class="font-bold text-foreground">审批说明：</span>{{ refundReview.decision_notes || '尚未填写' }}
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { Check, FilePlus2, Save, X } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type { AfterSalesCase, AfterSalesRefundReview } from '@/api/afterSales'
import { isAfterSalesRefundType } from '@/lib/afterSalesPresentation'

const props = withDefaults(defineProps<{
  record?: AfterSalesCase | null
  refundSubmitting?: boolean
}>(), {
  record: null,
  refundSubmitting: false,
})

const emit = defineEmits<{
  (event: 'save-refund-review', proposedAmount: number, currency: string, requestNotes: string): void
  (event: 'decide-refund-review', status: 'approved' | 'rejected' | 'cancelled', decisionNotes: string): void
  (event: 'create-pending-refund'): void
}>()

const refundForm = reactive({ amount: '', requestNotes: '', decisionNotes: '', draftConfirmed: false })

const refundReview = computed<AfterSalesRefundReview | null>(() => props.record?.refund_review || null)
const isRefundReviewCase = computed(() => isAfterSalesRefundType(props.record?.type))
const refundMaximumAmount = computed(() => Number(props.record?.refund_review_maximum_amount || 0))
const canEditRefundReview = computed(() => (
  isRefundReviewCase.value &&
  props.record?.status === 'resolving' &&
  (!refundReview.value || refundReview.value.status === 'pending')
))
const canDecideRefundReview = computed(() => canEditRefundReview.value && Boolean(refundReview.value))
const canCreatePendingRefund = computed(() => (
  isRefundReviewCase.value &&
  props.record?.status === 'resolving' &&
  refundReview.value?.status === 'approved' &&
  !refundReview.value.linked_refund_id
))
const canDecide = computed(() => Boolean(refundForm.decisionNotes.trim()))
const canSaveRefundReview = computed(() => (
  Number(refundForm.amount) > 0 &&
  Boolean(refundForm.requestNotes.trim()) &&
  Number(refundForm.amount) <= refundMaximumAmount.value + 0.000001
))
const refundReviewStatusLabel = computed(() => ({
  pending: '待审批',
  approved: '已批准',
  rejected: '已拒绝',
  cancelled: '已取消',
}[refundReview.value?.status || ''] || '未提交'))
const refundReviewStatusClass = computed(() => ({
  pending: 'status-amber',
  approved: 'status-green',
  rejected: 'status-coral',
  cancelled: 'status-gray',
}[refundReview.value?.status || ''] || 'status-gray'))

const formatMoney = (amount?: number | null, currency?: string | null): string => (
  amount && currency ? `${currency} ${Number(amount).toFixed(currency === 'JPY' || currency === 'KRW' || currency === 'CLP' ? 0 : 2)}` : '-'
)
const operatorFallback = (id?: string | number | null): string => Number(id || 0) > 0 ? `账号 #${id}` : '尚未处理'

const resetForm = (): void => {
  refundForm.amount = props.record?.refund_review?.proposed_amount
    ? String(props.record.refund_review.proposed_amount)
    : props.record?.refund_review_maximum_amount
      ? String(props.record.refund_review_maximum_amount)
      : ''
  refundForm.requestNotes = props.record?.refund_review?.request_notes || ''
  refundForm.decisionNotes = ''
  refundForm.draftConfirmed = false
}

const saveRefundReview = (): void => {
  if (!canSaveRefundReview.value) return
  emit(
    'save-refund-review',
    Number(refundForm.amount),
    props.record?.refund_review_currency || '',
    refundForm.requestNotes.trim(),
  )
}

const decideRefundReview = (status: 'approved' | 'rejected' | 'cancelled'): void => {
  if (!canDecide.value) return
  emit('decide-refund-review', status, refundForm.decisionNotes.trim())
}

const createPendingRefund = (): void => {
  if (!canCreatePendingRefund.value || !refundForm.draftConfirmed) return
  emit('create-pending-refund')
}

watch(
  () => [props.record?.id, props.record?.status, props.record?.refund_review?.status, props.record?.refund_review?.proposed_amount],
  () => resetForm(),
  { immediate: true },
)
</script>

<style scoped>
.field-label {
  display: block;
  color: hsl(var(--muted-foreground) / 0.7);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.currency-box {
  display: flex;
  height: 2.5rem;
  align-items: center;
  border: 1px dashed hsl(var(--border) / 0.8);
  border-radius: 0.375rem;
  padding: 0 0.75rem;
  font-size: 0.875rem;
  font-weight: 700;
}

.status-mark {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.25rem 0.55rem;
  font-size: 0.6875rem;
  font-weight: 800;
  white-space: nowrap;
}

.status-gray { background: rgb(148 163 184 / 0.12); color: rgb(71 85 105); }
.status-green { background: rgb(16 185 129 / 0.12); color: rgb(4 120 87); }
.status-amber { background: rgb(245 158 11 / 0.14); color: rgb(180 83 9); }
.status-coral { background: rgb(244 63 94 / 0.12); color: rgb(190 24 93); }
</style>
