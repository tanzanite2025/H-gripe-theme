<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 运维总览"
      description="集中查看目标环境的 VPS、项目、域名和外部连接的声明式状态。"
    >
      <template #actions>
        <select
          v-model="environmentFilter"
          class="h-9 rounded-md border bg-background px-3 text-sm"
          aria-label="筛选运维总览环境"
          :disabled="loading"
          @change="changeEnvironment"
        >
          <option v-for="option in opsEnvironmentOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
        <Button variant="outline" :disabled="loading" @click="loadOverview">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <section class="flex flex-col gap-3 rounded-2xl border border-dashed border-amber-500/30 bg-amber-500/5 p-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-start gap-3">
        <ShieldAlert class="mt-0.5 size-5 shrink-0 text-amber-600" />
        <div>
          <p class="text-sm font-black">当前为声明式运维台账</p>
          <p class="mt-1 text-xs text-muted-foreground">
            当前页面聚合声明式台账与已落库的观察状态；“未同步”不会被当作“健康”。
          </p>
        </div>
      </div>
      <AdminStatusBadge tone="amber">{{ environmentLabel(overview?.environment) }} / {{ generatedLabel }}</AdminStatusBadge>
    </section>

    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <button
        v-for="item in summaryCards"
        :key="item.key"
        type="button"
        class="group relative flex min-h-28 items-start justify-between gap-3 overflow-hidden rounded-2xl border border-dashed border-border/80 bg-card p-4 text-left transition-all hover:border-primary/40 hover:bg-muted/30"
        @click="navigateTo(item.path)"
      >
        <div class="uds-glow-bg opacity-30 transition-opacity group-hover:opacity-100" />
        <div class="relative z-10 min-w-0">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ item.label }}</span>
          <strong class="mt-2 block text-3xl font-black tabular-nums">{{ item.total }}</strong>
          <p class="mt-1 text-[10px] text-muted-foreground">
            启用 {{ item.enabled }} · {{ item.secondaryLabel }} {{ item.secondary }}
          </p>
        </div>
        <span class="relative z-10 flex size-9 shrink-0 items-center justify-center rounded-full" :class="item.toneClass">
          <component :is="item.icon" class="size-4" />
        </span>
      </button>
    </section>

    <section class="grid gap-4 xl:grid-cols-[minmax(0,1.25fr)_minmax(0,1fr)]">
      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <CardTitle>{{ environmentLabel(overview?.environment) }}拓扑</CardTitle>
          <CardDescription>当前登记的 VPS、Compose 项目和公网域名关系</CardDescription>
        </CardHeader>
        <CardContent class="space-y-3 pt-1">
          <div v-if="vpsList.length === 0" class="py-6 text-center text-xs text-muted-foreground">尚未登记 VPS</div>
          <div v-for="vps in vpsList" :key="`vps-${vps.id}`" class="rounded-xl border border-dashed border-border/70 p-3">
            <button type="button" class="flex w-full items-start justify-between gap-3 text-left" @click="navigateTo('/ops/vps')">
              <div class="min-w-0">
                <p class="truncate text-sm font-black">{{ vps.name }}</p>
                <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">
                  {{ vps.hostname || vps.ipv4 || '未登记地址' }}
                </p>
              </div>
              <AdminStatusBadge :tone="observedTone(vps.observed_status)">
                {{ observedLabel(vps.observed_status) }}
              </AdminStatusBadge>
            </button>

            <div class="my-2 ml-4 border-l border-dashed border-border/80 pl-4">
              <div v-if="projectsForVPS(vps.id).length === 0" class="text-[10px] text-muted-foreground">未绑定项目</div>
              <button
                v-for="project in projectsForVPS(vps.id)"
                :key="`project-${project.id}`"
                type="button"
                class="mb-2 flex w-full items-center justify-between gap-3 rounded-lg bg-muted/40 px-3 py-2 text-left last:mb-0 hover:bg-muted"
                @click="navigateTo('/ops/projects')"
              >
                <span class="min-w-0">
                  <span class="block truncate text-xs font-bold">{{ project.name }}</span>
                  <span class="mt-0.5 block truncate font-mono text-[10px] text-muted-foreground">
                    {{ project.compose_project_name || '未登记 Compose 项目' }}
                  </span>
                </span>
                <AdminStatusBadge :tone="healthTone(project.health_status)">
                  {{ healthLabel(project.health_status) }}
                </AdminStatusBadge>
              </button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <CardTitle>公网域名</CardTitle>
          <CardDescription>当前域名绑定和网关目标</CardDescription>
        </CardHeader>
        <CardContent class="pt-1">
          <div v-if="domainList.length === 0" class="py-6 text-center text-xs text-muted-foreground">尚未登记域名</div>
          <button
            v-for="domain in domainList"
            :key="domain.id"
            type="button"
            class="flex w-full items-center justify-between gap-3 border-b border-dashed border-border/60 py-3 text-left last:border-b-0"
            @click="navigateTo('/ops/domains')"
          >
            <span class="min-w-0">
              <span class="block truncate text-xs font-black">{{ domain.domain }}</span>
              <span class="mt-1 block truncate font-mono text-[10px] text-muted-foreground">
                {{ domain.target || '未登记目标' }}
              </span>
            </span>
            <span class="flex shrink-0 flex-col items-end gap-1">
              <AdminStatusBadge :tone="statusTone(domain.status)">
                期望：{{ statusLabel(domain.status) }}
              </AdminStatusBadge>
              <AdminStatusBadge :tone="observedStatusTone(domain.observed_status)">
                实际：{{ observedStatusLabel(domain.observed_status) }}
              </AdminStatusBadge>
            </span>
          </button>
        </CardContent>
      </Card>
    </section>

    <section class="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]">
      <Card size="sm">
        <CardHeader class="flex flex-row items-center justify-between border-b border-dashed border-border/70">
          <div>
            <CardTitle>待处理项</CardTitle>
            <CardDescription>需要补录、测试、同步或人工确认的资源</CardDescription>
          </div>
          <AdminStatusBadge :tone="attention.length ? 'amber' : 'green'">
            {{ attention.length }} 项
          </AdminStatusBadge>
        </CardHeader>
        <CardContent class="pt-1">
          <div v-if="attention.length === 0" class="flex items-center gap-2 py-6 text-xs text-emerald-700">
            <CircleCheck class="size-4" />
            当前没有待处理项
          </div>
          <div v-for="item in attention" :key="`${item.kind}-${item.id}`" class="flex items-start gap-3 border-b border-dashed border-border/60 py-3 last:border-b-0">
            <span class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full bg-amber-500/10 text-amber-700">
              <component :is="kindIcon(item.kind)" class="size-3.5" />
            </span>
            <button type="button" class="min-w-0 flex-1 text-left" @click="navigateTo(kindPath(item.kind))">
              <span class="flex flex-wrap items-center gap-2">
                <span class="truncate text-xs font-black">{{ item.name }}</span>
                <AdminStatusBadge :tone="resourceTone(item)">{{ kindLabel(item.kind) }}</AdminStatusBadge>
              </span>
              <span class="mt-1 block text-[10px] text-muted-foreground">{{ item.message }}</span>
              <span v-if="item.target" class="mt-1 block truncate font-mono text-[10px] text-muted-foreground/70">{{ item.target }}</span>
            </button>
          </div>
        </CardContent>
      </Card>

      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <CardTitle>最近运维操作</CardTitle>
          <CardDescription>仅显示运维中心相关审计记录</CardDescription>
        </CardHeader>
        <CardContent class="pt-1">
          <div v-if="recentAudit.length === 0" class="py-6 text-center text-xs text-muted-foreground">暂无运维审计记录</div>
          <div v-for="log in recentAudit" :key="log.id" class="border-b border-dashed border-border/60 py-3 last:border-b-0">
            <div class="flex items-center justify-between gap-3">
              <span class="truncate text-xs font-bold">{{ resourceLabel(log.resource) }}</span>
              <AdminStatusBadge :tone="log.status === 'success' ? 'green' : 'coral'">
                {{ log.status === 'success' ? '成功' : '失败' }}
              </AdminStatusBadge>
            </div>
            <p class="mt-1 text-[10px] text-muted-foreground">
              {{ actionLabel(log.action) }} · {{ log.username || '系统' }} · {{ formatDate(log.created_at) }}
            </p>
          </div>
        </CardContent>
      </Card>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  CircleCheck,
  Cloud,
  Globe2,
  HardDrive,
  Link2,
  PlugZap,
  RefreshCw,
  Server,
  ShieldAlert,
  Workflow,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import opsApi, {
  type OpsDomain,
  type OpsEnvironment,
  type OpsOverview,
  type OpsOverviewAuditLog,
  type OpsProject,
  type OpsVPS,
} from '@/api/ops'
import {
  opsEnvironmentOptions,
  readOpsEnvironmentQuery,
  withOpsEnvironmentQuery,
} from '@/lib/opsEnvironment'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const overview = ref<OpsOverview | null>(null)
const environmentFilter = ref<OpsEnvironment>(
  readOpsEnvironmentQuery(route.query.environment, 'production') || 'production',
)

