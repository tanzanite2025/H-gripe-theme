<template>
  <div class="space-y-5">
    <AdminPageHeader :title="pageTitle" :description="pageDescription">
      <template #actions>
        <span
          v-if="activeTab === 'api'"
          class="rounded-full border px-3 py-1.5 text-[10px] font-black uppercase tracking-widest"
          :class="apiStatusClass"
        >
          {{ apiStatusLabel }}
        </span>
      </template>
    </AdminPageHeader>

    <CurrencyPolicySettingsCard
      v-if="activeTab === 'overview'"
      :can-edit="canEdit"
    />

    <div v-else class="space-y-5">
      <section class="rounded-[28px] border border-sky-500/20 bg-sky-50/70 p-5">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="flex min-w-0 items-start gap-3">
            <span class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-slate-950 text-white shadow-[0_12px_24px_rgb(15_23_42_/_0.18)]">
              <Cable class="size-5" />
            </span>
            <div class="min-w-0">
              <p class="text-[10px] font-black uppercase tracking-[0.18em] text-sky-800/70">Exchange Rate Provider</p>
              <h2 class="mt-1 text-xl font-black tracking-tight text-slate-950">汇率 API 连接</h2>
              <p class="mt-1 max-w-2xl text-xs leading-relaxed text-slate-600">
                这里仅维护汇率服务的连接状态和私有 Key；请求 base 来自后台录入币种，quote targets 来自市场展示币种。
              </p>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:justify-end">
            <div class="rounded-2xl border border-sky-500/15 bg-white/80 px-3 py-2">
              <p class="text-[9px] font-black uppercase tracking-widest text-slate-500">Provider</p>
              <p class="mt-1 text-sm font-black text-slate-950">{{ apiSettings.exchange_rate_provider || 'ExchangeRate-API' }}</p>
            </div>
            <div class="rounded-2xl border border-sky-500/15 bg-white/80 px-3 py-2">
              <p class="text-[9px] font-black uppercase tracking-widest text-slate-500">请求 base</p>
              <p class="mt-1 font-mono text-sm font-black text-slate-950">{{ primaryPricingCurrency }}</p>
            </div>
          </div>
        </div>
      </section>

      <ExchangeRateApiSettingsCard
        :api-settings="apiSettings"
        :primary-pricing-currency="primaryPricingCurrency"
        :can-edit="canEdit"
        :syncing="syncingExchangeRates"
        :saving="loadingApiSettings"
        @sync="syncExchangeRates"
      />

      <section class="grid gap-3 sm:grid-cols-3">
        <div class="rounded-2xl border bg-card p-4">
          <div class="flex items-center gap-2">
            <ShieldCheck class="size-4 text-emerald-600" />
            <p class="text-xs font-black text-foreground">后端代理</p>
          </div>
          <p class="mt-2 text-xs leading-relaxed text-muted-foreground">第三方请求在后端完成，前台不会拿到接口地址或密钥。</p>
        </div>
        <div class="rounded-2xl border bg-card p-4">
          <div class="flex items-center gap-2">
            <KeyRound class="size-4 text-sky-600" />
            <p class="text-xs font-black text-foreground">私有 Key</p>
          </div>
          <p class="mt-2 text-xs leading-relaxed text-muted-foreground">API Key 只写入后台私有设置，不会打包到 Nuxt 前台。</p>
        </div>
        <div class="rounded-2xl border bg-card p-4">
          <div class="flex items-center gap-2">
            <RefreshCw class="size-4 text-amber-600" />
            <p class="text-xs font-black text-foreground">缓存优先</p>
          </div>
          <p class="mt-2 text-xs leading-relaxed text-muted-foreground">服务短暂不可用时，业务仍可读取最近一次有效汇率缓存。</p>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { Cable, KeyRound, RefreshCw, ShieldCheck } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import CurrencyPolicySettingsCard from '@/components/admin/settings/CurrencyPolicySettingsCard.vue'
import ExchangeRateApiSettingsCard from '@/components/admin/settings/ExchangeRateApiSettingsCard.vue'
import { useRouteTab } from '@/composables/useRouteTab'
import { useAdminI18n } from '@/i18n'
import type { ExchangeRateSettings } from '@/modules/settings/types'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const { t } = useAdminI18n()
const authStore = useAuthStore()
const DAILY_API_REFRESH_MINUTES = 1440
const DEFAULT_PRICING_CURRENCY = 'USD'
const EXCHANGE_RATE_PROVIDER = 'ExchangeRate-API'
const EXCHANGE_RATE_ENDPOINT = 'https://v6.exchangerate-api.com/v6/{apiKey}/latest/{base}'

