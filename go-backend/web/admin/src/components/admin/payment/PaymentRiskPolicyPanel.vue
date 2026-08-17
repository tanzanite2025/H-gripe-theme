<template>
  <section class="space-y-3">
    <div class="border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Effective Payment Policy</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">当前生效策略</h2>
          <p class="mt-1 max-w-4xl text-xs leading-5 text-muted-foreground">
            这里展示的是代码实际会采用的基础模式和触发条件。人工保护只会增加限制，不会把高风险订单降级成更弱的认证。
          </p>
        </div>
        <div class="flex flex-wrap gap-2 text-[10px] font-black uppercase tracking-wider">
          <span :class="statusClass(Boolean(threeDS.adaptive_enabled))" class="border px-2 py-1">
            自适应 3DS {{ threeDS.adaptive_enabled ? '已开启' : '已关闭' }}
          </span>
          <span :class="statusClass(Boolean(protection.enabled))" class="border px-2 py-1">
            人工保护 {{ protection.enabled ? '可用' : '未启用' }}
          </span>
        </div>
      </div>

      <div class="mt-4 grid grid-cols-1 gap-3 xl:grid-cols-2">
        <section class="border border-border/70 p-3">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h3 class="text-xs font-black uppercase tracking-widest">Stripe 基础 3DS 模式</h3>
              <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
                这是创建 PaymentIntent 时的起点；自适应策略只能把认证强度往上推。
              </p>
            </div>
            <span class="border border-sky-500/25 bg-sky-500/10 px-2 py-1 font-mono text-[10px] font-black text-sky-700">
              {{ modeLabel(stripeMode) }}
            </span>
          </div>
          <dl class="mt-3 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70">
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">实际值</dt>
              <dd class="mt-1 text-sm font-semibold">{{ modeLabel(stripeMode) }}</dd>
 <p class="mt-1 font-mono text-[11px] text-muted-foreground">{{ stripeMode || '未读取'}}</p>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">运行环境</dt>
 <dd class="mt-1 text-sm font-semibold">{{ stripeRuntime?.environment || '未读取'}}</dd>
            </div>
          </dl>
          <p class="mt-3 text-xs leading-5 text-muted-foreground">{{ modeDescription(stripeMode) }}</p>
          <p class="mt-2 text-[11px] leading-5 text-muted-foreground">
            自适应风控：{{ threeDS.enabled ? '配置已开启' : '配置未开启' }} · 策略服务：{{ threeDS.runtime_available ? '后端已挂载' : '后端未挂载' }}
          </p>
          <p v-if="stripeRuntime" class="mt-2 text-[11px] text-muted-foreground">
            配置来源：{{ runtimeSourceLabel(stripeRuntime.runtime_source) }} · 配置状态：{{ stripeRuntime.production_ready ? '生产就绪' : '未达到生产就绪' }}
          </p>
        </section>

        <section class="border border-border/70 p-3">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h3 class="text-xs font-black uppercase tracking-widest">自适应升级条件</h3>
              <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
                每次 Stripe 支付启动时，系统会重新评估这些条件，并记录最终策略和原因。
              </p>
            </div>
            <span class="border border-emerald-500/25 bg-emerald-500/10 px-2 py-1 text-[10px] font-black text-emerald-700">
              {{ threeDS.runtime_available ? '后端已挂载' : '后端不可用' }}
            </span>
          </div>
          <dl class="mt-3 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70">
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Step-up 分数</dt>
              <dd class="mt-1 font-mono text-sm font-semibold">≥ {{ number(threeDS.step_up_risk_score) }}</dd>
              <p class="mt-1 text-[11px] text-muted-foreground">升级为 any</p>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Challenge 分数</dt>
              <dd class="mt-1 font-mono text-sm font-semibold">≥ {{ number(threeDS.challenge_risk_score) }}</dd>
              <p class="mt-1 text-[11px] text-muted-foreground">升级为 challenge</p>
            </div>
          </dl>
          <p class="mt-3 text-xs leading-5 text-muted-foreground">
            分数来自支付风险与访客风险评估；一旦命中高风险，最终模式不会被低金额或老客条件覆盖。
          </p>
        </section>
      </div>

      <section class="mt-3 border border-border/70 p-3">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h3 class="text-xs font-black uppercase tracking-widest">3DS 基础策略对照</h3>
            <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
              设置页只保存基础策略；支付创建时，人工保护和自适应风控可以把它向更强方向提升。
            </p>
          </div>
          <span class="text-[11px] font-semibold text-foreground">当前：{{ modeLabel(stripeMode) }}</span>
        </div>
        <div class="mt-3 grid grid-cols-1 gap-2 md:grid-cols-3">
          <article
            v-for="mode in threeDSModeCards"
            :key="mode.value"
            class="border p-3"
 :class="stripeMode === mode.value ? 'border-primary/50 bg-primary/5': 'border-border/70'"
          >
            <div class="flex items-start justify-between gap-2">
              <h4 class="text-xs font-black">{{ mode.label }}</h4>
              <code class="font-mono text-[10px] text-muted-foreground">{{ mode.value }}</code>
            </div>
            <p class="mt-2 text-[11px] leading-5 text-muted-foreground">{{ mode.description }}</p>
            <p class="mt-2 text-[11px] leading-5">
              <span class="font-semibold">最终表现：</span>{{ mode.outcome }}
            </p>
          </article>
        </div>
      </section>
    </div>

    <section class="border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Runtime &amp; Circuit State</p>
          <h2 class="mt-1 text-base font-black tracking-tight">支付渠道现在能不能收款</h2>
          <p class="mt-1 max-w-4xl text-xs leading-5 text-muted-foreground">
            这里是运行时状态，不是配置文件里的理论值。熔断只统计最近窗口内的网关调用失败；它和拒付率、退款率是两套不同的指标。
          </p>
        </div>
        <span class="border border-sky-500/25 bg-sky-500/10 px-2 py-1 text-[10px] font-black text-sky-700">
          运行源：{{ runtimeSourceLabel(gatewayRuntime?.runtime_source) }}
        </span>
      </div>

      <div v-if="gatewayCards.length" class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2">
        <article v-for="gateway in gatewayCards" :key="gateway.provider" class="border border-border/70 p-3">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-xs font-black uppercase tracking-widest">{{ gateway.label }}</p>
              <p class="mt-1 text-[11px] text-muted-foreground">
                {{ gateway.runtime?.environment || '未读取环境' }} · {{ runtimeSourceLabel(gateway.runtime?.runtime_source) }}
              </p>
            </div>
            <span :class="gatewayStatusClass(gateway)" class="border px-2 py-1 text-[10px] font-black">
              {{ gatewayStatusLabel(gateway) }}
            </span>
          </div>
          <dl class="mt-3 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70">
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">网关调用</dt>
 <dd class="mt-1 text-sm font-semibold">{{ gateway.health?.allowed === false ? '当前拦截': '允许调用'}}</dd>
              <p class="mt-1 text-[11px] text-muted-foreground">
                {{ number(gateway.health?.failure_count) }} / {{ number(gateway.health?.sample_count) }} 次失败
              </p>
            </div>
            <div class="bg-card p-3">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">失败率</dt>
              <dd class="mt-1 font-mono text-sm font-semibold">{{ percent(gateway.health?.failure_rate) }}</dd>
              <p class="mt-1 text-[11px] text-muted-foreground">
                {{ gateway.health?.retry_after_seconds ? `恢复倒计时 ${seconds(gateway.health.retry_after_seconds)}` : '未进入熔断倒计时' }}
              </p>
            </div>
          </dl>
          <p v-if="gateway.provider === 'stripe'" class="mt-3 text-xs leading-5 text-muted-foreground">
            实际基础 3DS：<span class="font-mono font-semibold text-foreground">{{ modeLabel(gateway.runtime?.three_ds_mode || 'automatic') }}</span>。
            这只是 PaymentIntent 的起点，最终模式还会被人工保护、组合风险、单笔风险和访客风险向上提升。
          </p>
          <p v-else class="mt-3 text-xs leading-5 text-muted-foreground">
            PayPal 当前参与统一风险快照与网关健康监控；本页的自适应 3DS 决策只在 Stripe 支付创建链路执行。
          </p>
          <p v-if="gateway.runtime?.missing?.length" class="mt-2 text-[11px] text-amber-700">
            运行配置缺失：{{ gateway.runtime.missing.join('、') }}
          </p>
          <p v-if="gateway.runtime?.blockers?.length" class="mt-2 text-[11px] text-rose-700">
            生产阻断：{{ gateway.runtime.blockers.join('、') }}
          </p>
          <p v-else-if="gateway.runtime?.warnings?.length" class="mt-2 text-[11px] text-amber-700">
            运行提醒：{{ gateway.runtime.warnings.join('、') }}
          </p>
        </article>
      </div>
      <div v-else class="mt-4 border border-dashed border-border/80 p-4 text-sm text-muted-foreground">
        尚未读取到支付网关运行状态。
      </div>
    </section>

    <div class="grid grid-cols-1 gap-3 xl:grid-cols-3">
      <section class="border border-dashed border-border/80 bg-card p-4">
        <h3 class="text-xs font-black uppercase tracking-widest">豁免候选</h3>
        <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
          这只是记录为候选，不等于 Stripe 一定免验证，也不代表订单已经通过。
        </p>
        <dl class="mt-3 space-y-2 text-xs">
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">低金额上限</dt>
            <dd class="font-mono font-semibold">{{ money(threeDS.low_risk_max_amount) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">老客已支付订单</dt>
            <dd class="font-mono font-semibold">≥ {{ number(threeDS.trusted_paid_orders) }} 笔</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="text-muted-foreground">访客观察窗口</dt>
            <dd class="font-mono font-semibold">{{ number(threeDS.visitor_risk_lookback_days) }} 天</dd>
          </div>
        </dl>
      </section>

      <section class="border border-dashed border-border/80 bg-card p-4">
        <h3 class="text-xs font-black uppercase tracking-widest">支付风险预检</h3>
        <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
          发生在创建 Stripe PaymentIntent 之前；它可以延迟、要求站点挑战，或直接阻止支付启动。
        </p>
        <dl class="mt-3 space-y-2 text-xs">
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">状态</dt>
 <dd class="font-semibold">{{ paymentRisk.enabled ? '已开启': '未开启'}}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">失败计数窗口</dt>
            <dd class="font-mono font-semibold">{{ seconds(paymentRisk.failure_window_seconds) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">重复失败阈值</dt>
            <dd class="font-mono font-semibold">{{ number(paymentRisk.failure_threshold) }} 次</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">高风险分数</dt>
            <dd class="font-mono font-semibold">≥ {{ number(paymentRisk.high_risk_score) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="text-muted-foreground">命中后延迟</dt>
            <dd class="font-mono font-semibold">{{ seconds(paymentRisk.delay_seconds) }}</dd>
          </div>
        </dl>
      </section>

      <section class="border border-dashed border-border/80 bg-card p-4">
        <h3 class="text-xs font-black uppercase tracking-widest">卡测与网关保护</h3>
        <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
          这些规则不改变 3DS 模式，而是在支付尝试前或网关调用层阻止异常流量。
        </p>
        <dl class="mt-3 space-y-2 text-xs">
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">BIN 限流</dt>
 <dd class="font-semibold">{{ binRateLimit.enabled ? '已开启': '未开启'}}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">网关熔断</dt>
 <dd class="font-semibold">{{ circuit.enabled ? '已开启': '未开启'}}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">BIN 失败阈值</dt>
            <dd class="font-mono font-semibold">{{ number(binRateLimit.failure_threshold) }} / {{ seconds(binRateLimit.window_seconds) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">BIN 封锁时间</dt>
            <dd class="font-mono font-semibold">{{ seconds(binRateLimit.block_duration_seconds) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">网关失败率</dt>
            <dd class="font-mono font-semibold">≥ {{ percent(circuit.failure_rate_threshold) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="text-muted-foreground">熔断样本 / 开启</dt>
            <dd class="font-mono font-semibold">{{ number(circuit.minimum_sample_count) }} / {{ seconds(circuit.open_duration_seconds) }}</dd>
          </div>
        </dl>
      </section>
    </div>

    <section class="grid grid-cols-1 gap-3 xl:grid-cols-3">
      <section class="border border-dashed border-border/80 bg-card p-4">
        <h3 class="text-xs font-black uppercase tracking-widest">访客风险画像</h3>
        <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
          记录 IP / UA 的历史行为，给 3DS 决策提供观察信号；它本身不是 Stripe 的分数。
        </p>
        <dl class="mt-3 space-y-2 text-xs">
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">采集状态</dt>
 <dd class="font-semibold">{{ visitorRisk.enabled ? '已开启': '未开启'}}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">观察窗口</dt>
            <dd class="font-mono font-semibold">{{ number(threeDS.visitor_risk_lookback_days) }} 天</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">数据保留</dt>
            <dd class="font-mono font-semibold">{{ number(visitorRisk.retention_days) }} 天</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="text-muted-foreground">刷新 / 待处理上限</dt>
            <dd class="font-mono font-semibold">{{ seconds(visitorRisk.flush_interval_seconds) }} / {{ number(visitorRisk.max_pending_facts) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="text-muted-foreground">路径采样上限</dt>
            <dd class="font-mono font-semibold">{{ number(visitorRisk.sample_path_limit) }} 条</dd>
          </div>
        </dl>
      </section>

      <section class="border border-dashed border-border/80 bg-card p-4">
        <h3 class="text-xs font-black uppercase tracking-widest">站点挑战与反滥用</h3>
        <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
          当重复支付失败达到阈值时，支付接口先要求 Turnstile/站点挑战；这和银行 3DS 不是同一个验证。
        </p>
        <dl class="mt-3 space-y-2 text-xs">
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">Turnstile</dt>
 <dd class="font-semibold">{{ antiAbuse.turnstile_required ? (antiAbuse.turnstile_configured ? '已启用且已配置': '启用但缺密钥') : '未强制'}}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">支付失败窗口</dt>
            <dd class="font-mono font-semibold">{{ seconds(paymentRisk.failure_window_seconds) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">验证每日上限</dt>
            <dd class="font-mono font-semibold">{{ number(antiAbuse.verification_daily_limit) }} 次</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="text-muted-foreground">全局验证窗口</dt>
            <dd class="font-mono font-semibold">{{ seconds(antiAbuse.verification_global_window_seconds) }} · {{ number(antiAbuse.verification_global_limit) }} 次</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">重复失败 / 高风险</dt>
            <dd class="font-mono font-semibold">{{ number(paymentRisk.failure_threshold) }} 次 / ≥ {{ number(paymentRisk.high_risk_score) }} 分</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="text-muted-foreground">命中后处理</dt>
            <dd class="font-semibold">延迟 {{ seconds(paymentRisk.delay_seconds) }}，必要时站点挑战</dd>
          </div>
        </dl>
      </section>

      <section class="border border-dashed border-border/80 bg-card p-4">
        <h3 class="text-xs font-black uppercase tracking-widest">订单创建保护</h3>
        <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
          这是订单层的滥用限流，早于 PaymentIntent；它不是拒付率，也不会修改 Stripe 的 3DS。
        </p>
        <dl class="mt-3 space-y-2 text-xs">
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">状态</dt>
 <dd class="font-semibold">{{ orderAbuse.enabled ? '已开启': '未开启'}}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">计数窗口</dt>
            <dd class="font-mono font-semibold">{{ seconds(orderAbuse.order_create_window_seconds) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3 border-b border-dashed border-border/70 pb-2">
            <dt class="text-muted-foreground">用户 / 会话</dt>
            <dd class="font-mono font-semibold">{{ number(orderAbuse.max_order_creations_per_user) }} / {{ number(orderAbuse.max_order_creations_per_session) }}</dd>
          </div>
          <div class="flex items-start justify-between gap-3">
            <dt class="text-muted-foreground">IP 上限</dt>
            <dd class="font-mono font-semibold">{{ number(orderAbuse.max_order_creations_per_ip) }} 次</dd>
          </div>
        </dl>
      </section>
    </section>

    <section class="border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h3 class="text-xs font-black uppercase tracking-widest">一次支付的实际决策链路</h3>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
            3DS 不是第一道拦截。页面上的“3DS 升级率”只统计已经进入支付创建并被记录的最终模式；
            EFW 到达后当前代码会写入风险事件并生成退款建议/人工处理任务，不会在 webhook 里自动退款。
          </p>
        </div>
 <span :class="circuitOpen ? 'border-rose-500/25 bg-rose-500/10 text-rose-700': 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'" class="border px-2 py-1 text-[10px] font-black uppercase tracking-wider">
          Stripe 网关 {{ circuitOpen ? '当前熔断' : '当前可用' }}
        </span>
      </div>
      <ol class="mt-4 grid grid-cols-1 gap-2 md:grid-cols-2 xl:grid-cols-4">
        <li v-for="step in decisionSteps" :key="step.index" class="border border-border/70 p-3">
          <div class="flex items-center gap-2">
            <span class="flex size-6 items-center justify-center rounded-full bg-primary text-[10px] font-black text-primary-foreground">{{ step.index }}</span>
            <span class="text-xs font-black">{{ step.title }}</span>
          </div>
          <p class="mt-2 text-[11px] leading-5 text-muted-foreground">{{ step.description }}</p>
        </li>
      </ol>
    </section>

    <section class="border border-dashed border-border/80 bg-card p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h3 class="text-xs font-black uppercase tracking-widest">风控监控本身</h3>
          <p class="mt-1 text-[11px] leading-5 text-muted-foreground">
            这是 30 天滚动统计和自动升级的控制面，不等于支付创建时的单笔风险分数。
          </p>
        </div>
        <span :class="statusClass(Boolean(monitoring.worker_enabled))" class="border px-2 py-1 text-[10px] font-black">
          定时重算 {{ monitoring.worker_enabled ? '已开启' : '未开启' }}
        </span>
      </div>
      <dl class="mt-3 grid grid-cols-2 gap-px overflow-hidden border border-border/70 bg-border/70 md:grid-cols-4">
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">滚动窗口</dt>
          <dd class="mt-1 font-mono text-sm font-semibold">{{ number(monitoring.window_days) }} 天</dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">最小样本</dt>
          <dd class="mt-1 font-mono text-sm font-semibold">{{ number(monitoring.minimum_successful_payments) }} 笔</dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">预警 / 严重争议率</dt>
          <dd class="mt-1 font-mono text-sm font-semibold">{{ percent(monitoring.warning_dispute_activity_rate) }} / {{ percent(monitoring.critical_dispute_activity_rate) }}</dd>
        </div>
        <div class="bg-card p-3">
          <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground">自动升级</dt>
 <dd class="mt-1 text-sm font-semibold">{{ monitoring.auto_step_up_enabled ? '风险快照异常时开启': '关闭'}}</dd>
        </div>
      </dl>
      <p class="mt-3 text-[11px] text-muted-foreground">
        监控任务间隔：{{ seconds(monitoring.worker_interval_seconds) }} · 退款阈值 {{ percent(monitoring.warning_refund_rate) }} / {{ percent(monitoring.critical_refund_rate) }} · EFW 阈值 {{ percent(monitoring.warning_early_fraud_rate) }} / {{ percent(monitoring.critical_early_fraud_rate) }}
      </p>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type {
  PaymentGatewayHealthMap,
  PaymentGatewayRuntime,
  PaymentGatewayRuntimeStatus,
  PaymentRiskConfiguration,
} from './paymentRiskTypes'

const props = withDefaults(defineProps<{
  configuration?: PaymentRiskConfiguration | null
  gatewayRuntime?: PaymentGatewayRuntime | null
  gatewayHealth?: PaymentGatewayHealthMap
}>(), {
  configuration: null,
  gatewayRuntime: null,
  gatewayHealth: () => ({}),
})

const configuration = computed(() => props.configuration || {})
const threeDS = computed(() => configuration.value.three_ds || {})
const paymentRisk = computed(() => configuration.value.payment_risk || {})
const binRateLimit = computed(() => configuration.value.bin_rate_limit || {})
const circuit = computed(() => configuration.value.gateway_circuit_breaker || {})
const protection = computed(() => configuration.value.protection || {})
const monitoring = computed(() => configuration.value.monitoring || {})
const antiAbuse = computed(() => configuration.value.anti_abuse || {})
const orderAbuse = computed(() => configuration.value.order_abuse || {})
const visitorRisk = computed(() => configuration.value.visitor_risk || {})
const stripeRuntime = computed<PaymentGatewayRuntimeStatus | null>(() => (
  props.gatewayRuntime?.gateways?.find((gateway) => gateway.provider === 'stripe') || null
))
const stripeHealth = computed(() => props.gatewayHealth?.stripe || {})
const circuitOpen = computed(() => Boolean(stripeHealth.value.circuit_open))
const stripeMode = computed(() => stripeRuntime.value?.three_ds_mode || '')
const gatewayCards = computed(() => ['stripe', 'paypal'].map((provider) => ({
  provider,
  label: provider === 'stripe' ? 'Stripe' : 'PayPal',
  runtime: props.gatewayRuntime?.gateways?.find((gateway) => gateway.provider === provider) || null,
  health: props.gatewayHealth?.[provider] || null,
})))

const decisionSteps = [
  { index: 1, title: '订单可支付', description: '先检查订单是否已支付、取消或过期；命中人工暂停新支付时在这里直接结束。' },
  { index: 2, title: '单笔预检', description: '按用户、会话、匿名标识、IP 和 UA 读取失败计数，叠加国家不一致等信号，必要时延迟或要求站点挑战。' },
  { index: 3, title: 'BIN 与熔断', description: '检查卡 BIN 是否暂时封锁，再读取网关最近窗口失败率；熔断时不会创建新的 PaymentIntent。' },
  { index: 4, title: '最终 3DS', description: '以 Stripe 基础模式为起点，合并人工保护、30 天组合风险、单笔分数、访客画像和客户历史，模式只向 any/challenge 提升。' },
  { index: 5, title: '支付与记录', description: '使用最终模式创建 PaymentIntent，并把策略、风险分数、原因和豁免候选写入风险决策表。' },
  { index: 6, title: '事后事件', description: '支付失败、EFW、争议和退款再通过 webhook/数据库进入 30 天快照；EFW 当前生成退款建议，需人工确认执行。' },
]

const threeDSModeCards = [
  {
    value: 'automatic',
    label: '自动判断',
    description: '以 Stripe 的自动 SCA/3DS 判断为基础，系统不会因为这个基础值强制所有买家验证。',
    outcome: '可能直接完成、走无感 3DS，或由风险条件升级。',
  },
  {
    value: 'any',
    label: '要求 3DS',
    description: '请求在支持的情况下使用 3DS；认证可能无感完成，也可能需要买家完成银行验证。',
    outcome: '认证强度高于 automatic，但不等于每笔都会出现 challenge 弹窗。',
  },
  {
    value: 'challenge',
    label: '更强挑战',
    description: '请求发卡行进入更强的挑战路径，适合风险升高或人工保护时使用。',
    outcome: '更可能出现银行挑战，但最终是否展示以及如何验证仍由发卡行决定。',
  },
]

const number = (value: unknown): string => {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) ? parsed.toLocaleString('zh-CN') : '0'
}
const percent = (value: unknown): string => {
  const parsed = Number(value || 0)
  return `${(Number.isFinite(parsed) ? parsed * 100 : 0).toFixed(2)}%`
}
const seconds = (value: unknown): string => {
  const parsed = Number(value || 0)
  if (!Number.isFinite(parsed) || parsed <= 0) return '未设置'
  if (parsed >= 86400 && parsed % 86400 === 0) return `${parsed / 86400} 天`
  if (parsed >= 3600 && parsed % 3600 === 0) return `${parsed / 3600} 小时`
  if (parsed >= 60 && parsed % 60 === 0) return `${parsed / 60} 分钟`
  return `${parsed} 秒`
}
const money = (value: unknown): string => {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) && parsed > 0 ? `≤ ${parsed.toFixed(2)}（订单币种）` : '未设置'
}
const statusClass = (enabled: boolean): string => (
  enabled
    ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'
    : 'border-amber-500/25 bg-amber-500/10 text-amber-700'
)
const modeLabel = (mode: string): string => ({
  automatic: '自动判断（automatic）',
  any: '要求 3DS（any）',
  challenge: '更强挑战（challenge）',
}[mode] || mode || '未读取')
const modeDescription = (mode: string): string => ({
  automatic: '自动判断（automatic）：把是否需要 SCA/3DS 的最终判断交给 Stripe；本系统仍可能因风险把模式升级。',
  any: '要求 3DS（any）：请求在支持的情况下使用 3DS，但可能以无感方式完成，也可能进入银行 challenge。',
  challenge: '更强挑战（challenge）：请求发卡行进入更强的挑战路径，但最终交互仍由发卡行决定。',
}[mode] || '当前没有读取到 Stripe 的基础 3DS 模式。')
const gatewayStatusLabel = (gateway: { runtime?: PaymentGatewayRuntimeStatus | null; health?: { allowed?: boolean; circuit_open?: boolean } | null }): string => {
  if (gateway.health?.circuit_open) return '熔断中'
  if (gateway.health?.allowed === false) return '当前拦截'
  if (gateway.runtime?.production_ready) return '运行就绪'
  if (gateway.runtime?.configured || gateway.runtime?.webhook_configured) return '配置未完成'
  return '未配置'
}
const gatewayStatusClass = (gateway: { runtime?: PaymentGatewayRuntimeStatus | null; health?: { allowed?: boolean; circuit_open?: boolean } | null }): string => {
  if (gateway.health?.circuit_open || gateway.health?.allowed === false) return 'border-rose-500/25 bg-rose-500/10 text-rose-700'
  if (gateway.runtime?.production_ready) return 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700'
  return 'border-amber-500/25 bg-amber-500/10 text-amber-700'
}
const runtimeSourceLabel = (source?: string): string => ({
  environment: '环境变量',
  'admin-encrypted': '后台加密配置',
  mixed: '混合来源',
}[source || ''] || source || '未知')
</script>
