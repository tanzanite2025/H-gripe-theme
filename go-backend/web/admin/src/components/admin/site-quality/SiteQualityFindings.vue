<template>
 <section class="border bg-card">
 <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
      <div>
 <p class="text-sm font-black">质量事项</p>
 <p class="mt-1 text-xs text-muted-foreground">跨样本确认且未关闭的质量事项</p>
      </div>
 <div class="flex items-center gap-2">
 <select v-model="state" class="h-8 border bg-background px-2 text-xs" :disabled="loading" @change="$emit('change-filter', state)">
          <option value="active">待处理</option>
          <option value="open">未确认</option>
          <option value="acknowledged">处理中</option>
          <option value="resolved">待验证</option>
          <option value="verified">已验证</option>
          <option value="all">全部</option>
        </select>
        <Button size="icon" variant="outline" title="刷新质量事项" :disabled="loading" @click="$emit('refresh')">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
        </Button>
      </div>
    </div>
 <div v-if="loading && findings.length === 0" class="p-8 text-center text-sm text-muted-foreground">正在读取质量事项</div>
 <div v-else-if="findings.length === 0" class="p-8 text-center text-sm text-muted-foreground">当前筛选下没有质量事项</div>
 <div v-else class="divide-y">
 <article v-for="finding in findings" :key="finding.id" class="grid gap-3 px-4 py-3 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto] lg:items-center">
 <div class="min-w-0">
 <div class="flex flex-wrap items-center gap-2">
 <p class="truncate text-sm font-black">{{ finding.title || finding.audit_id }}</p>
            <AdminStatusBadge :tone="issueTone(finding.severity)">{{ issueSeverityLabel(finding.severity) }}</AdminStatusBadge>
            <AdminStatusBadge :tone="findingStateTone(finding.state)">{{ findingStateLabel(finding.state) }}</AdminStatusBadge>
          </div>
 <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">{{ finding.target_url }} · {{ strategyLabel(finding.strategy) }} · {{ finding.rule_id || finding.audit_id }}</p>
        </div>
 <span class="text-xs font-bold text-muted-foreground">{{ findingSavings(finding) || '-'}}</span>
 <span class="text-xs text-muted-foreground">{{ finding.resource_count }} 个资源</span>
        <Button size="icon" variant="ghost" title="查看质量事项详情" :disabled="loading" @click="$emit('open', finding.id)">
 <Eye class="size-4" />
        </Button>
      </article>
    </div>
 <div class="flex items-center justify-between border-t px-4 py-2">
 <span class="text-xs text-muted-foreground">共 {{ pagination.total }} 项</span>
 <div class="flex items-center gap-1">
        <Button size="icon" variant="outline" title="上一页" :disabled="loading || pagination.page <= 1" @click="$emit('change-page', pagination.page - 1)">
 <ChevronLeft class="size-4" />
        </Button>
        <Button size="icon" variant="outline" title="下一页" :disabled="loading || pagination.page >= pagination.total_pages" @click="$emit('change-page', pagination.page + 1)">
 <ChevronRight class="size-4" />
        </Button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ChevronLeft, ChevronRight, Eye, RefreshCw } from '@lucide/vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import type { SiteQualityFinding, SiteQualityFindingState, SiteQualityFindingStateFilter } from '@/api/preflight'

const props = defineProps<{
  loading: boolean
  findings: SiteQualityFinding[]
  state: SiteQualityFindingStateFilter
  pagination: { page: number; page_size: number; total: number; total_pages: number }
}>()
const state = ref<SiteQualityFindingStateFilter>(props.state)
watch(() => props.state, (value) => { state.value = value })

defineEmits<{
  'change-filter': [state: SiteQualityFindingStateFilter]
  refresh: []
  open: [id: number]
  'change-page': [page: number]
}>()

const strategyLabel = (value: SiteQualityFinding['strategy']): string => value === 'mobile' ? '移动端' : '桌面端'
const issueTone = (severity: string): AdminStatusTone => severity === 'critical' || severity === 'high' ? 'coral' : severity === 'medium' ? 'amber' : 'gray'
const issueSeverityLabel = (severity: string): string => ({ critical: '严重', high: '高', medium: '中', low: '低' }[severity] || severity)
const findingStateLabel = (state: SiteQualityFindingState): string => ({ open: '未确认', acknowledged: '处理中', resolved: '待验证', verified: '已验证' }[state])
const findingStateTone = (state: SiteQualityFindingState): AdminStatusTone => (
  { open: 'coral', acknowledged: 'amber', resolved: 'blue', verified: 'green' } as const
)[state]
const findingSavings = (finding: SiteQualityFinding): string => {
  if (typeof finding.latest_savings_ms === 'number' && finding.latest_savings_ms > 0) return `预计节省 ${Math.round(finding.latest_savings_ms)} ms`
  if (typeof finding.latest_savings_bytes === 'number' && finding.latest_savings_bytes > 0) return `预计节省 ${(finding.latest_savings_bytes / 1024).toFixed(1)} KiB`
  return ''
}
</script>
