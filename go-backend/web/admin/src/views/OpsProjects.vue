<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 项目中心"
      description="维护 Compose 项目、服务边界、当前版本和备份恢复记录；这里只做台账维护。"
    >
      <template #actions>
        <select
          v-model="environmentFilter"
          class="h-9 rounded-md border bg-background px-3 text-sm"
          aria-label="筛选项目环境"
          :disabled="loading"
          @change="changeEnvironment"
        >
          <option value="production">生产</option>
          <option value="staging">预发布</option>
          <option value="test">测试</option>
          <option value="local">本地</option>
          <option value="">全部环境</option>
        </select>
        <Button variant="outline" :disabled="loading" @click="refreshProjectsPage">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button v-if="canEdit" @click="openCreate">
          <Plus class="size-4" />
          新增项目
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 md:grid-cols-4">
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已登记项目</p>
        <p class="mt-2 text-2xl font-black">{{ projects.length }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">当前启用</p>
        <p class="mt-2 text-2xl font-black text-emerald-600">{{ enabledCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">健康状态</p>
        <p class="mt-2 text-2xl font-black text-primary">{{ healthyCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">需处理状态</p>
        <p class="mt-2 text-2xl font-black text-amber-600">{{ attentionCount }}</p>
      </div>
    </section>

    <AdminTablePanel :loading="loading" :batch-visible="false">
      <Table class="min-w-[1480px]">
        <TableHeader>
          <TableRow>
            <TableHead>项目</TableHead>
            <TableHead>VPS</TableHead>
            <TableHead>Compose / 网关</TableHead>
            <TableHead>Provider 实际</TableHead>
            <TableHead>服务边界</TableHead>
            <TableHead>当前版本</TableHead>
            <TableHead>健康 / 期望</TableHead>
            <TableHead>最近记录</TableHead>
            <TableHead class="w-40 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="projects.length === 0" :colspan="9">
            <div class="py-8 text-center text-xs text-muted-foreground">还没有登记项目</div>
          </TableEmpty>
          <TableRow v-for="project in projects" :key="project.id">
            <TableCell>
              <p class="truncate font-bold">{{ project.name }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ environmentLabel(project.environment) }}</p>
            </TableCell>
            <TableCell>
 <p class="text-xs font-bold">{{ project.vps_name || '未绑定 VPS'}}</p>
 <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ project.vps_hostname || project.vps_ipv4 || '-'}}</p>
              <p class="mt-1 truncate text-[10px] text-muted-foreground" :title="projectConnectorLabel(project)">
                连接器：{{ projectConnectorLabel(project) }}
              </p>
            </TableCell>
            <TableCell>
 <p class="font-mono text-xs">{{ project.compose_project_name || '-'}}</p>
 <p class="mt-1 truncate text-[10px] text-muted-foreground">{{ project.compose_source || '未登记 Compose 来源'}}</p>
              <p v-if="project.gateway_network || project.gateway_alias" class="mt-1 text-[10px] text-muted-foreground">
                {{ project.gateway_network || '-' }} / {{ project.gateway_alias || '-' }}
              </p>
            </TableCell>
            <TableCell>
 <p class="font-mono text-xs">{{ project.observed_state || '未同步'}}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">{{ project.observed_source || '无同步来源'}}</p>
              <p class="mt-1 font-mono text-[10px] text-muted-foreground">
                {{ containerSummary(project) }}
              </p>
            </TableCell>
            <TableCell>
 <p class="max-w-56 truncate text-xs" :title="project.services">{{ project.services || '-'}}</p>
              <p class="mt-1 max-w-56 truncate text-[10px] text-muted-foreground" :title="project.networks">
                网络：{{ project.networks || '-' }}
              </p>
              <p class="mt-1 max-w-56 truncate text-[10px] text-muted-foreground" :title="project.volumes">
                卷：{{ project.volumes || '-' }}
              </p>
              <p class="mt-1 max-w-56 truncate text-[10px] text-muted-foreground">
                Quick Buy：{{ projectQuickBuyRateLimitSummary(project) }}
              </p>
            </TableCell>
            <TableCell>
 <p class="font-mono text-xs">{{ project.current_image_tag || '-'}}</p>
              <p class="mt-1 max-w-36 truncate font-mono text-[10px] text-muted-foreground" :title="project.current_commit_sha">
                {{ project.current_commit_sha || '未登记 Commit SHA' }}
              </p>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="healthTone(project.health_status)">
                {{ healthLabel(project.health_status) }}
              </AdminStatusBadge>
              <p class="mt-1 text-[10px] text-muted-foreground">
                期望：{{ statusLabel(project.status) }}
              </p>
            </TableCell>
            <TableCell>
              <p class="text-[10px] text-muted-foreground">
                发布：{{ project.last_deployment_at ? formatDate(project.last_deployment_at) : '未登记' }}
              </p>
              <p class="mt-1 text-[10px] text-muted-foreground">
                检查：{{ project.last_checked_at ? formatDate(project.last_checked_at) : '未同步' }}
              </p>
              <p v-if="project.last_error" class="mt-1 max-w-36 truncate text-[10px] text-rose-600" :title="project.last_error">
                {{ project.last_error }}
              </p>
            </TableCell>
            <TableCell class="text-right">
              <div class="flex justify-end gap-1">
                <Button
                  v-if="canSync && project.vps_provider === 'hostinger'"
                  size="icon"
                  variant="ghost"
                  :title="`同步 ${project.name}`"
                  :disabled="syncingId === project.id"
                  @click="syncProject(project)"
                >
                  <LoaderCircle v-if="syncingId === project.id" class="size-4 animate-spin" />
                  <RefreshCw v-else class="size-4" />
                </Button>
                <Button
                  v-if="canEdit"
                  size="icon"
                  variant="ghost"
                  :title="project.enabled ? '停用项目记录' : '启用项目记录'"
                  @click="toggleProject(project)"
                >
                  <Power class="size-4" />
                </Button>
                <Button v-if="canEdit" size="icon" variant="ghost" title="编辑项目" @click="openEdit(project)">
                  <Pencil class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <OpsProjectBindingFormDialog
      :open="dialogOpen"
      :form="form"
      :vps="vpsList"
      :connectors="connectors"
      :saving="saving"
      :can-edit="canEdit"
      @update:open="dialogOpen = $event"
      @save="saveProject"
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
import OpsProjectBindingFormDialog from '@/components/admin/ops/OpsProjectBindingFormDialog.vue'
import {
  assignOpsProjectForm,
  emptyOpsProjectForm,
  opsProjectEnvironmentOptions,
  opsProjectStatusOptions,
  projectQuickBuyRateLimitSummary,
  serializeQuickBuyRateLimitPolicy,
  toOptionalISOTime,
  type OpsProjectForm,
} from '@/components/admin/ops/opsProjectBindingForm'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import opsApi, {
  type OpsConnector,
  type OpsEnvironment,
  type OpsProject,
  type OpsProjectPayload,
  type OpsVPS,
} from '@/api/ops'
import { readOpsEnvironmentQuery, withOpsEnvironmentQuery } from '@/lib/opsEnvironment'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('ops:project:edit'))
const canSync = computed(() => authStore.hasPermission('ops:project:sync'))
const projects = ref<OpsProject[]>([])
const vpsList = ref<OpsVPS[]>([])
const connectors = ref<OpsConnector[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const syncingId = ref(0)
const environmentFilter = ref<OpsEnvironment | ''>(readOpsEnvironmentQuery(route.query.environment))

const form = reactive<OpsProjectForm>(emptyOpsProjectForm())

const enabledCount = computed(() => projects.value.filter((project) => project.enabled).length)
const healthyCount = computed(() => projects.value.filter((project) => project.health_status === 'healthy').length)
const attentionCount = computed(() => projects.value.filter((project) => (
  ['pending', 'drifted', 'error'].includes(project.status) ||
  ['degraded', 'offline'].includes(project.health_status)
)).length)

const assignForm = (project?: Partial<OpsProject>): void => {
  assignOpsProjectForm(form, project)
  if (!project && environmentFilter.value) {
    form.environment = environmentFilter.value
  }
}

const loadProjectList = async (): Promise<void> => {
  loading.value = true
  try {
    const projectData = await opsApi.listProjects(environmentFilter.value || undefined)
    projects.value = Array.isArray(projectData?.projects) ? projectData.projects : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '项目列表加载失败')
  } finally {
    loading.value = false
  }
}

const loadVPSOptions = async (): Promise<void> => {
  if (!canEdit.value) {
    vpsList.value = []
    return
  }
  try {
    const vpsData = await opsApi.listVPS()
    vpsList.value = Array.isArray(vpsData?.vps) ? vpsData.vps : []
  } catch (error: any) {
    vpsList.value = []
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'VPS 选项加载失败，项目列表仍可使用')
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
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接器选项加载失败，项目列表仍可使用')
  }
}

