<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增运费模板' : '编辑运费模板' }}</DialogTitle>
          <DialogDescription>
            先按当前后端模型维护区域与重量/数量/金额规则；后续会扩展到承运商线路、SKU 绑定和体积重。
          </DialogDescription>
        </DialogHeader>

        <section class="grid gap-4 lg:grid-cols-4">
          <AdminFormField label="模板名称" required :error="errors.name" class="lg:col-span-2">
            <Input v-model.trim="form.name" placeholder="例如 全球空运标准模板" @input="emit('clear-error', 'name')" />
          </AdminFormField>

          <AdminFormField label="计费类型" required :error="errors.type">
            <Select v-model="form.type" @update:model-value="emit('clear-error', 'type')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择计费类型" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="weight">按重量</SelectItem>
                <SelectItem value="quantity">按数量</SelectItem>
                <SelectItem value="price">按订单金额</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">启用模板 / ENABLED</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后不会参与运费规则计算。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用运费模板" />
          </div>

          <AdminFormField label="默认运费" required :error="errors.default_fee">
            <Input v-model.number="form.default_fee" type="number" min="0" step="0.01" @input="handleDefaultFeeInput" />
          </AdminFormField>

          <AdminFormField label="免运门槛">
            <Input v-model.number="form.free_threshold" type="number" min="0" step="0.01" @input="clearTemplateDisplayPrice('free_threshold')" />
          </AdminFormField>

          <div class="flex items-end justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">开启免运 / FREE SHIPPING</span>
              <p class="mt-0.5 text-xs text-muted-foreground">订单金额达到门槛时返回 0 运费。</p>
            </div>
            <Switch v-model="form.free_shipping" aria-label="开启免运" />
          </div>

          <AdminFormField label="说明" class="lg:col-span-4">
            <Textarea v-model="form.description" class="min-h-20" placeholder="内部说明、适用渠道或注意事项" />
          </AdminFormField>
        </section>

        <section class="rounded-lg border bg-background px-3 py-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="text-xs leading-5 text-muted-foreground">
              运费金额按后台主基准币种 {{ primaryPricingCurrency }} 录入；按钮会读取后台缓存汇率并按金额字段填充次展示价，随模板保存。
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="displayPriceLoading || !hasFillableShippingAmounts()"
              @click="fillShippingDisplayPrices"
            >
              <RefreshCw class="size-3.5" :class="{ 'animate-spin': displayPriceLoading }" />
              {{ displayPriceLoading ? '填充中' : '按汇率填充次币种' }}
            </Button>
            <p v-if="displayPriceError" class="basis-full text-xs font-medium text-destructive">{{ displayPriceError }}</p>
          </div>

          <div v-if="displayPriceRows.length" class="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-3">
            <div v-for="row in displayPriceRows" :key="row.key" class="rounded-md border bg-muted/20 px-2.5 py-2">
              <div class="flex items-center justify-between gap-2">
                <span class="truncate text-xs font-semibold text-foreground">{{ row.label }}</span>
                <span class="font-mono text-[11px] text-muted-foreground">{{ row.baseCurrency }} {{ Number(row.amount || 0).toFixed(2) }}</span>
              </div>
              <div class="mt-1.5 flex flex-wrap gap-1.5">
                <span
                  v-for="price in row.prices"
                  :key="price.quote_currency || price.currency"
                  class="rounded-md border px-1.5 py-0.5 font-mono text-[11px]"
 :class="price.fallback_reason ? 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-200': 'bg-background text-foreground'"
                >
                  {{ formatDisplayPriceResult(price) }}
                </span>
              </div>
            </div>
          </div>
        </section>

        <section class="space-y-3 border-t border-dashed pt-5">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-sm font-black tracking-tighter uppercase text-foreground">规则矩阵</h3>
              <p class="mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">Region 建议填国家/区域代码，例如 US、EU、CN；最大值为 0 表示不设上限。</p>
            </div>
            <Button type="button" variant="outline" size="sm" @click="addRule">
              <Plus class="size-3.5" />
              新增规则
            </Button>
          </div>

          <div v-if="!form.rules?.length" class="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
            暂无规则；没有匹配规则时会使用模板默认运费。
          </div>

          <div v-for="(rule, index) in form.rules" :key="index" class="grid gap-3 rounded-lg border p-3 lg:grid-cols-12">
            <AdminFormField label="Region" class="lg:col-span-2">
              <Input v-model.trim="rule.region" class="font-mono uppercase" placeholder="US" />
            </AdminFormField>
            <AdminFormField label="最小值" class="lg:col-span-2">
              <Input v-model.number="rule.min_value" type="number" min="0" step="0.001" @input="clearRuleDisplayPrice(rule, 'min_value')" />
            </AdminFormField>
            <AdminFormField label="最大值" class="lg:col-span-2">
              <Input v-model.number="rule.max_value" type="number" min="0" step="0.001" @input="clearRuleDisplayPrice(rule, 'max_value')" />
            </AdminFormField>
            <AdminFormField label="运费" class="lg:col-span-2">
              <Input v-model.number="rule.fee" type="number" min="0" step="0.01" @input="clearRuleDisplayPrice(rule, 'fee')" />
            </AdminFormField>
            <AdminFormField label="续费" class="lg:col-span-2">
              <Input v-model.number="rule.additional" type="number" min="0" step="0.01" @input="clearRuleDisplayPrice(rule, 'additional')" />
            </AdminFormField>
            <div class="flex items-end justify-end lg:col-span-2">
              <Button type="button" variant="ghost" size="icon-sm" class="text-destructive hover:text-destructive" @click="removeRule(index)">
                <Trash2 class="size-4" />
                <span class="sr-only">删除规则</span>
              </Button>
            </div>
          </div>
        </section>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存模板' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Plus, RefreshCw, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import axios from '@/utils/axios'

