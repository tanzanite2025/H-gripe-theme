<template>
  <section class="space-y-3">
    <div class="border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">3DS Policy Center</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">3DS 策略与执行解释</h2>
          <p class="mt-1 max-w-4xl text-xs leading-5 text-muted-foreground">
            这页把“支付设置里的基础模式”和“风控运行时的动态升级”合并展示。你可以直接看出当前每一步会做什么、哪些值能修改，以及 3DS 统计到底在数什么。
          </p>
        </div>
        <div class="flex flex-wrap gap-2 text-[10px] font-black uppercase tracking-wider">
          <span :class="badgeClass(Boolean(stripeRuntime?.production_ready))" class="border px-2 py-1">
            Stripe {{ stripeRuntime?.production_ready ? '生产就绪' : '未达到生产就绪' }}
          </span>
          <span :class="badgeClass(Boolean(threeDS.adaptive_enabled))" class="border px-2 py-1">
            自适应 3DS {{ threeDS.adaptive_enabled ? '已开启' : '已关闭' }}
          </span>
        </div>
      </div>

      <div class="mt-4 grid grid-cols-1 gap-px overflow-hidden border border-border/70 bg-border/70 sm:grid-cols-2 xl:grid-cols-4">
        <div class="bg-card p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">基础模式</p>
          <p class="mt-1 text-sm font-black">{{ modeLabel(stripeMode) }}</p>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">PaymentIntent 创建时的起点</p>
        </div>
        <div class="bg-card p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Step-up</p>
          <p class="mt-1 font-mono text-sm font-black">≥ {{ number(threeDS.step_up_risk_score) }} 分</p>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">至少升到 any</p>
        </div>
        <div class="bg-card p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Challenge</p>
          <p class="mt-1 font-mono text-sm font-black">≥ {{ number(threeDS.challenge_risk_score) }} 分</p>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">请求更强挑战路径</p>
        </div>
        <div class="bg-card p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">人工保护</p>
 <p class="mt-1 text-sm font-black">{{ protection.enabled ? '可临时强制': '未启用'}}</p>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">强制 3DS 至少 any</p>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 xl:grid-cols-[minmax(0,1.2fr)_minmax(360px,0.8fr)]">
      <section class="border border-dashed border-border/80 bg-card p-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Runtime Decision Flow</p>
            <h3 class="mt-1 text-base font-black tracking-tight">一笔 Stripe 支付实际怎么走</h3>
          </div>
 <span :class="stripeHealth.circuit_open ? 'border-rose-500/25 bg-rose-500/10 text-rose-700': 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'" class="border px-2 py-1 text-[10px] font-black">
            {{ stripeHealth.circuit_open ? '网关熔断中' : '网关允许调用' }}
          </span>
        </div>

        <ol class="mt-4 space-y-2">
          <li v-for="step in decisionSteps" :key="step.index" class="flex gap-3 border border-border/70 p-3">
            <span class="flex size-7 flex-none items-center justify-center rounded-full bg-primary text-[11px] font-black text-primary-foreground">{{ step.index }}</span>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h4 class="text-xs font-black">{{ step.title }}</h4>
                <span class="font-mono text-[10px] text-muted-foreground">{{ step.code }}</span>
              </div>
              <p class="mt-1 text-[11px] leading-5 text-muted-foreground">{{ step.description }}</p>
            </div>
          </li>
        </ol>
      </section>

      <section class="border border-dashed border-border/80 bg-card p-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Effective Settings</p>
            <h3 class="mt-1 text-base font-black tracking-tight">当前生效设置</h3>
          </div>
          <span class="border border-sky-500/25 bg-sky-500/10 px-2 py-1 text-[10px] font-black text-sky-700">
            运行源：{{ runtimeSourceLabel(stripeRuntime?.runtime_source) }}
          </span>
        </div>

        <dl class="mt-4 divide-y divide-dashed divide-border/70 border-y border-dashed border-border/70 text-xs">
          <div class="grid grid-cols-[minmax(0,1fr)_auto] gap-3 py-3">
            <div>
              <dt class="font-semibold">Stripe 基础 3DS 模式</dt>
              <dd class="mt-1 text-[11px] leading-5 text-muted-foreground">可在支付设置中修改；这里显示真实运行值。</dd>
            </div>
            <dd class="text-right font-mono font-semibold">{{ modeLabel(stripeMode) }}</dd>
          </div>
          <div class="grid grid-cols-[minmax(0,1fr)_auto] gap-3 py-3">
            <div>
              <dt class="font-semibold">自适应 3DS</dt>
              <dd class="mt-1 text-[11px] leading-5 text-muted-foreground">根据单笔风险、访客画像和历史支付动态升级。</dd>
            </div>
 <dd class="text-right font-semibold">{{ threeDS.adaptive_enabled ? '开启': '关闭'}}</dd>
          </div>
          <div class="grid grid-cols-[minmax(0,1fr)_auto] gap-3 py-3">
            <div>
              <dt class="font-semibold">30 天组合风险升级</dt>
              <dd class="mt-1 text-[11px] leading-5 text-muted-foreground">争议、EFW 或退款进入预警/严重级别时生效。</dd>
            </div>
 <dd class="text-right font-semibold">{{ monitoring.auto_step_up_enabled ? '开启': '关闭'}}</dd>
          </div>
          <div class="grid grid-cols-[minmax(0,1fr)_auto] gap-3 py-3">
            <div>
              <dt class="font-semibold">低风险豁免候选</dt>
              <dd class="mt-1 text-[11px] leading-5 text-muted-foreground">低金额或老客条件只记录为候选，不保证银行免验证。</dd>
            </div>
            <dd class="text-right font-mono font-semibold">
              {{ lowRiskMaxAmount > 0 ? `≤ ${lowRiskMaxAmount.toFixed(2)}` : '未设金额' }}
            </dd>
          </div>
          <div class="grid grid-cols-[minmax(0,1fr)_auto] gap-3 py-3">
            <div>
              <dt class="font-semibold">人工保护动作</dt>
              <dd class="mt-1 text-[11px] leading-5 text-muted-foreground">强制 3DS 会把新支付至少提高到 any；暂停支付会在创建 PaymentIntent 前拦截。</dd>
            </div>
 <dd class="text-right font-semibold">{{ protection.enabled ? '可用': '不可用'}}</dd>
          </div>
        </dl>

        <div class="mt-4 border border-sky-500/20 bg-sky-500/5 p-3 text-[11px] leading-5 text-sky-900 dark:text-sky-100">
          <p class="font-black">可修改位置</p>
          <p class="mt-1">基础模式：支付设置 → Stripe → 基础 3DS 认证策略；人工保护：本页“人工保护”；Step-up、Challenge 和 30 天阈值：当前由服务端运行配置提供，页面只读展示。</p>
        </div>

        <Button v-if="canViewPaymentSettings" type="button" variant="outline" class="mt-3 w-full rounded-full font-black" @click="goToPaymentSettings">
          <Settings2 class="size-4" />
          去支付设置修改 Stripe 基础模式
        </Button>
      </section>
    </div>

    <section class="border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Mode Comparison</p>
          <h3 class="mt-1 text-base font-black tracking-tight">三个模式分别代表什么</h3>
        </div>
        <p class="max-w-xl text-[11px] leading-5 text-muted-foreground">
          这是“请求强度”的比较，不是银行最终结果的保证。发卡行、卡片能力和 SCA 规则仍会决定最终是无感完成、challenge、失败还是拒绝。
        </p>
      </div>

      <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-3">
        <article
          v-for="mode in modeCards"
          :key="mode.value"
          class="border p-3"
 :class="stripeMode === mode.value ? 'border-primary/50 bg-primary/5': 'border-border/70'"
        >
          <div class="flex items-start justify-between gap-2">
            <h4 class="text-xs font-black">{{ mode.title }}</h4>
            <code class="font-mono text-[10px] text-muted-foreground">{{ mode.value }}</code>
          </div>
          <p class="mt-2 text-[11px] leading-5 text-muted-foreground">{{ mode.description }}</p>
          <dl class="mt-3 space-y-2 text-[11px]">
            <div class="flex items-start justify-between gap-3 border-t border-dashed border-border/70 pt-2">
              <dt class="text-muted-foreground">系统请求</dt>
              <dd class="text-right font-semibold">{{ mode.request }}</dd>
            </div>
            <div class="flex items-start justify-between gap-3 border-t border-dashed border-border/70 pt-2">
              <dt class="text-muted-foreground">常见表现</dt>
              <dd class="max-w-[190px] text-right font-semibold">{{ mode.outcome }}</dd>
            </div>
          </dl>
        </article>
      </div>
    </section>

    <section class="grid grid-cols-1 gap-3 xl:grid-cols-[minmax(0,1fr)_minmax(360px,0.8fr)]">
      <section class="border border-dashed border-border/80 bg-card p-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">3DS Observability</p>
            <h3 class="mt-1 text-base font-black tracking-tight">最近 30 天实际记录</h3>
          </div>
 <span class="text-[11px] text-muted-foreground">{{ stripeReport?.snapshot ? formatPeriod(stripeReport.snapshot) : '尚未生成快照'}}</span>
        </div>

        <template v-if="stripeReport?.snapshot">
          <dl class="mt-4 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70 sm:grid-cols-4">
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">支付启动</dt>
              <dd class="mt-1 font-mono text-base font-semibold">{{ count(stripeReport.snapshot.checkout_attempt_count) }}</dd>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">要求 3DS</dt>
              <dd class="mt-1 font-mono text-base font-semibold">{{ count(stripeReport.snapshot.three_ds_upgrade_count) }}</dd>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">升级率</dt>
              <dd class="mt-1 font-mono text-base font-semibold">{{ percent(stripeReport.snapshot.three_ds_upgrade_rate) }}</dd>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">challenge 模式</dt>
              <dd class="mt-1 font-mono text-base font-semibold">{{ count(stripeReport.snapshot.three_ds_challenge_count) }}</dd>
            </div>
          </dl>
          <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2">
            <div class="border border-border/70 p-3 text-xs leading-5">
              <p class="font-black">页面这个百分比怎么算</p>
              <p class="mt-1 text-muted-foreground">
                3DS 要求认证率 = {{ count(stripeReport.snapshot.three_ds_upgrade_count) }} ÷ {{ count(stripeReport.snapshot.checkout_attempt_count) }} = {{ percent(stripeReport.snapshot.three_ds_upgrade_rate) }}。
              </p>
            </div>
            <div class="border border-border/70 p-3 text-xs leading-5">
              <p class="font-black">不要把它当成什么</p>
              <p class="mt-1 text-muted-foreground">
                这个数字表示系统记录了多少次升级请求，不代表银行实际弹出了多少次 challenge，也不代表认证成功率。
              </p>
            </div>
          </div>
        </template>
        <p v-else class="mt-4 border border-dashed border-border/80 p-4 text-sm text-muted-foreground">
          {{ enabled ? '还没有可展示的 Stripe 3DS 决策快照；有新的支付启动并完成记录后会出现在这里。' : '风控策略监控未启用。' }}
        </p>
      </section>

      <section class="border border-dashed border-border/80 bg-card p-4">
        <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Important Distinctions</p>
        <h3 class="mt-1 text-base font-black tracking-tight">几个容易混淆的点</h3>
        <div class="mt-4 space-y-3 text-xs leading-5">
          <div class="border-l-2 border-sky-500/50 pl-3">
            <p class="font-black">要求 3DS ≠ 一定弹窗</p>
            <p class="mt-1 text-muted-foreground">`any` 可以无感完成，`challenge` 也只是请求更强路径，最终交互由发卡行决定。</p>
          </div>
          <div class="border-l-2 border-amber-500/50 pl-3">
            <p class="font-black">站点挑战 ≠ 银行 3DS</p>
            <p class="mt-1 text-muted-foreground">Turnstile 是支付启动前的反滥用验证；3DS 是 PaymentIntent 认证流程，两者不是同一个拦截。</p>
          </div>
          <div class="border-l-2 border-rose-500/50 pl-3">
            <p class="font-black">人工强制 ≠ 自动退款</p>
            <p class="mt-1 text-muted-foreground">人工保护只影响匹配范围内的新支付；EFW 当前进入风险事件和退款建议链路，不会在 webhook 里直接退款。</p>
          </div>
        </div>
      </section>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Settings2 } from '@lucide/vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import type {
  RiskStrategyGatewayHealthMap,
  RiskStrategyGatewayRuntime,
  RiskStrategyGatewayRuntimeStatus,
  RiskStrategyConfiguration,
  RiskStrategyProviderReport,
  RiskStrategyReports,
  RiskStrategySnapshot,
} from './riskStrategyTypes'

