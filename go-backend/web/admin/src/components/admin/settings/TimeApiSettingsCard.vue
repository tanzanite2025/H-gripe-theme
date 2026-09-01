<template>
  <section class="rounded-2xl border bg-muted/30 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <span class="flex size-9 items-center justify-center rounded-xl border bg-background/70 text-admin-selected">
          <Clock3 class="size-4" />
        </span>
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Timezone Policy</p>
          <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">时区规则</h3>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <span class="rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2.5 py-1 text-[11px] font-black text-emerald-700 dark:text-emerald-200">
          内置
        </span>
        <Button v-if="canEdit" type="button" size="sm" :disabled="effectiveSaving" @click="saveTimezoneSettings">
          <LoaderCircle v-if="effectiveSaving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          {{ effectiveSaving ? '保存中' : '保存时区规则' }}
        </Button>
      </div>
    </div>

    <p v-if="lastSaveMessage" class="mt-3 rounded-lg border px-3 py-2 text-xs font-semibold" :class="lastSaveMessageClass">
      {{ lastSaveMessage }}
    </p>

    <p
      v-if="hasLegacyExternalTimeApi"
      class="mt-3 rounded-lg border border-amber-500/20 bg-amber-500/10 px-3 py-2 text-xs font-semibold text-amber-700 dark:text-amber-200"
    >
      检测到旧时间 API 配置；保存本卡片会清空 Endpoint、参数规则和 Key 引用，切换为内置时区。
    </p>

    <div class="mt-4 grid gap-4 md:grid-cols-2">
      <div class="rounded-xl border bg-background/70 px-3 py-2.5 md:col-span-2">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">时间来源</p>
        <p class="mt-1 text-sm font-bold leading-relaxed text-foreground">
          后端时间统一使用 UTC 存储和计算；时区只负责后台展示、统计窗口和排查口径，不再接第三方时间 API。
        </p>
      </div>

      <AdminFormField label="默认经营时区" description="使用 IANA 时区名称；当前默认用于后台口径说明和后续需要统一本地日界线的功能。">
        <Input v-model="apiSettings.time_api_default_timezone" :disabled="controlsDisabled" placeholder="Asia/Shanghai" />
      </AdminFormField>

      <div class="rounded-xl border bg-background/70 px-3 py-2.5">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">外部时间 API</p>
        <p class="mt-1 text-sm font-black text-foreground">关闭</p>
        <p class="mt-1 text-xs leading-relaxed text-muted-foreground">只有以后出现跨系统签名校时需求，才重新评估服务端校时。</p>
      </div>
    </div>

    <div class="mt-4 grid gap-3 sm:grid-cols-3">
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Timezone</p>
        <p class="mt-1 text-sm font-black text-foreground">{{ normalizedTimezone }}</p>
      </div>
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Source</p>
        <p class="mt-1 text-sm font-black text-foreground">Built-in</p>
      </div>
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Refresh</p>
        <p class="mt-1 text-sm font-black text-foreground">不刷新</p>
      </div>
    </div>

    <div class="mt-4 rounded-xl border bg-background/70 px-3 py-3">
      <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">不要重复建设</p>
      <div class="mt-2 grid gap-2 text-xs leading-relaxed text-muted-foreground md:grid-cols-3">
        <p><span class="font-black text-foreground">业务存储：</span>继续用 UTC，不按时区写入数据库。</p>
        <p><span class="font-black text-foreground">本地统计：</span>需要日界线时由调用端传 offset 或显式时区。</p>
        <p><span class="font-black text-foreground">访客画像：</span>继续记录 CF-Timezone / X-Timezone 请求头。</p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Clock3, LoaderCircle, Save } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { apiSettingPayload, postApiSettingsBatch } from '@/modules/settings/apiSettingsPersistence'
import type { ApiManagementSettings } from '@/modules/settings/types'

const props = withDefaults(defineProps<{
  apiSettings: ApiManagementSettings
  canEdit?: boolean
  saving?: boolean
}>(), {
  canEdit: false,
  saving: false,
})

const BUILT_IN_PROVIDER = 'built-in'
const DEFAULT_TIMEZONE = 'Asia/Shanghai'
const DISABLED_REFRESH_MINUTES = 0

