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

    <div v-if="presentableSpecDefinitions.length" class="space-y-3 rounded-lg border bg-muted/10 p-3">
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

        <div v-if="optionsForSpec(spec).length" class="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
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
                <input class="sr-only" type="file" accept="image/jpeg,image/png,image/webp" :disabled="swatchUploadingKey === optionKey(option)" @change="uploadSwatch($event, option)" />
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
        <p v-else class="rounded-md border border-dashed px-3 py-3 text-center text-xs text-muted-foreground">
          还没有配置展示选项；先添加颜色或图片选项，再在下方 SKU 表中选择。
        </p>
      </div>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-2 rounded-lg border bg-background px-3 py-2">
      <div class="text-xs leading-5 text-muted-foreground">
        主价格按 {{ currency || 'USD' }} 录入；按钮会读取后台缓存汇率并填充本 SKU 的次展示价，随商品保存。
      </div>
      <Button
        type="button"
        variant="outline"
        size="sm"
        :disabled="displayPriceLoading || !hasFillableVariantPrices()"
        @click="fillDisplayPrices"
      >
        <RefreshCw class="size-3.5" :class="{ 'animate-spin': displayPriceLoading }" />
        {{ displayPriceLoading ? '填充中' : '按汇率填充次币种' }}
      </Button>
      <p v-if="displayPriceError" class="basis-full text-xs font-medium text-destructive">{{ displayPriceError }}</p>
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
          <TableHead class="w-56">次展示价</TableHead>
          <TableHead class="w-28">重量（克）</TableHead>
          <TableHead class="w-44">运费模板</TableHead>
          <TableHead class="w-24">库存</TableHead>
          <TableHead class="w-20 text-center">启用</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="(variant, index) in variants" :key="variant.id || `variant-${index}`">
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
              <Input v-model.number="variant.price" type="number" min="0" step="0.01" @input="clearVariantDisplayPrices(variant, index)" />
            </TableCell>
            <TableCell>
              <Input v-model.number="variant.sale_price" type="number" min="0" step="0.01" placeholder="可选" @input="clearVariantDisplayPrices(variant, index)" />
            </TableCell>
            <TableCell>
              <div v-if="displayPricesForVariant(variant, index).length" class="flex max-w-56 flex-wrap gap-1.5">
                <span
                  v-for="price in displayPricesForVariant(variant, index)"
                  :key="price.quote_currency || price.currency"
                  class="rounded-md border px-1.5 py-0.5 font-mono text-[11px]"
 :class="price.fallback_reason ? 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-200': 'bg-muted/40 text-foreground'"
                >
                  {{ formatDisplayPriceResult(price) }}
                </span>
              </div>
              <span v-else class="text-xs text-muted-foreground">未填充</span>
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
import axios from 'axios'
import { toast } from 'vue-sonner'
import { ImageUp, Info, LoaderCircle, Plus, RefreshCw, Trash2 } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import mediaApi from '@/api/media'
import type {
  ProductDisplayPriceResult,
  ProductSpecDefinition,
  ProductVariantForm,
  ProductVariantOptionValueForm,
  ShippingTemplateRecord,
} from './productEditorTypes'

const defaultRadioName = `product-variant-default-${Math.random().toString(36).slice(2)}`

const props = withDefaults(defineProps<{
  variants: ProductVariantForm[]
  currency?: string
  specDefinitions?: ProductSpecDefinition[]
  defaultIndex?: number
  shippingTemplates?: ShippingTemplateRecord[]
  optionValues?: ProductVariantOptionValueForm[]
}>(), {
  currency: 'USD',
  specDefinitions: () => [],
  defaultIndex: 0,
  shippingTemplates: () => [],
  optionValues: () => [],
})

const emit = defineEmits<{
  (event: 'add'): void
  (event: 'remove', index: number): void
  (event: 'set-default', index: number): void
  (event: 'set-active', index: number, active: boolean): void
}>()

const displayPriceRows = ref<Record<string, ProductDisplayPriceResult[]>>({})
const displayPriceLoading = ref(false)
const displayPriceError = ref('')
const swatchUploadingKey = ref('')

const presentableSpecDefinitions = computed(() => (
  props.specDefinitions.filter((spec) => spec?.presentation === 'color' || spec?.presentation === 'image')
))

