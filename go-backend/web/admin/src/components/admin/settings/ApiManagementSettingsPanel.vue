<template>
  <div class="grid items-start gap-4 xl:grid-cols-2">
    <ExchangeRateApiSettingsCard
      :api-settings="apiSettings"
      :currency-options="currencyOptions"
      :loading-currencies="loadingCurrencies"
    />

    <TimeApiSettingsCard :api-settings="apiSettings" />

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

<script setup>
import { KeyRound, RefreshCw, ShieldCheck } from '@lucide/vue'
import ExchangeRateApiSettingsCard from '@/components/admin/settings/ExchangeRateApiSettingsCard.vue'
import TimeApiSettingsCard from '@/components/admin/settings/TimeApiSettingsCard.vue'

defineProps({
  apiSettings: { type: Object, required: true },
  currencyOptions: { type: Array, default: () => [] },
  loadingCurrencies: { type: Boolean, default: false },
})
</script>
