<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="人工复核"
      description="人工判断风控策略风险，仅记录处理结论，不直接改变支付渠道状态。"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="loading" @click="fetchReviews">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid class="shrink-0" :items="statItems" />

    <AdminFilterPanel class="shrink-0">
      <form class="grid grid-cols-1 gap-3 md:grid-cols-[220px_1fr_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
          <select v-model="filters.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="">全部</option>
            <option value="pending">待复核</option>
            <option value="approved">已通过</option>
            <option value="rejected">已拒绝</option>
            <option value="cancelled">已取消</option>
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

    <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden max-xl:overflow-auto xl:grid-cols-[minmax(0,1fr)_minmax(460px,520px)] 2xl:grid-cols-[minmax(0,1fr)_520px]">
      <AdminTablePanel class="h-full min-h-0" :loading="loading" scroll-body>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>来源</TableHead>
              <TableHead>订单</TableHead>
              <TableHead>PaymentIntent</TableHead>
              <TableHead>原因</TableHead>
              <TableHead>创建时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="review in reviews"
              :key="review.id"
              class="cursor-pointer"
              :class="selectedReview?.id === review.id ? 'bg-primary/5' : ''"
              @click="selectReview(review)"
            >
              <TableCell class="font-mono text-xs">#{{ review.id }}</TableCell>
              <TableCell><RiskStrategyStatusPill :status="review.status" /></TableCell>
              <TableCell class="font-mono text-xs uppercase">{{ review.source || '-' }}</TableCell>
              <TableCell class="font-mono text-xs">{{ review.order_id ? `#${review.order_id}` : '-' }}</TableCell>
              <TableCell class="max-w-[180px] truncate font-mono text-xs">{{ review.payment_intent_id || '-' }}</TableCell>
              <TableCell class="max-w-[220px] truncate text-xs">{{ review.reason || '-' }}</TableCell>
              <TableCell class="text-xs text-muted-foreground">{{ formatDate(review.created_at) }}</TableCell>
            </TableRow>
            <TableRow v-if="!loading && reviews.length === 0">
              <TableCell colspan="7" class="h-24 text-center text-sm text-muted-foreground">暂无支付复核记录</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <template #footer>
          <RiskStrategyPaginationBar :pagination="pagination" @page="updatePage" @page-size="updatePageSize" />
        </template>
      </AdminTablePanel>

      <aside class="h-full min-h-[420px] min-w-0 overflow-y-auto overscroll-contain rounded-[24px] border border-dashed border-border/80 bg-card p-5">
        <div class="mb-4">
          <h2 class="text-sm font-black uppercase tracking-tight">复核处理</h2>
          <p class="mt-1 text-xs text-muted-foreground">仅记录人工判断，不直接改变 Stripe 或订单支付状态。</p>
        </div>

        <div v-if="selectedReview" class="space-y-3">
          <dl class="grid grid-cols-2 gap-2 text-xs">
            <div class="rounded-xl bg-muted/40 p-3">
              <dt class="font-black uppercase text-muted-foreground">Review</dt>
              <dd class="mt-1 font-mono">#{{ selectedReview.id }}</dd>
            </div>
            <div class="rounded-xl bg-muted/40 p-3">
              <dt class="font-black uppercase text-muted-foreground">Status</dt>
              <dd class="mt-1"><RiskStrategyStatusPill :status="selectedReview.status" /></dd>
            </div>
            <div class="col-span-2 rounded-xl bg-muted/40 p-3">
              <dt class="font-black uppercase text-muted-foreground">PaymentIntent</dt>
              <dd class="mt-1 break-all font-mono">{{ selectedReview.payment_intent_id || '-' }}</dd>
            </div>
            <div class="col-span-2 rounded-xl bg-muted/40 p-3">
              <dt class="font-black uppercase text-muted-foreground">Notes</dt>
              <dd class="mt-1 whitespace-pre-wrap">{{ selectedReview.notes || '-' }}</dd>
            </div>
          </dl>

          <label class="block space-y-1">
            <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">DECISION / 处理结果</span>
            <select v-model="decision.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm" :disabled="selectedReview.status !== 'pending'">
              <option value="pending">保持待复核</option>
              <option value="approved">通过</option>
              <option value="rejected">拒绝</option>
              <option value="cancelled">取消</option>
            </select>
          </label>
          <label class="block space-y-1">
            <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">OPERATOR NOTES / 操作备注</span>
            <Textarea v-model="decision.notes" rows="5" :disabled="selectedReview.status !== 'pending'" placeholder="记录证据、沟通结论或拒绝原因" />
          </label>
          <Button class="w-full rounded-full font-black uppercase tracking-wider" :disabled="selectedReview.status !== 'pending' || saving" @click="submitDecision">
            <CheckCircle2 class="size-4" />
            保存复核
          </Button>
        </div>
        <div v-else class="rounded-2xl border border-dashed border-border/80 p-6 text-center text-sm text-muted-foreground">
          选择左侧复核记录后处理。
        </div>

        <div class="mt-5 border-t border-dashed border-border/70 pt-4">
          <h3 class="mb-3 text-xs font-black uppercase tracking-widest text-muted-foreground">新建人工复核</h3>
          <form class="space-y-3" @submit.prevent="createManualReview">
            <Input v-model="manualReview.orderId" inputmode="numeric" placeholder="Order ID，可选" />
            <Input v-model="manualReview.paymentIntentId" placeholder="PaymentIntent ID，可选" />
            <Input v-model="manualReview.reason" required placeholder="复核原因" />
            <Textarea v-model="manualReview.notes" rows="3" placeholder="备注" />
            <Button type="submit" variant="outline" class="w-full rounded-full font-black uppercase tracking-wider" :disabled="saving">
              创建复核
            </Button>
          </form>
        </div>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { CheckCircle2, CreditCard, RefreshCw, Search, AlertTriangle } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import RiskStrategyPaginationBar from '@/components/admin/payment/RiskStrategyPaginationBar.vue'
