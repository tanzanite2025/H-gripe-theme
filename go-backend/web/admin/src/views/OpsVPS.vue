<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / VPS 中心"
      description="维护各 Provider 的 VPS 资源基线和观察状态；这里不创建、删除或部署 VPS。"
    >
      <template #actions>
        <select
          v-model="environmentFilter"
          class="h-9 rounded-md border bg-background px-3 text-sm"
          aria-label="筛选 VPS 环境"
          :disabled="loading"
          :options="opsVPSEnvironmentOptions"
          @update:model-value="changeEnvironment"
        />
        <Button variant="outline" :disabled="loading" @click="refreshVPSPage">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button v-if="canEdit" @click="openCreate">
          <Plus class="size-4" />
          新增 VPS
        </Button>
      </template>
    </AdminPageHeader>

    <OpsSummaryCards :items="summaryCards" />

    <AdminTablePanel :loading="loading" :batch-visible="false">
      <Table class="min-w-[1280px]">
        <TableHeader>
          <TableRow>
            <TableHead>资源名称</TableHead>
            <TableHead>Provider / ID</TableHead>
            <TableHead>Provider 实际</TableHead>
            <TableHead>主机地址</TableHead>
            <TableHead>环境 / 系统</TableHead>
            <TableHead>期望状态</TableHead>
            <TableHead>观察状态</TableHead>
            <TableHead class="w-40 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="vpsList.length === 0" :colspan="8">
            <div class="py-8 text-center text-xs text-muted-foreground">还没有登记 VPS</div>
          </TableEmpty>
          <TableRow v-for="vps in vpsList" :key="vps.id">
            <TableCell>
              <p class="truncate font-bold">{{ vps.name }}</p>
              <p v-if="vps.region" class="mt-1 text-[10px] text-muted-foreground">{{ vps.region }}</p>
            </TableCell>
            <TableCell>
              <p class="text-xs font-bold">{{ providerLabel(vps.provider) }}</p>
 <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ vps.provider_resource_id || '-'}}</p>
            </TableCell>
            <TableCell>
 <p class="font-mono text-xs">{{ vps.observed_state || '未同步'}}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">{{ vps.observed_source || '无同步来源'}}</p>
              <p v-if="vps.observed_hostname || vps.observed_ipv4" class="mt-1 truncate font-mono text-[10px] text-muted-foreground">
                {{ vps.observed_hostname || '-' }} / {{ vps.observed_ipv4 || '-' }}
              </p>
              <p v-if="vps.observed_operating_system" class="mt-1 truncate text-[10px] text-muted-foreground">
                {{ vps.observed_operating_system }}
              </p>
              <p v-if="vps.observed_plan || vps.observed_region" class="mt-1 truncate text-[10px] text-muted-foreground">
                {{ vps.observed_plan || '-' }} / {{ vps.observed_region || '-' }}
              </p>
            </TableCell>
            <TableCell>
 <p class="font-mono text-xs">{{ vps.hostname || '-'}}</p>
 <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ vps.ipv4 || '-'}}</p>
            </TableCell>
            <TableCell>
              <p class="text-xs">{{ environmentLabel(vps.environment) }}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">{{ vps.operating_system || '未登记系统'}}</p>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(vps.status)">{{ statusLabel(vps.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="observedTone(vps.observed_status)">
                {{ observedLabel(vps.observed_status) }}
              </AdminStatusBadge>
              <p v-if="vps.last_observed_at" class="mt-1 text-[10px] text-muted-foreground">
                {{ formatDate(vps.last_observed_at) }}
              </p>
              <p v-else class="mt-1 text-[10px] text-muted-foreground">尚未同步</p>
              <p v-if="vps.last_error" class="mt-1 max-w-40 truncate text-[10px] text-rose-600" :title="vps.last_error">
                {{ vps.last_error }}
              </p>
            </TableCell>
            <TableCell class="text-right">
              <div class="flex justify-end gap-1">
                <Button
                  v-if="canSync && vps.provider === 'hostinger'"
                  size="icon"
                  variant="ghost"
                  :title="`同步 ${vps.name}`"
                  :disabled="syncingId === vps.id"
                  @click="syncVPS(vps)"
                >
                  <LoaderCircle v-if="syncingId === vps.id" class="size-4 animate-spin" />
                  <RefreshCw v-else class="size-4" />
                </Button>
                <Button
                  v-if="canEdit"
                  size="icon"
                  variant="ghost"
                  :title="vps.enabled ? '停用 VPS 记录' : '启用 VPS 记录'"
                  @click="toggleVPS(vps)"
                >
                  <Power class="size-4" />
                </Button>
                <Button v-if="canEdit" size="icon" variant="ghost" title="编辑 VPS" @click="openEdit(vps)">
                  <Pencil class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <OpsVPSBindingFormDialog
      :open="dialogOpen"
      :form="form"
      :connectors="connectors"
      :saving="saving"
      :can-edit="canEdit"
      @update:open="dialogOpen = $event"
      @save="saveVPS"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { LoaderCircle, Pencil, Plus, Power, RefreshCw } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import OpsEnvironmentSelect from '@/components/admin/ops/OpsEnvironmentSelect.vue'
import OpsSummaryCards, { type OpsSummaryCardItem } from '@/components/admin/ops/OpsSummaryCards.vue'
import OpsVPSBindingFormDialog from '@/components/admin/ops/OpsVPSBindingFormDialog.vue'
import {
  assignOpsVPSForm,
  emptyOpsVPSForm,
  opsVPSEnvironmentOptions,
  opsVPSObservedOptions,
  opsVPSProviderOptions,
  opsVPSStatusOptions,
  type OpsVPSForm,
} from '@/components/admin/ops/opsVPSBindingForm'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import opsApi, { type OpsConnector, type OpsEnvironment, type OpsVPS, type OpsVPSPayload } from '@/api/ops'
import { readOpsEnvironmentQuery, withOpsEnvironmentQuery } from '@/lib/opsEnvironment'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('ops:vps:edit'))
const canSync = computed(() => authStore.hasPermission('ops:vps:sync'))
const vpsList = ref<OpsVPS[]>([])
const connectors = ref<OpsConnector[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const syncingId = ref(0)
const environmentFilter = ref<OpsEnvironment | ''>(readOpsEnvironmentQuery(route.query.environment))

const form = reactive<OpsVPSForm>(emptyOpsVPSForm())

const enabledCount = computed(() => vpsList.value.filter((vps) => vps.enabled).length)
const observedCount = computed(() => vpsList.value.filter((vps) => vps.observed_status !== 'unknown').length)
const attentionCount = computed(() => vpsList.value.filter((vps) => ['pending', 'drifted', 'error'].includes(vps.status) || ['degraded', 'offline'].includes(vps.observed_status)).length)
const summaryCards = computed<readonly OpsSummaryCardItem[]>(() => [
  { key: 'vps', label: '已登记 VPS', value: vpsList.value.length, detail: `当前启用 ${enabledCount.value}` },
  { key: 'enabled', label: '当前启用', value: enabledCount.value, detail: `共 ${vpsList.value.length} 台 VPS`, tone: 'green' },
  { key: 'observed', label: '已确认观察', value: observedCount.value, detail: `待观察 ${Math.max(vpsList.value.length - observedCount.value, 0)}`, tone: 'primary' },
  { key: 'attention', label: '需处理状态', value: attentionCount.value, detail: '期望状态或实际状态异常', tone: 'amber' },
])

const assignForm = (vps?: Partial<OpsVPS>): void => {
  assignOpsVPSForm(form, vps)
}

const loadVPS = async (): Promise<void> => {
  loading.value = true
  try {
    const vpsData = await opsApi.listVPS(environmentFilter.value || undefined)
    vpsList.value = Array.isArray(vpsData?.vps) ? vpsData.vps : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'VPS 列表加载失败')
  } finally {
    loading.value = false
  }
}

const loadConnectorOptions = async (): Promise<void> => {
  if (!canEdit.value) {
    connectors.value = []
    return
  }
  try {
    const connectorData = await opsApi.listConnectors()
    connectors.value = Array.isArray(connectorData?.connectors) ? connectorData.connectors : []
  } catch (error: any) {
    connectors.value = []
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接器选项加载失败，VPS 列表仍可使用')
  }
}

const changeEnvironment = (): void => {
  void router.replace({ query: withOpsEnvironmentQuery(route.query, environmentFilter.value) })
  void loadVPS()
}

const refreshVPSPage = async (): Promise<void> => {
  await Promise.all([
    loadVPS(),
    loadConnectorOptions(),
  ])
}

const openCreate = (): void => {
  assignForm()
  dialogOpen.value = true
  if (!connectors.value.length) void loadConnectorOptions()
}

const openEdit = (vps: OpsVPS): void => {
  assignForm(vps)
  dialogOpen.value = true
  if (!connectors.value.length) void loadConnectorOptions()
}

const saveVPS = async (): Promise<void> => {
  if (!form.name.trim()) {
    toast.error('请输入资源名称')
    return
  }
  saving.value = true
  try {
    const payload: OpsVPSPayload = {
      name: form.name.trim(),
      provider: form.provider,
      environment: form.environment,
      connector_id: form.connector_id || null,
      provider_resource_id: form.provider_resource_id.trim(),
      hostname: form.hostname.trim(),
      ipv4: form.ipv4.trim(),
      region: form.region.trim(),
      operating_system: form.operating_system.trim(),
      status: form.status,
      enabled: form.enabled,
      notes: form.notes.trim(),
    }
    if (form.id) {
      await opsApi.updateVPS(form.id, payload)
      toast.success('VPS 绑定已保存')
    } else {
      await opsApi.createVPS(payload)
      toast.success('VPS 绑定已创建')
    }
    dialogOpen.value = false
    await loadVPS()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'VPS 保存失败')
  } finally {
    saving.value = false
  }
}

const toggleVPS = async (vps: OpsVPS): Promise<void> => {
  try {
    const updated = await opsApi.setVPSEnabled(vps.id, !vps.enabled)
    const index = vpsList.value.findIndex((item) => item.id === vps.id)
    if (index >= 0) vpsList.value[index] = updated
    toast.success(updated.enabled ? 'VPS 记录已启用' : 'VPS 记录已停用')
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'VPS 状态更新失败')
  }
}

