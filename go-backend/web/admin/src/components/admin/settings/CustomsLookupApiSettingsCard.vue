<template>
  <section class="rounded-2xl border bg-muted/30 p-4 xl:col-span-2">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <span class="flex size-9 items-center justify-center rounded-xl border bg-background/70 text-admin-selected">
          <FileSearch class="size-4" />
        </span>
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Customs Lookup API</p>
          <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">清关编码查询接口</h3>
        </div>
      </div>
      <Button v-if="canEdit" type="button" :disabled="effectiveSaving" @click="saveSettings">
        <LoaderCircle v-if="effectiveSaving" class="size-3.5 animate-spin" />
        <Save v-else class="size-3.5" />
        {{ effectiveSaving ? '保存中' : '保存清关接口' }}
      </Button>
    </div>

    <p v-if="lastSaveMessage" class="mt-3 rounded-lg border px-3 py-2 text-xs font-semibold" :class="lastSaveMessageClass">
      {{ lastSaveMessage }}
    </p>

    <div class="mt-4 grid gap-4 md:grid-cols-2">
      <div class="rounded-xl border bg-background/70 p-4">
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">US HTS</p>
            <p class="mt-1 text-sm font-black text-foreground">美国商品协调关税表</p>
          </div>
          <span class="rounded-full border px-2.5 py-1 text-[11px] font-black" :class="statusBadgeClass(apiSettings.customs_lookup_us_hts_enabled, apiSettings.customs_lookup_us_hts_endpoint)">
            {{ statusLabel(apiSettings.customs_lookup_us_hts_enabled, apiSettings.customs_lookup_us_hts_endpoint) }}
          </span>
        </div>
        <div class="mt-4 space-y-4">
          <div class="flex items-center justify-between gap-3 rounded-xl border bg-muted/30 px-3 py-2.5">
            <div>
              <span class="text-xs font-bold text-foreground">启用 US HTS</span>
              <p class="mt-0.5 text-xs text-muted-foreground">公开接口默认可用，无 Key 也可以查询。</p>
            </div>
            <Switch v-model="usHTSEnabled" :disabled="controlsDisabled" aria-label="启用 US HTS" />
          </div>
          <AdminFormField label="API 地址">
            <Input v-model="apiSettings.customs_lookup_us_hts_endpoint" :disabled="controlsDisabled" placeholder="https://hts.usitc.gov/reststop/search" />
          </AdminFormField>
          <AdminFormField label="API Key" description="可选；如服务商要求，可在 Endpoint 中使用 {apiKey}，或通过请求头发送。">
            <Input
              v-model="apiSettings.customs_lookup_us_hts_api_key"
              :disabled="controlsDisabled"
              type="password"
              autocomplete="new-password"
              placeholder="可留空"
            />
          </AdminFormField>
          <AdminFormField label="Key 请求头" description="默认 X-API-Key；如不需要可留空。">
            <Input v-model="apiSettings.customs_lookup_us_hts_api_key_header" :disabled="controlsDisabled" placeholder="X-API-Key" />
          </AdminFormField>
        </div>
      </div>

      <div class="rounded-xl border bg-background/70 p-4">
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">UK TRADE TARIFF</p>
            <p class="mt-1 text-sm font-black text-foreground">英国商品编码查询</p>
          </div>
          <span class="rounded-full border px-2.5 py-1 text-[11px] font-black" :class="statusBadgeClass(apiSettings.customs_lookup_uk_trade_tariff_enabled, apiSettings.customs_lookup_uk_trade_tariff_endpoint)">
            {{ statusLabel(apiSettings.customs_lookup_uk_trade_tariff_enabled, apiSettings.customs_lookup_uk_trade_tariff_endpoint) }}
          </span>
        </div>
        <div class="mt-4 space-y-4">
          <div class="flex items-center justify-between gap-3 rounded-xl border bg-muted/30 px-3 py-2.5">
            <div>
              <span class="text-xs font-bold text-foreground">启用 UK Trade Tariff</span>
              <p class="mt-0.5 text-xs text-muted-foreground">默认根据 8 或 10 位 commodity code 查询。</p>
            </div>
            <Switch v-model="ukTradeTariffEnabled" :disabled="controlsDisabled" aria-label="启用 UK Trade Tariff" />
          </div>
          <AdminFormField label="API 地址">
            <Input v-model="apiSettings.customs_lookup_uk_trade_tariff_endpoint" :disabled="controlsDisabled" placeholder="https://www.trade-tariff.service.gov.uk/api/v2/commodities" />
          </AdminFormField>
          <AdminFormField label="API Key" description="可选；如服务商要求，可在 Endpoint 中使用 {apiKey}，或通过请求头发送。">
            <Input
              v-model="apiSettings.customs_lookup_uk_trade_tariff_api_key"
              :disabled="controlsDisabled"
              type="password"
              autocomplete="new-password"
              placeholder="可留空"
            />
          </AdminFormField>
          <AdminFormField label="Key 请求头" description="默认 X-API-Key；如不需要可留空。">
            <Input v-model="apiSettings.customs_lookup_uk_trade_tariff_api_key_header" :disabled="controlsDisabled" placeholder="X-API-Key" />
          </AdminFormField>
        </div>
      </div>
    </div>

    <div class="mt-4 flex items-start gap-3 rounded-xl border border-sky-500/20 bg-sky-500/10 p-3 text-sky-900 dark:text-sky-100">
      <KeyRound class="mt-0.5 size-4 flex-none" />
      <p class="text-xs leading-relaxed text-sky-800/80 dark:text-sky-100/75">
        API Key 只保存在后台私有设置中，由后端代理请求；清关资料中心不会把 Key 下发到前台或浏览器。
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { FileSearch, KeyRound, LoaderCircle, Save } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { apiSettingPayload, postApiSettingsBatch } from '@/components/admin/settings/apiSettingsPersistence'

