<template>
  <TabsContent value="claims" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">保修申请处理</h2>
        <p class="mt-1 text-xs text-muted-foreground">保修申请按订单号进入这里，必要时绑定具体订单行。</p>
      </div>
      <div class="w-full sm:w-48">
        <Select v-model="filters.status">
          <SelectTrigger class="h-9 w-full rounded-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem v-for="option in statusOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <section class="grid gap-4 2xl:grid-cols-[minmax(0,1fr)_380px]">
      <AdminTablePanel :loading="loading">
        <Table class="min-w-[1180px]">
          <TableHeader>
            <TableRow>
              <TableHead>申请来源</TableHead>
              <TableHead>关联商品</TableHead>
              <TableHead>问题说明</TableHead>
              <TableHead>证据</TableHead>
              <TableHead>提交时间</TableHead>
              <TableHead class="w-40">状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="claims.length === 0" :colspan="6">
              <div class="flex flex-col items-center text-muted-foreground">
                <FileWarning class="mb-2 size-7 opacity-55" />
                <span class="text-xs">暂无保修申请</span>
              </div>
            </TableEmpty>
            <TableRow
              v-for="claim in claims"
              :key="claim.id"
              class="cursor-pointer"
 :class="selectedClaim?.id === claim.id ? 'bg-admin-selected-soft': ''"
              @click="$emit('select-claim', claim)"
            >
              <TableCell>
                <span class="block font-mono text-xs font-bold">#{{ claim.id }}</span>
                <span class="block max-w-60 truncate text-[10px] text-muted-foreground/70">
                  订单：{{ claim.order_number || '-' }}
                </span>
                <span class="block max-w-60 truncate text-[10px] text-muted-foreground/70">
                  邮箱：{{ claim.email || '-' }}
                </span>
              </TableCell>
              <TableCell>
                <span class="block text-xs font-bold">{{ claimProductName(claim) }}</span>
                <span class="block font-mono text-[10px] text-muted-foreground/70">
                  order_item_id={{ claim.order_item_id || '-' }}
                </span>
              </TableCell>
              <TableCell>
                <span class="block text-xs font-bold">{{ issueTypeLabel(claim.issue_type) }}</span>
                <span class="block max-w-[26rem] truncate text-[10px] text-muted-foreground/70">
                  {{ claim.description || '-' }}
                </span>
                <span v-if="claim.tire_pressure || claim.is_tubeless" class="block text-[10px] text-muted-foreground/70">
                  胎压 {{ claim.tire_pressure || '-' }} / {{ claim.is_tubeless ? '真空胎' : '非真空胎' }}
                </span>
              </TableCell>
              <TableCell>
                <div class="flex flex-wrap gap-1.5">
                  <Badge v-for="(image, index) in claimImages(claim)" :key="`${claim.id}-image-${index}`" variant="outline" class="rounded-full">
                    <a :href="image" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1 no-underline">
                      <ImageIcon class="size-3" />
                      图 {{ index + 1 }}
                    </a>
                  </Badge>
                  <Badge v-if="claim.video_url" variant="outline" class="rounded-full">
                    <a :href="claim.video_url" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1 no-underline">
                      <Video class="size-3" />
                      视频
                    </a>
                  </Badge>
                  <span v-if="claimImages(claim).length === 0 && !claim.video_url" class="text-xs text-muted-foreground">无附件</span>
                </div>
              </TableCell>
              <TableCell>
                <span class="block text-xs">{{ formatDateTime(claim.created_at) }}</span>
                <span class="block text-[10px] text-muted-foreground/70">处理：{{ formatDateTime(claim.processed_at) }}</span>
              </TableCell>
              <TableCell>
                <Select
                  :model-value="claim.status || 'submitted'"
                  :disabled="statusUpdating.claim === claim.id || !canEdit"
                  @update:model-value="$emit('update-status', claim, String($event))"
                  @click.stop
                >
                  <SelectTrigger class="h-8 w-full rounded-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="option in statusOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <template #footer>
          <AdminPagination
            :page="pagination.page"
            :page-size="pagination.pageSize"
            :total="pagination.total"
            @update:page="$emit('update-page', $event)"
            @update:page-size="$emit('update-page-size', $event)"
          />
        </template>
      </AdminTablePanel>

      <WarrantyClaimDetailPanel
        :detail-loading="detailLoading"
        :resolution-saving="resolutionSaving"
        :order-items-loading="orderItemsLoading"
        :order-item-binding="orderItemBinding"
        :service-records-loading="serviceRecordsLoading"
        :service-record-creating="serviceRecordCreating"
        :selected-claim="selectedClaim"
        :resolution-draft="resolutionDraft"
        :order-items="orderItems"
        :order-item-selection="orderItemSelection"
        :service-records="serviceRecords"
        :service-record-form="serviceRecordForm"
        :service-type-options="serviceTypeOptions"
        :service-status-options="serviceStatusOptions"
        :can-edit="canEdit"
        @update-order-item-selection="$emit('update-order-item-selection', $event)"
        @bind-order-item="$emit('bind-order-item')"
        @update-resolution-draft="$emit('update-resolution-draft', $event)"
        @save-resolution="$emit('save-resolution')"
        @update-service-record-form="$emit('update-service-record-form', $event)"
        @create-service-record="$emit('create-service-record')"
      />
    </section>
  </TabsContent>