const props = withDefaults(defineProps<{
  configuration?: RiskStrategyConfiguration | null
  gatewayRuntime?: RiskStrategyGatewayRuntime | null
  gatewayHealth?: RiskStrategyGatewayHealthMap
  reports?: RiskStrategyReports
  enabled?: boolean
}>(), {
  configuration: null,
  gatewayRuntime: null,
  gatewayHealth: () => ({}),
  reports: () => ({}),
  enabled: false,
})

const router = useRouter()
const authStore = useAuthStore()
const configuration = computed(() => props.configuration || {})
const threeDS = computed(() => configuration.value.three_ds || {})
const monitoring = computed(() => configuration.value.monitoring || {})
const protection = computed(() => configuration.value.protection || {})
const lowRiskMaxAmount = computed(() => Number(threeDS.value.low_risk_max_amount || 0))
const stripeRuntime = computed<RiskStrategyGatewayRuntimeStatus | null>(() => (
  props.gatewayRuntime?.gateways?.find((gateway) => gateway.provider === 'stripe') || null
))
const stripeHealth = computed(() => props.gatewayHealth?.stripe || {})
const stripeMode = computed(() => stripeRuntime.value?.three_ds_mode || 'automatic')
const stripeReport = computed<RiskStrategyProviderReport | null>(() => props.reports?.stripe || null)
const canViewPaymentSettings = computed(() => authStore.hasPermission('settings:view'))

