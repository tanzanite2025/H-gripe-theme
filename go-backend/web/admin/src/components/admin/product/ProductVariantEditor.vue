<template>
  <div class="space-y-3" :class="{ 'carbon-rim-sales-variants': dedicatedCarbonRim }">
    <Alert>
      <Info class="size-4" />
      <AlertTitle>价格与库存按 SKU 维护</AlertTitle>
      <AlertDescription>
        每个 SKU 独立维护重量、价格、运费模板和库存。附属品可配置自己的运费模板；未配置模板的 SKU 无法计算运费。
      </AlertDescription>
    </Alert>

    <div class="flex flex-wrap items-center justify-between gap-2 rounded-lg border bg-background px-3 py-2">
      <div class="text-xs leading-5 text-muted-foreground">
        主价格按商品基准币种录入；展示价由独立读模型刷新任务生成。
      </div>
    </div>

    <Table class="min-w-[1080px]" :class="{ 'carbon-rim-sales-variants-table': dedicatedCarbonRim }">
      <TableHeader>
        <TableRow>
          <TableHead class="w-16 text-center">默认</TableHead>
          <TableHead class="min-w-40">SKU</TableHead>
          <TableHead class="min-w-40">SKU 名称</TableHead>
          <TableHead class="w-32">价格</TableHead>
          <TableHead class="w-28">重量（克）</TableHead>
          <TableHead class="w-44">线路计费模板</TableHead>
          <TableHead class="w-24">库存</TableHead>
          <TableHead class="w-20 text-center">启用</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <template v-for="(variant, index) in variants" :key="variant.id || `variant-${index}`">
        <TableRow :class="{ 'carbon-rim-sales-fields-row': dedicatedCarbonRim }">
          <TableCell class="text-center">
            <input
              type="radio"
              :name="defaultRadioName"
              :checked="index === defaultIndex"
              class="size-4 accent-primary"
              :aria-label="`设为默认变体 ${index + 1}`"
              @change="emit('set-default', index)"
            >
          </TableCell>
            <TableCell>
              <Input v-model="variant.sku" placeholder="变体 SKU" />
            </TableCell>
            <TableCell><Input v-model="variant.title" placeholder="例如 700C 前轮" /></TableCell>
            <TableCell>
              <Input v-model="variant.price" inputmode="decimal" type="text" min="0" step="0.01" />
            </TableCell>
            <TableCell>
              <Input v-model.number="variant.weight_grams" type="number" min="0" step="1" placeholder="克" />
            </TableCell>
            <TableCell>
              <ProductShippingBillingTemplateSelectionDialog
                v-model="variant.shipping_template_ids"
                :shipping-templates="shippingTemplates"
                :selection-context="`SKU ${variant.sku || index + 1}`"
              />
            </TableCell>
            <TableCell>
              <Input v-model.number="variant.stock" type="number" min="0" step="1" />
            </TableCell>
            <TableCell class="text-center">
              <input
                type="checkbox"
                :checked="variant.is_active !== false"
                class="size-4 accent-primary"
                :aria-label="`启用变体 ${variant.sku || index + 1}`"
                @change="emit('set-active', index, checkboxValue($event))"
              >
            </TableCell>
            <TableCell class="text-right">
              <Tooltip>
                <TooltipTrigger as-child>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="text-destructive hover:text-destructive"
                    :disabled="variants.length <= 1 || (dedicatedCarbonRim && Boolean(variant.id))"
                    :aria-label="`删除变体 ${variant.sku || index + 1}`"
                    @click="emit('remove', index)"
                  >
                    <Trash2 class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>删除变体</TooltipContent>
              </Tooltip>
            </TableCell>
        </TableRow>
        <TableRow v-if="dedicatedCarbonRim && variant.carbon_rim_specification" class="carbon-rim-engineering-fields-row">
          <TableCell :colspan="9">
            <CarbonRimSalesVariantSpecificationFields :specification="variant.carbon_rim_specification" :persisted="Boolean(variant.id)" :locked="engineeringFactsLocked" />
          </TableCell>
        </TableRow>
        </template>
      </TableBody>
    </Table>

    <Button type="button" variant="outline" size="sm" @click="emit('add')">
      <Plus class="size-3.5" />
      添加变体
    </Button>
  </div>
</template>

<script setup lang="ts">
import ProductShippingBillingTemplateSelectionDialog from './ProductShippingBillingTemplateSelectionDialog.vue'
import CarbonRimSalesVariantSpecificationFields from './CarbonRimSalesVariantSpecificationFields.vue'
import { Info, Plus, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import type {
  ProductVariantForm,
  ShippingTemplateRecord,
} from '@/modules/product/productEditorTypes'

const defaultRadioName = `product-variant-default-${Math.random().toString(36).slice(2)}`

const props = withDefaults(defineProps<{
  dedicatedCarbonRim?: boolean
  engineeringFactsLocked?: boolean
  variants: ProductVariantForm[]
  defaultIndex?: number
  shippingTemplates?: ShippingTemplateRecord[]
}>(), {
  defaultIndex: 0,
  shippingTemplates: () => [],
})

const emit = defineEmits<{
  (event: 'add'): void
  (event: 'remove', index: number): void
  (event: 'set-default', index: number): void
  (event: 'set-active', index: number, active: boolean): void
}>()

const checkboxValue = (event: Event): boolean => Boolean((event.target as HTMLInputElement | null)?.checked)


</script>

<style scoped>
@media (max-width: 767px) {
  .carbon-rim-sales-variants :deep(table) { display: block; width: 100%; }
  .carbon-rim-sales-variants :deep(thead) { display: none; }
  .carbon-rim-sales-variants :deep(tbody) { display: block; }
  .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row) { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.5rem; padding-top: 1rem; }
  .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td) { display: block; min-width: 0; text-align: left; }
  .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td::before) { display: block; font-size: 0.75rem; margin-bottom: 0.375rem; }
  .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(1)::before) { content: '默认'; }
  .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(2)::before) { content: 'SKU'; }
   .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(3)::before) { content: 'SKU 名称'; }
   .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(4)::before) { content: '价格'; }
   .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(5)::before) { content: '重量（克）'; }
   .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(6)::before) { content: '线路计费模板'; }
   .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(7)::before) { content: '库存'; }
   .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(8)::before) { content: '启用'; }
   .carbon-rim-sales-variants :deep(.carbon-rim-sales-fields-row > td:nth-child(9)::before) { content: '操作'; }
  .carbon-rim-sales-variants :deep(.carbon-rim-engineering-fields-row),
  .carbon-rim-sales-variants :deep(.carbon-rim-engineering-fields-row > td) { display: block; width: 100%; }
}
</style>