const summary = computed(() => overview.value?.summary || {})
const vpsList = computed<OpsVPS[]>(() => overview.value?.topology?.vps || [])
const projects = computed<OpsProject[]>(() => overview.value?.topology?.projects || [])
const domainList = computed<OpsDomain[]>(() => overview.value?.topology?.domains || [])
const attention = computed(() => overview.value?.attention || [])
const recentAudit = computed<OpsOverviewAuditLog[]>(() => Array.isArray(overview.value?.recent_audit) ? overview.value.recent_audit : [])
const generatedLabel = computed(() => overview.value?.generated_at ? formatDate(overview.value.generated_at) : '尚未生成')
const environmentLabel = (value?: string): string => ({
  production: '生产',
  staging: '预发布',
  test: '测试',
  local: '本地',
}[value || ''] || value || '目标环境')

const summaryCards = computed(() => [
  {
    key: 'vps',
    label: 'VPS 资源',
    total: summary.value.vps?.total || 0,
    enabled: summary.value.vps?.enabled || 0,
    secondaryLabel: '未同步',
    secondary: summary.value.vps?.unknown || 0,
    path: '/ops/vps',
    icon: Server,
    toneClass: 'bg-blue-500/10 text-blue-700',
  },
  {
    key: 'projects',
    label: '项目绑定',
    total: summary.value.projects?.total || 0,
    enabled: summary.value.projects?.enabled || 0,
    secondaryLabel: '健康',
    secondary: summary.value.projects?.healthy || 0,
    path: '/ops/projects',
    icon: Workflow,
    toneClass: 'bg-emerald-500/10 text-emerald-700',
  },
  {
    key: 'domains',
    label: '域名绑定',
    total: summary.value.domains?.total || 0,
    enabled: summary.value.domains?.enabled || 0,
    secondaryLabel: '待处理',
    secondary: summary.value.domains?.attention || 0,
    path: '/ops/domains',
    icon: Globe2,
    toneClass: 'bg-amber-500/10 text-amber-700',
  },
  {
    key: 'connectors',
    label: '连接器',
    total: summary.value.connectors?.total || 0,
    enabled: summary.value.connectors?.enabled || 0,
    secondaryLabel: '待测试',
    secondary: summary.value.connectors?.unknown || 0,
    path: '/ops/connectors',
    icon: PlugZap,
    toneClass: 'bg-rose-500/10 text-rose-700',
  },
])

