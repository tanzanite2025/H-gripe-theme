<template>
  <section class="account-tab-panel referral-tab">
    <div v-if="loading" class="referral-state" role="status">
      <Icon name="lucide:loader-circle" class="referral-state__spin" />
      <span>{{ t('accountSidebar.referral.loading', 'Loading your referral center...') }}</span>
    </div>

    <div v-else-if="error" class="referral-state referral-state--error" role="alert">
      <Icon name="lucide:triangle-alert" />
      <span>{{ error }}</span>
      <button type="button" class="referral-icon-button" :aria-label="t('accountSidebar.actions.refresh', 'Refresh')" @click="loadReferralData">
        <Icon name="lucide:refresh-cw" />
      </button>
    </div>

    <template v-else-if="dashboard?.enabled">
      <div class="referral-hero">
        <div class="referral-hero__copy">
          <p>{{ t('accountSidebar.referral.kicker', 'Refer a friend') }}</p>
          <h3>{{ t('accountSidebar.referral.title', 'Share the ride') }}</h3>
          <span>{{ t('accountSidebar.referral.description', 'Give a friend a welcome benefit and keep reward points moving toward your next build.') }}</span>
        </div>
        <Icon name="lucide:users-round" class="referral-hero__icon" aria-hidden="true" />
      </div>

      <div class="referral-code-block">
        <div>
          <span>{{ t('accountSidebar.referral.yourCode', 'Your referral code') }}</span>
          <strong>{{ dashboard.referral_code || '—' }}</strong>
        </div>
        <button
          type="button"
          class="referral-copy-button"
          :aria-label="copied ? t('accountSidebar.referral.copied', 'Referral link copied') : t('accountSidebar.referral.copy', 'Copy referral link')"
          :title="copied ? t('accountSidebar.referral.copied', 'Referral link copied') : t('accountSidebar.referral.copy', 'Copy referral link')"
          :disabled="!dashboard.share_url"
          @click="copyReferralLink"
        >
          <Icon :name="copied ? 'lucide:check' : 'lucide:copy'" />
        </button>
      </div>

      <div class="referral-stats" aria-label="Referral summary">
        <div>
          <span>{{ t('accountSidebar.referral.invited', 'Invited') }}</span>
          <strong>{{ dashboard.stats?.total_invited_count || 0 }}</strong>
        </div>
        <div>
          <span>{{ t('accountSidebar.referral.successful', 'Successful') }}</span>
          <strong>{{ dashboard.stats?.successful_orders_count || 0 }}</strong>
        </div>
        <div>
          <span>{{ t('accountSidebar.referral.pendingPoints', 'Pending pts') }}</span>
          <strong>{{ dashboard.stats?.pending_reward_points || 0 }}</strong>
        </div>
        <div>
          <span>{{ t('accountSidebar.referral.settledPoints', 'Settled pts') }}</span>
          <strong>{{ dashboard.stats?.settled_reward_points || 0 }}</strong>
        </div>
      </div>

      <div class="referral-rule-strip">
        <Icon name="lucide:shield-check" aria-hidden="true" />
        <span>{{ ruleSummary }}</span>
      </div>

      <div class="referral-history">
        <div class="referral-history__header">
          <div>
            <p>{{ t('accountSidebar.referral.historyKicker', 'Activity') }}</p>
            <h4>{{ t('accountSidebar.referral.historyTitle', 'Referral history') }}</h4>
          </div>
          <button type="button" class="referral-icon-button" :aria-label="t('accountSidebar.actions.refresh', 'Refresh')" :disabled="historyLoading" @click="loadReferralData">
            <Icon name="lucide:refresh-cw" :class="{ 'referral-state__spin': historyLoading }" />
          </button>
        </div>

        <div v-if="historyLoading" class="referral-history__empty" role="status">
          {{ t('common.loading', 'Loading...') }}
        </div>
        <div v-else-if="history.items.length === 0" class="referral-history__empty">
          <Icon name="lucide:route" aria-hidden="true" />
          <span>{{ t('accountSidebar.referral.empty', 'Your first referral will appear here.') }}</span>
        </div>
        <ul v-else class="referral-history__list">
          <li v-for="item in history.items" :key="item.id">
            <div class="referral-history__item-main">
              <strong>{{ item.referee_name_mask || t('accountSidebar.referral.newCustomer', 'New customer') }}</strong>
              <span>{{ formatDate(item.order_date || item.vesting_until) }}</span>
            </div>
            <div class="referral-history__item-meta">
              <span class="referral-status" :class="`referral-status--${item.status}`">{{ statusLabel(item.status) }}</span>
              <strong>{{ item.potential_reward_points || 0 }} {{ t('member.points.unit', 'pts') }}</strong>
            </div>
          </li>
        </ul>
      </div>
    </template>

    <div v-else class="referral-state referral-state--disabled">
      <Icon name="lucide:megaphone-off" />
      <strong>{{ t('accountSidebar.referral.disabledTitle', 'Referral center is paused') }}</strong>
      <span>{{ t('accountSidebar.referral.disabledDescription', 'This program is not available right now. Your account and points remain unchanged.') }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'

const props = withDefaults(defineProps<{ active?: boolean }>(), { active: true })
const { t } = useI18n()
const { request } = useApiRequest()

interface ReferralDashboard {
  enabled?: boolean
  referral_code?: string
  share_url?: string
  reward_rules?: {
    currency?: string
    min_order_amount_minor?: number
    referrer_reward_points?: number
    vesting_period_days?: number
  }
  stats?: {
    total_invited_count?: number
    successful_orders_count?: number
    pending_reward_points?: number
    settled_reward_points?: number
  }
}

interface ReferralHistoryItem {
  id: number
  referee_name_mask?: string
  status: string
  order_date?: string
  potential_reward_points?: number
  vesting_until?: string
}

interface ReferralHistory {
  items: ReferralHistoryItem[]
  total?: number
}

const dashboard = ref<ReferralDashboard | null>(null)
const history = ref<ReferralHistory>({ items: [] })
const loading = ref(false)
const historyLoading = ref(false)
const error = ref('')
const copied = ref(false)

const unwrap = <T,>(payload: T | { data?: T }) => {
  if (payload && typeof payload === 'object' && 'data' in payload && payload.data !== undefined) {
    return payload.data as T
  }
  return payload as T
}

const ruleSummary = computed(() => {
  const rules = dashboard.value?.reward_rules
  if (!rules) return t('accountSidebar.referral.ruleFallback', 'Rewards are held until the order clears its delivery and cooling-off checks.')
  const points = Number(rules.referrer_reward_points || 0)
  const days = Number(rules.vesting_period_days || 0)
  const translated = t('accountSidebar.referral.ruleSummary', { points, days })
  return translated === 'accountSidebar.referral.ruleSummary'
    ? `Earn ${points} pts after delivery plus a ${days}-day cooling-off period.`
    : translated
})

const loadReferralData = async () => {
  if (loading.value) return
  loading.value = true
  error.value = ''
  try {
    const dashboardPayload = await request<ReferralDashboard | { data?: ReferralDashboard }>('/customer/referral/me')
    dashboard.value = unwrap(dashboardPayload)
    historyLoading.value = true
    if (dashboard.value?.enabled) {
      const historyPayload = await request<ReferralHistory | { data?: ReferralHistory }>('/customer/referral/history', {
        params: { page: 1, page_size: 8 },
      })
      history.value = unwrap(historyPayload) || { items: [] }
      history.value.items = Array.isArray(history.value.items) ? history.value.items : []
    } else {
      history.value = { items: [] }
    }
  } catch (requestError) {
    error.value = requestError instanceof Error
      ? requestError.message
      : t('accountSidebar.referral.loadFailed', 'Referral data could not be loaded.')
  } finally {
    loading.value = false
    historyLoading.value = false
  }
}

const copyReferralLink = async () => {
  const value = dashboard.value?.share_url
  if (!value || typeof navigator === 'undefined' || !navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(value)
    copied.value = true
    window.setTimeout(() => { copied.value = false }, 1800)
  } catch {
    copied.value = false
  }
}

const statusLabel = (status: string) => {
  const labels: Record<string, string> = {
    pending: t('accountSidebar.referral.status.pending', 'Pending'),
    ordered: t('accountSidebar.referral.status.ordered', 'Order paid'),
    vesting: t('accountSidebar.referral.status.vesting', 'Cooling off'),
    settled: t('accountSidebar.referral.status.settled', 'Settled'),
    expired: t('accountSidebar.referral.status.expired', 'Expired'),
    revoked: t('accountSidebar.referral.status.revoked', 'Revoked'),
    reversed: t('accountSidebar.referral.status.reversed', 'Reversed'),
  }
  return labels[status] || status
}

const formatDate = (value?: string) => {
  if (!value) return t('accountSidebar.referral.awaitingOrder', 'Awaiting first order')
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(undefined, { year: 'numeric', month: 'short', day: 'numeric' }).format(date)
}

onMounted(() => {
  if (props.active) loadReferralData()
})

watch(() => props.active, active => {
  if (active && !dashboard.value && !loading.value) loadReferralData()
})
</script>

<style scoped>
.referral-tab { min-width: 0; }
.referral-hero, .referral-code-block, .referral-rule-strip, .referral-history {
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-card-surface);
}
.referral-hero {
  display: flex; align-items: center; justify-content: space-between; gap: 0.8rem;
  border-radius: 1.35rem; padding: 1rem;
}
.referral-hero__copy { min-width: 0; }
.referral-hero p, .referral-history__header p {
  margin: 0; color: #059669; font-size: var(--tz-type-micro-label); font-weight: 850;
  letter-spacing: 0.12em; text-transform: uppercase;
}
.referral-hero h3, .referral-history h4 { margin: 0.25rem 0 0; color: var(--tz-text-primary); font-size: 1.18rem; font-weight: 900; }
.referral-hero span { display: block; margin-top: 0.45rem; color: var(--tz-text-secondary); font-size: 0.78rem; line-height: 1.45; }
.referral-hero__icon { width: 2.2rem; height: 2.2rem; flex: 0 0 auto; color: #059669; }
.referral-code-block { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; border-radius: 1rem; background: var(--tz-input-surface); padding: 0.8rem 0.9rem; }
.referral-code-block span, .referral-code-block strong { display: block; }
.referral-code-block span { color: var(--tz-text-muted); font-size: var(--tz-type-micro-label); font-weight: 750; }
.referral-code-block strong { margin-top: 0.25rem; color: var(--tz-text-primary); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 1.12rem; letter-spacing: 0.12em; }
.referral-copy-button, .referral-icon-button { display: inline-flex; align-items: center; justify-content: center; width: 2.25rem; height: 2.25rem; flex: 0 0 auto; border: 1px solid var(--tz-border-subtle); border-radius: 0.75rem; background: var(--tz-card-surface); color: var(--tz-text-secondary); }
.referral-copy-button:hover:not(:disabled), .referral-icon-button:hover:not(:disabled) { color: #059669; border-color: #059669; }
.referral-copy-button :deep(svg), .referral-icon-button :deep(svg) { width: 1rem; height: 1rem; }
.referral-stats { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.55rem; }
.referral-stats > div { min-height: 3.8rem; border: 1px solid var(--tz-border-subtle); border-radius: 0.95rem; background: var(--tz-input-surface); padding: 0.7rem; }
.referral-stats span, .referral-stats strong { display: block; }
.referral-stats span { color: var(--tz-text-muted); font-size: var(--tz-type-micro-label); font-weight: 750; }
.referral-stats strong { margin-top: 0.3rem; color: var(--tz-text-primary); font-size: 1.05rem; font-weight: 900; }
.referral-rule-strip { display: flex; align-items: flex-start; gap: 0.5rem; border-radius: 0.95rem; background: color-mix(in srgb, #059669 7%, var(--tz-card-surface)); padding: 0.75rem; color: var(--tz-text-secondary); font-size: 0.75rem; line-height: 1.4; }
.referral-rule-strip :deep(svg) { width: 1rem; height: 1rem; flex: 0 0 auto; color: #059669; }
.referral-history { border-radius: 1.1rem; padding: 0.85rem; }
.referral-history__header { display: flex; align-items: center; justify-content: space-between; gap: 0.6rem; }
.referral-history h4 { font-size: 1rem; }
.referral-history__list { display: grid; gap: 0.55rem; margin: 0.75rem 0 0; padding: 0; list-style: none; }
.referral-history__list li { display: flex; align-items: center; justify-content: space-between; gap: 0.6rem; border-top: 1px solid var(--tz-border-subtle); padding-top: 0.65rem; }
.referral-history__item-main, .referral-history__item-meta { display: grid; min-width: 0; gap: 0.18rem; }
.referral-history__item-main strong, .referral-history__item-meta strong { overflow: hidden; color: var(--tz-text-primary); font-size: 0.78rem; font-weight: 850; text-overflow: ellipsis; white-space: nowrap; }
.referral-history__item-main span, .referral-history__item-meta > strong { color: var(--tz-text-muted); font-size: var(--tz-type-micro-label); }
.referral-history__item-meta { justify-items: end; flex: 0 0 auto; }
.referral-status { display: inline-flex; min-height: 1.25rem; align-items: center; border-radius: 999px; padding: 0 0.45rem; font-size: 0.62rem; font-weight: 850; text-transform: uppercase; }
.referral-status--pending, .referral-status--ordered, .referral-status--vesting { background: #fef3c7; color: #92400e; }
.referral-status--settled { background: #d1fae5; color: #065f46; }
.referral-status--expired, .referral-status--revoked, .referral-status--reversed { background: #fee2e2; color: #991b1b; }
.referral-history__empty, .referral-state { display: flex; min-height: 8rem; align-items: center; justify-content: center; gap: 0.55rem; border: 1px dashed var(--tz-border-subtle); border-radius: 1.1rem; color: var(--tz-text-secondary); font-size: 0.78rem; line-height: 1.45; text-align: center; }
.referral-state { flex-wrap: wrap; align-content: center; padding: 1rem; }
.referral-state--disabled { flex-direction: column; }
.referral-state--disabled strong { color: var(--tz-text-primary); }
.referral-state--error { color: #991b1b; }
.referral-state__spin { animation: referral-spin 0.9s linear infinite; }
@keyframes referral-spin { to { transform: rotate(360deg); } }
</style>
