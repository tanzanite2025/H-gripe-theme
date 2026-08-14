<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 项目中心"
      description="维护 Hostinger Compose 项目、服务边界、当前版本和备份恢复记录；这里只做台账维护。"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="loadProjects">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
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
            <TableHead>Hostinger 实际</TableHead>
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
              <p class="text-xs font-bold">{{ project.vps_name || '未绑定 VPS' }}</p>
              <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ project.vps_hostname || project.vps_ipv4 || '-' }}</p>
            </TableCell>
            <TableCell>
              <p class="font-mono text-xs">{{ project.compose_project_name || '-' }}</p>
              <p class="mt-1 truncate text-[10px] text-muted-foreground">{{ project.compose_source || '未登记 Compose 来源' }}</p>
              <p v-if="project.gateway_network || project.gateway_alias" class="mt-1 text-[10px] text-muted-foreground">
                {{ project.gateway_network || '-' }} / {{ project.gateway_alias || '-' }}
              </p>
            </TableCell>
            <TableCell>
              <p class="font-mono text-xs">{{ project.observed_state || '未同步' }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ project.observed_source || '无同步来源' }}</p>
              <p class="mt-1 font-mono text-[10px] text-muted-foreground">
                {{ containerSummary(project) }}
              </p>
            </TableCell>
            <TableCell>
              <p class="max-w-56 truncate text-xs" :title="project.services">{{ project.services || '-' }}</p>
              <p class="mt-1 max-w-56 truncate text-[10px] text-muted-foreground" :title="project.networks">
                网络：{{ project.networks || '-' }}
              </p>
              <p class="mt-1 max-w-56 truncate text-[10px] text-muted-foreground" :title="project.volumes">
                卷：{{ project.volumes || '-' }}
              </p>
              <p class="mt-1 max-w-56 truncate text-[10px] text-muted-foreground">
                Quick Buy：{{ quickBuyRateLimitSummary(project) }}
              </p>
            </TableCell>
            <TableCell>
              <p class="font-mono text-xs">{{ project.current_image_tag || '-' }}</p>
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

    <Dialog v-model:open="dialogOpen">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑项目绑定' : '新增项目绑定' }}</DialogTitle>
          <DialogDescription>
            项目状态来自后台维护的声明式记录。填写实际检查结果前，请保留“未同步”，不要把文档基线当成实时健康状态。
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-4" @submit.prevent="saveProject">
          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField label="项目名称" required>
              <Input v-model="form.name" placeholder="commerce-platform" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="绑定 VPS" required>
              <select v-model.number="form.vps_binding_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option :value="0">请选择 VPS</option>
                <option v-for="vps in vpsList" :key="vps.id" :value="vps.id">{{ vps.name }} · {{ vps.hostname || vps.ipv4 || '未登记地址' }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="环境" required>
              <select v-model="form.environment" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in environmentOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="Hostinger 项目 ID">
              <Input v-model="form.provider_resource_id" placeholder="可在 Hostinger 确认后补录" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="Compose 来源">
              <Input v-model="form.compose_source" placeholder="compose.prod.yml" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="Compose 项目名">
              <Input v-model="form.compose_project_name" placeholder="commerce-platform" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="共享网关网络">
              <Input v-model="form.gateway_network" placeholder="shared-edge" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="网关别名">
              <Input v-model="form.gateway_alias" placeholder="theme-web" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="服务列表" description="用逗号分隔，例如 db, redis, api。">
              <Input v-model="form.services" placeholder="db, redis, api, storefront, admin, web" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="网络边界" description="用逗号分隔。">
              <Input v-model="form.networks" placeholder="db, cache, app, shared-edge" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="卷边界" description="用逗号分隔。">
              <Input v-model="form.volumes" placeholder="postgres, redis, uploads" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="当前镜像标签">
              <Input v-model="form.current_image_tag" placeholder="master 或 sha-..." :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="当前 Commit SHA">
              <Input v-model="form.current_commit_sha" placeholder="可选：完整 SHA" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="期望状态">
              <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="最后部署时间">
              <Input v-model="form.last_deployment_at" type="datetime-local" :disabled="saving" />
            </AdminFormField>
            <div class="md:col-span-2 rounded-lg border border-border/80 p-3">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <p class="text-xs font-black text-foreground">Quick Buy 限流</p>
                <label class="inline-flex items-center gap-2 text-xs font-bold">
                  <input v-model="form.quickBuyRateLimit.enabled" type="checkbox" class="size-4" :disabled="saving" />
                  启用
                </label>
              </div>
              <div class="mt-3 grid gap-3 md:grid-cols-4">
                <AdminFormField label="IP 每分钟">
                  <Input v-model.number="form.quickBuyRateLimit.ipRequestsPerMinute" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled" />
                </AdminFormField>
                <AdminFormField label="IP Burst">
                  <Input v-model.number="form.quickBuyRateLimit.ipBurst" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled" />
                </AdminFormField>
                <AdminFormField label="Session 每分钟">
                  <Input v-model.number="form.quickBuyRateLimit.sessionRequestsPerMinute" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled" />
                </AdminFormField>
                <AdminFormField label="Session Burst">
                  <Input v-model.number="form.quickBuyRateLimit.sessionBurst" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled" />
                </AdminFormField>
                <AdminFormField label="Edge IP 每分钟">
                  <Input v-model.number="form.quickBuyRateLimit.edgeIPRequestsPerMinute" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled || !form.quickBuyRateLimit.caddyRateLimitEnabled" />
                </AdminFormField>
                <AdminFormField label="Edge IP Burst">
                  <Input v-model.number="form.quickBuyRateLimit.edgeIPBurst" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled || !form.quickBuyRateLimit.caddyRateLimitEnabled" />
                </AdminFormField>
                <div class="flex items-center md:col-span-2">
                  <label class="inline-flex items-center gap-2 text-xs font-bold">
                    <input v-model="form.quickBuyRateLimit.caddyRateLimitEnabled" type="checkbox" class="size-4" :disabled="saving || !form.quickBuyRateLimit.enabled" />
                    Caddy 边缘桶
                  </label>
                </div>
              </div>
            </div>
            <AdminFormField label="备份策略" class="md:col-span-2">
              <Textarea v-model="form.backup_policy" class="min-h-20" placeholder="每日 PostgreSQL、uploads 备份，异地保存和恢复演练。" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="恢复说明" class="md:col-span-2">
              <Textarea v-model="form.restore_notes" class="min-h-20" placeholder="记录最近一次恢复演练、备份位置和待处理事项。" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="备注" class="md:col-span-2">
              <Textarea v-model="form.notes" class="min-h-24" placeholder="记录项目边界、发布来源和维护说明。" :disabled="saving" />
            </AdminFormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving || !canEdit">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              <Save v-else class="size-4" />
              {{ saving ? '保存中' : '保存项目' }}
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
import opsApi, { type OpsProjectPayload } from '@/api/ops'
import { useAuthStore } from '@/stores/auth'

interface OpsVPS {
  id: number
  name: string
  provider: string
  hostname: string
  ipv4: string
}

interface OpsProject {
  id: number
  name: string
  vps_binding_id: number
  provider_resource_id: string
  environment: string
  compose_source: string
  compose_project_name: string
  gateway_network: string
  gateway_alias: string
  services: string
  networks: string
  volumes: string
  current_image_tag: string
  current_commit_sha: string
  status: string
  health_status: string
  observed_state: string
  observed_source: string
  observed_container_count: number
  observed_running_container_count: number
  observed_healthy_container_count: number
  enabled: boolean
  last_deployment_at?: string
  last_checked_at?: string
  last_error: string
  backup_policy: string
  restore_notes: string
  quick_buy_rate_limit_policy: string
  notes: string
  vps_name: string
  vps_provider: string
  vps_hostname: string
  vps_ipv4: string
}

interface QuickBuyRateLimitForm {
  enabled: boolean
  ipRequestsPerMinute: number
  ipBurst: number
  sessionRequestsPerMinute: number
  sessionBurst: number
  edgeIPRequestsPerMinute: number
  edgeIPBurst: number
  caddyRateLimitEnabled: boolean
}

interface OpsProjectForm {
  id: number
  name: string
  vps_binding_id: number
  provider_resource_id: string
  environment: string
  compose_source: string
  compose_project_name: string
  gateway_network: string
  gateway_alias: string
  services: string
  networks: string
  volumes: string
  current_image_tag: string
  current_commit_sha: string
  status: string
  enabled: boolean
  last_deployment_at: string
  backup_policy: string
  restore_notes: string
  quickBuyRateLimit: QuickBuyRateLimitForm
  notes: string
}

const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('ops:project:edit'))
const canSync = computed(() => authStore.hasPermission('ops:project:sync'))
const projects = ref<OpsProject[]>([])
const vpsList = ref<OpsVPS[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const syncingId = ref(0)

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

const defaultQuickBuyRateLimit = (): QuickBuyRateLimitForm => ({
  enabled: true,
  ipRequestsPerMinute: 120,
  ipBurst: 40,
  sessionRequestsPerMinute: 60,
  sessionBurst: 20,
  edgeIPRequestsPerMinute: 240,
  edgeIPBurst: 80,
  caddyRateLimitEnabled: false,
})

const positiveInteger = (value: unknown, fallback: number): number => {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) return fallback
  return Math.round(parsed)
}

