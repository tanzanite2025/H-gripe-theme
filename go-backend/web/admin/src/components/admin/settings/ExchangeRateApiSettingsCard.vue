<template>
  <section class="rounded-2xl border bg-muted/30 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <span class="flex size-9 items-center justify-center rounded-xl border bg-background/70 text-admin-selected">
          <DollarSign class="size-4" />
        </span>
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">ExchangeRate-API</p>
          <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">汇率接口缓存</h3>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <span class="rounded-full border px-2.5 py-1 text-[11px] font-black" :class="statusBadgeClass">
          {{ statusLabel }}
        </span>
        <button
          v-if="canEdit"
          class="inline-flex h-7 items-center justify-center gap-1 rounded-full border border-transparent bg-admin-selected px-2.5 text-xs font-black uppercase tracking-wider text-admin-selected-foreground shadow-[var(--admin-control-selected-shadow)] transition active:scale-[0.98] disabled:pointer-events-none disabled:opacity-50"
          type="button"
          :disabled="effectiveSaving"
          title="保存并启用 ExchangeRate-API Key；请求 base 跟随价格币种页的主基准币种"
          @click.prevent.stop="saveExchangeRateApiEnableStateAndApiKey"
        >
          <LoaderCircle v-if="effectiveSaving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          {{ exchangeRateApiSaveButtonLabel }}
        </button>
        <button
          v-if="canEdit"
          class="inline-flex h-7 items-center justify-center gap-1 rounded-full border border-dashed border-border/80 bg-background px-2.5 text-xs font-black uppercase tracking-wider text-foreground transition hover:bg-muted active:scale-[0.98] disabled:pointer-events-none disabled:opacity-50"
          type="button"
          :disabled="syncDisabled"
          title="先保存当前 ExchangeRate-API Key 和启用状态，再同步一次汇率缓存；同步目标由价格币种页决定"
          @click.prevent.stop="syncExchangeRateApiSettingsAndRates"
        >
 <RefreshCw :class="['size-3.5', syncing ? 'animate-spin': '']" />
          {{ syncing ? '同步中' : '同步汇率' }}
        </button>
      </div>
    </div>

    <p v-if="lastSaveMessage" class="mt-3 rounded-lg border px-3 py-2 text-xs font-semibold" :class="lastSaveMessageClass">
      {{ lastSaveMessage }}
    </p>

    <div class="mt-4 grid gap-4 md:grid-cols-2">
      <div class="flex items-center justify-between gap-3 rounded-xl border bg-background/70 px-3 py-2.5 md:col-span-2">
        <div>
          <span class="text-xs font-bold text-foreground">启用汇率接口</span>
          <p class="mt-0.5 text-xs text-muted-foreground">这里只管理 API 启用状态和 Key；请求 base 来自“价格币种”页的主基准币种。</p>
        </div>
        <Switch
          :model-value="exchangeRateApiEnabledInput"
          :disabled="!canEdit || effectiveSaving"
          aria-label="启用汇率接口"
          @update:model-value="setExchangeRateApiEnabledInput"
        />
      </div>

      <div class="rounded-xl border bg-background/70 px-3 py-2.5">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">Provider</p>
        <p class="mt-1 text-sm font-black text-foreground">ExchangeRate-API</p>
      </div>
      <div class="rounded-xl border bg-background/70 px-3 py-2.5">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">请求 base</p>
        <p class="mt-1 text-sm font-black text-foreground">{{ sourceBaseCurrency }}</p>
      </div>

      <AdminFormField
        label="API Key"
        class="md:col-span-2"
        description="这里只填 ExchangeRate-API 后台生成的 Key；完整请求地址由后端自动拼成 https://v6.exchangerate-api.com/v6/你的Key/latest/主基准币种，不公开给 Nuxt 前台。"
      >
        <div class="space-y-2">
          <div class="relative">
            <Input
              v-model="apiSettings.exchange_rate_api_key"
              :type="showAPIKey ? 'text' : 'password'"
              class="pr-10 font-mono"
              autocomplete="new-password"
              :disabled="!canEdit || effectiveSaving"
              placeholder="例如 778629c8508daca15ac81b24"
            />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              class="absolute right-0 top-0"
              :disabled="!canEdit || effectiveSaving"
              :aria-label="showAPIKey ? '隐藏 API Key' : '显示 API Key'"
              @click="showAPIKey = !showAPIKey"
            >
              <EyeOff v-if="showAPIKey" class="size-4" />
              <Eye v-else class="size-4" />
            </Button>
          </div>
          <p class="break-all font-mono text-[11px] font-semibold text-muted-foreground">
            后端实际请求示例：{{ maskedExchangeRateApiRequestPreview }}
          </p>
        </div>
      </AdminFormField>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue'
