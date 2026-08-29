<template>
 <div class="space-y-4">
    <AdminPageHeader title="上线前检查 / 页面质量" description="H1 层级与 Schema 独立检查，仅手动触发">
      <template #actions>
        <Button
          v-if="canManage && terminalJobCount > 0"
          size="icon"
          variant="outline"
          title="清理失败和历史死信任务"
          aria-label="清理失败和历史死信任务"
          :disabled="loading || targetOptionsLoading || cleaningJobs"
          @click="cleanupTerminalJobs"
        >
          <LoaderCircle v-if="cleaningJobs" class="size-4 animate-spin" />
          <Trash2 v-else class="size-4 text-destructive" />
        </Button>
        <Button size="icon" variant="outline" title="刷新页面质量检查" :disabled="loading || targetOptionsLoading || cleaningJobs" @click="refreshSiteQualityData">
 <RefreshCw :class="['size-4', loading || targetOptionsLoading ? 'animate-spin': '']" />
        </Button>
      </template>
    </AdminPageHeader>

    <Alert v-if="activeJob" :variant="activeJob.status === 'failed' || activeJob.status === 'dead_letter' ? 'destructive' : 'default'" class="rounded-[20px] border-dashed">
      <component :is="activeJobIcon" class="size-4" />
      <div class="min-w-0 space-y-1">
        <div class="flex flex-wrap items-center gap-2">
          <p class="text-xs font-black">{{ activeJobTitle }}</p>
          <AdminStatusBadge :tone="activeJobTone">
            {{ activeJobStatusLabel }}
          </AdminStatusBadge>
          <span class="font-mono text-[10px] text-muted-foreground">#{{ activeJob.id }}</span>
        </div>
        <p class="text-xs leading-5 text-foreground">
          {{ activeJobSummary }}
        </p>
        <p class="text-[11px] leading-5 text-muted-foreground">
          {{ activeJobDetail }}
        </p>
        <p v-if="activeJob.last_error" class="text-xs leading-5 text-destructive">
          {{ activeJob.last_error }}
        </p>
      </div>
    </Alert>

    <SiteQualitySummary :items="summaryItems" :warnings="summaryWarnings" />

    <section class="border bg-card p-4">
 <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_10rem_auto] lg:items-end">
 <label class="grid gap-1.5">
 <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">页面地址</span>
          <Select
            v-model="targetURL"
            :disabled="running || !canManage || targetOptionsLoading"
          >
            <SelectTrigger class="h-10 w-full min-w-0">
              <SelectValue placeholder="选择要检测的页面" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem
                v-for="option in targetOptions"
                :key="option.url"
                :value="option.url"
              >
                {{ targetOptionLabel(option) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </label>
 <label class="grid gap-1.5">
 <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">策略</span>
 <select v-model="strategy" class="h-10 border bg-background px-3 text-sm" :disabled="running">
            <option value="mobile">移动端</option>
            <option value="desktop">桌面端</option>
          </select>
        </label>
        <Button :disabled="running || targetOptionsLoading || !canManage || !targetURL" @click="runInspection">
 <LoaderCircle v-if="running" class="size-4 animate-spin" />
 <Play v-else class="size-4" />
          运行检测
        </Button>
      </div>
    </section>

    <Tabs v-model="activeQualityTab" class="gap-4">
      <TabsList variant="line" class="h-9 w-full justify-start border-b bg-transparent p-0">
        <TabsTrigger value="headings" class="max-w-40 flex-none rounded-none px-3">
          <Heading2 class="size-3.5" />
          H1 层级
        </TabsTrigger>
        <TabsTrigger value="schema" class="max-w-40 flex-none rounded-none px-3">
          <Braces class="size-3.5" />
          Schema
        </TabsTrigger>
        <TabsTrigger value="links" class="max-w-40 flex-none rounded-none px-3">
          <Link2 class="size-3.5" />
          链接文字
        </TabsTrigger>
      </TabsList>

      <TabsContent value="headings" class="mt-0">
        <SiteQualityHeadings
          :loading="headingsLoading"
          :findings="headingFindings"
          :state="headingStateFilter"
          :pagination="headingPagination"
          @change-filter="applyHeadingFilter"
          @refresh="() => loadHeadingFindings()"
          @open="openFinding"
          @change-page="changeHeadingPage"
        />
      </TabsContent>

      <TabsContent value="schema" class="mt-0">
        <SiteQualityStructuredData
          :loading="schemaLoading"
          :findings="schemaFindings"
          :state="schemaStateFilter"
          :pagination="schemaPagination"
          @change-filter="applySchemaFilter"
          @refresh="() => loadSchemaFindings()"
          @open="openFinding"
          @change-page="changeSchemaPage"
        />
      </TabsContent>

      <TabsContent value="links" class="mt-0">
        <SiteQualityLinks
          :loading="linksLoading"
          :findings="linkFindings"
          :state="linkStateFilter"
          :pagination="linkPagination"
          @change-filter="applyLinkFilter"
          @refresh="() => loadLinkFindings()"
          @open="openFinding"
          @change-page="changeLinkPage"
        />
      </TabsContent>
    </Tabs>

    <SiteQualityFindingDialog
      v-model:open="findingDetailOpen"
      :detail-loading="findingDetailLoading"
      :selected-finding="selectedFinding"
      :selected-finding-evidence="selectedFindingEvidence"
      :finding-events="findingEvents"
      :finding-events-pagination="findingEventsPagination"
      :finding-action-key="findingActionKey"
      :can-manage="canManage"
      :resolution-note="findingResolutionNote"
      @update:resolution-note="findingResolutionNote = $event"
      @acknowledge="acknowledgeFinding"
      @resolve="resolveFinding"
      @recheck="recheckFinding"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { AlertTriangle, Braces, CheckCircle2, Clock3, Heading2, Link2, LoaderCircle, Play, RefreshCw, Trash2 } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import preflightApi, {
  SITE_QUALITY_RULE_ID_DESCRIPTIVE_LINK_TEXT,
  type SiteQualityJob,
  type SiteQualityOperationalSummary,
  type SiteQualityJobCleanupResult,
  type SiteQualityFinding,
  type SiteQualityFindingEvidence,
  type SiteQualityFindingEvent,
  type SiteQualityFindingStateFilter,
  type SiteQualityTargetOption,
  type SiteQualityStrategy,
} from '@/api/preflight'
import { useAuthStore } from '@/stores/auth'
import SiteQualityFindingDialog from '@/components/admin/site-quality/SiteQualityFindingDialog.vue'
import SiteQualityHeadings from '@/components/admin/site-quality/SiteQualityHeadings.vue'
import SiteQualityStructuredData from '@/components/admin/site-quality/SiteQualityStructuredData.vue'
import SiteQualityLinks from '@/components/admin/site-quality/SiteQualityLinks.vue'
import SiteQualitySummary from '@/components/admin/site-quality/SiteQualitySummary.vue'
import { useRouteTab } from '@/composables/useRouteTab'
import { useSiteQualityJobs } from '@/composables/useSiteQualityJobs'

const authStore = useAuthStore()
const { enqueueInspection, enqueueFindingRecheck } = useSiteQualityJobs()
const canManage = computed(() => authStore.hasPermission('services:manage'))
const loading = ref(false)
const running = ref(false)
const activeQualityTab = useRouteTab<'headings' | 'schema' | 'links'>({
  defaultValue: 'headings',
  values: ['headings', 'schema', 'links'],
  routes: {
    headings: 'PreflightSiteQualityHeadings',
    schema: 'PreflightSiteQualitySchema',
    links: 'PreflightSiteQualityLinks',
  },
})
const targetURL = ref('')
const targetOptionsLoading = ref(false)
const targetOptions = ref<SiteQualityTargetOption[]>([])
const strategy = ref<SiteQualityStrategy>('mobile')
const runnerConfigured = ref(false)
const operationalSummary = ref<SiteQualityOperationalSummary | null>(null)
const activeJob = ref<SiteQualityJob | null>(null)
const activeJobContext = ref<{ title: string; target: string } | null>(null)
const activeJobClock = ref(Date.now())
let activeJobClockTimer: number | null = null
const cleaningJobs = ref(false)
const headingsLoading = ref(false)
const headingFindings = ref<SiteQualityFinding[]>([])
const headingStateFilter = ref<SiteQualityFindingStateFilter>('active')
const headingPagination = ref({ page: 1, page_size: 20, total: 0, total_pages: 1 })
const schemaLoading = ref(false)
const schemaFindings = ref<SiteQualityFinding[]>([])
const schemaStateFilter = ref<SiteQualityFindingStateFilter>('active')
const schemaPagination = ref({ page: 1, page_size: 20, total: 0, total_pages: 1 })
const linksLoading = ref(false)
const linkFindings = ref<SiteQualityFinding[]>([])
const linkStateFilter = ref<SiteQualityFindingStateFilter>('active')
const linkPagination = ref({ page: 1, page_size: 20, total: 0, total_pages: 1 })
const findingDetailOpen = ref(false)
const findingDetailLoading = ref(false)
const findingActionKey = ref<string | null>(null)
const selectedFinding = ref<SiteQualityFinding | null>(null)
const findingEvents = ref<SiteQualityFindingEvent[]>([])
const findingEventsPagination = ref({ page: 1, page_size: 50, total: 0, total_pages: 1 })
const findingResolutionNote = ref('')

const summaryItems = computed(() => [
  {
    label: '系统状态',
    value: operationalStatusLabel(operationalSummary.value?.status),
    tone: operationalStatusTone(operationalSummary.value?.status),
    badge: operationalStatusBadgeLabel(operationalSummary.value?.status),
    badgeTone: operationalSummary.value?.status ? operationalStatusBadgeTone(operationalSummary.value.status) : undefined,
    hint: operationalSummary.value?.warnings?.[0] || '',
  },
  {
    label: '运行器',
    value: runnerConfigured.value ? '就绪' : '未配置',
    tone: runnerConfigured.value ? 'text-emerald-600' : 'text-rose-600',
    hint: operationalSummary.value?.default_url || '仅允许同源 storefront 目标',
  },
  {
    label: '调度',
    value: operationalSummary.value?.worker_enabled
      ? operationalSummary.value.auto_scan_enabled ? '自动扫描' : '手动模式'
      : '已暂停',
    tone: operationalSummary.value?.worker_enabled ? 'text-emerald-600' : 'text-amber-600',
    badge: operationalSummary.value?.worker_enabled
      ? operationalSummary.value.auto_scan_enabled
        ? `${operationalSummary.value.worker_interval_seconds}s`
        : '仅手动'
      : undefined,
    badgeTone: operationalSummary.value?.worker_enabled ? 'green' as const : 'amber' as const,
    hint: operationalSummary.value?.worker_enabled
      ? operationalSummary.value.auto_scan_enabled
        ? '到期目标、运营检测和复检都进入任务队列'
        : '不会按 30 秒自动扫描，仅执行手动检测和问题复检'
      : '任务会保留在队列中等待 worker 恢复',
  },
  {
    label: '队列',
    value: operationalSummary.value ? `${operationalSummary.value.jobs.claimable}` : '-',
    tone: operationalSummary.value?.jobs.failed || operationalSummary.value?.jobs.dead_letter
      ? 'text-rose-600'
      : operationalSummary.value?.jobs.claimable
        ? 'text-amber-600'
        : 'text-emerald-600',
    hint: operationalSummary.value
      ? `处理中 ${operationalSummary.value.jobs.processing} · 重试 ${operationalSummary.value.jobs.failed} · 死信 ${operationalSummary.value.jobs.dead_letter}`
      : '',
  },
  {
    label: '目标',
    value: operationalSummary.value
      ? `${operationalSummary.value.targets.enabled}/${operationalSummary.value.targets.total}`
      : '-',
    tone: operationalSummary.value?.targets.enabled ? 'text-emerald-600' : 'text-muted-foreground',
    hint: operationalSummary.value
      ? `已登记 ${operationalSummary.value.targets.enabled} 个目标 · 活跃事项 ${operationalSummary.value.findings.active}`
      : '',
  },
  {
    label: '租约',
    value: operationalSummary.value ? `${operationalSummary.value.jobs.stale_leases}` : '-',
    tone: operationalSummary.value?.jobs.stale_leases ? 'text-amber-600' : 'text-emerald-600',
    hint: operationalSummary.value
      ? `可回收 ${operationalSummary.value.provider_slots.available} · 并发 ${operationalSummary.value.provider_slots.configured}`
      : '',
  },
  {
    label: '最近成功',
    value: operationalSummary.value?.latest_success_at
      ? formatDate(operationalSummary.value.latest_success_at)
      : '-',
    tone: operationalSummary.value?.latest_success_at ? '' : 'text-muted-foreground',
    hint: operationalSummary.value?.latest_run ? `Run #${operationalSummary.value.latest_run.id}` : '',
  },
])

const summaryWarnings = computed(() => operationalSummary.value?.warnings || [])

const terminalJobCount = computed(() => {
  const jobs = operationalSummary.value?.jobs
  return (jobs?.failed || 0) + (jobs?.dead_letter || 0)
})

const activeJobIsLive = computed(() => activeJob.value !== null && (activeJob.value.status === 'queued' || activeJob.value.status === 'processing'))
const activeJobTone = computed<AdminStatusTone>(() => ({
  queued: 'amber',
  processing: 'blue',
  succeeded: 'green',
  failed: 'coral',
  dead_letter: 'coral',
}[activeJob.value?.status || 'queued'] as AdminStatusTone))
const activeJobStatusLabel = computed(() => ({
  queued: '排队中',
  processing: '处理中',
  succeeded: '已完成',
  failed: '失败',
  dead_letter: '死信',
}[activeJob.value?.status || 'queued']))
const activeJobTitle = computed(() => activeJobContext.value?.title || '页面质量任务')
const activeJobTarget = computed(() => activeJobContext.value?.target || targetURL.value || '未指定目标')
const activeJobIcon = computed(() => ({
  queued: Clock3,
  processing: LoaderCircle,
  succeeded: CheckCircle2,
  failed: AlertTriangle,
  dead_letter: AlertTriangle,
}[activeJob.value?.status || 'queued']))
const activeJobSummary = computed(() => {
  const job = activeJob.value
  if (!job) return ''
  const jobLabel = `任务 #${job.id}`
  switch (job.status) {
    case 'queued':
      return `${jobLabel} 已进入队列，worker 取到后会继续执行。`
    case 'processing':
      return `${jobLabel} 正在执行采样，关闭页面不会中断后台任务。`
    case 'succeeded':
      return `${jobLabel} 已完成，下面的检查结果已经刷新。`
    case 'failed':
      return `${jobLabel} 本次执行失败，请查看最后错误。`
    case 'dead_letter':
      return `${jobLabel} 已达到最大重试次数，进入死信状态。`
    default:
      return `${jobLabel} 正在执行。`
  }
})
const activeJobDetail = computed(() => {
  const job = activeJob.value
  if (!job) return ''
  const parts: string[] = [`目标 ${activeJobTarget.value}`, `策略 ${strategyLabel(job.strategy)}`, `采样 ${job.sample_count} 次 · 确认 ${job.required_confirmations} 次`]
  const startedAt = job.started_at || job.available_at || job.created_at
  if (job.status === 'queued') {
    parts.unshift(`已排队 ${formatElapsed(startedAt)}`)
  } else if (job.status === 'processing') {
    parts.unshift(`已运行 ${formatElapsed(startedAt)}`)
  } else if (job.status === 'succeeded') {
    parts.unshift(`完成于 ${formatDate(job.finished_at || job.updated_at)}`)
  } else {
    parts.unshift(`更新于 ${formatDate(job.updated_at)}`)
  }
  if (job.heartbeat_at) {
    parts.push(`最近心跳 ${formatRelativeTime(job.heartbeat_at)}`)
  }
  if (job.lease_expires_at && job.status === 'processing') {
    parts.push(`租约到期 ${formatRelativeTime(job.lease_expires_at)}`)
  }
  if (job.status === 'failed' || job.status === 'dead_letter') {
    parts.push(`重试 ${job.attempts}/${job.max_attempts}`)
  }
  if (job.status === 'queued' || job.status === 'processing') {
    parts.push('当前列表仍显示上一次完成后的结果，任务结束后会自动刷新。')
  }
  if (job.status === 'succeeded') {
    parts.push('结果已经同步到下方列表。')
  }
  return parts.join(' · ')
})

const selectedFindingEvidence = computed<SiteQualityFindingEvidence | null>(() => {
  const raw = selectedFinding.value?.latest_evidence
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed === 'object' ? parsed as SiteQualityFindingEvidence : null
  } catch {
    return null
  }
})

const fallbackTargetOption = (url: string): SiteQualityTargetOption => ({
  url,
  path: '/',
  title: '',
  locale: '',
  source_type: 'static',
  is_home: true,
})

const loadTargets = async (): Promise<void> => {
  targetOptionsLoading.value = true
  try {
    const data = await preflightApi.getSiteQualityTargets()
    const nextOptions = data.items.length > 0
      ? data.items
      : data.default_url
        ? [fallbackTargetOption(data.default_url)]
        : []
    const currentTarget = targetURL.value.trim()
    targetOptions.value = nextOptions
    if (!currentTarget || !nextOptions.some((option) => option.url === currentTarget)) {
      targetURL.value = data.default_url || nextOptions[0]?.url || ''
    }
  } catch {
    // The run history endpoint still provides a same-origin fallback URL.
  } finally {
    targetOptionsLoading.value = false
  }
}

const loadRuns = async (): Promise<void> => {
  loading.value = true
  try {
    const data = await preflightApi.getSiteQualityRuns({ page: 1, pageSize: 1 })
    runnerConfigured.value = data.runner_configured
    operationalSummary.value = data.summary
    if (!targetURL.value && data.default_url) targetURL.value = data.default_url
    if (targetOptions.value.length === 0 && data.default_url) {
      targetOptions.value = [fallbackTargetOption(data.default_url)]
    }
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '页面质量检测历史加载失败')
  } finally {
    loading.value = false
  }
}

