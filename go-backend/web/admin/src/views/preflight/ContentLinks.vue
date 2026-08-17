<template>
  <div class="space-y-4">
    <AdminPageHeader title="上线前检查 / 内容链接" description="检测泛化链接文案并修正可定位内容源">
      <template #actions>
        <Button
          size="icon"
          variant="outline"
          title="刷新内容链接检查"
          :disabled="loading || targetOptionsLoading"
          @click="refresh"
        >
          <RefreshCw :class="['size-4', { 'animate-spin': loading || targetOptionsLoading }]" />
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
      <div v-for="item in summaryItems" :key="item.label" class="border bg-card px-4 py-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">{{ item.label }}</p>
        <p class="mt-2 text-2xl font-black" :class="item.tone">{{ item.value }}</p>
      </div>
    </section>

    <section class="border bg-card p-4">
      <div class="grid gap-3 xl:grid-cols-[minmax(0,1fr)_auto] xl:items-end">
        <label class="grid gap-1.5">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">检测页面</span>
          <Select v-model="targetURL" :disabled="running || targetOptionsLoading">
            <SelectTrigger class="h-10 w-full min-w-0">
              <SelectValue placeholder="选择页面" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="option in targetOptions" :key="option.url" :value="option.url">
                {{ targetOptionLabel(option) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </label>
        <Button :disabled="running || targetOptionsLoading || !canManage || !targetURL" @click="runCheck">
          <LoaderCircle v-if="running" class="size-4 animate-spin" />
          <Play v-else class="size-4" />
          运行检测
        </Button>
      </div>
    </section>

    <section class="border bg-card p-4">
      <div class="grid gap-3 lg:grid-cols-[11rem_11rem_minmax(0,1fr)_auto] lg:items-end">
        <label class="grid gap-1.5">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">状态</span>
          <Select v-model="stateFilter">
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="active">待处理</SelectItem>
              <SelectItem value="open">未处理</SelectItem>
              <SelectItem value="resolved">待复检</SelectItem>
              <SelectItem value="verified">已验证</SelectItem>
              <SelectItem value="ignored">已忽略</SelectItem>
              <SelectItem value="all">全部</SelectItem>
            </SelectContent>
          </Select>
        </label>
        <label class="grid gap-1.5">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">修正</span>
          <Select v-model="fixableFilter">
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="fixable">可直接修正</SelectItem>
              <SelectItem value="source">需编辑来源</SelectItem>
            </SelectContent>
          </Select>
        </label>
        <label class="grid gap-1.5">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">搜索</span>
          <Input v-model="search" placeholder="页面、目标或文案" @keyup.enter="applyFilters" />
        </label>
        <Button variant="outline" :disabled="loading" @click="applyFilters">
          <Search class="size-4" />
          筛选
        </Button>
      </div>
    </section>

    <section class="overflow-hidden border bg-card">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[1120px] text-sm">
          <thead class="border-b bg-muted/30 text-left text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            <tr>
              <th class="px-4 py-3">页面</th>
              <th class="px-4 py-3">链接目标</th>
              <th class="px-4 py-3">当前文案</th>
              <th class="px-4 py-3">建议文案</th>
              <th class="px-4 py-3">来源</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3 text-right">处理</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            <tr v-if="loading">
              <td colspan="7" class="px-4 py-12 text-center text-sm text-muted-foreground">正在加载内容链接问题</td>
            </tr>
            <tr v-else-if="issues.length === 0">
              <td colspan="7" class="px-4 py-12 text-center text-sm text-muted-foreground">当前筛选没有问题</td>
            </tr>
            <tr v-for="issue in issues" :key="issue.id">
              <td class="px-4 py-3">
                <div class="max-w-72">
                  <p class="truncate font-mono text-xs" :title="issue.target_url">{{ pathLabel(issue.target_url) }}</p>
                  <p class="mt-1 text-[10px] text-muted-foreground">{{ formatDate(issue.last_detected_at) }}</p>
                </div>
              </td>
              <td class="px-4 py-3">
                <button
                  type="button"
                  class="flex max-w-72 items-center gap-1 truncate font-mono text-xs text-primary"
                  :title="issue.link_url"
                  @click="openExternal(issue.link_url)"
                >
                  <ExternalLink class="size-3 shrink-0" />
                  <span class="truncate">{{ pathLabel(issue.link_url) }}</span>
                </button>
                <p class="mt-1 truncate text-[10px] text-muted-foreground" :title="issue.selector">{{ issue.selector }}</p>
              </td>
              <td class="px-4 py-3">
                <p class="max-w-44 truncate font-bold" :title="issue.link_text">{{ issue.link_text }}</p>
              </td>
              <td class="px-4 py-3">
                <p class="max-w-64 truncate font-bold" :title="issue.suggested_text">{{ issue.suggested_text || '-' }}</p>
                <p v-if="issue.fix_error" class="mt-1 max-w-64 truncate text-[10px] text-rose-600" :title="issue.fix_error">{{ issue.fix_error }}</p>
              </td>
              <td class="px-4 py-3">
                <AdminStatusBadge :tone="fixStatusTone(issue.fix_status)">
                  {{ fixStatusLabel(issue.fix_status) }}
                </AdminStatusBadge>
                <p class="mt-1 max-w-48 truncate font-mono text-[10px] text-muted-foreground" :title="sourceLabel(issue)">
                  {{ sourceLabel(issue) }}
                </p>
              </td>
              <td class="px-4 py-3">
                <AdminStatusBadge :tone="stateTone(issue.state)">{{ stateLabel(issue.state) }}</AdminStatusBadge>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-2">
                  <Button
                    v-if="issue.fix_status === 'pending'"
                    size="sm"
                    :disabled="!canManage || actionID === issue.id"
                    @click="applySuggestion(issue.id)"
                  >
                    <LoaderCircle v-if="actionID === issue.id" class="size-3.5 animate-spin" />
                    <WandSparkles v-else class="size-3.5" />
                    修正
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    title="重新检测"
                    :disabled="!canManage || actionID === issue.id"
                    @click="recheckIssue(issue.id)"
                  >
                    <RefreshCw :class="['size-3.5', { 'animate-spin': actionID === issue.id }]" />
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="border-t px-4 py-3">
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          @update:page="changePage"
          @update:page-size="changePageSize"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ExternalLink, LoaderCircle, Play, RefreshCw, Search, WandSparkles } from '@lucide/vue'
import { toast } from 'vue-sonner'
import preflightApi, {
  type PreflightContentLinkFixStatus,
  type PreflightContentLinkIssue,
  type PreflightContentLinkIssueState,
  type PreflightContentLinkIssueStateFilter,
  type PreflightContentLinkStats,
  type PreflightContentLinkTargetOption,
} from '@/api/preflight'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('services:manage'))
const loading = ref(false)
const running = ref(false)
const targetOptionsLoading = ref(false)
const targetURL = ref('')
const targetOptions = ref<PreflightContentLinkTargetOption[]>([])
const issues = ref<PreflightContentLinkIssue[]>([])
const stats = ref<PreflightContentLinkStats>({
  active: 0,
  open: 0,
  resolved: 0,
  verified: 0,
  ignored: 0,
  fixable: 0,
  needs_source: 0,
  applied: 0,
})
const pagination = ref({ page: 1, page_size: 50, total: 0, total_pages: 1 })
const stateFilter = ref<PreflightContentLinkIssueStateFilter>('active')
const fixableFilter = ref<'all' | 'fixable' | 'source'>('all')
const search = ref('')
const actionID = ref<number | null>(null)

