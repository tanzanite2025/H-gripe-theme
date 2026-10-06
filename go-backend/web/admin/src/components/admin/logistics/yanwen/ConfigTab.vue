<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { CheckCircle2, RefreshCw, Save } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAuthStore } from '@/stores/auth'
import yanwenLogisticsAdminApi, { type YanwenAPIConfigView, type YanwenPingResult, type YanwenProductSyncSummary, type YanwenCountrySyncSummary, type YanwenWarehouseSyncSummary } from '@/api/yanwenLogisticsAdminApi'

type Environment = 'fat' | 'production'

type GatewayForm = YanwenAPIConfigView & {
  user_id: string
  api_token: string
}

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:yanwen:manage'))
const environment = ref<Environment>('fat')
const endpointFor = (value: Environment) => value === 'fat'
  ? 'https://open-fat.yw56.com.cn/api/order'
  : 'https://open.yw56.com.cn/api/order'

const emptyForm = (value: Environment): GatewayForm => ({
  environment: value,
  endpoint: endpointFor(value),
  user_id: '',
  api_token: '',
  user_id_configured: false,
  api_token_configured: false,
  enabled: false,
})

const form = ref<GatewayForm>(emptyForm('fat'))
const loading = ref(false)
const saving = ref(false)
const pinging = ref(false)
const syncingCountries = ref(false)
const syncingWarehouses = ref(false)
const syncingProducts = ref(false)
const message = ref('')
const error = ref('')
const pingResult = ref<YanwenPingResult | null>(null)
const productSyncSummary = ref<YanwenProductSyncSummary | null>(null)
const countrySyncSummary = ref<YanwenCountrySyncSummary | null>(null)
const warehouseSyncSummary = ref<YanwenWarehouseSyncSummary | null>(null)
const countryCatalogLoaded = ref(false)
const warehouseCatalogLoaded = ref(false)
const productCatalogLoaded = ref(false)
let loadSequence = 0

const canSave = computed(() => canManage.value
  && (form.value.user_id.trim().length > 0 || form.value.user_id_configured)
  && (form.value.api_token.trim().length > 0 || form.value.api_token_configured))

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
  countrySyncSummary.value = null
  warehouseSyncSummary.value = null
  productSyncSummary.value = null
  countryCatalogLoaded.value = false
  warehouseCatalogLoaded.value = false
  productCatalogLoaded.value = false
  form.value = emptyForm(selectedEnvironment)
  try {
    const saved = await yanwenLogisticsAdminApi.getYanwenApiConfig(selectedEnvironment)
    if (requestSequence !== loadSequence) return
    form.value = {
      ...emptyForm(selectedEnvironment),
      ...saved,
      environment: selectedEnvironment,
      user_id: '',
      api_token: '',
    }
  } catch (cause) {
    if (requestSequence === loadSequence) error.value = getErrorMessage(cause, '加载燕文配置失败')
  } finally {
    if (requestSequence === loadSequence) loading.value = false
  }
}

async function loadProductCatalogStatus() {
  try {
    const products = await yanwenLogisticsAdminApi.listYanwenProducts(environment.value)
    productCatalogLoaded.value = products.length > 0
  } catch {
    productCatalogLoaded.value = false
  }
}

async function loadCountryCatalogStatus() {
  try {
    const countries = await yanwenLogisticsAdminApi.listYanwenCountries(environment.value)
    countryCatalogLoaded.value = countries.length > 0
  } catch {
    countryCatalogLoaded.value = false
  }
}

async function loadWarehouseCatalogStatus() {
  try {
    const warehouses = await yanwenLogisticsAdminApi.listYanwenWarehouses(environment.value)
    warehouseCatalogLoaded.value = warehouses.length > 0
  } catch {
    warehouseCatalogLoaded.value = false
  }
}

async function save() {
  if (!canSave.value) return
  saving.value = true
  error.value = ''
  message.value = ''
  pingResult.value = null
  try {
    await yanwenLogisticsAdminApi.saveYanwenApiConfig({
      environment: environment.value,
      endpoint: form.value.endpoint,
      user_id: form.value.user_id,
      api_token: form.value.api_token,
      enabled: form.value.enabled,
    })
    await load()
    message.value = '当前环境配置已安全保存；留空的凭据保持原值不变。'
  } catch (cause) {
    error.value = getErrorMessage(cause, '保存燕文配置失败')
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
    pingResult.value = await yanwenLogisticsAdminApi.pingYanwenApi({
      environment: environment.value,
      endpoint: form.value.endpoint,
      user_id: form.value.user_id,
      api_token: form.value.api_token,
    })
  } catch (cause) {
    error.value = getErrorMessage(cause, '燕文网关 Ping 失败')
  } finally {
    pinging.value = false
  }
}

