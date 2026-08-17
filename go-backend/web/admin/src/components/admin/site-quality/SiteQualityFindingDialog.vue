<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
 <DialogContent size="xl" class="max-h-[calc(100dvh-1rem)]">
      <DialogHeader>
 <div class="flex min-w-0 items-start justify-between gap-3 pr-8">
 <div class="min-w-0">
            <DialogTitle>{{ selectedFinding?.title || selectedFinding?.audit_id || '质量事项' }}</DialogTitle>
 <DialogDescription class="mt-1 truncate font-mono text-[10px]">
              {{ selectedFinding?.target_url || '正在读取质量事项' }}
            </DialogDescription>
          </div>
 <div v-if="selectedFinding" class="flex shrink-0 gap-1">
            <AdminStatusBadge :tone="issueTone(selectedFinding.severity)">
              {{ issueSeverityLabel(selectedFinding.severity) }}
            </AdminStatusBadge>
            <AdminStatusBadge :tone="findingStateTone(selectedFinding.state)">
              {{ findingStateLabel(selectedFinding.state) }}
            </AdminStatusBadge>
          </div>
        </div>
      </DialogHeader>

 <div v-if="detailLoading" class="flex min-h-56 items-center justify-center text-sm text-muted-foreground">
        正在读取质量事项
      </div>

 <div v-else-if="selectedFinding" class="space-y-4 overflow-y-auto pr-1">
 <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">策略</p>
 <p class="mt-1 text-sm font-bold">{{ strategyLabel(selectedFinding.strategy) }}</p>
 <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ selectedFinding.audit_id }}</p>
          </div>
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">最近证据</p>
 <p class="mt-1 text-sm font-bold">{{ findingSavings(selectedFinding) || '无节省值'}}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">运行 #{{ selectedFinding.latest_run_id }}</p>
          </div>
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">{{ evidenceCountLabel(selectedFinding) }}</p>
 <p class="mt-1 text-sm font-bold">{{ selectedFinding.resource_count }} {{ evidenceCountUnit(selectedFinding) }}</p>
