<template>
 <div class="space-y-4">
    <AdminPageHeader
      title="URL 管理 / 问题队列"
      description="刷新列表只读取已保存结果；重新扫描会先更新 URL 来源清单，再分批检查全部待处理 URL。修复后会自动验证关闭，历史记录仍保留。"
    >
      <template #actions>
        <Button variant="outline" title="只重新读取已保存的问题结果，不会访问 URL" :disabled="loading || scanning || syncing || retiringStaleRoutes" @click="refreshIssueQueueList">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新列表
        </Button>
        <Button variant="outline" title="从当前站点清单更新 URL 台账" :disabled="loading || scanning || syncing || retiringStaleRoutes || !canEdit" @click="synchronizeStorefrontUrlCatalog">
 <RefreshCw :class="['size-4', syncing ? 'animate-spin': '']" />
          同步 URL 台账
        </Button>
        <Button variant="outline" title="先同步最新 URL 来源，再实际访问全部可检查的问题 URL" :disabled="loading || scanning || syncing || retiringStaleRoutes || !canEdit" @click="scanCurrentIssueQueueUrls">
 <ScanSearch :class="['size-4', scanning ? 'animate-spin': '']" />
          更新来源并扫描全部问题 URL
        </Button>
        <Button v-if="canEdit && issueStats.stale_route" variant="ghost" :disabled="loading || scanning || syncing || retiringStaleRoutes" @click="retireAllActiveStaleRouteIssues">
 <Trash2 :class="['size-4', retiringStaleRoutes ? 'animate-pulse': '']" />
          清理 {{ issueStats.stale_route }} 条失效路径
        </Button>
      </template>
    </AdminPageHeader>

    <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
      <span v-if="syncing">正在更新 URL 来源清单…</span>
      <span v-else-if="scanning">正在重新检查问题 URL：已检查 {{ urlOperationStore.checked }} / {{ urlOperationStore.eligible }} 条</span>
      <span v-else-if="lastListRefreshAt">列表读取于 {{ formatRouteCatalogDate(lastListRefreshAt) }}</span>
      <span>“最近发现”是问题被检测到的时间，不是列表刷新时间。</span>
    </div>

    <AdminStatsGrid :items="statItems" />

 <div class="flex flex-wrap items-center gap-2 border-y border-dashed border-border/70 py-3">
      <Select v-model="stateFilter">
 <SelectTrigger class="w-36"><SelectValue /></SelectTrigger>
        <SelectContent>
          <SelectItem value="active">待处理</SelectItem>
          <SelectItem value="open">未认领</SelectItem>
          <SelectItem value="acknowledged">处理中</SelectItem>
          <SelectItem value="resolved">待验证</SelectItem>
          <SelectItem value="verified">已验证</SelectItem>
          <SelectItem value="suppressed">已抑制</SelectItem>
          <SelectItem value="all">全部状态</SelectItem>
        </SelectContent>
      </Select>
      <Select v-model="severityFilter">
 <SelectTrigger class="w-32"><SelectValue placeholder="全部等级" /></SelectTrigger>
        <SelectContent>
          <SelectItem value="all">全部等级</SelectItem>
          <SelectItem value="critical">严重</SelectItem>
          <SelectItem value="high">高</SelectItem>
          <SelectItem value="medium">中</SelectItem>
          <SelectItem value="low">低</SelectItem>
        </SelectContent>
      </Select>
      <Button variant="outline" size="sm" :disabled="loading" @click="applyFilters">
 <Filter class="size-3.5" />
        筛选
      </Button>
      <Button variant="ghost" size="sm" :disabled="loading" @click="resetFilters">
        重置筛选
      </Button>
 <span class="ml-auto text-xs text-muted-foreground">共 {{ pagination.total }} 项</span>
    </div>

    <AdminTablePanel :loading="loading">
 <Table class="min-w-[1080px]">
        <TableHeader>
          <TableRow>
 <TableHead class="w-48">问题</TableHead>
 <TableHead class="w-[300px]">路径 / 来源</TableHead>
 <TableHead class="w-24">等级</TableHead>
 <TableHead class="w-28">状态</TableHead>
            <TableHead>处置路径</TableHead>
 <TableHead class="w-40">最近发现</TableHead>
 <TableHead class="w-16 text-right">详情</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="issues.length === 0">
 <TableCell colspan="7" class="h-40 text-center text-sm text-muted-foreground">
              {{ loading ? '正在加载 URL 问题' : '当前筛选下没有 URL 问题' }}
            </TableCell>
          </TableRow>
          <TableRow
            v-for="issue in issues"
            :key="issue.id"
 class="cursor-pointer"
            @click="openDetail(issue.id)"
          >
            <TableCell>
 <p class="font-bold">{{ issueLabel(issue.issue_type) }}</p>
 <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ issue.issue_type }}</p>
            </TableCell>
            <TableCell>
 <p class="truncate font-mono text-xs">{{ issue.route_entry?.path || '-'}}</p>
 <p class="mt-1 truncate text-[10px] text-muted-foreground">
                {{ issue.route_entry?.source_type || '未知来源' }} · {{ issue.route_entry?.locale || '-' }}
              </p>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="severityTone(issue.severity)">
                {{ severityLabel(issue.severity) }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="stateTone(issue.state)">
                {{ stateLabel(issue.state) }}
              </AdminStatusBadge>
 <p v-if="issue.assignee_id" class="mt-1 font-mono text-[10px] text-muted-foreground">#{{ issue.assignee_id }}</p>
            </TableCell>
 <TableCell class="max-w-72">
 <p class="truncate text-xs">{{ prescribedAction(issue.issue_type) }}</p>
 <p v-if="issue.resolution_note" class="mt-1 truncate text-[10px] text-muted-foreground" :title="issue.resolution_note">
                {{ issue.resolution_note }}
              </p>
            </TableCell>
 <TableCell class="text-xs text-muted-foreground">
              {{ formatRouteCatalogDate(issue.last_detected_at) }}
            </TableCell>
 <TableCell class="text-right">
              <Button
                variant="ghost"
                size="icon"
                title="查看问题详情"
                aria-label="查看问题详情"
                @click.stop="openDetail(issue.id)"
              >
 <Eye class="size-4" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="detailOpen">
 <DialogContent size="xl" class="max-h-[calc(100dvh-1rem)]">
        <DialogHeader>
 <div class="flex min-w-0 items-start justify-between gap-3 pr-8">
 <div class="min-w-0">
              <DialogTitle>{{ selectedIssue ? issueLabel(selectedIssue.issue_type) : 'URL 问题' }}</DialogTitle>
 <DialogDescription class="mt-1 truncate font-mono text-[10px]">
                {{ selectedIssue?.route_entry?.path || '正在读取问题详情' }}
              </DialogDescription>
            </div>
 <div v-if="selectedIssue" class="flex shrink-0 gap-1">
              <AdminStatusBadge :tone="severityTone(selectedIssue.severity)">
                {{ severityLabel(selectedIssue.severity) }}
              </AdminStatusBadge>
              <AdminStatusBadge :tone="stateTone(selectedIssue.state)">
                {{ stateLabel(selectedIssue.state) }}
              </AdminStatusBadge>
            </div>
          </div>
        </DialogHeader>

 <div v-if="detailLoading" class="flex min-h-56 items-center justify-center text-sm text-muted-foreground">
          正在加载问题详情
        </div>

 <div v-else-if="selectedIssue" class="space-y-4 overflow-y-auto pr-1">
 <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">来源</p>
 <p class="mt-1 text-sm font-bold">{{ selectedIssue.route_entry?.source_type || '-'}}</p>
 <p class="mt-1 font-mono text-[10px] text-muted-foreground">
                {{ selectedIssue.route_entry?.source_key || '无来源键' }}
              </p>
            </div>
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">证据</p>
 <p class="mt-1 text-sm font-bold">
                {{ checkLabel(selectedIssue.route_entry?.last_check_status, selectedIssue.route_entry?.last_http_status) }}
              </p>
 <p class="mt-1 font-mono text-[10px] text-muted-foreground">
                {{ selectedIssue.latest_check_result_id ? `检查 #${selectedIssue.latest_check_result_id}` : '快照证据' }}
              </p>
            </div>
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">责任人</p>
 <p class="mt-1 text-sm font-bold">{{ selectedIssue.assignee_id ? `用户 #${selectedIssue.assignee_id}` : '未认领'}}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">{{ formatRouteCatalogDate(selectedIssue.last_detected_at) }}</p>
            </div>
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">关联跳转</p>
 <p class="mt-1 text-sm font-bold">{{ selectedIssue.linked_redirect_rule_id ? `规则 #${selectedIssue.linked_redirect_rule_id}` : '未关联'}}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">{{ prescribedAction(selectedIssue.issue_type) }}</p>
            </div>
          </div>

 <div class="flex flex-wrap gap-2 border-y border-dashed border-border/70 py-3">
            <Button
              v-if="selectedIssue.state === 'open'"
              variant="outline"
              size="sm"
              :disabled="!canEdit || actionKey !== null"
              @click="acknowledge"
            >
 <Check class="size-3.5" />
              确认处理
            </Button>
            <Button
              v-if="selectedIssue.state === 'open' || selectedIssue.state === 'acknowledged'"
              variant="outline"
              size="sm"
              :disabled="!canEdit || actionKey !== null"
              @click="claim"
            >
 <UserCheck class="size-3.5" />
              由我认领
            </Button>
            <Button
              v-if="canCreateRedirect"
              variant="outline"
              size="sm"
              :disabled="!canEdit || actionKey !== null"
              @click="openRedirectCreate"
            >
 <GitBranch class="size-3.5" />
              建立重定向
            </Button>
            <Button
              v-if="selectedIssue.issue_type === 'path_collision' || selectedIssue.route_entry?.is_checkable"
              variant="outline"
              size="sm"
              :disabled="!canEdit || actionKey !== null || scanning || syncing"
              @click="recheck"
            >
 <RefreshCw :class="['size-3.5', actionKey === 'recheck'? 'animate-spin': '']" />
              {{ selectedIssue.issue_type === 'path_collision' ? '同步来源并复核' : '重新检查' }}
            </Button>
            <span v-else-if="selectedIssue.route_entry" class="self-center text-xs text-muted-foreground">
              该路径在 URL 台账中标记为不可检查。
            </span>
            <Button
              v-if="selectedIssue.state === 'resolved'"
              size="sm"
              :disabled="!canEdit || actionKey !== null || scanning || syncing"
              @click="verify"
            >
 <BadgeCheck :class="['size-3.5', actionKey === 'verify'? 'animate-spin': '']" />
              验证关闭
            </Button>
          </div>

 <div v-if="isActionable" class="grid gap-4 xl:grid-cols-2">
 <form class="space-y-2 border border-dashed border-border/80 p-3" @submit.prevent="addComment">
 <div class="flex items-center justify-between gap-3">
 <p class="text-xs font-bold">处理备注</p>
                <Button type="submit" size="sm" variant="outline" :disabled="!canEdit || actionKey !== null || !commentNote.trim()">
 <MessageSquarePlus class="size-3.5" />
                  记录
                </Button>
              </div>
              <Textarea v-model="commentNote" placeholder="记录判断、外部工单或接手信息" />
            </form>

 <form class="space-y-2 border border-dashed border-border/80 p-3" @submit.prevent="resolve">
 <div class="flex items-center justify-between gap-3">
 <p class="text-xs font-bold">处理完成</p>
                <Button type="submit" size="sm" :disabled="!canEdit || actionKey !== null || !resolutionNote.trim()">
 <CheckCheck class="size-3.5" />
                  记录解决
                </Button>
              </div>
              <Select v-model="resolutionType">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="redirect_published">重定向已发布</SelectItem>
                  <SelectItem value="source_restored">来源已恢复</SelectItem>
                  <SelectItem value="source_path_changed">来源路径已调整</SelectItem>
                  <SelectItem value="canonical_fixed">Canonical 已修正</SelectItem>
                  <SelectItem value="runtime_fixed">运行环境已修复</SelectItem>
                  <SelectItem v-if="selectedIssue?.issue_type === 'stale_route'" value="retired">页面已正式退役</SelectItem>
                  <SelectItem value="not_applicable">确认无需处理</SelectItem>
                </SelectContent>
              </Select>
              <Textarea v-model="resolutionNote" placeholder="说明已完成的修复与可验证结果" />
            </form>
          </div>

 <form v-if="isActionable" class="flex flex-col gap-2 border border-dashed border-border/80 p-3 sm:flex-row sm:items-end" @submit.prevent="suppress">
 <label class="min-w-0 flex-1 space-y-1">
 <span class="text-xs font-bold">抑制原因</span>
              <Input v-model="suppressionReason" autocomplete="off" placeholder="说明暂不处理的原因" />
            </label>
 <label class="space-y-1">
 <span class="text-xs font-bold">复审时间</span>
              <Input v-model="suppressedUntil" type="datetime-local" />
            </label>
            <Button type="submit" variant="outline" :disabled="!canEdit || actionKey !== null || !suppressionReason.trim() || !suppressedUntil">
 <BellOff class="size-3.5" />
              暂时抑制
            </Button>
          </form>

 <section class="border border-dashed border-border/80">
 <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/70 px-3 py-3">
              <div>
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">TIMELINE</p>
 <h3 class="mt-1 text-sm font-black">处理时间线</h3>
              </div>
 <span class="font-mono text-[10px] text-muted-foreground">共 {{ eventsPagination.total }} 项</span>
            </div>
  <div class="max-h-64 overflow-auto">
  <div v-if="eventsLoading" class="px-3 py-8 text-center text-xs text-muted-foreground">
                正在加载处理事件
              </div>
  <div v-else-if="events.length === 0" class="px-3 py-8 text-center text-xs text-muted-foreground">
                暂无处理事件
              </div>
 <div v-for="event in events" :key="event.id" class="border-b border-dashed border-border/60 px-3 py-3 last:border-b-0">
 <div class="flex items-center justify-between gap-3">
 <p class="text-xs font-bold">{{ eventLabel(event.event_type) }}</p>
 <p class="whitespace-nowrap text-[10px] text-muted-foreground">{{ formatRouteCatalogDate(event.created_at) }}</p>
                </div>
 <p v-if="event.note" class="mt-1 whitespace-pre-wrap text-xs text-muted-foreground">{{ event.note }}</p>
 <p v-if="event.actor_user_id" class="mt-1 font-mono text-[10px] text-muted-foreground">用户 #{{ event.actor_user_id }}</p>
              </div>
            </div>
            <div v-if="eventsPagination.total > eventsPagination.page_size" class="border-t border-dashed border-border/70 px-3 py-2">
              <AdminPagination
                :page="eventsPagination.page"
                :page-size="eventsPagination.page_size"
                :total="eventsPagination.total"
                :page-sizes="[20, 50, 100]"
                @update:page="updateEventsPage"
                @update:page-size="updateEventsPageSize"
              />
            </div>
          </section>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="detailOpen = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  BadgeCheck,
  BellOff,
  Check,
  CheckCheck,
  Eye,
  Filter,
  GitBranch,
  MessageSquarePlus,
  RefreshCw,
  ScanSearch,
  ShieldAlert,
  Trash2,
  UserCheck,
} from '@lucide/vue'
import { toast } from 'vue-sonner'
import { useRouter } from 'vue-router'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { formatRouteCatalogDate, checkLabel } from '@/modules/url-management/routeCatalogPresentation'
import { storefrontRouteCatalogApi } from '@/modules/url-management/routeCatalog'
import {
  storefrontURLIssuesApi,
  type StorefrontURLIssue,
  type StorefrontURLIssueEvent,
  type StorefrontURLIssueSeverity,
  type StorefrontURLIssueState,
  type StorefrontURLIssueStateFilter,
  type StorefrontURLIssueStats,
} from '@/modules/url-management/urlIssues'
import type { SEOResourcePagination } from '@/modules/seo/types'
import { useAuthStore } from '@/stores/auth'
import { useURLOperationStore } from '@/stores/urlOperation'