const decisionSteps = [
  {
    index: 1,
    code: 'preflight',
    title: '支付启动前预检',
    description: '先检查订单状态、人工暂停、重复失败、Turnstile、BIN 限流和网关熔断；命中时可能要求站点验证、延迟或直接不创建 PaymentIntent。',
  },
  {
    index: 2,
    code: 'base_mode',
    title: '读取 Stripe 基础模式',
    description: `当前是 ${modeLabel(stripeMode.value)}。它只是每笔支付的起点，保存位置在支付设置的 Stripe 加密配置中。`,
  },
  {
    index: 3,
    code: 'risk_merge',
    title: '合并动态风险信号',
    description: `叠加人工保护、30 天组合风险、单笔风控、访客画像和客户历史；模式只会向更强方向提升，当前 Step-up 阈值为 ${number(threeDS.value.step_up_risk_score)} 分，Challenge 阈值为 ${number(threeDS.value.challenge_risk_score)} 分。`,
  },
  {
    index: 4,
    code: 'payment_intent',
    title: '使用最终模式创建 PaymentIntent',
    description: '系统记录最终模式、策略、风险分数、组合风险级别、原因和豁免候选；这就是本页 3DS 统计的来源。',
  },
]

const modeCards = [
  {
    value: 'automatic',
    title: '自动判断',
    description: '把基础 SCA/3DS 判断交给 Stripe，并允许本系统在风险升高时继续升级。',
    request: '不主动强制',
    outcome: '可能直接完成、无感认证或被动态升级',
  },
  {
    value: 'any',
    title: '要求 3DS',
    description: '请求在支持时使用 3DS；认证可以无感完成，也可能需要买家完成银行验证。',
    request: '至少要求认证路径',
    outcome: '更可能进入 3DS，但不等于一定弹窗',
  },
  {
    value: 'challenge',
    title: '更强挑战',
    description: '请求发卡行进入更强的挑战路径，适合高风险基础策略或高风险订单。',
    request: '请求 challenge 路径',
    outcome: '更可能出现银行挑战，最终仍由发卡行决定',
  },
]

