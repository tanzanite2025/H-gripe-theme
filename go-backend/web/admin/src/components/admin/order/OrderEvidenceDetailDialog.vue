<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent size="xl">
      <DialogHeader>
        <div class="flex flex-wrap items-start justify-between gap-3 pr-8">
          <div>
            <DialogTitle>履约证据包</DialogTitle>
            <DialogDescription>
              {{ order?.order_number || `订单 #${order?.order_id || '-'}` }} · 订单域独立证据记录
            </DialogDescription>
          </div>
          <div v-if="result?.package" class="flex flex-wrap gap-2">
            <Button
              v-if="result.package.status === 'locked'"
              variant="outline"
              size="sm"
              :disabled="exporting"
              @click="$emit('export')"
            >
              <LoaderCircle v-if="exporting" class="size-3.5 animate-spin" />
              <Download v-else class="size-3.5" />
              导出锁定清单
            </Button>
            <Button
              v-if="canEdit && result.package.status === 'ready'"
              size="sm"
              :disabled="locking"
              @click="$emit('lock')"
            >
              <LockKeyhole class="size-3.5" />
              锁定证据包
            </Button>
            <Button
              v-if="canEdit && result.package.status === 'locked'"
              variant="outline"
              size="sm"
              :disabled="revising"
              @click="$emit('revision')"
            >
              <GitBranch class="size-3.5" />
              创建修订版
            </Button>
          </div>
        </div>
      </DialogHeader>

      <div v-if="loading" class="flex min-h-56 items-center justify-center text-muted-foreground">
        <LoaderCircle class="size-5 animate-spin text-primary" />
      </div>

      <div v-else-if="!result?.package" class="flex min-h-56 flex-col items-center justify-center text-center">
        <FileWarning class="mb-2 size-8 text-amber-500" />
        <p class="text-sm font-black">该订单尚未生成证据包</p>
        <p class="mt-1 max-w-md text-xs leading-5 text-muted-foreground">
          当前列表仍保留该订单，便于核对历史数据。证据包生成属于订单创建边界，不在后台凭空创建。
        </p>
      </div>

      <div v-else class="space-y-4">
        <section class="grid grid-cols-2 gap-2 sm:grid-cols-4">
          <DetailItem label="包状态">
            <AdminStatusBadge :tone="orderEvidencePackageStatusTone(result.package.status)">
              {{ orderEvidencePackageStatusName(result.package.status) }}
            </AdminStatusBadge>
          </DetailItem>
          <DetailItem label="版本">v{{ result.package.package_version }}</DetailItem>
          <DetailItem label="订单 USD 快照">
            {{ formatMoney(result.package.order_total_usd_snapshot, 'USD') }}
          </DetailItem>
          <DetailItem label="完整度">
            {{ result.completeness.percent }}% · {{ result.completeness.satisfied }}/{{ result.completeness.total }}
          </DetailItem>
        </section>

        <section class="rounded-2xl border border-dashed border-border/80 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <h3 class="text-xs font-black uppercase tracking-wider">证据项</h3>
              <p class="mt-1 text-[11px] text-muted-foreground">
                配置确认单是订单创建时快照；其他履约记录在此补录。
              </p>
            </div>
            <span class="text-[10px] font-mono text-muted-foreground/70">
              {{ result.completeness.complete }} 完成 · {{ result.completeness.waived }} 豁免 · {{ result.completeness.pending }} 待处理
            </span>
          </div>

          <div class="mt-3 overflow-x-auto">
            <Table class="min-w-[800px]">
              <TableHeader>
                <TableRow>
                  <TableHead>证据项</TableHead>
                  <TableHead>范围</TableHead>
                  <TableHead>要求来源</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>附件</TableHead>
                  <TableHead class="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in result.package.items || []" :key="String(item.id)">
                  <TableCell>
                    <span class="block text-xs font-black">{{ orderEvidenceItemTypeName(item.item_type) }}</span>
                    <span class="block text-[10px] text-muted-foreground/70">#{{ item.id }}</span>
                  </TableCell>
                  <TableCell class="text-xs">{{ orderEvidenceItemScopeLabel(item) }}</TableCell>
                  <TableCell class="text-xs">{{ orderEvidenceRequiredReasonName(item.required_reason) }}</TableCell>
                  <TableCell>
                    <AdminStatusBadge :tone="orderEvidenceItemStatusTone(item.status)">
                      {{ orderEvidenceItemStatusName(item.status) }}
                    </AdminStatusBadge>
                  </TableCell>
                  <TableCell class="text-xs text-muted-foreground">
                    {{ item.attachments?.length || 0 }} 个
                  </TableCell>
                  <TableCell class="text-right">
                    <Button
                      v-if="!orderEvidenceItemIsReadOnly(item)"
                      variant="outline"
                      size="sm"
                      class="rounded-full"
                      :disabled="!canEdit || result.package.status === 'locked' || result.package.status === 'superseded' || savingItemId === item.id"
                      @click="$emit('edit-item', item)"
                    >
                      <LoaderCircle v-if="savingItemId === item.id" class="size-3.5 animate-spin" />
                      <Pencil v-else class="size-3.5" />
                      录入
                    </Button>
                    <span v-else class="text-[10px] text-muted-foreground/70">订单快照</span>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </section>

        <section class="grid gap-3 sm:grid-cols-2">
          <div class="rounded-2xl border border-dashed border-border/80 p-3">
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">DELIVERY CONTEXT / 物流上下文</span>
            <div v-if="result.tracking_context?.shipments?.length" class="mt-2 space-y-2 text-xs">
              <div v-for="shipment in result.tracking_context.shipments" :key="String(shipment.id)" class="rounded-lg border border-dashed p-2">
                <p class="font-bold">
                  {{ shipment.tracking_number || '-' }}
                  <span v-if="shipment.provider_name" class="font-normal text-muted-foreground">
                    · {{ shipment.provider_name }}
                  </span>
                </p>
                <p class="text-muted-foreground">
                  同步 {{ shipment.sync_status || '-' }} · 事件 {{ shipment.event_count }}
                </p>
                <p v-for="event in (result.tracking_context.latest_delivery_events || []).filter((candidate) => candidate.tracking_number === shipment.tracking_number)" :key="String(event.id)" class="text-muted-foreground">
                  最新交付：{{ event.status || '-' }} · {{ formatDate(event.event_time) }}
                  <span v-if="event.location"> · {{ event.location }}</span>
                </p>
              </div>
              <a
                v-for="url in result.tracking_context.provider_pod_urls || []"
                :key="url"
                :href="url"
                target="_blank"
                rel="noreferrer"
                class="inline-flex text-primary underline underline-offset-2"
              >
                查看供应商 POD
              </a>
            </div>
            <p v-else class="mt-2 text-xs text-muted-foreground">尚未关联物流追踪单。</p>
          </div>
          <div class="rounded-2xl border border-dashed border-border/80 p-3">
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">MANUAL POD / 人工证据</span>
            <div v-if="result.tracking_context?.manual_pod" class="mt-2 space-y-1 text-xs">
              <p class="flex items-center gap-1.5">
                <span>状态：</span>
                <AdminStatusBadge :tone="orderEvidenceItemStatusTone(result.tracking_context.manual_pod.status)">
                  {{ orderEvidenceItemStatusName(result.tracking_context.manual_pod.status) }}
                </AdminStatusBadge>
              </p>
              <p class="text-muted-foreground">
                附件 {{ result.tracking_context.manual_pod.attachment_count }} 个
                <span v-if="result.tracking_context.manual_pod.captured_at">
                  · {{ formatDate(result.tracking_context.manual_pod.captured_at) }}
                </span>
              </p>
              <p class="flex flex-wrap items-center gap-1.5 text-muted-foreground">
                <span>物流关联：</span>
                <AdminStatusBadge :tone="orderEvidenceTrackingAssociationTone(result.tracking_context.manual_pod.tracking_association_status)">
                  {{ orderEvidenceTrackingAssociationName(result.tracking_context.manual_pod.tracking_association_status) }}
                </AdminStatusBadge>
                <span v-if="result.tracking_context.manual_pod.tracking_number">
                  · {{ result.tracking_context.manual_pod.tracking_number }}
                </span>
              </p>
            </div>
            <p v-else class="mt-2 text-xs text-muted-foreground">未找到人工 POD 证据项。</p>
          </div>
          <div class="rounded-2xl border border-dashed border-border/80 p-3">
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">TRIGGERS / 触发条件</span>
            <div class="mt-2 flex flex-wrap gap-2">
              <AdminStatusBadge v-if="result.package.is_high_value" tone="amber">订单 ≥ 750 USD</AdminStatusBadge>
              <AdminStatusBadge v-if="result.package.has_spoke_tension_qc" tone="blue">产品规则要求张力表</AdminStatusBadge>
              <span v-if="!result.package.is_high_value && !result.package.has_spoke_tension_qc" class="text-xs text-muted-foreground">
                仅基础订单证据
              </span>
            </div>
          </div>
          <div class="rounded-2xl border border-dashed border-border/80 p-3">
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">IMMUTABILITY / 版本</span>
            <p class="mt-2 text-xs leading-5 text-muted-foreground">
              {{ result.package.status === 'locked' ? `已于 ${formatDate(result.package.locked_at)} 锁定；修改请创建修订版。` : '当前版本尚未锁定，可补录履约证据。' }}
            </p>
          </div>
        </section>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { defineComponent, h } from 'vue'