const authStore = useAuthStore()
const urlOperationStore = useURLOperationStore()
const router = useRouter()
const canEdit = authStore.hasPermission('url:edit')
const loading = ref(false)
const eventsLoading = ref(false)
const syncing = ref(false)
const retiringStaleRoutes = ref(false)
const lastListRefreshAt = ref<string | null>(null)
const detailLoading = ref(false)
const actionKey = ref<string | null>(null)
const issues = ref<StorefrontURLIssue[]>([])
const events = ref<StorefrontURLIssueEvent[]>([])
const selectedIssue = ref<StorefrontURLIssue | null>(null)
const detailOpen = ref(false)
const stateFilter = ref<StorefrontURLIssueStateFilter>('active')
const severityFilter = ref<StorefrontURLIssueSeverity | 'all'>('all')
const pagination = ref<SEOResourcePagination>({ page: 1, page_size: 50, total: 0, total_pages: 0 })
const eventsPagination = ref<SEOResourcePagination>({ page: 1, page_size: 50, total: 0, total_pages: 0 })
const commentNote = ref('')
const resolutionType = ref('runtime_fixed')
const resolutionNote = ref('')
const suppressionReason = ref('')
const suppressedUntil = ref('')
const issueStats = ref<StorefrontURLIssueStats>({
  active: 0,
  open: 0,
  acknowledged: 0,
  resolved: 0,
  verified: 0,
  suppressed: 0,
  critical: 0,
  high: 0,
  stale_route: 0,
})