<p class="mt-1 text-[10px] text-muted-foreground">{{ formatDate(selectedFinding.last_detected_at) }}</p>
          </div>
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">统计确认</p>
 <p class="mt-1 text-sm font-bold">{{ formatConfidence(selectedFinding.confidence) }}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">
              {{ selectedFinding.confirmations }}/{{ selectedFinding.sample_count }} 次样本命中
            </p>
          </div>
 <div class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">处理状态</p>
 <p class="mt-1 text-sm font-bold">{{ findingStateLabel(selectedFinding.state) }}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">
              {{ selectedFinding.verified_at
                ? `验证于 ${formatDate(selectedFinding.verified_at)}`
                : `连续清洁评估 ${selectedFinding.consecutive_clean} 次` }}
            </p>
          </div>
        </div>

 <div v-if="selectedFindingEvidence?.description" class="border-l-2 border-amber-500 bg-card px-3 py-3 text-xs leading-5 text-muted-foreground">
          {{ selectedFindingEvidence.description }}
        </div>

 <section v-if="selectedFindingEvidence?.headings?.length" class="border border-dashed border-border/80">
 <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/70 px-3 py-3">
            <div>
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">HEADINGS</p>
 <h3 class="mt-1 text-sm font-black">命中标题节点</h3>
            </div>
 <span class="font-mono text-[10px] text-muted-foreground">{{ selectedFindingEvidence.headings.length }} 项</span>
          </div>
 <div class="divide-y">
            <div
              v-for="(heading, index) in selectedFindingEvidence.headings"
              :key="`${heading.selector || heading.snippet || heading.text || 'heading'}-${index}`"
 class="grid gap-2 px-3 py-3 sm:grid-cols-[auto_minmax(0,1fr)]"
            >
 <span class="font-mono text-xs font-black text-amber-700">
                H{{ heading.level || '?' }}
              </span>
 <div class="min-w-0">
 <p class="truncate text-xs font-bold">{{ heading.text || heading.snippet || '未命名标题' }}</p>
 <p v-if="heading.selector" class="mt-1 truncate font-mono text-[10px] text-muted-foreground" :title="heading.selector">
                  {{ heading.selector }}
                </p>
 <p v-if="heading.snippet" class="mt-1 truncate font-mono text-[10px] text-muted-foreground" :title="heading.snippet">
                  {{ heading.snippet }}
                </p>
 <p v-if="heading.explanation" class="mt-1 text-[11px] leading-5 text-muted-foreground">{{ heading.explanation }}</p>
              </div>
            </div>
          </div>
        </section>

        <section v-if="selectedFindingEvidence?.structured_data?.length" class="border border-dashed border-border/80">
          <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/70 px-3 py-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STRUCTURED DATA</p>
              <h3 class="mt-1 text-sm font-black">命中结构化数据</h3>
            </div>
            <span class="font-mono text-[10px] text-muted-foreground">
              {{ selectedFindingEvidence.structured_data.length }} 项
            </span>
          </div>
          <div class="divide-y">
            <div
              v-for="(schema, index) in selectedFindingEvidence.structured_data"
              :key="`${schema.selector || schema.type || schema.property || 'schema'}-${index}`"
              class="space-y-1.5 px-3 py-3"
            >
              <div class="flex flex-wrap items-center gap-2">
                <Braces class="size-3.5 text-amber-700" />
                <span v-if="schema.format" class="font-mono text-[10px] font-bold text-muted-foreground">
                  {{ schema.format }}
                </span>
                <span v-if="schema.type" class="text-xs font-black">{{ schema.type }}</span>
                <span v-if="schema.property" class="font-mono text-[10px] text-amber-700">
                  {{ schema.property }}
                </span>
              </div>
              <p v-if="schema.name" class="text-xs font-bold">{{ schema.name }}</p>
              <p v-if="schema.url" class="truncate font-mono text-[10px] text-muted-foreground" :title="schema.url">
                {{ schema.url }}
              </p>
              <p v-if="schema.id" class="truncate font-mono text-[10px] text-muted-foreground" :title="schema.id">
                {{ schema.id }}
              </p>
              <p v-if="schema.selector" class="truncate font-mono text-[10px] text-muted-foreground" :title="schema.selector">
                {{ schema.selector }}
              </p>
              <pre v-if="schema.snippet" class="max-h-32 overflow-auto whitespace-pre-wrap break-words bg-muted/30 p-2 font-mono text-[10px] leading-4 text-muted-foreground">{{ schema.snippet }}</pre>
              <p v-if="schema.explanation" class="text-[11px] leading-5 text-muted-foreground">
                {{ schema.explanation }}
              </p>
            </div>
          </div>
        </section>

        <section class="border border-dashed border-border/80">
 <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/70 px-3 py-3">
            <div>
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">EVIDENCE</p>
 <h3 class="mt-1 text-sm font-black">受影响资源</h3>
            </div>
 <span class="font-mono text-[10px] text-muted-foreground">{{ selectedFinding.resource_count }} 项</span>
          </div>
 <div v-if="!selectedFindingEvidence?.resources?.length" class="px-3 py-6 text-center text-xs text-muted-foreground">
            该审计未提供资源级明细
          </div>
 <div v-else class="divide-y">
            <div
              v-for="resource in selectedFindingEvidence.resources"
              :key="resource.url"
 class="grid gap-2 px-3 py-3 sm:grid-cols-[minmax(0,1fr)_auto_auto]"
            >
 <p class="truncate font-mono text-xs" :title="resource.url">{{ resource.url }}</p>
 <p class="text-xs text-muted-foreground">{{ formatBytes(resource.total_bytes) }}</p>
 <p class="text-xs font-bold text-amber-700">{{ resource.wasted_ms ? `预计延迟 ${formatMilliseconds(resource.wasted_ms)}` : '-'}}</p>
            </div>
          </div>
        </section>

 <div class="flex flex-wrap gap-2 border-y border-dashed border-border/70 py-3">
          <Button
            v-if="selectedFinding.state === 'open'"
            variant="outline"
            size="sm"
            :disabled="!canManage || findingActionKey !== null"
            @click="$emit('acknowledge')"
          >
 <Check class="size-3.5" />
            确认处理
          </Button>
          <Button
            v-if="selectedFinding.state === 'resolved'"
            size="sm"
            :disabled="!canManage || findingActionKey !== null"
            @click="$emit('recheck')"
          >
 <BadgeCheck :class="['size-3.5', findingActionKey === 'recheck'? 'animate-spin': '']" />
            运行复检
          </Button>
        </div>

        <form
          v-if="selectedFinding.state === 'open' || selectedFinding.state === 'acknowledged'"
 class="space-y-2 border border-dashed border-border/80 p-3"
          @submit.prevent="$emit('resolve')"
        >
 <div class="flex items-center justify-between gap-3">
            <div>
 <p class="text-xs font-bold">记录解决方案</p>
            </div>
            <Button type="submit" size="sm" :disabled="!canManage || findingActionKey !== null || !resolutionNote.trim()">
 <CheckCheck class="size-3.5" />
              记录解决
            </Button>
          </div>
          <Textarea
            :model-value="resolutionNote"
            placeholder="说明已经完成的修复，以及预期如何在复检中验证"
            @update:model-value="$emit('update:resolution-note', String($event))"
          />
        </form>

 <div v-if="selectedFinding.resolution_note" class="border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">RESOLUTION NOTE</p>
 <p class="mt-2 whitespace-pre-wrap text-xs text-muted-foreground">{{ selectedFinding.resolution_note }}</p>
        </div>

 <section class="border border-dashed border-border/80">
 <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/70 px-3 py-3">
            <div>
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">TIMELINE</p>
 <h3 class="mt-1 text-sm font-black">处理时间线</h3>
            </div>
 <span class="font-mono text-[10px] text-muted-foreground">共 {{ findingEventsPagination.total }} 项</span>
          </div>
 <div v-if="findingEvents.length === 0" class="px-3 py-8 text-center text-xs text-muted-foreground">
            暂无处理事件
          </div>
 <div v-else class="max-h-64 divide-y overflow-auto">
 <div v-for="event in findingEvents" :key="event.id" class="px-3 py-3">
 <div class="flex items-center justify-between gap-3">
 <p class="text-xs font-bold">{{ findingEventLabel(event.event_type) }}</p>
 <p class="whitespace-nowrap text-[10px] text-muted-foreground">{{ formatDate(event.created_at) }}</p>
              </div>
 <p v-if="event.note" class="mt-1 whitespace-pre-wrap text-xs text-muted-foreground">{{ event.note }}</p>
 <p v-if="event.actor_user_id" class="mt-1 font-mono text-[10px] text-muted-foreground">用户 #{{ event.actor_user_id }}</p>
            </div>
          </div>
        </section>
      </div>

      <DialogFooter>
        <Button type="button" variant="outline" @click="$emit('update:open', false)">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { BadgeCheck, Braces, Check, CheckCheck } from '@lucide/vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import type {
  SiteQualityFinding,
  SiteQualityFindingEvidence,
  SiteQualityFindingEvent,
  SiteQualityFindingState,
  SiteQualityStrategy,
} from '@/api/preflight'

