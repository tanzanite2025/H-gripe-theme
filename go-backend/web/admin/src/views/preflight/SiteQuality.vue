<template>
 <div class="space-y-4">
    <AdminPageHeader title="上线前检查 / 页面质量" description="按需运行的内部 Lighthouse 页面质量检测">
      <template #actions>
        <Button size="icon" variant="outline" title="刷新检测历史" :disabled="loading || targetOptionsLoading" @click="refreshSiteQualityData">
 <RefreshCw :class="['size-4', loading || targetOptionsLoading ? 'animate-spin': '']" />
        </Button>
      </template>
    </AdminPageHeader>

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
        <TabsTrigger value="overview" class="max-w-36 flex-none rounded-none px-3">
          <ListChecks class="size-3.5" />
          总览
        </TabsTrigger>
        <TabsTrigger value="headings" class="max-w-36 flex-none rounded-none px-3">
          <Heading2 class="size-3.5" />
          Headings
          <span
            v-if="headingPagination.total > 0"
            class="rounded-full bg-rose-500/10 px-1.5 py-0.5 text-[9px] text-rose-600"
          >
            {{ headingPagination.total }}
          </span>
        </TabsTrigger>
        <TabsTrigger value="schema" class="max-w-40 flex-none rounded-none px-3">
          <Braces class="size-3.5" />
          Schema
          <span
            v-if="schemaPagination.total > 0"
            class="rounded-full bg-rose-500/10 px-1.5 py-0.5 text-[9px] text-rose-600"
          >
            {{ schemaPagination.total }}
          </span>
        </TabsTrigger>
      </TabsList>

      <TabsContent value="overview" class="mt-0 space-y-4">
    <SiteQualityFindings
      :loading="findingsLoading"
      :findings="findings"
      :state="findingStateFilter"
      :pagination="findingPagination"
      @change-filter="applyFindingFilter"
      @refresh="() => loadFindings()"
      @open="openFinding"
      @change-page="changeFindingPage"
    />

 <section v-if="selectedRun" class="space-y-3">
 <div class="flex flex-wrap items-start justify-between gap-3 border bg-card px-4 py-3">
 <div class="min-w-0">
 <p class="truncate text-sm font-black">{{ selectedRun.final_url || selectedRun.target_url }}</p>
 <p class="mt-1 text-xs text-muted-foreground">
            {{ strategyLabel(selectedRun.strategy) }} · {{ formatDate(selectedRun.created_at) }}
          </p>
        </div>
        <AdminStatusBadge :tone="runTone(selectedRun.status)">
          {{ selectedRun.status === 'success' ? '已完成' : '检测失败' }}
        </AdminStatusBadge>
      </div>

 <div v-if="selectedRun.status === 'failed'" class="border-l-2 border-rose-500 bg-card px-4 py-3 text-sm text-rose-700">
        {{ selectedRun.error_message || '内部 Lighthouse 运行器未返回可用结果' }}
      </div>

      <template v-else>
 <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
 <div v-for="score in scoreItems(selectedRun)" :key="score.label" class="border bg-card p-4">
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">{{ score.label }}</p>
 <p class="mt-3 text-3xl font-black" :class="scoreTone(score.value)">{{ displayScore(score.value) }}</p>
          </div>
        </section>

 <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
 <div v-for="metric in metricItems(selectedRun)" :key="metric.label" class="border bg-card p-3">
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">{{ metric.label }}</p>
 <p class="mt-2 text-lg font-black">{{ metric.value }}</p>
          </div>
        </section>

 <section class="border bg-card">
 <div class="flex items-center justify-between gap-3 border-b px-4 py-3">
            <div>
 <p class="text-sm font-black">待处理项</p>
 <p class="mt-1 text-xs text-muted-foreground">{{ selectedRun.issues.length }} 个检测项</p>
            </div>
 <Gauge class="size-4 text-muted-foreground" />
          </div>
 <div v-if="selectedRun.issues.length === 0" class="p-8 text-center text-sm text-muted-foreground">
            当前检测未发现低分项
          </div>
 <div v-else class="divide-y">
            <article
              v-for="issue in selectedRun.issues"
              :key="issue.id"
 class="grid gap-3 px-4 py-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center"
            >
 <div class="min-w-0">
 <div class="flex flex-wrap items-center gap-2">
 <p class="text-sm font-black">{{ issue.title }}</p>
                  <AdminStatusBadge :tone="issueTone(issue.severity)">
                    {{ issueSeverityLabel(issue.severity) }}
                  </AdminStatusBadge>
                </div>
 <p v-if="issue.display_value || issueSavings(issue)" class="mt-1 text-xs font-bold text-muted-foreground">
                  {{ issue.display_value || issueSavings(issue) }}
                </p>
 <p v-if="issue.description" class="mt-2 text-xs leading-5 text-muted-foreground">{{ issue.description }}</p>
              </div>
              <Button
                v-if="issue.remediation"
                variant="outline"
                size="sm"
                :title="issue.remediation.label"
                @click="openRemediation(issue.remediation.route)"
              >
 <Wrench class="size-3.5" />
                {{ issue.remediation.label }}
              </Button>
            </article>
          </div>
        </section>
      </template>
    </section>

    <SiteQualityRuns
      :loading="loading"
      :runs="runs"
      :selected-run="selectedRun"
      :pagination="pagination"
      @select-run="selectedRun = $event"
      @change-page="changePage"
    />
      </TabsContent>

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
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Braces, Gauge, Heading2, ListChecks, LoaderCircle, Play, RefreshCw, Wrench } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import preflightApi, {
  type SiteQualityOperationalSummary,
  type SiteQualityFinding,
  type SiteQualityFindingEvidence,
  type SiteQualityFindingEvent,
  type SiteQualityFindingStateFilter,
  type SiteQualityRun,
  type SiteQualityTargetOption,
  type SiteQualityStrategy,
} from '@/api/preflight'
import { useAuthStore } from '@/stores/auth'
import SiteQualityFindingDialog from '@/components/admin/site-quality/SiteQualityFindingDialog.vue'
import SiteQualityFindings from '@/components/admin/site-quality/SiteQualityFindings.vue'
import SiteQualityHeadings from '@/components/admin/site-quality/SiteQualityHeadings.vue'
import SiteQualityStructuredData from '@/components/admin/site-quality/SiteQualityStructuredData.vue'
import SiteQualityRuns from '@/components/admin/site-quality/SiteQualityRuns.vue'
import SiteQualitySummary from '@/components/admin/site-quality/SiteQualitySummary.vue'
import { useSiteQualityJobs } from '@/composables/useSiteQualityJobs'

