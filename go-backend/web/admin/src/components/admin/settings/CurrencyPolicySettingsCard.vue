<template>
  <section class="rounded-2xl border bg-muted/30 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Pricing Currency</p>
        <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">价格币种</h3>
        <p class="mt-1 max-w-2xl text-xs leading-relaxed text-muted-foreground">
          这里定义后台录入商品、SKU、运费和订单金额的主基准币种，并添加用于 Nuxt 展示的次展示币种；汇率同步时后端直接读取这里。
        </p>
      </div>
      <div class="flex items-center gap-2">
        <span class="rounded-full border border-admin-selected-border bg-admin-selected-soft px-2.5 py-1 text-[11px] font-black text-admin-selected">
          {{ selectedDisplayCurrencies.length }} 个
        </span>
        <Button type="button" variant="outline" size="sm" :disabled="loading || saving" @click="loadPolicy">
          <RefreshCw :class="['size-3.5', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEdit" type="button" size="sm" :disabled="loading || saving" @click="savePolicy">
          <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          {{ saving ? '保存中' : '保存主基准与次币种' }}
        </Button>
        <Button
          v-if="canEdit"
          type="button"
          variant="outline"
          size="sm"
          :disabled="syncRateDisabled"
          :title="syncRateDisabledReason"
          @click="syncExchangeRates"
        >
          <RefreshCw :class="['size-3.5', syncingRates ? 'animate-spin' : '']" />
          {{ syncingRates ? '同步中' : '同步汇率缓存' }}
        </Button>
      </div>
    </div>

    <div v-if="loading" class="mt-5 flex h-28 items-center justify-center text-xs text-muted-foreground">
      <LoaderCircle class="mr-2 size-4 animate-spin" />
      正在读取价格币种策略
    </div>

    <div v-else class="mt-5 space-y-5">
      <AdminFormField label="主基准币种" description="后台商品、SKU、运费和订单金额的录入基准币种；支付按钮不由这里决定。">
        <Select v-model="primaryCurrencyInput" :disabled="!canEdit">
          <SelectTrigger class="w-full"><SelectValue placeholder="选择主基准币种" /></SelectTrigger>
          <SelectContent>
            <SelectItem v-for="option in catalog" :key="option.code" :value="option.code">
              {{ option.code }} · {{ option.name }}
            </SelectItem>
          </SelectContent>
        </Select>
      </AdminFormField>

      <AdminFormField label="次展示币种" description="这些币种是汇率缓存的同步目标，用于后台一键填充商品、SKU 和运费展示价；不作为支付网关白名单。">
        <div class="space-y-3">
          <Input
            v-model="displayCurrencyInput"
            class="font-mono uppercase"
            placeholder="USD, EUR, CNY"
            :disabled="!canEdit"
            @blur="normalizeDisplayCurrencyInput"
          />
          <div class="flex flex-wrap gap-2">
            <button
              v-for="option in secondaryCatalog"
              :key="`display-${option.code}`"
              type="button"
              class="rounded-full border px-3 py-1.5 text-xs font-black transition hover:border-admin-selected-border hover:bg-admin-selected-soft disabled:cursor-not-allowed disabled:opacity-45"
              :class="isDisplayCurrencySelected(option.code) ? 'border-admin-selected-border bg-admin-selected-soft text-admin-selected shadow-[var(--admin-control-selected-surface-shadow)]' : 'bg-background/70 text-foreground'"
              :disabled="!canEdit"
              :title="`${option.code} · ${option.name}`"
              @click="toggleDisplayCurrency(option.code)"
            >
              {{ option.code }}
            </button>
          </div>
        </div>
      </AdminFormField>

      <div class="rounded-xl border bg-background/70 p-3 text-xs leading-relaxed text-muted-foreground">
        商品、SKU 和运费录入主基准币种金额；次展示币种按 ExchangeRate-API 缓存汇率一键填充给 Nuxt 展示。用户付款时后端按订单金额和币种发起支付，并用网关回调核对金额、币种、签名和幂等。
      </div>

      <div class="overflow-hidden rounded-xl border bg-background/70">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b bg-muted/40 px-3 py-2.5">
          <div>
            <p class="text-xs font-black text-foreground">当前汇率缓存</p>
            <p class="mt-0.5 text-[11px] text-muted-foreground">
              主基准 {{ exchangeRateBaseCurrency }} -> 次展示币种；这里只展示后台缓存，不给 Nuxt 前台直接调第三方 API。
            </p>
          </div>
          <Button type="button" variant="ghost" size="sm" :disabled="exchangeRateLoading" @click="loadExchangeRates">
            <RefreshCw :class="['size-3.5', exchangeRateLoading ? 'animate-spin' : '']" />
            刷新汇率视图
          </Button>
        </div>

        <div v-if="exchangeRateLoading" class="flex h-24 items-center justify-center text-xs text-muted-foreground">
          <LoaderCircle class="mr-2 size-4 animate-spin" />
          正在读取汇率缓存
        </div>
        <div v-else-if="selectedDisplayCurrencies.length === 0" class="px-3 py-5 text-xs text-muted-foreground">
          先添加次展示币种，保存后汇率 API 会按这些币种维护缓存。
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[720px] text-left text-xs">
            <thead class="border-b bg-muted/60 text-[10px] font-black uppercase tracking-widest text-muted-foreground">
              <tr>
                <th class="px-3 py-2">次展示币种</th>
                <th class="px-3 py-2">缓存汇率</th>
                <th class="px-3 py-2">来源</th>
                <th class="px-3 py-2">更新时间</th>
                <th class="px-3 py-2">状态</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="currency in selectedDisplayCurrencies" :key="`rate-${currency}`" class="border-b last:border-b-0">
                <td class="px-3 py-2 font-mono font-black text-foreground">{{ exchangeRateBaseCurrency }} -> {{ currency }}</td>
                <td class="px-3 py-2 font-mono text-foreground">{{ rateValueLabel(currency) }}</td>
                <td class="px-3 py-2 text-muted-foreground">{{ rateSourceLabel(currency) }}</td>
                <td class="px-3 py-2 text-muted-foreground">{{ rateFetchedAtLabel(currency) }}</td>
                <td class="px-3 py-2">
                  <span class="rounded-full border px-2 py-0.5 text-[10px] font-black" :class="rateStatusClass(currency)">
                    {{ rateStatusLabel(currency) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { LoaderCircle, RefreshCw, Save } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
}

interface ExchangeRateRecord {
  quote_currency?: string
  rate?: number
  source?: string
  fetched_at?: string
  expires_at?: string
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
const syncingRates = ref(false)
const primaryCurrencyInput = ref('USD')
const displayCurrencyInput = ref('')
const exchangeRateConfig = ref<ExchangeRateConfig | null>(null)
const exchangeRates = ref<ExchangeRateRecord[]>([])
const policy = reactive<CurrencyPolicyState>({
  primary_currency: 'USD',
  display_currencies: [],
  available_currencies: [],
})

const catalog = computed<CurrencyCatalogOption[]>(() => policy.available_currencies || [])
const primaryCurrency = computed(() => normalizeCurrencyCode(primaryCurrencyInput.value) || 'USD')
const secondaryCatalog = computed(() => catalog.value.filter((option) => option.code !== primaryCurrency.value))
const selectedDisplayCurrencies = computed(() => splitCurrencyCodes(displayCurrencyInput.value))
const exchangeRateBaseCurrency = computed(() => normalizeCurrencyCode(exchangeRateConfig.value?.base_currency) || primaryCurrency.value)
const exchangeRateApiEnabled = computed(() => Boolean(exchangeRateConfig.value?.enabled))
const exchangeRateApiKeySet = computed(() => Boolean(exchangeRateConfig.value?.api_key_set))
const ratesByQuoteCurrency = computed<Map<string, ExchangeRateRecord>>(() => {
  const result = new Map()
  for (const rate of exchangeRates.value || []) {
    const quoteCurrency = normalizeCurrencyCode(rate?.quote_currency)
    if (quoteCurrency) result.set(quoteCurrency, rate)
  }
  return result
})
const syncRateDisabled = computed(() => (
  !canEdit.value ||
  loading.value ||
  saving.value ||
  syncingRates.value ||
  selectedDisplayCurrencies.value.length === 0 ||
  !exchangeRateApiEnabled.value ||
  !exchangeRateApiKeySet.value
))
const syncRateDisabledReason = computed(() => {
  if (!canEdit.value) return '当前账号不能编辑设置'
  if (loading.value || saving.value || syncingRates.value) return '正在处理设置'
  if (selectedDisplayCurrencies.value.length === 0) return '请先添加次展示币种'
  if (!exchangeRateApiEnabled.value) return '请先在 API 管理启用 ExchangeRate-API'
  if (!exchangeRateApiKeySet.value) return '请先在 API 管理保存 ExchangeRate-API Key'
  return '保存当前主基准币种与次展示币种后同步汇率缓存'
})

const normalizeCurrencyCode = (value: unknown): string => String(value || '').trim().toUpperCase()
const uniqueCurrencyCodes = (values: unknown): string[] => {
  const seen = new Set()
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
const splitCurrencyCodes = (value: unknown): string[] => uniqueCurrencyCodes(String(value || '').split(/[\s,;，；]+/))

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
      minor_units: Number(option.minor_units),
    }))
    .filter((option) => /^[A-Z]{3}$/.test(option.code))
}