defineProps<{
  open: boolean
  detailLoading: boolean
  selectedFinding: SiteQualityFinding | null
  selectedFindingEvidence: SiteQualityFindingEvidence | null
  findingEvents: SiteQualityFindingEvent[]
  findingEventsPagination: { page: number; page_size: number; total: number; total_pages: number }
  findingActionKey: string | null
  canManage: boolean
  resolutionNote: string
}>()

defineEmits<{
  'update:open': [value: boolean]
  'update:resolution-note': [value: string]
  acknowledge: []
  resolve: []
  recheck: []
}>()

const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')
const strategyLabel = (value: SiteQualityStrategy): string => value === 'mobile' ? '移动端' : '桌面端'
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
const findingStateLabel = (state: SiteQualityFindingState): string => ({
  open: '未确认',
  acknowledged: '处理中',
  resolved: '待验证',
  verified: '已验证',
}[state])
const findingStateTone = (state: SiteQualityFindingState): AdminStatusTone => ({
  open: 'coral',
  acknowledged: 'amber',
  resolved: 'blue',
  verified: 'green',
} as const)[state]
const findingSavings = (finding: SiteQualityFinding): string => {
  if (typeof finding.latest_savings_ms === 'number' && finding.latest_savings_ms > 0) {
    return `预计节省 ${formatMilliseconds(finding.latest_savings_ms)}`
  }
  if (typeof finding.latest_savings_bytes === 'number' && finding.latest_savings_bytes > 0) {
    return `预计节省 ${(finding.latest_savings_bytes / 1024).toFixed(1)} KiB`
  }
  return ''
}
const evidenceCountLabel = (finding: SiteQualityFinding): string => (
  finding.audit_id === 'site-heading-rendered-scan-failed'
    || finding.audit_id === 'site-schema-rendered-scan-failed'
    ? '检测状态'
    : finding.finding_kind === 'headings'
      ? '标题节点'
      : finding.finding_kind === 'schema'
        ? '结构化证据'
      : '资源影响'
)
const evidenceCountUnit = (finding: SiteQualityFinding): string => (
  finding.audit_id === 'site-heading-rendered-scan-failed'
    || finding.audit_id === 'site-schema-rendered-scan-failed'
    ? '项'
    : finding.finding_kind === 'headings'
      ? '个节点'
      : finding.finding_kind === 'schema'
        ? '项证据'
      : '个资源'
)
const formatConfidence = (value?: number): string => (
  typeof value === 'number' ? `${Math.round(value * 100)}% 置信度` : '-'
)
const formatMilliseconds = (value?: number): string => {
  if (typeof value !== 'number') return '-'
  if (value < 1000) return `${Math.round(value)} ms`
  return `${(value / 1000).toFixed(2)} s`
}
const formatBytes = (value?: number): string => {
  if (typeof value !== 'number' || value <= 0) return '-'
  if (value < 1024) return `${Math.round(value)} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`
  return `${(value / (1024 * 1024)).toFixed(2)} MiB`
}
const findingEventLabel = (eventType: string): string => ({
  detected: '检测到问题',
  reopened: '问题重新出现',
  acknowledged: '确认处理',
  resolution_recorded: '记录解决方案',
  verification_passed: '复检验证通过',
}[eventType] || eventType)
</script>
