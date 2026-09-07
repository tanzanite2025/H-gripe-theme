<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl">
      <form @submit.prevent="submit">
        <DialogHeader>
          <DialogTitle>发货取证 · {{ order?.order_number || `订单 #${order?.id || '-'}` }}</DialogTitle>
          <DialogDescription>
            发货前一次性保存产品、包装和必要的张力表照片；提交成功后才会写入发货状态。
          </DialogDescription>
        </DialogHeader>

        <div class="max-h-[70vh] space-y-4 overflow-y-auto py-4 pr-1">
          <div v-if="loading" class="flex min-h-32 items-center justify-center text-muted-foreground">
            <LoaderCircle class="size-5 animate-spin text-primary" />
          </div>

          <template v-else-if="result?.package">
            <section class="grid gap-3 sm:grid-cols-3">
              <div class="summary-box">
                <span class="field-label">ORDER / 订单金额</span>
                <strong>{{ formatMoney(result.package.order_total_usd_snapshot, 'USD') }} USD</strong>
                <span v-if="result.package.is_high_value" class="summary-note text-amber-700">高价值订单策略</span>
              </div>
              <div class="summary-box">
                <span class="field-label">EVIDENCE / 当前完整度</span>
                <strong>{{ result.completeness.percent }}%</strong>
                <span class="summary-note">{{ result.completeness.satisfied }}/{{ result.completeness.total }} 已满足</span>
              </div>
              <div class="summary-box">
                <span class="field-label">POD / 签收证据</span>
                <strong class="text-blue-700">发货后补录</strong>
                <span class="summary-note">不会阻塞本次发货</span>
              </div>
            </section>

            <section class="space-y-3">
              <div>
                <h3 class="section-title">产品照片 / PRODUCT IDENTITY</h3>
                <p class="section-description">每个订单行至少上传一张发货前实物照片。没有序列号时填写批次、SKU 或人工识别备注即可。</p>
              </div>

              <div
                v-for="orderItem in order?.items || []"
                :key="String(orderItem.id)"
                class="evidence-row"
              >
                <div class="min-w-0">
                  <p class="truncate text-sm font-black">{{ orderItem.product_name || '未命名产品' }}</p>
                  <p class="mt-1 text-[11px] text-muted-foreground">
                    {{ orderItem.sku || '无 SKU' }} · 数量 {{ orderItem.quantity || 0 }} · 订单行 #{{ orderItem.id || '-' }}
                  </p>
                  <p class="mt-1 text-[11px] text-muted-foreground">
                    已有附件 {{ attachmentCount(identityItem(orderItem.id)) }} 个
                  </p>
                </div>

                <div class="grid gap-3 sm:grid-cols-3">
                  <label class="block space-y-1">
                    <span class="field-label">SERIAL / 序列号</span>
                    <Input v-model="identityDraft(orderItem.id).serial_number" placeholder="没有可留空" />
                  </label>
                  <label class="block space-y-1">
                    <span class="field-label">BATCH / 批次</span>
                    <Input v-model="identityDraft(orderItem.id).batch_number" placeholder="批次或生产批号" />
                  </label>
                  <label class="block space-y-1">
                    <span class="field-label">IDENTIFIER / 人工识别备注</span>
                    <Input v-model="identityDraft(orderItem.id).capture_note" placeholder="例如 SKU + 订单行唯一标识" />
                  </label>
                </div>

                <div class="flex flex-wrap items-center gap-2">
                  <input
                    :id="`identity-file-${orderItem.id}`"
                    type="file"
                    accept="image/jpeg,image/png,image/webp,image/gif"
                    multiple
                    class="sr-only"
                    :disabled="submitting"
                    @change="handleFiles(orderItem.id, 'identity', $event)"
                  >
                  <label :for="`identity-file-${orderItem.id}`" class="upload-button">
                    <Camera class="size-3.5" />
                    添加产品照片
                  </label>
                  <span v-for="file in identityDraft(orderItem.id).files" :key="fileKey(file)" class="file-chip">
                    {{ file.name }}
                    <button
                      type="button"
                      class="remove-file"
                      :aria-label="`移除 ${file.name}`"
                      @click="removeFile(orderItem.id, 'identity', file)"
                    >
                      <X class="size-3" />
                    </button>
                  </span>
                </div>
              </div>
            </section>

            <section v-if="outboundItem" class="space-y-3">
              <div>
                <h3 class="section-title">出库包装照片 / OUTBOUND PACKAGING</h3>
                <p class="section-description">记录实际打包后的毛重、包裹数量和包装方式，并上传照片。</p>
              </div>
              <div class="evidence-row">
                <div class="grid gap-3 sm:grid-cols-2">
                  <label class="block space-y-1">
                    <span class="field-label">GROSS WEIGHT / 毛重（克）</span>
                    <Input v-model="outbound.gross_weight_g" type="number" min="1" step="1" placeholder="例如 12640" />
                  </label>
                  <label class="block space-y-1">
                    <span class="field-label">PACKAGES / 包裹数量</span>
                    <Input v-model="outbound.package_count" type="number" min="1" max="100" step="1" placeholder="例如 1" />
                  </label>
                </div>
                <label class="block space-y-1">
                  <span class="field-label">PACKAGING METHOD / 包装方式</span>
                  <Input v-model="outbound.packaging_method" placeholder="例如双层瓦楞纸箱、木箱" />
                </label>
                <label class="block space-y-1">
                  <span class="field-label">PACKAGING NOTE / 包装备注（可选）</span>
                  <Textarea v-model="outbound.packaging_note" class="min-h-16" placeholder="防护、加固或分包情况" />
                </label>
                <div class="flex flex-wrap items-center gap-2">
                  <input
                    id="outbound-files"
                    type="file"
                    accept="image/jpeg,image/png,image/webp,image/gif"
                    multiple
                    class="sr-only"
                    :disabled="submitting"
                    @change="handleFiles(outboundItem.id, 'outbound', $event)"
                  >
                  <label for="outbound-files" class="upload-button">
                    <PackageOpen class="size-3.5" />
                    添加包装照片
                  </label>
                  <span class="text-[11px] text-muted-foreground">
                    已有附件 {{ attachmentCount(outboundItem) }} 个
                  </span>
                  <span v-for="file in outbound.files" :key="fileKey(file)" class="file-chip">
                    {{ file.name }}
                    <button
                      type="button"
                      class="remove-file"
                      :aria-label="`移除 ${file.name}`"
                      @click="removeFile(outboundItem.id, 'outbound', file)"
                    >
                      <X class="size-3" />
                    </button>
                  </span>
                </div>
              </div>
            </section>

            <section v-if="tensionItems.length" class="space-y-3">
              <div>
                <h3 class="section-title">编轮质检张力表 / SPOKE QC TENSION</h3>
                <p class="section-description">只对订单创建时快照明确要求的订单行显示。系统只保存人工照片、录入和备注，不判断张力合格或不合格。</p>
              </div>
              <div v-for="item in tensionItems" :key="String(item.id)" class="evidence-row">
                <div>
                  <p class="text-sm font-black">订单行 #{{ item.order_item_id }}</p>
                  <p class="mt-1 text-[11px] text-muted-foreground">
                    已有附件 {{ attachmentCount(item) }} 个 · 规则已在下单时锁定
                  </p>
                </div>
                <div class="grid gap-3 sm:grid-cols-2">
                  <label class="block space-y-1">
                    <span class="field-label">ASSEMBLY / 装配引用（可选）</span>
                    <Input v-model="tensionDraft(item.id).assembly_reference" placeholder="例如 wheelset-001" />
                  </label>
                  <label class="block space-y-1">
                    <span class="field-label">TOOL / 测量工具（可选）</span>
                    <Input v-model="tensionDraft(item.id).measurement_tool" placeholder="工具编号或型号" />
                  </label>
                </div>
                <label class="block space-y-1">
                  <span class="field-label">NOTE / 人工备注（可选）</span>
                  <Textarea v-model="tensionDraft(item.id).capture_note" class="min-h-16" placeholder="原样记录人工说明，不填写系统判定" />
                </label>
                <div class="flex flex-wrap items-center gap-2">
                  <input
                    :id="`tension-file-${item.id}`"
                    type="file"
                    accept="image/jpeg,image/png,image/webp,image/gif"
                    multiple
                    class="sr-only"
                    :disabled="submitting"
                    @change="handleFiles(item.id, 'tension', $event)"
                  >
                  <label :for="`tension-file-${item.id}`" class="upload-button">
                    <FileImage class="size-3.5" />
                    添加张力表照片
                  </label>
                  <span v-for="file in tensionDraft(item.id).files" :key="fileKey(file)" class="file-chip">
                    {{ file.name }}
                    <button
                      type="button"
                      class="remove-file"
                      :aria-label="`移除 ${file.name}`"
                      @click="removeFile(item.id, 'tension', file)"
                    >
                      <X class="size-3" />
                    </button>
                  </span>
                </div>
              </div>
            </section>

            <section class="space-y-3">
              <div>
                <h3 class="section-title">物流信息 / SHIPPING</h3>
                <p class="section-description">物流信息和上述证据保存完成后，系统才会原子写入发货状态。</p>
              </div>
              <div class="grid gap-3 sm:grid-cols-2">
                <label class="block space-y-1">
                  <span class="field-label">TRACKING / 物流单号</span>
                  <Input v-model="trackingNumber" placeholder="填写承运商单号" />
                </label>
                <label class="block space-y-1">
                  <span class="field-label">PROVIDER / 追踪 Provider</span>
                  <select v-model="trackingProviderId" class="field-select">
                    <option value="none">请选择追踪 Provider</option>
                    <option v-for="provider in trackingProviders" :key="provider.id" :value="String(provider.id)">
                      {{ provider.provider_name }} / {{ provider.provider_code }}
                    </option>
                  </select>
                </label>
              </div>
              <div class="grid gap-3 sm:grid-cols-2">
                <label class="block space-y-1">
                  <span class="field-label">CARRIER / 本地承运商</span>
                  <select v-model="carrierId" class="field-select">
                    <option value="none">不指定承运商</option>
                    <option v-for="carrier in carriers" :key="carrier.id" :value="String(carrier.id)">
                      {{ carrier.name }} / {{ carrier.code }}
                    </option>
                  </select>
                </label>
                <label class="block space-y-1">
                  <span class="field-label">SERVICE / 线路服务</span>
                  <select v-model="carrierServiceId" class="field-select">
                    <option value="none">不指定线路服务</option>
                    <option v-for="service in filteredCarrierServices" :key="service.id" :value="String(service.id)">
                      {{ service.service_name }} / {{ service.service_code }}
                    </option>
                  </select>
                </label>
              </div>
              <div class="rounded-xl border border-dashed border-border/80 bg-muted/30 p-3 text-xs text-muted-foreground">
                当前 Provider Carrier Code：
                <span class="font-mono font-bold text-foreground">{{ resolvedProviderCarrierCodeLabel }}</span>
              </div>
            </section>

            <label v-if="order?.signature_required" class="flex items-start gap-2 rounded-xl border border-amber-300/70 bg-amber-50/70 p-3 text-xs text-amber-900">
              <input v-model="signatureConfirmed" type="checkbox" class="mt-0.5 size-4 accent-amber-600">
              <span>我确认该订单达到高价签收要求，已选择需要签名的运输服务。</span>
            </label>
          </template>

          <div v-else class="rounded-xl border border-rose-200 bg-rose-50 p-3 text-xs text-rose-700">
            该订单没有可用的证据包，不能从此处直接发货。请先检查订单创建时的证据快照。
          </div>

          <p v-if="validationError" class="text-xs font-bold text-rose-600">{{ validationError }}</p>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="submitting" @click="emit('update:open', false)">
            取消
          </Button>
          <Button type="submit" :disabled="submitting || loading || !result?.package">
            <LoaderCircle v-if="submitting" class="size-3.5 animate-spin" />
            <Truck v-else class="size-3.5" />
            {{ submitting ? '正在保存取证并发货' : '保存取证并发货' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Camera, FileImage, LoaderCircle, PackageOpen, Truck, X } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { formatMoney } from '@/lib/orderPresentation'
import type {
  OrderEvidenceItem,
  OrderEvidencePackageResult,
  OrderFulfillmentSubmitInput,
} from '@/modules/order/orderEvidenceTypes'
import type {
  OrderID,
  OrderRecord,
  ShippingCarrier,
  ShippingCarrierService,
  TrackingCarrierMapping,
  TrackingProvider,
} from '@/modules/order/orderTypes'

type EvidenceDraftKind = 'identity' | 'outbound' | 'tension'

interface IdentityDraft {
  serial_number: string
  batch_number: string
  capture_note: string
  files: File[]
}

interface TensionDraft {
  assembly_reference: string
  measurement_tool: string
  capture_note: string
  files: File[]
}

interface OutboundDraft {
  gross_weight_g: string
  package_count: string
  packaging_method: string
  packaging_note: string
  files: File[]
}

const props = withDefaults(defineProps<{
  open?: boolean
  order?: OrderRecord | null
  result?: OrderEvidencePackageResult | null
  loading?: boolean
  submitting?: boolean
  trackingProviders?: TrackingProvider[]
  carriers?: ShippingCarrier[]
  carrierServices?: ShippingCarrierService[]
  trackingCarrierMappings?: TrackingCarrierMapping[]
}>(), {
  open: false,
  order: null,
  result: null,
  loading: false,
  submitting: false,
  trackingProviders: () => [],
  carriers: () => [],
  carrierServices: () => [],
  trackingCarrierMappings: () => [],
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit', value: OrderFulfillmentSubmitInput): void
}>()

const trackingNumber = ref('')
const trackingProviderId = ref('none')
const carrierId = ref('none')
const carrierServiceId = ref('none')
const signatureConfirmed = ref(false)
const validationError = ref('')
const identityDrafts = reactive<Record<string, IdentityDraft>>({})
const tensionDrafts = reactive<Record<string, TensionDraft>>({})
const outbound = reactive<OutboundDraft>({
  gross_weight_g: '',
  package_count: '',
  packaging_method: '',
  packaging_note: '',
  files: [],
})
const initializedKey = ref('')

const asRecord = (value: unknown): Record<string, unknown> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
  return value as Record<string, unknown>
}

const readText = (record: Record<string, unknown>, key: string): string => String(record[key] || '')

const itemKey = (itemID: OrderID | null | undefined): string => String(itemID || '')

const identityItem = (orderItemID: OrderID | null | undefined): OrderEvidenceItem | null => (
  props.result?.package.items?.find((item) =>
    item.item_type === 'product_identity' && String(item.order_item_id) === itemKey(orderItemID)
  ) || null
)

const tensionItems = computed(() => (
  props.result?.package.items?.filter((item) => item.item_type === 'spoke_qc_tension') || []
))

const outboundItem = computed(() => (
  props.result?.package.items?.find((item) => item.item_type === 'outbound_weight_packaging') || null
))

const selectedCarrierService = computed(() => {
  const serviceID = toID(carrierServiceId.value)
  if (!serviceID) return null
  return props.carrierServices.find((service) => Number(service.id) === serviceID) || null
})

const filteredCarrierServices = computed(() => {
  const carrierID = toID(carrierId.value) || Number(selectedCarrierService.value?.carrier_id || 0)
  if (!carrierID) return props.carrierServices
  return props.carrierServices.filter((service) => Number(service.carrier_id) === carrierID)
})

const resolvedProviderCarrierCodeLabel = computed(() => {
  const providerID = toID(trackingProviderId.value)
  if (!providerID) return '未匹配映射'

  const serviceID = toID(carrierServiceId.value)
  if (serviceID) {
    const serviceMapping = props.trackingCarrierMappings.find((mapping) =>
      Number(mapping.provider_id) === providerID &&
      mapping.scope === 'carrier_service' &&
      Number(mapping.carrier_service_id) === serviceID
    )
    if (serviceMapping) {
      return `${serviceMapping.provider_carrier_code || '未匹配'}${serviceMapping.provider_carrier_name ? ` / ${serviceMapping.provider_carrier_name}` : ''}`
    }
  }

  const carrierID = toID(carrierId.value) || Number(selectedCarrierService.value?.carrier_id || 0)
  if (!carrierID) return '未匹配映射'
  const carrierMapping = props.trackingCarrierMappings.find((mapping) =>
    Number(mapping.provider_id) === providerID &&
    mapping.scope === 'carrier' &&
    Number(mapping.carrier_id) === carrierID
  )
  if (!carrierMapping) return '未匹配映射'
  return `${carrierMapping.provider_carrier_code || '未匹配'}${carrierMapping.provider_carrier_name ? ` / ${carrierMapping.provider_carrier_name}` : ''}`
})

const identityDraft = (orderItemID: OrderID | null | undefined): IdentityDraft => {
  const key = itemKey(orderItemID)
  if (!identityDrafts[key]) {
    const record = asRecord(identityItem(orderItemID)?.data_json)
    identityDrafts[key] = {
      serial_number: readText(record, 'serial_number'),
      batch_number: readText(record, 'batch_number') || readText(record, 'production_lot'),
      capture_note: readText(record, 'capture_note') || readText(record, 'identifier_note'),
      files: [],
    }
  }
  return identityDrafts[key]
}

const tensionDraft = (evidenceItemID: OrderID): TensionDraft => {
  const key = itemKey(evidenceItemID)
  if (!tensionDrafts[key]) {
    const record = asRecord(props.result?.package.items?.find((item) => item.id === evidenceItemID)?.data_json)
    tensionDrafts[key] = {
      assembly_reference: readText(record, 'assembly_reference'),
      measurement_tool: readText(record, 'measurement_tool'),
      capture_note: readText(record, 'capture_note') || readText(record, 'conclusion'),
      files: [],
    }
  }
  return tensionDrafts[key]
}

const clearDrafts = (): void => {
  Object.keys(identityDrafts).forEach((key) => delete identityDrafts[key])
  Object.keys(tensionDrafts).forEach((key) => delete tensionDrafts[key])
  Object.assign(outbound, {
    gross_weight_g: '',
    package_count: '',
    packaging_method: '',
    packaging_note: '',
    files: [],
  })
}

const initialize = (): void => {
  const order = props.order
  const result = props.result
  if (!order?.id || !result?.package?.id) return

  clearDrafts()
  trackingNumber.value = String(order.tracking_number || '')
  trackingProviderId.value = order.tracking_provider_id ? String(order.tracking_provider_id) : 'none'
  carrierId.value = order.carrier_id ? String(order.carrier_id) : 'none'
  carrierServiceId.value = order.carrier_service_id ? String(order.carrier_service_id) : 'none'
  signatureConfirmed.value = false
  validationError.value = ''

  const outboundRecord = asRecord(outboundItem.value?.data_json)
  outbound.gross_weight_g = readText(outboundRecord, 'gross_weight_g')
  outbound.package_count = readText(outboundRecord, 'package_count')
  outbound.packaging_method = readText(outboundRecord, 'packaging_method')
  outbound.packaging_note = readText(outboundRecord, 'packaging_note')

  initializedKey.value = `${order.id}:${result.package.id}`
}

watch(
  () => [props.open, props.order?.id, props.result?.package?.id] as const,
  () => {
    const key = `${props.order?.id || ''}:${props.result?.package?.id || ''}`
    if (props.open && key !== initializedKey.value) initialize()
    if (!props.open) initializedKey.value = ''
  },
  { immediate: true },
)

const attachmentCount = (item: OrderEvidenceItem | null): number => Number(item?.attachments?.length || 0)

const hasFiles = (item: OrderEvidenceItem | null, files: File[]): boolean => attachmentCount(item) > 0 || files.length > 0

const fileKey = (file: File): string => `${file.name}:${file.size}:${file.lastModified}`

const handleFiles = (
  evidenceID: OrderID | null | undefined,
  kind: EvidenceDraftKind,
  event: Event,
): void => {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (kind === 'outbound') {
    outbound.files.push(...files)
    return
  }
  if (kind === 'identity') {
    identityDraft(evidenceID).files.push(...files)
    return
  }
  tensionDraft(evidenceID || '').files.push(...files)
}

const removeFile = (
  evidenceID: OrderID | null | undefined,
  kind: EvidenceDraftKind,
  file: File,
): void => {
  const target = kind === 'outbound'
    ? outbound.files
    : kind === 'identity'
      ? identityDraft(evidenceID).files
      : tensionDraft(evidenceID || '').files
  const index = target.findIndex((candidate) => fileKey(candidate) === fileKey(file))
  if (index >= 0) target.splice(index, 1)
}

const toPositiveInteger = (value: string): number | null => {
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null
}

const toID = (value: string): OrderID | null => {
  const parsed = Number(value)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null
}

const buildIdentityData = (draft: IdentityDraft): Record<string, unknown> => ({
  schema_version: 1,
  ...(draft.serial_number.trim() ? { serial_number: draft.serial_number.trim() } : {}),
  ...(draft.batch_number.trim() ? { batch_number: draft.batch_number.trim() } : {}),
  ...(draft.capture_note.trim() ? { capture_note: draft.capture_note.trim() } : {}),
})

const buildTensionData = (item: OrderEvidenceItem, draft: TensionDraft): Record<string, unknown> => ({
  schema_version: 1,
  order_item_id: Number(item.order_item_id),
  ...(draft.assembly_reference.trim() ? { assembly_reference: draft.assembly_reference.trim() } : {}),
  ...(draft.measurement_tool.trim() ? { measurement_tool: draft.measurement_tool.trim() } : {}),
  ...(draft.capture_note.trim() ? { capture_note: draft.capture_note.trim() } : {}),
})

const submit = (): void => {
  validationError.value = ''
  if (!props.order || !props.result?.package) return
  if (props.result.package.status === 'locked' || props.result.package.status === 'superseded') {
    validationError.value = '当前证据包已锁定或已修订，请先在履约证据页创建可编辑修订版。'
    return
  }

  if (!trackingNumber.value.trim()) {
    validationError.value = '请填写物流单号。'
    return
  }
  const provider = toID(trackingProviderId.value)
  if (!provider) {
    validationError.value = '请选择追踪 Provider。'
    return
  }
  const carrier = toID(carrierId.value)
  const carrierService = toID(carrierServiceId.value)
  if (!carrier && !carrierService) {
    validationError.value = '请选择本地承运商或线路服务。'
    return
  }
  if (props.order.signature_required && !signatureConfirmed.value) {
    validationError.value = '该订单需要签名运输确认。'
    return
  }

  const capturedAt = new Date().toISOString()
  const evidence: OrderFulfillmentSubmitInput['evidence'] = []
  for (const orderItem of props.order.items || []) {
    if (!orderItem.id) {
      validationError.value = '订单存在没有订单行 ID 的产品，不能安全发货。'
      return
    }
    const item = identityItem(orderItem.id)
    const draft = identityDraft(orderItem.id)
    if (!item || !hasFiles(item, draft.files)) {
      validationError.value = `订单行 #${orderItem.id} 缺少产品照片。`
      return
    }
    evidence.push({
      item_id: item.id,
      data_json: buildIdentityData(draft),
      captured_at: capturedAt,
      files: draft.files,
    })
  }

  if (!outboundItem.value) {
    validationError.value = '证据包缺少出库称重 / 包装项。'
    return
  }
  const grossWeight = toPositiveInteger(outbound.gross_weight_g)
  const packageCount = toPositiveInteger(outbound.package_count)
  if (!grossWeight || !packageCount || packageCount > 100 || !outbound.packaging_method.trim()) {
    validationError.value = '请完整填写出库毛重、包裹数量和包装方式。'
    return
  }
  if (!hasFiles(outboundItem.value, outbound.files)) {
    validationError.value = '请至少上传一张出库包装照片。'
    return
  }
  evidence.push({
    item_id: outboundItem.value.id,
    data_json: {
      schema_version: 1,
      gross_weight_g: grossWeight,
      package_count: packageCount,
      packaging_method: outbound.packaging_method.trim(),
      ...(outbound.packaging_note.trim() ? { packaging_note: outbound.packaging_note.trim() } : {}),
    },
    captured_at: capturedAt,
    files: outbound.files,
  })

  for (const item of tensionItems.value) {
    const draft = tensionDraft(item.id)
    if (!hasFiles(item, draft.files)) {
      validationError.value = `订单行 #${item.order_item_id} 缺少张力表照片。`
      return
    }
    evidence.push({
      item_id: item.id,
      data_json: buildTensionData(item, draft),
      captured_at: capturedAt,
      files: draft.files,
    })
  }

  emit('submit', {
    tracking_number: trackingNumber.value.trim(),
    tracking_provider_id: provider,
    carrier_id: carrier,
    carrier_service_id: carrierService,
    signature_confirmed: Boolean(signatureConfirmed.value),
    evidence,
  })
}
</script>

<style scoped>
.field-label {
  display: block;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: hsl(var(--muted-foreground) / 0.75);
}

.field-select {
  height: 2.25rem;
  width: 100%;
  border-radius: 0.75rem;
  border: 1px dashed hsl(var(--border));
  background: hsl(var(--background));
  padding: 0 0.75rem;
  font-size: 0.75rem;
  font-weight: 700;
}

.section-title {
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.08em;
}

.section-description {
  margin-top: 0.25rem;
  font-size: 0.6875rem;
  line-height: 1.25rem;
  color: hsl(var(--muted-foreground));
}

.summary-box {
  display: flex;
  min-height: 4.25rem;
  flex-direction: column;
  justify-content: center;
  gap: 0.25rem;
  border-radius: 0.75rem;
  border: 1px solid hsl(var(--border));
  padding: 0.75rem;
}

.summary-box strong {
  font-size: 0.875rem;
}

.summary-note {
  font-size: 0.6875rem;
  color: hsl(var(--muted-foreground));
}

.evidence-row {
  display: grid;
  gap: 0.75rem;
  border-radius: 0.75rem;
  border: 1px dashed hsl(var(--border));
  padding: 0.75rem;
}

.upload-button {
  display: inline-flex;
  min-height: 2rem;
  cursor: pointer;
  align-items: center;
  gap: 0.375rem;
  border-radius: 9999px;
  border: 1px solid hsl(var(--border));
  padding: 0.375rem 0.75rem;
  font-size: 0.6875rem;
  font-weight: 800;
}

.upload-button:hover {
  background: hsl(var(--muted) / 0.5);
}

.file-chip {
  display: inline-flex;
  max-width: 16rem;
  align-items: center;
  gap: 0.25rem;
  overflow: hidden;
  border-radius: 9999px;
  background: hsl(var(--muted));
  padding: 0.375rem 0.5rem 0.375rem 0.625rem;
  font-size: 0.6875rem;
}

.file-chip::first-line {
  overflow: hidden;
  text-overflow: ellipsis;
}

.remove-file {
  display: inline-flex;
  flex: 0 0 auto;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  padding: 0.125rem;
}

.remove-file:hover {
  background: hsl(var(--background));
}
</style>
