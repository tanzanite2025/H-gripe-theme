<template>
  <TabsContent value="shipments" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">已发货</h2>
        <p class="mt-1 text-xs text-muted-foreground">直接读取订单的已发货状态；这里仅补充售后凭据。</p>
      </div>
      <div class="flex w-full flex-wrap gap-2 sm:w-auto">
        <div class="relative w-full sm:w-72">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            :model-value="filters.keyword || ''"
            class="h-9 rounded-full pl-9"
            placeholder="订单号 / 客户 / 物流单号"
            @update:model-value="$emit('update-search', String($event))"
          />
        </div>
        <Select v-model="filters.status">
          <SelectTrigger class="h-9 w-full rounded-full sm:w-36">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="active">保修有效</SelectItem>
            <SelectItem value="expired">已过期</SelectItem>
            <SelectItem value="cancelled">已取消</SelectItem>
            <SelectItem value="unbound">待补充凭据</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <section class="grid gap-4 2xl:grid-cols-[minmax(0,1fr)_390px]">
      <AdminTablePanel :loading="loading">
        <Table class="min-w-[1160px]">
          <TableHeader>
            <TableRow>
              <TableHead>订单 / 客户</TableHead>
              <TableHead>发货商品快照</TableHead>
              <TableHead>物流</TableHead>
              <TableHead>发货时间</TableHead>
              <TableHead>保修</TableHead>
              <TableHead>附加资料</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="shipments.length === 0" :colspan="6">
              <div class="flex flex-col items-center text-muted-foreground">
                <PackageCheck class="mb-2 size-7 opacity-55" />
                <span class="text-xs">暂无已发货订单</span>
              </div>
            </TableEmpty>
            <TableRow
              v-for="shipment in shipments"
              :key="shipment.order_id || shipment.id"
              class="cursor-pointer"
              :class="selectedShipment?.order_id === shipment.order_id ? 'bg-admin-selected-soft' : ''"
              @click="$emit('select-shipment', shipment)"
            >
              <TableCell>
                <span class="block font-mono text-xs font-black">{{ shipment.order_number || '-' }}</span>
                <span class="block max-w-64 truncate text-xs font-bold">{{ shipment.customer_name || '未填写姓名' }}</span>
                <span class="block max-w-64 truncate text-[10px] text-muted-foreground/70">
                  {{ shipment.customer_email || '-' }}
                </span>
              </TableCell>
              <TableCell>
                <span class="block max-w-72 truncate text-xs font-bold">
                  {{ firstItemLabel(shipment) }}
                </span>
                <span class="block text-[10px] text-muted-foreground/70">
                  {{ itemCountLabel(shipment) }}
                </span>
              </TableCell>
              <TableCell>
                <span class="block max-w-44 truncate font-mono text-xs font-bold">
                  {{ shipment.tracking_number || '无物流单号' }}
                </span>
                <span class="block text-[10px] text-muted-foreground/70">
                  {{ shipment.order_status || '-' }} / {{ shipment.shipping_status || '-' }}
                </span>
              </TableCell>
              <TableCell>
                <span class="block text-xs">{{ formatDateTime(shipment.shipped_at) }}</span>
                <span class="block text-[10px] text-muted-foreground/70">
                  起算 {{ formatDate(shipment.warranty_start_at) }}
                </span>
              </TableCell>
              <TableCell>
                <AdminStatusBadge :tone="warrantyTone(shipment)">
                  {{ warrantyLabel(shipment) }}
                </AdminStatusBadge>
                <span class="mt-1 block text-[10px] text-muted-foreground/70">
                  {{ shipment.warranty_months || 0 }} 个月 · 到期 {{ formatDate(shipment.warranty_expires) }}
                </span>
              </TableCell>
              <TableCell>
                <span class="block text-xs">{{ shipment.shipping_images?.length || 0 }} 张图片</span>
                <span class="block text-[10px] text-muted-foreground/70">
                  {{ shipment.product_codes?.length || 0 }} 个可选产品标识 · {{ shipment.record_bound ? '已补充' : '待补充' }}
                </span>
                <span class="block max-w-36 truncate text-[10px] text-muted-foreground/70">
                  {{ shipment.shipping_note || '无发货备注' }}
                </span>
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

      <Card class="relative overflow-hidden py-0">
        <div
          v-if="detailLoading"
          class="absolute inset-0 z-10 flex items-center justify-center bg-card/80 backdrop-blur-[1px]"
        >
          <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载发货详情" />
        </div>
        <CardHeader class="border-b bg-muted/30 px-4 py-3">
          <CardTitle class="flex items-center gap-2 text-sm">
            <PackageCheck class="size-4 text-primary" />
            订单售后凭据
          </CardTitle>
          <CardDescription>订单主数据只读；在此维护可选产品标识、备注、图片和保修期限。</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4 p-4">
          <div v-if="!selectedShipment" class="flex min-h-72 flex-col items-center justify-center text-center text-muted-foreground">
            <PackageCheck class="mb-2 size-8 opacity-55" />
            <p class="text-xs leading-6">从左侧选择一条已发货订单。</p>
          </div>

          <template v-else>
            <section class="rounded-2xl border p-3">
              <div class="mb-3 flex items-center justify-between gap-2">
                <h3 class="text-xs font-black uppercase tracking-wider">{{ selectedShipment.order_number }}</h3>
                <AdminStatusBadge :tone="warrantyTone(selectedShipment)">
                  {{ warrantyLabel(selectedShipment) }}
                </AdminStatusBadge>
              </div>
              <dl class="grid grid-cols-2 gap-2 text-xs">
                <DetailItem label="客户">{{ selectedShipment.customer_name || '-' }}</DetailItem>
                <DetailItem label="订单 ID">#{{ selectedShipment.order_id || '-' }}</DetailItem>
                <DetailItem label="邮箱" class="col-span-2">{{ selectedShipment.customer_email || '-' }}</DetailItem>
                <DetailItem label="物流单号" class="col-span-2">{{ selectedShipment.tracking_number || '-' }}</DetailItem>
                <DetailItem label="发货时间">{{ formatDateTime(selectedShipment.shipped_at) }}</DetailItem>
                <DetailItem label="保修到期">{{ formatDate(selectedShipment.warranty_expires) }}</DetailItem>
              </dl>
            </section>

            <section class="rounded-2xl border p-3">
              <div class="mb-2 flex items-center justify-between gap-2">
                <h3 class="text-xs font-black uppercase tracking-wider">可选产品标识</h3>
                <span class="text-[10px] font-mono text-muted-foreground/70">OPTIONAL EVIDENCE</span>
              </div>
              <Textarea
                :model-value="draft.productCodes.join('\n')"
                class="min-h-20"
                placeholder="每行填写一个可选产品标识；没有标识的商品可以留空。"
                :disabled="!canEdit || saving"
                @update:model-value="$emit('update-draft', { productCodes: String($event).split(/\r?\n/).map((value) => value.trim()).filter(Boolean) })"
              />
            </section>

            <section class="rounded-2xl border p-3">
              <div class="mb-2 flex items-center justify-between gap-2">
                <h3 class="text-xs font-black uppercase tracking-wider">商品快照</h3>
                <span class="text-[10px] font-mono text-muted-foreground/70">ORDER SNAPSHOT</span>
              </div>
              <div class="space-y-2">
                <div
                  v-for="(item, index) in selectedShipment.items_snapshot || []"
                  :key="item.id || `${item.product_id}-${index}`"
                  class="rounded-xl bg-muted/35 p-2"
                >
                  <div class="flex items-start justify-between gap-2">
                    <span class="text-xs font-bold">{{ item.product_name || item.sku || '商品' }}</span>
                    <span class="shrink-0 text-[10px] font-mono text-muted-foreground/70">×{{ item.quantity || 1 }}</span>
                  </div>
                  <span class="mt-1 block text-[10px] text-muted-foreground/70">
                    SKU {{ item.sku || '-' }} · variant {{ item.variant_id || '-' }}
                  </span>
                </div>
                <p v-if="!selectedShipment.items_snapshot?.length" class="text-xs text-muted-foreground">没有商品快照。</p>
              </div>
            </section>

            <section class="rounded-2xl border p-3">
              <div class="mb-2 flex items-center justify-between gap-2">
                <h3 class="text-xs font-black uppercase tracking-wider">发货备注</h3>
                <span class="text-[10px] font-mono text-muted-foreground/70">SHIPPING NOTE</span>
              </div>
              <Textarea
                :model-value="draft.shippingNote"
                class="min-h-28"
                placeholder="记录打包、外观、配件或特殊发货情况。"
                :disabled="!canEdit || saving"
                @update:model-value="$emit('update-draft', { shippingNote: String($event) })"
              />
            </section>

            <section class="rounded-2xl border p-3">
              <div class="mb-2 flex items-center justify-between gap-2">
                <h3 class="text-xs font-black uppercase tracking-wider">保修期限</h3>
                <span class="text-[10px] font-mono text-muted-foreground/70">WARRANTY WINDOW</span>
              </div>
              <div class="grid grid-cols-2 gap-2">
                <Input
                  :model-value="String(draft.warrantyMonths)"
                  type="number"
                  min="1"
                  max="120"
                  class="h-9 rounded-full"
                  :disabled="!canEdit || saving"
                  @update:model-value="$emit('update-draft', { warrantyMonths: Number($event) })"
                />
                <Input
                  :model-value="draft.warrantyStart"
                  type="date"
                  class="h-9 rounded-full"
                  :disabled="!canEdit || saving"
                  @update:model-value="$emit('update-draft', { warrantyStart: String($event) })"
                />
              </div>
              <div class="mt-2 flex items-center justify-between text-[10px] text-muted-foreground/70">
                <span>保修月数 / 起算日</span>
                <span>预计到期 {{ projectedWarrantyEnd }}</span>
              </div>
            </section>

            <section class="rounded-2xl border p-3">
              <div class="mb-2 flex items-center justify-between gap-2">
                <h3 class="text-xs font-black uppercase tracking-wider">发货凭据图片</h3>
                <span class="text-[10px] font-mono text-muted-foreground/70">{{ draft.shippingImages.length }} / 10</span>
              </div>
              <div v-if="draft.shippingImages.length" class="grid grid-cols-3 gap-2">
                <div v-for="image in draft.shippingImages" :key="image" class="group relative aspect-square overflow-hidden rounded-xl border bg-muted/30">
                  <a :href="image" target="_blank" rel="noopener noreferrer">
                    <img :src="image" alt="发货记录图片" class="size-full object-cover">
                  </a>
                  <Button
                    v-if="canEdit"
                    variant="destructive"
                    size="icon-sm"
                    class="absolute right-1 top-1 size-6 rounded-full opacity-0 transition-opacity group-hover:opacity-100"
                    aria-label="移除图片"
                    @click="$emit('remove-image', image)"
                  >
                    <X class="size-3" />
                  </Button>
                </div>
              </div>
              <p v-else class="text-xs text-muted-foreground">暂无图片。</p>
              <input
                ref="fileInput"
                type="file"
                :accept="uploadSpecAccept('warranty_evidence')"
                multiple
                class="hidden"
                :disabled="!canEdit || uploading"
                @change="handleFileChange"
              >
              <Button
                variant="outline"
                size="sm"
                class="mt-3 rounded-full"
                :disabled="!canEdit || uploading || draft.shippingImages.length >= 10"
                @click="openFilePicker"
              >
                <LoaderCircle v-if="uploading" class="size-3.5 animate-spin" />
                <ImagePlus v-else class="size-3.5" />
                {{ uploading ? '上传中' : '上传图片' }}
              </Button>
              <UploadSpecHint code="warranty_evidence" />
            </section>

            <div class="flex justify-end">
              <Button
                class="rounded-full font-black uppercase tracking-wider"
                :disabled="!canEdit || saving || !draft.warrantyStart || draft.warrantyMonths < 1"
                @click="$emit('save')"
              >
                <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
                <Save v-else class="size-3.5" />
                保存售后凭据
              </Button>
            </div>
          </template>
        </CardContent>
      </Card>
    </section>
  </TabsContent>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref } from 'vue'