const scanning = computed(() => urlOperationStore.running)

let issueListRequestSequence = 0
let issueSummaryRequestSequence = 0
let detailRequestSequence = 0
let eventRequestSequence = 0

const requestErrorMessage = (error: unknown): string => {
  if (error && typeof error === 'object') {
    const responseData = (error as { response?: { data?: unknown } }).response?.data
    if (responseData && typeof responseData === 'object') {
      const message = (responseData as Record<string, unknown>).error
      if (typeof message === 'string' && message.trim()) return message.trim()
    }
  }
  return error instanceof Error ? error.message : ''
}

const issueLabel = (value: string): string => ({
  redirect_chain: '重定向链',
  redirect_target_mismatch: '跳转目标不一致',
  redirect_status_mismatch: '规范路径发生跳转',
  not_found: '页面不存在',
  server_error: '服务端错误',
  canonical_mismatch: 'Canonical 不一致',
  path_collision: '路径冲突',
  stale_route: '失效路径',
  check_error: '检查失败',
}[value] || value || '-')

const prescribedAction = (value: string): string => ({
  redirect_chain: '改为一次直达目标路径后重新检查',
  redirect_target_mismatch: '校正跳转规则或静态别名',
  redirect_status_mismatch: '恢复规范路径的直接 200 响应',
  not_found: '恢复来源，或建立到替代页面的跳转',
  server_error: '修复运行环境后重新检查',
  canonical_mismatch: '在来源域修正 Canonical 输出',
  path_collision: '保留唯一来源并调整或退役冲突来源',
  stale_route: '恢复来源，建立跳转，或正式退役',
  check_error: '确认网络或运行错误后重新检查',
}[value] || '确认来源与实际响应后处理')

