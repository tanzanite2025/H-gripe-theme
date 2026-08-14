<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / VPS 中心"
      description="维护 Hostinger VPS 资源基线和观察状态；这里不创建、删除或部署 VPS。"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="loadVPS">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEdit" @click="openCreate">
          <Plus class="size-4" />
          新增 VPS
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 md:grid-cols-4">
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已登记 VPS</p>
        <p class="mt-2 text-2xl font-black">{{ vpsList.length }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">当前启用</p>
        <p class="mt-2 text-2xl font-black text-emerald-600">{{ enabledCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已确认观察</p>
        <p class="mt-2 text-2xl font-black text-primary">{{ observedCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">需处理状态</p>
        <p class="mt-2 text-2xl font-black text-amber-600">{{ attentionCount }}</p>
      </div>
    </section>

    <AdminTablePanel :loading="loading" :batch-visible="false">
      <Table class="min-w-[1280px]">
        <TableHeader>
          <TableRow>
            <TableHead>资源名称</TableHead>
            <TableHead>Provider / ID</TableHead>
            <TableHead>Hostinger 实际</TableHead>
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
              <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ vps.provider_resource_id || '-' }}</p>
            </TableCell>
            <TableCell>
              <p class="font-mono text-xs">{{ vps.observed_state || '未同步' }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ vps.observed_source || '无同步来源' }}</p>
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
              <p class="font-mono text-xs">{{ vps.hostname || '-' }}</p>
              <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ vps.ipv4 || '-' }}</p>
            </TableCell>
            <TableCell>
              <p class="text-xs">{{ environmentLabel(vps.environment) }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ vps.operating_system || '未登记系统' }}</p>
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

    <Dialog v-model:open="dialogOpen">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑 VPS 绑定' : '新增 VPS 绑定' }}</DialogTitle>
          <DialogDescription>
            这里记录 Hostinger 或其他提供商的资源基线。观察状态只有同步后才代表实际状态。
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-4" @submit.prevent="saveVPS">
          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField label="资源名称" required>
              <Input v-model="form.name" placeholder="Hostinger Production VPS" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="提供商" required>
              <select v-model="form.provider" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in providerOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="环境" required>
              <select v-model="form.environment" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in environmentOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="连接器">
              <select v-model="form.connector_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option :value="null">未绑定连接器</option>
                <option v-for="connector in connectors" :key="connector.id" :value="connector.id">{{ connector.name }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="Hostinger VPS ID / 资源 ID">
              <Input v-model="form.provider_resource_id" placeholder="1834903" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="主机名">
              <Input v-model="form.hostname" placeholder="srv1834903.hstgr.cloud" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="IPv4">
              <Input v-model="form.ipv4" placeholder="2.25.85.201" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="区域">
              <Input v-model="form.region" placeholder="Hostinger region" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="操作系统">
              <Input v-model="form.operating_system" placeholder="Ubuntu 24.04 LTS" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="期望状态">
              <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="备注" class="md:col-span-2">
              <Textarea v-model="form.notes" class="min-h-24" placeholder="记录区域、维护窗口、资源边界或确认来源。" :disabled="saving" />
            </AdminFormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving || !canEdit">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              <Save v-else class="size-4" />
              {{ saving ? '保存中' : '保存 VPS' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Pencil, Plus, Power, RefreshCw, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import opsApi, { type OpsVPSPayload } from '@/api/ops'
import { useAuthStore } from '@/stores/auth'

interface OpsVPS {
  id: number
  name: string
  provider: string
  environment: string
  connector_id?: number | null
  provider_resource_id: string
  hostname: string
  ipv4: string
  region: string
  operating_system: string
  status: string
  observed_status: string
  observed_state: string
  observed_source: string
  observed_hostname: string
  observed_ipv4: string
  observed_operating_system: string
  observed_plan: string
  observed_region: string
  enabled: boolean
  last_observed_at?: string
  last_error: string
  notes: string
}

interface OpsConnector {
  id: number
  name: string
}

interface OpsVPSForm {
  id: number
  name: string
  provider: string
  environment: string
  connector_id: number | null
  provider_resource_id: string
  hostname: string
  ipv4: string
  region: string
  operating_system: string
  status: string
  enabled: boolean
  notes: string
}

const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('ops:vps:edit'))
const canSync = computed(() => authStore.hasPermission('ops:vps:sync'))
const vpsList = ref<OpsVPS[]>([])
const connectors = ref<OpsConnector[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const syncingId = ref(0)

const providerOptions = [
  { value: 'hostinger', label: 'Hostinger' },
  { value: 'other', label: '其他' },
]
const environmentOptions = [
  { value: 'production', label: '生产' },
  { value: 'staging', label: '预发布' },
  { value: 'test', label: '测试' },
  { value: 'local', label: '本地' },
]
const statusOptions = [
  { value: 'active', label: '正常' },
  { value: 'pending', label: '待确认' },
  { value: 'disabled', label: '已停用' },
  { value: 'drifted', label: '配置漂移' },
  { value: 'error', label: '错误' },
]
const observedOptions = [
  { value: 'healthy', label: '健康' },
  { value: 'degraded', label: '降级' },
  { value: 'unknown', label: '未同步' },
  { value: 'offline', label: '离线' },
]

const emptyForm = (): OpsVPSForm => ({
  id: 0,
  name: '',
  provider: 'hostinger',
  environment: 'production',
  connector_id: null,
  provider_resource_id: '',
  hostname: '',
  ipv4: '',
  region: '',
  operating_system: '',
  status: 'pending',
  enabled: true,
  notes: '',
})
const form = reactive<OpsVPSForm>(emptyForm())

const enabledCount = computed(() => vpsList.value.filter((vps) => vps.enabled).length)
const observedCount = computed(() => vpsList.value.filter((vps) => vps.observed_status !== 'unknown').length)
const attentionCount = computed(() => vpsList.value.filter((vps) => ['pending', 'drifted', 'error'].includes(vps.status) || ['degraded', 'offline'].includes(vps.observed_status)).length)

const assignForm = (vps?: Partial<OpsVPS>): void => {
  Object.assign(form, emptyForm(), vps || {})
  form.connector_id = vps?.connector_id ?? null
}

const loadVPS = async (): Promise<void> => {
  loading.value = true
  try {
    const [vpsData, connectorData] = await Promise.all([opsApi.listVPS(), opsApi.listConnectors()])
    vpsList.value = Array.isArray(vpsData?.vps) ? vpsData.vps : []
    connectors.value = Array.isArray(connectorData?.connectors) ? connectorData.connectors : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'VPS 列表加载失败')
  } finally {
    loading.value = false
  }
}

const openCreate = (): void => {
  assignForm()
  dialogOpen.value = true
}

const openEdit = (vps: OpsVPS): void => {
  assignForm(vps)
  dialogOpen.value = true
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
    await loadVPS()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || `${vps.name} 同步失败`)
  } finally {
    syncingId.value = 0
  }
}

const optionLabel = (options: Array<{ value: string; label: string }>, value: string): string => (
  options.find((option) => option.value === value)?.label || value || '-'
)
const providerLabel = (value: string): string => optionLabel(providerOptions, value)
const environmentLabel = (value: string): string => optionLabel(environmentOptions, value)
const statusLabel = (value: string): string => optionLabel(statusOptions, value)
const observedLabel = (value: string): string => optionLabel(observedOptions, value)
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

onMounted(loadVPS)
</script>
