<template>
  <section class="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
    <div class="rounded-lg border bg-card p-4 shadow-xs">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 class="text-sm font-black tracking-tighter italic uppercase">运费试算器</h2>
          <p class="mt-1 text-xs text-muted-foreground">
            输入国家和商品/SKU，系统从数据库读取真实价格与 SKU 重量，再走后端报价规则。
          </p>
        </div>
        <Badge variant="outline" class="w-fit border-blue-200 bg-blue-50 text-blue-700">QUOTE API</Badge>
      </div>

      <form class="mt-5 space-y-4" @submit.prevent="submitQuote">
        <div class="grid gap-3 sm:grid-cols-2">
          <AdminFormField label="国家/地区代码" required :error="errors.country">
            <Input
              v-model.trim="form.country"
              class="font-mono uppercase"
              placeholder="US"
              maxlength="8"
              @input="clearError('country')"
            />
          </AdminFormField>

          <AdminFormField label="币种">
            <Input v-model.trim="form.currency" class="font-mono uppercase" placeholder="USD" maxlength="10" />
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
          <h2 class="text-sm font-black tracking-tighter italic uppercase">报价结果</h2>
          <p class="mt-1 text-xs text-muted-foreground">
            明细会显示命中的模板、线路候选、SKU 实重、包装重量、计费重量和分摊运费，方便排查绑定和规则矩阵。
          </p>
        </div>
        <Badge v-if="quote" variant="outline" :class="quote.free_shipping ? 'border-emerald-200 bg-emerald-50 text-emerald-700' : 'border-amber-200 bg-amber-50 text-amber-700'">
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
            <p class="mt-1 text-2xl font-black tracking-tighter">{{ formatMoney(quote.shipping_fee) }}</p>
          </div>
          <div class="rounded-lg border bg-muted/35 p-3">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">币种</span>
            <p class="mt-1 text-2xl font-black tracking-tighter">{{ quote.currency || 'USD' }}</p>
          </div>
          <div class="rounded-lg border bg-muted/35 p-3">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">免运</span>
            <p class="mt-1 text-2xl font-black tracking-tighter">{{ quote.free_shipping ? '是' : '否' }}</p>
          </div>
          <div class="rounded-lg border bg-muted/35 p-3">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">默认线路</span>
            <p class="mt-1 truncate text-sm font-black tracking-tighter">{{ selectedOptionLabel(quote.selected_option) }}</p>
            <p class="mt-1 text-[10px] text-muted-foreground">{{ quote.source === 'carrier_service' ? '线路服务报价' : '模板基础报价' }}</p>
          </div>
        </div>

        <div class="rounded-lg border bg-card p-3">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-xs font-black uppercase tracking-wider">可选线路</h3>
              <p class="mt-1 text-[10px] text-muted-foreground">
                只展示已启用、国家/币种匹配、且绑定当前模板的线路；体积重线路必须有包装尺寸才会进入候选。
              </p>
            </div>
            <Badge variant="outline" class="w-fit">{{ quote.options?.length || 0 }} OPTIONS</Badge>
          </div>

          <div v-if="!quote.options?.length" class="mt-3 rounded-lg border border-dashed p-4 text-xs text-muted-foreground">
            当前没有可用线路服务，系统仍返回模板基础报价。请检查：线路是否启用、是否绑定同一个运费模板、国家/币种是否匹配、体积重线路是否已配置包装尺寸。
          </div>

          <AdminTablePanel v-else :loading="false" class="mt-3">
            <Table class="min-w-[1080px]">
              <TableHeader>
                <TableRow>
                  <TableHead>线路服务</TableHead>
                  <TableHead class="w-32">计费模式</TableHead>
                  <TableHead class="w-44 text-right">重量口径</TableHead>
                  <TableHead class="w-40 text-right">费用拆分</TableHead>
                  <TableHead class="w-28 text-right">总运费</TableHead>
                  <TableHead class="w-28 text-right">时效</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="option in quote.options" :key="option.carrier_service_id">
                  <TableCell>
                    <div class="flex items-start gap-2">
                      <Badge
                        v-if="isSelectedOption(option)"
                        variant="outline"
                        class="mt-0.5 border-emerald-200 bg-emerald-50 text-[10px] text-emerald-700"
                      >
                        默认
                      </Badge>
                      <div class="min-w-0">
                        <span class="block truncate text-xs font-bold">{{ option.service_name || '-' }}</span>
                        <span class="block truncate font-mono text-[10px] text-muted-foreground">
                          {{ option.carrier_name || '-' }} · {{ option.service_code || '-' }} · template_id={{ option.template_id || '-' }}
                        </span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>{{ billingModeLabel(option.billing_mode) }}</TableCell>
                  <TableCell class="text-right font-mono text-[10px] text-muted-foreground">
                    实重 {{ formatGrams(option.actual_weight_grams) }}<br />
                    体积 {{ formatGrams(option.volumetric_weight_grams) }}<br />
                    计费 {{ formatGrams(option.billable_weight_grams) }}
                  </TableCell>
                  <TableCell class="text-right font-mono text-[10px] text-muted-foreground">
                    base {{ formatMoney(option.base_fee) }}<br />
                    fuel {{ formatMoney(option.fuel_surcharge) }} · remote {{ formatMoney(option.remote_surcharge) }}
                  </TableCell>
                  <TableCell class="text-right text-sm font-black tabular-nums">{{ formatMoney(option.shipping_fee) }}</TableCell>
                  <TableCell class="text-right text-xs tabular-nums">{{ formatEta(option) }}</TableCell>
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
                  <span class="block font-bold text-xs">{{ item.template_name || '-' }}</span>
                  <span class="block text-[10px] text-muted-foreground">template_id={{ item.template_id || '-' }}</span>
                </TableCell>
                <TableCell class="text-right tabular-nums">{{ item.quantity || 0 }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatMoney(item.unit_price) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatGrams(item.weight_grams) }}</TableCell>
                <TableCell>
                  <span class="block text-xs font-bold">{{ item.packaging_rule_name || '未绑定包装' }}</span>
                  <span class="block text-[10px] text-muted-foreground">
                    {{ item.packaging_rule_id ? `rule_id=${item.packaging_rule_id}` : '按 SKU 实重计费' }}
                    · {{ formatGrams(item.packaging_weight_grams) }}
                  </span>
                </TableCell>
                <TableCell class="text-right tabular-nums">{{ formatGrams(item.charge_weight_grams || item.weight_grams) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatMoney(item.shipping_fee) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </div>
    </div>
  </section>
</template>

<script setup>
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

const form = reactive({
  country: 'US',
  currency: 'USD',
  items: [
    {
      product_id: '',
      variant_id: '',
      quantity: 1,
    },
  ],
})

const errors = reactive({})
const quote = ref(null)
const quoteError = ref('')
const submitting = ref(false)

const defaultItem = () => ({
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

const itemErrors = (index) => ({
  product_id: errors[itemErrorKey(index, 'product_id')],
  quantity: errors[itemErrorKey(index, 'quantity')],
})

const clearItemError = (index, field) => {
  delete errors[itemErrorKey(index, field)]
}

const validate = () => {
  Object.keys(errors).forEach((key) => delete errors[key])
  if (!form.country?.trim()) errors.country = '请输入国家/地区代码'

  form.items.forEach((item, index) => {
    if (!Number(item.product_id || 0)) errors[itemErrorKey(index, 'product_id')] = '请输入 Product ID'
    if (Number(item.quantity || 0) <= 0) errors[itemErrorKey(index, 'quantity')] = '数量必须大于 0'
  })

  return Object.keys(errors).length === 0
}

const buildPayload = () => ({
  country: form.country.trim().toUpperCase(),
  currency: form.currency?.trim().toUpperCase() || 'USD',
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

const formatMoney = (value) => Number(value || 0).toFixed(2)
const formatGrams = (value) => `${Number(value || 0).toLocaleString()} g`
const selectedOptionLabel = (option) => {
  if (!option) return '未命中线路'
  return [option.carrier_name, option.service_name].filter(Boolean).join(' / ') || `Service #${option.carrier_service_id || '-'}`
}
const billingModeLabel = (mode) => ({
  actual_weight: '实重计费',
  volumetric_weight: '体积重计费',
  greater_of_actual_and_volumetric: '实重/体积重取大',
}[mode] || mode || '-')
const formatEta = (option) => {
  const min = Number(option?.eta_min_days || 0)
  const max = Number(option?.eta_max_days || 0)
  if (min > 0 && max > 0) return min === max ? `${min} 天` : `${min}-${max} 天`
  if (min > 0) return `${min}+ 天`
  if (max > 0) return `${max} 天内`
  return '-'
}
const isSelectedOption = (option) => Number(option?.carrier_service_id || 0) === Number(quote.value?.selected_option?.carrier_service_id || 0)
</script>