const loadOverview = async (): Promise<void> => {
  loading.value = true
  try {
    overview.value = await opsApi.getOverview(environmentFilter.value)
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '运维总览加载失败')
  } finally {
    loading.value = false
  }
}

const changeEnvironment = (): void => {
  void router.replace({ query: withOpsEnvironmentQuery(route.query, environmentFilter.value) })
  void loadOverview()
}

const navigateTo = (path: string): void => {
  void router.push({
    path,
    query: { environment: environmentFilter.value },
  })
}

const projectsForVPS = (vpsID: number): OpsProject[] => projects.value.filter((project) => project.vps_binding_id === vpsID)
const formatDate = (value?: string): string => value ? new Date(value).toLocaleString('zh-CN') : '-'
const optionLabel = (options: Array<{ value: string; label: string }>, value: string): string => (
  options.find((option) => option.value === value)?.label || value || '-'
)
const statusLabel = (value: string): string => optionLabel([
  { value: 'active', label: '正常' },
  { value: 'pending', label: '待确认' },
  { value: 'disabled', label: '已停用' },
  { value: 'drifted', label: '配置漂移' },
  { value: 'error', label: '错误' },
], value)
const observedStatusLabel = (value: string): string => optionLabel([
  { value: 'unknown', label: '未同步' },
  { value: 'matched', label: '已匹配' },
  { value: 'drifted', label: '漂移' },
  { value: 'error', label: '检查错误' },
], value)
const observedLabel = (value: string): string => optionLabel([
  { value: 'healthy', label: '健康' },
  { value: 'degraded', label: '降级' },
  { value: 'unknown', label: '未同步' },
  { value: 'offline', label: '离线' },
], value)
const healthLabel = (value: string): string => optionLabel([
  { value: 'healthy', label: '健康' },
  { value: 'degraded', label: '降级' },
  { value: 'unknown', label: '未同步' },
  { value: 'offline', label: '离线' },
], value)
const statusTone = (value: string): AdminStatusTone => {
  if (value === 'active' || value === 'healthy') return 'green'
  if (value === 'pending' || value === 'drifted' || value === 'degraded' || value === 'unknown') return 'amber'
  if (value === 'error' || value === 'offline') return 'coral'
  return 'gray'
}
const observedStatusTone = (value: string): AdminStatusTone => {
  if (value === 'matched' || value === 'healthy') return 'green'
  if (value === 'drifted' || value === 'unknown' || value === 'degraded') return 'amber'
  if (value === 'error' || value === 'offline') return 'coral'
  return 'gray'
}
const resourceTone = (item: { status: string; observed_status?: string; health_status?: string }): AdminStatusTone => (
  item.observed_status
    ? observedStatusTone(item.observed_status)
    : item.health_status
      ? statusTone(item.health_status)
      : statusTone(item.status)
)
const observedTone = (value: string): AdminStatusTone => statusTone(value)
const healthTone = (value: string): AdminStatusTone => statusTone(value)
const kindLabel = (value: string): string => ({ domain: '域名', connector: '连接器', vps: 'VPS', project: '项目' }[value] || value)
const kindPath = (value: string): string => ({ domain: '/ops/domains', connector: '/ops/connectors', vps: '/ops/vps', project: '/ops/projects' }[value] || '/ops/overview')
const kindIcon = (value: string) => ({ domain: Globe2, connector: Link2, vps: HardDrive, project: Workflow }[value] || Cloud)
const resourceLabel = (value?: string): string => ({
  ops_domain_binding: '域名绑定',
  ops_connector: '连接器',
  ops_vps_binding: 'VPS 绑定',
  ops_project_binding: '项目绑定',
}[value || ''] || value || '运维资源')
const actionLabel = (value?: string): string => ({ create: '创建', update: '更新', probe: '测试' }[value || ''] || value || '操作')

watch(() => route.query.environment, (value) => {
  const nextEnvironment = readOpsEnvironmentQuery(value, 'production') || 'production'
  if (nextEnvironment === environmentFilter.value) return
  environmentFilter.value = nextEnvironment
  void loadOverview()
})

onMounted(loadOverview)
</script>
