<template>
  <section class="space-y-3">
    <div class="border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Risk Strategy Observability</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">风控策略实时总览</h2>
          <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">
            所有比例都展示实际分子 / 分母。风险快照来自最近一次计算，3DS 数据来自支付启动时记录的策略决定。
          </p>
        </div>
        <span
 :class="enabled ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700': 'border-amber-500/25 bg-amber-500/10 text-amber-700'"
          class="border px-2 py-1 text-[10px] font-black uppercase tracking-wider"
        >
          {{ enabled ? '监控已启用' : '监控未启用' }}
        </span>
      </div>

      <dl class="mt-4 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70 md:grid-cols-4">
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">观察窗口</dt>
          <dd class="mt-1 font-mono text-sm font-semibold">{{ windowDays }} 天</dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">最小样本</dt>
          <dd class="mt-1 font-mono text-sm font-semibold">{{ minimumPayments }} 笔成功支付</dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">自动升级 3DS</dt>
 <dd class="mt-1 text-sm font-semibold">{{ policy?.auto_step_up_enabled ? '已开启': '未开启'}}</dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">外部告警</dt>
 <dd class="mt-1 text-sm font-semibold">{{ policy?.alerting_enabled ? '已接入': '未接入'}}</dd>
        </div>
      </dl>

      <div class="mt-3 border-t border-dashed border-border/70 pt-3 text-[11px] leading-5 text-muted-foreground">
        <span class="font-semibold text-foreground">数据来源：</span>
        成功支付取已完成交易；Stripe 争议合并拒付表与标准化拒付事件；EFW 取风控策略事件；退款取已完成退款；
        3DS 取支付启动时的决策记录。3DS 这里表示“系统要求升级认证”，不等于银行挑战已经成功或失败。
      </div>
    </div>

    <section class="grid grid-cols-1 gap-3 lg:grid-cols-2">
      <article
        v-for="provider in providers"
        :key="provider.key"
        class="border border-dashed border-border/80 bg-card p-4"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Provider</p>
            <h3 class="mt-1 text-base font-black uppercase tracking-wide">{{ provider.label }}</h3>
            <p class="mt-1 text-xs text-muted-foreground">
              {{ provider.snapshot ? formatPeriod(provider.snapshot) : '尚未生成风险快照' }}
            </p>
          </div>
          <span :class="levelClass(provider.snapshot?.level)" class="border px-2 py-1 text-[10px] font-black uppercase tracking-wider">
            {{ levelLabel(provider.snapshot?.level) }}
          </span>
        </div>

        <div v-if="loading" class="mt-5 text-sm text-muted-foreground">正在读取风险快照...</div>
        <template v-else-if="provider.snapshot">
          <dl class="mt-4 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70 xl:grid-cols-3">
            <div v-for="metric in provider.metrics" :key="metric.label" class="min-w-0 bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">{{ metric.label }}</dt>
              <dd class="mt-1 truncate font-mono text-base font-semibold">{{ metric.value }}</dd>
              <dd class="mt-1 truncate text-[11px] text-muted-foreground">{{ metric.detail }}</dd>
            </div>
          </dl>

          <div class="mt-4 grid grid-cols-1 gap-3 xl:grid-cols-2">
            <section class="border border-border/70 p-3">
              <h4 class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">计算口径</h4>
              <div class="mt-2 space-y-2 text-xs leading-5">
                <p><span class="font-semibold">争议 / 拒付率</span> = {{ count(provider.snapshot.dispute_count) }} ÷ {{ count(provider.snapshot.successful_payment_count) }} = {{ percent(provider.snapshot.dispute_activity_rate) }}</p>
                <p><span class="font-semibold">早期欺诈预警率</span> = {{ count(provider.snapshot.early_fraud_warning_count) }} ÷ {{ count(provider.snapshot.successful_payment_count) }} = {{ percent(provider.snapshot.early_fraud_warning_rate) }}</p>
                <p><span class="font-semibold">退款率</span> = {{ count(provider.snapshot.refund_count) }} ÷ {{ count(provider.snapshot.successful_payment_count) }} = {{ percent(provider.snapshot.refund_rate) }}</p>
                <p><span class="font-semibold">3DS 要求认证率</span> = {{ count(provider.snapshot.three_ds_upgrade_count) }} ÷ {{ count(provider.snapshot.checkout_attempt_count) }} = {{ percent(provider.snapshot.three_ds_upgrade_rate) }}</p>
              </div>
            </section>

            <section class="border border-border/70 p-3">
              <h4 class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">当前阈值与动作</h4>
              <div class="mt-2 space-y-2 text-xs leading-5">
                <p>争议率：预警 {{ threshold(policy?.warning_dispute_activity_rate) }} / 严重 {{ threshold(policy?.critical_dispute_activity_rate) }}</p>
                <p>EFW：预警 {{ threshold(policy?.warning_early_fraud_rate) }} / 严重 {{ threshold(policy?.critical_early_fraud_rate) }}</p>
                <p>退款率：预警 {{ threshold(policy?.warning_refund_rate) }} / 严重 {{ threshold(policy?.critical_refund_rate) }}</p>
                <p class="font-semibold text-foreground">系统动作：{{ actionLabel(provider.snapshot.recommended_action) }}</p>
              </div>
            </section>
          </div>

          <div class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-dashed border-border/70 pt-3 text-xs">
            <p v-if="provider.reasons.length" class="text-amber-700">
              触发原因：{{ provider.reasons.map(reasonLabel).join(' · ') }}
            </p>
            <p v-else class="text-emerald-700">当前没有触发风险阈值。</p>
            <p class="text-muted-foreground">最后计算：{{ formatDateTime(provider.snapshot.computed_at) }}</p>
          </div>
        </template>
        <div v-else class="mt-5 border border-dashed border-border/80 p-4 text-sm text-muted-foreground">
          {{ enabled ? '等待定时任务或新的风控策略事件生成首次计算。' : '风控策略监控当前未启用。' }}
        </div>
      </article>
    </section>

    <p class="text-[11px] text-muted-foreground">
      注：成功支付金额、争议金额和退款金额按原始币种累计，当前页面不做跨币种换汇；金额仅用于量级观察，比例以笔数为准。
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type {
  RiskStrategyMonitoringPolicy,
  RiskStrategyProviderReport,
  RiskStrategyReports,
  RiskStrategySnapshot,
} from './riskStrategyTypes'

