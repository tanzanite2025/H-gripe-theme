<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>发起售后</DialogTitle>
        <DialogDescription>{{ order?.order_number || '订单售后申请' }}</DialogDescription>
      </DialogHeader>

      <form class="space-y-5" @submit.prevent="submit">
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block space-y-1.5">
            <span class="field-label">TYPE / 售后类型</span>
            <select v-model="form.type" class="field-select" :disabled="submitting">
              <option v-for="option in typeOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="block space-y-1.5">
            <span class="field-label">REASON / 申请原因</span>
            <Input v-model="form.reason" maxlength="500" :disabled="submitting" />
          </label>
        </div>

        <label class="block space-y-1.5">
          <span class="field-label">DESCRIPTION / 补充说明</span>
          <Textarea v-model="form.description" rows="4" maxlength="2000" :disabled="submitting" />
        </label>

        <section class="space-y-3">
          <div>
            <h3 class="text-sm font-black uppercase text-foreground">申请商品</h3>
            <p class="mt-1 text-xs text-muted-foreground">选择本次售后涉及的订单商品及数量。</p>
          </div>

          <div class="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead class="w-12">
                    <span class="sr-only">选择</span>
                  </TableHead>
                  <TableHead>商品</TableHead>
                  <TableHead>SKU</TableHead>
                  <TableHead class="w-40 text-right">申请数量</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in eligibleItems" :key="itemKey(item)">
                  <TableCell>
                    <Checkbox
                      :model-value="isSelected(item)"
                      :disabled="submitting"
                      :aria-label="`选择 ${item.product_name || '商品'}`"
                      @update:model-value="toggleItem(item, $event)"
                    />
                  </TableCell>
                  <TableCell class="font-medium">{{ item.product_name || '-' }}</TableCell>
                  <TableCell class="font-mono text-xs text-muted-foreground">{{ item.sku || '-' }}</TableCell>
                  <TableCell>
                    <Input
                      :model-value="quantityFor(item)"
                      type="number"
                      min="1"
                      :max="maxQuantity(item)"
                      class="ml-auto w-24 text-right"
                      :disabled="submitting || !isSelected(item)"
                      :aria-label="`${item.product_name || '商品'}的申请数量`"
                      @update:model-value="setQuantity(item, $event)"
                    />
                    <p class="mt-1 text-right text-[10px] font-medium text-muted-foreground">最多 {{ maxQuantity(item) }} 件</p>
                  </TableCell>
                </TableRow>
                <TableEmpty v-if="eligibleItems.length === 0" :colspan="4">订单中没有可申请售后的商品</TableEmpty>
              </TableBody>
            </Table>
          </div>
        </section>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="submitting" @click="emit('update:open', false)">
            取消
          </Button>
          <Button type="submit" :disabled="!canSubmit || submitting">
            <RotateCcw :class="['size-4', submitting ? 'animate-spin' : '']" />
            {{ submitting ? '提交中' : '创建售后单' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { RotateCcw } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import type { CreateAfterSalesCaseInput } from '@/api/afterSales'
import type { OrderItem, OrderRecord } from './orderTypes'

const typeOptions: Array<{ value: CreateAfterSalesCaseInput['type']; label: string }> = [
  { value: 'return_refund', label: '退货退款' },
  { value: 'exchange', label: '换货' },
  { value: 'refund_only', label: '仅退款' },
  { value: 'reshipment', label: '补发' },
]

const props = withDefaults(defineProps<{
  open?: boolean
  order?: OrderRecord | null
  submitting?: boolean
}>(), {
  open: false,
  order: null,
  submitting: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit', input: CreateAfterSalesCaseInput): void
}>()

const form = reactive({
  type: 'return_refund' as CreateAfterSalesCaseInput['type'],
  reason: '',
  description: '',
  selected: {} as Record<string, boolean>,
  quantities: {} as Record<string, number>,
})

const itemKey = (item: OrderItem): string => String(item.id)
const maxQuantity = (item: OrderItem): number => {
  const quantity = Number(item.quantity || 0)
  return Number.isInteger(quantity) && quantity > 0 ? quantity : 0
}
const eligibleItems = computed<OrderItem[]>(() => (
  (props.order?.items || []).filter((item) => item.id != null && maxQuantity(item) > 0)
))
const isSelected = (item: OrderItem): boolean => Boolean(form.selected[itemKey(item)])
const quantityFor = (item: OrderItem): number => form.quantities[itemKey(item)] || 1
const selectedItems = computed(() => eligibleItems.value.filter(isSelected))
const canSubmit = computed(() => (
  form.reason.trim().length > 0 &&
  selectedItems.value.length > 0 &&
  selectedItems.value.every((item) => {
    const quantity = quantityFor(item)
    return Number.isInteger(quantity) && quantity > 0 && quantity <= maxQuantity(item)
  })
))

const resetForm = (): void => {
  form.type = 'return_refund'
  form.reason = ''
  form.description = ''
  form.selected = {}
  form.quantities = {}
  eligibleItems.value.forEach((item) => {
    form.quantities[itemKey(item)] = 1
  })
}

const toggleItem = (item: OrderItem, value: boolean | 'indeterminate'): void => {
  const key = itemKey(item)
  form.selected[key] = value === true
  if (!form.quantities[key]) form.quantities[key] = 1
}

const setQuantity = (item: OrderItem, value: string | number): void => {
  const parsed = Number(value)
  const quantity = Number.isFinite(parsed) ? Math.trunc(parsed) : 1
  form.quantities[itemKey(item)] = Math.min(Math.max(quantity, 1), maxQuantity(item))
}

const submit = (): void => {
  if (!canSubmit.value) return
  emit('submit', {
    type: form.type,
    reason: form.reason.trim(),
    ...(form.description.trim() ? { description: form.description.trim() } : {}),
    items: selectedItems.value.map((item) => ({
      order_item_id: item.id!,
      quantity: quantityFor(item),
    })),
  })
}

watch(
  () => [props.open, props.order?.id],
  ([open]) => {
    if (open) resetForm()
  },
)
</script>

<style scoped>
.field-label {
  display: block;
  color: hsl(var(--muted-foreground) / 0.7);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.field-select {
  height: 2.5rem;
  width: 100%;
  border: 1px dashed hsl(var(--border) / 0.8);
  border-radius: 0.375rem;
  background: hsl(var(--background));
  padding: 0 0.75rem;
  color: hsl(var(--foreground));
  font-size: 0.875rem;
}
</style>
