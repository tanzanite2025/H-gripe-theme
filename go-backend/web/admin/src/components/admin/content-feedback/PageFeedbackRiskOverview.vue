<template>
  <section class="shrink-0 space-y-3 rounded-lg border border-dashed border-border/80 bg-card p-3" aria-label="留言安全概览">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div class="flex min-w-0 items-center gap-2">
        <ShieldAlert class="size-4 shrink-0 text-muted-foreground" />
        <div class="min-w-0">
          <h2 class="truncate text-sm font-black">留言安全概览</h2>
          <p class="text-[11px] text-muted-foreground">最近 {{ overview?.window_hours || 24 }} 小时</p>
        </div>
      </div>
      <AdminStatusBadge v-if="overview" :tone="pageFeedbackRiskTone(overview.level)">
        {{ pageFeedbackRiskLabel(overview.level) }}
      </AdminStatusBadge>
      <span v-else class="text-xs text-muted-foreground">{{ loading ? '加载中...' : '暂不可用' }}</span>
    </div>

    <div v-if="overview" class="space-y-3">
      <AdminStatsGrid :items="statItems" />

      <div class="grid grid-cols-1 gap-3 xl:grid-cols-2">
        <div class="min-w-0 rounded-md border border-border/70 bg-muted/10 p-3">
          <div class="mb-2 flex items-center gap-2">
            <MessageSquareWarning class="size-3.5 text-muted-foreground" />
            <h3 class="text-xs font-bold">热点页面</h3>
          </div>
          <div v-if="overview.hot_pages.length" class="space-y-1">
            <button
              v-for="page in overview.hot_pages"
              :key="`${page.page_path}:${page.thread_key}`"
              type="button"
              class="flex w-full min-w-0 items-center justify-between gap-3 rounded-md px-2 py-1.5 text-left transition-colors hover:bg-muted"
              :title="`按页面筛选 ${displayRiskPagePath(page)}`"
              @click="emit('filter-page', {
                kind: page.filter_kind || (page.page_path ? 'page_path' : 'thread_key'),
                value: page.filter_value || page.page_path || page.thread_key,
              })"
            >
              <span class="min-w-0">
                <span class="block truncate text-xs font-semibold">{{ page.page_title || page.page_path || page.thread_key }}</span>
                <span class="block truncate font-mono text-[10px] text-muted-foreground">{{ displayRiskPagePath(page) }}</span>
              </span>
              <span class="shrink-0 text-right text-[10px] text-muted-foreground">
                <strong class="block text-xs tabular-nums text-foreground">{{ page.feedback_count }}</strong>
                {{ page.pending_count }} 待处理
              </span>
            </button>
          </div>
          <p v-else class="text-xs text-muted-foreground">最近窗口没有形成热点的页面。</p>
        </div>

        <div class="min-w-0 rounded-md border border-border/70 bg-muted/10 p-3">
          <div class="mb-2 flex items-center gap-2">
            <Fingerprint class="size-3.5 text-muted-foreground" />
            <h3 class="text-xs font-bold">来源突发</h3>
          </div>
          <div v-if="overview.source_bursts.length" class="space-y-1">
            <div
              v-for="source in overview.source_bursts"
              :key="source.source_hash_preview"
              class="flex min-w-0 items-center justify-between gap-3 rounded-md px-2 py-1.5"
            >
              <span class="min-w-0">
                <span class="block truncate font-mono text-xs">{{ displaySourceHashPreview(source.source_hash_preview) }}</span>
                <span class="block text-[10px] text-muted-foreground">{{ source.page_count }} 个页面有提交</span>
              </span>
              <span class="shrink-0 text-right text-[10px] text-muted-foreground">
                <strong class="block text-xs tabular-nums text-foreground">{{ source.feedback_count }}</strong>
                {{ source.pending_count }} 待处理
              </span>
            </div>
          </div>
          <p v-else class="text-xs text-muted-foreground">最近窗口没有检测到跨页面来源突发。</p>
        </div>
      </div>

      <p v-if="overview.rate_limit.unavailable" class="text-[11px] text-amber-700">
        Redis 分布式限流计数暂不可用，当前显示本机兜底统计。
      </p>
      <p v-else-if="overview.rate_limit.redis_unavailable > 0 || overview.rate_limit.fallback_total > 0" class="text-[11px] text-muted-foreground">
        最近窗口 Redis 限流不可用 {{ overview.rate_limit.redis_unavailable }} 次，本机兜底拦截 {{ overview.rate_limit.fallback_total }} 次。
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Fingerprint, MessageSquareWarning, ShieldAlert } from '@lucide/vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import type { PageFeedbackRiskOverview, PageFeedbackRiskPage } from '@/api/pageFeedback'
import {
  displaySourceHashPreview,
  pageFeedbackRiskLabel,
  pageFeedbackRiskTone,
} from './pageFeedbackTypes'

const props = defineProps<{
  overview: PageFeedbackRiskOverview | null
  loading: boolean
}>()

const emit = defineEmits<{
  'filter-page': [filter: { kind: 'page_path' | 'thread_key'; value: string }]
}>()

const statItems = computed(() => {
  if (!props.overview) return []
  const totals = props.overview.totals
  const rateLimit = props.overview.rate_limit
  return [
    { key: 'pending', label: '待处理', value: totals.pending_total, tone: 'amber', icon: ShieldAlert },
    { key: 'stale', label: '超 24 小时', value: totals.pending_over_24h, tone: 'coral', icon: MessageSquareWarning },
    { key: 'last-hour', label: '近 1 小时', value: totals.last_hour_total, tone: 'blue', icon: Fingerprint },
    {
      key: 'blocked',
      label: '限流拦截',
      value: rateLimit.total,
      tone: rateLimit.total > 0 ? 'coral' : 'green',
      icon: ShieldAlert,
    },
  ]
})

const displayRiskPagePath = (page: PageFeedbackRiskPage): string => (
  page.page_path || `thread:${page.thread_key}`
)
</script>