const refreshSiteQualityData = async (): Promise<void> => {
  await loadTargets()
  await Promise.all([loadRuns(), loadActiveFindings()])
}

const runInspection = async (): Promise<void> => {
  if (!canManage.value || !targetURL.value) return
  running.value = true
  activeJobContext.value = { title: '页面质量检测', target: targetURL.value }
  try {
    const jobPromise = enqueueInspection(targetURL.value, strategy.value, (job) => {
      activeJob.value = job
    })
    void loadRuns()
    const job = await jobPromise
    activeJob.value = job
    if (job.status !== 'succeeded') {
      throw new Error(job.last_error || `页面质量任务 ${job.status}`)
    }
    toast.success('页面质量检测完成')
    await Promise.all([loadRuns(), loadActiveFindings(1)])
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '页面质量检测失败')
    await loadRuns()
  } finally {
    running.value = false
  }
}

const cleanupTerminalJobs = async (): Promise<void> => {
  if (!canManage.value || cleaningJobs.value || terminalJobCount.value <= 0) return
  const jobs = operationalSummary.value?.jobs
  const failed = jobs?.failed || 0
  const deadLetter = jobs?.dead_letter || 0
  if (!window.confirm(`确定清理 ${failed} 个重试任务和 ${deadLetter} 个历史死信任务吗？历史 Run、Findings 和成功任务会保留。`)) return

  cleaningJobs.value = true
  try {
    const result: SiteQualityJobCleanupResult = await preflightApi.cleanupSiteQualityJobs()
    toast.success(`已清理 ${result.deleted} 个旧任务（重试 ${result.failed}，死信 ${result.dead_letter}）`)
    await loadRuns()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '旧任务清理失败')
  } finally {
    cleaningJobs.value = false
  }
}

