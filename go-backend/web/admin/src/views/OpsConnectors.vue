<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 连接器中心"
      description="网页登录授权并自动绑定 Cloudflare、Hostinger 资源；发现和同步只读，不直接执行发布或 DNS 写入。"
    >
      <template #actions>
        <select
          v-model="environmentFilter"
          class="h-9 rounded-md border bg-background px-3 text-sm"
          aria-label="筛选连接器环境"
          :disabled="loading || Boolean(oauthProvider)"
          @change="changeEnvironment"
        >
          <option v-for="option in opsConnectorEnvironmentOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
          <option value="">全部环境</option>
        </select>
        <Button variant="outline" :disabled="loading || Boolean(oauthProvider)" @click="loadConnectors">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button
          v-if="canEdit"
          variant="outline"
          :disabled="Boolean(oauthProvider) || !environmentFilter"
          title="选择具体环境后，网页登录并自动绑定 Cloudflare 账号与已有域名"
          @click="startOAuth('cloudflare')"
        >
          <LoaderCircle v-if="oauthProvider === 'cloudflare'" class="size-4 animate-spin" />
          <Cloud v-else class="size-4" />
          连接 Cloudflare
        </Button>
        <Button
          v-if="canEdit"
          variant="outline"
          :disabled="Boolean(oauthProvider) || !environmentFilter"
          title="选择具体环境后，网页登录并自动绑定 Hostinger 服务器与 Docker 项目"
          @click="startOAuth('hostinger')"
        >
          <LoaderCircle v-if="oauthProvider === 'hostinger'" class="size-4 animate-spin" />
          <Server v-else class="size-4" />
          连接 Hostinger
        </Button>
        <Button
          v-if="canEdit"
          :disabled="Boolean(oauthProvider) || !environmentFilter"
          title="选择具体环境后，依次网页登录 Hostinger 和 Cloudflare，并自动绑定服务器、项目与域名"
          @click="connectAll"
        >
          <LoaderCircle v-if="oauthProvider === 'all'" class="size-4 animate-spin" />
          <Link2 v-else class="size-4" />
          一键连接全部
        </Button>
        <Button v-if="canEdit" @click="openCreate">
          <Plus class="size-4" />
          新增连接
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 md:grid-cols-4">
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已登记连接</p>
        <p class="mt-2 text-2xl font-black">{{ connectors.length }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">当前启用</p>
        <p class="mt-2 text-2xl font-black text-emerald-600">{{ enabledCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">凭据已配置</p>
        <p class="mt-2 text-2xl font-black text-primary">{{ credentialCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">需处理状态</p>
        <p class="mt-2 text-2xl font-black text-amber-600">{{ attentionCount }}</p>
      </div>
    </section>

    <AdminTablePanel :loading="loading" :batch-visible="false">
      <Table class="min-w-[1120px]">
        <TableHeader>
          <TableRow>
            <TableHead>连接名称</TableHead>
            <TableHead>提供商 / 环境</TableHead>
            <TableHead>认证方式</TableHead>
            <TableHead>凭据状态</TableHead>
            <TableHead>范围</TableHead>
            <TableHead>测试状态</TableHead>
            <TableHead>状态</TableHead>
            <TableHead class="w-48 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="connectors.length === 0" :colspan="8">
            <div class="py-8 text-center text-xs text-muted-foreground">还没有登记连接器</div>
          </TableEmpty>
          <TableRow v-for="connector in connectors" :key="connector.id">
            <TableCell>
              <div class="min-w-0">
                <p class="truncate font-bold">{{ connector.name }}</p>
                <p class="mt-1 truncate text-[10px] text-muted-foreground" :title="connector.endpoint">
                  {{ connector.endpoint || defaultEndpointLabel(connector.provider) }}
                </p>
              </div>
            </TableCell>
            <TableCell>
              <p class="text-xs font-bold">{{ providerLabel(connector.provider) }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ environmentLabel(connector.environment) }}</p>
            </TableCell>
            <TableCell class="text-xs">{{ authTypeLabel(connector.auth_type) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="connector.credential_configured ? 'green' : 'amber'">
                {{ connector.credential_configured ? '已配置' : '未配置' }}
              </AdminStatusBadge>
              <p v-if="connector.credential_fields?.length" class="mt-1 text-[10px] text-muted-foreground">
                {{ connector.credential_fields.join(', ') }}
              </p>
              <p v-else-if="connector.credential_ref" class="mt-1 truncate text-[10px] text-muted-foreground">
                {{ connector.credential_ref }}
              </p>
            </TableCell>
            <TableCell class="max-w-44 truncate text-xs" :title="connector.scopes">
              {{ connector.scopes || '-' }}
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="testTone(connector.last_test_status)">
                {{ testLabel(connector.last_test_status) }}
              </AdminStatusBadge>
              <p v-if="connector.last_tested_at" class="mt-1 text-[10px] text-muted-foreground">
                {{ formatDate(connector.last_tested_at) }}
              </p>
              <p v-if="connector.last_error" class="mt-1 max-w-36 truncate text-[10px] text-rose-600" :title="connector.last_error">
                {{ connector.last_error }}
              </p>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(connector.status)">
                {{ statusLabel(connector.status) }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="flex justify-end gap-1">
                <Button
                  v-if="canEdit"
                  size="icon"
                  variant="ghost"
                  :title="`测试连接 ${connector.name}`"
                  :disabled="testingID === connector.id"
                  @click="testConnector(connector)"
                >
                  <LoaderCircle v-if="testingID === connector.id" class="size-4 animate-spin" />
                  <PlugZap v-else class="size-4" />
                </Button>
                <Button
                  v-if="canEdit && oauthProviderFor(connector)"
                  size="icon"
                  variant="ghost"
                  :title="`重新授权 ${connector.name}`"
                  :disabled="Boolean(oauthProvider)"
                  @click="reauthorizeConnector(connector)"
                >
                  <KeyRound class="size-4" />
                </Button>
                <Button
                  v-if="canEdit"
                  size="icon"
                  variant="ghost"
                  :title="connector.enabled ? '停用连接' : '启用连接'"
                  @click="toggleConnector(connector)"
                >
                  <Power class="size-4" />
                </Button>
                <Button v-if="canEdit" size="icon" variant="ghost" title="编辑连接" @click="openEdit(connector)">
                  <Pencil class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <OpsConnectorBindingFormDialog
      :open="dialogOpen"
      :form="form"
      :saving="saving"
      :can-edit="canEdit"
      @update:open="dialogOpen = $event"
      @save="saveConnector"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { Cloud, KeyRound, Link2, LoaderCircle, Pencil, PlugZap, Plus, Power, RefreshCw, Server } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import OpsConnectorBindingFormDialog from '@/components/admin/ops/OpsConnectorBindingFormDialog.vue'
import {
  assignOpsConnectorForm,
  emptyOpsConnectorForm,
  opsConnectorAuthTypeOptions,
  opsConnectorEnvironmentOptions,
  opsConnectorProviderOptions,
  opsConnectorStatusOptions,
  type OpsConnectorForm,
} from '@/components/admin/ops/opsConnectorBindingForm'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import opsApi, {
  type OpsConnector,
  type OpsConnectorPayload,
  type OpsConnectorTestResult,
  type OpsEnvironment,
} from '@/api/ops'
import { readOpsEnvironmentQuery, withOpsEnvironmentQuery } from '@/lib/opsEnvironment'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('ops:connector:edit'))
const connectors = ref<OpsConnector[]>([])
const loading = ref(false)
const saving = ref(false)
const testingID = ref(0)
const oauthProvider = ref('')
const dialogOpen = ref(false)
const environmentFilter = ref<OpsEnvironment | ''>(readOpsEnvironmentQuery(route.query.environment))

const form = reactive<OpsConnectorForm>(emptyOpsConnectorForm())

const enabledCount = computed(() => connectors.value.filter((connector) => connector.enabled).length)
const credentialCount = computed(() => connectors.value.filter((connector) => connector.credential_configured).length)
const attentionCount = computed(() => connectors.value.filter((connector) => ['pending', 'error'].includes(connector.status)).length)

const assignForm = (connector?: Partial<OpsConnector>): void => {
  assignOpsConnectorForm(form, connector)
  if (!connector && environmentFilter.value) {
    form.environment = environmentFilter.value
  }
}

const loadConnectors = async (): Promise<void> => {
  loading.value = true
  try {
    const data = await opsApi.listConnectors(environmentFilter.value || undefined)
    connectors.value = Array.isArray(data?.connectors) ? data.connectors : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接器列表加载失败')
  } finally {
    loading.value = false
  }
}

const changeEnvironment = (): void => {
  void router.replace({ query: withOpsEnvironmentQuery(route.query, environmentFilter.value) })
  void loadConnectors()
}

const openCreate = (): void => {
  assignForm()
  dialogOpen.value = true
}

const openEdit = (connector: OpsConnector): void => {
  assignForm(connector)
  dialogOpen.value = true
}

const saveConnector = async (): Promise<void> => {
  if (!form.name.trim()) {
    toast.error('请输入连接名称')
    return
  }
  if (form.auth_type === 'environment' && !form.credential_ref.trim()) {
    toast.error('请输入后端环境变量名')
    return
  }
  saving.value = true
  try {
    const credentials = Object.fromEntries(
      Object.entries(form.credentials).filter(([, value]) => String(value || '').trim())
    )
    const payload: OpsConnectorPayload = {
      name: form.name.trim(),
      provider: form.provider,
      environment: form.environment,
      endpoint: form.endpoint.trim(),
      auth_type: form.auth_type,
      credential_ref: form.credential_ref.trim(),
      credentials,
      scopes: form.scopes.trim(),
      status: form.status,
      enabled: form.enabled,
      notes: form.notes.trim(),
    }
    if (form.id) {
      await opsApi.updateConnector(form.id, payload)
      toast.success('连接器已保存')
    } else {
      await opsApi.createConnector(payload)
      toast.success('连接器已创建')
    }
    dialogOpen.value = false
    await loadConnectors()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接器保存失败')
  } finally {
    saving.value = false
  }
}

const toggleConnector = async (connector: OpsConnector): Promise<void> => {
  try {
    const updated = await opsApi.setConnectorEnabled(connector.id, !connector.enabled)
    const index = connectors.value.findIndex((item) => item.id === connector.id)
    if (index >= 0) connectors.value[index] = updated
    toast.success(updated.enabled ? '连接器已启用' : '连接器已停用')
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接器状态更新失败')
  }
}

const testConnector = async (connector: OpsConnector): Promise<void> => {
  testingID.value = connector.id
  try {
    const result = await opsApi.testConnector(connector.id)
    toast[result.success ? 'success' : 'error'](result.message || (result.success ? '连接测试成功' : '连接测试失败'))
    applyTestResult(connector, result)
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接测试失败')
  } finally {
    testingID.value = 0
  }
}

const applyTestResult = (connector: OpsConnector, result: OpsConnectorTestResult): void => {
  const index = connectors.value.findIndex((item) => item.id === connector.id)
  if (index < 0) return
  connectors.value[index] = {
    ...connector,
    credential_configured: result.credential_configured,
    last_test_status: result.success ? 'success' : 'failed',
    last_tested_at: result.checked_at,
    last_error: result.success ? '' : result.message,
    status: result.success ? 'active' : connector.enabled ? 'error' : 'disabled',
  }
}

const oauthProviderFor = (connector: OpsConnector): 'cloudflare' | 'hostinger' | '' => (
  connector.provider === 'cloudflare' || connector.provider === 'hostinger'
    ? connector.provider
    : ''
)

const reauthorizeConnector = async (connector: OpsConnector): Promise<void> => {
  const provider = oauthProviderFor(connector)
  if (!provider) return
  await startOAuth(
    provider,
    connector.id,
    isOpsEnvironment(connector.environment) ? connector.environment : 'production',
  )
}

const isOpsEnvironment = (value: string | null): value is OpsEnvironment => (
  value === 'production' || value === 'staging' || value === 'test' || value === 'local'
)

const oauthReturnPath = (environment: OpsEnvironment, nextProvider?: 'cloudflare'): string => {
  const query = new URLSearchParams({ ops_oauth_environment: environment })
  if (nextProvider) query.set('ops_oauth_next', nextProvider)
  return `/ops/connectors?${query.toString()}`
}

const startOAuth = async (
  provider: 'cloudflare' | 'hostinger',
  connectorID?: number,
  environment: OpsEnvironment | '' = environmentFilter.value,
): Promise<void> => {
  const resolvedEnvironment = environment || 'production'
  oauthProvider.value = provider
  try {
    const result = await opsApi.startConnectorOAuth(
      provider,
      connectorID,
      oauthReturnPath(resolvedEnvironment),
      resolvedEnvironment,
    )
    window.location.assign(result.authorization_url)
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '网页登录授权启动失败')
    oauthProvider.value = ''
  }
}

const connectAll = async (): Promise<void> => {
  const environment = environmentFilter.value || 'production'
  oauthProvider.value = 'all'
  try {
    const result = await opsApi.startConnectorOAuth(
      'hostinger',
      undefined,
      oauthReturnPath(environment, 'cloudflare'),
      environment,
    )
    window.location.assign(result.authorization_url)
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '一键网页登录授权启动失败')
    oauthProvider.value = ''
  }
}

const handleOAuthReturn = async (): Promise<void> => {
  const query = new URLSearchParams(window.location.search)
  const status = query.get('ops_oauth_status')
  if (!status) return
  const message = query.get('ops_oauth_message') || (status === 'connected' ? '网页登录授权成功' : '网页登录授权失败')
  const nextProvider = query.get('ops_oauth_next')
  const oauthEnvironment = query.get('ops_oauth_environment')
  if (isOpsEnvironment(oauthEnvironment)) {
    environmentFilter.value = oauthEnvironment
  }
  const connected = status === 'connected' || status === 'connected_with_warnings'
  if (connected) {
    const provider = query.get('ops_oauth_provider') === 'hostinger' ? 'Hostinger' : 'Cloudflare'
    const vps = query.get('ops_oauth_bound_vps') || '0'
    const projects = query.get('ops_oauth_bound_projects') || '0'
    const domains = query.get('ops_oauth_bound_domains') || '0'
    const warnings = query.get('ops_oauth_binding_warnings') || '0'
    const summary = `${provider} 已连接，自动绑定 ${vps} 台服务器、${projects} 个项目、${domains} 个域名；${warnings} 项绑定或同步需要复核`
    if (status === 'connected_with_warnings') {
      toast.warning(message || summary)
    } else {
      toast.success(summary)
    }
  } else {
    toast.error(message)
  }
  const cleanURL = `${window.location.pathname}${window.location.hash}`
  window.history.replaceState({}, document.title, cleanURL)
  if (connected && nextProvider === 'cloudflare') {
    toast.info('Hostinger 已连接，继续授权 Cloudflare')
    await startOAuth('cloudflare', undefined, isOpsEnvironment(oauthEnvironment) ? oauthEnvironment : environmentFilter.value)
  }
}

const optionLabel = (options: Array<{ value: string; label: string }>, value: string): string => (
  options.find((option) => option.value === value)?.label || value || '-'
)
const providerLabel = (value: string): string => optionLabel(opsConnectorProviderOptions, value)
const environmentLabel = (value: string): string => optionLabel(opsConnectorEnvironmentOptions, value)
const authTypeLabel = (value: string): string => optionLabel(opsConnectorAuthTypeOptions, value)
const statusLabel = (value: string): string => optionLabel(opsConnectorStatusOptions, value)
const testLabel = (value: string): string => value === 'success' ? '成功' : value === 'failed' ? '失败' : '未测试'
const defaultEndpointLabel = (provider: string): string => provider === 'cloudflare' ? 'Cloudflare Token 验证接口' : '未设置测试接口'
const formatDate = (value: string): string => new Date(value).toLocaleString()
const statusTone = (value: string): AdminStatusTone => {
  if (value === 'active') return 'green'
  if (value === 'error') return 'coral'
  if (value === 'pending') return 'amber'
  return 'gray'
}
const testTone = (value: string): AdminStatusTone => value === 'success' ? 'green' : value === 'failed' ? 'coral' : 'gray'

watch(() => route.query.environment, (value) => {
  const nextEnvironment = readOpsEnvironmentQuery(value)
  if (nextEnvironment === environmentFilter.value) return
  environmentFilter.value = nextEnvironment
  void loadConnectors()
})

onMounted(async () => {
  await handleOAuthReturn()
  await loadConnectors()
})
</script>