interface CustomsLookupAPISettings {
  customs_lookup_us_hts_enabled: boolean | string | number
  customs_lookup_us_hts_endpoint: string
  customs_lookup_us_hts_api_key: string
  customs_lookup_us_hts_api_key_header: string
  customs_lookup_uk_trade_tariff_enabled: boolean | string | number
  customs_lookup_uk_trade_tariff_endpoint: string
  customs_lookup_uk_trade_tariff_api_key: string
  customs_lookup_uk_trade_tariff_api_key_header: string
}

const props = withDefaults(defineProps<{
  apiSettings: CustomsLookupAPISettings
  canEdit?: boolean
  saving?: boolean
}>(), {
  canEdit: false,
  saving: false,
})

const localSaving = ref(false)
const lastSaveMessage = ref('')
const lastSaveStatus = ref<'success' | 'error' | ''>('')
const effectiveSaving = computed(() => props.saving || localSaving.value)
const controlsDisabled = computed(() => !props.canEdit || effectiveSaving.value)

const normalizeBooleanSetting = (value: unknown): boolean => (
  value === true || value === 'true' || value === '1' || value === 1
)

const usHTSEnabled = computed<boolean>({
  get: () => normalizeBooleanSetting(props.apiSettings.customs_lookup_us_hts_enabled),
  set: (value) => { props.apiSettings.customs_lookup_us_hts_enabled = value },
})

const ukTradeTariffEnabled = computed<boolean>({
  get: () => normalizeBooleanSetting(props.apiSettings.customs_lookup_uk_trade_tariff_enabled),
  set: (value) => { props.apiSettings.customs_lookup_uk_trade_tariff_enabled = value },
})

const statusLabel = (enabled: unknown, endpoint: unknown): string => {
  if (!normalizeBooleanSetting(enabled)) return '未启用'
  if (!String(endpoint || '').trim()) return '缺 Endpoint'
  return '已启用'
}