const severityLabel = (value: StorefrontURLIssueSeverity): string => ({
  critical: '严重',
  high: '高',
  medium: '中',
  low: '低',
}[value] || value)

const severityTone = (value: StorefrontURLIssueSeverity): AdminStatusTone => ({
  critical: 'coral',
  high: 'coral',
  medium: 'amber',
  low: 'gray',
} as const)[value] || 'gray'

const stateLabel = (value: StorefrontURLIssueState): string => ({
  open: '未认领',
  acknowledged: '处理中',
  resolved: '待验证',
  verified: '已验证',
  suppressed: '已抑制',
}[value] || value)

const stateTone = (value: StorefrontURLIssueState): AdminStatusTone => ({
  open: 'coral',
  acknowledged: 'amber',
  resolved: 'blue',
  verified: 'green',
  suppressed: 'gray',
} as const)[value] || 'gray'

const eventLabel = (value: string): string => ({
  detected: '检测到问题',
  reopened: '问题重新出现',
  acknowledged: '确认处理',
  claimed: '已认领',
  commented: '新增备注',
  redirect_linked: '关联重定向',
  resolution_recorded: '记录解决方案',
  suppressed: '暂时抑制',
  snapshot_invalidated: '快照使旧观测失效',
  verification_passed: '验证通过',
  verification_failed: '验证未通过',
  stale_route_retired: '失效路径已清理',
}[value] || value)