const parseQuickBuyRateLimitPolicy = (raw?: string): QuickBuyRateLimitForm => {
  const defaults = defaultQuickBuyRateLimit()
  if (!raw?.trim()) return defaults
  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>
    if (!parsed || typeof parsed !== 'object') return defaults
    return {
      enabled: Boolean(parsed.enabled ?? defaults.enabled),
      ipRequestsPerMinute: positiveInteger(parsed.ip_requests_per_minute, defaults.ipRequestsPerMinute),
      ipBurst: positiveInteger(parsed.ip_burst, defaults.ipBurst),
      sessionRequestsPerMinute: positiveInteger(parsed.session_requests_per_minute, defaults.sessionRequestsPerMinute),
      sessionBurst: positiveInteger(parsed.session_burst, defaults.sessionBurst),
      edgeIPRequestsPerMinute: positiveInteger(parsed.edge_ip_requests_per_minute, defaults.edgeIPRequestsPerMinute),
      edgeIPBurst: positiveInteger(parsed.edge_ip_burst, defaults.edgeIPBurst),
      caddyRateLimitEnabled: Boolean(parsed.caddy_rate_limit_enabled ?? defaults.caddyRateLimitEnabled),
    }
  } catch {
    return defaults
  }
}

const serializeQuickBuyRateLimitPolicy = (policy: QuickBuyRateLimitForm): string => {
  const defaults = defaultQuickBuyRateLimit()
  return JSON.stringify({
    enabled: Boolean(policy.enabled),
    ip_requests_per_minute: positiveInteger(policy.ipRequestsPerMinute, defaults.ipRequestsPerMinute),
    ip_burst: positiveInteger(policy.ipBurst, defaults.ipBurst),
    session_requests_per_minute: positiveInteger(policy.sessionRequestsPerMinute, defaults.sessionRequestsPerMinute),
    session_burst: positiveInteger(policy.sessionBurst, defaults.sessionBurst),
    edge_ip_requests_per_minute: positiveInteger(policy.edgeIPRequestsPerMinute, defaults.edgeIPRequestsPerMinute),
    edge_ip_burst: positiveInteger(policy.edgeIPBurst, defaults.edgeIPBurst),
    caddy_rate_limit_enabled: Boolean(policy.caddyRateLimitEnabled),
  })
}
const healthOptions = [
  { value: 'healthy', label: '健康' },
  { value: 'degraded', label: '降级' },
  { value: 'unknown', label: '未同步' },
  { value: 'offline', label: '离线' },
]

