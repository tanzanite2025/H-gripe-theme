<template>
  <section class="rounded-2xl border bg-muted/30 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <span class="flex size-9 items-center justify-center rounded-xl border bg-background/70 text-admin-selected">
          <Clock3 class="size-4" />
        </span>
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Time API</p>
          <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">时间接口</h3>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <span class="rounded-full border px-2.5 py-1 text-[11px] font-black" :class="statusBadgeClass(apiSettings.time_api_enabled, apiSettings.time_api_endpoint)">
          {{ statusLabel(apiSettings.time_api_enabled, apiSettings.time_api_endpoint) }}
        </span>
        <Button v-if="isConfigured" type="button" variant="outline" size="sm" @click="toggleEditing">
          <LockKeyhole v-if="isEditing" class="size-3.5" />
          <Pencil v-else class="size-3.5" />
          {{ isEditing ? '锁定' : '修改' }}
        </Button>
      </div>
    </div>

    <div class="mt-4 grid gap-4 md:grid-cols-2">
      <div class="flex items-center justify-between gap-3 rounded-xl border bg-background/70 px-3 py-2.5 md:col-span-2">
        <div>
          <span class="text-xs font-bold text-foreground">启用时间接口</span>
          <p class="mt-0.5 text-xs text-muted-foreground">用于统一时区、接口时间戳校准和后台排查。</p>
        </div>
        <Switch v-model="apiSettings.time_api_enabled" :disabled="controlsDisabled" aria-label="启用时间接口" />
      </div>

      <div class="rounded-xl border bg-background/70 px-3 py-2.5 md:col-span-2">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">服务商</p>
        <p class="mt-1 text-sm font-black text-foreground">{{ providerName || '填入 API 地址后自动识别' }}</p>
      </div>
      <AdminFormField label="API 地址" class="md:col-span-2">
        <Input v-model="apiSettings.time_api_endpoint" :disabled="controlsDisabled" type="url" placeholder="https://api.example.com/time" @input="beginEditing" />
      </AdminFormField>
      <AdminFormField label="参数规则" class="md:col-span-2">
        <div class="grid gap-2 sm:grid-cols-3">
          <button
            v-for="option in timeQueryTemplateOptions"
            :key="option.value"
            type="button"
            class="rounded-xl border bg-background/70 px-3 py-2 text-left text-xs font-bold transition hover:border-admin-selected-border hover:bg-admin-selected-soft"
            :class="apiSettings.time_api_query_template === option.value ? 'border-admin-selected-border bg-admin-selected-soft text-admin-selected shadow-[var(--admin-control-selected-surface-shadow)]' : 'text-foreground'"
            :disabled="controlsDisabled"
            @click="selectQueryTemplate(option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </AdminFormField>
      <AdminFormField label="默认时区">
        <Input v-model="apiSettings.time_api_default_timezone" :disabled="controlsDisabled" placeholder="Asia/Shanghai" @input="beginEditing" />
      </AdminFormField>
      <AdminFormField label="密钥引用名" description="需要 Key 的接口只填安全凭据引用，不在这里写明文。">
        <Input v-model="apiSettings.time_api_key_ref" :disabled="controlsDisabled" autocomplete="off" placeholder="time_api_key" @input="beginEditing" />
      </AdminFormField>
    </div>

    <div class="mt-4 grid gap-3 sm:grid-cols-3">
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Timezone</p>
        <p class="mt-1 text-sm font-black text-foreground">{{ apiSettings.time_api_default_timezone || 'Asia/Shanghai' }}</p>
      </div>
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Source</p>
          <p class="mt-1 text-sm font-black text-foreground">{{ providerName || '待填写' }}</p>
      </div>
      <div class="rounded-xl border bg-background/70 p-3">
        <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">Refresh</p>
        <p class="mt-1 text-sm font-black text-foreground">每日自动</p>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, ref, watch, watchEffect } from 'vue'
import { Clock3, LockKeyhole, Pencil } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { inferAPIProviderFromEndpoint } from '@/components/admin/settings/apiProviderDetection'

const props = defineProps({
  apiSettings: { type: Object, required: true },
})

const timeQueryTemplateOptions = [
  { value: 'timezone={timezone}', label: 'timezone' },
  { value: 'tz={timezone}', label: 'tz' },
  { value: 'zone={timezone}', label: 'zone' },
]

const isEditing = ref(false)
const isConfigured = computed(() => Boolean(String(props.apiSettings.time_api_endpoint || '').trim()))
const controlsDisabled = computed(() => isConfigured.value && !isEditing.value)
const providerName = computed(() => props.apiSettings.time_api_provider || inferAPIProviderFromEndpoint(props.apiSettings.time_api_endpoint))

watch(
  () => props.apiSettings.time_api_endpoint,
  (endpoint) => {
    props.apiSettings.time_api_provider = inferAPIProviderFromEndpoint(endpoint)
  },
  { immediate: true }
)

const beginEditing = () => {
  if (controlsDisabled.value) return
  isEditing.value = true
}

const toggleEditing = () => {
  isEditing.value = !isEditing.value
}

const selectQueryTemplate = (value) => {
  beginEditing()
  props.apiSettings.time_api_query_template = value
}

watchEffect(() => {
  if (!timeQueryTemplateOptions.some((option) => option.value === props.apiSettings.time_api_query_template)) {
    props.apiSettings.time_api_query_template = timeQueryTemplateOptions[0].value
  }
})

const statusLabel = (enabled, endpoint) => {
  if (!enabled) return '未启用'
  if (!endpoint) return '缺 Endpoint'
  return '已启用'
}

const statusBadgeClass = (enabled, endpoint) => {
  if (!enabled) return 'border-border bg-muted text-muted-foreground'
  if (!endpoint) return 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-200'
  return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200'
}
</script>