const router = useRouter()
const authStore = useAuthStore()
const { enqueueInspection, enqueueFindingRecheck } = useSiteQualityJobs()
const canManage = computed(() => authStore.hasPermission('services:manage'))
const loading = ref(false)
const running = ref(false)
const activeQualityTab = ref<'overview' | 'headings' | 'schema'>('overview')
const targetURL = ref('')
const targetOptionsLoading = ref(false)
const targetOptions = ref<SiteQualityTargetOption[]>([])
const strategy = ref<SiteQualityStrategy>('mobile')
const runs = ref<SiteQualityRun[]>([])
const selectedRun = ref<SiteQualityRun | null>(null)
const pagination = ref({ page: 1, page_size: 20, total: 0, total_pages: 1 })
const runnerConfigured = ref(false)
const operationalSummary = ref<SiteQualityOperationalSummary | null>(null)
const findingsLoading = ref(false)
const findings = ref<SiteQualityFinding[]>([])
const findingStateFilter = ref<SiteQualityFindingStateFilter>('active')
const findingPagination = ref({ page: 1, page_size: 20, total: 0, total_pages: 1 })
const headingsLoading = ref(false)
const headingFindings = ref<SiteQualityFinding[]>([])
const headingStateFilter = ref<SiteQualityFindingStateFilter>('active')
const headingPagination = ref({ page: 1, page_size: 20, total: 0, total_pages: 1 })
const schemaLoading = ref(false)
const schemaFindings = ref<SiteQualityFinding[]>([])
const schemaStateFilter = ref<SiteQualityFindingStateFilter>('active')
const schemaPagination = ref({ page: 1, page_size: 20, total: 0, total_pages: 1 })
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
    value: operationalSummary.value?.worker_enabled ? '运行中' : '已暂停',
    tone: operationalSummary.value?.worker_enabled ? 'text-emerald-600' : 'text-amber-600',
    badge: operationalSummary.value ? `${operationalSummary.value.worker_interval_seconds}s` : undefined,
    badgeTone: operationalSummary.value?.worker_enabled ? 'green' as const : 'amber' as const,
    hint: operationalSummary.value?.worker_enabled
      ? '到期目标、运营检测和复检都进入任务队列'
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

