<template>
  <section class="space-y-3 rounded-xl border border-emerald-500/20 bg-emerald-500/[0.035] p-3">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <div class="flex flex-wrap items-center gap-2">
          <h2 class="text-sm font-bold text-foreground">商品成本与利润预估</h2>
          <span class="rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-semibold text-emerald-700 dark:text-emerald-300">
            独立附加域
          </span>
          <span v-if="!canEdit" class="rounded-full bg-muted px-2 py-0.5 text-[10px] font-semibold text-muted-foreground">
            只读
          </span>
        </div>
        <p class="mt-1 text-xs leading-5 text-muted-foreground">
          资料按 SKU 保存，不会写入商品目录；预计毛利 = 实际售价 - 采购价 - 运费、包装和其他附加成本。
        </p>
      </div>
      <div class="text-right text-xs text-muted-foreground">
        <span v-if="loading">正在读取成本资料...</span>
        <span v-else-if="saving">正在保存成本与利润...</span>
        <span v-else-if="lastSavedAt">最近保存 {{ formatSavedAt(lastSavedAt) }}</span>
      </div>
    </div>

    <div v-if="pending || error" class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2.5">
      <div class="flex min-w-0 items-start gap-2 text-xs leading-5 text-amber-800 dark:text-amber-200">
        <TriangleAlert class="mt-0.5 size-4 shrink-0" />
        <span>{{ error || '商品已保存，但成本与利润资料尚未保存。' }}</span>
      </div>
      <Button
        v-if="canEdit && pending"
        type="button"
        variant="outline"
        size="sm"
        :disabled="saving"
        @click="emit('retry')"
      >
        <RefreshCw class="size-3.5" :class="{ 'animate-spin': saving }" />
        独立重试
      </Button>
    </div>

    <div v-if="!variants.length" class="rounded-lg border border-dashed px-3 py-5 text-center text-xs text-muted-foreground">
        保存商品 SKU 后，这里会显示对应的成本和利润资料。
    </div>

    <div v-for="(variant, index) in variants" :key="variant.id || `profit-variant-${index}`" class="space-y-3 rounded-xl border bg-background/80 p-3">
      <div class="flex flex-wrap items-center justify-between gap-2 border-b pb-2">
        <div class="flex min-w-0 items-center gap-2">
          <span class="rounded-md bg-muted px-2 py-1 font-mono text-xs font-bold">
            {{ variant.sku || `待填写 SKU ${index + 1}` }}
          </span>
          <span v-if="variant.title" class="min-w-0 truncate text-xs text-muted-foreground">{{ variant.title }}</span>
        </div>
        <div class="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
          <span>商品售价 {{ formatMoney(effectiveSellingPrice(variant), currency) }}</span>
          <span v-if="calculationFor(index).status === 'ready' || calculationFor(index).status === 'warning'" class="font-semibold text-emerald-700 dark:text-emerald-300">
            预计毛利 {{ formatMoney(calculationFor(index).grossProfit, currency) }}
          </span>
        </div>
      </div>

      <div class="grid gap-3 xl:grid-cols-[minmax(0,1.05fr)_minmax(0,1.35fr)]">
        <div class="space-y-3">
          <div class="grid gap-3 sm:grid-cols-2">
            <AdminFormField label="采购价" description="空值不会按零成本计算">
              <Input
                :model-value="drafts[index].purchasePrice ?? undefined"
                type="number"
                min="0"
                step="0.01"
                placeholder="未填写"
                :disabled="!canEdit"
                @update:model-value="setPurchasePrice(drafts[index], $event)"
              />
            </AdminFormField>
            <AdminFormField label="采购币种" description="必须与商品主币种一致">
              <Input
                :model-value="drafts[index].currency"
                class="font-mono uppercase"
                maxlength="3"
                :disabled="!canEdit"
                @update:model-value="setDraftCurrency(drafts[index], $event)"
              />
            </AdminFormField>
            <AdminFormField label="入库运费 / 件">
              <Input v-model.number="drafts[index].inboundShippingUnitCost" type="number" min="0" step="0.01" :disabled="!canEdit" />
            </AdminFormField>
            <AdminFormField label="包装成本 / 件">
              <Input v-model.number="drafts[index].packagingUnitCost" type="number" min="0" step="0.01" :disabled="!canEdit" />
            </AdminFormField>
            <AdminFormField label="其他成本 / 件">
              <Input v-model.number="drafts[index].otherUnitCost" type="number" min="0" step="0.01" :disabled="!canEdit" />
            </AdminFormField>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <AdminFormField label="供应商">
              <Input v-model="drafts[index].supplierName" placeholder="供应商名称" :disabled="!canEdit" />
            </AdminFormField>
            <AdminFormField label="供应商联系人">
              <Input v-model="drafts[index].supplierContactName" placeholder="联系人姓名" :disabled="!canEdit" />
            </AdminFormField>
            <AdminFormField label="联系电话">
              <Input v-model="drafts[index].supplierPhone" placeholder="电话" :disabled="!canEdit" />
            </AdminFormField>
            <AdminFormField label="供应商邮箱">
              <Input v-model="drafts[index].supplierEmail" type="email" placeholder="邮箱" :disabled="!canEdit" />
            </AdminFormField>
            <AdminFormField label="到货周期（天）">
              <Input v-model.number="drafts[index].leadTimeDays" type="number" min="0" step="1" :disabled="!canEdit" />
            </AdminFormField>
            <AdminFormField label="最小起订量">
              <Input v-model.number="drafts[index].minimumOrderQuantity" type="number" min="1" step="1" :disabled="!canEdit" />
            </AdminFormField>
          </div>
        </div>

        <div class="space-y-3 rounded-xl border border-dashed bg-muted/20 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <h3 class="text-sm font-semibold text-foreground">利润预估</h3>
              <p class="mt-1 text-xs text-muted-foreground">后端保存时会重新计算并作为最终结果。</p>
            </div>
            <span :class="statusClass(calculationFor(index).status)" class="rounded-full px-2.5 py-1 text-[11px] font-semibold">
              {{ statusLabel(calculationFor(index).status) }}
            </span>
          </div>

          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <Metric label="常规售价" :value="formatMoney(Number(variant.price || 0), currency)" />
            <Metric label="实际售价" :value="formatMoney(effectiveSellingPrice(variant), currency)" />
            <Metric label="含附加成本" :value="calculationFor(index).landedCost == null ? '待填写' : formatMoney(calculationFor(index).landedCost, currency)" />
            <Metric label="预计毛利率" :value="calculationFor(index).grossMargin == null ? '待填写' : `${calculationFor(index).grossMargin.toFixed(2)}%`" />
          </div>

          <div class="rounded-lg border bg-background px-3 py-3">
            <div class="flex flex-wrap items-end justify-between gap-3">
              <div>
                <p class="text-xs text-muted-foreground">预计单位毛利</p>
                <p
                  :class="calculationFor(index).grossProfit != null && calculationFor(index).grossProfit < 0 ? 'text-destructive' : 'text-emerald-700 dark:text-emerald-300'"
                  class="mt-1 text-2xl font-black tabular-nums"
                >
                  {{ calculationFor(index).grossProfit == null ? '待填写' : formatMoney(calculationFor(index).grossProfit, currency) }}
                </p>
              </div>
              <div v-if="calculationFor(index).currencyMismatch" class="max-w-xs text-right text-xs leading-5 text-amber-700 dark:text-amber-300">
                采购币种 {{ drafts[index].currency || '未设置' }} 与商品币种 {{ currency }} 不一致
              </div>
            </div>
          </div>

          <ul v-if="calculationFor(index).warnings.length" class="space-y-1.5 text-xs leading-5 text-amber-700 dark:text-amber-300">
            <li v-for="warning in calculationFor(index).warnings" :key="warning" class="flex items-start gap-1.5">
              <TriangleAlert class="mt-0.5 size-3.5 shrink-0" />
              <span>{{ warningLabel(warning) }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, defineComponent, h } from 'vue'
