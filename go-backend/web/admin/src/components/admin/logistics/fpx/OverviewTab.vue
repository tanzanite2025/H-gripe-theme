<script setup lang="ts">
import { Activity, ClipboardCheck, Truck } from '@lucide/vue'


const overviewMetrics = [
  { label: '今日直发大包', value: '—', detail: '等待直发订单数据接入', icon: Truck },
  { label: '待处理退件', value: '—', detail: '等待内部退件台账接入', icon: ClipboardCheck },
  { label: '待核对费用', value: '—', detail: '等待订单费用查询接入', icon: Activity },
]

</script>

<template>
      <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3" aria-label="4PX 运营指标">
        <article
          v-for="metric in overviewMetrics"
          :key="metric.label"
          class="group relative min-h-36 overflow-hidden rounded-[24px] border border-dashed border-border/80 bg-card p-4"
        >
          <div class="pointer-events-none absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-transparent" />
          <div class="relative flex h-full flex-col justify-between gap-4">
            <div class="flex items-center justify-between gap-2">
              <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">{{ metric.label }}</p>
              <component :is="metric.icon" class="size-4 text-muted-foreground/70" aria-hidden="true" />
            </div>
            <div>
              <p class="font-mono text-3xl font-black tracking-tight text-foreground">{{ metric.value }}</p>
              <p class="mt-1 text-[11px] leading-4 text-muted-foreground">{{ metric.detail }}</p>
            </div>
          </div>
        </article>
      </section>

      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-muted/5 p-5 sm:p-6">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground/70">4PX Shipping Operations</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">大件发货链路监控</h2>
            <p class="mt-1 text-xs text-muted-foreground">直发订单、面单、交运状态与轨迹将在 4PX 发货接口接入后展示。</p>
          </div>
          <span class="inline-flex w-fit items-center gap-2 rounded-full border border-amber-500/30 bg-amber-500/10 px-3 py-1.5 text-[11px] font-bold text-amber-700 dark:text-amber-300">
            <span class="size-1.5 rounded-full bg-amber-500" />
            接口待接入
          </span>
        </div>

        <div class="flex items-start gap-2 rounded-2xl border border-border/70 bg-card/70 px-4 py-3 text-xs leading-5 text-muted-foreground">
          <Activity class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
          <p>此大盘当前展示的是待接入状态，不包含模拟运营数据。完成 4PX 凭据配置后，再接入直发、面单、状态、轨迹和订单费用查询。</p>
        </div>
      </section>
</template>
