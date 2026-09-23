<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent size="lg">
      <DialogHeader>
        <DialogTitle>{{ itemTypeName }}</DialogTitle>
        <DialogDescription>
          {{ scopeLabel }} · {{ requiredReasonName }}
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4">
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <label class="block space-y-1">
            <span class="field-label">STATUS / 状态</span>
            <select
              v-model="status"
              class="field-select"
              :disabled="!canEdit || saving"
            >
              <option value="missing">缺失</option>
              <option value="draft">草稿</option>
              <option value="complete">已完成</option>
              <option value="waived">已豁免</option>
            </select>
          </label>
          <label class="block space-y-1">
            <span class="field-label">CAPTURED AT / 采集时间</span>
            <Input
              v-model="capturedAtInput"
              type="datetime-local"
              class="h-9"
              :disabled="!canEdit || saving || status === 'missing'"
            />
          </label>
        </div>

        <label class="block space-y-1">
          <span v-if="!isStructuredEvidence" class="field-label">DATA JSON / 人工备注（可选）</span>
          <Textarea
            v-if="!isStructuredEvidence"
            v-model="dataText"
            class="min-h-64 font-mono text-xs leading-5"
            :placeholder="dataPlaceholder"
            :disabled="!canEdit || saving"
          />
        </label>

        <OrderEvidenceOutboundFields
          v-if="isOutboundEvidence"
          v-model="outboundRecord"
          :disabled="!canEdit || saving"
        />
        <OrderEvidencePODFields
          v-else-if="isPODEvidence"
          v-model="podRecord"
          :disabled="!canEdit || saving"
        />

        <section class="rounded-2xl border border-dashed border-border/80 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <span class="field-label">ATTACHMENTS / 附件</span>
              <p class="mt-1 text-[11px] text-muted-foreground">{{ isPODEvidence ? '上传承运商官方签字 POD PDF，提交 PayPal 争议时会作为 documents 附件发送。' : '唯一编码照片、张力表扫描件和其他履约凭据保存在独立附件记录中。' }}</p>
            </div>
            <input
              ref="fileInput"
              type="file"
              :accept="isPODEvidence ? 'application/pdf,image/jpeg,image/png,image/webp,image/gif' : 'image/jpeg,image/png,image/webp,image/gif'"
              class="hidden"
              :disabled="!canEdit || saving || uploadingAttachment || readOnly"
              @change="handleAttachmentChange"
            >
            <Button
              type="button"
              variant="outline"
              size="sm"
              class="rounded-full"
              :disabled="!canEdit || saving || uploadingAttachment || readOnly"
              @click="openFilePicker"
            >
              <LoaderCircle v-if="uploadingAttachment" class="size-3.5 animate-spin" />
              <Paperclip v-else class="size-3.5" />
              {{ uploadingAttachment ? '上传中' : '添加附件' }}
            </Button>
          </div>
          <div v-if="item?.attachments?.length" class="mt-3 space-y-1.5">
            <div
              v-for="attachment in item.attachments"
              :key="String(attachment.id)"
              class="flex items-center gap-3 rounded-xl border px-3 py-2 text-xs hover:bg-muted/40"
            >
              <a
                :href="attachmentUrl(attachment.id)"
                target="_blank"
                rel="noopener noreferrer"
                class="flex min-w-0 flex-1 items-center gap-2"
              >
                <FileImage class="size-3.5 shrink-0 text-primary" />
                <span class="truncate">{{ attachment.original_filename || `附件 #${attachment.id}` }}</span>
                <span class="ml-auto shrink-0 text-[10px] font-mono text-muted-foreground/70">
                  {{ formatBytes(attachment.size_bytes) }}
                </span>
              </a>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                class="size-7 shrink-0 text-destructive"
                :aria-label="`移除附件 ${attachment.original_filename || attachment.id}`"
                title="移除错误附件"
                :disabled="!canEdit || saving || uploadingAttachment || deletingAttachmentID === attachment.id || readOnly"
                @click.prevent.stop="$emit('delete-attachment', attachment.id)"
              >
                <LoaderCircle v-if="deletingAttachmentID === attachment.id" class="size-3.5 animate-spin" />
                <Trash2 v-else class="size-3.5" />
              </Button>
            </div>
          </div>
          <p v-else class="mt-3 text-xs text-muted-foreground">暂无附件。</p>
        </section>

        <p v-if="isSpokeTensionEvidence" class="text-xs leading-5 text-amber-700">
          张力表请在发货时上传照片或扫描件。系统只保存人工提交的附件和备注，不判断张力是否合格，也不会根据数值自动判定。
        </p>
        <p class="text-xs leading-5 text-muted-foreground">
          照片或扫描件类证据标记为完成前必须先上传附件；配置确认单来自下单快照，不能在此修改。
        </p>
        <p v-if="validationError" class="text-xs font-bold text-rose-600">{{ validationError }}</p>
      </div>

      <DialogFooter>
        <Button variant="outline" :disabled="saving" @click="$emit('update:open', false)">
          取消
        </Button>
        <Button :disabled="!canEdit || saving || readOnly" @click="submit">
          <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          保存证据项
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { FileImage, LoaderCircle, Paperclip, Save } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import OrderEvidenceOutboundFields from '@/components/admin/order/OrderEvidenceOutboundFields.vue'
import OrderEvidencePODFields from '@/components/admin/order/OrderEvidencePODFields.vue'
import {
  orderEvidenceItemIsReadOnly,
  orderEvidenceItemScopeLabel,
  orderEvidenceItemTypeName,
  orderEvidenceRequiredReasonName,
} from '@/lib/orderEvidencePresentation'
import type { OrderEvidenceItem, OrderEvidenceItemStatus } from '@/modules/order/orderEvidenceTypes'
import type { OrderID } from '@/modules/order/orderTypes'
import type {
  OrderEvidenceOutboundWeightPackagingForm,
  OrderEvidenceSignedPODForm,
} from '@/modules/order/orderEvidenceStructuredTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  item?: OrderEvidenceItem | null
  canEdit?: boolean
  saving?: boolean
  uploadingAttachment?: boolean
  deletingAttachmentID?: OrderID | null
  attachmentUrl?: (attachmentID: OrderID) => string
}>(), {
  open: false,
  item: null,
  canEdit: false,
  saving: false,
  uploadingAttachment: false,
  deletingAttachmentID: null,
  attachmentUrl: () => '#',
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'upload-attachment', file: File): void
  (event: 'delete-attachment', attachmentID: OrderID): void
  (event: 'submit', payload: {
    status: OrderEvidenceItemStatus
    data_json: unknown
    captured_at: string | null
  }): void
}>()