import RiskStrategyStatusPill from '@/components/admin/payment/RiskStrategyStatusPill.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { paymentRiskApi as riskStrategyApi } from '@/api/paymentRisk'
import { applyPaged, formatDate, type RiskStrategyPagination } from '@/lib/riskStrategyViewUtils'

const loading = ref(false)
const saving = ref(false)
const reviews = ref<any[]>([])
const selectedReview = ref<any | null>(null)
const filters = reactive({ status: 'pending' })
const pagination = reactive<RiskStrategyPagination>({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const decision = reactive({ status: 'pending', notes: '' })
const manualReview = reactive({ orderId: '', paymentIntentId: '', reason: '', notes: '' })

const pendingReviews = computed(() => reviews.value.filter((item) => item.status === 'pending').length)
const approvedReviews = computed(() => reviews.value.filter((item) => item.status === 'approved').length)
const rejectedReviews = computed(() => reviews.value.filter((item) => ['rejected', 'cancelled'].includes(item.status)).length)
const statItems = computed(() => [
  { key: 'reviews', label: '复核记录', value: pagination.total || reviews.value.length, icon: CreditCard, tone: 'gray' },
  { key: 'pending_reviews', label: '待复核', value: pendingReviews.value, icon: AlertTriangle, tone: pendingReviews.value ? 'amber' : 'green' },
  { key: 'approved_reviews', label: '已通过', value: approvedReviews.value, icon: CheckCircle2, tone: approvedReviews.value ? 'green' : 'gray' },
  { key: 'rejected_reviews', label: '已拒绝/取消', value: rejectedReviews.value, icon: AlertTriangle, tone: rejectedReviews.value ? 'coral' : 'gray' },
])

const fetchReviews = async (): Promise<void> => {
  loading.value = true
  try {
    const payload = await riskStrategyApi.listReviews({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
    })
    applyPaged(reviews, pagination, payload)
    if (selectedReview.value) {
      selectedReview.value = reviews.value.find((item) => item.id === selectedReview.value.id) || null
    }
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void fetchReviews()
}

const resetFilters = (): void => {
  filters.status = 'pending'
  pagination.page = 1
  void fetchReviews()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void fetchReviews()
}

const updatePageSize = (pageSize: number): void => {
  pagination.page_size = pageSize
  pagination.page = 1
  void fetchReviews()
}

const selectReview = (review: any): void => {
  selectedReview.value = review
  decision.status = review.status || 'pending'
  decision.notes = review.notes || ''
}

const submitDecision = async (): Promise<void> => {
  if (!selectedReview.value) return
  saving.value = true
  try {
    await riskStrategyApi.updateReview(selectedReview.value.id, {
      status: decision.status,
      notes: decision.notes,
    })
    toast.success('复核已保存')
    await fetchReviews()
  } finally {
    saving.value = false
  }
}

const createManualReview = async (): Promise<void> => {
  saving.value = true
  try {
    await riskStrategyApi.createReview({
      order_id: manualReview.orderId ? Number(manualReview.orderId) : undefined,
      payment_intent_id: manualReview.paymentIntentId.trim() || undefined,
      reason: manualReview.reason.trim(),
      notes: manualReview.notes.trim(),
    })
    Object.assign(manualReview, { orderId: '', paymentIntentId: '', reason: '', notes: '' })
    toast.success('人工复核已创建')
    pagination.page = 1
    await fetchReviews()
  } finally {
    saving.value = false
  }
}

onMounted(fetchReviews)
</script>