const summaryItems = computed(() => [
  { label: '待处理', value: stats.value.active, tone: stats.value.active ? 'text-amber-600' : 'text-emerald-600' },
  { label: '未处理', value: stats.value.open, tone: stats.value.open ? 'text-rose-600' : 'text-emerald-600' },
  { label: '可直接修正', value: stats.value.fixable, tone: stats.value.fixable ? 'text-emerald-600' : 'text-muted-foreground' },
  { label: '需编辑来源', value: stats.value.needs_source, tone: stats.value.needs_source ? 'text-amber-600' : 'text-emerald-600' },
  { label: '已验证', value: stats.value.verified, tone: 'text-emerald-600' },
])

const loadTargets = async (): Promise<void> => {
  targetOptionsLoading.value = true
  try {
    const result = await preflightApi.getContentLinkTargets()
    targetOptions.value = result.items
    if (!targetURL.value || !result.items.some((option) => option.url === targetURL.value)) {
      targetURL.value = result.default_url || result.items[0]?.url || ''
    }
  } catch (error: any) {
    toast.error(errorMessage(error, '内容链接检测目标加载失败'))
  } finally {
    targetOptionsLoading.value = false
  }
}

const loadIssues = async (page = pagination.value.page): Promise<void> => {
  loading.value = true
  try {
    const result = await preflightApi.getContentLinkIssues({
      page,
      pageSize: pagination.value.page_size,
      state: stateFilter.value,
      search: search.value.trim() || undefined,
      fixable: fixableFilter.value === 'all' ? undefined : fixableFilter.value === 'fixable',
    })
    issues.value = result.items
    stats.value = result.stats
    pagination.value = result.pagination
  } catch (error: any) {
    toast.error(errorMessage(error, '内容链接问题加载失败'))
  } finally {
    loading.value = false
  }
}