const specOptions = (spec: ProductSpecDefinition): unknown[] => {
  const configuredValues = props.optionValues
    .filter((item) => Number(item?.spec_definition_id) === Number(spec?.id))
    .map((item) => String(item?.value_key || '').trim())
    .filter(Boolean)
  if (configuredValues.length) return configuredValues
  if (!spec?.options) return []
  try {
    const options = JSON.parse(spec.options)
    return Array.isArray(options) ? options : []
  } catch {
    return []
  }
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
  const index = props.optionValues.indexOf(option)
  if (index >= 0) props.optionValues.splice(index, 1)
}
const clearSwatch = (option: ProductVariantOptionValueForm): void => {
  option.swatch_media_asset_id = null
  option.swatch_url = ''
}
const uploadSwatch = async (event: Event, option: ProductVariantOptionValueForm): Promise<void> => {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (input) input.value = ''
  if (!file) return
  const key = optionKey(option)
  swatchUploadingKey.value = key
  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('media_type', 'image')
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
const normalizeCurrencyCode = (value: unknown): string => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

const effectiveVariantPrice = (variant: ProductVariantForm): number => {
  const salePrice = Number(variant?.sale_price || 0)
  if (salePrice > 0) return salePrice
  const price = Number(variant?.price || 0)
  return Number.isFinite(price) ? price : 0
}

const displayPriceKey = (variant: ProductVariantForm, index: number): string => variant?.id ? `id:${variant.id}` : `index:${index}`
const displayPricesForVariant = (variant: ProductVariantForm, index: number): ProductDisplayPriceResult[] => displayPriceRows.value[displayPriceKey(variant, index)] || variant?.display_prices || []
const hasFillableVariantPrices = (): boolean => props.variants.some(variant => effectiveVariantPrice(variant) > 0)

const clearVariantDisplayPrices = (variant: ProductVariantForm, index: number): void => {
  if (!variant) return
  variant.display_prices = []
  const nextRows = { ...displayPriceRows.value }
  delete nextRows[displayPriceKey(variant, index)]
  displayPriceRows.value = nextRows
}

const fillDisplayPrices = async (): Promise<void> => {
  displayPriceError.value = ''
  const baseCurrency = normalizeCurrencyCode(props.currency) || 'USD'
  const fillable = props.variants
    .map((variant, index) => ({ variant, index, amount: effectiveVariantPrice(variant) }))
    .filter(item => item.amount > 0)

  if (!fillable.length) {
    displayPriceError.value = '请先录入至少一个 SKU 主价格'
    return
  }

  displayPriceLoading.value = true
  try {
    const entries = await Promise.all(fillable.map(async ({ variant, index, amount }) => {
      const response = await axios.post('/api/admin/pricing/exchange-rates/convert', {
        amount,
        base_currency: baseCurrency
      })
      const data = response.data?.data || response.data || {}
      const prices = Array.isArray(data.prices) ? data.prices : []
      variant.display_prices = prices.filter(price => !price?.fallback_reason && price?.converted !== false)
      return [displayPriceKey(variant, index), prices] as [string, ProductDisplayPriceResult[]]
    }))
    displayPriceRows.value = Object.fromEntries(entries)
    toast.success('次展示价已按缓存汇率填充')
  } catch (error) {
    const message = errorMessage(error, '次展示价填充失败')
    displayPriceError.value = message
    toast.error(message)
  } finally {
    displayPriceLoading.value = false
  }
}

interface ErrorLike {
  response?: { data?: { message?: string; error?: string } }
  message?: string
}

const errorMessage = (error: unknown, fallback: string): string => {
  const value = error as ErrorLike
  return value?.response?.data?.message || value?.response?.data?.error || value?.message || fallback
}

const formatDisplayPriceResult = (price: ProductDisplayPriceResult): string => {
  const quoteCurrency = normalizeCurrencyCode(price?.quote_currency)
  if (price?.fallback_reason) {
    return `${quoteCurrency || normalizeCurrencyCode(price?.currency) || '---'} 缺汇率`
  }
  const currency = normalizeCurrencyCode(price?.currency) || quoteCurrency || 'USD'
  const amount = Number(price?.amount || 0)
  try {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency }).format(amount)
  } catch {
    return `${currency} ${amount.toFixed(2)}`
  }
}

const setSelectValue = (variant: ProductVariantForm, slug: string, value: unknown): void => {
  variant.option_values[slug] = value === '__empty__' ? '' : value
}

const setShippingTemplateValue = (variant: ProductVariantForm, value: unknown): void => {
  variant.shipping_template_id = value === '__inherit__' ? null : Number(value)
}

</script>