const loadHeadingFindings = async (page = headingPagination.value.page): Promise<void> => {
  headingsLoading.value = true
  try {
    const data = await preflightApi.getSiteQualityFindings({
      page,
      pageSize: headingPagination.value.page_size,
      state: headingStateFilter.value,
      kind: 'headings',
    })
    headingFindings.value = data.items
    headingPagination.value = data.pagination
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '标题检查事项加载失败')
  } finally {
    headingsLoading.value = false
  }
}

const applyHeadingFilter = (state: SiteQualityFindingStateFilter): void => {
  headingStateFilter.value = state
  headingPagination.value.page = 1
  void loadHeadingFindings(1)
}

const loadSchemaFindings = async (page = schemaPagination.value.page): Promise<void> => {
  schemaLoading.value = true
  try {
    const data = await preflightApi.getSiteQualityFindings({
      page,
      pageSize: schemaPagination.value.page_size,
      state: schemaStateFilter.value,
      kind: 'schema',
    })
    schemaFindings.value = data.items
    schemaPagination.value = data.pagination
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '结构化数据事项加载失败')
  } finally {
    schemaLoading.value = false
  }
}

const applySchemaFilter = (state: SiteQualityFindingStateFilter): void => {
  schemaStateFilter.value = state
  schemaPagination.value.page = 1
  void loadSchemaFindings(1)
}