const loadRuns = async (page = pagination.value.page, preserveSelection = false): Promise<void> => {
  loading.value = true
  try {
    const data = await preflightApi.getSiteQualityRuns({ page, pageSize: 20 })
    runs.value = data.items
    pagination.value = data.pagination
    runnerConfigured.value = data.runner_configured
    operationalSummary.value = data.summary
    if (!targetURL.value && data.default_url) targetURL.value = data.default_url
    if (targetOptions.value.length === 0 && data.default_url) {
      targetOptions.value = [fallbackTargetOption(data.default_url)]
    }
    if (!preserveSelection) selectedRun.value = data.items[0] || null
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '页面质量检测历史加载失败')
  } finally {
    loading.value = false
  }
}

const refreshSiteQualityData = async (): Promise<void> => {
  await loadTargets()
  await Promise.all([loadRuns(1), loadFindings(1), loadHeadingFindings(1), loadSchemaFindings(1)])
}

const runInspection = async (): Promise<void> => {
  if (!canManage.value || !targetURL.value) return
  running.value = true
  try {
    const job = await enqueueInspection(targetURL.value, strategy.value)
    if (job.status !== 'succeeded') {
      throw new Error(job.last_error || `页面质量任务 ${job.status}`)
    }
    toast.success('页面质量检测完成')
    await Promise.all([loadRuns(1, true), loadFindings(1), loadHeadingFindings(1), loadSchemaFindings(1)])
    selectedRun.value = runs.value[0] || null
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '页面质量检测失败')
    await loadRuns(1, true)
  } finally {
    running.value = false
  }
}

const loadFindings = async (page = findingPagination.value.page): Promise<void> => {
  findingsLoading.value = true
  try {
    const data = await preflightApi.getSiteQualityFindings({
      page,
      pageSize: findingPagination.value.page_size,
      state: findingStateFilter.value,
      kind: 'opportunity',
    })
    findings.value = data.items
    findingPagination.value = data.pagination
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '页面质量事项加载失败')
  } finally {
    findingsLoading.value = false
  }
}