async function syncProducts() {
  if (!canManage.value) return
  syncingProducts.value = true
  error.value = ''
  message.value = ''
  productSyncSummary.value = null
  try {
    productSyncSummary.value = await yanwenLogisticsAdminApi.syncYanwenProducts({
      environment: environment.value,
      endpoint: form.value.endpoint,
      user_id: form.value.user_id,
      api_token: form.value.api_token,
    })
    productCatalogLoaded.value = true
    message.value = '燕文已开通产品目录同步完成，可在服务集合中选择官方产品。'
  } catch (cause) {
    error.value = getErrorMessage(cause, '同步燕文产品目录失败')
  } finally {
    syncingProducts.value = false
  }
}

async function syncCountries() {
  if (!canManage.value) return
  syncingCountries.value = true
  error.value = ''
  message.value = ''
  countrySyncSummary.value = null
  try {
    countrySyncSummary.value = await yanwenLogisticsAdminApi.syncYanwenCountries({
      environment: environment.value,
      endpoint: form.value.endpoint,
      user_id: form.value.user_id,
      api_token: form.value.api_token,
    })
    countryCatalogLoaded.value = true
    message.value = '燕文通达国家目录同步完成。'
  } catch (cause) {
    error.value = getErrorMessage(cause, '同步燕文国家目录失败')
  } finally {
    syncingCountries.value = false
  }
}

async function syncWarehouses() {
  if (!canManage.value) return
  syncingWarehouses.value = true
  error.value = ''
  message.value = ''
  warehouseSyncSummary.value = null
  try {
    warehouseSyncSummary.value = await yanwenLogisticsAdminApi.syncYanwenWarehouses({
      environment: environment.value,
      endpoint: form.value.endpoint,
      user_id: form.value.user_id,
      api_token: form.value.api_token,
    })
    warehouseCatalogLoaded.value = true
    message.value = '燕文交货仓目录同步完成。'
  } catch (cause) {
    error.value = getErrorMessage(cause, '同步燕文交货仓目录失败')
  } finally {
    syncingWarehouses.value = false
  }
}

watch(environment, () => {
  void load()
  void loadCountryCatalogStatus()
  void loadWarehouseCatalogStatus()
  void loadProductCatalogStatus()
})

onMounted(() => {
  void load()
  void loadCountryCatalogStatus()
  void loadWarehouseCatalogStatus()
  void loadProductCatalogStatus()
})
</script>

