<template>
  <section class="space-y-3">
    <div class="border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Provider Risk Strategy Snapshot</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">{{ paymentProviderDisplayName }} 风控策略口径</h2>
          <p class="mt-1 max-w-4xl text-xs leading-5 text-muted-foreground">
            本页只读取 {{ paymentProviderDisplayName }} provider 数据；争议率和退款率分开看，分母均为同一 provider 的 30 天成功支付笔数。
          </p>
        </div>
        <span class="border px-2 py-1 text-[10px] font-black uppercase tracking-wider" :class="monitoringEnabledClass">
          {{ monitoringEnabled ? '监控已启用' : '监控未启用' }}
        </span>
      </div>

      <dl class="mt-4 grid grid-cols-1 gap-px overflow-hidden border border-border/70 bg-border/70 sm:grid-cols-2 xl:grid-cols-4">
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">观察窗口</dt>
          <dd class="mt-1 font-mono text-sm font-semibold">{{ windowDays }} 天</dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">成功支付分母</dt>
          <dd class="mt-1 font-mono text-sm font-semibold">{{ count(riskStrategySnapshot?.successful_payment_count) }} 笔</dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">当前灯号</dt>
          <dd class="mt-1">
            <span class="border px-2 py-1 text-[10px] font-black uppercase tracking-wider" :class="riskStrategyLevelClass(riskStrategySnapshot?.level)">
              {{ riskStrategyLevelLabel(riskStrategySnapshot?.level) }}
            </span>
          </dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">外部告警</dt>
          <dd class="mt-1 text-sm font-semibold">{{ riskStrategyMonitoringPolicy?.alerting_enabled ? '已接入' : '未接入' }}</dd>
        </div>
      </dl>
    </div>

    <div v-if="loading" class="border border-dashed border-border/80 bg-card p-6 text-sm text-muted-foreground">
      正在读取 {{ paymentProviderDisplayName }} 风控策略总览...
    </div>

    <template v-else-if="riskStrategySnapshot">
      <section class="grid grid-cols-1 gap-3 lg:grid-cols-2">
        <article class="border border-dashed border-border/80 bg-card p-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">30D Dispute Rate</p>
              <h3 class="mt-1 text-base font-black tracking-tight">30D 争议率</h3>
              <p class="mt-1 text-xs text-muted-foreground">
                {{ count(riskStrategySnapshot.dispute_count) }} 笔争议 / {{ count(riskStrategySnapshot.successful_payment_count) }} 笔成功支付
              </p>
            </div>
            <strong class="font-mono text-2xl font-black tabular-nums">{{ percent(riskStrategySnapshot.dispute_activity_rate) }}</strong>
          </div>
          <dl class="mt-4 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70">
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">预警阈值</dt>
              <dd class="mt-1 font-mono text-sm font-semibold">{{ percent(riskStrategyMonitoringPolicy?.warning_dispute_activity_rate) }}</dd>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">严重阈值</dt>
              <dd class="mt-1 font-mono text-sm font-semibold">{{ percent(riskStrategyMonitoringPolicy?.critical_dispute_activity_rate) }}</dd>
            </div>
          </dl>
        </article>

        <article class="border border-dashed border-border/80 bg-card p-4">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">30D Refund Rate</p>
              <h3 class="mt-1 text-base font-black tracking-tight">30D 退款率</h3>
              <p class="mt-1 text-xs text-muted-foreground">
                {{ count(riskStrategySnapshot.refund_count) }} 笔退款 / {{ count(riskStrategySnapshot.successful_payment_count) }} 笔成功支付
              </p>
            </div>
            <strong class="font-mono text-2xl font-black tabular-nums">{{ percent(riskStrategySnapshot.refund_rate) }}</strong>
          </div>
          <dl class="mt-4 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70">
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">预警阈值</dt>
              <dd class="mt-1 font-mono text-sm font-semibold">{{ percent(riskStrategyMonitoringPolicy?.warning_refund_rate) }}</dd>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">严重阈值</dt>
              <dd class="mt-1 font-mono text-sm font-semibold">{{ percent(riskStrategyMonitoringPolicy?.critical_refund_rate) }}</dd>
            </div>
          </dl>
        </article>
      </section>

      <section class="grid grid-cols-1 gap-3 xl:grid-cols-[minmax(0,1.1fr)_minmax(360px,0.9fr)]">
        <article class="border border-dashed border-border/80 bg-card p-4">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Alert & Trigger</p>
              <h3 class="mt-1 text-base font-black tracking-tight">触发时间和规则</h3>
            </div>
            <span class="text-[11px] text-muted-foreground">最后计算：{{ formatDateTime(riskStrategySnapshot.computed_at) }}</span>
          </div>
          <dl class="mt-4 grid grid-cols-1 gap-px overflow-hidden border border-border/70 bg-border/70 md:grid-cols-3">
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">当前动作</dt>
              <dd class="mt-1 text-sm font-semibold">{{ riskStrategyRecommendedActionLabel(riskStrategySnapshot.recommended_action) }}</dd>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">快照周期</dt>
              <dd class="mt-1 text-sm font-semibold">{{ riskStrategySnapshotPeriodLabel }}</dd>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">规则来源</dt>
              <dd class="mt-1 text-sm font-semibold">风控策略监控规则</dd>
            </div>
          </dl>
          <div class="mt-4 rounded-xl border border-dashed border-border/80 p-3 text-xs leading-5">
            <p v-if="riskStrategyTriggerRuleLabels.length" class="font-semibold text-amber-700">
              已触发：{{ riskStrategyTriggerRuleLabels.join(' · ') }}
            </p>
            <p v-else class="font-semibold text-emerald-700">当前没有触发争议率或退款率阈值。</p>
            <p class="mt-2 text-muted-foreground">
              争议率 = 30 天 {{ paymentProviderDisplayName }} 争议笔数 / 30 天 {{ paymentProviderDisplayName }} 成功支付笔数；退款率同理，两个指标独立判断。
            </p>
          </div>
        </article>

        <article class="border border-dashed border-border/80 bg-card p-4">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Gateway State</p>
              <h3 class="mt-1 text-base font-black tracking-tight">告警 / 熔断状态</h3>
            </div>
            <span class="border px-2 py-1 text-[10px] font-black uppercase tracking-wider" :class="paymentGatewayCircuitBreakerClass">
              {{ paymentGatewayCircuitBreakerLabel }}
            </span>
          </div>
          <dl class="mt-4 space-y-2 text-xs">
            <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
              <dt class="text-muted-foreground">网关调用</dt>
              <dd class="font-semibold">{{ gatewayHealth?.allowed === false ? '当前拦截' : '允许调用' }}</dd>
            </div>
            <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
              <dt class="text-muted-foreground">失败率</dt>
              <dd class="font-mono font-semibold">{{ percent(gatewayHealth?.failure_rate) }}</dd>
            </div>
            <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
              <dt class="text-muted-foreground">失败样本</dt>
              <dd class="font-mono font-semibold">{{ count(gatewayHealth?.failure_count) }} / {{ count(gatewayHealth?.sample_count) }}</dd>
            </div>
            <div class="flex items-start justify-between gap-3">
              <dt class="text-muted-foreground">恢复倒计时</dt>
              <dd class="font-mono font-semibold">{{ retryAfterSecondsLabel }}</dd>
            </div>
          </dl>
          <p v-if="gatewayHealth?.error" class="mt-3 rounded-xl border border-rose-500/30 bg-rose-500/10 p-3 text-xs font-semibold text-rose-700">
            {{ gatewayHealth.error }}
          </p>
        </article>
      </section>

      <p class="text-[11px] text-muted-foreground">
        注：金额按原始币种累计，当前页面不做跨币种换汇；本页核心判断以笔数比例为准。
      </p>
    </template>

    <div v-else class="border border-dashed border-border/80 bg-card p-6 text-sm text-muted-foreground">
      {{ monitoringEnabled ? `等待定时任务或新的 ${paymentProviderDisplayName} 风控策略事件生成首次计算。` : '风控策略监控当前未启用。' }}
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type {
  RiskStrategyGatewayHealth,
  RiskStrategyMonitoringPolicy,
  RiskStrategyProviderReport,
  RiskStrategySnapshot,
} from './riskStrategyTypes'