const applyPolicy = (next: CurrencyPolicyResponse = {}) => {
  const primaryCurrencyValue = normalizeCurrencyCode(next.primary_currency) || 'USD'
  const displayCurrencies = Array.isArray(next.display_currencies) ? next.display_currencies : []
  Object.assign(policy, {
    primary_currency: primaryCurrencyValue,
    display_currencies: uniqueCurrencyCodes(displayCurrencies),
    available_currencies: normalizeCatalog(next.available_currencies),
  })
  primaryCurrencyInput.value = policy.primary_currency
  displayCurrencyInput.value = removePrimaryCurrency(policy.display_currencies).join(',')
}

const normalizeDisplayCurrencyInput = () => {
  displayCurrencyInput.value = removePrimaryCurrency(selectedDisplayCurrencies.value).join(',')
}

const isDisplayCurrencySelected = (code: string): boolean => selectedDisplayCurrencies.value.includes(code)

const toggleDisplayCurrency = (code: string) => {
  if (normalizeCurrencyCode(code) === primaryCurrency.value) return
  const next = new Set(selectedDisplayCurrencies.value)
  if (next.has(code)) next.delete(code)
  else next.add(code)
  displayCurrencyInput.value = removePrimaryCurrency(Array.from(next)).join(',')
}

const removePrimaryCurrency = (values: unknown): string[] => uniqueCurrencyCodes(values).filter((code) => code !== primaryCurrency.value)