import { DollarSign, Eye, EyeOff, LoaderCircle, RefreshCw, Save } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { apiSettingPayload, postApiSettingsBatch } from '@/components/admin/settings/apiSettingsPersistence'

const EXCHANGE_RATE_PROVIDER = 'ExchangeRate-API'
const DEFAULT_PRICING_CURRENCY = 'USD'
const EXCHANGE_RATE_ENDPOINT = 'https://v6.exchangerate-api.com/v6/{apiKey}/latest/{base}'
const DAILY_API_REFRESH_MINUTES = 1440

interface ExchangeRateAPISettings {
  exchange_rate_enabled: boolean | string | number
  exchange_rate_provider: string
  exchange_rate_endpoint: string
  exchange_rate_query_template: string
  exchange_rate_refresh_minutes: number
  exchange_rate_api_key: string
}

const props = withDefaults(defineProps<{
  apiSettings: ExchangeRateAPISettings
  primaryPricingCurrency?: string
  canEdit?: boolean
  syncing?: boolean
  saving?: boolean
}>(), {
  primaryPricingCurrency: '',
  canEdit: false,
  syncing: false,
  saving: false,
})
const emit = defineEmits<{
  sync: []
}>()

const showAPIKey = ref(false)
const localSaving = ref(false)
const lastSaveMessage = ref('')
const lastSaveStatus = ref<'success' | 'error' | 'saving' | ''>('')
const exchangeRateApiEnabledInput = ref(false)

const normalizeCurrencyCode = (currency: unknown): string => String(currency || '').trim().toUpperCase()
const isCurrencyCode = (currency: unknown): boolean => /^[A-Z]{3}$/.test(normalizeCurrencyCode(currency))
const normalizeBooleanSetting = (value: unknown): boolean => value === true || value === 'true' || value === '1' || value === 1

const sourceBaseCurrency = computed(() => {
  const policyBaseCurrency = normalizeCurrencyCode(props.primaryPricingCurrency)
  if (isCurrencyCode(policyBaseCurrency)) return policyBaseCurrency
  return DEFAULT_PRICING_CURRENCY
})

const maskedApiKey = computed(() => {
  const key = String(props.apiSettings.exchange_rate_api_key || '').trim()
  if (!key) return '你的Key'
  if (key.length <= 8) return '****'
  return `${key.slice(0, 4)}****${key.slice(-4)}`
})
const maskedExchangeRateApiRequestPreview = computed(() =>
  `https://v6.exchangerate-api.com/v6/${maskedApiKey.value}/latest/${sourceBaseCurrency.value}`
)

watchEffect(() => {
  props.apiSettings.exchange_rate_provider = EXCHANGE_RATE_PROVIDER
  props.apiSettings.exchange_rate_endpoint = EXCHANGE_RATE_ENDPOINT
  props.apiSettings.exchange_rate_query_template = ''
  props.apiSettings.exchange_rate_refresh_minutes = DAILY_API_REFRESH_MINUTES
})

watch(
  () => props.apiSettings.exchange_rate_enabled,
  (enabled) => {
    exchangeRateApiEnabledInput.value = normalizeBooleanSetting(enabled)
  },
  { immediate: true }
)

const setExchangeRateApiEnabledInput = (value: unknown) => {
  const enabled = normalizeBooleanSetting(value)
  exchangeRateApiEnabledInput.value = enabled
  props.apiSettings.exchange_rate_enabled = enabled
  lastSaveStatus.value = ''
  lastSaveMessage.value = enabled
    ? '已切换为启用，点击“保存 API 启用/Key”后写入后台'
    : '已切换为停用，点击“保存 API 启用/Key”后写入后台'
}