const changeEnvironment = (): void => {
  void router.replace({ query: withOpsEnvironmentQuery(route.query, environmentFilter.value) })
  void loadProjectList()
}

const refreshProjectsPage = async (): Promise<void> => {
  await Promise.all([
    loadProjectList(),
    loadVPSOptions(),
    loadConnectorOptions(),
  ])
}

const openCreate = (): void => {
  assignForm()
  dialogOpen.value = true
  if (!vpsList.value.length) void loadVPSOptions()
  if (!connectors.value.length) void loadConnectorOptions()
}

const openEdit = (project: OpsProject): void => {
  assignForm(project)
  dialogOpen.value = true
  if (!vpsList.value.length) void loadVPSOptions()
  if (!connectors.value.length) void loadConnectorOptions()
}

const saveProject = async (): Promise<void> => {
  if (!form.name.trim()) {
    toast.error('请输入项目名称')
    return
  }
  if (!form.vps_binding_id) {
    toast.error('请选择绑定 VPS')
    return
  }
  saving.value = true
  try {
    const payload: OpsProjectPayload = {
      name: form.name.trim(),
      vps_binding_id: form.vps_binding_id,
      connector_id: form.connector_id || null,
      provider_resource_id: form.provider_resource_id.trim(),
      environment: form.environment,
      compose_source: form.compose_source.trim(),
      compose_project_name: form.compose_project_name.trim(),
      gateway_network: form.gateway_network.trim(),
      gateway_alias: form.gateway_alias.trim(),
      services: form.services.trim(),
      networks: form.networks.trim(),
      volumes: form.volumes.trim(),
      current_image_tag: form.current_image_tag.trim(),
      current_commit_sha: form.current_commit_sha.trim(),
      status: form.status,
      enabled: form.enabled,
      last_deployment_at: toOptionalISOTime(form.last_deployment_at),
      backup_policy: form.backup_policy.trim(),
      restore_notes: form.restore_notes.trim(),
      quick_buy_rate_limit_policy: serializeQuickBuyRateLimitPolicy(form.quickBuyRateLimit),
      notes: form.notes.trim(),
    }
    if (form.id) {
      await opsApi.updateProject(form.id, payload)
      toast.success('项目绑定已保存')
    } else {
      await opsApi.createProject(payload)
      toast.success('项目绑定已创建')
    }
    dialogOpen.value = false
    await loadProjectList()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '项目保存失败')
  } finally {
    saving.value = false
  }
}