const props = defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])
const displayPriceLoading = ref(false)
const displayPriceError = ref('')
const primaryPricingCurrency = ref('USD')

const displayPriceRows = computed(() => shippingAmountEntries()
  .map(entry => ({
    ...entry,
    baseCurrency: '主币种',
    prices: displayPricesForEntry(entry),
  }))
  .filter(row => row.prices.length))

const ensureRules = () => {
  if (!Array.isArray(props.form.rules)) {
    props.form.rules = []
  }
}

const addRule = () => {
  ensureRules()
  props.form.rules.push({
    region: '',
    min_value: 0,
    max_value: 0,
    fee: 0,
    additional: 0,
    display_price_snapshots: {},
  })
}

const removeRule = (index) => {
  ensureRules()
  props.form.rules.splice(index, 1)
}

const normalizeCurrencyCode = (value) => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

const loadPrimaryPricingCurrencyForShippingDisplayPriceFill = async () => {
  try {
    const response = await axios.get('/api/admin/settings/currency-policy')
    const policy = response.data?.policy || response.data?.data?.policy || {}
    primaryPricingCurrency.value = normalizeCurrencyCode(policy.primary_currency) || 'USD'
  } catch (error) {
    primaryPricingCurrency.value = 'USD'
  }
}

const numericAmount = (value) => {
  const amount = Number(value || 0)
  return Number.isFinite(amount) && amount > 0 ? amount : 0
}

const normalizeDisplayPrices = (values) => {
  const list = Array.isArray(values) ? values : []
  const seen = new Set()
  return list
    .map((price) => {
      const quoteCurrency = normalizeCurrencyCode(price?.quote_currency || price?.currency)
      if (!quoteCurrency || price?.fallback_reason) return null
      return {
        amount: Number(price?.amount || 0),
        currency: quoteCurrency,
        quote_currency: quoteCurrency,
        rate: Number(price?.rate || 0),
        source: String(price?.source || '').trim(),
        converted: price?.converted !== false,
      }
    })
    .filter(Boolean)
    .filter(price => price.amount > 0 && price.converted !== false)
    .filter((price) => {
      if (seen.has(price.currency)) return false
      seen.add(price.currency)
      return true
    })
}

const ensureTemplateDisplaySnapshots = () => {
  if (!props.form.display_price_snapshots || typeof props.form.display_price_snapshots !== 'object' || Array.isArray(props.form.display_price_snapshots)) {
    props.form.display_price_snapshots = {}
  }
  return props.form.display_price_snapshots
}

const ensureRuleDisplaySnapshots = (rule) => {
  if (!rule.display_price_snapshots || typeof rule.display_price_snapshots !== 'object' || Array.isArray(rule.display_price_snapshots)) {
    rule.display_price_snapshots = {}
  }
  return rule.display_price_snapshots
}

const clearTemplateDisplayPrice = (field) => {
  const snapshots = ensureTemplateDisplaySnapshots()
  delete snapshots[field]
}

