<template>
 <section class="border bg-card">
 <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
      <div>
 <p class="text-sm font-black">检测历史</p>
 <p class="mt-1 text-xs text-muted-foreground">{{ pagination.total }} 条记录</p>
      </div>
 <div class="flex items-center gap-1">
        <Button size="icon" variant="outline" title="上一页" :disabled="loading || pagination.page <= 1" @click="$emit('change-page', pagination.page - 1)">
 <ChevronLeft class="size-4" />
        </Button>
        <Button size="icon" variant="outline" title="下一页" :disabled="loading || pagination.page >= pagination.total_pages" @click="$emit('change-page', pagination.page + 1)">
 <ChevronRight class="size-4" />
        </Button>
      </div>
    </div>
 <div v-if="loading && runs.length === 0" class="p-8 text-center text-sm text-muted-foreground">正在读取检测历史</div>
 <div v-else-if="runs.length === 0" class="p-8 text-center text-sm text-muted-foreground">尚无页面质量检测记录</div>
 <div v-else class="divide-y">
      <button
        v-for="run in runs"
        :key="run.id"
        type="button"
 class="grid w-full gap-3 px-4 py-3 text-left transition-colors hover:bg-muted/40 sm:grid-cols-[minmax(0,1fr)_auto_auto_auto] sm:items-center"
 :class="selectedRun?.id === run.id ? 'bg-muted/50': ''"
        @click="$emit('select-run', run)"
      >
 <span class="min-w-0">
 <span class="block truncate text-sm font-bold">{{ run.final_url || run.target_url }}</span>
 <span class="mt-1 block text-[10px] text-muted-foreground">{{ formatDate(run.created_at) }}</span>
        </span>
 <span class="text-xs font-bold">{{ strategyLabel(run.strategy) }}</span>
 <span class="text-sm font-black" :class="scoreTone(run.performance_score)">{{ displayScore(run.performance_score) }}</span>
        <AdminStatusBadge :tone="runTone(run.status)">
          {{ run.status === 'success' ? '完成' : '失败' }}
        </AdminStatusBadge>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ChevronLeft, ChevronRight } from '@lucide/vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import type { SiteQualityRun } from '@/api/preflight'

defineProps<{
  loading: boolean
  runs: SiteQualityRun[]
  selectedRun: SiteQualityRun | null
  pagination: { page: number; page_size: number; total: number; total_pages: number }
}>()

defineEmits<{
  'select-run': [run: SiteQualityRun]
  'change-page': [page: number]
}>()

const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')
const strategyLabel = (value: SiteQualityRun['strategy']): string => value === 'mobile' ? '移动端' : '桌面端'
const displayScore = (score?: number): string => typeof score === 'number' ? String(score) : '-'
const scoreTone = (score?: number): string => {
  if (typeof score !== 'number') return ''
  if (score >= 90) return 'text-emerald-600'
  if (score >= 50) return 'text-amber-600'
  return 'text-rose-600'
}
const runTone = (status: SiteQualityRun['status']): AdminStatusTone => status === 'success' ? 'green' : 'coral'
</script>