import { ImagePlus, LoaderCircle, PackageCheck, Save, Search, X } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import { uploadSpecAccept } from '@/lib/uploadSpecs'
import { formatDate, formatDateTime } from '@/lib/warrantyPresentation'
import type {
  WarrantyFilters,
  WarrantyPagination,
  WarrantyShipmentDraft,
  WarrantyShipmentRecord
} from './warrantyTypes'

const DetailItem = defineComponent({
  inheritAttrs: false,
  props: {
    label: { type: String, required: true }
  },
  setup(props, { slots, attrs }) {
    return () => h('div', { class: ['space-y-1 rounded-xl border p-2', attrs.class] }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70' }, props.label),
      h('dd', { class: 'break-words font-bold text-foreground' }, slots.default ? slots.default() : '-')
    ])
  }
})

const props = withDefaults(defineProps<{
  shipments?: WarrantyShipmentRecord[]
  loading?: boolean
  detailLoading?: boolean
  saving?: boolean
  uploading?: boolean
  filters: WarrantyFilters
  pagination: WarrantyPagination
  selectedShipment?: WarrantyShipmentRecord | null
  draft: WarrantyShipmentDraft
  canEdit?: boolean
}>(), {
  shipments: () => [],
  loading: false,
  detailLoading: false,
  saving: false,
  uploading: false,
  selectedShipment: null,
  canEdit: false
})