const props = withDefaults(defineProps<{
  paymentProviderDisplayName: string
  riskStrategyProviderReport?: RiskStrategyProviderReport | null
  riskStrategyMonitoringPolicy?: RiskStrategyMonitoringPolicy | null
  gatewayHealth?: RiskStrategyGatewayHealth | null
  monitoringEnabled?: boolean
  loading?: boolean
}>(), {
  riskStrategyProviderReport: null,
  riskStrategyMonitoringPolicy: null,
  gatewayHealth: null,
  monitoringEnabled: false,
  loading: false,
})

const riskStrategySnapshot = computed<RiskStrategySnapshot | null>(() => props.riskStrategyProviderReport?.snapshot || null)
const riskStrategyTriggerRuleLabels = computed(() => (
  Array.isArray(props.riskStrategyProviderReport?.reasons)
    ? props.riskStrategyProviderReport.reasons.map(riskStrategyTriggerRuleLabel)
    : []
))
const windowDays = computed(() => Number(props.riskStrategyMonitoringPolicy?.window_days || riskStrategySnapshot.value?.window_days || 30))
const monitoringEnabledClass = computed(() => (
  props.monitoringEnabled
    ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'
    : 'border-amber-500/25 bg-amber-500/10 text-amber-700'
))
const riskStrategySnapshotPeriodLabel = computed(() => {
  const snapshot = riskStrategySnapshot.value
  if (!snapshot?.window_start || !snapshot?.window_end) return `${snapshot?.window_days || windowDays.value} 天窗口`
  const start = new Date(snapshot.window_start).toLocaleDateString('zh-CN')
  const end = new Date(snapshot.window_end).toLocaleDateString('zh-CN')
  return `${start} - ${end}`
})
const paymentGatewayCircuitBreakerLabel = computed(() => {
  if (!props.gatewayHealth) return '未读取'
  if (props.gatewayHealth.circuit_open) return '熔断中'
  if (props.gatewayHealth.allowed === false) return '当前拦截'
  return '允许调用'
})
const paymentGatewayCircuitBreakerClass = computed(() => {
  if (!props.gatewayHealth) return 'border-border bg-muted text-muted-foreground'
  if (props.gatewayHealth.circuit_open || props.gatewayHealth.allowed === false) {
    return 'border-rose-500/25 bg-rose-500/10 text-rose-700'
  }
  return 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'
})
const retryAfterSecondsLabel = computed(() => {
  const seconds = Number(props.gatewayHealth?.retry_after_seconds || 0)
  return seconds > 0 ? `${Math.round(seconds)} 秒` : '无'
})