const hasAPIKey = computed(() => Boolean(String(props.apiSettings.exchange_rate_api_key || '').trim()))
const canEdit = computed(() => props.canEdit)
const syncing = computed(() => props.syncing)
const saving = computed(() => props.saving)
const effectiveSaving = computed(() => saving.value || localSaving.value)
const exchangeRateApiSaveButtonLabel = computed(() => {
  if (effectiveSaving.value) return '保存中'
  if (!exchangeRateApiEnabledInput.value && hasAPIKey.value) return '启用并保存 Key'
  return '保存 API Key'
})
const syncDisabled = computed(() => (
  !canEdit.value ||
  syncing.value ||
  effectiveSaving.value ||
  !exchangeRateApiEnabledInput.value ||
  !hasAPIKey.value
))
const lastSaveMessageClass = computed(() => {
  if (lastSaveStatus.value === 'success') return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
  if (lastSaveStatus.value === 'error') return 'border-red-500/20 bg-red-500/10 text-red-700 dark:text-red-200'
  return 'border-border bg-muted text-muted-foreground'
})

const errorMessage = (error: unknown, fallback: string): string =>
  error instanceof Error ? error.message : fallback

const saveExchangeRateApiEnableStateAndApiKey = async (): Promise<boolean> => {
  if (!canEdit.value || effectiveSaving.value) return false
  localSaving.value = true
  lastSaveStatus.value = 'saving'
  lastSaveMessage.value = '正在保存 ExchangeRate-API Key...'
  try {
    const apiKey = String(props.apiSettings.exchange_rate_api_key || '').trim()
    const enabled = Boolean(apiKey)
    await postApiSettingsBatch([
      apiSettingPayload('exchange_rate_enabled', enabled, 'boolean', '是否启用 ExchangeRate-API 汇率缓存'),
      apiSettingPayload('exchange_rate_provider', EXCHANGE_RATE_PROVIDER, 'string', '汇率接口提供方'),
      apiSettingPayload('exchange_rate_endpoint', EXCHANGE_RATE_ENDPOINT, 'string', '后端内置汇率接口地址模板'),
      apiSettingPayload('exchange_rate_query_template', '', 'string', '汇率接口查询模板'),
      apiSettingPayload('exchange_rate_refresh_minutes', DAILY_API_REFRESH_MINUTES, 'number', '汇率缓存刷新间隔分钟数'),
      apiSettingPayload('exchange_rate_api_key', apiKey, 'string', 'ExchangeRate-API Key，仅后台私有保存'),
    ], { label: 'ExchangeRate-API 设置' })
    props.apiSettings.exchange_rate_enabled = enabled
    props.apiSettings.exchange_rate_api_key = apiKey
    exchangeRateApiEnabledInput.value = enabled
    lastSaveStatus.value = 'success'
    lastSaveMessage.value = enabled
      ? `ExchangeRate-API 已启用并保存 Key；后端请求 base=${sourceBaseCurrency.value}`
      : 'ExchangeRate-API Key 已清空，接口保持停用'
    toast.success(lastSaveMessage.value)
    return true
  } catch (error) {
    lastSaveStatus.value = 'error'
    lastSaveMessage.value = errorMessage(error, 'ExchangeRate-API 设置保存失败')
    toast.error(lastSaveMessage.value)
    return false
  } finally {
    localSaving.value = false
  }
}

const syncExchangeRateApiSettingsAndRates = async (): Promise<void> => {
  const saved = await saveExchangeRateApiEnableStateAndApiKey()
  if (saved) emit('sync')
}
const statusLabel = computed(() => {
  if (!exchangeRateApiEnabledInput.value) return '未启用'
  if (!hasAPIKey.value) return '缺 API Key'
  return '已启用'
})
const statusBadgeClass = computed(() => {
  if (!exchangeRateApiEnabledInput.value) return 'border-border bg-muted text-muted-foreground'
  if (!hasAPIKey.value) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-200'
  return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
})
</script>
