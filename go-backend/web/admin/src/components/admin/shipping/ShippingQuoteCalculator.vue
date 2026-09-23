<template>
  <section class="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
    <div class="rounded-lg border bg-card p-4 shadow-xs">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 class="text-sm font-black tracking-tighter uppercase">运费试算器</h2>
          <p class="mt-1 text-xs text-muted-foreground">
            输入国家和商品/SKU，系统从数据库读取真实价格与 SKU 重量，再走后端报价规则。
          </p>
        </div>
        <Badge variant="outline" class="w-fit border-blue-200 bg-blue-50 text-blue-700">QUOTE API</Badge>
      </div>

      <form class="mt-5 space-y-4" @submit.prevent="submitQuote">
        <div class="grid gap-3 sm:grid-cols-3">
          <AdminFormField label="国家/地区代码" required :error="errors.country">
            <Input
              v-model.trim="form.country"
              class="font-mono uppercase"
              placeholder="US"
              maxlength="8"
              @input="clearError('country')"
            />
          </AdminFormField>

          <AdminFormField label="邮编 / Postal Code" description="用于命中线路的偏远邮编段规则。">
            <Input v-model.trim="form.postal_code" class="font-mono" placeholder="可选，例如 10005" />
          </AdminFormField>

          <AdminFormField label="币种" required :error="errors.currency">
            <Input v-model.trim="form.currency" class="font-mono uppercase" placeholder="ISO 4217" maxlength="10" @input="clearError('currency')" />
          </AdminFormField>
        </div>

        <div class="space-y-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-xs font-black uppercase tracking-wider">试算商品</h3>
              <p class="mt-1 text-[10px] text-muted-foreground">variant_id 可空；为空时后端会选择可购买的默认 SKU。</p>
            </div>
            <Button type="button" variant="outline" size="sm" @click="addItem">
              <Plus class="size-3.5" />
              添加一行
            </Button>
          </div>

          <div v-for="(item, index) in form.items" :key="index" class="grid gap-3 rounded-lg border p-3 lg:grid-cols-12">
            <AdminFormField label="Product ID" class="lg:col-span-3" :error="itemErrors(index).product_id">
              <Input v-model.number="item.product_id" type="number" min="1" step="1" @input="clearItemError(index, 'product_id')" />
            </AdminFormField>

            <AdminFormField label="Variant / SKU ID" class="lg:col-span-3">
              <Input v-model.number="item.variant_id" type="number" min="1" step="1" placeholder="可空" />
            </AdminFormField>

            <AdminFormField label="数量" class="lg:col-span-2" :error="itemErrors(index).quantity">
              <Input v-model.number="item.quantity" type="number" min="1" step="1" @input="clearItemError(index, 'quantity')" />
            </AdminFormField>

            <div class="flex items-end justify-end lg:col-span-4">
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                class="text-destructive hover:text-destructive"
                :disabled="form.items.length === 1"
                @click="removeItem(index)"
              >
                <Trash2 class="size-4" />
                <span class="sr-only">删除试算商品</span>
              </Button>
            </div>
          </div>
        </div>

        <div v-if="quoteError" class="rounded-lg border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
          {{ quoteError }}
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-muted/35 p-3">
          <p class="text-xs text-muted-foreground">
            报价结果只用于验证规则，不会创建订单，也不会写入库存/订单数据。
          </p>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '试算中' : '开始试算' }}
          </Button>
        </div>
      </form>
    </div>

    <div class="rounded-lg border bg-card p-4 shadow-xs">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 class="text-sm font-black tracking-tighter uppercase">报价结果</h2>
          <p class="mt-1 text-xs text-muted-foreground">
            明细会显示商品/SKU 设置的模板、线路候选、SKU 实重、包装重量、计费重量和分摊运费，方便排查规则矩阵。
          </p>
        </div>
 <Badge v-if="quote" variant="outline" :class="quote.free_shipping ? 'border-emerald-200 bg-emerald-50 text-emerald-700': 'border-amber-200 bg-amber-50 text-amber-700'">
          {{ quote.free_shipping ? 'FREE SHIPPING' : 'CHARGED' }}
        </Badge>
      </div>

      <div v-if="!quote" class="mt-5 rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
        还没有试算结果。先输入国家和商品/SKU，然后点击“开始试算”。
      </div>

      <div v-else class="mt-5 space-y-4">
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-lg border bg-muted/35 p-3">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">运费</span>
            <p class="mt-1 text-2xl font-black tracking-tighter">{{ formatMoney(quote.shipping_fee_minor, quote.currency) }}</p>
          </div>
          <div class="rounded-lg border bg-muted/35 p-3">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">币种</span>
 <p class="mt-1 text-2xl font-black tracking-tighter">{{ quote.currency || '币种缺失'}}</p>
          </div>
          <div class="rounded-lg border bg-muted/35 p-3">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">免运</span>
 <p class="mt-1 text-2xl font-black tracking-tighter">{{ quote.free_shipping ? '是': '否'}}</p>
          </div>
          <div class="rounded-lg border bg-muted/35 p-3">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">默认方案</span>
            <p class="mt-1 text-sm font-black">{{ selectedPlanLabel(quote.selected_plan) }}</p>
            <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ quote.selected_plan?.id || '-' }}</p>
          </div>
        </div>

        <div class="rounded-lg border bg-card p-3">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-xs font-black uppercase tracking-wider">可选运输方案</h3>
              <p class="mt-1 text-[10px] text-muted-foreground">
                每个方案覆盖全部商品分组；多模板商品会显示为包含多段线路的完整方案。
              </p>
            </div>
            <Badge variant="outline" class="w-fit">{{ quote.plans?.length || 0 }} PLANS</Badge>
          </div>

          <div v-if="!quote.plans?.length" class="mt-3 rounded-lg border border-dashed p-4 text-xs text-muted-foreground">
            当前没有覆盖全部商品分组的运输方案。
          </div>

          <AdminTablePanel v-else :loading="false" class="mt-3">
            <Table class="min-w-[1080px]">
              <TableHeader>
                <TableRow>
                  <TableHead>运输方案</TableHead>
                  <TableHead class="w-24 text-right">分段</TableHead>
                  <TableHead class="w-64 text-right">计费明细</TableHead>
                  <TableHead class="w-28 text-right">总运费</TableHead>
                  <TableHead class="w-28 text-right">时效</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="plan in quote.plans" :key="plan.id || selectedPlanLabel(plan)">
                  <TableCell>
                    <div class="flex items-start gap-2">
                      <Badge
                        v-if="isSelectedPlan(plan)"
                        variant="outline"
                        class="mt-0.5 border-emerald-200 bg-emerald-50 text-[10px] text-emerald-700"
                      >
                        默认
                      </Badge>
                      <div class="min-w-0 space-y-1">
                        <div v-for="leg in plan.legs || []" :key="leg.group_key || `${leg.template_id}-${leg.carrier_service_id}`">
                          <span class="block text-xs font-bold">{{ leg.service_name || leg.template_name || '-' }}</span>
                          <span class="block font-mono text-[10px] text-muted-foreground">
                            {{ leg.carrier_name || '模板费率' }} · {{ leg.service_code || leg.group_key || '-' }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="text-right font-mono text-xs">{{ plan.legs?.length || 0 }}</TableCell>
                  <TableCell class="text-right font-mono text-[10px] text-muted-foreground">
                    <div v-for="leg in plan.legs || []" :key="`fee-${leg.group_key || leg.template_id}`">
                      {{ billingModeLabel(leg.billing_mode) }} · {{ formatGrams(leg.billable_weight_grams) }} · {{ formatMoney(leg.shipping_fee_minor, plan.currency || quote.currency) }}
                    </div>
                  </TableCell>
                  <TableCell class="text-right text-sm font-black tabular-nums">{{ formatMoney(plan.shipping_fee_minor, plan.currency || quote.currency) }}</TableCell>
                  <TableCell class="text-right text-xs tabular-nums">{{ formatEta(plan) }}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </AdminTablePanel>
        </div>

        <AdminTablePanel :loading="false">
          <Table class="min-w-[1120px]">
            <TableHeader>
              <TableRow>
                <TableHead>商品 / SKU</TableHead>
                <TableHead>命中模板</TableHead>
                <TableHead class="w-24 text-right">数量</TableHead>
                <TableHead class="w-28 text-right">单价</TableHead>
                <TableHead class="w-32 text-right">SKU 重量</TableHead>
                <TableHead class="w-44">包装规则</TableHead>
                <TableHead class="w-32 text-right">计费重量</TableHead>
                <TableHead class="w-32 text-right">分摊运费</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="!quote.items?.length" :colspan="8">暂无明细</TableEmpty>
              <TableRow v-for="item in quote.items" :key="`${item.product_id}-${item.variant_id || 0}`">
                <TableCell class="font-mono text-xs">
                  product_id={{ item.product_id }}<br />
                  variant_id={{ item.variant_id || '-' }}
                </TableCell>
                <TableCell>
 <span class="block font-bold text-xs">{{ item.template_name || '-'}}</span>
 <span class="block text-[10px] text-muted-foreground">template_id={{ item.template_id || '-'}}</span>
                </TableCell>
                <TableCell class="text-right tabular-nums">{{ item.quantity || 0 }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatMoney(item.unit_price_minor, quote.currency) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatGrams(item.weight_grams) }}</TableCell>
                <TableCell>
 <span class="block text-xs font-bold">{{ item.packaging_rule_name || '未绑定包装'}}</span>
                  <span class="block text-[10px] text-muted-foreground">
                    {{ item.packaging_rule_id ? `rule_id=${item.packaging_rule_id}` : '按 SKU 实重计费' }}
                    · {{ formatGrams(item.packaging_weight_grams) }}
                  </span>
                </TableCell>
                <TableCell class="text-right tabular-nums">{{ formatGrams(item.charge_weight_grams || item.weight_grams) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatMoney(item.shipping_fee_minor, quote.currency) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { LoaderCircle, Plus, Trash2 } from '@lucide/vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type {
  ShippingQuoteForm,
  ShippingQuoteItemInput,
  ShippingQuotePlan,
  ShippingQuoteResult
} from '@/modules/shipping/shippingTypes'
import { formatMinorMoney } from '@/lib/dashboardPresentation'

const form = reactive<ShippingQuoteForm>({
  country: 'US',
  postal_code: '',
  currency: '',
  items: [
    {
      product_id: '',
      variant_id: '',
      quantity: 1,
    },
  ],
})

const errors = reactive<Record<string, string>>({})
const quote = ref<ShippingQuoteResult | null>(null)
const quoteError = ref('')
const submitting = ref(false)

const defaultItem = (): ShippingQuoteItemInput => ({
  product_id: '',
  variant_id: '',
  quantity: 1,
})

const addItem = () => {
  form.items.push(defaultItem())
}

const removeItem = (index) => {
  if (form.items.length === 1) return
  form.items.splice(index, 1)
}

const clearError = (field) => {
  delete errors[field]
}

const itemErrorKey = (index, field) => `items.${index}.${field}`

const itemErrors = (index: number) => ({
  product_id: errors[itemErrorKey(index, 'product_id')],
  quantity: errors[itemErrorKey(index, 'quantity')],
})

const clearItemError = (index: number, field: string) => {
  delete errors[itemErrorKey(index, field)]
}

const validate = () => {
  Object.keys(errors).forEach((key) => delete errors[key])
  if (!form.country?.trim()) errors.country = '请输入国家/地区代码'
  if (!form.currency?.trim()) errors.currency = '请输入报价币种'

  form.items.forEach((item, index) => {
    if (!Number(item.product_id || 0)) errors[itemErrorKey(index, 'product_id')] = '请输入 Product ID'
    if (Number(item.quantity || 0) <= 0) errors[itemErrorKey(index, 'quantity')] = '数量必须大于 0'
  })

  return Object.keys(errors).length === 0
}

const buildPayload = () => ({
  country: form.country.trim().toUpperCase(),
  ...(form.postal_code.trim() ? { postal_code: form.postal_code.trim() } : {}),
  currency: form.currency.trim().toUpperCase(),
  items: form.items.map((item) => ({
    product_id: Number(item.product_id),
    variant_id: Number(item.variant_id || 0) > 0 ? Number(item.variant_id) : null,
    quantity: Number(item.quantity || 1),
  })),
})

const submitQuote = async () => {
  if (!validate()) return

  submitting.value = true
  quoteError.value = ''
  try {
    quote.value = await shippingApi.quote(buildPayload())
    toast.success('运费试算完成')
  } catch (error) {
    quote.value = null
    quoteError.value = error?.response?.data?.error || error?.message || '运费试算失败'
  } finally {
    submitting.value = false
  }
}

const formatMoney = (value: unknown, currency?: string | null) => formatMinorMoney(value as number | string | null | undefined, currency)
const formatGrams = (value: unknown) => `${Number(value || 0).toLocaleString()} g`
const selectedPlanLabel = (plan?: ShippingQuotePlan | null) => {
  if (!plan) return '未命中方案'
  const labels = (plan.legs || []).map(leg => leg.service_name || leg.template_name || '').filter(Boolean)
  return labels.join(' + ') || '未命名方案'
}
const billingModeLabel = (mode?: string | null) => ({
  actual_weight: '实重计费',
  volumetric_weight: '体积重计费',
  greater_of_actual_and_volumetric: '实重/体积重取大',
}[mode] || mode || '-')
const formatEta = (plan?: ShippingQuotePlan | null) => {
  const min = Number(plan?.eta_min_days || 0)
  const max = Number(plan?.eta_max_days || 0)
  if (min > 0 && max > 0) return min === max ? `${min} 天` : `${min}-${max} 天`
  if (min > 0) return `${min}+ 天`
  if (max > 0) return `${max} 天内`
  return '-'
}
const isSelectedPlan = (plan?: ShippingQuotePlan | null) =>
  Boolean(plan?.id) && plan?.id === quote.value?.selected_plan?.id
</script>