const emptyForm = (): OpsProjectForm => ({
  id: 0,
  name: '',
  vps_binding_id: 0,
  provider_resource_id: '',
  environment: 'production',
  compose_source: '',
  compose_project_name: '',
  gateway_network: '',
  gateway_alias: '',
  services: '',
  networks: '',
  volumes: '',
  current_image_tag: '',
  current_commit_sha: '',
  status: 'pending',
  enabled: true,
  last_deployment_at: '',
  backup_policy: '',
  restore_notes: '',
  quickBuyRateLimit: defaultQuickBuyRateLimit(),
  notes: '',
})
const form = reactive<OpsProjectForm>(emptyForm())

const enabledCount = computed(() => projects.value.filter((project) => project.enabled).length)
const healthyCount = computed(() => projects.value.filter((project) => project.health_status === 'healthy').length)
const attentionCount = computed(() => projects.value.filter((project) => ['pending', 'drifted', 'error'].includes(project.status) || ['degraded', 'offline'].includes(project.health_status)).length)

const assignForm = (project?: Partial<OpsProject>): void => {
  Object.assign(form, emptyForm(), project || {})
  form.vps_binding_id = project?.vps_binding_id ?? 0
  form.quickBuyRateLimit = parseQuickBuyRateLimitPolicy(project?.quick_buy_rate_limit_policy)
}

const loadProjects = async (): Promise<void> => {
  loading.value = true
  try {
    const [projectData, vpsData] = await Promise.all([opsApi.listProjects(), opsApi.listVPS()])
    projects.value = Array.isArray(projectData?.projects) ? projectData.projects : []
    vpsList.value = Array.isArray(vpsData?.vps) ? vpsData.vps : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '项目列表加载失败')
  } finally {
    loading.value = false
  }
}

const openCreate = (): void => {
  assignForm()
  dialogOpen.value = true
}

const openEdit = (project: OpsProject): void => {
  assignForm(project)
  dialogOpen.value = true
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
      last_deployment_at: form.last_deployment_at,
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
    await loadProjects()
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
    await loadProjects()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || `${project.name} 同步失败`)
  } finally {
    syncingId.value = 0
  }
}

const optionLabel = (options: Array<{ value: string; label: string }>, value: string): string => (
  options.find((option) => option.value === value)?.label || value || '-'
)
const environmentLabel = (value: string): string => optionLabel(environmentOptions, value)
const statusLabel = (value: string): string => optionLabel(statusOptions, value)
const healthLabel = (value: string): string => optionLabel(healthOptions, value)
const formatDate = (value: string): string => new Date(value).toLocaleString()
const containerSummary = (project: OpsProject): string => {
  if (!project.last_checked_at) return '容器：未同步'
  return `容器：${project.observed_container_count || 0} / 运行 ${project.observed_running_container_count || 0} / 健康 ${project.observed_healthy_container_count || 0}`
}
const quickBuyRateLimitSummary = (project: OpsProject): string => {
  const policy = parseQuickBuyRateLimitPolicy(project.quick_buy_rate_limit_policy)
  if (!policy.enabled) return '关闭'
  return `${policy.ipRequestsPerMinute}/min IP · ${policy.sessionRequestsPerMinute}/min Session`
}
const healthTone = (value: string): AdminStatusTone => {
  if (value === 'healthy') return 'green'
  if (value === 'degraded') return 'amber'
  if (value === 'offline') return 'coral'
  return 'gray'
}

onMounted(loadProjects)
</script>
