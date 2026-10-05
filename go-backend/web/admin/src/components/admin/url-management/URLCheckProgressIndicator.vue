<template>
  <aside
    v-if="operation.running || operation.status === 'failed'"
    class="fixed bottom-4 right-4 z-[90] w-[min(22rem,calc(100vw-2rem))] rounded-xl border border-border bg-card p-3 text-card-foreground shadow-lg"
    aria-live="polite"
  >
    <div class="flex items-center justify-between gap-3">
      <div class="min-w-0">
        <p class="truncate text-xs font-black">
          {{ operation.status === 'failed' ? 'URL 检查失败' : operation.status === 'queued' ? '正在同步最新路由清单' : '正在检查 URL' }}
        </p>
        <p v-if="operation.locale" class="mt-0.5 font-mono text-[10px] text-muted-foreground">
          {{ operation.locale }}
        </p>
      </div>
      <span v-if="operation.status !== 'failed'" class="shrink-0 font-mono text-xs font-bold tabular-nums">
        {{ operation.progress.percentage }}%
      </span>
    </div>

    <div v-if="operation.status !== 'failed'" class="mt-2 h-1.5 overflow-hidden rounded-full bg-muted" role="progressbar" :aria-valuenow="operation.progress.percentage" aria-valuemin="0" aria-valuemax="100">
      <div
        class="h-full rounded-full bg-primary transition-[width] duration-300"
        :style="{ width: `${operation.progress.percentage}%` }"
      />
    </div>
    <p class="mt-2 text-[11px] text-muted-foreground">
      本次 {{ operation.progress.eligible }} 条 · 每批 {{ operation.batchSize || operation.summary.batch_size || 200 }} 条
      <span v-if="operation.summary.total_batches"> · 第 {{ operation.summary.current_batch || 0 }} / {{ operation.summary.total_batches }} 批</span>
    </p>
    <p class="mt-1 text-[11px] text-muted-foreground">
      已完成 {{ operation.progress.checked }} 条 · 剩余 {{ operation.progress.remaining }} 条
      <span v-if="operation.progress.estimatedSeconds"> · 预计 {{ operation.progress.estimatedSeconds }} 秒</span>
    </p>
    <p class="mt-1 text-[10px] text-muted-foreground">
      正常 {{ operation.summary.ok }} · 跳转 {{ operation.summary.redirects }} · 404 {{ operation.summary.not_found }} · 5xx {{ operation.summary.server_errors }} · Canonical {{ operation.summary.canonical_mismatch }} · 请求错误 {{ operation.summary.errors }}
    </p>
    <p v-if="operation.status === 'failed'" class="mt-1 text-[11px] text-destructive">{{ operation.error || '任务中断，剩余 URL 可按“未检查”筛选后重跑' }}</p>
  </aside>
</template>

<script setup lang="ts">
import { useURLOperationStore } from '@/stores/urlOperation'

const operation = useURLOperationStore()
</script>