const isActionable = computed(() => (
  selectedIssue.value?.state === 'open' || selectedIssue.value?.state === 'acknowledged'
))

const canCreateRedirect = computed(() => (
  selectedIssue.value?.issue_type === 'stale_route'
))

const statItems = computed(() => [
  { key: 'active', label: '未关闭工单', value: issueStats.value.active, icon: RefreshCw, tone: issueStats.value.active ? 'amber' : 'gray' },
  { key: 'open', label: '未认领', value: issueStats.value.open, icon: UserCheck, tone: issueStats.value.open ? 'coral' : 'gray' },
  { key: 'acknowledged', label: '处理中', value: issueStats.value.acknowledged, icon: RefreshCw, tone: issueStats.value.acknowledged ? 'blue' : 'gray' },
  { key: 'resolved', label: '待验证', value: issueStats.value.resolved, icon: BadgeCheck, tone: issueStats.value.resolved ? 'blue' : 'gray' },
  { key: 'critical', label: '严重等级', value: issueStats.value.critical, icon: ShieldAlert, tone: issueStats.value.critical ? 'coral' : 'gray' },
])

const loadSummary = async (): Promise<boolean> => {
  const requestSequence = ++issueSummaryRequestSequence
  try {
    const nextStats = await storefrontURLIssuesApi.summary()
    if (requestSequence !== issueSummaryRequestSequence) return false
    issueStats.value = { ...issueStats.value, ...nextStats }
    return true
  } catch (error) {
    console.error('Failed to load storefront URL issue summary:', error)
    if (requestSequence === issueSummaryRequestSequence) {
      toast.error('URL 问题统计加载失败')
    }
    return false
  }
}

