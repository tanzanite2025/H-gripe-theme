<template>
  <section class="border bg-card">
    <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
      <div>
        <h2 class="text-sm font-black">标题层级</h2>
        <p class="mt-1 text-xs text-muted-foreground">H1 / H2 / H3</p>
      </div>
      <div class="flex items-center gap-2">
        <select
          v-model="state"
          class="h-8 border bg-background px-2 text-xs"
          :disabled="loading"
          @change="$emit('change-filter', state)"
        >
          <option value="active">待处理</option>
          <option value="open">未确认</option>
          <option value="acknowledged">处理中</option>
          <option value="resolved">待验证</option>
          <option value="verified">已验证</option>
          <option value="all">全部</option>
        </select>
        <Button size="icon" variant="outline" title="刷新标题检查" :disabled="loading" @click="$emit('refresh')">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
        </Button>
      </div>
    </div>

    <div v-if="loading && findings.length === 0" class="p-8 text-center text-sm text-muted-foreground">
      正在读取标题检查
    </div>
    <div v-else-if="findings.length === 0" class="p-8 text-center text-sm text-muted-foreground">
      当前筛选下没有标题层级问题
    </div>
    <div v-else class="divide-y">
      <article
        v-for="finding in findings"
        :key="finding.id"
        class="grid gap-3 px-4 py-3 lg:grid-cols-[minmax(0,1fr)_auto_auto] lg:items-center"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <Heading2 class="size-4 text-muted-foreground" aria-hidden="true" />
            <p class="truncate text-sm font-black">{{ headingRuleLabel(finding.audit_id) }}</p>
            <AdminStatusBadge :tone="issueTone(finding.severity)">
              {{ issueSeverityLabel(finding.severity) }}
            </AdminStatusBadge>
            <AdminStatusBadge :tone="findingStateTone(finding.state)">
              {{ findingStateLabel(finding.state) }}
            </AdminStatusBadge>
          </div>
          <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground" :title="finding.target_url">
            {{ finding.target_url }}
          </p>
          <p v-if="finding.description" class="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">
            {{ finding.description }}
          </p>
        </div>
        <span class="text-xs text-muted-foreground">
          {{ strategyLabel(finding.strategy) }} · {{ formatDate(finding.last_detected_at) }}
        </span>
        <Button
          size="icon"
          variant="ghost"
          title="查看标题检查详情"
          :disabled="loading"
          @click="$emit('open', finding.id)"
        >
          <Eye class="size-4" />
        </Button>
      </article>
    </div>

    <div class="flex items-center justify-between border-t px-4 py-2">
      <span class="text-xs text-muted-foreground">共 {{ pagination.total }} 项</span>
      <div class="flex items-center gap-1">
        <Button
          size="icon"
          variant="outline"
          title="上一页"
          :disabled="loading || pagination.page <= 1"
          @click="$emit('change-page', pagination.page - 1)"
        >
          <ChevronLeft class="size-4" />
        </Button>
        <Button
          size="icon"
          variant="outline"
          title="下一页"
          :disabled="loading || pagination.page >= pagination.total_pages"
          @click="$emit('change-page', pagination.page + 1)"
        >
          <ChevronRight class="size-4" />
        </Button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, Eye, Heading2, RefreshCw } from '@lucide/vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import type {
  SiteQualityFinding,
  SiteQualityFindingState,
  SiteQualityFindingStateFilter,
  SiteQualityStrategy,
} from '@/api/preflight'

const props = defineProps<{
  loading: boolean
  findings: SiteQualityFinding[]
  state: SiteQualityFindingStateFilter
  pagination: { page: number; page_size: number; total: number; total_pages: number }
}>()

const state = ref<SiteQualityFindingStateFilter>(props.state)
watch(() => props.state, (value) => {
  state.value = value
})

defineEmits<{
  'change-filter': [state: SiteQualityFindingStateFilter]
  refresh: []
  open: [id: number]
  'change-page': [page: number]
}>()

const headingRuleLabel = (auditID: string): string => ({
  'heading-order': '标题层级跳级',
  'site-heading-missing-h1': '缺少主 H1',
  'site-heading-multiple-h1': '存在多个 H1',
  'site-heading-rendered-scan-failed': '渲染标题快照失败',
}[auditID] || auditID)

const strategyLabel = (value: SiteQualityStrategy): string => value === 'mobile' ? '移动端' : '桌面端'
const issueTone = (severity: string): AdminStatusTone => (
  severity === 'critical' || severity === 'high' ? 'coral' : severity === 'medium' ? 'amber' : 'gray'
)
const issueSeverityLabel = (severity: string): string => (
  { critical: '严重', high: '高', medium: '中', low: '低' }[severity] || severity
)
const findingStateLabel = (value: SiteQualityFindingState): string => (
  { open: '未确认', acknowledged: '处理中', resolved: '待验证', verified: '已验证' }[value]
)
const findingStateTone = (value: SiteQualityFindingState): AdminStatusTone => (
  { open: 'coral', acknowledged: 'amber', resolved: 'blue', verified: 'green' } as const
)[value]
const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')
</script>