const statusBadgeClass = (enabled: unknown, endpoint: unknown): string => {
  if (!normalizeBooleanSetting(enabled)) return 'border-border bg-muted text-muted-foreground'
  if (!String(endpoint || '').trim()) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-200'
  return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
}

const lastSaveMessageClass = computed(() => (
  lastSaveStatus.value === 'success'
    ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
    : 'border-red-500/20 bg-red-500/10 text-red-700 dark:text-red-200'
))

const saveSettings = async () => {
  if (effectiveSaving.value) return
  localSaving.value = true
  lastSaveStatus.value = ''
  lastSaveMessage.value = '正在保存清关接口设置...'
  try {
    const values = {
      customs_lookup_us_hts_enabled: normalizeBooleanSetting(props.apiSettings.customs_lookup_us_hts_enabled),
      customs_lookup_us_hts_endpoint: String(props.apiSettings.customs_lookup_us_hts_endpoint || '').trim(),
      customs_lookup_us_hts_api_key: String(props.apiSettings.customs_lookup_us_hts_api_key || '').trim(),
      customs_lookup_us_hts_api_key_header: String(props.apiSettings.customs_lookup_us_hts_api_key_header || '').trim(),
      customs_lookup_uk_trade_tariff_enabled: normalizeBooleanSetting(props.apiSettings.customs_lookup_uk_trade_tariff_enabled),
      customs_lookup_uk_trade_tariff_endpoint: String(props.apiSettings.customs_lookup_uk_trade_tariff_endpoint || '').trim(),
      customs_lookup_uk_trade_tariff_api_key: String(props.apiSettings.customs_lookup_uk_trade_tariff_api_key || '').trim(),
      customs_lookup_uk_trade_tariff_api_key_header: String(props.apiSettings.customs_lookup_uk_trade_tariff_api_key_header || '').trim(),
    }
    await postApiSettingsBatch([
      apiSettingPayload('customs_lookup_us_hts_enabled', values.customs_lookup_us_hts_enabled, 'boolean', '是否启用 US HTS 清关编码查询'),
      apiSettingPayload('customs_lookup_us_hts_endpoint', values.customs_lookup_us_hts_endpoint, 'string', 'US HTS 清关编码查询地址'),
      apiSettingPayload('customs_lookup_us_hts_api_key', values.customs_lookup_us_hts_api_key, 'string', 'US HTS API Key，仅后台私有保存'),
      apiSettingPayload('customs_lookup_us_hts_api_key_header', values.customs_lookup_us_hts_api_key_header, 'string', 'US HTS API Key 请求头'),
      apiSettingPayload('customs_lookup_uk_trade_tariff_enabled', values.customs_lookup_uk_trade_tariff_enabled, 'boolean', '是否启用 UK Trade Tariff 清关编码查询'),
      apiSettingPayload('customs_lookup_uk_trade_tariff_endpoint', values.customs_lookup_uk_trade_tariff_endpoint, 'string', 'UK Trade Tariff 清关编码查询地址'),
      apiSettingPayload('customs_lookup_uk_trade_tariff_api_key', values.customs_lookup_uk_trade_tariff_api_key, 'string', 'UK Trade Tariff API Key，仅后台私有保存'),
      apiSettingPayload('customs_lookup_uk_trade_tariff_api_key_header', values.customs_lookup_uk_trade_tariff_api_key_header, 'string', 'UK Trade Tariff API Key 请求头'),
    ], { label: '清关接口设置' })
    Object.assign(props.apiSettings, values)
    lastSaveStatus.value = 'success'
    lastSaveMessage.value = '清关接口设置已保存'
    toast.success(lastSaveMessage.value)
  } catch (error) {
    lastSaveStatus.value = 'error'
    lastSaveMessage.value = error instanceof Error ? error.message : '清关接口设置保存失败'
    toast.error(lastSaveMessage.value)
  } finally {
    localSaving.value = false
  }
}
</script>