const loadLinkFindings = async (page = linkPagination.value.page): Promise<void> => {
  linksLoading.value = true
  try {
    const data = await preflightApi.getSiteQualityFindings({
      page,
      pageSize: linkPagination.value.page_size,
      state: linkStateFilter.value,
      kind: 'links',
      ruleID: SITE_QUALITY_RULE_ID_DESCRIPTIVE_LINK_TEXT,
    })
    linkFindings.value = data.items
    linkPagination.value = data.pagination
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '链接文字检查事项加载失败')
  } finally {
    linksLoading.value = false
  }
}

const applyLinkFilter = (state: SiteQualityFindingStateFilter): void => {
  linkStateFilter.value = state
  linkPagination.value.page = 1
  void loadLinkFindings(1)
}

const changeHeadingPage = (page: number): void => {
  if (page < 1 || page > headingPagination.value.total_pages || page === headingPagination.value.page) return
  void loadHeadingFindings(page)
}

const changeSchemaPage = (page: number): void => {
  if (page < 1 || page > schemaPagination.value.total_pages || page === schemaPagination.value.page) return
  void loadSchemaFindings(page)
}

const changeLinkPage = (page: number): void => {
  if (page < 1 || page > linkPagination.value.total_pages || page === linkPagination.value.page) return
  void loadLinkFindings(page)
}

