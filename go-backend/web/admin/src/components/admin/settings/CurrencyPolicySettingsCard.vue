<template>
  <div class="space-y-5">
    <section class="overflow-hidden rounded-[28px] border border-emerald-500/20 bg-emerald-50/70 p-5 shadow-[0_18px_40px_rgb(15_23_42_/_0.06)]">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="flex min-w-0 items-start gap-3">
          <span class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-emerald-600 text-white shadow-[0_12px_24px_rgb(5_150_105_/_0.22)]">
            <Coins class="size-5" />
          </span>
          <div class="min-w-0">
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-700/70">Currency &amp; FX Center</p>
            <h2 class="mt-1 text-xl font-black tracking-tight text-slate-950">后台录入币种与汇率缓存</h2>
            <p class="mt-1 max-w-2xl text-xs leading-relaxed text-slate-600">
              后台只维护一个商品录入币种；前台地区展示币种来自“市场与本地化语种”TAB，汇率 API 每日同步这些市场目标。
            </p>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:justify-end">
          <div class="rounded-2xl border border-emerald-500/15 bg-white/75 px-3 py-2">
            <p class="text-[9px] font-black uppercase tracking-widest text-slate-500">录入币种</p>
            <p class="mt-1 font-mono text-sm font-black text-slate-950">{{ primaryCurrency }}</p>
          </div>
          <div class="rounded-2xl border border-emerald-500/15 bg-white/75 px-3 py-2">
            <p class="text-[9px] font-black uppercase tracking-widest text-slate-500">市场目标</p>
            <p class="mt-1 text-sm font-black text-slate-950">{{ rateTargetCurrencies.length }} 个</p>
          </div>
          <div class="rounded-2xl border border-emerald-500/15 bg-white/75 px-3 py-2">
            <p class="text-[9px] font-black uppercase tracking-widest text-slate-500">缓存状态</p>
            <p class="mt-1 text-sm font-black text-slate-950">{{ cachedRateCount }}/{{ rateTargetCurrencies.length }}</p>
          </div>
        </div>
      </div>
    </section>

    <section class="uds-card-box space-y-5 p-4 sm:p-5">
      <div class="flex flex-col gap-3 border-b border-dashed border-border/80 pb-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground/60">Live Rate Board</p>
          <h3 class="mt-1 text-base font-black tracking-tight text-foreground">当前汇率看板</h3>
          <p class="mt-1 text-xs text-muted-foreground">
            Base 使用后台录入币种 {{ primaryCurrency }}；Quote targets 来自启用市场的展示币种配置。
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            :disabled="loading || saving || exchangeRateLoading || auditLoading"
            title="重新读取币种策略、市场目标和汇率缓存"
            @click="loadPolicy"
          >
            <RefreshCw :class="['size-3.5', loading || exchangeRateLoading || auditLoading ? 'animate-spin' : '']" />
            刷新
          </Button>
          <Button
            v-if="canEdit"
            type="button"
            size="sm"
            :disabled="syncRateDisabled"
            :title="syncRateDisabledReason"
            @click="syncExchangeRates"
          >
            <LoaderCircle v-if="syncingRates" class="size-3.5 animate-spin" />
            <RefreshCw v-else class="size-3.5" />
            {{ syncingRates ? '同步中' : '同步汇率' }}
          </Button>
        </div>
      </div>

      <div v-if="loading" class="flex min-h-56 items-center justify-center text-xs text-muted-foreground">
        <LoaderCircle class="mr-2 size-4 animate-spin" />
        正在读取币种策略
      </div>

      <div v-else-if="currencyCards.length === 1" class="rounded-2xl border border-dashed border-border bg-muted/20 px-4 py-10 text-center">
        <Coins class="mx-auto size-8 text-muted-foreground/50" />
        <p class="mt-3 text-sm font-black text-foreground">暂无市场展示币种目标</p>
        <p class="mt-1 text-xs text-muted-foreground">请在“设置 / 市场与本地化语种”TAB 为启用市场配置默认展示币种和展示币种。</p>
      </div>

      <div v-else class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <article
          v-for="currency in currencyCards"
          :key="currency.code"
          class="relative flex min-h-56 flex-col overflow-hidden rounded-2xl border p-4 transition-shadow hover:shadow-[0_14px_28px_rgb(15_23_42_/_0.08)]"
          :class="currency.isBase
            ? 'border-emerald-300 bg-emerald-50/80'
            : 'border-dashed border-border/90 bg-card'"
        >
          <div v-if="currency.isBase" class="absolute inset-x-0 top-0 h-1 bg-emerald-600" />

          <div class="flex items-start gap-3">
            <span
              class="flex size-9 shrink-0 items-center justify-center rounded-xl border"
              :class="currency.isBase ? 'border-emerald-300 bg-emerald-600 text-white' : 'border-border bg-muted text-foreground'"
            >
              <Coins class="size-4" />
            </span>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h4 class="truncate text-sm font-black text-foreground">{{ currency.name }}</h4>
                <span class="font-mono text-xs font-black text-muted-foreground">({{ currency.code }})</span>
              </div>
              <p class="mt-1 text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
                精度：{{ currency.minor_units }} 位小数
              </p>
            </div>
          </div>

          <div class="mt-auto pt-8">
            <p class="text-[10px] font-black uppercase tracking-[0.16em] text-muted-foreground/65">
              {{ currency.isBase ? '后台录入币种' : '当前汇率' }}
            </p>
            <p class="mt-1 font-mono text-3xl font-black tracking-tight text-foreground">
              {{ currency.isBase ? '1.0000' : rateValueLabel(currency.code) }}
            </p>
            <p class="mt-1 font-mono text-[11px] text-muted-foreground">
              1 {{ primaryCurrency }} = {{ currency.isBase ? '1.0000' : rateValueLabel(currency.code) }} {{ currency.code }}
            </p>
          </div>

          <div class="mt-4 flex items-center justify-between gap-2 border-t border-border/60 pt-3">
            <span class="text-[10px] text-muted-foreground">
              {{ currency.isBase ? 'Entry currency' : rateFetchedAtLabel(currency.code) }}
            </span>
            <span
              class="rounded-full border px-2 py-0.5 text-[10px] font-black"
              :class="currency.isBase ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700' : rateStatusClass(currency.code)"
            >
              {{ currency.isBase ? '录入基准' : rateStatusLabel(currency.code) }}
            </span>
          </div>
        </article>
      </div>

      <div class="flex flex-col gap-2 rounded-2xl border border-sky-500/20 bg-sky-50/70 px-4 py-3 text-xs text-sky-900 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex items-start gap-2">
          <Info class="mt-0.5 size-4 shrink-0 text-sky-700" />
          <p class="leading-relaxed">
            汇率 API 由后端代理并写入缓存，前台不会暴露第三方地址或 API Key；同步不会修改商品主金额。
          </p>
        </div>
        <span class="shrink-0 font-mono text-[10px] text-sky-800/70">
          {{ latestFetchedAtLabel }}
        </span>
      </div>
    </section>

    <section class="rounded-2xl border bg-card p-4 sm:p-5">
      <div class="flex flex-col gap-3 border-b border-dashed border-border/80 pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground/60">Backend Entry Currency</p>
          <h3 class="mt-1 text-base font-black tracking-tight text-foreground">后台默认录入币种</h3>
          <p class="mt-1 text-xs text-muted-foreground">
            商品、SKU、运费和商业金额统一按这个币种录入；前台展示币种请在市场 TAB 配置。
          </p>
        </div>
        <Button v-if="canEdit" type="button" :disabled="saving || loading" @click="savePolicy">
          <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          {{ saving ? '保存中' : '保存录入币种' }}
        </Button>
      </div>

      <div v-if="!loading" class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.4fr)]">
        <AdminFormField label="后台录入币种" description="例如选择 CNY，则商品和 SKU 原始价格都应按人民币录入。">
          <Select v-model="primaryCurrencyInput" :disabled="!canEdit || saving">
            <SelectTrigger>
              <SelectValue placeholder="选择后台录入币种" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="option in catalog" :key="option.code" :value="option.code">
                {{ option.code }} · {{ option.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </AdminFormField>

        <div class="rounded-2xl border border-dashed border-border bg-muted/20 p-4 text-xs leading-relaxed text-muted-foreground">
          <p class="font-black text-foreground">市场展示币种不在这里新增。</p>
          <p class="mt-1">
            当前汇率目标：<span class="font-mono font-black text-foreground">{{ rateTargetLabel }}</span>
          </p>
          <p class="mt-1">
            如果把后台录入币种从 CNY 改为 USD，历史商品仍保留原 currency/price/sale_price，下方检测会提示哪些商品需要人工修正。
          </p>
        </div>
      </div>

      <div class="mt-5 grid gap-3 sm:grid-cols-3">
        <div class="rounded-xl border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Base</p>
          <p class="mt-1 font-mono text-sm font-black text-foreground">{{ primaryCurrency }}</p>
        </div>
        <div class="rounded-xl border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Market targets</p>
          <p class="mt-1 text-sm font-black text-foreground">{{ rateTargetCurrencies.length }} 个</p>
        </div>
        <div class="rounded-xl border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Refresh policy</p>
          <p class="mt-1 text-sm font-black text-foreground">每日自动 / 手动同步</p>
        </div>
      </div>

      <div class="mt-5 rounded-2xl border p-4" :class="currencyAuditHasMismatch ? 'border-amber-500/25 bg-amber-50/70' : 'border-emerald-500/20 bg-emerald-50/60'">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div class="flex items-start gap-3">
            <span
              class="flex size-9 shrink-0 items-center justify-center rounded-xl"
              :class="currencyAuditHasMismatch ? 'bg-amber-500/10 text-amber-700' : 'bg-emerald-500/10 text-emerald-700'"
            >
              <AlertTriangle v-if="currencyAuditHasMismatch" class="size-4" />
              <ShieldCheck v-else class="size-4" />
            </span>
            <div>
              <p class="text-sm font-black text-foreground">商品录入币种一致性检测</p>
              <p v-if="auditLoading" class="mt-1 text-xs text-muted-foreground">正在检测商品和 SKU 币种...</p>
              <p v-else-if="currencyAuditHasMismatch" class="mt-1 text-xs leading-relaxed text-amber-900/80">
                发现 {{ currencyAudit?.total_mismatch_count || 0 }} 条商品/SKU 币种与当前后台录入币种 {{ auditExpectedCurrency }} 不一致，请人工确认金额后再修正币种。
              </p>
              <p v-else class="mt-1 text-xs leading-relaxed text-emerald-900/80">
                当前商品和 SKU 币种都与后台录入币种 {{ auditExpectedCurrency }} 一致。
              </p>
            </div>
          </div>
          <Button type="button" variant="outline" size="sm" :disabled="auditLoading" @click="loadCurrencyAudit">
            <RefreshCw :class="['size-3.5', auditLoading ? 'animate-spin' : '']" />
            重新检测
          </Button>
        </div>

        <div v-if="currencyAuditHasMismatch" class="mt-4 overflow-hidden rounded-xl border border-amber-500/20 bg-white/75">
          <table class="w-full min-w-[560px] text-left text-xs">
            <thead class="border-b bg-amber-50 text-[10px] font-black uppercase tracking-widest text-amber-900/70">
              <tr>
                <th class="px-3 py-2">类型</th>
                <th class="px-3 py-2">ID</th>
                <th class="px-3 py-2">SKU</th>
                <th class="px-3 py-2">名称</th>
                <th class="px-3 py-2">当前币种</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="sample in currencyAudit?.samples || []" :key="`${sample.kind}-${sample.id}`" class="border-b last:border-b-0">
                <td class="px-3 py-2 font-black">{{ sample.kind === 'variant' ? 'SKU' : '商品' }}</td>
                <td class="px-3 py-2 font-mono">{{ sample.id }}</td>
                <td class="px-3 py-2 font-mono">{{ sample.sku || '-' }}</td>
                <td class="px-3 py-2">{{ sample.name || sample.title || '-' }}</td>
                <td class="px-3 py-2 font-mono font-black text-amber-800">{{ sample.currency }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { AlertTriangle, Coins, Info, LoaderCircle, RefreshCw, Save, ShieldCheck } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import axios from '@/utils/axios'

interface CurrencyCatalogOption {
  code: string
  name: string
  minor_units?: number
}

interface CurrencyPolicyState {
  primary_currency: string
  display_currencies: string[]
  available_currencies: CurrencyCatalogOption[]
}

interface CurrencyPolicyResponse {
  primary_currency?: string
  display_currencies?: unknown
  available_currencies?: unknown
}

interface ExchangeRateConfig {
  enabled?: boolean
  api_key_set?: boolean
  base_currency?: string
  quote_currencies?: unknown
}

interface ExchangeRateRecord {
  base_currency?: string
  quote_currency?: string
  rate?: number
  source?: string
  fetched_at?: string
  expires_at?: string
}

interface CurrencyAuditSample {
  kind?: string
  id?: number | string
  product_id?: number | string
  sku?: string
  name?: string
  title?: string
  currency?: string
}

interface CurrencyAudit {
  expected_currency?: string
  product_mismatch_count?: number
  variant_mismatch_count?: number
  total_mismatch_count?: number
  samples?: CurrencyAuditSample[]
}

type UnknownRecord = Record<string, unknown>

const props = withDefaults(defineProps<{
  canEdit?: boolean
}>(), {
  canEdit: false,
})

const emit = defineEmits<{
  saved: [policy: CurrencyPolicyState]
}>()

const canEdit = computed(() => props.canEdit)
const loading = ref(false)
const saving = ref(false)
const exchangeRateLoading = ref(false)
const auditLoading = ref(false)
const syncingRates = ref(false)
const primaryCurrencyInput = ref('USD')
const exchangeRateConfig = ref<ExchangeRateConfig | null>(null)
const exchangeRates = ref<ExchangeRateRecord[]>([])
const currencyAudit = ref<CurrencyAudit | null>(null)
const policy = reactive<CurrencyPolicyState>({
  primary_currency: 'USD',
  display_currencies: [],
  available_currencies: [],
})

const normalizeCurrencyCode = (value: unknown): string => String(value || '').trim().toUpperCase()
const uniqueCurrencyCodes = (values: unknown): string[] => {
  const seen = new Set<string>()
  const list = Array.isArray(values) ? values : []
  return list
    .map(normalizeCurrencyCode)
    .filter((code) => /^[A-Z]{3}$/.test(code))
    .filter((code) => {
      if (seen.has(code)) return false
      seen.add(code)
      return true
    })
}

const asRecord = (value: unknown): UnknownRecord =>
  value && typeof value === 'object' ? value as UnknownRecord : {}

const responsePayload = (response: { data?: unknown }): UnknownRecord => {
  const data = asRecord(response.data)
  const nested = asRecord(data.data)
  const doubleNested = asRecord(nested.data)
  if (Object.keys(doubleNested).length > 0) return doubleNested
  if (Object.keys(nested).length > 0) return nested
  return data
}

const errorResponseMessage = (error: unknown, fallback: string): string => {
  const response = asRecord(asRecord(error).response)
  const data = asRecord(response.data)
  return String(data.message || data.error || fallback)
}

const normalizeCatalog = (value: unknown): CurrencyCatalogOption[] => {
  if (!Array.isArray(value)) return []
  return value
    .map((option) => asRecord(option))
    .map((option) => ({
      code: normalizeCurrencyCode(option.code),
      name: String(option.name || option.code || ''),
      minor_units: Number.isFinite(Number(option.minor_units)) ? Number(option.minor_units) : 2,
    }))
    .filter((option) => /^[A-Z]{3}$/.test(option.code))
}

const catalog = computed(() => policy.available_currencies || [])
const primaryCurrency = computed(() => normalizeCurrencyCode(primaryCurrencyInput.value) || 'USD')
const rateTargetCurrencies = computed(() => uniqueCurrencyCodes(exchangeRateConfig.value?.quote_currencies).filter((code) => code !== primaryCurrency.value))
const rateTargetLabel = computed(() => rateTargetCurrencies.value.length ? rateTargetCurrencies.value.join(' / ') : '暂无市场展示币种目标')
const exchangeRateApiEnabled = computed(() => Boolean(exchangeRateConfig.value?.enabled))
const exchangeRateApiKeySet = computed(() => Boolean(exchangeRateConfig.value?.api_key_set))
const auditExpectedCurrency = computed(() => normalizeCurrencyCode(currencyAudit.value?.expected_currency) || primaryCurrency.value)
const currencyAuditHasMismatch = computed(() => Number(currencyAudit.value?.total_mismatch_count || 0) > 0)

const ratesByQuoteCurrency = computed<Map<string, ExchangeRateRecord>>(() => {
  const result = new Map<string, ExchangeRateRecord>()
  for (const rate of exchangeRates.value || []) {
    const quoteCurrency = normalizeCurrencyCode(rate?.quote_currency)
    if (quoteCurrency) result.set(quoteCurrency, rate)
  }
  return result
})

const currencyOption = (code: string): CurrencyCatalogOption => (
  catalog.value.find((option) => option.code === code) || {
    code,
    name: code,
    minor_units: 2,
  }
)

const currencyCards = computed(() => [
  {
    ...currencyOption(primaryCurrency.value),
    code: primaryCurrency.value,
    isBase: true,
  },
  ...rateTargetCurrencies.value.map((code) => ({
    ...currencyOption(code),
    code,
    isBase: false,
  })),
])

const cachedRateCount = computed(() => rateTargetCurrencies.value.filter((code) => Boolean(ratesByQuoteCurrency.value.get(code))).length)

const latestFetchedAtLabel = computed(() => {
  const dates = exchangeRates.value
    .map((rate) => rate.fetched_at)
    .filter(Boolean)
    .map((value) => new Date(String(value)).getTime())
    .filter((value) => Number.isFinite(value))
  if (!dates.length) return '尚无同步记录'
  return `最近同步 ${new Date(Math.max(...dates)).toLocaleString()}`
})

const syncRateDisabled = computed(() => (
  !canEdit.value ||
  loading.value ||
  saving.value ||
  exchangeRateLoading.value ||
  syncingRates.value ||
  rateTargetCurrencies.value.length === 0 ||
  !exchangeRateApiEnabled.value ||
  !exchangeRateApiKeySet.value
))

const syncRateDisabledReason = computed(() => {
  if (!canEdit.value) return '当前账号不能编辑设置'
  if (loading.value || saving.value || exchangeRateLoading.value || syncingRates.value) return '正在处理设置'
  if (rateTargetCurrencies.value.length === 0) return '请先在“市场与本地化语种”TAB 配置展示币种'
  if (!exchangeRateApiEnabled.value) return '请先在“汇率 API”页启用接口'
  if (!exchangeRateApiKeySet.value) return '请先在“汇率 API”页保存 API Key'
  return '同步启用市场展示币种的汇率缓存'
})

const applyPolicy = (next: CurrencyPolicyResponse = {}) => {
  const nextPrimaryCurrency = normalizeCurrencyCode(next.primary_currency) || 'USD'
  Object.assign(policy, {
    primary_currency: nextPrimaryCurrency,
    display_currencies: [],
    available_currencies: normalizeCatalog(next.available_currencies),
  })
  primaryCurrencyInput.value = policy.primary_currency
}

const applyExchangeRatePayload = (payload: UnknownRecord) => {
  const config = payload.config
  exchangeRateConfig.value = config && typeof config === 'object' ? config as ExchangeRateConfig : null
  exchangeRates.value = Array.isArray(payload.rates) ? payload.rates as ExchangeRateRecord[] : []
}

const applyCurrencyAuditPayload = (payload: UnknownRecord) => {
  const audit = payload.audit
  currencyAudit.value = audit && typeof audit === 'object' ? audit as CurrencyAudit : null
}

const loadExchangeRates = async () => {
  exchangeRateLoading.value = true
  try {
    const response = await axios.get('/api/admin/settings/exchange-rates')
    applyExchangeRatePayload(responsePayload(response))
  } catch (error) {
    toast.error(errorResponseMessage(error, '汇率缓存读取失败'))
    exchangeRateConfig.value = null
    exchangeRates.value = []
  } finally {
    exchangeRateLoading.value = false
  }
}

const loadCurrencyAudit = async () => {
  auditLoading.value = true
  try {
    const response = await axios.get('/api/admin/settings/currency-policy/audit')
    applyCurrencyAuditPayload(responsePayload(response))
  } catch (error) {
    toast.error(errorResponseMessage(error, '商品币种一致性检测失败'))
    currencyAudit.value = null
  } finally {
    auditLoading.value = false
  }
}

const loadPolicy = async () => {
  loading.value = true
  try {
    const [policyResponse] = await Promise.all([
      axios.get('/api/admin/settings/currency-policy'),
      loadExchangeRates(),
      loadCurrencyAudit(),
    ])
    applyPolicy(asRecord(policyResponse.data).policy as CurrencyPolicyResponse || {})
  } catch (error) {
    toast.error(errorResponseMessage(error, '价格币种策略读取失败'))
  } finally {
    loading.value = false
  }
}

const persistCurrencyPolicy = async () => {
  const response = await axios.put('/api/admin/settings/currency-policy', {
    primary_currency: primaryCurrency.value,
  })
  const payload = responsePayload(response)
  const nextPolicy = asRecord(payload.policy) as CurrencyPolicyResponse || {}
  applyPolicy(nextPolicy)
  applyCurrencyAuditPayload(payload)
  emit('saved', {
    primary_currency: policy.primary_currency,
    display_currencies: [],
    available_currencies: [...policy.available_currencies],
  })
}

const savePolicy = async () => {
  saving.value = true
  try {
    await persistCurrencyPolicy()
    await Promise.all([loadExchangeRates(), loadCurrencyAudit()])
    if (currencyAuditHasMismatch.value) {
      toast.warning('后台录入币种已保存，但存在商品/SKU 币种不一致，请查看检测结果')
    } else {
      toast.success('后台录入币种已保存')
    }
  } catch (error) {
    toast.error(errorResponseMessage(error, '后台录入币种保存失败'))
  } finally {
    saving.value = false
  }
}

const syncExchangeRates = async () => {
  if (syncRateDisabled.value) return
  syncingRates.value = true
  try {
    await persistCurrencyPolicy()
    const response = await axios.post('/api/admin/settings/exchange-rates/sync')
    const payload = responsePayload(response)
    const rates = Array.isArray(payload.rates) ? payload.rates : []
    await Promise.all([loadExchangeRates(), loadCurrencyAudit()])
    toast.success(`汇率缓存已同步：${rates.length} 个币种`)
  } catch (error) {
    toast.error(errorResponseMessage(error, '汇率同步失败'))
  } finally {
    syncingRates.value = false
  }
}

const exchangeRateForCurrency = (currency: string): ExchangeRateRecord | null =>
  ratesByQuoteCurrency.value.get(normalizeCurrencyCode(currency)) || null

const formatRate = (value: unknown): string => {
  const rate = Number(value)
  if (!Number.isFinite(rate) || rate <= 0) return '未缓存'
  if (rate >= 100) return rate.toFixed(2)
  if (rate >= 1) return rate.toFixed(4)
  return rate.toPrecision(5)
}

const rateValueLabel = (currency: string): string => formatRate(exchangeRateForCurrency(currency)?.rate)

const formatDateTime = (value?: string): string => {
  if (!value) return '尚未同步'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '尚未同步'
  return date.toLocaleString()
}

const rateFetchedAtLabel = (currency: string): string => formatDateTime(exchangeRateForCurrency(currency)?.fetched_at)

const rateExpired = (rate: ExchangeRateRecord | null): boolean => {
  if (!rate?.expires_at) return false
  const expiresAt = new Date(rate.expires_at).getTime()
  return Number.isFinite(expiresAt) && expiresAt <= Date.now()
}

const rateStatusLabel = (currency: string): string => {
  const rate = exchangeRateForCurrency(currency)
  if (!rate) return '未缓存'
  if (rateExpired(rate)) return '已过期'
  return '已缓存'
}

const rateStatusClass = (currency: string): string => {
  const rate = exchangeRateForCurrency(currency)
  if (!rate) return 'border-amber-500/20 bg-amber-500/10 text-amber-700'
  if (rateExpired(rate)) return 'border-rose-500/20 bg-rose-500/10 text-rose-700'
  return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700'
}

onMounted(loadPolicy)
</script>