function number(value: unknown): string {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) ? parsed.toLocaleString('zh-CN') : '0'
}
const count = (value: unknown): string => number(value)
const percent = (value: unknown): string => `${(Number(value || 0) * 100).toFixed(2)}%`
function modeLabel(mode: string): string {
  return ({
  automatic: '自动判断（automatic）',
  any: '要求 3DS（any）',
  challenge: '更强挑战（challenge）',
  }[mode] || mode || '未读取')
}
const badgeClass = (enabled: boolean): string => (
  enabled
    ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'
    : 'border-amber-500/25 bg-amber-500/10 text-amber-700'
)
const runtimeSourceLabel = (source?: string): string => ({
  environment: '环境变量',
  'admin-encrypted': '后台加密配置',
  mixed: '混合来源',
}[source || ''] || source || '未知')
const formatPeriod = (snapshot?: RiskStrategySnapshot | null): string => {
  if (!snapshot?.window_start || !snapshot?.window_end) return `${snapshot?.window_days || 30} 天窗口`
  return `${new Date(snapshot.window_start).toLocaleDateString('zh-CN')} - ${new Date(snapshot.window_end).toLocaleDateString('zh-CN')} · ${snapshot.window_days || 30} 天窗口`
}
const goToPaymentSettings = (): void => {
  router.push({ name: 'PaymentSettings' })
}
</script>
