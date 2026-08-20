<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="退款建议"
      description="处理风险事件生成的退款建议，并在确认后创建或执行退款。"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="loading" @click="fetchRecommendations">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid class="shrink-0" :items="statItems" />

    <AdminFilterPanel class="shrink-0">
      <form class="grid grid-cols-1 gap-3 md:grid-cols-[180px_180px_1fr_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
          <select v-model="filters.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="">全部</option>
            <option value="pending">待处理</option>
            <option value="accepted">已采纳</option>
            <option value="dismissed">已驳回</option>
            <option value="cancelled">已取消</option>
          </select>
        </label>
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">PROVIDER / 渠道</span>
          <select v-model="filters.provider" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="">全部</option>
            <option value="stripe">Stripe</option>
            <option value="paypal">PayPal</option>
          </select>
        </label>
        <div class="hidden md:block" />
        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" :disabled="loading">
            <Search class="size-3.5" />
            查询
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" @click="resetFilters">
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden xl:grid-cols-[minmax(0,1fr)_400px]">
      <AdminTablePanel :loading="loading" scroll-body>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>渠道</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>来源</TableHead>
              <TableHead>建议金额</TableHead>
              <TableHead>订单</TableHead>
              <TableHead>支付引用</TableHead>
              <TableHead>处理期限</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="item in recommendations"
              :key="item.id"
              class="cursor-pointer"
              :class="selectedRecommendation?.id === item.id ? 'bg-primary/5' : ''"
              @click="selectRecommendation(item)"
            >
              <TableCell class="font-mono text-xs">#{{ item.id }}</TableCell>
              <TableCell class="font-mono text-xs uppercase">{{ item.provider || '-' }}</TableCell>
              <TableCell><RiskStrategyStatusPill :status="item.status" /></TableCell>
              <TableCell class="text-xs">{{ refundRecommendationSourceLabel(item.source_kind) }}</TableCell>
              <TableCell class="font-mono text-xs">{{ formatMoney(item.recommended_amount, item.currency) }}</TableCell>
              <TableCell class="font-mono text-xs">{{ item.order_id ? `#${item.order_id}` : '-' }}</TableCell>
              <TableCell class="max-w-[180px] truncate font-mono text-xs">{{ item.provider_payment_id || item.payment_intent_id || item.charge_id || '-' }}</TableCell>
              <TableCell class="text-xs" :class="isEvidenceSoon(item.review_by) ? 'font-semibold text-rose-600' : 'text-muted-foreground'">{{ formatDate(item.review_by) }}</TableCell>
            </TableRow>
            <TableRow v-if="!loading && recommendations.length === 0">
              <TableCell colspan="8" class="h-24 text-center text-sm text-muted-foreground">暂无退款建议</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <template #footer>
          <RiskStrategyPaginationBar :pagination="pagination" @page="updatePage" @page-size="updatePageSize" />
        </template>
      </AdminTablePanel>

      <RiskStrategyRefundRecommendationDetailPanel
        v-model:decision-status="decision.status"
        v-model:decision-notes="decision.decision_notes"
        :recommendation="selectedRecommendation"
        :saving="saving"
        :draft-saving="draftSaving"
        :execution-saving="executionSaving"
        :execution-completed="selectedExecutionCompleted"
        @save-decision="submitDecision"
        @create-draft="createPendingRefund"
        @execute-refund="executeProviderRefund"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { AlertTriangle, CheckCircle2, CreditCard, RefreshCw, Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import RiskStrategyRefundRecommendationDetailPanel from '@/components/admin/payment/RiskStrategyRefundRecommendationDetailPanel.vue'
import RiskStrategyPaginationBar from '@/components/admin/payment/RiskStrategyPaginationBar.vue'
import RiskStrategyStatusPill from '@/components/admin/payment/RiskStrategyStatusPill.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { paymentRefundApi } from '@/api/paymentRefunds'
import { paymentRiskApi as riskStrategyApi } from '@/api/paymentRisk'
import {
  applyPaged,
  formatDate,
  formatMoney,
  isEvidenceSoon,
  refundRecommendationSourceLabel,
  type RiskStrategyPagination,
} from '@/lib/riskStrategyViewUtils'

const loading = ref(false)
const saving = ref(false)
const draftSaving = ref(false)
const executionSaving = ref(false)
const recommendations = ref<any[]>([])
const selectedRecommendation = ref<any | null>(null)
const executedRefundIds = ref<Set<string | number>>(new Set())
const filters = reactive({ status: 'pending', provider: '' })
const pagination = reactive<RiskStrategyPagination>({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const decision = reactive({ status: 'pending', decision_notes: '' })

const pendingRecommendations = computed(() => recommendations.value.filter((item) => item.status === 'pending').length)
const acceptedRecommendations = computed(() => recommendations.value.filter((item) => item.status === 'accepted').length)
const recommendationsDueSoon = computed(() => recommendations.value.filter((item) => item.status === 'pending' && isEvidenceSoon(item.review_by)).length)
const statItems = computed(() => [
  { key: 'refunds', label: '退款建议', value: pagination.total || recommendations.value.length, icon: CreditCard, tone: 'gray' },
  { key: 'pending_refunds', label: '待处理', value: pendingRecommendations.value, icon: AlertTriangle, tone: pendingRecommendations.value ? 'amber' : 'green' },
  { key: 'accepted_refunds', label: '已采纳', value: acceptedRecommendations.value, icon: CheckCircle2, tone: acceptedRecommendations.value ? 'green' : 'gray' },
  { key: 'due_soon', label: '即将到期', value: recommendationsDueSoon.value, icon: AlertTriangle, tone: recommendationsDueSoon.value ? 'coral' : 'green' },
])
const selectedExecutionCompleted = computed(() => {
  const refundID = selectedRecommendation.value?.linked_refund_id
  return Boolean(refundID && executedRefundIds.value.has(refundID))
})

const fetchRecommendations = async (): Promise<void> => {
  loading.value = true
  try {
    const payload = await riskStrategyApi.listRefundRecommendations({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      provider: filters.provider || undefined,
    })
    applyPaged(recommendations, pagination, payload)
    if (selectedRecommendation.value) {
      selectedRecommendation.value = recommendations.value.find((item) => item.id === selectedRecommendation.value.id) || null
    }
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void fetchRecommendations()
}

const resetFilters = (): void => {
  filters.status = 'pending'
  filters.provider = ''
  pagination.page = 1
  void fetchRecommendations()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void fetchRecommendations()
}

const updatePageSize = (pageSize: number): void => {
  pagination.page_size = pageSize
  pagination.page = 1
  void fetchRecommendations()
}

const selectRecommendation = (recommendation: any): void => {
  selectedRecommendation.value = recommendation
  decision.status = recommendation.status || 'pending'
  decision.decision_notes = recommendation.decision_notes || ''
}

const submitDecision = async (): Promise<void> => {
  if (!selectedRecommendation.value) return
  saving.value = true
  try {
    await riskStrategyApi.updateRefundRecommendation(selectedRecommendation.value.id, {
      status: decision.status,
      decision_notes: decision.decision_notes,
    })
    toast.success('退款建议处理已保存')
    await fetchRecommendations()
  } finally {
    saving.value = false
  }
}

const createPendingRefund = async (payload: any): Promise<void> => {
  if (!selectedRecommendation.value) return
  draftSaving.value = true
  try {
    const result = await riskStrategyApi.createPendingRefundFromRecommendation(selectedRecommendation.value.id, payload)
    const refundID = result?.refund?.id
    selectedRecommendation.value = result?.recommendation || selectedRecommendation.value
    toast.success(refundID ? `已生成待处理退款 #${refundID}` : '已生成待处理退款')
    await fetchRecommendations()
  } finally {
    draftSaving.value = false
  }
}

const executeProviderRefund = async (payload: any): Promise<void> => {
  const refundID = payload?.refund_id || selectedRecommendation.value?.linked_refund_id
  if (!refundID) return
  executionSaving.value = true
  try {
    const result = await paymentRefundApi.executePendingRefund(refundID, { confirm: Boolean(payload?.confirm) })
    const providerRefundID = result?.execution?.provider_refund_id || result?.refund?.refund_id
    const next = new Set(executedRefundIds.value)
    next.add(refundID)
    executedRefundIds.value = next
    toast.success(providerRefundID ? `支付渠道退款已执行：${providerRefundID}` : '支付渠道退款已执行')
  } finally {
    executionSaving.value = false
  }
}

onMounted(fetchRecommendations)
</script>

