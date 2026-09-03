<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="min-w-0">
        <h2 class="text-sm font-black text-foreground">优惠叠加风险</h2>
        <p class="mt-1 text-xs text-muted-foreground">最近计算：{{ generatedAtLabel }}</p>
      </div>
      <Button variant="outline" size="sm" :disabled="loading" @click="emit('refresh')">
        <RefreshCw :class="['size-3.5', loading ? 'animate-spin' : '']" />
        重新计算
      </Button>
    </div>

    <Alert
      v-if="analysisState !== 'ready'"
      :variant="analysisState === 'error' || analysisState === 'invalid' ? 'destructive' : 'default'"
    >
      <LoaderCircle v-if="analysisState === 'loading'" class="size-4 animate-spin" />
      <AlertTriangle v-else-if="analysisState === 'error' || analysisState === 'invalid'" class="size-4" />
      <ShieldCheck v-else class="size-4" />
      <AlertTitle>{{ analysisStateTitle }}</AlertTitle>
      <AlertDescription>{{ analysisStateMessage }}</AlertDescription>
    </Alert>

    <template v-else>
      <AdminStatsGrid compact :items="statItems" />

      <AdminTablePanel :loading="loading">
        <Table class="min-w-[1120px]">
          <TableHeader>
            <TableRow>
              <TableHead class="w-28">风险</TableHead>
              <TableHead class="w-36">场景</TableHead>
              <TableHead class="w-40">优惠规则</TableHead>
              <TableHead class="w-44">会员 / 积分</TableHead>
              <TableHead class="w-48 text-right">危险小计</TableHead>
              <TableHead class="w-52 text-right">估算抵扣</TableHead>
              <TableHead>触发因素</TableHead>
              <TableHead class="w-56">建议</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="riskItems?.length === 0" :colspan="8">
              <div class="flex flex-col items-center text-muted-foreground">
                <ShieldCheck class="mb-2 size-7 opacity-55" />
                <span class="text-xs">当前规则未发现低实付风险</span>
              </div>
            </TableEmpty>
            <TableRow v-for="item in riskItems" :key="riskItemKey(item)">
              <TableCell>
                <AdminStatusBadge :tone="severityTone(item.severity)">
                  {{ severityLabel(item.severity) }}
                </AdminStatusBadge>
              </TableCell>
              <TableCell class="text-xs font-medium">{{ scenarioLabel(item.scenario) }}</TableCell>
              <TableCell>
                <div class="space-y-1">
                  <div class="font-mono text-xs font-bold">{{ item.coupon_code || '无优惠券' }}</div>
                  <div class="text-[11px] text-muted-foreground">
                    {{ couponTypeLabel(item.coupon_type) }}
                    <span v-if="item.coupon_status"> · {{ couponStatusLabel(item.coupon_status) }}</span>
                  </div>
                </div>
              </TableCell>
              <TableCell class="text-xs text-muted-foreground">
                <div>{{ item.member_level_name || '-' }} · {{ formatRate(item.member_discount_rate) }}</div>
                <div>积分直抵 · {{ formatRate(item.points_discount_rate) }}</div>
              </TableCell>
              <TableCell class="text-right text-xs tabular-nums">
                <div>{{ thresholdLabel(item) }}</div>
                <div class="text-[11px] text-muted-foreground">样本 {{ money(item.estimated_subtotal) }}</div>
              </TableCell>
              <TableCell class="text-right text-xs tabular-nums">
                <div>{{ money(item.estimated_discount_amount) }}</div>
                <div class="text-[11px] text-muted-foreground">实付 {{ money(item.estimated_payable_amount) }}</div>
              </TableCell>
              <TableCell>
                <div class="flex flex-wrap gap-1">
                  <AdminStatusBadge
                    v-for="factor in item.factors"
                    :key="factor"
                    tone="gray"
                  >
                    {{ factorLabel(factor) }}
                  </AdminStatusBadge>
                </div>
              </TableCell>
              <TableCell class="text-xs text-muted-foreground">{{ recommendationLabel(item.kind) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </AdminTablePanel>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, BadgePercent, Coins, Crown, LoaderCircle, RefreshCw, ShieldCheck } from '@lucide/vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type { PromotionRiskAnalysis, PromotionRiskItem, PromotionRiskSummary } from '@/modules/marketing/marketingTypes'

const props = withDefaults(defineProps<{
  loading?: boolean
  analysis?: PromotionRiskAnalysis | null
  error?: string | null
  formatCurrency: (value: unknown, currency?: string) => string
  formatDate: (value: unknown) => string
  formatRate: (value: unknown) => string
}>(), {
  loading: false,
  analysis: null,
  error: null,
})

const emit = defineEmits<{
  (event: 'refresh'): void
}>()

type PromotionRiskAnalysisState = 'idle' | 'loading' | 'error' | 'invalid' | 'ready'

const isValidSummary = (summary: unknown): summary is PromotionRiskSummary => {
  if (!summary || typeof summary !== 'object') return false
  const candidate = summary as Record<string, unknown>
  return typeof candidate.severity === 'string'
    && typeof candidate.candidate_coupon_count === 'number'
    && typeof candidate.risk_item_count === 'number'
    && typeof candidate.zero_total_risk_count === 'number'
    && typeof candidate.gateway_minimum_risk_count === 'number'
    && typeof candidate.member_level_count === 'number'
    && typeof candidate.max_member_discount_rate === 'number'
    && typeof candidate.max_member_discount_level_name === 'string'
    && typeof candidate.points_redemption_enabled === 'boolean'
    && typeof candidate.direct_points_discount_cap_rate === 'number'
    && typeof candidate.max_redeem_gift_card_value === 'number'
}

const isValidRiskItem = (item: unknown): item is PromotionRiskItem => {
  if (!item || typeof item !== 'object') return false
  const candidate = item as Record<string, unknown>
  return typeof candidate.severity === 'string'
    && typeof candidate.kind === 'string'
    && typeof candidate.scenario === 'string'
    && Array.isArray(candidate.factors)
}

const isValidAnalysis = (analysis: PromotionRiskAnalysis | null | undefined) =>
  Boolean(
    analysis
    && typeof analysis.generated_at === 'string'
    && analysis.generated_at.length > 0
    && typeof analysis.currency === 'string'
    && analysis.currency.length > 0
    && isValidSummary(analysis.summary)
    && Array.isArray(analysis.items)
    && analysis.items.every(isValidRiskItem),
  )

const analysisReady = computed(() => isValidAnalysis(props.analysis))
const analysisState = computed<PromotionRiskAnalysisState>(() => {
  if (props.error) return 'error'
  if (analysisReady.value) return 'ready'
  if (props.loading) return 'loading'
  if (props.analysis) return 'invalid'
  return 'idle'
})
const analysisStateTitle = computed(() => {
  if (analysisState.value === 'loading') return '正在加载优惠风险分析'
  if (analysisState.value === 'error') return '优惠风险分析加载失败'
  if (analysisState.value === 'invalid') return '优惠风险分析数据异常'
  if (analysisState.value === 'idle') return '优惠风险分析尚未加载'
  return ''
})
const analysisStateMessage = computed(() => {
  if (analysisState.value === 'loading') return '分析结果尚未就绪，请稍候。'
  if (analysisState.value === 'error') return props.error
  if (analysisState.value === 'invalid') return '接口返回缺少必需字段，已停止渲染默认结果。'
  if (analysisState.value === 'idle') return '当前没有可展示的分析结果，请点击“重新计算”。'
  return ''
})
const summary = computed(() => props.analysis?.summary)
const currency = computed(() => props.analysis?.currency)
const riskItems = computed(() => props.analysis?.items)
const generatedAtLabel = computed(() => props.analysis?.generated_at ? props.formatDate(props.analysis.generated_at) : '-')
const money = (value: unknown) => props.formatCurrency(value ?? 0, currency.value)

const statItems = computed(() => [
  {
    key: 'severity',
    label: '风险等级',
    value: severityLabel(summary.value?.severity),
    icon: summary.value?.severity === 'critical' ? AlertTriangle : ShieldCheck,
    tone: summaryTone(summary.value?.severity),
  },
  {
    key: 'zero',
    label: '零元风险',
    value: Number(summary.value?.zero_total_risk_count ?? 0),
    icon: AlertTriangle,
    tone: Number(summary.value?.zero_total_risk_count ?? 0) > 0 ? 'coral' : 'gray',
  },
  {
    key: 'gateway',
    label: '低于网关',
    value: Number(summary.value?.gateway_minimum_risk_count ?? 0),
    icon: BadgePercent,
    tone: Number(summary.value?.gateway_minimum_risk_count ?? 0) > 0 ? 'amber' : 'gray',
  },
  {
    key: 'coupons',
    label: '候选券',
    value: Number(summary.value?.candidate_coupon_count ?? 0),
    icon: BadgePercent,
    tone: 'blue',
  },
  {
    key: 'member',
    label: '最高会员折扣',
    value: props.formatRate(summary.value?.max_member_discount_rate ?? 0),
    icon: Crown,
    tone: Number(summary.value?.max_member_discount_rate ?? 0) > 0 ? 'green' : 'gray',
  },
  {
    key: 'points',
    label: '积分直抵上限',
    value: props.formatRate(summary.value?.direct_points_discount_cap_rate ?? 0),
    icon: Coins,
    tone: summary.value?.points_redemption_enabled ? 'amber' : 'gray',
  },
])

const summaryTone = (severity?: string): string => {
  if (severity === 'critical') return 'coral'
  if (severity === 'warning') return 'amber'
  return 'green'
}

const severityTone = (severity?: string): AdminStatusTone => {
  if (severity === 'critical') return 'coral'
  if (severity === 'warning') return 'amber'
  return 'green'
}

const severityLabel = (severity?: string) => {
  if (severity === 'critical') return '高危'
  if (severity === 'warning') return '预警'
  return '正常'
}

const couponTypeLabel = (type?: string) => {
  if (type === 'fixed') return '固定金额'
  if (type === 'percentage') return '百分比'
  return '组合规则'
}

const couponStatusLabel = (status?: string) => {
  if (status === 'active') return '生效中'
  if (status === 'scheduled') return '未开始'
  if (status === 'expired') return '已过期'
  if (status === 'disabled') return '已停用'
  return status || '-'
}

const scenarioLabel = (scenario?: string) => {
  if (scenario === 'member_points') return '会员 + 积分'
  if (scenario === 'coupon_member_points_before_cap') return '券封顶前叠加'
  if (scenario === 'coupon_member_points_after_cap') return '券封顶后叠加'
  return '券 + 会员 + 积分'
}

const factorLabel = (factor: string) => {
  if (factor === 'fixed_coupon') return '固定券'
  if (factor === 'percentage_coupon') return '百分比券'
  if (factor === 'member_level_discount') return '会员折扣'
  if (factor === 'direct_points_discount') return '积分直抵'
  return factor
}

const recommendationLabel = (kind?: string) => {
  if (kind === 'zero_total') return '提高最低消费、设置折扣封顶，或保留零元订单内部结算。'
  if (kind === 'below_gateway_minimum') return '提高最低消费，确保实付金额高于支付网关最低金额。'
  return '复核优惠叠加边界。'
}

const thresholdLabel = (item: PromotionRiskItem) => {
  const threshold = item.kind === 'zero_total'
    ? item.full_cover_subtotal_threshold
    : item.gateway_minimum_threshold
  if (!threshold || threshold <= 0) return '全部满足门槛的小计'
  return `≤ ${money(threshold)}`
}

const riskItemKey = (item: PromotionRiskItem) => [
  item.kind,
  item.scenario,
  item.coupon_id || 'none',
  item.full_cover_subtotal_threshold || item.gateway_minimum_threshold || 0,
].join(':')
</script>