import { Download, FileWarning, GitBranch, LoaderCircle, LockKeyhole, Pencil } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatDate, formatMoney } from '@/lib/orderPresentation'
import {
  orderEvidenceItemIsReadOnly,
  orderEvidenceItemScopeLabel,
  orderEvidenceItemStatusName,
  orderEvidenceItemStatusTone,
  orderEvidenceItemTypeName,
  orderEvidencePackageStatusName,
  orderEvidencePackageStatusTone,
  orderEvidenceRequiredReasonName,
  orderEvidenceTrackingAssociationName,
  orderEvidenceTrackingAssociationTone,
} from '@/lib/orderEvidencePresentation'
import type { OrderEvidenceItem, OrderEvidenceOrderSummary, OrderEvidencePackageResult } from '@/modules/order/orderEvidenceTypes'

const DetailItem = defineComponent({
  inheritAttrs: false,
  props: { label: { type: String, required: true } },
  setup(props, { slots, attrs }) {
    return () => h('div', { class: ['space-y-1 rounded-xl border p-2', attrs.class] }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70' }, props.label),
      h('dd', { class: 'break-words text-xs font-bold text-foreground' }, slots.default ? slots.default() : '-'),
    ])
  },
})

withDefaults(defineProps<{
  open?: boolean
  loading?: boolean
  result?: OrderEvidencePackageResult | null
  order?: OrderEvidenceOrderSummary | null
  canEdit?: boolean
  savingItemId?: number | string | null
  locking?: boolean
  revising?: boolean
  exporting?: boolean
}>(), {
  open: false,
  loading: false,
  result: null,
  order: null,
  canEdit: false,
  savingItemId: null,
  locking: false,
  revising: false,
  exporting: false,
})

defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'edit-item', item: OrderEvidenceItem): void
  (event: 'lock'): void
  (event: 'revision'): void
  (event: 'export'): void
}>()
</script>
