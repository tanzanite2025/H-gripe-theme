<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="支付风控"
      description="Stripe 拒付监控、Radar 复核与人工支付复核队列"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="currentLoading" @click="refreshCurrent">
          <RefreshCw :class="['size-3.5', { 'animate-spin': currentLoading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid class="shrink-0" :items="statItems" />
    <PaymentRiskSummaryPanel
      class="shrink-0"
      :reports="riskSummary.reports"
      :enabled="riskSummary.enabled"
      :loading="summaryLoading"
    />

    <Tabs :model-value="activeTab" class="min-h-0 flex-1 overflow-hidden">
      <TabsContent value="reviews" class="min-h-0 flex flex-col gap-3 overflow-hidden">
        <AdminFilterPanel class="shrink-0">
          <form class="grid grid-cols-1 gap-3 md:grid-cols-[220px_1fr_auto]" @submit.prevent="applyReviewFilters">
            <label class="block space-y-1">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
              <select v-model="reviewFilters.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
                <option value="">全部</option>
                <option value="pending">待复核</option>
                <option value="approved">已通过</option>
                <option value="rejected">已拒绝</option>
                <option value="cancelled">已取消</option>
              </select>
            </label>
            <div class="hidden md:block" />
            <div class="flex items-end gap-2">
              <Button type="submit" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" :disabled="reviewLoading">
                <Search class="size-3.5" />
                查询
              </Button>
              <Button type="button" variant="outline" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" @click="resetReviewFilters">
                重置
              </Button>
            </div>
          </form>
        </AdminFilterPanel>

        <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden xl:grid-cols-[minmax(0,1fr)_380px]">
          <AdminTablePanel :loading="reviewLoading" scroll-body>
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
                  <TableCell><StatusPill :status="review.status" /></TableCell>
                  <TableCell class="font-mono text-xs uppercase">{{ review.source || '-' }}</TableCell>
                  <TableCell class="font-mono text-xs">{{ review.order_id ? `#${review.order_id}` : '-' }}</TableCell>
                  <TableCell class="max-w-[180px] truncate font-mono text-xs">{{ review.payment_intent_id || '-' }}</TableCell>
                  <TableCell class="max-w-[220px] truncate text-xs">{{ review.reason || '-' }}</TableCell>
                  <TableCell class="text-xs text-muted-foreground">{{ formatDate(review.created_at) }}</TableCell>
                </TableRow>
                <TableRow v-if="!reviewLoading && reviews.length === 0">
                  <TableCell colspan="7" class="h-24 text-center text-sm text-muted-foreground">暂无支付复核记录</TableCell>
                </TableRow>
              </TableBody>
            </Table>
            <template #footer>
              <PaginationBar :pagination="reviewPagination" @page="updateReviewPage" @page-size="updateReviewPageSize" />
            </template>
          </AdminTablePanel>

          <aside class="min-h-0 overflow-auto rounded-[24px] border border-dashed border-border/80 bg-card p-4">
            <div class="mb-4">
              <h2 class="text-sm font-black uppercase italic tracking-tight">复核处理</h2>
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
                  <dd class="mt-1"><StatusPill :status="selectedReview.status" /></dd>
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
                <select v-model="reviewDecision.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm" :disabled="selectedReview.status !== 'pending'">
                  <option value="pending">保持待复核</option>
                  <option value="approved">通过</option>
                  <option value="rejected">拒绝</option>
                  <option value="cancelled">取消</option>
                </select>
              </label>
              <label class="block space-y-1">
                <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">OPERATOR NOTES / 操作备注</span>
                <Textarea v-model="reviewDecision.notes" rows="5" :disabled="selectedReview.status !== 'pending'" placeholder="记录证据、沟通结论或拒绝原因" />
              </label>
              <Button class="w-full rounded-full font-black uppercase tracking-wider" :disabled="selectedReview.status !== 'pending' || reviewSaving" @click="submitReviewDecision">
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
                <Button type="submit" variant="outline" class="w-full rounded-full font-black uppercase tracking-wider" :disabled="reviewSaving">
                  创建复核
                </Button>
              </form>
            </div>
          </aside>
        </section>
      </TabsContent>

      <TabsContent value="refunds" class="min-h-0 flex flex-col gap-3 overflow-hidden">
        <AdminFilterPanel class="shrink-0">
          <form class="grid grid-cols-1 gap-3 md:grid-cols-[180px_180px_1fr_auto]" @submit.prevent="applyRefundRecommendationFilters">
            <label class="block space-y-1">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
              <select v-model="refundRecommendationFilters.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
                <option value="">全部</option>
                <option value="pending">待处理</option>
                <option value="accepted">已采纳</option>
                <option value="dismissed">已驳回</option>
                <option value="cancelled">已取消</option>
              </select>
            </label>
            <label class="block space-y-1">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">PROVIDER / 渠道</span>
              <select v-model="refundRecommendationFilters.provider" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
                <option value="">全部</option>
                <option value="stripe">Stripe</option>
                <option value="paypal">PayPal</option>
              </select>
            </label>
            <div class="hidden md:block" />
            <div class="flex items-end gap-2">
              <Button type="submit" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" :disabled="refundRecommendationLoading">
                <Search class="size-3.5" />
                查询
              </Button>
              <Button type="button" variant="outline" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" @click="resetRefundRecommendationFilters">
                重置
              </Button>
            </div>
          </form>
        </AdminFilterPanel>

        <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden xl:grid-cols-[minmax(0,1fr)_400px]">
          <AdminTablePanel :loading="refundRecommendationLoading" scroll-body>
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
                  v-for="item in refundRecommendations"
                  :key="item.id"
                  class="cursor-pointer"
                  :class="selectedRefundRecommendation?.id === item.id ? 'bg-primary/5' : ''"
                  @click="selectRefundRecommendation(item)"
                >
                  <TableCell class="font-mono text-xs">#{{ item.id }}</TableCell>
                  <TableCell class="font-mono text-xs uppercase">{{ item.provider || '-' }}</TableCell>
                  <TableCell><StatusPill :status="item.status" /></TableCell>
                  <TableCell class="text-xs">{{ refundRecommendationSourceLabel(item.source_kind) }}</TableCell>
                  <TableCell class="font-mono text-xs">{{ formatMoney(item.recommended_amount, item.currency) }}</TableCell>
                  <TableCell class="font-mono text-xs">{{ item.order_id ? `#${item.order_id}` : '-' }}</TableCell>
                  <TableCell class="max-w-[180px] truncate font-mono text-xs">{{ item.provider_payment_id || item.payment_intent_id || item.charge_id || '-' }}</TableCell>
                  <TableCell class="text-xs" :class="isEvidenceSoon(item.review_by) ? 'font-semibold text-rose-600' : 'text-muted-foreground'">{{ formatDate(item.review_by) }}</TableCell>
                </TableRow>
                <TableRow v-if="!refundRecommendationLoading && refundRecommendations.length === 0">
                  <TableCell colspan="8" class="h-24 text-center text-sm text-muted-foreground">暂无退款建议</TableCell>
                </TableRow>
              </TableBody>
            </Table>
            <template #footer>
              <PaginationBar :pagination="refundRecommendationPagination" @page="updateRefundRecommendationPage" @page-size="updateRefundRecommendationPageSize" />
            </template>
          </AdminTablePanel>

          <PaymentRefundRecommendationDetailPanel
            v-model:decision-status="refundRecommendationDecision.status"
            v-model:decision-notes="refundRecommendationDecision.decision_notes"
            :recommendation="selectedRefundRecommendation"
            :saving="refundRecommendationSaving"
            :draft-saving="refundDraftSaving"
            :execution-saving="refundExecutionSaving"
            :execution-completed="selectedRefundExecutionCompleted"
            @save-decision="submitRefundRecommendationDecision"
            @create-draft="createPendingRefundFromRecommendation"
            @execute-refund="executeProviderRefundFromRecommendation"
          />
        </section>
      </TabsContent>

      <TabsContent value="disputes" class="min-h-0 flex flex-col gap-3 overflow-hidden">
        <AdminFilterPanel class="shrink-0">
          <form class="grid grid-cols-1 gap-3 md:grid-cols-[220px_1fr_auto]" @submit.prevent="applyDisputeFilters">
            <label class="block space-y-1">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
              <select v-model="disputeFilters.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
                <option value="">全部</option>
                <option value="needs_response">needs_response</option>
                <option value="warning_needs_response">warning_needs_response</option>
                <option value="under_review">under_review</option>
                <option value="won">won</option>
                <option value="lost">lost</option>
                <option value="closed">closed</option>
              </select>
            </label>
            <div class="hidden md:block" />
            <div class="flex items-end gap-2">
              <Button type="submit" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" :disabled="disputeLoading">
                <Search class="size-3.5" />
                查询
              </Button>
              <Button type="button" variant="outline" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" @click="resetDisputeFilters">
                重置
              </Button>
            </div>
          </form>
        </AdminFilterPanel>

        <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden xl:grid-cols-[minmax(0,1fr)_460px]">
          <AdminTablePanel :loading="disputeLoading" class="min-h-0" scroll-body>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Stripe Dispute</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>金额</TableHead>
                  <TableHead>原因</TableHead>
                  <TableHead>订单</TableHead>
                  <TableHead>证据截止</TableHead>
                  <TableHead>更新时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow
                  v-for="dispute in disputes"
                  :key="dispute.id"
                  class="cursor-pointer"
                  :class="selectedDispute?.id === dispute.id ? 'bg-primary/5' : ''"
                  @click="selectDispute(dispute)"
                >
                  <TableCell class="font-mono text-xs">#{{ dispute.id }}</TableCell>
                  <TableCell class="max-w-[220px] truncate font-mono text-xs">{{ dispute.stripe_dispute_id }}</TableCell>
                  <TableCell><StatusPill :status="dispute.status" /></TableCell>
                  <TableCell class="font-mono text-xs">{{ formatMoney(dispute.amount, dispute.currency) }}</TableCell>
                  <TableCell class="text-xs">{{ dispute.reason || '-' }}</TableCell>
                  <TableCell class="font-mono text-xs">{{ dispute.order_id ? `#${dispute.order_id}` : '-' }}</TableCell>
                  <TableCell class="text-xs" :class="isEvidenceSoon(dispute.evidence_due_at) ? 'font-semibold text-rose-600' : 'text-muted-foreground'">{{ formatDate(dispute.evidence_due_at) }}</TableCell>
                  <TableCell class="text-xs text-muted-foreground">{{ formatDate(dispute.updated_at) }}</TableCell>
                </TableRow>
                <TableRow v-if="!disputeLoading && disputes.length === 0">
                  <TableCell colspan="8" class="h-24 text-center text-sm text-muted-foreground">暂无 Stripe 拒付记录</TableCell>
                </TableRow>
              </TableBody>
            </Table>
            <template #footer>
              <PaginationBar :pagination="disputePagination" @page="updateDisputePage" @page-size="updateDisputePageSize" />
            </template>
          </AdminTablePanel>

          <aside class="min-h-0 overflow-auto rounded-[24px] border border-dashed border-border/80 bg-card p-4">
            <div class="mb-4 flex items-start justify-between gap-3">
              <div>
                <h2 class="text-sm font-black uppercase italic tracking-tight">拒付证据工作台</h2>
                <p class="mt-1 text-xs text-muted-foreground">先预览证据，再人工确认提交到 Stripe。</p>
              </div>
              <Button v-if="selectedDispute" variant="outline" size="sm" class="rounded-full text-xs font-black" :disabled="evidenceLoading" @click="fetchDisputeEvidence">
                <RefreshCw :class="['size-3.5', { 'animate-spin': evidenceLoading }]" />
              </Button>
            </div>

            <div v-if="selectedDispute" class="space-y-4">
              <dl class="grid grid-cols-2 gap-2 text-xs">
                <div class="rounded-xl bg-muted/40 p-3">
                  <dt class="font-black uppercase text-muted-foreground">Dispute</dt>
                  <dd class="mt-1 break-all font-mono">{{ selectedDispute.stripe_dispute_id }}</dd>
                </div>
                <div class="rounded-xl bg-muted/40 p-3">
                  <dt class="font-black uppercase text-muted-foreground">Status</dt>
                  <dd class="mt-1"><StatusPill :status="selectedDispute.status" /></dd>
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
                  <h3 class="mb-2 text-xs font-black uppercase tracking-widest text-muted-foreground">Stripe Evidence Preview</h3>
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
                  <label class="mb-2 flex items-center gap-2 text-xs font-bold text-muted-foreground">
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

                <section class="space-y-3 rounded-2xl border border-dashed border-border/80 p-3">
                  <h3 class="text-xs font-black uppercase tracking-widest text-muted-foreground">Stripe File IDs</h3>
                  <Input v-model="evidenceForm.shipping_documentation_file_id" placeholder="Shipping documentation File ID (file_...)" />
                  <Input v-model="evidenceForm.customer_communication_file_id" placeholder="Customer communication File ID (file_...)" />
                  <Input v-model="evidenceForm.receipt_file_id" placeholder="Receipt File ID (file_...)" />
                  <Input v-model="evidenceForm.uncategorized_file_id" placeholder="Uncategorized File ID (file_...)" />
                  <Textarea v-model="evidenceForm.additional_statement" rows="4" placeholder="人工补充说明，可选" />
                  <label class="flex items-start gap-2 text-xs font-bold text-muted-foreground">
                    <input v-model="evidenceForm.confirm" type="checkbox" class="mt-0.5 size-4 accent-primary" />
                    我已检查证据文本、物流凭证和沟通记录，确认提交给 Stripe。
                  </label>
                  <Button class="w-full rounded-full font-black uppercase tracking-wider" :disabled="!disputeEvidence.can_submit || !evidenceForm.confirm || evidenceSubmitting" @click="submitDisputeEvidence">
                    <Send class="size-4" />
                    提交证据到 Stripe
                  </Button>
                </section>
              </div>
            </div>
            <div v-else class="rounded-2xl border border-dashed border-border/80 p-6 text-center text-sm text-muted-foreground">
              选择左侧拒付记录后查看证据包。
            </div>
          </aside>
        </section>
      </TabsContent>

      <TabsContent v-if="activeTab === 'controls'" value="controls" class="min-h-0 flex flex-col gap-3 overflow-hidden">
        <PaymentRiskControlPanel />
      </TabsContent>
    </Tabs>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { AlertTriangle, CheckCircle2, CreditCard, RefreshCw, Search, Send, ShieldAlert } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import PaymentRefundRecommendationDetailPanel from '@/components/admin/payment/PaymentRefundRecommendationDetailPanel.vue'