import { RefreshCw, TriangleAlert } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { ProcurementProfitDraft } from '@/composables/product/useProcurementProfitDraft'
import type { ProductVariantForm } from './productEditorTypes'

const props = withDefaults(defineProps<{
  variants: ProductVariantForm[]
  drafts: ProcurementProfitDraft[]
  currency?: string
  canEdit?: boolean
  loading?: boolean
  saving?: boolean
  pending?: boolean
  error?: string
  lastSavedAt?: string
}>(), {
  currency: 'USD',
  canEdit: false,
  loading: false,
  saving: false,
  pending: false,
  error: '',
  lastSavedAt: '',
})

const emit = defineEmits<{
  (event: 'retry'): void
}>()

const currencyMinorUnits: Record<string, number> = {
  BHD: 3,
  JOD: 3,
  KWD: 3,
  OMR: 3,
  TND: 3,
  JPY: 0,
  KRW: 0,
  VND: 0,
}

type CalculationStatus = 'ready' | 'warning' | 'missing_purchase_price' | 'currency_mismatch' | 'invalid'

interface LocalCalculation {
  status: CalculationStatus
  landedCost: number | null
  grossProfit: number | null
  grossMargin: number | null
  currencyMismatch: boolean
  warnings: string[]
}

const normalizeCurrency = (value: unknown): string => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : 'USD'
}

