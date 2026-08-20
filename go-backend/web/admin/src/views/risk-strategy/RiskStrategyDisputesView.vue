<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="拒付处理"
      description="管理 Stripe / PayPal 拒付、证据预览、提交期限和商业发票。"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" @click="paypalInvoicePreviewOpen = true">
          <FileText class="size-3.5" />
          发票样式
        </Button>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="loading" @click="fetchDisputes">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <PayPalInvoicePreviewDialog v-model:open="paypalInvoicePreviewOpen" />
    <AdminStatsGrid class="shrink-0" :items="statItems" />

    <AdminFilterPanel class="shrink-0">
      <form class="grid grid-cols-1 gap-3 md:grid-cols-[180px_220px_1fr_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">PROVIDER / 渠道</span>
          <select v-model="disputeProvider" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="stripe">Stripe</option>
            <option value="paypal">PayPal</option>
          </select>
        </label>
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
          <select v-model="filters.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="">全部</option>
            <option v-for="status in disputeStatusOptions" :key="status" :value="status">{{ status }}</option>
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

    <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden xl:grid-cols-[minmax(0,1fr)_460px]">
      <AdminTablePanel :loading="loading" class="min-h-0" scroll-body>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>{{ disputeProviderLabel }} Dispute</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>金额</TableHead>
              <TableHead>原因</TableHead>
              <TableHead>订单</TableHead>
              <TableHead>{{ disputeProvider === 'paypal' ? '证据提交' : '证据截止' }}</TableHead>
              <TableHead>更新时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="dispute in disputes"
              :key="`${disputeProvider}-${dispute.id}`"
              class="cursor-pointer"
              :class="selectedDispute?.id === dispute.id ? 'bg-primary/5' : ''"
              @click="selectDispute(dispute)"
            >
              <TableCell class="font-mono text-xs">#{{ dispute.id }}</TableCell>
              <TableCell class="max-w-[220px] truncate font-mono text-xs">{{ disputeReference(dispute) }}</TableCell>
              <TableCell><RiskStrategyStatusPill :status="dispute.status" /></TableCell>
              <TableCell class="font-mono text-xs">{{ formatMoney(dispute.amount, dispute.currency) }}</TableCell>
              <TableCell class="text-xs">{{ dispute.reason || '-' }}</TableCell>
              <TableCell class="font-mono text-xs">{{ dispute.order_id ? `#${dispute.order_id}` : '-' }}</TableCell>
              <TableCell class="text-xs" :class="isEvidenceSoon(dispute.evidence_due_at) ? 'font-semibold text-rose-600' : 'text-muted-foreground'">
                {{ formatDate(disputeProvider === 'paypal' ? dispute.evidence_submitted_at : dispute.evidence_due_at) }}
              </TableCell>
              <TableCell class="text-xs text-muted-foreground">{{ formatDate(dispute.updated_at) }}</TableCell>
            </TableRow>
            <TableRow v-if="!loading && disputes.length === 0">
              <TableCell colspan="8" class="h-24 text-center text-sm text-muted-foreground">暂无 {{ disputeProviderLabel }} 拒付记录</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <template #footer>
          <RiskStrategyPaginationBar :pagination="pagination" @page="updatePage" @page-size="updatePageSize" />
        </template>
      </AdminTablePanel>

      <aside class="min-h-0 overflow-auto rounded-[24px] border border-dashed border-border/80 bg-card p-4">
        <div class="mb-4 flex items-start justify-between gap-3">
          <div>
            <h2 class="text-sm font-black uppercase tracking-tight">拒付证据工作台</h2>
            <p class="mt-1 text-xs text-muted-foreground">{{ disputeWorkbenchSubtitle }}</p>
          </div>
          <div v-if="selectedDispute" class="flex items-center gap-2">
            <Button v-if="disputeProvider === 'paypal'" variant="outline" size="sm" class="rounded-full text-xs font-black" @click="openPayPalInvoicePDF">
              <FileText class="size-3.5" />
              PDF
            </Button>
            <Button variant="outline" size="sm" class="rounded-full text-xs font-black" :disabled="evidenceLoading" @click="fetchDisputeEvidence">
              <RefreshCw :class="['size-3.5', { 'animate-spin': evidenceLoading }]" />
            </Button>
          </div>
        </div>

        <div v-if="selectedDispute" class="space-y-4">
          <dl class="grid grid-cols-2 gap-2 text-xs">
            <div class="rounded-xl bg-muted/40 p-3">
              <dt class="font-black uppercase text-muted-foreground">Dispute</dt>
              <dd class="mt-1 break-all font-mono">{{ disputeReference(selectedDispute) }}</dd>
            </div>
            <div class="rounded-xl bg-muted/40 p-3">
              <dt class="font-black uppercase text-muted-foreground">Status</dt>
              <dd class="mt-1"><RiskStrategyStatusPill :status="selectedDispute.status" /></dd>
            </div>
            <div class="rounded-xl bg-muted/40 p-3">
              <dt class="font-black uppercase text-muted-foreground">Evidence Due</dt>
              <dd class="mt-1">{{ formatDate(selectedDispute.evidence_due_at) }}</dd>
            </div>
            <div class="rounded-xl bg-muted/40 p-3">
              <dt class="font-black uppercase text-muted-foreground">Submitted</dt>
              <dd class="mt-1">{{ formatDate(selectedDispute.evidence_submitted_at) }}</dd>
            </div>
          </dl>

          <div v-if="evidenceLoading" class="rounded-2xl border border-dashed border-border/80 p-6 text-center text-sm text-muted-foreground">
            正在生成证据包...
          </div>

          <div v-else-if="disputeEvidence" class="space-y-4">
            <div v-if="disputeEvidence.warnings?.length" class="space-y-2">
              <div v-for="warning in disputeEvidence.warnings" :key="warning" class="rounded-xl border border-amber-500/20 bg-amber-500/10 p-3 text-xs text-amber-800">
                {{ warning }}
              </div>
            </div>

            <section class="rounded-2xl border border-dashed border-border/80 p-3">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h3 class="text-xs font-black uppercase tracking-widest">7 项拒付证据链</h3>
                  <p class="mt-1 text-[11px] text-muted-foreground">只显示系统实际找到的证据；“尚未接入”不会被当成已具备。</p>
                </div>
                <span class="shrink-0 rounded-full bg-muted px-2.5 py-1 text-[11px] font-black">
                  {{ disputeEvidence.evidence_checklist?.ready_count || 0 }}/{{ disputeEvidence.evidence_checklist?.total_count || 7 }} 已核验
                </span>
              </div>
              <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-[10px] text-muted-foreground">
                <span>需人工：{{ disputeEvidence.evidence_checklist?.manual_required_count || 0 }}</span>
                <span>缺失：{{ disputeEvidence.evidence_checklist?.missing_count || 0 }}</span>
                <span>尚未接入：{{ disputeEvidence.evidence_checklist?.unavailable_count || 0 }}</span>
                <span v-if="disputeEvidence.evidence_checklist?.complete" class="font-black text-emerald-700">7 项完整</span>
                <span v-else class="font-black text-amber-700">证据链未完整</span>
              </div>

              <div class="mt-3 space-y-2">
                <div
                  v-for="item in (disputeEvidence.evidence_checklist?.items || [])"
                  :key="item.key"
                  class="rounded-xl border p-3"
                  :class="evidenceStatusClass(item.status)"
                >
                  <div class="flex items-start justify-between gap-2">
                    <div class="min-w-0">
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="text-xs font-black">{{ item.title }}</span>
                        <span v-if="item.required" class="text-[10px] font-black uppercase tracking-widest opacity-70">必需</span>
                      </div>
                      <div class="mt-1 text-[10px] font-mono opacity-70">{{ item.provider_field || '-' }}</div>
                    </div>
                    <span class="shrink-0 rounded-full px-2 py-1 text-[10px] font-black" :class="evidenceStatusPillClass(item.status)">
                      {{ evidenceStatusLabel(item.status) }}
                    </span>
                  </div>
                  <p class="mt-2 text-xs leading-5">{{ item.summary || '-' }}</p>
                  <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-[10px] opacity-75">
                    <span>来源：{{ item.source || '-' }}</span>
                    <span>时间：{{ formatDate(item.observed_at) }}</span>
                  </div>
                  <p v-if="item.missing_reason" class="mt-2 text-[11px] font-semibold leading-5 opacity-80">
                    缺失说明：{{ item.missing_reason }}
                  </p>
                </div>
              </div>

              <div
                v-if="disputeEvidence.submission_check"
                class="mt-3 rounded-xl border p-3 text-xs"
                :class="disputeEvidence.submission_check.ready ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-800' : 'border-rose-500/30 bg-rose-500/10 text-rose-800'"
              >
                <div class="font-black">
                  {{ disputeEvidence.submission_check.ready ? '渠道硬性提交条件已满足' : '当前不能提交：存在渠道硬性阻断' }}
                </div>
                <p v-if="disputeEvidence.submission_check.ready" class="mt-1 leading-5">
                  这不代表发卡行或支付渠道必然支持申诉结果；仍需确认下方人工补充项和证据文本。
                </p>
                <div v-else class="mt-1 space-y-1">
                  <p v-for="blocker in disputeEvidence.submission_check.blockers || []" :key="blocker">• {{ blocker }}</p>
                </div>
                <p v-if="disputeEvidence.submission_check.warnings?.length" class="mt-2 leading-5 opacity-80">
                  尚未完整的非阻断项：{{ disputeEvidence.submission_check.warnings.length }} 项
                </p>
              </div>
            </section>

            <section class="rounded-2xl border border-dashed border-border/80 p-3">
              <h3 class="mb-2 text-xs font-black uppercase tracking-widest text-muted-foreground">{{ disputeEvidencePreviewTitle }}</h3>
              <dl class="space-y-2 text-xs">
                <div v-for="field in evidenceFields" :key="field.key" class="grid gap-1">
                  <dt class="font-black uppercase text-muted-foreground">{{ field.label }}</dt>
                  <dd class="whitespace-pre-wrap break-words rounded-lg bg-muted/35 p-2 font-mono">{{ field.value || '-' }}</dd>
                </div>
              </dl>
            </section>

            <section class="rounded-2xl border border-dashed border-border/80 p-3">
              <h3 class="mb-2 text-xs font-black uppercase tracking-widest text-muted-foreground">Tracking Events</h3>
              <div v-if="disputeEvidence.tracking_events?.length" class="max-h-44 space-y-2 overflow-auto pr-1">
                <div v-for="event in disputeEvidence.tracking_events.slice(0, 12)" :key="event.id || `${event.event_time}-${event.status}`" class="rounded-lg bg-muted/35 p-2 text-xs">
                  <div class="font-mono text-muted-foreground">{{ formatDate(event.event_time) }}</div>
                  <div class="font-semibold">{{ event.status || '-' }}</div>
                  <div class="text-muted-foreground">{{ event.location || '-' }}</div>
                  <p class="mt-1 whitespace-pre-wrap">{{ event.description || '-' }}</p>
                </div>
              </div>
              <p v-else class="text-xs text-muted-foreground">暂无物流事件。</p>
            </section>

            <section class="rounded-2xl border border-dashed border-border/80 p-3">
              <h3 class="mb-2 text-xs font-black uppercase tracking-widest text-muted-foreground">Customer Communication</h3>
              <label v-if="disputeProvider === 'stripe'" class="mb-2 flex items-center gap-2 text-xs font-bold text-muted-foreground">
                <input v-model="evidenceForm.include_customer_communication" type="checkbox" class="size-4 accent-primary" />
                将沟通摘要放入 Uncategorized Text
              </label>
              <div v-if="disputeEvidence.communications?.length" class="max-h-44 space-y-2 overflow-auto pr-1">
                <div v-for="message in disputeEvidence.communications.slice(0, 16)" :key="message.id" class="rounded-lg bg-muted/35 p-2 text-xs">
                  <div class="flex items-center justify-between gap-2">
                    <span class="font-semibold">{{ message.sender }}</span>
                    <span class="font-mono text-muted-foreground">{{ formatDate(message.created_at) }}</span>
                  </div>
                  <p class="mt-1 whitespace-pre-wrap">{{ message.content }}</p>
                </div>
              </div>
              <p v-else class="text-xs text-muted-foreground">没有找到可关联的客服沟通。</p>
            </section>

            <section v-if="disputeProvider === 'stripe'" class="space-y-3 rounded-2xl border border-dashed border-border/80 p-3">
              <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">Stripe File IDs</h3>
              <p class="text-[11px] leading-5 text-muted-foreground">这些是 Stripe 外部文件引用，不是系统自动生成的凭证。上传后把 File ID 填入对应项。</p>
              <label class="grid gap-1">
                <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">履约 / POD</span>
                <Input v-model="evidenceForm.shipping_documentation_file_id" placeholder="Shipping documentation File ID (file_...)" />
              </label>
              <label class="grid gap-1">
                <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">客服沟通</span>
                <Input v-model="evidenceForm.customer_communication_file_id" placeholder="Customer communication File ID (file_...)" />
              </label>
              <label class="grid gap-1">
                <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">采购收据</span>
                <Input v-model="evidenceForm.receipt_file_id" placeholder="Receipt File ID (file_...)" />
              </label>
              <label class="grid gap-1">
                <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">其他说明文件</span>
                <Input v-model="evidenceForm.uncategorized_file_id" placeholder="Uncategorized File ID (file_...)" />
              </label>
              <Textarea v-model="evidenceForm.additional_statement" rows="4" placeholder="人工补充说明，可选" />
              <label class="flex items-start gap-2 text-xs font-bold text-muted-foreground">
                <input v-model="evidenceForm.confirm" type="checkbox" class="mt-0.5 size-4 accent-primary" />
                我已检查 7 项证据链、缺失说明、证据文本和附件，确认提交给 Stripe。
              </label>
              <Button class="w-full rounded-full font-black uppercase tracking-wider" :disabled="!disputeEvidence.submission_check?.ready || !evidenceForm.confirm || evidenceSubmitting" @click="submitEvidence">
                <Send class="size-4" />
                提交证据到 Stripe
              </Button>
            </section>

            <section v-else class="space-y-3 rounded-2xl border border-dashed border-border/80 p-3">
              <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">PayPal 手动提交</h3>
              <p class="text-[11px] leading-5 text-muted-foreground">系统会自动生成商业发票 PDF；提交前仍需核对物流妥投、沟通记录和 7 项证据链状态。</p>
              <Textarea v-model="evidenceForm.additional_statement" rows="4" placeholder="人工补充说明，可选" />
              <label class="flex items-start gap-2 text-xs font-bold text-muted-foreground">
                <input v-model="evidenceForm.confirm" type="checkbox" class="mt-0.5 size-4 accent-primary" />
                我已检查 7 项证据链和 PayPal 证据说明，确认提交。
              </label>
              <Button class="w-full rounded-full font-black uppercase tracking-wider" :disabled="!disputeEvidence.submission_check?.ready || !evidenceForm.confirm || evidenceSubmitting" @click="submitEvidence">
                <Send class="size-4" />
                手动提交证据到 PayPal
              </Button>
              <Button variant="outline" class="w-full rounded-full font-black uppercase tracking-wider" @click="openPayPalInvoicePDF">
                <FileText class="size-4" />
                打开 PDF 预览
              </Button>
            </section>
          </div>
        </div>
        <div v-else class="rounded-2xl border border-dashed border-border/80 p-6 text-center text-sm text-muted-foreground">
          选择左侧拒付记录后查看证据包。
        </div>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { AlertTriangle, CheckCircle2, CreditCard, FileText, RefreshCw, Search, Send, ShieldAlert } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import PayPalInvoicePreviewDialog from '@/components/admin/payment/PayPalInvoicePreviewDialog.vue'