const load = async (): Promise<boolean> => {
  const requestSequence = ++issueListRequestSequence
  loading.value = true
  try {
    const response = await storefrontURLIssuesApi.list({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      state: stateFilter.value,
      ...(severityFilter.value !== 'all' ? { severity: severityFilter.value } : {}),
    })
    if (requestSequence !== issueListRequestSequence) return false
    issues.value = response.items
    pagination.value = response.pagination
    return true
  } catch (error) {
    console.error('Failed to load storefront URL issues:', error)
    if (requestSequence === issueListRequestSequence) toast.error('URL 问题队列加载失败')
    return false
  } finally {
    if (requestSequence === issueListRequestSequence) loading.value = false
  }
}

const refreshIssueQueueList = async (): Promise<void> => {
  const [listLoaded] = await Promise.all([load(), loadSummary()])
  if (listLoaded) lastListRefreshAt.value = new Date().toISOString()
}

const scanCurrentIssueQueueUrls = async (): Promise<void> => {
  if (!canEdit || scanning.value || syncing.value || retiringStaleRoutes.value) return
  if (!await synchronizeStorefrontUrlCatalog(false)) return
  const completed = await urlOperationStore.run({
    needs_attention: true,
    include_stale_when_checking: true,
    include_aliases: true,
    sync_latest: false,
    batch_size: 200,
  })
  if (!completed) return
  await refreshIssueQueueList()
}

const synchronizeStorefrontUrlCatalog = async (showSuccessMessage = true): Promise<boolean> => {
  if (!canEdit || syncing.value || scanning.value || retiringStaleRoutes.value) return false
  syncing.value = true
  try {
    const summary = await storefrontRouteCatalogApi.sync()
    await refreshIssueQueueList()
    if (showSuccessMessage) {
      toast.success(`URL 台账已同步：${summary.entries || 0} 条，重复 ${summary.duplicates || 0} 条`)
    }
    return true
  } catch (error) {
    console.error('Failed to sync storefront route catalog:', error)
    const detail = requestErrorMessage(error)
    toast.error(detail ? `URL 台账同步失败：${detail}` : 'URL 台账同步失败，请检查站点清单')
    return false
  } finally {
    syncing.value = false
  }
}

const retireAllActiveStaleRouteIssues = async (): Promise<void> => {
  const staleRouteCount = Number(issueStats.value.stale_route || 0)
  if (!canEdit || staleRouteCount < 1 || retiringStaleRoutes.value) return
  if (!window.confirm(`确认清理 ${staleRouteCount} 条失效路径吗？系统会保留历史记录，并将它们标记为“已验证 / 已退役”。`)) return
  retiringStaleRoutes.value = true
  try {
    const retiredCount = await storefrontURLIssuesApi.retireAllActiveStaleRouteIssues()
    detailOpen.value = false
    await refreshIssueQueueList()
    toast.success(retiredCount ? `已清理 ${retiredCount} 条失效路径，历史记录已保留` : '没有需要清理的失效路径')
  } catch (error) {
    console.error('Failed to retire stale storefront URL issues:', error)
    toast.error('失效路径清理失败')
  } finally {
    retiringStaleRoutes.value = false
  }
}

