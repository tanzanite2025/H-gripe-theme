<template>
  <section class="grid grid-cols-1 gap-3 lg:grid-cols-2">
    <article
      v-for="provider in providers"
      :key="provider.key"
      class="border border-dashed border-border/80 bg-card p-4"
    >
      <div class="flex items-start justify-between gap-3">
        <div>
          <h2 class="text-sm font-black uppercase tracking-wide">{{ provider.label }}</h2>
          <p class="mt-1 text-xs text-muted-foreground">{{ provider.snapshot ? formatPeriod(provider.snapshot) : '尚未生成风险快照' }}</p>
        </div>
        <span :class="levelClass(provider.snapshot?.level)" class="border px-2 py-1 text-[10px] font-black uppercase tracking-wider">
          {{ levelLabel(provider.snapshot?.level) }}
        </span>
      </div>

      <div v-if="loading" class="mt-5 text-sm text-muted-foreground">正在读取...</div>
      <div v-else-if="provider.snapshot" class="mt-4 grid grid-cols-3 gap-px overflow-hidden border border-border/70 bg-border/70">
        <div v-for="metric in provider.metrics" :key="metric.label" class="min-w-0 bg-card p-3">
          <div class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">{{ metric.label }}</div>
          <div class="mt-1 truncate font-mono text-sm font-semibold">{{ metric.value }}</div>
        </div>
      </div>
      <div v-else class="mt-5 text-sm text-muted-foreground">
        {{ enabled ? '等待定时任务或新支付风险事件触发首次计算。' : '支付风险监控当前未启用。' }}
      </div>

      <p v-if="provider.reasons.length" class="mt-3 text-xs text-muted-foreground">
        {{ provider.reasons.join(' · ') }}
      </p>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { PaymentRiskProviderReport, PaymentRiskReports, PaymentRiskSnapshot } from './paymentRiskTypes'

const props = withDefaults(defineProps<{
  reports?: PaymentRiskReports
  enabled?: boolean
  loading?: boolean
}>(), {
  reports: () => ({}),
  enabled: false,
  loading: false,
})

const providers = computed(() => ['stripe', 'paypal'].map((key) => {
  const report: PaymentRiskProviderReport | null = props.reports?.[key] || null
  const snapshot = report?.snapshot || null
  return {
    key,
    label: key === 'stripe' ? 'Stripe' : 'PayPal',
    snapshot,
    reasons: Array.isArray(report?.reasons) ? report.reasons : [],
    metrics: snapshot ? [
      { label: '成功支付', value: formatCount(snapshot.successful_payment_count) },
      { label: '争议率', value: formatPercent(snapshot.dispute_activity_rate) },
      { label: 'EFW', value: formatPercent(snapshot.early_fraud_warning_rate) },
      { label: '退款率', value: formatPercent(snapshot.refund_rate) },
      { label: '争议数', value: formatCount(snapshot.dispute_count) },
      { label: '退款数', value: formatCount(snapshot.refund_count) },
    ] : [],
  }
}))

const formatCount = (value: unknown): string => Number(value || 0).toLocaleString('zh-CN')
const formatPercent = (value: unknown): string => `${(Number(value || 0) * 100).toFixed(2)}%`
const formatPeriod = (snapshot?: PaymentRiskSnapshot | null): string => {
  if (!snapshot?.window_start || !snapshot?.window_end) return `${snapshot?.window_days || 30} 天窗口`
  const start = new Date(snapshot.window_start).toLocaleDateString('zh-CN')
  const end = new Date(snapshot.window_end).toLocaleDateString('zh-CN')
  return `${start} - ${end}`
}
const levelLabel = (level?: string): string => ({
  normal: '正常',
  warning: '预警',
  critical: '严重',
} as Record<string, string>)[level || ''] || '未计算'
const levelClass = (level?: string): string => ({
  normal: 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700',
  warning: 'border-amber-500/25 bg-amber-500/10 text-amber-700',
  critical: 'border-rose-500/25 bg-rose-500/10 text-rose-700',
} as Record<string, string>)[level || ''] || 'border-border bg-muted text-muted-foreground'
</script>
