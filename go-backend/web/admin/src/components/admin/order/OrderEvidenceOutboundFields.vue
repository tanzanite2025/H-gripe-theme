<template>
  <section class="space-y-3 rounded-2xl border border-dashed border-border/80 p-3">
    <div>
      <span class="field-label">OUTBOUND RECORD / 出库记录</span>
      <p class="mt-1 text-[11px] text-muted-foreground">
        记录实际完成包装后的总毛重；照片或扫描件请通过下方附件上传。
      </p>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <label class="block space-y-1">
        <span class="field-label">GROSS WEIGHT / 出库毛重（克）</span>
        <Input
          type="number"
          min="1"
          step="1"
          inputmode="numeric"
          :model-value="displayNumber(modelValue.gross_weight_g)"
          placeholder="例如 12640"
          :disabled="disabled"
          @update:model-value="updateNumber('gross_weight_g', $event)"
        />
      </label>
      <label class="block space-y-1">
        <span class="field-label">PACKAGES / 包裹数量</span>
        <Input
          type="number"
          min="1"
          max="100"
          step="1"
          inputmode="numeric"
          :model-value="displayNumber(modelValue.package_count)"
          placeholder="例如 2"
          :disabled="disabled"
          @update:model-value="updateNumber('package_count', $event)"
        />
      </label>
    </div>

    <label class="block space-y-1">
      <span class="field-label">PACKAGING METHOD / 包装方式</span>
      <Input
        :model-value="modelValue.packaging_method"
        placeholder="例如双层瓦楞纸箱、木箱"
        :disabled="disabled"
        @update:model-value="updateText('packaging_method', $event)"
      />
    </label>

    <label class="block space-y-1">
      <span class="field-label">PACKAGING NOTE / 包装备注（可选）</span>
      <Textarea
        :model-value="modelValue.packaging_note"
        class="min-h-20"
        placeholder="记录防护、加固或分包情况"
        :disabled="disabled"
        @update:model-value="updateText('packaging_note', $event)"
      />
    </label>
  </section>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type { OrderEvidenceOutboundWeightPackagingForm } from '@/modules/order/orderEvidenceStructuredTypes'

const props = withDefaults(defineProps<{
  modelValue: OrderEvidenceOutboundWeightPackagingForm
  disabled?: boolean
}>(), {
  disabled: false,
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: OrderEvidenceOutboundWeightPackagingForm): void
}>()

const displayNumber = (value: number | null): string => (
  value === null || value === undefined ? '' : String(value)
)

const updateNumber = (
  key: 'gross_weight_g' | 'package_count',
  value: string | number,
): void => {
  const parsed = Number(String(value).trim())
  emit('update:modelValue', {
    ...props.modelValue,
    [key]: Number.isFinite(parsed) && parsed > 0 ? Math.trunc(parsed) : null,
  })
}

const updateText = (
  key: 'packaging_method' | 'packaging_note',
  value: string | number,
): void => {
  emit('update:modelValue', {
    ...props.modelValue,
    [key]: String(value),
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
</style>
