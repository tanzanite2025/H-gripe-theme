<template>
  <div class="space-y-3">
    <Alert>
      <Info class="size-4" />
      <AlertTitle>价格与库存按 SKU 维护</AlertTitle>
      <AlertDescription>
        每个 SKU 都可以单独维护重量、价格、运费模板和库存；主基准币种继承商品设置。
      </AlertDescription>
    </Alert>

    <div class="mb-3 rounded-lg border border-dashed bg-muted/20 px-3 py-2 text-xs leading-5 text-muted-foreground">
      <span v-if="specDefinitions.length">
          SKU 选项列来自已绑定商品规格模板；每一行仍单独维护价格、重量、运费模板和库存，主基准币种继承商品。
      </span>
      <span v-else>
        当前没有模板 SKU 选项字段；先维护默认 SKU、价格、重量、运费模板和库存即可，主基准币种继承商品。
      </span>
    </div>

    <div v-if="presentableSpecDefinitions.length || customOptionDefinitions.length" class="space-y-3 rounded-lg border bg-muted/10 p-3">
      <div class="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h3 class="text-sm font-semibold text-foreground">选项展示值</h3>
          <p class="mt-1 text-xs leading-5 text-muted-foreground">
            颜色和图片选项使用稳定值标识参与 SKU 匹配；色板颜色或上传图片只负责前台展示。
          </p>
        </div>
      </div>

      <div v-for="spec in presentableSpecDefinitions" :key="`option-meta-${spec.id}`" class="space-y-2 rounded-lg border bg-background/70 p-3">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <span class="text-xs font-bold">{{ specLabel(spec) }}</span>
            <span class="rounded-full bg-muted px-2 py-0.5 text-[10px] text-muted-foreground">
              {{ spec.presentation === 'image' ? '图片展示' : '色板展示' }}
            </span>
          </div>
          <Button type="button" variant="outline" size="sm" @click="addOptionValue(spec)">
            <Plus class="size-3.5" />
            添加选项
          </Button>
        </div>

        <template v-if="optionsForSpec(spec).length">
          <div class="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
            <div v-for="option in optionsForSpec(spec)" :key="option.local_key || option.id || `${spec.id}-${option.value_key}`" class="min-w-0 rounded-lg border bg-background p-2.5">
              <div class="mb-2 flex items-center justify-between gap-2">
                <div
                  class="size-9 shrink-0 overflow-hidden rounded-md border bg-muted"
                  :style="option.swatch_url ? undefined : (option.color_hex ? { backgroundColor: option.color_hex } : undefined)"
                >
                  <img v-if="option.swatch_url" :src="option.swatch_url" alt="" class="h-full w-full object-cover" />
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="size-8 text-destructive hover:text-destructive"
                  :aria-label="`删除${spec.name}选项`"
                  @click="removeOptionValue(option)"
                >
                  <Trash2 class="size-4" />
                </Button>
              </div>

              <div class="space-y-2">
                <Input v-model="option.value_key" class="font-mono text-xs" placeholder="稳定值，如 carbon-red" />
                <Input v-model="option.label" class="text-xs" :placeholder="`${spec.name}显示名称`" />
                <Input
                  v-if="spec.presentation === 'color'"
                  v-model="option.color_hex"
                  class="font-mono text-xs uppercase"
                  placeholder="#8F2028"
                />
                <label class="flex cursor-pointer items-center justify-center gap-1.5 rounded-md border border-dashed px-2 py-1.5 text-xs text-muted-foreground transition hover:border-primary/50 hover:text-foreground">
                  <input class="sr-only" type="file" :accept="uploadSpecAccept('product_variant_swatch')" :disabled="swatchUploadingKey === optionKey(option)" @change="uploadSwatch($event, option)" />
                  <LoaderCircle v-if="swatchUploadingKey === optionKey(option)" class="size-3.5 animate-spin" />
                  <ImageUp v-else class="size-3.5" />
                  {{ option.swatch_url ? '更换展示图片' : '上传展示图片' }}
                </label>
                <button
                  v-if="option.swatch_url"
                  type="button"
                  class="text-[11px] text-muted-foreground underline underline-offset-2"
                  @click="clearSwatch(option)"
                >
                  清除图片，使用色板颜色
                </button>
              </div>
            </div>
          </div>
        </template>
        <template v-else>
          <p class="rounded-md border border-dashed px-3 py-3 text-center text-xs text-muted-foreground">
          还没有配置展示选项；先添加颜色或图片选项，再在下方 SKU 表中选择。
          </p>
        </template>
        <UploadSpecHint code="product_variant_swatch" />
      </div>

      <div v-for="spec in customOptionDefinitions" :key="`custom-option-${spec.id}`" class="space-y-2 rounded-lg border bg-background/70 p-3">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <span class="text-xs font-bold">{{ specLabel(spec) }}</span>
            <span class="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] text-primary">{{ spec.selection_mode === 'multiple' ? '多选' : '单选' }}</span>
          </div>
          <Button type="button" variant="outline" size="sm" @click="addOptionValue(spec)">
            <Plus class="size-3.5" />
            添加选配
          </Button>
        </div>
        <div v-if="optionsForSpec(spec).length" class="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
          <div v-for="option in optionsForSpec(spec)" :key="option.local_key || option.id || `${spec.id}-${option.value_key}`" class="min-w-0 rounded-lg border bg-background p-2.5">
            <div class="mb-2 flex items-center justify-between gap-2">
              <span class="text-[10px] text-muted-foreground">价格增量（最小货币单位）</span>
              <Button type="button" variant="ghost" size="icon" class="size-8 text-destructive hover:text-destructive" :aria-label="`删除${spec.name}选配`" @click="removeOptionValue(option)"><Trash2 class="size-4" /></Button>
            </div>
            <div class="space-y-2">
              <Input v-model="option.value_key" class="font-mono text-xs" placeholder="稳定值，如 xdr" />
              <Input v-model="option.label" class="text-xs" :placeholder="`${spec.name}显示名称`" />
              <Input v-model.number="option.price_delta_minor" type="number" min="0" step="1" placeholder="0" />
              <div class="grid grid-cols-2 gap-2">
                <Input v-model.number="option.weight_delta_grams" type="number" min="0" step="1" placeholder="增重（克）" />
                <Input v-model.number="option.packaging_weight_delta_grams" type="number" min="0" step="1" placeholder="包装增重（克）" />
              </div>
              <Input v-model.number="option.production_lead_time_days" type="number" min="0" step="1" placeholder="生产周期（天）" />
              <label class="flex items-center justify-between gap-2 text-xs"><span>需要生产</span><Switch v-model="option.requires_production" :aria-label="`${option.label || option.value_key || spec.name}需要生产`" /></label>
              <Select v-model="option.cancellation_policy">
                <SelectTrigger class="w-full text-xs"><SelectValue placeholder="取消策略" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="standard">标准取消</SelectItem>
                  <SelectItem value="before_production">生产前可取消</SelectItem>
                  <SelectItem value="never">不可取消</SelectItem>
                </SelectContent>
              </Select>
              <Select v-model="option.return_policy">
                <SelectTrigger class="w-full text-xs"><SelectValue placeholder="退货策略" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="standard">标准退货</SelectItem>
                  <SelectItem value="not_allowed">不可退货/换货</SelectItem>
                </SelectContent>
              </Select>
              <Select
                :model-value="option.inventory_policy || 'none'"
                @update:model-value="value => setInventoryPolicy(option, String(value || 'none'))"
              >
                <SelectTrigger class="w-full text-xs">
                  <SelectValue placeholder="库存策略" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">无独立库存</SelectItem>
                  <SelectItem value="component">绑定组件库存</SelectItem>
                </SelectContent>
              </Select>
              <template v-if="option.inventory_policy === 'component'">
                <Input v-model.number="option.component_variant_id" type="number" min="1" step="1" placeholder="组件变体 ID" />
                <Input v-model.number="option.component_quantity" type="number" min="1" step="1" placeholder="每件所需数量" />
              </template>
              <label class="flex items-center justify-between gap-2 text-xs"><span>默认选中</span><Switch v-model="option.is_default" :aria-label="`${option.label || option.value_key || spec.name}默认选中`" /></label>
              <label class="flex items-center justify-between gap-2 text-xs"><span>启用</span><Switch v-model="option.is_enabled" :aria-label="`${option.label || option.value_key || spec.name}启用`" /></label>
            </div>
          </div>
        </div>
        <p v-else class="rounded-md border border-dashed px-3 py-3 text-center text-xs text-muted-foreground">还没有配置买家选配值。</p>
      </div>
    </div>

    <div class="space-y-3 rounded-lg border bg-muted/10 p-3">
      <div class="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h3 class="text-sm font-semibold text-foreground">选项依赖关系</h3>
          <p class="mt-1 text-xs leading-5 text-muted-foreground">
            “必须同时选择”用于组合要求，“不能同时选择”用于互斥选项。
          </p>
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          :disabled="!canAddOptionValueRelation"
          @click="addOptionValueRelation"
        >
          <Plus class="size-3.5" />
          添加关系
        </Button>
      </div>

      <p v-if="!canConfigureOptionValueRelations" class="rounded-md border border-dashed px-3 py-3 text-center text-xs text-muted-foreground">
        保存商品后，才能使用已生成 ID 的选项值配置依赖关系。
      </p>
      <p v-else-if="persistedOptionValues.length < 2" class="rounded-md border border-dashed px-3 py-3 text-center text-xs text-muted-foreground">
        至少需要两个已保存的选项值才能配置依赖关系。
      </p>
      <template v-else>
        <div v-if="optionValueRelations.length" class="divide-y rounded-md border bg-background">
          <div
            v-for="(relation, index) in optionValueRelations"
            :key="relation.id || `option-relation-${index}`"
            class="grid gap-2 p-2.5 lg:grid-cols-[minmax(0,1fr)_11rem_minmax(0,1fr)_2.25rem] lg:items-center"
          >
            <Select
              :model-value="relationValue(relation.source_option_value_id)"
              @update:model-value="value => setRelationOptionValue(relation, 'source_option_value_id', value)"
            >
              <SelectTrigger class="w-full text-xs"><SelectValue placeholder="来源选项" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in persistedOptionValues" :key="`source-${option.id}`" :value="String(option.id)">
                  {{ relationOptionLabel(option) }}
                </SelectItem>
              </SelectContent>
            </Select>
            <Select
              :model-value="relation.relation_type"
              @update:model-value="value => setRelationType(relation, value)"
            >
              <SelectTrigger class="w-full text-xs"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="requires">必须同时选择</SelectItem>
                <SelectItem value="conflicts">不能同时选择</SelectItem>
              </SelectContent>
            </Select>
            <Select
              :model-value="relationValue(relation.target_option_value_id)"
              @update:model-value="value => setRelationOptionValue(relation, 'target_option_value_id', value)"
            >
              <SelectTrigger class="w-full text-xs"><SelectValue placeholder="目标选项" /></SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="option in persistedOptionValues"
                  :key="`target-${option.id}`"
                  :value="String(option.id)"
                  :disabled="Number(option.id) === Number(relation.source_option_value_id)"
                >
                  {{ relationOptionLabel(option) }}
                </SelectItem>
              </SelectContent>
            </Select>
            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="size-9 text-destructive hover:text-destructive"
                  :aria-label="`删除选项关系 ${index + 1}`"
                  @click="removeOptionValueRelation(index)"
                >
                  <Trash2 class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>删除关系</TooltipContent>
            </Tooltip>
          </div>
        </div>
        <p v-else class="rounded-md border border-dashed px-3 py-3 text-center text-xs text-muted-foreground">
          尚未配置选项依赖关系。
        </p>
        <p v-if="hasUnsavedOptionValues" class="text-[11px] leading-5 text-muted-foreground">
          本次新增的选项值需要先保存商品，之后才能加入依赖关系。
        </p>
      </template>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-2 rounded-lg border bg-background px-3 py-2">
      <div class="text-xs leading-5 text-muted-foreground">
        主价格按商品基准币种录入；展示价由独立读模型刷新任务生成。
      </div>
    </div>

    <Table class="min-w-[1380px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-16 text-center">默认</TableHead>
          <TableHead class="min-w-40">SKU</TableHead>
          <TableHead v-for="spec in specDefinitions" :key="spec.id" class="min-w-36">
            {{ specLabel(spec) }}
          </TableHead>
          <TableHead class="w-32">价格</TableHead>
          <TableHead class="w-32">促销价</TableHead>
          <TableHead class="w-28">重量（克）</TableHead>
          <TableHead class="w-44">运费模板</TableHead>
          <TableHead class="w-24">库存</TableHead>
          <TableHead class="w-20 text-center">启用</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <template v-for="(variant, index) in variants" :key="variant.id || `variant-${index}`">
        <TableRow>
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
            <TableCell v-for="spec in specDefinitions" :key="spec.id">
              <Input
                v-if="spec.field_type === 'number'"
                v-model.number="variant.option_values[spec.slug]"
                type="number"
                min="0"
              />
              <Select
                v-else-if="spec.field_type === 'select' && specOptions(spec).length"
                :model-value="selectValue(variant.option_values[spec.slug])"
                @update:model-value="setSelectValue(variant, spec.slug, $event)"
              >
                <SelectTrigger class="w-full">
                  <SelectValue placeholder="请选择" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__empty__">未设置</SelectItem>
                  <SelectItem v-for="option in specOptions(spec)" :key="String(option)" :value="String(option)">
                    {{ optionLabel(spec, option) }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <Input
                v-else-if="spec.field_type === 'select'"
                v-model="variant.option_values[spec.slug]"
                :placeholder="`${spec.name}（可动态录入）`"
              />
              <div v-else-if="spec.field_type === 'boolean'" class="flex h-8 items-center">
                <Switch v-model="variant.option_values[spec.slug]" :aria-label="spec.name" />
              </div>
              <Input v-else v-model="variant.option_values[spec.slug]" :placeholder="spec.name" />
            </TableCell>
            <TableCell>
              <Input v-model="variant.price" inputmode="decimal" type="text" min="0" step="0.01" />
            </TableCell>
            <TableCell>
              <Input v-model="variant.sale_price" inputmode="decimal" type="text" min="0" step="0.01" placeholder="可选" />
            </TableCell>
            <TableCell>
              <Input v-model.number="variant.weight_grams" type="number" min="0" step="1" placeholder="克" />
            </TableCell>
            <TableCell>
              <Select
                :model-value="shippingTemplateSelectValue(variant.shipping_template_id)"
                @update:model-value="setShippingTemplateValue(variant, $event)"
              >
                <SelectTrigger class="w-full">
                  <SelectValue placeholder="继承商品" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__inherit__">继承商品</SelectItem>
                  <SelectItem v-for="template in shippingTemplates" :key="template.id" :value="String(template.id)">
                    {{ template.name }}{{ template.enabled === false ? '（停用）' : '' }}
                  </SelectItem>
                </SelectContent>
              </Select>
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
                    :disabled="variants.length <= 1"
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
        <TableRow v-if="customOptionDefinitions.length" :key="`${variant.id || `variant-${index}`}-custom-rules`">
          <TableCell :colspan="specDefinitions.length + 10" class="bg-muted/10 p-3">
            <div class="grid gap-2 lg:grid-cols-2">
              <div v-for="spec in customOptionDefinitions" :key="`rule-${variant.id || index}-${spec.id}`" class="rounded-md border bg-background p-2.5">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <span class="text-xs font-semibold">{{ specLabel(spec) }}</span>
                  <label class="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                    <Switch :model-value="ensureGroupRule(variant, spec).is_applicable" :aria-label="`${spec.name}适用于此 SKU`" @update:model-value="value => ensureGroupRule(variant, spec).is_applicable = Boolean(value)" />
                    适用于此 SKU
                  </label>
                </div>
                <div class="mt-2 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                  <span>选择数覆盖</span>
                  <Input v-model.number="ensureGroupRule(variant, spec).min_selections_override" class="h-7 w-20 text-xs" type="number" min="0" placeholder="最少" />
                  <span>-</span>
                  <Input v-model.number="ensureGroupRule(variant, spec).max_selections_override" class="h-7 w-20 text-xs" type="number" min="0" placeholder="最多" />
                </div>
                <div v-if="optionsForSpec(spec).length" class="mt-2 grid gap-1.5">
                  <div v-for="option in optionsForSpec(spec)" :key="`${variant.id || index}-${option.id || option.local_key}`" class="grid grid-cols-[minmax(0,1fr)_auto_7rem_minmax(7rem,1fr)] items-center gap-2 rounded border px-2 py-1.5">
                    <span class="min-w-0 truncate text-xs">{{ option.label || option.value_key }}</span>
                    <Switch
                      v-if="option.id"
                      :model-value="ensureValueRule(variant, option).is_enabled"
                      :aria-label="`${option.label || option.value_key}在此 SKU 可用`"
                      @update:model-value="value => ensureValueRule(variant, option).is_enabled = Boolean(value)"
                    />
                    <Input
                      v-if="option.id"
                      v-model.number="ensureValueRule(variant, option).price_delta_minor_override"
                      class="h-7 text-xs"
                      type="number"
                      min="0"
                      placeholder="价格覆盖"
                    />
                    <Input
                      v-if="option.id"
                      v-model="ensureValueRule(variant, option).unavailable_reason"
                      class="h-7 text-xs"
                      placeholder="不可用原因"
                    />
                    <span v-else class="col-span-2 text-[10px] text-muted-foreground">保存商品后可配置 SKU 覆盖</span>
                  </div>
                </div>
              </div>
            </div>
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
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import { ImageUp, Info, LoaderCircle, Plus, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import mediaApi from '@/api/media'
import UploadSpecHint from '@/components/admin/UploadSpecHint.vue'
import { uploadSpecAccept, validateUploadFile } from '@/lib/uploadSpecs'
import type {
  ProductOptionValueRelationForm,
  ProductSpecDefinition,
  ProductVariantForm,
  ProductVariantOptionValueForm,
  ProductOptionGroupVariantRuleForm,
  ProductOptionValueVariantRuleForm,
  ShippingTemplateRecord,
} from '@/modules/product/productEditorTypes'

const defaultRadioName = `product-variant-default-${Math.random().toString(36).slice(2)}`

const props = withDefaults(defineProps<{
  variants: ProductVariantForm[]
  specDefinitions?: ProductSpecDefinition[]
  defaultIndex?: number
  shippingTemplates?: ShippingTemplateRecord[]
  optionValues?: ProductVariantOptionValueForm[]
  optionValueRelations?: ProductOptionValueRelationForm[]
  productId?: number | string | null
  customOptionDefinitions?: ProductSpecDefinition[]
}>(), {
  specDefinitions: () => [],
  defaultIndex: 0,
  shippingTemplates: () => [],
  optionValues: () => [],
  optionValueRelations: () => [],
  productId: null,
  customOptionDefinitions: () => [],
})

const emit = defineEmits<{
  (event: 'add'): void
  (event: 'remove', index: number): void
  (event: 'set-default', index: number): void
  (event: 'set-active', index: number, active: boolean): void
}>()

const swatchUploadingKey = ref('')

const presentableSpecDefinitions = computed(() => (
  props.specDefinitions.filter((spec) => spec?.presentation === 'color' || spec?.presentation === 'image')
))
const persistedOptionValues = computed(() => props.optionValues.filter((option) => Number(option.id || 0) > 0))
const hasUnsavedOptionValues = computed(() => props.optionValues.some((option) => Number(option.id || 0) <= 0))
const canConfigureOptionValueRelations = computed(() => Number(props.productId || 0) > 0)
const canAddOptionValueRelation = computed(() => canConfigureOptionValueRelations.value && persistedOptionValues.value.length >= 2)

const specOptions = (spec: ProductSpecDefinition): unknown[] => {
  const configuredValues = props.optionValues
    .filter((item) => Number(item?.spec_definition_id) === Number(spec?.id))
    .map((item) => String(item?.value_key || '').trim())
    .filter(Boolean)
  if (configuredValues.length) return configuredValues
  const templateValues = (spec.option_items || [])
    .map((item) => String(item?.value_key || '').trim())
    .filter(Boolean)
  if (templateValues.length) return templateValues
  return []
}

const specLabel = (spec: ProductSpecDefinition): string => spec.unit ? `${spec.name} (${spec.unit})` : spec.name
const optionLabel = (spec: ProductSpecDefinition, option: unknown): string => {
  const metadata = props.optionValues.find((item) => (
    Number(item?.spec_definition_id) === Number(spec?.id)
      && String(item?.value_key || '') === String(option)
  ))
  return metadata?.label || String(option).replace(/_/g, ' ')
}
const optionKey = (option: ProductVariantOptionValueForm): string => `${option?.spec_definition_id || 'option'}:${option?.id || option?.local_key || option?.value_key || 'new'}`
const optionsForSpec = (spec: ProductSpecDefinition): ProductVariantOptionValueForm[] => props.optionValues.filter((item) => Number(item?.spec_definition_id) === Number(spec?.id))
const relationOptionLabel = (option: ProductVariantOptionValueForm): string => {
  const definition = [...props.specDefinitions, ...props.customOptionDefinitions]
    .find((spec) => Number(spec.id) === Number(option.spec_definition_id))
  const valueLabel = option.label || option.value_key || `选项值 #${option.id}`
  return definition?.name ? `${definition.name} / ${valueLabel}` : valueLabel
}
const relationValue = (value: number | string | null): string => Number(value || 0) > 0 ? String(value) : ''
const setRelationOptionValue = (
  relation: ProductOptionValueRelationForm,
  field: 'source_option_value_id' | 'target_option_value_id',
  value: unknown,
): void => {
  relation[field] = Number(value || 0) || null
  if (field === 'source_option_value_id' && Number(relation.target_option_value_id) === Number(relation.source_option_value_id)) {
    relation.target_option_value_id = null
  }
}
const setRelationType = (relation: ProductOptionValueRelationForm, value: unknown): void => {
  relation.relation_type = value === 'conflicts' ? 'conflicts' : 'requires'
}
const addOptionValueRelation = (): void => {
  const [source, target] = persistedOptionValues.value
  if (!source?.id || !target?.id) return
  props.optionValueRelations.push({
    id: null,
    source_option_value_id: Number(source.id),
    target_option_value_id: Number(target.id),
    relation_type: 'requires',
  })
}
const removeOptionValueRelation = (index: number): void => {
  props.optionValueRelations.splice(index, 1)
}
const removeRelationsForOption = (optionID: number): void => {
  for (let index = props.optionValueRelations.length - 1; index >= 0; index -= 1) {
    const relation = props.optionValueRelations[index]
    if (Number(relation.source_option_value_id) === optionID || Number(relation.target_option_value_id) === optionID) {
      props.optionValueRelations.splice(index, 1)
    }
  }
}
const ensureGroupRule = (variant: ProductVariantForm, spec: ProductSpecDefinition): ProductOptionGroupVariantRuleForm => {
  variant.option_group_rules ||= []
  const existing = variant.option_group_rules.find(rule => Number(rule.spec_definition_id) === Number(spec.id))
  if (existing) return existing
  const created: ProductOptionGroupVariantRuleForm = {
    spec_definition_id: Number(spec.id || 0),
    is_applicable: true,
    min_selections_override: null,
    max_selections_override: null,
  }
  variant.option_group_rules.push(created)
  return created
}
const ensureValueRule = (variant: ProductVariantForm, option: ProductVariantOptionValueForm): ProductOptionValueVariantRuleForm => {
  variant.option_value_rules ||= []
  const optionID = Number(option.id || 0)
  const existing = variant.option_value_rules.find(rule => Number(rule.product_variant_option_value_id) === optionID)
  if (existing) return existing
  const created: ProductOptionValueVariantRuleForm = {
    product_variant_option_value_id: optionID,
    is_enabled: true,
    price_delta_minor_override: null,
    unavailable_reason: '',
  }
  variant.option_value_rules.push(created)
  return created
}
const addOptionValue = (spec: ProductSpecDefinition): void => {
  props.optionValues.push({
    id: null,
    local_key: `option-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    spec_definition_id: spec.id ?? '',
    value_key: '',
    label: '',
    color_hex: '',
    swatch_media_asset_id: null,
    swatch_url: '',
    sort_order: optionsForSpec(spec).length * 10,
    is_enabled: true
  })
}
const removeOptionValue = (option: ProductVariantOptionValueForm): void => {
  const optionID = Number(option.id || 0)
  if (optionID > 0) removeRelationsForOption(optionID)
  const index = props.optionValues.indexOf(option)
  if (index >= 0) props.optionValues.splice(index, 1)
}
const clearSwatch = (option: ProductVariantOptionValueForm): void => {
  option.swatch_media_asset_id = null
  option.swatch_url = ''
}
const setInventoryPolicy = (option: ProductVariantOptionValueForm, policy: string): void => {
  option.inventory_policy = policy === 'component' ? 'component' : 'none'
  if (option.inventory_policy !== 'component') {
    option.component_variant_id = null
    option.component_quantity = 0
  }
}
const uploadSwatch = async (event: Event, option: ProductVariantOptionValueForm): Promise<void> => {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (input) input.value = ''
  if (!file) return
  const key = optionKey(option)
  swatchUploadingKey.value = key
  try {
    const validation = await validateUploadFile(file, 'product_variant_swatch')
    if (!validation.ok) {
      toast.error(validation.error || 'SKU 展示图片不符合上传规范')
      return
    }
    if (validation.warning) toast.warning(validation.warning)
    const formData = new FormData()
    formData.append('file', file)
    formData.append('media_type', 'image')
    formData.append('image_purpose', 'product_variant_swatch')
    const asset = await mediaApi.uploadAsset(formData)
    option.swatch_media_asset_id = asset?.id || null
    option.swatch_url = String(asset?.url || asset?.access_url || '')
    toast.success('展示图片已上传')
  } catch (error) {
    console.error('Failed to upload variant option swatch:', error)
    toast.error('展示图片上传失败')
  } finally {
    swatchUploadingKey.value = ''
  }
}
const selectValue = (value: unknown): string => value === undefined || value === null || value === '' ? '__empty__' : String(value)
const shippingTemplateSelectValue = (value: unknown): string => value === undefined || value === null || value === '' ? '__inherit__' : String(value)
const checkboxValue = (event: Event): boolean => Boolean((event.target as HTMLInputElement | null)?.checked)
const setSelectValue = (variant: ProductVariantForm, slug: string, value: unknown): void => {
  variant.option_values[slug] = value === '__empty__' ? '' : value
}

const setShippingTemplateValue = (variant: ProductVariantForm, value: unknown): void => {
  variant.shipping_template_id = value === '__inherit__' ? null : Number(value)
}

</script>