import PaymentRiskControlPanel from '@/components/admin/payment/PaymentRiskControlPanel.vue'
import PaymentRiskSummaryPanel from '@/components/admin/payment/PaymentRiskSummaryPanel.vue'
import { paymentRefundApi } from '@/api/paymentRefunds'
import { paymentRiskApi } from '@/api/paymentRisk'
import { useRouteTab } from '@/composables/useRouteTab'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent } from '@/components/ui/tabs'
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
  }
})

const PaginationBar = defineComponent({
  props: { pagination: { type: Object, required: true } },
  emits: ['page', 'page-size'],
  setup(props, { emit }) {
    return () => h('div', { class: 'flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground' }, [
      h('span', `共 ${props.pagination.total || 0} 条，第 ${props.pagination.page || 1} / ${props.pagination.total_pages || 1} 页`),
      h('div', { class: 'flex items-center gap-2' }, [
        h('button', {
          class: 'rounded-full border border-dashed px-3 py-1 font-bold disabled:opacity-40',
          disabled: (props.pagination.page || 1) <= 1,
          onClick: () => emit('page', Math.max(1, (props.pagination.page || 1) - 1)),
        }, '上一页'),
        h('button', {
          class: 'rounded-full border border-dashed px-3 py-1 font-bold disabled:opacity-40',
          disabled: (props.pagination.page || 1) >= (props.pagination.total_pages || 1),
          onClick: () => emit('page', (props.pagination.page || 1) + 1),
        }, '下一页'),
        h('select', {
          class: 'h-7 rounded-full border border-dashed bg-background px-2',
          value: props.pagination.page_size || 20,
          onChange: (event) => emit('page-size', Number(event.target.value)),
        }, [10, 20, 50, 100].map(size => h('option', { value: size }, `${size}/页`))),
      ]),
    ])
  }
})