const loadEvents = async (issueID: number, detailSequence = detailRequestSequence): Promise<boolean> => {
  const requestSequence = ++eventRequestSequence
  const page = eventsPagination.value.page
  const pageSize = eventsPagination.value.page_size
  eventsLoading.value = true
  try {
    const response = await storefrontURLIssuesApi.events(issueID, {
      page,
      page_size: pageSize,
    })
    if (detailSequence !== detailRequestSequence || requestSequence !== eventRequestSequence) return false
    events.value = response.items
    eventsPagination.value = response.pagination
    return true
  } catch (error) {
    console.error('Failed to load storefront URL issue events:', error)
    if (detailSequence === detailRequestSequence && requestSequence === eventRequestSequence) {
      toast.error('URL 问题时间线加载失败')
    }
    return false
  } finally {
    if (requestSequence === eventRequestSequence) eventsLoading.value = false
  }
}

const openDetail = async (issueID: number): Promise<void> => {
  const requestSequence = ++detailRequestSequence
  eventRequestSequence += 1
  detailOpen.value = true
  detailLoading.value = true
  selectedIssue.value = null
  events.value = []
  eventsPagination.value = { page: 1, page_size: 50, total: 0, total_pages: 0 }
  eventsPagination.value.page = 1
  try {
    const [issue] = await Promise.all([
      storefrontURLIssuesApi.get(issueID),
      loadEvents(issueID, requestSequence),
    ])
    if (requestSequence !== detailRequestSequence) return
    selectedIssue.value = issue
    commentNote.value = ''
    resolutionType.value = issue.resolution_type || 'runtime_fixed'
    resolutionNote.value = ''
    suppressionReason.value = ''
    suppressedUntil.value = ''
  } catch (error) {
    console.error('Failed to load storefront URL issue detail:', error)
    if (requestSequence === detailRequestSequence) {
      toast.error('URL 问题详情加载失败')
      detailOpen.value = false
    }
  } finally {
    if (requestSequence === detailRequestSequence) detailLoading.value = false
  }
}

watch(detailOpen, (open) => {
  if (!open) {
    detailRequestSequence += 1
    eventRequestSequence += 1
  }
})

const refreshSelectedIssue = async (
  issue: StorefrontURLIssue,
  expectedDetailSequence = detailRequestSequence,
): Promise<void> => {
  if (expectedDetailSequence !== detailRequestSequence || selectedIssue.value?.id !== issue.id) {
    await refreshIssueQueueList()
    return
  }
  if (issue.state === 'verified') {
    detailOpen.value = false
    selectedIssue.value = null
    await Promise.all([load(), loadSummary()])
    return
  }
  selectedIssue.value = issue
  await Promise.all([load(), loadEvents(issue.id, expectedDetailSequence), loadSummary()])
}

const withAction = async (
  key: string,
  run: (issueID: number) => Promise<StorefrontURLIssue>,
  successMessage: string,
): Promise<boolean> => {
  if (!selectedIssue.value || !canEdit || actionKey.value) return false
  const issueID = selectedIssue.value.id
  const detailSequence = detailRequestSequence
  actionKey.value = key
  try {
    const issue = await run(issueID)
    await refreshSelectedIssue(issue, detailSequence)
    toast.success(successMessage)
    return true
  } catch (error) {
    console.error(`Failed to ${key} storefront URL issue:`, error)
    const detail = requestErrorMessage(error)
    toast.error(detail ? `URL 问题操作失败：${detail}` : 'URL 问题操作失败')
    return false
  } finally {
    actionKey.value = null
  }
}

const acknowledge = async (): Promise<void> => {
  await withAction('acknowledge', (issueID) => storefrontURLIssuesApi.acknowledge(issueID), '已确认处理')
}

const claim = async (): Promise<void> => {
  await withAction('claim', (issueID) => storefrontURLIssuesApi.claim(issueID), '已由你认领')
}

const addComment = async (): Promise<void> => {
  const note = commentNote.value.trim()
  if (!note) return
  if (await withAction('comment', (issueID) => storefrontURLIssuesApi.comment(issueID, note), '处理备注已记录')) {
    commentNote.value = ''
  }
}