const enteredCurrency = (value: unknown): string => String(value || '').trim().toUpperCase()

const roundMoney = (value: number, currency: string): number => {
  const units = currencyMinorUnits[normalizeCurrency(currency)] ?? 2
  const scale = 10 ** units
  return Math.round(value * scale) / scale
}

const finiteNumber = (value: unknown): number | null => {
  const number = Number(value)
  return Number.isFinite(number) ? number : null
}

const draftAt = (index: number): ProcurementProfitDraft => props.drafts[index] || {
  productCode: '',
  productName: '',
  purchasePrice: null,
  purchasePriceKnown: false,
  currency: props.currency,
  supplierName: '',
  supplierContactName: '',
  supplierPhone: '',
  supplierEmail: '',
  leadTimeDays: 0,
  minimumOrderQuantity: 1,
  inboundShippingUnitCost: 0,
  packagingUnitCost: 0,
  otherUnitCost: 0,
}

const effectiveSellingPrice = (variant: ProductVariantForm): number => {
  const salePrice = finiteNumber(variant.sale_price)
  if (salePrice != null) return salePrice
  return finiteNumber(variant.price) || 0
}

const calculationFor = (index: number): LocalCalculation => {
  const draft = draftAt(index)
  const currency = normalizeCurrency(props.currency)
  const rawCostCurrency = enteredCurrency(draft.currency)
  const costCurrency = rawCostCurrency || currency
  const warnings: string[] = []
  const variant = props.variants[index] as ProductVariantForm | undefined
  const listPrice = finiteNumber(variant?.price) || 0
  const salePrice = finiteNumber(variant?.sale_price)
  const sellingPrice = variant ? roundMoney(effectiveSellingPrice(variant), currency) : 0

  if (salePrice == null) warnings.push('sale_price_missing')
  if (salePrice != null && salePrice > listPrice) warnings.push('sale_price_above_list_price')
  if (!/^[A-Z]{3}$/.test(costCurrency)) {
    return { status: 'invalid', landedCost: null, grossProfit: null, grossMargin: null, currencyMismatch: false, warnings: [...warnings, 'invalid_currency'] }
  }
  if (currency !== costCurrency) {
    warnings.push('currency_mismatch')
    return { status: 'currency_mismatch', landedCost: null, grossProfit: null, grossMargin: null, currencyMismatch: true, warnings }
  }
  if (sellingPrice <= 0) return { status: 'invalid', landedCost: null, grossProfit: null, grossMargin: null, currencyMismatch: false, warnings: ['invalid_selling_price'] }
  if (!draft.purchasePriceKnown || draft.purchasePrice == null) {
    return { status: 'missing_purchase_price', landedCost: null, grossProfit: null, grossMargin: null, currencyMismatch: false, warnings: [...warnings, 'missing_purchase_price'] }
  }

  const costValues = [
    finiteNumber(draft.purchasePrice),
    finiteNumber(draft.inboundShippingUnitCost),
    finiteNumber(draft.packagingUnitCost),
    finiteNumber(draft.otherUnitCost),
  ]
  if (costValues.some((value) => value == null || value < 0)) {
    return { status: 'invalid', landedCost: null, grossProfit: null, grossMargin: null, currencyMismatch: false, warnings: [...warnings, 'invalid_cost'] }
  }
  const landedCost = roundMoney(costValues.reduce((total, value) => total + (value || 0), 0), currency)
  const grossProfit = roundMoney(sellingPrice - landedCost, currency)
  const grossMargin = grossProfit / sellingPrice * 100
  if (grossProfit < 0) warnings.push('negative_gross_profit')
  return {
    status: warnings.length ? 'warning' : 'ready',
    landedCost,
    grossProfit,
    grossMargin,
    currencyMismatch: false,
    warnings,
  }
}