const activeTab = useRouteTab({
  defaultValue: 'reviews',
  values: ['reviews', 'refunds', 'disputes', 'controls'],
  routes: {
    reviews: 'PaymentRiskReviews',
    refunds: 'PaymentRiskRefundRecommendations',
    disputes: 'PaymentRiskDisputes',
    controls: 'PaymentRiskControls',
  },
})
const reviewLoading = ref(false)
const refundRecommendationLoading = ref(false)
const disputeLoading = ref(false)
const evidenceLoading = ref(false)
const reviewSaving = ref(false)
const refundRecommendationSaving = ref(false)
const refundDraftSaving = ref(false)
const refundExecutionSaving = ref(false)
const evidenceSubmitting = ref(false)
const summaryLoading = ref(false)
const reviews = ref([])
const refundRecommendations = ref([])
const disputes = ref([])
const riskSummary = reactive({ enabled: false, reports: {} })
const selectedReview = ref(null)
const selectedRefundRecommendation = ref(null)
const selectedDispute = ref(null)
const disputeEvidence = ref(null)
const reviewFilters = reactive({ status: 'pending' })
const refundRecommendationFilters = reactive({ status: 'pending', provider: '' })
const disputeFilters = reactive({ status: '' })
const reviewPagination = reactive({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const refundRecommendationPagination = reactive({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const disputePagination = reactive({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const reviewDecision = reactive({ status: 'pending', notes: '' })
const refundRecommendationDecision = reactive({ status: 'pending', decision_notes: '' })
const manualReview = reactive({ orderId: '', paymentIntentId: '', reason: '', notes: '' })
const evidenceForm = reactive({
  include_customer_communication: false,
  shipping_documentation_file_id: '',
  customer_communication_file_id: '',
  receipt_file_id: '',
  uncategorized_file_id: '',
  additional_statement: '',
  confirm: false,
})

const currentLoading = computed(() => {
  if (summaryLoading.value) return true
  if (activeTab.value === 'disputes') return disputeLoading.value
  if (activeTab.value === 'refunds') return refundRecommendationLoading.value
  return reviewLoading.value
})
const pendingReviews = computed(() => reviews.value.filter((item) => item.status === 'pending').length)
const pendingRefundRecommendations = computed(() => refundRecommendations.value.filter((item) => item.status === 'pending').length)
const urgentDisputes = computed(() => disputes.value.filter((item) => ['needs_response', 'warning_needs_response'].includes(item.status)).length)
const statItems = computed(() => [
  { key: 'reviews', label: '复核记录', value: reviewPagination.total || reviews.value.length, icon: CreditCard, tone: 'gray' },
  { key: 'pending', label: '待复核', value: pendingReviews.value, icon: AlertTriangle, tone: pendingReviews.value ? 'amber' : 'green' },
  { key: 'refund_recommendations', label: '退款建议', value: pendingRefundRecommendations.value, icon: AlertTriangle, tone: pendingRefundRecommendations.value ? 'amber' : 'green' },
  { key: 'disputes', label: '拒付记录', value: disputePagination.total || disputes.value.length, icon: ShieldAlert, tone: 'gray' },
  { key: 'urgent', label: '需响应拒付', value: urgentDisputes.value, icon: AlertTriangle, tone: urgentDisputes.value ? 'coral' : 'green' },
])
const evidenceFields = computed(() => {
  const evidence = disputeEvidence.value?.evidence || {}
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
const executedRefundIds = ref(new Set())
const selectedRefundExecutionCompleted = computed(() => {
  const refundID = selectedRefundRecommendation.value?.linked_refund_id
  return Boolean(refundID && executedRefundIds.value.has(refundID))
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
const refundRecommendationSourceLabel = (sourceKind) => {
  if (sourceKind === 'early_fraud_warning') return '早期欺诈预警'
  if (sourceKind === 'dispute') return '争议/拒付'
  return sourceKind || '-'
}
const isEvidenceSoon = (dateString) => {
  if (!dateString) return false
  const due = new Date(dateString).getTime()
  return Number.isFinite(due) && due - Date.now() < 3 * 24 * 60 * 60 * 1000
}

const applyPaged = (target, pagination, payload) => {
  target.value = payload.data || []
  Object.assign(pagination, {
    page: payload.pagination.page || 1,
    page_size: payload.pagination.page_size || 20,
    total: payload.pagination.total || 0,
    total_pages: payload.pagination.total_pages || 1,
  })
}

const fetchReviews = async () => {
  reviewLoading.value = true
  try {
    const payload = await paymentRiskApi.listReviews({
      page: reviewPagination.page,
      page_size: reviewPagination.page_size,
      status: reviewFilters.status || undefined,
    })
    applyPaged(reviews, reviewPagination, payload)
    if (selectedReview.value) {
      selectedReview.value = reviews.value.find((item) => item.id === selectedReview.value.id) || null
    }
  } finally {
    reviewLoading.value = false
  }
}

const fetchRefundRecommendations = async () => {
  refundRecommendationLoading.value = true
  try {
    const payload = await paymentRiskApi.listRefundRecommendations({
      page: refundRecommendationPagination.page,
      page_size: refundRecommendationPagination.page_size,
      status: refundRecommendationFilters.status || undefined,
      provider: refundRecommendationFilters.provider || undefined,
    })
    applyPaged(refundRecommendations, refundRecommendationPagination, payload)
    if (selectedRefundRecommendation.value) {
      selectedRefundRecommendation.value = refundRecommendations.value.find((item) => item.id === selectedRefundRecommendation.value.id) || null
    }
  } finally {
    refundRecommendationLoading.value = false
  }
}

const fetchDisputes = async () => {
  disputeLoading.value = true
  try {
    const payload = await paymentRiskApi.listDisputes({
      page: disputePagination.page,
      page_size: disputePagination.page_size,
      status: disputeFilters.status || undefined,
    })
    applyPaged(disputes, disputePagination, payload)
    if (selectedDispute.value) {
      selectedDispute.value = disputes.value.find((item) => item.id === selectedDispute.value.id) || null
      if (!selectedDispute.value) {
        disputeEvidence.value = null
      }
    }
  } finally {
    disputeLoading.value = false
  }
}

const fetchRiskSummary = async () => {
  summaryLoading.value = true
  try {
    const payload = await paymentRiskApi.getSummary()
    riskSummary.enabled = Boolean(payload.enabled)
    riskSummary.reports = payload.reports || {}
  } finally {
    summaryLoading.value = false
  }
}

const fetchActiveTabData = () => {
  if (activeTab.value === 'disputes') return fetchDisputes()
  if (activeTab.value === 'refunds') return fetchRefundRecommendations()
  return fetchReviews()
}

const refreshCurrent = async () => {
  await Promise.all([
    fetchRiskSummary(),
    fetchActiveTabData(),
  ])
}
const applyReviewFilters = () => { reviewPagination.page = 1; fetchReviews() }
const resetReviewFilters = () => { reviewFilters.status = 'pending'; reviewPagination.page = 1; fetchReviews() }
const applyRefundRecommendationFilters = () => { refundRecommendationPagination.page = 1; fetchRefundRecommendations() }
const resetRefundRecommendationFilters = () => {
  refundRecommendationFilters.status = 'pending'
  refundRecommendationFilters.provider = ''
  refundRecommendationPagination.page = 1
  fetchRefundRecommendations()
}
const applyDisputeFilters = () => { disputePagination.page = 1; fetchDisputes() }
const resetDisputeFilters = () => { disputeFilters.status = ''; disputePagination.page = 1; fetchDisputes() }
const updateReviewPage = (page) => { reviewPagination.page = page; fetchReviews() }
const updateReviewPageSize = (pageSize) => { reviewPagination.page_size = pageSize; reviewPagination.page = 1; fetchReviews() }
const updateRefundRecommendationPage = (page) => { refundRecommendationPagination.page = page; fetchRefundRecommendations() }
const updateRefundRecommendationPageSize = (pageSize) => {
  refundRecommendationPagination.page_size = pageSize
  refundRecommendationPagination.page = 1
  fetchRefundRecommendations()
}
const updateDisputePage = (page) => { disputePagination.page = page; fetchDisputes() }
const updateDisputePageSize = (pageSize) => { disputePagination.page_size = pageSize; disputePagination.page = 1; fetchDisputes() }

const selectReview = (review) => {
  selectedReview.value = review
  reviewDecision.status = review.status || 'pending'
  reviewDecision.notes = review.notes || ''
}

const selectRefundRecommendation = (recommendation) => {
  selectedRefundRecommendation.value = recommendation
  refundRecommendationDecision.status = recommendation.status || 'pending'
  refundRecommendationDecision.decision_notes = recommendation.decision_notes || ''
}

const resetEvidenceForm = () => {
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

const selectDispute = async (dispute) => {
  selectedDispute.value = dispute
  disputeEvidence.value = null
  resetEvidenceForm()
  await fetchDisputeEvidence()
}

const fetchDisputeEvidence = async () => {
  if (!selectedDispute.value) return
  evidenceLoading.value = true
  try {
    disputeEvidence.value = await paymentRiskApi.getDisputeEvidence(selectedDispute.value.id)
  } finally {
    evidenceLoading.value = false
  }
}

const submitReviewDecision = async () => {
  if (!selectedReview.value) return
  reviewSaving.value = true
  try {
    await paymentRiskApi.updateReview(selectedReview.value.id, {
      status: reviewDecision.status,
      notes: reviewDecision.notes,
    })
    toast.success('复核已保存')
    await fetchReviews()
  } finally {
    reviewSaving.value = false
  }
}

const submitRefundRecommendationDecision = async () => {
  if (!selectedRefundRecommendation.value) return
  refundRecommendationSaving.value = true
  try {
    await paymentRiskApi.updateRefundRecommendation(selectedRefundRecommendation.value.id, {
      status: refundRecommendationDecision.status,
      decision_notes: refundRecommendationDecision.decision_notes,
    })
    toast.success('退款建议处理已保存')
    await fetchRefundRecommendations()
  } finally {
    refundRecommendationSaving.value = false
  }
}

const createPendingRefundFromRecommendation = async (payload) => {
  if (!selectedRefundRecommendation.value) return
  refundDraftSaving.value = true
  try {
    const result = await paymentRiskApi.createPendingRefundFromRecommendation(selectedRefundRecommendation.value.id, payload)
    const refundID = result?.refund?.id
    selectedRefundRecommendation.value = result?.recommendation || selectedRefundRecommendation.value
    toast.success(refundID ? `已生成待处理退款 #${refundID}` : '已生成待处理退款')
    await fetchRefundRecommendations()
  } finally {
    refundDraftSaving.value = false
  }
}

const executeProviderRefundFromRecommendation = async (payload) => {
  const refundID = payload?.refund_id || selectedRefundRecommendation.value?.linked_refund_id
  if (!refundID) return
  refundExecutionSaving.value = true
  try {
    const result = await paymentRefundApi.executePendingRefund(refundID, { confirm: Boolean(payload?.confirm) })
    const providerRefundID = result?.execution?.provider_refund_id || result?.refund?.refund_id
    const next = new Set(executedRefundIds.value)
    next.add(refundID)
    executedRefundIds.value = next
    toast.success(providerRefundID ? `支付渠道退款已执行：${providerRefundID}` : '支付渠道退款已执行')
  } finally {
    refundExecutionSaving.value = false
  }
}

const createManualReview = async () => {
  reviewSaving.value = true
  try {
    await paymentRiskApi.createReview({
      order_id: manualReview.orderId ? Number(manualReview.orderId) : undefined,
      payment_intent_id: manualReview.paymentIntentId.trim() || undefined,
      reason: manualReview.reason.trim(),
      notes: manualReview.notes.trim(),
    })
    Object.assign(manualReview, { orderId: '', paymentIntentId: '', reason: '', notes: '' })
    toast.success('人工复核已创建')
    reviewPagination.page = 1
    await fetchReviews()
  } finally {
    reviewSaving.value = false
  }
}

const submitDisputeEvidence = async () => {
  if (!selectedDispute.value) return
  evidenceSubmitting.value = true
  try {
    await paymentRiskApi.submitDisputeEvidence(selectedDispute.value.id, {
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
    evidenceForm.confirm = false
    await fetchDisputes()
    await fetchDisputeEvidence()
  } finally {
    evidenceSubmitting.value = false
  }
}

onMounted(() => {
  fetchRiskSummary()
  fetchReviews()
  fetchRefundRecommendations()
  fetchDisputes()
})

watch(activeTab, refreshCurrent)
</script>