const props = withDefaults(defineProps<{
  reports?: RiskStrategyReports
  policy?: RiskStrategyMonitoringPolicy | null
  enabled?: boolean
  loading?: boolean
}>(), {
  reports: () => ({}),
  policy: null,
  enabled: false,
  loading: false,
})

const windowDays = computed(() => Number(props.policy?.window_days || 30))
const minimumPayments = computed(() => Number(props.policy?.minimum_successful_payments || 20))

const providers = computed(() => ['stripe', 'paypal'].map((key) => {
  const report: RiskStrategyProviderReport | null = props.reports?.[key] || null
  const snapshot = report?.snapshot || null
  return {
    key,
    label: key === 'stripe' ? 'Stripe' : 'PayPal',
    snapshot,
    reasons: Array.isArray(report?.reasons) ? report.reasons : [],
    metrics: snapshot ? [
      {
        label: '成功支付',
        value: count(snapshot.successful_payment_count),
        detail: `金额 ${amount(snapshot.successful_payment_amount)}`,
      },
      {
         label: '争议 / 拒付率',
        value: percent(snapshot.dispute_activity_rate),
        detail: `${count(snapshot.dispute_count)} 笔 · 金额 ${amount(snapshot.dispute_amount)}`,
      },
      {
         label: '早期欺诈预警率',
        value: percent(snapshot.early_fraud_warning_rate),
        detail: `${count(snapshot.early_fraud_warning_count)} 笔早期欺诈预警`,
      },
      {
        label: '退款率',
        value: percent(snapshot.refund_rate),
        detail: `${count(snapshot.refund_count)} 笔 · 金额 ${amount(snapshot.refund_amount)}`,
      },
      {
         label: '3DS 要求认证率',
        value: percent(snapshot.three_ds_upgrade_rate),
        detail: `${count(snapshot.three_ds_upgrade_count)} / ${count(snapshot.checkout_attempt_count)} 次启动`,
      },
      {
         label: '银行挑战次数',
        value: count(snapshot.three_ds_challenge_count),
         detail: `免验证候选 ${count(snapshot.three_ds_exemption_count)} 次（候选，不代表已豁免）`,
      },
    ] : [],
  }
}))

const numeric = (value: unknown): number => {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) ? parsed : 0
}
const count = (value: unknown): string => numeric(value).toLocaleString('zh-CN')
const percent = (value: unknown): string => `${(numeric(value) * 100).toFixed(2)}%`
const threshold = (value: unknown): string => percent(value)
const amount = (value: unknown): string => numeric(value).toLocaleString('zh-CN', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})
const formatDateTime = (value: unknown): string => value ? new Date(value as string | number | Date).toLocaleString('zh-CN') : '-'
const formatPeriod = (snapshot?: RiskStrategySnapshot | null): string => {
  if (!snapshot?.window_start || !snapshot?.window_end) return `${snapshot?.window_days || windowDays.value} 天窗口`
  const start = new Date(snapshot.window_start).toLocaleDateString('zh-CN')
  const end = new Date(snapshot.window_end).toLocaleDateString('zh-CN')
  return `${start} - ${end} · ${snapshot.window_days || windowDays.value} 天窗口`
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
const reasonLabel = (reason: string): string => ({
  dispute_activity_rate_warning: '争议率达到预警阈值',
  dispute_activity_rate_critical: '争议率达到严重阈值',
  early_fraud_warning_rate_warning: 'EFW 达到预警阈值',
  early_fraud_warning_rate_critical: 'EFW 达到严重阈值',
  refund_rate_warning: '退款率达到预警阈值',
  refund_rate_critical: '退款率达到严重阈值',
}[reason] || reason)
const actionLabel = (action?: string): string => ({
  continue_monitoring: '继续监控',
  collect_more_volume_and_review_efw: '样本不足，继续收集数据并重点复核 EFW',
  force_3ds_for_high_risk_and_notify_operator: '高风险支付升级 3DS，并通知管理员',
  force_3ds_for_high_risk_and_operator_review: '高风险支付升级 3DS，并要求人工复核',
}[action || ''] || action || '未指定')
</script>