const syncVPS = async (vps: OpsVPS): Promise<void> => {
  syncingId.value = vps.id
  try {
    const result = await opsApi.syncVPS(vps.id)
    const tone = result.observed_status === 'healthy'
      ? 'success'
      : result.observed_status === 'degraded'
        ? 'warning'
        : result.observed_status === 'offline'
          ? 'error'
          : 'warning'
    toast[tone](result.message || `${vps.name} 同步完成`)
    const index = vpsList.value.findIndex((item) => item.id === vps.id)
    if (index >= 0) {
      vpsList.value[index] = {
        ...vps,
        observed_status: result.observed_status,
        observed_state: result.remote_state || '',
        observed_source: result.observed_source,
        observed_hostname: result.hostname || '',
        observed_ipv4: result.ipv4 || '',
        observed_operating_system: result.operating_system || '',
        observed_plan: result.observed_plan || '',
        observed_region: result.observed_region || '',
        last_observed_at: result.last_observed_at,
        last_error: result.observed_error || '',
      }
    }
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || `${vps.name} 同步失败`)
  } finally {
    syncingId.value = 0
  }
}

const optionLabel = (options: Array<{ value: string; label: string }>, value: string): string => (
  options.find((option) => option.value === value)?.label || value || '-'
)
const providerLabel = (value: string): string => optionLabel(opsVPSProviderOptions, value)
const environmentLabel = (value: string): string => optionLabel(opsVPSEnvironmentOptions, value)
const statusLabel = (value: string): string => optionLabel(opsVPSStatusOptions, value)
const observedLabel = (value: string): string => optionLabel(opsVPSObservedOptions, value)
const formatDate = (value: string): string => new Date(value).toLocaleString()
const statusTone = (value: string): AdminStatusTone => {
  if (value === 'active') return 'green'
  if (value === 'pending' || value === 'drifted') return 'amber'
  if (value === 'error') return 'coral'
  return 'gray'
}
const observedTone = (value: string): AdminStatusTone => {
  if (value === 'healthy') return 'green'
  if (value === 'degraded') return 'amber'
  if (value === 'offline') return 'coral'
  return 'gray'
}

watch(() => route.query.environment, (value) => {
  const nextEnvironment = readOpsEnvironmentQuery(value)
  if (nextEnvironment === environmentFilter.value) return
  environmentFilter.value = nextEnvironment
  void loadVPS()
})

onMounted(refreshVPSPage)
</script>