const localSaving = ref(false)
const lastSaveMessage = ref('')
const lastSaveStatus = ref<'success' | 'error' | 'saving' | ''>('')
const canEdit = computed(() => props.canEdit)
const effectiveSaving = computed(() => props.saving || localSaving.value)
const controlsDisabled = computed(() => !canEdit.value || effectiveSaving.value)
const normalizedTimezone = computed(() => normalizeTimezone(props.apiSettings.time_api_default_timezone))
const hasLegacyExternalTimeApi = computed(() =>
  normalizeBooleanSetting(props.apiSettings.time_api_enabled)
  || hasValue(props.apiSettings.time_api_endpoint)
  || hasValue(props.apiSettings.time_api_key_ref)
  || (hasValue(props.apiSettings.time_api_provider) && props.apiSettings.time_api_provider !== BUILT_IN_PROVIDER)
)
const lastSaveMessageClass = computed(() => {
  if (lastSaveStatus.value === 'success') return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
  if (lastSaveStatus.value === 'error') return 'border-red-500/20 bg-red-500/10 text-red-700 dark:text-red-200'
  return 'border-border bg-muted text-muted-foreground'
})

const hasValue = (value: unknown): boolean => String(value || '').trim().length > 0

const normalizeBooleanSetting = (value: unknown): boolean =>
  value === true || value === 'true' || value === '1' || value === 1

const normalizeTimezone = (value: unknown): string => String(value || '').trim() || DEFAULT_TIMEZONE

const isValidTimezone = (value: string): boolean => {
  try {
    new Intl.DateTimeFormat('en-US', { timeZone: value }).format(new Date())
    return true
  } catch {
    return false
  }
}

const setErrorMessage = (message: string) => {
  lastSaveStatus.value = 'error'
  lastSaveMessage.value = message
  toast.error(message)
}

const errorMessage = (error: unknown, fallback: string): string =>
  error instanceof Error ? error.message : fallback

const saveTimezoneSettings = async (): Promise<void> => {
  if (!canEdit.value || effectiveSaving.value) return
  const defaultTimezone = normalizeTimezone(props.apiSettings.time_api_default_timezone)
  if (!isValidTimezone(defaultTimezone)) {
    setErrorMessage('请输入有效 IANA 时区，例如 Asia/Shanghai')
    return
  }

  localSaving.value = true
  lastSaveStatus.value = 'saving'
  lastSaveMessage.value = '正在保存时区规则...'
  try {
    await postApiSettingsBatch([
      apiSettingPayload('time_api_enabled', false, 'boolean', 'External Time API disabled; timezone is built in'),
      apiSettingPayload('time_api_provider', BUILT_IN_PROVIDER, 'string', 'Timezone source'),
      apiSettingPayload('time_api_endpoint', '', 'string', 'External Time API endpoint disabled'),
      apiSettingPayload('time_api_query_template', '', 'string', 'External Time API query template disabled'),
      apiSettingPayload('time_api_default_timezone', defaultTimezone, 'string', 'Default business timezone'),
      apiSettingPayload('time_api_refresh_minutes', DISABLED_REFRESH_MINUTES, 'number', 'External Time API refresh interval disabled'),
      apiSettingPayload('time_api_key_ref', '', 'string', 'External Time API key reference disabled'),
    ], { label: '时区规则' })
    props.apiSettings.time_api_enabled = false
    props.apiSettings.time_api_provider = BUILT_IN_PROVIDER
    props.apiSettings.time_api_endpoint = ''
    props.apiSettings.time_api_query_template = ''
    props.apiSettings.time_api_default_timezone = defaultTimezone
    props.apiSettings.time_api_refresh_minutes = DISABLED_REFRESH_MINUTES
    props.apiSettings.time_api_key_ref = ''
    lastSaveStatus.value = 'success'
    lastSaveMessage.value = '时区规则已保存'
    toast.success(lastSaveMessage.value)
  } catch (error) {
    lastSaveStatus.value = 'error'
    lastSaveMessage.value = errorMessage(error, '时区规则保存失败')
    toast.error(lastSaveMessage.value)
  } finally {
    localSaving.value = false
  }
}
</script>