import RiskStrategyPaginationBar from '@/components/admin/payment/RiskStrategyPaginationBar.vue'
import RiskStrategyStatusPill from '@/components/admin/payment/RiskStrategyStatusPill.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { paymentRiskApi as riskStrategyApi } from '@/api/paymentRisk'
import { getPaymentChannelLabel } from '@/lib/paymentChannels'
import { applyPaged, formatDate, formatMoney, isEvidenceSoon, type RiskStrategyPagination } from '@/lib/riskStrategyViewUtils'

const props = withDefaults(defineProps<{
  defaultDisputeProvider?: string
}>(), {
  defaultDisputeProvider: '',
})

const normalizeDisputeProvider = (provider?: string): 'stripe' | 'paypal' => (
  String(provider || '').trim().toLowerCase() === 'paypal' ? 'paypal' : 'stripe'
)

const loading = ref(false)
const evidenceLoading = ref(false)
const evidenceSubmitting = ref(false)
const paypalInvoicePreviewOpen = ref(false)
const disputes = ref<any[]>([])
const selectedDispute = ref<any | null>(null)
const disputeEvidence = ref<any | null>(null)
const disputeProvider = ref<'stripe' | 'paypal'>(normalizeDisputeProvider(props.defaultDisputeProvider))
const filters = reactive({ status: '' })
const pagination = reactive<RiskStrategyPagination>({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const evidenceForm = reactive({
  include_customer_communication: false,
  shipping_documentation_file_id: '',
  customer_communication_file_id: '',
  receipt_file_id: '',
  uncategorized_file_id: '',
  additional_statement: '',
  confirm: false,
})

const disputeProviderLabel = computed(() => getPaymentChannelLabel(disputeProvider.value, 'Stripe'))
const disputeWorkbenchSubtitle = computed(() => (
  disputeProvider.value === 'paypal'
    ? '预览 PayPal 证据包和商业发票 PDF。'
    : '先预览证据，再人工确认提交到 Stripe。'
))
const disputeEvidencePreviewTitle = computed(() => disputeProvider.value === 'paypal' ? 'PayPal Evidence Preview' : 'Stripe Evidence Preview')
const disputeStatusOptions = computed(() => (
  disputeProvider.value === 'paypal'
    ? ['WAITING_FOR_SELLER_RESPONSE', 'OPEN', 'UNDER_REVIEW', 'WAITING_FOR_BUYER_RESPONSE', 'RESOLVED']
    : ['needs_response', 'warning_needs_response', 'under_review', 'won', 'lost', 'closed']
))
const urgentDisputes = computed(() => {
  const urgentStatuses = disputeProvider.value === 'paypal'
    ? ['WAITING_FOR_SELLER_RESPONSE', 'OPEN']
    : ['needs_response', 'warning_needs_response']
  return disputes.value.filter((item) => urgentStatuses.includes(item.status)).length
})
const resolvedDisputes = computed(() => {
  const resolvedStatuses = disputeProvider.value === 'paypal'
    ? ['RESOLVED']
    : ['won', 'lost', 'closed']
  return disputes.value.filter((item) => resolvedStatuses.includes(item.status)).length
})
const statItems = computed(() => [
  { key: 'disputes', label: `${disputeProviderLabel.value} 拒付`, value: pagination.total || disputes.value.length, icon: ShieldAlert, tone: 'gray' },
  { key: 'urgent_disputes', label: '需立即响应', value: urgentDisputes.value, icon: AlertTriangle, tone: urgentDisputes.value ? 'coral' : 'green' },
  { key: 'resolved_disputes', label: '已结束', value: resolvedDisputes.value, icon: CheckCircle2, tone: resolvedDisputes.value ? 'green' : 'gray' },
  { key: 'provider', label: '当前渠道', value: disputeProviderLabel.value, icon: CreditCard, tone: 'gray' },
])
const evidenceFields = computed(() => {
  const evidence = disputeEvidence.value?.evidence || {}
  if (disputeProvider.value === 'paypal') {
    return [
      { key: 'customer_name', label: 'Customer Name', value: evidence.customer_name },
      { key: 'customer_email_address', label: 'Customer Email', value: evidence.customer_email_address },
      { key: 'shipping_address', label: 'Shipping Address', value: evidence.shipping_address },
      { key: 'product_description', label: 'Product Description', value: evidence.product_description },
      { key: 'shipping_carrier', label: 'Shipping Carrier', value: evidence.shipping_carrier },
      { key: 'shipping_date', label: 'Shipping Date', value: evidence.shipping_date },
      { key: 'shipping_tracking_number', label: 'Tracking Number', value: evidence.shipping_tracking_number },
      { key: 'delivered_at', label: 'Delivered At', value: evidence.delivered_at },
      { key: 'invoice_summary', label: 'Invoice Summary', value: evidence.invoice_summary },
      { key: 'proof_of_delivery_summary', label: 'Proof Of Delivery', value: evidence.proof_of_delivery_summary },
      { key: 'communication_summary', label: 'Communication Summary', value: evidence.communication_summary },
      { key: 'notes', label: 'Notes', value: evidence.notes },
    ]
  }
  return [
    { key: 'customer_name', label: 'Customer Name', value: evidence.customer_name },
    { key: 'customer_email_address', label: 'Customer Email', value: evidence.customer_email_address },
    { key: 'billing_address', label: 'Billing Address', value: evidence.billing_address },
    { key: 'shipping_address', label: 'Shipping Address', value: evidence.shipping_address },
    { key: 'product_description', label: 'Product Description', value: evidence.product_description },
    { key: 'shipping_carrier', label: 'Shipping Carrier', value: evidence.shipping_carrier },
    { key: 'shipping_date', label: 'Shipping Date', value: evidence.shipping_date },
    { key: 'shipping_tracking_number', label: 'Tracking Number', value: evidence.shipping_tracking_number },
    { key: 'uncategorized_text', label: 'Uncategorized Text', value: evidence.uncategorized_text },
  ]
})

const evidenceStatusLabel = (status?: string): string => ({
  ready: '已核验',
  missing: '缺失',
  manual_required: '需人工补充',
  unavailable: '尚未接入',
}[status || ''] || status || '未知')

const evidenceStatusClass = (status?: string): string => ({
  ready: 'border-emerald-500/25 bg-emerald-500/5 text-emerald-900',
  missing: 'border-rose-500/25 bg-rose-500/5 text-rose-900',
  manual_required: 'border-amber-500/25 bg-amber-500/5 text-amber-900',
  unavailable: 'border-slate-500/25 bg-slate-500/5 text-slate-800',
}[status || ''] || 'border-border bg-muted/20 text-foreground')

const evidenceStatusPillClass = (status?: string): string => ({
  ready: 'bg-emerald-600/15 text-emerald-700',
  missing: 'bg-rose-600/15 text-rose-700',
  manual_required: 'bg-amber-600/15 text-amber-700',
  unavailable: 'bg-slate-600/15 text-slate-700',
}[status || ''] || 'bg-muted text-muted-foreground')

const disputeReference = (dispute: any): string => {
  if (!dispute) return '-'
  return disputeProvider.value === 'paypal'
    ? dispute.paypal_dispute_id || '-'
    : dispute.stripe_dispute_id || '-'
}

const fetchDisputes = async (): Promise<void> => {
  loading.value = true
  try {
    const apiCall = disputeProvider.value === 'paypal' ? riskStrategyApi.listPayPalDisputes : riskStrategyApi.listDisputes
    const payload = await apiCall({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
    })
    applyPaged(disputes, pagination, payload)
    if (selectedDispute.value) {
      selectedDispute.value = disputes.value.find((item) => item.id === selectedDispute.value.id) || null
      if (!selectedDispute.value) disputeEvidence.value = null
    }
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void fetchDisputes()
}

const resetFilters = (): void => {
  filters.status = ''
  pagination.page = 1
  void fetchDisputes()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void fetchDisputes()
}

const updatePageSize = (pageSize: number): void => {
  pagination.page_size = pageSize
  pagination.page = 1
  void fetchDisputes()
}

const resetEvidenceForm = (): void => {
  Object.assign(evidenceForm, {
    include_customer_communication: false,
    shipping_documentation_file_id: '',
    customer_communication_file_id: '',
    receipt_file_id: '',
    uncategorized_file_id: '',
    additional_statement: '',
    confirm: false,
  })
}

const fetchDisputeEvidence = async (): Promise<void> => {
  if (!selectedDispute.value) return
  evidenceLoading.value = true
  try {
    disputeEvidence.value = disputeProvider.value === 'paypal'
      ? await riskStrategyApi.getPayPalDisputeEvidence(selectedDispute.value.id)
      : await riskStrategyApi.getDisputeEvidence(selectedDispute.value.id)
  } finally {
    evidenceLoading.value = false
  }
}

const selectDispute = async (dispute: any): Promise<void> => {
  selectedDispute.value = dispute
  disputeEvidence.value = null
  resetEvidenceForm()
  await fetchDisputeEvidence()
}

const openPayPalInvoicePDF = (): void => {
  if (!selectedDispute.value) return
  window.open(riskStrategyApi.paypalDisputeInvoicePDFUrl(selectedDispute.value.id), '_blank', 'noopener,noreferrer')
}

const submitEvidence = async (): Promise<void> => {
  if (!selectedDispute.value) return
  evidenceSubmitting.value = true
  try {
    if (disputeProvider.value === 'paypal') {
      await riskStrategyApi.submitPayPalDisputeEvidence(selectedDispute.value.id, {
        confirm: evidenceForm.confirm,
        additional_statement: evidenceForm.additional_statement.trim(),
      })
      toast.success('PayPal 拒付证据已提交')
    } else {
      await riskStrategyApi.submitDisputeEvidence(selectedDispute.value.id, {
        confirm: evidenceForm.confirm,
        submit: true,
        include_customer_communication: evidenceForm.include_customer_communication,
        shipping_documentation_file_id: evidenceForm.shipping_documentation_file_id.trim(),
        customer_communication_file_id: evidenceForm.customer_communication_file_id.trim(),
        receipt_file_id: evidenceForm.receipt_file_id.trim(),
        uncategorized_file_id: evidenceForm.uncategorized_file_id.trim(),
        additional_statement: evidenceForm.additional_statement.trim(),
      })
      toast.success('Stripe 拒付证据已提交')
    }
    evidenceForm.confirm = false
    await fetchDisputes()
    await fetchDisputeEvidence()
  } finally {
    evidenceSubmitting.value = false
  }
}

watch(() => props.defaultDisputeProvider, (provider) => {
  const nextProvider = normalizeDisputeProvider(provider)
  if (disputeProvider.value !== nextProvider) disputeProvider.value = nextProvider
}, { immediate: true })

watch(disputeProvider, () => {
  filters.status = ''
  pagination.page = 1
  selectedDispute.value = null
  disputeEvidence.value = null
  resetEvidenceForm()
  void fetchDisputes()
})

onMounted(fetchDisputes)
</script>