const emit = defineEmits<{
  (event: 'select-shipment', shipment: WarrantyShipmentRecord): void
  (event: 'update-search', keyword: string): void
  (event: 'update-draft', patch: Partial<WarrantyShipmentDraft>): void
  (event: 'remove-image', image: string): void
  (event: 'save'): void
  (event: 'upload-images', files: File[]): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const draft = computed(() => props.draft)
const projectedWarrantyEnd = computed(() => {
  if (!draft.value.warrantyStart || draft.value.warrantyMonths < 1) return '-'
  const date = new Date(`${draft.value.warrantyStart}T00:00:00`)
  date.setMonth(date.getMonth() + Number(draft.value.warrantyMonths))
  return formatDate(date)
})

const openFilePicker = (): void => {
  fileInput.value?.click()
}

const handleFileChange = (event: Event): void => {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (files.length) emit('upload-images', files)
}

const remainingDays = (shipment: WarrantyShipmentRecord): number => {
  const expiry = shipment.warranty_expires ? new Date(shipment.warranty_expires).getTime() : 0
  if (!expiry) return 0
  return Math.max(0, Math.ceil((expiry - Date.now()) / 86_400_000))
}

const warrantyLabel = (shipment: WarrantyShipmentRecord): string => {
  if (shipment.status === 'cancelled') return '已取消'
  const days = remainingDays(shipment)
  return days > 0 ? `剩余 ${days} 天` : '已过期'
}

const warrantyTone = (shipment: WarrantyShipmentRecord): 'green' | 'amber' | 'coral' | 'gray' => {
  if (shipment.status === 'cancelled') return 'gray'
  return remainingDays(shipment) > 0 ? 'green' : 'amber'
}

const firstItemLabel = (shipment: WarrantyShipmentRecord): string => {
  const item = shipment.items_snapshot?.[0]
  if (!item) return '无商品快照'
  const suffix = (shipment.items_snapshot?.length || 0) > 1 ? ` 等 ${shipment.items_snapshot?.length} 项` : ''
  return `${item.product_name || item.sku || '商品'}${suffix}`
}

const itemCountLabel = (shipment: WarrantyShipmentRecord): string => {
  const quantity = (shipment.items_snapshot || []).reduce((total, item) => total + Number(item.quantity || 0), 0)
  return `${shipment.items_snapshot?.length || 0} 个商品行 · 数量 ${quantity || 0}`
}

</script>