const calculations = computed(() => props.variants.map((_, index) => calculationFor(index)))

const statusLabel = (status: CalculationStatus): string => ({
  ready: '可计算',
  warning: '需要关注',
  missing_purchase_price: '待填写采购价',
  currency_mismatch: '币种不一致',
  invalid: '输入有误',
}[status])

const statusClass = (status: CalculationStatus): string => ({
  ready: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
  warning: 'bg-amber-500/10 text-amber-700 dark:text-amber-300',
  missing_purchase_price: 'bg-muted text-muted-foreground',
  currency_mismatch: 'bg-amber-500/10 text-amber-700 dark:text-amber-300',
  invalid: 'bg-red-500/10 text-red-700 dark:text-red-300',
}[status])

const warningLabel = (warning: string): string => ({
  sale_price_missing: '未填写促销价，当前按常规售价计算。',
  sale_price_above_list_price: '促销价高于常规售价，请确认价格关系。',
  negative_gross_profit: '预计单位毛利为负，请确认售价或采购成本。',
  missing_purchase_price: '采购价未确认，暂不生成利润数值。',
  currency_mismatch: '采购币种与商品币种不一致，后端不会自动换算。',
  invalid_selling_price: '实际售价必须大于 0。',
  invalid_cost: '采购价和附加成本必须是有效的非负金额。',
  invalid_currency: '采购币种必须填写有效的三位字母币种代码。',
}[warning] || warning)

const formatMoney = (value: number, currency: string): string => {
  const code = normalizeCurrency(currency)
  try {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: code }).format(value)
  } catch {
    return `${code} ${value.toFixed(2)}`
  }
}

const formatSavedAt = (value: string): string => {
  const timestamp = new Date(value)
  if (Number.isNaN(timestamp.getTime())) return value
  return timestamp.toLocaleString('zh-CN', { hour12: false })
}

const setPurchasePrice = (draft: ProcurementProfitDraft, value: string | number): void => {
  const rawValue = String(value ?? '').trim()
  const parsedValue = rawValue === '' ? null : Number(rawValue)
  draft.purchasePrice = parsedValue != null && Number.isFinite(parsedValue) ? parsedValue : null
  draft.purchasePriceKnown = draft.purchasePrice != null
}

const setDraftCurrency = (draft: ProcurementProfitDraft, value: string | number): void => {
  draft.currency = String(value || '').toUpperCase()
}

const Metric = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
  },
  setup(metricProps) {
    return () => h('div', { class: 'rounded-lg border bg-background px-2.5 py-2' }, [
      h('p', { class: 'text-[11px] text-muted-foreground' }, metricProps.label),
      h('p', { class: 'mt-1 truncate text-sm font-bold tabular-nums text-foreground' }, metricProps.value),
    ])
  },
})
</script>
