<template>
  <section class="space-y-3 rounded-2xl border border-dashed border-border/80 p-3">
    <div>
      <span class="field-label">MANUAL POD / 人工签收记录</span>
      <p class="mt-1 text-[11px] text-muted-foreground">
        这里登记人工证据对应的物流单号和送达信息；物流商事件只作为只读上下文，不会自动完成 POD。
      </p>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <label class="block space-y-1">
        <span class="field-label">TRACKING NUMBER / 物流单号</span>
        <Input
          :model-value="modelValue.tracking_number"
          placeholder="填写照片或签收单对应的物流单号"
          :disabled="disabled"
          @update:model-value="updateText('tracking_number', $event)"
        />
      </label>
      <label class="block space-y-1">
        <span class="field-label">DELIVERED AT / 人工记录送达时间</span>
        <Input
          type="datetime-local"
          :model-value="modelValue.delivered_at"
          :disabled="disabled"
          @update:model-value="updateText('delivered_at', $event)"
        />
      </label>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <label class="block space-y-1">
        <span class="field-label">RECIPIENT / 收件人（可选）</span>
        <Input
          :model-value="modelValue.recipient_name"
          placeholder="签收人姓名"
          :disabled="disabled"
          @update:model-value="updateText('recipient_name', $event)"
        />
      </label>
      <label class="block space-y-1">
        <span class="field-label">PROOF REFERENCE / 证明引用（可选）</span>
        <Input
          :model-value="modelValue.proof_reference"
          placeholder="例如纸质 POD 编号"
          :disabled="disabled"
          @update:model-value="updateText('proof_reference', $event)"
        />
      </label>
    </div>

    <label class="block space-y-1">
      <span class="field-label">DELIVERY NOTE / 签收备注（可选）</span>
      <Textarea
        :model-value="modelValue.delivery_note"
        class="min-h-20"
        placeholder="记录人工核对情况"
        :disabled="disabled"
        @update:model-value="updateText('delivery_note', $event)"
      />
    </label>
  </section>
</template>

<script setup lang="ts">
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type { OrderEvidenceSignedPODForm } from '@/modules/order/orderEvidenceStructuredTypes'

const props = withDefaults(defineProps<{
  modelValue: OrderEvidenceSignedPODForm
  disabled?: boolean
}>(), {
  disabled: false,
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: OrderEvidenceSignedPODForm): void
}>()

const updateText = (
  key: keyof OrderEvidenceSignedPODForm,
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
