<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700">
    <header class="rounded-[32px] border border-dashed border-border/80 bg-muted/5 p-6">
      <div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
        <div>
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">
            {{ t('payment.onboardingGuide.eyebrow') }}
          </span>
          <h1 class="mt-1 text-lg font-black uppercase italic tracking-tighter text-foreground">
            {{ providerLabel }} {{ t('payment.onboardingGuide.titleSuffix') }}
          </h1>
          <p class="mt-1 max-w-3xl text-[10px] leading-relaxed tracking-wide text-muted-foreground">
            {{ t('payment.onboardingGuide.description') }}
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <span class="rounded-full border border-admin-selected-border/60 bg-admin-selected-soft/50 px-3 py-1 text-[9px] font-black uppercase tracking-widest text-admin-selected">
            {{ t('payment.onboardingGuide.readOnly') }}
          </span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="h-9 rounded-full px-3 text-[10px] font-black uppercase"
            :disabled="!webhookURL"
            @click="copyWebhookURL"
          >
            <CheckCircle2 v-if="copied" class="size-3.5 text-emerald-500" />
            <Copy v-else class="size-3.5" />
            {{ copied ? t('payment.onboardingGuide.copied') : t('payment.onboardingGuide.copyWebhook') }}
          </Button>
          <a
            v-if="guideProvider"
            :href="consoleURL"
            target="_blank"
            rel="noreferrer"
            class="inline-flex h-9 items-center gap-1.5 rounded-full bg-foreground px-3 text-[10px] font-black uppercase text-background transition-opacity hover:opacity-85"
          >
            <ExternalLink class="size-3.5" />
            {{ t('payment.onboardingGuide.openConsole') }}
          </a>
        </div>
      </div>

      <div class="mt-5 grid gap-3 sm:grid-cols-3">
        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">
            {{ t('payment.onboardingGuide.provider') }}
          </p>
          <p class="mt-1 text-sm font-black text-foreground">{{ providerLabel }}</p>
        </div>
        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">
            {{ t('payment.onboardingGuide.runtime') }}
          </p>
          <p class="mt-1 text-sm font-black" :class="status?.production_ready ? 'text-emerald-600 dark:text-emerald-300' : 'text-amber-600 dark:text-amber-300'">
            {{ status?.production_ready ? t('payment.productionReady') : t('payment.pendingConfig') }}
          </p>
        </div>
        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">
            {{ t('payment.onboardingGuide.webhookURL') }}
          </p>
          <code class="mt-1 block truncate text-[11px] font-bold text-foreground" :title="webhookURL || t('payment.onboardingGuide.webhookMissing')">
            {{ webhookURL || t('payment.onboardingGuide.webhookMissing') }}
          </code>
        </div>
      </div>
    </header>

    <div v-if="!guideProvider" class="rounded-[24px] border border-dashed border-amber-500/30 bg-amber-500/10 p-6">
      <div class="flex items-start gap-3">
        <AlertTriangle class="mt-0.5 size-5 shrink-0 text-amber-600 dark:text-amber-300" />
        <div>
          <h2 class="text-sm font-black text-foreground">{{ t('payment.onboardingGuide.unsupportedTitle') }}</h2>
          <p class="mt-1 text-xs leading-relaxed text-muted-foreground">{{ t('payment.onboardingGuide.unsupportedDescription') }}</p>
        </div>
      </div>
    </div>

    <template v-else>
      <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
        <section class="flex flex-col justify-between rounded-[24px] border border-dashed border-border/80 bg-card p-6 shadow-sm">
          <div class="space-y-4">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">01</p>
                <h2 class="mt-1 text-sm font-black italic tracking-tighter text-foreground">
                  {{ t('payment.onboardingGuide.settlementTitle') }}
                </h2>
              </div>
              <span class="rounded-full bg-rose-500/10 px-2 py-0.5 text-[8px] font-mono font-bold uppercase text-rose-600 dark:text-rose-300">
                {{ t('payment.onboardingGuide.critical') }}
              </span>
            </div>
            <p class="text-xs leading-relaxed text-muted-foreground">
              {{ guideProvider === 'paypal' ? t('payment.onboardingGuide.paypalSettlement') : t('payment.onboardingGuide.stripeSettlement') }}
            </p>
            <ol class="space-y-2 text-[11px] leading-relaxed text-muted-foreground">
              <li v-for="(step, index) in settlementSteps" :key="step" class="flex items-start gap-2">
                <span class="font-mono font-black text-admin-selected">{{ index + 1 }}.</span>
                <span>{{ step }}</span>
              </li>
            </ol>
            <div class="rounded-xl bg-muted/40 p-3 text-[11px] leading-relaxed">
              <p class="font-mono text-muted-foreground">
                <span class="font-black uppercase tracking-widest">{{ t('payment.onboardingGuide.path') }}:</span>
                {{ guideProvider === 'paypal' ? t('payment.onboardingGuide.paypalSettlementPath') : t('payment.onboardingGuide.stripeSettlementPath') }}
              </p>
              <p class="mt-2 flex items-start gap-1.5 font-bold text-emerald-700 dark:text-emerald-300">
                <CheckCircle2 class="mt-0.5 size-3.5 shrink-0" />
                {{ t('payment.onboardingGuide.settlementOutcome') }}
              </p>
            </div>
          </div>
        </section>

        <section class="rounded-[24px] border border-dashed border-border/80 bg-card p-6 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">02</p>
              <h2 class="mt-1 text-sm font-black italic tracking-tighter text-foreground">
                {{ t('payment.onboardingGuide.webhookTitle') }}
              </h2>
            </div>
            <span class="rounded-full bg-rose-500/10 px-2 py-0.5 text-[8px] font-mono font-bold uppercase text-rose-600 dark:text-rose-300">
              {{ eventCount }} {{ t('payment.onboardingGuide.events') }}
            </span>
          </div>
          <p class="mt-3 text-xs leading-relaxed text-muted-foreground">
            {{ t('payment.onboardingGuide.webhookDescription') }}
          </p>
          <ul class="mt-4 grid gap-2 sm:grid-cols-2">
            <li
              v-for="event in currentEvents"
              :key="event"
              class="flex min-w-0 items-start gap-2 rounded-lg border bg-muted/20 px-2.5 py-2"
            >
              <CheckCircle2 class="mt-0.5 size-3.5 shrink-0 text-emerald-500" />
              <code class="break-all text-[10px] font-bold leading-relaxed text-foreground">{{ event }}</code>
            </li>
          </ul>
          <div class="mt-4 rounded-xl border border-admin-selected-border/40 bg-admin-selected-soft/30 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.onboardingGuide.webhookURL') }}</p>
            <p class="mt-1 break-all font-mono text-[11px] font-bold text-foreground">{{ webhookURL || t('payment.onboardingGuide.webhookMissing') }}</p>
            <p class="mt-2 text-[11px] leading-relaxed text-amber-700 dark:text-amber-300">
              {{ status?.webhook_events_verified ? 'Webhook 事件清单已确认。' : '系统无法读取官方控制台订阅状态；请在官方后台逐项核对并完成事件清单。' }}
            </p>
            <a
              v-if="status?.webhook_event_setup_url"
              :href="status.webhook_event_setup_url"
              target="_blank"
              rel="noreferrer"
              class="mt-2 inline-flex items-center gap-1 text-[10px] font-black uppercase text-admin-selected hover:underline"
            >
              <ExternalLink class="size-3" />
              {{ t('payment.onboardingGuide.openConsole') }}
            </a>
          </div>
        </section>

        <section class="rounded-[24px] border border-dashed border-border/80 bg-card p-6 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">03</p>
              <h2 class="mt-1 text-sm font-black italic tracking-tighter text-foreground">
                {{ guideProvider === 'paypal' ? t('payment.onboardingGuide.paypalProtectionTitle') : t('payment.onboardingGuide.stripeRiskTitle') }}
              </h2>
            </div>
            <ShieldCheck class="size-5 text-admin-selected" />
          </div>
          <p class="mt-3 text-xs leading-relaxed text-muted-foreground">
            {{ guideProvider === 'paypal' ? t('payment.onboardingGuide.paypalProtectionDescription') : t('payment.onboardingGuide.stripeRiskDescription') }}
          </p>
          <ul class="mt-4 space-y-2 text-[11px] leading-relaxed text-muted-foreground">
            <li v-for="rule in riskRules" :key="rule" class="flex items-start gap-2">
              <CheckCircle2 class="mt-0.5 size-3.5 shrink-0 text-emerald-500" />
              <span>{{ rule }}</span>
            </li>
          </ul>
        </section>

        <section class="rounded-[24px] border border-dashed border-border/80 bg-card p-6 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">04</p>
              <h2 class="mt-1 text-sm font-black italic tracking-tighter text-foreground">
                {{ t('payment.onboardingGuide.billingTitle') }}
              </h2>
            </div>
            <KeyRound class="size-5 text-admin-selected" />
          </div>
          <p class="mt-3 text-xs leading-relaxed text-muted-foreground">
            {{ guideProvider === 'paypal' ? t('payment.onboardingGuide.paypalBilling') : t('payment.onboardingGuide.stripeBilling') }}
          </p>
          <div class="mt-4 rounded-xl bg-muted/40 p-3 text-[11px] leading-relaxed text-muted-foreground">
            <p class="font-black uppercase tracking-widest text-foreground">{{ t('payment.onboardingGuide.beforeGoLive') }}</p>
            <p class="mt-1">{{ t('payment.onboardingGuide.addressOutcome') }}</p>
          </div>
        </section>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import { AlertTriangle, CheckCircle2, Copy, ExternalLink, KeyRound, ShieldCheck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { useAdminI18n } from '@/i18n'
import type { PaymentGatewayRuntimeStatus } from '@/modules/settings/types'

const props = withDefaults(defineProps<{
  selectedGateway?: string
  status?: PaymentGatewayRuntimeStatus | null
}>(), {
  selectedGateway: 'stripe',
  status: null,
})

const { t } = useAdminI18n()
const copied = ref(false)

const guideProvider = computed<'stripe' | 'paypal' | null>(() => {
  const provider = String(props.selectedGateway || '').trim().toLowerCase()
  return provider === 'stripe' || provider === 'paypal' ? provider : null
})

const providerLabel = computed(() => {
  if (guideProvider.value === 'stripe') return t('payment.stripe')
  if (guideProvider.value === 'paypal') return t('payment.paypal')
  return String(props.selectedGateway || '').trim() || t('payment.selectGateway')
})

const webhookURL = computed(() => String(props.status?.callback_url || '').trim())
const consoleURL = computed(() => guideProvider.value === 'paypal'
  ? 'https://developer.paypal.com/dashboard/applications/live'
  : 'https://dashboard.stripe.com/webhooks')

const stripeEvents = [
  'payment_intent.succeeded',
  'payment_intent.payment_failed',
  'payment_intent.requires_action',
  'payment_intent.processing',
  'charge.refunded',
  'refund.created',
  'refund.updated',
  'charge.dispute.created',
  'charge.dispute.updated',
  'charge.dispute.funds_withdrawn',
  'charge.dispute.funds_reinstated',
  'charge.dispute.closed',
  'radar.early_fraud_warning.created',
  'review.opened / review.closed',
]

const paypalEvents = [
  'CHECKOUT.ORDER.APPROVED',
  'PAYMENT.CAPTURE.COMPLETED',
  'PAYMENT.CAPTURE.DENIED',
  'PAYMENT.CAPTURE.REFUNDED',
  'CUSTOMER.DISPUTE.CREATED',
  'CUSTOMER.DISPUTE.RESOLVED',
  'CUSTOMER.DISPUTE.UPDATED',
  'RISK.DISPUTE.CREATED',
]

const currentEvents = computed(() => {
  if (props.status?.webhook_event_checklist?.length) return props.status.webhook_event_checklist
  if (props.status?.required_webhook_events?.length) return props.status.required_webhook_events
  return guideProvider.value === 'paypal' ? paypalEvents : stripeEvents
})
const eventCount = computed(() => currentEvents.value.length)

const settlementSteps = computed(() => guideProvider.value === 'paypal'
  ? [
      t('payment.onboardingGuide.paypalSettlementStep1'),
      t('payment.onboardingGuide.paypalSettlementStep2'),
      t('payment.onboardingGuide.paypalSettlementStep3'),
    ]
  : [
      t('payment.onboardingGuide.stripeSettlementStep1'),
      t('payment.onboardingGuide.stripeSettlementStep2'),
      t('payment.onboardingGuide.stripeSettlementStep3'),
    ])

const riskRules = computed(() => guideProvider.value === 'paypal'
  ? [
      t('payment.onboardingGuide.paypalProtectionRule1'),
      t('payment.onboardingGuide.paypalProtectionRule2'),
      t('payment.onboardingGuide.paypalProtectionRule3'),
    ]
  : [
      t('payment.onboardingGuide.stripeRiskRule1'),
      t('payment.onboardingGuide.stripeRiskRule2'),
      t('payment.onboardingGuide.stripeRiskRule3'),
    ])

const copyWebhookURL = async (): Promise<void> => {
  if (!webhookURL.value || typeof navigator === 'undefined' || !navigator.clipboard) {
    toast.error(t('payment.onboardingGuide.copyFailed'))
    return
  }
  try {
    await navigator.clipboard.writeText(webhookURL.value)
    copied.value = true
    toast.success(t('payment.onboardingGuide.copied'))
    window.setTimeout(() => { copied.value = false }, 1800)
  } catch (error) {
    console.error('Failed to copy payment onboarding webhook URL:', error)
    toast.error(t('payment.onboardingGuide.copyFailed'))
  }
}
</script>
