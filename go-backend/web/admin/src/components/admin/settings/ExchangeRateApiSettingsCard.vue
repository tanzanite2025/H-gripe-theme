<template>
  <section class="rounded-2xl border bg-muted/30 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <span class="flex size-9 items-center justify-center rounded-xl border bg-background/70 text-admin-selected">
          <DollarSign class="size-4" />
        </span>
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">ExchangeRate-API</p>
          <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">汇率接口</h3>
        </div>
      </div>
      <span class="rounded-full border px-2.5 py-1 text-[11px] font-black" :class="statusBadgeClass">
        {{ statusLabel }}
      </span>
    </div>

    <div class="mt-4 grid gap-4 md:grid-cols-2">
      <div class="flex items-center justify-between gap-3 rounded-xl border bg-background/70 px-3 py-2.5 md:col-span-2">
        <div>
          <span class="text-xs font-bold text-foreground">启用汇率接口</span>
          <p class="mt-0.5 text-xs text-muted-foreground">后端每天同步一次，Nuxt 前台只读取缓存后的汇率结果。</p>
        </div>
        <Switch v-model="apiSettings.exchange_rate_enabled" aria-label="启用汇率接口" />
      </div>

      <div class="rounded-xl border bg-background/70 px-3 py-2.5">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">Provider</p>
        <p class="mt-1 text-sm font-black text-foreground">ExchangeRate-API</p>
      </div>
      <div class="rounded-xl border bg-background/70 px-3 py-2.5">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">Base</p>
        <p class="mt-1 text-sm font-black text-foreground">USD</p>
      </div>

      <AdminFormField
        label="API Key"
        class="md:col-span-2"
        description="填 ExchangeRate-API 后台生成的 Key；该设置不公开给 Nuxt 前台。"
      >
        <div class="relative">
          <Input
            v-model="apiSettings.exchange_rate_api_key"
            :type="showAPIKey ? 'text' : 'password'"
            class="pr-10 font-mono"
            autocomplete="new-password"
            placeholder="输入 ExchangeRate-API key"
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            class="absolute right-0 top-0"
            :aria-label="showAPIKey ? '隐藏 API Key' : '显示 API Key'"
            @click="showAPIKey = !showAPIKey"
          >
            <EyeOff v-if="showAPIKey" class="size-4" />
            <Eye v-else class="size-4" />
          </Button>
        </div>
      </AdminFormField>

      <div class="rounded-xl border bg-background/70 px-3 py-2.5 md:col-span-2">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">内置请求地址</p>
        <p class="mt-1 break-all font-mono text-xs font-bold text-foreground">{{ presetEndpoint }}</p>
      </div>

      <AdminFormField label="目标币种" class="md:col-span-2">
        <div class="flex flex-wrap gap-2">
          <p v-if="quoteCurrencyOptions.length === 0" class="text-xs text-destructive">
            请先在系统设置的收款货币中配置除 USD 外的可用币种。
          </p>
          <button
            v-for="currency in quoteCurrencyOptions"
            :key="`quote-${currency}`"
            type="button"
            class="rounded-full border px-3 py-1.5 text-xs font-black transition hover:border-admin-selected-border hover:bg-admin-selected-soft disabled:cursor-not-allowed disabled:opacity-45"
            :class="isQuoteCurrencySelected(currency) ? 'border-admin-selected-border bg-admin-selected-soft text-admin-selected shadow-[var(--admin-control-selected-surface-shadow)]' : 'bg-background/70 text-foreground'"
            :disabled="loadingCurrencies"
            @click="toggleQuoteCurrency(currency)"
          >
            {{ currency }}
          </button>
        </div>
      </AdminFormField>
    </div>

    <div class="mt-4 grid gap-3 sm:grid-cols-3">
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Base</p>
        <p class="mt-1 text-sm font-black text-foreground">USD</p>
      </div>
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Currencies</p>
        <p class="mt-1 text-sm font-black text-foreground">{{ currencyCount(apiSettings.exchange_rate_quote_currencies) }}</p>
      </div>
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Refresh</p>
        <p class="mt-1 text-sm font-black text-foreground">每日自动</p>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, ref, watchEffect } from 'vue'
import { DollarSign, Eye, EyeOff } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

const EXCHANGE_RATE_PROVIDER = 'ExchangeRate-API'
const EXCHANGE_RATE_BASE_CURRENCY = 'USD'
const EXCHANGE_RATE_ENDPOINT = 'https://v6.exchangerate-api.com/v6/{apiKey}/latest/{base}'
const DAILY_API_REFRESH_MINUTES = 1440

const props = defineProps({
  apiSettings: { type: Object, required: true },
  currencyOptions: { type: Array, default: () => [] },
  loadingCurrencies: { type: Boolean, default: false },
})

const showAPIKey = ref(false)
const presetEndpoint = EXCHANGE_RATE_ENDPOINT

watchEffect(() => {
  props.apiSettings.exchange_rate_provider = EXCHANGE_RATE_PROVIDER
  props.apiSettings.exchange_rate_endpoint = EXCHANGE_RATE_ENDPOINT
  props.apiSettings.exchange_rate_query_template = ''
  props.apiSettings.exchange_rate_base_currency = EXCHANGE_RATE_BASE_CURRENCY
  props.apiSettings.exchange_rate_refresh_minutes = DAILY_API_REFRESH_MINUTES
})

const normalizeCurrencyCode = (currency) => String(currency || '').trim().toUpperCase()

const quoteCurrencyOptions = computed(() =>
  props.currencyOptions
    .map(normalizeCurrencyCode)
    .filter((currency, index, currencies) =>
      /^[A-Z]{3}$/.test(currency) &&
      currency !== EXCHANGE_RATE_BASE_CURRENCY &&
      currencies.indexOf(currency) === index
    )
)

const selectedQuoteCurrencies = computed(() =>
  String(props.apiSettings.exchange_rate_quote_currencies || '')
    .split(/[\s,;，；]+/)
    .map(normalizeCurrencyCode)
    .filter((currency) => /^[A-Z]{3}$/.test(currency) && currency !== EXCHANGE_RATE_BASE_CURRENCY)
)

const isQuoteCurrencySelected = (currency) =>
  selectedQuoteCurrencies.value.includes(currency)

const toggleQuoteCurrency = (currency) => {
  const current = new Set(selectedQuoteCurrencies.value)
  if (current.has(currency)) current.delete(currency)
  else current.add(currency)
  props.apiSettings.exchange_rate_quote_currencies = Array.from(current).join(',')
}

const hasAPIKey = computed(() => Boolean(String(props.apiSettings.exchange_rate_api_key || '').trim()))
const statusLabel = computed(() => {
  if (!props.apiSettings.exchange_rate_enabled) return '未启用'
  if (!hasAPIKey.value) return '缺 API Key'
  return '已启用'
})
const statusBadgeClass = computed(() => {
  if (!props.apiSettings.exchange_rate_enabled) return 'border-border bg-muted text-muted-foreground'
  if (!hasAPIKey.value) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-200'
  return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
})

const currencyCount = (value) => {
  const codes = String(value || '')
    .split(/[\s,;，；]+/)
    .map(normalizeCurrencyCode)
    .filter(Boolean)
  return `${codes.length} 个`
}
</script>