const clearRuleDisplayPrice = (rule, field) => {
  if (!rule) return
  const snapshots = ensureRuleDisplaySnapshots(rule)
  delete snapshots[field]
}

const handleDefaultFeeInput = () => {
  clearTemplateDisplayPrice('default_fee')
  emit('clear-error', 'default_fee')
}

const shippingAmountEntries = () => {
  const entries = []
  const defaultFee = numericAmount(props.form.default_fee)
  if (defaultFee > 0) entries.push({ key: 'default_fee', target: 'template', field: 'default_fee', label: '默认运费', amount: defaultFee })
  const freeThreshold = numericAmount(props.form.free_threshold)
  if (freeThreshold > 0) entries.push({ key: 'free_threshold', target: 'template', field: 'free_threshold', label: '免运门槛', amount: freeThreshold })

  if (Array.isArray(props.form.rules)) {
    props.form.rules.forEach((rule, index) => {
      const region = String(rule.region || `规则 ${index + 1}`).toUpperCase()
      if (props.form.type === 'price') {
        const minValue = numericAmount(rule.min_value)
        const maxValue = numericAmount(rule.max_value)
        if (minValue > 0) entries.push({ key: `rule_${index}_min_value`, target: 'rule', ruleIndex: index, field: 'min_value', label: `${region} 最小订单金额`, amount: minValue })
        if (maxValue > 0) entries.push({ key: `rule_${index}_max_value`, target: 'rule', ruleIndex: index, field: 'max_value', label: `${region} 最大订单金额`, amount: maxValue })
      }
      const fee = numericAmount(rule.fee)
      if (fee > 0) entries.push({ key: `rule_${index}_fee`, target: 'rule', ruleIndex: index, field: 'fee', label: `${region} 运费`, amount: fee })
      const additional = numericAmount(rule.additional)
      if (additional > 0) entries.push({ key: `rule_${index}_additional`, target: 'rule', ruleIndex: index, field: 'additional', label: `${region} 续费`, amount: additional })
    })
  }
  return entries
}

const hasFillableShippingAmounts = () => shippingAmountEntries().length > 0

const displayPricesForEntry = (entry) => {
  if (entry.target === 'template') {
    return normalizeDisplayPrices(props.form.display_price_snapshots?.[entry.field])
  }
  const rule = props.form.rules?.[entry.ruleIndex]
  return normalizeDisplayPrices(rule?.display_price_snapshots?.[entry.field])
}

const setDisplayPricesForEntry = (entry, prices) => {
  const snapshots = normalizeDisplayPrices(prices)
  if (entry.target === 'template') {
    const target = ensureTemplateDisplaySnapshots()
    if (snapshots.length) target[entry.field] = snapshots
    else delete target[entry.field]
    return snapshots.length
  }

  const rule = props.form.rules?.[entry.ruleIndex]
  if (!rule) return 0
  const target = ensureRuleDisplaySnapshots(rule)
  if (snapshots.length) target[entry.field] = snapshots
  else delete target[entry.field]
  return snapshots.length
}

const fillShippingDisplayPrices = async () => {
  displayPriceError.value = ''
  const entries = shippingAmountEntries()
  if (!entries.length) {
    displayPriceError.value = '请先录入至少一个运费金额'
    return
  }

  displayPriceLoading.value = true
  try {
    const rows = await Promise.all(entries.map(async (entry) => {
      const response = await axios.post('/api/admin/pricing/exchange-rates/convert', {
        amount: entry.amount,
        base_currency: primaryPricingCurrency.value
      })
      const data = response.data?.data || response.data || {}
      return {
        entry,
        prices: Array.isArray(data.prices) ? data.prices : []
      }
    }))
    const filledCount = rows.reduce((count, row) => count + setDisplayPricesForEntry(row.entry, row.prices), 0)
    if (filledCount > 0) {
      toast.success('运费次展示价已按缓存汇率填充')
    } else {
      toast.warning('没有可保存的运费次展示价，请检查缓存汇率')
    }
  } catch (error) {
    const message = error?.response?.data?.message || error?.response?.data?.error || error?.message || '运费次展示价填充失败'
    displayPriceError.value = message
    toast.error(message)
  } finally {
    displayPriceLoading.value = false
  }
}

const formatDisplayPriceResult = (price) => {
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

onMounted(loadPrimaryPricingCurrencyForShippingDisplayPriceFill)

watch(() => props.open, (open) => {
  if (open) loadPrimaryPricingCurrencyForShippingDisplayPriceFill()
})
</script>