const resolve = async (): Promise<void> => {
  const note = resolutionNote.value.trim()
  if (!note) return
  const resolutionTypeSnapshot = resolutionType.value
  const linkedRedirectRuleID = selectedIssue.value?.linked_redirect_rule_id
  const retiresStaleRoute = selectedIssue.value?.issue_type === 'stale_route' && resolutionType.value === 'retired'
  if (await withAction('resolve', (issueID) => storefrontURLIssuesApi.resolve(issueID, {
    resolution_type: resolutionTypeSnapshot,
    resolution_note: note,
    ...(linkedRedirectRuleID
      ? { linked_redirect_rule_id: linkedRedirectRuleID }
      : {}),
  }), retiresStaleRoute ? '失效路径已退役并关闭' : '已记录解决方案，等待验证')) {
    resolutionNote.value = ''
  }
}

const suppress = async (): Promise<void> => {
  const reason = suppressionReason.value.trim()
  if (!reason || !suppressedUntil.value) return
  const parsedSuppressedUntil = new Date(suppressedUntil.value)
  if (Number.isNaN(parsedSuppressedUntil.getTime())) {
    toast.error('复审时间格式无效')
    return
  }
  const suppressedUntilISO = parsedSuppressedUntil.toISOString()
  if (await withAction('suppress', (issueID) => storefrontURLIssuesApi.suppress(issueID, {
    reason,
    suppressed_until: suppressedUntilISO,
  }), '该问题已抑制，届时将重新复审')) {
    suppressionReason.value = ''
    suppressedUntil.value = ''
  }
}

const recheck = async (): Promise<void> => {
  if (!selectedIssue.value || !canEdit || actionKey.value || scanning.value || syncing.value) return
  const issueID = selectedIssue.value.id
  const detailSequence = detailRequestSequence
  actionKey.value = 'recheck'
  try {
    const result = await storefrontURLIssuesApi.recheck(issueID)
    await refreshSelectedIssue(result.issue, detailSequence)
    toast.success(result.issue.state === 'verified'
      ? '已确认恢复，问题已自动关闭并从待处理列表移除'
      : '复检完成，问题仍未恢复')
  } catch (error) {
    console.error('Failed to recheck storefront URL issue:', error)
    const detail = requestErrorMessage(error)
    toast.error(detail ? `URL 重新检查失败：${detail}` : 'URL 重新检查失败')
  } finally {
    actionKey.value = null
  }
}

const verify = async (): Promise<void> => {
  if (!selectedIssue.value || !canEdit || actionKey.value || scanning.value || syncing.value) return
  const issueID = selectedIssue.value.id
  const detailSequence = detailRequestSequence
  actionKey.value = 'verify'
  try {
    const result = await storefrontURLIssuesApi.verify(issueID)
    await refreshSelectedIssue(result.issue, detailSequence)
    toast.success(result.issue.state === 'verified' ? 'URL 问题已验证关闭' : '验证未通过，问题已重新打开')
  } catch (error) {
    console.error('Failed to verify storefront URL issue:', error)
    const detail = requestErrorMessage(error)
    toast.error(detail ? `URL 问题验证失败：${detail}` : 'URL 问题验证失败')
  } finally {
    actionKey.value = null
  }
}

const openRedirectCreate = (): void => {
  if (!selectedIssue.value?.route_entry?.path) return
  detailOpen.value = false
  void router.push({
    name: 'URLRedirects',
    query: {
      source_path: selectedIssue.value.route_entry.path,
      issue_id: String(selectedIssue.value.id),
    },
  })
}

const applyFilters = (): void => {
  pagination.value.page = 1
  void load()
}

const resetFilters = (): void => {
  stateFilter.value = 'active'
  severityFilter.value = 'all'
  pagination.value.page = 1
  void load()
}

const updatePage = (page: number): void => {
  pagination.value.page = page
  void load()
}

const updatePageSize = (pageSize: number): void => {
  pagination.value.page_size = pageSize
  pagination.value.page = 1
  void load()
}

const updateEventsPage = (page: number): void => {
  if (!selectedIssue.value) return
  eventsPagination.value.page = page
  void loadEvents(selectedIssue.value.id)
}

const updateEventsPageSize = (pageSize: number): void => {
  if (!selectedIssue.value) return
  eventsPagination.value.page_size = pageSize
  eventsPagination.value.page = 1
  void loadEvents(selectedIssue.value.id)
}

onMounted(() => {
  void refreshIssueQueueList()
})
</script>