const loadExchangeRates = async () => {
  exchangeRateLoading.value = true
  try {
    const response = await axios.get('/api/admin/settings/exchange-rates')
    const payload = responsePayload(response)
    const config = payload.config
    exchangeRateConfig.value = config && typeof config === 'object' ? config as ExchangeRateConfig : null
    exchangeRates.value = Array.isArray(payload.rates) ? payload.rates as ExchangeRateRecord[] : []
  } catch (error) {
    toast.error(errorResponseMessage(error, '汇率缓存读取失败'))
    exchangeRateConfig.value = null
    exchangeRates.value = []
  } finally {
    exchangeRateLoading.value = false
  }
}

const exchangeRateForCurrency = (currency: string): ExchangeRateRecord | null =>
  ratesByQuoteCurrency.value.get(normalizeCurrencyCode(currency)) || null

const rateValueLabel = (currency: string): string => {
  const rate = exchangeRateForCurrency(currency)
  const value = Number(rate?.rate || 0)
  return value > 0 ? value.toPrecision(8) : '未缓存'
}

const rateSourceLabel = (currency: string): string => exchangeRateForCurrency(currency)?.source || '-'

const formatDateTime = (value?: string): string => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
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
  if (!rate) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-200'
  if (rateExpired(rate)) return 'border-red-500/20 bg-red-500/10 text-red-700 dark:text-red-200'
  return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
}

const persistCurrencyPolicy = async () => {
  const response = await axios.put('/api/admin/settings/currency-policy', {
    primary_currency: primaryCurrency.value,
    display_currencies: removePrimaryCurrency(selectedDisplayCurrencies.value),
  })
  const nextPolicy = asRecord(response.data).policy as CurrencyPolicyResponse || {}
  applyPolicy(nextPolicy)
  emit('saved', {
    primary_currency: policy.primary_currency,
    display_currencies: [...policy.display_currencies],
    available_currencies: [...policy.available_currencies],
  })
}

const loadPolicy = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/settings/currency-policy')
    applyPolicy(asRecord(response.data).policy as CurrencyPolicyResponse || {})
    await loadExchangeRates()
  } catch (error) {
    toast.error(errorResponseMessage(error, '价格币种策略读取失败'))
  } finally {
    loading.value = false
  }
}

const savePolicy = async () => {
  saving.value = true
  try {
    await persistCurrencyPolicy()
    await loadExchangeRates()
    toast.success('主基准币种与次展示币种已保存')
  } catch (error) {
    toast.error(errorResponseMessage(error, '价格币种保存失败'))
  } finally {
    saving.value = false
  }
}

const syncExchangeRates = async () => {
  syncingRates.value = true
  try {
    await persistCurrencyPolicy()
    const response = await axios.post('/api/admin/settings/exchange-rates/sync')
    const payload = responsePayload(response)
    const rates = Array.isArray(payload.rates) ? payload.rates : []
    await loadExchangeRates()
    toast.success(`汇率缓存已同步：${rates.length} 个币种`)
  } catch (error) {
    toast.error(errorResponseMessage(error, '汇率同步失败'))
  } finally {
    syncingRates.value = false
  }
}

watch(primaryCurrencyInput, normalizeDisplayCurrencyInput)
onMounted(loadPolicy)
</script>