const activeTab = useRouteTab({
  defaultValue: 'overview',
  values: ['overview', 'api'],
  routes: {
    overview: 'CurrencyExchangeOverview',
    api: 'CurrencyExchangeApi',
  },
})

const apiSettings = reactive<ExchangeRateSettings>({
  exchange_rate_enabled: false as boolean | string | number,
  exchange_rate_provider: EXCHANGE_RATE_PROVIDER,
  exchange_rate_endpoint: EXCHANGE_RATE_ENDPOINT,
  exchange_rate_query_template: '',
  exchange_rate_refresh_minutes: DAILY_API_REFRESH_MINUTES,
  exchange_rate_api_key: '',
})

const primaryPricingCurrency = ref(DEFAULT_PRICING_CURRENCY)
const loadingApiSettings = ref(false)
const syncingExchangeRates = ref(false)

const canEdit = computed(() => authStore.hasPermission('settings:edit'))
const pageTitle = computed(() => activeTab.value === 'api' ? '汇率 API' : '币种与汇率')
const pageDescription = computed(() => activeTab.value === 'api'
  ? '管理汇率服务连接、私有凭据与同步入口'
  : '管理后台录入币种、市场汇率目标和汇率缓存')

const normalizeCurrencyCode = (value: unknown): string => String(value || '').trim().toUpperCase()
const normalizeBooleanSetting = (value: unknown): boolean => (
  value === true || value === 'true' || value === '1' || value === 1
)

const apiStatusLabel = computed(() => {
  if (!normalizeBooleanSetting(apiSettings.exchange_rate_enabled)) return '接口未启用'
  if (!String(apiSettings.exchange_rate_api_key || '').trim()) return '缺少 API Key'
  return '接口已启用'
})

const apiStatusClass = computed(() => {
  if (!normalizeBooleanSetting(apiSettings.exchange_rate_enabled)) return 'border-border bg-muted text-muted-foreground'
  if (!String(apiSettings.exchange_rate_api_key || '').trim()) return 'border-amber-500/20 bg-amber-500/10 text-amber-700'
  return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700'
})

const coerceSettingValue = (value: unknown, key: string): string | number | boolean => {
  if (key === 'exchange_rate_enabled') return normalizeBooleanSetting(value)
  if (key === 'exchange_rate_refresh_minutes') {
    const parsed = Number(value)
    return Number.isFinite(parsed) && parsed > 0 ? parsed : DAILY_API_REFRESH_MINUTES
  }
  return String(value ?? '')
}

const loadPrimaryPricingCurrency = async (): Promise<void> => {
  const response = await axios.get('/api/admin/settings/currency-policy')
  const policy = response.data?.policy || {}
  primaryPricingCurrency.value = normalizeCurrencyCode(policy.primary_currency) || DEFAULT_PRICING_CURRENCY
}

const loadApiSettings = async (): Promise<void> => {
  loadingApiSettings.value = true
  try {
    const response = await axios.get('/api/admin/settings/api', { params: { locale: 'en' } })
    const settings = Array.isArray(response.data?.settings) ? response.data.settings : []
    for (const setting of settings) {
      const key = String(setting?.key || '').replace(/^api_/, '')
      if (!(key in apiSettings)) continue
      apiSettings[key] = coerceSettingValue(setting.value, key) as never
    }
    apiSettings.exchange_rate_provider = EXCHANGE_RATE_PROVIDER
    apiSettings.exchange_rate_endpoint = EXCHANGE_RATE_ENDPOINT
    apiSettings.exchange_rate_query_template = ''
    apiSettings.exchange_rate_refresh_minutes = DAILY_API_REFRESH_MINUTES
  } catch (error) {
    console.error('Failed to load exchange rate API settings:', error)
    toast.error('汇率 API 设置读取失败')
  } finally {
    loadingApiSettings.value = false
  }
}

const syncExchangeRates = async (): Promise<void> => {
  syncingExchangeRates.value = true
  try {
    const response = await axios.post('/api/admin/settings/exchange-rates/sync')
    const rates = response.data?.data?.rates || response.data?.rates || []
    toast.success(`汇率缓存已同步：${Array.isArray(rates) ? rates.length : 0} 个币种`)
  } catch (error) {
    const responseData = error?.response?.data
    toast.error(responseData?.message || responseData?.error || '汇率同步失败')
  } finally {
    syncingExchangeRates.value = false
  }
}

watch(activeTab, async (tab) => {
  if (tab !== 'api') return
  await Promise.all([loadPrimaryPricingCurrency(), loadApiSettings()])
}, { immediate: true })
</script>
