<template>
  <aside
    v-if="operation.running || operation.status === 'failed'"
    class="fixed bottom-4 right-4 z-[90] w-[min(22rem,calc(100vw-2rem))] rounded-xl border border-border bg-card p-3 text-card-foreground shadow-lg"
    aria-live="polite"
  >
    <div class="flex items-center justify-between gap-3">
      <div class="min-w-0">
        <p class="truncate text-xs font-black">
          {{ operation.status === 'failed' ? 'URL 检查失败' : '正在检查 URL' }}
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
    <p v-if="operation.status !== 'failed'" class="mt-2 text-[11px] text-muted-foreground">
      {{ operation.progress.checked }} / {{ operation.progress.eligible }} 条 · 剩余 {{ operation.progress.remaining }} 条
      <span v-if="operation.progress.estimatedSeconds"> · 预计 {{ operation.progress.estimatedSeconds }} 秒</span>
    </p>
    <p v-else class="mt-2 text-[11px] text-destructive">{{ operation.error || '请稍后重试' }}</p>
  </aside>
</template>

<script setup lang="ts">
import { useURLOperationStore } from '@/stores/urlOperation'

const operation = useURLOperationStore()
</script>