const numeric = (value: unknown): number => {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) ? parsed : 0
}
const count = (value: unknown): string => numeric(value).toLocaleString('zh-CN')
const percent = (value: unknown): string => `${(numeric(value) * 100).toFixed(2)}%`
const formatDateTime = (value: unknown): string => value ? new Date(value as string | number | Date).toLocaleString('zh-CN') : '-'
const riskStrategyLevelLabel = (level?: string): string => ({
  normal: '正常',
  warning: '预警',
  critical: '严重',
} as Record<string, string>)[level || ''] || '未计算'
const riskStrategyLevelClass = (level?: string): string => ({
  normal: 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700',
  warning: 'border-amber-500/25 bg-amber-500/10 text-amber-700',
  critical: 'border-rose-500/25 bg-rose-500/10 text-rose-700',
} as Record<string, string>)[level || ''] || 'border-border bg-muted text-muted-foreground'
const riskStrategyTriggerRuleLabel = (rule: string): string => ({
  dispute_activity_rate_warning: '30D 争议率达到预警阈值',
  dispute_activity_rate_critical: '30D 争议率达到严重阈值',
  early_fraud_warning_rate_warning: '30D EFW 达到预警阈值',
  early_fraud_warning_rate_critical: '30D EFW 达到严重阈值',
  refund_rate_warning: '30D 退款率达到预警阈值',
  refund_rate_critical: '30D 退款率达到严重阈值',
}[rule] || rule)
const riskStrategyRecommendedActionLabel = (action?: string): string => ({
  continue_monitoring: '继续监控',
  collect_more_volume_and_review_efw: '样本不足，继续收集数据并重点复核 EFW',
  force_3ds_for_high_risk_and_notify_operator: '高风险支付升级 3DS，并通知管理员',
  force_3ds_for_high_risk_and_operator_review: '高风险支付升级 3DS，并要求人工复核',
}[action || ''] || action || '未指定')
</script>