const applyFindingFilter = (state: SiteQualityFindingStateFilter): void => {
  findingStateFilter.value = state
  findingPagination.value.page = 1
  void loadFindings(1)
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

const changeFindingPage = (page: number): void => {
  if (page < 1 || page > findingPagination.value.total_pages || page === findingPagination.value.page) return
  void loadFindings(page)
}

const changeHeadingPage = (page: number): void => {
  if (page < 1 || page > headingPagination.value.total_pages || page === headingPagination.value.page) return
  void loadHeadingFindings(page)
}

const changeSchemaPage = (page: number): void => {
  if (page < 1 || page > schemaPagination.value.total_pages || page === schemaPagination.value.page) return
  void loadSchemaFindings(page)
}

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
  await Promise.all([loadFindings(), loadHeadingFindings(), loadSchemaFindings(), loadFindingEvents(finding.id)])
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
  try {
    const job = await enqueueFindingRecheck(selectedFinding.value.id)
    if (job.status !== 'succeeded') {
      throw new Error(job.last_error || `页面质量任务 ${job.status}`)
    }
    const finding = await preflightApi.getSiteQualityFinding(selectedFinding.value.id)
    await refreshSelectedFinding(finding)
    await loadRuns(1, true)
    selectedRun.value = runs.value[0] || null
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

const changePage = (page: number): void => {
  if (page < 1 || page > pagination.value.total_pages || page === pagination.value.page) return
  void loadRuns(page)
}

const openRemediation = (route: string): void => {
  void router.push(route)
}

const displayScore = (score?: number): string => typeof score === 'number' ? String(score) : '-'
const scoreTone = (score?: number): string => {
  if (typeof score !== 'number') return ''
  if (score >= 90) return 'text-emerald-600'
  if (score >= 50) return 'text-amber-600'
  return 'text-rose-600'
}
const scoreItems = (run: SiteQualityRun) => [
  { label: '性能', value: run.performance_score },
  { label: '无障碍', value: run.accessibility_score },
  { label: '最佳实践', value: run.best_practices_score },
  { label: 'SEO', value: run.seo_score },
]
const metricItems = (run: SiteQualityRun) => [
  { label: 'FCP', value: formatMilliseconds(run.first_contentful_paint_ms) },
  { label: 'LCP', value: formatMilliseconds(run.largest_contentful_paint_ms) },
  { label: 'INP', value: formatMilliseconds(run.interaction_to_next_paint_ms) },
  { label: 'CLS', value: typeof run.cumulative_layout_shift === 'number' ? run.cumulative_layout_shift.toFixed(3) : '-' },
  { label: 'TBT', value: formatMilliseconds(run.total_blocking_time_ms) },
  { label: 'Speed Index', value: formatMilliseconds(run.speed_index_ms) },
]
const issueSavings = (issue: SiteQualityRun['issues'][number]): string => {
  if (typeof issue.savings_ms === 'number' && issue.savings_ms > 0) {
    return `预计节省 ${formatMilliseconds(issue.savings_ms)}`
  }
  if (typeof issue.savings_bytes === 'number' && issue.savings_bytes > 0) {
    return `预计节省 ${(issue.savings_bytes / 1024).toFixed(1)} KiB`
  }
  return ''
}
const formatMilliseconds = (value?: number): string => {
  if (typeof value !== 'number') return '-'
  if (value < 1000) return `${Math.round(value)} ms`
  return `${(value / 1000).toFixed(2)} s`
}
const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')
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
const strategyLabel = (value: SiteQualityStrategy): string => value === 'mobile' ? '移动端' : '桌面端'
const runTone = (status: SiteQualityRun['status']): AdminStatusTone => status === 'success' ? 'green' : 'coral'
const issueTone = (severity: string): AdminStatusTone => (
  severity === 'critical' || severity === 'high'
    ? 'coral'
    : severity === 'medium'
      ? 'amber'
      : 'gray'
)
const issueSeverityLabel = (severity: string): string => ({
  critical: '严重',
  high: '高',
  medium: '中',
  low: '低',
}[severity] || severity)

const targetOptionLabel = (option: SiteQualityTargetOption): string => {
  const parts: string[] = []
  const title = option.title.trim()
  if (option.is_home) parts.push('首页')
  if (title && (!option.is_home || title !== '首页')) parts.push(title)
  if (option.path.trim()) parts.push(option.path.trim())
  if (option.locale.trim()) parts.push(option.locale.trim())
  return parts.join(' · ') || option.url
}

onMounted(() => {
  void (async () => {
    await loadTargets()
    await Promise.all([loadRuns(1), loadFindings(1), loadHeadingFindings(1), loadSchemaFindings(1)])
  })()
})
</script>