const status = ref<OrderEvidenceItemStatus>('draft')
const dataText = ref('{}')
const capturedAtInput = ref('')
const validationError = ref('')

const emptyOutboundRecord = (): OrderEvidenceOutboundWeightPackagingForm => ({
  gross_weight_g: null,
  package_count: null,
  packaging_method: '',
  packaging_note: '',
})

const emptyPODRecord = (): OrderEvidenceSignedPODForm => ({
  tracking_number: '',
  delivered_at: '',
  recipient_name: '',
  proof_reference: '',
  delivery_note: '',
})

const asRecord = (value: unknown): Record<string, unknown> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
  return value as Record<string, unknown>
}

const readNumber = (value: unknown): number | null => {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? Math.trunc(parsed) : null
}

const toLocalDateTime = (value?: string | null): string => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const offset = date.getTimezoneOffset()
  return new Date(date.getTime() - offset * 60 * 1000).toISOString().slice(0, 16)
}

const readOutboundRecord = (value: unknown): OrderEvidenceOutboundWeightPackagingForm => {
  const record = asRecord(value)
  return {
    gross_weight_g: readNumber(record.gross_weight_g),
    package_count: readNumber(record.package_count),
    packaging_method: String(record.packaging_method || ''),
    packaging_note: String(record.packaging_note || ''),
  }
}

const readPODRecord = (value: unknown): OrderEvidenceSignedPODForm => {
  const record = asRecord(value)
  return {
    tracking_number: String(record.tracking_number || ''),
    delivered_at: toLocalDateTime(String(record.delivered_at || '')),
    recipient_name: String(record.recipient_name || ''),
    proof_reference: String(record.proof_reference || ''),
    delivery_note: String(record.delivery_note || ''),
  }
}

const toISOStringOrNull = (value: string): string | null => {
  const normalized = value.trim()
  if (!normalized) return null
  const date = new Date(normalized)
  if (Number.isNaN(date.getTime())) return null
  return date.toISOString()
}

const outboundRecord = ref<OrderEvidenceOutboundWeightPackagingForm>(emptyOutboundRecord())
const podRecord = ref<OrderEvidenceSignedPODForm>(emptyPODRecord())