const toggleProject = async (project: OpsProject): Promise<void> => {
  try {
    const updated = await opsApi.setProjectEnabled(project.id, !project.enabled)
    const index = projects.value.findIndex((item) => item.id === project.id)
    if (index >= 0) projects.value[index] = updated
    toast.success(updated.enabled ? '项目记录已启用' : '项目记录已停用')
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '项目状态更新失败')
  }
}

const syncProject = async (project: OpsProject): Promise<void> => {
  syncingId.value = project.id
  try {
    const result = await opsApi.syncProject(project.id)
    const tone = result.health_status === 'healthy'
      ? 'success'
      : result.health_status === 'degraded'
        ? 'warning'
        : result.health_status === 'offline'
          ? 'error'
          : 'warning'
    toast[tone](result.message || `${project.name} 同步完成`)
    const index = projects.value.findIndex((item) => item.id === project.id)
    if (index >= 0) {
      projects.value[index] = {
        ...project,
        health_status: result.health_status,
        observed_state: result.remote_state || '',
        observed_source: result.observed_source,
        observed_container_count: result.container_count,
        observed_running_container_count: result.running_container_count,
        observed_healthy_container_count: result.healthy_container_count,
        last_checked_at: result.last_checked_at,
        last_error: result.observed_error || '',
      }
    }
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || `${project.name} 同步失败`)
  } finally {
    syncingId.value = 0
  }
}

const healthOptions = [
  { value: 'healthy', label: '健康' },
  { value: 'degraded', label: '降级' },
  { value: 'unknown', label: '未同步' },
  { value: 'offline', label: '离线' },
]
const optionLabel = (options: Array<{ value: string; label: string }>, value: string): string => (
  options.find((option) => option.value === value)?.label || value || '-'
)
const environmentLabel = (value: string): string => optionLabel(opsProjectEnvironmentOptions, value)
const statusLabel = (value: string): string => optionLabel(opsProjectStatusOptions, value)
const healthLabel = (value: string): string => optionLabel(healthOptions, value)
const formatDate = (value: string): string => new Date(value).toLocaleString()
const containerSummary = (project: OpsProject): string => {
  if (!project.last_checked_at) return '容器：未同步'
  return `容器：${project.observed_container_count || 0} / 运行 ${project.observed_running_container_count || 0} / 健康 ${project.observed_healthy_container_count || 0}`
}
const projectConnectorLabel = (project: OpsProject): string => {
  const connectorID = project.connector_id || project.vps_connector_id
  const connector = connectorID
    ? connectors.value.find((item) => item.id === connectorID)
    : undefined
  if (project.connector_id) {
    return connector ? `${connector.name} · 项目指定` : '项目指定连接器'
  }
  if (project.vps_connector_id) {
    return connector ? `${connector.name} · 沿用 VPS` : '沿用 VPS 连接器'
  }
  return 'VPS 未绑定连接器'
}
const healthTone = (value: string): AdminStatusTone => {
  if (value === 'healthy') return 'green'
  if (value === 'degraded') return 'amber'
  if (value === 'offline') return 'coral'
  return 'gray'
}

watch(() => route.query.environment, (value) => {
  const nextEnvironment = readOpsEnvironmentQuery(value)
  if (nextEnvironment === environmentFilter.value) return
  environmentFilter.value = nextEnvironment
  void loadProjectList()
})

onMounted(refreshProjectsPage)
</script>