const refresh = async (): Promise<void> => {
  await loadTargets()
  await loadIssues(1)
}

const runCheck = async (): Promise<void> => {
  if (!targetURL.value || !canManage.value || running.value) return
  running.value = true
  try {
    const result = await preflightApi.runContentLinkCheck(targetURL.value)
    stats.value = result.stats
    if (result.run.status === 'failed') {
      toast.error(result.run.error_message || '内容链接检测失败')
    } else {
      toast.success(`检测完成，发现 ${result.run.issue_count} 个问题`)
    }
    await loadIssues(1)
  } catch (error: any) {
    toast.error(errorMessage(error, '内容链接检测失败'))
  } finally {
    running.value = false
  }
}

const applySuggestion = async (id: number): Promise<void> => {
  if (!canManage.value || actionID.value) return
  actionID.value = id
  try {
    await preflightApi.applyContentLinkSuggestion(id)
    toast.success('已修正来源文案，等待复检验证')
    await loadIssues()
  } catch (error: any) {
    toast.error(errorMessage(error, '内容链接修正失败'))
  } finally {
    actionID.value = null
  }
}

const recheckIssue = async (id: number): Promise<void> => {
  if (!canManage.value || actionID.value) return
  actionID.value = id
  try {
    const result = await preflightApi.recheckContentLinkIssue(id)
    if (result.run.status === 'failed') {
      toast.error(result.run.error_message || '重新检测失败')
    } else {
      toast.success('重新检测完成')
    }
    await loadIssues()
  } catch (error: any) {
    toast.error(errorMessage(error, '重新检测失败'))
  } finally {
    actionID.value = null
  }
}

const applyFilters = (): void => {
  pagination.value.page = 1
  void loadIssues(1)
}

const changePage = (page: number): void => {
  pagination.value.page = page
  void loadIssues(page)
}

const changePageSize = (pageSize: number): void => {
  pagination.value.page_size = pageSize
  pagination.value.page = 1
  void loadIssues(1)
}

const targetOptionLabel = (option: PreflightContentLinkTargetOption): string => {
  const parts = []
  if (option.is_home) parts.push('首页')
  if (option.title && (!option.is_home || option.title !== '首页')) parts.push(option.title)
  if (option.path) parts.push(option.path)
  if (option.locale) parts.push(option.locale)
  return parts.join(' · ') || option.url
}

const pathLabel = (value: string): string => {
  try {
    const parsed = new URL(value)
    return `${parsed.pathname}${parsed.search}`
  } catch {
    return value || '-'
  }
}

const sourceLabel = (issue: PreflightContentLinkIssue): string => {
  if (issue.source_type === 'blog_post') {
    return `${issue.source_key} · ${issue.source_field}`
  }
  return issue.source_key || 'storefront render'
}

const fixStatusLabel = (value: PreflightContentLinkFixStatus): string => ({
  not_fixable: '需编辑来源',
  pending: '可直接修正',
  applied: '已修正',
  failed: '修正失败',
}[value])

const fixStatusTone = (value: PreflightContentLinkFixStatus): AdminStatusTone => ({
  not_fixable: 'amber',
  pending: 'green',
  applied: 'blue',
  failed: 'coral',
} as const)[value]

const stateLabel = (value: PreflightContentLinkIssueState): string => ({
  open: '未处理',
  resolved: '待复检',
  verified: '已验证',
  ignored: '已忽略',
}[value])

const stateTone = (value: PreflightContentLinkIssueState): AdminStatusTone => ({
  open: 'coral',
  resolved: 'blue',
  verified: 'green',
  ignored: 'gray',
} as const)[value]

const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')

const openExternal = (value: string): void => {
  if (!value) return
  window.open(value, '_blank', 'noopener,noreferrer')
}

const errorMessage = (error: any, fallback: string): string => (
  error?.response?.data?.message || error?.response?.data?.error || fallback
)

onMounted(() => {
  void refresh()
})
</script>