const loadActiveFindings = (page?: number): Promise<void> => (
  activeQualityTab.value === 'schema'
    ? loadSchemaFindings(page)
    : activeQualityTab.value === 'links'
      ? loadLinkFindings(page)
      : loadHeadingFindings(page)
)

const loadFindingEvents = async (findingID: number): Promise<void> => {
  const data = await preflightApi.getSiteQualityFindingEvents(findingID, {
    page: findingEventsPagination.value.page,
    pageSize: findingEventsPagination.value.page_size,
  })
  findingEvents.value = data.items
  findingEventsPagination.value = data.pagination
}

const openFinding = async (findingID: number): Promise<void> => {
  findingDetailOpen.value = true
  findingDetailLoading.value = true
  findingEventsPagination.value.page = 1
  try {
    const [finding] = await Promise.all([
      preflightApi.getSiteQualityFinding(findingID),
      loadFindingEvents(findingID),
    ])
    selectedFinding.value = finding
    findingResolutionNote.value = ''
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '页面质量事项详情加载失败')
    findingDetailOpen.value = false
  } finally {
    findingDetailLoading.value = false
  }
}

const refreshSelectedFinding = async (finding: SiteQualityFinding): Promise<void> => {
  selectedFinding.value = finding
  const reloadFindingList = finding.finding_kind === 'schema'
    ? loadSchemaFindings()
    : finding.finding_kind === 'links'
      ? loadLinkFindings()
      : loadHeadingFindings()
  await Promise.all([reloadFindingList, loadFindingEvents(finding.id)])
}