</template>

<script setup lang="ts">
import { FileWarning, Image as ImageIcon, LoaderCircle, Video } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import WarrantyClaimDetailPanel from '@/components/admin/warranty/WarrantyClaimDetailPanel.vue'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import {
  claimImages,
  claimProductName,
  formatDateTime,
  issueTypeLabel
} from '@/lib/warrantyPresentation'
import type {
  WarrantyClaim,
  WarrantyFilters,
  WarrantyOrderItem,
  WarrantyPagination,
  WarrantyServiceRecord,
  WarrantyServiceRecordForm,
  WarrantyStatusOption,
  WarrantyStatusUpdating
} from '@/modules/warranty/warrantyTypes'

withDefaults(defineProps<{
  claims?: WarrantyClaim[]
  loading?: boolean
  detailLoading?: boolean
  resolutionSaving?: boolean
  orderItemsLoading?: boolean
  orderItemBinding?: boolean
  serviceRecordsLoading?: boolean
  serviceRecordCreating?: boolean
  filters: WarrantyFilters
  pagination: WarrantyPagination
  statusUpdating: WarrantyStatusUpdating
  statusOptions: WarrantyStatusOption[]
  selectedClaim?: WarrantyClaim | null
  resolutionDraft?: string
  orderItems?: WarrantyOrderItem[]
  orderItemSelection?: string
  serviceRecords?: WarrantyServiceRecord[]
  serviceRecordForm: WarrantyServiceRecordForm
  serviceTypeOptions: WarrantyStatusOption[]
  serviceStatusOptions: WarrantyStatusOption[]
  canEdit?: boolean
}>(), {
  claims: () => [],
  loading: false,
  detailLoading: false,
  resolutionSaving: false,
  orderItemsLoading: false,
  orderItemBinding: false,
  serviceRecordsLoading: false,
  serviceRecordCreating: false,
  selectedClaim: null,
  resolutionDraft: '',
  orderItems: () => [],
  orderItemSelection: 'none',
  serviceRecords: () => [],
  canEdit: false
})

defineEmits<{
  (event: 'update-status', claim: WarrantyClaim, status: string): void
  (event: 'select-claim', claim: WarrantyClaim): void
  (event: 'update-order-item-selection', value: string): void
  (event: 'bind-order-item'): void
  (event: 'update-resolution-draft', value: string): void
  (event: 'save-resolution'): void
  (event: 'update-service-record-form', patch: Partial<WarrantyServiceRecordForm>): void
  (event: 'create-service-record'): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()
</script>

