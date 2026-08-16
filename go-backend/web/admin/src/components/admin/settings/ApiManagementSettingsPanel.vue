<template>
  <div class="grid items-start gap-4 xl:grid-cols-2">
    <ExchangeRateApiSettingsCard
      :api-settings="apiSettings"
      :primary-pricing-currency="primaryPricingCurrency"
      :can-edit="canEdit"
      :syncing="syncingExchangeRates"
      :saving="savingApiSettings"
      @sync="emit('sync-exchange-rates')"
    />

    <TimeApiSettingsCard
      :api-settings="apiSettings"
      :can-edit="canEdit"
      :saving="savingApiSettings"
    />

    <CustomsLookupApiSettingsCard
      :api-settings="apiSettings"
      :can-edit="canEdit"
      :saving="savingApiSettings"
    />

    <section class="rounded-2xl border bg-muted/30 p-4 xl:col-span-2">
      <div class="grid gap-4 md:grid-cols-3">
        <div class="flex items-start gap-3">
          <ShieldCheck class="mt-0.5 size-4 flex-none text-admin-selected" />
          <div>
            <p class="text-xs font-black text-foreground">后端代理</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">第三方接口由后端请求并缓存，前台不直接拿接口地址和密钥。</p>
          </div>
        </div>
        <div class="flex items-start gap-3">
          <KeyRound class="mt-0.5 size-4 flex-none text-admin-selected" />
          <div>
            <p class="text-xs font-black text-foreground">私有 Key</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">ExchangeRate-API Key 只保存在后台私有设置，不写进 Nuxt 前台包。</p>
          </div>
        </div>
        <div class="flex items-start gap-3">
          <RefreshCw class="mt-0.5 size-4 flex-none text-admin-selected" />
          <div>
            <p class="text-xs font-black text-foreground">缓存优先</p>
            <p class="mt-1 text-xs leading-relaxed text-muted-foreground">汇率和时间接口失败时，业务可以继续使用最近一次有效缓存。</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { KeyRound, RefreshCw, ShieldCheck } from '@lucide/vue'
import CustomsLookupApiSettingsCard from '@/components/admin/settings/CustomsLookupApiSettingsCard.vue'
import ExchangeRateApiSettingsCard from '@/components/admin/settings/ExchangeRateApiSettingsCard.vue'
import TimeApiSettingsCard from '@/components/admin/settings/TimeApiSettingsCard.vue'

interface APISettings {
  exchange_rate_enabled: boolean | string | number
  exchange_rate_provider: string
  exchange_rate_endpoint: string
  exchange_rate_query_template: string
  exchange_rate_refresh_minutes: number
  exchange_rate_api_key: string
  time_api_enabled: boolean | string | number
  time_api_provider: string
  time_api_endpoint: string
  time_api_query_template: string
  time_api_default_timezone: string
  time_api_refresh_minutes: number
  time_api_key_ref: string
  customs_lookup_us_hts_enabled: boolean | string | number
  customs_lookup_us_hts_endpoint: string
  customs_lookup_us_hts_api_key: string
  customs_lookup_us_hts_api_key_header: string
  customs_lookup_uk_trade_tariff_enabled: boolean | string | number
  customs_lookup_uk_trade_tariff_endpoint: string
  customs_lookup_uk_trade_tariff_api_key: string
  customs_lookup_uk_trade_tariff_api_key_header: string
}

withDefaults(defineProps<{
  apiSettings: APISettings
  primaryPricingCurrency?: string
  canEdit?: boolean
  syncingExchangeRates?: boolean
  savingApiSettings?: boolean
}>(), {
  primaryPricingCurrency: '',
  canEdit: false,
  syncingExchangeRates: false,
  savingApiSettings: false,
})

const emit = defineEmits<{
  'sync-exchange-rates': []
}>()
</script>