const itemTypeName = computed(() => orderEvidenceItemTypeName(props.item?.item_type))
const requiredReasonName = computed(() => orderEvidenceRequiredReasonName(props.item?.required_reason))
const scopeLabel = computed(() => orderEvidenceItemScopeLabel(props.item))
const readOnly = computed(() => orderEvidenceItemIsReadOnly(props.item))
const isSpokeTensionEvidence = computed(() => props.item?.item_type === 'spoke_qc_tension')
const isOutboundEvidence = computed(() => props.item?.item_type === 'outbound_weight_packaging')
const isPODEvidence = computed(() => props.item?.item_type === 'signed_pod')
const isStructuredEvidence = computed(() => isOutboundEvidence.value || isPODEvidence.value)
const dataPlaceholder = computed(() => {
  switch (props.item?.item_type) {
    case 'product_identity':
      return '{\n  "serial_number": "",\n  "batch_number": "",\n  "production_lot": ""\n}'
    case 'spoke_qc_tension':
      return '{\n  "note": ""\n}'
    case 'outbound_weight_packaging':
      return '{\n  "weight_g": 0,\n  "package_count": 1,\n  "packaging_note": ""\n}'
    case 'signed_pod':
      return '{\n  "carrier": "",\n  "tracking_number": "",\n  "delivered_at": "",\n  "proof_reference": ""\n}'
    default:
      return '{}'
  }
})
const fileInput = ref<HTMLInputElement | null>(null)

const stringifyJSON = (value: unknown): string => {
  if (typeof value === 'string') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2)
    } catch {
      return value
    }
  }
  return JSON.stringify(value ?? {}, null, 2)
}

watch(
  () => [props.open, props.item] as const,
  ([open, item]) => {
    if (!open || !item) return
    status.value = item.status || 'draft'
    dataText.value = stringifyJSON(item.data_json)
    capturedAtInput.value = toLocalDateTime(item.captured_at)
    outboundRecord.value = readOutboundRecord(item.data_json)
    podRecord.value = readPODRecord(item.data_json)
    validationError.value = ''
  },
  { immediate: true },
)

const submit = (): void => {
  if (!props.item || readOnly.value) return
  validationError.value = ''

  const capturedAt = toISOStringOrNull(capturedAtInput.value)
  if (capturedAtInput.value.trim() && !capturedAt) {
    validationError.value = '采集时间格式无效。'
    return
  }

  let data: unknown
  if (isOutboundEvidence.value) {
    data = {
      schema_version: 1,
      gross_weight_g: outboundRecord.value.gross_weight_g,
      package_count: outboundRecord.value.package_count,
      packaging_method: outboundRecord.value.packaging_method.trim(),
      ...(outboundRecord.value.packaging_note.trim()
        ? { packaging_note: outboundRecord.value.packaging_note.trim() }
        : {}),
    }
  } else if (isPODEvidence.value) {
    const deliveredAt = toISOStringOrNull(podRecord.value.delivered_at)
    if (podRecord.value.delivered_at.trim() && !deliveredAt) {
      validationError.value = '人工送达时间格式无效。'
      return
    }
    data = {
      schema_version: 1,
      tracking_number: podRecord.value.tracking_number.trim(),
      delivered_at: deliveredAt || '',
      ...(podRecord.value.recipient_name.trim()
        ? { recipient_name: podRecord.value.recipient_name.trim() }
        : {}),
      ...(podRecord.value.proof_reference.trim()
        ? { proof_reference: podRecord.value.proof_reference.trim() }
        : {}),
      ...(podRecord.value.delivery_note.trim()
        ? { delivery_note: podRecord.value.delivery_note.trim() }
        : {}),
    }
  } else {
    try {
      data = JSON.parse(dataText.value || '{}')
    } catch {
      validationError.value = '结构化记录必须是合法 JSON。'
      return
    }
  }

  if (status.value === 'complete' && !capturedAtInput.value) {
    validationError.value = '完成状态必须填写采集时间。'
    return
  }

  if (status.value === 'complete' && (data === null || typeof data !== 'object')) {
    validationError.value = '完成状态的结构化记录必须是 JSON 对象或数组。'
    return
  }

  emit('submit', {
    status: status.value,
    data_json: data,
    captured_at: capturedAt,
  })
}

const openFilePicker = (): void => {
  fileInput.value?.click()
}

const handleAttachmentChange = (event: Event): void => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (file) emit('upload-attachment', file)
}

const attachmentUrl = (attachmentID: OrderID): string => props.attachmentUrl(attachmentID)

const formatBytes = (value?: number | null): string => {
  const size = Number(value || 0)
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
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
  border: 1px dashed hsl(var(--border));
  border-radius: 0.75rem;
  background: hsl(var(--background));
  padding: 0 0.75rem;
  font-size: 0.875rem;
}
</style>
