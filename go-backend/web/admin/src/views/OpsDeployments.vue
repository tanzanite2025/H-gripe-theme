<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 部署中心"
      description="生成项目级 preflight 报告；当前只读取台账和观察状态，不执行部署。"
    >
      <template #actions>
        <Button variant="outline" :disabled="loadingProjects || generating" @click="loadProjects">
          <RefreshCw :class="['size-4', loadingProjects ? 'animate-spin' : '']" />
          刷新项目
        </Button>
        <Button variant="outline" :disabled="!overview || loadingProjects || generating" @click="copyOverview">
          <Copy class="size-4" />
          复制总览
        </Button>
        <Button :disabled="!selectedProjectId || generating" @click="generateReport">
          <LoaderCircle v-if="generating" class="size-4 animate-spin" />
          <FileSearch v-else class="size-4" />
          生成报告
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <div
        v-for="item in overviewStats"
        :key="item.key"
        class="rounded-xl border border-dashed border-border/80 bg-background p-3"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ item.label }}</p>
            <p class="mt-2 text-2xl font-black" :class="item.valueClass">{{ item.value }}</p>
          </div>
          <span class="flex size-8 shrink-0 items-center justify-center rounded-full" :class="item.iconClass">
            <component :is="item.icon" class="size-4" />
          </span>
        </div>
        <p class="mt-2 truncate text-[10px] text-muted-foreground">{{ item.detail }}</p>
      </div>
    </section>

    <section v-if="overviewRiskCategories.length" class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
      <button
        v-for="group in overviewRiskCategories"
        :key="group.category"
        type="button"
        class="rounded-xl border border-dashed border-border/80 bg-background p-3 text-left transition hover:bg-muted/40"
        @click="focusCategory(group.category)"
      >
        <div class="flex items-center justify-between gap-3">
          <p class="text-xs font-black">{{ group.label }}</p>
          <AdminStatusBadge :tone="groupTone(group)">{{ groupStatus(group) }}</AdminStatusBadge>
        </div>
        <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-muted">
          <div class="h-full rounded-full" :class="groupBarClass(group)" :style="{ width: groupRiskWidth(group) }" />
        </div>
        <p class="mt-2 text-[10px] text-muted-foreground">
          阻断 {{ group.blocking_count }} · 警告 {{ group.warning_count }} · 总检查 {{ group.total_count }}
        </p>
      </button>
    </section>

    <section class="grid gap-3 xl:grid-cols-[minmax(18rem,0.64fr)_minmax(0,1.36fr)]">
      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <CardTitle>项目选择</CardTitle>
          <CardDescription>{{ overview ? `总览生成于 ${formatDate(overview.generated_at)}` : '报告按项目生成，覆盖 Compose、镜像、边界、VPS、连接器、健康和备份记录' }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-3 pt-3">
          <div class="grid gap-2 sm:grid-cols-3 xl:grid-cols-1 2xl:grid-cols-3">
            <select
              v-model="overviewStatusFilter"
              class="h-9 w-full rounded-md border bg-background px-3 text-xs"
              :disabled="loadingProjects || generating"
            >
              <option value="all">全部状态</option>
              <option value="blocked">只看阻断</option>
              <option value="review">只看 REVIEW</option>
              <option value="ready">只看 READY</option>
            </select>
            <select
              v-model="overviewEnvironmentFilter"
              class="h-9 w-full rounded-md border bg-background px-3 text-xs"
              :disabled="loadingProjects || generating"
            >
              <option value="all">全部环境</option>
              <option v-for="environment in overviewEnvironments" :key="environment" :value="environment">
                {{ environmentLabel(environment) }}
              </option>
            </select>
            <select
              v-model="overviewSort"
              class="h-9 w-full rounded-md border bg-background px-3 text-xs"
              :disabled="loadingProjects || generating"
            >
              <option value="risk">风险优先</option>
              <option value="name">项目名</option>
              <option value="generated">生成时间</option>
            </select>
          </div>

          <select
            v-model.number="selectedProjectId"
            class="h-10 w-full rounded-md border bg-background px-3 text-sm"
            :disabled="loadingProjects || generating"
            @change="handleProjectChange"
          >
            <option :value="0">请选择项目</option>
            <option v-for="project in projectOptions" :key="project.id" :value="project.id">
              {{ project.name }} · {{ environmentLabel(project.environment) }}
            </option>
          </select>

          <div v-if="selectedProject" class="rounded-xl border border-dashed border-border/70 p-3">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="truncate text-sm font-black">{{ selectedProject.name }}</p>
                <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">
                  {{ selectedProject.compose_project_name || '未登记 Compose 项目名' }}
                </p>
              </div>
              <AdminStatusBadge :tone="healthTone(selectedProject.health_status)">
                {{ healthLabel(selectedProject.health_status) }}
              </AdminStatusBadge>
            </div>
            <dl class="mt-3 grid gap-2 text-[10px] text-muted-foreground">
              <div class="flex justify-between gap-3">
                <dt>Compose</dt>
                <dd class="truncate font-mono">{{ selectedProject.compose_source || '-' }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>镜像</dt>
                <dd class="truncate font-mono">{{ selectedProject.current_image_tag || '-' }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>Commit</dt>
                <dd class="truncate font-mono">{{ selectedProject.current_commit_sha || '-' }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>VPS</dt>
                <dd class="truncate">{{ selectedProject.vps_name || '-' }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>最近观察</dt>
                <dd class="truncate">{{ selectedProject.last_checked_at ? formatDate(selectedProject.last_checked_at) : '未同步' }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt>容器</dt>
                <dd class="truncate font-mono">{{ containerSummary(selectedProject) }}</dd>
              </div>
            </dl>
          </div>

          <div v-if="overviewProjects.length" class="space-y-2">
            <div class="flex items-center justify-between gap-3">
              <p class="text-xs font-black">项目 Preflight</p>
              <span class="text-[10px] text-muted-foreground">{{ filteredOverviewProjects.length }} / {{ overviewProjects.length }} 个项目</span>
            </div>
            <button
              v-for="summary in filteredOverviewProjects"
              :key="summary.project_id"
              type="button"
              class="w-full rounded-xl border border-dashed px-3 py-2 text-left transition hover:bg-muted/40 disabled:pointer-events-none disabled:opacity-60"
              :class="summary.project_id === selectedProjectId ? 'border-primary/40 bg-muted/60' : 'border-border/70'"
              :disabled="generating"
              @click="selectOverviewProject(summary.project_id)"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-xs font-black">{{ summary.project }}</p>
                  <p class="mt-0.5 truncate text-[10px] text-muted-foreground">
                    {{ environmentLabel(summary.environment) }} · 阻断 {{ summary.blocking_count }} · 警告 {{ summary.warning_count }}
                  </p>
                </div>
                <AdminStatusBadge :tone="statusLevelTone(summary.status_level)">
                  {{ statusLevelLabel(summary.status_level) }}
                </AdminStatusBadge>
              </div>
              <p class="mt-2 line-clamp-2 text-[10px] text-muted-foreground">{{ summaryReason(summary) }}</p>
            </button>
            <div v-if="!filteredOverviewProjects.length" class="rounded-xl border border-dashed border-border/70 p-3 text-center">
              <p class="text-xs font-bold">没有匹配的项目</p>
              <p class="mt-1 text-[10px] text-muted-foreground">调整状态或环境筛选后再查看。</p>
            </div>
          </div>

          <div v-if="report" class="grid gap-2">
            <button
              v-for="item in categoryCards"
              :key="item.key"
              type="button"
              class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-left transition hover:bg-muted/40"
              :class="activeCategory === item.key ? 'border-primary/40 bg-muted/60' : 'border-border/70'"
              @click="activeCategory = item.key"
            >
              <span class="min-w-0">
                <span class="block text-xs font-black">{{ item.label }}</span>
                <span class="mt-0.5 block text-[10px] text-muted-foreground">
                  {{ item.total }} 项 · 阻断 {{ item.block }} · 警告 {{ item.warning }}
                </span>
              </span>
              <AdminStatusBadge :tone="itemTone(item)">{{ itemStatus(item) }}</AdminStatusBadge>
            </button>
          </div>

          <div class="rounded-xl border border-dashed border-amber-500/30 bg-amber-500/5 p-3">
            <div class="flex items-start gap-2">
              <ShieldAlert class="mt-0.5 size-4 shrink-0 text-amber-600" />
              <p class="text-xs text-muted-foreground">
                第一版部署中心不会调用 Hostinger 更新、Cloudflare 写入、SSH、Docker restart 或缓存清理接口。
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <CardTitle>Preflight 报告</CardTitle>
              <CardDescription>{{ report ? `生成于 ${formatDate(report.generated_at)}` : '选择项目后生成只读检查报告' }}</CardDescription>
            </div>
            <div v-if="report" class="flex items-center gap-2">
              <Button variant="outline" size="sm" type="button" @click="copyReportSummary">
                <Copy class="size-4" />
                复制摘要
              </Button>
              <AdminStatusBadge :tone="statusLevelTone(report.status_level)">
                {{ statusLevelLabel(report.status_level) }}
              </AdminStatusBadge>
            </div>
          </div>
        </CardHeader>
        <CardContent class="space-y-4 pt-3">
          <div v-if="!report && !generating" class="flex min-h-56 items-center justify-center rounded-xl border border-dashed border-border/70 text-center">
            <div>
              <FileSearch class="mx-auto size-8 text-muted-foreground/50" />
              <p class="mt-3 text-sm font-bold">尚未生成报告</p>
              <p class="mt-1 text-xs text-muted-foreground">报告会显示阻断项、警告项和每个只读检查的证据摘要。</p>
            </div>
          </div>

          <div v-if="generating" class="flex min-h-56 items-center justify-center rounded-xl border border-dashed border-border/70">
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <LoaderCircle class="size-4 animate-spin" />
              正在读取项目台账并生成报告
            </div>
          </div>

          <template v-if="report && !generating">
            <section class="flex flex-wrap items-center gap-2">
              <Button
                type="button"
                size="sm"
                :variant="detailMode === 'all' ? 'default' : 'outline'"
                @click="detailMode = 'all'"
              >
                全部
              </Button>
              <Button
                type="button"
                size="sm"
                :variant="detailMode === 'needs-action' ? 'default' : 'outline'"
                @click="detailMode = 'needs-action'"
              >
                需处理项
              </Button>
              <Button
                type="button"
                size="sm"
                variant="outline"
                @click="copyFullChecklist"
              >
                <Copy class="size-4" />
                复制清单
              </Button>
            </section>

            <section class="grid gap-3 sm:grid-cols-4">
              <div class="rounded-xl border border-dashed border-border/80 p-3">
                <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">阻断项</p>
                <p class="mt-2 text-2xl font-black" :class="report.blocking_count ? 'text-rose-600' : 'text-emerald-600'">{{ report.blocking_count }}</p>
              </div>
              <div class="rounded-xl border border-dashed border-border/80 p-3">
                <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">警告项</p>
                <p class="mt-2 text-2xl font-black" :class="report.warning_count ? 'text-amber-600' : 'text-emerald-600'">{{ report.warning_count }}</p>
              </div>
              <div class="rounded-xl border border-dashed border-border/80 p-3">
                <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">通过项</p>
                <p class="mt-2 text-2xl font-black text-emerald-600">{{ report.pass_count }}</p>
              </div>
              <div class="rounded-xl border border-dashed border-border/80 p-3">
                <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">信息项</p>
                <p class="mt-2 text-2xl font-black text-blue-600">{{ report.info_count }}</p>
              </div>
            </section>

            <section class="rounded-xl border border-dashed border-border/70 bg-muted/20 p-3">
              <p class="text-sm font-bold">{{ report.summary }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">
                项目：{{ report.project }} · {{ environmentLabel(report.environment) }} · {{ statusLevelLabel(report.status_level) }}
              </p>
            </section>

            <section v-if="reportNextActions.length" class="rounded-xl border border-dashed border-border/70 bg-background p-3">
              <div class="flex items-center justify-between gap-3">
                <p class="text-xs font-black">下一步建议</p>
                <span class="text-[10px] text-muted-foreground">{{ reportNextActions.length }} 条</span>
              </div>
              <ul class="mt-2 space-y-1">
                <li v-for="action in reportNextActions" :key="action" class="text-[10px] text-muted-foreground">
                  {{ action }}
                </li>
              </ul>
            </section>

            <section v-if="reportCategories.length" class="rounded-xl border border-dashed border-border/70 bg-background p-3">
              <div class="flex items-center justify-between gap-3">
                <p class="text-xs font-black">类别分布</p>
                <span class="text-[10px] text-muted-foreground">{{ reportCategories.length }} 类</span>
              </div>
              <div class="mt-3 grid gap-2 md:grid-cols-2">
                <button
                  v-for="group in reportCategories"
                  :key="group.category"
                  type="button"
                  class="rounded-lg border border-dashed px-3 py-2 text-left transition hover:bg-muted/40"
                  :class="activeCategory === group.category ? 'border-primary/40 bg-muted/60' : 'border-border/70'"
                  @click="activeCategory = group.category"
                >
                  <div class="flex items-center justify-between gap-3">
                    <p class="truncate text-xs font-black">{{ group.label }}</p>
                    <span class="font-mono text-[10px] text-muted-foreground">{{ group.total_count }}</span>
                  </div>
                  <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-muted">
                    <div class="h-full rounded-full" :class="groupBarClass(group)" :style="{ width: groupRiskWidth(group) }" />
                  </div>
                  <p class="mt-1 text-[10px] text-muted-foreground">
                    阻断 {{ group.blocking_count }} · 警告 {{ group.warning_count }}
                  </p>
                </button>
              </div>
            </section>

            <section class="flex flex-wrap gap-2">
              <Button
                v-for="item in categoryCards"
                :key="`filter-${item.key}`"
                type="button"
                size="sm"
                :variant="activeCategory === item.key ? 'default' : 'outline'"
                @click="activeCategory = item.key"
              >
                <component :is="categoryIcon(item.key)" class="size-4" />
                {{ item.label }}
              </Button>
            </section>

            <section class="overflow-hidden rounded-xl border border-dashed border-border/70">
              <div
                v-for="check in visibleChecks"
                :key="check.key"
                class="grid gap-3 border-b border-dashed border-border/70 p-3 last:border-b-0 md:grid-cols-[10rem_minmax(0,1fr)_6rem]"
              >
                <div class="flex items-center gap-2">
                  <span class="flex size-7 shrink-0 items-center justify-center rounded-full" :class="checkIconClass(check.status)">
                    <component :is="checkIcon(check.status)" class="size-4" />
                  </span>
                  <span class="min-w-0">
                    <span class="block truncate text-sm font-black">{{ check.label }}</span>
                    <span class="mt-0.5 block text-[9px] font-bold uppercase tracking-widest text-muted-foreground/60">{{ categoryLabel(check.category) }}</span>
                  </span>
                </div>
                <div class="min-w-0">
                  <p class="text-xs font-bold">{{ check.message }}</p>
                  <p v-if="check.detail" class="mt-1 break-words font-mono text-[10px] text-muted-foreground">{{ check.detail }}</p>
                </div>
                <div class="flex items-start md:justify-end">
                  <AdminStatusBadge :tone="checkTone(check.status)">{{ checkStatusLabel(check.status) }}</AdminStatusBadge>
                </div>
              </div>
            </section>
          </template>
        </CardContent>
      </Card>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { useRoute, useRouter } from 'vue-router'
import {
  Boxes,
  CircleCheck,
  Cloud,
  Copy,
  FileSearch,
  GitCommit,
  Info,
  LoaderCircle,
  Network,
  RefreshCw,
  Server,
  ShieldAlert,
  TriangleAlert,
  XCircle,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import opsApi, {
  type OpsDeploymentPreflight,
  type OpsDeploymentPreflightGroup,
  type OpsDeploymentPreflightOverview,
} from '@/api/ops'
import {
  useDeploymentPreflightOverview,
  type DeploymentPreflightOverviewSort,
  type DeploymentPreflightOverviewStatus,
} from '@/composables/useDeploymentPreflightOverview'
import {
  buildPreflightChecklistMarkdown,
  buildPreflightOverviewMarkdown,
  buildPreflightReportMarkdown,
  categoryLabel,
  checkStatusLabel,
  environmentLabel,
  formatDeploymentPreflightDate as formatDate,
  statusLevelLabel,
  summaryReason,
} from '@/lib/deploymentPreflightPresentation'

const route = useRoute()
const router = useRouter()

const queryValue = (value: unknown): string => typeof value === 'string' ? value : ''
const queryNumber = (value: unknown): number => {
  const parsed = Number.parseInt(queryValue(value), 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0
}
const queryChoice = <T extends string>(value: unknown, allowed: readonly T[], fallback: T): T => {
  const candidate = queryValue(value) as T
  return allowed.includes(candidate) ? candidate : fallback
}

interface OpsProject {
  id: number
  name: string
  vps_binding_id: number
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
  backup_policy: string
  restore_notes: string
  vps_name: string
}

const projects = ref<OpsProject[]>([])
const overview = ref<OpsDeploymentPreflightOverview | null>(null)
const selectedProjectId = ref(queryNumber(route.query.project))
const report = ref<OpsDeploymentPreflight | null>(null)
const loadingProjects = ref(false)
const generating = ref(false)
const activeCategory = ref(queryValue(route.query.category) || 'all')
const detailMode = ref(queryChoice(route.query.mode, ['all', 'needs-action'] as const, 'all'))
const overviewStatusFilter = ref<DeploymentPreflightOverviewStatus>(queryChoice(route.query.status, ['all', 'blocked', 'review', 'ready'] as const, 'all'))
const overviewEnvironmentFilter = ref(queryValue(route.query.environment) || 'all')
const overviewSort = ref<DeploymentPreflightOverviewSort>(queryChoice(route.query.sort, ['risk', 'name', 'generated'] as const, 'risk'))

const selectedProject = computed(() => projects.value.find((project) => project.id === selectedProjectId.value) || null)
const {
  overviewProjects,
  overviewCategories,
  overviewRiskCategories,
  overviewEnvironments,
  filteredOverviewProjects,
  projectOptions,
  overviewStats: overviewStatsBase,
  mergeReportSummary,
} = useDeploymentPreflightOverview({
  projects,
  overview,
  statusFilter: overviewStatusFilter,
  environmentFilter: overviewEnvironmentFilter,
  sort: overviewSort,
})
const overviewStats = computed(() => overviewStatsBase.value.map((item) => ({
  ...item,
  icon: item.key === 'projects' ? FileSearch : item.key === 'ready' ? CircleCheck : item.key === 'review' ? TriangleAlert : XCircle,
  valueClass: item.key === 'projects'
    ? 'text-foreground'
    : item.key === 'ready'
      ? 'text-emerald-600'
      : item.key === 'review'
        ? (item.value > 0 ? 'text-amber-600' : 'text-emerald-600')
        : (item.value > 0 ? 'text-rose-600' : 'text-emerald-600'),
  iconClass: item.key === 'projects'
    ? 'bg-blue-500/10 text-blue-700'
    : item.key === 'ready'
      ? 'bg-emerald-500/10 text-emerald-700'
      : item.key === 'review'
        ? 'bg-amber-500/10 text-amber-700'
        : 'bg-rose-500/10 text-rose-700',
})))
const filteredChecks = computed(() => {
  if (!report.value) return []
  if (activeCategory.value === 'all') return report.value.checks
  return report.value.checks.filter((check) => (check.category || 'other') === activeCategory.value)
})
const reportNextActions = computed(() => report.value?.next_actions || [])
const reportCategories = computed(() => report.value?.categories || [])
const visibleChecks = computed(() => {
  const checks = filteredChecks.value
  if (detailMode.value === 'needs-action') {
    return checks.filter((check) => check.status === 'block' || check.status === 'warning')
  }
  return checks
})
const categoryCards = computed(() => {
  const checks = report.value?.checks || []
  const groups = reportCategories.value.map((group) => ({
    key: group.category,
    label: group.label,
    total: group.total_count,
    block: group.blocking_count,
    warning: group.warning_count,
    pass: group.pass_count,
  }))
  return [
    {
      key: 'all',
      label: '全部',
      total: checks.length,
      block: checks.filter((check) => check.status === 'block').length,
      warning: checks.filter((check) => check.status === 'warning').length,
      pass: checks.filter((check) => check.status === 'pass').length,
    },
    ...groups,
  ]
})

const loadProjects = async (): Promise<void> => {
  loadingProjects.value = true
  try {
    const [projectsResult, overviewResult] = await Promise.allSettled([
      opsApi.listProjects(),
      opsApi.getDeploymentPreflightOverview(),
    ])

    if (projectsResult.status === 'fulfilled') {
      projects.value = Array.isArray(projectsResult.value?.projects) ? projectsResult.value.projects : []
    } else {
      projects.value = []
    }

    if (overviewResult.status === 'fulfilled') {
      overview.value = overviewResult.value
    }

    if (projectsResult.status === 'rejected' && overviewResult.status === 'rejected') {
      throw overviewResult.reason || projectsResult.reason
    }

    if (projectsResult.status === 'rejected') {
      toast.warning('项目台账明细加载失败，已显示 preflight 总览')
    }
    if (overviewResult.status === 'rejected') {
      toast.warning('preflight 总览加载失败，仍可手动生成单项目报告')
    }

    if (selectedProjectId.value && projectOptions.value.length > 0 && !projectOptions.value.some((project) => project.id === selectedProjectId.value)) {
      selectedProjectId.value = projectOptions.value[0].id
      handleProjectChange()
    } else if (!selectedProjectId.value && projectOptions.value.length > 0) {
      selectedProjectId.value = projectOptions.value[0].id
    }
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '部署项目列表加载失败')
  } finally {
    loadingProjects.value = false
  }
}

const handleProjectChange = (): void => {
  report.value = null
  activeCategory.value = 'all'
  detailMode.value = 'all'
}

const selectOverviewProject = async (projectID: number): Promise<void> => {
  if (!projectID || projectID === selectedProjectId.value) {
    if (!report.value) {
      await generateReport()
    }
    return
  }
  selectedProjectId.value = projectID
  handleProjectChange()
  await generateReport()
}

const generateReport = async (): Promise<void> => {
  if (!selectedProjectId.value) {
    toast.error('请选择项目')
    return
  }
  generating.value = true
  try {
    report.value = await opsApi.getProjectPreflight(selectedProjectId.value)
    mergeReportSummary(report.value)
    activeCategory.value = 'all'
    detailMode.value = 'all'
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'preflight 报告生成失败')
  } finally {
    generating.value = false
  }
}

const focusCategory = (category: string): void => {
  activeCategory.value = category || 'all'
  detailMode.value = 'needs-action'
}

const copyReportSummary = async (): Promise<void> => {
  if (!report.value) return
  const markdown = buildPreflightReportMarkdown(report.value)
  try {
    await navigator.clipboard.writeText(markdown)
    toast.success('preflight 摘要已复制')
  } catch {
    toast.error('复制失败，请检查浏览器剪贴板权限')
  }
}

const copyFullChecklist = async (): Promise<void> => {
  if (!report.value) return
  const markdown = buildPreflightChecklistMarkdown(report.value)
  try {
    await navigator.clipboard.writeText(markdown)
    toast.success('完整清单已复制')
  } catch {
    toast.error('复制失败，请检查浏览器剪贴板权限')
  }
}

const copyOverview = async (): Promise<void> => {
  if (!overview.value) return
  try {
    await navigator.clipboard.writeText(buildPreflightOverviewMarkdown({
      overview: overview.value,
      projects: filteredOverviewProjects.value,
      filterLabel: overviewFilterLabel(),
    }))
    toast.success('当前总览已复制')
  } catch {
    toast.error('复制失败，请检查浏览器剪贴板权限')
  }
}

const overviewFilterLabel = (): string => [
  overviewStatusFilter.value === 'all' ? '' : `状态=${statusLevelLabel(overviewStatusFilter.value)}`,
  overviewEnvironmentFilter.value === 'all' ? '' : `环境=${environmentLabel(overviewEnvironmentFilter.value)}`,
  `排序=${overviewSort.value === 'risk' ? '风险优先' : overviewSort.value === 'name' ? '项目名' : '生成时间'}`,
].filter(Boolean).join(' · ')

const groupTone = (group: OpsDeploymentPreflightGroup): AdminStatusTone => {
  if (group.blocking_count > 0) return 'coral'
  if (group.warning_count > 0) return 'amber'
  return 'green'
}

const groupStatus = (group: OpsDeploymentPreflightGroup): string => {
  if (group.blocking_count > 0) return 'BLOCK'
  if (group.warning_count > 0) return 'WARN'
  return 'PASS'
}

const groupBarClass = (group: OpsDeploymentPreflightGroup): string => {
  if (group.blocking_count > 0) return 'bg-rose-500'
  if (group.warning_count > 0) return 'bg-amber-500'
  return 'bg-emerald-500'
}

const groupRiskWidth = (group: OpsDeploymentPreflightGroup): string => {
  const risky = group.blocking_count + group.warning_count
  const total = Math.max(group.total_count, 1)
  return `${Math.max(Math.round((risky / total) * 100), risky > 0 ? 12 : 0)}%`
}

const statusLevelTone = (value?: string): AdminStatusTone => {
  if (value === 'ready') return 'green'
  if (value === 'review') return 'amber'
  if (value === 'blocked') return 'coral'
  return 'gray'
}

const healthLabel = (value: string): string => ({
  healthy: '健康',
  degraded: '降级',
  unknown: '未同步',
  offline: '离线',
}[value] || value || '-')

const healthTone = (value: string): AdminStatusTone => {
  if (value === 'healthy') return 'green'
  if (value === 'degraded' || value === 'unknown') return 'amber'
  if (value === 'offline') return 'coral'
  return 'gray'
}

const checkTone = (status: string): AdminStatusTone => {
  if (status === 'pass') return 'green'
  if (status === 'warning') return 'amber'
  if (status === 'block') return 'coral'
  if (status === 'info') return 'blue'
  return 'gray'
}

const categoryIcon = (value: string) => ({
  all: FileSearch,
  compose: Boxes,
  version: GitCommit,
  boundary: Network,
  infra: Server,
  connector: Cloud,
  domain: Cloud,
  runtime: CircleCheck,
  evidence: Info,
}[value] || Info)

const itemTone = (item: { block: number; warning: number; total: number }): AdminStatusTone => {
  if (item.block) return 'coral'
  if (item.warning) return 'amber'
  if (item.total) return 'green'
  return 'gray'
}

const itemStatus = (item: { block: number; warning: number; total: number }): string => {
  if (item.block) return 'BLOCK'
  if (item.warning) return 'WARN'
  if (item.total) return 'PASS'
  return '-'
}

const checkIcon = (status: string) => ({
  pass: CircleCheck,
  warning: TriangleAlert,
  block: XCircle,
  info: Info,
}[status] || Info)

const checkIconClass = (status: string): string => ({
  pass: 'bg-emerald-500/10 text-emerald-700',
  warning: 'bg-amber-500/10 text-amber-700',
  block: 'bg-rose-500/10 text-rose-700',
  info: 'bg-blue-500/10 text-blue-700',
}[status] || 'bg-muted text-muted-foreground')

const containerSummary = (project: OpsProject): string => {
  if (!project.last_checked_at) return '未同步'
  return `${project.observed_container_count || 0} / ${project.observed_running_container_count || 0} / ${project.observed_healthy_container_count || 0}`
}

watch(
  [selectedProjectId, activeCategory, detailMode, overviewStatusFilter, overviewEnvironmentFilter, overviewSort],
  async ([projectID, category, mode, status, environment, sort]) => {
    const query: Record<string, string> = {}
    if (projectID) query.project = String(projectID)
    if (category && category !== 'all') query.category = category
    if (mode && mode !== 'all') query.mode = mode
    if (status && status !== 'all') query.status = status
    if (environment && environment !== 'all') query.environment = environment
    if (sort && sort !== 'risk') query.sort = sort
    if (JSON.stringify(route.query) !== JSON.stringify(query)) {
      await router.replace({ query })
    }
  },
)

onMounted(async () => {
  await loadProjects()
  if (!selectedProjectId.value && projectOptions.value.length) {
    selectedProjectId.value = projectOptions.value[0].id
  }
  if (selectedProjectId.value) {
    await generateReport()
  }
})
</script>