const acknowledgeFinding = async (): Promise<void> => {
  if (!selectedFinding.value || !canManage.value || findingActionKey.value) return
  findingActionKey.value = 'acknowledge'
  try {
    const finding = await preflightApi.acknowledgeSiteQualityFinding(selectedFinding.value.id)
    await refreshSelectedFinding(finding)
    toast.success('已确认处理该质量事项')
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '确认处理失败')
  } finally {
    findingActionKey.value = null
  }
}

const resolveFinding = async (): Promise<void> => {
  if (!selectedFinding.value || !canManage.value || findingActionKey.value) return
  const note = findingResolutionNote.value.trim()
  if (!note) return
  findingActionKey.value = 'resolve'
  try {
    const finding = await preflightApi.resolveSiteQualityFinding(selectedFinding.value.id, note)
    findingResolutionNote.value = ''
    await refreshSelectedFinding(finding)
    toast.success('已记录解决方案，等待复检验证')
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '记录解决方案失败')
  } finally {
    findingActionKey.value = null
  }
}

const recheckFinding = async (): Promise<void> => {
  if (!selectedFinding.value || !canManage.value || findingActionKey.value) return
  findingActionKey.value = 'recheck'
  activeJobContext.value = {
    title: '页面质量复检',
    target: selectedFinding.value.target_url,
  }
  try {
    const jobPromise = enqueueFindingRecheck(selectedFinding.value.id, (job) => {
      activeJob.value = job
    })
    void loadRuns()
    const job = await jobPromise
    activeJob.value = job
    if (job.status !== 'succeeded') {
      throw new Error(job.last_error || `页面质量任务 ${job.status}`)
    }
    const finding = await preflightApi.getSiteQualityFinding(selectedFinding.value.id)
    await refreshSelectedFinding(finding)
    await loadRuns()
    if (finding.state === 'verified') {
      toast.success('复检通过，质量事项已验证关闭')
    } else if (finding.state === 'open') {
      toast.success('复检确认问题仍存在，质量事项已重新打开')
    } else {
      toast.success('复检样本已记录，等待连续清洁评估完成验证')
    }
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '页面质量复检失败')
  } finally {
    findingActionKey.value = null
  }
}

