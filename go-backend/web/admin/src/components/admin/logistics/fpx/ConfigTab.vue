<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { CheckCircle2, RefreshCw, Save } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAuthStore } from '@/stores/auth'
import fpxLogisticsAdminApi, { type FpxAPIConfigView, type FpxChannelSyncSummary, type FpxPingResult } from '@/api/fpxLogisticsAdminApi'

type Environment = 'production' | 'test'
type GatewayForm = FpxAPIConfigView & {
  app_key: string
  app_secret: string
  access_token: string
}

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:fpx:manage'))
const environment = ref<Environment>('production')
const endpointFor = (value: Environment) => value === 'test'
  ? 'https://open-test.4px.com/router/api/service'
  : 'https://open.4px.com/router/api/service'
const emptyForm = (value: Environment): GatewayForm => ({
  environment: value,
  endpoint: endpointFor(value),
  app_key: '',
  app_secret: '',
  access_token: '',
  app_key_configured: false,
  app_secret_configured: false,
  access_token_configured: false,
  enabled: false,
  last_sync_status: '',
  last_sync_scanned: 0,
  last_sync_added: 0,
  last_sync_updated: 0,
  last_sync_preserved_enabled: 0,
})

const form = ref<GatewayForm>(emptyForm('production'))
const loading = ref(false)
const saving = ref(false)
const pinging = ref(false)
const syncing = ref(false)
const message = ref('')
const error = ref('')
const pingResult = ref<FpxPingResult | null>(null)
const syncResult = ref<FpxChannelSyncSummary | null>(null)
let loadSequence = 0

const syncSummary = computed(() => {
  const result = syncResult.value
  if (!result) return ''
  return `同步成功：扫描官方服务 ${result.scanned} 项，新增 ${result.added} 项，更新 ${result.updated} 项，保留已启用 ${result.preserved_enabled} 项。`
})

function getErrorMessage(cause: unknown, fallback: string): string {
  if (typeof cause === 'object' && cause !== null) {
    const candidate = cause as {
      message?: unknown
      response?: { data?: { message?: unknown; error?: unknown } }
    }
    const apiMessage = candidate.response?.data?.message ?? candidate.response?.data?.error
    if (typeof apiMessage === 'string' && apiMessage.trim()) return apiMessage
    if (typeof candidate.message === 'string' && candidate.message.trim()) return candidate.message
  }
  return fallback
}

async function load() {
  const requestSequence = ++loadSequence
  const selectedEnvironment = environment.value
  loading.value = true
  error.value = ''
  message.value = ''
  pingResult.value = null
  syncResult.value = null
  form.value = emptyForm(selectedEnvironment)
  try {
    const saved = await fpxLogisticsAdminApi.getFpxApiConfig(selectedEnvironment)
    if (requestSequence !== loadSequence) return
    form.value = {
      ...emptyForm(selectedEnvironment),
      ...saved,
      environment: selectedEnvironment,
      app_key: '',
      app_secret: '',
      access_token: '',
    }
  } catch (cause) {
    if (requestSequence === loadSequence) error.value = getErrorMessage(cause, '加载 4PX 配置失败')
  } finally {
    if (requestSequence === loadSequence) loading.value = false
  }
}

watch(environment, () => { void load() })

async function save() {
  if (!canManage.value) return
  saving.value = true
  error.value = ''
  message.value = ''
  try {
    await fpxLogisticsAdminApi.saveFpxApiConfig({ ...form.value, environment: environment.value })
    form.value.app_key = ''
    form.value.app_secret = ''
    form.value.access_token = ''
    await load()
    message.value = '当前环境配置已独立保存'
  } catch (cause) {
    error.value = getErrorMessage(cause, '保存配置失败')
  } finally {
    saving.value = false
  }
}

async function ping() {
  if (!canManage.value) return
  pinging.value = true
  error.value = ''
  message.value = ''
  pingResult.value = null
  try {
    pingResult.value = await fpxLogisticsAdminApi.pingFpxApi({ ...form.value, environment: environment.value })
  } catch (cause) {
    error.value = getErrorMessage(cause, '连通性自检失败')
  } finally {
    pinging.value = false
  }
}