<template>
  <section class="space-y-5 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
    <div>
      <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Yanwen Gateway Console</p>
      <h2 class="mt-1 text-lg font-black">网关配置</h2>
      <p class="mt-1 text-xs leading-5 text-muted-foreground">保存燕文 FAT 或 PRD 网关凭据，并通过 common.country.getlist 验证签名和网络连通性。凭据只在后端加密保存。</p>
    </div>

    <p v-if="error" role="alert" class="rounded-xl border border-red-500/40 bg-red-500/10 px-3 py-3 text-sm font-medium text-red-700 dark:text-red-300">{{ error }}</p>
    <p v-if="message" role="status" class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-xs text-emerald-700 dark:text-emerald-300">{{ message }}</p>

    <div class="grid gap-3 sm:grid-cols-2">
      <label class="space-y-1.5">
        <span class="text-xs font-bold">环境</span>
        <select v-model="environment" :disabled="loading || saving || pinging || syncingCountries || syncingWarehouses || syncingProducts" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm">
          <option value="fat">FAT 测试</option>
          <option value="production">PRD 生产</option>
        </select>
      </label>
      <label class="space-y-1.5">
        <span class="text-xs font-bold">官方网关地址</span>
        <Input v-model="form.endpoint" readonly />
      </label>
      <label class="space-y-1.5">
        <span class="text-xs font-bold">user_id（客户商户号）</span>
        <Input v-model="form.user_id" :disabled="!canManage || loading" autocomplete="off" :placeholder="form.user_id_configured ? '已保存，留空保持不变' : '输入燕文客户商户号'" />
      </label>
      <label class="space-y-1.5">
        <span class="text-xs font-bold">apitoken（制单账号秘钥）</span>
        <Input v-model="form.api_token" type="password" :disabled="!canManage || loading" autocomplete="new-password" :placeholder="form.api_token_configured ? '已保存，留空保持不变' : '输入燕文 apitoken'" />
      </label>
    </div>

    <label class="flex items-center gap-2 text-xs font-bold">
      <input v-model="form.enabled" type="checkbox" :disabled="!canManage || loading" class="size-4 rounded border-border" />
      启用当前环境配置
    </label>

    <div class="flex flex-wrap gap-2">
      <Button :disabled="!canSave || loading || saving || pinging || syncingCountries || syncingWarehouses || syncingProducts" @click="save">
        <Save class="size-3.5" />{{ saving ? '保存中…' : '保存配置' }}
      </Button>
      <Button variant="outline" :disabled="!canManage || loading || saving || pinging || syncingCountries || syncingWarehouses || syncingProducts" @click="ping">
        <RefreshCw class="size-3.5" :class="{ 'animate-spin': pinging }" />{{ pinging ? 'Ping 中…' : '执行网关 Ping' }}
      </Button>
      <Button variant="outline" :disabled="!canManage || loading || saving || pinging || syncingCountries || syncingWarehouses || syncingProducts" @click="syncCountries">
        <RefreshCw class="size-3.5" :class="{ 'animate-spin': syncingCountries }" />{{ syncingCountries ? '同步中…' : '同步通达国家' }}
      </Button>
      <Button variant="outline" :disabled="!canManage || loading || saving || pinging || syncingCountries || syncingWarehouses || syncingProducts" @click="syncWarehouses">
        <RefreshCw class="size-3.5" :class="{ 'animate-spin': syncingWarehouses }" />{{ syncingWarehouses ? '同步中…' : '同步交货仓' }}
      </Button>
      <Button variant="outline" :disabled="!canManage || loading || saving || pinging || syncingCountries || syncingWarehouses || syncingProducts" @click="syncProducts">
        <RefreshCw class="size-3.5" :class="{ 'animate-spin': syncingProducts }" />{{ syncingProducts ? '同步中…' : '同步已开通产品' }}
      </Button>
    </div>

    <div v-if="pingResult" role="status" class="flex items-center gap-2 rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
      <CheckCircle2 class="size-4 shrink-0" />
      <span>{{ pingResult.message }} · {{ pingResult.latency_ms }} ms</span>
    </div>
    <div v-if="productSyncSummary" role="status" class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
      产品目录已同步：扫描 {{ productSyncSummary.scanned }} 项，新增 {{ productSyncSummary.added }} 项，更新 {{ productSyncSummary.updated }} 项。
    </div>
    <div v-if="countrySyncSummary" role="status" class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
      国家目录已同步：扫描 {{ countrySyncSummary.scanned }} 项，新增 {{ countrySyncSummary.added }} 项，更新 {{ countrySyncSummary.updated }} 项。
    </div>
    <div v-if="warehouseSyncSummary" role="status" class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
      交货仓目录已同步：扫描 {{ warehouseSyncSummary.scanned }} 项，新增 {{ warehouseSyncSummary.added }} 项，更新 {{ warehouseSyncSummary.updated }} 项。
    </div>
  </section>

  <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
    <div>
      <h2 class="text-base font-black">主数据同步</h2>
      <p class="mt-1 text-xs leading-5 text-muted-foreground">国家、交货仓与已开通产品可通过上方按钮从燕文官方目录同步；服务集合只允许选择已缓存的官方产品。</p>
    </div>
    <div class="grid gap-3 md:grid-cols-3">
      <article v-for="item in [
        { label: '通达国家', api: 'common.country.getlist', detail: '国家代码与中英文名称' },
        { label: '交货仓', api: 'common.warehouse.getlist', detail: '燕文揽收仓代码、名称和区域' },
        { label: '已开通产品', api: 'express.channel.getlist', detail: '当前商户协议下可用的产品 ID' },
      ]" :key="item.api" class="rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4">
        <div class="flex items-center justify-between gap-2">
          <p class="text-sm font-black">{{ item.label }}</p>
          <span :class="((item.api === 'common.country.getlist' && (countrySyncSummary || countryCatalogLoaded)) || (item.api === 'common.warehouse.getlist' && (warehouseSyncSummary || warehouseCatalogLoaded)) || (item.api === 'express.channel.getlist' && (productSyncSummary || productCatalogLoaded))) ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' : 'bg-muted text-muted-foreground'" class="rounded-full px-2 py-1 text-[10px] font-bold">{{ ((item.api === 'common.country.getlist' && (countrySyncSummary || countryCatalogLoaded)) || (item.api === 'common.warehouse.getlist' && (warehouseSyncSummary || warehouseCatalogLoaded)) || (item.api === 'express.channel.getlist' && (productSyncSummary || productCatalogLoaded))) ? '已同步' : '待接入' }}</span>
        </div>
        <p class="mt-2 font-mono text-[10px] text-muted-foreground">{{ item.api }}</p>
        <p class="mt-2 text-xs leading-5 text-muted-foreground">{{ item.detail }}</p>
      </article>
    </div>
  </section>
</template>