const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')
const formatElapsed = (value: string): string => {
  const start = new Date(value).getTime()
  if (!Number.isFinite(start)) return '-'
  const diffSeconds = Math.max(0, Math.floor((activeJobClock.value - start) / 1000))
  if (diffSeconds < 60) return `${diffSeconds}s`
  const minutes = Math.floor(diffSeconds / 60)
  const seconds = diffSeconds % 60
  if (minutes < 60) return `${minutes}m${seconds.toString().padStart(2, '0')}s`
  const hours = Math.floor(minutes / 60)
  return `${hours}h${(minutes % 60).toString().padStart(2, '0')}m`
}
const formatRelativeTime = (value: string): string => {
  const timestamp = new Date(value).getTime()
  if (!Number.isFinite(timestamp)) return '-'
  const diffSeconds = Math.round((activeJobClock.value - timestamp) / 1000)
  const absSeconds = Math.abs(diffSeconds)
  if (absSeconds < 60) return diffSeconds >= 0 ? `${absSeconds}s前` : `${absSeconds}s后`
  const minutes = Math.round(absSeconds / 60)
  if (minutes < 60) return diffSeconds >= 0 ? `${minutes} 分钟前` : `${minutes} 分钟后`
  const hours = Math.round(absSeconds / 3600)
  return diffSeconds >= 0 ? `${hours} 小时前` : `${hours} 小时后`
}
const operationalStatusLabel = (status?: SiteQualityOperationalSummary['status']): string => ({
  healthy: '健康',
  degraded: '降级',
  not_configured: '未配置',
  unavailable: '不可用',
}[status || ''] || '-')
const operationalStatusBadgeLabel = (status?: SiteQualityOperationalSummary['status']): string => ({
  healthy: '正常',
  degraded: '注意',
  not_configured: '未配置',
  unavailable: '异常',
}[status || ''] || '-')
const operationalStatusTone = (status?: SiteQualityOperationalSummary['status']): string => ({
  healthy: 'text-emerald-600',
  degraded: 'text-amber-600',
  not_configured: 'text-muted-foreground',
  unavailable: 'text-rose-600',
}[status || ''] || '')
const operationalStatusBadgeTone = (status?: SiteQualityOperationalSummary['status']): AdminStatusTone => ({
  healthy: 'green',
  degraded: 'amber',
  not_configured: 'gray',
  unavailable: 'coral',
} as const)[status || 'not_configured']

const strategyLabel = (value: SiteQualityStrategy): string => (value === 'mobile' ? '移动端' : '桌面端')

const targetOptionLabel = (option: SiteQualityTargetOption): string => {
  const parts: string[] = []
  const title = option.title.trim()
  if (option.is_home) parts.push('首页')
  if (title && (!option.is_home || title !== '首页')) parts.push(title)
  if (option.path.trim()) parts.push(option.path.trim())
  if (option.locale.trim()) parts.push(option.locale.trim())
  return parts.join(' · ') || option.url
}

const startActiveJobClock = (): void => {
  if (activeJobClockTimer !== null) return
  activeJobClock.value = Date.now()
  activeJobClockTimer = window.setInterval(() => {
    activeJobClock.value = Date.now()
  }, 1000)
}

const stopActiveJobClock = (): void => {
  if (activeJobClockTimer === null) return
  window.clearInterval(activeJobClockTimer)
  activeJobClockTimer = null
}

watch(activeJobIsLive, (isLive) => {
  if (isLive) {
    startActiveJobClock()
  } else {
    stopActiveJobClock()
  }
}, { immediate: true })

onMounted(() => {
  void (async () => {
    await loadTargets()
    await loadRuns()
    await loadActiveFindings(1)
  })()
})

watch(activeQualityTab, () => {
  void loadActiveFindings(1)
})

onBeforeUnmount(() => {
  stopActiveJobClock()
})
</script>