async function syncChannels() {
  if (!canManage.value) return
  syncing.value = true
  error.value = ''
  message.value = ''
  syncResult.value = null
  try {
    const result = await fpxLogisticsAdminApi.syncFpxChannels({ ...form.value, environment: environment.value })
    syncResult.value = result
    form.value.last_sync_status = 'success'
    form.value.last_synced_at = new Date().toISOString()
    form.value.last_error = ''
    form.value.last_sync_scanned = result.scanned
    form.value.last_sync_added = result.added
    form.value.last_sync_updated = result.updated
    form.value.last_sync_preserved_enabled = result.preserved_enabled
  } catch (cause) {
    error.value = getErrorMessage(cause, '同步渠道失败')
    form.value.last_sync_status = 'failed'
    form.value.last_error = error.value
  } finally {
    syncing.value = false
  }
}

onMounted(() => { void load() })
</script>

<template>
  <section class="space-y-5 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
    <div>
      <p class="text-[10px] font-black uppercase tracking-[0.18em] text-violet-600">4PX Gateway Console</p>
      <h2 class="mt-1 text-lg font-black">网关配置</h2>
      <p class="mt-1 text-xs leading-5 text-muted-foreground">生产与测试凭据独立保存；Ping 只读验证，不写渠道或同步状态。</p>
    </div>

    <p v-if="error" role="alert" class="rounded-xl border border-red-500/40 bg-red-500/10 px-3 py-3 text-sm font-medium text-red-700 dark:text-red-300">{{ error }}</p>
    <p v-if="message" role="status" class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-xs text-emerald-700">{{ message }}</p>
    <p v-if="syncResult" role="status" class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 px-3 py-3 text-sm font-semibold text-emerald-700 dark:text-emerald-300">{{ syncSummary }}</p>

    <div class="grid gap-3 sm:grid-cols-2">
      <label class="space-y-1.5">
        <span class="text-xs font-bold">环境</span>
        <select v-model="environment" :disabled="loading || saving || pinging || syncing" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
          <option value="production">生产</option>
          <option value="test">测试</option>
        </select>
      </label>
      <label class="space-y-1.5">
        <span class="text-xs font-bold">官方网关地址</span>
        <Input v-model="form.endpoint" readonly />
      </label>
      <label class="space-y-1.5">
        <span class="text-xs font-bold">AppKey</span>
        <Input v-model="form.app_key" type="password" :disabled="!canManage || loading" :placeholder="form.app_key_configured ? '已保存，留空保持不变' : '输入 AppKey'" />
      </label>
      <label class="space-y-1.5">
        <span class="text-xs font-bold">AppSecret</span>
        <Input v-model="form.app_secret" type="password" :disabled="!canManage || loading" :placeholder="form.app_secret_configured ? '已保存，留空保持不变' : '输入 AppSecret'" />
      </label>
      <label class="space-y-1.5 sm:col-span-2">
        <span class="text-xs font-bold">Access Token（可选）</span>
        <Input v-model="form.access_token" type="password" :disabled="!canManage || loading" :placeholder="form.access_token_configured ? '已保存，留空保持不变' : '输入 Access Token（可选）'" />
      </label>
    </div>

    <div class="flex flex-wrap gap-2">
      <Button :disabled="!canManage || loading || saving || pinging || syncing" @click="save">
        <Save class="size-3.5" />{{ saving ? '保存中…' : '保存配置' }}
      </Button>
      <Button variant="outline" :disabled="!canManage || loading || saving || pinging || syncing" @click="ping">
        <RefreshCw class="size-3.5" :class="{ 'animate-spin': pinging }" />{{ pinging ? '自检中…' : '连通性自检' }}
      </Button>
      <Button variant="outline" :disabled="!canManage || loading || saving || pinging || syncing" @click="syncChannels">
        <RefreshCw class="size-3.5" :class="{ 'animate-spin': syncing }" />{{ syncing ? '同步中…' : '一键同步渠道' }}
      </Button>
    </div>

    <div v-if="pingResult" role="status" class="flex items-center gap-2 rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
      <CheckCircle2 class="size-4 shrink-0" />
      <span>{{ pingResult.message }} · {{ pingResult.latency_ms }} ms</span>
    </div>

    <div v-if="form.last_sync_status" class="rounded-xl border border-border/70 bg-muted/30 p-3 text-xs text-muted-foreground">
      <p class="font-bold text-foreground">此环境最近同步：{{ form.last_sync_status === 'success' ? '成功' : '失败' }}</p>
      <p class="mt-1">扫描 {{ form.last_sync_scanned }} 项，新增 {{ form.last_sync_added }} 项，更新 {{ form.last_sync_updated }} 项，保留已启用 {{ form.last_sync_preserved_enabled }} 项</p>
      <p v-if="form.last_error" class="mt-1 break-words text-red-700 dark:text-red-300">{{ form.last_error }}</p>
    </div>
  </section>
</template>
